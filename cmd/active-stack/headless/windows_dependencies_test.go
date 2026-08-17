// Package headless (internal test) — tests for toWindowsDependenciesResponse,
// unexported because it is a pure translation seam (gui-installer-dependency-detection).
package headless

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/system"
)

// TestToWindowsDependenciesResponse_InstallCommand asserts that a missing
// dependency's install command is joined into a single display string.
func TestToWindowsDependenciesResponse_InstallCommand(t *testing.T) {
	report := system.DependencyReport{
		Dependencies: []system.Dependency{
			{Name: "git", Required: true, Installed: false},
		},
	}
	profile := system.PlatformProfile{OS: "windows"}

	resp := toWindowsDependenciesResponse(report, profile)

	if len(resp.Dependencies) != 1 {
		t.Fatalf("Dependencies len = %d, want 1", len(resp.Dependencies))
	}
	want := "winget install --id Git.Git -e --accept-source-agreements --accept-package-agreements"
	if resp.Dependencies[0].InstallCommand != want {
		t.Errorf("InstallCommand = %q, want %q", resp.Dependencies[0].InstallCommand, want)
	}
}

// TestToWindowsDependenciesResponse_NoCommandFallsBackToHint triangulates
// with a dependency/platform combo where InstallCommandsForDep returns nil
// (curl on Windows) — install_command must be empty while install_hint stays
// populated.
func TestToWindowsDependenciesResponse_NoCommandFallsBackToHint(t *testing.T) {
	report := system.DependencyReport{
		Dependencies: []system.Dependency{
			{Name: "curl", Required: true, Installed: false, InstallHint: "curl is pre-installed on Windows 10+"},
		},
	}
	profile := system.PlatformProfile{OS: "windows"}

	resp := toWindowsDependenciesResponse(report, profile)

	if len(resp.Dependencies) != 1 {
		t.Fatalf("Dependencies len = %d, want 1", len(resp.Dependencies))
	}
	if resp.Dependencies[0].InstallCommand != "" {
		t.Errorf("InstallCommand = %q, want empty (no automatic command for curl on windows)", resp.Dependencies[0].InstallCommand)
	}
	if resp.Dependencies[0].InstallHint != "curl is pre-installed on Windows 10+" {
		t.Errorf("InstallHint = %q, want the hint text preserved", resp.Dependencies[0].InstallHint)
	}
}

// TestToWindowsDependenciesResponse_InstalledDependencyHasNoCommand asserts
// that an already-installed dependency never gets an install_command, even
// if InstallCommandsForDep would return one for its name/platform.
func TestToWindowsDependenciesResponse_InstalledDependencyHasNoCommand(t *testing.T) {
	report := system.DependencyReport{
		Dependencies: []system.Dependency{
			{Name: "git", Required: true, Installed: true, Version: "2.43.0"},
		},
	}
	profile := system.PlatformProfile{OS: "windows"}

	resp := toWindowsDependenciesResponse(report, profile)

	if resp.Dependencies[0].InstallCommand != "" {
		t.Errorf("InstallCommand = %q, want empty for an installed dependency", resp.Dependencies[0].InstallCommand)
	}
}
