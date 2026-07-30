package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/stuffbucket/vale/internal/buildinfo"
	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
	"github.com/stuffbucket/vale/internal/lsp"
)

// cmdLSP runs the lsp subcommand. It starts a stdio Language Server that
// publishes Simplified Technical English diagnostics for open documents.
func cmdLSP(args []string) int {
	fs := flag.NewFlagSet("lsp", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to a config file")
	slopFlag := fs.Bool("slop", false, "enable the opt-in STE.Slop* rule family")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load(*configPath, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "lsp: %v\n", err)
		return 2
	}
	if *slopFlag {
		cfg.Slop.Enabled = true
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := lsp.Serve(ctx, os.Stdin, os.Stdout, linter.New(cfg), buildinfo.Get().Version); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "lsp: %v\n", err)
		return 1
	}
	return 0
}
