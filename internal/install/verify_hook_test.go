package install_test

import (
	"errors"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// TestVerifyHookStepRunsAfterApply verifies that when Options.VerifyHook is
// provided, the plan includes a "verify-hook" step as the last Apply step and
// it executes after all harness Apply steps.
func TestVerifyHookStepRunsAfterApply(t *testing.T) {
	var order []string

	// Record external step execution.
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	// Record config step execution.
	restoreConfig := install.SetConfigInstallFn(func(h model.Harness, _ interface{}, _ string) error {
		order = append(order, "config:"+h.ID)
		return nil
	})
	defer restoreConfig()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	// Verify hook records its call.
	hookCalled := false
	verifyHook := func() error {
		hookCalled = true
		order = append(order, "verify-hook")
		return nil
	}

	h := model.Harness{
		ID:           "cfg-h",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	homeDir := t.TempDir()
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, verifyHook))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// The last Apply step must be the verify-hook.
	applyIDs := make([]string, len(plan.Apply))
	for i, s := range plan.Apply {
		applyIDs[i] = s.ID()
	}
	if len(applyIDs) == 0 {
		t.Fatal("Apply must not be empty")
	}
	lastID := applyIDs[len(applyIDs)-1]
	if lastID != "verify-hook" {
		t.Errorf("last Apply step ID = %q, want %q; all IDs: %v", lastID, "verify-hook", applyIDs)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !hookCalled {
		t.Error("verify hook must have been called")
	}

	// Verify hook ran after the config step.
	if len(order) < 2 {
		t.Fatalf("expected at least 2 order entries, got %v", order)
	}
	if order[len(order)-1] != "verify-hook" {
		t.Errorf("verify-hook must be last in execution order; got: %v", order)
	}
}

// TestVerifyHookFailureTriggersRollback verifies that when the verify hook
// returns an error, the pipeline rolls back the Apply stage.
func TestVerifyHookFailureTriggersRollback(t *testing.T) {
	rollbackCount := 0

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	restoreRestore := install.SetRestoreFn(func(_ backup.Manifest) error {
		rollbackCount++
		return nil
	})
	defer restoreRestore()

	// Verify hook always fails.
	failHook := func() error {
		return errors.New("verification failed: SKILL.md missing")
	}

	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	homeDir := t.TempDir()
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, failHook))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)

	if result.Err == nil {
		t.Fatal("expected error when verify hook fails")
	}

	// Rollback must have been triggered.
	if rollbackCount == 0 {
		t.Errorf("rollback must have been called when verify hook fails (rollbackCount=%d)", rollbackCount)
	}
}

// TestVerifyHookNilIsNoOp verifies that when VerifyHook is nil, no verify-hook
// step is added and the plan executes without error.
func TestVerifyHookNilIsNoOp(t *testing.T) {
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil)) // nil hook
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, s := range plan.Apply {
		if s.ID() == "verify-hook" {
			t.Error("verify-hook step must NOT be present when VerifyHook is nil")
		}
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v (nil hook should be no-op)", result.Err)
	}
}
