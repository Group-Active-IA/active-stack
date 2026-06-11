package opencode_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestOpenCodeAdapter_ProjectTarget verifies that the OpenCode adapter resolves
// the correct paths when the install target is Project.
//
// OpenCode project layout (D2, confirmed against official docs):
//   <root>/.opencode/skills          (NOT <root>/.config/opencode/skills)
//   <root>/.opencode/AGENTS.md
//   <root>/.opencode/opencode.json
//
// The machine layout uses ~/.config/opencode/ (XDG convention).
// The project layout uses <root>/.opencode/ (project-local convention).
// This difference MUST live inside the adapter — never in the caller.
func TestOpenCodeAdapter_ProjectTarget(t *testing.T) {
	a := opencode.NewAdapter()
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
			want: filepath.Join(root, ".opencode", "skills"),
		},
		{
			name: "InstructionsPath",
			got:  paths.InstructionsPath,
			want: filepath.Join(root, ".opencode", "AGENTS.md"),
		},
		{
			name: "SettingsPath",
			got:  paths.SettingsPath,
			want: filepath.Join(root, ".opencode", "opencode.json"),
		},
		{
			name: "MCPConfigPath(context7)",
			got:  paths.MCPConfigPath("context7"),
			// OpenCode merges MCP into the settings file (StrategyMergeIntoSettings).
			want: filepath.Join(root, ".opencode", "opencode.json"),
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

// TestOpenCodeAdapter_ProjectSkillsDirIsNotConfigOpenCode explicitly asserts that
// the project skills dir does NOT use .config/opencode (machine convention).
func TestOpenCodeAdapter_ProjectSkillsDirIsNotConfigOpenCode(t *testing.T) {
	a := opencode.NewAdapter()
	root := "/project/myapp"

	paths := a.PathsFor(root, model.Project)
	bad := filepath.Join(root, ".config", "opencode", "skills")

	if paths.SkillsDir == bad {
		t.Errorf("OpenCode project SkillsDir must NOT be %q (machine XDG path), got that path", bad)
	}
}

// TestOpenCodeAdapter_MachineTargetViaPathsFor verifies that PathsFor with
// Machine target returns exactly the same paths as the individual methods.
// This is the zero-regression guard from the project side.
func TestOpenCodeAdapter_MachineTargetViaPathsFor(t *testing.T) {
	a := opencode.NewAdapter()
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
