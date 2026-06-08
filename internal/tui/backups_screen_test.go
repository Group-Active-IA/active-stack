package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// makeTestBackupsState returns a Model at ScreenBackups with pre-loaded manifests.
func makeTestBackupsState(manifests []backup.Manifest) Model {
	m := newModel(ModelDeps{})
	m.Screen = ScreenBackups
	m.prevScreen = ScreenWelcome
	m.backups = backupsState{
		manifests: manifests,
	}
	return m
}

// sampleManifest creates a test manifest with known values.
func sampleManifest(id string) backup.Manifest {
	return backup.Manifest{
		ID:        id,
		CreatedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
		Source:    backup.BackupSourceInstall,
		FileCount: 3,
	}
}

// TestBackups_ViewListsManifests verifies the backups view shows manifests.
func TestBackups_ViewListsManifests(t *testing.T) {
	manifests := []backup.Manifest{
		sampleManifest("b1"),
		sampleManifest("b2"),
	}
	m := makeTestBackupsState(manifests)
	view := m.viewBackups()

	// Both manifests' display labels should be present.
	for _, mf := range manifests {
		label := mf.DisplayLabel()
		if !contains(view, label) {
			t.Errorf("view missing manifest label %q:\n%s", label, view)
		}
	}
}

// TestBackups_EmptyListDoesNotCrash verifies the empty-list case.
func TestBackups_EmptyListDoesNotCrash(t *testing.T) {
	m := makeTestBackupsState(nil)
	view := m.viewBackups()
	if !contains(view, "No backups") {
		t.Errorf("expected 'No backups' in empty view:\n%s", view)
	}
}

// TestBackups_CursorDown verifies Down key moves the cursor.
func TestBackups_CursorDown(t *testing.T) {
	manifests := []backup.Manifest{sampleManifest("b1"), sampleManifest("b2")}
	m := makeTestBackupsState(manifests)
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	state := updated.(Model)
	if state.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", state.Cursor)
	}
}

// TestBackups_EscReturnsToHub verifies Esc returns to prevScreen.
func TestBackups_EscReturnsToHub(t *testing.T) {
	m := makeTestBackupsState(nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)
	if state.Screen != ScreenWelcome {
		t.Errorf("Screen = %v, want ScreenWelcome", state.Screen)
	}
}
