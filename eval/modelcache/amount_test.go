package modelcache

import (
	"math"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

func number(v float64) *float64 { return &v }

func TestReconcileMeasureDoesNotSumInclusiveParentAndChildren(t *testing.T) {
	got, err := ReconcileMeasure(MeasureInput{Aggregate: number(1), Scope: usage.ScopeIncludesChildren, Expected: []string{"worker-1"}, Components: []Component{{ID: "worker-1", Role: "worker", Value: number(.4)}}})
	if err != nil || got.Total.Value == nil || *got.Total.Value != 1 || !got.Total.Complete {
		t.Fatalf("inclusive aggregate doubled: %+v %v", got, err)
	}
}

func TestReconcileMeasureParentOnlyAndUnknownScope(t *testing.T) {
	for _, tc := range []struct {
		name     string
		scope    usage.Scope
		want     float64
		complete bool
	}{
		{"parent-only", usage.ScopeParentOnly, 1.4, true},
		{"unknown", usage.ScopeUnknown, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReconcileMeasure(MeasureInput{Aggregate: number(1), Scope: tc.scope, Expected: []string{"worker-1"}, Components: []Component{{ID: "worker-1", Role: "worker", Value: number(.4)}}})
			if err != nil || got.Total.Value == nil || *got.Total.Value != tc.want || got.Total.Complete != tc.complete {
				t.Fatalf("unexpected reconciliation %+v %v", got, err)
			}
		})
	}
}

func TestReconcileMeasureMissingDuplicateAndInvalidValues(t *testing.T) {
	missing, err := ReconcileMeasure(MeasureInput{Expected: []string{"parent", "worker"}, Components: []Component{{ID: "parent", Role: "parent", Value: number(0)}}})
	if err != nil || missing.Total.Value == nil || *missing.Total.Value != 0 || missing.Total.Complete {
		t.Fatalf("missing is not observed zero: %+v %v", missing, err)
	}
	for _, in := range []MeasureInput{
		{Aggregate: number(-1), Scope: usage.ScopeIncludesChildren},
		{Aggregate: number(math.NaN()), Scope: usage.ScopeIncludesChildren},
		{Expected: []string{"x"}, Components: []Component{{ID: "x", Value: number(1)}, {ID: "x", Value: number(1)}}},
	} {
		if _, err := ReconcileMeasure(in); err == nil {
			t.Fatal("invalid or duplicated measure accepted")
		}
	}
}

func TestReconcileTokensUsesBaselineWithoutDoublingAggregate(t *testing.T) {
	var total int64 = 100
	got, err := ReconcileTokens(TokenInput{Aggregate: &total, Scope: usage.ScopeIncludesChildren, Expected: []string{"parent", "worker"}, Entries: []usage.Entry{{ID: "parent", Role: "parent", Tokens: 60}, {ID: "worker", Role: "worker", Tokens: 40}}})
	if err != nil || got.Value == nil || *got.Value != 100 || !got.Complete {
		t.Fatalf("token aggregate doubled: %+v %v", got, err)
	}
	got, err = ReconcileTokens(TokenInput{Aggregate: &total, Scope: usage.ScopeUnknown, Expected: []string{"parent", "worker"}, Entries: []usage.Entry{{ID: "parent", Role: "parent", Tokens: 60}, {ID: "worker", Role: "worker", Tokens: 40}}})
	if err != nil || got.Complete {
		t.Fatalf("unknown scope called complete: %+v %v", got, err)
	}
}

func TestReconcileTokensParentOnlyIncludesTheParent(t *testing.T) {
	var total int64 = 100
	for _, tc := range []struct {
		name     string
		entries  []usage.Entry
		expected []string
		want     float64
	}{
		{"parent-without-workers", nil, nil, 100},
		{"parent-and-worker", []usage.Entry{{ID: "w", Role: "worker", Tokens: 40}}, []string{"w"}, 140},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReconcileTokens(TokenInput{Aggregate: &total, Scope: usage.ScopeParentOnly, Expected: tc.expected, Entries: tc.entries})
			if err != nil || got.Value == nil || *got.Value != tc.want || !got.Complete {
				t.Fatalf("parent amount lost: %+v %v", got, err)
			}
		})
	}
}

func TestReconcileTokensRejectsOverflowAndScopeWithoutAggregate(t *testing.T) {
	for _, in := range []TokenInput{
		{Expected: []string{"a", "b"}, Entries: []usage.Entry{{ID: "a", Role: "parent", Tokens: math.MaxInt64}, {ID: "b", Role: "worker", Tokens: math.MaxInt64}}},
		{Scope: usage.ScopeIncludesChildren},
		{Scope: usage.ScopeParentOnly},
	} {
		if got, err := ReconcileTokens(in); err == nil {
			t.Fatalf("invalid input accepted: %+v", got)
		}
	}
}
