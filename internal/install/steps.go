package install

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	cfginstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
	perminstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config/permissions"
	skillinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// ─────────────────────────────────────────────────────────────────
// snapshotStep — Prepare stage; creates the backup snapshot.
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
		return fmt.Errorf("snapshot: %w", err)
	}
	if s.manifestOut != nil {
		*s.manifestOut = manifest
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────
// externalStep — wraps external.Install.
// ─────────────────────────────────────────────────────────────────

// externalInstallFn is the backing function for external installation.
// It is a package-level variable so tests can inject a fake.
var externalInstallFn = func(
	ctx context.Context,
	h model.Harness,
	profile system.PlatformProfile,
	adapters []extinstaller.AgentAdapter,
	homeDir string,
) (extinstaller.Result, error) {
	return extinstaller.Install(ctx, h, profile, adapters, homeDir)
}

type externalStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *externalStep) ID() string { return "external:" + s.h.ID }
func (s *externalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *externalStep) Run() error {
	ext := toExternalAdapters(s.adapters)
	_, err := externalInstallFn(context.Background(), s.h, system.PlatformProfile{}, ext, s.homeDir)
	return err
}

func (s *externalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// skillStep — wraps skill.NewInstaller(...).Install.
// ─────────────────────────────────────────────────────────────────

// skillInstallFn is the backing function for skill installation.
// It is a package-level variable so tests can inject a fake.
var skillInstallFn = func(
	runner skillRunner,
	embeddedFS fs.FS,
	ctx context.Context,
	h model.Harness,
	adapters []skillinstaller.AgentAdapter,
	homeDir, backupDir string,
) ([]skillinstaller.Result, error) {
	ins := skillinstaller.NewInstaller(runner, embeddedFS)
	return ins.Install(ctx, h, adapters, homeDir, backupDir)
}

type skillStep struct {
	h          model.Harness
	adapters   []AgentAdapter
	homeDir    string
	backupDir  string
	embeddedFS fs.FS
	runner     skillRunner
	manifest   *backup.Manifest
}

func (s *skillStep) ID() string { return "skill:" + s.h.ID }
func (s *skillStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *skillStep) Run() error {
	skill := toSkillAdapters(s.adapters)
	_, err := skillInstallFn(s.runner, s.embeddedFS, context.Background(), s.h, skill, s.homeDir, s.backupDir)
	return err
}

func (s *skillStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// configStep — wraps config.Install.
// ─────────────────────────────────────────────────────────────────

// configInstallFn is the backing function for config installation.
// It is a package-level variable so tests can inject a fake.
var configInstallFn = func(
	h model.Harness,
	adapters []cfginstaller.AgentAdapter,
	homeDir string,
) (cfginstaller.Result, error) {
	return cfginstaller.Install(h, adapters, homeDir)
}

type configStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *configStep) ID() string { return "config:" + s.h.ID }
func (s *configStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *configStep) Run() error {
	cfg := toConfigAdapters(s.adapters)
	_, err := configInstallFn(s.h, cfg, s.homeDir)
	return err
}

func (s *configStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// permissionsStep — wraps permissions.Install.
// ─────────────────────────────────────────────────────────────────

// permissionsInstallFn is the backing function for permissions installation.
// It is a package-level variable so tests can inject a fake.
var permissionsInstallFn = func(
	homeDir string,
	adapters []perminstaller.PermissionsAdapter,
) (perminstaller.Result, error) {
	return perminstaller.Install(homeDir, adapters)
}

type permissionsStep struct {
	h        model.Harness
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *permissionsStep) ID() string { return "permissions:" + s.h.ID }
func (s *permissionsStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *permissionsStep) Run() error {
	perm := toPermissionsAdapters(s.adapters)
	_, err := permissionsInstallFn(s.homeDir, perm)
	return err
}

func (s *permissionsStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}

// ─────────────────────────────────────────────────────────────────
// Adapter coercion helpers
// ─────────────────────────────────────────────────────────────────

// toExternalAdapters narrows the full AgentAdapter to external.AgentAdapter.
func toExternalAdapters(adapters []AgentAdapter) []extinstaller.AgentAdapter {
	out := make([]extinstaller.AgentAdapter, len(adapters))
	for i, a := range adapters {
		out[i] = a
	}
	return out
}

// toSkillAdapters narrows the full AgentAdapter to skill.AgentAdapter.
func toSkillAdapters(adapters []AgentAdapter) []skillinstaller.AgentAdapter {
	out := make([]skillinstaller.AgentAdapter, len(adapters))
	for i, a := range adapters {
		out[i] = a
	}
	return out
}

// toConfigAdapters narrows the full AgentAdapter to config.AgentAdapter.
func toConfigAdapters(adapters []AgentAdapter) []cfginstaller.AgentAdapter {
	out := make([]cfginstaller.AgentAdapter, len(adapters))
	for i, a := range adapters {
		out[i] = a
	}
	return out
}

// toPermissionsAdapters narrows the full AgentAdapter to permissions.PermissionsAdapter.
func toPermissionsAdapters(adapters []AgentAdapter) []perminstaller.PermissionsAdapter {
	out := make([]perminstaller.PermissionsAdapter, len(adapters))
	for i, a := range adapters {
		out[i] = a
	}
	return out
}
