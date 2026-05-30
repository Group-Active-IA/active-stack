package install_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// permissionsHarness returns a catalog entry mirroring the real "permissions"
// harness: a security-first config harness supported by a subset of agents.
func permissionsHarness() model.Harness {
	return model.Harness{
		ID:           "permissions",
		Name:         "Permissions (security-first)",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode, model.AgentGemini, model.AgentVSCode},
	}
}

// TestBuildPlanCustomForcesPermissions verifies the C-21 guarantee: in Custom
// mode, the security-first "permissions" harness is always installed even when
// the user did not list it in Intent.Custom — provided the selected agent
// supports it.
func TestBuildPlanCustomForcesPermissions(t *testing.T) {
	other := model.Harness{
		ID: "h1", Name: "H1", Type: model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{other, permissionsHarness()}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"h1"}, // NOTE: permissions deliberately omitted
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	if !containsID(ids, "permissions") {
		t.Errorf("permissions must be force-included in Custom mode, got %v", ids)
	}
	if !containsID(ids, "h1") {
		t.Errorf("user-selected h1 must still be present, got %v", ids)
	}
}

// TestBuildPlanCustomDoesNotDuplicatePermissions verifies that when the user
// DID select permissions explicitly, it appears exactly once (no duplicate from
// the forced injection).
func TestBuildPlanCustomDoesNotDuplicatePermissions(t *testing.T) {
	cat := &fakeCatalog{harnesses: []model.Harness{permissionsHarness()}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"permissions"}, // user picked it explicitly
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	count := 0
	for _, id := range ids {
		if id == "permissions" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("permissions must appear exactly once, got %d (ids=%v)", count, ids)
	}
}

// TestBuildPlanCustomPermissionsDroppedForUnsupportedAgent documents the
// boundary of the guarantee: if the selected agent does NOT support the
// permissions overlay (e.g. codex/cursor), the agent filter drops it — we
// cannot force an overlay that does not exist for that agent.
func TestBuildPlanCustomPermissionsDroppedForUnsupportedAgent(t *testing.T) {
	other := model.Harness{
		ID: "h1", Name: "H1", Type: model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
		Agents:       []model.Agent{model.AgentCodex},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{other, permissionsHarness()}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentCodex: fakeAdapter{agent: model.AgentCodex},
	}}

	intent := install.Intent{
		Agents: []model.Agent{model.AgentCodex}, // codex not in permissions.agents
		Mode:   model.ModeCustom,
		Custom: []string{"h1"},
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan.StagePlan)
	if containsID(ids, "permissions") {
		t.Errorf("permissions cannot be forced for an agent that does not support it, got %v", ids)
	}
}

// ── C-24: single source of truth — install.SelectHarnesses ────────────────────

// TestSelectHarnessesForcesPermissions verifies the canonical selector forces
// the security-first harness in Custom mode even when not requested. This is the
// ONE place the rule lives; the TUI gate, the picker and the verify hook all
// delegate here (C-24).
func TestSelectHarnessesForcesPermissions(t *testing.T) {
	cat := &fakeCatalog{harnesses: []model.Harness{
		{ID: "engram", Type: model.HarnessExternal, Agents: []model.Agent{model.AgentClaude}},
		permissionsHarness(),
	}}

	intent := install.Intent{
		Mode:   model.ModeCustom,
		Agents: []model.Agent{model.AgentClaude},
		Custom: []string{"engram"}, // permissions deliberately omitted
	}

	got, err := install.SelectHarnesses(cat, intent)
	if err != nil {
		t.Fatalf("SelectHarnesses() error = %v", err)
	}

	if !harnessSetContains(got, install.SecurityFirstHarnessID) {
		t.Errorf("expected %q forced in Custom mode, got %v", install.SecurityFirstHarnessID, idsOf(got))
	}
}

// TestSelectHarnessesUnknownIDIsError verifies the canonical selector keeps the
// strict semantics: an unknown id in Custom mode is an error (safer than the
// laxer "ignore unknown" behaviour the TUI duplicates used to have).
func TestSelectHarnessesUnknownIDIsError(t *testing.T) {
	cat := &fakeCatalog{harnesses: []model.Harness{permissionsHarness()}}

	intent := install.Intent{
		Mode:   model.ModeCustom,
		Agents: []model.Agent{model.AgentClaude},
		Custom: []string{"does-not-exist"},
	}

	if _, err := install.SelectHarnesses(cat, intent); err == nil {
		t.Fatalf("expected error for unknown harness id, got nil")
	}
}

func harnessSetContains(harnesses []model.Harness, id string) bool {
	for _, h := range harnesses {
		if h.ID == id {
			return true
		}
	}
	return false
}

func idsOf(harnesses []model.Harness) []string {
	ids := make([]string, 0, len(harnesses))
	for _, h := range harnesses {
		ids = append(ids, h.ID)
	}
	return ids
}
