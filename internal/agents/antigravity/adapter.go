// Package antigravity provides the Active Stack adapter for Google Antigravity.
package antigravity

import (
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Adapter resolves Antigravity-specific filesystem paths and config strategies.
type Adapter struct{}

// NewAdapter returns a ready-to-use Antigravity adapter.
func NewAdapter() *Adapter { return &Adapter{} }

// Agent returns the model.Agent identifier for Antigravity.
func (a *Adapter) Agent() model.Agent { return model.AgentAntigravity }

// InstructionsPath returns the path to Antigravity's global instructions file.
// Antigravity loads ~/.gemini/config/AGENTS.md as an always-on global rule.
func (a *Adapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "config", "AGENTS.md")
}

// SkillsDir returns the path to Antigravity's global skills directory.
func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "config", "skills")
}

// CommandsDir returns empty because Antigravity exposes skills automatically as
// slash commands and does not use a flat markdown command directory.
func (a *Adapter) CommandsDir(string) string { return "" }

// SettingsPath returns the path to Antigravity's global settings file.
func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "config", "config.json")
}

// MCPConfigPath returns the path to Antigravity's global MCP config file.
// Antigravity stores MCP servers under the top-level "mcpServers" key in
// ~/.gemini/config/mcp_config.json.
func (a *Adapter) MCPConfigPath(homeDir, _ string) string {
	return filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
}

// MCPStrategy returns StrategyMergeIntoSettings so MCP entries are merged under
// the top-level "mcpServers" object inside mcp_config.json.
func (a *Adapter) MCPStrategy() external.MCPStrategy {
	return external.StrategyMergeIntoSettings
}

// VariantKey returns "antigravity", resolving to internal/harness/config/assets/antigravity/.
func (a *Adapter) VariantKey() string { return "antigravity" }

// PathsFor resolves Antigravity filesystem paths for Machine (~/.gemini/config/...)
// or Project (<root>/AGENTS.md and <root>/.agents/...).
func (a *Adapter) PathsFor(base string, target model.InstallTarget) model.AgentPaths {
	if target == model.Project {
		return model.AgentPaths{
			InstructionsPath: filepath.Join(base, "AGENTS.md"),
			SkillsDir:        filepath.Join(base, ".agents", "skills"),
			SettingsPath:     filepath.Join(base, ".agents", "config.json"),
		}.WithMCPConfigFn(func(string) string {
			return filepath.Join(base, ".agents", "mcp_config.json")
		}).WithMCPStrategy(model.MCPStrategyMergeIntoSettings)
	}

	return model.AgentPaths{
		InstructionsPath: a.InstructionsPath(base),
		SkillsDir:        a.SkillsDir(base),
		SettingsPath:     a.SettingsPath(base),
	}.WithMCPConfigFn(func(name string) string {
		return a.MCPConfigPath(base, name)
	}).WithMCPStrategy(model.MCPStrategyMergeIntoSettings)
}

// ConfigDelivery returns ConfigDeliveryInstructions so sdd-orchestrator and
// active-orchestrator inject via markers into AGENTS.md.
func (a *Adapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
