package claude_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestClaudeAdapter_ProjectTarget_Paths verifies that the Claude adapter resolves
// the correct non-MCP paths when the install target is Project.
//
// Claude project layout (D2):
//   <root>/.claude/skills
//   <root>/.claude/CLAUDE.md
//   <root>/.claude/settings.json
func TestClaudeAdapter_ProjectTarget_Paths(t *testing.T) {
	a := claude.NewAdapter()
	root := "/project/myapp"

	paths := a.PathsFor(root, model.Project)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "SkillsDir",
			got:  paths.SkillsDir,
			want: filepath.Join(root, ".claude", "skills"),
		},
		{
			name: "InstructionsPath",
			got:  paths.InstructionsPath,
			want: filepath.Join(root, ".claude", "CLAUDE.md"),
		},
		{
			name: "SettingsPath",
			got:  paths.SettingsPath,
			want: filepath.Join(root, ".claude", "settings.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestClaudeAdapter_ProjectTarget_MCPPath asserts that the Claude project-target
// MCP path resolves to <root>/.mcp.json (D2), server name ignored.
//
// Spec: PathsFor(<root>, Project).MCPConfigPath("any-server") == <root>/.mcp.json
// AND  the path must NOT contain a ".claude/mcp/" segment
// AND  the path is identical regardless of the server name passed
func TestClaudeAdapter_ProjectTarget_MCPPath(t *testing.T) {
	a := claude.NewAdapter()
	root := "/project/myapp"

	paths := a.PathsFor(root, model.Project)

	wantPath := filepath.Join(root, ".mcp.json")

	t.Run("server_context7", func(t *testing.T) {
		got := paths.MCPConfigPath("context7")
		if got != wantPath {
			t.Errorf("MCPConfigPath(context7) = %q, want %q", got, wantPath)
		}
	})

	t.Run("server_other", func(t *testing.T) {
		got := paths.MCPConfigPath("other-server")
		if got != wantPath {
			t.Errorf("MCPConfigPath(other-server) = %q, want %q", got, wantPath)
		}
	})

	t.Run("no_claude_mcp_subdir", func(t *testing.T) {
		got := paths.MCPConfigPath("context7")
		if strings.Contains(filepath.ToSlash(got), ".claude/mcp/") {
			t.Errorf("project MCP path must not contain .claude/mcp/, got %q", got)
		}
	})
}

// TestClaudeAdapter_ProjectTarget_MCPStrategy asserts that the project-target
// resolved strategy is MCPStrategySingleFileMerge (D2, spec: single-file merge
// into mcpServers).
func TestClaudeAdapter_ProjectTarget_MCPStrategy(t *testing.T) {
	a := claude.NewAdapter()
	root := "/project/myapp"

	paths := a.PathsFor(root, model.Project)

	if paths.MCPStrategy != model.MCPStrategySingleFileMerge {
		t.Errorf("project MCPStrategy = %v, want MCPStrategySingleFileMerge", paths.MCPStrategy)
	}
}

// TestClaudeAdapter_MachineTargetViaPathsFor verifies that PathsFor with
// Machine target returns exactly the same paths as the individual methods.
// This is the zero-regression guard from the project side.
func TestClaudeAdapter_MachineTargetViaPathsFor(t *testing.T) {
	a := claude.NewAdapter()
	home := testHome

	paths := a.PathsFor(home, model.Machine)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "SkillsDir",
			got:  paths.SkillsDir,
			want: a.SkillsDir(home),
		},
		{
			name: "InstructionsPath",
			got:  paths.InstructionsPath,
			want: a.InstructionsPath(home),
		},
		{
			name: "SettingsPath",
			got:  paths.SettingsPath,
			want: a.SettingsPath(home),
		},
		{
			name: "MCPConfigPath(context7)",
			got:  paths.MCPConfigPath("context7"),
			want: a.MCPConfigPath(home, "context7"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
