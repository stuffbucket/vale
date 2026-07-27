// Package config loads and merges the vale configuration. The configuration is
// a small YAML file. All fields have safe defaults, so a missing file is fine.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stuffbucket/vale/internal/lint"
	"gopkg.in/yaml.v3"
)

// DefaultFileNames are the file names that Load looks for when the caller does
// not give an explicit path.
var DefaultFileNames = []string{".vale-ste.yml", ".vale-ste.yaml"}

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

// Config is the full configuration.
type Config struct {
	// MinSeverity is the gate for the command line. The linter fails when it
	// finds a problem at this level or higher.
	MinSeverity string `yaml:"minSeverity"`
	// StrictVocabulary turns on the full unapproved-word list. It is off by
	// default because it makes many findings on ordinary prose.
	StrictVocabulary bool                   `yaml:"strictVocabulary"`
	Sentence         Sentence               `yaml:"sentence"`
	Rules            map[string]RuleSetting `yaml:"rules"`
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

// Load reads a configuration file. When path is empty, it looks for a default
// file name in dir and its parents. When it finds no file, it returns the
// default configuration.
func Load(path, dir string) (*Config, error) {
	if path == "" {
		path = search(dir)
	}
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Rules == nil {
		cfg.Rules = map[string]RuleSetting{}
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// search walks up from dir and returns the first default config file it finds.
func search(dir string) string {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		for _, name := range DefaultFileNames {
			candidate := filepath.Join(abs, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
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
