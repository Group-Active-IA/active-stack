package external

// Tests for the Claude MCP fix — overlay type discriminators + integration.
//
// RED before:
//  - generic StrategyMergeIntoSettings buildOverlay branch lacks "type":"http"
//  - generic StrategyMergeIntoSettings buildStdioOverlay branch lacks "type":"stdio"
//  - installMCP for Claude writes to wrong path and with wrong overlay shape
//  - registerStdioMCP for Claude writes to wrong path and with wrong overlay shape

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ── Step 2: type discriminators in generic StrategyMergeIntoSettings ─────────

// TestBuildOverlay_Generic_MergeIntoSettings_HasTypeHTTP asserts that when the
// adapter is NOT OpenCode (generic branch) and strategy is StrategyMergeIntoSettings,
// buildOverlay includes "type":"http" in the server entry.
//
// RED: fails before fix because current generic branch omits the "type" field.
func TestBuildOverlay_Generic_MergeIntoSettings_HasTypeHTTP(t *testing.T) {
	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	// Use AgentClaude (generic path — not OpenCode).
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategyMergeIntoSettings}

	overlay := buildOverlay(h, adapter)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("overlay should have 'mcpServers' key for generic MergeIntoSettings, got keys: %v", mapKeys(overlay))
	}
	server, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers should have 'context7', got: %v", mapKeys(servers))
	}
	if server["type"] != "http" {
		t.Errorf(`server["type"] = %v, want "http"`, server["type"])
	}
	if server["url"] == "" {
		t.Error(`server["url"] must not be empty`)
	}
}

// TestBuildStdioOverlay_Generic_MergeIntoSettings_HasTypeStdio asserts that when
// the adapter is NOT OpenCode (generic branch) and strategy is StrategyMergeIntoSettings,
// buildStdioOverlay includes "type":"stdio" in the server entry wrapped in the
// __replace__ sentinel.
//
// Updated: the server entry is now {"__replace__": {type:"stdio", command:..., ...}}
// so MergeJSONObjects replaces any stale base entry atomically.
func TestBuildStdioOverlay_Generic_MergeIntoSettings_HasTypeStdio(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	// Use AgentClaude (generic path — not OpenCode).
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("overlay should have 'mcpServers' key for generic MergeIntoSettings, got keys: %v", mapKeys(overlay))
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
	if server["type"] != "stdio" {
		t.Errorf(`server["type"] = %v, want "stdio"`, server["type"])
	}
	if server["command"] != "engram" {
		t.Errorf(`server["command"] = %v, want "engram"`, server["command"])
	}
}

// TestBuildStdioOverlay_Generic_MergeIntoSettings_HasTypeStdio_WithEnv asserts
// that buildStdioOverlay for generic MergeIntoSettings includes "type":"stdio"
// and "env" even when Env is set, inside the __replace__ sentinel.
func TestBuildStdioOverlay_Generic_MergeIntoSettings_HasTypeStdio_WithEnv(t *testing.T) {
	mcp := model.MCP{
		Name:    "engram",
		Command: "engram",
		Args:    []string{"mcp"},
		Env:     map[string]string{"LOG": "1"},
	}
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	servers := overlay["mcpServers"].(map[string]any)
	// Unwrap the __replace__ sentinel to inspect actual server fields.
	sentinel := servers["engram"].(map[string]any)
	server := sentinel["__replace__"].(map[string]any)

	if server["type"] != "stdio" {
		t.Errorf(`server["type"] = %v, want "stdio"`, server["type"])
	}
	if server["env"] == nil {
		t.Error("server[\"env\"] must be present when MCP.Env is set")
	}
}

// TestBuildOverlay_OpenCode_Unchanged asserts that the OpenCode sub-branch in
// buildOverlay is NOT affected by the generic branch fix (regression guard).
func TestBuildOverlay_OpenCode_Unchanged(t *testing.T) {
	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	overlay := buildOverlay(h, adapter)

	mcp, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("OpenCode overlay should use 'mcp' key, got keys: %v", mapKeys(overlay))
	}
	server := mcp["context7"].(map[string]any)
	if server["type"] != "remote" {
		t.Errorf("OpenCode overlay type = %v, want remote (regression)", server["type"])
	}
	if server["enabled"] != true {
		t.Errorf("OpenCode overlay enabled = %v, want true (regression)", server["enabled"])
	}
}

// TestBuildStdioOverlay_OpenCode_Unchanged asserts that the OpenCode sub-branch
// in buildStdioOverlay uses the __replace__ sentinel (consistent with the generic
// branch) and still produces type:"local" + enabled:true inside the sentinel.
//
// Updated: entry is now {"__replace__": {type:"local", enabled:true, ...}}.
func TestBuildStdioOverlay_OpenCode_Unchanged(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	mcpMap, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("OpenCode stdio overlay should use 'mcp' key, got keys: %v", mapKeys(overlay))
	}
	// Unwrap the __replace__ sentinel to inspect actual server fields.
	sentinel, ok := mcpMap["engram"].(map[string]any)
	if !ok {
		t.Fatalf("mcp['engram'] should be a sentinel map, got: %T %v", mcpMap["engram"], mcpMap["engram"])
	}
	server, ok := sentinel["__replace__"].(map[string]any)
	if !ok {
		t.Fatalf("sentinel should have '__replace__' key with server map, got: %v", mapKeys(sentinel))
	}
	if server["type"] != "local" {
		t.Errorf("OpenCode stdio overlay type = %v, want local", server["type"])
	}
	if server["enabled"] != true {
		t.Errorf("OpenCode stdio overlay enabled = %v, want true", server["enabled"])
	}
}

// ── Step 3: integration — installMCP + registerStdioMCP with Claude adapter ──

// claudeAdapter is a minimal AgentAdapter that mimics the FIXED Claude adapter:
// MCPConfigPath returns ~/.claude.json and MCPStrategy returns StrategyMergeIntoSettings.
// Used in integration tests before the real adapter is updated, then kept to test
// the real adapter later.
type claudeFixAdapter struct {
	homeDir string
}

func (c *claudeFixAdapter) Agent() model.Agent { return model.AgentClaude }
func (c *claudeFixAdapter) MCPConfigPath(homeDir, _ string) string {
	return filepath.Join(homeDir, ".claude.json")
}
func (c *claudeFixAdapter) MCPStrategy() MCPStrategy { return StrategyMergeIntoSettings }

// TestInstallMCP_Claude_WritesToClaudeJSON asserts that after the fix, installMCP
// for context7 with the Claude adapter writes to ~/.claude.json with:
//
//	{"mcpServers":{"context7":{"type":"http","url":"https://mcp.context7.com/mcp"}}}
//
// RED: fails before fix (wrong path + missing type field).
func TestInstallMCP_Claude_WritesToClaudeJSON(t *testing.T) {
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
		},
	}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	result, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	wantPath := filepath.Join(homeDir, ".claude.json")

	// ConfigFiles should list ~/.claude.json.
	if len(result.ConfigFiles) == 0 {
		t.Fatal("result.ConfigFiles should not be empty")
	}
	if result.ConfigFiles[0] != wantPath {
		t.Errorf("ConfigFiles[0] = %q, want %q", result.ConfigFiles[0], wantPath)
	}

	// Content shape: {"mcpServers":{"context7":{"type":"http","url":"..."}}}
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read %q: %v", wantPath, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}
	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json should have 'mcpServers', got: %v", mapKeys(obj))
	}
	server, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers should have 'context7', got: %v", mapKeys(servers))
	}
	if server["type"] != "http" {
		t.Errorf(`server["type"] = %v, want "http"`, server["type"])
	}
	if server["url"] == "" {
		t.Error(`server["url"] must not be empty`)
	}
}

// TestInstallMCP_Claude_Idempotent asserts that running installMCP twice for
// context7 with the Claude adapter does NOT produce duplicate entries and the
// second call returns AlreadyInstalled=true.
//
// RED: fails before fix.
func TestInstallMCP_Claude_Idempotent(t *testing.T) {
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
		},
	}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	// First install.
	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("first installMCP failed: %v", err)
	}

	// Second install — idempotent.
	result2, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("second installMCP failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second installMCP: AlreadyInstalled should be true")
	}
	if len(result2.ConfigFiles) != 0 {
		t.Errorf("second installMCP: ConfigFiles should be empty, got %v", result2.ConfigFiles)
	}
}

// TestInstallMCP_Claude_PreservesUnrelatedKeys asserts that if ~/.claude.json
// already contains unrelated top-level keys (e.g. theme settings), those keys
// are preserved after merging in the MCP entry (governance ALTO: no truncation).
//
// RED: fails before fix because the wrong path is used.
func TestInstallMCP_Claude_PreservesUnrelatedKeys(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	claudeJSON := filepath.Join(homeDir, ".claude.json")
	// Pre-seed with unrelated existing config keys.
	existing := `{"theme":"dark","autoUpdates":true,"someOtherSetting":42}`
	if err := os.WriteFile(claudeJSON, []byte(existing), 0o644); err != nil {
		t.Fatalf("setup: write existing config: %v", err)
	}

	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("read %q: %v", claudeJSON, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	// Unrelated keys must be preserved.
	if obj["theme"] != "dark" {
		t.Errorf(`"theme" key lost after merge, got: %v`, obj["theme"])
	}
	if obj["autoUpdates"] != true {
		t.Errorf(`"autoUpdates" key lost after merge, got: %v`, obj["autoUpdates"])
	}
	if _, ok := obj["someOtherSetting"]; !ok {
		t.Error(`"someOtherSetting" key lost after merge (governance ALTO violation: truncated user config)`)
	}

	// And the new MCP entry is also there.
	if _, ok := obj["mcpServers"]; !ok {
		t.Error(`"mcpServers" key not found after merge`)
	}
}

// TestRegisterStdioMCP_Claude_WritesToClaudeJSON asserts that registerStdioMCP
// for engram with the Claude adapter writes to ~/.claude.json with:
//
//	{"mcpServers":{"engram":{"type":"stdio","command":"engram","args":["mcp"]}}}
//
// RED: fails before fix.
func TestRegisterStdioMCP_Claude_WritesToClaudeJSON(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	result, err := registerStdioMCP(mcp, []AgentAdapter{adapter}, homeDir, "engram")
	if err != nil {
		t.Fatalf("registerStdioMCP failed: %v", err)
	}

	wantPath := filepath.Join(homeDir, ".claude.json")

	if len(result.ConfigFiles) == 0 {
		t.Fatal("result.ConfigFiles should not be empty")
	}
	if result.ConfigFiles[0] != wantPath {
		t.Errorf("ConfigFiles[0] = %q, want %q", result.ConfigFiles[0], wantPath)
	}

	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read %q: %v", wantPath, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}
	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json should have 'mcpServers', got: %v", mapKeys(obj))
	}
	server, ok := servers["engram"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers should have 'engram', got: %v", mapKeys(servers))
	}
	if server["type"] != "stdio" {
		t.Errorf(`server["type"] = %v, want "stdio"`, server["type"])
	}
	if server["command"] != "engram" {
		t.Errorf(`server["command"] = %v, want "engram"`, server["command"])
	}
}

// TestRegisterStdioMCP_Claude_Idempotent asserts that calling registerStdioMCP
// twice for engram with Claude adapter is idempotent (second call AlreadyInstalled=true).
func TestRegisterStdioMCP_Claude_Idempotent(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	_, err := registerStdioMCP(mcp, []AgentAdapter{adapter}, homeDir, "engram")
	if err != nil {
		t.Fatalf("first registerStdioMCP failed: %v", err)
	}

	result2, err := registerStdioMCP(mcp, []AgentAdapter{adapter}, homeDir, "engram")
	if err != nil {
		t.Fatalf("second registerStdioMCP failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second call: AlreadyInstalled should be true")
	}
}

// TestRegisterStdioMCP_Claude_PreservesUnrelatedKeys asserts that existing
// unrelated keys in ~/.claude.json are preserved when merging engram stdio entry.
// This is the governance ALTO proof: no truncation of user config.
func TestRegisterStdioMCP_Claude_PreservesUnrelatedKeys(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	claudeJSON := filepath.Join(homeDir, ".claude.json")
	existing := `{"theme":"dark","numericAISuggestions":5}`
	if err := os.WriteFile(claudeJSON, []byte(existing), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &claudeFixAdapter{homeDir: homeDir}

	_, err := registerStdioMCP(mcp, []AgentAdapter{adapter}, homeDir, "engram")
	if err != nil {
		t.Fatalf("registerStdioMCP failed: %v", err)
	}

	data, _ := os.ReadFile(claudeJSON)
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	if obj["theme"] != "dark" {
		t.Errorf(`"theme" lost after merge, got: %v`, obj["theme"])
	}
	if _, ok := obj["numericAISuggestions"]; !ok {
		t.Error(`"numericAISuggestions" lost after merge (governance ALTO violation)`)
	}
	if _, ok := obj["mcpServers"]; !ok {
		t.Error(`"mcpServers" not found after merge`)
	}
}

// TestRegisterStdioMCP_Claude_BothServicesCoexist asserts that after registering
// both context7 (via installMCP) and engram (via registerStdioMCP) for Claude,
// both entries coexist in ~/.claude.json under "mcpServers".
func TestRegisterStdioMCP_Claude_BothServicesCoexist(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	adapter := &claudeFixAdapter{homeDir: homeDir}

	// Install context7 (remote MCP).
	h := model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
		},
	}
	if _, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir); err != nil {
		t.Fatalf("installMCP (context7) failed: %v", err)
	}

	// Register engram (stdio MCP).
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	if _, err := registerStdioMCP(mcp, []AgentAdapter{adapter}, homeDir, "engram"); err != nil {
		t.Fatalf("registerStdioMCP (engram) failed: %v", err)
	}

	// Read final state.
	claudeJSON := filepath.Join(homeDir, ".claude.json")
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("read %q: %v", claudeJSON, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json missing 'mcpServers', got: %v", mapKeys(obj))
	}
	if _, ok := servers["context7"]; !ok {
		t.Errorf("mcpServers missing 'context7', got: %v", mapKeys(servers))
	}
	if _, ok := servers["engram"]; !ok {
		t.Errorf("mcpServers missing 'engram', got: %v", mapKeys(servers))
	}
}

// ── Context7 stdio-vs-remote dispatch (fix for auth-required remote endpoint) ──

// context7HarnessWithMCP returns the context7 harness as it will appear in the
// catalog after Cambio 1: External.MCP set to the npx stdio spec.
func context7HarnessWithMCP() model.Harness {
	return model.Harness{
		ID:   "context7",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: "mcp",
			URL:    "https://mcp.context7.com",
			MCP: &model.MCP{
				Name:    "context7",
				Command: "npx",
				Args:    []string{"-y", "--package=@upstash/context7-mcp@2.2.5", "--", "context7-mcp"},
			},
		},
	}
}

// TestInstallMCP_Context7_Claude_UsesStdio asserts that when context7 has
// External.MCP set (the stdio spec) and the adapter is Claude, installMCP writes
// a stdio entry (type:"stdio", command:"npx") to ~/.claude.json, NOT a remote URL.
//
// RED: fails before Cambio 2 because installMCP always calls buildOverlay (remote),
// ignoring External.MCP for MCP-method harnesses.
func TestInstallMCP_Context7_Claude_UsesStdio(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := context7HarnessWithMCP()
	adapter := &claudeFixAdapter{homeDir: homeDir}

	result, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	wantPath := filepath.Join(homeDir, ".claude.json")
	if len(result.ConfigFiles) == 0 {
		t.Fatal("result.ConfigFiles should not be empty")
	}
	if result.ConfigFiles[0] != wantPath {
		t.Errorf("ConfigFiles[0] = %q, want %q", result.ConfigFiles[0], wantPath)
	}

	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read %q: %v", wantPath, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	// Must use mcpServers (not mcp), with type:stdio — NOT type:http or a url.
	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json should have 'mcpServers' for stdio, got keys: %v", mapKeys(obj))
	}
	server, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers should have 'context7', got: %v", mapKeys(servers))
	}
	if server["type"] != "stdio" {
		t.Errorf(`server["type"] = %v, want "stdio" (context7 remote exige auth → usar stdio local)`, server["type"])
	}
	if server["command"] != "npx" {
		t.Errorf(`server["command"] = %v, want "npx"`, server["command"])
	}
	if _, hasURL := server["url"]; hasURL {
		t.Error(`server["url"] must not be present for stdio entry (no remote auth endpoint)`)
	}
}

// ── TRIANGULATE ───────────────────────────────────────────────────────────────

// TestInstallMCP_Context7_OpenCode_StaysRemote asserts that context7 with
// External.MCP set still uses the remote overlay for OpenCode (the endpoint works
// without auth for OpenCode — do NOT change that behaviour).
func TestInstallMCP_Context7_OpenCode_StaysRemote(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := context7HarnessWithMCP()

	configPath := filepath.Join(homeDir, "opencode.json")
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
		t.Fatalf("read %q: %v", configPath, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	// OpenCode must still use "mcp" key with type:"remote" (unchanged).
	mcpMap, ok := obj["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("OpenCode config should have 'mcp' key (remote), got keys: %v", mapKeys(obj))
	}
	server, ok := mcpMap["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcp should have 'context7', got: %v", mapKeys(mcpMap))
	}
	if server["type"] != "remote" {
		t.Errorf(`OpenCode server["type"] = %v, want "remote" (regression: must not become stdio for OpenCode)`, server["type"])
	}
	if server["enabled"] != true {
		t.Errorf(`OpenCode server["enabled"] = %v, want true`, server["enabled"])
	}
	if _, hasURL := server["url"]; !hasURL {
		t.Error(`OpenCode server must have "url" key for remote transport`)
	}
	// Must NOT have "mcpServers" key (that is the Claude/generic shape).
	if _, hasMCPServers := obj["mcpServers"]; hasMCPServers {
		t.Error(`OpenCode config must not have "mcpServers" key — that is the Claude shape`)
	}
}

// TestInstallMCP_Context7_Claude_Idempotent_WithMCP asserts that installing
// context7 stdio for Claude twice is idempotent: second call returns
// AlreadyInstalled=true and does not duplicate entries.
func TestInstallMCP_Context7_Claude_Idempotent_WithMCP(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := context7HarnessWithMCP()
	adapter := &claudeFixAdapter{homeDir: homeDir}

	// First install.
	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("first installMCP failed: %v", err)
	}

	// Second install — must be idempotent.
	result2, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("second installMCP failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second installMCP: AlreadyInstalled should be true (idempotent)")
	}
	if len(result2.ConfigFiles) != 0 {
		t.Errorf("second installMCP: ConfigFiles should be empty, got %v", result2.ConfigFiles)
	}

	// Verify no duplication in the file: exactly one mcpServers.context7 entry.
	claudeJSON := filepath.Join(homeDir, ".claude.json")
	data, err := os.ReadFile(claudeJSON)
	if err != nil {
		t.Fatalf("read %q: %v", claudeJSON, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}
	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json missing 'mcpServers', got: %v", mapKeys(obj))
	}
	if len(servers) != 1 {
		t.Errorf("mcpServers has %d entries after idempotent reinstall, want exactly 1: %v", len(servers), mapKeys(servers))
	}
}
