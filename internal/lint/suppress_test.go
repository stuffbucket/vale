package lint

import "testing"

func TestSuppressRegionOffOn(t *testing.T) {
	text := "line1\n<!-- vale off -->\nline3\nline4\n<!-- vale on -->\nline6"
	s := ParseSuppressions(text)
	if s.Suppressed(1, "STE.X") {
		t.Errorf("line 1 before off should not be suppressed")
	}
	if !s.Suppressed(3, "STE.X") || !s.Suppressed(4, "STE.Anything") {
		t.Errorf("lines inside off/on region should be suppressed for all rules")
	}
	if s.Suppressed(6, "STE.X") {
		t.Errorf("line 6 after on should not be suppressed")
	}
}

func TestSuppressRuleRegion(t *testing.T) {
	text := "<!-- vale STE.PassiveVoice = NO -->\nhere\n<!-- vale STE.PassiveVoice = YES -->\nthere"
	s := ParseSuppressions(text)
	if !s.Suppressed(2, "STE.PassiveVoice") {
		t.Errorf("named rule should be suppressed in region")
	}
	if s.Suppressed(2, "STE.Contractions") {
		t.Errorf("only the named rule should be suppressed")
	}
	if s.Suppressed(4, "STE.PassiveVoice") {
		t.Errorf("rule re-enabled after = YES")
	}
}

func TestSuppressDisableLineAndNextLine(t *testing.T) {
	text := "bad here <!-- vale disable-line STE.Vocabulary -->\n<!-- vale disable-next-line -->\ntarget line"
	s := ParseSuppressions(text)
	if !s.Suppressed(1, "STE.Vocabulary") {
		t.Errorf("disable-line should suppress its own line for the named rule")
	}
	if s.Suppressed(1, "STE.PassiveVoice") {
		t.Errorf("disable-line with an id should not suppress other rules")
	}
	if !s.Suppressed(3, "STE.PassiveVoice") || !s.Suppressed(3, "STE.Anything") {
		t.Errorf("disable-next-line without ids should suppress all rules on the next line")
	}
}

func TestFilterSuppressed(t *testing.T) {
	findings := []Finding{
		{RuleID: "STE.Contractions", Line: 1},
		{RuleID: "STE.Contractions", Line: 2},
	}
	text := "keep this\n<!-- vale disable-line STE.Contractions -->"
	got := FilterSuppressed(findings, text)
	if len(got) != 1 || got[0].Line != 1 {
		t.Errorf("FilterSuppressed = %+v, want only the line-1 finding", got)
	}
}

func TestSuppressNoDirectivesIsNoop(t *testing.T) {
	if (&Suppressions{}).Suppressed(1, "X") {
		t.Errorf("empty suppressions must suppress nothing")
	}
	if ParseSuppressions("plain text\nno directives").Suppressed(1, "X") {
		t.Errorf("text without directives must suppress nothing")
	}
}
