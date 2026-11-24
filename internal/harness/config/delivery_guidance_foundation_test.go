package config

import (
	"os"
	"strings"
	"testing"
)

// The project foundation (active-orchestrator phases such as the Knowledge Base)
// is not a delivery unit: the delivery guidance must not make the agent ask for
// ODD or SDD while it runs.
const foundationExclusion = "Phases of the project foundation run by `active-orchestrator` are not delivery units: do not ask for ODD or SDD for them."

func TestDeliveryGuidanceExcludesFoundationPhases(t *testing.T) {
	for name, text := range map[string]string{
		"claude": claudeDeliveryGuidance,
		"codex":  codexDeliveryGuidance,
	} {
		if !strings.Contains(text, foundationExclusion) {
			t.Errorf("%s delivery guidance lacks the foundation exclusion:\n%s", name, text)
		}
	}
}

func TestCodexOrchestratorAssetKeepsTheFoundationExclusion(t *testing.T) {
	b, err := os.ReadFile("assets/codex/sdd-orchestrator.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), codexDeliveryGuidance) {
		t.Errorf("assets/codex/sdd-orchestrator.md does not carry codexDeliveryGuidance")
	}
}
