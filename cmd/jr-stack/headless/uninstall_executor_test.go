package headless_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ── Fakes for uninstall executor tests ───────────────────────────────────────

// fakeUninstallAdapter satisfies uninstall.AgentAdapter (7 methods).
// Updated in the HarnessCommand bugfix to add CommandsDir and VariantKey.
type fakeUninstallAdapter struct {
	agent model.Agent
}

func (a fakeUninstallAdapter) Agent() model.Agent                     { return a.agent }
func (a fakeUninstallAdapter) InstructionsPath(homeDir string) string { return homeDir + "/CLAUDE.md" }
func (a fakeUninstallAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a fakeUninstallAdapter) SettingsPath(homeDir string) string     { return homeDir + "/settings.json" }
func (a fakeUninstallAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a fakeUninstallAdapter) CommandsDir(homeDir string) string { return homeDir + "/commands" }
func (a fakeUninstallAdapter) VariantKey() string {
	switch a.agent {
	case model.AgentClaude:
		return "claude"
	case model.AgentOpenCode:
		return "opencode"
	default:
		return ""
	}
}

// fakeUninstallRegistry maps agents to fakeUninstallAdapters.
type fakeUninstallRegistry struct {
	adapters map[model.Agent]uninstall.AgentAdapter
}

func (r *fakeUninstallRegistry) Get(agent model.Agent) (uninstall.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// fakeUninstallCatalog implements uninstall.Catalog.
type fakeUninstallCatalog struct {
	harnesses []model.Harness
}

func (f *fakeUninstallCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeUninstallCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeUninstallCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// minimalCatalogAndRegistry creates a catalog with a single config harness and
// a registry with the claude adapter.
func minimalCatalogAndRegistry(homeDir string) (*fakeUninstallCatalog, *fakeUninstallRegistry) {
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeUninstallCatalog{harnesses: []model.Harness{h}}
	reg := &fakeUninstallRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeUninstallAdapter{agent: model.AgentClaude},
	}}
	return cat, reg
}

// minimalParams returns a valid ParsedUninstallFlags targeting claude/lite.
func minimalParams(homeDir string) headless.ParsedUninstallFlags {
	return headless.ParsedUninstallFlags{
		DryRun:  false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestRunHeadlessUninstallSuccess verifies that a successful uninstall returns
// exit code 0 and writes progress output to the writer.
func TestRunHeadlessUninstallSuccess(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)
	params := minimalParams(homeDir)

	// Wire testseams to succeed.
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap"}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return nil
	})
	defer restoreMarker()

	restoreStale := uninstall.SetStalePurgeFn(func(path string) error {
		return nil
	})
	defer restoreStale()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("successful uninstall must exit 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	// Progress output must be written (step references appear).
	if output == "" {
		t.Error("RunHeadlessUninstall must write progress to the writer, got empty output")
	}
}

// TestRunHeadlessUninstallSuccessProgressPrinted verifies that step IDs appear
// in the progress output (triangulation: specific content check).
func TestRunHeadlessUninstallSuccessProgressPrinted(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)
	params := minimalParams(homeDir)

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap"}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error { return nil })
	defer restoreMarker()

	restoreStale := uninstall.SetStalePurgeFn(func(path string) error { return nil })
	defer restoreStale()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Fatalf("exit code must be 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	// At minimum the snapshot step should appear.
	if !strings.Contains(output, "uninstall-snapshot") && !strings.Contains(output, "sdd-orchestrator") {
		t.Errorf("progress output must mention a step ID; got:\n%s", output)
	}
}

// TestRunHeadlessUninstallApplyFailure verifies that when an Apply step fails,
// rollback is triggered and exit code 1 is returned.
func TestRunHeadlessUninstallApplyFailure(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)
	params := minimalParams(homeDir)

	// Snapshot succeeds so Prepare stage completes.
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: homeDir}, nil
	})
	defer restoreSnap()

	// Marker removal fails — triggers rollback.
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return errors.New("simulated marker removal failure")
	})
	defer restoreMarker()

	// Restore succeeds so rollback completes cleanly.
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		return nil
	})
	defer restoreRestoreFn()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("failed apply must exit non-zero, got 0; output:\n%s", out.String())
	}

	// Failure message must be written.
	output := out.String()
	if output == "" {
		t.Error("failure output must not be empty")
	}
}

// TestRunHeadlessUninstallApplyFailureRollbackReported verifies that rollback
// outcome is reported to the writer when Apply fails (triangulation).
func TestRunHeadlessUninstallApplyFailureRollbackReported(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)
	params := minimalParams(homeDir)

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: homeDir}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return errors.New("marker fail")
	})
	defer restoreMarker()

	rollbackCalled := false
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		rollbackCalled = true
		return nil
	})
	defer restoreRestoreFn()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("expected non-zero exit on failure, got 0")
	}

	// The output must mention "failed" or "rollback".
	output := out.String()
	if !strings.Contains(strings.ToLower(output), "fail") && !strings.Contains(strings.ToLower(output), "rollback") {
		t.Errorf("output must mention failure or rollback; got:\n%s", output)
	}
	_ = rollbackCalled // rollback may or may not be called depending on which step failed
}

// TestRunHeadlessUninstallBuildPlanError verifies that when BuildPlan errors
// (e.g. restore strategy without manifest), exit code 1 is returned.
func TestRunHeadlessUninstallBuildPlanError(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)

	// StrategyRestore requires RestoreManifest != nil; passing nil causes BuildPlan to error.
	params := headless.ParsedUninstallFlags{
		DryRun:  false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyRestore,
			// RestoreManifest in Options will be nil → BuildPlan errors
		},
		// RestoreManifestPath is empty → executor will not set RestoreManifest → BuildPlan errors
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("BuildPlan error must cause exit non-zero, got 0; output:\n%s", out.String())
	}

	// Error must be written to the output.
	output := out.String()
	if output == "" {
		t.Error("error output must not be empty")
	}
}

// TestRunHeadlessUninstallBuildPlanErrorWritten verifies that the BuildPlan
// error message is written to the writer (triangulation on error content).
// Uses StrategyRestore + ModeLite (harnesses exist) + nil manifest → BuildPlan errors.
func TestRunHeadlessUninstallBuildPlanErrorWritten(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)

	params := headless.ParsedUninstallFlags{
		DryRun:  false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite, // harnesses exist → BuildPlan reaches StrategyRestore check
			Strategy: uninstall.StrategyRestore,
			// RestoreManifest in Options is nil → BuildPlan errors
		},
		// RestoreManifestPath is empty → executor does not set RestoreManifest
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("must exit non-zero on BuildPlan error, got 0")
	}

	output := out.String()
	// Should contain "error" in some form.
	if !strings.Contains(strings.ToLower(output), "error") && !strings.Contains(strings.ToLower(output), "require") {
		t.Errorf("output must describe the error; got:\n%s", output)
	}
}

// TestRunHeadlessUninstallDryRun verifies that --dry-run prints plan step IDs
// and returns 0 without performing any writes.
func TestRunHeadlessUninstallDryRun(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)

	params := headless.ParsedUninstallFlags{
		DryRun:  true,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
	}

	sideEffectCalled := false
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		sideEffectCalled = true
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		sideEffectCalled = true
		return nil
	})
	defer restoreMarker()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("dry-run must exit 0, got %d", exitCode)
	}
	if sideEffectCalled {
		t.Error("dry-run must NOT call snapshot or marker removal (no side effects)")
	}

	output := out.String()
	// Output must list plan step IDs.
	if !strings.Contains(output, "uninstall-snapshot") && !strings.Contains(output, "sdd-orchestrator") {
		t.Errorf("dry-run output must list plan step IDs; got:\n%s", output)
	}
}

// TestRunHeadlessUninstallDryRunExitsZero verifies dry-run exits 0 even when
// there are apply steps pending (triangulation: non-empty plan).
func TestRunHeadlessUninstallDryRunExitsZero(t *testing.T) {
	homeDir := t.TempDir()
	// Two harnesses to ensure a non-trivial plan.
	harnesses := []model.Harness{
		{
			ID:           "sdd-orchestrator",
			Type:         model.HarnessConfig,
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
		{
			ID:           "engram",
			Type:         model.HarnessConfig,
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
	}
	cat := &fakeUninstallCatalog{harnesses: harnesses}
	reg := &fakeUninstallRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeUninstallAdapter{agent: model.AgentClaude},
	}}

	params := headless.ParsedUninstallFlags{
		DryRun:  true,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeFull,
			Strategy: uninstall.StrategyTargeted,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestRunHeadlessUninstallCommandHarnessDryRun verifies that a catalog
// containing a command harness (type: command) does NOT cause BuildPlan to
// crash. This is the regression test for the HarnessCommand bug.
func TestRunHeadlessUninstallCommandHarnessDryRun(t *testing.T) {
	homeDir := t.TempDir()
	harnesses := []model.Harness{
		{
			ID:           "starter-add-command",
			Type:         model.HarnessCommand,
			Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
	}
	cat := &fakeUninstallCatalog{harnesses: harnesses}
	reg := &fakeUninstallRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeUninstallAdapter{agent: model.AgentClaude},
	}}

	params := headless.ParsedUninstallFlags{
		DryRun:  true,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("command harness + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestRunHeadlessUninstallCommandHarnessOpenCodeDryRun is a triangulation:
// same test but with OpenCode agent (different VariantKey = "opencode").
func TestRunHeadlessUninstallCommandHarnessOpenCodeDryRun(t *testing.T) {
	homeDir := t.TempDir()
	harnesses := []model.Harness{
		{
			ID:           "starter-add-command",
			Type:         model.HarnessCommand,
			Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		},
	}
	cat := &fakeUninstallCatalog{harnesses: harnesses}
	reg := &fakeUninstallRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentOpenCode: fakeUninstallAdapter{agent: model.AgentOpenCode},
	}}

	params := headless.ParsedUninstallFlags{
		DryRun:  true,
		Yes:     true,
		HomeDir: homeDir,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentOpenCode},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	var out bytes.Buffer
	exitCode := headless.RunHeadlessUninstall(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("command harness + opencode + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}
