package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// opencodeAdapter is a fake adapter that delivers config harnesses as a
// primary agent registered in a settings JSON file (the OpenCode contract).
type opencodeAdapter struct {
	instr    string
	settings string
}

func (o opencodeAdapter) Agent() model.Agent                  { return model.AgentOpenCode }
func (o opencodeAdapter) InstructionsPath(_ string) string    { return o.instr }
func (o opencodeAdapter) VariantKey() string                  { return "opencode" }
func (o opencodeAdapter) SettingsPath(_ string) string        { return o.settings }
func (o opencodeAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryPrimaryAgent
}

const sddHarnessID = "sdd-orchestrator"

// readAgentEntry parses settingsPath and returns the agent.<id> sub-object.
func readAgentEntry(t *testing.T, settingsPath, id string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings %q: %v", settingsPath, err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("settings is not valid JSON: %v\n%s", err, raw)
	}
	agentSection, ok := root["agent"].(map[string]any)
	if !ok {
		t.Fatalf("settings has no \"agent\" object: %s", raw)
	}
	entry, ok := agentSection[id].(map[string]any)
	if !ok {
		t.Fatalf("settings has no agent.%s entry: %s", id, raw)
	}
	return entry
}

// TestInstall_PrimaryAgent_RegistersTabableAgent is the core regression test for
// the OpenCode tab bug: the orchestrator must be registered as a primary agent
// in opencode.json (mode:primary), NOT injected into the shared AGENTS.md.
func TestInstall_PrimaryAgent_RegistersTabableAgent(t *testing.T) {
	dir := t.TempDir()
	instrPath := filepath.Join(dir, "AGENTS.md")
	settingsPath := filepath.Join(dir, "opencode.json")

	// Pre-existing AGENTS.md carrying a STALE orchestrator section from an older
	// (buggy) install, plus user content that must survive.
	staleInstr := "# My global rules\n\nkeep me\n\n" +
		"<!-- jr-stack:sdd-orchestrator -->\nOLD orchestrator block\n<!-- /jr-stack:sdd-orchestrator -->\n"
	if err := os.WriteFile(instrPath, []byte(staleInstr), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-existing opencode.json with user config that must be preserved.
	if err := os.WriteFile(settingsPath, []byte(`{"theme":"tokyonight","agent":{"build":{"mode":"primary"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	h := model.Harness{
		ID:      sddHarnessID,
		Type:    model.HarnessConfig,
		Toggles: []string{"model-routing", "delegation"},
	}
	adapter := opencodeAdapter{instr: instrPath, settings: settingsPath}

	result, err := config.Install(h, []config.AgentAdapter{adapter}, dir)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if result.AllAlready {
		t.Error("first install should NOT be AllAlready")
	}

	// 1. The orchestrator is a primary agent in opencode.json.
	entry := readAgentEntry(t, settingsPath, sddHarnessID)
	if entry["mode"] != "primary" {
		t.Errorf("agent.%s mode = %v, want \"primary\" (must be tab-able)", sddHarnessID, entry["mode"])
	}
	prompt, _ := entry["prompt"].(string)
	if strings.TrimSpace(prompt) == "" {
		t.Error("agent.sdd-orchestrator must carry a non-empty prompt")
	}

	// 2. The prompt is the SAME composed content the instructions file would
	//    have received — same orchestrator, different home.
	wantComposed, err := config.Compose(h.Toggles, "opencode")
	if err != nil {
		t.Fatal(err)
	}
	if prompt != wantComposed {
		t.Errorf("prompt does not match composed orchestrator block.\n got: %q\nwant: %q", prompt, wantComposed)
	}

	// 3. User config preserved.
	raw, _ := os.ReadFile(settingsPath)
	cs := string(raw)
	if !strings.Contains(cs, "tokyonight") {
		t.Error("existing user setting (theme) was clobbered")
	}
	if !strings.Contains(cs, "\"build\"") {
		t.Error("existing agent.build entry was clobbered")
	}

	// 4. The stale AGENTS.md orchestrator section was purged (no leak into the
	//    shared instructions — this is what made build behave like the orchestrator).
	instrAfter, _ := os.ReadFile(instrPath)
	if strings.Contains(string(instrAfter), "jr-stack:sdd-orchestrator") {
		t.Error("stale sdd-orchestrator section must be purged from AGENTS.md")
	}
	if !strings.Contains(string(instrAfter), "keep me") {
		t.Error("user content in AGENTS.md must be preserved")
	}

	// 5. Backup created before any write.
	entries, err := os.ReadDir(filepath.Join(dir, ".jr-stack", "backups"))
	if err != nil || len(entries) == 0 {
		t.Errorf("backup should have been created on first install (err=%v)", err)
	}

	// 6. Idempotency — second install with same toggles changes nothing.
	result2, err := config.Install(h, []config.AgentAdapter{adapter}, dir)
	if err != nil {
		t.Fatalf("second Install error: %v", err)
	}
	if !result2.AllAlready {
		t.Error("second install with identical content should be AllAlready")
	}
}

// TestInstall_PrimaryAgent_FreshMachine verifies the orchestrator is registered
// even when no opencode.json exists yet (the file is created).
func TestInstall_PrimaryAgent_FreshMachine(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "opencode.json")
	// No AGENTS.md, no opencode.json — clean machine.

	h := model.Harness{ID: sddHarnessID, Type: model.HarnessConfig, Toggles: []string{"engram"}}
	adapter := opencodeAdapter{instr: filepath.Join(dir, "AGENTS.md"), settings: settingsPath}

	result, err := config.Install(h, []config.AgentAdapter{adapter}, dir)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if result.AllAlready {
		t.Error("install on fresh machine should NOT be AllAlready")
	}

	entry := readAgentEntry(t, settingsPath, sddHarnessID)
	if entry["mode"] != "primary" {
		t.Errorf("agent.%s mode = %v, want \"primary\"", sddHarnessID, entry["mode"])
	}
}

// TestInstall_PrimaryAgent_NoSettingsPath verifies that an empty SettingsPath is
// a graceful skip (no panic, no error), mirroring the instructions-path skip.
func TestInstall_PrimaryAgent_NoSettingsPath(t *testing.T) {
	dir := t.TempDir()
	h := model.Harness{ID: sddHarnessID, Type: model.HarnessConfig, Toggles: []string{"engram"}}
	adapter := opencodeAdapter{instr: "", settings: ""}

	result, err := config.Install(h, []config.AgentAdapter{adapter}, dir)
	if err != nil {
		t.Fatalf("Install with empty settings path should skip gracefully, got: %v", err)
	}
	if len(result.Files) != 0 {
		t.Errorf("empty-path primary adapter should write nothing, got %v", result.Files)
	}
}
