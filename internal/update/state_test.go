package update_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/update"
)

func TestSaveIntentMergesAgentsAndBacksUpPreviousState(t *testing.T) {
	home := t.TempDir()
	if err := update.SaveIntent(home, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeFull}); err != nil {
		t.Fatal(err)
	}
	if err := update.SaveIntent(home, install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeCustom, Custom: []string{"design-mockups"}}); err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Agents) != 2 || state.Agents[model.AgentCodex].Mode != model.ModeFull || state.Agents[model.AgentClaude].Mode != model.ModeCustom {
		t.Fatalf("state = %+v", state)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".active-stack", "backups", "state"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("backups=%v, err=%v", entries, err)
	}
}

func TestDropAgentsPreventsUninstalledAgentFromReapplying(t *testing.T) {
	home := t.TempDir()
	if err := update.SaveIntent(home, install.Intent{Agents: []model.Agent{model.AgentCodex, model.AgentClaude}, Mode: model.ModeFull}); err != nil {
		t.Fatal(err)
	}
	if err := update.DropAgents(home, []model.Agent{model.AgentCodex}); err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := state.Agents[model.AgentCodex]; ok {
		t.Fatal("uninstalled codex still recorded")
	}
	if _, ok := state.Agents[model.AgentClaude]; !ok {
		t.Fatal("unrelated claude removed")
	}
}

func TestRecordProjectKeepsDistinctRoots(t *testing.T) {
	home := t.TempDir()
	a, b := t.TempDir(), t.TempDir()
	if err := update.RecordProject(home, a); err != nil {
		t.Fatal(err)
	}
	if err := update.RecordProject(home, b); err != nil {
		t.Fatal(err)
	}
	if err := update.RecordProject(home, a); err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Projects) != 2 {
		t.Fatalf("projects=%v", state.Projects)
	}
}

func TestSaveIntentAccumulatesCustomHarnessesWithPreset(t *testing.T) {
	home := t.TempDir()
	agent := model.AgentCodex
	for _, intent := range []install.Intent{
		{Agents: []model.Agent{agent}, Mode: model.ModeFull},
		{Agents: []model.Agent{agent}, Mode: model.ModeCustom, Custom: []string{"design-mockups"}},
		{Agents: []model.Agent{agent}, Mode: model.ModeCustom, Custom: []string{"extra", "design-mockups"}},
	} {
		if err := update.SaveIntent(home, intent); err != nil {
			t.Fatal(err)
		}
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	selection := state.Agents[agent]
	if selection.Mode != model.ModeFull || len(selection.Custom) != 2 {
		t.Fatalf("selection=%+v", selection)
	}
}

func TestSaveIntentRecordsOnlySelectedExistingSkills(t *testing.T) {
	home := t.TempDir()
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := reg.Get(model.AgentCodex)
	managed := filepath.Join(adapter.SkillsDir(home), "design-system")
	foreign := filepath.Join(adapter.SkillsDir(home), "find-skills")
	for _, path := range []string{managed, foreign} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := update.SaveIntent(home, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeCustom, Custom: []string{"design-system"}}); err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	paths := state.ManagedSkills[model.AgentCodex]
	if len(paths) != 1 || paths[0] != managed {
		t.Fatalf("managed paths=%v", paths)
	}
}
