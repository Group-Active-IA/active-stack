package external

// Tests for the __replace__ sentinel fix in buildStdioOverlay for Claude.
//
// Problem: deep-merge of stdio overlay onto a stale remote entry leaves orphan
// keys (e.g. "url") in the resulting config. The fix wraps the server object in
// the __replace__ sentinel so MergeJSONObjects discards the base value entirely.
//
// RED test: pre-seed ~/.claude.json with a broken remote context7 entry, install
// stdio context7, assert the "url" key is gone and only stdio fields survive.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// TestInstallMCP_Context7_Claude_OverwritesBrokenRemote is the RED test.
//
// It pre-seeds ~/.claude.json with the stale remote context7 config that a
// user may have from a previous broken install:
//
//	{"mcpServers":{"context7":{"type":"http","url":"https://mcp.context7.com/mcp"}}}
//
// Then installs context7 using the stdio overlay (External.MCP set).
// Asserts that the resulting context7 entry is EXACTLY
//
//	{"type":"stdio","command":"npx","args":["-y","--package=@upstash/context7-mcp@2.2.5","--","context7-mcp"]}
//
// and that the "url" key NO LONGER EXISTS (deep merge would leave it; __replace__
// discards the whole stale base value atomically).
//
// RED: fails before the __replace__ fix because MergeJSONObjects deep-merges the
// stale {"type":"http","url":"..."} with the new stdio entry, leaving url orphaned.
func TestInstallMCP_Context7_Claude_OverwritesBrokenRemote(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	// Pre-seed with stale remote (broken) config.
	claudeJSON := filepath.Join(homeDir, ".claude.json")
	brokenRemote := `{"mcpServers":{"context7":{"type":"http","url":"https://mcp.context7.com/mcp"}}}`
	if err := os.WriteFile(claudeJSON, []byte(brokenRemote), 0o644); err != nil {
		t.Fatalf("setup: write broken remote config: %v", err)
	}

	h := context7HarnessWithMCP()
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

	servers, ok := obj["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("~/.claude.json should have 'mcpServers', got keys: %v", mapKeys(obj))
	}
	server, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers should have 'context7', got: %v", mapKeys(servers))
	}

	// Assert "url" key is gone — the stale remote value must be fully discarded.
	if _, hasURL := server["url"]; hasURL {
		t.Errorf("server still has orphan 'url' key after reinstall — deep merge left stale remote data; want __replace__ to discard it entirely. server = %v", server)
	}

	// Assert the stdio fields are correct.
	if server["type"] != "stdio" {
		t.Errorf(`server["type"] = %v, want "stdio"`, server["type"])
	}
	if server["command"] != "npx" {
		t.Errorf(`server["command"] = %v, want "npx"`, server["command"])
	}
	args, ok := server["args"].([]interface{})
	if !ok || len(args) == 0 {
		t.Errorf(`server["args"] = %v, want non-empty slice`, server["args"])
	}

	// "type" must be exactly "stdio" — not "http" leftover.
	if server["type"] == "http" {
		t.Error(`server["type"] is still "http" — stale remote type not replaced`)
	}
}

// TestInstallMCP_Context7_Claude_CleanInstall_NoURL asserts that a fresh install
// (no pre-existing config) also produces a clean stdio entry with no "url" key.
// TRIANGULATE: clean install path.
func TestInstallMCP_Context7_Claude_CleanInstall_NoURL(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	h := context7HarnessWithMCP()
	adapter := &claudeFixAdapter{homeDir: homeDir}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	claudeJSON := filepath.Join(homeDir, ".claude.json")
	data, _ := os.ReadFile(claudeJSON)
	var obj map[string]any
	json.Unmarshal(data, &obj)

	servers := obj["mcpServers"].(map[string]any)
	server := servers["context7"].(map[string]any)

	if _, hasURL := server["url"]; hasURL {
		t.Errorf("clean install: server should not have 'url' key, got: %v", server)
	}
	if server["type"] != "stdio" {
		t.Errorf("clean install: type = %v, want stdio", server["type"])
	}
	if server["command"] != "npx" {
		t.Errorf("clean install: command = %v, want npx", server["command"])
	}
}

// TestInstallMCP_Context7_Claude_IdempotentAfterReplace asserts that reinstalling
// context7 after the __replace__ fix is idempotent: the second call returns
// AlreadyInstalled=true and does not corrupt the entry.
// TRIANGULATE: idempotency after __replace__.
func TestInstallMCP_Context7_Claude_IdempotentAfterReplace(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	// Start from broken remote.
	claudeJSON := filepath.Join(homeDir, ".claude.json")
	brokenRemote := `{"mcpServers":{"context7":{"type":"http","url":"https://mcp.context7.com/mcp"}}}`
	os.WriteFile(claudeJSON, []byte(brokenRemote), 0o644)

	h := context7HarnessWithMCP()
	adapter := &claudeFixAdapter{homeDir: homeDir}

	// First install: replaces broken remote.
	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("first installMCP failed: %v", err)
	}

	// Second install: must be idempotent.
	result2, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("second installMCP failed: %v", err)
	}
	if !result2.AlreadyInstalled {
		t.Error("second installMCP: AlreadyInstalled should be true after __replace__ idempotent reinstall")
	}

	// Verify entry is still clean after second call.
	data, _ := os.ReadFile(claudeJSON)
	var obj map[string]any
	json.Unmarshal(data, &obj)
	servers := obj["mcpServers"].(map[string]any)
	server := servers["context7"].(map[string]any)

	if _, hasURL := server["url"]; hasURL {
		t.Error("idempotent reinstall: orphan 'url' reappeared")
	}
	if server["type"] != "stdio" {
		t.Errorf("idempotent reinstall: type = %v, want stdio", server["type"])
	}
}

// TestInstallMCP_Context7_Claude_PreservesOtherServers asserts that when
// reinstalling context7 on a ~/.claude.json that has OTHER mcpServers entries
// (e.g. engram), those other entries are NOT wiped by the __replace__ sentinel.
// The sentinel applies only to the context7 key, not the whole mcpServers map.
// TRIANGULATE: __replace__ scope is limited to the named server.
func TestInstallMCP_Context7_Claude_PreservesOtherServers(t *testing.T) {
	homeDir := t.TempDir()

	origSnap := snapshotterCreate
	snapshotterCreate = func(dir string, paths []string) error { return nil }
	defer func() { snapshotterCreate = origSnap }()

	claudeJSON := filepath.Join(homeDir, ".claude.json")
	// Pre-seed with both: stale context7 remote + existing engram stdio.
	existing := `{"mcpServers":{"context7":{"type":"http","url":"https://mcp.context7.com/mcp"},"engram":{"type":"stdio","command":"engram","args":["mcp"]}}}`
	os.WriteFile(claudeJSON, []byte(existing), 0o644)

	h := context7HarnessWithMCP()
	adapter := &claudeFixAdapter{homeDir: homeDir}

	_, err := installMCP(context.Background(), h, []AgentAdapter{adapter}, homeDir)
	if err != nil {
		t.Fatalf("installMCP failed: %v", err)
	}

	data, _ := os.ReadFile(claudeJSON)
	var obj map[string]any
	json.Unmarshal(data, &obj)
	servers := obj["mcpServers"].(map[string]any)

	// context7 must be clean stdio.
	c7, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("context7 missing from mcpServers: %v", mapKeys(servers))
	}
	if _, hasURL := c7["url"]; hasURL {
		t.Error("context7 still has orphan 'url' after __replace__")
	}
	if c7["type"] != "stdio" {
		t.Errorf("context7 type = %v, want stdio", c7["type"])
	}

	// engram must be preserved (unrelated server).
	engram, ok := servers["engram"].(map[string]any)
	if !ok {
		t.Fatalf("engram entry was wiped — __replace__ scope too wide, got servers: %v", mapKeys(servers))
	}
	if engram["type"] != "stdio" {
		t.Errorf("engram type = %v, want stdio (preserved)", engram["type"])
	}
}

// TestBuildStdioOverlay_Generic_HasReplaceSentinel asserts that after the fix,
// buildStdioOverlay for the generic StrategyMergeIntoSettings (Claude) path
// wraps the server object in the __replace__ sentinel. This verifies the
// overlay structure produced by the function, not the merged result.
func TestBuildStdioOverlay_Generic_HasReplaceSentinel(t *testing.T) {
	mcp := model.MCP{
		Name:    "context7",
		Command: "npx",
		Args:    []string{"-y", "--package=@upstash/context7-mcp@2.2.5", "--", "context7-mcp"},
	}
	adapter := &fakeAdapter{agent: model.AgentClaude, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	servers, ok := overlay["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("overlay should have 'mcpServers', got keys: %v", mapKeys(overlay))
	}

	// The server entry must be a sentinel: {"__replace__": {...actual fields...}}.
	sentinel, ok := servers["context7"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers['context7'] should be a map (sentinel), got: %T %v", servers["context7"], servers["context7"])
	}
	if len(sentinel) != 1 {
		t.Errorf("sentinel map should have exactly 1 key '__replace__', got keys: %v", mapKeys(sentinel))
	}
	inner, ok := sentinel["__replace__"].(map[string]any)
	if !ok {
		t.Fatalf("sentinel['__replace__'] should be a map, got: %T %v", sentinel["__replace__"], sentinel["__replace__"])
	}
	if inner["type"] != "stdio" {
		t.Errorf("inner type = %v, want stdio", inner["type"])
	}
	if inner["command"] != "npx" {
		t.Errorf("inner command = %v, want npx", inner["command"])
	}
}

// TestBuildStdioOverlay_OpenCode_HasReplaceSentinel asserts that after the fix,
// buildStdioOverlay for OpenCode also wraps localEntry in the __replace__ sentinel.
func TestBuildStdioOverlay_OpenCode_HasReplaceSentinel(t *testing.T) {
	mcp := model.MCP{Name: "engram", Command: "engram", Args: []string{"mcp"}}
	adapter := &fakeAdapter{agent: model.AgentOpenCode, strategy: StrategyMergeIntoSettings}

	overlay := buildStdioOverlay(mcp, adapter)

	mcpMap, ok := overlay["mcp"].(map[string]any)
	if !ok {
		t.Fatalf("OpenCode overlay should have 'mcp' key, got keys: %v", mapKeys(overlay))
	}

	sentinel, ok := mcpMap["engram"].(map[string]any)
	if !ok {
		t.Fatalf("mcp['engram'] should be a map (sentinel), got: %T %v", mcpMap["engram"], mcpMap["engram"])
	}
	if len(sentinel) != 1 {
		t.Errorf("sentinel map should have exactly 1 key '__replace__', got keys: %v", mapKeys(sentinel))
	}
	inner, ok := sentinel["__replace__"].(map[string]any)
	if !ok {
		t.Fatalf("sentinel['__replace__'] should be a map, got: %T %v", sentinel["__replace__"], sentinel["__replace__"])
	}
	if inner["type"] != "local" {
		t.Errorf("inner type = %v, want local", inner["type"])
	}
	if inner["enabled"] != true {
		t.Errorf("inner enabled = %v, want true", inner["enabled"])
	}
}
