package skill

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// cloneInstaller performs a shallow git clone into a temp directory, copies
// the skill subdirectory to the agent's skills dir, and cleans up on exit.
func cloneInstaller(
	ctx context.Context,
	runner Runner,
	skillID, repo, ref, path, skillsDir, backupDir string,
) (Result, error) {
	if repo == "" {
		return Result{}, fmt.Errorf("skill %q: source.repo is empty (placeholder not yet confirmed)", skillID)
	}

	// If repo already contains a URL scheme (e.g. "file://", "https://"), use it
	// as-is; otherwise treat it as a GitHub owner/repo short path.
	// This allows hermetic E2E tests to supply file:// fixture repos without
	// network access while keeping the default production behaviour unchanged.
	repoURL := repo
	if !strings.Contains(repo, "://") {
		repoURL = "https://github.com/" + repo
	}

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

	// Resolve the skill source directory.
	//
	// When path is set (C-22), the skill lives in an explicit subdir of the
	// repo (e.g. third-party monorepos). We use it directly — no root fallback.
	//
	// When path is empty (C-16), use root-first with subdir fallback:
	//   1. <tempDir>/SKILL.md exists → root layout (our convention).
	//   2. <tempDir>/<skillID>/SKILL.md exists → legacy subdir layout.
	//   3. Neither → descriptive error.
	var srcDir string
	switch {
	case path != "":
		srcDir = filepath.Join(tempDir, path)
		if !fileExists(filepath.Join(srcDir, "SKILL.md")) {
			return Result{}, fmt.Errorf(
				"skill %q: SKILL.md not found in %q subdir of repo %q",
				skillID, path, repoURL)
		}
	case fileExists(filepath.Join(tempDir, "SKILL.md")):
		srcDir = tempDir
	case fileExists(filepath.Join(tempDir, skillID, "SKILL.md")):
		srcDir = filepath.Join(tempDir, skillID)
	default:
		return Result{}, fmt.Errorf(
			"skill %q: SKILL.md not found at clone root nor in %q subdir of repo %q",
			skillID, skillID, repoURL)
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

// copyDir recursively copies src directory to dst, excluding the .git/
// subtree. Non-.git dotfiles (e.g. .gitignore) are preserved.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		// Skip the .git/ directory and everything inside it.
		// We match exactly ".git" or ".git" + separator to avoid accidentally
		// dropping .gitignore, .gitattributes, etc.
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

// fileExists returns true if path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
