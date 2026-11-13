package backup

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPersistentRestoreSurvivesManifestReload(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(config, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := NewSnapshotter().Create(filepath.Join(dir, "backup"), []string{config})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte("installed"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest.GuardPersistentRestore(config, sha256.Sum256([]byte("installed")))
	path := filepath.Join(manifest.RootDir, ManifestFilename)
	if err := WriteManifest(path, manifest); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := (RestoreService{}).Restore(loaded); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(config)
	if err != nil || string(raw) != "original" {
		t.Fatalf("restored %q, error %v", raw, err)
	}
}

func TestPersistentRestorePreservesLaterChanges(t *testing.T) {
	for _, change := range []string{"edit", "delete", "symlink"} {
		t.Run(change, func(t *testing.T) {
			dir := t.TempDir()
			config := filepath.Join(dir, "config")
			snapshot := filepath.Join(dir, "snapshot")
			if err := os.WriteFile(snapshot, []byte("original"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config, []byte("installed"), 0o600); err != nil {
				t.Fatal(err)
			}
			m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: config, SnapshotPath: snapshot, Existed: true}}}
			m.GuardPersistentRestore(config, sha256.Sum256([]byte("installed")))
			switch change {
			case "edit":
				if err := os.WriteFile(config, []byte("user"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "delete":
				if err := os.Remove(config); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(config); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(snapshot, config); err != nil {
					t.Skip(err)
				}
			}
			if err := (RestoreService{}).Restore(m); err == nil {
				t.Fatal("expected preservation conflict")
			}
			switch change {
			case "edit":
				raw, _ := os.ReadFile(config)
				if string(raw) != "user" {
					t.Fatalf("lost user edit: %q", raw)
				}
			case "delete":
				if _, err := os.Lstat(config); !os.IsNotExist(err) {
					t.Fatal("recreated deleted file")
				}
			case "symlink":
				info, err := os.Lstat(config)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatal("overwrote symlink")
				}
			}
		})
	}
}

func TestPersistentRestorePreservesUnownedDirectoryFiles(t *testing.T) {
	dir := t.TempDir()
	created := filepath.Join(dir, "skill")
	if err := os.Mkdir(created, 0o700); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(created, "managed")
	user := filepath.Join(created, "user")
	for _, path := range []string{managed, user} {
		if err := os.WriteFile(path, []byte("installed"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: created, IsDir: true}, {OriginalPath: managed}}}
	m.GuardPersistentRestore(managed, sha256.Sum256([]byte("installed")))
	if err := (RestoreService{}).Restore(m); err == nil {
		t.Fatal("expected retained directory conflict")
	}
	raw, err := os.ReadFile(user)
	if err != nil || string(raw) != "installed" {
		t.Fatalf("lost unowned file: %q %v", raw, err)
	}
	if _, err := os.Stat(managed); !os.IsNotExist(err) {
		t.Fatal("managed file not removed")
	}
}

func TestPersistentRestoreRejectsUnguardedEntry(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config")
	snapshot := filepath.Join(dir, "snapshot")
	for path, content := range map[string]string{config: "user", snapshot: "original"} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: config, SnapshotPath: snapshot, Existed: true}}}
	m.GuardPersistentRestore(filepath.Join(dir, "other"), sha256.Sum256([]byte("installed")))
	if err := (RestoreService{}).Restore(m); err == nil {
		t.Fatal("expected incomplete guard error")
	}
	raw, _ := os.ReadFile(config)
	if string(raw) != "user" {
		t.Fatal("unguarded write")
	}
}

func TestPersistentRestoreRetryAndCreatedFile(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config")
	created := filepath.Join(dir, "new")
	if err := os.WriteFile(config, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, err := NewSnapshotter().Create(filepath.Join(dir, "backup"), []string{config, created})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{config, created} {
		if err := os.WriteFile(path, []byte("installed"), 0o600); err != nil {
			t.Fatal(err)
		}
		m.GuardPersistentRestore(path, sha256.Sum256([]byte("installed")))
	}
	manifestPath := filepath.Join(m.RootDir, ManifestFilename)
	if err := WriteManifest(manifestPath, m); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		loaded, err := ReadManifest(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := (RestoreService{}).Restore(loaded); err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
	}
	raw, err := os.ReadFile(config)
	if err != nil || string(raw) != "original" {
		t.Fatalf("restored %q %v", raw, err)
	}
	if _, err := os.Stat(created); !os.IsNotExist(err) {
		t.Fatal("created file still exists")
	}
}

func TestPersistentRestoreIncompleteVersionedManifest(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config")
	if err := os.WriteFile(config, []byte("user"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{RootDir: dir, RestoreGuardVersion: 1, Entries: []ManifestEntry{{OriginalPath: config}}}
	path := filepath.Join(dir, ManifestFilename)
	if err := WriteManifest(path, m); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := (RestoreService{}).Restore(loaded); err == nil {
		t.Fatal("expected missing guard conflict")
	}
	if raw, _ := os.ReadFile(config); string(raw) != "user" {
		t.Fatal("unguarded file removed")
	}
}

func TestPersistentRestoreCleanCreatedDirectories(t *testing.T) {
	dir := t.TempDir()
	created := filepath.Join(dir, "skill")
	nested := filepath.Join(created, "modules")
	managed := filepath.Join(nested, "managed")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managed, []byte("installed"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: created, IsDir: true}, {OriginalPath: nested, IsDir: true}, {OriginalPath: managed}}}
	m.GuardPersistentRestore(managed, sha256.Sum256([]byte("installed")))
	if err := (RestoreService{}).Restore(m); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(created); !os.IsNotExist(err) {
		t.Fatal("empty created directory not removed")
	}
}

func TestPersistentRestorePreservesReplacedDirectory(t *testing.T) {
	for _, kind := range []string{"file", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			created := filepath.Join(dir, "created")
			marker := filepath.Join(dir, "marker")
			if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			if kind == "file" {
				if err := os.WriteFile(created, []byte("user"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(marker, created); err != nil {
				t.Skip(err)
			}
			m := Manifest{RootDir: dir, RestoreGuardVersion: 1, Entries: []ManifestEntry{{OriginalPath: created, IsDir: true}}, RestoreGuards: map[string][32]byte{}}
			if err := (RestoreService{}).Restore(m); err == nil {
				t.Fatal("expected directory replacement conflict")
			}
			if _, err := os.Lstat(created); err != nil {
				t.Fatal("replacement removed", err)
			}
		})
	}
}

func TestPersistentRestoreRejectsUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config")
	if err := os.WriteFile(config, []byte("installed"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: config}}}
	m.GuardPersistentRestore(config, sha256.Sum256([]byte("installed")))
	m.RestoreGuardVersion = 2
	if err := (RestoreService{}).Restore(m); err == nil {
		t.Fatal("expected unsupported version")
	}
	if _, err := os.Stat(config); err != nil {
		t.Fatal("file removed before validation")
	}
}

func TestPersistentRestorePreservesChildThroughReplacedDirectory(t *testing.T) {
	for _, existed := range []bool{false, true} {
		t.Run(map[bool]string{false: "created", true: "preexisting"}[existed], func(t *testing.T) {
			dir := t.TempDir()
			foreign := t.TempDir()
			parent := filepath.Join(dir, "skill")
			child := filepath.Join(parent, "managed")
			target := filepath.Join(foreign, "managed")
			if err := os.WriteFile(target, []byte("installed"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(foreign, parent); err != nil {
				t.Skip(err)
			}
			m := Manifest{RootDir: dir, Entries: []ManifestEntry{{OriginalPath: parent, IsDir: true, Existed: existed}, {OriginalPath: child}}}
			m.GuardPersistentRestore(child, sha256.Sum256([]byte("installed")))
			if err := (RestoreService{}).Restore(m); err == nil {
				t.Fatal("expected replaced parent conflict")
			}
			raw, err := os.ReadFile(target)
			if err != nil || string(raw) != "installed" {
				t.Fatalf("foreign child modified: %q %v", raw, err)
			}
		})
	}
}

func TestRollbackExpectedAbsent(t *testing.T) {
	for _, recreated := range []bool{false, true} {
		t.Run(fmt.Sprint(recreated), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "file")
			os.WriteFile(path, []byte("before"), 0600)
			m, err := NewSnapshotter().Create(filepath.Join(root, "snapshot"), []string{path})
			if err != nil {
				t.Fatal(err)
			}
			os.Remove(path)
			m.GuardRollbackAbsent(path)
			if recreated {
				os.WriteFile(path, []byte("new user file"), 0600)
			}
			err = (RestoreService{}).Restore(m)
			raw, _ := os.ReadFile(path)
			if recreated {
				if err == nil || string(raw) != "new user file" {
					t.Fatalf("recreated raw=%s err=%v", raw, err)
				}
			} else if err != nil || string(raw) != "before" {
				t.Fatalf("restored raw=%s err=%v", raw, err)
			}
		})
	}
}

func TestGuardedRestorePreservesParentPermissions(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		t.Run(fmt.Sprint(persistent), func(t *testing.T) {
			home := t.TempDir()
			parent := filepath.Join(home, "private")
			os.Mkdir(parent, 0700)
			path := filepath.Join(parent, "config")
			os.WriteFile(path, []byte("before"), 0600)
			m, e := NewSnapshotter().Create(filepath.Join(home, "backup"), []string{path})
			if e != nil {
				t.Fatal(e)
			}
			os.WriteFile(path, []byte("installed"), 0600)
			if persistent {
				m.GuardPersistentRestore(path, sha256.Sum256([]byte("installed")))
			} else {
				m.GuardRollback(path, sha256.Sum256([]byte("installed")))
			}
			if e = (RestoreService{}).Restore(m); e != nil {
				t.Fatal(e)
			}
			info, _ := os.Stat(parent)
			if info.Mode().Perm() != 0700 {
				t.Fatalf("private directory permissions widened to %o", info.Mode().Perm())
			}
		})
	}
}
