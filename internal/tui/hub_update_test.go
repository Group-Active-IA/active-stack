package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHub_UpdateStackRunsCallbackAndShowsResult(t *testing.T) {
	called := false
	m := newModel(ModelDeps{RunUpdate: func() (string, error) {
		called = true
		return "Skills: 2 actualizadas", nil
	}})
	m.Cursor = hubUpdate
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	if state.hubNotice != "Updating stack..." || cmd == nil {
		t.Fatalf("notice=%q, cmd=%v", state.hubNotice, cmd)
	}
	updated, _ = state.Update(cmd())
	state = updated.(Model)
	if !called || state.hubNotice != "Skills: 2 actualizadas" {
		t.Fatalf("called=%v, notice=%q", called, state.hubNotice)
	}
}

func TestHub_UpdateStackShowsFailure(t *testing.T) {
	m := newModel(ModelDeps{RunUpdate: func() (string, error) { return "partial", errors.New("offline") }})
	m.Cursor = hubUpdate
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	updated, _ = state.Update(cmd())
	state = updated.(Model)
	if !contains(state.hubNotice, "offline") {
		t.Fatalf("notice=%q", state.hubNotice)
	}
}

func TestHub_UpdateStackUnavailableWithoutCallback(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Cursor = hubUpdate
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	if cmd != nil || !contains(state.hubNotice, "unavailable") {
		t.Fatalf("notice=%q, cmd=%v", state.hubNotice, cmd)
	}
}
