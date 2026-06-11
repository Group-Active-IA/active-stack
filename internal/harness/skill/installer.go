package skill

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Runner is the interface for executing external commands. It is satisfied by
// the real OS exec wrapper and by stub implementations in tests.
type Runner interface {
	Run(ctx context.Context, args []string) error
}

// Installer dispatches skill installation to the correct method based on
// Source.Method (clone | embed).
type Installer struct {
	runner     Runner
	embeddedFS fs.FS
}

// NewInstaller creates a new Installer.
//   - runner is used for the clone method (can be nil if only embed is needed).
//   - embeddedFS is the FS for the embed method (can be nil if only clone is needed).
func NewInstaller(runner Runner, embeddedFS fs.FS) *Installer {
	return &Installer{runner: runner, embeddedFS: embeddedFS}
}

// Install installs the skill harness h for each adapter that supports skills.
// It returns one Result per adapter that was not skipped.
//
// Parameters:
//   - ctx: context for cancellation.
//   - h: the harness to install (must be of type HarnessSkill).
//   - adapters: per-agent adapters; adapters with an empty SkillsDir are skipped.
//   - homeDir: the user's home directory, passed to adapter.SkillsDir.
//   - backupDir: directory where pre-overwrite snapshots are stored.
func (ins *Installer) Install(
	ctx context.Context,
	h model.Harness,
	adapters []AgentAdapter,
	homeDir, backupDir string,
) ([]Result, error) {
	if h.Source == nil {
		return nil, fmt.Errorf("harness %q has no Source", h.ID)
	}
	if h.Source.Method == "" {
		return nil, fmt.Errorf("harness %q: source.method is empty (should be inferred by catalog.Load)", h.ID)
	}

	var results []Result
	for _, adapter := range adapters {
		skillsDir := adapter.SkillsDir(homeDir)
		if skillsDir == "" {
			// This adapter does not support skills — skip silently.
			continue
		}

		result, err := ins.installForAdapter(ctx, h, skillsDir, backupDir)
		if err != nil {
			return nil, fmt.Errorf("agent %q: %w", adapter.Agent(), err)
		}
		results = append(results, result)
	}
	return results, nil
}

// installForAdapter runs the correct installer for a single agent's skills dir.
func (ins *Installer) installForAdapter(
	ctx context.Context,
	h model.Harness,
	skillsDir, backupDir string,
) (Result, error) {
	switch h.Source.Method {
	case "clone":
		return cloneInstaller(ctx, ins.runner, h.ID, h.Source.Repo, h.Source.Ref, h.Source.Path, skillsDir, backupDir)
	case "embed":
		if ins.embeddedFS == nil {
			return Result{}, fmt.Errorf("harness %q: embed method requires an embedded FS", h.ID)
		}
		return embedInstaller(ins.embeddedFS, h.ID, skillsDir, backupDir)
	default:
		return Result{}, fmt.Errorf("harness %q: unsupported source.method %q (supported: clone, embed)", h.ID, h.Source.Method)
	}
}
