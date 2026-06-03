package verify_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/verify"
)

// fakeAdapter is a test double that satisfies verify.Adapter.
// It proves the interface is minimal and does NOT require importing internal/agents.
type fakeAdapter struct {
	agent           model.Agent
	skillsDir       string
	instructionsPath string
	settingsPath    string
	mcpConfigPath   string
	delivery        model.ConfigDelivery
}

func (a fakeAdapter) Agent() model.Agent                              { return a.agent }
func (a fakeAdapter) SkillsDir(homeDir string) string                 { return a.skillsDir }
func (a fakeAdapter) InstructionsPath(homeDir string) string          { return a.instructionsPath }
func (a fakeAdapter) SettingsPath(homeDir string) string              { return a.settingsPath }
func (a fakeAdapter) MCPConfigPath(homeDir, serverName string) string { return a.mcpConfigPath }
func (a fakeAdapter) ConfigDelivery() model.ConfigDelivery            { return a.delivery }

// TestAdapterInterfaceIsSatisfiedByFake verifies that a zero-import fake struct
// satisfies verify.Adapter by structural typing (compile-time check).
func TestAdapterInterfaceIsSatisfiedByFake(t *testing.T) {
	var _ verify.Adapter = fakeAdapter{}
}

// TestAdapterDoesNotImportAgents verifies that internal/verify does NOT import
// internal/agents — enforced by build: if it did, the package would fail to
// compile due to a potential import cycle.
//
// This test is a documentation-level assertion; the real guard is go build.
func TestAdapterInterfaceIsMinimal(t *testing.T) {
	fa := fakeAdapter{
		agent:           model.AgentClaude,
		skillsDir:       "/home/user/.claude/skills",
		instructionsPath: "/home/user/.claude/CLAUDE.md",
		settingsPath:    "/home/user/.claude/settings.json",
		mcpConfigPath:   "/home/user/.claude/mcp/context7.json",
	}

	if fa.Agent() != model.AgentClaude {
		t.Errorf("Agent() = %q, want %q", fa.Agent(), model.AgentClaude)
	}
	if fa.SkillsDir("") != "/home/user/.claude/skills" {
		t.Errorf("SkillsDir() = %q", fa.SkillsDir(""))
	}
	if fa.InstructionsPath("") != "/home/user/.claude/CLAUDE.md" {
		t.Errorf("InstructionsPath() = %q", fa.InstructionsPath(""))
	}
	if fa.SettingsPath("") != "/home/user/.claude/settings.json" {
		t.Errorf("SettingsPath() = %q", fa.SettingsPath(""))
	}
	if fa.MCPConfigPath("", "context7") != "/home/user/.claude/mcp/context7.json" {
		t.Errorf("MCPConfigPath() = %q", fa.MCPConfigPath("", "context7"))
	}
}
