package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

type layeredEnv struct {
	root, home, md, skills string
}

func newLayeredEnv(t *testing.T) layeredEnv {
	t.Helper()
	root := t.TempDir()
	e := layeredEnv{
		root:   root,
		home:   filepath.Join(root, "home"),
		md:     filepath.Join(root, "home", ".claude", "CLAUDE.md"),
		skills: filepath.Join(root, "home", ".claude", "skills"),
	}
	legacy := filemerge.InjectMarkdownSection("# Mine\n\nuser text\n", "sdd-orchestrator", "LEGACY-ORCH\n")
	legacy = filemerge.InjectMarkdownSection(legacy, "sdd-delegation", "LEGACY-DELEG\n")
	if err := os.MkdirAll(filepath.Dir(e.md), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(e.md, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e layeredEnv) install() LayeredInstall {
	return LayeredInstall{
		HomeDir:          e.home,
		InstructionsPath: e.md,
		SkillsDir:        e.skills,
		Skill: SkillSpec{
			ID: "sdd-orchestrator", Description: "Coordinator", Body: "NEW-BODY\n",
			Modules: map[string]string{"tdd": "NEW-TDD\n"},
		},
		UniversalContent: "UNIVERSAL\n",
	}
}

func readStr(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Errorf("expected %s to exist: %v", p, err)
	}
	return string(b)
}

func treeOf(root string) map[string]string {
	out := map[string]string{}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			b, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(root, p)
			out[rel] = string(b)
		}
		return nil
	})
	return out
}

func equalTrees(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

func resetLayeredSeams(t *testing.T) {
	t.Helper()
	oldSnap, oldHooks := layeredSnapshot, layeredHooks
	t.Cleanup(func() { layeredSnapshot, layeredHooks = oldSnap, oldHooks })
}

func TestInstallLayered_Success(t *testing.T) {
	e := newLayeredEnv(t)
	res, err := InstallLayered(e.install())
	if err != nil {
		t.Fatalf("InstallLayered: %v", err)
	}
	md := readStr(t, e.md)
	if strings.Contains(md, "active-stack:sdd-orchestrator") || strings.Contains(md, "active-stack:sdd-delegation") || strings.Contains(md, "LEGACY") {
		t.Errorf("legacy sections must be purged:\n%s", md)
	}
	if strings.Count(md, "<!-- active-stack:"+UniversalSectionID+" -->") != 1 || !strings.Contains(md, "user text") {
		t.Errorf("want one universal section and the user text kept:\n%s", md)
	}
	if !strings.Contains(readStr(t, filepath.Join(e.skills, "sdd-orchestrator", "SKILL.md")), "NEW-BODY") {
		t.Error("skill not written")
	}
	if !res.SkillChanged || !res.InstructionsChanged || res.ManifestDir == "" {
		t.Errorf("unexpected result %+v", res)
	}
}

// 5.4: a failing snapshot leaves everything untouched.
func TestInstallLayered_SnapshotFailureChangesNothing(t *testing.T) {
	resetLayeredSeams(t)
	e := newLayeredEnv(t)
	before := readStr(t, e.md)
	layeredSnapshot = func(string, []string, map[string]bool) (backup.Manifest, error) {
		return backup.Manifest{}, errors.New("disk full")
	}
	if _, err := InstallLayered(e.install()); err == nil {
		t.Fatal("expected an error when the snapshot fails")
	}
	if readStr(t, e.md) != before {
		t.Error("CLAUDE.md changed despite snapshot failure")
	}
	if _, err := os.Stat(e.skills); err == nil {
		t.Error("skills dir was created despite snapshot failure")
	}
}

// 5.4: failure while activating the skill leaves CLAUDE.md byte-identical.
func TestInstallLayered_SkillFailureKeepsInstructions(t *testing.T) {
	resetLayeredSeams(t)
	e := newLayeredEnv(t)
	before := readStr(t, e.md)
	layeredHooks.beforeSkill = func() error { return errors.New("boom activating skill") }
	if _, err := InstallLayered(e.install()); err == nil {
		t.Fatal("expected an error")
	}
	if readStr(t, e.md) != before {
		t.Error("CLAUDE.md must be byte-identical after a skill failure")
	}
	if _, err := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); err == nil {
		t.Error("skill dir must not exist")
	}
}

// 5.4: failure while writing CLAUDE.md (after the skill was activated) restores
// both files: a new skill is removed, a previous owned skill comes back.
func TestInstallLayered_InstructionsFailureRestoresSkillAndFile(t *testing.T) {
	for _, withPrevious := range []bool{false, true} {
		name := "no previous skill"
		if withPrevious {
			name = "previous owned skill"
		}
		t.Run(name, func(t *testing.T) {
			resetLayeredSeams(t)
			e := newLayeredEnv(t)
			if withPrevious {
				prev := e.install()
				prev.Skill.Body = "OLD-BODY\n"
				prev.Skill.Modules = map[string]string{"old-only": "OLD-MODULE\n"}
				if _, err := InstallLayered(prev); err != nil {
					t.Fatal(err)
				}
				// Back to a CLAUDE.md with the legacy block for the real run.
				legacy := filemerge.InjectMarkdownSection("# Mine\n\nuser text\n", "sdd-orchestrator", "LEGACY-ORCH\n")
				if err := os.WriteFile(e.md, []byte(legacy), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			beforeMD := readStr(t, e.md)
			beforeSkills := treeOf(e.skills)

			layeredHooks.beforeInstructions = func() error { return errors.New("boom writing CLAUDE.md") }
			if _, err := InstallLayered(e.install()); err == nil {
				t.Fatal("expected an error")
			}
			if readStr(t, e.md) != beforeMD {
				t.Error("CLAUDE.md must be byte-identical after rollback")
			}
			if after := treeOf(e.skills); !equalTrees(beforeSkills, after) {
				t.Errorf("skills dir not restored.\nbefore=%v\nafter=%v", beforeSkills, after)
			}
		})
	}
}

// 5.4 + 01's GuardRollback: if another process changed CLAUDE.md after the
// installer wrote it, the rollback keeps that change and names the path.
func TestInstallLayered_RollbackDoesNotOverwriteConcurrentEdit(t *testing.T) {
	resetLayeredSeams(t)
	e := newLayeredEnv(t)
	const other = "# EDITED BY ANOTHER PROCESS\n"
	layeredHooks.afterInstructions = func() error {
		if err := os.WriteFile(e.md, []byte(other), 0o644); err != nil {
			t.Fatal(err)
		}
		return errors.New("later step failed")
	}
	_, err := InstallLayered(e.install())
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := readStr(t, e.md); got != other {
		t.Errorf("rollback overwrote a concurrent edit; CLAUDE.md = %q", got)
	}
	if !strings.Contains(err.Error(), e.md) {
		t.Errorf("error must name %q, got: %v", e.md, err)
	}
	if _, serr := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); serr == nil {
		t.Error("the new skill dir must still be rolled back")
	}
}

// 5.6 (C8): in a Project target the manifest lives under <home>/.active-stack/backups,
// never under the project root.
func TestInstallLayered_ProjectManifestUnderHome(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	project := filepath.Join(root, "project")
	md := filepath.Join(project, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(md), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(md, []byte("# Mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := LayeredInstall{
		HomeDir: home, InstructionsPath: md,
		SkillsDir:        filepath.Join(project, ".claude", "skills"),
		Skill:            SkillSpec{ID: "sdd-orchestrator", Description: "d", Body: "B\n"},
		UniversalContent: "U\n",
	}
	res, err := InstallLayered(in)
	if err != nil {
		t.Fatalf("InstallLayered: %v", err)
	}
	backups := filepath.Join(home, ".active-stack", "backups")
	if !strings.HasPrefix(res.ManifestDir, backups+string(filepath.Separator)) {
		t.Errorf("ManifestDir = %q, want under %q", res.ManifestDir, backups)
	}
	if _, err := os.Stat(filepath.Join(res.ManifestDir, backup.ManifestFilename)); err != nil {
		t.Errorf("manifest file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".active-stack")); err == nil {
		t.Error("nothing may be written under <root>/.active-stack")
	}
}

// 5.4: when a later step fails after CLAUDE.md was purged and rewritten, and
// nobody else touched the file, it is restored byte for byte.
func TestInstallLayered_FailureAfterPurgeRestoresInstructions(t *testing.T) {
	resetLayeredSeams(t)
	e := newLayeredEnv(t)
	before := readStr(t, e.md)
	layeredHooks.afterInstructions = func() error {
		if strings.Contains(readStr(t, e.md), "LEGACY-ORCH") {
			t.Error("precondition: the legacy block must already be purged when the hook runs")
		}
		return errors.New("later step failed")
	}
	if _, err := InstallLayered(e.install()); err == nil {
		t.Fatal("expected an error")
	}
	if got := readStr(t, e.md); got != before {
		t.Errorf("CLAUDE.md not restored byte for byte.\nwant %q\ngot  %q", before, got)
	}
	if _, err := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); err == nil {
		t.Error("new skill dir must be removed by the rollback")
	}
}

func modeOf(t *testing.T, p string) os.FileMode {
	t.Helper()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

// 5.5 (G6, small scope): CLAUDE.md, ~/.claude and ~/.claude/skills keep their mode.
func TestInstallLayered_PreservesModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful on Windows")
	}
	e := newLayeredEnv(t)
	claudeDir := filepath.Dir(e.md)
	if err := os.MkdirAll(e.skills, 0o700); err != nil {
		t.Fatal(err)
	}
	for p, m := range map[string]os.FileMode{e.md: 0o600, e.skills: 0o700, claudeDir: 0o700} {
		if err := os.Chmod(p, m); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := InstallLayered(e.install()); err != nil {
		t.Fatalf("InstallLayered: %v", err)
	}
	for p, want := range map[string]os.FileMode{e.md: 0o600, e.skills: 0o700, claudeDir: 0o700} {
		if got := modeOf(t, p); got != want {
			t.Errorf("%s mode = %o, want %o", p, got, want)
		}
	}
}

// 5.5: a directory left read-only by an old installer (0500) only gains the
// owner write bit and the full write succeeds.
func TestInstallLayered_ReadOnlyDirsGetOnlyOwnerWrite(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions enforced for a non-root user")
	}
	e := newLayeredEnv(t)
	claudeDir := filepath.Dir(e.md)
	if err := os.MkdirAll(e.skills, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{e.skills, claudeDir} {
		if err := os.Chmod(d, 0o500); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = os.Chmod(claudeDir, 0o700); _ = os.Chmod(e.skills, 0o700) })
	if _, err := InstallLayered(e.install()); err != nil {
		t.Fatalf("InstallLayered on read-only dirs: %v", err)
	}
	if !strings.Contains(readStr(t, e.md), UniversalSectionID) {
		t.Error("CLAUDE.md was not written")
	}
	for _, d := range []string{e.skills, claudeDir} {
		if got := modeOf(t, d); got != 0o700 {
			t.Errorf("%s mode = %o, want 700 (only the owner write bit added)", d, got)
		}
	}
}
