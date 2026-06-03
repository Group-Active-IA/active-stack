// Package uninstall wires the headless uninstall flow:
// catalog selection → backup snapshot → reversal steps.
// It mirrors the structure of internal/install (Intent → BuildPlan → Plan).
package uninstall

import (
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// Strategy controls how the uninstall reverses an installation.
type Strategy string

const (
	// StrategyTargeted reverses each harness individually:
	// marker removal for config harnesses, skill-dir removal for skill harnesses.
	StrategyTargeted Strategy = "targeted"
	// StrategyRestore performs a full restore from the install-time backup manifest,
	// reverting all agent config to the pre-install state.
	StrategyRestore Strategy = "restore"
)

// Intent describes what the user wants to uninstall.
type Intent struct {
	// Agents is the ordered list of agents to target.
	Agents []model.Agent
	// Mode selects the harness bundle (Lite, Full, or Custom).
	Mode model.InstallMode
	// Custom lists explicit harness IDs when Mode == ModeCustom.
	Custom []string
	// Strategy controls the reversal mechanism (targeted or restore).
	Strategy Strategy
}

// Catalog is the read interface consumed by BuildPlan.
// It is satisfied by *catalog.Catalog from internal/catalog.
// Mirrored from internal/install — do NOT import the install package.
type Catalog interface {
	ByID(id string) (model.Harness, bool)
	ForMode(m model.InstallMode) []model.Harness
	ForAgent(a model.Agent) []model.Harness
}

// AgentAdapter is the superset interface needed by the uninstall steps.
// It is satisfied by agents.Adapter from internal/agents.
// Mirrored from internal/install — do NOT import the install package.
type AgentAdapter interface {
	Agent() model.Agent
	InstructionsPath(homeDir string) string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string
	ConfigDelivery() model.ConfigDelivery
}

// Registry maps agents to their adapters. It is satisfied by *agents.Registry.
// Mirrored from internal/install — do NOT import the install package.
type Registry interface {
	Get(agent model.Agent) (AgentAdapter, bool)
}

// Options carries the dependencies and configuration for BuildPlan.
type Options struct {
	// HomeDir is the user's home directory, passed to adapters for path resolution.
	HomeDir string
	// Registry maps agents to their concrete adapters.
	Registry Registry
	// OnProgress receives progress events during uninstall.
	// When nil no progress events are emitted.
	OnProgress pipeline.ProgressFunc
	// RestoreManifest is the install-time backup manifest to use when
	// Intent.Strategy == StrategyRestore. Ignored for StrategyTargeted.
	RestoreManifest *backup.Manifest
}

// Plan is the output of BuildPlan. It combines the pipeline.StagePlan with
// the ProgressFunc so callers can wire both into the Orchestrator.
type Plan struct {
	// StagePlan is the pipeline-ready execution plan.
	pipeline.StagePlan
	// OnProgress is the ProgressFunc from Options, ready to wire into
	// pipeline.NewOrchestrator via pipeline.WithProgressFunc.
	OnProgress pipeline.ProgressFunc
}
