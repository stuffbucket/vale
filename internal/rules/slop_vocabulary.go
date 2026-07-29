package rules

import (
	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// SlopVocabularyRule reports low-baseline "spike" words that surged in
// LLM-generated text and are also non-plain diction STE discourages. It is the
// one slop lexical category where per-occurrence flagging is defensible, because
// the human baseline is low. It is a plain-language nudge, never a claim that
// the text was written by a model. Part of the opt-in STE.Slop* family.
type SlopVocabularyRule struct {
	allowed map[string]bool
}

// NewSlopVocabularyRule builds the rule. The allowed set (built-in technical
// terms plus vocabulary.allow, including terms learned in a session) is never
// flagged, so a user can approve a watchlist word and stop the warning.
func NewSlopVocabularyRule(allowed map[string]bool) *SlopVocabularyRule {
	return &SlopVocabularyRule{allowed: allowed}
}

// ID gives the stable identifier of the rule.
func (r *SlopVocabularyRule) ID() string { return "STE.SlopVocabulary" }

// Description gives a short summary of the rule.
func (r *SlopVocabularyRule) Description() string {
	return "Avoid low-baseline words common to LLM output; prefer a plain, concrete term."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopVocabularyRule) DefaultSeverity() lint.Severity { return lint.SeverityWarning }

// Check reports each watchlist word.
func (r *SlopVocabularyRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		for _, t := range s.Tokens {
			if t.Kind != lint.KindWord {
				continue
			}
			if r.allowed[t.Lower] {
				continue
			}
			family, ok := slop.SlopWords[t.Lower]
			if !ok {
				continue
			}
			findings = append(findings, lint.Finding{
				Message: "The word \"" + t.Text + "\" is common to LLM output (" + family + ").",
				Hint:    "Prefer a plain, concrete term.",
				Line:    t.Line, Col: t.Col, EndLine: t.EndLine, EndCol: t.EndCol,
				Match: t.Text,
			})
		}
	}
	return findings
}

// slopRules returns the opt-in STE.Slop* family. Empty unless slop is enabled.
// The allowed set lets the watchlist honor learned vocabulary.
func slopRules(enabled bool, allowed map[string]bool) []lint.Rule {
	if !enabled {
		return nil
	}
	return []lint.Rule{
		NewSlopVocabularyRule(allowed),
		NewSlopRestatementRule(),
		NewSlopImpersonalHedgeRule(),
		NewSlopNegativeParallelismRule(),
		NewSlopEvaluativeRule(),
		NewSlopHedgeDensityRule(),
		NewSlopNominalizationRule(),
	}
}
