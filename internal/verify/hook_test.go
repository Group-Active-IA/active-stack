package verify_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// ─────────────────────────────────────────────────────────────────────────────
// 3.1 — BuildHook
// ─────────────────────────────────────────────────────────────────────────────

// TestBuildHookReturnsNilWhenAllChecksPass exercises the path where every
// harness check passes. The hook must return nil (install completes).
func TestBuildHookReturnsNilWhenAllChecksPass(t *testing.T) {
	homeDir := t.TempDir()

	h := model.Harness{
		ID:   "active-orchestrator",
		Type: model.HarnessSkill,
	}

	// Create the expected SKILL.md.
	fa := fakeAdapter{
		agent:     model.AgentClaude,
		skillsDir: filepath.Join(homeDir, "skills"),
	}
	skillDir := filepath.Join(fa.SkillsDir(homeDir), h.ID)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# content"), 0o644); err != nil {
		t.Fatal(err)
	}

	hook := verify.BuildHook([]model.Harness{h}, []verify.Adapter{fa}, homeDir)
	if err := hook(); err != nil {
		t.Errorf("hook() = %v, want nil (all checks passed)", err)
	}
}

// TestBuildHookReturnsErrorWhenHardCheckFails exercises the path where a hard
// check fails. The hook must return a non-nil error so the pipeline rolls back.
func TestBuildHookReturnsErrorWhenHardCheckFails(t *testing.T) {
	homeDir := t.TempDir()

	h := model.Harness{
		ID:   "active-orchestrator",
		Type: model.HarnessSkill,
	}

	fa := fakeAdapter{
		agent:     model.AgentClaude,
		skillsDir: filepath.Join(homeDir, "skills"),
	}
	// Do NOT create SKILL.md — hard check should fail.

	hook := verify.BuildHook([]model.Harness{h}, []verify.Adapter{fa}, homeDir)
	if err := hook(); err == nil {
		t.Error("hook() = nil, want error when hard check fails")
	}
}

// TestBuildHookReturnsNilWhenOnlySoftWarnings exercises the path where all
// hard checks pass but a soft check warns. Hook must return nil.
func TestBuildHookReturnsNilWhenOnlySoftWarnings(t *testing.T) {
	homeDir := t.TempDir()

	// External homebrew harness → its checks are all Soft.
	h := model.Harness{
		ID:       "engram",
		Type:     model.HarnessExternal,
		External: &model.External{Method: "homebrew", Pkg: "engram"},
	}

	fa := fakeAdapter{
		agent:        model.AgentClaude,
		mcpConfigPath: filepath.Join(homeDir, "mcp", "engram.json"),
	}

	hook := verify.BuildHook([]model.Harness{h}, []verify.Adapter{fa}, homeDir)
	if err := hook(); err != nil {
		t.Errorf("hook() = %v, want nil when only Soft checks fail (warnings allowed)", err)
	}
}

// TestBuildHookWithEmptyHarnessesPasses verifies that an empty harness list
// returns a hook that always succeeds.
func TestBuildHookWithEmptyHarnessesPasses(t *testing.T) {
	hook := verify.BuildHook(nil, nil, t.TempDir())
	if err := hook(); err != nil {
		t.Errorf("hook() = %v, want nil for empty harness set", err)
	}
}

// TestBuildHookIsNotImportCycleOnAgents is a compile-time guard.
// If internal/verify imported internal/agents, this package (verify_test) would
// already fail to compile. This test documents the constraint.
func TestBuildHookDoesNotImportAgentsPackage(t *testing.T) {
	// The fakeAdapter in this file satisfies verify.Adapter without any import
	// from internal/agents. If verify imported agents, the test binary would
	// not compile (potential import cycle). This test passing IS the guard.
	var _ verify.Adapter = fakeAdapter{}
}
