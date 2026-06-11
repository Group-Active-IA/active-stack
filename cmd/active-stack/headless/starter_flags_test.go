// Package headless — tests for C-29 ParseStarterAddFlags (Task 2.1 RED).
package headless_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestParseStarterAddFlags covers the flag parsing rules from design D1/D2:
//   - <starter-id> is the first positional arg (required)
//   - --project <path> (optional; default = resolved cwd)
//   - --dry-run (optional bool)
//   - --yes / -y (optional bool)
//   - --agent <csv> (optional; default = focal agents claude+opencode)
//
// RED: this test fails because ParseStarterAddFlags does not exist yet.
func TestParseStarterAddFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantID       string
		wantProject  string // empty = skip (default cwd is not predictable)
		wantDryRun   bool
		wantYes      bool
		wantAgents   []model.Agent // nil = skip (don't check)
	}{
		// ── Happy path: id + explicit --project ────────────────────────────────
		{
			name:        "id and project",
			args:        []string{"active-ia", "--project", "/tmp/myproj"},
			wantID:      "active-ia",
			wantProject: "/tmp/myproj",
		},
		// ── --dry-run ──────────────────────────────────────────────────────────
		{
			name:       "id + dry-run",
			args:       []string{"backend", "--project", "/tmp/p", "--dry-run"},
			wantID:     "backend",
			wantDryRun: true,
		},
		// ── --yes / -y ─────────────────────────────────────────────────────────
		{
			name:    "--yes flag",
			args:    []string{"ux-ui", "--project", "/tmp/p", "--yes"},
			wantID:  "ux-ui",
			wantYes: true,
		},
		{
			name:    "-y short flag",
			args:    []string{"ux-ui", "--project", "/tmp/p", "-y"},
			wantID:  "ux-ui",
			wantYes: true,
		},
		// ── --agent CSV ────────────────────────────────────────────────────────
		{
			name:       "explicit agents",
			args:       []string{"active-ia", "--project", "/tmp/p", "--agent", "claude,opencode"},
			wantID:     "active-ia",
			wantAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		},
		// ── Default --agent omitted → focal agents ─────────────────────────────
		{
			name:       "default agents when omitted",
			args:       []string{"active-ia", "--project", "/tmp/p"},
			wantID:     "active-ia",
			wantAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		},
		// ── --project omitted → default cwd (don't check value, just no error) ─
		{
			name:   "project defaults to cwd when omitted",
			args:   []string{"active-ia"},
			wantID: "active-ia",
		},
		// ── Missing id → error ─────────────────────────────────────────────────
		{
			name:    "missing starter id",
			args:    []string{},
			wantErr: true,
		},
		// ── id looks like a flag (starts with -) → error ───────────────────────
		{
			name:    "id starts with dash",
			args:    []string{"--not-an-id"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := headless.ParseStarterAddFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseStarterAddFlags(%v) expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStarterAddFlags(%v) unexpected error: %v", tt.args, err)
			}

			if result.StarterID != tt.wantID {
				t.Errorf("StarterID = %q, want %q", result.StarterID, tt.wantID)
			}
			if tt.wantProject != "" && result.ProjectPath != tt.wantProject {
				t.Errorf("ProjectPath = %q, want %q", result.ProjectPath, tt.wantProject)
			}
			if result.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", result.DryRun, tt.wantDryRun)
			}
			if result.Yes != tt.wantYes {
				t.Errorf("Yes = %v, want %v", result.Yes, tt.wantYes)
			}
			if tt.wantAgents != nil {
				if len(result.Agents) != len(tt.wantAgents) {
					t.Errorf("Agents = %v, want %v", result.Agents, tt.wantAgents)
				} else {
					for i, a := range tt.wantAgents {
						if result.Agents[i] != a {
							t.Errorf("Agents[%d] = %q, want %q", i, result.Agents[i], a)
						}
					}
				}
			}
		})
	}
}
