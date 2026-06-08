package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// TestBackNav_EnterStartersSetsPreScreen verifies that selecting Starters from
// the hub sets prevScreen = ScreenWelcome.
func TestBackNav_EnterStartersSetsPreScreen(t *testing.T) {
	m := newModel(ModelDeps{
		Starters: []model.Starter{{ID: "s1", Name: "S1"}},
	})
	m.Screen = ScreenWelcome
	m.Cursor = hubStarters // Starters

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenStarters {
		t.Fatalf("expected ScreenStarters, got %v", state.Screen)
	}
	if state.prevScreen != ScreenWelcome {
		t.Errorf("prevScreen = %v, want ScreenWelcome", state.prevScreen)
	}
}

// TestBackNav_EnterBackupsSetsPreScreen verifies that selecting Manage backups
// from the hub sets prevScreen = ScreenWelcome.
func TestBackNav_EnterBackupsSetsPreScreen(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = hubBackups // Manage backups

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenBackups {
		t.Fatalf("expected ScreenBackups, got %v", state.Screen)
	}
	if state.prevScreen != ScreenWelcome {
		t.Errorf("prevScreen = %v, want ScreenWelcome", state.prevScreen)
	}
}

// TestBackNav_EscFromStartersReturnsToHub verifies Esc on ScreenStarters
// returns to prevScreen (ScreenWelcome).
func TestBackNav_EscFromStartersReturnsToHub(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenStarters
	m.prevScreen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Errorf("Esc from Starters: Screen = %v, want ScreenWelcome", state.Screen)
	}
}

// TestBackNav_EscFromBackupsReturnsToHub verifies Esc on ScreenBackups
// returns to prevScreen (ScreenWelcome).
func TestBackNav_EscFromBackupsReturnsToHub(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenBackups
	m.prevScreen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Errorf("Esc from Backups: Screen = %v, want ScreenWelcome", state.Screen)
	}
}
