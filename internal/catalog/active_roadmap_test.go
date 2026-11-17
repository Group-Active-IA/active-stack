package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestActiveRoadmapCatalog(t *testing.T) {
	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	harness, ok := cat.ByID("active-roadmap")
	if !ok {
		t.Fatal("active-roadmap missing from catalog")
	}
	if !harness.InMode(model.ModeFull) || harness.InMode(model.ModeLite) {
		t.Fatal("active-roadmap must be included in Full and available separately in Custom")
	}
	if !harness.SupportsAgent(model.AgentCodex) || harness.SupportsAgent(model.AgentClaude) ||
		harness.SupportsAgent(model.AgentOpenCode) || harness.SupportsAgent(model.AgentAntigravity) {
		t.Fatal("active-roadmap must be scoped to Codex")
	}
}
