package install_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// buildOptionsWithProgress is like buildOptions but with a ProgressFunc.
func buildOptionsWithProgress(homeDir string, reg install.Registry, verify func() error, onProgress pipeline.ProgressFunc) install.Options {
	return install.Options{
		HomeDir:    homeDir,
		Registry:   reg,
		VerifyHook: verify,
		OnProgress: onProgress,
	}
}

// TestProgressFuncReceivesEvents verifies that when a ProgressFunc is provided
// via Options, the pipeline emits running→succeeded events for each step.
func TestProgressFuncReceivesEvents(t *testing.T) {
	var events []pipeline.ProgressEvent

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
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptionsWithProgress(homeDir, reg, nil, func(e pipeline.ProgressEvent) {
		events = append(events, e)
	}))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// BuildPlan itself doesn't run the plan; wire the ProgressFunc into the Orchestrator.
	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(plan.OnProgress),
	)
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	// Expect at least 2 events: running + succeeded for the ext-h step.
	foundRunning := false
	foundSucceeded := false
	for _, e := range events {
		if e.StepID == "external:ext-h" {
			if e.Status == pipeline.StepStatusRunning {
				foundRunning = true
			}
			if e.Status == pipeline.StepStatusSucceeded {
				foundSucceeded = true
			}
		}
	}
	if !foundRunning {
		t.Error("expected running event for external:ext-h")
	}
	if !foundSucceeded {
		t.Error("expected succeeded event for external:ext-h")
	}
}

// TestProgressFuncNilIsNoop verifies that not providing a ProgressFunc does
// not panic and execution completes normally.
func TestProgressFuncNilIsNoop(t *testing.T) {
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
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// No progress func — should not panic.
	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}
}

// TestVerifyHookRunsAfterSuccessfulApply verifies that a provided verify hook
// is called after all Apply steps succeed.
func TestVerifyHookRunsAfterSuccessfulApply(t *testing.T) {
	hookCalled := false

	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	verifyHook := func() error {
		hookCalled = true
		return nil
	}

	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, verifyHook))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !hookCalled {
		t.Error("verify hook must be called after successful Apply")
	}
}

// TestVerifyHookNilIsNoop verifies that omitting a verify hook does not panic.
func TestVerifyHookNilIsNoop(t *testing.T) {
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
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	// nil verify hook
	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}
}
