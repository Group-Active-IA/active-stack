package config

import (
	"strings"
	"testing"
)

func TestUniversalWorkflowGuidance(t *testing.T) {
	claude, err := ComposeLayers(nil)
	if err != nil {
		t.Fatal(err)
	}
	codex, err := ComposeCodexPackage(nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, client := range []struct {
		name, universal, invocation, otherInvocation string
	}{
		{"Claude", claude.Universal, "/sdd-orchestrator", "$sdd-orchestrator"},
		{"Codex", codex.Universal, "$sdd-orchestrator", "/sdd-orchestrator"},
	} {
		t.Run(client.name, func(t *testing.T) {
			for _, want := range []string{"Handle routine tasks directly", "proposal, specs, design, and tasks", "installed skill", client.invocation} {
				if !strings.Contains(client.universal, want) {
					t.Errorf("missing %q in universal instructions", want)
				}
			}
			if strings.Contains(client.universal, client.otherInvocation) {
				t.Errorf("found another client's invocation")
			}
		})
	}
}

func TestEvidenceContractDelivery(t *testing.T) {
	claude, err := ComposeLayers(nil)
	if err != nil {
		t.Fatal(err)
	}
	codex, err := ComposeCodexPackage(nil)
	if err != nil {
		t.Fatal(err)
	}
	contract := claude.Modules["evidence"]
	if contract == "" || contract != codex.Files["modules/evidence.md"] {
		t.Fatal("Claude and Codex must install the same evidence contract")
	}
	for _, text := range []string{"### Review step", "Direct execution:", "Delegated execution:", "independent source"} {
		if !strings.Contains(contract, text) {
			t.Errorf("contract missing %q", text)
		}
	}
	for _, skill := range []string{RenderSkillMD(claude.Skill), codex.Files["SKILL.md"]} {
		if !strings.Contains(skill, "modules/evidence.md") {
			t.Error("coordinator must load the contract")
		}
	}
}

func TestClaudeVerifyInstallsCaptureModule(t *testing.T) {
	layers, err := ComposeLayers(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(layers.Skill.Body, "| verify |") || !strings.Contains(layers.Skill.Body, "modules/bounded-command-output.md") || layers.Modules["bounded-command-output"] == "" {
		t.Fatal("Claude verify must load the installed capture instructions without delegation")
	}
}
