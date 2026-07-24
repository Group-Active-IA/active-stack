package headless

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
)

// windowsBackupSlots are the only backup slots the Windows GUI's "Gestionar
// backups" screen surfaces: the ones internal/install/plan.go and
// internal/uninstall/plan.go write to (design D4). <home>/.active-stack/backups
// also holds several OTHER, unrelated per-harness/per-adapter scratch backup
// dirs used purely as internal rollback state for a single install run (e.g.
// "sdd-orchestrator-claude", "engram-claude-mcp" — see internal/harness/config
// install.go and internal/install/plan.go's skillStep/commandStep backupDir).
// Those must never be listed here: they are not meaningful restore points to
// a user, and scanning the whole backups/ dir picks them up as noise.
var windowsBackupSlots = []string{"install", "uninstall"}

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

// RunWindowsBackupsList reads the manifest for each windowsBackupSlots entry
// under <home>/.active-stack/backups and emits {"backups":[...]} (design D4).
// Read-only; internal/backup is untouched.
//
// Each slot's manifest.json is read DIRECTLY (backup.ReadManifest), not via
// backup.ListManifests: internal/install/plan.go and internal/uninstall/plan.go
// always snapshot to the FIXED, non-timestamped paths
// <home>/.active-stack/backups/install and .../uninstall — each run
// overwrites the previous one, so "install"/"uninstall" ARE the backup slots
// themselves (manifest.json lives directly inside them), not containers of
// multiple per-event subdirectories the way ListManifests expects. Calling
// ListManifests on "backups/install" looks one level too deep and always
// finds nothing (the bug); calling it on "backups/" itself finds the slots
// but ALSO picks up unrelated per-harness scratch backup dirs as noise (see
// windowsBackupSlots) — reading the two known paths directly avoids both.
// A missing or empty store yields {"backups":[]}.
func RunWindowsBackupsList(homeDir string, lang i18n.Lang, w io.Writer) error {
	manifests, err := readWindowsBackupSlots(homeDir)
	if err != nil {
		return err
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

// readWindowsBackupSlots reads the manifest for each windowsBackupSlots entry
// that actually exists under <home>/.active-stack/backups, skipping any slot
// that hasn't been written yet (e.g. "uninstall" before the first uninstall
// ever ran). Shared by RunWindowsBackupsList and findWindowsBackupManifest so
// the two cannot drift.
func readWindowsBackupSlots(homeDir string) ([]backup.Manifest, error) {
	root := filepath.Join(homeDir, ".active-stack", "backups")

	var manifests []backup.Manifest
	for _, slot := range windowsBackupSlots {
		m, err := backup.ReadManifest(filepath.Join(root, slot, backup.ManifestFilename))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
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
