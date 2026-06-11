package install_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
)

// ─────────────────────────────────────────────────────────────────
// 2.4 / 2.5 — selfInstallStep.Run() happy path
// ─────────────────────────────────────────────────────────────────

// TestSelfInstallStep_HappyPath verifies that Run copies the binary to the
// injected bin dir and calls the PATH registration function.
func TestSelfInstallStep_HappyPath(t *testing.T) {
	// Prepare a fake source binary.
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "active-stack-src")
	if err := os.WriteFile(src, []byte("binary-v1"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Inject a temp bin dir.
	binDir := t.TempDir()

	// Track which dir was passed to AddToUserPath.
	var calledWithDir string
	restoreExec := install.SetExecPathFn(func() (string, error) { return src, nil })
	defer restoreExec()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()
	restorePath := install.SetAddToUserPathFn(func(dir string) error {
		calledWithDir = dir
		return nil
	})
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Binary must be present in binDir.
	target := filepath.Join(binDir, "active-stack")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "binary-v1" {
		t.Errorf("target content = %q, want %q", got, "binary-v1")
	}

	// AddToUserPath must have been called with binDir.
	if calledWithDir != binDir {
		t.Errorf("AddToUserPath called with %q, want %q", calledWithDir, binDir)
	}
}

// TestSelfInstallStep_WindowsExeName verifies the target filename is active-stack.exe on Windows.
func TestSelfInstallStep_WindowsExeName(t *testing.T) {
	src := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(src, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}
	binDir := t.TempDir()

	restoreExec := install.SetExecPathFn(func() (string, error) { return src, nil })
	defer restoreExec()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()
	restorePath := install.SetAddToUserPathFn(func(_ string) error { return nil })
	defer restorePath()

	step := install.NewSelfInstallStep("windows", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	target := filepath.Join(binDir, "active-stack.exe")
	if _, err := os.Stat(target); err != nil {
		t.Errorf("expected active-stack.exe in bin dir, got err: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────
// 2.6 — Graceful degradation
// ─────────────────────────────────────────────────────────────────

// TestSelfInstallStep_ExecPathError verifies that when execPathFn errors,
// Run returns nil (warning emitted, install not aborted).
func TestSelfInstallStep_ExecPathError(t *testing.T) {
	restoreExec := install.SetExecPathFn(func() (string, error) {
		return "", errors.New("cannot resolve executable")
	})
	defer restoreExec()

	var pathCalled bool
	restorePath := install.SetAddToUserPathFn(func(_ string) error {
		pathCalled = true
		return nil
	})
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() must return nil on exec error, got: %v", err)
	}
	if pathCalled {
		t.Error("AddToUserPath must NOT be called when exec resolution fails")
	}
}

// TestSelfInstallStep_CopyError verifies that when the copy fails (source does
// not exist), Run returns nil and PATH fn was not called.
func TestSelfInstallStep_CopyError(t *testing.T) {
	restoreExec := install.SetExecPathFn(func() (string, error) {
		return "/nonexistent/binary-does-not-exist", nil
	})
	defer restoreExec()

	binDir := t.TempDir()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()

	var pathCalled bool
	restorePath := install.SetAddToUserPathFn(func(_ string) error {
		pathCalled = true
		return nil
	})
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() must return nil on copy error, got: %v", err)
	}
	if pathCalled {
		t.Error("AddToUserPath must NOT be called when copy fails")
	}
}

// ─────────────────────────────────────────────────────────────────
// 2.7 — Idempotency: re-run twice → target replaced, no error
// ─────────────────────────────────────────────────────────────────

// TestSelfInstallStep_Idempotent verifies re-running twice succeeds and updates the target.
func TestSelfInstallStep_Idempotent(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "binary")
	if err := os.WriteFile(src, []byte("v2"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	binDir := t.TempDir()
	// Pre-populate the target to simulate a prior install.
	target := filepath.Join(binDir, "active-stack")
	if err := os.WriteFile(target, []byte("v1"), 0o755); err != nil {
		t.Fatalf("write prior target: %v", err)
	}

	restoreExec := install.SetExecPathFn(func() (string, error) { return src, nil })
	defer restoreExec()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()
	pathCallCount := 0
	restorePath := install.SetAddToUserPathFn(func(_ string) error {
		pathCallCount++
		return nil
	})
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)

	// First run.
	if err := step.Run(); err != nil {
		t.Fatalf("first Run() error: %v", err)
	}
	// Second run.
	if err := step.Run(); err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "v2" {
		t.Errorf("target content = %q, want %q", got, "v2")
	}
	// PATH fn was invoked on both runs (dedup is system.AddToUserPath's job).
	if pathCallCount != 2 {
		t.Errorf("AddToUserPath called %d times, want 2", pathCallCount)
	}
}
