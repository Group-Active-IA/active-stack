package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
)

// SkillOwnerKey is the SKILL.md frontmatter key that marks a skill directory as
// owned by Active Stack (claude-native-opsx D1, C9).
const SkillOwnerKey = "active-stack-managed"

// SkillSpec describes the managed coordinator skill to write.
type SkillSpec struct {
	ID          string
	Description string
	Body        string
	// Modules maps a module name to its content; written as modules/<name>.md.
	Modules map[string]string
}

// SkillInjectOptions tunes InjectManagedSkill.
type SkillInjectOptions struct {
	// BackupDir receives a snapshot of an existing skill directory before it is
	// replaced. Required unless ExternalSnapshot is true.
	BackupDir string
	// ExternalSnapshot means the caller already snapshotted the directory in a
	// transaction manifest, so no extra snapshot is taken here.
	ExternalSnapshot bool
	// ShadowSkillsDir is the Machine skills dir to check when installing into a
	// Project target: a skill with the same ID there takes precedence.
	ShadowSkillsDir string
}

// SkillInjectResult is the outcome of InjectManagedSkill.
type SkillInjectResult struct {
	Path     string
	Changed  bool
	Warnings []string
}

// RenderSkillMD builds SKILL.md: frontmatter (name, description,
// disable-model-invocation, ownership mark) followed by the body.
func RenderSkillMD(spec SkillSpec) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + spec.ID + "\n")
	b.WriteString("description: " + strings.ReplaceAll(strings.TrimSpace(spec.Description), "\n", " ") + "\n")
	b.WriteString("disable-model-invocation: true\n")
	b.WriteString(SkillOwnerKey + ": true\n")
	b.WriteString("---\n\n")
	b.WriteString(spec.Body)
	return b.String()
}

// HasOwnerMark reports whether SKILL.md content carries the Active Stack ownership mark.
func HasOwnerMark(content string) bool { return hasOwnerMark(content) }

// hasOwnerMark reports whether SKILL.md content carries the ownership mark in
// its frontmatter.
func hasOwnerMark(content string) bool {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false
	}
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "---" {
			return false
		}
		if strings.TrimSpace(l) == SkillOwnerKey+": true" {
			return true
		}
	}
	return false
}

func validateSkillSpec(spec SkillSpec) error {
	if spec.ID == "" || spec.ID != filepath.Base(spec.ID) || spec.ID == "." || spec.ID == ".." {
		return fmt.Errorf("managed skill: invalid skill ID %q", spec.ID)
	}
	for name := range spec.Modules {
		if name == "" || name == "." || name == ".." || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) {
			return fmt.Errorf("managed skill %q: invalid module name %q", spec.ID, name)
		}
	}
	return nil
}

// InjectManagedSkill writes <skillsDir>/<spec.ID>/ (SKILL.md + modules/) with
// the staged swap of the skill installer. It refuses to replace a directory
// that lacks the Active Stack ownership mark (C9) and warns when a Machine skill
// with the same ID would take precedence over a Project one.
func InjectManagedSkill(skillsDir string, spec SkillSpec, opts SkillInjectOptions) (SkillInjectResult, error) {
	if err := validateSkillSpec(spec); err != nil {
		return SkillInjectResult{}, err
	}
	if opts.BackupDir == "" && !opts.ExternalSnapshot {
		return SkillInjectResult{}, fmt.Errorf("managed skill %q: a backup dir or an external snapshot is required", spec.ID)
	}
	dest := filepath.Join(skillsDir, spec.ID)

	if info, err := os.Lstat(dest); err == nil {
		if !info.IsDir() {
			return SkillInjectResult{}, fmt.Errorf("managed skill %q: %q exists and is not a directory; left untouched", spec.ID, dest)
		}
		raw, rerr := os.ReadFile(filepath.Join(dest, "SKILL.md"))
		if rerr != nil || !hasOwnerMark(string(raw)) {
			return SkillInjectResult{}, fmt.Errorf("managed skill %q: %q exists without the %s ownership mark; refusing to overwrite it", spec.ID, dest, SkillOwnerKey)
		}
	} else if !os.IsNotExist(err) {
		return SkillInjectResult{}, fmt.Errorf("managed skill %q: inspect %q: %w", spec.ID, dest, err)
	}

	var warnings []string
	if opts.ShadowSkillsDir != "" {
		shadow := filepath.Join(opts.ShadowSkillsDir, spec.ID)
		if _, err := os.Lstat(shadow); err == nil {
			warnings = append(warnings, fmt.Sprintf("personal skill %q takes precedence over the project skill %q (personal over project)", shadow, dest))
		}
	}

	backupDir := opts.BackupDir
	if opts.ExternalSnapshot {
		backupDir = ""
	}
	// The skill directory keeps its previous mode (0755 when new): staging dirs
	// are created 0700.
	destMode := os.FileMode(0o755)
	if info, err := os.Stat(dest); err == nil {
		destMode = info.Mode().Perm()
	}
	res, err := skill.RefreshManagedDirectory(skillsDir, spec.ID, backupDir, func(stage string) error {
		if err := os.WriteFile(filepath.Join(stage, "SKILL.md"), []byte(RenderSkillMD(spec)), 0o644); err != nil {
			return err
		}
		if len(spec.Modules) == 0 {
			return nil
		}
		if err := os.MkdirAll(filepath.Join(stage, "modules"), 0o755); err != nil {
			return err
		}
		for name, content := range spec.Modules {
			if err := os.WriteFile(filepath.Join(stage, "modules", name+".md"), []byte(content), 0o644); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return SkillInjectResult{}, err
	}
	if !res.AlreadyInstalled {
		_ = os.Chmod(res.SkillPath, destMode)
	}
	return SkillInjectResult{Path: res.SkillPath, Changed: !res.AlreadyInstalled, Warnings: warnings}, nil
}
