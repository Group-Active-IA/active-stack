package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestActiveOrchestratorFullUsesApprovedClaudeFix(t *testing.T) {
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
				if h.Source.Repo != "Group-Active-IA/active-orchestrator" || h.Source.Ref != "5d758fceb3b560b934c3348eaa20e944f79503a6" {
					t.Fatalf("Full source = %s@%s, want approved upstream merge", h.Source.Repo, h.Source.Ref)
				}
				return
			}
			t.Fatal("active-orchestrator missing from Full for agent")
		})
	}
}
