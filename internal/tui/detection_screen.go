package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/system"
)

// viewDetection renders ScreenDetection: the full system/tools/dependencies/
// configs report from m.deps.Detection, shown before the user picks an
// agent. It never blocks — Enter always advances to ScreenAgents regardless
// of missing dependencies (the reactive gate in gate.go is the hard stop).
func (m Model) viewDetection() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Detecting environment"))
	sb.WriteString("\n\n")

	writeSystemInfoSection(&sb, m.deps.Detection.System)
	writeToolsSection(&sb, m.deps.Detection.Tools)
	writeDependenciesSection(&sb, m.deps.Detection.Dependencies, m.deps.Detection.System.Profile)
	writeConfigsSection(&sb, m.deps.Detection.Configs)

	sb.WriteString("\nPress Enter to continue.\n")
	return sb.String()
}

func writeConfigsSection(sb *strings.Builder, configs []system.ConfigState) {
	if len(configs) == 0 {
		return
	}

	sb.WriteString("\nDetected Configs:\n")
	for _, cfg := range configs {
		indicator := errorStyle.Render("missing")
		if cfg.Exists {
			indicator = selectedStyle.Render("present")
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", cfg.Agent, indicator))
	}
}

func writeDependenciesSection(sb *strings.Builder, report system.DependencyReport, profile system.PlatformProfile) {
	if len(report.Dependencies) == 0 {
		return
	}

	sb.WriteString("\nDependencies:\n")
	for _, dep := range report.Dependencies {
		var indicator string
		if dep.Installed {
			version := dep.Version
			if version == "" {
				version = "found"
			}
			indicator = selectedStyle.Render(version)
		} else {
			label := "not found"
			if dep.Required {
				label = "NOT FOUND (required)"
			}
			indicator = errorStyle.Render(label)
		}

		suffix := ""
		if !dep.Required {
			suffix = dimStyle.Render(" (optional)")
		}

		sb.WriteString(fmt.Sprintf("  %s: %s%s\n", dep.Name, indicator, suffix))

		if !dep.Installed && dep.InstallHint != "" {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("    %s\n", dep.InstallHint)))
		}

		if !dep.Installed {
			for _, cmd := range system.InstallCommandsForDep(dep.Name, profile) {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("    Run: %s\n", strings.Join(cmd, " "))))
			}
		}
	}
}

func writeToolsSection(sb *strings.Builder, tools map[string]system.ToolStatus) {
	if len(tools) == 0 {
		return
	}

	sb.WriteString("\nTools:\n")
	keys := make([]string, 0, len(tools))
	for k := range tools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		status := tools[k]
		indicator := errorStyle.Render("not found")
		if status.Installed {
			indicator = selectedStyle.Render("found")
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, indicator))
	}
}

func writeSystemInfoSection(sb *strings.Builder, sys system.SystemInfo) {
	supported := errorStyle.Render("No")
	if sys.Supported {
		supported = selectedStyle.Render("Yes")
	}
	shellName := filepath.Base(sys.Shell)

	sb.WriteString(fmt.Sprintf("OS: %s (%s)\n", sys.OS, sys.Arch))
	sb.WriteString(fmt.Sprintf("Shell: %s\n", shellName))
	sb.WriteString(fmt.Sprintf("Supported: %s\n", supported))
}
