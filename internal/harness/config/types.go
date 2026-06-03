package config

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// AgentAdapter is the minimal interface the config installer needs per agent.
// Full adapters are implemented in internal/agents (C-10).
// This is a local mirror of the pattern established in internal/harness/external/types.go.
type AgentAdapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent
	// InstructionsPath returns the path to the agent's instructions file
	// (e.g. ~/.claude/CLAUDE.md, AGENTS.md in a project root).
	// An empty string means this agent should be skipped for injection.
	InstructionsPath(homeDir string) string
	// VariantKey returns the asset base key for this agent
	// (e.g. "claude", "codex", "generic").
	// If no asset directory matches the key, the installer falls back to "generic".
	VariantKey() string
	// SettingsPath returns the path to the agent's settings JSON file
	// (e.g. ~/.config/opencode/opencode.json). Used only when ConfigDelivery
	// is ConfigDeliveryPrimaryAgent. An empty string means the agent has no
	// settings file and primary-agent delivery is skipped.
	SettingsPath(homeDir string) string
	// ConfigDelivery reports how this agent expects a config-type harness to
	// materialize: injected into the instructions file (default) or registered
	// as a primary agent in the settings JSON.
	ConfigDelivery() model.ConfigDelivery
}

// Result describes the outcome of a config harness installation.
type Result struct {
	// Files lists the instruction files that were written or updated.
	Files []string
	// AllAlready is true when every adapter produced byte-identical content
	// and no files were changed.
	AllAlready bool
}
