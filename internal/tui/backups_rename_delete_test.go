package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// TestBackups_DeleteShowsConfirmation verifies 'd' enters the delete
// confirmation state.
func TestBackups_DeleteShowsConfirmation(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	state := updated.(Model)

	if state.backups.action != backupActionDelete {
		t.Errorf("action = %v, want backupActionDelete", state.backups.action)
	}
	view := state.viewBackups()
	if !contains(view, "delete") && !contains(view, "Delete") {
		t.Errorf("delete confirmation not shown in view:\n%s", view)
	}
}

// TestBackups_DeleteCancelDoesNotDelete verifies that Esc on delete
// confirmation cancels without deleting.
func TestBackups_DeleteCancelDoesNotDelete(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	// Enter delete confirmation.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	state := updated.(Model)

	// Cancel with Esc.
	updated2, _ := state.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state2 := updated2.(Model)

	if state2.backups.action != backupActionNone {
		t.Errorf("after cancel: action = %v, want backupActionNone", state2.backups.action)
	}
	if len(state2.backups.manifests) != 1 {
		t.Errorf("manifests len = %d, want 1 (no delete happened)", len(state2.backups.manifests))
	}
}

// TestBackups_RenameEntersInputMode verifies 'n' enters the rename input state.
func TestBackups_RenameEntersInputMode(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	state := updated.(Model)

	if state.backups.action != backupActionRename {
		t.Errorf("action = %v, want backupActionRename", state.backups.action)
	}
}

// TestBackups_RenameAccumulates verifies that typing characters accumulates
// the rename input.
func TestBackups_RenameAccumulates(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0
	m.backups.action = backupActionRename
	m.backups.actionTarget = 0
	m.backups.renameInput = ""

	// Type "abc".
	for _, r := range []rune("abc") {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if m.backups.renameInput != "abc" {
		t.Errorf("renameInput = %q, want %q", m.backups.renameInput, "abc")
	}
}
