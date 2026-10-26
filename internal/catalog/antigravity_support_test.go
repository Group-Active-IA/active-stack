package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestCatalogAntigravityCoreHarnesses(t *testing.T) {
	cat, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, id := range []string{"openspec", "engram", "context7", "sdd-orchestrator", "active-orchestrator"} {
		h, ok := cat.ByID(id)
		if !ok {
			t.Fatalf("missing harness %q", id)
		}
		if !h.SupportsAgent(model.AgentAntigravity) {
			t.Errorf("%s does not support antigravity", id)
		}
	}

	if h, _ := cat.ByID("starter-add-command"); h.SupportsAgent(model.AgentAntigravity) {
		t.Error("starter-add-command must remain excluded for antigravity")
	}
	if h, _ := cat.ByID("permissions"); h.SupportsAgent(model.AgentAntigravity) {
		t.Error("permissions must remain excluded for antigravity (no-op overlay)")
	}
}
