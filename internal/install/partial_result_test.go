package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

type blockedLegacyAdapter struct{ AgentAdapter }

func (a blockedLegacyAdapter) Agent() model.Agent { return model.AgentOpenCode }
func (a blockedLegacyAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a blockedLegacyAdapter) VariantKey() string { return "opencode" }
func (a blockedLegacyAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, "blocked", "AGENTS.md")
}

func TestConfigStepFailurePreservesNativeUserEditOnRollback(t *testing.T) {
	home := t.TempDir()
	a := codex.NewAdapter()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	paths, err := config.ManagedSkillOperationPaths(h, a, home)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := backup.NewSnapshotter().Create(filepath.Join(home, "snapshot"), paths)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(home, "blocked"), []byte("user file"), 0600); err != nil {
		t.Fatal(err)
	}
	step := configStep{h: h, adapters: []AgentAdapter{a, blockedLegacyAdapter{AgentAdapter: a}}, homeDir: home, manifest: &manifest}
	if err = step.Run(); err == nil {
		t.Fatal("expected later adapter failure")
	}
	path := a.InstructionsPath(home)
	if err = os.WriteFile(path, []byte("user edit after native install"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = step.Rollback(); err == nil {
		t.Fatal("expected preserved-user-edit conflict")
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "user edit after native install" {
		t.Fatalf("user edit lost during rollback: %q, %v", raw, err)
	}
}
