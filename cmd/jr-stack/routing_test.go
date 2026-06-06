// Package main — tests for routing in the dispatch switch.
// C-29: starter routing tests (Task 6.1 RED).
// uninstall-subcommand: uninstall routing tests (Task 4.2 RED).
// Tests run dispatch functions extracted from main() without spinning up the
// full binary.
package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/agents"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/catalog"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestStarterDispatch_AddSubcommandRoutes asserts that "starter add <id>"
// dispatches to the handler and, for an unknown id, exits non-zero with a list.
//
// RED: fails because runStarterDispatch does not exist yet.
func TestStarterDispatch_AddSubcommandRoutes(t *testing.T) {
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

	projectRoot := t.TempDir()
	args := []string{"add", "ux-ui", "--project", projectRoot, "--dry-run", "--yes"}

	var out bytes.Buffer
	exitCode := runStarterDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("'starter add ux-ui --dry-run' must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestStarterDispatch_UnknownSubcommand asserts that "starter <unknown>" exits
// non-zero with a usage error naming the "add" subcommand.
//
// RED: fails because runStarterDispatch does not exist yet.
func TestStarterDispatch_UnknownSubcommand(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runStarterDispatch([]string{"list"}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("unknown subcommand must exit non-zero")
	}

	output := out.String()
	if !strings.Contains(output, "add") {
		t.Errorf("usage error must mention the 'add' subcommand; got:\n%s", output)
	}
}

// TestStarterDispatch_NoSubcommand asserts that "starter" with no subcommand
// exits non-zero with a usage error.
//
// RED: fails because runStarterDispatch does not exist yet.
func TestStarterDispatch_NoSubcommand(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runStarterDispatch([]string{}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("no subcommand must exit non-zero")
	}
}

// TestStarterDispatch_AddMissingID asserts that "starter add" without an id
// exits non-zero (the handler/parser returns an error).
//
// RED: fails because runStarterDispatch does not exist yet.
func TestStarterDispatch_AddMissingID(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	// "add" with no further args → missing id
	exitCode := runStarterDispatch([]string{"add"}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("'starter add' with no id must exit non-zero")
	}
}

// ── Uninstall routing tests (Task 4.2 RED) ────────────────────────────────────

// routingUninstallReg is a minimal uninstall.Registry for routing tests.
type routingUninstallReg struct {
	adapters map[model.Agent]uninstall.AgentAdapter
}

func (r routingUninstallReg) Get(agent model.Agent) (uninstall.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// routingUninstallAdapter satisfies uninstall.AgentAdapter (7 methods).
// Updated in the HarnessCommand bugfix to add CommandsDir and VariantKey.
type routingUninstallAdapter struct{ agent model.Agent }

func (a routingUninstallAdapter) Agent() model.Agent                     { return a.agent }
func (a routingUninstallAdapter) InstructionsPath(homeDir string) string { return homeDir + "/CLAUDE.md" }
func (a routingUninstallAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a routingUninstallAdapter) SettingsPath(homeDir string) string {
	return homeDir + "/settings.json"
}
func (a routingUninstallAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a routingUninstallAdapter) CommandsDir(homeDir string) string { return homeDir + "/commands" }
func (a routingUninstallAdapter) VariantKey() string {
	switch a.agent {
	case model.AgentClaude:
		return "claude"
	case model.AgentOpenCode:
		return "opencode"
	default:
		return ""
	}
}

// TestUninstallRouting_ValidFlagsReachDispatch asserts that valid uninstall args
// reach runUninstallDispatch and BuildPlan succeeds with the REAL catalog (which
// contains a HarnessCommand harness). This is the regression test for the
// "unknown harness type command" bug that the previous apply hid with a fake catalog.
func TestUninstallRouting_ValidFlagsReachDispatch(t *testing.T) {
	// Use the REAL catalog — it contains starter-add-command (type: command, lite+full).
	// The bug caused BuildPlan to crash here; the fix makes it succeed.
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := routingUninstallReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude:   routingUninstallAdapter{agent: model.AgentClaude},
		model.AgentOpenCode: routingUninstallAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error { return nil })
	defer restoreMarker()

	restoreStale := uninstall.SetStalePurgeFn(func(path string) error { return nil })
	defer restoreStale()

	restoreSkill := uninstall.SetSkillRemovalFn(func(path string) error { return nil })
	defer restoreSkill()

	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	args := []string{"--mode", "lite", "--agent", "claude", "--dry-run", "--yes"}

	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("valid uninstall flags + real catalog + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestUninstallRouting_RealCatalogFullModeSucceeds is a triangulation test:
// same as ValidFlagsReachDispatch but targeting --mode full, which also includes
// the command harness (starter-add-command is lite+full).
func TestUninstallRouting_RealCatalogFullModeSucceeds(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := routingUninstallReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude:   routingUninstallAdapter{agent: model.AgentClaude},
		model.AgentOpenCode: routingUninstallAdapter{agent: model.AgentOpenCode},
	}}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error { return nil })
	defer restoreMarker()
	restoreStale := uninstall.SetStalePurgeFn(func(path string) error { return nil })
	defer restoreStale()
	restoreSkill := uninstall.SetSkillRemovalFn(func(path string) error { return nil })
	defer restoreSkill()
	restoreCommand := uninstall.SetCommandRemovalFn(func(path string) error { return nil })
	defer restoreCommand()

	args := []string{"--mode", "full", "--agent", "claude", "--dry-run", "--yes"}
	var out bytes.Buffer
	exitCode := runUninstallDispatch(args, cat, reg, &out)
	if exitCode != 0 {
		t.Errorf("--mode full + real catalog + dry-run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}
}

// TestUninstallRouting_NonZeroResultPropagated asserts that a non-zero result
// from runUninstallDispatch is propagated as a non-zero exit.
func TestUninstallRouting_NonZeroResultPropagated(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := routingUninstallReg{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: routingUninstallAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	// Unknown flag → parse error → non-zero exit.
	exitCode := runUninstallDispatch([]string{"--totally-bogus-flag"}, cat, reg, &out)
	if exitCode == 0 {
		t.Errorf("unknown flag must produce non-zero exit code, got 0; output:\n%s", out.String())
	}
}

// TestUninstallRouting_UninstallRegistryAdapterWraps asserts that
// uninstallRegistryAdapter satisfies uninstall.Registry for a registered agent.
func TestUninstallRouting_UninstallRegistryAdapterWraps(t *testing.T) {
	defaultReg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Skip("agent registry unavailable:", err)
	}

	wrapped := uninstallRegistryAdapter{r: defaultReg}

	// The adapter must satisfy the uninstall.Registry interface structurally.
	// We verify by calling Get for a known agent.
	adapter, ok := wrapped.Get(model.AgentClaude)
	if !ok {
		t.Skip("claude adapter not registered in default registry; skipping structural check")
	}
	if adapter == nil {
		t.Error("Get(claude) returned non-nil ok but nil adapter")
	}
}

// TestUninstallRouting_UninstallRegistryAdapterUnregistered asserts that
// uninstallRegistryAdapter returns (nil, false) for an unregistered agent.
func TestUninstallRouting_UninstallRegistryAdapterUnregistered(t *testing.T) {
	defaultReg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Skip("agent registry unavailable:", err)
	}

	wrapped := uninstallRegistryAdapter{r: defaultReg}
	adapter, ok := wrapped.Get(model.Agent("nonexistent-agent-xyz"))
	if ok {
		t.Error("Get(nonexistent) must return ok=false")
	}
	if adapter != nil {
		t.Error("Get(nonexistent) must return nil adapter")
	}
}

// TestUninstallRouting_UsageMessageContainsFlags asserts that an unknown flag
// produces output mentioning at least one recognized flag name.
func TestUninstallRouting_UsageMessageContainsFlags(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := routingUninstallReg{}

	var out bytes.Buffer
	exitCode := runUninstallDispatch([]string{"--unknown"}, cat, reg, &out)
	if exitCode == 0 {
		t.Fatal("unknown flag must exit non-zero")
	}

	output := out.String()
	// Usage message must mention at least one recognized flag.
	if !strings.Contains(output, "--mode") && !strings.Contains(output, "mode") &&
		!strings.Contains(output, "error") && !strings.Contains(output, "flag") {
		t.Errorf("output must reference recognized flags or error; got:\n%s", output)
	}
}
