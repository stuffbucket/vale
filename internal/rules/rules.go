// Package rules holds the Simplified Technical English rules. Each rule lives in
// its own file with a table-driven test. The Default function builds the full
// set of rules from the configuration.
package rules

import (
	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
)

// Default builds the standard set of rules. The configuration sets the sentence
// limits, the vocabulary options, and whether the opt-in STE.Slop* family runs.
func Default(cfg *config.Config) []lint.Rule {
	if cfg == nil {
		cfg = config.Default()
	}
	base := []lint.Rule{
		NewSentenceLengthRule(cfg.Sentence.ProcedureMax, cfg.Sentence.DescriptionMax),
		NewContractionRule(),
		NewPassiveVoiceRule(),
		NewIngFormRule(),
		NewPhrasalVerbRule(),
		NewOneInstructionRule(),
		NewVocabularyRule(cfg.StrictVocabulary, cfg.AllowedVocabulary()),
	}
	return append(base, slopRules(cfg.Slop.Enabled)...)
}
