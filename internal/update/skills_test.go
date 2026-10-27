package update_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/update"
)

func TestScanInstalledSkillsFindsOnlyManagedExistingSkills(t *testing.T) {
	home := t.TempDir()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, agent := range []model.Agent{model.AgentCodex, model.AgentOpenCode} {
		adapter, _ := reg.Get(agent)
		path := filepath.Join(adapter.SkillsDir(home), "design-system")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("# old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// An unmanaged user skill is never selected for update.
	adapter, _ := reg.Get(model.AgentCodex)
	unknown := filepath.Join(adapter.SkillsDir(home), "user-authored")
	if err := os.MkdirAll(unknown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unknown, "SKILL.md"), []byte("# custom"), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := update.ScanInstalledSkills(cat, reg, home)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("targets = %+v, want two installed managed skills", targets)
	}
	for _, target := range targets {
		if target.Harness.ID != "design-system" {
			t.Fatalf("unexpected harness %q", target.Harness.ID)
		}
	}
}

func TestRefreshInstalledSkillsUpdatesOnlyScannedTargets(t *testing.T) {
	home := t.TempDir()
	backupDir := filepath.Join(home, "backups")
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	h, ok := cat.ByID("design-system")
	if !ok {
		t.Fatal("design-system missing")
	}
	adapter, _ := reg.Get(model.AgentCodex)
	path := filepath.Join(adapter.SkillsDir(home), h.ID)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	userPath := filepath.Join(adapter.SkillsDir(home), "personal")
	if err := os.MkdirAll(userPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userPath, "SKILL.md"), []byte("untouched"), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := update.ScanInstalledSkills(cat, reg, home)
	if err != nil {
		t.Fatal(err)
	}
	fs := fstest.MapFS{
		"skills/design-system/SKILL.md":        &fstest.MapFile{Data: []byte("new")},
		"skills/design-system/assets/logo.svg": &fstest.MapFile{Data: []byte("<svg/>")},
	}
	outcomes := update.RefreshInstalledSkills(context.Background(), targets, reg, home, backupDir, nil, fs)
	if len(outcomes) != 1 || outcomes[0].Err != nil || outcomes[0].Result.AlreadyInstalled {
		t.Fatalf("outcomes = %+v", outcomes)
	}
	got, err := os.ReadFile(filepath.Join(path, "SKILL.md"))
	if err != nil || string(got) != "new" {
		t.Fatalf("skill = %q, %v", got, err)
	}
	got, err = os.ReadFile(filepath.Join(userPath, "SKILL.md"))
	if err != nil || string(got) != "untouched" {
		t.Fatalf("personal = %q, %v", got, err)
	}
}

func TestScanProjectSkillsUsesAdapterProjectLayout(t *testing.T) {
	project := t.TempDir()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := reg.Get(model.AgentCodex)
	path := filepath.Join(adapter.PathsFor(project, model.Project).SkillsDir, "tailwind-design-system")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	targets, err := update.ScanProjectSkills(cat, reg, project)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].Harness.ID != "tailwind-design-system" || targets[0].ProjectRoot != project {
		t.Fatalf("targets=%+v", targets)
	}
}
