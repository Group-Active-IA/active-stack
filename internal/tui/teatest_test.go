package tui

// Teatest integration test for the installing→complete full flow.
// Skill applied: go-testing Pattern 3 (teatest.NewTestModel) — see decision tree:
//   "Testing TUI component? → Full flow? → Use teatest.NewTestModel()".
//
// This test complements review_test.go#TestInstallFlowCompletes (direct
// Model.Update drain) with a teatest driver that runs the real Bubbletea
// event loop and asserts on rendered output.

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// TestFullFlowInstallToComplete drives the TUI from ScreenReview through
// ScreenInstalling to ScreenComplete using teatest.NewTestModel.
//
// Flow:
//  1. Model starts at ScreenReview with a pre-resolved plan (two fake steps).
//  2. Enter key → startInstall → ScreenInstalling.
//  3. Fake RunPlanFn emits two progress events via the bridge, closes it, then
//     delivers a doneMsg via the TestModel's program Send channel.
//  4. Test waits for the final output to contain "Installation complete!".
func TestFullFlowInstallToComplete(t *testing.T) {
	wantIDs := []string{"snapshot", "engram"}

	// tmRef is filled after teatest.NewTestModel returns so the fake RunPlanFn
	// closure can send doneMsg through the real tea.Program.
	var tmRef *teatest.TestModel

	fakeRun := func(plan install.Plan, bridge *progressBridge, _ func(tea.Msg)) {
		go func() {
			defer bridge.close()

			// Emit two progress events through the bridge (model drains via listenCmd).
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

		// Deliver the doneMsg through the running tea.Program.
		// We spin briefly until tmRef is non-nil (it's set synchronously below
		// before any message can arrive, but the goroutine might race).
		go func() {
			deadline := time.Now().Add(2 * time.Second)
			for tmRef == nil && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if tmRef != nil {
				tmRef.Send(doneMsg{result: pipeline.ExecutionResult{}})
			}
		}()
	}

	deps := ModelDeps{
		Catalog:     &fakeCatalog{},
		Registry:    &fakeRegistry{},
		BuildPlanFn: fakeBuildPlanWith(wantIDs),
		RunPlanFn:   fakeRun,
	}

	// Build model pre-positioned at ScreenReview with resolved IDs already set.
	m := newModel(deps)
	m.Screen = ScreenReview
	m.ResolvedIDs = wantIDs
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Selection.Mode = model.ModeLite

	// teatest.NewTestModel spins up the real Bubbletea event loop in background.
	// Pattern from go-testing skill §Pattern 3.
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 24),
	)
	tmRef = tm // expose to fakeRun goroutine BEFORE any Send

	// Send Enter → transitions ScreenReview → ScreenInstalling → (async) ScreenComplete.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait until the rendered output contains the completion frame.
	// WaitFor polls Output() (live stream), WithDuration sets the overall timeout.
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Installation complete!"))
		},
		teatest.WithDuration(5*time.Second),
		teatest.WithCheckInterval(20*time.Millisecond),
	)

	// On ScreenComplete the model only quits on 'q' or ctrl+c.
	// Send 'q' to terminate the program cleanly.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Wait for the program to finish (tea.Quit was issued by the model).
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second)).(Model)
	if !ok {
		t.Fatal("FinalModel is not a tui.Model")
	}

	if final.Screen != ScreenComplete {
		t.Errorf("final screen = %v, want ScreenComplete", final.Screen)
	}

	if final.ExecutionResult.Err != nil {
		t.Errorf("ExecutionResult.Err = %v, want nil", final.ExecutionResult.Err)
	}

	// Verify the complete view rendered the success text.
	view := final.View()
	if !strings.Contains(view, "Installation complete!") {
		t.Errorf("complete view missing success text; got:\n%s", view)
	}
}
