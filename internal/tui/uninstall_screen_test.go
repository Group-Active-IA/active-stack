package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestUninstall_AgentsToMode verifies that Enter on ScreenUninstallAgents
// advances to ScreenUninstallMode.
func TestUninstall_AgentsToMode(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude},
	})
	m.Screen = ScreenUninstallAgents
	m.prevScreen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenUninstallMode {
		t.Errorf("Screen = %v, want ScreenUninstallMode", state.Screen)
	}
}

// TestUninstall_ModeToStrategy verifies that Enter on ScreenUninstallMode
// advances to ScreenUninstallStrategy.
func TestUninstall_ModeToStrategy(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenUninstallMode
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenUninstallStrategy {
		t.Errorf("Screen = %v, want ScreenUninstallStrategy", state.Screen)
	}
	if state.uninstallSel.mode != model.ModeLite {
		t.Errorf("mode = %v, want ModeLite (cursor=0)", state.uninstallSel.mode)
	}
}

// TestUninstall_StrategyToConfirm verifies that Enter on ScreenUninstallStrategy
// advances to ScreenUninstallConfirm.
func TestUninstall_StrategyToConfirm(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenUninstallStrategy
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenUninstallConfirm {
		t.Errorf("Screen = %v, want ScreenUninstallConfirm", state.Screen)
	}
	if state.uninstallSel.strategy != uninstall.StrategyTargeted {
		t.Errorf("strategy = %v, want StrategyTargeted", state.uninstallSel.strategy)
	}
}

// TestUninstall_SelectionPreservedAcrossScreens verifies that agent selection
// from ScreenUninstallAgents is preserved when transitioning to Mode.
func TestUninstall_SelectionPreservedAcrossScreens(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
	})
	m.Screen = ScreenUninstallAgents
	m.Cursor = 0

	// Toggle Claude.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	state := updated.(Model)

	// Advance to Mode.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenUninstallMode {
		t.Fatalf("expected ScreenUninstallMode, got %v", state.Screen)
	}
	if !isAgentSelected(state.uninstallSel.agents, model.AgentClaude) {
		t.Errorf("agent selection lost: uninstallSel.agents = %v", state.uninstallSel.agents)
	}
}

// TestUninstall_EscFromModeGoesBackToAgents verifies Esc on Mode goes to Agents.
func TestUninstall_EscFromModeGoesBackToAgents(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenUninstallMode

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenUninstallAgents {
		t.Errorf("Screen = %v, want ScreenUninstallAgents", state.Screen)
	}
}

// TestUninstall_EscFromAgentsGoesBackToHub verifies Esc on Agents goes to prevScreen.
func TestUninstall_EscFromAgentsGoesBackToHub(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenUninstallAgents
	m.prevScreen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Errorf("Screen = %v, want ScreenWelcome", state.Screen)
	}
}
