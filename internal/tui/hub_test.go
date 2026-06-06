package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// hubMenuItems is the canonical list to test against.
var hubMenuItems = []string{
	"Install",
	"Starters",
	"Manage backups",
	"Uninstall",
	"Update stack",
	"Quit",
}

// TestHub_ViewShowsSixItems verifies the hub renders all 6 menu items.
func TestHub_ViewShowsSixItems(t *testing.T) {
	m := newModel(ModelDeps{})
	view := m.viewWelcome()
	for _, item := range hubMenuItems {
		if !contains(view, item) {
			t.Errorf("viewWelcome() missing item %q\n--- view ---\n%s", item, view)
		}
	}
}

// TestHub_CursorDownMoves verifies Down key increments cursor from 0.
func TestHub_CursorDownMoves(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	state := updated.(Model)
	if state.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1 after Down", state.Cursor)
	}
}

// TestHub_CursorUpAtZeroStays verifies Up key at 0 does not go negative.
func TestHub_CursorUpAtZeroStays(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	state := updated.(Model)
	if state.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0 at top boundary", state.Cursor)
	}
}

// TestHub_CursorDownAtLastStays verifies Down key at last item stays.
func TestHub_CursorDownAtLastStays(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = len(hubMenuItems) - 1 // last item

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	state := updated.(Model)
	if state.Cursor != len(hubMenuItems)-1 {
		t.Errorf("Cursor = %d, want %d at bottom boundary", state.Cursor, len(hubMenuItems)-1)
	}
}

// TestHub_SelectInstallRoutes verifies selecting Install (cursor=0) enters the install flow.
func TestHub_SelectInstallRoutes(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = 0 // Install

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	// Install should advance to ScreenDetection (start of existing install flow).
	if state.Screen != ScreenDetection {
		t.Errorf("selecting Install got Screen=%v, want ScreenDetection", state.Screen)
	}
}

// TestHub_SelectQuitEmitsQuit verifies selecting Quit (cursor=5) emits tea.Quit.
func TestHub_SelectQuitEmitsQuit(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = 5 // Quit (6th item, index 5)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("selecting Quit must return a non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("cmd() returned %T, want tea.QuitMsg", msg)
	}
}
