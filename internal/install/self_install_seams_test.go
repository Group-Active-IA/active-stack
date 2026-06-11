package install_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
)

// TestSelfInstallSeams_ExecPathFnOverride verifies that SetExecPathFn is used
// by the step instead of the real os.Executable.
func TestSelfInstallSeams_ExecPathFnOverride(t *testing.T) {
	src := filepath.Join(t.TempDir(), "injected-binary")
	if err := os.WriteFile(src, []byte("injected"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	binDir := t.TempDir()
	restoreExec := install.SetExecPathFn(func() (string, error) { return src, nil })
	defer restoreExec()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()
	restorePath := install.SetAddToUserPathFn(func(_ string) error { return nil })
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	target := filepath.Join(binDir, "active-stack")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "injected" {
		t.Errorf("target content = %q, want %q (injected execPathFn not used)", got, "injected")
	}
}

// TestSelfInstallSeams_AddToUserPathFnOverride verifies that SetAddToUserPathFn
// replaces the real system.AddToUserPath call.
func TestSelfInstallSeams_AddToUserPathFnOverride(t *testing.T) {
	src := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(src, []byte("data"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	binDir := t.TempDir()
	restoreExec := install.SetExecPathFn(func() (string, error) { return src, nil })
	defer restoreExec()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()

	var capturedDir string
	restorePath := install.SetAddToUserPathFn(func(dir string) error {
		capturedDir = dir
		return nil
	})
	defer restorePath()

	step := install.NewSelfInstallStep("linux", "", nil)
	if err := step.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if capturedDir != binDir {
		t.Errorf("injected AddToUserPathFn called with %q, want %q", capturedDir, binDir)
	}
}
