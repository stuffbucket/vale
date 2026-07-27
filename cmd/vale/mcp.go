package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
	"github.com/stuffbucket/vale/internal/mcp"
)

// cmdMCP runs the mcp subcommand. It starts the stdio MCP server.
func cmdMCP(args []string) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to a config file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load(*configPath, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	lnt := linter.New(cfg)
	if err := mcp.Serve(ctx, os.Stdin, os.Stdout, lnt, version); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		return 1
	}
	return 0
}
