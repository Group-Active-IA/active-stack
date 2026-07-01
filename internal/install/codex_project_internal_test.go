package install

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
)

func TestProjectTargetAdapterUsesCodexProjectLayout(t *testing.T) {
	root := t.TempDir()
	adapter := projectTargetAdapter{AgentAdapter: codex.NewAdapter(), base: root}

	if got, want := adapter.InstructionsPath(root), filepath.Join(root, "AGENTS.md"); got != want {
		t.Errorf("InstructionsPath = %q, want %q", got, want)
	}
	if got, want := adapter.SettingsPath(root), filepath.Join(root, ".codex", "config.toml"); got != want {
		t.Errorf("SettingsPath = %q, want %q", got, want)
	}
	if got := adapter.MCPStrategy(); got != external.StrategyMergeIntoTOML {
		t.Errorf("MCPStrategy = %v, want TOML", got)
	}
}
