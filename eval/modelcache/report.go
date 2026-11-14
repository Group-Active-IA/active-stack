package modelcache

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/record"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/summary"
)

type Attempt struct {
	SourceID     string        `json:"source_id"`
	Tokens       *TokenInput   `json:"tokens"`
	Money        *MeasureInput `json:"money"`
	Quota        *MeasureInput `json:"quota"`
	LogicalInput *MeasureInput `json:"logical_input"`
	CacheRead    *MeasureInput `json:"cache_read"`
	CacheWrite   *MeasureInput `json:"cache_write"`
	Output       *MeasureInput `json:"output"`
	ElapsedMS    *float64      `json:"elapsed_ms"`
	HumanMS      *float64      `json:"human_ms"`
}

type Observation struct {
	ID              string                    `json:"id"`
	ArmID           string                    `json:"arm_id"`
	CohortID        string                    `json:"cohort_id"`
	TaskID          string                    `json:"task_id"`
	Repetition      int                       `json:"repetition"`
	PairID          string                    `json:"pair_id"`
	BlockID         string                    `json:"block_id"`
	FixtureRevision string                    `json:"fixture_revision"`
	RubricVersion   string                    `json:"rubric_version"`
	Cell            manifest.Cell             `json:"cell"`
	Effective       Parameters                `json:"effective"`
	Status          manifest.RunStatus        `json:"status"`
	Quality         []fixture.CriterionResult `json:"quality"`
	MoneyBasis      string                    `json:"money_basis"`
	MoneyUnit       string                    `json:"money_unit"`
	QuotaUnit       string                    `json:"quota_unit"`
	Attempts        []Attempt                 `json:"attempts"`
	Cache           CacheObservation          `json:"cache"`
}

type Preparation struct {
	ExpectedCache string  `json:"expected_cache"`
	ID            string  `json:"id"`
	ArmID         string  `json:"arm_id"`
	CohortID      string  `json:"cohort_id"`
	BlockID       string  `json:"block_id"`
	MoneyBasis    string  `json:"money_basis"`
	MoneyUnit     string  `json:"money_unit"`
	QuotaUnit     string  `json:"quota_unit"`
	Cost          Attempt `json:"cost"`
}

type Dataset struct {
	Protocol Protocol      `json:"protocol"`
	Runs     []Observation `json:"runs"`
	// Nil means unavailable; [] declares no preparation events in this input.
	Preparation []Preparation `json:"preparation"`
}

type CostSummary struct {
	Tasks       Metric `json:"tasks"`
	Preparation Metric `json:"preparation"`
	Total       Metric `json:"total"`
	PerAccepted Metric `json:"per_accepted"`
}

type Row struct {
	ExpectedCache string      `json:"expected_cache"`
	Pilot         string      `json:"pilot"`
	ArmID         string      `json:"arm_id"`
	CohortID      string      `json:"cohort_id"`
	Client        string      `json:"client"`
	OS            string      `json:"os"`
	Interface     string      `json:"interface"`
	Modality      string      `json:"modality"`
	MoneyBasis    string      `json:"money_basis"`
	MoneyUnit     string      `json:"money_unit"`
	QuotaUnit     string      `json:"quota_unit"`
	Started       int         `json:"started"`
	Accepted      int         `json:"accepted"`
	Rejected      int         `json:"rejected"`
	Pending       int         `json:"pending"`
	Attempts      int         `json:"attempts"`
	Preparations  int         `json:"preparations"`
	Tokens        CostSummary `json:"tokens"`
	Money         CostSummary `json:"money"`
	Quota         CostSummary `json:"quota"`
	LogicalInput  Metric      `json:"logical_input"`
	CacheRead     Metric      `json:"cache_read"`
	CacheWrite    Metric      `json:"cache_write"`
	Output        Metric      `json:"output"`
	// These are sums of declared run/attempt durations, not wall-clock time of
	// concurrently running workers or an invented approval-adjusted duration.
	SummedElapsedMS     Metric         `json:"summed_elapsed_ms"`
	SummedHumanMS       Metric         `json:"summed_human_ms"`
	ObservedCache       map[string]int `json:"observed_cache"`
	MandatoryRegression bool           `json:"mandatory_regression"`
	FragmentedCoverage  bool           `json:"fragmented_coverage"`
}

type PairResult struct {
	ID           string     `json:"id"`
	ControlRun   string     `json:"control_run"`
	CandidateRun string     `json:"candidate_run"`
	Comparison   Comparison `json:"comparison"`
}
type AdoptionGate struct {
	Status  string   `json:"status"`
	Reasons []string `json:"reasons"`
}
type CacheAssessment struct {
	RunID  string      `json:"run_id"`
	Result CacheResult `json:"result"`
}

type ArmLayoutSource struct {
	ArmID  string `json:"arm_id"`
	Layout Layout `json:"layout"`
}
type SourceSummary struct {
	BaselineRevision        string             `json:"baseline_revision"`
	DelegationRevision      string             `json:"delegation_revision"`
	BaselineProtocolVersion string             `json:"baseline_protocol_version"`
	FixtureRevision         string             `json:"fixture_revision"`
	RubricVersion           string             `json:"rubric_version"`
	Layouts                 []ArmLayoutSource  `json:"layouts"`
	DatasetFingerprint      record.Fingerprint `json:"dataset_fingerprint"`
	FingerprintMethod       string             `json:"fingerprint_method"`
	SourceIDs               []string           `json:"source_ids"`
}

type Report struct {
	Source    SourceSummary     `json:"source"`
	Caches    []CacheAssessment `json:"caches"`
	Version   string            `json:"version"`
	Basis     string            `json:"basis"`
	Rows      []Row             `json:"rows"`
	Pairs     []PairResult      `json:"pairs"`
	Contrasts []Contrast        `json:"contrasts"`
	Adoption  AdoptionGate      `json:"adoption"`
	Decisions *Decisions        `json:"decisions"`
}

// ValidateLayouts checks the named bytes again. It still cannot establish that
// a client loaded them or that a source labelled materialized-prompt is authentic.
func ValidateLayouts(p Protocol, files fs.FS) error {
	for _, a := range p.Arms {
		var paths []string
		for _, f := range a.Layout.Files {
			paths = append(paths, f.Path)
		}
		actual, err := FreezeLayout(files, a.Layout.SourceScope, paths)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(actual, a.Layout) {
			return fmt.Errorf("layout file bytes changed for arm %s", a.ID)
		}
	}
	return nil
}

func fromAmount(a summary.Amount) Metric {
	if !a.Available {
		return Metric{Reason: a.Reason}
	}
	return metric(a.Value, !a.Partial, a.Reason)
}
func toAmount(m Metric) summary.Amount {
	if m.Value == nil {
		return summary.Amount{Reason: m.Reason}
	}
	return summary.Amount{Available: true, Value: *m.Value, Partial: !m.Complete, Reason: m.Reason}
}
func sumMetrics(ms []Metric, emptyKnown bool) Metric {
	if len(ms) == 0 {
		if emptyKnown {
			return metric(0, true, "")
		}
		return Metric{Reason: "metric unavailable"}
	}
	complete := true
	var vs []float64
	for _, m := range ms {
		if m.Value == nil {
			complete = false
			continue
		}
		vs = append(vs, *m.Value)
		complete = complete && m.Complete
	}
	if len(vs) == 0 {
		return Metric{Reason: "metric unavailable for all declared events"}
	}
	sort.Float64s(vs)
	var v float64
	for _, x := range vs {
		v += x
	}
	if !validNumber(v) {
		return Metric{Reason: "metric sum overflow"}
	}
	reason := ""
	if !complete {
		reason = "incomplete declared event coverage"
	}
	return metric(v, complete, reason)
}
func scalar(v *float64) (Metric, error) {
	if v == nil {
		return Metric{Reason: "duration unavailable"}, nil
	}
	if !validNumber(*v) {
		return Metric{}, fmt.Errorf("invalid duration")
	}
	return metric(*v, true, ""), nil
}

// Missing aggregate signals are retained as partial data at the dataset
// boundary. Low-level reconciliation remains strict about contradictory input.
func observedAmount(in *MeasureInput) (Metric, error) {
	if in == nil {
		return Metric{Reason: "measure unavailable"}, nil
	}
	copy := *in
	missing := copy.Aggregate == nil && copy.Scope != ""
	if missing {
		copy.Scope = ""
	}
	r, err := ReconcileMeasure(copy)
	if err != nil {
		return Metric{}, err
	}
	if missing {
		r.Total.Complete = false
		r.Total.Reason = "aggregate signal unavailable; components are partial"
	}
	return r.Total, nil
}
func observedTokens(in *TokenInput) (Metric, error) {
	if in == nil {
		return Metric{Reason: "tokens unavailable"}, nil
	}
	copy := *in
	missing := copy.Aggregate == nil && copy.Scope != ""
	if missing {
		copy.Scope = ""
	}
	m, err := ReconcileTokens(copy)
	if err != nil {
		return Metric{}, err
	}
	if missing {
		m.Complete = false
		m.Reason = "aggregate token signal unavailable"
	}
	return m, nil
}

type eventMetrics struct {
	cost    summary.Cost
	context [6]Metric
}

func evaluateAttempt(a Attempt) (eventMetrics, error) {
	var e eventMetrics
	t, err := observedTokens(a.Tokens)
	if err != nil {
		return e, err
	}
	m, err := observedAmount(a.Money)
	if err != nil {
		return e, err
	}
	q, err := observedAmount(a.Quota)
	if err != nil {
		return e, err
	}
	e.cost = summary.Cost{Tokens: toAmount(t), Money: toAmount(m), Quota: toAmount(q)}
	for i, v := range []*MeasureInput{a.LogicalInput, a.CacheRead, a.CacheWrite, a.Output} {
		e.context[i], err = observedAmount(v)
		if err != nil {
			return e, err
		}
	}
	e.context[4], err = scalar(a.ElapsedMS)
	if err != nil {
		return e, err
	}
	e.context[5], err = scalar(a.HumanMS)
	return e, err
}

type reportGroup struct {
	expected           string
	arm                Arm
	cohort             manifest.Cohort
	basis, unit, quota string
	runs               []summary.RunResult
	prep               []summary.Cost
	events             []eventMetrics
	cache              map[string]int
	regression         bool
}

// Analyze is a conditional calculation over offline inputs. No input flag,
// receipt hash or self-declared quality result can enable adoption here.
func Analyze(d Dataset) (Report, error) {
	if err := d.Protocol.Validate(); err != nil {
		return Report{}, err
	}
	p := d.Protocol
	plan, err := p.Baseline.Plan()
	if err != nil {
		return Report{}, err
	}
	arms := map[string]Arm{}
	for _, a := range p.Arms {
		arms[a.ID] = a
	}
	cohorts := map[string]manifest.Cohort{}
	for _, c := range p.Baseline.Cohorts {
		cohorts[c.ID] = c
	}
	groups := map[string]*reportGroup{}
	group := func(arm, cohort, expected, basis, unit, quota string) (*reportGroup, error) {
		a, ok := arms[arm]
		if !ok {
			return nil, fmt.Errorf("unknown arm %s", arm)
		}
		c, ok := cohorts[cohort]
		if !ok {
			return nil, fmt.Errorf("unknown cohort %s", cohort)
		}
		if basis != "" && basis != "observed" && basis != "estimated" {
			return nil, fmt.Errorf("money basis must be observed, estimated or unavailable")
		}
		key := strings.Join([]string{a.Pilot, arm, cohort, expected, basis, unit, quota}, "\x00")
		if groups[key] == nil {
			groups[key] = &reportGroup{arm: a, cohort: c, expected: expected, basis: basis, unit: unit, quota: quota, cache: map[string]int{}}
		}
		return groups[key], nil
	}
	sources := map[string]bool{}
	event := func(a Attempt, basis, unit, quota string) (eventMetrics, error) {
		if strings.TrimSpace(a.SourceID) == "" || sources[a.SourceID] {
			return eventMetrics{}, fmt.Errorf("missing or repeated charge/event source_id")
		}
		sources[a.SourceID] = true
		e, err := evaluateAttempt(a)
		if err != nil {
			return e, err
		}
		if e.cost.Money.Available && (basis == "" || unit == "") {
			return e, fmt.Errorf("available money requires explicit basis and unit")
		}
		if e.cost.Quota.Available && quota == "" {
			return e, fmt.Errorf("available quota requires an explicit unit")
		}
		return e, nil
	}
	var cacheAssessments []CacheAssessment
	seenRuns := map[string]bool{}
	seenTrials := map[string]bool{}
	for _, o := range d.Runs {
		if strings.TrimSpace(o.ID) == "" || seenRuns[o.ID] || o.Repetition <= 0 {
			return Report{}, fmt.Errorf("invalid or repeated run id/repetition")
		}
		seenRuns[o.ID] = true
		trial := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s", o.ArmID, o.CohortID, o.TaskID, o.Repetition, o.Cache.Expected)
		if seenTrials[trial] {
			return Report{}, fmt.Errorf("repeated experimental trial across groups")
		}
		seenTrials[trial] = true
		g, err := group(o.ArmID, o.CohortID, o.Cache.Expected, o.MoneyBasis, o.MoneyUnit, o.QuotaUnit)
		if err != nil {
			return Report{}, err
		}
		st := o.Status
		if len(o.Quality) > 0 {
			verdict, err := fixture.Decide(o.Quality)
			if err != nil {
				return Report{}, err
			}
			if verdict.Status == fixture.StatusRejected {
				g.regression = true
				if st == manifest.RunAccepted {
					st = manifest.RunRejectedQuality
				}
			}
			if verdict.Status == fixture.StatusPending && st == manifest.RunAccepted {
				st = run.StatusPendingReview
			}
		} else if st == manifest.RunAccepted {
			st = run.StatusPendingReview
		}
		rr := summary.RunResult{CohortID: o.CohortID, TaskID: o.TaskID, Repetition: o.Repetition, Status: st}
		for _, a := range o.Attempts {
			e, err := event(a, o.MoneyBasis, o.MoneyUnit, o.QuotaUnit)
			if err != nil {
				return Report{}, err
			}
			rr.Attempts = append(rr.Attempts, e.cost)
			g.events = append(g.events, e)
		}
		g.runs = append(g.runs, rr)
		cache, err := ObserveCache(g.arm.Layout, o.Cache)
		if err != nil {
			return Report{}, err
		}
		g.cache[cache.Observed]++
		cacheAssessments = append(cacheAssessments, CacheAssessment{RunID: o.ID, Result: cache})
	}
	prepIDs := map[string]bool{}
	for _, pre := range d.Preparation {
		if strings.TrimSpace(pre.ID) == "" || prepIDs[pre.ID] || strings.TrimSpace(pre.BlockID) == "" {
			return Report{}, fmt.Errorf("invalid or repeated preparation id/block")
		}
		prepIDs[pre.ID] = true
		expected := pre.ExpectedCache
		if expected != "" && expected != "cold" && expected != "warm" {
			return Report{}, fmt.Errorf("invalid preparation cache assignment")
		}
		if expected == "" {
			conditions := map[string]bool{}
			for _, o := range d.Runs {
				if o.ArmID == pre.ArmID && o.CohortID == pre.CohortID && o.BlockID == pre.BlockID {
					conditions[o.Cache.Expected] = true
				}
			}
			if len(conditions) == 1 {
				for condition := range conditions {
					expected = condition
				}
			}
		}
		g, err := group(pre.ArmID, pre.CohortID, expected, pre.MoneyBasis, pre.MoneyUnit, pre.QuotaUnit)
		if err != nil {
			return Report{}, err
		}
		e, err := event(pre.Cost, pre.MoneyBasis, pre.MoneyUnit, pre.QuotaUnit)
		if err != nil {
			return Report{}, err
		}
		g.prep = append(g.prep, e.cost)
		g.events = append(g.events, e)
	}
	rep := Report{Version: p.Version, Basis: "offline conditional calculation; input declarations are not verified execution or client behavior", Decisions: p.Decisions, Caches: cacheAssessments,
		Adoption: AdoptionGate{Status: "unavailable", Reasons: []string{"calibration only", "independent 06 evidence assessment is not integrated", "07 control-result and human adoption review are not integrated"}}}
	normalized, err := json.Marshal(d)
	if err != nil {
		return Report{}, err
	}
	rep.Source = SourceSummary{BaselineRevision: p.BaselineRevision, DelegationRevision: p.DelegationRevision, BaselineProtocolVersion: p.Baseline.ProtocolVersion, FixtureRevision: p.Baseline.FixtureRevision, RubricVersion: p.Baseline.RubricVersion, DatasetFingerprint: record.SHA256(normalized), FingerprintMethod: "modelcache/v1 Go encoding/json of typed Dataset; not raw file bytes, a receipt or independent provenance"}
	for _, a := range p.Arms {
		rep.Source.Layouts = append(rep.Source.Layouts, ArmLayoutSource{ArmID: a.ID, Layout: a.Layout})
	}
	sort.Slice(rep.Source.Layouts, func(i, j int) bool { return rep.Source.Layouts[i].ArmID < rep.Source.Layouts[j].ArmID })
	for source := range sources {
		rep.Source.SourceIDs = append(rep.Source.SourceIDs, source)
	}
	sort.Strings(rep.Source.SourceIDs)
	if !p.BaselineStable {
		rep.Adoption.Reasons = append(rep.Adoption.Reasons, "02 baseline stability not established")
	}
	var keys []string
	fragments := map[string]int{}
	conditions := map[string]map[string]bool{}
	for key := range groups {
		keys = append(keys, key)
		g := groups[key]
		baseKey := g.arm.ID + "\x00" + g.cohort.ID
		fragments[baseKey+"\x00"+g.expected]++
		if conditions[baseKey] == nil {
			conditions[baseKey] = map[string]bool{}
		}
		conditions[baseKey][g.expected] = true
	}
	sort.Strings(keys)
	for _, key := range keys {
		g := groups[key]
		m := p.Baseline
		m.Cohorts = []manifest.Cohort{g.cohort}
		part := manifest.Plan{}
		for _, x := range plan.Executable {
			if x.CohortID == g.cohort.ID {
				part.Executable = append(part.Executable, x)
			}
		}
		for _, x := range plan.Incompatible {
			if x.CohortID == g.cohort.ID {
				part.Incompatible = append(part.Incompatible, x)
			}
		}
		base, err := summary.Compute(m, part, g.runs)
		if err != nil {
			return Report{}, err
		}
		if len(base.Cohorts) == 0 {
			return Report{}, fmt.Errorf("preparation or run belongs to an incompatible cohort")
		}
		x := base.Cohorts[0]
		row := Row{ExpectedCache: g.expected, Pilot: g.arm.Pilot, ArmID: g.arm.ID, CohortID: g.cohort.ID, Client: x.Client, OS: string(g.cohort.Cell.OS), Interface: x.Interface, Modality: x.Modality, MoneyBasis: g.basis, MoneyUnit: g.unit, QuotaUnit: g.quota, Started: x.Started, Accepted: x.Accepted, Rejected: x.Rejected, Pending: x.Pending, Attempts: x.Attempts, Preparations: len(g.prep), ObservedCache: g.cache, MandatoryRegression: g.regression}
		var prepT, prepM, prepQ []Metric
		for _, c := range g.prep {
			prepT = append(prepT, fromAmount(c.Tokens))
			prepM = append(prepM, fromAmount(c.Money))
			prepQ = append(prepQ, fromAmount(c.Quota))
		}
		row.Tokens = costSummary(x.Tokens, prepT, d.Preparation != nil, x.Accepted)
		row.Money = costSummary(x.Money, prepM, d.Preparation != nil, x.Accepted)
		row.Quota = costSummary(x.Quota, prepQ, d.Preparation != nil, x.Accepted)
		baseKey := g.arm.ID + "\x00" + g.cohort.ID
		if fragments[baseKey+"\x00"+g.expected] > 1 || conditions[baseKey][""] && len(conditions[baseKey]) > 1 {
			row.FragmentedCoverage = true
			for _, measure := range []*CostSummary{&row.Tokens, &row.Money, &row.Quota} {
				measure.Tasks.Complete = false
				measure.Preparation.Complete = false
				measure.Total.Complete = false
				measure.PerAccepted = Metric{Reason: "cohort fragmented across measurement bases or units; complete denominator unavailable"}
			}
		}
		var context [6][]Metric
		for _, e := range g.events {
			for i, v := range e.context {
				context[i] = append(context[i], v)
			}
		}
		row.LogicalInput = sumMetrics(context[0], false)
		row.CacheRead = sumMetrics(context[1], false)
		row.CacheWrite = sumMetrics(context[2], false)
		row.Output = sumMetrics(context[3], false)
		row.SummedElapsedMS = sumMetrics(context[4], false)
		row.SummedHumanMS = sumMetrics(context[5], false)
		rep.Rows = append(rep.Rows, row)
	}
	sort.Slice(rep.Caches, func(i, j int) bool { return rep.Caches[i].RunID < rep.Caches[j].RunID })
	rep.Pairs = pairResults(p, d.Runs)
	rep.Contrasts = contrastResults(p, rep.Rows, rep.Pairs, d.Runs)
	return rep, nil
}

func costSummary(task summary.Measure, prep []Metric, prepKnown bool, accepted int) CostSummary {
	r := CostSummary{Tasks: fromAmount(task.Total), Preparation: sumMetrics(prep, prepKnown)}
	r.Total = sumMetrics([]Metric{r.Tasks, r.Preparation}, false)
	switch {
	case accepted == 0:
		r.PerAccepted.Reason = "no accepted tasks"
	case r.Total.Value == nil || !r.Total.Complete:
		r.PerAccepted.Reason = "incomplete measure or preparation coverage"
	default:
		r.PerAccepted = metric(*r.Total.Value/float64(accepted), true, "")
	}
	return r
}

func pairResults(p Protocol, runs []Observation) []PairResult {
	arms := map[string]Arm{}
	for _, a := range p.Arms {
		arms[a.ID] = a
	}
	var out []PairResult
	for _, o := range runs {
		a := arms[o.ArmID]
		if a.Factor == "control" {
			continue
		}
		result := PairResult{ID: o.PairID, CandidateRun: o.ID, Comparison: Comparison{Basis: "conditional on offline declarations", Reasons: []string{"control run missing or ambiguous"}}}
		var matches []Observation
		for _, c := range runs {
			if o.PairID != "" && c.PairID == o.PairID && c.ArmID == a.ControlID {
				matches = append(matches, c)
			}
		}
		if len(matches) == 1 {
			c := matches[0]
			result.ControlRun = c.ID
			result.Comparison = CompareArms(arms[c.ArmID], a)
			bad := func(s string) {
				result.Comparison.Comparable = false
				result.Comparison.Reasons = append(result.Comparison.Reasons, s)
			}
			if c.CohortID != o.CohortID || c.TaskID != o.TaskID || c.Repetition != o.Repetition || c.BlockID == "" || c.BlockID != o.BlockID || c.FixtureRevision == "" || c.FixtureRevision != o.FixtureRevision || c.FixtureRevision != p.Baseline.FixtureRevision || c.RubricVersion == "" || c.RubricVersion != o.RubricVersion || c.RubricVersion != p.Baseline.RubricVersion {
				bad("pair task, fixture, rubric, repetition or block differs")
			}
			var frozen manifest.Cell
			for _, cohort := range p.Baseline.Cohorts {
				if cohort.ID == c.CohortID {
					frozen = cohort.Cell
				}
			}
			if !reflect.DeepEqual(fixedCell(c.Cell), fixedCell(o.Cell)) || !reflect.DeepEqual(fixedCell(c.Cell), fixedCell(frozen)) || manifest.CheckCell(c.Cell).Status != manifest.CellExecutable || manifest.CheckCell(o.Cell).Status != manifest.CellExecutable {
				bad("effective cohort/version/interface/permissions/tools differ or are unavailable")
			}
			if !reflect.DeepEqual(c.Effective, arms[c.ArmID].Effective) || !reflect.DeepEqual(o.Effective, a.Effective) {
				bad("per-run effective parameters differ or are missing")
			}
			if c.Cache.Expected != o.Cache.Expected {
				bad("assigned cache conditions differ")
			}
			if a.Pilot == "B" && (c.Cache.Expected == "" || o.Cache.Expected == "") {
				bad("assigned cache condition unavailable")
			}
			if c.MoneyBasis != o.MoneyBasis || c.MoneyUnit != o.MoneyUnit || c.QuotaUnit != o.QuotaUnit {
				bad("measurement basis or unit differs")
			}
		}
		out = append(out, result)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].CandidateRun < out[j].CandidateRun
	})
	return out
}

func fixedCell(c manifest.Cell) manifest.Cell {
	// Model/effort are the declared pilot factors, checked per role separately.
	c.Models = nil
	c.Effort = ""
	c.Interface = manifest.CheckCell(c).Interface
	c.Observed.CLIInIDETerminal = false
	c.Tools = append([]string(nil), c.Tools...)
	sort.Strings(c.Tools)
	c.Permissions.Flags = append([]string(nil), c.Permissions.Flags...)
	sort.Strings(c.Permissions.Flags)
	c.Observed.Permissions.Flags = append([]string(nil), c.Observed.Permissions.Flags...)
	sort.Strings(c.Observed.Permissions.Flags)
	return c
}
