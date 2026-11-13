package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type diagnosticAdapter struct {
	agent    model.Agent
	target   model.InstallTarget
	delivery model.ConfigDelivery
}

func (a diagnosticAdapter) Agent() model.Agent                   { return a.agent }
func (a diagnosticAdapter) InstallTarget() model.InstallTarget   { return a.target }
func (a diagnosticAdapter) ConfigDelivery() model.ConfigDelivery { return a.delivery }
func (a diagnosticAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, "AGENTS.md")
}
func (a diagnosticAdapter) SettingsPath(string) string { return "" }
func (a diagnosticAdapter) VariantKey() string         { return string(a.agent) }

func TestCodexProjectLegacyDiagnostic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		agent   model.Agent
		target  model.InstallTarget
		warning bool
	}{
		{"project-codex", model.AgentCodex, model.Project, true},
		{"machine-codex", model.AgentCodex, model.Machine, false},
		{"project-claude", model.AgentClaude, model.Project, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := diagnosticAdapter{agent: tc.agent, target: tc.target}
			result, err := Install(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, []AgentAdapter{a}, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			warnings := result.Warnings
			if tc.warning && len(warnings) != 1 {
				t.Fatal("Codex Project legacy delivery diagnostic not emitted")
			}
			if !tc.warning && len(warnings) != 0 {
				t.Fatal("unexpected legacy diagnostic")
			}
			if tc.warning && (!strings.Contains(warnings[0], "Project") || !strings.Contains(warnings[0], "legacy")) {
				t.Fatal("diagnostic must explain Project legacy delivery")
			}
		})
	}
}
func TestManagedDeliveryRequiresSkillsCapability(t *testing.T) {
	a := diagnosticAdapter{agent: model.AgentCodex, target: model.Machine, delivery: model.ConfigDeliveryManagedSkill}
	_, err := Install(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, []AgentAdapter{a}, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "does not expose skill paths") {
		t.Fatalf("expected skill path capability error, got %v", err)
	}
}

func TestNativeMachineHasNoLegacyDiagnostic(t *testing.T) {
	result, err := Install(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, []AgentAdapter{managedTestAdapter{}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("native Machine warnings: %v", result.Warnings)
	}
}

func TestProjectReinstallKeepsDiagnosticAndLegacyDelivery(t *testing.T) {
	home := t.TempDir()
	a := diagnosticAdapter{agent: model.AgentCodex, target: model.Project}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := Install(h, []AgentAdapter{a}, home); err != nil {
		t.Fatal(err)
	}
	result, err := Install(h, []AgentAdapter{a}, home)
	if err != nil {
		t.Fatal(err)
	}
	if !result.AllAlready || len(result.Warnings) != 1 {
		t.Fatalf("reinstall = %+v", result)
	}
	content, err := os.ReadFile(a.InstructionsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "<!-- active-stack:sdd-orchestrator -->") != 1 {
		t.Fatal("Project must retain one legacy coordinator block")
	}
}
