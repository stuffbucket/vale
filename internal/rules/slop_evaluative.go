package rules

import (
	"strconv"
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// evaluativeThreshold is the count of praise adjectives in one sentence at or
// above which the density reads as promotional. The evidence is explicit that
// co-occurrence (not a single hit) is the signal, so the floor is two.
const evaluativeThreshold = 2

var evaluativeSet = toLowerSet(slop.EvaluativeAdjectives)

// toLowerSet builds a lower-cased lookup set.
func toLowerSet(words []string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[strings.ToLower(w)] = true
	}
	return m
}

// setMatches returns the sentence's word tokens whose lower form is in set.
func setMatches(toks []lint.Token, set map[string]bool) []lint.Token {
	var out []lint.Token
	for _, t := range toks {
		if set[t.Lower] {
			out = append(out, t)
		}
	}
	return out
}

// SlopEvaluativeRule flags an elevated density of positive-evaluative adjectives
// within one sentence. A density metric, never a per-hit flag. Opt-in family.
type SlopEvaluativeRule struct{}

// NewSlopEvaluativeRule builds the rule.
func NewSlopEvaluativeRule() *SlopEvaluativeRule { return &SlopEvaluativeRule{} }

// ID gives the stable identifier of the rule.
func (r *SlopEvaluativeRule) ID() string { return "STE.SlopEvaluative" }

// Description gives a short summary of the rule.
func (r *SlopEvaluativeRule) Description() string {
	return "Avoid clustering positive-evaluative adjectives; the density reads as promotional."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopEvaluativeRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// Check reports each sentence whose praise-adjective count reaches the threshold.
func (r *SlopEvaluativeRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		m := setMatches(wordTokens(s), evaluativeSet)
		if len(m) < evaluativeThreshold {
			continue
		}
		findings = append(findings, lint.Finding{
			Message: "This sentence stacks " + strconv.Itoa(len(m)) + " evaluative adjectives.",
			Hint:    "Replace praise words with concrete, factual detail.",
			Line:    m[0].Line, Col: m[0].Col, EndLine: m[0].EndLine, EndCol: m[0].EndCol,
			Match: m[0].Text,
		})
	}
	return findings
}
