// Package claude provides the JR Stack agent adapter for Claude Code.
package claude

import (
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/model"
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

// MCPConfigPath returns the path to Claude Code's user-scope MCP config file.
// Claude Code reads user-scope MCP servers from the top-level "mcpServers" key
// in ~/.claude.json — there is no ~/.claude/mcp/ auto-discovery. The serverName
// argument is ignored because all servers share the single ~/.claude.json file.
// Example: /home/user/.claude.json
func (a *Adapter) MCPConfigPath(homeDir, _ string) string {
	return filepath.Join(homeDir, ".claude.json")
}

// MCPStrategy returns external.StrategyMergeIntoSettings — Claude Code reads
// user-scope MCP servers from the top-level "mcpServers" key in ~/.claude.json,
// so new entries are merged into that single shared file rather than written to
// per-server files under ~/.claude/mcp/ (which Claude Code does not read).
func (a *Adapter) MCPStrategy() external.MCPStrategy {
	return external.StrategyMergeIntoSettings
}

// VariantKey returns "claude", used to select the claude/ asset directory in
// the config installer.
func (a *Adapter) VariantKey() string {
	return "claude"
}

// PathsFor returns the resolved AgentPaths for the given base directory and
// install target. For Project, non-MCP paths use the .claude/ subdirectory
// layout and the MCP path resolves to <root>/.mcp.json (D2). For Machine,
// non-MCP paths are identical to the existing per-method outputs; the MCP path
// resolves to <base>/.claude.json (the single shared user-scope file that
// Claude Code actually reads under the top-level "mcpServers" key).
//
// Claude layout by target:
//
//	Machine: <base>/.claude/...  + MCP: <base>/.claude.json  (SingleFileMerge)
//	Project: <root>/.claude/...  + MCP: <root>/.mcp.json     (SingleFileMerge)
//
// The MCP strategy is resolved together with the path (D1) so the two cannot
// contradict each other. Both targets now use MCPStrategySingleFileMerge:
//
//	Machine → MCPStrategySingleFileMerge, target file ~/.claude.json
//	Project → MCPStrategySingleFileMerge, target file <root>/.mcp.json
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
		// Machine target: Claude Code reads user-scope MCP servers from the
		// top-level "mcpServers" key in ~/.claude.json. There is no
		// ~/.claude/mcp/ auto-discovery. All servers share the single file;
		// the server name is ignored when resolving the path.
		return paths.
			WithMCPConfigFn(func(_ string) string {
				return filepath.Join(base, ".claude.json")
			}).
			WithMCPStrategy(model.MCPStrategySingleFileMerge)
	}
}

// ConfigDelivery returns model.ConfigDeliveryInstructions — Claude Code reads a
// flat instructions file (~/.claude/CLAUDE.md), so config harnesses inject
// there.
func (a *Adapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
