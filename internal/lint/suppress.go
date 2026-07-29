package lint

import (
	"regexp"
	"strings"
)

// directiveRe finds a `<!-- vale ... -->` inline directive and captures its body.
var directiveRe = regexp.MustCompile(`(?i)<!--\s*vale\s+(.*?)\s*-->`)

// Suppressions records, per source line, which rule ids (or "*" for all) are
// silenced by inline directives. See docs/decisions/0007.
type Suppressions struct {
	lineRules map[int]map[string]bool
}

// Suppressed reports whether a finding for ruleID on the given 1-based line is
// silenced.
func (s *Suppressions) Suppressed(line int, ruleID string) bool {
	if s == nil {
		return false
	}
	rs := s.lineRules[line]
	if rs == nil {
		return false
	}
	return rs["*"] || rs[ruleID]
}

// mark records that ruleID (or "*") is suppressed on a line.
func (s *Suppressions) mark(line int, rule string) {
	if line < 1 {
		return
	}
	if s.lineRules[line] == nil {
		s.lineRules[line] = map[string]bool{}
	}
	s.lineRules[line][rule] = true
}

// ParseSuppressions scans the source for `<!-- vale ... -->` directives and
// builds the per-line suppression map. Supported directives:
//
//	<!-- vale off -->            <!-- vale on -->              (region, all rules)
//	<!-- vale STE.Rule = NO -->  <!-- vale STE.Rule = YES -->  (region, one rule)
//	<!-- vale disable-line [ids] -->        (this line)
//	<!-- vale disable-next-line [ids] -->   (the next line)
//
// Region directives take effect from the line they appear on (inclusive), at
// line granularity.
func ParseSuppressions(text string) *Suppressions {
	s := &Suppressions{lineRules: map[int]map[string]bool{}}
	allOff := false
	ruleOff := map[string]bool{}
	for i, line := range strings.Split(text, "\n") {
		lineNum := i + 1
		for _, m := range directiveRe.FindAllStringSubmatch(line, -1) {
			applyDirective(s, strings.Fields(m[1]), lineNum, &allOff, ruleOff)
		}
		// Record the region state that applies to this line.
		if allOff {
			s.mark(lineNum, "*")
		}
		for r := range ruleOff {
			s.mark(lineNum, r)
		}
	}
	return s
}

// applyDirective interprets one directive's tokens and updates the running
// region state or records a line-scoped suppression.
func applyDirective(s *Suppressions, tok []string, lineNum int, allOff *bool, ruleOff map[string]bool) {
	if len(tok) == 0 {
		return
	}
	switch strings.ToLower(tok[0]) {
	case "off":
		*allOff = true
	case "on":
		*allOff = false
	case "disable-line":
		markLine(s, lineNum, tok[1:])
	case "disable-next-line":
		markLine(s, lineNum+1, tok[1:])
	default:
		// "<RuleID> = NO" / "<RuleID> = YES" region toggle for one rule.
		if len(tok) == 3 && tok[1] == "=" {
			switch strings.ToUpper(tok[2]) {
			case "NO":
				ruleOff[tok[0]] = true
			case "YES":
				delete(ruleOff, tok[0])
			}
		}
	}
}

// markLine records a line-scoped suppression: the given rule ids, or "*" when
// none are named.
func markLine(s *Suppressions, line int, ids []string) {
	if len(ids) == 0 {
		s.mark(line, "*")
		return
	}
	for _, id := range ids {
		s.mark(line, id)
	}
}

// FilterSuppressed drops findings silenced by the directives in text.
func FilterSuppressed(findings []Finding, text string) []Finding {
	sup := ParseSuppressions(text)
	out := findings[:0:0]
	for _, f := range findings {
		if !sup.Suppressed(f.Line, f.RuleID) {
			out = append(out, f)
		}
	}
	return out
}
