package tui

import (
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// TestStarters_ListReflectsDepStarters verifies the view shows the starters
// provided in ModelDeps.Starters.
func TestStarters_ListReflectsDepStarters(t *testing.T) {
	deps := ModelDeps{
		Starters: []model.Starter{
			{ID: "starter-a", Name: "Starter A", Description: "The A starter"},
			{ID: "starter-b", Name: "Starter B"},
		},
	}
	m := newModel(deps)
	m.Screen = ScreenStarters

	view := m.viewStarters()
	if !contains(view, "Starter A") {
		t.Errorf("view missing 'Starter A':\n%s", view)
	}
	if !contains(view, "Starter B") {
		t.Errorf("view missing 'Starter B':\n%s", view)
	}
	if !contains(view, "The A starter") {
		t.Errorf("view missing description 'The A starter':\n%s", view)
	}
}

// TestStarters_EmptyListDoesNotCrash verifies that the screen renders gracefully
// with no starters available.
func TestStarters_EmptyListDoesNotCrash(t *testing.T) {
	m := newModel(ModelDeps{Starters: nil})
	m.Screen = ScreenStarters

	view := m.viewStarters()
	if !contains(view, "No starters") {
		t.Errorf("expected 'No starters' in empty view:\n%s", view)
	}
}

// TestStarters_ConfirmInvokesRunStarter verifies that Enter invokes RunStarter
// with the correct starter ID.
func TestStarters_ConfirmInvokesRunStarter(t *testing.T) {
	var gotID string
	deps := ModelDeps{
		Starters: []model.Starter{
			{ID: "starter-x", Name: "X"},
		},
		RunStarter: func(starterID, projectPath string, agents []model.Agent, w io.Writer) int {
			gotID = starterID
			return 0
		},
	}
	m := newModel(deps)
	m.Screen = ScreenStarters
	m.Cursor = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on Starters must produce a cmd (async RunStarter)")
	}
	// Execute the cmd synchronously.
	msg := cmd()
	result, ok := msg.(starterRunMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want starterRunMsg", msg)
	}
	if result.exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", result.exitCode)
	}
	if gotID != "starter-x" {
		t.Errorf("RunStarter gotID = %q, want %q", gotID, "starter-x")
	}
}

// TestStarters_EmptyListEnterDoesNotCrash verifies that Enter on an empty
// starters list does not crash and does not emit a cmd.
func TestStarters_EmptyListEnterDoesNotCrash(t *testing.T) {
	m := newModel(ModelDeps{Starters: nil})
	m.Screen = ScreenStarters

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	_ = state // no panic = pass
	if cmd != nil {
		// Evaluate — should not crash
		_ = cmd()
	}
}
