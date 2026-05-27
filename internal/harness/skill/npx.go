package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// npxInstaller installs a third-party skill via the `npx skills` CLI.
//
// Command executed:
//
//	npx skills add <skillID> --skills-dir <skillsDir>
//
// TODO: verificar flag contra `npx skills --help` en máquina real.
// El flag `--skills-dir` es el que usamos según el diseño D-4, pero la CLI de
// Vercel podría llamarlo `--dir` o `--output` en versiones futuras.
func npxInstaller(
	ctx context.Context,
	runner Runner,
	skillID, skillsDir, backupDir string,
) (Result, error) {
	args := []string{
		"npx", "skills", "add", skillID,
		"--skills-dir", skillsDir, // TODO: verificar flag contra npx skills --help en máquina real
	}

	if err := runner.Run(ctx, args); err != nil {
		return Result{}, fmt.Errorf("skill %q: npx skills add: %w", skillID, err)
	}

	// npx manages idempotence internally; we do a post-install check.
	destDir := filepath.Join(skillsDir, skillID)
	skillMDPath := filepath.Join(destDir, "SKILL.md")

	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, fmt.Errorf("skill %q: SKILL.md not found after npx install (expected at %q)", skillID, skillMDPath)
		}
		return Result{}, fmt.Errorf("skill %q: read post-install SKILL.md: %w", skillID, err)
	}

	// We cannot compare pre/post content reliably for npx (no source FS),
	// so AlreadyInstalled is always false — npx itself is idempotent.
	_ = data
	return Result{SkillPath: destDir, AlreadyInstalled: false}, nil
}
