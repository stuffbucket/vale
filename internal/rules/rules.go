// Package rules holds the Simplified Technical English rules. Each rule lives in
// its own file with a table-driven test. The Default function builds the full
// set of rules from the configuration.
package rules

import (
	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
)

// Default builds the standard set of rules. The configuration sets the sentence
// limits and turns the strict vocabulary list on or off.
func Default(cfg *config.Config) []lint.Rule {
	if cfg == nil {
		cfg = config.Default()
	}
	return []lint.Rule{
		NewSentenceLengthRule(cfg.Sentence.ProcedureMax, cfg.Sentence.DescriptionMax),
		NewContractionRule(),
		NewPassiveVoiceRule(),
		NewIngFormRule(),
		NewPhrasalVerbRule(),
		NewOneInstructionRule(),
		NewVocabularyRule(cfg.StrictVocabulary),
	}
}
