package headless

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// windowsBackupActionResponse is the response shape for
// "windows backups restore|delete|rename" (design D5).
type windowsBackupActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

// RunWindowsBackupsAction dispatches "windows backups restore|delete|rename"
// (design D5). It locates the manifest by id among the known backup slots
// (readWindowsBackupSlots), then acts on it using ONLY the public
// internal/backup API: RestoreService{}.Restore, DeleteBackup, RenameBackup.
// internal/backup is NOT modified (governance ALTO — callers only).
func RunWindowsBackupsAction(homeDir, action, id, description string, w io.Writer) int {
	manifest, found, err := findWindowsBackupManifest(homeDir, id)
	if err != nil {
		return encodeWindowsBackupAction(w, false, fmt.Sprintf("list backups: %v", err), id)
	}
	if !found {
		return encodeWindowsBackupAction(w, false, fmt.Sprintf("backup %q not found", id), id)
	}

	switch action {
	case "restore":
		if err := (backup.RestoreService{}).Restore(manifest); err != nil {
			return encodeWindowsBackupAction(w, false, err.Error(), id)
		}
		return encodeWindowsBackupAction(w, true, "Backup restored successfully.", id)

	case "delete":
		if err := backup.DeleteBackup(manifest); err != nil {
			return encodeWindowsBackupAction(w, false, err.Error(), id)
		}
		return encodeWindowsBackupAction(w, true, "Backup deleted successfully.", id)

	case "rename":
		if err := backup.RenameBackup(manifest, description); err != nil {
			return encodeWindowsBackupAction(w, false, err.Error(), id)
		}
		return encodeWindowsBackupAction(w, true, "Backup renamed successfully.", id)

	default:
		return encodeWindowsBackupAction(w, false, fmt.Sprintf("unknown backup action %q", action), id)
	}
}

// findWindowsBackupManifest looks up the manifest whose ID matches id among
// the known backup slots (readWindowsBackupSlots, shared with
// RunWindowsBackupsList — see windows_backups.go for why this reads each
// slot's manifest.json directly instead of via backup.ListManifests).
func findWindowsBackupManifest(homeDir, id string) (backup.Manifest, bool, error) {
	manifests, err := readWindowsBackupSlots(homeDir)
	if err != nil {
		return backup.Manifest{}, false, err
	}
	for _, m := range manifests {
		if m.ID == id {
			return m, true, nil
		}
	}
	return backup.Manifest{}, false, nil
}

func encodeWindowsBackupAction(w io.Writer, success bool, message, id string) int {
	_ = json.NewEncoder(w).Encode(windowsBackupActionResponse{Success: success, Message: message, ID: id})
	if success {
		return 0
	}
	return 1
}
