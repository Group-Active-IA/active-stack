package install_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestCustomPickerHarnesses_DesignMockupsOnlyForClaude verifies, against the
// real embedded catalog, that the opt-in design-mockups harness is pickable
// when claude is selected and hidden when only opencode/codex are selected.
func TestCustomPickerHarnesses_DesignMockupsOnlyForClaude(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	cases := []struct {
		name   string
		agents []model.Agent
		want   bool
	}{
		{"claude offers it", []model.Agent{model.AgentClaude}, true},
		{"claude+opencode offers it", []model.Agent{model.AgentClaude, model.AgentOpenCode}, true},
		{"opencode hides it", []model.Agent{model.AgentOpenCode}, false},
		{"codex hides it", []model.Agent{model.AgentCodex}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var found bool
			for _, h := range install.CustomPickerHarnesses(cat, tc.agents) {
				if h.ID == "design-mockups" {
					found = true
				}
			}
			if found != tc.want {
				t.Errorf("design-mockups in picker = %v, want %v", found, tc.want)
			}
		})
	}
}

// TestBuildPlanCustom_DesignMockupsPullsOrchestrator verifies that selecting
// only design-mockups in Custom mode resolves sdd-orchestrator as an
// auto-dependency ordered before it.
func TestBuildPlanCustom_DesignMockupsPullsOrchestrator(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"design-mockups"},
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	orch := indexOf(ids, "sdd-orchestrator")
	dm := indexOf(ids, "design-mockups")
	if orch == -1 || dm == -1 {
		t.Fatalf("plan must contain sdd-orchestrator and design-mockups, got %v", ids)
	}
	if orch >= dm {
		t.Errorf("sdd-orchestrator must precede design-mockups, got %v", ids)
	}
}
