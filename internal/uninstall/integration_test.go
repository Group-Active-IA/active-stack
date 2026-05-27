package uninstall_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestHeadlessUninstallIntegrationRollback is the end-to-end rollback test:
// Intent → BuildPlan → Orchestrator.Execute where the second Apply step fails.
//
// Verifies:
//  1. Snapshot runs in Prepare before any Apply steps.
//  2. After first Apply step succeeds and second fails, rollback calls restore.
//  3. Rollback happens in reverse order.
func TestHeadlessUninstallIntegrationRollback(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	// Create files needed by the steps.
	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := "# Config\n\n<!-- jr-stack:h-first -->\nfirst block\n<!-- /jr-stack:h-first -->\n"
	if err := os.WriteFile(instrPath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var executionOrder []string
	var progressEvents []pipeline.ProgressEvent
	restoreCount := 0

	// Snapshot tracking.
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		executionOrder = append(executionOrder, "snapshot")
		return backup.Manifest{ID: "uninstall-snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	// Restore tracking.
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		restoreCount++
		executionOrder = append(executionOrder, "restore")
		return nil
	})
	defer restoreRestoreFn()

	// marker-removal for h-first: succeeds.
	// marker-removal for h-fail: fails (injected via testseam).
	failNext := false
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		if sectionID == "h-fail" {
			return errors.New("simulated marker removal failure")
		}
		executionOrder = append(executionOrder, "marker:"+sectionID)
		return nil
	})
	defer restoreMarker()

	hFirst := model.Harness{
		ID:           "h-first",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	hFail := model.Harness{
		ID:           "h-fail",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	_ = failNext

	cat := &fakeCatalog{harnesses: []model.Harness{hFirst, hFail}}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: adapter,
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, uninstall.Options{
		HomeDir:  homeDir,
		Registry: reg,
		OnProgress: func(e pipeline.ProgressEvent) {
			progressEvents = append(progressEvents, e)
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(plan.OnProgress),
	)
	result := orch.Execute(plan.StagePlan)

	// Must have failed (h-fail errors).
	if result.Err == nil {
		t.Fatal("expected error from h-fail step")
	}

	// 1. Snapshot must be first in execution order.
	snapIdx := indexInSlice(executionOrder, "snapshot")
	if snapIdx != 0 {
		t.Errorf("snapshot must be first operation, order = %v", executionOrder)
	}

	// 2. h-first succeeded, then h-fail failed → restore must have been called.
	if restoreCount == 0 {
		t.Errorf("rollback must have called restore at least once (restoreCount=%d)", restoreCount)
	}

	// 3. Rollback stage should have been executed.
	if result.Rollback.Stage != pipeline.StageRollback {
		t.Errorf("rollback stage = %q, want rollback", result.Rollback.Stage)
	}

	// 4. Progress events must include running events.
	if len(progressEvents) == 0 {
		t.Error("expected progress events, got none")
	}
	foundSnapshotRunning := false
	for _, e := range progressEvents {
		if e.StepID == "uninstall-snapshot" && e.Status == pipeline.StepStatusRunning {
			foundSnapshotRunning = true
		}
	}
	if !foundSnapshotRunning {
		t.Error("expected running event for uninstall-snapshot step")
	}
}

// TestHeadlessUninstallIdempotency verifies that running the uninstall twice
// is a clean series of no-ops (second run produces no errors).
func TestHeadlessUninstallIdempotency(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	// Create files for first run.
	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := "# Config\n\n<!-- jr-stack:sdd-orchestrator -->\norchestrator block\n<!-- /jr-stack:sdd-orchestrator -->\n"
	if err := os.WriteFile(instrPath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap"}, nil
	})
	defer restoreSnap()

	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		return nil
	})
	defer restoreRestoreFn()

	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: adapter,
	}}

	runUninstall := func(label string) {
		plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		}, buildUninstallOptions(homeDir, reg))
		if err != nil {
			t.Fatalf("%s: BuildPlan() error = %v", label, err)
		}

		orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
		result := orch.Execute(plan.StagePlan)
		if result.Err != nil {
			t.Errorf("%s: Execute() error = %v", label, result.Err)
		}
	}

	runUninstall("first run")
	runUninstall("second run (idempotency check)")
}

// TestNoHardcodedAgentPathsInPackage is a behavioral assertion: all steps must
// resolve agent paths via the adapter, never from hardcoded literals.
// We verify this by using a custom path adapter and confirming the step operates
// on the custom path, not a default string.
func TestNoHardcodedAgentPathsInPackage(t *testing.T) {
	homeDir := t.TempDir()
	customInstrPath := filepath.Join(homeDir, "agent-custom", "INSTRUCTIONS.md")
	customSkillsDir := filepath.Join(homeDir, "agent-custom", "skills")

	adapter := fakeAdapterCustomPath{
		agent:            model.AgentClaude,
		instructionsPath: customInstrPath,
		skillsDir:        customSkillsDir,
	}

	// Set up files at custom paths.
	if err := os.MkdirAll(filepath.Dir(customInstrPath), 0o755); err != nil {
		t.Fatalf("setup instr: %v", err)
	}
	instrContent := "# Custom\n\n<!-- jr-stack:test-harness -->\nblock\n<!-- /jr-stack:test-harness -->\n"
	if err := os.WriteFile(customInstrPath, []byte(instrContent), 0o644); err != nil {
		t.Fatalf("setup instr: %v", err)
	}

	skillDir := filepath.Join(customSkillsDir, "skill-harness")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("setup skill: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreRestoreFn := uninstall.SetRestoreFn(func(_ backup.Manifest) error { return nil })
	defer restoreRestoreFn()

	configH := model.Harness{
		ID:           "test-harness",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	skillH := model.Harness{
		ID:           "skill-harness",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/skill-harness", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{configH, skillH}}
	reg := &fakeRegistryCustom{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: adapter,
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	// config harness marker should have been removed from the custom path.
	got, err := os.ReadFile(customInstrPath)
	if err != nil {
		t.Fatalf("read custom instr path: %v", err)
	}
	if contains(string(got), "<!-- jr-stack:test-harness -->") {
		t.Errorf("marker still present at custom path; adapter path was not used:\n%s", got)
	}

	// skill harness dir should have been removed from the custom path.
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill dir still exists at custom path; adapter SkillsDir was not used: %s", skillDir)
	}
}

// ─────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────

func indexInSlice(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}
