package install_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
)

// A `go test` binary must never be copied over the installed active-stack.
func TestSelfInstallSkipsTestBinary(t *testing.T) {
	dir := t.TempDir()
	defer install.SetSelfInstallBinaryInstallDirFn(func(string) string { return dir })()
	src := filepath.Join(t.TempDir(), "install.test.exe")
	if err := os.WriteFile(src, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	defer install.SetExecPathFn(func() (string, error) { return src, nil })()
	step := install.NewSelfInstallStep("", "", nil)
	if err := step.Run(); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("a test binary was installed: %v", entries)
	}
}
