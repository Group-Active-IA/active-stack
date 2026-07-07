package model

import "testing"

// TestTierCapable verifies the single source of truth for which agents can
// meaningfully differentiate the three permission tiers. Both the TUI
// permission screen and the windows options command consume this function.
func TestTierCapable(t *testing.T) {
	tests := []struct {
		name   string
		agents []Agent
		want   bool
	}{
		{"claude only → true", []Agent{AgentClaude}, true},
		{"gemini only → false", []Agent{AgentGemini}, false},
		{"opencode only → true", []Agent{AgentOpenCode}, true},
		{"codex only → true", []Agent{AgentCodex}, true},
		{"mixed gemini+claude → true", []Agent{AgentGemini, AgentClaude}, true},
		{"empty → false", []Agent{}, false},
		{"nil → false", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TierCapable(tt.agents)
			if got != tt.want {
				t.Errorf("TierCapable(%v) = %v, want %v", tt.agents, got, tt.want)
			}
		})
	}
}
