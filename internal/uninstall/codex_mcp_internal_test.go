package uninstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type codexUninstallAdapter struct{}

func (codexUninstallAdapter) Agent() model.Agent { return model.AgentCodex }
func (codexUninstallAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, ".codex", "AGENTS.md")
}
func (codexUninstallAdapter) SkillsDir(home string) string {
	return filepath.Join(home, ".agents", "skills")
}
func (codexUninstallAdapter) SettingsPath(home string) string {
	return filepath.Join(home, ".codex", "config.toml")
}
func (codexUninstallAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (codexUninstallAdapter) CommandsDir(string) string { return "" }
func (codexUninstallAdapter) VariantKey() string        { return "codex" }

func TestCodexMCPRemovalStepPreservesUnrelatedConfig(t *testing.T) {
	home := t.TempDir()
	path := codexUninstallAdapter{}.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	input := "model = \"keep\"\n\n[mcp_servers.engram]\ncommand = \"engram\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	h := model.Harness{
		ID:   "engram",
		Type: model.HarnessExternal,
		External: &model.External{
			MCP: &model.MCP{Name: "engram"},
		},
	}
	step := &codexMCPRemovalStep{h: h, adapters: []AgentAdapter{codexUninstallAdapter{}}, homeDir: home}
	if err := step.Run(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	got := string(raw)
	if strings.Contains(got, "[mcp_servers.engram]") {
		t.Fatalf("managed MCP table remains:\n%s", got)
	}
	if !strings.Contains(got, `model = "keep"`) || !strings.Contains(got, "[mcp_servers.other]") {
		t.Fatalf("unrelated config lost:\n%s", got)
	}
}
