// Package headless — tests for "windows backups list" (windows-contract-hub-operations,
// Task 4.1 RED). Manifests are fabricated directly via backup.WriteManifest
// under a temp --home backup store; internal/backup itself is NOT modified
// (governance ALTO — callers only).
package headless_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bytes"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// fabricateManifest writes a manifest.json under
// <home>/.active-stack/backups/<sub>/<id>/manifest.json using ONLY the
// public backup.WriteManifest API, and returns the manifest with RootDir set.
func fabricateManifest(t *testing.T, home, sub, id string, m backup.Manifest) backup.Manifest {
	t.Helper()
	rootDir := filepath.Join(home, ".active-stack", "backups", sub, id)
	m.ID = id
	m.RootDir = rootDir
	if err := backup.WriteManifest(filepath.Join(rootDir, backup.ManifestFilename), m); err != nil {
		t.Fatalf("fabricate manifest %s/%s: %v", sub, id, err)
	}
	return m
}

// TestRunWindowsBackupsList_ListsFabricatedManifests asserts one entry per
// fabricated manifest, with all D4 fields populated and manifest_path valid.
func TestRunWindowsBackupsList_ListsFabricatedManifests(t *testing.T) {
	home := t.TempDir()

	created := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)
	m1 := fabricateManifest(t, home, "install", "backup-1", backup.Manifest{
		CreatedAt:   created,
		Source:      backup.BackupSourceInstall,
		Description: "before upgrade",
		FileCount:   3,
		Pinned:      true,
	})
	m2 := fabricateManifest(t, home, "uninstall", "backup-2", backup.Manifest{
		CreatedAt:  created.Add(time.Hour),
		Source:     backup.BackupSourceSync,
		FileCount:  1,
		Compressed: true,
	})

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, &out); err != nil {
		t.Fatalf("RunWindowsBackupsList() error = %v", err)
	}

	var resp struct {
		Backups []struct {
			ID           string `json:"id"`
			CreatedAt    string `json:"created_at"`
			Source       string `json:"source"`
			Description  string `json:"description"`
			FileCount    int    `json:"file_count"`
			Pinned       bool   `json:"pinned"`
			Compressed   bool   `json:"compressed"`
			DisplayLabel string `json:"display_label"`
			ManifestPath string `json:"manifest_path"`
		} `json:"backups"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal backups list json: %v\nbody=%s", err, out.String())
	}

	if len(resp.Backups) != 2 {
		t.Fatalf("backups len = %d, want 2; body=%s", len(resp.Backups), out.String())
	}

	byID := make(map[string]struct {
		ID           string `json:"id"`
		CreatedAt    string `json:"created_at"`
		Source       string `json:"source"`
		Description  string `json:"description"`
		FileCount    int    `json:"file_count"`
		Pinned       bool   `json:"pinned"`
		Compressed   bool   `json:"compressed"`
		DisplayLabel string `json:"display_label"`
		ManifestPath string `json:"manifest_path"`
	})
	for _, b := range resp.Backups {
		byID[b.ID] = b
	}

	b1, ok := byID["backup-1"]
	if !ok {
		t.Fatalf("backup-1 not found in response: %+v", resp.Backups)
	}
	if b1.Source != "install" || b1.Description != "before upgrade" || b1.FileCount != 3 || !b1.Pinned {
		t.Errorf("backup-1 = %+v, want source=install description='before upgrade' file_count=3 pinned=true", b1)
	}
	if b1.DisplayLabel != m1.DisplayLabel() {
		t.Errorf("backup-1 display_label = %q, want %q", b1.DisplayLabel, m1.DisplayLabel())
	}
	wantPath1 := filepath.Join(m1.RootDir, backup.ManifestFilename)
	if b1.ManifestPath != wantPath1 {
		t.Errorf("backup-1 manifest_path = %q, want %q", b1.ManifestPath, wantPath1)
	}
	if _, err := os.Stat(b1.ManifestPath); err != nil {
		t.Errorf("backup-1 manifest_path does not exist on disk: %v", err)
	}

	b2, ok := byID["backup-2"]
	if !ok {
		t.Fatalf("backup-2 not found in response: %+v", resp.Backups)
	}
	if b2.Source != "sync" || b2.FileCount != 1 || !b2.Compressed || b2.Pinned {
		t.Errorf("backup-2 = %+v, want source=sync file_count=1 compressed=true pinned=false", b2)
	}
	if b2.DisplayLabel != m2.DisplayLabel() {
		t.Errorf("backup-2 display_label = %q, want %q", b2.DisplayLabel, m2.DisplayLabel())
	}
}

// TestRunWindowsBackupsList_EmptyStore asserts a home with no backup store
// yields {"backups":[]} and exit is implied zero (no error).
func TestRunWindowsBackupsList_EmptyStore(t *testing.T) {
	home := t.TempDir()

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, &out); err != nil {
		t.Fatalf("RunWindowsBackupsList() error = %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v\nbody=%s", err, out.String())
	}
	if string(raw["backups"]) != "[]" {
		t.Fatalf("backups = %s, want []", raw["backups"])
	}
}
