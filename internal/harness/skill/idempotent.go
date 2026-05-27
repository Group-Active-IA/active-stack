package skill

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// checkIdempotent inspects the destination directory for skillID inside
// skillsDir and compares the existing SKILL.md (if any) against newContent.
//
// Returns:
//   - identical=true when the existing content matches — caller must no-op.
//   - identical=false, nil error when different — caller must backup and overwrite.
//   - identical=false, non-nil error when the existing file cannot be read.
func checkIdempotent(skillsDir, skillID string, newContent []byte) (identical bool, err error) {
	existing := filepath.Join(skillsDir, skillID, "SKILL.md")
	data, err := os.ReadFile(existing)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // fresh install — not identical, no error
		}
		return false, fmt.Errorf("skill %q: read existing SKILL.md: %w", skillID, err)
	}
	return bytes.Equal(data, newContent), nil
}

// snapshotSkillDir creates a backup snapshot of the existing skill directory
// before overwriting it.  backupDir is the parent directory for all snapshots;
// a timestamped sub-directory is created inside it.
func snapshotSkillDir(backupDir, skillsDir, skillID string) error {
	skillDir := filepath.Join(skillsDir, skillID)

	// Collect files to back up.
	var paths []string
	err := filepath.WalkDir(skillDir, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to back up
		}
		return fmt.Errorf("skill %q: walk dir for backup: %w", skillID, err)
	}

	// Name the snapshot directory with a timestamp + skillID.
	snapName := fmt.Sprintf("%s-%s", time.Now().UTC().Format("20060102-150405"), skillID)
	snapDir := filepath.Join(backupDir, snapName)

	s := backup.NewSnapshotter()
	if _, err := s.Create(snapDir, paths); err != nil {
		return fmt.Errorf("skill %q: snapshot: %w", skillID, err)
	}
	return nil
}
