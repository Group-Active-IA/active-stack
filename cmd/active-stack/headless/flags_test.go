// Package headless implements the non-interactive install mode for active-stack.
// This file contains table-driven tests for the flag parser (task 4.0a).
package headless_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestParseInstallFlags covers all cases from D4: flag parsing correctness,
// implicit headless activation, validation errors, and TUI fallback.
func TestParseInstallFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantTUI     bool   // true means headless=false (launch TUI)
		wantErr     bool
		wantMode    model.InstallMode
		wantAgents  []model.Agent
		wantCustom  []string
		wantDryRun  bool
		wantYes     bool
		wantHomeDir string // empty = uses os.UserHomeDir() (we can't test that value)
	}{
		// ── No flags → TUI mode ────────────────────────────────────────────
		{
			name:    "no flags → TUI",
			args:    []string{},
			wantTUI: true,
		},
		// ── Explicit --headless ────────────────────────────────────────────
		{
			name:     "explicit headless with mode lite",
			args:     []string{"--headless", "--mode", "lite"},
			wantTUI:  false,
			wantMode: model.ModeLite,
		},
		{
			name:     "explicit headless with mode full",
			args:     []string{"--headless", "--mode", "full"},
			wantTUI:  false,
			wantMode: model.ModeFull,
		},
		// ── --mode implies headless ────────────────────────────────────────
		{
			name:     "mode flag implies headless",
			args:     []string{"--mode", "lite"},
			wantTUI:  false,
			wantMode: model.ModeLite,
		},
		{
			name:     "mode full implies headless",
			args:     []string{"--mode", "full"},
			wantTUI:  false,
			wantMode: model.ModeFull,
		},
		// ── --agent implies headless ───────────────────────────────────────
		{
			name:       "agent flag implies headless",
			args:       []string{"--agent", "claude"},
			wantTUI:    false,
			wantAgents: []model.Agent{model.AgentClaude},
		},
		{
			name:       "multiple agents csv",
			args:       []string{"--agent", "claude,opencode"},
			wantTUI:    false,
			wantAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		},
		// ── mode + agent together ──────────────────────────────────────────
		{
			name:       "mode + agent",
			args:       []string{"--mode", "lite", "--agent", "claude"},
			wantTUI:    false,
			wantMode:   model.ModeLite,
			wantAgents: []model.Agent{model.AgentClaude},
		},
		// ── --mode custom + --custom ───────────────────────────────────────
		{
			name:       "custom mode with custom harnesses",
			args:       []string{"--mode", "custom", "--custom", "engram,openspec"},
			wantTUI:    false,
			wantMode:   model.ModeCustom,
			wantCustom: []string{"engram", "openspec"},
		},
		// ── --dry-run ──────────────────────────────────────────────────────
		{
			name:       "dry-run with mode",
			args:       []string{"--mode", "lite", "--dry-run"},
			wantTUI:    false,
			wantMode:   model.ModeLite,
			wantDryRun: true,
		},
		// ── --yes / -y ─────────────────────────────────────────────────────
		{
			name:    "--yes flag",
			args:    []string{"--mode", "lite", "--yes"},
			wantTUI: false,
			wantMode: model.ModeLite,
			wantYes: true,
		},
		{
			name:    "-y short flag",
			args:    []string{"--mode", "lite", "-y"},
			wantTUI: false,
			wantMode: model.ModeLite,
			wantYes: true,
		},
		// ── --home override ────────────────────────────────────────────────
		{
			name:        "--home override",
			args:        []string{"--mode", "lite", "--home", "/tmp/sandbox"},
			wantTUI:     false,
			wantMode:    model.ModeLite,
			wantHomeDir: "/tmp/sandbox",
		},
		// ── Validation errors ──────────────────────────────────────────────
		{
			name:    "invalid mode value",
			args:    []string{"--mode", "mega"},
			wantErr: true,
		},
		{
			name:    "--custom without --mode custom",
			args:    []string{"--mode", "lite", "--custom", "engram"},
			wantErr: true,
		},
		{
			name:    "--custom without any --mode",
			args:    []string{"--custom", "engram"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := headless.ParseInstallFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseInstallFlags() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseInstallFlags() unexpected error: %v", err)
			}

			// TUI mode check.
			if result.TUI != tt.wantTUI {
				t.Errorf("TUI = %v, want %v", result.TUI, tt.wantTUI)
			}

			// If TUI mode, stop here — intent fields are irrelevant.
			if tt.wantTUI {
				return
			}

			// Intent.Mode
			if tt.wantMode != "" && result.Intent.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", result.Intent.Mode, tt.wantMode)
			}

			// Intent.Agents
			if len(tt.wantAgents) > 0 {
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

			// Intent.Custom
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

			// DryRun
			if result.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", result.DryRun, tt.wantDryRun)
			}

			// Yes
			if result.Yes != tt.wantYes {
				t.Errorf("Yes = %v, want %v", result.Yes, tt.wantYes)
			}

			// HomeDir (only test when explicitly expected)
			if tt.wantHomeDir != "" && result.HomeDir != tt.wantHomeDir {
				t.Errorf("HomeDir = %q, want %q", result.HomeDir, tt.wantHomeDir)
			}
		})
	}
}

// TestParseInstallFlagsIntent verifies that the produced install.Intent has the
// right shape when mapped from flags (not TUI).
func TestParseInstallFlagsIntent(t *testing.T) {
	result, err := headless.ParseInstallFlags([]string{
		"--mode", "custom",
		"--agent", "claude,opencode",
		"--custom", "engram,openspec",
		"--yes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := install.Intent{
		Mode:   model.ModeCustom,
		Agents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		Custom: []string{"engram", "openspec"},
	}

	if result.Intent.Mode != want.Mode {
		t.Errorf("Mode = %q, want %q", result.Intent.Mode, want.Mode)
	}
	if len(result.Intent.Agents) != len(want.Agents) {
		t.Errorf("Agents len = %d, want %d", len(result.Intent.Agents), len(want.Agents))
	}
	if len(result.Intent.Custom) != len(want.Custom) {
		t.Errorf("Custom len = %d, want %d", len(result.Intent.Custom), len(want.Custom))
	}
	if !result.Yes {
		t.Error("Yes must be true")
	}
}
