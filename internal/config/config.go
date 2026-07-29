// Package config loads and merges the vale configuration. The configuration is
// a small YAML file. All fields have safe defaults, so a missing file is fine.
package config

import (
	"fmt"
	"strings"

	"github.com/stuffbucket/vale/internal/lint"
	"github.com/stuffbucket/vale/internal/vocab"
)

// Sentence holds the word limits for the sentence-length rule.
type Sentence struct {
	ProcedureMax   int `yaml:"procedureMax"`
	DescriptionMax int `yaml:"descriptionMax"`
}

// RuleSetting overrides the behavior of one rule.
type RuleSetting struct {
	Disabled bool   `yaml:"disabled"`
	Severity string `yaml:"severity"`
}

// Vocabulary tunes which words the STE.Vocabulary rule treats as approved. The
// built-in software and design technical terms are on by default; allow adds
// project terms, and deny removes terms from the approved set so STE checks them
// again.
type Vocabulary struct {
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
	// BuiltinTechnicalTerms turns the built-in software and design terms on or
	// off. A nil pointer (the field absent) means on.
	BuiltinTechnicalTerms *bool `yaml:"builtinTechnicalTerms"`
}

// Files scopes which paths the linter considers when it walks a directory.
// Both lists are globs (filepath.Match syntax, matched against the path and the
// base name); a leading "**/" also matches the base name at any depth. exclude
// wins over include. Explicitly named file arguments are always linted.
type Files struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// Config is the full configuration.
type Config struct {
	// MinSeverity is the gate for the command line. The linter fails when it
	// finds a problem at this level or higher.
	MinSeverity string `yaml:"minSeverity"`
	// StrictVocabulary turns on the full unapproved-word list. It is off by
	// default because it makes many findings on ordinary prose.
	StrictVocabulary bool                   `yaml:"strictVocabulary"`
	Sentence         Sentence               `yaml:"sentence"`
	Vocabulary       Vocabulary             `yaml:"vocabulary"`
	Files            Files                  `yaml:"files"`
	Slop             Slop                   `yaml:"slop"`
	Rules            map[string]RuleSetting `yaml:"rules"`
}

// Slop turns on the STE.Slop* rule family: opt-in, advisory checks for markers
// of LLM-generated "slop." Off by default. See research/ai-slop/.
type Slop struct {
	Enabled bool `yaml:"enabled"`
}

// Default returns a configuration with the standard values.
func Default() *Config {
	return &Config{
		MinSeverity:      string(lint.SeverityError),
		StrictVocabulary: false,
		Sentence:         Sentence{ProcedureMax: 20, DescriptionMax: 25},
		Rules:            map[string]RuleSetting{},
	}
}

// Load resolves the layered configuration. It merges (lowest to highest) the
// system, user (XDG), project, and project-local files, then an explicit path
// when given. See layers.go for the precedence and merge rules. A missing file
// at any layer is skipped; the result always has safe defaults.
func Load(path, dir string) (*Config, error) {
	cfg := Default()
	for _, layer := range discoverLayers(dir, path) {
		fc, err := loadFileLayer(layer)
		if err != nil {
			return nil, err
		}
		mergeFile(cfg, fc)
	}
	if cfg.Rules == nil {
		cfg.Rules = map[string]RuleSetting{}
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate checks the configuration values.
func (c *Config) validate() error {
	if c.MinSeverity != "" {
		if _, err := lint.ParseSeverity(c.MinSeverity); err != nil {
			return err
		}
	}
	for id, r := range c.Rules {
		if r.Severity != "" {
			if _, err := lint.ParseSeverity(r.Severity); err != nil {
				return fmt.Errorf("rule %s: %w", id, err)
			}
		}
	}
	if c.Sentence.ProcedureMax < 0 || c.Sentence.DescriptionMax < 0 {
		return fmt.Errorf("sentence limits must not be negative")
	}
	return nil
}

// EngineConfig converts the rule settings into the form the engine needs.
func (c *Config) EngineConfig() lint.EngineConfig {
	ec := lint.EngineConfig{
		Disabled:          map[string]bool{},
		SeverityOverrides: map[string]lint.Severity{},
	}
	for id, r := range c.Rules {
		if r.Disabled {
			ec.Disabled[id] = true
		}
		if r.Severity != "" {
			if sev, err := lint.ParseSeverity(r.Severity); err == nil {
				ec.SeverityOverrides[id] = sev
			}
		}
	}
	return ec
}

// AllowedVocabulary returns the set of lower-case words and phrases that the
// STE.Vocabulary rule must treat as approved: the built-in software and design
// technical terms (unless turned off), plus vocabulary.allow, minus
// vocabulary.deny. deny wins over both allow and the built-in set.
func (c *Config) AllowedVocabulary() map[string]bool {
	allowed := map[string]bool{}
	if c.Vocabulary.BuiltinTechnicalTerms == nil || *c.Vocabulary.BuiltinTechnicalTerms {
		for _, t := range vocab.TechnicalTerms {
			allowed[normalizeTerm(t)] = true
		}
	}
	for _, t := range c.Vocabulary.Allow {
		if n := normalizeTerm(t); n != "" {
			allowed[n] = true
		}
	}
	for _, t := range c.Vocabulary.Deny {
		delete(allowed, normalizeTerm(t))
	}
	return allowed
}

// normalizeTerm lower-cases a term and collapses its surrounding whitespace.
func normalizeTerm(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

// Gate returns the minimum severity for the command-line exit code.
func (c *Config) Gate() lint.Severity {
	if c.MinSeverity == "" {
		return lint.SeverityError
	}
	sev, err := lint.ParseSeverity(c.MinSeverity)
	if err != nil {
		return lint.SeverityError
	}
	return sev
}
