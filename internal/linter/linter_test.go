package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
)

func TestLintTextFindsContraction(t *testing.T) {
	l := New(config.Default())
	got := l.LintText("doc.txt", "Don't stop the fan.", MarkdownOff)
	var found *lint.Finding
	for i := range got {
		if got[i].RuleID == "STE.Contractions" {
			found = &got[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("no contraction finding: %+v", got)
	}
	if found.Match != "Don't" {
		t.Errorf("match = %q, want Don't", found.Match)
	}
	if found.Path != "doc.txt" {
		t.Errorf("path = %q", found.Path)
	}
	if found.Severity != lint.SeverityError {
		t.Errorf("severity = %q, want error", found.Severity)
	}
}

func TestLintTextNilConfigUsesDefaults(t *testing.T) {
	l := New(nil)
	if l.Config() == nil {
		t.Fatal("Config() nil")
	}
	if l.Config().Sentence.DescriptionMax != 25 {
		t.Errorf("default description max = %d", l.Config().Sentence.DescriptionMax)
	}
}

func TestMarkdownModeAuto(t *testing.T) {
	l := New(config.Default())
	// A heading in a .md file is exempt from the ing-form rule, so "Installing"
	// in a heading is not flagged when markdown auto-detects the .md extension.
	mdFindings := l.LintText("doc.md", "# Installing the system now", MarkdownAuto)
	for _, f := range mdFindings {
		if f.RuleID == "STE.IngForms" {
			t.Errorf("heading should not flag ing form in markdown mode: %+v", f)
		}
	}
	// The same text in a .txt file is treated as prose, so the heading marker is
	// just punctuation and "Installing" is flagged.
	txtFindings := l.LintText("doc.txt", "# Installing the system now", MarkdownAuto)
	var ing bool
	for _, f := range txtFindings {
		if f.RuleID == "STE.IngForms" {
			ing = true
		}
	}
	if !ing {
		t.Errorf("expected ing-form finding in non-markdown text: %+v", txtFindings)
	}
}

func TestMarkdownModeForced(t *testing.T) {
	l := New(config.Default())
	// MarkdownOn on a .txt path treats it as markdown (heading exempt).
	on := l.LintText("doc.txt", "# Installing the system now", MarkdownOn)
	for _, f := range on {
		if f.RuleID == "STE.IngForms" {
			t.Errorf("MarkdownOn should treat as heading: %+v", f)
		}
	}
	// MarkdownOff on a .md path treats it as prose (ing flagged).
	off := l.LintText("doc.md", "# Installing the system now", MarkdownOff)
	var ing bool
	for _, f := range off {
		if f.RuleID == "STE.IngForms" {
			ing = true
		}
	}
	if !ing {
		t.Errorf("MarkdownOff should treat .md as prose: %+v", off)
	}
}

func TestLintFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(path, []byte("Don't stop."), 0o644); err != nil {
		t.Fatal(err)
	}
	l := New(config.Default())
	got, err := l.LintFile(path, MarkdownAuto)
	if err != nil {
		t.Fatalf("LintFile error: %v", err)
	}
	var ok bool
	for _, f := range got {
		if f.RuleID == "STE.Contractions" {
			ok = true
		}
	}
	if !ok {
		t.Errorf("expected contraction finding from file: %+v", got)
	}
}

func TestLintFileMissing(t *testing.T) {
	l := New(config.Default())
	if _, err := l.LintFile(filepath.Join(t.TempDir(), "nope.txt"), MarkdownAuto); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRulesAndGate(t *testing.T) {
	l := New(config.Default())
	if len(l.Rules()) != 7 {
		t.Errorf("rules = %d, want 7", len(l.Rules()))
	}
	if l.Gate() != lint.SeverityError {
		t.Errorf("gate = %q, want error", l.Gate())
	}
}

func TestDisabledRuleThroughConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Rules = map[string]config.RuleSetting{
		"STE.Contractions": {Disabled: true},
	}
	l := New(cfg)
	got := l.LintText("doc.txt", "Don't stop.", MarkdownOff)
	for _, f := range got {
		if f.RuleID == "STE.Contractions" {
			t.Errorf("contraction rule should be disabled: %+v", f)
		}
	}
}

func TestSeverityOverrideThroughConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Rules = map[string]config.RuleSetting{
		"STE.Contractions": {Severity: "suggestion"},
	}
	l := New(cfg)
	got := l.LintText("doc.txt", "Don't stop.", MarkdownOff)
	var found bool
	for _, f := range got {
		if f.RuleID == "STE.Contractions" {
			found = true
			if f.Severity != lint.SeveritySuggestion {
				t.Errorf("severity = %q, want suggestion", f.Severity)
			}
		}
	}
	if !found {
		t.Fatal("no contraction finding")
	}
}
