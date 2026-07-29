package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/eval"
	"github.com/stuffbucket/vale/internal/fix"
	"github.com/stuffbucket/vale/internal/linter"
)

// fixSettings carries the resolved --fix options.
type fixSettings struct {
	cfg         *config.Config
	paths       []string
	model       string
	endpoint    string
	temperature *float64
	output      string
	maxTokens   int
}

// runFix rewrites one file with a model so it resolves its lint findings, then
// writes the corrected document to stdout, or to --output when set.
func runFix(s fixSettings) int {
	if len(s.paths) != 1 {
		fmt.Fprintln(os.Stderr, "fix: --fix takes exactly one file")
		return 2
	}
	path := s.paths[0]
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		return 2
	}

	endpoint := firstNonEmpty(s.endpoint, s.cfg.Model.Endpoint)
	model := firstNonEmpty(s.model, s.cfg.Model.Name)
	temp := s.temperature
	if temp == nil {
		temp = s.cfg.Model.Temperature
	}

	// Lint with the slop family on so the model fixes slop as well.
	lintCfg := *s.cfg
	lintCfg.Slop.Enabled = true
	findings := linter.New(&lintCfg).LintText(path, string(raw), linter.MarkdownAuto)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if isTerminal(os.Stderr) {
		fmt.Fprintf(os.Stderr, "fixing %s with %s (%d findings)…\n", path, model, len(findings))
	}
	fixed, err := fix.Fix(ctx, fix.Options{
		Client:      eval.NewClient(endpoint, os.Getenv("OPENAI_API_KEY")),
		Model:       model,
		Path:        path,
		Text:        string(raw),
		Findings:    findings,
		MaxTokens:   s.maxTokens,
		Temperature: temp,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		return 1
	}
	if !strings.HasSuffix(fixed, "\n") {
		fixed += "\n"
	}

	if s.output == "" {
		fmt.Print(fixed)
		return 0
	}
	if err := os.WriteFile(s.output, []byte(fixed), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", s.output)
	return 0
}

// firstNonEmpty returns a if it is non-empty, else b.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
