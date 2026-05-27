package uninstall_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ─────────────────────────────────────────────────────────────────
// skillRemovalStep tests
// ─────────────────────────────────────────────────────────────────

// TestSkillRemovalStepRemovesDirectory verifies that the step removes the
// skill directory under the adapter's SkillsDir.
func TestSkillRemovalStepRemovesDirectory(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	// Create the skill directory with some files.
	skillDir := filepath.Join(adapter.SkillsDir(homeDir), "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill content"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "my-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/my-skill", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill directory still exists after removal: %s", skillDir)
	}
}

// TestSkillRemovalStepMissingDirIsNoop verifies that when the skill directory
// does not exist the step completes without error.
func TestSkillRemovalStepMissingDirIsNoop(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	// Do NOT create the skills directory — it should be absent.
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "missing-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/missing-skill", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Errorf("step.Run() on missing dir returned error = %v", err)
		}
	}
}

// TestSkillRemovalStepPathResolvedViaAdapter verifies that the skill directory
// path comes from the adapter, not from a hardcoded literal.
func TestSkillRemovalStepPathResolvedViaAdapter(t *testing.T) {
	homeDir := t.TempDir()
	customSkillsDir := filepath.Join(homeDir, "custom-agent-skills")
	adapter := fakeAdapterCustomPath{
		agent:     model.AgentClaude,
		skillsDir: customSkillsDir,
	}

	// Create the skill directory at the custom path.
	skillDir := filepath.Join(customSkillsDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "my-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/my-skill", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistryCustom{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill dir still exists; adapter path was not used: %s", skillDir)
	}
}
