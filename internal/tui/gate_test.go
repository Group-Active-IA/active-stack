package tui

import (
	"context"
	"errors"
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
