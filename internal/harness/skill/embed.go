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
	if len(content) == 0 {
		return Result{}, fmt.Errorf("skill %q: embedded SKILL.md is empty", skillID)
	}

	return refreshSkillDirectory(skillsDir, skillID, backupDir, func(stage string) error {
		return fs.WalkDir(embeddedFS, "skills/"+skillID, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel("skills/"+skillID, path)
			if err != nil {
				return err
			}
			target := filepath.Join(stage, rel)
			if d.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			data, err := fs.ReadFile(embeddedFS, path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o644)
		})
	})
}
