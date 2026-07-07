package headless

import (
	"encoding/json"
	"io"
	"path/filepath"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// windowsBackupStoreSubdirs are the two backup roots under
// <home>/.active-stack/backups, mirroring internal/install/plan.go and
// internal/uninstall/plan.go (design D4). This change does NOT modify
// internal/backup, internal/install, or internal/uninstall — it only reads
// via the public backup.ListManifests API.
var windowsBackupStoreSubdirs = []string{"install", "uninstall"}

// windowsBackupEntry is one entry in the "windows backups list" response
// (design D4).
type windowsBackupEntry struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	Source       string `json:"source"`
	Description  string `json:"description"`
	FileCount    int    `json:"file_count"`
	Pinned       bool   `json:"pinned"`
	Compressed   bool   `json:"compressed"`
	DisplayLabel string `json:"display_label"`
	ManifestPath string `json:"manifest_path"`
}

type windowsBackupsListResponse struct {
	Backups []windowsBackupEntry `json:"backups"`
}

// RunWindowsBackupsList aggregates backup.ListManifests over
// <home>/.active-stack/backups/{install,uninstall} and emits
// {"backups":[...]} (design D4). Read-only; internal/backup is untouched.
// A missing or empty store yields {"backups":[]}.
func RunWindowsBackupsList(homeDir string, w io.Writer) error {
	root := filepath.Join(homeDir, ".active-stack", "backups")

	var manifests []backup.Manifest
	for _, sub := range windowsBackupStoreSubdirs {
		ms, err := backup.ListManifests(filepath.Join(root, sub))
		if err != nil {
			return err
		}
		manifests = append(manifests, ms...)
	}

	resp := windowsBackupsListResponse{Backups: make([]windowsBackupEntry, 0, len(manifests))}
	for _, m := range manifests {
		resp.Backups = append(resp.Backups, windowsBackupEntry{
			ID:           m.ID,
			CreatedAt:    m.CreatedAt.UTC().Format(time.RFC3339),
			Source:       m.Source.Label(),
			Description:  m.Description,
			FileCount:    m.FileCount,
			Pinned:       m.Pinned,
			Compressed:   m.Compressed,
			DisplayLabel: m.DisplayLabel(),
			ManifestPath: filepath.Join(m.RootDir, backup.ManifestFilename),
		})
	}

	return json.NewEncoder(w).Encode(resp)
}
