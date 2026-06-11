package system

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBinaryInstallDir_Windows verifies Windows returns %LOCALAPPDATA%\active-stack\bin.
func TestBinaryInstallDir_Windows(t *testing.T) {
	localAppData := t.TempDir()
	orig := os.Getenv("LOCALAPPDATA")
	t.Cleanup(func() { os.Setenv("LOCALAPPDATA", orig) })
	os.Setenv("LOCALAPPDATA", localAppData)

	got := BinaryInstallDir("windows")
	want := filepath.Join(localAppData, "active-stack", "bin")
	if got != want {
		t.Errorf("BinaryInstallDir(windows) = %q, want %q", got, want)
	}
}

// TestBinaryInstallDir_Unix_Writable verifies Unix returns /usr/local/bin when writable.
// We inject a temp writable dir as the candidate via the isWritableDir helper.
func TestBinaryInstallDir_Unix_Writable(t *testing.T) {
	// Override the writable-dir check so we can simulate /usr/local/bin being writable
	// without actually probing the real /usr/local/bin (which is not writable in CI).
	writableDir := t.TempDir()
	old := binaryInstallDirWritableFn
	t.Cleanup(func() { binaryInstallDirWritableFn = old })
	binaryInstallDirWritableFn = func(dir string) bool {
		return dir == writableDir
	}

	got := binaryInstallDirWithCandidate("linux", writableDir)
	if got != writableDir {
		t.Errorf("binaryInstallDirWithCandidate(linux, writable) = %q, want %q", got, writableDir)
	}
}

// TestBinaryInstallDir_Unix_Fallback verifies that when /usr/local/bin is not writable,
// the fallback is ~/.local/bin.
func TestBinaryInstallDir_Unix_Fallback(t *testing.T) {
	fakeHome := t.TempDir()
	// os.UserHomeDir() uses USERPROFILE on Windows, HOME on Unix.
	for _, key := range []string{"HOME", "USERPROFILE"} {
		orig := os.Getenv(key)
		t.Cleanup(func() { os.Setenv(key, orig) })
		os.Setenv(key, fakeHome)
	}

	old := binaryInstallDirWritableFn
	t.Cleanup(func() { binaryInstallDirWritableFn = old })
	// Nothing is writable → always fall back.
	binaryInstallDirWritableFn = func(dir string) bool { return false }

	got := BinaryInstallDir("linux")
	want := filepath.Join(fakeHome, ".local", "bin")
	if got != want {
		t.Errorf("BinaryInstallDir(linux, not writable) = %q, want %q", got, want)
	}
}
