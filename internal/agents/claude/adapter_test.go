package claude_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

const testHome = "/home/testuser"

func TestClaudeAdapter_Agent(t *testing.T) {
	a := claude.NewAdapter()
	if got := a.Agent(); got != model.AgentClaude {
		t.Errorf("Agent() = %q, want %q", got, model.AgentClaude)
	}
}

func TestClaudeAdapter_VariantKey(t *testing.T) {
	a := claude.NewAdapter()
	if got := a.VariantKey(); got != "claude" {
		t.Errorf("VariantKey() = %q, want %q", got, "claude")
	}
}

func TestClaudeAdapter_MCPStrategy(t *testing.T) {
	// Post-fix: Claude Code reads user-scope MCP servers from ~/.claude.json under
	// a top-level "mcpServers" key. There is no ~/.claude/mcp/ auto-discovery.
	// Strategy is StrategyMergeIntoSettings (merge into the shared file).
	a := claude.NewAdapter()
	if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
		t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
	}
}

// TestClaudeAdapter_CommandsDir_Machine asserts that the Claude adapter resolves
// the command directory to <homeDir>/.claude/commands for the Machine target.
// RED: fails until CommandsDir(homeDir) is added to the Claude adapter.
func TestClaudeAdapter_CommandsDir_Machine(t *testing.T) {
	a := claude.NewAdapter()
	home := testHome
	want := filepath.Join(home, ".claude", "commands")
	if got := a.CommandsDir(home); got != want {
		t.Errorf("CommandsDir(%q) = %q, want %q", home, got, want)
	}
}

// TestClaudeAdapter_CommandsDir_ProjectTarget asserts that PathsFor(root, Project)
// populates CommandsDir as <root>/.claude/commands.
// RED: fails until CommandsDir is populated in PathsFor.
func TestClaudeAdapter_CommandsDir_ProjectTarget(t *testing.T) {
	a := claude.NewAdapter()
	root := "/project/myapp"
	want := filepath.Join(root, ".claude", "commands")
	paths := a.PathsFor(root, model.Project)
	if paths.CommandsDir != want {
		t.Errorf("PathsFor(root, Project).CommandsDir = %q, want %q", paths.CommandsDir, want)
	}
}

func TestClaudeAdapter_PathMethods(t *testing.T) {
	a := claude.NewAdapter()
	home := testHome

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "InstructionsPath",
			got:  a.InstructionsPath(home),
			want: filepath.Join(home, ".claude", "CLAUDE.md"),
		},
		{
			name: "SkillsDir",
			got:  a.SkillsDir(home),
			want: filepath.Join(home, ".claude", "skills"),
		},
		{
			name: "SettingsPath",
			got:  a.SettingsPath(home),
			want: filepath.Join(home, ".claude", "settings.json"),
		},
		{
			// Post-fix: MCPConfigPath returns ~/.claude.json regardless of server name.
			// The serverName arg is ignored since all servers share the single file.
			name: "MCPConfigPath",
			got:  a.MCPConfigPath(home, "context7"),
			want: filepath.Join(home, ".claude.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
