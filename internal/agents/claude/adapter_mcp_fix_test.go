package claude_test

// Tests for the Claude MCP machine-target fix: ~/.claude.json + MergeIntoSettings.
//
// These tests are intentionally RED before the fix is applied (Step 3, TDD RED phase).
// They assert:
//  1. MCPConfigPath returns ~/.claude.json (not ~/.claude/mcp/<server>.json)
//  2. MCPStrategy returns StrategyMergeIntoSettings (not StrategySeparateFile)
//  3. PathsFor(Machine) MCP path == ~/.claude.json + MCPStrategySingleFileMerge
//  4. PathsFor(Project) UNCHANGED (.mcp.json + SingleFileMerge)

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestClaudeAdapter_MCPConfigPath_ReturnsClaudeJSON asserts that MCPConfigPath
// returns ~/.claude.json (the single shared user-scope file that Claude Code
// actually reads) regardless of the serverName argument.
//
// Spec (from root-cause analysis):
//
//	MCPConfigPath(homeDir, "context7") == filepath.Join(homeDir, ".claude.json")
//	MCPConfigPath(homeDir, "engram")   == filepath.Join(homeDir, ".claude.json")
//	MCPConfigPath(homeDir, "")         == filepath.Join(homeDir, ".claude.json")
//
// RED: fails before fix because the current impl returns ~/.claude/mcp/<server>.json.
func TestClaudeAdapter_MCPConfigPath_ReturnsClaudeJSON(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/testuser"
	want := filepath.Join(home, ".claude.json")

	cases := []struct {
		server string
	}{
		{"context7"},
		{"engram"},
		{"any-server"},
		{""},
	}
	for _, c := range cases {
		got := a.MCPConfigPath(home, c.server)
		if got != want {
			t.Errorf("MCPConfigPath(%q, %q) = %q, want %q", home, c.server, got, want)
		}
	}
}

// TestClaudeAdapter_MCPStrategy_IsMergeIntoSettings asserts that MCPStrategy
// returns StrategyMergeIntoSettings so that the installMCP / registerStdioMCP
// flow merges into ~/.claude.json under the "mcpServers" key.
//
// RED: fails before fix because the current impl returns StrategySeparateFile.
func TestClaudeAdapter_MCPStrategy_IsMergeIntoSettings(t *testing.T) {
	a := claude.NewAdapter()
	if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
		t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
	}
}

// TestClaudeAdapter_PathsFor_Machine_MCPPath_IsClaudeJSON asserts that
// PathsFor(homeDir, Machine).MCPConfigPath("any-server") returns ~/.claude.json.
//
// RED: fails before fix because current PathsFor(Machine) returns ~/.claude/mcp/<s>.json.
func TestClaudeAdapter_PathsFor_Machine_MCPPath_IsClaudeJSON(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/testuser"
	want := filepath.Join(home, ".claude.json")

	paths := a.PathsFor(home, model.Machine)
	for _, server := range []string{"context7", "engram", "any-server"} {
		got := paths.MCPConfigPath(server)
		if got != want {
			t.Errorf("PathsFor(Machine).MCPConfigPath(%q) = %q, want %q", server, got, want)
		}
	}
}

// TestClaudeAdapter_PathsFor_Machine_MCPStrategy_IsSingleFileMerge asserts that
// PathsFor(homeDir, Machine).MCPStrategy == MCPStrategySingleFileMerge.
//
// RED: fails before fix because current PathsFor(Machine) returns MCPStrategySeparateFile.
func TestClaudeAdapter_PathsFor_Machine_MCPStrategy_IsSingleFileMerge(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/testuser"
	paths := a.PathsFor(home, model.Machine)
	if paths.MCPStrategy != model.MCPStrategySingleFileMerge {
		t.Errorf("PathsFor(Machine).MCPStrategy = %v, want MCPStrategySingleFileMerge", paths.MCPStrategy)
	}
}

// TestClaudeAdapter_PathsFor_Project_Unchanged asserts that the Project target
// is NOT affected by the machine-target MCP fix: the MCP path is still
// <root>/.mcp.json and the strategy is still MCPStrategySingleFileMerge.
//
// This is the regression guard for the project target.
func TestClaudeAdapter_PathsFor_Project_Unchanged(t *testing.T) {
	a := claude.NewAdapter()
	root := "/project/myapp"
	want := filepath.Join(root, ".mcp.json")

	paths := a.PathsFor(root, model.Project)

	t.Run("MCPPath", func(t *testing.T) {
		if got := paths.MCPConfigPath("any-server"); got != want {
			t.Errorf("PathsFor(Project).MCPConfigPath = %q, want %q", got, want)
		}
	})
	t.Run("MCPStrategy", func(t *testing.T) {
		if paths.MCPStrategy != model.MCPStrategySingleFileMerge {
			t.Errorf("PathsFor(Project).MCPStrategy = %v, want MCPStrategySingleFileMerge", paths.MCPStrategy)
		}
	})
}

// TestClaudeAdapter_PathsFor_Machine_NonMCP_Unchanged asserts that the machine
// non-MCP paths (InstructionsPath, SkillsDir, SettingsPath, CommandsDir) are
// NOT altered by the MCP machine-target fix.
//
// Regression guard for Step 1.
func TestClaudeAdapter_PathsFor_Machine_NonMCP_Unchanged(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/testuser"
	paths := a.PathsFor(home, model.Machine)
	claudeDir := filepath.Join(home, ".claude")

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"InstructionsPath", paths.InstructionsPath, filepath.Join(claudeDir, "CLAUDE.md")},
		{"SkillsDir", paths.SkillsDir, filepath.Join(claudeDir, "skills")},
		{"SettingsPath", paths.SettingsPath, filepath.Join(claudeDir, "settings.json")},
		{"CommandsDir", paths.CommandsDir, filepath.Join(claudeDir, "commands")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Errorf("PathsFor(Machine).%s = %q, want %q", c.name, c.got, c.want)
			}
		})
	}
}
