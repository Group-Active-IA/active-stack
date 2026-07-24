// Package headless — tests for "windows backups list" (windows-contract-hub-operations,
// Task 4.1 RED). Manifests are fabricated directly via backup.WriteManifest
// under a temp --home backup store; internal/backup itself is NOT modified
// (governance ALTO — callers only).
//
// Fabrication mirrors the REAL shape internal/install/plan.go and
// internal/uninstall/plan.go actually produce: manifest.json lives directly
// inside <home>/.active-stack/backups/<sub> (sub = "install" or "uninstall"),
// with Manifest.ID == sub (backup.Snapshotter.Create sets
// ID = filepath.Base(snapshotDir), and snapshotDir IS that fixed path — see
// resolveSnapshotHomeDir in internal/install/plan.go). There is only ever ONE
// slot per operation type today (each run overwrites the previous one) — NOT
// an arbitrary per-event subdirectory keyed by a custom id. An earlier
// version of this test fabricated one extra level of nesting
// (backups/<sub>/<custom-id>/manifest.json) that the real snapshotter never
// writes, which is exactly why "Gestionar backups" showed nothing in
// production despite real backups existing on disk (RunWindowsBackupsList
// was calling ListManifests one level too deep).
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
	"github.com/Group-Active-IA/active-stack/internal/i18n"
)

// fabricateManifest writes a manifest.json under
// <home>/.active-stack/backups/<sub>/manifest.json (sub is the ID: "install"
// or "uninstall") using ONLY the public backup.WriteManifest API, and returns
// the manifest with ID/RootDir set.
func fabricateManifest(t *testing.T, home, sub string, m backup.Manifest) backup.Manifest {
	t.Helper()
	rootDir := filepath.Join(home, ".active-stack", "backups", sub)
	m.ID = sub
	m.RootDir = rootDir
	if err := backup.WriteManifest(filepath.Join(rootDir, backup.ManifestFilename), m); err != nil {
		t.Fatalf("fabricate manifest %s: %v", sub, err)
	}
	return m
}

// TestRunWindowsBackupsList_ListsFabricatedManifests asserts one entry per
// fabricated manifest, with all D4 fields populated and manifest_path valid.
func TestRunWindowsBackupsList_ListsFabricatedManifests(t *testing.T) {
	home := t.TempDir()

	created := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)
	m1 := fabricateManifest(t, home, "install", backup.Manifest{
		CreatedAt:   created,
		Source:      backup.BackupSourceInstall,
		Description: "before upgrade",
		FileCount:   3,
		Pinned:      true,
	})
	m2 := fabricateManifest(t, home, "uninstall", backup.Manifest{
		CreatedAt:  created.Add(time.Hour),
		Source:     backup.BackupSourceSync,
		FileCount:  1,
		Compressed: true,
	})

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, i18n.LangEN, &out); err != nil {
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

	b1, ok := byID["install"]
	if !ok {
		t.Fatalf("install backup not found in response: %+v", resp.Backups)
	}
	if b1.Source != "install" || b1.Description != "before upgrade" || b1.FileCount != 3 || !b1.Pinned {
		t.Errorf("install backup = %+v, want source=install description='before upgrade' file_count=3 pinned=true", b1)
	}
	if b1.DisplayLabel != m1.DisplayLabel() {
		t.Errorf("install backup display_label = %q, want %q", b1.DisplayLabel, m1.DisplayLabel())
	}
	wantPath1 := filepath.Join(m1.RootDir, backup.ManifestFilename)
	if b1.ManifestPath != wantPath1 {
		t.Errorf("install backup manifest_path = %q, want %q", b1.ManifestPath, wantPath1)
	}
	if _, err := os.Stat(b1.ManifestPath); err != nil {
		t.Errorf("install backup manifest_path does not exist on disk: %v", err)
	}

	b2, ok := byID["uninstall"]
	if !ok {
		t.Fatalf("uninstall backup not found in response: %+v", resp.Backups)
	}
	if b2.Source != "sync" || b2.FileCount != 1 || !b2.Compressed || b2.Pinned {
		t.Errorf("uninstall backup = %+v, want source=sync file_count=1 compressed=true pinned=false", b2)
	}
	if b2.DisplayLabel != m2.DisplayLabel() {
		t.Errorf("uninstall backup display_label = %q, want %q", b2.DisplayLabel, m2.DisplayLabel())
	}
}

// TestRunWindowsBackupsList_FindsARealSnapshotterOutput is the regression
// guard for the production bug: rather than fabricating a manifest by hand,
// it drives the SAME public backup.Snapshotter API internal/install/plan.go
// and internal/uninstall/plan.go actually call, writing to the exact fixed
// path (<home>/.active-stack/backups/install) they use, then asserts
// RunWindowsBackupsList finds it. This is what a real "windows install" run
// produces on disk (confirmed manually), and what the previous version of
// this test suite never actually exercised.
func TestRunWindowsBackupsList_FindsARealSnapshotterOutput(t *testing.T) {
	home := t.TempDir()

	configuredFile := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(configuredFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configuredFile, []byte("existing content"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	snapshotDir := filepath.Join(home, ".active-stack", "backups", "install")
	if _, err := backup.NewSnapshotter().Create(snapshotDir, []string{configuredFile}); err != nil {
		t.Fatalf("Snapshotter.Create: %v", err)
	}

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, i18n.LangEN, &out); err != nil {
		t.Fatalf("RunWindowsBackupsList() error = %v", err)
	}

	var resp struct {
		Backups []struct {
			ID string `json:"id"`
		} `json:"backups"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal backups list json: %v\nbody=%s", err, out.String())
	}

	if len(resp.Backups) != 1 || resp.Backups[0].ID != "install" {
		t.Fatalf("backups = %+v, want exactly one entry with id=install", resp.Backups)
	}
}

// TestRunWindowsBackupsList_EmptyStore asserts a home with no backup store
// yields {"backups":[]} and exit is implied zero (no error).
func TestRunWindowsBackupsList_EmptyStore(t *testing.T) {
	home := t.TempDir()

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, i18n.LangEN, &out); err != nil {
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

// TestRunWindowsBackupsList_LocalizedSourceWithRawFallback covers task 6.1:
// a backup with a known source label ("install") localizes under --lang es,
// while a source with no registered i18n key falls back to the raw label
// string returned by Source.Label() exactly.
func TestRunWindowsBackupsList_LocalizedSourceWithRawFallback(t *testing.T) {
	home := t.TempDir()

	fabricateManifest(t, home, "install", backup.Manifest{
		Source: backup.BackupSourceInstall,
	})
	fabricateManifest(t, home, "uninstall", backup.Manifest{
		Source: backup.BackupSource("other"),
	})

	var out bytes.Buffer
	if err := headless.RunWindowsBackupsList(home, i18n.LangES, &out); err != nil {
		t.Fatalf("RunWindowsBackupsList() error = %v", err)
	}

	var resp struct {
		Backups []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		} `json:"backups"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal backups list json: %v\nbody=%s", err, out.String())
	}

	byID := make(map[string]string, len(resp.Backups))
	for _, b := range resp.Backups {
		byID[b.ID] = b.Source
	}

	if got := byID["install"]; got != "instalación" {
		t.Errorf("install source = %q, want localized %q", got, "instalación")
	}
	if got := byID["uninstall"]; got != backup.BackupSource("other").Label() {
		t.Errorf("uninstall source = %q, want raw fallback %q", got, backup.BackupSource("other").Label())
	}
}
