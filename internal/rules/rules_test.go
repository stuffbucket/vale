package rules

import (
	"testing"

	"github.com/stuffbucket/vale/internal/config"
	"github.com/stuffbucket/vale/internal/lint"
)

// parseMD parses text as Markdown.
func parseMD(text string) *lint.Document {
	return lint.Parse("t.md", text, lint.ParseOptions{Markdown: true})
}

// parsePlain parses text without Markdown.
func parsePlain(text string) *lint.Document {
	return lint.Parse("t.txt", text, lint.ParseOptions{Markdown: false})
}

func TestSentenceLengthRuleMetadata(t *testing.T) {
	r := NewSentenceLengthRule(20, 25)
	if r.ID() != "STE.SentenceLength" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.DefaultSeverity() != lint.SeverityError {
		t.Errorf("severity = %q", r.DefaultSeverity())
	}
	if r.Description() == "" {
		t.Error("empty description")
	}
}

func TestSentenceLengthRule(t *testing.T) {
	tests := []struct {
		name      string
		procMax   int
		descMax   int
		text      string
		markdown  bool
		wantCount int
		wantMsg   string
	}{
		{
			name: "paragraph over description limit", descMax: 5, procMax: 3,
			text:      "one two three four five six.",
			wantCount: 1,
			wantMsg:   "Sentence has 6 words; the description limit is 5.",
		},
		{
			name: "paragraph at limit ok", descMax: 6, procMax: 3,
			text: "one two three four five six.", wantCount: 0,
		},
		{
			name: "list item uses procedure limit", procMax: 3, descMax: 25,
			text: "- one two three four", markdown: true, wantCount: 1,
			wantMsg: "Sentence has 4 words; the procedure limit is 3.",
		},
		{
			name: "heading exempt", procMax: 1, descMax: 1,
			text: "# one two three four", markdown: true, wantCount: 0,
		},
		{
			name: "description limit disabled", descMax: 0, procMax: 3,
			text: "one two three four five six seven.", wantCount: 0,
		},
		{
			name: "procedure limit disabled", procMax: 0, descMax: 25,
			text: "- one two three four five", markdown: true, wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *lint.Document
			if tt.markdown {
				doc = parseMD(tt.text)
			} else {
				doc = parsePlain(tt.text)
			}
			got := NewSentenceLengthRule(tt.procMax, tt.descMax).Check(doc)
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 && got[0].Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", got[0].Message, tt.wantMsg)
			}
		})
	}
}

func TestSentenceLengthRulePositions(t *testing.T) {
	doc := parseMD("- one two three four")
	got := NewSentenceLengthRule(3, 25).Check(doc)
	if len(got) != 1 {
		t.Fatalf("count = %d", len(got))
	}
	f := got[0]
	if f.Line != 1 || f.Col != 3 {
		t.Errorf("start = %d:%d, want 1:3", f.Line, f.Col)
	}
	if f.EndLine != 1 || f.EndCol != 21 {
		t.Errorf("end = %d:%d, want 1:21", f.EndLine, f.EndCol)
	}
	if f.Hint != "Divide the sentence into shorter sentences." {
		t.Errorf("hint = %q", f.Hint)
	}
}

func TestContractionRule(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCount int
		wantMatch string
		wantHint  string
	}{
		{"straight apostrophe", "Don't stop.", 1, "Don't", `Write "do not".`},
		{"its is flagged", "It's here.", 1, "It's", `Write "it is".`},
		{"cannot", "You can't do it.", 1, "can't", `Write "cannot".`},
		{"possessive not flagged", "The system's fan is here.", 0, "", ""},
		{"no contraction", "Do not stop the fan.", 0, "", ""},
		{"multiple", "Don't and can't work.", 2, "Don't", `Write "do not".`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewContractionRule().Check(parsePlain(tt.text))
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 {
				if got[0].Match != tt.wantMatch {
					t.Errorf("match = %q, want %q", got[0].Match, tt.wantMatch)
				}
				if got[0].Hint != tt.wantHint {
					t.Errorf("hint = %q, want %q", got[0].Hint, tt.wantHint)
				}
				wantMsg := "Contraction \"" + tt.wantMatch + "\" is not allowed."
				if got[0].Message != wantMsg {
					t.Errorf("message = %q, want %q", got[0].Message, wantMsg)
				}
			}
		})
	}
}

func TestContractionRuleCurlyApostrophe(t *testing.T) {
	got := NewContractionRule().Check(parsePlain("It’s broken."))
	if len(got) != 1 {
		t.Fatalf("count = %d, want 1", len(got))
	}
	if got[0].Match != "It’s" {
		t.Errorf("match = %q", got[0].Match)
	}
	if got[0].Hint != `Write "it is".` {
		t.Errorf("hint = %q", got[0].Hint)
	}
}

func TestContractionRulePositions(t *testing.T) {
	got := NewContractionRule().Check(parsePlain("Don't stop."))
	f := got[0]
	if f.Line != 1 || f.Col != 1 || f.EndCol != 6 {
		t.Errorf("span = %d:%d..%d, want 1:1..6", f.Line, f.Col, f.EndCol)
	}
}

func TestPassiveVoiceRule(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCount int
		wantMatch string
	}{
		{"simple ed participle", "The door was opened.", 1, "was opened"},
		{"with adverb", "It was quickly removed.", 1, "was removed"},
		{"irregular participle", "The file was written.", 1, "was written"},
		{"be form", "The parts are checked.", 1, "are checked"},
		{"non participle ed excluded", "It is indeed here.", 0, ""},
		{"be verb no participle", "The fan is here.", 0, ""},
		{"no be verb", "He opened the door.", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPassiveVoiceRule().Check(parsePlain(tt.text))
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 {
				if got[0].Match != tt.wantMatch {
					t.Errorf("match = %q, want %q", got[0].Match, tt.wantMatch)
				}
				wantMsg := "Passive voice: \"" + tt.wantMatch + "\"."
				if got[0].Message != wantMsg {
					t.Errorf("message = %q, want %q", got[0].Message, wantMsg)
				}
			}
		})
	}
}

func TestPassiveVoiceRuleMetadata(t *testing.T) {
	r := NewPassiveVoiceRule()
	if r.ID() != "STE.PassiveVoice" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.DefaultSeverity() != lint.SeverityWarning {
		t.Errorf("severity = %q", r.DefaultSeverity())
	}
}

func TestPassiveVoicePositions(t *testing.T) {
	got := NewPassiveVoiceRule().Check(parsePlain("The door was opened."))
	f := got[0]
	if f.Col != 10 || f.EndCol != 20 {
		t.Errorf("span cols = %d..%d, want 10..20", f.Col, f.EndCol)
	}
}

func TestIngFormRule(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		markdown  bool
		wantCount int
		wantMatch string
	}{
		{"flags ing verb", "Running is hard.", false, 1, "Running"},
		{"allow-list words", "The warning string setting is here.", false, 0, ""},
		{"short ing excluded", "The ping is here.", false, 0, ""},
		{"heading exempt", "# Installing the system", true, 0, ""},
		{"multiple", "Testing and checking now.", false, 2, "Testing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *lint.Document
			if tt.markdown {
				doc = parseMD(tt.text)
			} else {
				doc = parsePlain(tt.text)
			}
			got := NewIngFormRule().Check(doc)
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 {
				if got[0].Match != tt.wantMatch {
					t.Errorf("match = %q, want %q", got[0].Match, tt.wantMatch)
				}
				wantMsg := "The -ing form \"" + tt.wantMatch + "\" is hard to read in an instruction."
				if got[0].Message != wantMsg {
					t.Errorf("message = %q, want %q", got[0].Message, wantMsg)
				}
			}
		})
	}
}

func TestIngFormRuleMetadata(t *testing.T) {
	r := NewIngFormRule()
	if r.ID() != "STE.IngForms" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.DefaultSeverity() != lint.SeveritySuggestion {
		t.Errorf("severity = %q", r.DefaultSeverity())
	}
}

func TestPhrasalVerbRule(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCount int
		wantMatch string
		wantHint  string
	}{
		{"turn on", "Turn on the light.", 1, "turn on", `Use "start".`},
		{"set up", "Set up the system.", 1, "set up", `Use "install or prepare".`},
		{"carry out", "Carry out the test.", 1, "carry out", `Use "do or complete".`},
		{"shut down", "Shut down the system.", 1, "shut down", `Use "stop".`},
		{"no phrasal", "Start the light.", 0, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPhrasalVerbRule().Check(parsePlain(tt.text))
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 {
				if got[0].Match != tt.wantMatch {
					t.Errorf("match = %q, want %q", got[0].Match, tt.wantMatch)
				}
				if got[0].Hint != tt.wantHint {
					t.Errorf("hint = %q, want %q", got[0].Hint, tt.wantHint)
				}
				wantMsg := "Phrasal verb \"" + tt.wantMatch + "\" is not clear."
				if got[0].Message != wantMsg {
					t.Errorf("message = %q, want %q", got[0].Message, wantMsg)
				}
			}
		})
	}
}

func TestPhrasalVerbPositions(t *testing.T) {
	got := NewPhrasalVerbRule().Check(parsePlain("Turn on the light."))
	f := got[0]
	if f.Col != 1 || f.EndCol != 8 {
		t.Errorf("span cols = %d..%d, want 1..8", f.Col, f.EndCol)
	}
}

func TestOneInstructionRule(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantCount int
		wantMsg   string
	}{
		{"semicolon in list", "- Open the door; close it.", 1, "The semicolon joins two instructions."},
		{"and then in list", "- Open the door and then close it.", 1, `The words "and then" join two instructions.`},
		{"paragraph not checked", "Open the door; close it.", 0, ""},
		{"list with no joiner", "- Open the door.", 0, ""},
		{"and without then", "- Open the door and the window.", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOneInstructionRule().Check(parseMD(tt.text))
			if len(got) != tt.wantCount {
				t.Fatalf("count = %d, want %d: %+v", len(got), tt.wantCount, got)
			}
			if tt.wantCount > 0 {
				if got[0].Message != tt.wantMsg {
					t.Errorf("message = %q, want %q", got[0].Message, tt.wantMsg)
				}
				if got[0].Hint != "Divide the step into one instruction for each sentence." {
					t.Errorf("hint = %q", got[0].Hint)
				}
			}
		})
	}
}

func TestOneInstructionRuleMetadata(t *testing.T) {
	r := NewOneInstructionRule()
	if r.ID() != "STE.OneInstruction" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.DefaultSeverity() != lint.SeverityWarning {
		t.Errorf("severity = %q", r.DefaultSeverity())
	}
}

func TestVocabularyRuleSingle(t *testing.T) {
	got := NewVocabularyRule(false, nil).Check(parsePlain("Open the valve able tool."))
	if len(got) != 1 {
		t.Fatalf("count = %d, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Match != "able" {
		t.Errorf("match = %q, want able", f.Match)
	}
	if f.Message != `The word "able" is not approved Simplified Technical English.` {
		t.Errorf("message = %q", f.Message)
	}
	if f.Hint != `Use "can" or "possible".` {
		t.Errorf("hint = %q", f.Hint)
	}
}

func TestVocabularyRulePhrase(t *testing.T) {
	got := NewVocabularyRule(false, nil).Check(parsePlain("Open the valve according to the manual."))
	if len(got) != 1 {
		t.Fatalf("count = %d, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Match != "according to" {
		t.Errorf("match = %q, want 'according to'", f.Match)
	}
	if f.Hint != `Use "refer" or "tell".` {
		t.Errorf("hint = %q", f.Hint)
	}
}

func TestVocabularyRuleStrictBareWord(t *testing.T) {
	// "bolt" is an unapproved bare word with no substitution.
	strict := NewVocabularyRule(true, nil).Check(parsePlain("Open the bolt here."))
	if len(strict) != 1 {
		t.Fatalf("strict count = %d, want 1: %+v", len(strict), strict)
	}
	f := strict[0]
	if f.Match != "bolt" {
		t.Errorf("match = %q, want bolt", f.Match)
	}
	if f.Message != `The word "bolt" is not in the approved vocabulary.` {
		t.Errorf("message = %q", f.Message)
	}
	if f.Hint != "Use an approved word with the same meaning." {
		t.Errorf("hint = %q", f.Hint)
	}

	// Without strict, the bare word produces no finding.
	loose := NewVocabularyRule(false, nil).Check(parsePlain("Open the bolt here."))
	if len(loose) != 0 {
		t.Fatalf("loose count = %d, want 0: %+v", len(loose), loose)
	}
}

func TestVocabularyRuleMetadata(t *testing.T) {
	r := NewVocabularyRule(false, nil)
	if r.ID() != "STE.Vocabulary" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.DefaultSeverity() != lint.SeveritySuggestion {
		t.Errorf("severity = %q", r.DefaultSeverity())
	}
}

func TestVocabularyRuleAllowSkipsWord(t *testing.T) {
	// "able" is flagged by default, but an allow entry approves it.
	allowed := map[string]bool{"able": true}
	got := NewVocabularyRule(false, allowed).Check(parsePlain("Open the valve able tool."))
	if len(got) != 0 {
		t.Fatalf("count = %d, want 0 (allowed): %+v", len(got), got)
	}
}

func TestVocabularyRuleAllowSkipsPhrase(t *testing.T) {
	// An allowed multiword technical phrase is not flagged.
	allowed := map[string]bool{"according to": true}
	got := NewVocabularyRule(false, allowed).Check(parsePlain("Open the valve according to the manual."))
	if len(got) != 0 {
		t.Fatalf("count = %d, want 0 (allowed phrase): %+v", len(got), got)
	}
}

func TestSlopRulesOptIn(t *testing.T) {
	cfg := config.Default()
	// Off by default: no STE.Slop* rule present.
	for _, r := range Default(cfg) {
		if r.ID() == "STE.SlopVocabulary" {
			t.Fatal("slop rules must be off by default")
		}
	}
	// Enabled: the family appears.
	cfg.Slop.Enabled = true
	found := false
	for _, r := range Default(cfg) {
		if r.ID() == "STE.SlopVocabulary" {
			found = true
		}
	}
	if !found {
		t.Fatal("slop rules should register when enabled")
	}
}

func TestSlopVocabularyRuleFlagsWatchlist(t *testing.T) {
	r := NewSlopVocabularyRule(nil)
	got := r.Check(parsePlain("We delve into the intricate realm of valves."))
	if len(got) < 2 {
		t.Fatalf("expected several slop words flagged, got %d: %+v", len(got), got)
	}
	if r.DefaultSeverity() != lint.SeverityWarning {
		t.Errorf("severity = %q, want warning", r.DefaultSeverity())
	}
	// A plain word is not flagged.
	if plain := r.Check(parsePlain("Open the valve.")); len(plain) != 0 {
		t.Errorf("plain sentence should not be flagged: %+v", plain)
	}
}
