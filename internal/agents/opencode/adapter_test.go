package opencode_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

const testHome = "/home/testuser"

func TestOpenCodeAdapter_Agent(t *testing.T) {
	a := opencode.NewAdapter()
	if got := a.Agent(); got != model.AgentOpenCode {
		t.Errorf("Agent() = %q, want %q", got, model.AgentOpenCode)
	}
}

func TestOpenCodeAdapter_VariantKey(t *testing.T) {
	a := opencode.NewAdapter()
	if got := a.VariantKey(); got != "opencode" {
		t.Errorf("VariantKey() = %q, want %q", got, "opencode")
	}
}

func TestOpenCodeAdapter_MCPStrategy(t *testing.T) {
	a := opencode.NewAdapter()
	if got := a.MCPStrategy(); got != external.StrategyMergeIntoSettings {
		t.Errorf("MCPStrategy() = %v, want StrategyMergeIntoSettings", got)
	}
}

// TestOpenCodeAdapter_CommandsDir_Machine asserts that the OpenCode adapter
// resolves the machine-target command directory to
// <homeDir>/.config/opencode/commands (NOT .opencode/).
// RED: fails until CommandsDir(homeDir) is added to the OpenCode adapter.
func TestOpenCodeAdapter_CommandsDir_Machine(t *testing.T) {
	a := opencode.NewAdapter()
	home := testHome
	want := filepath.Join(home, ".config", "opencode", "commands")
	if got := a.CommandsDir(home); got != want {
		t.Errorf("CommandsDir(%q) = %q, want %q", home, got, want)
	}
}

// TestOpenCodeAdapter_CommandsDir_DiffersbyTarget asserts that:
//   - Project target → <root>/.opencode/commands
//   - Machine target → <home>/.config/opencode/commands  (different!)
//
// RED: fails until CommandsDir is populated in PathsFor.
func TestOpenCodeAdapter_CommandsDir_DiffersbyTarget(t *testing.T) {
	a := opencode.NewAdapter()
	root := "/project/myapp"
	home := testHome

	projectPaths := a.PathsFor(root, model.Project)
	machinePaths := a.PathsFor(home, model.Machine)

	wantProject := filepath.Join(root, ".opencode", "commands")
	wantMachine := filepath.Join(home, ".config", "opencode", "commands")

	if projectPaths.CommandsDir != wantProject {
		t.Errorf("PathsFor(root, Project).CommandsDir = %q, want %q", projectPaths.CommandsDir, wantProject)
	}
	if machinePaths.CommandsDir != wantMachine {
		t.Errorf("PathsFor(home, Machine).CommandsDir = %q, want %q", machinePaths.CommandsDir, wantMachine)
	}
	// Assert they actually differ (per spec).
	if projectPaths.CommandsDir == machinePaths.CommandsDir {
		t.Errorf("Project and Machine CommandsDir should differ but both = %q", projectPaths.CommandsDir)
	}
}

func TestOpenCodeAdapter_PathMethods(t *testing.T) {
	a := opencode.NewAdapter()
	home := testHome

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "InstructionsPath",
			got:  a.InstructionsPath(home),
			want: filepath.Join(home, ".config", "opencode", "AGENTS.md"),
		},
		{
			name: "SkillsDir",
			got:  a.SkillsDir(home),
			want: filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name: "SettingsPath",
			got:  a.SettingsPath(home),
			want: filepath.Join(home, ".config", "opencode", "opencode.json"),
		},
		{
			name: "MCPConfigPath",
			got:  a.MCPConfigPath(home, "context7"),
			want: filepath.Join(home, ".config", "opencode", "opencode.json"),
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
