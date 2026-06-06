package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// backupAction represents the current action being performed on a backup.
type backupAction int

const (
	backupActionNone    backupAction = iota
	backupActionRestore              // waiting for restore confirmation
	backupActionDelete               // waiting for delete confirmation
	backupActionRename               // capturing new description
)

// backupsState holds the per-screen state for ScreenBackups.
type backupsState struct {
	manifests      []backup.Manifest
	action         backupAction
	actionTarget   int    // index into manifests of the item being acted on
	renameInput    string // accumulated rename input (simple char-by-char)
	lastMsg        string // result message (success/error)
}

// enterBackups transitions the model to ScreenBackups and loads the manifest list.
func (m Model) enterBackups() (tea.Model, tea.Cmd) {
	manifests, err := backup.ListManifests(m.deps.BackupDir)
	var msg string
	if err != nil {
		msg = "Error loading backups: " + err.Error()
	}
	m.backups = backupsState{
		manifests: manifests,
		lastMsg:   msg,
	}
	m.Screen = ScreenBackups
	m.Cursor = 0
	return m, nil
}

// viewBackups renders the backups list screen.
func (m Model) viewBackups() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Backups") + "\n\n")

	switch m.backups.action {
	case backupActionRestore:
		manifest := m.backups.manifests[m.backups.actionTarget]
		sb.WriteString("Restore backup:\n")
		sb.WriteString("  " + manifest.DisplayLabel() + "\n\n")
		sb.WriteString("This will OVERWRITE your current configuration.\n\n")
		sb.WriteString("Enter = confirm restore  Esc = cancel\n")
		return sb.String()

	case backupActionDelete:
		manifest := m.backups.manifests[m.backups.actionTarget]
		sb.WriteString("Delete backup:\n")
		sb.WriteString("  " + manifest.DisplayLabel() + "\n\n")
		sb.WriteString("This will PERMANENTLY DELETE this backup.\n\n")
		sb.WriteString("Enter = confirm delete  Esc = cancel\n")
		return sb.String()

	case backupActionRename:
		manifest := m.backups.manifests[m.backups.actionTarget]
		sb.WriteString("Rename backup:\n")
		sb.WriteString("  " + manifest.DisplayLabel() + "\n\n")
		sb.WriteString("New description: " + m.backups.renameInput + "_\n\n")
		sb.WriteString("Type new description  Enter = confirm  Esc = cancel\n")
		return sb.String()
	}

	// Normal list view.
	if len(m.backups.manifests) == 0 {
		sb.WriteString(dimStyle.Render("No backups found.") + "\n")
		sb.WriteString("\nEsc = back\n")
		return sb.String()
	}

	for i, mf := range m.backups.manifests {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		sb.WriteString(cursor + mf.DisplayLabel() + "\n")
	}

	if m.backups.lastMsg != "" {
		sb.WriteString("\n" + m.backups.lastMsg + "\n")
	}

	sb.WriteString("\nEnter = actions  r = restore  d = delete  n = rename  Esc = back\n")
	return sb.String()
}

// handleBackupsKey handles keyboard input on ScreenBackups.
func (m Model) handleBackupsKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.backups.action {
	case backupActionRestore:
		return m.handleBackupsRestoreConfirm(key)
	case backupActionDelete:
		return m.handleBackupsDeleteConfirm(key)
	case backupActionRename:
		return m.handleBackupsRenameInput(key)
	}

	// Normal list navigation.
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
			m.backups.lastMsg = ""
		}
	case tea.KeyDown:
		if m.Cursor < len(m.backups.manifests)-1 {
			m.Cursor++
			m.backups.lastMsg = ""
		}
	case tea.KeyRunes:
		if len(m.backups.manifests) == 0 {
			break
		}
		switch string(key.Runes) {
		case "r":
			m.backups.action = backupActionRestore
			m.backups.actionTarget = m.Cursor
			m.backups.lastMsg = ""
		case "d":
			m.backups.action = backupActionDelete
			m.backups.actionTarget = m.Cursor
			m.backups.lastMsg = ""
		case "n":
			m.backups.action = backupActionRename
			m.backups.actionTarget = m.Cursor
			m.backups.renameInput = ""
			m.backups.lastMsg = ""
		}
	case tea.KeyEsc:
		m.Screen = m.prevScreen
		m.Cursor = 0
	}
	return m, nil
}

// handleBackupsRestoreConfirm handles the restore confirmation dialog.
func (m Model) handleBackupsRestoreConfirm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		manifest := m.backups.manifests[m.backups.actionTarget]
		svc := backup.RestoreService{}
		if err := svc.Restore(manifest); err != nil {
			m.backups.lastMsg = fmt.Sprintf("Restore failed: %v", err)
		} else {
			m.backups.lastMsg = "Restore succeeded."
		}
		m.backups.action = backupActionNone
	case tea.KeyEsc:
		m.backups.action = backupActionNone
	}
	return m, nil
}

// handleBackupsDeleteConfirm handles the delete confirmation dialog.
func (m Model) handleBackupsDeleteConfirm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		manifest := m.backups.manifests[m.backups.actionTarget]
		if err := backup.DeleteBackup(manifest); err != nil {
			m.backups.lastMsg = fmt.Sprintf("Delete failed: %v", err)
		} else {
			// Remove from list.
			idx := m.backups.actionTarget
			m.backups.manifests = append(
				m.backups.manifests[:idx:idx],
				m.backups.manifests[idx+1:]...,
			)
			if m.Cursor >= len(m.backups.manifests) && m.Cursor > 0 {
				m.Cursor--
			}
			m.backups.lastMsg = "Backup deleted."
		}
		m.backups.action = backupActionNone
	case tea.KeyEsc:
		m.backups.action = backupActionNone
	}
	return m, nil
}

// handleBackupsRenameInput handles the rename input.
func (m Model) handleBackupsRenameInput(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		manifest := m.backups.manifests[m.backups.actionTarget]
		newDesc := m.backups.renameInput
		if err := backup.RenameBackup(manifest, newDesc); err != nil {
			m.backups.lastMsg = fmt.Sprintf("Rename failed: %v", err)
		} else {
			m.backups.manifests[m.backups.actionTarget].Description = newDesc
			m.backups.lastMsg = "Backup renamed."
		}
		m.backups.action = backupActionNone
		m.backups.renameInput = ""
	case tea.KeyEsc:
		m.backups.action = backupActionNone
		m.backups.renameInput = ""
	case tea.KeyBackspace:
		if len(m.backups.renameInput) > 0 {
			m.backups.renameInput = m.backups.renameInput[:len(m.backups.renameInput)-1]
		}
	case tea.KeyRunes:
		m.backups.renameInput += string(key.Runes)
	}
	return m, nil
}
