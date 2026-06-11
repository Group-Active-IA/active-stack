// Package headless — tests for C-29 RunHeadless with Target/ProjectRoot/Starter
// (Task 4.1 RED). These tests assert that:
//   - RunHeadless passes Target=Project, ProjectRoot, and Starter into install.Options.
//   - The dry-run path does not execute (no filesystem writes).
//   - The install (Machine) path is unaffected (zero-regression).
package headless_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestRunHeadless_StarterAdd_ProjectTarget asserts that when ParsedFlags carries
// Target=Project, ProjectRoot, and a non-nil Starter, RunHeadless passes those
// fields into the install.Options that BuildPlanFn receives.
//
// RED: fails because ParsedFlags does not have Target/ProjectRoot/Starter fields yet.
func TestRunHeadless_StarterAdd_ProjectTarget(t *testing.T) {
	projectRoot := t.TempDir()

	starter := &model.Starter{
		ID:   "test-starter",
		Name: "Test Starter",
	}

	// Capture the options that BuildPlanFn receives.
	var capturedOpts install.Options

	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:         false,
		Yes:         true,
		HomeDir:     t.TempDir(),
		Target:      model.Project,
		ProjectRoot: projectRoot,
		Starter:     starter,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeCustom,
			Custom: []string{},
		},
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			capturedOpts = opts
			return install.BuildPlan(c, intent, opts)
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("RunHeadless exited %d; output:\n%s", exitCode, out.String())
	}

	// Assert Target and ProjectRoot were threaded into Options.
	if capturedOpts.Target != model.Project {
		t.Errorf("Options.Target = %v, want %v", capturedOpts.Target, model.Project)
	}
	if capturedOpts.ProjectRoot != projectRoot {
		t.Errorf("Options.ProjectRoot = %q, want %q", capturedOpts.ProjectRoot, projectRoot)
	}
	if capturedOpts.Starter != starter {
		t.Errorf("Options.Starter = %v, want the injected starter", capturedOpts.Starter)
	}
}

// TestRunHeadless_StarterAdd_DryRunNoSideEffects asserts that dry-run with
// Target=Project writes nothing.
//
// RED: fails because ParsedFlags does not have Target/ProjectRoot/Starter fields yet.
func TestRunHeadless_StarterAdd_DryRunNoSideEffects(t *testing.T) {
	projectRoot := t.TempDir()
	buildPlanCalled := false

	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	params := headless.ParsedFlags{
		TUI:         false,
		DryRun:      true,
		Yes:         true,
		HomeDir:     t.TempDir(),
		Target:      model.Project,
		ProjectRoot: projectRoot,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeCustom,
		},
		BuildPlanFn: func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
			buildPlanCalled = true
			return install.BuildPlan(c, intent, opts)
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
	_ = buildPlanCalled

	output := out.String()
	if !strings.Contains(output, "Dry-run") {
		t.Errorf("dry-run output must mention 'Dry-run'; got:\n%s", output)
	}
}

// TestRunHeadless_MachineTarget_ZeroRegression asserts that the existing
// Machine-target install path is unaffected when Target/ProjectRoot/Starter
// are at their zero values (nil/empty/"").
//
// RED: fails only if extending ParsedFlags breaks existing callers.
func TestRunHeadless_MachineTarget_ZeroRegression(t *testing.T) {
	cat := &fakeExecCatalog{harnesses: []model.Harness{}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	// No Target/ProjectRoot/Starter set → zero values → Machine behavior.
	params := headless.ParsedFlags{
		TUI:     false,
		Yes:     true,
		HomeDir: t.TempDir(),
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunHeadless(params, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("machine-target zero-regression must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}
