// Package model defines the core domain types of the JR Stack installer.
package model

// HarnessType is how a harness materializes when installed.
type HarnessType string

const (
	// HarnessSkill is a SKILL.md module fetched from a git repo and copied
	// into each agent's skill directory.
	HarnessSkill HarnessType = "skill"
	// HarnessConfig is text/files bundled in the installer that configure the
	// agent (e.g. the sdd-orchestrator block, permissions).
	HarnessConfig HarnessType = "config"
	// HarnessExternal is a third-party binary or service we install/configure
	// but do not own (e.g. Engram, OpenSpec CLI, Context7).
	HarnessExternal HarnessType = "external"
	// HarnessCommand is an embedded slash-command .md file written into each
	// focused agent's command directory. Added in C-31 (TBD-2 option a).
	HarnessCommand HarnessType = "command"
)

// IsValid reports whether t is a known harness type.
func (t HarnessType) IsValid() bool {
	switch t {
	case HarnessSkill, HarnessConfig, HarnessExternal, HarnessCommand:
		return true
	default:
		return false
	}
}

// InstallMode is a preset bundle of harnesses.
type InstallMode string

const (
	// ModeLite is the minimum needed to start working with the methodology.
	ModeLite InstallMode = "lite"
	// ModeFull is the complete methodology ecosystem.
	ModeFull InstallMode = "full"
	// ModeCustom lets the user pick each harness.
	ModeCustom InstallMode = "custom"
)

// IsValid reports whether m is a known install mode.
func (m InstallMode) IsValid() bool {
	switch m {
	case ModeLite, ModeFull, ModeCustom:
		return true
	default:
		return false
	}
}

// Agent identifies a supported AI coding agent.
type Agent string

const (
	AgentClaude      Agent = "claude"
	AgentOpenCode    Agent = "opencode"
	AgentGemini      Agent = "gemini"
	AgentCodex       Agent = "codex"
	AgentCursor      Agent = "cursor"
	AgentVSCode      Agent = "vscode"
	AgentWindsurf    Agent = "windsurf"
	AgentAntigravity Agent = "antigravity"
)

// ConfigDelivery is how a config-type harness materializes for a given agent.
//
// Why this exists: a config harness like sdd-orchestrator is the SAME content
// for every agent, but agents disagree on WHERE it must live to take effect.
// Claude reads a flat instructions file (CLAUDE.md); OpenCode only treats a
// block as a tab-able "primary agent" when it is registered under the "agent"
// key of opencode.json with "mode": "primary". Injecting the orchestrator into
// OpenCode's AGENTS.md instead leaks it into every agent (plan/build) and
// registers no new tab — the exact bug this type fixes.
type ConfigDelivery int

const (
	// ConfigDeliveryInstructions injects the composed block into the agent's
	// instructions file (e.g. ~/.claude/CLAUDE.md) via markdown markers.
	// This is the default and the zero value.
	ConfigDeliveryInstructions ConfigDelivery = iota
	// ConfigDeliveryPrimaryAgent registers the composed block as a dedicated
	// primary agent inside the agent's settings JSON (e.g. opencode.json under
	// agent.<id> with mode:primary), so the host TUI exposes it as a tab-able
	// agent instead of folding it into the shared instructions.
	ConfigDeliveryPrimaryAgent
)

// Source locates a skill harness in a git repository.
type Source struct {
	Repo   string `yaml:"repo"`             // e.g. JuanCruzRobledo/kb-creator
	Ref    string `yaml:"ref,omitempty"`    // tag/branch/commit; defaults to "latest"
	Method string `yaml:"method,omitempty"` // clone | embed; inferred if empty
	// Path is the subdir within the cloned repo where the SKILL.md lives.
	// Empty = repo root (C-16 root-first behavior). Used by third-party
	// monorepos where the skill is nested (e.g. skills/find-skills).
	Path string `yaml:"path,omitempty"`
}

// External describes how to install a third-party tool harness.
type External struct {
	Method string `yaml:"method"`         // homebrew | download | npm | go-install | mcp
	Pkg    string `yaml:"pkg,omitempty"`  // package/formula/module identifier (brew formula, npm pkg)
	Repo   string `yaml:"repo,omitempty"` // GitHub owner/repo for binary download fallback (distinct from Pkg)
	URL    string `yaml:"url,omitempty"`  // for download/mcp transports
}

// Harness is a single installable/configurable module of the stack.
type Harness struct {
	ID           string        `yaml:"id"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description,omitempty"`
	Type         HarnessType   `yaml:"type"`
	ThirdParty   bool          `yaml:"third_party,omitempty"` // not owned by us, bundled
	Source       *Source       `yaml:"source,omitempty"`      // skill harnesses
	External     *External     `yaml:"external,omitempty"`    // external harnesses
	Toggles      []string      `yaml:"toggles,omitempty"`     // config harnesses (composable pieces)
	InstallModes []InstallMode `yaml:"install_modes"`
	DependsOn    []string      `yaml:"depends_on,omitempty"`
	Agents       []Agent       `yaml:"agents,omitempty"` // empty = all agents
	// BestEffort marks a harness as tolerant of install failures.
	// When true, a failing install step emits a warning and continues
	// (the pipeline does not abort and roll back). Defaults to false.
	BestEffort bool `yaml:"best_effort,omitempty"`
}

// InMode reports whether this harness belongs to the given install mode.
// ModeCustom matches every harness (the user picks individually).
func (h Harness) InMode(m InstallMode) bool {
	if m == ModeCustom {
		return true
	}
	for _, mode := range h.InstallModes {
		if mode == m {
			return true
		}
	}
	return false
}

// SupportsAgent reports whether this harness applies to the given agent.
// An empty Agents list means it applies to all agents.
func (h Harness) SupportsAgent(a Agent) bool {
	if len(h.Agents) == 0 {
		return true
	}
	for _, agent := range h.Agents {
		if agent == a {
			return true
		}
	}
	return false
}
