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
	format := fs.String("format", "concise", "output format: concise, text, or json")
	markdownFlag := fs.String("markdown", "auto", "markdown mode: auto, on, or off")
	strict := fs.Bool("strict-vocabulary", false, "also report unapproved words with no replacement")
	slopFlag := fs.Bool("slop", false, "enable the opt-in STE.Slop* rule family")
	colorFlag := fs.String("color", "auto", "color: auto, always, or never")
	audit := fs.Bool("audit", false, "audit only: print findings but always exit 0")
	fix := fs.Bool("fix", false, "rewrite the file with a model to resolve findings; prints to stdout")
	model := fs.String("model", "", "model for --fix (default: the model.name config)")
	endpoint := fs.String("endpoint", "", "endpoint for --fix (default: the model.endpoint config)")
	temperature := fs.Float64("temperature", -1, "temperature for --fix; -1 uses the config or server default")
	output := fs.String("output", "", "write --fix output to this file instead of stdout")
	fixMaxTokens := fs.Int("max-tokens", 2048, "max tokens for the --fix rewrite")
	paths, err := parsePositional(fs, args)
	if err != nil {
		return 2
	}
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
	if *slopFlag {
		cfg.Slop.Enabled = true
	}
	if *minSeverity != "" {
		cfg.MinSeverity = *minSeverity
	}

	if *fix {
		return runFix(fixSettings{
			cfg:         cfg,
			paths:       paths,
			model:       *model,
			endpoint:    *endpoint,
			temperature: temperatureFlag(*temperature),
			output:      *output,
			maxTokens:   *fixMaxTokens,
		})
	}

	mode, err := parseMarkdownMode(*markdownFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lint: %v\n", err)
		return 2
	}
	if *format != "text" && *format != "json" && *format != "concise" {
		fmt.Fprintf(os.Stderr, "lint: unknown format %q\n", *format)
		return 2
	}
	if *colorFlag != "auto" && *colorFlag != "always" && *colorFlag != "never" {
		fmt.Fprintf(os.Stderr, "lint: unknown color mode %q\n", *colorFlag)
		return 2
	}

	lnt := linter.New(cfg)
	filter := newPathFilter(cfg.Files.Include, cfg.Files.Exclude, ".")
	files, err := collectFiles(paths, filter)
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

	color := shouldColor(*colorFlag, os.Stdout)
	switch *format {
	case "json":
		if err := report.JSON(os.Stdout, results); err != nil {
			fmt.Fprintf(os.Stderr, "lint: %v\n", err)
			return 2
		}
	case "text":
		// Verbose: one self-contained clickable line per finding.
		report.Text(os.Stdout, results, color)
		fmt.Fprintln(os.Stderr, report.SummaryLine(results))
	default: // concise
		// Compact, grouped by rule — the token-efficient default for tools/LLMs.
		report.Concise(os.Stdout, results, color)
		fmt.Fprintln(os.Stderr, report.SummaryLine(results))
	}

	// Exit-code contract for agentic harnesses:
	//   0  no findings at or above the gate (or --audit)
	//   1  findings at or above the gate (lint failures found)
	//   2  usage or runtime error (returned earlier)
	if !*audit && gateFailed(results, cfg.Gate()) {
		return 1
	}
	return 0
}

// shouldColor decides whether to color the output. "always" and "never" force
// it; "auto" (the default) turns color on only when the writer is a terminal
// and NO_COLOR is unset, so a pipe, a redirect, or CI gets plain text.
func shouldColor(mode string, f *os.File) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	default:
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		return isTerminal(f)
	}
}

// isTerminal reports whether f is a character device (a terminal).
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// parsePositional parses flags that may appear before, after, or between
// positional arguments. Go's flag package stops at the first non-flag, so this
// re-parses around each positional to allow "vale file.md --format json".
func parsePositional(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
	return positional, nil
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
// directories and keeps only files with a known ending that the filter allows.
// An explicitly named file argument is always kept; the filter applies only to
// files discovered by walking a directory.
func collectFiles(paths []string, filter pathFilter) ([]string, error) {
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
			add(p) // explicit file: always linted
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == ".git" || d.Name() == "node_modules" || filter.excludes(path) {
					return filepath.SkipDir
				}
				return nil
			}
			if lintExtensions[strings.ToLower(filepath.Ext(path))] && filter.allows(path) {
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

// pathFilter decides which walked files to lint, from the config's
// files.include / files.exclude globs plus a .valeignore file.
type pathFilter struct {
	include []string
	exclude []string
}

// newPathFilter builds a filter and folds in patterns from the nearest
// .valeignore, searching up from dir.
func newPathFilter(include, exclude []string, dir string) pathFilter {
	return pathFilter{
		include: include,
		exclude: append(append([]string{}, exclude...), readValeIgnore(dir)...),
	}
}

// allows reports whether a file passes the filter: not excluded, and (when an
// include list is set) matching it.
func (f pathFilter) allows(path string) bool {
	if f.excludes(path) {
		return false
	}
	return len(f.include) == 0 || matchAny(f.include, path)
}

// excludes reports whether a path matches any exclude pattern.
func (f pathFilter) excludes(path string) bool { return matchAny(f.exclude, path) }

// matchAny matches a path against glob patterns, testing the full slash path,
// the base name, and (for a "**/" prefix) the base name at any depth.
func matchAny(patterns []string, path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	base := filepath.Base(clean)
	for _, p := range patterns {
		if ok, _ := filepath.Match(p, clean); ok {
			return true
		}
		if ok, _ := filepath.Match(p, base); ok {
			return true
		}
		if rest, found := strings.CutPrefix(p, "**/"); found {
			if ok, _ := filepath.Match(rest, base); ok {
				return true
			}
		}
	}
	return false
}

// readValeIgnore returns the glob patterns from the nearest .valeignore, walking
// up from dir. Blank lines and lines starting with "#" are ignored.
func readValeIgnore(dir string) []string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}
	for {
		data, err := os.ReadFile(filepath.Join(abs, ".valeignore"))
		if err == nil {
			var pats []string
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					pats = append(pats, line)
				}
			}
			return pats
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return nil
		}
		abs = parent
	}
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
