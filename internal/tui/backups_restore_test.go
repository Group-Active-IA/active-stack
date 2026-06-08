package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// TestBackups_RestoreShowsConfirmation verifies that pressing 'r' enters
// the restore confirmation state (action = backupActionRestore).
func TestBackups_RestoreShowsConfirmation(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	// Press 'r' to initiate restore.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	state := updated.(Model)

	if state.backups.action != backupActionRestore {
		t.Errorf("action = %v, want backupActionRestore", state.backups.action)
	}
	view := state.viewBackups()
	if !contains(view, "confirm") && !contains(view, "Restore") {
		t.Errorf("restore confirmation not shown in view:\n%s", view)
	}
}

// TestBackups_RestoreDoesNotExecuteWithoutConfirm verifies that pressing 'r'
// alone (not confirmed) does NOT invoke restore.
// The action sentinel stays active; restore only runs on Enter.
func TestBackups_RestoreDoesNotExecuteWithoutConfirm(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	state := updated.(Model)

	// Action is pending — no restore yet. Esc cancels it.
	updated2, _ := state.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state2 := updated2.(Model)

	if state2.backups.action != backupActionNone {
		t.Errorf("after Esc: action = %v, want backupActionNone", state2.backups.action)
	}
	// The screen is still ScreenBackups (Esc only cancels the action, not the screen).
	if state2.Screen != ScreenBackups {
		t.Errorf("after cancel restore: Screen = %v, want ScreenBackups", state2.Screen)
	}
}
