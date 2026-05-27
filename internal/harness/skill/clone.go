package skill

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// cloneInstaller performs a shallow git clone into a temp directory, copies
// the skill subdirectory to the agent's skills dir, and cleans up on exit.
func cloneInstaller(
	ctx context.Context,
	runner Runner,
	skillID, repo, ref, skillsDir, backupDir string,
) (Result, error) {
	if repo == "" {
		return Result{}, fmt.Errorf("skill %q: source.repo is empty (placeholder not yet confirmed)", skillID)
	}

	repoURL := "https://github.com/" + repo

	// Clone into a fresh temp directory.
	tempDir, err := os.MkdirTemp("", "jr-stack-clone-*")
	if err != nil {
		return Result{}, fmt.Errorf("skill %q: create temp dir: %w", skillID, err)
	}
	defer os.RemoveAll(tempDir) // always clean up, success or failure

	args := []string{"git", "clone", "--depth", "1"}
	if ref != "" && ref != "latest" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repoURL, tempDir)

	if err := runner.Run(ctx, args); err != nil {
		return Result{}, fmt.Errorf("skill %q: git clone %q: %w", skillID, repoURL, err)
	}

	// The cloned repo is expected to have a top-level directory named after the
	// skill ID.
	srcDir := filepath.Join(tempDir, skillID)
	if _, err := os.Stat(srcDir); err != nil {
		return Result{}, fmt.Errorf("skill %q: expected directory %q in cloned repo %q", skillID, skillID, repoURL)
	}

	// Read SKILL.md content for idempotence check.
	srcSKILLmd := filepath.Join(srcDir, "SKILL.md")
	newContent, err := os.ReadFile(srcSKILLmd)
	if err != nil {
		return Result{}, fmt.Errorf("skill %q: SKILL.md not found in cloned repo: %w", skillID, err)
	}

	destDir := filepath.Join(skillsDir, skillID)

	identical, err := checkIdempotent(skillsDir, skillID, newContent)
	if err != nil {
		return Result{}, err
	}
	if identical {
		return Result{SkillPath: destDir, AlreadyInstalled: true}, nil
	}

	// Backup if destination exists.
	if _, statErr := os.Stat(destDir); statErr == nil {
		if backupErr := snapshotSkillDir(backupDir, skillsDir, skillID); backupErr != nil {
			return Result{}, backupErr
		}
	}

	// Copy srcDir → destDir recursively.
	if err := copyDir(srcDir, destDir); err != nil {
		return Result{}, fmt.Errorf("skill %q: copy to skills dir: %w", skillID, err)
	}

	return Result{SkillPath: destDir}, nil
}

// copyDir recursively copies src directory to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
