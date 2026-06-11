package system

import (
	"os"
	"path/filepath"
)

// binaryInstallDirWritableFn is the backing function for the writable-dir check.
// It is a package-level variable so tests can inject a fake.
var binaryInstallDirWritableFn = isWritableDirSystem

// BinaryInstallDir returns the directory where active-stack binaries should be installed.
// Windows: %LOCALAPPDATA%\active-stack\bin
// Unix: /usr/local/bin when writable, otherwise ~/.local/bin
func BinaryInstallDir(goos string) string {
	return binaryInstallDirWithCandidate(goos, "/usr/local/bin")
}

// binaryInstallDirWithCandidate is the testable core of BinaryInstallDir,
// accepting an injected candidate dir (for tests that can't use the real /usr/local/bin).
func binaryInstallDirWithCandidate(goos, candidate string) string {
	if goos == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "active-stack", "bin")
	}

	if binaryInstallDirWritableFn(candidate) {
		return candidate
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return candidate
	}
	return filepath.Join(home, ".local", "bin")
}

// isWritableDirSystem probes whether a directory is writable by creating a temp file.
func isWritableDirSystem(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	tmp, err := os.CreateTemp(dir, ".active-stack-write-test-*")
	if err != nil {
		return false
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return true
}
