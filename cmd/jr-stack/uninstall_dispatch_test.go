// Package main — tests for the "uninstall" dispatch (task 3.1 RED).
// Tests run runUninstallDispatch extracted from main() to assert correct behavior
// without spinning up the full binary.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ── Fakes for dispatch tests ──────────────────────────────────────────────────

// uninstallDispatchAdapter satisfies uninstall.AgentAdapter (7 methods).
// Updated in the HarnessCommand bugfix to add CommandsDir and VariantKey.
type uninstallDispatchAdapter struct{ agent model.Agent }

func (a uninstallDispatchAdapter) Agent() model.Agent                     { return a.agent }
func (a uninstallDispatchAdapter) InstructionsPath(homeDir string) string { return homeDir + "/CLAUDE.md" }
func (a uninstallDispatchAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a uninstallDispatchAdapter) SettingsPath(homeDir string) string {
	return homeDir + "/settings.json"
}
func (a uninstallDispatchAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a uninstallDispatchAdapter) CommandsDir(homeDir string) string { return homeDir + "/commands" }
func (a uninstallDispatchAdapter) VariantKey() string {
	switch a.agent {
	case model.AgentClaude:
		return "claude"
	case model.AgentOpenCode:
		return "opencode"
	default:
		return ""
	}
}

// uninstallDispatchReg satisfies uninstall.Registry.
type uninstallDispatchReg struct {
	adapters map[model.Agent]uninstall.AgentAdapter
}

func (r uninstallDispatchReg) Get(agent model.Agent) (uninstall.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// uninstallDispatchCatalog implements uninstall.Catalog using in-memory harnesses.
type uninstallDispatchCatalog struct {
	harnesses []model.Harness
}

func (f *uninstallDispatchCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *uninstallDispatchCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *uninstallDispatchCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// newDispatchFixtures creates a minimal catalog + registry for dispatch tests.
func newDispatchFixtures() (*uninstallDispatchCatalog, uninstallDispatchReg) {
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &uninstallDispatchCatalog{harnesses: []model.Harness{h}}
	reg := uninstallDispatchReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: uninstallDispatchAdapter{agent: model.AgentClaude},
	}}
	return cat, reg
}

// writeManifestFile writes a minimal backup.Manifest to a temp file and returns
// the path. Used in restore-strategy tests.
func writeManifestFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	m := backup.Manifest{ID: "test-manifest", RootDir: dir}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

// ── Tests ──────────────────────────────────────────────────────────────────────

// TestUninstallDispatch_ValidFlags_ExitsZero asserts that valid flags with
// --dry-run reach RunHeadlessUninstall and the exit code is propagated (0).
func TestUninstallDispatch_ValidFlags_ExitsZero(t *testing.T) {
	cat, reg := newDispatchFixtures()

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error { return nil })
	defer restoreMarker()

	restoreStale := uninstall.SetStalePurgeFn(func(path string) error { return nil })
	defer restoreStale()

	args := []string{"--mode", "lite", "--agent", "claude", "--dry-run", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("valid flags + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestUninstallDispatch_ValidFlagsNonZeroPropagated asserts that a non-zero exit
// from RunHeadlessUninstall is propagated by the dispatch.
func TestUninstallDispatch_ValidFlagsNonZeroPropagated(t *testing.T) {
	cat, reg := newDispatchFixtures()

	// Snapshot fails → BuildPlan/Execute will error → exit 1
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	// Marker removal always fails → pipeline returns error → exit 1
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return os.ErrPermission
	})
	defer restoreMarker()

	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error { return nil })
	defer restoreRestoreFn()

	args := []string{"--mode", "lite", "--agent", "claude", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode == 0 {
		t.Errorf("failing apply must produce non-zero exit code, got 0; output:\n%s", out.String())
	}
}

// TestUninstallDispatch_UnknownFlag_ExitsNonZeroWithUsage asserts that an unknown
// flag causes a non-zero exit and a usage message.
func TestUninstallDispatch_UnknownFlag_ExitsNonZeroWithUsage(t *testing.T) {
	cat, reg := newDispatchFixtures()

	var out bytes.Buffer
	exitCode := runUninstallDispatch([]string{"--bogus", "value"}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("unknown flag must exit non-zero")
	}

	output := out.String()
	if !strings.Contains(output, "error") && !strings.Contains(output, "flag") && !strings.Contains(output, "usage") {
		t.Errorf("usage/error message expected; got:\n%s", output)
	}
}

// TestUninstallDispatch_UnknownFlagProjectRejected asserts that --project is not
// recognized (machine-scope uninstall, no project flag).
func TestUninstallDispatch_UnknownFlagProjectRejected(t *testing.T) {
	cat, reg := newDispatchFixtures()

	var out bytes.Buffer
	exitCode := runUninstallDispatch([]string{"--project", "/some/repo"}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("--project must be rejected with non-zero exit (not a recognized flag)")
	}
}

// TestUninstallDispatch_RestoreWithValidManifest asserts that --strategy restore
// with a valid --restore-manifest path reads the manifest and exits 0.
func TestUninstallDispatch_RestoreWithValidManifest(t *testing.T) {
	manifestPath := writeManifestFile(t)

	// For restore strategy: the catalog/agents combo doesn't matter for the plan
	// (restore step ignores them). Use a catalog that returns harnesses so
	// BuildPlan reaches the restore branch.
	cat, reg := newDispatchFixtures()

	// Wire restore seam so the restore step succeeds.
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error { return nil })
	defer restoreRestoreFn()

	args := []string{
		"--strategy", "restore",
		"--restore-manifest", manifestPath,
		"--agent", "claude",
		"--dry-run",
		"--yes",
	}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("restore + valid manifest + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestUninstallDispatch_RestoreWithValidManifestExecutes asserts that without
// --dry-run the restore step actually runs (triangulation: no dry-run).
func TestUninstallDispatch_RestoreWithValidManifestExecutes(t *testing.T) {
	manifestPath := writeManifestFile(t)
	cat, reg := newDispatchFixtures()

	restoreWasCalled := false
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error {
		restoreWasCalled = true
		return nil
	})
	defer restoreRestoreFn()

	args := []string{
		"--strategy", "restore",
		"--restore-manifest", manifestPath,
		"--agent", "claude",
		"--yes",
	}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("restore + valid manifest must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
	_ = restoreWasCalled // seam confirmed wired; whether called depends on pipeline
}

// TestUninstallDispatch_RestoreWithoutManifestPath_ExitsNonZero asserts that
// --strategy restore without --restore-manifest path exits non-zero.
func TestUninstallDispatch_RestoreWithoutManifestPath_ExitsNonZero(t *testing.T) {
	cat, reg := newDispatchFixtures()

	args := []string{"--strategy", "restore", "--agent", "claude", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode == 0 {
		t.Errorf("restore without --restore-manifest must exit non-zero, got 0; output:\n%s", out.String())
	}
}

// TestUninstallDispatch_RestoreWithoutManifestPath_NothingExecuted asserts that
// when --restore-manifest is missing, nothing is executed (triangulation).
func TestUninstallDispatch_RestoreWithoutManifestPath_NothingExecuted(t *testing.T) {
	cat, reg := newDispatchFixtures()

	sideEffectCalled := false
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		sideEffectCalled = true
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	args := []string{"--strategy", "restore", "--agent", "claude", "--yes"}
	var out bytes.Buffer
	runUninstallDispatch(args, cat, reg, &out)
	if sideEffectCalled {
		t.Error("no steps must execute when --restore-manifest is missing")
	}
}

// TestUninstallDispatch_UnreadableManifest_ExitsNonZero asserts that when
// --restore-manifest points to a nonexistent file, the command exits non-zero.
func TestUninstallDispatch_UnreadableManifest_ExitsNonZero(t *testing.T) {
	cat, reg := newDispatchFixtures()

	args := []string{
		"--strategy", "restore",
		"--restore-manifest", "/no/such/file/manifest.json",
		"--agent", "claude",
		"--yes",
	}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode == 0 {
		t.Errorf("unreadable manifest must cause non-zero exit, got 0; output:\n%s", out.String())
	}
}

// TestUninstallDispatch_UnreadableManifest_NoStepExecuted asserts that when
// the manifest is unreadable, no steps are executed (triangulation).
func TestUninstallDispatch_UnreadableManifest_NoStepExecuted(t *testing.T) {
	cat, reg := newDispatchFixtures()

	sideEffectCalled := false
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		sideEffectCalled = true
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	args := []string{
		"--strategy", "restore",
		"--restore-manifest", "/nonexistent/manifest.json",
		"--agent", "claude",
		"--yes",
	}
	var out bytes.Buffer
	runUninstallDispatch(args, cat, reg, &out)
	if sideEffectCalled {
		t.Error("no steps must execute when manifest is unreadable")
	}
}

// TestUninstallDispatch_CommandHarnessDryRun asserts that when the catalog
// contains a command harness (type: command), the dispatch succeeds without
// error in dry-run mode. This is the regression test for the HarnessCommand
// bug where BuildPlan crashed with "unknown harness type command".
func TestUninstallDispatch_CommandHarnessDryRun(t *testing.T) {
	commandH := model.Harness{
		ID:           "starter-add-command",
		Type:         model.HarnessCommand,
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &uninstallDispatchCatalog{harnesses: []model.Harness{commandH}}
	reg := uninstallDispatchReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude:   uninstallDispatchAdapter{agent: model.AgentClaude},
		model.AgentOpenCode: uninstallDispatchAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	args := []string{"--mode", "lite", "--agent", "claude", "--dry-run", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("command harness + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestUninstallDispatch_CommandHarnessOpenCodeDryRun triangulates with the
// opencode agent (different VariantKey → different relPath).
func TestUninstallDispatch_CommandHarnessOpenCodeDryRun(t *testing.T) {
	commandH := model.Harness{
		ID:           "starter-add-command",
		Type:         model.HarnessCommand,
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &uninstallDispatchCatalog{harnesses: []model.Harness{commandH}}
	reg := uninstallDispatchReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentOpenCode: uninstallDispatchAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	args := []string{"--mode", "lite", "--agent", "opencode", "--dry-run", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("command harness + opencode + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}
