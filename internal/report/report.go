// Package report formats findings for people and for machines. The text format
// is for the terminal. The JSON format is for other programs.
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/stuffbucket/vale/internal/lint"
)

// FileResult holds the findings for one file.
type FileResult struct {
	Path     string         `json:"path"`
	Findings []lint.Finding `json:"findings"`
}

// Summary counts findings by severity across all files.
type Summary struct {
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Suggestions int `json:"suggestions"`
}

// Total gives the sum of all counts.
func (s Summary) Total() int { return s.Errors + s.Warnings + s.Suggestions }

// Summarize counts the findings in the results.
func Summarize(results []FileResult) Summary {
	var s Summary
	for _, r := range results {
		for _, f := range r.Findings {
			switch f.Severity {
			case lint.SeverityError:
				s.Errors++
			case lint.SeverityWarning:
				s.Warnings++
			case lint.SeveritySuggestion:
				s.Suggestions++
			}
		}
	}
	return s
}

// Text writes the results in a form for the terminal.
// Text writes one actionable line per finding, in the form
//
//	path:line:col: severity: message [RuleID] hint: <fix>
//
// Every line carries the file path and position so an editor, a terminal, or a
// CI problem matcher can jump straight to the location. The hint is appended
// when the finding has one. Text writes findings only; call SummaryLine for the
// count so the caller can keep it off stdout.
func Text(w io.Writer, results []FileResult) {
	for _, r := range results {
		for _, f := range r.Findings {
			line := fmt.Sprintf("%s:%d:%d: %s: %s [%s]", r.Path, f.Line, f.Col, f.Severity, f.Message, f.RuleID)
			if f.Hint != "" {
				line += " hint: " + f.Hint
			}
			fmt.Fprintln(w, line)
		}
	}
}

// SummaryLine returns the one-line count of findings by severity.
func SummaryLine(results []FileResult) string {
	s := Summarize(results)
	return fmt.Sprintf("%d problems (%d errors, %d warnings, %d suggestions)",
		s.Total(), s.Errors, s.Warnings, s.Suggestions)
}

// JSON writes the results as JSON.
func JSON(w io.Writer, results []FileResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	out := struct {
		Results []FileResult `json:"results"`
		Summary Summary      `json:"summary"`
	}{Results: results, Summary: Summarize(results)}
	return enc.Encode(out)
}
