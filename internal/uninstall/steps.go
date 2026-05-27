package uninstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ─────────────────────────────────────────────────────────────────
// snapshotStep — Prepare stage; creates the uninstall-time backup snapshot.
// ─────────────────────────────────────────────────────────────────

type snapshotStep struct {
	id          string
	paths       []string
	snapDir     string
	snapCreate  func(dir string, paths []string) (backup.Manifest, error)
	manifestOut *backup.Manifest // written after Run so Apply steps can rollback
}

func (s *snapshotStep) ID() string { return s.id }

func (s *snapshotStep) Run() error {
	manifest, err := s.snapCreate(s.snapDir, s.paths)
	if err != nil {
		return fmt.Errorf("uninstall snapshot: %w", err)
	}
	if s.manifestOut != nil {
		*s.manifestOut = manifest
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────
// markerRemovalStep — config harnesses: removes the marker section.
// ─────────────────────────────────────────────────────────────────

// markerRemovalFn is the backing function for reading a file and calling
// InjectMarkdownSection with empty content (idempotent removal).
// It is a package-level variable so tests can inject a fake.
var markerRemovalFn = func(path, sectionID string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist — no marker to remove; treat as no-op.
			return nil
		}
		return fmt.Errorf("read instructions file %q: %w", path, err)
	}
	updated := filemerge.InjectMarkdownSection(string(existing), sectionID, "")
	if updated == string(existing) {
		// No change — section was absent; no-op.
		return nil
	}
	_, err = filemerge.WriteFileAtomic(path, []byte(updated), 0o644)
	return err
}

type markerRemovalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *markerRemovalStep) ID() string { return "marker:" + s.h.ID }
func (s *markerRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *markerRemovalStep) Run() error {
	for _, a := range s.adapters {
		path := a.InstructionsPath(s.homeDir)
		if err := markerRemovalFn(path, s.h.ID); err != nil {
			return fmt.Errorf("marker removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
	}
	return nil
}

func (s *markerRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// skillRemovalStep — skill harnesses: removes the skill directory.
// ─────────────────────────────────────────────────────────────────

// skillRemovalFn is the backing function for removing a skill directory.
// It is a package-level variable so tests can inject a fake.
var skillRemovalFn = func(path string) error {
	return os.RemoveAll(path)
}

type skillRemovalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *skillRemovalStep) ID() string { return "skill-removal:" + s.h.ID }
func (s *skillRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *skillRemovalStep) Run() error {
	for _, a := range s.adapters {
		skillPath := filepath.Join(a.SkillsDir(s.homeDir), s.h.ID)
		if err := skillRemovalFn(skillPath); err != nil {
			return fmt.Errorf("skill removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
	}
	return nil
}

func (s *skillRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// permissionsRemovalStep — permissions harness: restore from snapshot.
// ─────────────────────────────────────────────────────────────────

type permissionsRemovalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *permissionsRemovalStep) ID() string { return "permissions-removal:" + s.h.ID }
func (s *permissionsRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *permissionsRemovalStep) Run() error {
	// The permissions harness writes JSON settings files.
	// The cleanest reversal is restore-to-snapshot (D3 from design).
	// The uninstall-time snapshot captured the current (installed) settings;
	// restoring it rolls back the uninstall if it fails mid-way.
	// If an install-time backup is provided (StrategyRestore), that is used
	// by the restoreStep instead — permissionsRemovalStep handles the targeted case.
	if s.manifest == nil {
		// No snapshot yet (shouldn't happen in correct usage); no-op to stay safe.
		return nil
	}
	return restoreFn(*s.manifest)
}

func (s *permissionsRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// restoreStep — StrategyRestore: replays the install-time backup.
// ─────────────────────────────────────────────────────────────────

type restoreStep struct {
	installManifest backup.Manifest
}

func (s *restoreStep) ID() string { return "restore-from-backup" }

func (s *restoreStep) Run() error {
	return restoreFn(s.installManifest)
}

// restoreStep intentionally does not implement RollbackStep: StrategyRestore
// has no Prepare snapshot to roll back to (the install-time manifest IS the
// recovery point), so there is nothing to revert.

// ─────────────────────────────────────────────────────────────────
// externalSkipStep — external harnesses: recorded no-op.
// ─────────────────────────────────────────────────────────────────

type externalSkipStep struct {
	h model.Harness
}

func (s *externalSkipStep) ID() string  { return "external-skip:" + s.h.ID }
func (s *externalSkipStep) Run() error  { return nil }

// ─────────────────────────────────────────────────────────────────
// Package-level function vars (testseams; real implementations below)
// ─────────────────────────────────────────────────────────────────

// snapshotCreate is the backing function for creating snapshots. It is a
// package-level variable so tests can inject a fake without reopening backup.
var snapshotCreate = func(snapshotDir string, paths []string) (backup.Manifest, error) {
	return backup.NewSnapshotter().Create(snapshotDir, paths)
}

// restoreFn is the backing function for restoring from a snapshot.
// It is a package-level variable so tests can inject a fake.
var restoreFn = func(m backup.Manifest) error {
	return backup.RestoreService{}.Restore(m)
}

// manifestReceiver is implemented by steps that need the snapshot manifest
// for rollback. The manifest is set after Prepare runs.
type manifestReceiver interface {
	setManifest(m *backup.Manifest)
}
