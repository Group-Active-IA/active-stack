// Package main — TUI wiring tests for the new ModelDeps fields (tui-menu-hub task 6.2).
// Tests use fakes and stubs to verify the callback wiring without touching the real environment.
package main

import (
	"bytes"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestTUIWiring_RunUninstallCallback verifies that buildRunUninstallCallback
// creates a callback that correctly delegates to RunHeadlessUninstall.
// We use --dry-run to avoid side effects.
func TestTUIWiring_RunUninstallCallback(t *testing.T) {
	cat, reg := newDispatchFixtures()

	// Wire test seams to prevent filesystem side-effects.
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error { return nil })
	defer restoreMarker()
	restoreStale := uninstall.SetStalePurgeFn(func(path string) error { return nil })
	defer restoreStale()

	runUninstall := buildRunUninstallCallback(cat, reg)

	flags := headless.ParsedUninstallFlags{
		DryRun: true,
		Yes:    true,
		Intent: uninstall.Intent{
			Mode:     model.ModeLite,
			Agents:   []model.Agent{model.AgentClaude},
			Strategy: uninstall.StrategyTargeted,
		},
	}

	var buf bytes.Buffer
	exitCode := runUninstall(flags, &buf)
	if exitCode != 0 {
		t.Errorf("RunUninstall dry-run must exit 0, got %d; output:\n%s", exitCode, buf.String())
	}
}

// TestTUIWiring_RunUninstallCallback_PropagatesNonZero verifies that a failing
// RunHeadlessUninstall propagates a non-zero exit code.
func TestTUIWiring_RunUninstallCallback_PropagatesNonZero(t *testing.T) {
	cat, reg := newDispatchFixtures()

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return errPermission // always fail
	})
	defer restoreMarker()
	restoreRestoreFn := uninstall.SetRestoreFn(func(m backup.Manifest) error { return nil })
	defer restoreRestoreFn()

	runUninstall := buildRunUninstallCallback(cat, reg)

	flags := headless.ParsedUninstallFlags{
		DryRun: false,
		Yes:    true,
		Intent: uninstall.Intent{
			Mode:     model.ModeLite,
			Agents:   []model.Agent{model.AgentClaude},
			Strategy: uninstall.StrategyTargeted,
		},
	}
	var buf bytes.Buffer
	exitCode := runUninstall(flags, &buf)
	if exitCode == 0 {
		t.Error("failing uninstall must produce non-zero exit code")
	}
}

// TestTUIWiring_RunStarterCallback_UnknownStarter verifies that calling RunStarter
// with an unknown starter ID returns exit code 1.
func TestTUIWiring_RunStarterCallback_UnknownStarter(t *testing.T) {
	// starterCatalog with one known starter (not "unknown-id").
	fakeSC := &tuiWiringStarterCatalog{knownID: "known-starter"}
	regWrapper := agentRegistryAdapter{r: nil} // nil registry: headless will handle it

	runStarter := buildRunStarterCallback(fakeSC, regWrapper)

	var buf bytes.Buffer
	exitCode := runStarter("unknown-starter-id", t.TempDir(), []model.Agent{model.AgentClaude}, &buf)
	if exitCode == 0 {
		t.Error("unknown starter must produce exit code 1")
	}
	output := buf.String()
	if !wiringContains(output, "unknown-starter-id") {
		t.Errorf("error output should mention the starter ID; got:\n%s", output)
	}
}

// wiringContains is a simple substring helper for this test file.
func wiringContains(s, sub string) bool {
	if len(s) < len(sub) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ── Helpers and fakes ─────────────────────────────────────────────────────────

// errPermission is a fake error for seam tests.
var errPermission = &wiringPermissionError{}

type wiringPermissionError struct{}

func (e *wiringPermissionError) Error() string { return "permission denied" }

// tuiWiringStarterCatalog implements the starterCatalog interface with a
// single known starter, for wiring tests.
type tuiWiringStarterCatalog struct {
	knownID string
	*uninstallDispatchCatalog
}

func (f *tuiWiringStarterCatalog) StarterByID(id string) (model.Starter, bool) {
	if id == f.knownID {
		return model.Starter{ID: id, Name: id}, true
	}
	return model.Starter{}, false
}

func (f *tuiWiringStarterCatalog) AllStarters() []model.Starter {
	return []model.Starter{{ID: f.knownID, Name: f.knownID}}
}

func (f *tuiWiringStarterCatalog) ResolveStarter(id string) ([]model.Harness, error) {
	return nil, nil
}

func (f *tuiWiringStarterCatalog) ResolveStarterMCPs(id string) ([]model.MCP, error) {
	return nil, nil
}

func (f *tuiWiringStarterCatalog) AllHarnesses() []model.Harness {
	return nil
}

func (f *tuiWiringStarterCatalog) ForAgent(a model.Agent) []model.Harness {
	return nil
}
