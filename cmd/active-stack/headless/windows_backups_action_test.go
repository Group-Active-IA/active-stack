// Package headless — tests for "windows backups restore|delete|rename"
// (windows-contract-hub-operations, Tasks 5.1/5.2/5.3 RED). Manifests are
// fabricated via the public backup API only; internal/backup is NOT modified
// (governance ALTO — callers only).
package headless_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
)

type backupActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

// TestRunWindowsBackupsAction_Restore asserts that "restore --id <known>"
// invokes RestoreService{}.Restore on the matching manifest (verified via the
// restored file content) and returns success:true with the id.
func TestRunWindowsBackupsAction_Restore(t *testing.T) {
	home := t.TempDir()

	// Fabricate a snapshot file and target: the target existed with old
	// content, RestoreService.Restore should overwrite it back from the
	// uncompressed snapshot.
	snapshotDir := filepath.Join(home, ".active-stack", "backups", "install", "backup-1")
	snapshotFile := filepath.Join(snapshotDir, "snapshot-config.json")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir: %v", err)
	}
	if err := os.WriteFile(snapshotFile, []byte("original content"), 0o644); err != nil {
		t.Fatalf("write snapshot file: %v", err)
	}

	targetFile := filepath.Join(home, "config.json")
	if err := os.WriteFile(targetFile, []byte("mutated content"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}

	manifest := backup.Manifest{
		ID:      "backup-1",
		RootDir: snapshotDir,
		Source:  backup.BackupSourceInstall,
		Entries: []backup.ManifestEntry{
			{OriginalPath: targetFile, SnapshotPath: snapshotFile, Existed: true, IsDir: false, Mode: 0o644},
		},
	}
	if err := backup.WriteManifest(filepath.Join(snapshotDir, backup.ManifestFilename), manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsBackupsAction(home, "restore", "backup-1", "", &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsBackupsAction(restore) exit = %d; output:\n%s", exitCode, out.String())
	}

	var resp backupActionResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal restore response: %v\nbody=%s", err, out.String())
	}
	if !resp.Success || resp.ID != "backup-1" {
		t.Fatalf("restore response = %+v, want success=true id=backup-1", resp)
	}

	restored, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("read restored target file: %v", err)
	}
	if string(restored) != "original content" {
		t.Fatalf("restored target file = %q, want %q (RestoreService.Restore side effect)", restored, "original content")
	}
}

// TestRunWindowsBackupsAction_Delete asserts "delete --id <known>" removes
// the backup via DeleteBackup and returns success:true.
func TestRunWindowsBackupsAction_Delete(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, ".active-stack", "backups", "install", "backup-del")
	manifest := backup.Manifest{ID: "backup-del", RootDir: backupDir, Source: backup.BackupSourceInstall}
	if err := backup.WriteManifest(filepath.Join(backupDir, backup.ManifestFilename), manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsBackupsAction(home, "delete", "backup-del", "", &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsBackupsAction(delete) exit = %d; output:\n%s", exitCode, out.String())
	}

	var resp backupActionResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal delete response: %v\nbody=%s", err, out.String())
	}
	if !resp.Success || resp.ID != "backup-del" {
		t.Fatalf("delete response = %+v, want success=true id=backup-del", resp)
	}

	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		t.Fatalf("backup dir %q still exists after delete (err=%v)", backupDir, err)
	}
}

// TestRunWindowsBackupsAction_Rename asserts "rename --id <known> --description"
// updates the manifest description via RenameBackup and returns success:true.
func TestRunWindowsBackupsAction_Rename(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, ".active-stack", "backups", "uninstall", "backup-ren")
	manifest := backup.Manifest{ID: "backup-ren", RootDir: backupDir, Source: backup.BackupSourceUpgrade, Description: "old description"}
	manifestPath := filepath.Join(backupDir, backup.ManifestFilename)
	if err := backup.WriteManifest(manifestPath, manifest); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsBackupsAction(home, "rename", "backup-ren", "before upgrade", &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsBackupsAction(rename) exit = %d; output:\n%s", exitCode, out.String())
	}

	var resp backupActionResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal rename response: %v\nbody=%s", err, out.String())
	}
	if !resp.Success || resp.ID != "backup-ren" {
		t.Fatalf("rename response = %+v, want success=true id=backup-ren", resp)
	}

	updated, err := backup.ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("read updated manifest: %v", err)
	}
	if updated.Description != "before upgrade" {
		t.Fatalf("updated manifest description = %q, want %q", updated.Description, "before upgrade")
	}
}

// TestRunWindowsBackupsAction_UnknownID asserts any action with an unknown
// --id returns success:false with a descriptive message and non-zero exit.
func TestRunWindowsBackupsAction_UnknownID(t *testing.T) {
	home := t.TempDir()

	for _, action := range []string{"restore", "delete", "rename"} {
		t.Run(action, func(t *testing.T) {
			var out bytes.Buffer
			exitCode := headless.RunWindowsBackupsAction(home, action, "does-not-exist", "desc", &out)
			if exitCode == 0 {
				t.Fatalf("%s with unknown id: expected non-zero exit; output:\n%s", action, out.String())
			}

			var resp backupActionResponse
			if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal %s response: %v\nbody=%s", action, err, out.String())
			}
			if resp.Success || resp.ID != "does-not-exist" || resp.Message == "" {
				t.Fatalf("%s response = %+v, want success=false id=does-not-exist with a message", action, resp)
			}
		})
	}
}
