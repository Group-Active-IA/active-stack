package install_test

import (
	"errors"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// TestBuildPlanPrepareHasSnapshotStepFirst verifies that the Prepare stage
// always starts with a snapshot step, and that it appears before any Apply
// write steps.
func TestBuildPlanPrepareHasSnapshotStepFirst(t *testing.T) {
	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	if len(plan.Prepare) == 0 {
		t.Fatal("Prepare stage must not be empty when there are Apply steps")
	}

	// First step in Prepare must be the snapshot step.
	if plan.Prepare[0].ID() != "snapshot" {
		t.Errorf("first Prepare step ID = %q, want %q", plan.Prepare[0].ID(), "snapshot")
	}

	// There must also be Apply steps.
	if len(plan.Apply) == 0 {
		t.Fatal("Apply stage must not be empty")
	}
}

// TestBuildPlanSnapshotStepRunsBeforeApply verifies that when the plan is
// executed, the snapshot step runs before any Apply step.
func TestBuildPlanSnapshotStepRunsBeforeApply(t *testing.T) {
	order := []string{}

	// Override snapshotCreate to record the call.
	restore := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		order = append(order, "snapshot")
		return backup.Manifest{}, nil
	})
	defer restore()

	// Override configInstallFn to record the call.
	restoreConfig := install.SetConfigInstallFn(func(h model.Harness, _ interface{}, homeDir string) error {
		order = append(order, "config:"+h.ID)
		return nil
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	homeDir := t.TempDir()
	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if len(order) < 2 {
		t.Fatalf("expected at least 2 operations (snapshot + config), got %v", order)
	}
	if order[0] != "snapshot" {
		t.Errorf("snapshot must run first, got order: %v", order)
	}
}

// TestBuildPlanSnapshotFailureAbortsApply verifies that if the snapshot step
// fails, no Apply step runs (Prepare failure aborts Apply — pipeline contract).
func TestBuildPlanSnapshotFailureAbortsApply(t *testing.T) {
	applyRan := false

	restore := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, errors.New("disk full")
	})
	defer restore()

	restoreConfig := install.SetConfigInstallFn(func(h model.Harness, _ interface{}, homeDir string) error {
		applyRan = true
		return nil
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)

	if result.Err == nil {
		t.Fatal("expected error from snapshot failure")
	}
	if applyRan {
		t.Error("Apply step must not run when Prepare (snapshot) fails")
	}
}
