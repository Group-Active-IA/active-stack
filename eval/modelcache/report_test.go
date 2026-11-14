package modelcache

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

func paidAttempt(id string, value float64) Attempt {
	return Attempt{SourceID: id, Money: &MeasureInput{Aggregate: number(value), Scope: usage.ScopeIncludesChildren}}
}

func TestAnalyzeCountsFailuresRetriesAndPreparationOnce(t *testing.T) {
	p := syntheticProtocol(t)
	d := Dataset{Protocol: p, Preparation: []Preparation{{ID: "prep", ArmID: "A0", CohortID: "c", BlockID: "block", MoneyBasis: "observed", MoneyUnit: "synthetic-units", Cost: paidAttempt("prep-source", 10)}}, Runs: []Observation{
		{ID: "failed", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 1, BlockID: "block", Status: manifest.RunClientFailure, MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("failed-1", 20), paidAttempt("retry-2", 30)}},
		{ID: "accepted", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 2, BlockID: "block", Status: manifest.RunAccepted, Quality: []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}}, MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("accepted-1", 50)}},
	}}
	r, err := Analyze(d)
	if err != nil || len(r.Rows) != 1 {
		t.Fatalf("analysis: %+v %v", r, err)
	}
	row := r.Rows[0]
	if row.Accepted != 1 || row.Attempts != 3 || row.Preparations != 1 || row.Money.Total.Value == nil || *row.Money.Total.Value != 110 || row.Money.PerAccepted.Value == nil || *row.Money.PerAccepted.Value != 110 {
		t.Fatalf("incorrect denominator or warmup charge: %+v", row)
	}
	if r.Adoption.Status != "unavailable" {
		t.Fatal("offline input authorized adoption")
	}
}

func TestAnalyzeRejectsDuplicateTrialAcrossMeasurementBases(t *testing.T) {
	d := pairedDataset(t)
	copy := d.Runs[0]
	copy.ID = "duplicate-trial"
	copy.MoneyBasis = "estimated"
	copy.Attempts = []Attempt{paidAttempt("different-event", 10)}
	d.Runs = append(d.Runs, copy)
	if _, err := Analyze(d); err == nil {
		t.Fatal("measurement-basis split hid a repeated experimental trial")
	}
}

func TestAnalyzeKeepsDeclaredSourceRevisionsAndDatasetIdentity(t *testing.T) {
	d := pairedDataset(t)
	r, err := Analyze(d)
	if err != nil || r.Source.BaselineRevision != d.Protocol.BaselineRevision || r.Source.DelegationRevision != d.Protocol.DelegationRevision || r.Source.FixtureRevision != d.Protocol.Baseline.FixtureRevision || r.Source.RubricVersion != d.Protocol.Baseline.RubricVersion || len(r.Source.Layouts) != 2 || !validFingerprint(r.Source.DatasetFingerprint) {
		t.Fatalf("source trace lost: %+v %v", r, err)
	}
	d.Runs[0].Attempts[0].Money.Aggregate = number(999)
	next, err := Analyze(d)
	if err != nil || next.Source.DatasetFingerprint == r.Source.DatasetFingerprint {
		t.Fatal("different offline data retained the same identity")
	}
}
