package rules

import "github.com/stuffbucket/vale/internal/lint"

// OneInstructionRule reports a procedure step that holds more than one
// instruction. Simplified Technical English keeps one instruction in one
// sentence. The rule looks at list items, which are the usual form of a
// procedure step.
type OneInstructionRule struct{}

// NewOneInstructionRule builds the rule.
func NewOneInstructionRule() *OneInstructionRule { return &OneInstructionRule{} }

// ID gives the stable identifier of the rule.
func (r *OneInstructionRule) ID() string { return "STE.OneInstruction" }

// Description gives a short summary of the rule.
func (r *OneInstructionRule) Description() string {
	return "Write one instruction in one sentence."
}

// DefaultSeverity gives the severity that the rule reports by default.
func (r *OneInstructionRule) DefaultSeverity() lint.Severity { return lint.SeverityWarning }

// Check reports each procedure step that joins two instructions. It finds a
// semicolon, or the words "and then", which are clear signs of two
// instructions in one step.
func (r *OneInstructionRule) Check(doc *lint.Document) []lint.Finding {
	var findings []lint.Finding
	for _, s := range doc.Sentences {
		if s.Block != lint.BlockListItem {
			continue
		}
		for i, t := range s.Tokens {
			if t.Kind == lint.KindPunct && t.Text == ";" {
				findings = append(findings, finding(t, "The semicolon joins two instructions."))
				continue
			}
			if t.Kind == lint.KindWord && t.Lower == "and" {
				if next, ok := nextWord(s.Tokens, i); ok && next.Lower == "then" {
					findings = append(findings, finding(t, "The words \"and then\" join two instructions."))
				}
			}
		}
	}
	return findings
}

// finding builds a one-instruction finding at a token.
func finding(t lint.Token, message string) lint.Finding {
	return lint.Finding{
		Message: message,
		Hint:    "Divide the step into one instruction for each sentence.",
		Line:    t.Line, Col: t.Col, EndLine: t.EndLine, EndCol: t.EndCol,
	}
}

// nextWord returns the next word token after index i.
func nextWord(toks []lint.Token, i int) (lint.Token, bool) {
	for j := i + 1; j < len(toks); j++ {
		if toks[j].Kind == lint.KindWord {
			return toks[j], true
		}
	}
	return lint.Token{}, false
}
