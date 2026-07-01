package codex

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestAdapterMachinePaths(t *testing.T) {
	home := t.TempDir()
	a := NewAdapter()

	if got, want := a.InstructionsPath(home), filepath.Join(home, ".codex", "AGENTS.md"); got != want {
		t.Fatalf("InstructionsPath() = %q, want %q", got, want)
	}
	if got, want := a.SkillsDir(home), filepath.Join(home, ".agents", "skills"); got != want {
		t.Fatalf("SkillsDir() = %q, want %q", got, want)
	}
	if got, want := a.SettingsPath(home), filepath.Join(home, ".codex", "config.toml"); got != want {
		t.Fatalf("SettingsPath() = %q, want %q", got, want)
	}
	if got := a.CommandsDir(home); got != "" {
		t.Fatalf("CommandsDir() = %q, want empty", got)
	}
	if got := a.MCPStrategy(); got != external.StrategyMergeIntoTOML {
		t.Fatalf("MCPStrategy() = %v, want TOML", got)
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
	if got, want := paths.SettingsPath, filepath.Join(root, ".codex", "config.toml"); got != want {
		t.Errorf("SettingsPath = %q, want %q", got, want)
	}
	if paths.CommandsDir != "" {
		t.Errorf("CommandsDir = %q, want empty", paths.CommandsDir)
	}
	if got := paths.MCPStrategy; got != model.MCPStrategyMergeIntoTOML {
		t.Errorf("MCPStrategy = %v, want TOML", got)
	}
	if got, want := paths.MCPConfigPath("context7"), filepath.Join(root, ".codex", "config.toml"); got != want {
		t.Errorf("MCPConfigPath = %q, want %q", got, want)
	}
}

func TestAdapterIdentityAndDelivery(t *testing.T) {
	a := NewAdapter()
	if a.Agent() != model.AgentCodex {
		t.Errorf("Agent() = %q", a.Agent())
	}
	if a.VariantKey() != "codex" {
		t.Errorf("VariantKey() = %q", a.VariantKey())
	}
	if a.ConfigDelivery() != model.ConfigDeliveryInstructions {
		t.Errorf("ConfigDelivery() = %v", a.ConfigDelivery())
	}
}
