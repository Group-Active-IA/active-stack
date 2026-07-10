package headless

import (
	"encoding/json"
	"io"
	"path/filepath"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
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
func RunWindowsBackupsList(homeDir string, lang i18n.Lang, w io.Writer) error {
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
			Source:       localizedBackupSource(lang, m.Source.Label()),
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

// localizedBackupSource maps a backup.Source.Label() value (e.g. "install",
// "sync", "upgrade", "unknown source") through the i18n "backup.source.*"
// tables (design D4, i18n-engine-locales). i18n.T's own missing-key fallback
// returns the composed key itself, which is not a usable label; when no
// "backup.source.<label>" key is registered (e.g. "unknown source",
// deliberately unregistered), this falls back to the raw label string
// exactly as internal/backup returns it.
func localizedBackupSource(lang i18n.Lang, rawLabel string) string {
	key := "backup.source." + rawLabel
	localized := i18n.T(lang, key)
	if localized == key {
		return rawLabel
	}
	return localized
}
