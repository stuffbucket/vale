package rules

import (
	"fmt"

	"github.com/stuffbucket/vale/internal/lint"
)

// SentenceLengthRule reports sentences that have too many words. Procedures use
// a smaller limit than descriptions. Simplified Technical English keeps
// sentences short so that they are easy to read.
type SentenceLengthRule struct {
	procedureMax   int
	descriptionMax int
}

// NewSentenceLengthRule builds the rule with word limits for procedures and
// descriptions. A limit of zero or less turns off that limit.
func NewSentenceLengthRule(procedureMax, descriptionMax int) *SentenceLengthRule {
	return &SentenceLengthRule{procedureMax: procedureMax, descriptionMax: descriptionMax}
}

// ID gives the stable identifier of the rule.
func (r *SentenceLengthRule) ID() string { return "STE.SentenceLength" }

// Description gives a short summary of the rule.
func (r *SentenceLengthRule) Description() string {
	return "Keep sentences short: procedures to 20 words, descriptions to 25 words."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SentenceLengthRule) DefaultSeverity() lint.Severity { return lint.SeverityError }

// Check reports each sentence that has too many words.
func (r *SentenceLengthRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		if s.Block == lint.BlockHeading {
			continue
		}
		limit := r.descriptionMax
		kind := "description"
		if s.Block == lint.BlockListItem {
			limit = r.procedureMax
			kind = "procedure"
		}
		if limit <= 0 {
			continue
		}
		count := s.WordCount()
		if count <= limit {
			continue
		}
		last := s.Tokens[len(s.Tokens)-1]
		findings = append(findings, lint.Finding{
			Message: fmt.Sprintf("Sentence has %d words; the %s limit is %d.", count, kind, limit),
			Hint:    "Divide the sentence into shorter sentences.",
			Line:    s.Line, Col: s.Col, EndLine: last.EndLine, EndCol: last.EndCol,
		})
	}
	return findings
}
