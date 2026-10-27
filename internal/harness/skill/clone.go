package skill

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var commitSHARe = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// isCommitSHA reports whether ref is a full 40-char hexadecimal commit SHA.
// Shorter hex strings are ambiguous with branch/tag names and are NOT SHAs.
func isCommitSHA(ref string) bool { return commitSHARe.MatchString(ref) }

// gitCommands builds the ordered git commands that materialize the repo in dir.
// A full commit SHA cannot be passed to `git clone --branch`, so it is fetched
// by commit (shallow) and checked out detached; any other ref keeps the
// shallow `git clone [--branch <ref>]` behaviour.
func gitCommands(ref, repoURL, dir string) [][]string {
	if isCommitSHA(ref) {
		return [][]string{
			{"git", "init", dir},
			{"git", "-C", dir, "fetch", "--depth", "1", repoURL, ref},
			{"git", "-C", dir, "checkout", "--detach", "FETCH_HEAD"},
		}
	}
	args := []string{"git", "clone", "--depth", "1"}
	if ref != "" && ref != "latest" {
		args = append(args, "--branch", ref)
	}
	return [][]string{append(args, repoURL, dir)}
}

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
	tempDir, err := os.MkdirTemp("", "active-stack-clone-*")
	if err != nil {
		return Result{}, fmt.Errorf("skill %q: create temp dir: %w", skillID, err)
	}
	defer os.RemoveAll(tempDir) // always clean up, success or failure

	for _, args := range gitCommands(ref, repoURL, tempDir) {
		if err := runner.Run(ctx, args); err != nil {
			if isCommitSHA(ref) {
				return Result{}, fmt.Errorf("skill %q: git fetch %q at commit %q: %w", skillID, repoURL, ref, err)
			}
			return Result{}, fmt.Errorf("skill %q: git clone %q: %w", skillID, repoURL, err)
		}
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

	return refreshSkillDirectory(skillsDir, skillID, backupDir, func(stage string) error {
		return copyDir(srcDir, stage)
	})
}

// copyDir recursively copies src directory to dst, excluding the .git/
// subtree. Non-.git dotfiles (e.g. .gitignore) are preserved.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("skill source contains symlink %q", path)
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
