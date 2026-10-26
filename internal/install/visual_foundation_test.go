package install_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// A Full installation must deliver the actual foundation skill to every
// supported agent, rather than only mentioning it in an instruction fragment.
func TestFullInstallsVisualFoundationSkillForFourAgents(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	h, ok := cat.ByID("design-system")
	if !ok {
		t.Fatal("Full is missing the installable design-system skill")
	}
	if h.Type != model.HarnessSkill || h.Source == nil || h.Source.Method != "embed" {
		t.Fatalf("design-system source = %+v, want embedded skill", h.Source)
	}
	orchestrator, ok := cat.ByID("active-orchestrator")
	if !ok || orchestrator.Source == nil || len(orchestrator.Source.Ref) != 40 {
		t.Fatalf("active-orchestrator must use a pinned visual-foundation version: %+v", orchestrator.Source)
	}

	for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode, model.AgentAntigravity} {
		t.Run(string(agent), func(t *testing.T) {
			selected, err := install.SelectHarnesses(cat, install.Intent{Agents: []model.Agent{agent}, Mode: model.ModeFull})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			foundOrchestrator := false
			for _, candidate := range selected {
				found = found || candidate.ID == h.ID
				foundOrchestrator = foundOrchestrator || candidate.ID == orchestrator.ID
			}
			if !found || !foundOrchestrator {
				t.Fatalf("Full %s does not select both design-system and active-orchestrator", agent)
			}
			adapter, err := agents.NewAdapter(agent)
			if err != nil {
				t.Fatal(err)
			}
			home := t.TempDir()
			results, err := skill.NewInstaller(nil, assets.SkillsFS).Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Fatalf("results = %+v, want one installed skill", results)
			}
			path := filepath.Join(adapter.SkillsDir(home), "design-system", "SKILL.md")
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(content), "design-system/README.md") {
				t.Fatalf("installed skill at %s lacks the foundation workflow", path)
			}
		})
	}
}

func TestCustomMockupsPullsInDesignSystemSkill(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeCustom, Custom: []string{"design-mockups"},
	}, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatal(err)
	}
	ids := applyStepIDs(plan.StagePlan)
	if indexOf(ids, "design-system") < 0 || indexOf(ids, "design-mockups") < 0 {
		t.Fatalf("Custom mockups must install the foundation skill: %v", ids)
	}
}

func TestCustomOrchestratorPullsInDesignSystemSkill(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentAntigravity: fakeAdapter{agent: model.AgentAntigravity},
	}}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentAntigravity}, Mode: model.ModeCustom, Custom: []string{"active-orchestrator"},
	}, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatal(err)
	}
	ids := applyStepIDs(plan.StagePlan)
	if indexOf(ids, "design-system") < 0 || indexOf(ids, "active-orchestrator") < 0 {
		t.Fatalf("Custom orchestrator must install the foundation skill: %v", ids)
	}
}
