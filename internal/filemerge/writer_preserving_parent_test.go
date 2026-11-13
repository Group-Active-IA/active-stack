package filemerge

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteFileAtomicPreservingParentKeepsPrivateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits differ on Windows")
	}
	dir := filepath.Join(t.TempDir(), "private")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteFileAtomicPreservingParent(filepath.Join(dir, "config"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("private directory permissions changed: %v, %v", info, err)
	}
}

func TestWriteFileAtomicPreservingParentKeepsGroupDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits differ on Windows")
	}
	dir := filepath.Join(t.TempDir(), "group")
	if err := os.Mkdir(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := WriteFileAtomicPreservingParent(path, []byte("after"), 0o600)
	if err != nil || !result.Changed || result.Created {
		t.Fatalf("replacement result=%+v, err=%v", result, err)
	}
	parent, err := os.Stat(dir)
	if err != nil || parent.Mode().Perm() != 0o750 {
		t.Fatalf("group directory permissions changed: %v, %v", parent, err)
	}
	file, err := os.Stat(path)
	if err != nil || file.Mode().Perm() != 0o600 {
		t.Fatalf("file permissions changed: %v, %v", file, err)
	}
	result, err = WriteFileAtomicPreservingParent(path, []byte("after"), 0o644)
	if err != nil || result.Changed || result.Created {
		t.Fatalf("identical write result=%+v, err=%v", result, err)
	}
	again, err := os.Stat(path)
	if err != nil || again.Mode().Perm() != 0o600 || !again.ModTime().Equal(file.ModTime()) {
		t.Fatalf("identical write changed file metadata: %v, %v", again, err)
	}
}

func TestWriteFileAtomicPreservingParentCreatesPrivateParents(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits differ on Windows")
	}
	base := t.TempDir()
	path := filepath.Join(base, "one", "two", "config")
	result, err := WriteFileAtomicPreservingParent(path, []byte("content"), 0o600)
	if err != nil || !result.Changed || !result.Created {
		t.Fatalf("creation result=%+v, err=%v", result, err)
	}
	for _, dir := range []string{filepath.Join(base, "one"), filepath.Join(base, "one", "two")} {
		info, err := os.Stat(dir)
		if err != nil || info.Mode().Perm()&0o077 != 0 || info.Mode().Perm()&0o700 != 0o700 {
			t.Fatalf("new directory is not private and owner-accessible: %v, %v", info, err)
		}
	}
}
