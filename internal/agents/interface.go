package agents

import (
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Adapter is the public interface for an agent adapter.
// It is the union of the four harness-installer local interfaces:
//   - skill.AgentAdapter     → Agent, SkillsDir
//   - config.AgentAdapter    → Agent, InstructionsPath, VariantKey
//   - permissions.PermissionsAdapter → Agent, SettingsPath
//   - external.AgentAdapter  → Agent, MCPConfigPath, MCPStrategy
//
// Callers (TUI, pipeline) hold one Adapter value and pass it to each
// installer unchanged — structural typing satisfies every narrow interface.
type Adapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent

	// InstructionsPath returns the path to the agent's primary instructions
	// file (e.g. ~/.claude/CLAUDE.md, ~/.config/opencode/AGENTS.md).
	InstructionsPath(homeDir string) string

	// SkillsDir returns the path to the agent's skills directory.
	SkillsDir(homeDir string) string

	// CommandsDir returns the path to the agent's slash-command directory
	// for the machine (global) target. An empty string signals that this
	// agent does not support commands and the command installer should skip it.
	// For the project target, use PathsFor(base, Project).CommandsDir.
	// Added in C-31 (D1) — mirrors SkillsDir.
	CommandsDir(homeDir string) string

	// SettingsPath returns the absolute path to the agent's settings file
	// (e.g. ~/.claude/settings.json, ~/.config/opencode/opencode.json).
	SettingsPath(homeDir string) string

	// MCPConfigPath returns the path where an MCP server config should be
	// written. serverName is the harness ID (e.g. "context7").
	MCPConfigPath(homeDir, serverName string) string

	// MCPStrategy returns how this agent expects MCP config to be injected.
	MCPStrategy() external.MCPStrategy

	// VariantKey returns the asset base key used to select variant-specific
	// config assets (e.g. "claude", "opencode", "generic").
	VariantKey() string

	// PathsFor returns the resolved model.AgentPaths for the given base directory
	// and install target. For Machine, the paths are identical to the existing
	// per-method results (zero regression). For Project, the paths resolve
	// under the agent's project layout (which may differ from machine layout).
	//
	// The per-agent layout difference lives inside the adapter implementation,
	// never in the caller. This is the single target-aware resolver added by C-27.
	PathsFor(base string, t model.InstallTarget) model.AgentPaths

	// ConfigDelivery reports how a config-type harness materializes for this
	// agent: injected into the instructions file (default) or registered as a
	// primary agent in the settings JSON.
	ConfigDelivery() model.ConfigDelivery
}
