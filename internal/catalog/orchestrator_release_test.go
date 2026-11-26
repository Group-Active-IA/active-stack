package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestActiveOrchestratorFullUsesRequiredTrackingMerge(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode, model.AgentAntigravity} {
		t.Run(string(agent), func(t *testing.T) {
			for _, h := range c.ForAgent(agent) {
				if h.ID != "active-orchestrator" || !h.InMode(model.ModeFull) {
					continue
				}
				if h.Source == nil {
					t.Fatal("active-orchestrator has no source")
				}
				if h.Source.Repo != "Group-Active-IA/active-orchestrator" || h.Source.Ref != "a3d6b0490940a2865beb16d3baf07a5a292eeafb" {
					t.Fatalf("Full source = %s@%s, want approved upstream merge", h.Source.Repo, h.Source.Ref)
				}
				return
			}
			t.Fatal("active-orchestrator missing from Full for agent")
		})
	}
}
