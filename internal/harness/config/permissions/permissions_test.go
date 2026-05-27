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

// --- Task 4.1: Claude uses acceptEdits + deny list ---

func TestInstallClaudeCode(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()})
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

	// --- OVERRIDE from orchestrator: must be acceptEdits, NOT bypassPermissions ---
	mode, ok := perms["defaultMode"].(string)
	if !ok {
		t.Fatalf("defaultMode missing from permissions")
	}
	if mode == "bypassPermissions" {
		t.Fatal("claude overlay uses bypassPermissions — MUST be acceptEdits per orchestrator decision")
	}
	if mode != "acceptEdits" {
		t.Fatalf("expected defaultMode=acceptEdits, got %q", mode)
	}

	// deny list must exist and contain Read(.env)
	denyList, ok := perms["deny"].([]any)
	if !ok {
		t.Fatalf("deny list missing or wrong type: %#v", perms["deny"])
	}

	wantDeny := []string{
		"Read(.env)",
		"Read(.env.*)",
		"Edit(.env)",
		"Edit(.env.*)",
		"Bash(rm -rf /)",
		"Bash(sudo rm -rf /)",
		"Bash(rm -rf ~)",
		"Bash(sudo rm -rf ~)",
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

// --- Task 4.2: OpenCode idempotent ---

func TestInstallOpenCodeIsIdempotent(t *testing.T) {
	home := t.TempDir()

	first, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()})
	if err != nil {
		t.Fatalf("Install() first error = %v", err)
	}
	if !first.Changed {
		t.Fatal("Install() first changed = false")
	}

	second, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()})
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

// --- Task 4.3: OpenCode git destructive requires ask ---

func TestInstallOpenCodeGitDestructiveRequiresAsk(t *testing.T) {
	home := t.TempDir()

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}); err != nil {
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

	destructive := []string{
		"git push *",
		"git push --force *",
		"git rebase *",
		"git reset --hard *",
	}

	for _, cmd := range destructive {
		val, ok := bash[cmd]
		if !ok {
			t.Errorf("bash[%q] missing", cmd)
			continue
		}
		if val != "ask" {
			t.Errorf("bash[%q] = %q, want \"ask\"", cmd, val)
		}
	}
}

// --- Task 4.4: Gemini uses auto_edit, no permissions key ---

func TestInstallGeminiCLI(t *testing.T) {
	home := t.TempDir()

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{geminiAdapter()})
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

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{adapter})
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

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{cursorAdapter()})
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

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{codexAdapter()})
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

	result, err := permissions.Install(home, []permissions.PermissionsAdapter{antigravityAdapter()})
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

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}); err != nil {
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

	_, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()})
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
