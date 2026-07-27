package lint

import "testing"

// fakeRule is a controllable Rule for engine tests.
type fakeRule struct {
	id       string
	desc     string
	sev      Severity
	findings []Finding
}

func (r fakeRule) ID() string                { return r.id }
func (r fakeRule) Description() string       { return r.desc }
func (r fakeRule) DefaultSeverity() Severity { return r.sev }
func (r fakeRule) Check(*Document) []Finding { return append([]Finding(nil), r.findings...) }

func newFake(id string, sev Severity, findings ...Finding) fakeRule {
	return fakeRule{id: id, desc: id + " desc", sev: sev, findings: findings}
}

func TestEngineRunSetsFields(t *testing.T) {
	rule := newFake("R1", SeverityWarning, Finding{Line: 1, Col: 1, Message: "m"})
	e := New([]Rule{rule}, EngineConfig{})
	doc := &Document{Path: "file.md"}
	got := e.Run(doc)
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1", len(got))
	}
	f := got[0]
	if f.RuleID != "R1" {
		t.Errorf("RuleID = %q, want R1", f.RuleID)
	}
	if f.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want warning", f.Severity)
	}
	if f.Path != "file.md" {
		t.Errorf("Path = %q, want file.md", f.Path)
	}
}

func TestEngineRunSortsAcrossRules(t *testing.T) {
	r1 := newFake("R1", SeverityError, Finding{Line: 3, Col: 1}, Finding{Line: 1, Col: 5})
	r2 := newFake("R2", SeverityError, Finding{Line: 1, Col: 1}, Finding{Line: 2, Col: 1})
	e := New([]Rule{r1, r2}, EngineConfig{})
	got := e.Run(&Document{})
	wantLines := []int{1, 1, 2, 3}
	if len(got) != len(wantLines) {
		t.Fatalf("findings = %d, want %d", len(got), len(wantLines))
	}
	for i, wl := range wantLines {
		if got[i].Line != wl {
			t.Errorf("finding %d line = %d, want %d", i, got[i].Line, wl)
		}
	}
	// At line 1: col 1 (R2) before col 5 (R1).
	if got[0].RuleID != "R2" || got[1].RuleID != "R1" {
		t.Errorf("order = %q,%q want R2,R1", got[0].RuleID, got[1].RuleID)
	}
}

func TestEngineDisabledRule(t *testing.T) {
	r1 := newFake("R1", SeverityError, Finding{Line: 1, Col: 1})
	r2 := newFake("R2", SeverityError, Finding{Line: 2, Col: 1})
	e := New([]Rule{r1, r2}, EngineConfig{Disabled: map[string]bool{"R1": true}})
	got := e.Run(&Document{})
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1", len(got))
	}
	if got[0].RuleID != "R2" {
		t.Errorf("RuleID = %q, want R2", got[0].RuleID)
	}
}

func TestEngineDisabledFalseKeepsRule(t *testing.T) {
	r1 := newFake("R1", SeverityError, Finding{Line: 1, Col: 1})
	e := New([]Rule{r1}, EngineConfig{Disabled: map[string]bool{"R1": false}})
	if got := e.Run(&Document{}); len(got) != 1 {
		t.Fatalf("findings = %d, want 1", len(got))
	}
}

func TestEngineSeverityOverride(t *testing.T) {
	r1 := newFake("R1", SeverityError, Finding{Line: 1, Col: 1})
	e := New([]Rule{r1}, EngineConfig{SeverityOverrides: map[string]Severity{"R1": SeveritySuggestion}})
	got := e.Run(&Document{})
	if got[0].Severity != SeveritySuggestion {
		t.Errorf("Severity = %q, want suggestion", got[0].Severity)
	}
}

func TestEngineIgnoresInvalidOverride(t *testing.T) {
	r1 := newFake("R1", SeverityError, Finding{Line: 1, Col: 1})
	e := New([]Rule{r1}, EngineConfig{SeverityOverrides: map[string]Severity{"R1": Severity("bogus")}})
	got := e.Run(&Document{})
	if got[0].Severity != SeverityError {
		t.Errorf("Severity = %q, want error (invalid override ignored)", got[0].Severity)
	}
}

func TestEngineRules(t *testing.T) {
	r1 := newFake("R1", SeverityError)
	r2 := newFake("R2", SeverityError)
	e := New([]Rule{r1, r2}, EngineConfig{Disabled: map[string]bool{"R1": true}})
	// Rules() includes disabled rules.
	if len(e.Rules()) != 2 {
		t.Errorf("Rules() = %d, want 2", len(e.Rules()))
	}
}

func TestEngineRunFiltered(t *testing.T) {
	rErr := newFake("E", SeverityError, Finding{Line: 1, Col: 1})
	rWarn := newFake("W", SeverityWarning, Finding{Line: 2, Col: 1})
	rSug := newFake("S", SeveritySuggestion, Finding{Line: 3, Col: 1})
	e := New([]Rule{rErr, rWarn, rSug}, EngineConfig{})

	if got := e.RunFiltered(&Document{}, SeverityWarning); len(got) != 2 {
		t.Errorf("filter warning = %d, want 2", len(got))
	}
	if got := e.RunFiltered(&Document{}, SeverityError); len(got) != 1 {
		t.Errorf("filter error = %d, want 1", len(got))
	}
	if got := e.RunFiltered(&Document{}, SeveritySuggestion); len(got) != 3 {
		t.Errorf("filter suggestion = %d, want 3", len(got))
	}
}

func TestEngineRunFilteredInvalidMinReturnsAll(t *testing.T) {
	rErr := newFake("E", SeverityError, Finding{Line: 1, Col: 1})
	rSug := newFake("S", SeveritySuggestion, Finding{Line: 2, Col: 1})
	e := New([]Rule{rErr, rSug}, EngineConfig{})
	if got := e.RunFiltered(&Document{}, Severity("")); len(got) != 2 {
		t.Errorf("invalid min = %d, want all 2", len(got))
	}
}

func TestEngineNoRules(t *testing.T) {
	e := New(nil, EngineConfig{})
	if got := e.Run(&Document{}); len(got) != 0 {
		t.Errorf("findings = %d, want 0", len(got))
	}
}
