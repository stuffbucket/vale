package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
)

// contractions is the set of contractions that Simplified Technical English
// does not allow. The keys are in lower case with a straight apostrophe.
var contractions = map[string]string{
	"aren't":    "are not",
	"can't":     "cannot",
	"couldn't":  "could not",
	"didn't":    "did not",
	"doesn't":   "does not",
	"don't":     "do not",
	"hadn't":    "had not",
	"hasn't":    "has not",
	"haven't":   "have not",
	"he'd":      "he had / he would",
	"he'll":     "he will",
	"he's":      "he is",
	"here's":    "here is",
	"i'd":       "I had / I would",
	"i'll":      "I will",
	"i'm":       "I am",
	"i've":      "I have",
	"isn't":     "is not",
	"it'd":      "it would",
	"it'll":     "it will",
	"it's":      "it is",
	"let's":     "let us",
	"mustn't":   "must not",
	"shan't":    "shall not",
	"she'd":     "she had / she would",
	"she'll":    "she will",
	"she's":     "she is",
	"shouldn't": "should not",
	"that's":    "that is",
	"there's":   "there is",
	"they'd":    "they would",
	"they'll":   "they will",
	"they're":   "they are",
	"they've":   "they have",
	"wasn't":    "was not",
	"we'd":      "we would",
	"we'll":     "we will",
	"we're":     "we are",
	"we've":     "we have",
	"weren't":   "were not",
	"what's":    "what is",
	"where's":   "where is",
	"who's":     "who is",
	"won't":     "will not",
	"wouldn't":  "would not",
	"you'd":     "you had / you would",
	"you'll":    "you will",
	"you're":    "you are",
	"you've":    "you have",
}

// ContractionRule reports contractions such as "do not" written as "don't".
type ContractionRule struct{}

// NewContractionRule builds the rule.
func NewContractionRule() *ContractionRule { return &ContractionRule{} }

// ID gives the stable identifier of the rule.
func (r *ContractionRule) ID() string { return "STE.Contractions" }

// Description gives a short summary of the rule.
func (r *ContractionRule) Description() string {
	return "Do not use contractions; write the full words."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *ContractionRule) DefaultSeverity() lint.Severity { return lint.SeverityError }

// Check reports each contraction in the document.
func (r *ContractionRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		for _, t := range s.Tokens {
			if t.Kind != lint.KindWord {
				continue
			}
			key := strings.ReplaceAll(t.Lower, "’", "'")
			full, ok := contractions[key]
			if !ok {
				continue
			}
			findings = append(findings, lint.Finding{
				Message: "Contraction \"" + t.Text + "\" is not allowed.",
				Hint:    "Write \"" + full + "\".",
				Line:    t.Line, Col: t.Col, EndLine: t.EndLine, EndCol: t.EndCol,
				Match: t.Text,
			})
		}
	}
	return findings
}
