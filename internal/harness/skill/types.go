// Package skill implements installation of skill harnesses — SKILL.md modules
// fetched from git repos, third-party registries, or embedded assets, then
// copied into each agent's skills directory.
package skill

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// AgentAdapter is the minimal interface the skill installer needs per agent.
// Full adapters are implemented in internal/agents (C-10).
// This follows the local-interface pattern established in
// internal/harness/external/types.go and internal/harness/config/types.go.
type AgentAdapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent
	// SkillsDir returns the path to the agent's skills directory.
	// An empty string signals that this adapter does not support skills and
	// should be skipped.
	SkillsDir(homeDir string) string
}

// Result describes the outcome of a skill installation for a single agent.
type Result struct {
	// SkillPath is the absolute path to the installed skill directory.
	SkillPath string
	// AlreadyInstalled is true when the destination already contained
	// byte-identical content and no changes were made.
	AlreadyInstalled bool
}
