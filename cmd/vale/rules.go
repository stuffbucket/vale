package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/linter"
)

// cmdRules runs the rules subcommand. It lists the rules with their identifiers
// and default severities.
func cmdRules(args []string) int {
	fs := flag.NewFlagSet("rules", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	lnt := linter.New(config.Default())
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSEVERITY\tDESCRIPTION")
	for _, r := range lnt.Rules() {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.ID(), r.DefaultSeverity(), r.Description())
	}
	tw.Flush()
	return 0
}
