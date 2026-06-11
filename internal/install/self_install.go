package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// ─────────────────────────────────────────────────────────────────
// Package-level injectable backing functions for selfInstallStep.
// Override via testseams.go SetXxxFn helpers in tests.
// ─────────────────────────────────────────────────────────────────

// execPathFn resolves the path of the currently running binary.
// Default: os.Executable (returns the real path).
var execPathFn func() (string, error) = os.Executable

// selfInstallBinaryInstallDirFn resolves the bin dir for self-install.
// Default: system.BinaryInstallDir. Separate from homebrew's package-level var.
var selfInstallBinaryInstallDirFn = system.BinaryInstallDir

// selfInstallAddToUserPathFn persists the bin dir to the user PATH.
var selfInstallAddToUserPathFn func(dir string) error = system.AddToUserPath

// ─────────────────────────────────────────────────────────────────
// selfInstallStep — copies the running binary into the PATH bin dir.
// Governance ALTO: uses snapshot + rollback.
// ─────────────────────────────────────────────────────────────────

type selfInstallStep struct {
	// sourcePath is an optional override for the source binary path.
	// When empty, the step resolves via execPathFn.
	sourcePath string
	// goos is the target OS (injected for cross-platform tests).
	// When empty, defaults to runtime.GOOS.
	goos       string
	// manifest is the snapshot manifest pointer set by setManifest.
	manifest   *backup.Manifest
	// onProgress receives warning events on graceful degradation.
	onProgress pipeline.ProgressFunc
}

// NewSelfInstallStep constructs a selfInstallStep.
// goos is the target OS (empty → runtime.GOOS at Run time).
// sourcePath is an optional override for the source binary (empty → os.Executable).
// onProgress receives warning events (nil → stderr fallback).
func NewSelfInstallStep(goos, sourcePath string, onProgress pipeline.ProgressFunc) *selfInstallStep {
	return &selfInstallStep{
		goos:       goos,
		sourcePath: sourcePath,
		onProgress: onProgress,
	}
}

func (s *selfInstallStep) ID() string { return "self-install" }

func (s *selfInstallStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *selfInstallStep) Run() error {
	goos := s.goos
	if goos == "" {
		goos = runtime.GOOS
	}

	// 1. Resolve source binary path.
	src := s.sourcePath
	if src == "" {
		var err error
		src, err = execPathFn()
		if err != nil {
			s.warn(fmt.Sprintf("self-install: resolve executable path: %v (skipping)", err))
			return nil // graceful degradation
		}
	}

	// 2. Resolve target directory and filename.
	dir := selfInstallBinaryInstallDirFn(goos)
	filename := binaryFilename(goos)
	target := filepath.Join(dir, filename)

	// 3. Ensure the target directory exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		s.warn(fmt.Sprintf("self-install: create bin dir %q: %v (skipping)", dir, err))
		return nil // graceful degradation
	}

	// 4. Copy via rename-then-replace (Windows exe-in-use safe).
	if err := copyBinaryRenameReplace(src, target); err != nil {
		s.warn(fmt.Sprintf("self-install: copy binary to %q: %v (skipping)", target, err))
		return nil // graceful degradation
	}

	// 5. Register the bin dir on the user PATH.
	if err := selfInstallAddToUserPathFn(dir); err != nil {
		s.warn(fmt.Sprintf("self-install: add %q to PATH: %v (skipping)", dir, err))
		return nil // graceful degradation
	}

	return nil
}

func (s *selfInstallStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// warn emits a warning via the progress callback or to stderr as fallback.
func (s *selfInstallStep) warn(msg string) {
	if s.onProgress != nil {
		s.onProgress(pipeline.ProgressEvent{
			StepID: s.ID(),
			Stage:  pipeline.StageApply,
			Status: pipeline.StepStatusFailed,
			Err:    fmt.Errorf("%s", msg),
		})
	} else {
		fmt.Fprintln(os.Stderr, "[WARN]", msg)
	}
}

// ─────────────────────────────────────────────────────────────────
// binaryFilename returns the platform-specific binary filename.
// ─────────────────────────────────────────────────────────────────

func binaryFilename(goos string) string {
	if goos == "windows" {
		return "active-stack.exe"
	}
	return "active-stack"
}

// ─────────────────────────────────────────────────────────────────
// copyBinaryRenameReplace copies src to dst using rename-then-replace
// so a locked (in-use) Windows target never aborts the copy.
//
// Strategy (D4):
//  1. Write src to dst+".new".
//  2. Rename existing dst aside to dst+".old" (best-effort: ok if absent).
//  3. Rename dst+".new" → dst.
//  4. Best-effort remove dst+".old" (ignore failure; harmless leftover).
//
// On Unix, rename is atomic and an in-use binary can be replaced.
// The exported name allows the _test package to call it directly.
// ─────────────────────────────────────────────────────────────────

func copyBinaryRenameReplace(src, dst string) error {
	tmpPath := dst + ".new"
	oldPath := dst + ".old"

	// Write to temp file in target dir.
	if err := copyFile(src, tmpPath); err != nil {
		return fmt.Errorf("write %q: %w", tmpPath, err)
	}

	// Move existing target aside (best-effort — ignore "not found").
	_ = os.Rename(dst, oldPath)

	// Rename the new file into place.
	if err := os.Rename(tmpPath, dst); err != nil {
		// Rollback: try to restore the old target.
		_ = os.Rename(oldPath, dst)
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %q → %q: %w", tmpPath, dst, err)
	}

	// Best-effort remove the old target.
	_ = os.Remove(oldPath)

	return nil
}

// copyFile copies the file at src to dst, creating or truncating dst.
// On Unix it sets the executable bit (0o755).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create dst %q: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	return nil
}
