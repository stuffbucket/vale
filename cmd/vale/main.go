// Command vale is a Simplified Technical English linter and a self-powered MCP
// server. It has three main subcommands: lint, mcp, and gen.
package main

import (
	"fmt"
	"os"

	"github.com/stuffbucket/vale/internal/buildinfo"
)

// usage is the top-level help text.
const usage = `vale - a Simplified Technical English linter and MCP server

Usage:
  vale [flags] <path>...        Check files or directories (the default action).
  vale lint [flags] <path>...   The same, stated explicitly.
  vale mcp                      Start the stdio MCP server.
  vale lsp                      Start the stdio Language Server (editor diagnostics).
  vale gen [flags]              Build the vocabulary rules from the wordset.
  vale rules                    List the rules.
  vale eval [flags]             Measure slop across an LLM endpoint's models.
  vale version                  Print the version.

Run "vale lint -h" for the lint flags.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

// run dispatches a subcommand and returns the exit code.
func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "lint":
		return cmdLint(rest)
	case "mcp":
		return cmdMCP(rest)
	case "lsp":
		return cmdLSP(rest)
	case "gen":
		return cmdGen(rest)
	case "rules":
		return cmdRules(rest)
	case "eval":
		return cmdEval(rest)
	case "version", "--version", "-v":
		fmt.Println(buildinfo.Get())
		return 0
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0
	default:
		// No known subcommand: treat every argument as lint input (paths and
		// lint flags). This makes "vale <path>..." the default action.
		return cmdLint(args)
	}
}
