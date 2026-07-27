package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
)

// PassiveVoiceRule reports the passive voice. Simplified Technical English uses
// the active voice for instructions. The rule finds a form of "be" and then a
// past participle, with an optional adverb between them.
type PassiveVoiceRule struct{}

// NewPassiveVoiceRule builds the rule.
func NewPassiveVoiceRule() *PassiveVoiceRule { return &PassiveVoiceRule{} }

// ID gives the stable identifier of the rule.
func (r *PassiveVoiceRule) ID() string { return "STE.PassiveVoice" }

// Description gives a short summary of the rule.
func (r *PassiveVoiceRule) Description() string {
	return "Use the active voice, not the passive voice."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *PassiveVoiceRule) DefaultSeverity() lint.Severity { return lint.SeverityWarning }

// words returns only the word tokens of a sentence.
func wordTokens(s lint.Sentence) []lint.Token {
	var out []lint.Token
	for _, t := range s.Tokens {
		if t.Kind == lint.KindWord {
			out = append(out, t)
		}
	}
	return out
}

// Check reports each passive-voice construction.
func (r *PassiveVoiceRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		for i := 0; i < len(toks); i++ {
			if !beVerbs[toks[i].Lower] {
				continue
			}
			j := i + 1
			// Step over one optional adverb that ends in "ly".
			if j < len(toks) && strings.HasSuffix(toks[j].Lower, "ly") {
				j++
			}
			if j >= len(toks) || !looksLikeParticiple(toks[j].Lower) {
				continue
			}
			be := toks[i]
			part := toks[j]
			findings = append(findings, lint.Finding{
				Message: "Passive voice: \"" + be.Text + " " + part.Text + "\".",
				Hint:    "Rewrite the sentence in the active voice. State who does the action.",
				Line:    be.Line, Col: be.Col, EndLine: part.EndLine, EndCol: part.EndCol,
				Match: be.Text + " " + part.Text,
			})
			i = j
		}
	}
	return findings
}
