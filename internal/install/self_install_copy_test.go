package install_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
)

// TestCopyBinaryRenameReplace_PreexistingTarget verifies that install.CopyBinaryRenameReplace
// overwrites a pre-existing target with the source bytes and leaves it executable.
func TestCopyBinaryRenameReplace_PreexistingTarget(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src-binary")
	if err := os.WriteFile(src, []byte("new-binary-content"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dstDir := t.TempDir()
	dst := filepath.Join(dstDir, "active-stack")
	// Pre-existing target with different content.
	if err := os.WriteFile(dst, []byte("old-binary-content"), 0o755); err != nil {
		t.Fatalf("write dst: %v", err)
	}

	if err := install.CopyBinaryRenameReplace(src, dst); err != nil {
		t.Fatalf("install.CopyBinaryRenameReplace error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "new-binary-content" {
		t.Errorf("dst content = %q, want %q", got, "new-binary-content")
	}

	// On Unix, verify the file is executable.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat dst: %v", err)
		}
		if info.Mode()&0o111 == 0 {
			t.Errorf("expected executable bit set, got mode %v", info.Mode())
		}
	}
}

// TestCopyBinaryRenameReplace_FreshInstall verifies that install.CopyBinaryRenameReplace
// works when the target does NOT pre-exist.
func TestCopyBinaryRenameReplace_FreshInstall(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src-binary")
	if err := os.WriteFile(src, []byte("fresh-binary"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dst := filepath.Join(t.TempDir(), "active-stack")
	// dst does NOT exist.

	if err := install.CopyBinaryRenameReplace(src, dst); err != nil {
		t.Fatalf("install.CopyBinaryRenameReplace error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "fresh-binary" {
		t.Errorf("dst content = %q, want %q", got, "fresh-binary")
	}
}

// TestCopyBinaryRenameReplace_OldRemovalNoOp verifies that if the .old file
// cannot be removed (simulated by already being absent), no error occurs.
func TestCopyBinaryRenameReplace_OldRemovalNoOp(t *testing.T) {
	// Build: src exists, dst exists. After rename, the .old is removed. Test
	// that if .old removal silently fails, install.CopyBinaryRenameReplace still returns nil.
	// This test achieves that by running the same scenario twice: on the second
	// run the .old from the first run is absent, simulating a no-op remove.
	src := filepath.Join(t.TempDir(), "src")
	if err := os.WriteFile(src, []byte("v2"), 0o755); err != nil {
		t.Fatalf("write src: %v", err)
	}

	dstDir := t.TempDir()
	dst := filepath.Join(dstDir, "active-stack")
	if err := os.WriteFile(dst, []byte("v1"), 0o755); err != nil {
		t.Fatalf("write dst: %v", err)
	}

	// First run: succeeds, .old is cleaned up.
	if err := install.CopyBinaryRenameReplace(src, dst); err != nil {
		t.Fatalf("first run error: %v", err)
	}
	// Second run: target exists again, .old absent → no-op remove on .old.
	if err := os.WriteFile(dst, []byte("v2"), 0o755); err != nil {
		t.Fatalf("write dst for second run: %v", err)
	}
	if err := install.CopyBinaryRenameReplace(src, dst); err != nil {
		t.Fatalf("second run error: %v", err)
	}
}
