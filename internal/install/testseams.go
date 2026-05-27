package install

import (
	"context"
	"io/fs"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	cfginstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
	perminstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config/permissions"
	skillinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// SetSnapshotCreate replaces the snapshotCreate function for testing.
func SetSnapshotCreate(fn func(dir string, paths []string) (backup.Manifest, error)) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
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
func SetPermissionsInstallFn(fn func(homeDir string, adapters []perminstaller.PermissionsAdapter) (perminstaller.Result, error)) (restore func()) {
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
