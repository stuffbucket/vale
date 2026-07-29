package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"

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
	temperature := fs.Float64("temperature", -1, "sampling temperature; -1 uses the config or server default")
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
		Temperature: temperatureFlag(*temperature),
		Concurrency: *concurrency,
		OnProgress:  progressReporter(os.Stderr),
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

// temperatureFlag converts the CLI sentinel (-1) into a nil pointer (use the
// config/server default) or a real value.
func temperatureFlag(v float64) *float64 {
	if v < 0 {
		return nil
	}
	return &v
}

// progressReporter returns an OnProgress callback that draws a live bar to w
// when w is a terminal, or nil otherwise. It serializes writes across goroutines.
func progressReporter(w *os.File) func(done, total int) {
	if !isTerminal(w) {
		return nil
	}
	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(28), progress.WithoutPercentage())
	var mu sync.Mutex
	return func(done, total int) {
		mu.Lock()
		defer mu.Unlock()
		pct := 0.0
		if total > 0 {
			pct = float64(done) / float64(total)
		}
		fmt.Fprintf(w, "\r  eval %s %3d/%d", bar.ViewAs(pct), done, total)
		if done >= total {
			fmt.Fprint(w, "\r\033[K") // clear the line; the report follows on stdout
		}
	}
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
