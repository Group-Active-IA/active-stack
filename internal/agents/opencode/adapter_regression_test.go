package opencode_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestOpenCodeAdapter_ProjectTarget_MCPPath_Unchanged asserts that the OpenCode
// project target MCP path stays at <root>/.opencode/opencode.json (merge-into-settings).
// No code changes are expected here — this is a regression guard for C-28.
func TestOpenCodeAdapter_ProjectTarget_MCPPath_Unchanged(t *testing.T) {
	a := opencode.NewAdapter()
	root := "/project/myapp"

	paths := a.PathsFor(root, model.Project)

	t.Run("MCPConfigPath_AnyServer", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "opencode.json")
		got := paths.MCPConfigPath("context7")
		if got != want {
			t.Errorf("OpenCode project MCPConfigPath = %q, want %q", got, want)
		}
	})

	t.Run("MCPConfigPath_OtherServer", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "opencode.json")
		got := paths.MCPConfigPath("any-other-server")
		if got != want {
			t.Errorf("OpenCode project MCPConfigPath(other) = %q, want %q", got, want)
		}
	})

	t.Run("MCPStrategy_MergeIntoSettings", func(t *testing.T) {
		if paths.MCPStrategy != model.MCPStrategyMergeIntoSettings {
			t.Errorf("OpenCode project MCPStrategy = %v, want MCPStrategyMergeIntoSettings", paths.MCPStrategy)
		}
	})
}

// TestOpenCodeAdapter_ZeroRegression_C31_PathsForFields asserts that adding
// CommandsDir in C-31 does not alter any previously established PathsFor fields
// for either target (zero-regression guard, mirrors the Claude adapter test).
func TestOpenCodeAdapter_ZeroRegression_C31_PathsForFields(t *testing.T) {
	a := opencode.NewAdapter()
	home := "/home/regressuser"
	root := "/project/myapp"

	t.Run("Machine_InstructionsPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "AGENTS.md")
		if got := a.PathsFor(home, model.Machine).InstructionsPath; got != want {
			t.Errorf("Machine PathsFor InstructionsPath = %q, want %q", got, want)
		}
	})
	t.Run("Machine_SkillsDir", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "skills")
		if got := a.PathsFor(home, model.Machine).SkillsDir; got != want {
			t.Errorf("Machine PathsFor SkillsDir = %q, want %q", got, want)
		}
	})
	t.Run("Machine_SettingsPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "opencode.json")
		if got := a.PathsFor(home, model.Machine).SettingsPath; got != want {
			t.Errorf("Machine PathsFor SettingsPath = %q, want %q", got, want)
		}
	})
	t.Run("Machine_MCPPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "opencode.json")
		if got := a.PathsFor(home, model.Machine).MCPConfigPath("any"); got != want {
			t.Errorf("Machine PathsFor MCPConfigPath = %q, want %q", got, want)
		}
	})
	t.Run("Machine_MCPStrategy", func(t *testing.T) {
		if got := a.PathsFor(home, model.Machine).MCPStrategy; got != model.MCPStrategyMergeIntoSettings {
			t.Errorf("Machine PathsFor MCPStrategy = %v, want MergeIntoSettings", got)
		}
	})
	t.Run("Project_InstructionsPath", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "AGENTS.md")
		if got := a.PathsFor(root, model.Project).InstructionsPath; got != want {
			t.Errorf("Project PathsFor InstructionsPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_SkillsDir", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "skills")
		if got := a.PathsFor(root, model.Project).SkillsDir; got != want {
			t.Errorf("Project PathsFor SkillsDir = %q, want %q", got, want)
		}
	})
	t.Run("Project_SettingsPath", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "opencode.json")
		if got := a.PathsFor(root, model.Project).SettingsPath; got != want {
			t.Errorf("Project PathsFor SettingsPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_MCPPath", func(t *testing.T) {
		want := filepath.Join(root, ".opencode", "opencode.json")
		if got := a.PathsFor(root, model.Project).MCPConfigPath("any"); got != want {
			t.Errorf("Project PathsFor MCPConfigPath = %q, want %q", got, want)
		}
	})
	t.Run("Project_MCPStrategy_MergeIntoSettings", func(t *testing.T) {
		if got := a.PathsFor(root, model.Project).MCPStrategy; got != model.MCPStrategyMergeIntoSettings {
			t.Errorf("Project PathsFor MCPStrategy = %v, want MergeIntoSettings", got)
		}
	})
}

// TestOpenCodeAdapter_ZeroRegression_MachineMethodsUnchanged is the explicit
// zero-regression guard for C-27. It pins the exact return values of the
// 7 pre-existing adapter methods (those that existed before C-27) and asserts
// they are UNCHANGED by the project-target addition.
//
// Machine layout uses ~/.config/opencode/ (XDG). This MUST remain unchanged.
func TestOpenCodeAdapter_ZeroRegression_MachineMethodsUnchanged(t *testing.T) {
	a := opencode.NewAdapter()
	home := "/home/regressuser"

	t.Run("Agent", func(t *testing.T) {
		if got := a.Agent(); got != model.AgentOpenCode {
			t.Errorf("Agent() = %q, want %q", got, model.AgentOpenCode)
		}
	})

	t.Run("InstructionsPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "AGENTS.md")
		if got := a.InstructionsPath(home); got != want {
			t.Errorf("InstructionsPath(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("SkillsDir", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "skills")
		if got := a.SkillsDir(home); got != want {
			t.Errorf("SkillsDir(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("SettingsPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "opencode.json")
		if got := a.SettingsPath(home); got != want {
			t.Errorf("SettingsPath(%q) = %q, want %q", home, got, want)
		}
	})

	t.Run("MCPConfigPath", func(t *testing.T) {
		want := filepath.Join(home, ".config", "opencode", "opencode.json")
		if got := a.MCPConfigPath(home, "context7"); got != want {
			t.Errorf("MCPConfigPath(%q, %q) = %q, want %q", home, "context7", got, want)
		}
	})

	t.Run("MCPStrategy", func(t *testing.T) {
		if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
			t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
		}
	})

	t.Run("VariantKey", func(t *testing.T) {
		if got := a.VariantKey(); got != "opencode" {
			t.Errorf("VariantKey() = %q, want %q", got, "opencode")
		}
	})
}
