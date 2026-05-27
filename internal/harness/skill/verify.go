package skill

import (
	"fmt"
	"os"
	"path/filepath"
)

// Verify checks that a skill was installed correctly by confirming that
// skillsDir/<skillID>/SKILL.md exists and is not empty.
func Verify(skillsDir, skillID string) error {
	skillMD := filepath.Join(skillsDir, skillID, "SKILL.md")
	info, err := os.Stat(skillMD)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q: SKILL.md not found at %q", skillID, skillMD)
		}
		return fmt.Errorf("skill %q: stat SKILL.md: %w", skillID, err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("skill %q: SKILL.md is empty at %q", skillID, skillMD)
	}
	return nil
}
