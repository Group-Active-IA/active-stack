package install_test

import (
	"context"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// fakeExternalSuccess is a shared SetExternalInstallFn handler that always succeeds.
func fakeExternalSuccess(
	_ context.Context,
	_ model.Harness,
	_ system.PlatformProfile,
	_ []external.AgentAdapter,
	_ string,
) (external.Result, error) {
	return external.Result{}, nil
}

// fakeCatalog implements install.Catalog using an in-memory harness list.
type fakeCatalog struct {
	harnesses []model.Harness
}

func (f *fakeCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// fakeRegistry returns a fixed adapter for every agent.
type fakeRegistry struct {
	adapters map[model.Agent]install.AgentAdapter
}

func (r *fakeRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// fakeAdapter satisfies install.AgentAdapter.
type fakeAdapter struct {
	agent model.Agent
}

func (a fakeAdapter) Agent() model.Agent                     { return a.agent }
func (a fakeAdapter) InstructionsPath(homeDir string) string { return homeDir + "/instr.md" }
func (a fakeAdapter) SkillsDir(homeDir string) string        { return homeDir + "/skills" }
func (a fakeAdapter) CommandsDir(homeDir string) string      { return homeDir + "/commands" }
func (a fakeAdapter) SettingsPath(homeDir string) string     { return homeDir + "/settings.json" }
func (a fakeAdapter) MCPConfigPath(homeDir, serverName string) string {
	return homeDir + "/mcp/" + serverName + ".json"
}
func (a fakeAdapter) MCPStrategy() external.MCPStrategy { return external.StrategySeparateFile }
func (a fakeAdapter) VariantKey() string                { return string(a.agent) }
func (a fakeAdapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/instr.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      base + "/commands",
	}.WithMCPConfigFn(func(serverName string) string {
		return base + "/mcp/" + serverName + ".json"
	})
}
func (a fakeAdapter) ConfigDelivery() model.ConfigDelivery { return model.ConfigDeliveryInstructions }

// buildOptions returns a minimal set of options for BuildPlan.
func buildOptions(homeDir string, reg install.Registry, verify func() error) install.Options {
	return install.Options{
		HomeDir:    homeDir,
		Registry:   reg,
		VerifyHook: verify,
	}
}

// --- Tests ---

// TestBuildPlanLiteModeSelectsLiteHarnesses verifies that Lite mode selects
// only harnesses whose InstallModes includes "lite".
func TestBuildPlanLiteModeSelectsLiteHarnesses(t *testing.T) {
	liteH := model.Harness{
		ID:           "lite-only",
		Name:         "Lite Only",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	fullH := model.Harness{
		ID:           "full-only",
		Name:         "Full Only",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{liteH, fullH}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	if !containsID(ids, "lite-only") {
		t.Errorf("expected lite-only in plan, got %v", ids)
	}
	if containsID(ids, "full-only") {
		t.Errorf("full-only should not be in Lite plan, got %v", ids)
	}
}

// TestBuildPlanCustomModeUsesCustomList verifies that Custom mode uses the
// explicit list from Intent.Custom, resolving deps from the catalog.
func TestBuildPlanCustomModeUsesCustomList(t *testing.T) {
	h1 := model.Harness{
		ID: "h1", Name: "H1", Type: model.HarnessExternal,
		External: &model.External{Method: "npm"}, InstallModes: []model.InstallMode{model.ModeFull},
	}
	h2 := model.Harness{
		ID: "h2", Name: "H2", Type: model.HarnessExternal,
		External: &model.External{Method: "npm"}, InstallModes: []model.InstallMode{model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h1, h2}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"h1"}, // only h1, not h2
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	if !containsID(ids, "h1") {
		t.Errorf("expected h1 in plan, got %v", ids)
	}
	if containsID(ids, "h2") {
		t.Errorf("h2 should not be in plan, got %v", ids)
	}
}

// TestBuildPlanAgentFiltering verifies that only harnesses supporting the
// selected agent are included.
func TestBuildPlanAgentFiltering(t *testing.T) {
	claudeOnly := model.Harness{
		ID:           "claude-skill",
		Name:         "Claude Skill",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	allAgents := model.Harness{
		ID:           "universal",
		Name:         "Universal",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		// Agents empty = all agents
	}

	cat := &fakeCatalog{harnesses: []model.Harness{claudeOnly, allAgents}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentOpenCode: fakeAdapter{agent: model.AgentOpenCode},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentOpenCode},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	if containsID(ids, "claude-skill") {
		t.Errorf("claude-skill should not appear for opencode agent, got %v", ids)
	}
	if !containsID(ids, "universal") {
		t.Errorf("universal should appear for all agents, got %v", ids)
	}
}

// TestBuildPlanTopologicalOrder verifies that a dependency is installed before
// the dependent harness in the Apply steps.
func TestBuildPlanTopologicalOrder(t *testing.T) {
	dep := model.Harness{
		ID:           "dep",
		Name:         "Dep",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	main := model.Harness{
		ID:           "main",
		Name:         "Main",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		DependsOn:    []string{"dep"},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{dep, main}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	depIdx := indexOf(ids, "dep")
	mainIdx := indexOf(ids, "main")

	if depIdx == -1 || mainIdx == -1 {
		t.Fatalf("both dep and main should be in plan, got %v", ids)
	}
	if depIdx >= mainIdx {
		t.Errorf("dep should come before main in topo order, got %v", ids)
	}
}

// TestBuildPlanReturnsStagePlan verifies that BuildPlan returns a pipeline.StagePlan
// with both Prepare and Apply slices.
func TestBuildPlanReturnsStagePlan(t *testing.T) {
	h := model.Harness{
		ID:           "h1",
		Name:         "H1",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Ensure plan embeds pipeline.StagePlan (compile-time check via use).
	var _ pipeline.StagePlan = plan.StagePlan
}

// --- helpers ---

// applyStepIDs returns the harness IDs of Apply steps.
// Step IDs have the form "<type>:<harnessID>"; we return only the harness part.
func applyStepIDs(plan pipeline.StagePlan) []string {
	var ids []string
	for _, s := range plan.Apply {
		id := s.ID()
		// Strip "<type>:" prefix if present.
		for _, prefix := range []string{"external:", "skill:", "config:", "permissions:", "verify-hook"} {
			if prefix == id {
				id = prefix
				break
			}
			if len(id) > len(prefix) && id[:len(prefix)] == prefix {
				id = id[len(prefix):]
				break
			}
		}
		ids = append(ids, id)
	}
	return ids
}

func containsID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func indexOf(ids []string, target string) int {
	for i, id := range ids {
		if id == target {
			return i
		}
	}
	return -1
}
