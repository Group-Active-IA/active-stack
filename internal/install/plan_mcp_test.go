package install_test

// C-28: Starter.MCPs[] wiring into the install flow (D5).
//
// Governance ALTO: every MCP write step MUST expose a non-nil Rollback().
// Backup before write; idempotent merge via filemerge.

import (
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// mcpProjectAdapter is a test double that implements install.AgentAdapter with
// target-aware path resolution for MCP. It resolves the project MCP path to
// <root>/.mcp.json (MCPStrategySingleFileMerge), mirroring the real Claude adapter.
type mcpProjectAdapter struct {
	agent model.Agent
}

func (a mcpProjectAdapter) Agent() model.Agent { return a.agent }
func (a mcpProjectAdapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "CLAUDE.md")
}
func (a mcpProjectAdapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
}
func (a mcpProjectAdapter) CommandsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "commands")
}
func (a mcpProjectAdapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "settings.json")
}
func (a mcpProjectAdapter) MCPConfigPath(homeDir, serverName string) string {
	return filepath.Join(homeDir, ".claude", "mcp", serverName+".json")
}

// MCPStrategy returns the legacy machine strategy (unchanged).
func (a mcpProjectAdapter) MCPStrategy() external.MCPStrategy { return external.StrategySeparateFile }
func (a mcpProjectAdapter) VariantKey() string                { return string(a.agent) }
func (a mcpProjectAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a mcpProjectAdapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	switch t {
	case model.Project:
		return model.AgentPaths{
			InstructionsPath: filepath.Join(base, ".claude", "CLAUDE.md"),
			SkillsDir:        filepath.Join(base, ".claude", "skills"),
			SettingsPath:     filepath.Join(base, ".claude", "settings.json"),
			CommandsDir:      filepath.Join(base, ".claude", "commands"),
		}.WithMCPConfigFn(func(_ string) string {
			return filepath.Join(base, ".mcp.json")
		}).WithMCPStrategy(model.MCPStrategySingleFileMerge)
	default:
		return model.AgentPaths{
			InstructionsPath: filepath.Join(base, ".claude", "CLAUDE.md"),
			SkillsDir:        filepath.Join(base, ".claude", "skills"),
			SettingsPath:     filepath.Join(base, ".claude", "settings.json"),
			CommandsDir:      filepath.Join(base, ".claude", "commands"),
		}.WithMCPConfigFn(func(serverName string) string {
			return filepath.Join(base, ".claude", "mcp", serverName+".json")
		}).WithMCPStrategy(model.MCPStrategySeparateFile)
	}
}

// mcpOpenCodeAdapter is a test double for OpenCode (project MCP merges into opencode.json).
type mcpOpenCodeAdapter struct{}

func (a mcpOpenCodeAdapter) Agent() model.Agent                              { return model.AgentOpenCode }
func (a mcpOpenCodeAdapter) InstructionsPath(homeDir string) string          { return filepath.Join(homeDir, ".opencode", "AGENTS.md") }
func (a mcpOpenCodeAdapter) SkillsDir(homeDir string) string                 { return filepath.Join(homeDir, ".opencode", "skills") }
func (a mcpOpenCodeAdapter) CommandsDir(homeDir string) string               { return filepath.Join(homeDir, ".opencode", "commands") }
func (a mcpOpenCodeAdapter) SettingsPath(homeDir string) string              { return filepath.Join(homeDir, ".opencode", "opencode.json") }
func (a mcpOpenCodeAdapter) MCPConfigPath(homeDir, _ string) string          { return filepath.Join(homeDir, ".opencode", "opencode.json") }
func (a mcpOpenCodeAdapter) MCPStrategy() external.MCPStrategy                { return external.StrategyMergeIntoSettings }
func (a mcpOpenCodeAdapter) VariantKey() string                              { return "opencode" }
func (a mcpOpenCodeAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryPrimaryAgent
}
func (a mcpOpenCodeAdapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	dir := filepath.Join(base, ".opencode")
	return model.AgentPaths{
		InstructionsPath: filepath.Join(dir, "AGENTS.md"),
		SkillsDir:        filepath.Join(dir, "skills"),
		SettingsPath:     filepath.Join(dir, "opencode.json"),
	}.WithMCPConfigFn(func(_ string) string {
		return filepath.Join(dir, "opencode.json")
	}).WithMCPStrategy(model.MCPStrategyMergeIntoSettings)
}

// TestBuildPlanWithStarter_MCPs_EmitsNxMSteps asserts that a starter with N MCPs
// and M focused agents produces N×M MCP write steps, each with a non-nil Rollback().
//
// Spec: "flow produces N×M MCP write steps routed via the target-aware MCPConfigPath
// AND each step exposes a Rollback()"
func TestBuildPlanWithStarter_MCPs_EmitsNxMSteps(t *testing.T) {
	projectRoot := t.TempDir()

	// Stub snapshot and config to avoid real FS side effects.
	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return nil
	})
	defer restoreConfig()

	// N=2 MCPs, M=2 focused agents (Claude, OpenCode) → want 4 MCP write steps.
	starter := &model.Starter{
		ID:   "active-ia",
		Name: "Active IA",
		MCPs: []model.MCP{
			{Name: "context7", Command: "npx", Args: []string{"-y", "@upstash/context7-mcp"}},
			{Name: "engram", Command: "uvx", Args: []string{"engram-mcp"}},
		},
	}

	// Empty harness catalog (we only care about MCP steps here).
	cat := &fakeCatalog{harnesses: []model.Harness{}}

	claudeAdapter := mcpProjectAdapter{agent: model.AgentClaude}
	openCodeAdapter := mcpOpenCodeAdapter{}

	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude:   claudeAdapter,
		model.AgentOpenCode: openCodeAdapter,
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     projectRoot,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
		Starter:     starter,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Count MCP write steps (identified by "mcp:" prefix).
	mcpSteps := mcpWriteStepsFromPlan(plan.StagePlan)
	wantCount := len(starter.MCPs) * 2 // N × M (M=2 agents)
	if len(mcpSteps) != wantCount {
		t.Errorf("MCP write steps = %d, want %d (N=%d × M=%d)",
			len(mcpSteps), wantCount, len(starter.MCPs), 2)
	}

	// Every MCP step must implement Rollback().
	for _, step := range mcpSteps {
		_, ok := step.(rollbacker)
		if !ok {
			t.Errorf("step %q does not implement Rollback()", step.ID())
		}
	}
}

// TestBuildPlanWithStarter_NoMCPs_EmitsZeroMCPSteps asserts that a starter with
// an empty MCPs slice produces no MCP write steps.
func TestBuildPlanWithStarter_NoMCPs_EmitsZeroMCPSteps(t *testing.T) {
	projectRoot := t.TempDir()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	starter := &model.Starter{
		ID:   "empty",
		Name: "Empty",
		MCPs: nil, // empty
	}

	cat := &fakeCatalog{harnesses: []model.Harness{}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: mcpProjectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     projectRoot,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
		Starter:     starter,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	mcpSteps := mcpWriteStepsFromPlan(plan.StagePlan)
	if len(mcpSteps) != 0 {
		t.Errorf("expected 0 MCP steps for empty MCPs[], got %d", len(mcpSteps))
	}
}

// TestBuildPlanWithStarter_Nil_EmitsZeroMCPSteps asserts that a nil Starter in
// options produces no MCP steps (backward-compat for callers without starters).
func TestBuildPlanWithStarter_Nil_EmitsZeroMCPSteps(t *testing.T) {
	projectRoot := t.TempDir()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: mcpProjectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:  projectRoot,
		Registry: reg,
		Starter:  nil, // no starter
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	mcpSteps := mcpWriteStepsFromPlan(plan.StagePlan)
	if len(mcpSteps) != 0 {
		t.Errorf("expected 0 MCP steps for nil Starter, got %d", len(mcpSteps))
	}
}

// TestBuildPlanWithStarter_MCPStep_Rollback_Restores asserts that an MCP write
// step's Rollback() can be called and restores the pre-write state via the
// backup infrastructure.
//
// Governance ALTO gate: every MCP write step must have a working Rollback().
func TestBuildPlanWithStarter_MCPStep_Rollback_Restores(t *testing.T) {
	projectRoot := t.TempDir()

	var manifestCaptured backup.Manifest
	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		manifestCaptured = backup.Manifest{/* sentinel non-zero */}
		return manifestCaptured, nil
	})
	defer restoreSnap()

	var rollbackCalled bool
	restoreRestore := install.SetRestoreFn(func(m backup.Manifest) error {
		rollbackCalled = true
		return nil
	})
	defer restoreRestore()

	starter := &model.Starter{
		ID:   "active-ia",
		Name: "Active IA",
		MCPs: []model.MCP{
			{Name: "context7", Command: "npx"},
		},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: mcpProjectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     projectRoot,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
		Starter:     starter,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Run Prepare (snapshot) then find and rollback the MCP step.
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatalf("Prepare step failed: %v", err)
		}
	}

	mcpSteps := mcpWriteStepsFromPlan(plan.StagePlan)
	if len(mcpSteps) == 0 {
		t.Fatal("no MCP steps to rollback")
	}

	for _, step := range mcpSteps {
		if rb, ok := step.(interface{ Rollback() error }); ok {
			if err := rb.Rollback(); err != nil {
				t.Errorf("Rollback() returned error: %v", err)
			}
		}
	}

	if !rollbackCalled {
		t.Error("Rollback() did not call restoreFn — snapshot not restored")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────

// rollbacker is a step that can be rolled back.
type rollbacker interface {
	Rollback() error
}

// mcpWriteStepsFromPlan returns all Apply steps whose ID starts with "mcp:".
func mcpWriteStepsFromPlan(plan pipeline.StagePlan) []pipeline.Step {
	var out []pipeline.Step
	for _, s := range plan.Apply {
		id := s.ID()
		if len(id) > 4 && id[:4] == "mcp:" {
			out = append(out, s)
		}
	}
	return out
}
