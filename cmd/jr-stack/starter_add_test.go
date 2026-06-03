// Package main — tests for C-29 runStarterAdd handler (Task 5.1 RED).
package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/catalog"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ── test helpers ─────────────────────────────────────────────────────────────

// starterAddTestRegistry satisfies install.Registry and agentRegistryIface for
// handler tests. Re-uses fakeExecAdapter from executor_test (same package-local
// fake, but executor_test is in package headless_test — we need one here in main).
type starterAddTestAdapter struct{ agent model.Agent }

func (a starterAddTestAdapter) Agent() model.Agent             { return a.agent }
func (a starterAddTestAdapter) InstructionsPath(d string) string { return d + "/CLAUDE.md" }
func (a starterAddTestAdapter) SkillsDir(d string) string        { return d + "/skills" }
func (a starterAddTestAdapter) CommandsDir(d string) string      { return d + "/commands" }
func (a starterAddTestAdapter) SettingsPath(d string) string     { return d + "/settings.json" }
func (a starterAddTestAdapter) MCPConfigPath(d, s string) string { return d + "/mcp/" + s + ".json" }
func (a starterAddTestAdapter) MCPStrategy() extinstaller.MCPStrategy {
	return extinstaller.StrategySeparateFile
}
func (a starterAddTestAdapter) VariantKey() string               { return string(a.agent) }
func (a starterAddTestAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a starterAddTestAdapter) PathsFor(base string, _ model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/CLAUDE.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      base + "/commands",
	}.WithMCPConfigFn(func(s string) string { return base + "/mcp/" + s + ".json" })
}

// starterAddTestReg satisfies the install.Registry interface.
type starterAddTestReg struct {
	adapters map[model.Agent]install.AgentAdapter
}

func (r starterAddTestReg) Get(agent model.Agent) (install.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// ─────────────────────────────────────────────────────────────────────────────

// TestRunStarterAdd_KnownID_BuildsPlanWithProjectTarget asserts that runStarterAdd
// with a known starter id builds a plan with Target=Project and ProjectRoot set,
// and exits 0 (dry-run, no writes).
//
// RED: fails because runStarterAdd does not exist yet.
func TestRunStarterAdd_KnownID_BuildsPlanWithProjectTarget(t *testing.T) {
	projectRoot := t.TempDir()

	// Load the real embedded catalog to get a real starter.
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude:   starterAddTestAdapter{agent: model.AgentClaude},
		model.AgentOpenCode: starterAddTestAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	// Capture the options passed to BuildPlan.
	var capturedOpts install.Options
	buildPlanFn := func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
		capturedOpts = opts
		return install.BuildPlan(c, intent, opts)
	}

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "active-ia",
		ProjectPath: projectRoot,
		DryRun:      true, // safe: no filesystem writes
		Yes:         true,
		Agents:      []model.Agent{model.AgentClaude, model.AgentOpenCode},
	}

	var out bytes.Buffer
	exitCode := runStarterAdd(flags, cat, reg, buildPlanFn, &out)
	if exitCode != 0 {
		t.Fatalf("runStarterAdd() exit = %d; output:\n%s", exitCode, out.String())
	}

	// Target must be Project.
	if capturedOpts.Target != model.Project {
		t.Errorf("Options.Target = %v, want model.Project", capturedOpts.Target)
	}
	// ProjectRoot must be the supplied project root.
	if capturedOpts.ProjectRoot != projectRoot {
		t.Errorf("Options.ProjectRoot = %q, want %q", capturedOpts.ProjectRoot, projectRoot)
	}
	// Starter must be non-nil and match the requested id.
	if capturedOpts.Starter == nil {
		t.Fatal("Options.Starter is nil, want non-nil")
	}
}

// TestRunStarterAdd_UnknownID_ExitsNonZeroWithList asserts that an unknown starter
// id exits non-zero and prints a list of available starters.
//
// RED: fails because runStarterAdd does not exist yet.
func TestRunStarterAdd_UnknownID_ExitsNonZeroWithList(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "does-not-exist",
		ProjectPath: t.TempDir(),
		Agents:      []model.Agent{model.AgentClaude, model.AgentOpenCode},
	}

	var out bytes.Buffer
	exitCode := runStarterAdd(flags, cat, reg, nil, &out)
	if exitCode == 0 {
		t.Fatal("expected non-zero exit for unknown starter id, got 0")
	}

	output := out.String()
	// Must mention the unknown id.
	if !strings.Contains(output, "does-not-exist") {
		t.Errorf("output must mention the unknown id; got:\n%s", output)
	}
	// Must list available starters.
	for _, id := range []string{"active-ia", "ux-ui", "backend"} {
		if !strings.Contains(output, id) {
			t.Errorf("output must list available starter %q; got:\n%s", id, output)
		}
	}
}

// TestRunStarterAdd_DryRun_WritesNothing asserts that --dry-run exits 0 and
// does not create any files under projectRoot.
//
// RED: fails because runStarterAdd does not exist yet.
func TestRunStarterAdd_DryRun_WritesNothing(t *testing.T) {
	projectRoot := t.TempDir()

	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude:   starterAddTestAdapter{agent: model.AgentClaude},
		model.AgentOpenCode: starterAddTestAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "ux-ui",
		ProjectPath: projectRoot,
		DryRun:      true,
		Yes:         true,
		Agents:      []model.Agent{model.AgentClaude, model.AgentOpenCode},
	}

	var out bytes.Buffer
	exitCode := runStarterAdd(flags, cat, reg, nil, &out)
	if exitCode != 0 {
		t.Fatalf("dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}
