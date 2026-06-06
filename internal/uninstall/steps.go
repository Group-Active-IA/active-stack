package uninstall

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/command"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
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

// stalePurgeFn removes every jr-stack-marked section the current installer no
// longer owns (legacy/renamed sections from older layouts: persona,
// engram-protocol, strict-tdd-mode, …). It mirrors the install-time cleanup so
// uninstall leaves no orphaned blocks behind, and reuses config.PurgeStaleSections
// as the single source of truth for the owned/stale policy.
//
// It is a package-level variable so tests can swap it out.
var stalePurgeFn = func(path string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist — nothing to purge; no-op.
			return nil
		}
		return fmt.Errorf("read instructions file %q: %w", path, err)
	}
	updated := config.PurgeStaleSections(string(existing))
	if updated == string(existing) {
		// No stale sections present; no-op.
		return nil
	}
	_, err = filemerge.WriteFileAtomic(path, []byte(updated), 0o644)
	return err
}

// primaryAgentRemovalFn removes the agent.<id> entry from a settings JSON file
// (e.g. opencode.json), preserving all other user config. It is the mirror of
// the install-time primary-agent registration. No-op when the file is missing,
// malformed, or already has no such entry. It is a package-level variable so
// tests can inject a fake.
var primaryAgentRemovalFn = func(settingsPath, agentID string) error {
	if settingsPath == "" {
		return nil
	}
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read settings file %q: %w", settingsPath, err)
	}

	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		// Malformed settings — leave it untouched rather than risk clobbering
		// the user's file. Restore-from-snapshot remains the safety net.
		return nil
	}

	agentSection, ok := root["agent"].(map[string]any)
	if !ok {
		return nil // nothing to remove
	}
	if _, present := agentSection[agentID]; !present {
		return nil // no-op
	}
	delete(agentSection, agentID)
	// Drop the agent object entirely if it is now empty, leaving no footprint.
	if len(agentSection) == 0 {
		delete(root, "agent")
	}

	encoded, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings %q: %w", settingsPath, err)
	}
	encoded = append(encoded, '\n')
	if _, err := filemerge.WriteFileAtomic(settingsPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write settings %q: %w", settingsPath, err)
	}
	return nil
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
		// Remove this harness's own section first…
		if err := markerRemovalFn(path, s.h.ID); err != nil {
			return fmt.Errorf("marker removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
		// …then purge any legacy jr-stack sections from older layouts so the
		// uninstall leaves no orphaned blocks behind.
		if err := stalePurgeFn(path); err != nil {
			return fmt.Errorf("stale section purge for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
		// Primary-agent delivery lives in the settings JSON, not the markdown
		// instructions file — remove it there too.
		if a.ConfigDelivery() == model.ConfigDeliveryPrimaryAgent {
			if err := primaryAgentRemovalFn(a.SettingsPath(s.homeDir), s.h.ID); err != nil {
				return fmt.Errorf("primary-agent removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
			}
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
// commandRemovalStep — command harnesses: removes the per-agent command file.
// ─────────────────────────────────────────────────────────────────

// commandRemovalFn is the backing function for removing a single command file.
// It is a package-level variable so tests can inject a fake.
var commandRemovalFn = func(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // already absent — idempotent no-op
	}
	return err
}

type commandRemovalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *commandRemovalStep) ID() string { return "command-removal:" + s.h.ID }
func (s *commandRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *commandRemovalStep) Run() error {
	for _, a := range s.adapters {
		commandsDir := a.CommandsDir(s.homeDir)
		if commandsDir == "" {
			// Adapter does not support commands; skip silently.
			continue
		}
		relPath := command.RelPathForVariant(a.VariantKey())
		if relPath == "" {
			// Unknown variant — no command file to remove; skip silently.
			continue
		}
		commandPath := filepath.Join(commandsDir, relPath)
		if err := commandRemovalFn(commandPath); err != nil {
			return fmt.Errorf("command removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
	}
	return nil
}

func (s *commandRemovalStep) Rollback() error {
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
