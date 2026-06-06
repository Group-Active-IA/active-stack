package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestHub_UpdateStackShowsComingSoon verifies that selecting "Update stack"
// (index 4) sets hubNotice to "coming soon" and stays on ScreenWelcome.
func TestHub_UpdateStackShowsComingSoon(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = hubUpdate // "Update stack"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenWelcome {
		t.Errorf("Screen = %v, want ScreenWelcome (no screen entered)", state.Screen)
	}
	if state.hubNotice == "" {
		t.Error("hubNotice should be set to 'coming soon'")
	}
	view := state.viewWelcome()
	if !contains(view, "coming soon") {
		t.Errorf("view should show 'coming soon':\n%s", view)
	}
}

// TestHub_UpdateStackHubNoticeClears verifies that navigating away and back
// to the hub resets the notice (via cursor movement which clears hubNotice).
func TestHub_UpdateStackHubNoticeClears(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenWelcome
	m.Cursor = hubUpdate

	// Select Update stack — sets notice.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	if state.hubNotice == "" {
		t.Fatal("hubNotice not set after Update stack selection")
	}

	// Move cursor — notice should clear.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyDown})
	state = updated.(Model)
	if state.hubNotice != "" {
		t.Errorf("hubNotice = %q, want empty after cursor move", state.hubNotice)
	}
}
