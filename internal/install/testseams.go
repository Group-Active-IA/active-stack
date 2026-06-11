package install

import (
	"context"
	"io/fs"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	extinstaller "github.com/Group-Active-IA/active-stack/internal/harness/external"
	cfginstaller "github.com/Group-Active-IA/active-stack/internal/harness/config"
	perminstaller "github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	skillinstaller "github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// SetSnapshotCreate replaces the snapshotCreate function for testing.
// It also overrides snapshotCreateWithHints (used by BuildPlan) so that
// injected fakes apply uniformly, regardless of whether dir hints are passed.
func SetSnapshotCreate(fn func(dir string, paths []string) (backup.Manifest, error)) (restore func()) {
	oldCreate := snapshotCreate
	oldWithHints := snapshotCreateWithHints
	snapshotCreate = fn
	snapshotCreateWithHints = func(dir string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		return fn(dir, paths)
	}
	return func() {
		snapshotCreate = oldCreate
		snapshotCreateWithHints = oldWithHints
	}
}

// SetSnapshotCreateWithHints replaces the snapshotCreateWithHints function for
// testing. Unlike SetSnapshotCreate, this preserves the dir hints parameter so
// tests can inspect which paths were recorded as directory hints (DirHints).
// This is the preferred seam for tests that exercise DirHint behaviour (e.g.
// skill harnesses must be recorded as DirHints for dir-aware rollback).
func SetSnapshotCreateWithHints(fn func(dir string, paths []string, dirHints map[string]bool) (backup.Manifest, error)) (restore func()) {
	oldCreate := snapshotCreate
	oldWithHints := snapshotCreateWithHints
	snapshotCreateWithHints = fn
	// Also override the non-hints variant so simple Set calls still use the same fn.
	snapshotCreate = func(dir string, paths []string) (backup.Manifest, error) {
		return fn(dir, paths, nil)
	}
	return func() {
		snapshotCreate = oldCreate
		snapshotCreateWithHints = oldWithHints
	}
}

// SetExternalInstallFn replaces externalInstallFn for testing.
func SetExternalInstallFn(fn func(
	ctx context.Context,
	h model.Harness,
	profile system.PlatformProfile,
	adapters []extinstaller.AgentAdapter,
	homeDir string,
) (extinstaller.Result, error)) (restore func()) {
	old := externalInstallFn
	externalInstallFn = fn
	return func() { externalInstallFn = old }
}

// SetSkillInstallFn replaces skillInstallFn for testing.
func SetSkillInstallFn(fn func(
	runner interface{},
	embeddedFS fs.FS,
	ctx context.Context,
	h model.Harness,
	adapters []skillinstaller.AgentAdapter,
	homeDir, backupDir string,
) ([]skillinstaller.Result, error)) (restore func()) {
	old := skillInstallFn
	skillInstallFn = func(
		runner skillRunner,
		embeddedFS fs.FS,
		ctx context.Context,
		h model.Harness,
		adapters []skillinstaller.AgentAdapter,
		homeDir, backupDir string,
	) ([]skillinstaller.Result, error) {
		return fn(runner, embeddedFS, ctx, h, adapters, homeDir, backupDir)
	}
	return func() { skillInstallFn = old }
}

// SetConfigInstallFn replaces configInstallFn for testing.
func SetConfigInstallFn(fn func(h model.Harness, adapters interface{}, homeDir string) error) (restore func()) {
	old := configInstallFn
	configInstallFn = func(h model.Harness, adapters []cfginstaller.AgentAdapter, homeDir string) (cfginstaller.Result, error) {
		return cfginstaller.Result{}, fn(h, adapters, homeDir)
	}
	return func() { configInstallFn = old }
}

// SetPermissionsInstallFn replaces permissionsInstallFn for testing.
func SetPermissionsInstallFn(fn func(homeDir string, adapters []perminstaller.PermissionsAdapter, tier model.PermissionTier) (perminstaller.Result, error)) (restore func()) {
	old := permissionsInstallFn
	permissionsInstallFn = fn
	return func() { permissionsInstallFn = old }
}

// SetRestoreFn replaces the restore function used by rollback steps for testing.
func SetRestoreFn(fn func(m backup.Manifest) error) (restore func()) {
	old := restoreFn
	restoreFn = fn
	return func() { restoreFn = old }
}

// SetMCPWriteFn replaces the mcpWriteFn for testing (C-28).
// Injected fake avoids real filesystem writes during plan-only tests.
func SetMCPWriteFn(fn func(mcp model.MCP, configPath string, strategy model.MCPStrategy) error) (restore func()) {
	old := mcpWriteFn
	mcpWriteFn = fn
	return func() { mcpWriteFn = old }
}

// SetCommandInstallFn replaces commandInstallFn for testing (C-31).
// Injected fake avoids real filesystem writes during pipeline-only tests.
func SetCommandInstallFn(fn func(adapters []AgentAdapter, homeDir, backupDir string) error) (restore func()) {
	old := commandInstallFn
	commandInstallFn = fn
	return func() { commandInstallFn = old }
}

// GetEmbeddedCommandsFS returns the current value of the package-level
// embeddedCommandsFS variable. Used by real-wiring tests (C-32) to assert
// that the binary entry point called WithEmbeddedCommandsFS before dispatch.
// This is the minimal observation point for the D2 wiring test.
func GetEmbeddedCommandsFS() fs.FS {
	return embeddedCommandsFS
}

// ResetEmbeddedCommandsFS sets embeddedCommandsFS back to nil.
// Call this at the start of each real-wiring test to ensure a clean state,
// since the global persists across test runs in the same binary.
func ResetEmbeddedCommandsFS() {
	embeddedCommandsFS = nil
}

// ─────────────────────────────────────────────────────────────────
// selfInstallStep seam setters (D6 — binary-self-install)
// ─────────────────────────────────────────────────────────────────

// SetExecPathFn replaces execPathFn (os.Executable) for testing.
// Returns a restore function; always defer the restore.
func SetExecPathFn(fn func() (string, error)) (restore func()) {
	old := execPathFn
	execPathFn = fn
	return func() { execPathFn = old }
}

// SetSelfInstallBinaryInstallDirFn replaces selfInstallBinaryInstallDirFn for testing.
func SetSelfInstallBinaryInstallDirFn(fn func(goos string) string) (restore func()) {
	old := selfInstallBinaryInstallDirFn
	selfInstallBinaryInstallDirFn = fn
	return func() { selfInstallBinaryInstallDirFn = old }
}

// SetAddToUserPathFn replaces selfInstallAddToUserPathFn for testing.
func SetAddToUserPathFn(fn func(dir string) error) (restore func()) {
	old := selfInstallAddToUserPathFn
	selfInstallAddToUserPathFn = fn
	return func() { selfInstallAddToUserPathFn = old }
}
