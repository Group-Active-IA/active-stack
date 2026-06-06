package uninstall

import "github.com/JuanCruzRobledo/jr-stack/internal/backup"

// SetSnapshotCreate replaces the snapshotCreate function for testing.
func SetSnapshotCreate(fn func(dir string, paths []string) (backup.Manifest, error)) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

// SetRestoreFn replaces the restore function used by rollback steps for testing.
func SetRestoreFn(fn func(m backup.Manifest) error) (restore func()) {
	old := restoreFn
	restoreFn = fn
	return func() { restoreFn = old }
}

// SetMarkerRemovalFn replaces the markerRemovalFn for testing.
func SetMarkerRemovalFn(fn func(path, sectionID string) error) (restore func()) {
	old := markerRemovalFn
	markerRemovalFn = fn
	return func() { markerRemovalFn = old }
}

// SetStalePurgeFn replaces the stalePurgeFn for testing.
func SetStalePurgeFn(fn func(path string) error) (restore func()) {
	old := stalePurgeFn
	stalePurgeFn = fn
	return func() { stalePurgeFn = old }
}

// SetSkillRemovalFn replaces the skillRemovalFn for testing.
func SetSkillRemovalFn(fn func(path string) error) (restore func()) {
	old := skillRemovalFn
	skillRemovalFn = fn
	return func() { skillRemovalFn = old }
}

// SetCommandRemovalFn replaces the commandRemovalFn for testing.
func SetCommandRemovalFn(fn func(path string) error) (restore func()) {
	old := commandRemovalFn
	commandRemovalFn = fn
	return func() { commandRemovalFn = old }
}
