package verify

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// Adapter is a minimal, local interface for resolving agent-specific paths.
// It is a structural subset of agents.Adapter (from internal/agents), covering
// only the methods needed by the verify health-check builders.
//
// By defining this interface locally, internal/verify avoids importing
// internal/agents and therefore avoids any import cycle. Concrete adapters
// (e.g. *claude.Adapter, *opencode.Adapter) satisfy this interface by
// structural typing without modification.
type Adapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent

	// SkillsDir returns the path to the agent's skills directory.
	// Used to verify that a skill harness was cloned/copied successfully.
	SkillsDir(homeDir string) string

	// InstructionsPath returns the path to the agent's primary instructions file.
	// Used to verify that a config harness block was injected with its markers.
	InstructionsPath(homeDir string) string

	// SettingsPath returns the path to the agent's settings file.
	// Used to verify that the permissions harness wrote its key.
	SettingsPath(homeDir string) string

	// MCPConfigPath returns the path where an MCP server config was written.
	// serverName is the harness ID (e.g. "context7", "engram").
	MCPConfigPath(homeDir, serverName string) string

	// ConfigDelivery reports how a config-type harness materializes for this
	// agent, so the health check looks in the right place: the marker in the
	// instructions file, or the primary-agent entry in the settings JSON.
	ConfigDelivery() model.ConfigDelivery
}
