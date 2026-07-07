package tui

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestBuildIntentLite verifies that Lite mode produces an Intent with no
// Custom field (the catalog's ForMode handles the harness set).
func TestBuildIntentLite(t *testing.T) {
	s := Selection{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	intent := s.BuildIntent()

	if intent.Mode != model.ModeLite {
		t.Errorf("Mode = %q, want %q", intent.Mode, model.ModeLite)
	}
	if len(intent.Agents) != 1 || intent.Agents[0] != model.AgentClaude {
		t.Errorf("Agents = %v, want [claude]", intent.Agents)
	}
	if len(intent.Custom) != 0 {
		t.Errorf("Custom = %v, want empty for Lite", intent.Custom)
	}
}

// TestBuildIntentFull verifies that Full mode works identically to Lite
// w.r.t. the Custom field.
func TestBuildIntentFull(t *testing.T) {
	s := Selection{
		Agents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		Mode:   model.ModeFull,
	}

	intent := s.BuildIntent()

	if intent.Mode != model.ModeFull {
		t.Errorf("Mode = %q, want %q", intent.Mode, model.ModeFull)
	}
	if len(intent.Agents) != 2 {
		t.Errorf("Agents len = %d, want 2", len(intent.Agents))
	}
	if len(intent.Custom) != 0 {
		t.Errorf("Custom = %v, want empty for Full", intent.Custom)
	}
}

// TestBuildIntentCustom verifies that Custom mode populates the Custom field
// from the selected harness IDs.
func TestBuildIntentCustom(t *testing.T) {
	s := Selection{
		Agents:         []model.Agent{model.AgentClaude},
		Mode:           model.ModeCustom,
		CustomHarnesses: []string{"sdd-orchestrator", "engram"},
	}

	intent := s.BuildIntent()

	if intent.Mode != model.ModeCustom {
		t.Errorf("Mode = %q, want %q", intent.Mode, model.ModeCustom)
	}
	if len(intent.Custom) != 2 {
		t.Errorf("Custom = %v, want [sdd-orchestrator engram]", intent.Custom)
	}
}

// TestAvailableAgentsIntersectsDetectedAndRegistered verifies that the
// available-agents helper returns only agents that are BOTH detected and have
// a registered adapter (intersection requirement from the spec).
func TestAvailableAgentsIntersectsDetectedAndRegistered(t *testing.T) {
	detected := []model.Agent{model.AgentClaude, model.AgentOpenCode, model.AgentGemini}
	registered := []model.Agent{model.AgentClaude, model.AgentOpenCode}

	available := availableAgents(detected, registered)

	if len(available) != 2 {
		t.Errorf("len = %d, want 2", len(available))
	}
	for _, a := range available {
		if a != model.AgentClaude && a != model.AgentOpenCode {
			t.Errorf("unexpected agent %q in available list", a)
		}
	}
}

// TestAvailableAgentsEmptyDetected returns empty when nothing is detected.
func TestAvailableAgentsEmptyDetected(t *testing.T) {
	available := availableAgents(nil, []model.Agent{model.AgentClaude})
	if len(available) != 0 {
		t.Errorf("len = %d, want 0", len(available))
	}
}

// Harness filtering by selected agents for the Custom picker is now covered by
// TestCustomPickerHarnesses in internal/install/custom_picker_test.go — the
// TUI delegates to install.CustomPickerHarnesses (design D5,
// windows-contract-tier-multiagent) instead of a local duplicate.
