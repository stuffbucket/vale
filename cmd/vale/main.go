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
  vale lint [flags] <path>...   Check files or directories.
  vale mcp                      Start the stdio MCP server.
  vale gen [flags]              Build the vocabulary rules from the wordset.
  vale rules                    List the rules.
  vale version                  Print the version.

Run "vale <command> -h" for the flags of a command.
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
	case "gen":
		return cmdGen(rest)
	case "rules":
		return cmdRules(rest)
	case "version", "--version", "-v":
		fmt.Println(buildinfo.Get())
		return 0
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		return 2
	}
}
