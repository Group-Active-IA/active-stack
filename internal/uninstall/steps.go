package uninstall

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/harness/command"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
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

// stalePurgeFn removes every active-stack-marked section the current installer no
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
	h              model.Harness
	adapters       []AgentAdapter
	homeDir        string
	manifest       *backup.Manifest
	nativeRollback *managedRollback
}

func (s *markerRemovalStep) ID() string                     { return "marker:" + s.h.ID }
func (s *markerRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *markerRemovalStep) Run() error {
	for _, a := range s.adapters {
		if s.h.ID == "sdd-orchestrator" && a.ConfigDelivery() == model.ConfigDeliveryManagedSkill {
			if a.Agent() == model.AgentClaude {
				if err := s.removeLayered(a); err != nil {
					return err
				}
				continue
			}
			if a.Agent() != model.AgentCodex {
				return fmt.Errorf("managed skill delivery unsupported for agent %q", a.Agent())
			}
			if s.nativeRollback == nil {
				s.nativeRollback = newManagedRollback()
			}
			paths, err := config.ManagedSkillOwnedPaths(a, s.homeDir, s.h.ID)
			if err != nil {
				return err
			}
			if err := s.nativeRollback.capture(paths, s.homeDir); err != nil {
				return err
			}
			instructions, err := os.ReadFile(a.InstructionsPath(s.homeDir))
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			changed, removeErr := config.RemoveManagedSkill(a, s.homeDir, s.h.ID)
			s.nativeRollback.record(s.manifest, changed, a.InstructionsPath(s.homeDir), string(instructions), s.h.ID)
			if removeErr != nil {
				return fmt.Errorf("managed skill removal for agent %q: %w", a.Agent(), removeErr)
			}
			continue
		}
		path := a.InstructionsPath(s.homeDir)
		// Remove this harness's own section first…
		if err := markerRemovalFn(path, s.h.ID); err != nil {
			return fmt.Errorf("marker removal for harness %q on agent %q: %w", s.h.ID, a.Agent(), err)
		}
		// …then purge any legacy active-stack sections from older layouts so the
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

// removeLayered reverses the layered Claude delivery (Machine target): it drops
// the universal block (and any legacy orchestrator section) from the instructions
// file and deletes the skill directory only when it carries the Active Stack
// ownership mark. If deleting the skill fails after the block was removed, the
// instructions file is put back byte for byte and the error names the path.
func (s *markerRemovalStep) removeLayered(a AgentAdapter) error {
	path := a.InstructionsPath(s.homeDir)
	original, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		return fmt.Errorf("read instructions file \"%s\": %w", path, readErr)
	}
	// WriteFileAtomic forces 0755 on the directory: put ~/.claude back as found.
	defer config.PreserveDirModes(filepath.Dir(path))()
	origMode := config.FileModeOr(path, 0o644)
	if readErr == nil {
		updated := string(original)
		for _, id := range []string{config.UniversalSectionID, s.h.ID} {
			updated = filemerge.InjectMarkdownSection(updated, id, "")
		}
		updated = config.PurgeStaleSections(updated)
		if updated != string(original) {
			if _, err := filemerge.WriteFileAtomic(path, []byte(updated), origMode); err != nil {
				return fmt.Errorf("write \"%s\": %w", path, err)
			}
		}
	}
	skillDir := filepath.Join(a.SkillsDir(s.homeDir), s.h.ID)
	raw, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil || !config.HasOwnerMark(string(raw)) {
		return nil // absent, or not ours: never delete it
	}
	if err := skillRemovalFn(skillDir); err != nil {
		if readErr == nil {
			if _, werr := filemerge.WriteFileAtomic(path, original, origMode); werr != nil {
				return fmt.Errorf("remove skill dir \"%s\": %w; additionally the block could not be restored in \"%s\": %v", skillDir, err, path, werr)
			}
		}
		return fmt.Errorf("remove skill dir \"%s\": %w; the universal block was restored in \"%s\"", skillDir, err, path)
	}
	return nil
}

func (s *markerRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	if s.nativeRollback != nil {
		return s.nativeRollback.restore(*s.manifest)
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

func (s *skillRemovalStep) ID() string                     { return "skill-removal:" + s.h.ID }
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

func (s *commandRemovalStep) ID() string                     { return "command-removal:" + s.h.ID }
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

func (s *permissionsRemovalStep) ID() string                     { return "permissions-removal:" + s.h.ID }
func (s *permissionsRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *permissionsRemovalStep) Run() error {
	// (TBD) There is no dedicated mechanism yet to selectively strip just the
	// permissions overlay back out of settings.json — the reverse of
	// internal/harness/config/permissions/install.go's mergeJSONFile/agentOverlay.
	//
	// This used to call restoreFn(*s.manifest) here (a copy of Rollback's own
	// body — see below), which was a real bug: s.manifest is the ENTIRE
	// uninstall-time snapshot shared across every harness being removed in
	// this run, not something scoped to permissions/settings.json alone.
	// Calling it from Run silently reverted every OTHER harness this uninstall
	// had already removed (e.g. markerRemovalStep's CLAUDE.md edit) back to
	// its pre-uninstall state, making a targeted uninstall of anything appear
	// to do nothing whenever "permissions" (the mandatory security floor) was
	// part of the selection — which it always is.
	//
	// Run is a safe no-op until a properly scoped removal exists: the
	// settings.json permissions overlay is left in place rather than risk
	// clobbering sibling harnesses' removals. Rollback (below) still
	// legitimately restores the full manifest — that IS its job when the
	// overall uninstall pipeline fails and everything applied so far must be
	// undone.
	return nil
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
	home            string
	overwriteEdited bool
	// skipPaths are never restored when the manifest is the registered
	// claude-native-opsx migration one (Claude settings.json: that migration
	// only reads it). Any other manifest restores them as usual.
	skipPaths  []string
	onProgress pipeline.ProgressFunc
}

func (s *restoreStep) ID() string { return "restore-from-backup" }

func (s *restoreStep) Run() error {
	m := s.installManifest
	if err := s.guardEdited(m); err != nil {
		return err
	}
	migration, err := registeredMigration(s.home, m)
	if err != nil {
		return err
	}
	if migration && len(s.skipPaths) > 0 {
		skip := map[string]bool{}
		for _, p := range s.skipPaths {
			skip[filepath.Clean(p)] = true
		}
		kept := make([]backup.ManifestEntry, 0, len(m.Entries))
		for _, e := range m.Entries {
			if !skip[filepath.Clean(e.OriginalPath)] {
				kept = append(kept, e)
			}
		}
		m.Entries = kept
	}
	return restoreFn(m)
}

// registeredMigration reports whether m is the manifest registered in
// claude-migration.json for one of its entries. An unreadable record is an
// error: the restore fails closed instead of guessing.
func registeredMigration(home string, m backup.Manifest) (bool, error) {
	for _, e := range m.Entries {
		rec, ok, err := config.LoadMigrationRecord(home, e.OriginalPath)
		if err != nil {
			return false, fmt.Errorf("not restoring: cannot read the migration record: %w", err)
		}
		if ok && filepath.Clean(rec.ManifestDir) == filepath.Clean(m.RootDir) {
			return true, nil
		}
	}
	return false, nil
}

// guardEdited applies the hash guard of the layered Claude delivery: for every
// manifest entry that is an instructions file with a registered migration, the
// current hash of CLAUDE.md and of each skill file is compared with the hash
// recorded when the migration finished. A mismatch means the user edited the
// file since; nothing is written, the current content is saved in a new
// snapshot, and the error asks for --overwrite-edited. With the flag the
// snapshot is still taken and the restore proceeds.
func (s *restoreStep) guardEdited(m backup.Manifest) error {
	var edited []string
	for _, e := range m.Entries {
		rec, ok, err := config.LoadMigrationRecord(s.home, e.OriginalPath)
		if err != nil {
			return fmt.Errorf("not restoring: cannot read the migration record: %w", err)
		}
		if !ok {
			continue
		}
		if filepath.Clean(rec.ManifestDir) != filepath.Clean(m.RootDir) {
			s.warn(fmt.Sprintf("warning: \"%s\" is not the registered migration manifest (\"%s\"); the hash guard still applies", m.RootDir, rec.ManifestDir))
		}
		expected := map[string]string{filepath.Clean(rec.InstructionsPath): rec.InstructionsSHA256}
		for p, h := range rec.SkillFiles {
			expected[filepath.Clean(p)] = h
		}
		for p, want := range expected {
			raw, err := os.ReadFile(p)
			if err != nil {
				continue // missing: nothing of the user's to lose
			}
			if config.SHA256Hex(raw) != want {
				edited = append(edited, p)
			}
		}
	}
	if len(edited) == 0 {
		return nil
	}
	sort.Strings(edited)
	snap := filepath.Join(s.home, ".active-stack", "backups", fmt.Sprintf("%s-uninstall-edited", time.Now().UTC().Format("20060102-150405.000000000")))
	if _, err := backup.NewSnapshotter().Create(snap, edited); err != nil {
		return fmt.Errorf("save the edited files before restoring: %w", err)
	}
	if s.overwriteEdited {
		s.warn(fmt.Sprintf("warning: restoring over files edited after the migration (%s); their current content is saved in %s", strings.Join(edited, ", "), snap))
		return nil
	}
	return fmt.Errorf("not restoring: %s changed after the migration; the current content is saved in %s (manifest %s); rerun with --overwrite-edited to restore anyway (--yes does not skip this guard)", strings.Join(edited, ", "), snap, m.RootDir)
}

func (s *restoreStep) warn(msg string) {
	if s.onProgress != nil {
		s.onProgress(pipeline.ProgressEvent{StepID: s.ID(), Stage: pipeline.StageApply, Status: pipeline.StepStatusDegraded, Err: fmt.Errorf("%s", msg)})
	}
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

func (s *externalSkipStep) ID() string { return "external-skip:" + s.h.ID }
func (s *externalSkipStep) Run() error { return nil }

type codexMCPRemovalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *codexMCPRemovalStep) ID() string                     { return "external-codex-mcp:" + s.h.ID }
func (s *codexMCPRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *codexMCPRemovalStep) Run() error {
	name := s.h.ID
	if s.h.External != nil && s.h.External.MCP != nil && s.h.External.MCP.Name != "" {
		name = s.h.External.MCP.Name
	}
	for _, adapter := range s.adapters {
		if adapter.Agent() != model.AgentCodex {
			continue
		}
		path := adapter.SettingsPath(s.homeDir)
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read Codex MCP config %q: %w", path, err)
		}
		merged, err := filemerge.RemoveTOMLSection(string(raw), "mcp_servers."+name)
		if err != nil {
			return fmt.Errorf("remove Codex MCP %q: %w", name, err)
		}
		if _, err := filemerge.WriteFileAtomic(path, []byte(merged), 0o644); err != nil {
			return fmt.Errorf("write Codex MCP config %q: %w", path, err)
		}
	}
	return nil
}

func (s *codexMCPRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

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
