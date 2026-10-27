package skill

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// refreshSkillDirectory stages a complete skill, compares every file, then
// snapshots and swaps the managed directory. A removed upstream file is removed
// from the installed copy; no partial source tree is written in place.
func refreshSkillDirectory(skillsDir, skillID, backupDir string, populate func(stage string) error) (Result, error) {
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("skill %q: create skills dir: %w", skillID, err)
	}
	stage, err := os.MkdirTemp(skillsDir, ".active-stack-stage-"+skillID+"-*")
	if err != nil {
		return Result{}, fmt.Errorf("skill %q: create staging dir: %w", skillID, err)
	}
	defer os.RemoveAll(stage)
	if err := populate(stage); err != nil {
		return Result{}, fmt.Errorf("skill %q: stage files: %w", skillID, err)
	}
	if info, err := os.Stat(filepath.Join(stage, "SKILL.md")); err != nil || info.IsDir() || info.Size() == 0 {
		return Result{}, fmt.Errorf("skill %q: staged SKILL.md missing or empty", skillID)
	}

	dest := filepath.Join(skillsDir, skillID)
	info, statErr := os.Lstat(dest)
	if statErr != nil && !os.IsNotExist(statErr) {
		return Result{}, fmt.Errorf("skill %q: inspect installed directory: %w", skillID, statErr)
	}
	exists := statErr == nil
	if exists && !info.IsDir() {
		return Result{}, fmt.Errorf("skill %q: installed path is not a directory: %q", skillID, dest)
	}
	if exists {
		identical, err := sameSkillTree(stage, dest)
		if err != nil {
			return Result{}, fmt.Errorf("skill %q: compare files: %w", skillID, err)
		}
		if identical {
			return Result{SkillPath: dest, AlreadyInstalled: true}, nil
		}
		if err := snapshotSkillDir(backupDir, skillsDir, skillID); err != nil {
			return Result{}, err
		}
	}

	old := stage + ".previous"
	if exists {
		if err := os.Rename(dest, old); err != nil {
			return Result{}, fmt.Errorf("skill %q: move installed directory aside: %w", skillID, err)
		}
	}
	if err := os.Rename(stage, dest); err != nil {
		if exists {
			_ = os.Rename(old, dest)
		}
		return Result{}, fmt.Errorf("skill %q: activate staged directory: %w", skillID, err)
	}
	if exists {
		_ = os.RemoveAll(old) // snapshot is already durable; a locked old dir may remain temporarily
	}
	return Result{SkillPath: dest}, nil
}

func sameSkillTree(a, b string) (bool, error) {
	left, err := skillTreeDigests(a)
	if err != nil {
		return false, err
	}
	right, err := skillTreeDigests(b)
	if err != nil {
		return false, err
	}
	if len(left) != len(right) {
		return false, nil
	}
	for path, digest := range left {
		other, ok := right[path]
		if !ok || other != digest {
			return false, nil
		}
	}
	return true, nil
}

func skillTreeDigests(root string) (map[string][32]byte, error) {
	files := make(map[string][32]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in skill directory: %q", path)
		}
		if entry.IsDir() {
			files[rel+string(os.PathSeparator)] = [32]byte{}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		var digest [32]byte
		copy(digest[:], hash.Sum(nil))
		files[rel] = digest
		return nil
	})
	return files, err
}

// snapshotSkillDir creates a backup snapshot of the existing skill directory
// before overwriting it.  backupDir is the parent directory for all snapshots;
// a timestamped sub-directory is created inside it.
func snapshotSkillDir(backupDir, skillsDir, skillID string) error {
	skillDir := filepath.Join(skillsDir, skillID)

	// Include directories so a restore also recreates empty asset folders.
	var paths []string
	err := filepath.WalkDir(skillDir, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		paths = append(paths, p)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to back up
		}
		return fmt.Errorf("skill %q: walk dir for backup: %w", skillID, err)
	}

	// Name the snapshot directory with a timestamp + skillID.
	snapName := fmt.Sprintf("%s-%s", time.Now().UTC().Format("20060102-150405.000000000"), skillID)
	snapDir := filepath.Join(backupDir, snapName)

	s := backup.NewSnapshotter()
	if _, err := s.Create(snapDir, paths); err != nil {
		return fmt.Errorf("skill %q: snapshot: %w", skillID, err)
	}
	return nil
}
