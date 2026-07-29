package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/eval"
	"github.com/stuffbucket/vale/internal/linter"
)

// cmdEval runs the eval subcommand: drive an LLM endpoint across model families
// and measure the slop in each model's output with the STE.Slop* rules.
func cmdEval(args []string) int {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	endpoint := fs.String("endpoint", eval.DefaultEndpoint, "OpenAI-compatible base URL")
	modelsFlag := fs.String("models", "", "comma-separated model ids; empty discovers all from the endpoint")
	promptsFile := fs.String("prompts", "", "file with one prompt per line; empty uses the built-in set")
	maxTokens := fs.Int("max-tokens", 400, "max tokens per completion")
	concurrency := fs.Int("concurrency", 4, "parallel requests")
	format := fs.String("format", "text", "output format: text or json")
	apiKey := fs.String("api-key", os.Getenv("OPENAI_API_KEY"), "bearer token (or the OPENAI_API_KEY env var)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "eval: unknown format %q\n", *format)
		return 2
	}

	prompts := eval.DefaultPrompts
	if *promptsFile != "" {
		loaded, err := readPrompts(*promptsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "eval: %v\n", err)
			return 2
		}
		prompts = loaded
	}

	// The linter runs with the opt-in slop family on — that is what we measure.
	cfg := config.Default()
	cfg.Slop.Enabled = true

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rep, err := eval.Run(ctx, eval.Options{
		Client:      eval.NewClient(*endpoint, *apiKey),
		Linter:      linter.New(cfg),
		Models:      splitCSV(*modelsFlag),
		Prompts:     prompts,
		MaxTokens:   *maxTokens,
		Concurrency: *concurrency,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "eval: %v\n", err)
		return 1
	}

	if *format == "json" {
		if err := eval.WriteJSON(os.Stdout, rep); err != nil {
			fmt.Fprintf(os.Stderr, "eval: %v\n", err)
			return 1
		}
	} else {
		eval.WriteText(os.Stdout, rep)
	}
	return 0
}

// splitCSV splits a comma list, trimming spaces and dropping empties.
func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// readPrompts reads one prompt per non-blank, non-comment line.
func readPrompts(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no prompts in %s", path)
	}
	return out, nil
}
