package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffbucket/vale/internal/lint"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.MinSeverity != "error" {
		t.Errorf("MinSeverity = %q, want error", c.MinSeverity)
	}
	if c.StrictVocabulary {
		t.Error("StrictVocabulary should default false")
	}
	if c.Sentence.ProcedureMax != 20 || c.Sentence.DescriptionMax != 25 {
		t.Errorf("sentence limits = %d/%d, want 20/25", c.Sentence.ProcedureMax, c.Sentence.DescriptionMax)
	}
	if c.Rules == nil {
		t.Error("Rules map should be non-nil")
	}
}

func TestLoadMissingPathReturnsDefaults(t *testing.T) {
	// An empty dir with no config file returns defaults, no error.
	dir := t.TempDir()
	c, err := Load("", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.MinSeverity != "error" || c.Sentence.DescriptionMax != 25 {
		t.Errorf("did not return defaults: %+v", c)
	}
}

func TestLoadExplicitFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	yaml := "" +
		"minSeverity: warning\n" +
		"strictVocabulary: true\n" +
		"sentence:\n" +
		"  procedureMax: 15\n" +
		"  descriptionMax: 30\n" +
		"rules:\n" +
		"  STE.PassiveVoice:\n" +
		"    disabled: true\n" +
		"  STE.IngForms:\n" +
		"    severity: error\n"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.MinSeverity != "warning" {
		t.Errorf("MinSeverity = %q, want warning", c.MinSeverity)
	}
	if !c.StrictVocabulary {
		t.Error("StrictVocabulary = false, want true")
	}
	if c.Sentence.ProcedureMax != 15 || c.Sentence.DescriptionMax != 30 {
		t.Errorf("sentence = %d/%d, want 15/30", c.Sentence.ProcedureMax, c.Sentence.DescriptionMax)
	}
	if !c.Rules["STE.PassiveVoice"].Disabled {
		t.Error("PassiveVoice should be disabled")
	}
	if c.Rules["STE.IngForms"].Severity != "error" {
		t.Errorf("IngForms severity = %q, want error", c.Rules["STE.IngForms"].Severity)
	}
}

func TestLoadSearchWalksUp(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".vale-ste.yml")
	if err := os.WriteFile(path, []byte("minSeverity: suggestion\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load("", sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.MinSeverity != "suggestion" {
		t.Errorf("MinSeverity = %q, want suggestion (found via parent search)", c.MinSeverity)
	}
}

func TestLoadYamlAlternateExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".vale-ste.yaml")
	if err := os.WriteFile(path, []byte("minSeverity: warning\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load("", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.MinSeverity != "warning" {
		t.Errorf("MinSeverity = %q", c.MinSeverity)
	}
}

func TestLoadReadError(t *testing.T) {
	// A path that does not exist is a read error.
	_, err := Load(filepath.Join(t.TempDir(), "missing.yml"), "")
	if err == nil {
		t.Fatal("expected read error for missing explicit path")
	}
}

func TestLoadBadYaml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(path, []byte("minSeverity: [not a string\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, ""); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadRejectsBadSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("minSeverity: bogus\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, ""); err == nil {
		t.Fatal("expected validation error for bad minSeverity")
	}
}

func TestLoadRejectsBadRuleSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	yaml := "rules:\n  STE.PassiveVoice:\n    severity: loud\n"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path, "")
	if err == nil {
		t.Fatal("expected validation error for bad rule severity")
	}
}

func TestLoadRejectsNegativeSentenceLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(path, []byte("sentence:\n  procedureMax: -1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, ""); err == nil {
		t.Fatal("expected validation error for negative limit")
	}
}

func TestEngineConfig(t *testing.T) {
	c := Default()
	c.Rules = map[string]RuleSetting{
		"STE.PassiveVoice": {Disabled: true},
		"STE.IngForms":     {Severity: "error"},
		"STE.Contractions": {Severity: ""}, // no override
	}
	ec := c.EngineConfig()
	if !ec.Disabled["STE.PassiveVoice"] {
		t.Error("PassiveVoice should be disabled in engine config")
	}
	if ec.SeverityOverrides["STE.IngForms"] != lint.SeverityError {
		t.Errorf("IngForms override = %q, want error", ec.SeverityOverrides["STE.IngForms"])
	}
	if _, ok := ec.SeverityOverrides["STE.Contractions"]; ok {
		t.Error("empty severity should not create an override")
	}
	if _, ok := ec.Disabled["STE.IngForms"]; ok {
		t.Error("IngForms should not be disabled")
	}
}

func TestGate(t *testing.T) {
	tests := []struct {
		name string
		min  string
		want lint.Severity
	}{
		{"warning", "warning", lint.SeverityWarning},
		{"empty defaults to error", "", lint.SeverityError},
		{"invalid defaults to error", "bogus", lint.SeverityError},
		{"suggestion", "suggestion", lint.SeveritySuggestion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{MinSeverity: tt.min}
			if got := c.Gate(); got != tt.want {
				t.Errorf("Gate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllowedVocabularyBuiltinOnByDefault(t *testing.T) {
	allowed := Default().AllowedVocabulary()
	for _, term := range []string{"commit", "component", "cache", "design token"} {
		if !allowed[term] {
			t.Errorf("built-in technical term %q missing from default allow set", term)
		}
	}
}

func TestAllowedVocabularyAllowAndDeny(t *testing.T) {
	c := Default()
	c.Vocabulary.Allow = []string{"Foobar", "  Widget  "} // mixed case and padding
	c.Vocabulary.Deny = []string{"Commit"}                // deny overrides the built-in
	allowed := c.AllowedVocabulary()
	if !allowed["foobar"] || !allowed["widget"] {
		t.Errorf("allow entries not normalized and added: %+v", allowed["foobar"])
	}
	if allowed["commit"] {
		t.Errorf("deny did not remove the built-in term")
	}
}

func TestAllowedVocabularyBuiltinOff(t *testing.T) {
	c := Default()
	off := false
	c.Vocabulary.BuiltinTechnicalTerms = &off
	c.Vocabulary.Allow = []string{"commit"}
	allowed := c.AllowedVocabulary()
	if allowed["component"] {
		t.Errorf("built-in terms present when builtinTechnicalTerms is off")
	}
	if !allowed["commit"] {
		t.Errorf("explicit allow must still work when the built-in set is off")
	}
}
