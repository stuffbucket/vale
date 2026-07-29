package rules

import (
	"strconv"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// hedgeThreshold is the count of hedge words in one sentence at or above which
// the clustering is a signal. Individual hedges are legitimate; only abnormal
// packing counts, so the floor is three.
const hedgeThreshold = 3

var hedgeSet = toLowerSet(slop.Hedges)

// SlopHedgeDensityRule flags a cluster of hedge words in one sentence. A density
// metric, never a per-hit flag. Opt-in STE.Slop* family.
type SlopHedgeDensityRule struct{}

// NewSlopHedgeDensityRule builds the rule.
func NewSlopHedgeDensityRule() *SlopHedgeDensityRule { return &SlopHedgeDensityRule{} }

// ID gives the stable identifier of the rule.
func (r *SlopHedgeDensityRule) ID() string { return "STE.SlopHedgeDensity" }

// Description gives a short summary of the rule.
func (r *SlopHedgeDensityRule) Description() string {
	return "Avoid stacking hedges in one sentence; commit to a clear statement."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopHedgeDensityRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// Check reports each sentence whose hedge count reaches the threshold.
func (r *SlopHedgeDensityRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		m := setMatches(wordTokens(s), hedgeSet)
		if len(m) < hedgeThreshold {
			continue
		}
		findings = append(findings, lint.Finding{
			Message: "This sentence stacks " + strconv.Itoa(len(m)) + " hedges.",
			Hint:    "State the condition once, or commit to a direct statement.",
			Line:    m[0].Line, Col: m[0].Col, EndLine: m[0].EndLine, EndCol: m[0].EndCol,
			Match: m[0].Text,
		})
	}
	return findings
}
