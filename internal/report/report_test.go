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
	Text(&buf, sampleResults())
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
	Text(&buf, results)
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
	Text(&buf, results)
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
