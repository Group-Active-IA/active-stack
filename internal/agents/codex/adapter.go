// Package codex provides the Active Stack adapter for OpenAI Codex.
package codex

import (
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Adapter resolves Codex-specific filesystem paths and config strategies.
type Adapter struct{}

func NewAdapter() *Adapter { return &Adapter{} }

func (a *Adapter) Agent() model.Agent { return model.AgentCodex }

func (a *Adapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "AGENTS.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".agents", "skills")
}

// Codex currently exposes built-in slash commands, not a custom command
// directory that Active Stack can materialize into.
func (a *Adapter) CommandsDir(string) string { return "" }

func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "config.toml")
}

func (a *Adapter) MCPConfigPath(homeDir, _ string) string {
	return a.SettingsPath(homeDir)
}

func (a *Adapter) MCPStrategy() external.MCPStrategy {
	return external.StrategyMergeIntoTOML
}

func (a *Adapter) VariantKey() string { return "codex" }

func (a *Adapter) PathsFor(base string, target model.InstallTarget) model.AgentPaths {
	if target == model.Project {
		return model.AgentPaths{
			InstructionsPath: filepath.Join(base, "AGENTS.md"),
			SkillsDir:        filepath.Join(base, ".agents", "skills"),
			SettingsPath:     filepath.Join(base, ".codex", "config.toml"),
		}.WithMCPConfigFn(func(string) string {
			return filepath.Join(base, ".codex", "config.toml")
		}).WithMCPStrategy(model.MCPStrategyMergeIntoTOML)
	}

	return model.AgentPaths{
		InstructionsPath: a.InstructionsPath(base),
		SkillsDir:        a.SkillsDir(base),
		SettingsPath:     a.SettingsPath(base),
	}.WithMCPConfigFn(func(string) string {
		return a.SettingsPath(base)
	}).WithMCPStrategy(model.MCPStrategyMergeIntoTOML)
}

func (a *Adapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
