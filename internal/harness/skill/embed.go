package skill

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// embedInstaller extracts a skill from the embedded FS and writes it to the
// agent's skills directory.  The FS is expected to contain the SKILL.md at
// "skills/<skillID>/SKILL.md".
func embedInstaller(embeddedFS fs.FS, skillID, skillsDir, backupDir string) (Result, error) {
	assetPath := "skills/" + skillID + "/SKILL.md"
	content, err := fs.ReadFile(embeddedFS, assetPath)
	if err != nil {
		return Result{}, fmt.Errorf("skill %q: embedded asset not found at %q: %w", skillID, assetPath, err)
	}

	destDir := filepath.Join(skillsDir, skillID)
	destFile := filepath.Join(destDir, "SKILL.md")

	// Idempotence check.
	identical, err := checkIdempotent(skillsDir, skillID, content)
	if err != nil {
		return Result{}, err
	}
	if identical {
		return Result{SkillPath: destDir, AlreadyInstalled: true}, nil
	}

	// Backup existing content if any.
	if _, statErr := os.Stat(destDir); statErr == nil {
		if backupErr := snapshotSkillDir(backupDir, skillsDir, skillID); backupErr != nil {
			return Result{}, backupErr
		}
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("skill %q: mkdir %q: %w", skillID, destDir, err)
	}
	if err := os.WriteFile(destFile, content, 0o644); err != nil {
		return Result{}, fmt.Errorf("skill %q: write SKILL.md: %w", skillID, err)
	}

	return Result{SkillPath: destDir}, nil
}
