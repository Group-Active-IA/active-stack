package uninstall_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

// ─────────────────────────────────────────────────────────────────
// StrategyRestore tests
// ─────────────────────────────────────────────────────────────────

// TestRestoreStrategyCallsRestoreService verifies that StrategyRestore builds
// a plan that calls backup.RestoreService{}.Restore with the provided
// install-time manifest.
func TestRestoreStrategyCallsRestoreService(t *testing.T) {
	var gotManifest backup.Manifest
	called := false

	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		called = true
		gotManifest = m
		return nil
	})
	defer restoreRestoreFn()

	installManifest := backup.Manifest{
		ID:      "install-backup-123",
		RootDir: "/tmp/backups/install-backup-123",
		Source:  backup.BackupSourceInstall,
	}

	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyRestore,
		},
		uninstall.Options{
			HomeDir:         homeDir,
			Registry:        reg,
			RestoreManifest: &installManifest,
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	if !called {
		t.Error("RestoreService.Restore must have been called")
	}
	if gotManifest.ID != installManifest.ID {
		t.Errorf("manifest ID = %q, want %q", gotManifest.ID, installManifest.ID)
	}
}

// TestRestoreStrategySourcesInstallManifestNotFreshSnapshot verifies that the
// restore strategy uses the install-time manifest (the authoritative pristine
// state), NOT the fresh uninstall-time snapshot.
// This is the D3/D4 correctness decision from the design.
func TestRestoreStrategySourcesInstallManifestNotFreshSnapshot(t *testing.T) {
	snapshotCalled := false
	_ = uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		snapshotCalled = true
		return backup.Manifest{ID: "fresh-snap"}, nil
	})
	// NOTE: do NOT defer — we want to check this was NOT called for StrategyRestore.

	installManifest := backup.Manifest{ID: "install-backup", Source: backup.BackupSourceInstall}

	var gotManifest backup.Manifest
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		gotManifest = m
		return nil
	})
	defer restoreRestoreFn()

	homeDir := t.TempDir()
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyRestore,
		},
		uninstall.Options{
			HomeDir:         homeDir,
			Registry:        reg,
			RestoreManifest: &installManifest,
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	// StrategyRestore must NOT take a fresh snapshot (the Prepare stage is skipped).
	if snapshotCalled {
		t.Error("StrategyRestore must NOT call snapshotCreate — it uses the install-time manifest")
	}
	// Must have restored from the install manifest.
	if gotManifest.ID != installManifest.ID {
		t.Errorf("restored from manifest ID %q, want %q", gotManifest.ID, installManifest.ID)
	}

	// Restore the testseam so subsequent tests work normally.
	uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		_ = snapshotCalled
		return backup.Manifest{}, nil
	})
}

// TestRestoreStrategyRequiresManifest verifies that BuildPlan returns an error
// when StrategyRestore is requested but no RestoreManifest is provided.
func TestRestoreStrategyRequiresManifest(t *testing.T) {
	homeDir := t.TempDir()
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	_, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyRestore,
		},
		uninstall.Options{
			HomeDir:         homeDir,
			Registry:        reg,
			RestoreManifest: nil, // explicitly nil
		},
	)
	if err == nil {
		t.Error("BuildPlan with StrategyRestore and nil RestoreManifest must return error")
	}
}

// ─────────────────────────────────────────────────────────────────
// permissionsRemovalStep tests
// ─────────────────────────────────────────────────────────────────

// TestPermissionsRemovalStepRunDoesNotCallRestoreFn is the regression guard
// for a real production bug: Run() used to call restoreFn(*s.manifest) —
// which restores the ENTIRE shared uninstall-time snapshot (every path
// captured for every harness in this uninstall run), not something scoped to
// permissions/settings.json. Since "permissions" (the mandatory security
// floor) is part of virtually every uninstall, this silently reverted
// whatever earlier steps (e.g. markerRemovalStep) had already removed in the
// same run — a targeted uninstall of anything appeared to do nothing. Run
// must NOT call restoreFn; only Rollback (used when the overall pipeline
// fails) legitimately does.
func TestPermissionsRemovalStepRunDoesNotCallRestoreFn(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	settingsPath := adapter.SettingsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"permissions":["allow:*"]}`), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	snapshotTaken := false
	restoreCount := 0

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		snapshotTaken = true
		return backup.Manifest{ID: "uninstall-snap"}, nil
	})
	defer restoreSnap()

	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		restoreCount++
		return nil
	})
	defer restoreRestoreFn()

	h := model.Harness{
		ID:           "permissions",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatalf("prepare step error = %v", err)
		}
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("apply step error = %v", err)
		}
	}

	if !snapshotTaken {
		t.Error("snapshotCreate must have been called in Prepare stage")
	}
	if restoreCount != 0 {
		t.Errorf("permissionsRemovalStep.Run() must NOT call restoreFn (that is Rollback's job); called %d times", restoreCount)
	}
}

// TestUninstall_PermissionsStep_DoesNotRevertSiblingHarnessRemoval is the
// end-to-end regression guard: with a real config harness (marker removal)
// AND the permissions harness in the SAME uninstall run, the config
// harness's removal must survive after the permissions step runs later in
// Apply order — reproducing the exact production report (selecting Claude to
// uninstall appeared to do nothing because CLAUDE.md silently reverted).
func TestUninstall_PermissionsStep_DoesNotRevertSiblingHarnessRemoval(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	instructionsPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instructionsPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	original := "# My Project Config\n\n<!-- active-stack:sdd-orchestrator -->\nbody\n<!-- /active-stack:sdd-orchestrator -->\n\n## Notes\nKeep this.\n"
	if err := os.WriteFile(instructionsPath, []byte(original), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	settingsPath := adapter.SettingsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"permissions":["allow:*"]}`), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "uninstall-snap"}, nil
	})
	defer restoreSnap()

	harnesses := []model.Harness{
		{
			ID:           "sdd-orchestrator",
			Type:         model.HarnessConfig,
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
		{
			ID:           "permissions",
			Type:         model.HarnessConfig,
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: harnesses},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatalf("prepare step %q error = %v", step.ID(), err)
		}
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("apply step %q error = %v", step.ID(), err)
		}
	}

	got, err := os.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("read instructions file: %v", err)
	}
	if strings.Contains(string(got), "active-stack:sdd-orchestrator") {
		t.Errorf("sdd-orchestrator marker removal was reverted after permissions step ran; got:\n%s", got)
	}
	if !strings.Contains(string(got), "Keep this.") {
		t.Errorf("user content outside the marker must survive; got:\n%s", got)
	}
}
