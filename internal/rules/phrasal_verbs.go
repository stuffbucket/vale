package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
)

// phrasalVerb is one phrasal verb and its approved replacement.
type phrasalVerb struct {
	words   []string // the lowercase words of the phrase, in order
	replace string   // the approved single verb or short phrase
}

// phrasalVerbs is a small curated set of phrasal verbs. Simplified Technical
// English prefers a single, clear verb.
var phrasalVerbs = []phrasalVerb{
	{[]string{"carry", "out"}, "do or complete"},
	{[]string{"check", "out"}, "examine"},
	{[]string{"fill", "in"}, "complete"},
	{[]string{"find", "out"}, "determine"},
	{[]string{"hook", "up"}, "connect"},
	{[]string{"look", "up"}, "find"},
	{[]string{"pick", "up"}, "lift or collect"},
	{[]string{"put", "in"}, "install"},
	{[]string{"set", "up"}, "install or prepare"},
	{[]string{"shut", "down"}, "stop"},
	{[]string{"take", "out"}, "remove"},
	{[]string{"turn", "off"}, "stop"},
	{[]string{"turn", "on"}, "start"},
}

// PhrasalVerbRule reports phrasal verbs and gives a single-verb replacement.
type PhrasalVerbRule struct{}

// NewPhrasalVerbRule builds the rule.
func NewPhrasalVerbRule() *PhrasalVerbRule { return &PhrasalVerbRule{} }

// ID gives the stable identifier of the rule.
func (r *PhrasalVerbRule) ID() string { return "STE.PhrasalVerbs" }

// Description gives a short summary of the rule.
func (r *PhrasalVerbRule) Description() string {
	return "Do not use phrasal verbs; use one clear verb."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *PhrasalVerbRule) DefaultSeverity() lint.Severity { return lint.SeverityWarning }

// Check reports each phrasal verb in the document.
func (r *PhrasalVerbRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		for i := range toks {
			for _, pv := range phrasalVerbs {
				if match := matchPhrase(toks, i, pv.words); match {
					first := toks[i]
					last := toks[i+len(pv.words)-1]
					phrase := strings.Join(pv.words, " ")
					findings = append(findings, lint.Finding{
						Message: "Phrasal verb \"" + phrase + "\" is not clear.",
						Hint:    "Use \"" + pv.replace + "\".",
						Line:    first.Line, Col: first.Col, EndLine: last.EndLine, EndCol: last.EndCol,
						Match: phrase,
					})
					break
				}
			}
		}
	}
	return findings
}

// matchPhrase tells if the words at index i match the phrase words in order.
func matchPhrase(toks []lint.Token, i int, words []string) bool {
	if i+len(words) > len(toks) {
		return false
	}
	for k, w := range words {
		if toks[i+k].Lower != w {
			return false
		}
	}
	return true
}
