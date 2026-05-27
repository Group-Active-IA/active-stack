package permissions

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// PermissionsAdapter is the minimal interface the permissions installer needs
// per agent. Full adapters are implemented in internal/agents (C-10).
// Any adapter that satisfies AgentAdapter in internal/harness/config also
// satisfies PermissionsAdapter if it implements SettingsPath — no modification
// of existing adapters is required (ISP, D2 in design.md).
type PermissionsAdapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent
	// SettingsPath returns the absolute path to the agent's settings.json (or
	// equivalent JSON config file) for the given home directory.
	// An empty string signals that this agent has no injectable settings file
	// and the installer must skip it without error (explicit no-op).
	SettingsPath(homeDir string) string
}

// Result describes the outcome of a permissions harness installation run.
type Result struct {
	// Changed is true when at least one settings file was written or updated.
	Changed bool
	// Files lists the settings files that were written or updated.
	Files []string
}
