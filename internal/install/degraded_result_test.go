package install_test

// C-32: honest structured result — tasks 2.1, 2.2, 2.3
//
// The D3 seam decision (checkpoint 1.6): we derive the degraded list by
// filtering progress events with Status == StepStatusDegraded. Zero new
// pipeline plumbing; the runner core is untouched.

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	skillinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// collectDegradedFromEvents is the D3 seam: derive the degraded list from
// progress events filtered by StepStatusDegraded.  This lives in the test
// package to prove the derivation pattern that headless/TUI will also use.
func collectDegradedFromEvents(events []pipeline.ProgressEvent) []pipeline.ProgressEvent {
	var out []pipeline.ProgressEvent
	for _, ev := range events {
		if ev.Status == pipeline.StepStatusDegraded {
			out = append(out, ev)
		}
	}
	return out
}

// Task 2.1 RED — after a run with a degraded best-effort harness, the degraded
// list derived from progress events lists that harness and is NOT empty (i.e.
// the run is NOT reported as a clean success).
func TestDegradedResult_SingleHarnessDegraded(t *testing.T) {
	restore := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		_ model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		return nil, errors.New("source.path not found in repo")
	})
	defer restore()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "code-review-excellence",
		Type:         model.HarnessSkill,
		BestEffort:   true,
		Source:       &model.Source{Repo: "some/third-party", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	var events []pipeline.ProgressEvent
	opts := buildOptionsWithProgress(t.TempDir(), reg, nil, func(ev pipeline.ProgressEvent) {
		events = append(events, ev)
	})

	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeFull,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(plan.OnProgress),
	)
	result := orch.Execute(plan.StagePlan)

	// The pipeline must NOT abort.
	if result.Err != nil {
		t.Fatalf("best-effort degradation must not abort pipeline: %v", result.Err)
	}

	// The degraded list must be non-empty (not a clean success).
	degraded := collectDegradedFromEvents(events)
	if len(degraded) == 0 {
		t.Fatalf("expected at least one degraded event; got events: %+v", events)
	}

	// The degraded event must name the harness.
	found := false
	for _, ev := range degraded {
		if strings.Contains(ev.StepID, "code-review-excellence") {
			found = true
		}
	}
	if !found {
		t.Errorf("degraded events do not include code-review-excellence; got: %+v", degraded)
	}
}

// Task 2.3 TRIANGULATE — two degraded harnesses + one clean success → exactly
// two degraded events; the clean harness does NOT appear in the degraded list.
func TestDegradedResult_TwoDegraded_OneClean(t *testing.T) {
	// skill-a and skill-b: degraded (best-effort, install fails)
	// skill-c: external (succeeds)
	restoreSkill := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		h model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		return nil, errors.New("upstream path missing for " + h.ID)
	})
	defer restoreSkill()

	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	skillA := model.Harness{
		ID:           "skill-a",
		Type:         model.HarnessSkill,
		BestEffort:   true,
		Source:       &model.Source{Repo: "owner/repo-a", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	skillB := model.Harness{
		ID:           "skill-b",
		Type:         model.HarnessSkill,
		BestEffort:   true,
		Source:       &model.Source{Repo: "owner/repo-b", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	extC := model.Harness{
		ID:           "ext-c",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{skillA, skillB, extC}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	var events []pipeline.ProgressEvent
	opts := buildOptionsWithProgress(t.TempDir(), reg, nil, func(ev pipeline.ProgressEvent) {
		events = append(events, ev)
	})

	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeFull,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(plan.OnProgress),
	)
	result := orch.Execute(plan.StagePlan)

	if result.Err != nil {
		t.Fatalf("run with degraded best-effort harnesses must not abort: %v", result.Err)
	}

	degraded := collectDegradedFromEvents(events)

	// Exactly two degraded events (skill-a and skill-b).
	if len(degraded) != 2 {
		t.Errorf("expected 2 degraded events, got %d: %+v", len(degraded), degraded)
	}

	// ext-c must NOT appear in the degraded list.
	for _, ev := range degraded {
		if strings.Contains(ev.StepID, "ext-c") {
			t.Errorf("clean harness ext-c must not appear in degraded list; got event: %+v", ev)
		}
	}

	// skill-a and skill-b must appear.
	gotA, gotB := false, false
	for _, ev := range degraded {
		if strings.Contains(ev.StepID, "skill-a") {
			gotA = true
		}
		if strings.Contains(ev.StepID, "skill-b") {
			gotB = true
		}
	}
	if !gotA {
		t.Errorf("skill-a missing from degraded list: %+v", degraded)
	}
	if !gotB {
		t.Errorf("skill-b missing from degraded list: %+v", degraded)
	}
}
