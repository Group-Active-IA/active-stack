package permissions_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config/permissions"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// --- stub adapters ---

type stubAdapter struct {
	agent       model.Agent
	settingsDir []string // path segments relative to homeDir; empty = no-op
}

func (s stubAdapter) Agent() model.Agent { return s.agent }
func (s stubAdapter) SettingsPath(homeDir string) string {
	if len(s.settingsDir) == 0 {
		return ""
	}
	parts := append([]string{homeDir}, s.settingsDir...)
	return filepath.Join(parts...)
}

func claudeAdapter() permissions.PermissionsAdapter {
	return stubAdapter{
		agent:       model.AgentClaude,
		settingsDir: []string{".claude", "settings.json"},
	}
}

func openCodeAdapter() permissions.PermissionsAdapter {
	return stubAdapter{
		agent:       model.AgentOpenCode,
		settingsDir: []string{".config", "opencode", "opencode.json"},
	}
}

func geminiAdapter() permissions.PermissionsAdapter {
	return stubAdapter{
		agent:       model.AgentGemini,
		settingsDir: []string{".gemini", "settings.json"},
	}
}

func vsCodeAdapter() permissions.PermissionsAdapter {
	return stubAdapter{
		agent:       model.AgentVSCode,
		settingsDir: []string{".config", "Code", "User", "settings.json"},
	}
}

func cursorAdapter() permissions.PermissionsAdapter {
	// cursor returns empty path to signal no-op
	return stubAdapter{agent: model.AgentCursor}
}

func codexAdapter() permissions.PermissionsAdapter {
	return stubAdapter{agent: model.AgentCodex}
}

func antigravityAdapter() permissions.PermissionsAdapter {
	return stubAdapter{agent: model.AgentAntigravity}
}

func windsurfAdapter() permissions.PermissionsAdapter {
	return stubAdapter{agent: model.AgentWindsurf}
}

// ── Task 2.1: securityFloorDeny contains the five required rules ──────────────

// TestSecurityFloorDenyContainsRequiredRules verifies that SecurityFloorDeny
// (the single source of truth) contains all five required deny rules from the spec
// (spec: "El piso proviene de una sola fuente", permission-tier §D2).
func TestSecurityFloorDenyContainsRequiredRules(t *testing.T) {
	floor := permissions.SecurityFloorDeny()

	required := []string{
		"Read(.env)",
		"Read(.env.*)",
		"Bash(rm -rf /)",
		"Bash(rm -rf ~)",
		"Bash(git push --force:*)",
	}

	floorSet := make(map[string]bool, len(floor))
	for _, rule := range floor {
		floorSet[rule] = true
	}

	for _, r := range required {
		if !floorSet[r] {
			t.Errorf("security floor missing required rule %q", r)
		}
	}
}

// --- Task 4.1 (updated): Claude tier-balanceado uses defaultMode default + allow-list ---
// The prior test checked for "acceptEdits" — that was the old fixed behavior.
// This change resolves the spec↔code drift: defaultMode is now derived from the tier.
// Balanceado produces "default" + allow-list; bypass produces "bypassPermissions".

func TestInstallClaudeCode(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Install() changed = false; expected first install to write the file")
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal settings.json: %v", err)
	}

	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("permissions node missing or wrong type: %#v", settings["permissions"])
	}

	// Balanceado tier → defaultMode must be "default" (NOT bypassPermissions, NOT acceptEdits).
	mode, ok := perms["defaultMode"].(string)
	if !ok {
		t.Fatalf("defaultMode missing from permissions")
	}
	if mode != "default" {
		t.Fatalf("expected defaultMode=default for balanceado tier, got %q", mode)
	}

	// deny list must exist and contain the security floor.
	denyList, ok := perms["deny"].([]any)
	if !ok {
		t.Fatalf("deny list missing or wrong type: %#v", perms["deny"])
	}

	wantDeny := []string{
		"Read(.env)",
		"Read(.env.*)",
		"Bash(rm -rf /)",
		"Bash(rm -rf ~)",
		"Bash(git push --force:*)",
	}

	denySet := make(map[string]bool, len(denyList))
	for _, entry := range denyList {
		if s, ok := entry.(string); ok {
			denySet[s] = true
		}
	}

	for _, want := range wantDeny {
		if !denySet[want] {
			t.Errorf("deny list missing %q", want)
		}
	}
}

// --- Task 4.2 (updated): OpenCode idempotent ---

func TestInstallOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Install() first changed = false")
	}

	second, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() second error = %v", err)
	}
	if second.Changed {
		t.Fatal("Install() second changed = true; expected idempotent (no change)")
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, `"permission"`) {
		t.Fatal("opencode.json missing permission key")
	}
	if strings.Contains(text, `"permissions"`) {
		t.Fatal("opencode.json uses 'permissions' (plural) — should be 'permission' (singular)")
	}
}

// --- Task 4.3 (updated): OpenCode balanceado tier — wildcard default is ask ---
// The prior test checked for specific git push/rebase keys in the bash object.
// The new design uses a wildcard {"*": "ask"} as the base for balanceado, so ALL
// unmatched commands (including git push) are "ask" by default. The specific
// allow-list only whitelists go test/build and git status/diff.

func TestInstallOpenCodeGitDefaultsToAsk(t *testing.T) {
	home := t.TempDir()

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	perm, ok := settings["permission"].(map[string]any)
	if !ok {
		t.Fatalf("permission node missing: %#v", settings)
	}

	bash, ok := perm["bash"].(map[string]any)
	if !ok {
		t.Fatalf("permission.bash missing: %#v", perm)
	}

	// Balanceado: wildcard "*" must be "ask" (destructive commands fall through to ask).
	if bash["*"] != "ask" {
		t.Errorf(`bash["*"] = %v, want "ask"`, bash["*"])
	}

	// Explicitly allowed safe operations.
	safeAllow := []string{"go test *", "go build *", "git status", "git diff *"}
	for _, cmd := range safeAllow {
		if bash[cmd] != "allow" {
			t.Errorf("bash[%q] = %v, want \"allow\"", cmd, bash[cmd])
		}
	}
}

// --- Task 4.4: Gemini uses auto_edit, no permissions key ---

func TestInstallGeminiCLI(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{geminiAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Install() changed = false")
	}

	settingsPath := filepath.Join(home, ".gemini", "settings.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	general, ok := settings["general"].(map[string]any)
	if !ok {
		t.Fatalf("general node missing: %#v", settings)
	}

	mode, ok := general["defaultApprovalMode"].(string)
	if !ok || mode != "auto_edit" {
		t.Fatalf("expected defaultApprovalMode=auto_edit, got %q", mode)
	}

	if _, exists := settings["permissions"]; exists {
		t.Fatal("gemini settings.json must NOT contain 'permissions' key (Claude key leak)")
	}
}

// --- Task 4.5: VSCode merges into JSONC, preserves existing settings ---

func TestInstallVSCodeCopilotMergesJSONC(t *testing.T) {
	home := t.TempDir()

	adapter := vsCodeAdapter()
	settingsPath := adapter.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Pre-existing JSONC with comments and trailing commas.
	baseSettings := `{
  // User has comments and trailing commas in VS Code settings
  "editor.formatOnSave": true,
  "files.exclude": {
    "**/.git": true,
  },
}
`
	if err := os.WriteFile(settingsPath, []byte(baseSettings), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{adapter}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if !result.Changed {
		t.Fatal("Install() changed = false")
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	autoApprove, ok := settings["chat.tools.autoApprove"].(bool)
	if !ok || !autoApprove {
		t.Fatalf("expected chat.tools.autoApprove=true, got %v", settings["chat.tools.autoApprove"])
	}

	if settings["editor.formatOnSave"] != true {
		t.Fatalf("editor.formatOnSave lost after merge, got %v", settings["editor.formatOnSave"])
	}
}

// --- Task 4.6: No-op agents return Changed=false, Files=[] ---

func TestInstallCursorNoOp(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{cursorAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Cursor: expected Changed=false (no-op)")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Cursor: expected Files=[], got %v", result.Files)
	}
}

func TestInstallCodexNoOp(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{codexAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Codex: expected Changed=false (no-op)")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Codex: expected Files=[], got %v", result.Files)
	}
}

func TestInstallAntigravityNoOp(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{antigravityAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Antigravity: expected Changed=false (no-op)")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Antigravity: expected Files=[], got %v", result.Files)
	}
}

func TestInstallWindsurfNoOp(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{windsurfAdapter()}, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Changed {
		t.Fatal("Windsurf: expected Changed=false (no-op)")
	}
	if len(result.Files) != 0 {
		t.Fatalf("Windsurf: expected Files=[], got %v", result.Files)
	}
}

// --- Task 4.7: Backup created before write ---

func TestInstallBackupCreatedBeforeWrite(t *testing.T) {
	home := t.TempDir()

	// Pre-create a settings.json so the backup has something to snapshot.
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	backupDir := filepath.Join(home, ".jr-stack", "backups", fmt.Sprintf("permissions-%s", string(model.AgentClaude)))
	info, err := os.Stat(backupDir)
	if err != nil {
		t.Fatalf("backup dir not created at %q: %v", backupDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected backup dir to be a directory, got file")
	}

	// Check that manifest.json exists (proof a real backup ran).
	manifestPath := filepath.Join(backupDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("manifest.json not found in backup dir: %v", err)
	}
}

// --- Task 4.8: Backup failure aborts write ---

func TestInstallBackupFailureAbortsWrite(t *testing.T) {
	home := t.TempDir()

	// Pre-create a settings.json.
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	original := []byte(`{"existing": true}`)
	if err := os.WriteFile(settingsPath, original, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Inject a snapshotter that always fails.
	restore := permissions.SetSnapshotterCreate(func(snapshotDir string, paths []string) error {
		return fmt.Errorf("injected backup failure")
	})
	defer restore()

	_, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado)
	if err == nil {
		t.Fatal("expected error when backup fails, got nil")
	}

	// settings.json must not have changed.
	content, err2 := os.ReadFile(settingsPath)
	if err2 != nil {
		t.Fatalf("read settings.json: %v", err2)
	}
	if string(content) != string(original) {
		t.Fatalf("settings.json was modified after backup failure\ngot: %s\nwant: %s", content, original)
	}
}

// ── Task 2.2: Claude overlay composition by tier ─────────────────────────────

// TestComposeClaudeOverlayByTier verifies the defaultMode and allow-list for
// each of the three Claude tiers (spec: permissions-harness — escenarios de tier).
func TestComposeClaudeOverlayByTier(t *testing.T) {
	tests := []struct {
		name            string
		tier            model.PermissionTier
		wantDefaultMode string
		wantAllowList   bool // true = allow-list should be present
	}{
		{
			name:            "estricto produces default, no allow-list",
			tier:            model.TierEstricto,
			wantDefaultMode: "default",
			wantAllowList:   false,
		},
		{
			name:            "balanceado produces default with allow-list",
			tier:            model.TierBalanceado,
			wantDefaultMode: "default",
			wantAllowList:   true,
		},
		{
			name:            "bypass produces bypassPermissions",
			tier:            model.TierBypass,
			wantDefaultMode: "bypassPermissions",
			wantAllowList:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, tt.tier); err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			settingsPath := filepath.Join(home, ".claude", "settings.json")
			content, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read settings.json: %v", err)
			}

			var settings map[string]any
			if err := json.Unmarshal(content, &settings); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			perms, ok := settings["permissions"].(map[string]any)
			if !ok {
				t.Fatalf("permissions node missing or wrong type: %#v", settings["permissions"])
			}

			mode, ok := perms["defaultMode"].(string)
			if !ok {
				t.Fatalf("defaultMode missing")
			}
			if mode != tt.wantDefaultMode {
				t.Errorf("defaultMode = %q, want %q", mode, tt.wantDefaultMode)
			}

			_, hasAllow := perms["allow"]
			if hasAllow != tt.wantAllowList {
				t.Errorf("allow-list present=%v, want %v", hasAllow, tt.wantAllowList)
			}

			// Balanceado allow-list must contain at least Bash(go test:*).
			if tt.wantAllowList {
				allowList, ok := perms["allow"].([]any)
				if !ok {
					t.Fatal("allow is not a list")
				}
				found := false
				for _, a := range allowList {
					if a == "Bash(go test:*)" {
						found = true
					}
				}
				if !found {
					t.Error("allow-list missing Bash(go test:*)")
				}
			}
		})
	}
}

// ── Task 2.3: Security floor in all three Claude tiers ─────────────────────

// TestSecurityFloorInAllClaudeTiers verifies that the security floor deny rules
// are present in all three tiers (including bypass) for Claude Code.
// Spec: "Deny floor sobrevive bajo tier bypass".
func TestSecurityFloorInAllClaudeTiers(t *testing.T) {
	requiredFloor := []string{
		"Read(.env)",
		"Bash(rm -rf /)",
		"Bash(rm -rf ~)",
	}

	tiers := []model.PermissionTier{model.TierEstricto, model.TierBalanceado, model.TierBypass}
	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, tier); err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			settingsPath := filepath.Join(home, ".claude", "settings.json")
			content, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read settings.json: %v", err)
			}

			var settings map[string]any
			if err := json.Unmarshal(content, &settings); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			perms, ok := settings["permissions"].(map[string]any)
			if !ok {
				t.Fatalf("permissions node missing")
			}

			denyList, ok := perms["deny"].([]any)
			if !ok {
				t.Fatalf("deny list missing or wrong type")
			}

			denySet := make(map[string]bool, len(denyList))
			for _, entry := range denyList {
				if s, ok := entry.(string); ok {
					denySet[s] = true
				}
			}

			for _, rule := range requiredFloor {
				if !denySet[rule] {
					t.Errorf("tier %q deny list missing floor rule %q", tier, rule)
				}
			}
		})
	}
}

// ── Task 2.4: Zero-value tier composes balanceado ─────────────────────────

// TestZeroValueTierComposesBalanceado verifies that an empty tier string normalizes
// to balanceado (never bypass) when composing the Claude overlay.
// Spec: "Sin tier especificado compone balanceado".
func TestZeroValueTierComposesBalanceado(t *testing.T) {
	home := t.TempDir()
	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, ""); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("permissions node missing")
	}

	mode, ok := perms["defaultMode"].(string)
	if !ok {
		t.Fatalf("defaultMode missing")
	}
	// Must be "default" (balanceado) — NEVER "bypassPermissions".
	if mode == "bypassPermissions" {
		t.Fatal("zero-value tier produced bypassPermissions — must normalize to balanceado")
	}
	if mode != "default" {
		t.Errorf("zero-value tier defaultMode = %q, want \"default\" (balanceado)", mode)
	}

	// Must have an allow-list (balanceado property).
	if _, hasAllow := perms["allow"]; !hasAllow {
		t.Error("zero-value tier did not produce allow-list — expected balanceado behavior")
	}
}

// ── Task 3.1: OpenCode overlay composition by tier ────────────────────────

// TestComposeOpencodeOverlayByTier verifies the opencode bash pattern-object
// and per-tool action strings for each tier.
func TestComposeOpencodeOverlayByTier(t *testing.T) {
	tests := []struct {
		name          string
		tier          model.PermissionTier
		wantReadAction string
		wantBashWildcard string
	}{
		{
			name:             "estricto — all ask",
			tier:             model.TierEstricto,
			wantReadAction:   "ask",
			wantBashWildcard: "ask",
		},
		{
			name:             "balanceado — read allow, bash wildcard ask",
			tier:             model.TierBalanceado,
			wantReadAction:   "allow",
			wantBashWildcard: "ask",
		},
		{
			name:             "bypass — all allow",
			tier:             model.TierBypass,
			wantReadAction:   "allow",
			wantBashWildcard: "allow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, tt.tier); err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
			content, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read opencode.json: %v", err)
			}

			var settings map[string]any
			if err := json.Unmarshal(content, &settings); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			perm, ok := settings["permission"].(map[string]any)
			if !ok {
				t.Fatalf("permission node missing: %#v", settings)
			}

			// Check read action.
			readAction, ok := perm["read"].(string)
			if !ok {
				t.Fatalf("permission.read missing")
			}
			if readAction != tt.wantReadAction {
				t.Errorf("permission.read = %q, want %q", readAction, tt.wantReadAction)
			}

			// Check bash wildcard.
			bash, ok := perm["bash"].(map[string]any)
			if !ok {
				t.Fatalf("permission.bash missing")
			}
			if bash["*"] != tt.wantBashWildcard {
				t.Errorf(`bash["*"] = %v, want %q`, bash["*"], tt.wantBashWildcard)
			}
		})
	}
}

// ── Task 3.2: Deny ordering — last-wins guarantee for opencode ────────────

// TestOpencodeLastWinsDenyOrdering verifies that in ALL three tiers, deny rules
// appear AFTER the wildcard "*" in the bash pattern-object when serialized.
// This is the CRITICAL last-match-wins guarantee.
// Spec: "El deny floor de opencode se posiciona último para no ser anulado".
func TestOpencodeLastWinsDenyOrdering(t *testing.T) {
	tiers := []model.PermissionTier{model.TierEstricto, model.TierBalanceado, model.TierBypass}

	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, tier); err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
			raw, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read opencode.json: %v", err)
			}

			// Extract the ordered list of key→value pairs in permission.bash
			// by scanning the raw JSON bytes for key order (position-based).
			pairs, err := extractBashOrderedPairs(raw)
			if err != nil {
				t.Fatalf("extractBashOrderedPairs: %v", err)
			}

			if len(pairs) == 0 {
				t.Fatal("permission.bash is empty")
			}

			// Find the index of the wildcard "*" and the first deny rule.
			wildcardIdx := -1
			firstDenyIdx := -1
			for i, p := range pairs {
				if p[0] == "*" {
					wildcardIdx = i
				}
				if p[1] == "deny" && firstDenyIdx == -1 {
					firstDenyIdx = i
				}
			}

			if wildcardIdx == -1 {
				t.Fatal(`bash pattern-object missing wildcard "*" key`)
			}
			if firstDenyIdx == -1 {
				t.Fatal("bash pattern-object missing any deny rule")
			}

			// THE CRITICAL ASSERTION: wildcard index MUST be BEFORE first deny index.
			if wildcardIdx >= firstDenyIdx {
				t.Errorf("tier %q: wildcard \"*\" at index %d is NOT before first deny at index %d (last-wins violated)",
					tier, wildcardIdx, firstDenyIdx)
			}
		})
	}
}

// extractBashOrderedPairs parses the opencode.json and returns the ordered list
// of [key, value] string pairs from the "permission.bash" object, preserving
// insertion order. Uses json.Decoder's token stream to maintain order.
func extractBashOrderedPairs(data []byte) ([][2]string, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))

	depth := 0
	inPermission := false
	inBash := false
	expectingKey := true
	var pendingKey string
	var result [][2]string

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch v := tok.(type) {
		case json.Delim:
			if v == '{' {
				depth++
				expectingKey = true
			} else if v == '}' {
				depth--
				if inBash && depth == 2 {
					// Exiting the bash object.
					return result, nil
				}
				if inPermission && depth == 1 {
					inPermission = false
				}
				expectingKey = true
			}
		case string:
			if depth == 1 && expectingKey && v == "permission" {
				inPermission = true
				expectingKey = false
			} else if inPermission && depth == 2 && expectingKey && v == "bash" {
				inBash = true
				expectingKey = false
			} else if inBash && depth == 3 {
				if expectingKey {
					pendingKey = v
					expectingKey = false
				} else {
					// This is the value for pendingKey.
					result = append(result, [2]string{pendingKey, v})
					expectingKey = true
				}
			} else {
				expectingKey = !expectingKey
			}
		default:
			expectingKey = true
		}
	}

	if len(result) > 0 {
		return result, nil
	}
	return nil, fmt.Errorf("permission.bash not found in opencode.json")
}

// ── Task 3.3: OpenCode does NOT receive Claude shape ──────────────────────

// TestOpencodeDoesNotReceiveClaudeShape verifies that the opencode.json output
// does NOT contain the "permissions" (plural) key with defaultMode/allow/deny.
// Spec: "OpenCode no recibe el shape de Claude Code".
func TestOpencodeDoesNotReceiveClaudeShape(t *testing.T) {
	tiers := []model.PermissionTier{model.TierEstricto, model.TierBalanceado, model.TierBypass}

	for _, tier := range tiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, tier); err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
			content, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read opencode.json: %v", err)
			}

			text := string(content)
			// Must NOT have "permissions" (plural — Claude shape).
			if strings.Contains(text, `"permissions"`) {
				t.Errorf("tier %q: opencode.json contains 'permissions' (Claude shape leak)", tier)
			}
			// Must NOT have "defaultMode" (Claude property).
			if strings.Contains(text, `"defaultMode"`) {
				t.Errorf("tier %q: opencode.json contains 'defaultMode' (Claude property leak)", tier)
			}
			// MUST have "permission" (singular — opencode shape).
			if !strings.Contains(text, `"permission"`) {
				t.Errorf("tier %q: opencode.json missing 'permission' (singular) key", tier)
			}
		})
	}
}

// ── Task 3b.1: Gemini and VSCode unaffected by tier ──────────────────────

// TestGeminiVSCodeUnaffectedByTier verifies that gemini and vscode overlays
// are NOT affected by the tier — they keep their fixed overlay behavior and do
// NOT receive the Claude "permissions" shape.
// Spec: "Gemini y VS Code no se ven afectados por el tier elegido",
//       "Gemini y VS Code no reciben el shape de Claude".
func TestGeminiVSCodeUnaffectedByTier(t *testing.T) {
	agents := []struct {
		name    string
		adapter permissions.PermissionsAdapter
		pathFn  func(home string) string
		wantKey string // key that should be present (not Claude's "permissions")
	}{
		{
			name:    "gemini",
			adapter: geminiAdapter(),
			pathFn:  func(home string) string { return filepath.Join(home, ".gemini", "settings.json") },
			wantKey: `"general"`,
		},
		{
			name:    "vscode",
			adapter: vsCodeAdapter(),
			pathFn: func(home string) string {
				return filepath.Join(home, ".config", "Code", "User", "settings.json")
			},
			wantKey: `"chat.tools.autoApprove"`,
		},
	}

	tiers := []model.PermissionTier{model.TierEstricto, model.TierBalanceado, model.TierBypass}

	for _, ag := range agents {
		for _, tier := range tiers {
			name := fmt.Sprintf("%s/%s", ag.name, tier)
			t.Run(name, func(t *testing.T) {
				home := t.TempDir()
				if _, err := permissions.Install(home, []permissions.PermissionsAdapter{ag.adapter}, tier); err != nil {
					t.Fatalf("Install() error = %v", err)
				}

				content, err := os.ReadFile(ag.pathFn(home))
				if err != nil {
					t.Fatalf("read settings: %v", err)
				}

				text := string(content)

				// Must NOT have "permissions" (plural — Claude shape).
				if strings.Contains(text, `"permissions"`) {
					t.Errorf("%s tier %q: settings contain 'permissions' (Claude shape leak)", ag.name, tier)
				}
				// Must NOT have "defaultMode" (Claude property).
				if strings.Contains(text, `"defaultMode"`) && ag.name != "gemini" {
					// gemini uses defaultApprovalMode (contains "Mode") but not "defaultMode"
					t.Errorf("%s tier %q: settings contain 'defaultMode' (Claude property)", ag.name, tier)
				}
				// Must have the expected key for this agent.
				if !strings.Contains(text, ag.wantKey) {
					t.Errorf("%s tier %q: missing expected key %q", ag.name, tier, ag.wantKey)
				}
			})
		}
	}
}
