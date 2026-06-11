package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestoreRestoresExistingAndRemovesCreated(t *testing.T) {
	home := t.TempDir()

	originalPath := filepath.Join(home, "config", "settings.json")
	if err := os.MkdirAll(filepath.Dir(originalPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(originalPath, []byte("new\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	removedPath := filepath.Join(home, "config", "extra.json")
	if err := os.WriteFile(removedPath, []byte("temporary\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() removed path error = %v", err)
	}

	snapshotPath := filepath.Join(home, "backup", "files", "settings.json")
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() snapshot error = %v", err)
	}
	if err := os.WriteFile(snapshotPath, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() snapshot error = %v", err)
	}

	manifest := Manifest{
		Entries: []ManifestEntry{
			{OriginalPath: originalPath, SnapshotPath: snapshotPath, Existed: true, Mode: 0o600},
			{OriginalPath: removedPath, Existed: false},
		},
	}

	service := RestoreService{}
	if err := service.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	restored, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("ReadFile() restored path error = %v", err)
	}
	if string(restored) != "old\n" {
		t.Fatalf("restored content = %q", string(restored))
	}

	if _, err := os.Stat(removedPath); !os.IsNotExist(err) {
		t.Fatalf("expected removed path %q to be deleted, err = %v", removedPath, err)
	}
}

func TestRestoreFailsWhenSnapshotMissing(t *testing.T) {
	service := RestoreService{}
	err := service.Restore(Manifest{Entries: []ManifestEntry{{
		OriginalPath: filepath.Join(t.TempDir(), "out.json"),
		SnapshotPath: filepath.Join(t.TempDir(), "missing.json"),
		Existed:      true,
		Mode:         0o644,
	}}})

	if err == nil {
		t.Fatalf("Restore() expected error for missing snapshot")
	}
}

// TestRestoreCompressedBackup verifies that Restore() correctly extracts files
// from a tar.gz archive when manifest.Compressed == true (BKUP-T31).
func TestRestoreCompressedBackup(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// Create a source file to snapshot.
	srcFile := filepath.Join(home, "config", "settings.json")
	if err := os.MkdirAll(filepath.Dir(srcFile), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("original content\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Use Snapshotter to create a compressed backup — this produces snapshot.tar.gz
	// and sets Compressed=true + relative SnapshotPaths in the manifest.
	snapshotter := Snapshotter{now: func() time.Time { return time.Now() }}
	manifest, err := snapshotter.Create(backupDir, []string{srcFile})
	if err != nil {
		t.Fatalf("Snapshotter.Create() error = %v", err)
	}
	if !manifest.Compressed {
		t.Fatalf("expected Compressed=true, got false")
	}

	// Overwrite the source file so we can verify restore brought back the original.
	if err := os.WriteFile(srcFile, []byte("modified content\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() overwrite error = %v", err)
	}

	service := RestoreService{}
	if err := service.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	restored, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("ReadFile() after restore error = %v", err)
	}
	if string(restored) != "original content\n" {
		t.Fatalf("restored content = %q, want %q", string(restored), "original content\n")
	}
}

// TestRestoreUncompressedBackup verifies backward compatibility: old-style backups
// with Compressed==false (plain files on disk) still restore correctly (BKUP-T30).
func TestRestoreUncompressedBackup(t *testing.T) {
	home := t.TempDir()

	originalPath := filepath.Join(home, "config", "app.json")
	if err := os.MkdirAll(filepath.Dir(originalPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(originalPath, []byte("modified\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	snapshotPath := filepath.Join(home, "backup", "files", "app.json")
	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() snapshot dir error = %v", err)
	}
	if err := os.WriteFile(snapshotPath, []byte("original\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() snapshot error = %v", err)
	}

	// Manifest with Compressed=false (zero value) — old-style plain files.
	manifest := Manifest{
		Compressed: false,
		Entries: []ManifestEntry{
			{OriginalPath: originalPath, SnapshotPath: snapshotPath, Existed: true, Mode: 0o600},
		},
	}

	service := RestoreService{}
	if err := service.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	got, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("ReadFile() after restore error = %v", err)
	}
	if string(got) != "original\n" {
		t.Fatalf("restored content = %q, want %q", string(got), "original\n")
	}
}

// TestRestoreCompressedMultipleFiles triangulates the compressed restore path
// with more than one file, ensuring the loop resolves all relative paths correctly.
func TestRestoreCompressedMultipleFiles(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	fileA := filepath.Join(home, "config", "a.json")
	fileB := filepath.Join(home, "config", "b.json")
	if err := os.MkdirAll(filepath.Dir(fileA), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(fileA, []byte("content-a\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() a error = %v", err)
	}
	if err := os.WriteFile(fileB, []byte("content-b\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() b error = %v", err)
	}

	snapshotter := Snapshotter{now: func() time.Time { return time.Now() }}
	manifest, err := snapshotter.Create(backupDir, []string{fileA, fileB})
	if err != nil {
		t.Fatalf("Snapshotter.Create() error = %v", err)
	}

	// Overwrite both files.
	if err := os.WriteFile(fileA, []byte("dirty-a\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() overwrite a error = %v", err)
	}
	if err := os.WriteFile(fileB, []byte("dirty-b\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() overwrite b error = %v", err)
	}

	service := RestoreService{}
	if err := service.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	gotA, err := os.ReadFile(fileA)
	if err != nil {
		t.Fatalf("ReadFile(a) error = %v", err)
	}
	if string(gotA) != "content-a\n" {
		t.Fatalf("fileA restored content = %q, want %q", string(gotA), "content-a\n")
	}

	gotB, err := os.ReadFile(fileB)
	if err != nil {
		t.Fatalf("ReadFile(b) error = %v", err)
	}
	if string(gotB) != "content-b\n" {
		t.Fatalf("fileB restored content = %q, want %q", string(gotB), "content-b\n")
	}
}

// TestRestoreCompressed_MissingArchive verifies that Restore returns an error
// when the manifest has Compressed==true but snapshot.tar.gz does not exist.
func TestRestoreCompressed_MissingArchive(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup-no-archive")
	// Create the backup directory but do NOT create snapshot.tar.gz inside it.
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	manifest := Manifest{
		RootDir:    backupDir,
		Compressed: true,
		Entries: []ManifestEntry{
			{
				OriginalPath: filepath.Join(home, "config", "settings.json"),
				SnapshotPath: "files/config/settings.json",
				Existed:      true,
				Mode:         0o644,
			},
		},
	}

	service := RestoreService{}
	err := service.Restore(manifest)
	if err == nil {
		t.Fatal("Restore() should return error when snapshot.tar.gz is missing")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-17 dir-aware rollback tests (RED before fix)
// ─────────────────────────────────────────────────────────────────────────────

// TestRestore_Security_PreexistingDirSurvivesRollback is the PRIMARY SECURITY test.
// A skills dir that existed BEFORE the install (Existed=true, IsDir=true) must
// NOT be touched by restore. The user's original skill inside it must survive.
//
// With the current code this test FAILS because buildEntry marks the dir as
// Existed=false and restore then calls os.Remove (which may fail or, after a
// naive RemoveAll fix, wipes user skills). After the fix the entry carries
// Existed=true+IsDir=true → NO-OP in restore → dir and user skill survive intact.
func TestRestore_Security_PreexistingDirSurvivesRollback(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// Setup: skills dir ALREADY EXISTS before any install.
	skillsDir := filepath.Join(home, ".config", "opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll skillsDir: %v", err)
	}
	// User's pre-existing skill.
	userSkill := filepath.Join(skillsDir, "my-existing-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(userSkill), 0o755); err != nil {
		t.Fatalf("MkdirAll user skill dir: %v", err)
	}
	if err := os.WriteFile(userSkill, []byte("# user skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile user skill: %v", err)
	}

	// Snapshot the skills dir BEFORE the install (as collectWritePaths would do).
	snap := NewSnapshotter()
	manifest, err := snap.Create(backupDir, []string{skillsDir})
	if err != nil {
		t.Fatalf("Snapshotter.Create: %v", err)
	}

	// Simulate what the install does: add another skill inside the dir.
	installSkill := filepath.Join(skillsDir, "active-orchestrator", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(installSkill), 0o755); err != nil {
		t.Fatalf("MkdirAll install skill dir: %v", err)
	}
	if err := os.WriteFile(installSkill, []byte("# installed skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile install skill: %v", err)
	}

	// Rollback: restore the snapshot.
	svc := RestoreService{}
	if err := svc.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// ASSERT: skills dir still exists.
	if _, statErr := os.Stat(skillsDir); statErr != nil {
		t.Errorf("skillsDir was removed by rollback — SECURITY VIOLATION: %v", statErr)
	}

	// ASSERT: user's original skill is still intact.
	content, err := os.ReadFile(userSkill)
	if err != nil {
		t.Errorf("user skill was removed by rollback — SECURITY VIOLATION: %v", err)
	} else if string(content) != "# user skill\n" {
		t.Errorf("user skill content corrupted = %q, want %q", string(content), "# user skill\n")
	}
}

// TestRestore_NewDir_RemovedByRollback verifies that a skills dir that did NOT
// exist before the install (Existed=false, IsDir=true) is removed entirely by
// RemoveAll — including any skills the install deposited inside.
func TestRestore_NewDir_RemovedByRollback(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// The skills dir does NOT exist yet when the snapshot is taken.
	skillsDir := filepath.Join(home, ".config", "opencode", "skills")

	// Use CreateWithDirHints so the missing dir is recorded as IsDir=true.
	// This mirrors what BuildPlan does via collectWritePaths for skill harnesses.
	snap := NewSnapshotter()
	dirHints := map[string]bool{filepath.Clean(skillsDir): true}
	manifest, err := snap.CreateWithDirHints(backupDir, []string{skillsDir}, dirHints)
	if err != nil {
		t.Fatalf("Snapshotter.CreateWithDirHints: %v", err)
	}

	// Simulate install: create the dir and add a skill inside.
	installSkill := filepath.Join(skillsDir, "active-orchestrator", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(installSkill), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(installSkill, []byte("# installed skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Rollback.
	svc := RestoreService{}
	if err := svc.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v (want no error — RemoveAll should succeed)", err)
	}

	// ASSERT: the dir created by the install is gone.
	if _, statErr := os.Stat(skillsDir); !os.IsNotExist(statErr) {
		t.Errorf("skillsDir still exists after rollback — want it removed; stat err = %v", statErr)
	}
}

// TestRestore_BugRegression_NonEmptyDirFails reproduces the original bug:
// a non-empty dir with Existed=false causes os.Remove to fail with
// "directory not empty". After the fix, RemoveAll handles it cleanly.
// CreateWithDirHints is used to record IsDir=true (mimicking BuildPlan behavior).
func TestRestore_BugRegression_NonEmptyDirFails(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// Dir does NOT exist at snapshot time.
	targetDir := filepath.Join(home, "skills")

	snap := NewSnapshotter()
	dirHints := map[string]bool{filepath.Clean(targetDir): true}
	manifest, err := snap.CreateWithDirHints(backupDir, []string{targetDir}, dirHints)
	if err != nil {
		t.Fatalf("Snapshotter.CreateWithDirHints: %v", err)
	}

	// The install creates the dir and a file inside (non-empty).
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "a.md"), []byte("content\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Rollback must NOT return "directory not empty".
	svc := RestoreService{}
	if err := svc.Restore(manifest); err != nil {
		t.Errorf("Restore() error = %v — want no error after fix (RemoveAll handles non-empty dir)", err)
	}

	// The dir should be gone.
	if _, statErr := os.Stat(targetDir); !os.IsNotExist(statErr) {
		t.Errorf("dir still exists after rollback; stat err = %v", statErr)
	}
}

// TestRestore_NewFile_Removed is a regression guard: file entries with
// Existed=false must still be removed via os.Remove (unchanged behavior).
func TestRestore_NewFile_Removed(t *testing.T) {
	home := t.TempDir()

	// A file that didn't exist at snapshot time.
	newFile := filepath.Join(home, "new.json")
	manifest := Manifest{
		Compressed: false,
		Entries: []ManifestEntry{
			{OriginalPath: newFile, Existed: false, IsDir: false},
		},
	}

	// Create the file (simulating what the install wrote).
	if err := os.WriteFile(newFile, []byte("created by install\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	svc := RestoreService{}
	if err := svc.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if _, statErr := os.Stat(newFile); !os.IsNotExist(statErr) {
		t.Errorf("new file still exists after rollback; stat err = %v", statErr)
	}
}

// TestRestore_PreexistingFile_ContentRestored is a regression guard: file entries
// with Existed=true must still have their content restored (unchanged behavior).
func TestRestore_PreexistingFile_ContentRestored(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// A file that existed before the install.
	origFile := filepath.Join(home, "config.json")
	if err := os.WriteFile(origFile, []byte(`{"original":true}`+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile orig: %v", err)
	}

	snap := NewSnapshotter()
	manifest, err := snap.Create(backupDir, []string{origFile})
	if err != nil {
		t.Fatalf("Snapshotter.Create: %v", err)
	}

	// Install overwrites the file.
	if err := os.WriteFile(origFile, []byte(`{"modified":true}`+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile modified: %v", err)
	}

	svc := RestoreService{}
	if err := svc.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	got, err := os.ReadFile(origFile)
	if err != nil {
		t.Fatalf("ReadFile after restore: %v", err)
	}
	if string(got) != `{"original":true}`+"\n" {
		t.Errorf("restored content = %q, want original", string(got))
	}
}

// TestRestoreCompressedRemovesCreatedFiles verifies that entries with Existed=false
// in a compressed backup cause the file at OriginalPath to be deleted (BKUP-T32).
func TestRestoreCompressedRemovesCreatedFiles(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backup")

	// Create a real file to snapshot (so the archive is valid).
	srcFile := filepath.Join(home, "config", "kept.json")
	if err := os.MkdirAll(filepath.Dir(srcFile), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(srcFile, []byte("data\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	snapshotter := Snapshotter{now: func() time.Time { return time.Now() }}
	manifest, err := snapshotter.Create(backupDir, []string{srcFile})
	if err != nil {
		t.Fatalf("Snapshotter.Create() error = %v", err)
	}

	// Add an entry that was NOT in the original snapshot (Existed=false).
	// This simulates a file created AFTER backup — restore should remove it.
	createdFile := filepath.Join(home, "config", "extra.json")
	if err := os.WriteFile(createdFile, []byte("should be removed\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() created file error = %v", err)
	}
	manifest.Entries = append(manifest.Entries, ManifestEntry{
		OriginalPath: createdFile,
		Existed:      false,
	})

	service := RestoreService{}
	if err := service.Restore(manifest); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if _, statErr := os.Stat(createdFile); !os.IsNotExist(statErr) {
		t.Fatalf("expected %q to be removed after restore, got stat err = %v", createdFile, statErr)
	}
}
