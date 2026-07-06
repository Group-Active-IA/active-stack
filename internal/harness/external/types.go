package external

import "github.com/Group-Active-IA/active-stack/internal/model"

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

type DownloadEventType string

const (
	DownloadStarted  DownloadEventType = "download_started"
	DownloadProgress DownloadEventType = "download_progress"
	DownloadFinished DownloadEventType = "download_finished"
)

type DownloadEvent struct {
	Type    DownloadEventType
	URL     string
	Message string
}

type DownloadEventFunc func(DownloadEvent)

// MCPStrategy controls how MCP server entries are injected into an agent config.
type MCPStrategy int

const (
	// StrategySeparateFile writes a standalone JSON file per MCP server
	// (Claude Code pattern: ~/.claude/mcp/<server>.json).
	StrategySeparateFile MCPStrategy = iota
	// StrategyMergeIntoSettings merges MCP entries into an existing settings
	// file (OpenCode opencode.json, Gemini settings.json).
	StrategyMergeIntoSettings
	// StrategyMergeIntoTOML merges server tables into Codex config.toml.
	StrategyMergeIntoTOML
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
