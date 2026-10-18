package antigravity

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestAdapterMachinePaths(t *testing.T) {
	home := t.TempDir()
	a := NewAdapter()

	if got, want := a.InstructionsPath(home), filepath.Join(home, ".gemini", "config", "AGENTS.md"); got != want {
		t.Fatalf("InstructionsPath() = %q, want %q", got, want)
	}
	if got, want := a.SkillsDir(home), filepath.Join(home, ".gemini", "config", "skills"); got != want {
		t.Fatalf("SkillsDir() = %q, want %q", got, want)
	}
	if got, want := a.SettingsPath(home), filepath.Join(home, ".gemini", "config", "config.json"); got != want {
		t.Fatalf("SettingsPath() = %q, want %q", got, want)
	}
	if got, want := a.MCPConfigPath(home, "engram"), filepath.Join(home, ".gemini", "config", "mcp_config.json"); got != want {
		t.Fatalf("MCPConfigPath() = %q, want %q", got, want)
	}
	if got := a.CommandsDir(home); got != "" {
		t.Fatalf("CommandsDir() = %q, want empty", got)
	}
	if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
		t.Fatalf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
	}

	paths := a.PathsFor(home, model.Machine)
	if got, want := paths.InstructionsPath, a.InstructionsPath(home); got != want {
		t.Errorf("PathsFor(Machine).InstructionsPath = %q, want %q", got, want)
	}
	if got, want := paths.SkillsDir, a.SkillsDir(home); got != want {
		t.Errorf("PathsFor(Machine).SkillsDir = %q, want %q", got, want)
	}
	if got, want := paths.SettingsPath, a.SettingsPath(home); got != want {
		t.Errorf("PathsFor(Machine).SettingsPath = %q, want %q", got, want)
	}
	if got, want := paths.MCPConfigPath("context7"), a.MCPConfigPath(home, "context7"); got != want {
		t.Errorf("PathsFor(Machine).MCPConfigPath = %q, want %q", got, want)
	}
	if got := paths.MCPStrategy; got != model.MCPStrategyMergeIntoSettings {
		t.Errorf("PathsFor(Machine).MCPStrategy = %v, want MCPStrategyMergeIntoSettings", got)
	}
}

func TestAdapterProjectPaths(t *testing.T) {
	root := t.TempDir()
	paths := NewAdapter().PathsFor(root, model.Project)

	if got, want := paths.InstructionsPath, filepath.Join(root, "AGENTS.md"); got != want {
		t.Errorf("InstructionsPath = %q, want %q", got, want)
	}
	if got, want := paths.SkillsDir, filepath.Join(root, ".agents", "skills"); got != want {
		t.Errorf("SkillsDir = %q, want %q", got, want)
	}
	if got, want := paths.SettingsPath, filepath.Join(root, ".agents", "config.json"); got != want {
		t.Errorf("SettingsPath = %q, want %q", got, want)
	}
	if paths.CommandsDir != "" {
		t.Errorf("CommandsDir = %q, want empty", paths.CommandsDir)
	}
	if got := paths.MCPStrategy; got != model.MCPStrategyMergeIntoSettings {
		t.Errorf("MCPStrategy = %v, want MCPStrategyMergeIntoSettings", got)
	}
	if got, want := paths.MCPConfigPath("context7"), filepath.Join(root, ".agents", "mcp_config.json"); got != want {
		t.Errorf("MCPConfigPath = %q, want %q", got, want)
	}
}

func TestAdapterIdentityAndDelivery(t *testing.T) {
	a := NewAdapter()
	if a.Agent() != model.AgentAntigravity {
		t.Errorf("Agent() = %q, want %q", a.Agent(), model.AgentAntigravity)
	}
	if a.VariantKey() != "antigravity" {
		t.Errorf("VariantKey() = %q, want antigravity", a.VariantKey())
	}
	if a.ConfigDelivery() != model.ConfigDeliveryInstructions {
		t.Errorf("ConfigDelivery() = %v, want ConfigDeliveryInstructions", a.ConfigDelivery())
	}
}
