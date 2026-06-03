// Package claude provides the JR Stack agent adapter for Claude Code.
package claude

import (
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Adapter resolves Claude Code-specific filesystem paths and config strategies.
// It satisfies skill.AgentAdapter, config.AgentAdapter,
// permissions.PermissionsAdapter, and external.AgentAdapter simultaneously.
type Adapter struct{}

// NewAdapter returns a ready-to-use Claude Code adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

// Agent returns the model.Agent identifier for Claude Code.
func (a *Adapter) Agent() model.Agent {
	return model.AgentClaude
}

// InstructionsPath returns the path to Claude Code's primary instructions file.
// Example: /home/user/.claude/CLAUDE.md
func (a *Adapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "CLAUDE.md")
}

// SkillsDir returns the path to Claude Code's skills directory.
// Example: /home/user/.claude/skills
func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
}

// CommandsDir returns the path to Claude Code's slash-command directory.
// Example: /home/user/.claude/commands
// Added in C-31 (D1) — mirrors SkillsDir on the adapter interface.
func (a *Adapter) CommandsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "commands")
}

// SettingsPath returns the path to Claude Code's settings file.
// Example: /home/user/.claude/settings.json
func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "settings.json")
}

// MCPConfigPath returns the path where an MCP server config file should be
// written for Claude Code. Claude uses one JSON file per server.
// Example: /home/user/.claude/mcp/context7.json
func (a *Adapter) MCPConfigPath(homeDir, serverName string) string {
	return filepath.Join(homeDir, ".claude", "mcp", serverName+".json")
}

// MCPStrategy returns external.StrategySeparateFile — Claude Code writes one
// JSON file per MCP server under ~/.claude/mcp/<server>.json.
func (a *Adapter) MCPStrategy() external.MCPStrategy {
	return external.StrategySeparateFile
}

// VariantKey returns "claude", used to select the claude/ asset directory in
// the config installer.
func (a *Adapter) VariantKey() string {
	return "claude"
}

// PathsFor returns the resolved AgentPaths for the given base directory and
// install target. For Machine, the result is identical to the existing per-method
// outputs (zero regression). For Project, non-MCP paths use the same .claude/
// subdirectory layout but the MCP path resolves to <root>/.mcp.json (D2).
//
// Claude layout by target:
//   Machine: <homeDir>/.claude/...  + MCP: <homeDir>/.claude/mcp/<server>.json
//   Project: <root>/.claude/...     + MCP: <root>/.mcp.json  (server name ignored)
//
// The MCP strategy is also resolved together with the path (D1) so the two
// cannot contradict each other:
//   Machine → MCPStrategySeparateFile (legacy, unchanged)
//   Project → MCPStrategySingleFileMerge (all servers share one .mcp.json)
func (a *Adapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	claudeDir := filepath.Join(base, ".claude")
	paths := model.AgentPaths{
		InstructionsPath: filepath.Join(claudeDir, "CLAUDE.md"),
		SkillsDir:        filepath.Join(claudeDir, "skills"),
		SettingsPath:     filepath.Join(claudeDir, "settings.json"),
		CommandsDir:      filepath.Join(claudeDir, "commands"),
	}

	switch t {
	case model.Project:
		// Project target (D2): a single .mcp.json at the repo root, independent
		// of the server name. Strategy is single-file merge into mcpServers key.
		return paths.
			WithMCPConfigFn(func(_ string) string {
				return filepath.Join(base, ".mcp.json")
			}).
			WithMCPStrategy(model.MCPStrategySingleFileMerge)
	default:
		// Machine target (zero-value): identical to pre-C-27 per-method results.
		// Strategy is separate-file (one JSON file per server).
		// (TBD) machine path correctness (~/.claude.json) is a separate follow-up.
		return paths.
			WithMCPConfigFn(func(serverName string) string {
				return filepath.Join(claudeDir, "mcp", serverName+".json")
			}).
			WithMCPStrategy(model.MCPStrategySeparateFile)
	}
}

// ConfigDelivery returns model.ConfigDeliveryInstructions — Claude Code reads a
// flat instructions file (~/.claude/CLAUDE.md), so config harnesses inject
// there.
func (a *Adapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
