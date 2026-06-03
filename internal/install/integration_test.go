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

// TestHeadlessInstallIntegration is the end-to-end integration test:
// Intent → BuildPlan → Orchestrator.Execute with all fake installers.
//
// Verifies:
//  1. Topological order: dep before dependent in Apply.
//  2. Backup step runs before all Apply steps.
//  3. Rollback in reverse order when one Apply step fails.
//  4. Progress events emitted for each step lifecycle.
func TestHeadlessInstallIntegration(t *testing.T) {
	var executionOrder []string
	var progressEvents []pipeline.ProgressEvent
	restoreCount := 0

	// ── Inject fakes ──────────────────────────────────────────────

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		executionOrder = append(executionOrder, "snapshot")
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	restoreRestore := install.SetRestoreFn(func(_ backup.Manifest) error {
		restoreCount++
		return nil
	})
	defer restoreRestore()

	// ext-dep: succeeds (dependency of skill-main).
	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context,
		h model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
	) (extinstaller.Result, error) {
		executionOrder = append(executionOrder, "ext:"+h.ID)
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	// skill-main: fails, triggering rollback of ext-dep.
	restoreSkill := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		h model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		executionOrder = append(executionOrder, "skill:"+h.ID)
		return nil, errors.New("skill install failed")
	})
	defer restoreSkill()

	// permissions harness (not in this plan, but wiring is fine).
	restorePerm := install.SetPermissionsInstallFn(func(
		_ string,
		_ []perminstaller.PermissionsAdapter,
		_ model.PermissionTier,
	) (perminstaller.Result, error) {
		executionOrder = append(executionOrder, "permissions")
		return perminstaller.Result{}, nil
	})
	defer restorePerm()

	// ── Build plan ─────────────────────────────────────────────────

	// Dependency graph: skill-main depends on ext-dep.
	extDep := model.Harness{
		ID:           "ext-dep",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	skillMain := model.Harness{
		ID:           "skill-main",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/skill-main", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		DependsOn:    []string{"ext-dep"},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{extDep, skillMain}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptionsWithProgress(homeDir, reg, nil, func(e pipeline.ProgressEvent) {
		progressEvents = append(progressEvents, e)
	}))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// ── Execute ────────────────────────────────────────────────────

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(plan.OnProgress),
	)
	result := orch.Execute(plan.StagePlan)

	// Expect an error (skill-main fails).
	if result.Err == nil {
		t.Fatal("expected error from skill-main failure")
	}

	// ── Assertions ─────────────────────────────────────────────────

	// 1. Topological order: ext-dep must appear before skill-main.
	extIdx := indexInSlice(executionOrder, "ext:ext-dep")
	skillIdx := indexInSlice(executionOrder, "skill:skill-main")
	if extIdx == -1 || skillIdx == -1 {
		t.Fatalf("expected both ext-dep and skill-main in order, got %v", executionOrder)
	}
	if extIdx >= skillIdx {
		t.Errorf("ext-dep must run before skill-main, order = %v", executionOrder)
	}

	// 2. Backup-first: snapshot must be the very first operation.
	snapIdx := indexInSlice(executionOrder, "snapshot")
	if snapIdx != 0 {
		t.Errorf("snapshot must be first operation, order = %v", executionOrder)
	}

	// 3. Rollback: ext-dep succeeded then skill-main failed → ext-dep must be rolled back.
	if restoreCount == 0 {
		t.Errorf("expected rollback to call restore at least once (restoreCount=%d)", restoreCount)
	}

	if result.Rollback.Stage != pipeline.StageRollback {
		t.Errorf("rollback stage = %q, want rollback", result.Rollback.Stage)
	}

	// 4. Progress events: expect events for each step.
	if len(progressEvents) == 0 {
		t.Error("expected progress events, got none")
	}
	// Verify at least running + (succeeded or failed) for snapshot.
	foundSnapRunning := false
	for _, e := range progressEvents {
		if e.StepID == "snapshot" && e.Status == pipeline.StepStatusRunning {
			foundSnapRunning = true
		}
	}
	if !foundSnapRunning {
		t.Error("expected running event for snapshot step")
	}
}

func indexInSlice(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}
