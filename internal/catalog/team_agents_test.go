package catalog

import (
	"reflect"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// [EQ] Fase 7 (R-20) T-F7-1: active-delivery y active-roadmap conservan sus
// agentes y modos.
func TestActiveDeliveryAndRoadmapKeepAgentsAndModes(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string][]model.InstallMode{
		"active-delivery": {model.ModeLite, model.ModeFull},
		"active-roadmap":  {model.ModeFull},
	}
	for id, modes := range cases {
		h, ok := c.ByID(id)
		if !ok {
			t.Fatalf("%s missing", id)
		}
		if want := []model.Agent{model.AgentCodex, model.AgentClaude}; !reflect.DeepEqual(h.Agents, want) {
			t.Errorf("%s agents = %v, want %v", id, h.Agents, want)
		}
		if !reflect.DeepEqual(h.InstallModes, modes) {
			t.Errorf("%s modes = %v, want %v", id, h.InstallModes, modes)
		}
	}
}
