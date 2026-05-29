package headless_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// TestGate_MissingDep_AbortsBeforePipeline verifies that when a required
// dependency for the selected harnesses is absent, RunHeadless:
//   - returns exit code 1
//   - prints the install hint / missing dep info
//   - does NOT execute the pipeline (extInstallFn is never called)
func TestGate_MissingDep_AbortsBeforePipeline(t *testing.T) {
	// Harness that needs npm (so RequiredDependencies → node+npm).
	h := model.Harness{
		ID:           "ext-npm",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	// Inject a fake detector: reports npm as missing.
	restore := headless.SetDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		return system.DependencyReport{
			Dependencies:    deps,
			AllPresent:      false,
			MissingRequired: []string{"npm"},
		}
	})
	defer restore()

	// Track whether the pipeline external installer is called.
	pipelineCalled := false
	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		pipelineCalled = true
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context, _ model.Harness, _ system.PlatformProfile,
		_ []extinstaller.AgentAdapter, _ string,
	) (extinstaller.Result, error) {
		pipelineCalled = true
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

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

	if exitCode == 0 {
		t.Fatalf("missing dep: want exit code != 0, got 0; output:\n%s", out.String())
	}
	if pipelineCalled {
		t.Fatalf("missing dep: pipeline must NOT be called when gate aborts")
	}

	output := out.String()
	// The report must mention the missing dep.
	if !strings.Contains(output, "npm") {
		t.Errorf("output must mention missing dep 'npm'; got:\n%s", output)
	}
}

// TestGate_AllDepsPresent_ProceedsNormally verifies that when all required
// deps are present the gate passes and execution continues normally.
func TestGate_AllDepsPresent_ProceedsNormally(t *testing.T) {
	h := model.Harness{
		ID:           "ext-npm",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	// Inject a fake detector: all deps present.
	restore := headless.SetDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		return system.DependencyReport{
			Dependencies: deps,
			AllPresent:   true,
		}
	})
	defer restore()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context, _ model.Harness, _ system.PlatformProfile,
		_ []extinstaller.AgentAdapter, _ string,
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
		t.Fatalf("all deps present: want exit code 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestGate_ConfigOnlyHarness_NoDepsRequired verifies that a config-only harness
// does not demand node/npm — the gate passes even with fake-missing node/npm.
func TestGate_ConfigOnlyHarness_NoDepsRequired(t *testing.T) {
	configH := model.Harness{
		ID:           "cfg-h",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{configH}}
	homeDir := t.TempDir()
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	// Detector should NOT be called at all (RequiredDependencies returns []).
	// If it IS called with non-empty deps, it reports everything missing — this
	// would catch a false gate.
	restore := headless.SetDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		if len(deps) > 0 {
			// Fail all deps — if gate incorrectly passes deps here, test will catch it.
			missing := make([]string, len(deps))
			for i, d := range deps {
				missing[i] = d.Name
			}
			return system.DependencyReport{
				Dependencies:    deps,
				AllPresent:      false,
				MissingRequired: missing,
			}
		}
		return system.DependencyReport{AllPresent: true}
	})
	defer restore()

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
		t.Fatalf("config-only: gate must NOT abort, got exit %d; output:\n%s", exitCode, out.String())
	}
}

// TestGate_DryRun_DoesNotRunGate verifies that --dry-run returns early before the
// gate runs (dry-run never executes, so detecting deps is unnecessary).
func TestGate_DryRun_DoesNotRunGate(t *testing.T) {
	h := model.Harness{
		ID:           "ext-npm",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	// Detector that always reports everything missing — if gate runs during dry-run,
	// it would abort and exit != 0, which would fail the test.
	restore := headless.SetDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		missing := make([]string, len(deps))
		for i, d := range deps {
			missing[i] = d.Name
		}
		return system.DependencyReport{
			Dependencies:    deps,
			AllPresent:      len(deps) == 0,
			MissingRequired: missing,
		}
	})
	defer restore()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

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
	exitCode := headless.RunHeadless(params, cat, reg, &out)

	if exitCode != 0 {
		t.Fatalf("dry-run: want exit 0 regardless of missing deps, got %d; output:\n%s", exitCode, out.String())
	}
}
