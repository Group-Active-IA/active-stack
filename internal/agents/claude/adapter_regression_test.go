package claude_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestClaudeAdapter_ZeroRegression_MachinePathsForMCP asserts that PathsFor
// with Machine target returns the correct post-fix machine MCP path
// (~/.claude.json) and strategy (MCPStrategySingleFileMerge). This test was
// originally a guard for C-28 (project target must not affect machine); it now
// also serves as the authoritative guard for the machine-target MCP fix that
// corrects the dead ~/.claude/mcp/ path.
//
// Claude Code reads user-scope MCP servers from ~/.claude.json under a top-level
// "mcpServers" key. There is no ~/.claude/mcp/ auto-discovery.
func TestClaudeAdapter_ZeroRegression_MachinePathsForMCP(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/regressuser"

	paths := a.PathsFor(home, model.Machine)

	t.Run("MachineTargetMCPPath", func(t *testing.T) {
		want := filepath.Join(home, ".claude.json")
		got := paths.MCPConfigPath("context7")
		if got != want {
			t.Errorf("Machine PathsFor MCPConfigPath = %q, want %q", got, want)
		}
	})

	t.Run("MachineTargetStrategy_SingleFileMerge", func(t *testing.T) {
		if paths.MCPStrategy != model.MCPStrategySingleFileMerge {
			t.Errorf("Machine MCPStrategy = %v, want MCPStrategySingleFileMerge", paths.MCPStrategy)
		}
	})

	t.Run("MachineLegacyMCPStrategy_IsMergeIntoSettings", func(t *testing.T) {
		if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
			t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
		}
	})
}

// TestClaudeAdapter_ZeroRegression_C31_PathsForFields asserts that adding
// CommandsDir in C-31 does not alter any previously established PathsFor fields
// (InstructionsPath, SkillsDir, SettingsPath, MCP path/strategy) for either target.
// This is the zero-regression guard for the C-31 PathsFor extension.
func TestClaudeAdapter_ZeroRegression_C31_PathsForFields(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/regressuser"
	root := "/project/myapp"

	t.Run("Machine_InstructionsPath", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "CLAUDE.md")
		if got := a.PathsFor(home, model.Machine).InstructionsPath; got != want {
			t.Errorf("Machine PathsFor InstructionsPath = %q, want %q", got, want)
		}
	})
	t.Run("Machine_SkillsDir", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "skills")
		if got := a.PathsFor(home, model.Machine).SkillsDir; got != want {
			t.Errorf("Machine PathsFor SkillsDir = %q, want %q", got, want)
		}
	})
	t.Run("Machine_SettingsPath", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "settings.json")
		if got := a.PathsFor(home, model.Machine).SettingsPath; got != want {
			t.Errorf("Machine PathsFor SettingsPath = %q, want %q", got, want)
		}
	})
	t.Run("Machine_MCPStrategy", func(t *testing.T) {
		if got := a.PathsFor(home, model.Machine).MCPStrategy; got != model.MCPStrategySingleFileMerge {
			t.Errorf("Machine PathsFor MCPStrategy = %v, want SingleFileMerge", got)
		}
	})
	t.Run("Project_InstructionsPath", func(t *testing.T) {
		want := filepath.Join(root, ".claude", "CLAUDE.md")
		if got := a.PathsFor(root, model.Project).InstructionsPath; got != want {
			t.Errorf("Project PathsFor InstructionsPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_SkillsDir", func(t *testing.T) {
		want := filepath.Join(root, ".claude", "skills")
		if got := a.PathsFor(root, model.Project).SkillsDir; got != want {
			t.Errorf("Project PathsFor SkillsDir = %q, want %q", got, want)
		}
	})
	t.Run("Project_SettingsPath", func(t *testing.T) {
		want := filepath.Join(root, ".claude", "settings.json")
		if got := a.PathsFor(root, model.Project).SettingsPath; got != want {
			t.Errorf("Project PathsFor SettingsPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_MCPPath_SingleFile", func(t *testing.T) {
		want := filepath.Join(root, ".mcp.json")
		if got := a.PathsFor(root, model.Project).MCPConfigPath("any-server"); got != want {
			t.Errorf("Project PathsFor MCPConfigPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_MCPStrategy_SingleFileMerge", func(t *testing.T) {
		if got := a.PathsFor(root, model.Project).MCPStrategy; got != model.MCPStrategySingleFileMerge {
			t.Errorf("Project PathsFor MCPStrategy = %v, want SingleFileMerge", got)
		}
	})
}

// TestClaudeAdapter_ZeroRegression_MachineMethodsUnchanged is the explicit
// zero-regression guard for C-27. It pins the exact return values of the
// 7 pre-existing adapter methods (those that existed before C-27) and asserts
// they are UNCHANGED by the project-target addition.
//
// Invariant: adding PathsFor must not alter the output of any existing method.
func TestClaudeAdapter_ZeroRegression_MachineMethodsUnchanged(t *testing.T) {
	a := claude.NewAdapter()
	home := "/home/regressuser"

	t.Run("Agent", func(t *testing.T) {
		if got := a.Agent(); got != model.AgentClaude {
			t.Errorf("Agent() = %q, want %q", got, model.AgentClaude)
		}
	})

	t.Run("InstructionsPath", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "CLAUDE.md")
		if got := a.InstructionsPath(home); got != want {
			t.Errorf("InstructionsPath(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("SkillsDir", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "skills")
		if got := a.SkillsDir(home); got != want {
			t.Errorf("SkillsDir(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("SettingsPath", func(t *testing.T) {
		want := filepath.Join(home, ".claude", "settings.json")
		if got := a.SettingsPath(home); got != want {
			t.Errorf("SettingsPath(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("MCPConfigPath", func(t *testing.T) {
		// Post-fix: MCPConfigPath returns ~/.claude.json regardless of server name.
		// Claude Code reads user-scope MCP servers from ~/.claude.json, not from
		// the old ~/.claude/mcp/<server>.json path which it never auto-discovers.
		want := filepath.Join(home, ".claude.json")
		if got := a.MCPConfigPath(home, "context7"); got != want {
			t.Errorf("MCPConfigPath(%q, %q) = %q, want %q", home, "context7", got, want)
		}
	})

	t.Run("MCPStrategy", func(t *testing.T) {
		// Post-fix: MCPStrategy returns StrategyMergeIntoSettings so that the
		// install flow merges into ~/.claude.json under the "mcpServers" key.
		if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
			t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
		}
	})

	t.Run("VariantKey", func(t *testing.T) {
		if got := a.VariantKey(); got != "claude" {
			t.Errorf("VariantKey() = %q, want %q", got, "claude")
		}
	})
}
