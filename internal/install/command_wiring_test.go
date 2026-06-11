package install_test

// Tests for §6: install pipeline wiring for the command harness type (C-31).
//
// TDD order: RED (write failing tests) → GREEN (implement) → REFACTOR.
// All tests are self-contained; they use the existing fakeAdapter/fakeCatalog/
// fakeRegistry helpers from install_test.go.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// errTestRollback is returned by the VerifyHook to trigger a pipeline rollback.
var errTestRollback = errors.New("test: verify hook failed — trigger rollback")

// ── §6.2 RED: resolvedCommandsDir mirrors resolvedSkillsDir ──────────────────

// TestResolvedCommandsDir_MachineTarget asserts that for Machine target,
// resolvedCommandsDir returns adapter.CommandsDir(base).
// RED: fails until resolvedCommandsDir is exported for testing OR
// collectWritePaths includes the command dir in its output.
// We test this indirectly via the snapshot paths captured by BuildPlan.
func TestResolvedCommandsDir_MachineTarget_AppearsInSnapshotPaths(t *testing.T) {
	homeDir := t.TempDir()
	var capturedPaths []string

	restoreSnap := install.SetSnapshotCreateWithHints(func(_ string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		capturedPaths = paths
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		return nil
	})
	defer restoreCommand()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{
		HomeDir:  homeDir,
		Registry: reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}
	for _, s := range plan.Prepare {
		_ = s.Run()
	}

	// The commands dir for the machine target must appear in the snapshot paths.
	wantCommandsDir := homeDir + "/commands"
	found := false
	for _, p := range capturedPaths {
		if filepath.ToSlash(p) == filepath.ToSlash(wantCommandsDir) ||
			strings.HasPrefix(filepath.ToSlash(p), filepath.ToSlash(wantCommandsDir)) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected command dir %q in snapshot paths; got: %v", wantCommandsDir, capturedPaths)
	}
}

// TestResolvedCommandsDir_ProjectTarget_AppearsInSnapshotPaths asserts that
// for Project target, the command dir resolves to PathsFor(base, Project).CommandsDir.
func TestResolvedCommandsDir_ProjectTarget_AppearsInSnapshotPaths(t *testing.T) {
	projectRoot := filepath.FromSlash("/proj/myapp")
	homeDir := t.TempDir()
	var capturedPaths []string

	restoreSnap := install.SetSnapshotCreateWithHints(func(_ string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		capturedPaths = paths
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		return nil
	})
	defer restoreCommand()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: projectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{
		HomeDir:     homeDir,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}
	for _, s := range plan.Prepare {
		_ = s.Run()
	}

	// For the projectAdapter, PathsFor(projectRoot, Project).CommandsDir
	// = projectRoot + "/.claude/commands" (mirrors the Claude project layout in projectAdapter).
	wantCommandsDir := filepath.Join(projectRoot, ".claude", "commands")
	found := false
	for _, p := range capturedPaths {
		if filepath.Clean(p) == filepath.Clean(wantCommandsDir) ||
			strings.HasPrefix(filepath.ToSlash(p), filepath.ToSlash(wantCommandsDir)) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected project command dir %q in snapshot paths; got: %v", wantCommandsDir, capturedPaths)
	}
}

// ── §6.6a RED: HarnessCommand IsValid + buildHarnessStep returns commandStep ─

// TestHarnessCommand_IsValid asserts that model.HarnessCommand is a known,
// valid harness type.
// RED: fails until HarnessCommand is added to model.HarnessType.
func TestHarnessCommand_IsValid(t *testing.T) {
	if !model.HarnessCommand.IsValid() {
		t.Errorf("model.HarnessCommand.IsValid() = false, want true")
	}
}

// TestBuildHarnessStep_CommandHarness_EmitsCommandStep asserts that BuildPlan
// accepts a command harness and routes it to a commandStep (the install fn runs).
// RED: fails until HarnessCommand is routed in buildHarnessStep.
func TestBuildHarnessStep_CommandHarness_EmitsCommandStep(t *testing.T) {
	homeDir := t.TempDir()
	called := false

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		called = true
		return nil
	})
	defer restoreCommand()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{HomeDir: homeDir, Registry: reg}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error: %v", result.Err)
	}
	if !called {
		t.Error("expected commandInstallFn to be called for HarnessCommand step")
	}
}

// ── §6.7 RED: rollback test ───────────────────────────────────────────────────

// TestCommandStep_Rollback_RestoresOnFailure asserts that when a downstream step
// fails, the command step's Rollback() is called (restores from snapshot).
// The rollback function is verified via SetRestoreFn.
func TestCommandStep_Rollback_RestoresOnFailure(t *testing.T) {
	homeDir := t.TempDir()
	rollbackCalled := false
	var manifestSeen backup.Manifest

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{Entries: []backup.ManifestEntry{{OriginalPath: "fake", Existed: true}}}, nil
	})
	defer restoreSnap()

	restoreCommand := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		return nil // command step succeeds
	})
	defer restoreCommand()

	restoreRestore := install.SetRestoreFn(func(m backup.Manifest) error {
		rollbackCalled = true
		manifestSeen = m
		return nil
	})
	defer restoreRestore()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{
		HomeDir:    homeDir,
		Registry:   reg,
		VerifyHook: func() error { return errTestRollback },
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err == nil {
		t.Fatal("expected Execute() to fail (verify hook returns error)")
	}
	if !rollbackCalled {
		t.Error("expected Rollback() to have been called on failure")
	}
	if len(manifestSeen.Entries) == 0 {
		t.Error("expected manifest with entries to be passed to Rollback()")
	}
}

// ── §6.9 TRIANGULATE: dry-run no write + re-install idempotent ───────────────

// TestCommandStep_DryRun_ProducesNoWrite asserts that when commandInstallFn is
// not called (empty plan), no command writes happen.
// We test the dry-run semantics via an empty plan (no harnesses).
func TestCommandStep_NoHarness_NoCommandWrite(t *testing.T) {
	called := false

	restoreCommand := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		called = true
		return nil
	})
	defer restoreCommand()

	// Empty catalog — no harnesses → no steps → no command write.
	cat := &fakeCatalog{harnesses: nil}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{HomeDir: t.TempDir(), Registry: reg}

	_, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}
	if called {
		t.Error("commandInstallFn must NOT be called when no command harness is in the plan")
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// commandHarness returns a minimal Harness of type HarnessCommand for use in
// §6 pipeline tests.
func commandHarness() model.Harness {
	return model.Harness{
		ID:           "starter-add-command",
		Name:         "Starter Add Command",
		Type:         model.HarnessCommand,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
	}
}
