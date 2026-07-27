// Package lint contains the Simplified Technical English analysis engine:
// the document model, the tokenizer, the rule interface, and the runner that
// applies rules to text and collects findings.
package lint

import (
	"fmt"
	"strings"
)

// Severity tells how important a finding is. The three levels match the model
// in the specification: error, warning, and suggestion.
type Severity string

// The severity levels, from most important to least important.
const (
	SeverityError      Severity = "error"
	SeverityWarning    Severity = "warning"
	SeveritySuggestion Severity = "suggestion"
)

// rank gives a number to each severity so that the engine can compare two
// levels. A larger number is a more important level.
func (s Severity) rank() int {
	switch s {
	case SeverityError:
		return 3
	case SeverityWarning:
		return 2
	case SeveritySuggestion:
		return 1
	default:
		return 0
	}
}

// Valid tells if the severity is one of the three known levels.
func (s Severity) Valid() bool {
	return s.rank() != 0
}

// AtLeast tells if this severity is as important as the other severity, or
// more important.
func (s Severity) AtLeast(other Severity) bool {
	return s.rank() >= other.rank()
}

// ParseSeverity reads a severity from text. It accepts the level names in any
// letter case. It gives an error if the text is not a known level.
func ParseSeverity(text string) (Severity, error) {
	s := Severity(strings.ToLower(strings.TrimSpace(text)))
	if !s.Valid() {
		return "", fmt.Errorf("unknown severity %q (use error, warning, or suggestion)", text)
	}
	return s, nil
}
