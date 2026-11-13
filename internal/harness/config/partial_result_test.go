package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type failingReviewAdapter struct {
	managedTestAdapter
	delivery model.ConfigDelivery
}

func (a failingReviewAdapter) Agent() model.Agent                   { return model.AgentClaude }
func (a failingReviewAdapter) ConfigDelivery() model.ConfigDelivery { return a.delivery }
func (a failingReviewAdapter) VariantKey() string                   { return "claude" }
func (a failingReviewAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, "blocked", "CLAUDE.md")
}
func (a failingReviewAdapter) SettingsPath(home string) string {
	return filepath.Join(home, "blocked", "settings.json")
}

func TestInstallRetainsNativeResultWhenLaterAdapterFails(t *testing.T) {
	for _, tc := range []struct {
		name    string
		adapter AgentAdapter
	}{
		{"instructions-injection", failingReviewAdapter{delivery: model.ConfigDeliveryInstructions}},
		{"primary-agent-injection", failingReviewAdapter{delivery: model.ConfigDeliveryPrimaryAgent}},
		{"missing-skill-capability", diagnosticAdapter{agent: model.AgentClaude, delivery: model.ConfigDeliveryManagedSkill}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if err := os.WriteFile(filepath.Join(home, "blocked"), []byte("user"), 0600); err != nil {
				t.Fatal(err)
			}
			a := managedTestAdapter{}
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
			result, err := Install(h, []AgentAdapter{a, tc.adapter}, home)
			if err == nil {
				t.Fatal("expected later adapter failure")
			}
			if _, err := os.Stat(filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")); err != nil {
				t.Fatalf("native install did not happen: %v", err)
			}
			if len(result.ManagedWrites) == 0 || len(result.Files) == 0 {
				t.Fatalf("lost prior native writes on later adapter failure: %+v", result)
			}
		})
	}
}
