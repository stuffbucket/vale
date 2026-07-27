// Package linter joins the configuration, the rules, and the engine into one
// simple object. The command-line tool and the MCP server both use it, so they
// give the same findings for the same text.
package linter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/rules"
)

// MarkdownMode tells the linter how to decide if text is Markdown.
type MarkdownMode int

const (
	// MarkdownAuto decides from the file name.
	MarkdownAuto MarkdownMode = iota
	// MarkdownOn always treats the text as Markdown.
	MarkdownOn
	// MarkdownOff never treats the text as Markdown.
	MarkdownOff
)

// markdownExts are the file name endings that mean Markdown.
var markdownExts = map[string]bool{
	".md": true, ".markdown": true, ".mdown": true, ".mkd": true,
}

// Linter holds the engine and the configuration.
type Linter struct {
	engine *lint.Engine
	cfg    *config.Config
}

// New builds a linter from a configuration. A nil configuration uses the
// defaults.
func New(cfg *config.Config) *Linter {
	if cfg == nil {
		cfg = config.Default()
	}
	engine := lint.New(rules.Default(cfg), cfg.EngineConfig())
	return &Linter{engine: engine, cfg: cfg}
}

// Config gives the configuration that the linter uses.
func (l *Linter) Config() *config.Config { return l.cfg }

// Rules gives the rules that the linter holds.
func (l *Linter) Rules() []lint.Rule { return l.engine.Rules() }

// LintText checks a block of text and returns the findings. The path is only
// for reports and for the Markdown decision.
func (l *Linter) LintText(path, text string, mode MarkdownMode) []lint.Finding {
	doc := lint.Parse(path, text, lint.ParseOptions{Markdown: markdownFor(path, mode)})
	return l.engine.Run(doc)
}

// LintFile reads a file and checks it.
func (l *Linter) LintFile(path string, mode MarkdownMode) ([]lint.Finding, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return l.LintText(path, string(raw), mode), nil
}

// Gate gives the minimum severity for the exit code.
func (l *Linter) Gate() lint.Severity { return l.cfg.Gate() }

// markdownFor decides if the linter treats a text as Markdown.
func markdownFor(path string, mode MarkdownMode) bool {
	switch mode {
	case MarkdownOn:
		return true
	case MarkdownOff:
		return false
	default:
		return markdownExts[strings.ToLower(filepath.Ext(path))]
	}
}
