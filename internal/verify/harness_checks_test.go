package verify_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/verify"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// newFakeAdapter builds a fakeAdapter wired to a temp dir.
func newFakeAdapter(t *testing.T, homeDir string, agent model.Agent) fakeAdapter {
	t.Helper()
	return fakeAdapter{
		agent:           agent,
		skillsDir:       filepath.Join(homeDir, "skills"),
		instructionsPath: filepath.Join(homeDir, "CLAUDE.md"),
		settingsPath:    filepath.Join(homeDir, "settings.json"),
		mcpConfigPath:   filepath.Join(homeDir, "mcp", "server.json"),
	}
}

// runCheck is a small helper that runs the first check for a harness and
// returns the result's status/error.
func runChecks(t *testing.T, checks []verify.Check) []verify.CheckResult {
	t.Helper()
	return verify.RunChecks(context.Background(), checks)
}

// ─────────────────────────────────────────────────────────────────────────────
// 2.2 — HarnessType == skill
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckForSkillHarness_Pass(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "jr-orchestrator",
		Type: model.HarnessSkill,
	}

	// Create the expected SKILL.md with content.
	skillDir := filepath.Join(fa.SkillsDir(homeDir), h.ID)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill content"), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for skill harness")
	}

	results := runChecks(t, checks)
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed", r.ID, r.Status, r.Error)
		}
	}
}

func TestCheckForSkillHarness_FailMissingSKILLMD(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "jr-orchestrator",
		Type: model.HarnessSkill,
	}

	// Do NOT create SKILL.md — should fail.
	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for skill harness")
	}

	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected at least one failed check when SKILL.md is missing")
	}
}

func TestCheckForSkillHarness_FailEmptySKILLMD(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "jr-orchestrator",
		Type: model.HarnessSkill,
	}

	// Create an empty SKILL.md — should fail (not useful as a skill).
	skillDir := filepath.Join(fa.SkillsDir(homeDir), h.ID)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected at least one failed check when SKILL.md is empty")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2.3 — HarnessType == config
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckForConfigHarness_Pass(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "sdd-orchestrator",
		Type: model.HarnessConfig,
	}

	// Write instructions file with the idempotent marker exactly once.
	marker := "<!-- jr-stack:sdd-orchestrator -->"
	closeMarker := "<!-- /jr-stack:sdd-orchestrator -->"
	content := "# Instructions\n\n" + marker + "\nsome content\n" + closeMarker + "\n"
	if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed", r.ID, r.Status, r.Error)
		}
	}
}

// TestCheckForConfigHarness_PrimaryAgent_Pass verifies that for an agent that
// delivers config as a primary agent (OpenCode), the health check looks in the
// settings JSON for agent.<id> with mode:primary — NOT the AGENTS.md marker,
// which is intentionally absent in this delivery mode.
func TestCheckForConfigHarness_PrimaryAgent_Pass(t *testing.T) {
	homeDir := t.TempDir()
	fa := fakeAdapter{
		agent:            model.AgentOpenCode,
		instructionsPath: filepath.Join(homeDir, "AGENTS.md"), // intentionally has no marker
		settingsPath:     filepath.Join(homeDir, "opencode.json"),
		delivery:         model.ConfigDeliveryPrimaryAgent,
	}

	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}

	settings := `{"agent":{"sdd-orchestrator":{"mode":"primary","prompt":"x"},"build":{"mode":"primary"}}}`
	if err := os.WriteFile(fa.SettingsPath(homeDir), []byte(settings), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	if len(results) == 0 {
		t.Fatal("expected at least one check")
	}
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed", r.ID, r.Status, r.Error)
		}
	}
}

// TestCheckForConfigHarness_PrimaryAgent_FailAbsent verifies the check fails
// when the primary agent entry is missing from the settings JSON.
func TestCheckForConfigHarness_PrimaryAgent_FailAbsent(t *testing.T) {
	homeDir := t.TempDir()
	fa := fakeAdapter{
		agent:            model.AgentOpenCode,
		instructionsPath: filepath.Join(homeDir, "AGENTS.md"),
		settingsPath:     filepath.Join(homeDir, "opencode.json"),
		delivery:         model.ConfigDeliveryPrimaryAgent,
	}

	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}

	// Settings exist but only have an unrelated agent — orchestrator absent.
	if err := os.WriteFile(fa.SettingsPath(homeDir), []byte(`{"agent":{"build":{"mode":"primary"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected failure when primary-agent entry is absent")
	}
}

func TestCheckForConfigHarness_FailMarkerAbsent(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "sdd-orchestrator",
		Type: model.HarnessConfig,
	}

	// Write instructions file WITHOUT the marker.
	if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte("# Instructions\nno marker here\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected failure when marker is absent")
	}
}

func TestCheckForConfigHarness_FailDuplicateMarker(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "sdd-orchestrator",
		Type: model.HarnessConfig,
	}

	// Write instructions file with duplicate markers (idempotency violated).
	marker := "<!-- jr-stack:sdd-orchestrator -->"
	closeMarker := "<!-- /jr-stack:sdd-orchestrator -->"
	content := marker + "\ncontent\n" + closeMarker + "\n" +
		marker + "\ncontent again\n" + closeMarker + "\n"
	if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected failure when marker appears more than once (idempotency violation)")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2.3 — special case: permissions harness → settings path
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckForPermissionsHarness_Pass(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "permissions",
		Type: model.HarnessConfig,
	}

	// Write a settings JSON that contains the permissions key.
	settings := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{"Read"},
		},
	}
	data, _ := json.Marshal(settings)
	if err := os.WriteFile(fa.SettingsPath(homeDir), data, 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed", r.ID, r.Status, r.Error)
		}
	}
}

func TestCheckForPermissionsHarness_FailMissingSettingsFile(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "permissions",
		Type: model.HarnessConfig,
	}

	// Settings file does not exist.
	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected failure when settings file is missing")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-18 — per-agent permissions key resolution
// ─────────────────────────────────────────────────────────────────────────────

// TestCheckPermissionsOpenCodeSingularKey is the RED test for C-18.
// OpenCode writes the key "permission" (singular) per overlays.go L32-33.
// Before the fix, checkPermissions searches "permissions" (plural) and
// MUST fail with a "permissions key not found" style error.
// After the fix it MUST pass.
func TestCheckPermissionsOpenCodeSingularKey(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentOpenCode)

	h := model.Harness{
		ID:   "permissions",
		Type: model.HarnessConfig,
	}

	// Write an opencode.json-style settings file with "permission" (singular),
	// exactly as the permissions installer writes for OpenCode.
	settings := map[string]interface{}{
		"permission": map[string]interface{}{
			"bash": map[string]interface{}{"*": "allow"},
			"read": map[string]interface{}{"*": "allow"},
		},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fa.SettingsPath(homeDir), data, 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for permissions harness")
	}

	results := runChecks(t, checks)
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed (opencode uses singular \"permission\" key)", r.ID, r.Status, r.Error)
		}
	}
}

// TestCheckPermissionsClaudePluralKey guards against regressions on Claude:
// Claude writes "permissions" (plural); the fix must not break it.
func TestCheckPermissionsClaudePluralKey(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:   "permissions",
		Type: model.HarnessConfig,
	}

	settings := map[string]interface{}{
		"permissions": map[string]interface{}{
			"defaultMode": "acceptEdits",
			"deny":        []string{"Bash(rm -rf /)"},
		},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fa.SettingsPath(homeDir), data, 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)
	for _, r := range results {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("check %q: status = %q, error = %q, want passed (claude uses plural \"permissions\" key)", r.ID, r.Status, r.Error)
		}
	}
}

// TestCheckPermissionsFailureMentionsCorrectKey asserts that when the
// agent-specific key is absent the error message names the searched key,
// not a hardcoded string. This covers D4 (self-describing failure message).
func TestCheckPermissionsFailureMentionsCorrectKey(t *testing.T) {
	tests := []struct {
		name        string
		agent       model.Agent
		settingsJSON string
		wantKeyInMsg string
	}{
		{
			name:         "opencode missing singular key",
			agent:        model.AgentOpenCode,
			settingsJSON: `{"someOtherKey": true}`,
			wantKeyInMsg: "permission",
		},
		{
			name:         "claude missing plural key",
			agent:        model.AgentClaude,
			settingsJSON: `{"someOtherKey": true}`,
			wantKeyInMsg: "permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := t.TempDir()
			fa := newFakeAdapter(t, homeDir, tt.agent)

			h := model.Harness{
				ID:   "permissions",
				Type: model.HarnessConfig,
			}

			if err := os.WriteFile(fa.SettingsPath(homeDir), []byte(tt.settingsJSON), 0o644); err != nil {
				t.Fatal(err)
			}

			checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
			results := runChecks(t, checks)

			anyFailed := false
			for _, r := range results {
				if r.Status == verify.CheckStatusFailed {
					anyFailed = true
					if r.Error == "" {
						t.Errorf("check %q: failed but Error is empty", r.ID)
						continue
					}
					if !containsKey(r.Error, tt.wantKeyInMsg) {
						t.Errorf("check %q: error = %q, want it to contain key name %q", r.ID, r.Error, tt.wantKeyInMsg)
					}
				}
			}
			if !anyFailed {
				t.Errorf("expected at least one failed check when key %q is absent from settings", tt.wantKeyInMsg)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-19: best-effort harness — tasks 4.1, 4.2
// ─────────────────────────────────────────────────────────────────────────────

// Task 4.1a — checkSkill for a best-effort harness emits a Check with Soft == true.
func TestCheckForSkillHarness_BestEffort_CheckIsSoft(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:         "find-skill",
		Type:       model.HarnessSkill,
		BestEffort: true,
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for best-effort skill harness")
	}

	for _, c := range checks {
		if !c.Soft {
			t.Errorf("check %q: Soft = false, want true for best-effort harness", c.ID)
		}
	}
}

// Task 4.1b — checkSkill for a NON-best-effort harness emits a Check with Soft == false (regression).
func TestCheckForSkillHarness_NonBestEffort_CheckIsHard(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:         "jr-orchestrator",
		Type:       model.HarnessSkill,
		BestEffort: false,
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for non-best-effort skill harness")
	}

	for _, c := range checks {
		if c.Soft {
			t.Errorf("check %q: Soft = true, want false for non-best-effort harness", c.ID)
		}
	}
}

// Task 4.2 — a best-effort skill check that fails (SKILL.md missing) produces a
// warning (CheckStatusWarning) — not a hard failure — and the report stays Ready.
func TestCheckForSkillHarness_BestEffort_FailIsWarningNotFailure(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)

	h := model.Harness{
		ID:         "find-skill",
		Type:       model.HarnessSkill,
		BestEffort: true,
	}

	// Intentionally do NOT create SKILL.md — the check should warn, not fail.
	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)

	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			t.Errorf("check %q: status = Failed, want Warning for best-effort harness", r.ID)
		}
	}

	// The report must still be Ready (warnings don't flip Ready).
	report := verify.BuildReport(results)
	if !report.Ready {
		t.Errorf("report.Ready = false, want true when only best-effort skill check fails (should be warning)")
	}
}

// containsKey reports whether s contains the quoted or unquoted form of key.
func containsKey(s, key string) bool {
	return containsStr(s, `"`+key+`"`) || containsStr(s, key)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && strings.Contains(s, substr))
}

// ─────────────────────────────────────────────────────────────────────────────
// 2.4 — HarnessType == external
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckForExternalHarness_MCP_Pass(t *testing.T) {
	homeDir := t.TempDir()
	mcpDir := filepath.Join(homeDir, "mcp")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fa := fakeAdapter{
		agent:        model.AgentClaude,
		mcpConfigPath: filepath.Join(mcpDir, "server.json"),
	}

	h := model.Harness{
		ID:       "context7",
		Type:     model.HarnessExternal,
		External: &model.External{Method: "mcp", URL: "https://mcp.context7.com"},
	}

	// Write a valid JSON MCP config.
	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"context7": map[string]interface{}{
				"url": "https://mcp.context7.com",
			},
		},
	}
	data, _ := json.Marshal(mcpConfig)
	if err := os.WriteFile(fa.MCPConfigPath(homeDir, h.ID), data, 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	// MCP config check is hard; binary check is Soft.
	hasSoftCheck := false
	for _, c := range checks {
		if c.Soft {
			hasSoftCheck = true
		}
	}
	if !hasSoftCheck {
		t.Error("expected at least one Soft check (binary/endpoint) for external harness")
	}

	results := runChecks(t, checks)
	report := verify.BuildReport(results)
	// MCP file exists and is valid JSON, so the hard check passes.
	// The soft check may warn (binary not in PATH in test env) — that's OK.
	if report.Failed > 0 {
		t.Errorf("expected no hard failures for valid MCP config, got %d failures: %+v", report.Failed, results)
	}
}

func TestCheckForExternalHarness_MCP_FailInvalidJSON(t *testing.T) {
	homeDir := t.TempDir()
	mcpDir := filepath.Join(homeDir, "mcp")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}

	fa := fakeAdapter{
		agent:        model.AgentClaude,
		mcpConfigPath: filepath.Join(mcpDir, "server.json"),
	}

	h := model.Harness{
		ID:       "context7",
		Type:     model.HarnessExternal,
		External: &model.External{Method: "mcp", URL: "https://mcp.context7.com"},
	}

	// Write invalid JSON.
	if err := os.WriteFile(fa.MCPConfigPath(homeDir, h.ID), []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	results := runChecks(t, checks)

	anyFailed := false
	for _, r := range results {
		if r.Status == verify.CheckStatusFailed {
			anyFailed = true
		}
	}
	if !anyFailed {
		t.Error("expected hard failure when MCP config JSON is invalid")
	}
}

func TestCheckForExternalHarness_SoftCheckProducesWarningNotFailure(t *testing.T) {
	homeDir := t.TempDir()

	fa := fakeAdapter{
		agent:        model.AgentClaude,
		mcpConfigPath: filepath.Join(homeDir, "mcp", "engram.json"),
	}

	// Engram uses homebrew, not mcp — its binary check is Soft.
	h := model.Harness{
		ID:       "engram",
		Type:     model.HarnessExternal,
		External: &model.External{Method: "homebrew", Pkg: "engram"},
	}

	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)

	// All checks for a homebrew harness should be Soft (binary in PATH).
	for _, c := range checks {
		if !c.Soft {
			t.Errorf("check %q should be Soft for homebrew external harness", c.ID)
		}
	}

	results := runChecks(t, checks)
	report := verify.BuildReport(results)

	// The binary is not installed in test env, so we get warnings — but NOT failures.
	if report.Failed > 0 {
		t.Errorf("Soft checks must not produce failures: got %d failures", report.Failed)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2.2 — Multiple adapters: one check per (harness × agent)
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckForSkillHarness_MultipleAdapters(t *testing.T) {
	homeDir := t.TempDir()

	// Two adapters for two agents.
	claudeSkillsDir := filepath.Join(homeDir, "claude", "skills")
	opencodeSkillsDir := filepath.Join(homeDir, "opencode", "skills")

	claudeAdapter := fakeAdapter{
		agent:     model.AgentClaude,
		skillsDir: claudeSkillsDir,
	}
	opencodeAdapter := fakeAdapter{
		agent:     model.AgentOpenCode,
		skillsDir: opencodeSkillsDir,
	}

	h := model.Harness{
		ID:   "jr-orchestrator",
		Type: model.HarnessSkill,
	}

	// Create SKILL.md for claude only.
	skillDir := filepath.Join(claudeSkillsDir, h.ID)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# content"), 0o644); err != nil {
		t.Fatal(err)
	}
	// opencode's SKILL.md is missing.

	checks := verify.ChecksForHarness(h, []verify.Adapter{claudeAdapter, opencodeAdapter}, homeDir)

	// Should have checks for both agents.
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 checks (one per adapter), got %d", len(checks))
	}

	results := runChecks(t, checks)
	report := verify.BuildReport(results)

	// Claude passes, opencode fails → report not ready.
	if report.Ready {
		t.Error("Ready = true, want false when opencode SKILL.md is missing")
	}
}
