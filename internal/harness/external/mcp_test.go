package external

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
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

// ── buildMCPOverlay: Claude project single-file strategy (C-28) ───────────

// TestBuildMCPOverlay_ClaudeProject asserts that buildMCPOverlay returns the
// correct overlay for a model.MCP with Command, Args, and Env:
//   {"mcpServers": {"<Name>": {"command": ..., "args": [...], "env": {...}}}}
//
// No hardcoded server name constants — the overlay key is MCP.Name.
func TestBuildMCPOverlay_ClaudeProject(t *testing.T) {
	mcp := model.MCP{
		Name:    "context7",
		Command: "npx",
		Args:    []string{"-y", "@upstash/context7-mcp"},
		Env:     map[string]string{"DEBUG": "1"},
	}

	overlay := buildMCPOverlay(mcp)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("overlay missing 'mcpServers' key, got: %v", overlay)
	}

	server, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing 'context7' key, got keys: %v", mapKeys(servers))
	}

	if server["command"] != "npx" {
		t.Errorf("command = %v, want npx", server["command"])
	}

	args, ok := server["args"].([]string)
	if !ok || len(args) != 2 || args[0] != "-y" || args[1] != "@upstash/context7-mcp" {
		t.Errorf("args = %v, want [-y @upstash/context7-mcp]", server["args"])
	}

	env, ok := server["env"].(map[string]string)
	if !ok || env["DEBUG"] != "1" {
		t.Errorf("env = %v, want {DEBUG:1}", server["env"])
	}
}

// TestBuildMCPOverlay_NoHardcodedName asserts the overlay uses MCP.Name as the
// server key, not any hardcoded string.
func TestBuildMCPOverlay_NoHardcodedName(t *testing.T) {
	mcp := model.MCP{Name: "my-custom-server", Command: "uvx", Args: []string{"mcp-server-fetch"}}

	overlay := buildMCPOverlay(mcp)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("overlay missing 'mcpServers'")
	}
	if _, exists := servers["my-custom-server"]; !exists {
		t.Errorf("mcpServers should have key 'my-custom-server', got: %v", mapKeys(servers))
	}
	if _, exists := servers["context7"]; exists {
		t.Error("mcpServers must not contain hardcoded 'context7'")
	}
}

// TestBuildMCPOverlay_EnvAbsent asserts that when Env is nil, the env key is
// absent from the overlay (no spurious empty map written to .mcp.json).
func TestBuildMCPOverlay_EnvAbsent(t *testing.T) {
	mcp := model.MCP{Name: "simple", Command: "node", Args: []string{"server.js"}}

	overlay := buildMCPOverlay(mcp)

	servers := overlay["mcpServers"].(map[string]any)
	server := servers["simple"].(map[string]any)

	if _, hasEnv := server["env"]; hasEnv {
		t.Error("env key must not appear when MCP.Env is nil")
	}
}

// ── Idempotent merge for the Claude project case (C-28, 4.4) ─────────────

// TestBuildMCPOverlay_IdempotentMerge asserts that merging the same server
// twice into a .mcp.json yields a single mcpServers.<Name> entry (no duplicate).
// This exercises the filemerge.MergeJSONObjects path used by the installer.
func TestBuildMCPOverlay_IdempotentMerge(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, ".mcp.json")

	// Stub backup so it doesn't create real backup directories.
	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{
		Name:    "context7",
		Command: "npx",
		Args:    []string{"-y", "@upstash/context7-mcp"},
	}

	writeOverlay := func() {
		overlay := buildMCPOverlay(mcp)
		overlayJSON, err := json.Marshal(overlay)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		base := readExistingJSON(configPath)
		merged, err := filemerge.MergeJSONObjects(base, overlayJSON)
		if err != nil {
			t.Fatalf("MergeJSONObjects: %v", err)
		}
		if _, err := filemerge.WriteFileAtomic(configPath, merged, 0o644); err != nil {
			t.Fatalf("WriteFileAtomic: %v", err)
		}
	}

	// First write.
	writeOverlay()
	// Second write (same entry) — idempotent.
	writeOverlay()

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("written config not valid JSON: %v\n%s", err, data)
	}

	servers, ok := result["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("missing mcpServers in %v", result)
	}

	if _, exists := servers["context7"]; !exists {
		t.Errorf("mcpServers should have 'context7', got: %v", mapKeys(servers))
	}

	// Exactly one entry — no duplication from the second write.
	if len(servers) != 1 {
		t.Errorf("mcpServers has %d entries, want exactly 1: %v", len(servers), mapKeys(servers))
	}
}

// ── stdio MCP overlay: per-strategy shape ────────────────────────────────

// TestBuildStdioOverlay_SeparateFile asserts that buildStdioOverlay for a
// StrategySeparateFile adapter returns the bare server object at the top level:
//   {"command":..., "args":[...]}
func TestBuildStdioOverlay_SeparateFile(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategySeparateFile}

	overlay := buildStdioOverlay(mcp, adapter)

	if overlay["command"] != "engram" {
		t.Errorf("command = %v, want engram", overlay["command"])
	}
	args, ok := overlay["args"].([]string)
	if !ok || len(args) != 1 || args[0] != "mcp" {
		t.Errorf("args = %v, want [mcp]", overlay["args"])
	}
	// The "url" key must be absent (this is stdio, not remote).
	if _, hasURL := overlay["url"]; hasURL {
		t.Error("stdio overlay must not have 'url' key")
	}
}

// TestBuildStdioOverlay_MergeIntoSettings_Generic asserts that
// buildStdioOverlay for a generic StrategyMergeIntoSettings (non-OpenCode)
// adapter wraps the server object in the __replace__ sentinel:
//   {"mcpServers": {"<name>": {"__replace__": {"type":"stdio","command":...,"args":[...]}}}}
//
// Updated: the server object is now wrapped in {"__replace__": ...} so that
// MergeJSONObjects replaces any stale base entry atomically (no orphan "url").
func TestBuildStdioOverlay_MergeIntoSettings_Generic(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentGemini, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("overlay should have 'mcpServers', got keys: %v", mapKeys(overlay))
	}
	// Server entry is a sentinel: {"__replace__": {actual fields}}.
	sentinel, ok := servers["engram"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers['engram'] should be a sentinel map, got: %T %v", servers["engram"], servers["engram"])
	}
	server, ok := sentinel["__replace__"].(map[string]any)
	if !ok {
		t.Fatalf("sentinel should have '__replace__' key with server map, got: %v", mapKeys(sentinel))
	}
	if server["command"] != "engram" {
		t.Errorf("command = %v, want engram", server["command"])
	}
	if _, hasURL := server["url"]; hasURL {
		t.Error("stdio overlay must not have 'url' key in server entry")
	}
}

// TestBuildStdioOverlay_OpenCode asserts that buildStdioOverlay for an
// OpenCode adapter wraps the localEntry in the __replace__ sentinel:
//   {"mcp": {"<name>": {"__replace__": {"type":"local","command":...,"args":[...],"enabled":true}}}}
//
// Updated: localEntry is now wrapped in {"__replace__": ...} so that
// MergeJSONObjects replaces any stale remote entry atomically (no orphan keys).
func TestBuildStdioOverlay_OpenCode(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	mcpMap, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("OpenCode overlay should have 'mcp' key, got keys: %v", mapKeys(overlay))
	}
	// Entry is a sentinel: {"__replace__": {actual fields}}.
	sentinel, ok := mcpMap["engram"].(map[string]any)
	if !ok {
		t.Fatalf("mcp['engram'] should be a sentinel map, got: %T %v", mcpMap["engram"], mcpMap["engram"])
	}
	server, ok := sentinel["__replace__"].(map[string]any)
	if !ok {
		t.Fatalf("sentinel should have '__replace__' key with server map, got: %v", mapKeys(sentinel))
	}
	if server["type"] != "local" {
		t.Errorf("type = %v, want local", server["type"])
	}
	if server["command"] != "engram" {
		t.Errorf("command = %v, want engram", server["command"])
	}
	if server["enabled"] != true {
		t.Errorf("enabled = %v, want true", server["enabled"])
	}
	// Must NOT have "url" key (this is local stdio, not remote).
	if _, hasURL := server["url"]; hasURL {
		t.Error("OpenCode local overlay must not have 'url' key")
	}
}

// TestBuildStdioOverlay_EnvOmittedWhenEmpty asserts that the "env" key is
// absent from the overlay when MCP.Env is nil (no spurious empty map written).
func TestBuildStdioOverlay_EnvOmittedWhenEmpty(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategySeparateFile}

	overlay := buildStdioOverlay(mcp, adapter)

	if _, hasEnv := overlay["env"]; hasEnv {
		t.Error("env must not appear when MCP.Env is nil")
	}
}

// TestBuildStdioOverlay_EnvPresent asserts that the "env" key appears in the
// overlay when MCP.Env is set (separate-file strategy).
func TestBuildStdioOverlay_EnvPresent(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}, Env: map[string]string{"LOG": "1"}}
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategySeparateFile}

	overlay := buildStdioOverlay(mcp, adapter)

	env, ok := overlay["env"].(map[string]string)
	if !ok || env["LOG"] != "1" {
		t.Errorf("env = %v, want {LOG:1}", overlay["env"])
	}
}

// ── registerStdioMCP: writes stdio entry after binary install ────────────

// TestRegisterStdioMCP_WritesEntryPerAdapter asserts that registerStdioMCP
// writes the stdio MCP entry into each adapter's config path using
// the correct per-strategy overlay.
func TestRegisterStdioMCP_WritesEntryPerAdapter(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}

	claudePath := filepath.Join(homeDir, "claude.json")
	opencodePath := filepath.Join(homeDir, "opencode.json")

	adapters := []AgentAdapter{
		&fakeAdapter{agent: model.AgentClaude, configPath: claudePath, strategy: StrategySeparateFile},
		&fakeAdapter{agent: model.AgentOpenCode, configPath: opencodePath, strategy: StrategyMergeIntoSettings},
	}

	result, err := registerStdioMCP(mcp, adapters, homeDir, "engram")
	if err != nil {
		t.Fatalf("registerStdioMCP failed: %v", err)
	}
	if len(result.ConfigFiles) != 2 {
		t.Errorf("ConfigFiles = %v, want 2 paths", result.ConfigFiles)
	}

	// Claude: StrategySeparateFile — top-level command field in the file.
	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read claude config: %v", err)
	}
	var claudeObj map[string]any
	if err := json.Unmarshal(claudeData, &claudeObj); err != nil {
		t.Fatalf("claude config not valid JSON: %v\n%s", err, claudeData)
	}
	if claudeObj["command"] != "engram" {
		t.Errorf("claude config command = %v, want engram", claudeObj["command"])
	}

	// OpenCode: StrategyMergeIntoSettings — nested under "mcp" key with type:local.
	ocData, err := os.ReadFile(opencodePath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	var ocObj map[string]any
	if err := json.Unmarshal(ocData, &ocObj); err != nil {
		t.Fatalf("opencode config not valid JSON: %v\n%s", err, ocData)
	}
	mcpMap, ok := ocObj["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("opencode config should have 'mcp' key, got: %v", mapKeys(ocObj))
	}
	if _, exists := mcpMap["engram"]; !exists {
		t.Errorf("opencode mcp should have 'engram', got: %v", mapKeys(mcpMap))
	}
}

// TestRegisterStdioMCP_Idempotent asserts that calling registerStdioMCP twice
// with the same MCP entry does not produce duplicate config entries.
func TestRegisterStdioMCP_Idempotent(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	configPath := filepath.Join(homeDir, "opencode.json")
	adapters := []AgentAdapter{
		&fakeAdapter{agent: model.AgentOpenCode, configPath: configPath, strategy: StrategyMergeIntoSettings},
	}

	// First call.
	_, err := registerStdioMCP(mcp, adapters, homeDir, "engram")
	if err != nil {
		t.Fatalf("first registerStdioMCP failed: %v", err)
	}

	// Second call — should be idempotent.
	result2, err := registerStdioMCP(mcp, adapters, homeDir, "engram")
	if err != nil {
		t.Fatalf("second registerStdioMCP failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second call: AlreadyInstalled should be true (idempotent)")
	}
}

// TestRegisterStdioMCP_BackupCalledForExistingFile asserts that when the
// adapter's config path already exists, a backup snapshot is taken before
// overwriting it (governance ALTO rule).
func TestRegisterStdioMCP_BackupCalledForExistingFile(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "opencode.json")
	os.WriteFile(configPath, []byte(`{"other":"value"}`), 0o644)

	var backupCalled bool
	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error {
		backupCalled = true
		return nil
	}
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapters := []AgentAdapter{
		&fakeAdapter{agent: model.AgentOpenCode, configPath: configPath, strategy: StrategyMergeIntoSettings},
	}

	_, err := registerStdioMCP(mcp, adapters, homeDir, "engram")
	if err != nil {
		t.Fatalf("registerStdioMCP failed: %v", err)
	}
	if !backupCalled {
		t.Error("backup was not called for an existing config file (governance ALTO violation)")
	}
}

// ── Install() integration: homebrew + External.MCP ───────────────────────

// TestInstall_Homebrew_WithMCP_CallsBinaryAndMCP asserts that when Install is
// called for a harness with method=homebrew and External.MCP set, BOTH the
// binary installer is invoked AND the stdio MCP entry is written to each
// adapter's config path. The Result merges BinaryPath + ConfigFiles.
func TestInstall_Homebrew_WithMCP_BinaryAndMCPWritten(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	// Override the brew runner so no real brew command runs.
	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return homeDir + "/" + name, nil
	})()

	mcp := &model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	h := model.Harness{
		ID:   "engram",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "homebrew",
			Pkg:    "gentleman-programming/tap/engram",
			Repo:   "Gentleman-Programming/engram",
			MCP:    mcp,
		},
	}

	configPath := filepath.Join(homeDir, "opencode.json")
	adapters := []AgentAdapter{
		&fakeAdapter{agent: model.AgentOpenCode, configPath: configPath, strategy: StrategyMergeIntoSettings},
	}

	// Use a brew-available profile so runBrew is called, not downloadBinary.
	profile := linuxProfile(true) // linux+apt — no brew → will attempt downloadBinary
	// Override to a brew profile.
	profile.PackageManager = "brew"

	result, err := Install(context.Background(), h, profile, adapters, homeDir)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Binary path must be set (from runBrew's lookPath result).
	if result.BinaryPath == "" {
		t.Error("result.BinaryPath should be set after homebrew install")
	}

	// MCP config file must be written.
	if len(result.ConfigFiles) == 0 {
		t.Error("result.ConfigFiles should include the written MCP config path")
	}

	// Verify the content of the written config file.
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("config not valid JSON: %v\n%s", err, data)
	}
	mcpMap, ok := obj["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("config should have 'mcp' key (OpenCode local format), got: %v", mapKeys(obj))
	}
	if _, exists := mcpMap["engram"]; !exists {
		t.Errorf("mcp should have 'engram' entry, got: %v", mapKeys(mcpMap))
	}
}

// TestInstall_Homebrew_NoMCP_UnchangedBehavior is a regression test asserting
// that when External.MCP is nil, Install behaves exactly as before: only the
// binary step runs, ConfigFiles is empty.
func TestInstall_Homebrew_NoMCP_UnchangedBehavior(t *testing.T) {
	homeDir := t.TempDir()

	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return homeDir + "/" + name, nil
	})()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := linuxProfile(true)
	profile.PackageManager = "brew"

	result, err := Install(context.Background(), h, profile, nil, homeDir)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if len(result.ConfigFiles) != 0 {
		t.Errorf("ConfigFiles should be empty when External.MCP is nil, got %v", result.ConfigFiles)
	}
}

// TestInstall_MCP_Remote_Unchanged is a regression test asserting that the
// existing remote MCP (context7) behavior is unchanged after the stdio-MCP
// registration code is added. The context7 overlay must still use type:remote.
func TestInstall_MCP_Remote_Unchanged(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
			// MCP is nil — remote MCPs have no stdio spec.
		},
	}

	configPath := filepath.Join(homeDir, "opencode.json")
	adapters := []AgentAdapter{
		&fakeAdapter{agent: model.AgentOpenCode, configPath: configPath, strategy: StrategyMergeIntoSettings},
	}

	_, err := Install(context.Background(), h, linuxProfile(true), adapters, homeDir)
	if err != nil {
		t.Fatalf("Install (remote MCP) failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("config not valid JSON: %v\n%s", err, data)
	}
	mcpMap, ok := obj["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("config should have 'mcp' key (OpenCode remote format), got: %v", mapKeys(obj))
	}
	server, ok := mcpMap["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcp should have 'context7', got: %v", mapKeys(mcpMap))
	}
	// Remote URL behavior: type must be "remote", not "local".
	if server["type"] != "remote" {
		t.Errorf("context7 server type = %v, want remote (regression: remote MCP overlay changed)", server["type"])
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
