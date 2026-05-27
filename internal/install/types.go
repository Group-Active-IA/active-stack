// Package install wires the headless installation flow:
// catalog selection → dependency resolution → backup snapshot → installer steps.
package install

import (
	"io/fs"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// Intent describes what the user wants to install.
type Intent struct {
	// Agents is the ordered list of agents to target.
	Agents []model.Agent
	// Mode selects the harness bundle (Lite, Full, or Custom).
	Mode model.InstallMode
	// Custom lists explicit harness IDs when Mode == ModeCustom.
	Custom []string
}

// Catalog is the read interface consumed by BuildPlan. It is satisfied by
// *catalog.Catalog from internal/catalog.
type Catalog interface {
	ByID(id string) (model.Harness, bool)
	ForMode(m model.InstallMode) []model.Harness
	ForAgent(a model.Agent) []model.Harness
}

// AgentAdapter is the superset interface needed by the install steps.
// It is satisfied by agents.Adapter from internal/agents.
type AgentAdapter interface {
	Agent() model.Agent
	InstructionsPath(homeDir string) string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string
	MCPConfigPath(homeDir, serverName string) string
	MCPStrategy() external.MCPStrategy
	VariantKey() string
}

// Registry maps agents to their adapters. It is satisfied by *agents.Registry.
type Registry interface {
	Get(agent model.Agent) (AgentAdapter, bool)
}

// Options carries the dependencies and configuration for BuildPlan.
type Options struct {
	// HomeDir is the user's home directory, passed to adapters for path resolution.
	HomeDir string
	// Registry maps agents to their concrete adapters.
	Registry Registry
	// VerifyHook is an optional function executed after a successful Apply stage.
	// A nil hook is a no-op.
	VerifyHook func() error
	// OnProgress receives progress events during installation.
	// When nil no progress events are emitted.
	OnProgress pipeline.ProgressFunc
	// embeddedSkillsFS is the fs.FS for the "embed" skill install method.
	// It is set via WithEmbeddedSkillsFS; nil means clone/npx only.
	embeddedSkillsFS fs.FS
}

// WithEmbeddedSkillsFS returns a copy of opts with the embedded skills FS set.
// Pass assets.SkillsFS from the binary entry point.
func WithEmbeddedSkillsFS(opts Options, f fs.FS) Options {
	opts.embeddedSkillsFS = f
	return opts
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
