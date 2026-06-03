package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// fakeCatalog satisfies install.Catalog for testing.
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
	if m == model.ModeCustom {
		return f.harnesses
	}
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

// fakeRegistry satisfies install.Registry for testing.
type fakeRegistry struct{}

func (r *fakeRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	return nil, false // no real adapters needed for review-only tests
}

// fakeBuildPlan returns a plan with the given ordered step IDs.
func fakeBuildPlanWith(orderedIDs []string) func(install.Catalog, install.Intent, install.Options) (install.Plan, error) {
	return func(cat install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
		steps := make([]pipeline.Step, len(orderedIDs))
		for i, id := range orderedIDs {
			steps[i] = &fakeStep{id: id}
		}
		return install.Plan{
			StagePlan: pipeline.StagePlan{Apply: steps},
		}, nil
	}
}

// fakeBuildPlanError always returns an error.
func fakeBuildPlanError(msg string) func(install.Catalog, install.Intent, install.Options) (install.Plan, error) {
	return func(cat install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
		return install.Plan{}, errors.New(msg)
	}
}

// fakeStep is a minimal pipeline.Step for testing.
type fakeStep struct{ id string }

func (s *fakeStep) ID() string  { return s.id }
func (s *fakeStep) Run() error  { return nil }

// TestReviewShowsResolvedPlanInOrder verifies that entering review calls
// BuildPlan and populates ResolvedIDs in topological order.
// Uses AgentGemini (not tier-capable) so ScreenPermissions is skipped and
// ScreenMode → ScreenReview is direct.
func TestReviewShowsResolvedPlanInOrder(t *testing.T) {
	wantIDs := []string{"snapshot", "sdd-orchestrator", "engram"}

	deps := ModelDeps{
		Catalog:     &fakeCatalog{},
		Registry:    &fakeRegistry{},
		BuildPlanFn: fakeBuildPlanWith(wantIDs),
	}
	m := newModel(deps)
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentGemini} // not tier-capable → skip ScreenPermissions
	m.Selection.Mode = model.ModeLite

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenReview {
		t.Fatalf("Screen = %v, want ScreenReview", state.Screen)
	}
	if state.ReviewErr != nil {
		t.Fatalf("ReviewErr = %v, want nil", state.ReviewErr)
	}
	if len(state.ResolvedIDs) != len(wantIDs) {
		t.Fatalf("ResolvedIDs len = %d, want %d", len(state.ResolvedIDs), len(wantIDs))
	}
	for i, id := range wantIDs {
		if state.ResolvedIDs[i] != id {
			t.Errorf("ResolvedIDs[%d] = %q, want %q", i, state.ResolvedIDs[i], id)
		}
	}
}

// TestReviewShowsErrorWithNoInstall verifies that a BuildPlan error is
// displayed on ScreenReview and prevents advancing to ScreenInstalling.
// Uses AgentGemini (not tier-capable) so ScreenPermissions is skipped.
func TestReviewShowsErrorWithNoInstall(t *testing.T) {
	deps := ModelDeps{
		Catalog:     &fakeCatalog{},
		Registry:    &fakeRegistry{},
		BuildPlanFn: fakeBuildPlanError("unknown harness \"foo\""),
	}
	m := newModel(deps)
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentGemini} // not tier-capable → skip ScreenPermissions
	m.Selection.Mode = model.ModeLite

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenReview {
		t.Fatalf("Screen = %v, want ScreenReview", state.Screen)
	}
	if state.ReviewErr == nil {
		t.Fatal("ReviewErr = nil, want non-nil error")
	}

	// Press Enter — should NOT advance past review because ReviewErr is set.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenReview {
		t.Errorf("Screen = %v, want ScreenReview (error blocks install)", state.Screen)
	}
}

// TestReviewViewContainsStepIDs verifies that the review view renders step IDs.
// Uses AgentGemini (not tier-capable) so ScreenPermissions is skipped.
func TestReviewViewContainsStepIDs(t *testing.T) {
	deps := ModelDeps{
		Catalog:     &fakeCatalog{},
		Registry:    &fakeRegistry{},
		BuildPlanFn: fakeBuildPlanWith([]string{"snapshot", "sdd-orchestrator"}),
	}
	m := newModel(deps)
	m.Screen = ScreenMode
	m.Selection.Mode = model.ModeLite
	m.Selection.Agents = []model.Agent{model.AgentGemini} // not tier-capable → skip ScreenPermissions

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	view := state.View()
	for _, id := range []string{"snapshot", "sdd-orchestrator"} {
		if !contains(view, id) {
			t.Errorf("view does not contain step ID %q", id)
		}
	}
}

// TestInstallFlowCompletes verifies the installing→complete transition with a
// fake orchestrator that immediately sends a doneMsg.
func TestInstallFlowCompletes(t *testing.T) {
	wantIDs := []string{"snapshot", "engram"}

	fakeRun := func(plan install.Plan, bridge *progressBridge, send func(tea.Msg)) {
		// Emit two progress events then close.
		go func() {
			defer bridge.close()
			bridge.OnProgress(pipeline.ProgressEvent{
				StepID: "snapshot",
				Stage:  pipeline.StagePrepare,
				Status: pipeline.StepStatusSucceeded,
			})
			bridge.OnProgress(pipeline.ProgressEvent{
				StepID: "engram",
				Stage:  pipeline.StageApply,
				Status: pipeline.StepStatusSucceeded,
			})
		}()
	}

	deps := ModelDeps{
		Catalog:     &fakeCatalog{},
		Registry:    &fakeRegistry{},
		BuildPlanFn: fakeBuildPlanWith(wantIDs),
		RunPlanFn:   fakeRun,
	}
	m := newModel(deps)
	m.Screen = ScreenReview
	m.ResolvedIDs = wantIDs
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Selection.Mode = model.ModeLite

	// Press Enter → transition to ScreenInstalling.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenInstalling {
		t.Fatalf("Screen = %v, want ScreenInstalling", state.Screen)
	}

	// Drain progress events.
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		updated, cmd = state.Update(msg)
		state = updated.(Model)
	}

	// After draining, send doneMsg.
	updated, _ = state.Update(doneMsg{result: pipeline.ExecutionResult{}})
	state = updated.(Model)

	if state.Screen != ScreenComplete {
		t.Errorf("Screen = %v, want ScreenComplete", state.Screen)
	}
}

// TestInstallFlowFailureRendersError verifies that a failed run renders the
// error on the complete screen.
func TestInstallFlowFailureRendersError(t *testing.T) {
	m := newModel(ModelDeps{
		BuildPlanFn: fakeBuildPlanWith([]string{"step1"}),
	})
	m.Screen = ScreenComplete
	m.ExecutionResult = pipeline.ExecutionResult{
		Err: errors.New("apply failed"),
	}

	view := m.View()
	if !contains(view, "failed") {
		t.Errorf("complete view should contain 'failed', got:\n%s", view)
	}
}

// contains is a helper for substring checks in view output.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
