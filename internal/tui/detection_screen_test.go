package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

func TestViewDetection_RendersSystemInfo(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			System: system.SystemInfo{
				OS:        "linux",
				Arch:      "amd64",
				Shell:     "/bin/bash",
				Supported: true,
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	for _, want := range []string{"linux", "amd64", "bash"} {
		if !strings.Contains(out, want) {
			t.Errorf("viewDetection() output missing %q\ngot: %s", want, out)
		}
	}
}

func TestViewDetection_RendersToolsSection(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Tools: map[string]system.ToolStatus{
				"git":  {Name: "git", Installed: true},
				"brew": {Name: "brew", Installed: false},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if !strings.Contains(out, "git") || !strings.Contains(out, "found") {
		t.Errorf("viewDetection() should show git as found\ngot: %s", out)
	}
	if !strings.Contains(out, "brew") || !strings.Contains(out, "not found") {
		t.Errorf("viewDetection() should show brew as not found\ngot: %s", out)
	}
}

func TestViewDetection_RendersDependenciesSection(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Dependencies: system.DependencyReport{
				Dependencies: []system.Dependency{
					{Name: "node", Required: true, Installed: true, Version: "20.10.0"},
					{Name: "git", Required: true, Installed: false},
					{Name: "go", Required: false, Installed: false},
				},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if !strings.Contains(out, "node") || !strings.Contains(out, "20.10.0") {
		t.Errorf("viewDetection() should show node's version\ngot: %s", out)
	}
	if !strings.Contains(out, "git") || !strings.Contains(out, "NOT FOUND (required)") {
		t.Errorf("viewDetection() should show git as missing and required\ngot: %s", out)
	}
	if !strings.Contains(out, "go") || !strings.Contains(out, "optional") {
		t.Errorf("viewDetection() should show go as optional\ngot: %s", out)
	}
}

// TestViewDetection_AllDependenciesPresent triangulates against
// TestViewDetection_RendersDependenciesSection with a different data shape
// (nothing missing) to guard against a hardcoded "not found"/"required".
func TestViewDetection_AllDependenciesPresent(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Dependencies: system.DependencyReport{
				Dependencies: []system.Dependency{
					{Name: "curl", Required: true, Installed: true, Version: "8.4.0"},
					{Name: "npm", Required: true, Installed: true, Version: "10.2.3"},
				},
				AllPresent: true,
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if strings.Contains(out, "not found") || strings.Contains(out, "NOT FOUND") {
		t.Errorf("viewDetection() should not report any dependency as missing\ngot: %s", out)
	}
	if !strings.Contains(out, "8.4.0") || !strings.Contains(out, "10.2.3") {
		t.Errorf("viewDetection() should show both detected versions\ngot: %s", out)
	}
}

func TestViewDetection_RendersConfigsSection(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Configs: []system.ConfigState{
				{Agent: "claude-code", Exists: true},
				{Agent: "codex", Exists: false},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if !strings.Contains(out, "claude-code") || !strings.Contains(out, "present") {
		t.Errorf("viewDetection() should show claude-code config as present\ngot: %s", out)
	}
	if !strings.Contains(out, "codex") || !strings.Contains(out, "missing") {
		t.Errorf("viewDetection() should show codex config as missing\ngot: %s", out)
	}
}

func TestViewDetection_MissingDependencyShowsInstallHint(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Dependencies: system.DependencyReport{
				Dependencies: []system.Dependency{
					{Name: "git", Required: true, Installed: false, InstallHint: "install git from https://git-scm.com/"},
				},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if !strings.Contains(out, "install git from https://git-scm.com/") {
		t.Errorf("viewDetection() should show git's install hint\ngot: %s", out)
	}
}

func TestViewDetection_MissingDependencyShowsCopyableCommand(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			System: system.SystemInfo{
				Profile: system.PlatformProfile{OS: "windows"},
			},
			Dependencies: system.DependencyReport{
				Dependencies: []system.Dependency{
					{Name: "git", Required: true, Installed: false},
				},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	want := "winget install --id Git.Git -e --accept-source-agreements --accept-package-agreements"
	if !strings.Contains(out, want) {
		t.Errorf("viewDetection() should show the copyable install command\nwant substring: %s\ngot: %s", want, out)
	}
}

func TestViewDetection_NoCommandFallsBackToHintOnly(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			System: system.SystemInfo{
				Profile: system.PlatformProfile{OS: "windows"},
			},
			Dependencies: system.DependencyReport{
				Dependencies: []system.Dependency{
					{Name: "curl", Required: true, Installed: false, InstallHint: "curl is pre-installed on Windows 10+"},
				},
			},
		},
	})
	m.Screen = ScreenDetection

	out := m.View()

	if !strings.Contains(out, "curl is pre-installed on Windows 10+") {
		t.Errorf("viewDetection() should still show curl's install hint\ngot: %s", out)
	}
	if strings.Contains(out, "Run:") {
		t.Errorf("viewDetection() should not render an empty command block when InstallCommandsForDep is nil\ngot: %s", out)
	}
}

func TestDetectionAdvancesOnEnter_EvenWithMissingRequiredDeps(t *testing.T) {
	m := newModel(ModelDeps{
		Detection: system.DetectionResult{
			Dependencies: system.DependencyReport{
				Dependencies:    []system.Dependency{{Name: "node", Required: true, Installed: false}},
				AllPresent:      false,
				MissingRequired: []string{"node"},
			},
		},
	})
	m.Screen = ScreenDetection

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenAgents {
		t.Errorf("Screen = %v, want %v (Enter must advance even with MissingRequired set)", state.Screen, ScreenAgents)
	}
}
