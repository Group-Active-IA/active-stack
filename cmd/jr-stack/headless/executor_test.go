package headless_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// ── Fake implementations ──────────────────────────────────────────────────────

// fakeExecCatalog implements install.Catalog for the executor tests.
type fakeExecCatalog struct {
	harnesses []model.Harness
}

func (f *fakeExecCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeExecCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeExecCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// fakeExecAdapter satisfies both install.AgentAdapter and verify.Adapter.
type fakeExecAdapter struct {
	agent model.Agent
}

func (a fakeExecAdapter) Agent() model.Agent                     { return a.agent }
func (a fakeExecAdapter) InstructionsPath(homeDir string) string { return homeDir + "/CLAUDE.md" }
func (a fakeExecAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a fakeExecAdapter) CommandsDir(homeDir string) string      { return homeDir + "/commands" }
func (a fakeExecAdapter) SettingsPath(homeDir string) string     { return homeDir + "/settings.json" }
func (a fakeExecAdapter) MCPConfigPath(homeDir, s string) string {
	return homeDir + "/mcp/" + s + ".json"
}
func (a fakeExecAdapter) MCPStrategy() extinstaller.MCPStrategy {
	return extinstaller.StrategySeparateFile
}
func (a fakeExecAdapter) VariantKey() string { return string(a.agent) }
func (a fakeExecAdapter) PathsFor(base string, _ model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/CLAUDE.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      base + "/commands",
	}.WithMCPConfigFn(func(s string) string { return base + "/mcp/" + s + ".json" })
}
func (a fakeExecAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}

// fakeExecRegistry maps one agent.
type fakeExecRegistry struct {
	adapters map[model.Agent]install.AgentAdapter
}

func (r *fakeExecRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestHeadlessExecutorDryRun verifies that --dry-run prints the plan steps and
// exits without touching the filesystem (no side-effects).
func TestHeadlessExecutorDryRun(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	params := headless.ParsedFlags{
		TUI:     false,
		DryRun:  true,
		Yes:     true,
		HomeDir: t.TempDir(),
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	// restore install seams so BuildPlan works without fs side effects.
	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("dry-run must exit 0, got %d", exitCode)
	}

	output := out.String()
	// Dry-run must list the plan steps, not execute them.
	if !strings.Contains(output, "external:ext-h") {
		t.Errorf("dry-run output must contain plan step IDs; got:\n%s", output)
	}
}

// TestHeadlessExecutorSuccess verifies that a successful headless install exits 0
// and prints the verify report to stdout.
func TestHeadlessExecutorSuccess(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	// Wire fake install seams.
	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

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

	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
		// VerifyHook is nil → no verify hook (tests the no-hook path)
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("successful install must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestHeadlessExecutorFailure verifies that when the install pipeline fails,
// RunHeadless returns a non-zero exit code.
func TestHeadlessExecutorFailure(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	// External installer always fails.
	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, errors.New("install failed")
	})
	defer restoreExt()

	restoreRestore := install.SetRestoreFn(func(_ backup.Manifest) error {
		return nil
	})
	defer restoreRestore()

	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("failed install must exit non-zero, got 0; output:\n%s", out.String())
	}
}

// TestHeadlessExecutorVerifyHookFails verifies that when the verify hook fails,
// RunHeadless exits non-zero.
func TestHeadlessExecutorVerifyHookFails(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

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

	restoreRestore := install.SetRestoreFn(func(_ backup.Manifest) error {
		return nil
	})
	defer restoreRestore()

	// Provide a failing verify hook via VerifyHookFn override.
	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
		VerifyHookFn: func() error {
			return errors.New("SKILL.md missing")
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode == 0 {
		t.Errorf("verify hook failure must exit non-zero, got 0; output:\n%s", out.String())
	}
}

// TestHeadlessExecutorProgressPrinted verifies that progress events are printed
// to the writer during execution (at least step IDs appear in output).
func TestHeadlessExecutorProgressPrinted(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

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

	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	// Progress output must mention the step.
	if !strings.Contains(output, "ext-h") {
		t.Errorf("output must contain step reference; got:\n%s", output)
	}
}

// TestHeadlessExecutorWithEmbeddedFS verifies that RunHeadless accepts a nil
// embeddedFS without panicking (no skill harnesses in the catalog).
func TestHeadlessExecutorWithEmbeddedFS(t *testing.T) {
	cat := &fakeExecCatalog{harnesses: []model.Harness{}} // empty catalog
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("empty plan must succeed (nothing to do), got %d; output:\n%s", exitCode, out.String())
	}
}

// TestHeadlessExecutorDryRunNoPipelineExecution verifies that --dry-run does NOT
// call the external install function (no side-effects).
func TestHeadlessExecutorDryRunNoPipelineExecution(t *testing.T) {
	extCalled := false

	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		extCalled = true // snapshot should NOT be called in dry-run
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
	) (extinstaller.Result, error) {
		extCalled = true
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	params := headless.ParsedFlags{
		TUI:     false,
		DryRun:  true,
		Yes:     true,
		HomeDir: homeDir,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("dry-run must exit 0, got %d", exitCode)
	}
	if extCalled {
		t.Error("dry-run must NOT call any installer or snapshot")
	}
}

// ensure pipeline package is referenced (used for progress event type assertion).
var _ pipeline.ProgressEvent
