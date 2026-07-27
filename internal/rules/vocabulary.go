package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/vocab"
)

// VocabularyRule reports words that are not in the approved Simplified Technical
// English vocabulary. It uses the substitution data that "vale gen" builds from
// the OpenSTE wordset. For each unapproved word or phrase, it gives the approved
// replacements. The strict option adds the bare unapproved words that have no
// direct replacement.
type VocabularyRule struct {
	strict  bool
	single  map[string][]string // one word -> approved replacements
	phrases []vocabPhrase       // multiword entries, longest first
	bare    map[string]bool     // strict-only words with no replacement
}

// vocabPhrase is a multiword unapproved entry.
type vocabPhrase struct {
	words   []string
	replace []string
}

// NewVocabularyRule builds the rule from the generated vocabulary data.
func NewVocabularyRule(strict bool) *VocabularyRule {
	r := &VocabularyRule{
		strict: strict,
		single: map[string][]string{},
		bare:   map[string]bool{},
	}
	for _, s := range vocab.Substitutions {
		if strings.Contains(s.Word, " ") {
			r.phrases = append(r.phrases, vocabPhrase{words: strings.Fields(s.Word), replace: s.Alternatives})
			continue
		}
		r.single[s.Word] = s.Alternatives
	}
	if strict {
		for _, w := range vocab.UnapprovedWords {
			r.bare[w] = true
		}
	}
	return r
}

// ID gives the stable identifier of the rule.
func (r *VocabularyRule) ID() string { return "STE.Vocabulary" }

// Description gives a short summary of the rule.
func (r *VocabularyRule) Description() string {
	return "Use only approved Simplified Technical English words."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *VocabularyRule) DefaultSeverity() lint.Severity { return lint.SeveritySuggestion }

// Check reports each unapproved word or phrase.
func (r *VocabularyRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		covered := make([]bool, len(toks))
		// Multiword phrases first, so a phrase wins over its single words.
		for i := range toks {
			for _, p := range r.phrases {
				if matchPhrase(toks, i, p.words) {
					first, last := toks[i], toks[i+len(p.words)-1]
					phrase := strings.Join(p.words, " ")
					findings = append(findings, vocabFinding(first, last, phrase, p.replace))
					for k := 0; k < len(p.words); k++ {
						covered[i+k] = true
					}
				}
			}
		}
		for i, t := range toks {
			if covered[i] {
				continue
			}
			if alts, ok := r.single[t.Lower]; ok {
				findings = append(findings, vocabFinding(t, t, t.Text, alts))
				continue
			}
			if r.strict && r.bare[t.Lower] {
				findings = append(findings, lint.Finding{
					Message: "The word \"" + t.Text + "\" is not in the approved vocabulary.",
					Hint:    "Use an approved word with the same meaning.",
					Line:    t.Line, Col: t.Col, EndLine: t.EndLine, EndCol: t.EndCol, Match: t.Text,
				})
			}
		}
	}
	return findings
}

// vocabFinding builds a finding for an unapproved word or phrase.
func vocabFinding(first, last lint.Token, text string, alts []string) lint.Finding {
	hint := "Use an approved word."
	if len(alts) > 0 {
		hint = "Use \"" + strings.Join(alts, "\" or \"") + "\"."
	}
	return lint.Finding{
		Message: "The word \"" + text + "\" is not approved Simplified Technical English.",
		Hint:    hint,
		Line:    first.Line, Col: first.Col, EndLine: last.EndLine, EndCol: last.EndCol,
		Match: text,
	}
}
