// Package report formats findings for people and for machines. The text format
// is for the terminal. The JSON format is for other programs.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

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
// when the finding has one. When color is true, the severity word is wrapped in
// an ANSI color; the rest of the line stays plain so it survives grep and
// click-through. Text writes findings only; call SummaryLine for the count so
// the caller can keep it off stdout.
func Text(w io.Writer, results []FileResult, color bool) {
	for _, r := range results {
		for _, f := range r.Findings {
			sev := string(f.Severity)
			if color {
				sev = colorize(f.Severity)
			}
			line := fmt.Sprintf("%s:%d:%d: %s: %s [%s]", r.Path, f.Line, f.Col, sev, f.Message, f.RuleID)
			if f.Hint != "" {
				line += " hint: " + f.Hint
			}
			fmt.Fprintln(w, line)
		}
	}
}

// colorize wraps a severity word in an ANSI color: error red, warning yellow,
// suggestion cyan.
func colorize(s lint.Severity) string {
	code := "0"
	switch s {
	case lint.SeverityError:
		code = "31"
	case lint.SeverityWarning:
		code = "33"
	case lint.SeveritySuggestion:
		code = "36"
	}
	return "\x1b[" + code + "m" + string(s) + "\x1b[0m"
}

// SummaryLine returns the one-line count of findings by severity.
func SummaryLine(results []FileResult) string {
	s := Summarize(results)
	return fmt.Sprintf("%d problems (%d errors, %d warnings, %d suggestions)",
		s.Total(), s.Errors, s.Warnings, s.Suggestions)
}

// conciseItem is one finding flattened for the concise report.
type conciseItem struct {
	path, match, hint, message string
	line, col                  int
}

// conciseGroup collects the findings of one rule.
type conciseGroup struct {
	ruleID     string
	severity   lint.Severity
	items      []conciseItem
	sharedHint string // set when every item has the same hint
}

// severityRank orders severities most-severe first.
func severityRank(s lint.Severity) int {
	switch s {
	case lint.SeverityError:
		return 0
	case lint.SeverityWarning:
		return 1
	default:
		return 2
	}
}

// groupByRule buckets findings by rule, most-severe rule first, keeping each
// rule's findings in their original (path, line) order.
func groupByRule(results []FileResult) []conciseGroup {
	order := []string{}
	byRule := map[string]*conciseGroup{}
	for _, r := range results {
		for _, f := range r.Findings {
			g, ok := byRule[f.RuleID]
			if !ok {
				g = &conciseGroup{ruleID: f.RuleID, severity: f.Severity}
				byRule[f.RuleID] = g
				order = append(order, f.RuleID)
			}
			g.items = append(g.items, conciseItem{
				path: r.Path, match: f.Match, hint: f.Hint, message: f.Message,
				line: f.Line, col: f.Col,
			})
		}
	}
	groups := make([]conciseGroup, 0, len(order))
	for _, id := range order {
		g := byRule[id]
		g.sharedHint = sharedHint(g.items)
		groups = append(groups, *g)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if ri, rj := severityRank(groups[i].severity), severityRank(groups[j].severity); ri != rj {
			return ri < rj
		}
		return groups[i].ruleID < groups[j].ruleID
	})
	return groups
}

// sharedHint returns the common hint when every item shares one, else "".
func sharedHint(items []conciseItem) string {
	if len(items) == 0 {
		return ""
	}
	h := items[0].hint
	for _, it := range items[1:] {
		if it.hint != h {
			return ""
		}
	}
	return h
}

// Concise writes a compact, LLM-friendly report: findings grouped by rule, the
// rule ID and a shared hint printed once, then one short "path:line:col match"
// line per finding. It drops the repeated per-line rule ID, severity word, and
// message that the verbose Text format emits, so the output costs far fewer
// tokens while keeping every location actionable. A per-finding hint is shown
// only when it differs within the rule.
func Concise(w io.Writer, results []FileResult, color bool) {
	for i, g := range groupByRule(results) {
		if i > 0 {
			fmt.Fprintln(w)
		}
		sev := string(g.severity)
		if color {
			sev = colorize(g.severity)
		}
		fmt.Fprintf(w, "%s · %s · %d\n", g.ruleID, sev, len(g.items))
		if g.sharedHint != "" {
			fmt.Fprintf(w, "  ↳ %s\n", g.sharedHint)
		}
		for _, it := range g.items {
			disp := it.match
			if disp == "" {
				disp = it.message
			}
			line := fmt.Sprintf("  %s:%d:%d  %s", it.path, it.line, it.col, disp)
			if g.sharedHint == "" && it.hint != "" {
				line += "  — " + it.hint
			}
			fmt.Fprintln(w, line)
		}
	}
}

// ConciseString returns the concise report as a string (no color), for callers
// like the MCP server that need it in a buffer.
func ConciseString(results []FileResult) string {
	var b strings.Builder
	Concise(&b, results, false)
	return b.String()
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
