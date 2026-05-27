package external

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ── fakeAdapter ───────────────────────────────────────────────────────────

type fakeAdapter struct {
	agent      model.Agent
	configPath string // resolved per-test to a temp path
	strategy   MCPStrategy
}

func (f *fakeAdapter) Agent() model.Agent { return f.agent }
func (f *fakeAdapter) MCPConfigPath(homeDir, serverName string) string {
	if f.configPath != "" {
		return f.configPath
	}
	return filepath.Join(homeDir, "."+string(f.agent), serverName+".json")
}
func (f *fakeAdapter) MCPStrategy() MCPStrategy { return f.strategy }

// ── MCP overlay generation ────────────────────────────────────────────────

func TestBuildOverlay_OpenCode_MergeIntoSettings(t *testing.T) {
	h := harnessWithMethod("mcp", "", "https://mcp.context7.com")
	h.ID = "context7"
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	overlay := buildOverlay(h, adapter)

	mcp, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("overlay missing 'mcp' key, got keys: %v", mapKeys(overlay))
	}

	server, ok := mcp["context7"].(map[string]any)
	if !ok {
		t.Fatalf("overlay['mcp'] missing 'context7' key")
	}

	if server["type"] != "remote" {
		t.Errorf("type = %v, want remote", server["type"])
	}
	if !strings.Contains(server["url"].(string), "https://mcp.context7.com") {
		t.Errorf("url = %v, want to contain https://mcp.context7.com", server["url"])
	}
	if server["enabled"] != true {
		t.Errorf("enabled = %v, want true", server["enabled"])
	}
}

func TestBuildOverlay_NoHardcodedConstants(t *testing.T) {
	// A completely different harness — the overlay must use its ID and URL,
	// never any hardcoded "context7" or "engram" string.
	h := model.Harness{
		ID:   "custom-mcp",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.example.com",
		},
	}
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}
	overlay := buildOverlay(h, adapter)

	mcp, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatal("overlay missing 'mcp' key")
	}
	if _, exists := mcp["custom-mcp"]; !exists {
		t.Errorf("overlay key should be harness ID 'custom-mcp', got keys: %v", mapKeys(mcp))
	}
	if _, exists := mcp["context7"]; exists {
		t.Error("overlay must not contain hardcoded 'context7' key")
	}
}

func TestBuildOverlay_SeparateFile(t *testing.T) {
	h := harnessWithMethod("mcp", "", "https://mcp.example.com")
	h.ID = "my-server"
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategySeparateFile}

	overlay := buildOverlay(h, adapter)

	urlVal, ok := overlay["url"].(string)
	if !ok {
		t.Fatalf("SeparateFile overlay should have 'url' at top level, got: %v", overlay)
	}
	if !strings.Contains(urlVal, "https://mcp.example.com") {
		t.Errorf("url = %q, want to contain https://mcp.example.com", urlVal)
	}
}

// ── MCP install: idempotency ───────────────────────────────────────────────

func TestInstallMCP_Idempotent(t *testing.T) {
	homeDir := t.TempDir()

	// Stub backup so it doesn't create real backup directories.
	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}

	configPath := filepath.Join(homeDir, "opencode.json")
	adapter := &fakeAdapter{
		agent:      model.AgentOpenCode,
		configPath: configPath,
		strategy:   StrategyMergeIntoSettings,
	}

	// First install.
	result1, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	if len(result1.ConfigFiles) == 0 {
		t.Error("first install should have written a config file")
	}

	// Second install on same config — should be idempotent.
	result2, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second install: AlreadyInstalled should be true")
	}
	if len(result2.ConfigFiles) != 0 {
		t.Errorf("second install: ConfigFiles should be empty, got %v", result2.ConfigFiles)
	}
}

// ── MCP install: backup called when file exists ────────────────────────────

func TestInstallMCP_BackupCalledWhenFileExists(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "existing.json")

	// Write an existing config.
	os.WriteFile(configPath, []byte(`{"other":"value"}`), 0o644)

	var backupCalled bool
	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error {
		backupCalled = true
		return nil
	}
	defer func() { snapshotterCreate = origSnap }()

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	adapter := &fakeAdapter{
		agent:      model.AgentClaude,
		configPath: configPath,
		strategy:   StrategySeparateFile,
	}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}
	if !backupCalled {
		t.Error("backup was not called for an existing config file")
	}
}

func TestInstallMCP_NoBackupWhenFileAbsent(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "does-not-exist.json")

	var backupCalled bool
	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error {
		backupCalled = true
		return nil
	}
	defer func() { snapshotterCreate = origSnap }()

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{Method: "mcp", URL: "https://mcp.context7.com"},
	}
	adapter := &fakeAdapter{
		agent:      model.AgentClaude,
		configPath: configPath,
		strategy:   StrategySeparateFile,
	}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}
	if backupCalled {
		t.Error("backup should not be called when the file doesn't exist")
	}
}

// ── MCP install: config content validation ────────────────────────────────

func TestInstallMCP_WrittenContentIsValid(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "opencode.json")

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	adapter := &fakeAdapter{
		agent:      model.AgentOpenCode,
		configPath: configPath,
		strategy:   StrategyMergeIntoSettings,
	}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("written config is not valid JSON: %v\ncontent: %s", err, data)
	}

	mcp, ok := result["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("config should have 'mcp' key")
	}
	if _, exists := mcp["context7"]; !exists {
		t.Errorf("config['mcp'] should have 'context7' key, keys: %v", mapKeys(mcp))
	}
}

// ── MCP install: no adapters ──────────────────────────────────────────────

func TestInstallMCP_NoAdapters(t *testing.T) {
	h := harnessWithMethod("mcp", "", "https://mcp.example.com")
	result, err := installMCP(context.Background(), h, nil, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.AlreadyInstalled {
		t.Error("empty adapters should return AlreadyInstalled=true")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
