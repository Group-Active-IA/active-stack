package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// buildRunUninstallCallback creates the RunUninstall callback for ModelDeps.
// It closes over the real catalog and registry, reducing the TUI's callback
// signature to (flags, writer) int.
func buildRunUninstallCallback(cat uninstall.Catalog, reg uninstall.Registry) func(headless.ParsedUninstallFlags, io.Writer) int {
	return func(flags headless.ParsedUninstallFlags, w io.Writer) int {
		return headless.RunHeadlessUninstall(flags, cat, reg, w)
	}
}

// buildRunStarterCallback creates the RunStarter callback for ModelDeps.
// It closes over the real catalog and registry, exposing only the high-level
// (starterID, projectPath, agents, writer) → int surface to the TUI.
func buildRunStarterCallback(cat starterCatalog, reg install.Registry) func(string, string, []model.Agent, io.Writer) int {
	return func(starterID, projectPath string, agents []model.Agent, w io.Writer) int {
		flags := headless.ParsedStarterAddFlags{
			StarterID:   starterID,
			ProjectPath: projectPath,
			DryRun:      false,
			Yes:         true,
			Agents:      agents,
		}
		return runStarterAdd(flags, cat, reg, nil, w)
	}
}

// resolveBackupDir returns the backup directory for the current user.
// It follows the same convention as the install engine: ~/.jr-stack/backups.
func resolveBackupDir(homeDir string) string {
	return filepath.Join(homeDir, ".jr-stack", "backups")
}

// defaultBackupDir resolves the backup directory for the current user.
// Falls back to os.UserHomeDir on error; if that also fails, returns ".".
func defaultBackupDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return resolveBackupDir(home)
}
