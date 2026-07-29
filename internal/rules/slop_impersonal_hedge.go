package rules

import (
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/slop"
)

// SlopImpersonalHedgeRule flags the impersonal-it modal padding pattern
// ("it can be argued that", "it is possible that"), which is wordy, evasive, and
// passive-adjacent. Opt-in STE.Slop* family.
type SlopImpersonalHedgeRule struct{}

// NewSlopImpersonalHedgeRule builds the rule.
func NewSlopImpersonalHedgeRule() *SlopImpersonalHedgeRule { return &SlopImpersonalHedgeRule{} }

// ID gives the stable identifier of the rule.
func (r *SlopImpersonalHedgeRule) ID() string { return "STE.SlopImpersonalHedge" }

// Description gives a short summary of the rule.
func (r *SlopImpersonalHedgeRule) Description() string {
	return "Avoid impersonal-it modal padding; state the fact directly."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *SlopImpersonalHedgeRule) DefaultSeverity() lint.Severity { return lint.SeverityWarning }

// matchTemplate reports whether the tokens at i match a per-position option
// pattern: each position's lower-cased token must be one of that position's
// options.
func matchTemplate(toks []lint.Token, i int, pattern [][]string) bool {
	if i+len(pattern) > len(toks) {
		return false
	}
	for k, opts := range pattern {
		if !containsString(opts, toks[i+k].Lower) {
			return false
		}
	}
	return true
}

// containsString reports whether s is in opts.
func containsString(opts []string, s string) bool {
	for _, o := range opts {
		if o == s {
			return true
		}
	}
	return false
}

// Check reports each impersonal hedge template.
func (r *SlopImpersonalHedgeRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		toks := wordTokens(s)
		for i := range toks {
			for _, pat := range slop.ImpersonalHedgePatterns {
				if matchTemplate(toks, i, pat) {
					first, last := toks[i], toks[i+len(pat)-1]
					words := make([]string, len(pat))
					for k := range pat {
						words[k] = toks[i+k].Text
					}
					text := strings.Join(words, " ")
					findings = append(findings, lint.Finding{
						Message: "Impersonal hedge \"" + text + "\" is wordy and evasive.",
						Hint:    "State the fact directly in the active voice.",
						Line:    first.Line, Col: first.Col, EndLine: last.EndLine, EndCol: last.EndCol,
						Match: text,
					})
				}
			}
		}
	}
	return findings
}
