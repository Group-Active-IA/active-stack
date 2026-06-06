package command_test

// ─────────────────────────────────────────────────────────────────────────────
// Tests for internal/harness/command
//
// TDD order: RED → GREEN → TRIANGULATE → REFACTOR.
// Mirrors internal/harness/skill/skill_test.go conventions.
// All file-system operations use t.TempDir().
// ─────────────────────────────────────────────────────────────────────────────

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/command"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// fakeAdapter implements command.AgentAdapter for tests.
// commandsDir controls the resolved directory; set to "" to test skip behavior.
type fakeAdapter struct {
	agent       model.Agent
	commandsDir string   // empty → skip
	variantKey  string
}

func (f fakeAdapter) Agent() model.Agent                  { return f.agent }
func (f fakeAdapter) CommandsDir(_ string) string         { return f.commandsDir }
func (f fakeAdapter) VariantKey() string                  { return f.variantKey }

// makeCommandsFS returns a minimal fstest.MapFS that contains both the Claude
// and OpenCode command variants at the expected asset paths.
func makeCommandsFS() fstest.MapFS {
	return fstest.MapFS{
		"commands/claude/jr/starter-add.md": {
			Data: []byte("---\nname: JR: Starter Add\ndescription: test\ncategory: Workflow\ntags:\n  - jr-stack\n---\njr-stack starter add $ARGUMENTS"),
		},
		"commands/opencode/jr-starter-add.md": {
			Data: []byte("---\ndescription: test\n---\njr-stack starter add $ARGUMENTS"),
		},
	}
}

// relPath returns the relative command file path for the given adapter's variant.
// Claude  → "jr/starter-add.md" (namespaced under a subdir)
// OpenCode→ "jr-starter-add.md" (flat, hyphenated)
func relPath(variantKey string) string {
	switch variantKey {
	case "claude":
		return filepath.Join("jr", "starter-add.md")
	case "opencode":
		return "jr-starter-add.md"
	default:
		return ""
	}
}

// ── §5.1 RED: installer writes the command file ───────────────────────────────

// TestInstaller_WritesCommandFileForAdapter asserts that Install writes the
// per-agent command file to the adapter's commandsDir.
// Mirrors: "Command is written for each focused adapter" scenario.
func TestInstaller_WritesCommandFileForAdapter(t *testing.T) {
	for _, tc := range []struct {
		name       string
		agentAgent model.Agent
		variant    string
		assetPath  string  // path inside CommandsFS
		expectRel  string  // relative path under commandsDir
	}{
		{
			name:       "claude",
			agentAgent: model.AgentClaude,
			variant:    "claude",
			assetPath:  "commands/claude/jr/starter-add.md",
			expectRel:  filepath.Join("jr", "starter-add.md"),
		},
		{
			name:       "opencode",
			agentAgent: model.AgentOpenCode,
			variant:    "opencode",
			assetPath:  "commands/opencode/jr-starter-add.md",
			expectRel:  "jr-starter-add.md",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			commandsDir := t.TempDir()
			backupDir := t.TempDir()
			adapter := fakeAdapter{agent: tc.agentAgent, commandsDir: commandsDir, variantKey: tc.variant}
			commandsFS := makeCommandsFS()

			ins := command.NewInstaller(commandsFS)
			results, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
			if err != nil {
				t.Fatalf("Install() error: %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			if results[0].AlreadyInstalled {
				t.Error("expected not AlreadyInstalled on fresh install")
			}

			destFile := filepath.Join(commandsDir, tc.expectRel)
			data, err := os.ReadFile(destFile)
			if err != nil {
				t.Fatalf("expected command file at %q: %v", destFile, err)
			}
			if !strings.Contains(string(data), "jr-stack starter add") {
				t.Errorf("written file should contain 'jr-stack starter add'; got:\n%s", data)
			}
		})
	}
}

// TestInstaller_SkipsAdapterWithEmptyCommandsDir asserts that an adapter whose
// CommandsDir returns "" is skipped silently (no error, no result).
// Mirrors: "Adapter with empty command directory is skipped" scenario.
func TestInstaller_SkipsAdapterWithEmptyCommandsDir(t *testing.T) {
	backupDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, commandsDir: "", variantKey: "claude"}
	commandsFS := makeCommandsFS()

	ins := command.NewInstaller(commandsFS)
	results, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
	if err != nil {
		t.Fatalf("Install() error for empty commandsDir: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when commandsDir is empty, got %d", len(results))
	}
}

// ── §5.3 RED: idempotence — identical content is a no-op ────────────────────

// TestInstaller_Idempotent_IdenticalContent asserts that a second install
// with byte-identical content does NOT rewrite and reports AlreadyInstalled.
// Mirrors: "Identical content is a no-op" scenario.
func TestInstaller_Idempotent_IdenticalContent(t *testing.T) {
	commandsDir := t.TempDir()
	backupDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, commandsDir: commandsDir, variantKey: "claude"}
	commandsFS := makeCommandsFS()
	ins := command.NewInstaller(commandsFS)

	// First install.
	_, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
	if err != nil {
		t.Fatalf("first Install() error: %v", err)
	}

	// Record mtime before the second install.
	destFile := filepath.Join(commandsDir, "jr", "starter-add.md")
	info1, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("stat after first install: %v", err)
	}

	// Second install — identical content.
	results2, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
	if err != nil {
		t.Fatalf("second Install() error: %v", err)
	}
	if len(results2) != 1 {
		t.Fatalf("expected 1 result on second install, got %d", len(results2))
	}
	if !results2[0].AlreadyInstalled {
		t.Error("expected AlreadyInstalled=true on second install with identical content")
	}

	// File must NOT have been rewritten (mtime identical; content unchanged).
	info2, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("stat after second install: %v", err)
	}
	if info2.ModTime() != info1.ModTime() {
		// ModTime may differ by < 1ns on some platforms; compare content instead.
		data, _ := os.ReadFile(destFile)
		orig, _ := commandsFS.ReadFile("commands/claude/jr/starter-add.md")
		if !strings.EqualFold(string(data), string(orig)) {
			t.Error("file content changed on second install (should be no-op)")
		}
	}
}

// ── §5.5 RED: backup — changed content backed up before overwrite ──────────

// TestInstaller_Backup_ChangedContent asserts that when the destination
// file exists with DIFFERENT content, a backup is taken and the file is
// overwritten with new content.
// Mirrors: "Changed content overwrites after backup" scenario.
func TestInstaller_Backup_ChangedContent(t *testing.T) {
	commandsDir := t.TempDir()
	backupDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, commandsDir: commandsDir, variantKey: "claude"}
	commandsFS := makeCommandsFS()
	ins := command.NewInstaller(commandsFS)

	// Pre-plant a different file at the destination.
	destFile := filepath.Join(commandsDir, "jr", "starter-add.md")
	if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	oldContent := []byte("# old version — DIFFERENT content")
	if err := os.WriteFile(destFile, oldContent, 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	// Install should backup and overwrite.
	results, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AlreadyInstalled {
		t.Error("expected AlreadyInstalled=false when content differs")
	}

	// File must now contain new content.
	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("read after overwrite: %v", err)
	}
	if !strings.Contains(string(data), "jr-stack starter add") {
		t.Errorf("overwritten file should contain 'jr-stack starter add'; got:\n%s", data)
	}

	// backupDir must contain at least one file (the backup of the old content).
	var backupFiles []string
	_ = filepath.WalkDir(backupDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			backupFiles = append(backupFiles, p)
		}
		return nil
	})
	if len(backupFiles) == 0 {
		t.Error("expected a backup file in backupDir after overwriting changed content")
	}
}

// ── §5.7 TRIANGULATE: focused-agents-only + both happy paths ─────────────────

// ── RelPathForVariant export ──────────────────────────────────────────────────

// TestRelPathForVariant_Claude asserts that the exported RelPathForVariant
// returns the correct relative path for the "claude" variant.
// RED: fails until RelPathForVariant is exported from installer.go.
func TestRelPathForVariant_Claude(t *testing.T) {
	got := command.RelPathForVariant("claude")
	want := filepath.Join("jr", "starter-add.md")
	if got != want {
		t.Errorf("RelPathForVariant(%q) = %q, want %q", "claude", got, want)
	}
}

// TestRelPathForVariant_OpenCode asserts that the exported RelPathForVariant
// returns the correct relative path for the "opencode" variant.
func TestRelPathForVariant_OpenCode(t *testing.T) {
	got := command.RelPathForVariant("opencode")
	want := "jr-starter-add.md"
	if got != want {
		t.Errorf("RelPathForVariant(%q) = %q, want %q", "opencode", got, want)
	}
}

// TestRelPathForVariant_UnknownVariant asserts that an unknown variant key
// returns "" (the caller should skip silently).
func TestRelPathForVariant_UnknownVariant(t *testing.T) {
	got := command.RelPathForVariant("gemini")
	if got != "" {
		t.Errorf("RelPathForVariant(%q) = %q, want empty string for unknown variant", "gemini", got)
	}
}

// ── §5.7 TRIANGULATE: focused-agents-only + both happy paths ─────────────────

// TestInstaller_FocusedAgentsOnly_NoOutputForUnknownAgent asserts that an
// adapter with an unknown variant key produces no result silently
// (the asset path doesn't exist, but since commandsDir is empty it's skipped).
func TestInstaller_FocusedAgentsOnly_NoOutputForUnknownAgent(t *testing.T) {
	backupDir := t.TempDir()
	// An agent with an unknown variantKey AND no commandsDir → skip.
	adapter := fakeAdapter{agent: model.AgentGemini, commandsDir: "", variantKey: "gemini"}
	commandsFS := makeCommandsFS()
	ins := command.NewInstaller(commandsFS)

	results, err := ins.Install([]command.AgentAdapter{adapter}, "", backupDir)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-focused agent, got %d", len(results))
	}
}

// TestInstaller_BothFocusedAgents_HappyPath asserts that when both Claude and
// OpenCode adapters are provided, each gets its own command file written
// under the correct relative path (not mixed up).
func TestInstaller_BothFocusedAgents_HappyPath(t *testing.T) {
	claudeDir := t.TempDir()
	openCodeDir := t.TempDir()
	backupDir := t.TempDir()
	adapters := []command.AgentAdapter{
		fakeAdapter{agent: model.AgentClaude, commandsDir: claudeDir, variantKey: "claude"},
		fakeAdapter{agent: model.AgentOpenCode, commandsDir: openCodeDir, variantKey: "opencode"},
	}
	commandsFS := makeCommandsFS()
	ins := command.NewInstaller(commandsFS)

	results, err := ins.Install(adapters, "", backupDir)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (one per focused agent), got %d", len(results))
	}

	// Claude: jr/starter-add.md
	claudeFile := filepath.Join(claudeDir, "jr", "starter-add.md")
	if _, err := os.Stat(claudeFile); err != nil {
		t.Errorf("Claude command file not found at %q: %v", claudeFile, err)
	}

	// OpenCode: jr-starter-add.md (flat)
	openCodeFile := filepath.Join(openCodeDir, "jr-starter-add.md")
	if _, err := os.Stat(openCodeFile); err != nil {
		t.Errorf("OpenCode command file not found at %q: %v", openCodeFile, err)
	}

	// Files must be different content (different frontmatter).
	cData, _ := os.ReadFile(claudeFile)
	oData, _ := os.ReadFile(openCodeFile)
	if string(cData) == string(oData) {
		t.Error("Claude and OpenCode command files should have different content (different frontmatter)")
	}
}
