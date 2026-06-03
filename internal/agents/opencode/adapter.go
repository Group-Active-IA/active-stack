// Package opencode provides the JR Stack agent adapter for OpenCode.
package opencode

import (
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Adapter resolves OpenCode-specific filesystem paths and config strategies.
// It satisfies skill.AgentAdapter, config.AgentAdapter,
// permissions.PermissionsAdapter, and external.AgentAdapter simultaneously.
//
// Note: internal/harness/config/assets/opencode/ exists in this repo, so
// VariantKey "opencode" resolves to a real asset directory (not "generic").
type Adapter struct{}

// NewAdapter returns a ready-to-use OpenCode adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Agent returns the model.Agent identifier for OpenCode.
func (a *Adapter) Agent() model.Agent {
	return model.AgentOpenCode
}

// InstructionsPath returns the path to OpenCode's primary instructions file.
// Example: /home/user/.config/opencode/AGENTS.md
func (a *Adapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
}

// SkillsDir returns the path to OpenCode's skills directory.
// Example: /home/user/.config/opencode/skills
func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "skills")
}

// CommandsDir returns the path to OpenCode's slash-command directory for the
// machine (global) target. Machine layout uses ~/.config/opencode/ (XDG).
// Example: /home/user/.config/opencode/commands
// Added in C-31 (D1) — mirrors SkillsDir on the adapter interface.
// For the project target, use PathsFor(base, Project).CommandsDir which
// resolves to <root>/.opencode/commands (project layout differs from machine).
func (a *Adapter) CommandsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "commands")
}

// SettingsPath returns the path to OpenCode's settings file.
// Example: /home/user/.config/opencode/opencode.json
func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

// MCPConfigPath returns the settings file path for MCP config. OpenCode merges
// MCP entries into opencode.json rather than using separate files.
// Example: /home/user/.config/opencode/opencode.json
func (a *Adapter) MCPConfigPath(homeDir, _ string) string {
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

// MCPStrategy returns external.StrategyMergeIntoSettings — OpenCode merges
// MCP entries into opencode.json rather than writing per-server files.
func (a *Adapter) MCPStrategy() external.MCPStrategy {
	return external.StrategyMergeIntoSettings
}

// VariantKey returns "opencode", used to select the opencode/ asset directory
// in the config installer.
func (a *Adapter) VariantKey() string {
	return "opencode"
}

// PathsFor returns the resolved AgentPaths for the given base directory and
// install target. For Machine, the result is identical to the existing per-method
// outputs (zero regression). For Project, OpenCode uses .opencode/ instead of
// .config/opencode/ — this difference lives here, never in the caller.
//
// OpenCode layout by target (D2, confirmed against official docs):
//   Machine: <homeDir>/.config/opencode/...   (XDG global convention)
//   Project: <root>/.opencode/...             (project-local convention)
func (a *Adapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	var openCodeDir string
	switch t {
	case model.Project:
		// Project layout: .opencode/ in the project root (NOT .config/opencode/).
		// Confirmed: official OpenCode docs use .opencode/ for per-project config.
		openCodeDir = filepath.Join(base, ".opencode")
	default:
		// Machine layout (default, zero-value): .config/opencode/ under home.
		// This is identical to the pre-C-27 per-method results → zero regression.
		openCodeDir = filepath.Join(base, ".config", "opencode")
	}
	return model.AgentPaths{
		InstructionsPath: filepath.Join(openCodeDir, "AGENTS.md"),
		SkillsDir:        filepath.Join(openCodeDir, "skills"),
		SettingsPath:     filepath.Join(openCodeDir, "opencode.json"),
		CommandsDir:      filepath.Join(openCodeDir, "commands"),
	}.WithMCPConfigFn(func(_ string) string {
		// OpenCode merges MCP entries into opencode.json (StrategyMergeIntoSettings),
		// so the MCP config path is always the settings file, regardless of server name.
		return filepath.Join(openCodeDir, "opencode.json")
	}).WithMCPStrategy(model.MCPStrategyMergeIntoSettings)
}

// ConfigDelivery returns model.ConfigDeliveryPrimaryAgent — OpenCode only treats
// a block as a tab-able agent when it is registered under agent.<id> with
// mode:primary in opencode.json. Injecting into the shared AGENTS.md instead
// leaks the orchestrator into every agent (plan/build) and registers no tab.
func (a *Adapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryPrimaryAgent
}
