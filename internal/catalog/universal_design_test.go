package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestUniversalDesignFullAndCustomScope(t *testing.T) {
	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"design-mockups", "playwright", "emil-design-eng", "design-taste-frontend-v1"} {
		h, ok := cat.ByID(id)
		if !ok {
			t.Fatalf("missing %s", id)
		}
		for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode} {
			if !h.SupportsAgent(agent) {
				t.Errorf("%s does not support %s", id, agent)
			}
		}
		if !h.InMode(model.ModeFull) || !h.InMode(model.ModeCustom) || h.InMode(model.ModeLite) {
			t.Errorf("%s should be in Full and Custom only", id)
		}
	}
}
