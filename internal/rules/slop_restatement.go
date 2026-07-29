package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// SlopRestatementRule flags phrases that announce a restatement — a proxy for a
// model reinventing a point it already made. Opt-in STE.Slop* family.
type SlopRestatementRule struct{}

// NewSlopRestatementRule builds the rule.
func NewSlopRestatementRule() *SlopRestatementRule { return &SlopRestatementRule{} }

// ID gives the stable identifier of the rule.
func (r *SlopRestatementRule) ID() string { return "STE.SlopRestatement" }

// Description gives a short summary of the rule.
func (r *SlopRestatementRule) Description() string {
	return "Avoid restatement markers that reintroduce a point already made."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopRestatementRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// Check reports each restatement marker.
func (r *SlopRestatementRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		for i := range toks {
			for _, phrase := range slop.RestatementMarkers {
				if matchPhrase(toks, i, phrase) {
					first, last := toks[i], toks[i+len(phrase)-1]
					text := strings.Join(phrase, " ")
					findings = append(findings, lint.Finding{
						Message: "Restatement marker \"" + text + "\" repeats an existing point.",
						Hint:    "Delete the restatement or merge it with the original.",
						Line:    first.Line, Col: first.Col, EndLine: last.EndLine, EndCol: last.EndCol,
						Match: text,
					})
				}
			}
		}
	}
	return findings
}
