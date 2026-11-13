package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

// layeredSnapshot and layeredHooks are test seams (like SnapshotterCreate).
var layeredSnapshot = func(dir string, paths []string, hints map[string]bool) (backup.Manifest, error) {
	return backup.NewSnapshotter().CreateWithDirHints(dir, paths, hints)
}

var layeredHooks struct {
	beforeSkill, beforeInstructions, afterInstructions func() error
}

// LayeredInstall describes one layered Claude delivery (claude-native-opsx D5):
// a managed coordinator skill plus a minimal universal block in the
// instructions file, written transactionally.
type LayeredInstall struct {
	// HomeDir holds the backups root (<HomeDir>/.active-stack/backups), also for
	// a Project target (C8).
	HomeDir          string
	InstructionsPath string
	SkillsDir        string
	// ShadowSkillsDir is the Machine skills dir when installing into Project.
	ShadowSkillsDir  string
	Skill            SkillSpec
	UniversalContent string
}

// LayeredResult is the outcome of InstallLayered.
type LayeredResult struct {
	// ManifestDir is the snapshot (manifest) taken before any write.
	ManifestDir         string
	SkillChanged        bool
	InstructionsChanged bool
	// MigratedLegacy is true when the instructions file held the legacy
	// sdd-orchestrator section that this run removed.
	MigratedLegacy bool
	Warnings       []string
}

// PreserveDirModes records the mode of each existing directory and, when it lacks the
// owner write bit (left 0500 by an old installer), adds only that bit so the
// write can proceed. The returned func puts every directory back to its recorded
// mode (plus the owner write bit), undoing the 0755 that WriteFileAtomic forces.
// Scope is G6 (small): only the directories of this install. Install and
// uninstall share it.
func PreserveDirModes(dirs ...string) func() {
	type rec struct {
		dir  string
		mode os.FileMode
	}
	var recs []rec
	for _, d := range dirs {
		info, err := os.Stat(d)
		if err != nil || !info.IsDir() {
			continue
		}
		m := info.Mode().Perm() | 0o200
		if m != info.Mode().Perm() {
			_ = os.Chmod(d, m)
		}
		recs = append(recs, rec{d, m})
	}
	return func() {
		for _, r := range recs {
			_ = os.Chmod(r.dir, r.mode)
		}
	}
}

// FileModeOr returns the mode of an existing file, or def when it does not exist.
func FileModeOr(path string, def os.FileMode) os.FileMode {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().Perm()
	}
	return def
}

// InstallLayered runs D5 steps 2-6: ownership/collision check, one snapshot
// manifest under <HomeDir>/.active-stack/backups, skill directory swap, then the
// CLAUDE.md edit (legacy coordinator sections purged, universal block injected).
// If a write step fails after the skill was activated, everything is restored
// from the manifest, but a file that changed after the installer wrote it is
// preserved (backup.Manifest.GuardRollback) and named in the error.
func InstallLayered(in LayeredInstall) (LayeredResult, error) {
	var res LayeredResult
	skillDest := filepath.Join(in.SkillsDir, in.Skill.ID)

	restoreModes := PreserveDirModes(filepath.Dir(in.InstructionsPath), in.SkillsDir)
	defer restoreModes()

	// Step 2: collision / precedence checks, before any write.
	if err := validateSkillSpec(in.Skill); err != nil {
		return res, err
	}
	skillExisted := false
	if info, err := os.Lstat(skillDest); err == nil {
		if !info.IsDir() {
			return res, fmt.Errorf("layered install: \"%s\" exists and is not a directory; left untouched", skillDest)
		}
		raw, rerr := os.ReadFile(filepath.Join(skillDest, "SKILL.md"))
		if rerr != nil || !hasOwnerMark(string(raw)) {
			return res, fmt.Errorf("layered install: \"%s\" exists without the %s ownership mark; refusing to overwrite it", skillDest, SkillOwnerKey)
		}
		skillExisted = true
	} else if !os.IsNotExist(err) {
		return res, fmt.Errorf("layered install: inspect \"%s\": %w", skillDest, err)
	}
	if in.ShadowSkillsDir != "" {
		shadow := filepath.Join(in.ShadowSkillsDir, in.Skill.ID)
		if _, err := os.Lstat(shadow); err == nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("personal skill \"%s\" takes precedence over the project skill \"%s\" (personal over project)", shadow, skillDest))
		}
	}

	// Nothing to do: same skill files and same CLAUDE.md. No snapshot, no write.
	if skillExisted && skillMatches(skillDest, in.Skill) {
		if want, werr := expectedInstructions(in); werr == nil && want == readFileOrEmpty(in.InstructionsPath) {
			return res, nil
		}
	}

	// Step 3: one snapshot of CLAUDE.md and the skill directory, always under
	// <HomeDir>/.active-stack/backups (also for a Project target, C8).
	paths := []string{in.InstructionsPath}
	hints := map[string]bool{}
	if skillExisted {
		err := filepath.WalkDir(skillDest, func(p string, _ os.DirEntry, werr error) error {
			if werr != nil {
				return werr
			}
			paths = append(paths, p)
			return nil
		})
		if err != nil {
			return res, fmt.Errorf("layered install: walk \"%s\" for backup: %w", skillDest, err)
		}
	} else {
		paths = append(paths, skillDest)
		hints[filepath.Clean(skillDest)] = true
	}
	snapDir := filepath.Join(in.HomeDir, ".active-stack", "backups",
		fmt.Sprintf("%s-%s-layered", time.Now().UTC().Format("20060102-150405.000000000"), in.Skill.ID))
	manifest, err := layeredSnapshot(snapDir, paths, hints)
	if err != nil {
		return res, fmt.Errorf("layered install: snapshot: %w", err)
	}
	res.ManifestDir = snapDir

	// Step 4: activate the skill directory.
	if layeredHooks.beforeSkill != nil {
		if err := layeredHooks.beforeSkill(); err != nil {
			return res, fmt.Errorf("layered install: activate skill: %w", err)
		}
	}
	sr, err := InjectManagedSkill(in.SkillsDir, in.Skill, SkillInjectOptions{ExternalSnapshot: true})
	if err != nil {
		return res, fmt.Errorf("layered install: activate skill: %w", err)
	}
	res.SkillChanged = sr.Changed
	if skillExisted && sr.Changed {
		res.Warnings = append(res.Warnings, fmt.Sprintf("skill \"%s\" was replaced; the previous version (including any edits) is saved in %s", skillDest, snapDir))
	}
	skillHashes, err := fileHashes(skillDest)
	if err != nil {
		return res, rollbackLayered(fmt.Errorf("layered install: hash skill files: %w", err), manifest, skillDest, skillHashes, nil, in.InstructionsPath)
	}

	// Step 5: purge the legacy coordinator sections and inject the universal block.
	written := false
	var mdHash [sha256.Size]byte
	step5 := func() error {
		if layeredHooks.beforeInstructions != nil {
			if err := layeredHooks.beforeInstructions(); err != nil {
				return err
			}
		}
		original := readFileOrEmpty(in.InstructionsPath)
		res.Warnings = append(res.Warnings, unclosedMarkerWarnings(in.InstructionsPath, original)...)
		for _, id := range filemerge.MarkedSectionIDs(original) {
			if id == "sdd-orchestrator" {
				res.MigratedLegacy = true
			}
		}
		content, _ := expectedInstructions(in)
		wr, err := filemerge.WriteFileAtomic(in.InstructionsPath, []byte(content), FileModeOr(in.InstructionsPath, 0o644))
		if err != nil {
			return fmt.Errorf("write \"%s\": %w", in.InstructionsPath, err)
		}
		written = true
		mdHash = sha256.Sum256([]byte(content))
		res.InstructionsChanged = wr.Changed
		if layeredHooks.afterInstructions != nil {
			return layeredHooks.afterInstructions()
		}
		return nil
	}
	if err := step5(); err != nil {
		var guard *[sha256.Size]byte
		if written {
			guard = &mdHash
		}
		return res, rollbackLayered(fmt.Errorf("layered install: update instructions: %w", err), manifest, skillDest, skillHashes, guard, in.InstructionsPath)
	}
	return res, nil
}

// expectedInstructions is the CLAUDE.md content after the layered install:
// legacy coordinator sections purged, stale sections purged, universal block in.
func expectedInstructions(in LayeredInstall) (string, error) {
	content := readFileOrEmpty(in.InstructionsPath)
	for _, id := range []string{"sdd-orchestrator", "sdd-delegation", "sdd-model-assignments"} {
		content = filemerge.InjectMarkdownSection(content, id, "")
	}
	content = PurgeStaleSections(content)
	return filemerge.InjectMarkdownSection(content, UniversalSectionID, in.UniversalContent), nil
}

// skillMatches reports whether the skill directory holds exactly the files
// InjectManagedSkill would write.
func skillMatches(dest string, spec SkillSpec) bool {
	want := map[string]string{"SKILL.md": RenderSkillMD(spec)}
	for n, c := range spec.Modules {
		want[filepath.ToSlash(filepath.Join("modules", n+".md"))] = c
	}
	got := map[string]string{}
	err := filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dest, p)
		got[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil || len(got) != len(want) {
		return false
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

func fileHashes(root string) (map[string][sha256.Size]byte, error) {
	out := map[string][sha256.Size]byte{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[p] = sha256.Sum256(b)
		return nil
	})
	return out, err
}

// rollbackLayered undoes steps 4-5 from the manifest. The skill directory is
// removed (and, when it existed, its old files restored) only if none of the
// files the installer wrote changed since; CLAUDE.md goes through GuardRollback
// and is only restored if the installer actually wrote it.
func rollbackLayered(cause error, m backup.Manifest, skillDest string, skillHashes map[string][sha256.Size]byte, mdGuard *[sha256.Size]byte, mdPath string) error {
	var conflicts error
	skillIntact := true
	for p, want := range skillHashes {
		raw, err := os.ReadFile(p)
		if err != nil || sha256.Sum256(raw) != want {
			skillIntact = false
			conflicts = errors.Join(conflicts, fmt.Errorf("preserved \"%s\" because it changed after installation; backup at \"%s\"", p, m.RootDir))
		}
	}
	if skillIntact {
		if err := os.RemoveAll(skillDest); err != nil {
			conflicts = errors.Join(conflicts, fmt.Errorf("remove \"%s\" during rollback: %w", skillDest, err))
		}
	}
	if mdGuard != nil {
		m.GuardRollback(mdPath, *mdGuard)
	}
	var entries []backup.ManifestEntry
	for _, e := range m.Entries {
		p := filepath.Clean(e.OriginalPath)
		switch {
		case p == filepath.Clean(mdPath):
			if mdGuard != nil {
				entries = append(entries, e)
			}
		case p == filepath.Clean(skillDest) || strings.HasPrefix(p, filepath.Clean(skillDest)+string(filepath.Separator)):
			if skillIntact {
				entries = append(entries, e)
			}
		}
	}
	m.Entries = entries
	rerr := backup.RestoreService{}.Restore(m)
	return errors.Join(cause, conflicts, rerr)
}
