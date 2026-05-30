package tui

import (
	"context"
	"errors"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// TestGateTUI_MissingDep_DoesNotStartProgress verifies that when a required
// dep for the selected harnesses is missing, pressing Enter on ScreenReview
// transitions to ScreenComplete with an error (NOT ScreenInstalling), and the
// RunPlanFn goroutine is never started.
func TestGateTUI_MissingDep_DoesNotStartProgress(t *testing.T) {
	h := model.Harness{
		ID:           "ext-npm",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	runPlanCalled := false

	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{h}},
		BuildPlanFn: func(_ install.Catalog, _ install.Intent, _ install.Options) (install.Plan, error) {
			return install.Plan{}, nil
		},
		RunPlanFn: func(_ install.Plan, _ *progressBridge, _ func(tea.Msg)) {
			runPlanCalled = true
		},
	}

	// Inject a fake detector that reports npm missing.
	restoreFn := setTUIDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		return system.DependencyReport{
			Dependencies:    deps,
			AllPresent:      false,
			MissingRequired: []string{"npm"},
		}
	})
	defer restoreFn()

	m := newModel(deps)
	m.Screen = ScreenReview
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Selection.Mode = model.ModeLite
	m.ResolvedIDs = []string{"external:ext-npm"}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenInstalling {
		t.Fatal("gate must NOT transition to ScreenInstalling when deps are missing")
	}
	if state.Screen != ScreenComplete {
		t.Fatalf("Screen = %v, want ScreenComplete (error path)", state.Screen)
	}
	if state.ExecutionResult.Err == nil {
		t.Fatal("ExecutionResult.Err must be set when gate aborts")
	}
	if runPlanCalled {
		t.Fatal("RunPlanFn must NOT be called when gate aborts")
	}
}

// TestGateTUI_AllDepsPresent_StartsInstall verifies that when all deps are
// present the gate passes and the model transitions to ScreenInstalling.
func TestGateTUI_AllDepsPresent_StartsInstall(t *testing.T) {
	h := model.Harness{
		ID:           "ext-npm",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	runPlanCalled := false

	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{h}},
		BuildPlanFn: func(_ install.Catalog, _ install.Intent, _ install.Options) (install.Plan, error) {
			return install.Plan{}, nil
		},
		RunPlanFn: func(_ install.Plan, _ *progressBridge, _ func(tea.Msg)) {
			runPlanCalled = true
		},
	}

	// Inject a detector that reports all deps present.
	restoreFn := setTUIDetectDepsForFn(func(_ context.Context, deps []system.Dependency) system.DependencyReport {
		return system.DependencyReport{
			Dependencies: deps,
			AllPresent:   true,
		}
	})
	defer restoreFn()

	m := newModel(deps)
	m.Screen = ScreenReview
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Selection.Mode = model.ModeLite
	m.ResolvedIDs = []string{"external:ext-npm"}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenInstalling {
		t.Fatalf("all deps present: Screen = %v, want ScreenInstalling", state.Screen)
	}
	_ = runPlanCalled // RunPlanFn is called in goroutine, not synchronously
}

// TestSelectTUIHarnesses_MatchesCanonical verifies the C-24 unification: the TUI
// selection path (selectTUIHarnesses) resolves the SAME set as the canonical
// install.SelectHarnesses for a Custom intent, including the forced
// security-first harness.
func TestSelectTUIHarnesses_MatchesCanonical(t *testing.T) {
	cat := &fakeTUICatalog{harnesses: []model.Harness{
		{
			ID:           install.SecurityFirstHarnessID,
			Type:         model.HarnessConfig,
			InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
			Agents:       []model.Agent{model.AgentClaude},
		},
		{
			ID:           "engram",
			Type:         model.HarnessExternal,
			External:     &model.External{Method: "npm"},
			InstallModes: []model.InstallMode{model.ModeFull},
			Agents:       []model.Agent{model.AgentClaude},
		},
	}}

	intent := install.Intent{
		Mode:   model.ModeCustom,
		Agents: []model.Agent{model.AgentClaude},
		Custom: []string{"engram"}, // permissions NOT requested
	}

	got := selectTUIHarnesses(cat, intent)

	want, err := install.SelectHarnesses(cat, intent)
	if err != nil {
		t.Fatalf("install.SelectHarnesses() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("TUI selection diverged from canonical:\n got=%v\nwant=%v",
			tuiHarnessIDs(got), tuiHarnessIDs(want))
	}

	found := false
	for _, h := range got {
		if h.ID == install.SecurityFirstHarnessID {
			found = true
		}
	}
	if !found {
		t.Errorf("expected security-first harness forced, got %v", tuiHarnessIDs(got))
	}
}

func tuiHarnessIDs(harnesses []model.Harness) []string {
	ids := make([]string, 0, len(harnesses))
	for _, h := range harnesses {
		ids = append(ids, h.ID)
	}
	return ids
}

// ── Fake catalog used only by TUI gate tests ──────────────────────────────────

type fakeTUICatalog struct{ harnesses []model.Harness }

func (f *fakeTUICatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *fakeTUICatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *fakeTUICatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// ensure errors package is used (used in gate implementation later)
var _ = errors.New

// ensure pipeline package is used
var _ pipeline.ExecutionResult
