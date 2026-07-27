package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
)

// ingAllowed holds words that end in "ing" but are acceptable. They are common
// nouns or technical terms, not verb forms in an instruction.
var ingAllowed = map[string]bool{
	"bearing": true, "being": true, "bring": true, "building": true, "casing": true,
	"ceiling": true, "coating": true, "coding": true, "during": true, "engineering": true,
	"everything": true, "fitting": true, "following": true, "grounding": true, "heading": true,
	"housing": true, "king": true, "lightning": true, "meaning": true, "morning": true,
	"nothing": true, "opening": true, "packaging": true, "ring": true, "setting": true,
	"something": true, "spring": true, "string": true, "swing": true, "thing": true,
	"timing": true, "training": true, "warning": true, "wing": true, "wiring": true,
	"according": true, "regarding": true, "sing": true,
}

// IngFormRule reports words that end in "ing" in instructions. Simplified
// Technical English prefers simple verb forms. The rule allows headings and a
// set of common nouns and technical terms.
type IngFormRule struct{}

// NewIngFormRule builds the rule.
func NewIngFormRule() *IngFormRule { return &IngFormRule{} }

// ID gives the stable identifier of the rule.
func (r *IngFormRule) ID() string { return "STE.IngForms" }

// Description gives a short summary of the rule.
func (r *IngFormRule) Description() string {
	return "Do not use the -ing verb form in instructions."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *IngFormRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// Check reports each -ing word that is not in the allowed set.
func (r *IngFormRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		if s.Block == lint.BlockHeading {
			continue
		}
		for _, t := range s.Tokens {
			if t.Kind != lint.KindWord {
				continue
			}
			word := t.Lower
			if !strings.HasSuffix(word, "ing") || len([]rune(word)) < 5 {
				continue
			}
			if ingAllowed[word] {
				continue
			}
			findings = append(findings, lint.Finding{
				Message: "The -ing form \"" + t.Text + "\" is hard to read in an instruction.",
				Hint:    "Use a simple verb form, such as the imperative.",
				Line:    t.Line, Col: t.Col, EndLine: t.EndLine, EndCol: t.EndCol,
				Match: t.Text,
			})
		}
	}
	return findings
}
