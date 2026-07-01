package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type codexVerifyAdapter struct{ root string }

func (a codexVerifyAdapter) Agent() model.Agent { return model.AgentCodex }
func (a codexVerifyAdapter) SkillsDir(string) string {
	return filepath.Join(a.root, ".agents", "skills")
}
func (a codexVerifyAdapter) InstructionsPath(string) string {
	return filepath.Join(a.root, ".codex", "AGENTS.md")
}
func (a codexVerifyAdapter) SettingsPath(string) string {
	return filepath.Join(a.root, ".codex", "config.toml")
}
func (a codexVerifyAdapter) MCPConfigPath(string, string) string { return a.SettingsPath("") }
func (a codexVerifyAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}

func TestCodexPermissionsAndMCPChecks(t *testing.T) {
	root := t.TempDir()
	a := codexVerifyAdapter{root: root}
	if err := os.MkdirAll(filepath.Dir(a.SettingsPath("")), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "approval_policy = \"on-request\"\nsandbox_mode = \"workspace-write\"\n\n[mcp_servers.context7]\ncommand = \"npx\"\n"
	if err := os.WriteFile(a.SettingsPath(""), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	permissionsHarness := model.Harness{ID: "permissions", Type: model.HarnessConfig}
	mcpHarness := model.Harness{ID: "context7", Type: model.HarnessExternal, External: &model.External{Method: "mcp"}}
	checks := append(ChecksForHarness(permissionsHarness, []Adapter{a}, root), ChecksForHarness(mcpHarness, []Adapter{a}, root)...)
	for _, result := range RunChecks(context.Background(), checks) {
		if result.Status == CheckStatusFailed {
			t.Errorf("%s failed: %s", result.ID, result.Error)
		}
	}
}
