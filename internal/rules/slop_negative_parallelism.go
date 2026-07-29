package rules

import (
	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// SlopNegativeParallelismRule flags the "not only X but also Y" / "not just X
// but Y" construction — a strong cross-family structural tell that is also wordy
// by STE standards. Opt-in STE.Slop* family.
type SlopNegativeParallelismRule struct{}

// NewSlopNegativeParallelismRule builds the rule.
func NewSlopNegativeParallelismRule() *SlopNegativeParallelismRule {
	return &SlopNegativeParallelismRule{}
}

// ID gives the stable identifier of the rule.
func (r *SlopNegativeParallelismRule) ID() string { return "STE.SlopNegativeParallelism" }

// Description gives a short summary of the rule.
func (r *SlopNegativeParallelismRule) Description() string {
	return "Avoid the \"not only X but also Y\" construction; state it directly."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopNegativeParallelismRule) DefaultSeverity() lint.Severity {
	return lint.SeverityWarning
}

// firstMatch returns the index of the first token where phrase matches, or -1.
func firstMatch(toks []lint.Token, from int, phrase []string) int {
	for i := from; i+len(phrase) <= len(toks); i++ {
		if matchPhrase(toks, i, phrase) {
			return i
		}
	}
	return -1
}

// Check reports one negative-parallelism construction per sentence per pattern.
func (r *SlopNegativeParallelismRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		for _, p := range slop.NegativeParallelismPatterns {
			i := firstMatch(toks, 0, p.First)
			if i < 0 {
				continue
			}
			j := firstMatch(toks, i+len(p.First), p.Second)
			if j < 0 {
				continue
			}
			first, last := toks[i], toks[j+len(p.Second)-1]
			findings = append(findings, lint.Finding{
				Message: "Negative parallelism (\"not only … but also …\") is wordy filler.",
				Hint:    "State the point directly.",
				Line:    first.Line, Col: first.Col, EndLine: last.EndLine, EndCol: last.EndCol,
				Match: first.Text,
			})
		}
	}
	return findings
}
