package rules

import (
	"strconv"
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
)

// nominalMinLen and nominalThreshold keep the rule conservative: only longer
// words count as nominalizations, and a sentence must stack several before it is
// flagged. LLM text runs high on nominalizations and STE penalizes them, but
// many legitimate technical nouns share these suffixes, so the floor is three.
const (
	nominalMinLen    = 6
	nominalThreshold = 3
)

// nominalSuffixes are the derived-noun endings the rule counts.
var nominalSuffixes = []string{"tion", "sion", "ment", "ance", "ence", "ity", "ness"}

// SlopNominalizationRule flags a high density of nominalizations (verbs frozen
// into nouns) in one sentence. A density metric, never a per-hit flag; it nudges
// toward the direct verb. Opt-in STE.Slop* family.
type SlopNominalizationRule struct{}

// NewSlopNominalizationRule builds the rule.
func NewSlopNominalizationRule() *SlopNominalizationRule { return &SlopNominalizationRule{} }

// ID gives the stable identifier of the rule.
func (r *SlopNominalizationRule) ID() string { return "STE.SlopNominalization" }

// Description gives a short summary of the rule.
func (r *SlopNominalizationRule) Description() string {
	return "Avoid stacking nominalizations; use the direct verb (configure, not perform the configuration)."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopNominalizationRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// isNominalization reports whether a lower-cased word looks like a nominalization.
func isNominalization(lower string) bool {
	if len(lower) < nominalMinLen {
		return false
	}
	for _, suf := range nominalSuffixes {
		if strings.HasSuffix(lower, suf) {
			return true
		}
	}
	return false
}

// Check reports each sentence whose nominalization count reaches the threshold.
func (r *SlopNominalizationRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		var hits []lint.Token
		for _, t := range wordTokens(s) {
			if isNominalization(t.Lower) {
				hits = append(hits, t)
			}
		}
		if len(hits) < nominalThreshold {
			continue
		}
		findings = append(findings, lint.Finding{
			Message: "This sentence stacks " + strconv.Itoa(len(hits)) + " nominalizations.",
			Hint:    "Rewrite with direct verbs.",
			Line:    hits[0].Line, Col: hits[0].Col, EndLine: hits[0].EndLine, EndCol: hits[0].EndCol,
			Match: hits[0].Text,
		})
	}
	return findings
}
