package install_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// TestProjectRollback_NewDirRemovedOnRollback verifies that when the install
// creates a project skills dir that did not exist before, rollback removes it
// (RemoveAll) — satisfying the "dir created by install" guarantee (D4/R1).
func TestProjectRollback_NewDirRemovedOnRollback(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()

	skillsDir := filepath.Join(projectRoot, ".claude", "skills")

	// Precondition: skills dir must not exist yet.
	if _, err := os.Stat(skillsDir); !os.IsNotExist(err) {
		t.Fatalf("precondition: skills dir %q should not exist yet", skillsDir)
	}

	failingStep := false

	h1 := model.Harness{
		ID:           "skill-ok",
		Name:         "Skill OK",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/repo", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
		// No DependsOn: h1 runs first, h2 second (catalog order → topo order with no deps).
	}
	h2 := model.Harness{
		ID:           "skill-fail",
		Name:         "Skill Fail",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/repo2", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
		DependsOn:    []string{"skill-ok"}, // ensures skill-ok runs first
	}

	// Wire skill install: h1 creates the dir, h2 fails → rollback removes the dir.
	restoreSkill := install.SetSkillInstallFn(func(_ interface{}, _ fs.FS, _ context.Context, h model.Harness, _ []skill.AgentAdapter, _, _ string) ([]skill.Result, error) {
		if h.ID == "skill-ok" {
			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return nil, err
			}
			// Create a file inside to verify RemoveAll handles non-empty dirs.
			_ = os.WriteFile(filepath.Join(skillsDir, "dummy.md"), []byte("test"), 0o644)
			return nil, nil
		}
		failingStep = true
		return nil, errors.New("simulated skill install failure")
	})
	defer restoreSkill()

	cat := &fakeCatalog{harnesses: []model.Harness{h1, h2}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: projectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     homeDir,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)

	if !failingStep {
		t.Error("expected skill-fail step to run and fail; it did not")
	}
	if result.Err == nil {
		t.Fatal("expected orchestrator to return error from failed step")
	}

	// After rollback: the dir created by skill-ok must have been removed.
	if _, err := os.Stat(skillsDir); !os.IsNotExist(err) {
		t.Errorf("after rollback: skills dir %q should have been removed (RemoveAll), but it still exists", skillsDir)
	}
}

// TestProjectRollback_PreexistingDirNotDeletedOnRollback verifies that when a
// project config dir already existed before the install and rollback runs, that
// directory is NOT deleted (R1/D4 — NO-OP for preexisting dirs).
func TestProjectRollback_PreexistingDirNotDeletedOnRollback(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()

	// Create the skills dir BEFORE the install (preexisting).
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("setup: create skills dir: %v", err)
	}
	// Put a file in it to confirm it's not empty.
	existingFile := filepath.Join(skillsDir, "existing-skill.md")
	if err := os.WriteFile(existingFile, []byte("preexisting"), 0o644); err != nil {
		t.Fatalf("setup: write existing file: %v", err)
	}

	// Wire a config harness that fails → triggers rollback.
	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return errors.New("simulated config install failure")
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: projectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     homeDir,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)

	if result.Err == nil {
		t.Fatal("expected error from failed config step")
	}

	// After rollback: the preexisting skills dir must still exist.
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Errorf("after rollback: preexisting skills dir %q was deleted; it must NOT be deleted", skillsDir)
	}

	// The existing file inside the dir must still be there.
	if _, err := os.Stat(existingFile); os.IsNotExist(err) {
		t.Errorf("after rollback: preexisting file %q was deleted; it must NOT be deleted", existingFile)
	}
}
