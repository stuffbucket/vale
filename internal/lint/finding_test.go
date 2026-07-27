package lint

import "testing"

func TestSortFindings(t *testing.T) {
	in := []Finding{
		{RuleID: "B", Line: 2, Col: 1},
		{RuleID: "A", Line: 1, Col: 5},
		{RuleID: "Z", Line: 1, Col: 2},
		{RuleID: "A", Line: 1, Col: 2}, // same line/col as Z, earlier RuleID
		{RuleID: "C", Line: 1, Col: 5},
	}
	SortFindings(in)
	want := []struct {
		RuleID string
		Line   int
		Col    int
	}{
		{"A", 1, 2},
		{"Z", 1, 2},
		{"A", 1, 5},
		{"C", 1, 5},
		{"B", 2, 1},
	}
	if len(in) != len(want) {
		t.Fatalf("len = %d, want %d", len(in), len(want))
	}
	for i, w := range want {
		if in[i].RuleID != w.RuleID || in[i].Line != w.Line || in[i].Col != w.Col {
			t.Errorf("index %d = {%s %d %d}, want {%s %d %d}",
				i, in[i].RuleID, in[i].Line, in[i].Col, w.RuleID, w.Line, w.Col)
		}
	}
}

func TestSortFindingsStableForEqualKeys(t *testing.T) {
	// Two findings with identical line, col and ruleID keep input order.
	in := []Finding{
		{RuleID: "A", Line: 1, Col: 1, Message: "first"},
		{RuleID: "A", Line: 1, Col: 1, Message: "second"},
	}
	SortFindings(in)
	if in[0].Message != "first" || in[1].Message != "second" {
		t.Errorf("stable sort broken: got %q then %q", in[0].Message, in[1].Message)
	}
}

func TestSortFindingsEmpty(t *testing.T) {
	var in []Finding
	SortFindings(in) // must not panic
	if len(in) != 0 {
		t.Errorf("len = %d, want 0", len(in))
	}
}
