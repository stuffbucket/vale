package lint

// Engine holds the set of rules and the per-rule settings, and applies them to
// documents. Build an Engine with New.
type Engine struct {
	rules     []Rule
	disabled  map[string]bool
	overrides map[string]Severity
}

// EngineConfig gives the engine its rule settings. The maps use rule
// identifiers as keys.
type EngineConfig struct {
	// Disabled turns off a rule when the value is true.
	Disabled map[string]bool
	// SeverityOverrides changes the severity that a rule reports.
	SeverityOverrides map[string]Severity
}

// New builds an engine from a set of rules and a configuration.
func New(rules []Rule, cfg EngineConfig) *Engine {
	e := &Engine{
		rules:     rules,
		disabled:  map[string]bool{},
		overrides: map[string]Severity{},
	}
	for id, off := range cfg.Disabled {
		e.disabled[id] = off
	}
	for id, sev := range cfg.SeverityOverrides {
		if sev.Valid() {
			e.overrides[id] = sev
		}
	}
	return e
}

// Rules gives the rules that the engine holds, including the disabled ones.
func (e *Engine) Rules() []Rule { return e.rules }

// Run applies every enabled rule to the document and returns the findings in
// reading order. The engine sets the severity and the path on each finding.
func (e *Engine) Run(doc *Document) []Finding {
	var findings []Finding
	for _, rule := range e.rules {
		if e.disabled[rule.ID()] {
			continue
		}
		sev := rule.DefaultSeverity()
		if o, ok := e.overrides[rule.ID()]; ok {
			sev = o
		}
		for _, f := range rule.Check(doc) {
			f.RuleID = rule.ID()
			f.Severity = sev
			f.Path = doc.Path
			findings = append(findings, f)
		}
	}
	SortFindings(findings)
	return findings
}

// RunFiltered runs the rules and keeps only findings at or above the minimum
// severity.
func (e *Engine) RunFiltered(doc *Document, min Severity) []Finding {
	all := e.Run(doc)
	if !min.Valid() {
		return all
	}
	out := all[:0:0]
	for _, f := range all {
		if f.Severity.AtLeast(min) {
			out = append(out, f)
		}
	}
	return out
}
