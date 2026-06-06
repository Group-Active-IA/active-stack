// Package headless implements the non-interactive (headless) install/uninstall mode.
// This file contains table-driven tests for the uninstall flag parser (task 1.1 RED).
package headless_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestParseUninstallFlags covers all behaviors from design D4 and spec requirements:
// default strategy, valid/invalid modes, --custom validation, --agent CSV, --strategy,
// --restore-manifest, --dry-run, --yes/-y, --home.
func TestParseUninstallFlags(t *testing.T) {
	tests := []struct {
		name                    string
		args                    []string
		wantErr                 bool
		wantMode                model.InstallMode
		wantAgents              []model.Agent
		wantCustom              []string
		wantStrategy            uninstall.Strategy
		wantRestoreManifestPath string
		wantDryRun              bool
		wantYes                 bool
		wantHomeDir             string // non-empty means we assert this exact value
	}{
		// ── Strategy defaults to targeted (D5) ────────────────────────────────
		{
			name:         "no flags → strategy defaults to targeted",
			args:         []string{},
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "mode full without strategy → strategy defaults to targeted",
			args:         []string{"--mode", "full"},
			wantMode:     model.ModeFull,
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── Valid --mode values ────────────────────────────────────────────────
		{
			name:         "mode lite",
			args:         []string{"--mode", "lite"},
			wantMode:     model.ModeLite,
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "mode full",
			args:         []string{"--mode", "full"},
			wantMode:     model.ModeFull,
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "mode custom with --custom",
			args:         []string{"--mode", "custom", "--custom", "engram,openspec"},
			wantMode:     model.ModeCustom,
			wantCustom:   []string{"engram", "openspec"},
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── Invalid --mode → error ─────────────────────────────────────────────
		{
			name:    "invalid mode value",
			args:    []string{"--mode", "bogus"},
			wantErr: true,
		},
		{
			name:    "invalid mode value mega",
			args:    []string{"--mode", "mega"},
			wantErr: true,
		},

		// ── --custom without --mode custom → error ─────────────────────────────
		{
			name:    "--custom without --mode custom (mode lite)",
			args:    []string{"--mode", "lite", "--custom", "engram"},
			wantErr: true,
		},
		{
			name:    "--custom without any --mode",
			args:    []string{"--custom", "engram"},
			wantErr: true,
		},

		// ── --agent CSV parsing ────────────────────────────────────────────────
		{
			name:         "single agent",
			args:         []string{"--agent", "claude"},
			wantAgents:   []model.Agent{model.AgentClaude},
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "multiple agents csv",
			args:         []string{"--agent", "claude,opencode"},
			wantAgents:   []model.Agent{model.AgentClaude, model.AgentOpenCode},
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── Empty agent list = all agents ──────────────────────────────────────
		{
			name:         "no --agent flag → empty agent list (all agents)",
			args:         []string{"--mode", "lite"},
			wantMode:     model.ModeLite,
			wantAgents:   nil, // empty means all agents
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── --strategy restore ─────────────────────────────────────────────────
		{
			name:                    "strategy restore with manifest",
			args:                    []string{"--strategy", "restore", "--restore-manifest", "/path/to/manifest.json"},
			wantStrategy:            uninstall.StrategyRestore,
			wantRestoreManifestPath: "/path/to/manifest.json",
		},
		{
			name:         "strategy targeted explicit",
			args:         []string{"--strategy", "targeted"},
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:    "invalid strategy value",
			args:    []string{"--strategy", "rollback"},
			wantErr: true,
		},

		// ── --restore-manifest path captured ──────────────────────────────────
		{
			name:                    "restore-manifest path is captured",
			args:                    []string{"--strategy", "restore", "--restore-manifest", "/some/manifest.json"},
			wantStrategy:            uninstall.StrategyRestore,
			wantRestoreManifestPath: "/some/manifest.json",
		},

		// ── --dry-run ──────────────────────────────────────────────────────────
		{
			name:         "--dry-run flag",
			args:         []string{"--mode", "lite", "--dry-run"},
			wantMode:     model.ModeLite,
			wantDryRun:   true,
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "--dry-run without mode",
			args:         []string{"--dry-run"},
			wantDryRun:   true,
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── --yes / -y ─────────────────────────────────────────────────────────
		{
			name:         "--yes flag",
			args:         []string{"--mode", "lite", "--yes"},
			wantMode:     model.ModeLite,
			wantYes:      true,
			wantStrategy: uninstall.StrategyTargeted,
		},
		{
			name:         "-y short flag",
			args:         []string{"--mode", "lite", "-y"},
			wantMode:     model.ModeLite,
			wantYes:      true,
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── --home ─────────────────────────────────────────────────────────────
		{
			name:         "--home override is honored",
			args:         []string{"--mode", "lite", "--home", "/custom/home"},
			wantMode:     model.ModeLite,
			wantHomeDir:  "/custom/home",
			wantStrategy: uninstall.StrategyTargeted,
		},

		// ── No --project flag (machine-scope) ──────────────────────────────────
		{
			// --project is not a recognized flag; flag.ContinueOnError should error.
			name:    "--project is not a recognized flag",
			args:    []string{"--project", "/repo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := headless.ParseUninstallFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseUninstallFlags(%v) expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUninstallFlags(%v) unexpected error: %v", tt.args, err)
			}

			// Strategy
			if result.Intent.Strategy != tt.wantStrategy {
				t.Errorf("Strategy = %q, want %q", result.Intent.Strategy, tt.wantStrategy)
			}

			// Mode
			if tt.wantMode != "" && result.Intent.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", result.Intent.Mode, tt.wantMode)
			}

			// Agents
			if tt.wantAgents != nil {
				if len(result.Intent.Agents) != len(tt.wantAgents) {
					t.Errorf("Agents = %v, want %v", result.Intent.Agents, tt.wantAgents)
				} else {
					for i, a := range tt.wantAgents {
						if result.Intent.Agents[i] != a {
							t.Errorf("Agents[%d] = %q, want %q", i, result.Intent.Agents[i], a)
						}
					}
				}
			}

			// Custom
			if len(tt.wantCustom) > 0 {
				if len(result.Intent.Custom) != len(tt.wantCustom) {
					t.Errorf("Custom = %v, want %v", result.Intent.Custom, tt.wantCustom)
				} else {
					for i, id := range tt.wantCustom {
						if result.Intent.Custom[i] != id {
							t.Errorf("Custom[%d] = %q, want %q", i, result.Intent.Custom[i], id)
						}
					}
				}
			}

			// RestoreManifestPath
			if tt.wantRestoreManifestPath != "" && result.RestoreManifestPath != tt.wantRestoreManifestPath {
				t.Errorf("RestoreManifestPath = %q, want %q", result.RestoreManifestPath, tt.wantRestoreManifestPath)
			}

			// DryRun
			if result.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", result.DryRun, tt.wantDryRun)
			}

			// Yes
			if result.Yes != tt.wantYes {
				t.Errorf("Yes = %v, want %v", result.Yes, tt.wantYes)
			}

			// HomeDir (only when explicitly expected)
			if tt.wantHomeDir != "" && result.HomeDir != tt.wantHomeDir {
				t.Errorf("HomeDir = %q, want %q", result.HomeDir, tt.wantHomeDir)
			}
		})
	}
}

// TestParseUninstallFlagsHomeDefault verifies that when --home is not provided,
// HomeDir is populated from os.UserHomeDir() (non-empty).
func TestParseUninstallFlagsHomeDefault(t *testing.T) {
	result, err := headless.ParseUninstallFlags([]string{"--mode", "lite"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HomeDir == "" {
		t.Error("HomeDir must default to os.UserHomeDir(), got empty string")
	}
}

// TestParseUninstallFlagsIntentShape verifies the full intent shape for a rich flag set.
func TestParseUninstallFlagsIntentShape(t *testing.T) {
	result, err := headless.ParseUninstallFlags([]string{
		"--mode", "custom",
		"--agent", "claude,opencode",
		"--custom", "engram,openspec",
		"--strategy", "targeted",
		"--yes",
		"--home", "/test/home",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Intent.Mode != model.ModeCustom {
		t.Errorf("Mode = %q, want custom", result.Intent.Mode)
	}
	if len(result.Intent.Agents) != 2 {
		t.Errorf("Agents len = %d, want 2", len(result.Intent.Agents))
	}
	if len(result.Intent.Custom) != 2 {
		t.Errorf("Custom len = %d, want 2", len(result.Intent.Custom))
	}
	if result.Intent.Strategy != uninstall.StrategyTargeted {
		t.Errorf("Strategy = %q, want targeted", result.Intent.Strategy)
	}
	if !result.Yes {
		t.Error("Yes must be true")
	}
	if result.HomeDir != "/test/home" {
		t.Errorf("HomeDir = %q, want /test/home", result.HomeDir)
	}
}
