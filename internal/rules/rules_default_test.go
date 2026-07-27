package rules

import (
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
)

func TestDefaultRuleSet(t *testing.T) {
	got := Default(config.Default())
	wantIDs := []string{
		"STE.SentenceLength",
		"STE.Contractions",
		"STE.PassiveVoice",
		"STE.IngForms",
		"STE.PhrasalVerbs",
		"STE.OneInstruction",
		"STE.Vocabulary",
	}
	if len(got) != len(wantIDs) {
		t.Fatalf("rule count = %d, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID() != id {
			t.Errorf("rule %d = %q, want %q", i, got[i].ID(), id)
		}
	}
}

func TestDefaultNilConfig(t *testing.T) {
	got := Default(nil)
	if len(got) != 7 {
		t.Fatalf("rule count = %d, want 7", len(got))
	}
}

func TestDefaultUsesSentenceLimits(t *testing.T) {
	cfg := config.Default()
	cfg.Sentence.ProcedureMax = 2
	cfg.Sentence.DescriptionMax = 100
	rs := Default(cfg)
	// Run the sentence-length rule (first rule) against a 3-word list item.
	doc := parseMD("- one two three")
	got := rs[0].Check(doc)
	if len(got) != 1 {
		t.Fatalf("expected the tightened procedure limit to fire: %+v", got)
	}
}

// FuzzEngineRun runs the full default rule set through the engine over fuzzed
// input. It must never panic, with markdown on or off.
func FuzzEngineRun(f *testing.F) {
	seeds := []string{
		"",
		"Don't turn on the system; and then shut down.",
		"# Heading\n\n- Running the test was quickly done.\n",
		"The door was opened according to the manual.",
		"one two three four five six seven eight nine ten.",
		"```\ncode block\n```\nprose after",
		"It's a well-made café setting.",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	rs := Default(config.Default())
	strictRules := Default(&config.Config{
		StrictVocabulary: true,
		Sentence:         config.Sentence{ProcedureMax: 20, DescriptionMax: 25},
	})
	f.Fuzz(func(t *testing.T, text string) {
		for _, md := range []bool{false, true} {
			doc := lint.Parse("fuzz.md", text, lint.ParseOptions{Markdown: md})
			eng := lint.New(rs, lint.EngineConfig{})
			for _, fnd := range eng.Run(doc) {
				if fnd.RuleID == "" {
					t.Fatalf("finding with empty RuleID for %q", text)
				}
			}
			strictEng := lint.New(strictRules, lint.EngineConfig{})
			_ = strictEng.Run(doc)
		}
	})
}
