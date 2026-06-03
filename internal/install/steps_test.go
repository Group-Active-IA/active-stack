package install_test

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	perminstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config/permissions"
	skillinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// ─────────────────────────────────────────────────────────────────
// external step
// ─────────────────────────────────────────────────────────────────

func TestExternalStepCallsExternalInstaller(t *testing.T) {
	called := false
	var gotHarness model.Harness
	var gotHomeDir string

	restore := install.SetExternalInstallFn(func(
		_ context.Context,
		h model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		homeDir string,
	) (extinstaller.Result, error) {
		called = true
		gotHarness = h
		gotHomeDir = homeDir
		return extinstaller.Result{}, nil
	})
	defer restore()

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

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !called {
		t.Error("external.Install must have been called")
	}
	if gotHarness.ID != h.ID {
		t.Errorf("harness ID = %q, want %q", gotHarness.ID, h.ID)
	}
	if gotHomeDir != homeDir {
		t.Errorf("homeDir = %q, want %q", gotHomeDir, homeDir)
	}
}

func TestExternalStepPathResolvedViaAdapter(t *testing.T) {
	var gotAdapterMCPPath string

	restore := install.SetExternalInstallFn(func(
		_ context.Context,
		h model.Harness,
		_ system.PlatformProfile,
		adapters []extinstaller.AgentAdapter,
		homeDir string,
	) (extinstaller.Result, error) {
		if len(adapters) > 0 {
			gotAdapterMCPPath = adapters[0].MCPConfigPath(homeDir, h.ID)
		}
		return extinstaller.Result{}, nil
	})
	defer restore()

	h := model.Harness{
		ID:           "context7",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "mcp"},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	_ = orch.Execute(plan.StagePlan)

	// Path must be derived from adapter, not hardcoded.
	expected := homeDir + "/mcp/context7.json"
	if gotAdapterMCPPath != expected {
		t.Errorf("MCPConfigPath = %q, want %q", gotAdapterMCPPath, expected)
	}
}

// ─────────────────────────────────────────────────────────────────
// skill step
// ─────────────────────────────────────────────────────────────────

func TestSkillStepCallsSkillInstaller(t *testing.T) {
	called := false
	var gotHarness model.Harness

	restore := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		h model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		called = true
		gotHarness = h
		return nil, nil
	})
	defer restore()

	h := model.Harness{
		ID:           "skill-h",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/skill-h", Method: "clone"},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !called {
		t.Error("skill.Installer.Install must have been called")
	}
	if gotHarness.ID != h.ID {
		t.Errorf("harness ID = %q, want %q", gotHarness.ID, h.ID)
	}
}

func TestSkillStepPathResolvedViaAdapter(t *testing.T) {
	var gotSkillsDir string

	restore := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		_ model.Harness,
		adapters []skillinstaller.AgentAdapter,
		homeDir, _ string,
	) ([]skillinstaller.Result, error) {
		if len(adapters) > 0 {
			gotSkillsDir = adapters[0].SkillsDir(homeDir)
		}
		return nil, nil
	})
	defer restore()

	h := model.Harness{
		ID:           "skill-h",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/skill-h", Method: "clone"},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	_ = orch.Execute(plan.StagePlan)

	// Path must be derived from adapter, not hardcoded.
	expected := homeDir + "/skills"
	if gotSkillsDir != expected {
		t.Errorf("SkillsDir = %q, want %q", gotSkillsDir, expected)
	}
}

// ─────────────────────────────────────────────────────────────────
// config step
// ─────────────────────────────────────────────────────────────────

func TestConfigStepCallsConfigInstaller(t *testing.T) {
	called := false
	var gotHarness model.Harness

	restore := install.SetConfigInstallFn(func(h model.Harness, _ interface{}, homeDir string) error {
		called = true
		gotHarness = h
		return nil
	})
	defer restore()

	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		Toggles:      []string{"tdd"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !called {
		t.Error("config.Install must have been called")
	}
	if gotHarness.ID != h.ID {
		t.Errorf("harness ID = %q, want %q", gotHarness.ID, h.ID)
	}
}

// ─────────────────────────────────────────────────────────────────
// permissions step
// ─────────────────────────────────────────────────────────────────

func TestPermissionsStepCallsPermissionsInstaller(t *testing.T) {
	called := false
	var gotHomeDir string

	restore := install.SetPermissionsInstallFn(func(
		homeDir string,
		_ []perminstaller.PermissionsAdapter,
		_ model.PermissionTier,
	) (perminstaller.Result, error) {
		called = true
		gotHomeDir = homeDir
		return perminstaller.Result{}, nil
	})
	defer restore()

	h := model.Harness{
		ID:           "permissions",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !called {
		t.Error("permissions.Install must have been called")
	}
	if gotHomeDir != homeDir {
		t.Errorf("homeDir = %q, want %q", gotHomeDir, homeDir)
	}
}

func TestPermissionsStepPathResolvedViaAdapter(t *testing.T) {
	var gotSettingsPath string

	restore := install.SetPermissionsInstallFn(func(
		homeDir string,
		adapters []perminstaller.PermissionsAdapter,
		_ model.PermissionTier,
	) (perminstaller.Result, error) {
		if len(adapters) > 0 {
			gotSettingsPath = adapters[0].SettingsPath(homeDir)
		}
		return perminstaller.Result{}, nil
	})
	defer restore()

	h := model.Harness{
		ID:           "permissions",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
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

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	_ = orch.Execute(plan.StagePlan)

	expected := homeDir + "/settings.json"
	if gotSettingsPath != expected {
		t.Errorf("SettingsPath = %q, want %q", gotSettingsPath, expected)
	}
}

// ─────────────────────────────────────────────────────────────────
// Rollback — restore from snapshot on Apply failure
// ─────────────────────────────────────────────────────────────────

// TestWriteStepRollbackRestoresFromSnapshot verifies that when an Apply step
// fails and rollback runs, it calls restore from the snapshot captured in
// Prepare.
//
// Setup: ext-first (external, succeeds) → cfg-fail (config, fails by dep order).
// Rollback of ext-first must call restoreFn.
func TestWriteStepRollbackRestoresFromSnapshot(t *testing.T) {
	restoreCount := 0

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "test-snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	// ext-first succeeds.
	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	// cfg-fail fails, triggering rollback.
	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return errors.New("config write failed")
	})
	defer restoreConfig()

	restoreRestore := install.SetRestoreFn(func(_ backup.Manifest) error {
		restoreCount++
		return nil
	})
	defer restoreRestore()

	// ext-first has no deps; cfg-fail depends on ext-first → topo: ext-first then cfg-fail.
	extH := model.Harness{
		ID:           "ext-first",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cfgH := model.Harness{
		ID:           "cfg-fail",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
		DependsOn:    []string{"ext-first"},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{extH, cfgH}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)

	if result.Err == nil {
		t.Fatal("expected error from failing config step")
	}

	if restoreCount == 0 {
		t.Errorf("Rollback must have called restore at least once (restoreCount=%d)", restoreCount)
	}
}
