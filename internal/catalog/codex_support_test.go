package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestCatalogCodexCoreHarnesses(t *testing.T) {
	cat, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, id := range []string{"sdd-orchestrator", "permissions", "engram", "context7"} {
		h, ok := cat.ByID(id)
		if !ok {
			t.Fatalf("missing harness %q", id)
		}
		if !h.SupportsAgent(model.AgentCodex) {
			t.Errorf("%s does not support codex", id)
		}
	}

	if h, _ := cat.ByID("starter-add-command"); h.SupportsAgent(model.AgentCodex) {
		t.Error("starter-add-command must remain excluded for codex")
	}
}
