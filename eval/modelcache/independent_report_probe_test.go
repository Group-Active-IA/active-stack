package modelcache

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/timing"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

type independentManualClock struct{ now time.Duration }

func (c *independentManualClock) Now() time.Duration { return c.now }

func TestIndependentReportZeroAcceptedRetainsKnownTokensTimeAndPartialMoney(t *testing.T) {
	for _, tc := range []struct {
		name       string
		money      *MeasureInput
		known      *float64
		isComplete bool
	}{
		{"unknown", &MeasureInput{Scope: usage.ScopeIncludesChildren}, nil, false},
		{"observed-zero", &MeasureInput{Aggregate: number(0), Scope: usage.ScopeIncludesChildren}, number(0), true},
		{"known-partial", &MeasureInput{Scope: usage.ScopeParentOnly, Expected: []string{"worker"}, Components: []Component{{ID: "worker", Role: "worker", Value: number(30)}}}, number(30), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Synthetic elapsed time goes through 02's public monotonic recorder;
			// unavailable human/approval waits are not filled with invented zeros.
			clock := &independentManualClock{}
			recorder := timing.NewRecorder(clock, false)
			if err := recorder.Start(); err != nil {
				t.Fatal(err)
			}
			clock.now = 100 * time.Millisecond
			measured, err := recorder.Stop()
			if err != nil || measured.Adjusted.Available {
				t.Fatalf("synthetic 02 timing measurement: %+v %v", measured, err)
			}
			elapsed := float64(measured.Elapsed.Milliseconds())
			d := Dataset{Protocol: syntheticProtocol(t), Preparation: []Preparation{}, Runs: []Observation{
				{ID: "failed-run", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 1, Status: manifest.RunClientFailure,
					MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{{SourceID: "failed-event", Money: tc.money,
						Tokens: &TokenInput{Aggregate: integer(100), Scope: usage.ScopeIncludesChildren}, ElapsedMS: &elapsed}}},
			}}
			report, err := Analyze(d)
			if err != nil || len(report.Rows) != 1 {
				t.Fatalf("missing monetary signal discarded a started run: %+v %v", report, err)
			}
			row := report.Rows[0]
			if row.Started != 1 || row.Attempts != 1 || row.Accepted != 0 || row.Money.PerAccepted.Value != nil || row.Tokens.PerAccepted.Value != nil {
				t.Fatalf("zero accepted denominator fabricated a ratio: %+v", row)
			}
			if row.Tokens.Total.Value == nil || *row.Tokens.Total.Value != 100 || !row.Tokens.Total.Complete || row.SummedElapsedMS.Value == nil || *row.SummedElapsedMS.Value != 100 || row.SummedHumanMS.Value != nil {
				t.Fatalf("known tokens/time lost or unknown human time invented: %+v", row)
			}
			if row.Money.Tasks.Complete != tc.isComplete || (row.Money.Tasks.Value == nil) != (tc.known == nil) {
				t.Fatalf("unknown, partial and observed zero were conflated: %+v", row.Money.Tasks)
			}
			if tc.known != nil && *row.Money.Tasks.Value != *tc.known {
				t.Fatalf("known monetary amount lost: %+v", row.Money.Tasks)
			}
		})
	}
}

func TestIndependentReportSubscriptionQuotaDoesNotInventMonetaryCost(t *testing.T) {
	p := syntheticProtocol(t)
	p.Baseline.Cohorts[0].Cell.Modality = manifest.ModalitySubscription
	d := Dataset{Protocol: p, Preparation: []Preparation{}, Runs: []Observation{
		{ID: "subscription-run", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 1, Status: manifest.RunAccepted,
			Quality: []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}}, QuotaUnit: "synthetic-quota-unit",
			Attempts: []Attempt{{SourceID: "subscription-event", Quota: &MeasureInput{Aggregate: number(10), Scope: usage.ScopeIncludesChildren},
				Tokens: &TokenInput{Aggregate: integer(100), Scope: usage.ScopeIncludesChildren}}}},
	}}
	report, err := Analyze(d)
	if err != nil || len(report.Rows) != 1 {
		t.Fatalf("subscription analysis: %+v %v", report, err)
	}
	row := report.Rows[0]
	if row.Modality != "subscription" || row.Quota.PerAccepted.Value == nil || *row.Quota.PerAccepted.Value != 10 || !row.Quota.PerAccepted.Complete {
		t.Fatalf("observed quota unavailable: %+v", row)
	}
	if row.Money.Tasks.Value != nil || row.Money.Total.Complete || row.Money.PerAccepted.Value != nil || row.MoneyBasis != "" {
		t.Fatalf("subscription tokens/quota became invented monetary cost: %+v", row.Money)
	}
}

// An unavailable failed-attempt amount must not disappear from the coverage
// denominator just because its monetary basis cannot be identified.
func TestIndependentReportMissingFailureBasisCannotMakeSuccessSubsetComplete(t *testing.T) {
	p := syntheticProtocol(t)
	d := Dataset{Protocol: p, Preparation: []Preparation{}, Runs: []Observation{
		{ID: "missing-failed-run", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 1, Status: manifest.RunClientFailure,
			Attempts: []Attempt{{SourceID: "missing-failed-attempt"}}},
		{ID: "observed-accepted-run", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 2, Status: manifest.RunAccepted,
			Quality:    []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}},
			MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("observed-accepted-attempt", 50)}},
	}}
	report, err := Analyze(d)
	if err != nil {
		t.Fatalf("missing observable signal must retain the started run: %v", err)
	}
	started, attempts := 0, 0
	known := float64(0)
	observedFound := false
	for _, row := range report.Rows {
		started += row.Started
		attempts += row.Attempts
		if row.Money.Tasks.Value != nil {
			known += *row.Money.Tasks.Value
		}
		if row.MoneyBasis == "observed" {
			observedFound = true
			if row.Money.PerAccepted.Value != nil || row.Money.PerAccepted.Complete {
				t.Fatalf("unpriced failed run hidden by successful observed subset: %+v", row.Money)
			}
		}
	}
	if !observedFound || started != 2 || attempts != 2 || known != 50 {
		t.Fatalf("started attempts or known cost were discarded: rows=%+v", report.Rows)
	}
}

func independentPrefixDataset(t *testing.T, expectedControl, expectedCandidate string, readControl, readCandidate int64) Dataset {
	t.Helper()
	p := syntheticProtocol(t)
	files := fstest.MapFS{"one": {Data: []byte("synthetic one")}, "two": {Data: []byte("synthetic two")}}
	c := p.Arms[0]
	c.ID, c.Pilot, c.CachePolicy, c.SemanticReviewRef = "B0", "B", "synthetic-fixed-cache-policy", "synthetic-semantic-review"
	var err error
	c.Layout, err = FreezeLayout(files, "materialized-prompt", []string{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}
	a := candidate(c, "prefix")
	a.ID = "B1"
	a.Layout, err = FreezeLayout(files, "materialized-prompt", []string{"two", "one"})
	if err != nil {
		t.Fatal(err)
	}
	p.Arms = []Arm{c, a}
	d := Dataset{Protocol: p, Preparation: []Preparation{}}
	for i, arm := range p.Arms {
		expected, read := expectedControl, readControl
		if i == 1 {
			expected, read = expectedCandidate, readCandidate
		}
		d.Runs = append(d.Runs, Observation{ID: "run-" + arm.ID, ArmID: arm.ID, CohortID: "c", TaskID: "task", Repetition: 1,
			PairID: "prefix-pair", BlockID: "prefix-block", FixtureRevision: p.Baseline.FixtureRevision, RubricVersion: p.Baseline.RubricVersion,
			Cell: p.Baseline.Cohorts[0].Cell, Effective: arm.Effective, Status: manifest.RunAccepted,
			Quality:    []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}},
			MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("attempt-"+arm.ID, 50)},
			Cache: CacheObservation{Expected: expected, Attribution: "prefix", PrefixFingerprint: arm.Layout.Fingerprint,
				ReadTokens: integer(read), EligibleTokens: integer(100), SignalRef: "synthetic-signal-" + arm.ID, IsolationRef: "synthetic-isolation-and-window"}})
	}
	return d
}

func TestIndependentReportPrefixContrastCannotCompareColdWithWarm(t *testing.T) {
	report, err := Analyze(independentPrefixDataset(t, "cold", "warm", 0, 20))
	if err != nil {
		return // Contradictory assigned conditions may be rejected as invalid data.
	}
	if len(report.Pairs) != 1 || report.Pairs[0].Comparison.Comparable {
		t.Fatalf("prefix comparison silently changed assigned cache condition: %+v", report.Pairs)
	}
}

func TestIndependentReportWarmMissAndPartialRemainAssignedResults(t *testing.T) {
	report, err := Analyze(independentPrefixDataset(t, "warm", "warm", 0, 20))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Pairs) != 1 || !report.Pairs[0].Comparison.Comparable {
		t.Fatalf("genuine miss/partial was excluded to improve cache performance: %+v", report.Pairs)
	}
	observed := map[string]int{}
	for _, row := range report.Rows {
		for state, count := range row.ObservedCache {
			observed[state] += count
		}
	}
	if observed["miss"] != 1 || observed["partial_read"] != 1 {
		t.Fatalf("assigned outcomes were rewritten: %+v", observed)
	}
}

func TestIndependentReportWarmMissAndPartialPermitOnlyConditionalArithmetic(t *testing.T) {
	d := independentPrefixDataset(t, "warm", "warm", 0, 20)
	n := 1
	d.Protocol.Decisions = &Decisions{Metric: "money", MinimumPairs: &n, MinImprovementFraction: number(.1), QualityMargin: number(0)}
	d.Runs[1].Attempts = []Attempt{paidAttempt("attempt-B1", 25)}
	report, err := Analyze(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Contrasts) != 1 || report.Contrasts[0].Status != "candidate_for_review" || report.Contrasts[0].ImprovementFraction == nil || *report.Contrasts[0].ImprovementFraction != .5 {
		t.Fatalf("genuine assigned miss/partial was dropped from conditional arithmetic: %+v", report.Contrasts)
	}
	if report.Adoption.Status != "unavailable" {
		t.Fatal("synthetic declared criteria or cache observations authorized adoption")
	}
	observed := map[string]int{}
	for _, row := range report.Rows {
		for state, count := range row.ObservedCache {
			observed[state] += count
		}
	}
	if observed["miss"] != 1 || observed["partial_read"] != 1 {
		t.Fatalf("conditional arithmetic rewrote genuine cache outcomes: %+v", observed)
	}
}

func TestIndependentReportRetainsCacheAssignmentAndDiagnosticFlags(t *testing.T) {
	d := independentPrefixDataset(t, "cold", "warm", 100, 0)
	d.Runs[0].Cache.IsolationRef = ""
	unknown := d.Runs[1]
	unknown.ID, unknown.Repetition, unknown.PairID = "unknown-cache-run", 2, ""
	unknown.Attempts = []Attempt{paidAttempt("unknown-cache-attempt", 25)}
	unknown.Cache.Attribution = "aggregate"
	d.Runs = append(d.Runs, unknown)
	report, err := Analyze(d)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	find := func(expected, observed, flag, reason string) bool {
		found := false
		var inspect func(any)
		inspect = func(v any) {
			switch v := v.(type) {
			case map[string]any:
				if v["expected"] == expected && v["observed"] == observed && (flag == "" || v[flag] == true) {
					r, _ := v["reason"].(string)
					found = found || strings.Contains(r, reason)
				}
				for _, item := range v {
					inspect(item)
				}
			case []any:
				for _, item := range v {
					inspect(item)
				}
			}
		}
		inspect(value)
		return found
	}
	if !find("cold", "full_read", "cold_contaminated", "cold isolation unavailable") || !find("warm", "miss", "warm_miss", "") || !find("warm", "unknown", "", "no attributable signal") {
		t.Fatalf("cache assignment, contamination/miss flags or unknown reason were discarded: %s", raw)
	}
}

func TestIndependentReportAmbiguousPreparationCannotBeHiddenOrChargedTwice(t *testing.T) {
	d := independentPrefixDataset(t, "cold", "warm", 0, 20)
	cold := d.Runs[0]
	warm := cold
	warm.ID, warm.Repetition, warm.PairID = "warm-same-arm-run", 2, ""
	warm.Cache.Expected, warm.Cache.ReadTokens = "warm", integer(20)
	warm.Attempts = []Attempt{paidAttempt("warm-same-arm-attempt", 50)}
	d.Runs = []Observation{cold, warm}
	d.Preparation = []Preparation{{ID: "shared-block-preparation", ArmID: cold.ArmID, CohortID: cold.CohortID, BlockID: cold.BlockID,
		MoneyBasis: "observed", MoneyUnit: "synthetic-units", Cost: paidAttempt("shared-preparation-source", 10)}}
	report, err := Analyze(d)
	if err != nil {
		return // An ambiguous preparation assignment may instead reject the input.
	}
	preparations, conditionRows := 0, 0
	var preparationCost, taskCost float64
	for _, row := range report.Rows {
		preparations += row.Preparations
		if row.Money.Preparation.Value != nil {
			preparationCost += *row.Money.Preparation.Value
		}
		if row.Money.Tasks.Value != nil {
			taskCost += *row.Money.Tasks.Value
		}
		if row.Started > 0 {
			conditionRows++
			if row.Money.PerAccepted.Value != nil || row.Money.PerAccepted.Complete {
				t.Fatalf("ambiguous shared preparation hidden from a condition ratio: %+v", row)
			}
		}
	}
	if conditionRows != 2 || preparations != 1 || preparationCost != 10 || taskCost != 100 {
		t.Fatalf("conditions mixed, preparation duplicated or input cost lost: %+v", report.Rows)
	}
}

func TestIndependentReportPreparationCoverageCannotHideUnknownCharge(t *testing.T) {
	for _, tc := range []struct {
		name     string
		basis    string
		value    *float64
		complete bool
	}{
		{"unknown-basis-and-amount", "", nil, false},
		{"known-basis-missing-amount", "observed", nil, false},
		{"known-observed-zero", "observed", number(0), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cost := Attempt{SourceID: "preparation-source"}
			unit := ""
			if tc.basis != "" {
				unit = "synthetic-units"
			}
			if tc.value != nil {
				cost = paidAttempt("preparation-source", *tc.value)
			}
			d := Dataset{Protocol: syntheticProtocol(t), Preparation: []Preparation{{ID: "prep", ArmID: "A0", CohortID: "c", BlockID: "block", MoneyBasis: tc.basis, MoneyUnit: unit, Cost: cost}},
				Runs: []Observation{{ID: "accepted-run", ArmID: "A0", CohortID: "c", TaskID: "task", Repetition: 1, BlockID: "block", Status: manifest.RunAccepted,
					Quality:    []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}},
					MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("accepted-attempt", 50)}}}}
			report, err := Analyze(d)
			if err != nil {
				t.Fatalf("unavailable preparation signal must remain reportable: %v", err)
			}
			accepted, preparations := 0, 0
			found := false
			for _, row := range report.Rows {
				accepted += row.Accepted
				preparations += row.Preparations
				if row.MoneyBasis != "observed" {
					continue
				}
				found = true
				m := row.Money.PerAccepted
				if tc.complete {
					if m.Value == nil || *m.Value != 50 || !m.Complete {
						t.Fatalf("observed preparation zero lost: %+v", m)
					}
				} else if m.Value != nil || m.Complete {
					t.Fatalf("unknown preparation charge hidden: %+v", m)
				}
			}
			if !found || accepted != 1 || preparations != 1 {
				t.Fatalf("accepted run or declared preparation discarded: %+v", report.Rows)
			}
		})
	}
}

// Both sides agreeing with one another is insufficient when both contradict
// the frozen cohort's client, version, permissions or tools.
func TestIndependentReportPairCellsMustMatchFrozenCohort(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*manifest.Cell)
	}{
		{"client", func(c *manifest.Cell) { c.Client = manifest.ClientClaude }},
		{"version", func(c *manifest.Cell) { c.Version = "2.0.0" }},
		{"permissions", func(c *manifest.Cell) {
			c.Permissions.Mode, c.Observed.Permissions.Mode = "other-synthetic-mode", "other-synthetic-mode"
		}},
		{"tools", func(c *manifest.Cell) { c.Tools = []string{"other-synthetic-tool"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := syntheticProtocol(t)
			control := p.Arms[0]
			variant := candidate(control, "model")
			variant.Requested.Models["worker"] = "synthetic-model-variant"
			variant.Effective.Models["worker"] = "synthetic-model-variant"
			p.Arms = append(p.Arms, variant)
			cell := p.Baseline.Cohorts[0].Cell
			tc.mutate(&cell)
			runs := []Observation{}
			for _, arm := range p.Arms {
				runs = append(runs, Observation{ID: "run-" + arm.ID, ArmID: arm.ID, CohortID: "c", TaskID: "task", Repetition: 1,
					PairID: "same-pair", BlockID: "same-block", FixtureRevision: p.Baseline.FixtureRevision, RubricVersion: p.Baseline.RubricVersion,
					Cell: cell, Effective: arm.Effective, Status: manifest.RunAccepted,
					Quality:    []fixture.CriterionResult{{ID: "required", Mandatory: true, Outcome: fixture.Met}},
					MoneyBasis: "observed", MoneyUnit: "synthetic-units", Attempts: []Attempt{paidAttempt("attempt-"+arm.ID, 50)}})
			}
			report, err := Analyze(Dataset{Protocol: p, Preparation: []Preparation{}, Runs: runs})
			if err != nil {
				return // Contradictory metadata may be rejected as invalid input.
			}
			if len(report.Pairs) != 1 {
				t.Fatalf("pair did not retain a diagnostic: %+v", report.Pairs)
			}
			if report.Pairs[0].Comparison.Comparable {
				t.Fatalf("both run cells contradict the frozen cohort but remain comparable: %+v", report.Pairs[0])
			}
		})
	}
}
