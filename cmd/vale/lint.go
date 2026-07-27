package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/linter"
	"github.com/stuffbucket/vale/internal/report"
)

// lintExtensions are the file endings that lint reads when it walks a
// directory.
var lintExtensions = map[string]bool{
	".md": true, ".markdown": true, ".mdown": true, ".mkd": true, ".txt": true,
}

// cmdLint runs the lint subcommand and returns the exit code.
func cmdLint(args []string) int {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to a config file")
	minSeverity := fs.String("min-severity", "", "gate: fail on this severity or higher")
	format := fs.String("format", "text", "output format: text or json")
	markdownFlag := fs.String("markdown", "auto", "markdown mode: auto, on, or off")
	strict := fs.Bool("strict-vocabulary", false, "also report unapproved words with no replacement")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "lint: give at least one path")
		return 2
	}

	cfg, err := config.Load(*configPath, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "lint: %v\n", err)
		return 2
	}
	if *strict {
		cfg.StrictVocabulary = true
	}
	if *minSeverity != "" {
		cfg.MinSeverity = *minSeverity
	}

	mode, err := parseMarkdownMode(*markdownFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lint: %v\n", err)
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "lint: unknown format %q\n", *format)
		return 2
	}

	lnt := linter.New(cfg)
	files, err := collectFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lint: %v\n", err)
		return 2
	}

	var results []report.FileResult
	for _, path := range files {
		findings, err := lnt.LintFile(path, mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lint: %v\n", err)
			return 2
		}
		results = append(results, report.FileResult{Path: path, Findings: findings})
	}

	if *format == "json" {
		if err := report.JSON(os.Stdout, results); err != nil {
			fmt.Fprintf(os.Stderr, "lint: %v\n", err)
			return 2
		}
	} else {
		report.Text(os.Stdout, results)
	}

	if gateFailed(results, cfg.Gate()) {
		return 1
	}
	return 0
}

// parseMarkdownMode reads the markdown flag value.
func parseMarkdownMode(value string) (linter.MarkdownMode, error) {
	switch strings.ToLower(value) {
	case "auto", "":
		return linter.MarkdownAuto, nil
	case "on", "true":
		return linter.MarkdownOn, nil
	case "off", "false":
		return linter.MarkdownOff, nil
	default:
		return linter.MarkdownAuto, fmt.Errorf("unknown markdown mode %q", value)
	}
}

// collectFiles turns the paths into a sorted list of files. It walks
// directories and keeps only files with a known ending.
func collectFiles(paths []string) ([]string, error) {
	var files []string
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			files = append(files, p)
		}
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			add(p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == ".git" || d.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if lintExtensions[strings.ToLower(filepath.Ext(path))] {
				add(path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

// gateFailed tells if any finding reaches the gate severity.
func gateFailed(results []report.FileResult, gate lint.Severity) bool {
	for _, r := range results {
		for _, f := range r.Findings {
			if f.Severity.AtLeast(gate) {
				return true
			}
		}
	}
	return false
}
