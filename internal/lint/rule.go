package lint

// Rule is one check that reads a document and reports findings. Every rule has
// a stable identifier, a short description, and a default severity. A rule must
// set the RuleID field on each finding, but it may leave the severity empty;
// the engine fills in the effective severity.
type Rule interface {
	ID() string
	Description() string
	DefaultSeverity() Severity
	Check(doc *Document) []Finding
}
