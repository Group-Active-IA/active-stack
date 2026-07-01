package model

// MCPStrategy controls how MCP server entries are injected into an agent config.
//
// This enum is defined in internal/model (not internal/harness/external) to avoid
// a cyclic import: external imports model, so model cannot import external. (D1a)
//
// internal/harness/external.MCPStrategy mirrors these values and is kept for
// backward compat with the legacy adapter.MCPStrategy() machine-target method.
type MCPStrategy int

const (
	// MCPStrategySeparateFile writes a standalone JSON file per MCP server.
	// (Claude machine target: ~/.claude/mcp/<server>.json)
	// This is the zero value so existing code that doesn't set the field is
	// backward compatible with the pre-C-28 machine behavior.
	MCPStrategySeparateFile MCPStrategy = iota
	// MCPStrategyMergeIntoSettings merges MCP entries into an existing settings
	// file (OpenCode: opencode.json — both machine and project targets).
	MCPStrategyMergeIntoSettings
	// MCPStrategySingleFileMerge merges all MCP servers into a single shared
	// file under the "mcpServers" key.
	// (Claude project target: <root>/.mcp.json — D2)
	MCPStrategySingleFileMerge
	// MCPStrategyMergeIntoTOML merges MCP server tables into a Codex
	// config.toml while preserving unrelated user configuration.
	MCPStrategyMergeIntoTOML
)

// AgentPaths holds the resolved filesystem paths for a given agent and install
// target. It is returned by the Adapter.PathsFor method and used by the install
// pipeline to route writes to the correct location without knowing per-agent
// directory layouts.
type AgentPaths struct {
	// InstructionsPath is the path to the agent's primary instructions file
	// (e.g. CLAUDE.md, AGENTS.md).
	InstructionsPath string
	// SkillsDir is the path to the agent's skills directory.
	SkillsDir string
	// SettingsPath is the path to the agent's settings/config file.
	SettingsPath string
	// CommandsDir is the path to the agent's slash-command directory.
	// Empty string means this agent/target does not support commands and
	// the command installer should skip it (mirrors the SkillsDir skip contract).
	// Added in C-31 (D1).
	CommandsDir string
	// MCPStrategy is the target-aware MCP injection strategy resolved alongside
	// the MCP config path. Path and strategy are resolved together in PathsFor
	// so they cannot contradict each other (D1).
	//
	// Zero value (MCPStrategySeparateFile) is the pre-C-28 machine default.
	MCPStrategy MCPStrategy
	// mcpConfigFn computes the MCP config path for a given server name.
	// It encapsulates the per-agent, per-target logic.
	// Set via WithMCPConfigFn; nil returns empty string.
	mcpConfigFn func(serverName string) string
}

// MCPConfigPath returns the resolved path for an MCP server config,
// using the agent-specific, target-aware logic captured at construction time.
//
//   - Claude (StrategySeparateFile): <agentDir>/mcp/<serverName>.json
//   - OpenCode (StrategyMergeIntoSettings): <agentDir>/opencode.json
func (p AgentPaths) MCPConfigPath(serverName string) string {
	if p.mcpConfigFn == nil {
		return ""
	}
	return p.mcpConfigFn(serverName)
}

// WithMCPConfigFn returns a copy of p with the MCP config path function set.
// Adapter implementations call this in PathsFor to wire the per-agent MCP logic.
func (p AgentPaths) WithMCPConfigFn(fn func(serverName string) string) AgentPaths {
	p.mcpConfigFn = fn
	return p
}

// WithMCPStrategy returns a copy of p with the MCPStrategy field set.
// Adapter implementations call this in PathsFor alongside WithMCPConfigFn so
// that path and strategy are resolved atomically (D1).
func (p AgentPaths) WithMCPStrategy(s MCPStrategy) AgentPaths {
	p.MCPStrategy = s
	return p
}

// WithCommandsDir returns a copy of p with the CommandsDir field set.
// Adapter implementations call this in PathsFor to wire the agent's command
// directory path (C-31 D1). An empty dir signals skip-on-install.
func (p AgentPaths) WithCommandsDir(dir string) AgentPaths {
	p.CommandsDir = dir
	return p
}
