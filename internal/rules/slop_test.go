package rules

import (
	"testing"

	"github.com/stuffbucket/vale/internal/lint"
)

func firstMatchText(fs []lint.Finding) string {
	if len(fs) == 0 {
		return ""
	}
	return fs[0].Match
}

func TestSlopRestatement(t *testing.T) {
	got := NewSlopRestatementRule().Check(parsePlain("The valve opens. In other words, flow starts."))
	if len(got) != 1 || got[0].Match != "in other words" {
		t.Fatalf("restatement = %+v", got)
	}
	if n := NewSlopRestatementRule().Check(parsePlain("Open the valve.")); len(n) != 0 {
		t.Errorf("no restatement should be flagged: %+v", n)
	}
}

func TestSlopImpersonalHedge(t *testing.T) {
	got := NewSlopImpersonalHedgeRule().Check(parsePlain("It could be argued that the fan is loud."))
	if len(got) != 1 {
		t.Fatalf("impersonal hedge = %+v", got)
	}
	if got[0].Match != "It could be argued that" {
		t.Errorf("match = %q", got[0].Match)
	}
	if NewSlopImpersonalHedgeRule().DefaultSeverity() != lint.SeverityWarning {
		t.Error("want warning severity")
	}
}

func TestSlopNegativeParallelism(t *testing.T) {
	got := NewSlopNegativeParallelismRule().Check(parsePlain("It is not only fast but also cheap."))
	if len(got) != 1 {
		t.Fatalf("negative parallelism = %+v", got)
	}
	if n := NewSlopNegativeParallelismRule().Check(parsePlain("It is fast and cheap.")); len(n) != 0 {
		t.Errorf("plain sentence flagged: %+v", n)
	}
}

func TestSlopEvaluativeDensity(t *testing.T) {
	// Two praise adjectives in one sentence reach the threshold.
	got := NewSlopEvaluativeRule().Check(parsePlain("The comprehensive and robust design works."))
	if len(got) != 1 {
		t.Fatalf("evaluative density = %+v", got)
	}
	// A single praise adjective does not.
	if n := NewSlopEvaluativeRule().Check(parsePlain("The robust design works.")); len(n) != 0 {
		t.Errorf("single adjective flagged: %+v", n)
	}
}

func TestSlopHedgeDensity(t *testing.T) {
	// Three hedges cluster; flagged.
	got := NewSlopHedgeDensityRule().Check(parsePlain("It may possibly perhaps work."))
	if len(got) != 1 {
		t.Fatalf("hedge density = %+v", got)
	}
	// One hedge is fine.
	if n := NewSlopHedgeDensityRule().Check(parsePlain("It may work.")); len(n) != 0 {
		t.Errorf("single hedge flagged: %+v", n)
	}
}

func TestSlopNominalizationDensity(t *testing.T) {
	got := NewSlopNominalizationRule().Check(parsePlain("The implementation of the configuration needs consideration."))
	if len(got) != 1 {
		t.Fatalf("nominalization density = %+v", got)
	}
	if n := NewSlopNominalizationRule().Check(parsePlain("Configure the system.")); len(n) != 0 {
		t.Errorf("plain verbs flagged: %+v", n)
	}
	_ = firstMatchText(got)
}
