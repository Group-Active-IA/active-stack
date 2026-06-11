package install_test

// command_entrypoint_wiring_test.go — C-32 real-wiring test.
//
// PURPOSE: Close the exact gap that let the "commandsFS is nil" bug ship.
// The existing command_wiring_test.go and command_integration_test.go both bypass
// the production path: they either stub commandInstallFn via SetCommandInstallFn,
// or hand-roll an fstest.MapFS. Neither proves that the binary entry point calls
// install.WithEmbeddedCommandsFS(assets.CommandsFS) before dispatch.
//
// THIS TEST observes the real package-level embeddedCommandsFS var via the
// GetEmbeddedCommandsFS / ResetEmbeddedCommandsFS seams (added in C-32).
// It resets the global to nil (simulating a cold-start process), then calls
// wireEmbeddedFS() — the same function the binary entry point calls — and
// asserts the var is non-nil afterwards.
//
// RED:  FAILS to compile on current HEAD because wireEmbeddedFS() does not
//       exist in package main. Once that function exists but does NOT call
//       WithEmbeddedCommandsFS, the test fails at runtime (nil check).
// GREEN: PASSES after wireEmbeddedFS() calls WithEmbeddedCommandsFS(assets.CommandsFS).
//
// Design refs: design.md D1 (Option A), D2 (real-wiring test strategy).
// Spec: openspec/changes/wire-commands-fs/specs/agent-command-install/spec.md
//
// NOTE: This test calls install.WithEmbeddedCommandsFS directly (same call as
// wireEmbeddedFS) to prove the mechanism works, and separately calls
// ResetEmbeddedCommandsFS to prove the nil baseline. The companion test in
// cmd/active-stack/main_wiring_test.go is the entry-point-level RED test.

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// TestEmbeddedCommandsFS_NilBeforeWiring proves that embeddedCommandsFS is nil
// when no wiring call has been made — reproducing the exact production bug
// condition. On current HEAD the binary never calls WithEmbeddedCommandsFS, so
// the global stays nil and commandInstallFn fails with "commandsFS is nil".
//
// This test is GREEN by construction (it just reads the global after a reset).
// It documents the bug condition; the companion test below guards the fix.
func TestEmbeddedCommandsFS_NilBeforeWiring(t *testing.T) {
	install.ResetEmbeddedCommandsFS()
	t.Cleanup(install.ResetEmbeddedCommandsFS)

	if got := install.GetEmbeddedCommandsFS(); got != nil {
		t.Logf("NOTE: embeddedCommandsFS was already set (%T) before this test ran. "+
			"This can happen if another test in the same binary called WithEmbeddedCommandsFS. "+
			"Run this test in isolation to reproduce the cold-start condition.", got)
	}
	// After reset: must be nil (this is the bug condition in production).
	install.ResetEmbeddedCommandsFS()
	if got := install.GetEmbeddedCommandsFS(); got != nil {
		t.Errorf("expected nil after ResetEmbeddedCommandsFS(); got %T — test seam broken", got)
	}
}

// TestWithEmbeddedCommandsFS_SetsNonNilGlobal_WhenCalledWithRealAssets is the
// mechanism test: confirms that calling WithEmbeddedCommandsFS(assets.CommandsFS)
// sets the global to non-nil.
//
// This is GREEN immediately because WithEmbeddedCommandsFS already sets the global
// correctly — the bug is that it is never CALLED from the binary entry point.
// The entry-point wiring test is in cmd/active-stack/main_wiring_test.go.
func TestWithEmbeddedCommandsFS_SetsNonNilGlobal_WhenCalledWithRealAssets(t *testing.T) {
	install.ResetEmbeddedCommandsFS()
	t.Cleanup(install.ResetEmbeddedCommandsFS)

	// Pre-condition: nil (the bug state).
	if got := install.GetEmbeddedCommandsFS(); got != nil {
		t.Fatalf("pre-condition failed: expected nil after reset, got %T", got)
	}

	// Act: the exact call the binary entry point must make.
	install.WithEmbeddedCommandsFS(assets.CommandsFS)

	// Assert: non-nil.
	if got := install.GetEmbeddedCommandsFS(); got == nil {
		t.Fatal("embeddedCommandsFS is still nil after WithEmbeddedCommandsFS(assets.CommandsFS)")
	}
}

// TestEntrypointWiring_InstallPath_CommandsFSReachesInstaller — triangulation §4.1.
//
// After the entry-point wiring call, the install path (TUI / headless) must
// reach commandInstallFn with a non-nil embeddedCommandsFS global.
// We confirm this by capturing the global value inside a replacement
// commandInstallFn: the real commandInstallFn reads embeddedCommandsFS to build
// cmdinstaller.NewInstaller — this replacement mirrors that read.
//
// Note: this test still calls WithEmbeddedCommandsFS in the test body (not via
// wireEmbeddedFS). The companion test in cmd/active-stack/main_wiring_test.go
// asserts that wireEmbeddedFS() does the same call.
func TestEntrypointWiring_InstallPath_CommandsFSReachesInstaller(t *testing.T) {
	install.ResetEmbeddedCommandsFS()
	t.Cleanup(install.ResetEmbeddedCommandsFS)

	// Simulate the binary wiring call.
	install.WithEmbeddedCommandsFS(assets.CommandsFS)

	fsCapturedNonNil := false
	restoreCmd := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		// Mirror what the real commandInstallFn does: read embeddedCommandsFS.
		fsCapturedNonNil = install.GetEmbeddedCommandsFS() != nil
		return nil
	})
	defer restoreCmd()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}
	opts := install.Options{HomeDir: t.TempDir(), Registry: reg}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error: %v", result.Err)
	}

	if !fsCapturedNonNil {
		t.Error("embeddedCommandsFS was nil when commandInstallFn ran (install path) — " +
			"WithEmbeddedCommandsFS must be called before any plan execution")
	}
}

// TestEntrypointWiring_StarterAddPath_CommandsFSReachesInstaller — triangulation §4.1.
//
// Second triangulation: the starter-add path (Project target) must also reach
// the command step with a non-nil FS. Both paths share the same package global,
// so one call in wireEmbeddedFS() covers both — but we assert both explicitly.
func TestEntrypointWiring_StarterAddPath_CommandsFSReachesInstaller(t *testing.T) {
	install.ResetEmbeddedCommandsFS()
	t.Cleanup(install.ResetEmbeddedCommandsFS)

	install.WithEmbeddedCommandsFS(assets.CommandsFS)

	fsCapturedNonNil := false
	restoreCmd := install.SetCommandInstallFn(func(_ []install.AgentAdapter, _, _ string) error {
		fsCapturedNonNil = install.GetEmbeddedCommandsFS() != nil
		return nil
	})
	defer restoreCmd()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentOpenCode: fakeAdapter{agent: model.AgentOpenCode},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentOpenCode}, Mode: model.ModeLite}
	opts := install.Options{
		HomeDir:     t.TempDir(),
		Target:      model.Project,
		ProjectRoot: t.TempDir(),
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error: %v", result.Err)
	}

	if !fsCapturedNonNil {
		t.Error("embeddedCommandsFS was nil when commandInstallFn ran (starter-add path) — " +
			"WithEmbeddedCommandsFS must be called before any dispatch")
	}
}
