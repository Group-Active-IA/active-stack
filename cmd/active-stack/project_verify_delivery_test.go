package main

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

type verifyDeliveryAdapter struct {
	agents.Adapter
	agent    model.Agent
	delivery model.ConfigDelivery
}

func (a verifyDeliveryAdapter) Agent() model.Agent                   { return a.agent }
func (a verifyDeliveryAdapter) ConfigDelivery() model.ConfigDelivery { return a.delivery }

func TestProjectVerifyCodexKeepsInstructionsDelivery(t *testing.T) {
	adapter := verifyDeliveryAdapter{agent: model.AgentCodex, delivery: model.ConfigDeliveryManagedSkill}
	wrapped := projectVerifyAdapter{Adapter: adapter, base: t.TempDir()}
	if got := wrapped.ConfigDelivery(); got != model.ConfigDeliveryInstructions {
		t.Fatalf("Project Codex delivery = %v, want instructions", got)
	}
}

func TestVerifyDeliveryPreservesOtherTargetsAndAgents(t *testing.T) {
	cases := []struct {
		name     string
		agent    model.Agent
		delivery model.ConfigDelivery
		target   model.InstallTarget
	}{
		{"OpenCode Project", model.AgentOpenCode, model.ConfigDeliveryPrimaryAgent, model.Project},
		{"Codex Machine", model.AgentCodex, model.ConfigDeliveryManagedSkill, model.Machine},
		{"legacy Codex Project", model.AgentCodex, model.ConfigDeliveryInstructions, model.Project},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			registry := agents.NewRegistry()
			adapter := verifyDeliveryAdapter{agent: tc.agent, delivery: tc.delivery}
			if err := registry.Register(adapter); err != nil {
				t.Fatal(err)
			}
			resolved := resolveVerifyAdaptersForTarget([]model.Agent{tc.agent}, registry, t.TempDir(), tc.target)
			if len(resolved) != 1 || resolved[0].ConfigDelivery() != tc.delivery {
				t.Fatalf("delivery for %s changed: %v", tc.name, resolved)
			}
		})
	}
}
