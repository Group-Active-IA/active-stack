package external

// Tests for agent scoping in external.Install (visual-verify, design D3):
// adapters of agents the harness does not support must be skipped, while
// harnesses without an Agents list keep installing for every adapter.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

func stubSnapshotter(t *testing.T) {
	t.Helper()
	orig := snapshotterCreate
	snapshotterCreate = func(string, []string) error { return nil }
	t.Cleanup(func() { snapshotterCreate = orig })
}

func playwrightHarness(t *testing.T) model.Harness {
	t.Helper()
	c, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load(): %v", err)
	}
	h, ok := c.ByID("playwright")
	if !ok {
		t.Fatal("playwright must exist in the catalog")
	}
	return h
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("%q is not valid JSON: %v\n%s", path, err, data)
	}
	return obj
}

// TestInstall_SkipsUnsupportedAgents: a stdio MCP harness scoped to claude
// must not touch the opencode config even when opencode is selected.
func TestInstall_SkipsUnsupportedAgents(t *testing.T) {
	stubSnapshotter(t)
	home := t.TempDir()
	h := model.Harness{
		ID:     "playwright",
		Type:   model.HarnessExternal,
		Agents: []model.Agent{model.AgentClaude},
		External: &model.External{
			Method: "mcp",
			MCP: &model.MCP{
				Name:    "playwright",
				Command: "npx",
				Args:    []string{"-y", "@playwright/mcp@0.0.83"},
			},
		},
	}
	claude := &claudeFixAdapter{homeDir: home}
	opencode := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	if _, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{claude, opencode}, home, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	if _, err := os.Stat(opencode.MCPConfigPath(home, h.ID)); !os.IsNotExist(err) {
		t.Errorf("opencode config must not be created, stat err = %v", err)
	}
	servers, _ := readJSONFile(t, filepath.Join(home, ".claude.json"))["mcpServers"].(map[string]any)
	entry, _ := servers["playwright"].(map[string]any)
	if entry["type"] != "stdio" || entry["command"] != "npx" {
		t.Fatalf("claude playwright entry = %v, want stdio/npx", entry)
	}
	wantArgs := []any{"-y", "@playwright/mcp@0.0.83"}
	if !reflect.DeepEqual(entry["args"], wantArgs) {
		t.Errorf("args = %v, want %v", entry["args"], wantArgs)
	}
}

// TestInstall_SkipsUnsupportedAgents_LeavesExistingFileUntouched: an existing
// opencode config stays byte-for-byte identical.
func TestInstall_SkipsUnsupportedAgents_LeavesExistingFileUntouched(t *testing.T) {
	stubSnapshotter(t)
	home := t.TempDir()
	h := playwrightHarness(t)
	claude := &claudeFixAdapter{homeDir: home}
	opencode := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}
	ocPath := opencode.MCPConfigPath(home, h.ID)
	if err := os.MkdirAll(filepath.Dir(ocPath), 0o755); err != nil {
		t.Fatal(err)
	}
	before := []byte("{\n  \"mcp\": {\"keep\": {\"type\": \"remote\", \"url\": \"x\"}}\n}\n")
	if err := os.WriteFile(ocPath, before, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{opencode, claude}, home, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	after, err := os.ReadFile(ocPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("opencode config changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestInstall_AgentAgnosticExternalStillInstallsForAll: regression guard — an
// external without Agents (context7 shape) writes to every selected adapter.
func TestInstall_AgentAgnosticExternalStillInstallsForAll(t *testing.T) {
	stubSnapshotter(t)
	home := t.TempDir()
	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	claude := &claudeFixAdapter{homeDir: home}
	opencode := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	if _, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{claude, opencode}, home, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	cl, _ := readJSONFile(t, filepath.Join(home, ".claude.json"))["mcpServers"].(map[string]any)
	if _, ok := cl["context7"]; !ok {
		t.Error("claude config must receive context7")
	}
	oc, _ := readJSONFile(t, opencode.MCPConfigPath(home, h.ID))["mcp"].(map[string]any)
	if _, ok := oc["context7"]; !ok {
		t.Error("opencode config must receive context7")
	}
}

// TestInstall_Playwright_Idempotent: the second install is a no-op and the
// config holds a single playwright entry.
func TestInstall_Playwright_Idempotent(t *testing.T) {
	stubSnapshotter(t)
	home := t.TempDir()
	h := playwrightHarness(t)
	claude := &claudeFixAdapter{homeDir: home}

	first, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{claude}, home, nil)
	if err != nil {
		t.Fatalf("first Install: %v", err)
	}
	if first.AlreadyInstalled {
		t.Error("first install must report a change")
	}
	second, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{claude}, home, nil)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !second.AlreadyInstalled {
		t.Error("second install must report AlreadyInstalled")
	}
	servers, _ := readJSONFile(t, filepath.Join(home, ".claude.json"))["mcpServers"].(map[string]any)
	if len(servers) != 1 {
		t.Errorf("want exactly 1 server entry, got %v", servers)
	}
}

// TestInstall_Playwright_PreservesOtherServers: existing MCP servers and
// unrelated top-level keys survive the install.
func TestInstall_Playwright_PreservesOtherServers(t *testing.T) {
	stubSnapshotter(t)
	home := t.TempDir()
	h := playwrightHarness(t)
	claude := &claudeFixAdapter{homeDir: home}
	seed := `{"numStartups": 7, "mcpServers": {"context7": {"type": "stdio", "command": "npx", "args": ["-y", "ctx"]}}}`
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(context.Background(), h, system.PlatformProfile{}, []AgentAdapter{claude}, home, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	obj := readJSONFile(t, filepath.Join(home, ".claude.json"))
	if obj["numStartups"] != float64(7) {
		t.Errorf("numStartups lost: %v", obj["numStartups"])
	}
	servers, _ := obj["mcpServers"].(map[string]any)
	ctx7, _ := servers["context7"].(map[string]any)
	if ctx7["command"] != "npx" || !reflect.DeepEqual(ctx7["args"], []any{"-y", "ctx"}) {
		t.Errorf("context7 entry altered: %v", ctx7)
	}
	if _, ok := servers["playwright"]; !ok {
		t.Error("playwright entry must be added")
	}
}
