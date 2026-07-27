package lint

import "sort"

// Finding is one problem that a rule reports at one place in a document. The
// line and column numbers start at 1. The column is a count of runes, not
// bytes, so that multibyte text keeps correct positions.
type Finding struct {
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Hint     string   `json:"hint,omitempty"`
	Path     string   `json:"path,omitempty"`
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	EndLine  int      `json:"endLine"`
	EndCol   int      `json:"endCol"`
	Match    string   `json:"match,omitempty"`
}

// SortFindings puts findings in reading order: first by line, then by column,
// then by rule identifier. The sort is stable and changes the slice in place.
func SortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Col != b.Col {
			return a.Col < b.Col
		}
		return a.RuleID < b.RuleID
	})
}
