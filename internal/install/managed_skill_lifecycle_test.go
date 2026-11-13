package install_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	codexadapter "github.com/Group-Active-IA/active-stack/internal/agents/codex"
	cfg "github.com/Group-Active-IA/active-stack/internal/harness/config"
	skill "github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

func TestNativeLateVerifyFailurePreservesUserNotes(t *testing.T) {
	home := t.TempDir()
	a := codexadapter.NewAdapter()
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	os.WriteFile(path, []byte("user rules"), 0600)
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite}}
	other := model.Harness{ID: "other-skill", Type: model.HarnessSkill, InstallModes: []model.InstallMode{model.ModeLite}}
	restoreSkill := install.SetSkillInstallFn(func(_ interface{}, _ fs.FS, _ context.Context, _ model.Harness, _ []skill.AgentAdapter, _, _ string) ([]skill.Result, error) {
		return nil, nil
	})
	defer restoreSkill()
	cat := &fakeCatalog{harnesses: []model.Harness{h, other}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentCodex: a}}
	note := filepath.Join(a.SkillsDir(home), h.ID, "user-note.md")
	hook := func() error {
		if e := cfg.VerifyManagedSkill(h, a, home); e != nil {
			t.Fatal(e)
		}
		os.WriteFile(note, []byte("user data after install"), 0600)
		return errors.New("late verify failure")
	}
	opts := buildOptions(home, reg, hook)
	opts.NoSelfInstall = true
	plan, e := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeLite}, opts)
	if e != nil {
		t.Fatal(e)
	}
	result := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
	if result.Err == nil {
		t.Fatal("verify failure ignored")
	}
	raw, e := os.ReadFile(note)
	if e != nil || string(raw) != "user data after install" {
		t.Fatalf("user note lost after rollback raw=%s err=%v", raw, e)
	}
	raw, e = os.ReadFile(path)
	if e != nil || string(raw) != "user rules" {
		t.Fatalf("instructions not restored: %s %v", raw, e)
	}
	if _, e = cfg.InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("unowned leftover directory should require explicit review")
	}
}

func TestNativeLateVerifyFailureCanRetry(t *testing.T) {
	home := t.TempDir()
	a := codexadapter.NewAdapter()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite}}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentCodex: a}}
	opts := buildOptions(home, reg, func() error { return errors.New("late verification failure") })
	opts.NoSelfInstall = true
	plan, e := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeLite}, opts)
	if e != nil {
		t.Fatal(e)
	}
	result := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
	if result.Err == nil {
		t.Fatal("expected failure")
	}
	if _, e = cfg.InstallManagedSkill(h, a, home); e != nil {
		t.Fatalf("retry failed: %v", e)
	}
}

func TestNativeRollbackPreservesReplacedExistingParent(t *testing.T) {
	home := t.TempDir()
	a := codexadapter.NewAdapter()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite}}
	if _, e := cfg.InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	h.Toggles = []string{"tdd"}
	root := filepath.Join(a.SkillsDir(home), h.ID)
	foreign := filepath.Join(home, "foreign")
	os.MkdirAll(filepath.Join(foreign, "modules"), 0700)
	var expected []byte
	hook := func() error {
		expected, _ = os.ReadFile(filepath.Join(root, "modules", "tdd.md"))
		os.WriteFile(filepath.Join(foreign, "modules", "tdd.md"), expected, 0600)
		if e := os.Rename(root, root+"-saved"); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(foreign, root); e != nil {
			t.Fatal(e)
		}
		return errors.New("late verify failed")
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentCodex: a}}
	opts := buildOptions(home, reg, hook)
	opts.NoSelfInstall = true
	plan, e := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeLite}, opts)
	if e != nil {
		t.Fatal(e)
	}
	result := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
	if result.Err == nil {
		t.Fatal("expected failure")
	}
	raw, e := os.ReadFile(filepath.Join(foreign, "modules", "tdd.md"))
	if e != nil || !bytes.Equal(raw, expected) {
		t.Fatalf("foreign file changed: %v", e)
	}
}
