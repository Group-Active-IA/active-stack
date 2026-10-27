package skill_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestCloneRefreshesChangedAssetEvenWhenSkillMDIsUnchanged(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	backupDir := t.TempDir()
	assetVersion := "first"
	runner := &stubRunner{sideEffect: func(args []string) {
		src := args[len(args)-1]
		_ = os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("# unchanged"), 0o644)
		_ = os.MkdirAll(filepath.Join(src, "references"), 0o755)
		_ = os.WriteFile(filepath.Join(src, "references", "guide.md"), []byte(assetVersion), 0o644)
	}}
	h := makeHarness("refreshable", "clone", "example/refreshable", false)
	adapter := fakeAdapter{agent: model.AgentCodex, skillsDir: skillsDir}
	installer := skill.NewInstaller(runner, nil)
	if _, err := installer.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, backupDir); err != nil {
		t.Fatal(err)
	}
	assetVersion = "second"
	results, err := installer.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, backupDir)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].AlreadyInstalled {
		t.Fatal("changed reference must update the installed skill")
	}
	installed, err := os.ReadFile(filepath.Join(skillsDir, "refreshable", "references", "guide.md"))
	if err != nil || string(installed) != "second" {
		t.Fatalf("installed asset = %q, err %v", installed, err)
	}
	if entries, _ := os.ReadDir(backupDir); len(entries) == 0 {
		t.Fatal("changed skill must be backed up")
	}
}

func TestCloneRefreshRemovesFilesDeletedUpstream(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	includeOld := true
	runner := &stubRunner{sideEffect: func(args []string) {
		src := args[len(args)-1]
		_ = os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("# unchanged"), 0o644)
		if includeOld {
			_ = os.WriteFile(filepath.Join(src, "obsolete.md"), []byte("old"), 0o644)
		}
	}}
	h := makeHarness("refreshable", "clone", "example/refreshable", false)
	adapter := fakeAdapter{agent: model.AgentCodex, skillsDir: skillsDir}
	installer := skill.NewInstaller(runner, nil)
	if _, err := installer.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	includeOld = false
	results, err := installer.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if results[0].AlreadyInstalled {
		t.Fatal("a deleted upstream file must change the installed skill")
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "refreshable", "obsolete.md")); !os.IsNotExist(err) {
		t.Fatalf("obsolete file remains: %v", err)
	}
}

func TestEmbeddedRefreshRemovesStaleManagedFiles(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	dest := filepath.Join(skillsDir, "design-system")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("# same"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "obsolete.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := fstest.MapFS{"skills/design-system/SKILL.md": &fstest.MapFile{Data: []byte("# same")}}
	h := makeHarness("design-system", "embed", "", false)
	adapter := fakeAdapter{agent: model.AgentCodex, skillsDir: skillsDir}
	results, err := skill.NewInstaller(nil, fs).Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if results[0].AlreadyInstalled {
		t.Fatal("stale file must trigger refresh")
	}
	if _, err := os.Stat(filepath.Join(dest, "obsolete.md")); !os.IsNotExist(err) {
		t.Fatalf("obsolete file remains: %v", err)
	}
}

func TestEmbeddedRefreshDetectsRenamedEmptyDirectory(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	dest := filepath.Join(skillsDir, "design-system")
	if err := os.MkdirAll(filepath.Join(dest, "old-empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("# same"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := fstest.MapFS{
		"skills/design-system/SKILL.md":  &fstest.MapFile{Data: []byte("# same")},
		"skills/design-system/new-empty": &fstest.MapFile{Mode: os.ModeDir},
	}
	h := makeHarness("design-system", "embed", "", false)
	adapter := fakeAdapter{agent: model.AgentCodex, skillsDir: skillsDir}
	results, err := skill.NewInstaller(nil, fs).Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if results[0].AlreadyInstalled {
		t.Fatal("renamed empty directory was considered identical")
	}
}
