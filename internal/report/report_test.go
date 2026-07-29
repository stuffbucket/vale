package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stuffbucket/vale/internal/lint"
)

func sampleResults() []FileResult {
	return []FileResult{
		{
			Path: "a.md",
			Findings: []lint.Finding{
				{RuleID: "STE.Contractions", Severity: lint.SeverityError, Message: "Contraction.", Hint: "Fix it.", Line: 1, Col: 1},
				{RuleID: "STE.PassiveVoice", Severity: lint.SeverityWarning, Message: "Passive.", Line: 2, Col: 3},
			},
		},
		{
			Path: "b.md",
			Findings: []lint.Finding{
				{RuleID: "STE.IngForms", Severity: lint.SeveritySuggestion, Message: "Ing.", Line: 5, Col: 2},
			},
		},
	}
}

func TestSummarize(t *testing.T) {
	s := Summarize(sampleResults())
	if s.Errors != 1 {
		t.Errorf("errors = %d, want 1", s.Errors)
	}
	if s.Warnings != 1 {
		t.Errorf("warnings = %d, want 1", s.Warnings)
	}
	if s.Suggestions != 1 {
		t.Errorf("suggestions = %d, want 1", s.Suggestions)
	}
	if s.Total() != 3 {
		t.Errorf("total = %d, want 3", s.Total())
	}
}

func TestSummarizeEmpty(t *testing.T) {
	s := Summarize(nil)
	if s.Total() != 0 {
		t.Errorf("total = %d, want 0", s.Total())
	}
}

func TestText(t *testing.T) {
	var buf bytes.Buffer
	Text(&buf, sampleResults(), false)
	out := buf.String()

	// Each finding is one actionable "path:line:col: severity: message [rule]" line.
	for _, want := range []string{
		"a.md:1:1: error: Contraction. [STE.Contractions] hint: Fix it.",
		"a.md:2:3: warning: Passive. [STE.PassiveVoice]",
		"b.md:5:2: suggestion: Ing. [STE.IngForms]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}
	// The summary is not part of Text; the caller writes it separately.
	if strings.Contains(out, "problems (") {
		t.Errorf("Text must not print the summary: %q", out)
	}
	// Plain mode carries no ANSI escapes.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain Text must not emit ANSI: %q", out)
	}
}

func TestTextColor(t *testing.T) {
	var buf bytes.Buffer
	Text(&buf, sampleResults(), true)
	out := buf.String()
	// The severity word is wrapped in ANSI; the path stays plain and clickable.
	if !strings.Contains(out, "\x1b[31merror\x1b[0m") {
		t.Errorf("color mode should wrap the severity: %q", out)
	}
	if !strings.Contains(out, "a.md:1:1: ") {
		t.Errorf("path:line:col prefix must stay plain: %q", out)
	}
}

func TestSummaryLine(t *testing.T) {
	if got := SummaryLine(sampleResults()); got != "3 problems (1 errors, 1 warnings, 1 suggestions)" {
		t.Errorf("SummaryLine = %q", got)
	}
	if got := SummaryLine(nil); got != "0 problems (0 errors, 0 warnings, 0 suggestions)" {
		t.Errorf("empty SummaryLine = %q", got)
	}
}

func TestTextSkipsFilesWithNoFindings(t *testing.T) {
	var buf bytes.Buffer
	results := []FileResult{
		{Path: "clean.md", Findings: nil},
	}
	Text(&buf, results, false)
	if out := buf.String(); out != "" {
		t.Errorf("clean file must produce no output: %q", out)
	}
}

func TestTextNoHintOmitsHintLine(t *testing.T) {
	var buf bytes.Buffer
	results := []FileResult{
		{Path: "a.md", Findings: []lint.Finding{
			{RuleID: "R", Severity: lint.SeverityError, Message: "No hint.", Line: 1, Col: 1},
		}},
	}
	Text(&buf, results, false)
	if strings.Contains(buf.String(), "hint:") {
		t.Errorf("should not print hint line when hint is empty: %q", buf.String())
	}
}

func TestJSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sampleResults()); err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	var decoded struct {
		Results []FileResult `json:"results"`
		Summary Summary      `json:"summary"`
	}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("decode error: %v\n%s", err, buf.String())
	}
	if len(decoded.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(decoded.Results))
	}
	if decoded.Results[0].Path != "a.md" {
		t.Errorf("path = %q, want a.md", decoded.Results[0].Path)
	}
	if decoded.Results[0].Findings[0].RuleID != "STE.Contractions" {
		t.Errorf("ruleId = %q", decoded.Results[0].Findings[0].RuleID)
	}
	if decoded.Summary.Errors != 1 || decoded.Summary.Warnings != 1 || decoded.Summary.Suggestions != 1 {
		t.Errorf("summary = %+v", decoded.Summary)
	}
}

func TestJSONFieldNames(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sampleResults()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{`"results"`, `"summary"`, `"ruleId"`, `"severity"`, `"errors"`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing key %q\n%s", want, out)
		}
	}
}

func TestConciseGroupsAndDedupes(t *testing.T) {
	var buf bytes.Buffer
	Concise(&buf, sampleResults(), false)
	out := buf.String()
	// Rule id appears once as a group header, not per finding.
	if strings.Count(out, "STE.Contractions") != 1 {
		t.Errorf("rule id should appear once as a header:\n%s", out)
	}
	// Locations are present and compact (path:line:col match).
	if !strings.Contains(out, "a.md:1:1  ") || !strings.Contains(out, "b.md:5:2  ") {
		t.Errorf("concise locations missing:\n%s", out)
	}
	// Error group precedes suggestion group (severity order).
	if strings.Index(out, "STE.Contractions") > strings.Index(out, "STE.IngForms") {
		t.Errorf("groups not ordered by severity:\n%s", out)
	}
	// No ANSI in plain mode.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain concise must not emit ANSI:\n%s", out)
	}
}

func TestConciseSharedHintPrintedOnce(t *testing.T) {
	results := []FileResult{{Path: "a.md", Findings: []lint.Finding{
		{RuleID: "STE.PassiveVoice", Severity: lint.SeverityWarning, Message: "p1", Hint: "Use active voice.", Line: 1, Col: 1, Match: "was done"},
		{RuleID: "STE.PassiveVoice", Severity: lint.SeverityWarning, Message: "p2", Hint: "Use active voice.", Line: 2, Col: 1, Match: "is tested"},
	}}}
	var buf bytes.Buffer
	Concise(&buf, results, false)
	out := buf.String()
	if strings.Count(out, "Use active voice.") != 1 {
		t.Errorf("shared hint should print once:\n%s", out)
	}
	if !strings.Contains(out, "was done") || !strings.Contains(out, "is tested") {
		t.Errorf("matches missing:\n%s", out)
	}
}
