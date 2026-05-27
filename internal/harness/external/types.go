package external

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// Result is the output of a successful external harness installation.
type Result struct {
	// BinaryPath is the path to the installed binary (npm/homebrew methods).
	// Empty for the mcp method.
	BinaryPath string
	// ConfigFiles lists the config files written or merged (mcp method).
	ConfigFiles []string
	// AlreadyInstalled is true when the tool was already present and no
	// changes were made.
	AlreadyInstalled bool
}

// MCPStrategy controls how MCP server entries are injected into an agent config.
type MCPStrategy int

const (
	// StrategySeparateFile writes a standalone JSON file per MCP server
	// (Claude Code pattern: ~/.claude/mcp/<server>.json).
	StrategySeparateFile MCPStrategy = iota
	// StrategyMergeIntoSettings merges MCP entries into an existing settings
	// file (OpenCode opencode.json, Gemini settings.json).
	StrategyMergeIntoSettings
)

// AgentAdapter is the minimal interface the mcp installer needs per agent.
// Full adapters are implemented in internal/agents (C-10).
type AgentAdapter interface {
	Agent() model.Agent
	// MCPConfigPath returns the path the MCP config should be written to.
	// serverName is the harness ID (e.g. "context7").
	MCPConfigPath(homeDir, serverName string) string
	// MCPStrategy returns how this agent expects MCP config to be injected.
	MCPStrategy() MCPStrategy
}
