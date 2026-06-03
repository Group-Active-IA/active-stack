package uninstall_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestUninstall_PrimaryAgent_RemovesKeyFromSettings verifies that uninstalling a
// config harness delivered as a primary agent removes the agent.<id> key from
// the settings JSON while preserving all other user config. This is the mirror
// of the install-time primary-agent registration: install writes opencode.json,
// so uninstall must clean opencode.json (not just AGENTS.md).
func TestUninstall_PrimaryAgent_RemovesKeyFromSettings(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{
		agent:    model.AgentOpenCode,
		homeDir:  homeDir,
		delivery: model.ConfigDeliveryPrimaryAgent,
	}

	settingsPath := adapter.SettingsPath(homeDir)
	seed := `{
  "theme": "tokyonight",
  "agent": {
    "build": { "mode": "primary" },
    "sdd-orchestrator": { "mode": "primary", "prompt": "orchestrator stuff" }
  }
}`
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatalf("setup settings: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentOpenCode},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentOpenCode: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("settings invalid after uninstall: %v\n%s", err, raw)
	}

	agentSection, _ := root["agent"].(map[string]any)
	if agentSection == nil {
		t.Fatalf("agent section vanished entirely: %s", raw)
	}
	if _, present := agentSection["sdd-orchestrator"]; present {
		t.Errorf("agent.sdd-orchestrator must be removed on uninstall:\n%s", raw)
	}
	if _, present := agentSection["build"]; !present {
		t.Errorf("agent.build (user config) must be preserved:\n%s", raw)
	}
	if root["theme"] != "tokyonight" {
		t.Errorf("theme (user config) must be preserved:\n%s", raw)
	}
}

// runPrimaryAgentUninstall is a small harness that builds and runs a targeted
// uninstall plan for a single primary-agent config harness against homeDir.
func runPrimaryAgentUninstall(t *testing.T, homeDir string) {
	t.Helper()
	adapter := fakeAdapter{agent: model.AgentOpenCode, homeDir: homeDir, delivery: model.ConfigDeliveryPrimaryAgent}
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}}
	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{Agents: []model.Agent{model.AgentOpenCode}, Mode: model.ModeLite, Strategy: uninstall.StrategyTargeted},
		uninstall.Options{HomeDir: homeDir, Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentOpenCode: adapter}}},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}
}

// TestUninstall_PrimaryAgent_DropsEmptyAgentObject verifies that when the
// orchestrator was the only entry, the now-empty "agent" object is dropped
// entirely so no orphaned {} footprint remains.
func TestUninstall_PrimaryAgent_DropsEmptyAgentObject(t *testing.T) {
	homeDir := t.TempDir()
	settingsPath := homeDir + "/settings.json"
	if err := os.WriteFile(settingsPath, []byte(`{"agent":{"sdd-orchestrator":{"mode":"primary"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	runPrimaryAgentUninstall(t, homeDir)

	raw, _ := os.ReadFile(settingsPath)
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("settings invalid: %v\n%s", err, raw)
	}
	if _, present := root["agent"]; present {
		t.Errorf("empty agent object should be dropped, got:\n%s", raw)
	}
}

// TestUninstall_PrimaryAgent_AbsentKeyIsNoop verifies uninstall is a no-op when
// the orchestrator key is not present (e.g. already uninstalled), leaving user
// config untouched.
func TestUninstall_PrimaryAgent_AbsentKeyIsNoop(t *testing.T) {
	homeDir := t.TempDir()
	settingsPath := homeDir + "/settings.json"
	const seed = `{"theme":"gruvbox","agent":{"build":{"mode":"primary"}}}`
	if err := os.WriteFile(settingsPath, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	runPrimaryAgentUninstall(t, homeDir)

	raw, _ := os.ReadFile(settingsPath)
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("settings invalid: %v\n%s", err, raw)
	}
	if root["theme"] != "gruvbox" {
		t.Errorf("theme must be preserved on no-op uninstall:\n%s", raw)
	}
	agentSection, _ := root["agent"].(map[string]any)
	if _, present := agentSection["build"]; !present {
		t.Errorf("agent.build must be preserved on no-op uninstall:\n%s", raw)
	}
}
