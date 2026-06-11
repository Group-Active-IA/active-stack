// Package command implements installation of slash-command harnesses — embedded
// .md files written into each focused agent's slash-command directory.
//
// This package mirrors internal/harness/skill in structure (embed → idempotent
// write → backup) but for slash-command files rather than SKILL.md directories.
// Added in C-31.
package command

import "github.com/Group-Active-IA/active-stack/internal/model"

// AgentAdapter is the minimal interface the command installer needs per agent.
// Follows the local-interface pattern established in internal/harness/skill/types.go.
type AgentAdapter interface {
	// Agent returns the model.Agent identifier for this adapter.
	Agent() model.Agent

	// CommandsDir returns the path to the agent's slash-command directory for
	// the machine (global) target. An empty string signals that this agent does
	// not support commands and should be skipped silently.
	// For the project target, the caller should use adapter.PathsFor(base, Project).CommandsDir
	// and pass it pre-resolved; the installer receives the already-resolved dir.
	CommandsDir(homeDir string) string

	// VariantKey returns the asset base key used to select the per-agent
	// command asset directory (e.g. "claude", "opencode").
	VariantKey() string
}

// Result describes the outcome of a command-file installation for a single agent.
type Result struct {
	// CommandPath is the absolute path to the installed command file.
	CommandPath string
	// AlreadyInstalled is true when the destination already contained
	// byte-identical content and no changes were made.
	AlreadyInstalled bool
}
