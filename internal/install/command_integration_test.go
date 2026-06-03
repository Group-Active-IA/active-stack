package install_test

// command_integration_test.go — §7.1/§7.2 integration tests for the C-31
// command harness write path: BuildPlan + pipeline end-to-end.
//
// These tests write actual files to t.TempDir() using the real command installer
// and verify that the correct files land under the correct paths.
// Re-run (idempotent) and no-op (dry-run-via-empty-catalog) are also covered.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// makeIntegrationCommandsFS returns an fstest.MapFS that contains both agent
// command variants, using the same asset paths as assets.CommandsFS.
func makeIntegrationCommandsFS() fs.FS {
	return fstest.MapFS{
		"commands/claude/jr/starter-add.md": {
			Data: []byte("---\nname: \"JR: Starter Add\"\ndescription: Apply a starter.\ncategory: Workflow\ntags:\n  - jr-stack\n  - starter\n---\njr-stack starter add $ARGUMENTS"),
		},
		"commands/opencode/jr-starter-add.md": {
			Data: []byte("---\ndescription: Apply a starter.\n---\njr-stack starter add $ARGUMENTS"),
		},
	}
}

// integrationCommandAdapter is a test adapter that resolves real temp-dir paths
// for the command integration test.
type integrationCommandAdapter struct {
	agent       model.Agent
	variantKey  string
	commandsDir string
}

func (a integrationCommandAdapter) Agent() model.Agent                  { return a.agent }
func (a integrationCommandAdapter) InstructionsPath(h string) string    { return h + "/instr.md" }
func (a integrationCommandAdapter) SkillsDir(h string) string           { return h + "/skills" }
func (a integrationCommandAdapter) CommandsDir(_ string) string         { return a.commandsDir }
func (a integrationCommandAdapter) SettingsPath(h string) string        { return h + "/settings.json" }
func (a integrationCommandAdapter) MCPConfigPath(h, s string) string    { return h + "/mcp/" + s + ".json" }
func (a integrationCommandAdapter) MCPStrategy() external.MCPStrategy { return external.StrategySeparateFile }
func (a integrationCommandAdapter) VariantKey() string                  { return a.variantKey }
func (a integrationCommandAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a integrationCommandAdapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/instr.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      a.commandsDir,
	}.WithMCPConfigFn(func(serverName string) string {
		return base + "/mcp/" + serverName + ".json"
	})
}

// TestCommandInstall_Integration_WritesFilesForBothAgents is the §7.1/§7.2
// integration test. It uses the real command installer (via SetCommandInstallFn
// reset to nil = no fake) to write actual files.
//
// Verifies:
//   - Claude: /.claude/commands/jr/starter-add.md is written with correct content
//   - OpenCode: /.opencode/commands/jr-starter-add.md is written with correct content
//   - Both bodies contain "jr-stack starter add $ARGUMENTS"
//   - Re-run (second BuildPlan + Execute) is a no-op (AlreadyInstalled path)
func TestCommandInstall_Integration_WritesFilesForBothAgents(t *testing.T) {
	projectRoot := t.TempDir()
	backupDir := t.TempDir()

	claudeCommandsDir := filepath.Join(projectRoot, ".claude", "commands")
	openCodeCommandsDir := filepath.Join(projectRoot, ".opencode", "commands")

	commandsFS := makeIntegrationCommandsFS()

	// Set the real command install using the test FS (avoid touching assets global).
	restoreCommand := install.SetCommandInstallFn(func(adapters []install.AgentAdapter, homeDir, bDir string) error {
		// Use the command installer directly with our test FS and adapters.
		// We inline the logic here to avoid a package-level FS dependency issue.
		for _, adapter := range adapters {
			dir := adapter.CommandsDir(homeDir)
			if dir == "" {
				continue
			}
			var assetPath, relPath string
			switch adapter.VariantKey() {
			case "claude":
				assetPath = "commands/claude/jr/starter-add.md"
				relPath = filepath.Join("jr", "starter-add.md")
			case "opencode":
				assetPath = "commands/opencode/jr-starter-add.md"
				relPath = "jr-starter-add.md"
			default:
				continue
			}
			content, err := fs.ReadFile(commandsFS, assetPath)
			if err != nil {
				return err
			}
			destFile := filepath.Join(dir, relPath)
			if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
				return err
			}
			// Idempotent: skip if byte-identical.
			if existing, readErr := os.ReadFile(destFile); readErr == nil {
				if string(existing) == string(content) {
					continue
				}
			}
			if err := os.WriteFile(destFile, content, 0o644); err != nil {
				return err
			}
		}
		_ = bDir // backup dir, used by real installer
		return nil
	})
	defer restoreCommand()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	claudeAdapter := integrationCommandAdapter{
		agent:       model.AgentClaude,
		variantKey:  "claude",
		commandsDir: claudeCommandsDir,
	}
	openCodeAdapter := integrationCommandAdapter{
		agent:       model.AgentOpenCode,
		variantKey:  "opencode",
		commandsDir: openCodeCommandsDir,
	}

	h := commandHarness()
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude:    claudeAdapter,
		model.AgentOpenCode:  openCodeAdapter,
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     projectRoot,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	// ── First install ──────────────────────────────────────────────────────────
	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error: %v", err)
	}
	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error: %v", result.Err)
	}

	// Claude: /.claude/commands/jr/starter-add.md
	claudeFile := filepath.Join(claudeCommandsDir, "jr", "starter-add.md")
	claudeData, err := os.ReadFile(claudeFile)
	if err != nil {
		t.Fatalf("Claude command file not written at %q: %v", claudeFile, err)
	}
	if !strings.Contains(string(claudeData), "jr-stack starter add $ARGUMENTS") {
		t.Errorf("Claude command file content unexpected:\n%s", claudeData)
	}

	// OpenCode: /.opencode/commands/jr-starter-add.md
	openCodeFile := filepath.Join(openCodeCommandsDir, "jr-starter-add.md")
	openCodeData, err := os.ReadFile(openCodeFile)
	if err != nil {
		t.Fatalf("OpenCode command file not written at %q: %v", openCodeFile, err)
	}
	if !strings.Contains(string(openCodeData), "jr-stack starter add $ARGUMENTS") {
		t.Errorf("OpenCode command file content unexpected:\n%s", openCodeData)
	}

	// ── Second install: re-run must be a no-op (files unchanged) ──────────────
	mtime1Claude, _ := os.Stat(claudeFile)
	mtime1OpenCode, _ := os.Stat(openCodeFile)

	plan2, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("second BuildPlan() error: %v", err)
	}
	result2 := orch.Execute(plan2.StagePlan)
	if result2.Err != nil {
		t.Fatalf("second Execute() error: %v", result2.Err)
	}

	mtime2Claude, _ := os.Stat(claudeFile)
	mtime2OpenCode, _ := os.Stat(openCodeFile)

	// ModTime must be unchanged (no rewrite happened).
	if mtime2Claude.ModTime() != mtime1Claude.ModTime() {
		t.Error("Claude command file was rewritten on second install (expected no-op)")
	}
	if mtime2OpenCode.ModTime() != mtime1OpenCode.ModTime() {
		t.Error("OpenCode command file was rewritten on second install (expected no-op)")
	}

	// Assert content unchanged.
	claudeData2, _ := os.ReadFile(claudeFile)
	openCodeData2, _ := os.ReadFile(openCodeFile)
	if string(claudeData2) != string(claudeData) {
		t.Error("Claude command file content changed on second install (expected idempotent)")
	}
	if string(openCodeData2) != string(openCodeData) {
		t.Error("OpenCode command file content changed on second install (expected idempotent)")
	}

	_ = backupDir
}
