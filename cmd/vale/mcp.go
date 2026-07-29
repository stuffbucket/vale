package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/stuffbucket/vale/internal/buildinfo"
	"github.com/stuffbucket/vale/internal/mcp"
)

// cmdMCP runs the mcp subcommand. It starts the stdio MCP server in session
// mode, so the update_vocabulary tool can learn terms during the session.
func cmdMCP(args []string) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to a config file")
	vocabStore := fs.String("vocab-store", "",
		"file for terms learned via update_vocabulary (default .vale-ste.vocab.yml; use a session-keyed path to scope to the session)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	server, err := mcp.NewSessionServer(mcp.SessionOptions{
		ConfigPath: *configPath,
		Dir:        ".",
		StorePath:  *vocabStore,
		Version:    buildinfo.Get().Version,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		return 1
	}
	return 0
}
