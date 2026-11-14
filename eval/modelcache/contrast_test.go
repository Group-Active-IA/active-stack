package modelcache

import (
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

func pairedDataset(t *testing.T) Dataset {
	p := syntheticProtocol(t)
	c := p.Arms[0]
	c.Layout, _ = FreezeLayout(fstest.MapFS{"p": {Data: []byte("synthetic prompt")}}, "materialized-prompt", []string{"p"})
	a := candidate(c, "model")
	a.Requested.Models["worker"] = "synthetic-other"
	a.Effective.Models["worker"] = "synthetic-other"
	p.Arms = []Arm{c, a}
	d := Dataset{Protocol: p, Preparation: []Preparation{}}
	for i, arm := range p.Arms {
		d.Runs = append(d.Runs, Observation{ID: "run-" + arm.ID, ArmID: arm.ID, CohortID: "c", TaskID: "task", Repetition: 1, PairID: "pair", BlockID: "block", FixtureRevision: p.Baseline.FixtureRevision, RubricVersion: p.Baseline.RubricVersion, Cell: p.Baseline.Cohorts[0].Cell, Effective: arm.Effective, Status: manifest.RunAccepted, Quality: []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}}, MoneyBasis: "observed", MoneyUnit: "synthetic-unit", Attempts: []Attempt{paidAttempt("event-"+arm.ID, float64(100-i*50))}, Cache: CacheObservation{Expected: "cold", Attribution: "prefix", PrefixFingerprint: arm.Layout.Fingerprint, ReadTokens: integer(0), EligibleTokens: integer(100), SignalRef: "synthetic-signal", IsolationRef: "synthetic-isolation"}})
	}
	return d
}

func TestContrastIsInconclusiveUntilObjectiveAndThresholdsAreExplicit(t *testing.T) {
	d := pairedDataset(t)
	r, err := Analyze(d)
	if err != nil || len(r.Contrasts) != 1 || r.Contrasts[0].Status != "inconclusive" {
		t.Fatalf("TBD criteria became defaults: %+v %v", r, err)
	}
}

func TestContrastComputesConditionalImprovementButNeverAuthorizesAdoption(t *testing.T) {
	d := pairedDataset(t)
	n := 1
	d.Protocol.Decisions = &Decisions{Metric: "money", MinimumPairs: &n, MinImprovementFraction: number(.1), QualityMargin: number(0)}
	r, err := Analyze(d)
	if err != nil || len(r.Contrasts) != 1 || r.Contrasts[0].Status != "candidate_for_review" || r.Contrasts[0].ImprovementFraction == nil || *r.Contrasts[0].ImprovementFraction != .5 {
		t.Fatalf("conditional contrast: %+v %v", r, err)
	}
	if r.Adoption.Status != "unavailable" {
		t.Fatal("declared input enabled adoption")
	}
	d.Runs[1].Quality[0].Outcome = fixture.NotMet
	r, err = Analyze(d)
	if err != nil || r.Contrasts[0].Status != "rejected_quality" {
		t.Fatalf("mandatory regression permitted: %+v %v", r, err)
	}
}

func TestContrastRetainsCostsButCannotAttributeUnconfirmedCacheConditions(t *testing.T) {
	for _, condition := range []string{"contaminated", "unknown-prefix", "missing-isolation"} {
		t.Run(condition, func(t *testing.T) {
			d := pairedDataset(t)
			n := 1
			d.Protocol.Decisions = &Decisions{Metric: "money", MinimumPairs: &n, MinImprovementFraction: number(.1), QualityMargin: number(0)}
			switch condition {
			case "contaminated":
				d.Runs[1].Cache.ReadTokens = integer(100)
			case "unknown-prefix":
				d.Runs[1].Cache.Attribution = "global"
			case "missing-isolation":
				d.Runs[1].Cache.IsolationRef = ""
			}
			r, err := Analyze(d)
			if err != nil || len(r.Contrasts) != 1 || r.Contrasts[0].Status != "inconclusive" || r.Contrasts[0].ImprovementFraction != nil {
				t.Fatalf("unconfirmed cache allowed attribution: %+v %v", r, err)
			}
			if len(r.Rows) != 2 || r.Rows[0].Money.Total.Value == nil {
				t.Fatal("lost incurred cost")
			}
		})
	}
}
