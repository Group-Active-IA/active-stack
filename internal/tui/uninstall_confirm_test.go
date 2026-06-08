package tui

import (
	"fmt"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// TestUninstallConfirm_EscDoesNotExecute verifies that Esc on the confirm screen
// goes back without executing uninstall.
func TestUninstallConfirm_EscDoesNotExecute(t *testing.T) {
	executed := false
	deps := ModelDeps{
		RunUninstall: func(_ headless.ParsedUninstallFlags, _ io.Writer) int {
			executed = true
			return 0
		},
	}
	m := newModel(deps)
	m.Screen = ScreenUninstallConfirm
	m.prevScreen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if executed {
		t.Error("RunUninstall must not be called when Esc is pressed")
	}
	if state.Screen != ScreenUninstallStrategy {
		t.Errorf("Screen = %v, want ScreenUninstallStrategy", state.Screen)
	}
}

// TestUninstallConfirm_EnterLaunchesGoroutine verifies that Enter on confirm
// screen transitions to ScreenUninstalling and the goroutine eventually reports
// the exit code.
func TestUninstallConfirm_EnterLaunchesGoroutine(t *testing.T) {
	deps := ModelDeps{
		RunUninstall: func(_ headless.ParsedUninstallFlags, w io.Writer) int {
			fmt.Fprintln(w, "Removing harnesses...")
			return 0
		},
	}
	m := newModel(deps)
	m.Screen = ScreenUninstallConfirm
	m.uninstallSel = uninstallSelections{
		agents:   []model.Agent{model.AgentClaude},
		mode:     model.ModeLite,
		strategy: uninstall.StrategyTargeted,
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenUninstalling {
		t.Fatalf("Screen = %v, want ScreenUninstalling", state.Screen)
	}
	if cmd == nil {
		t.Fatal("cmd must not be nil — goroutine should be listening")
	}

	// Drain cmd loop until done.
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		updated, cmd = state.Update(msg)
		state = updated.(Model)
	}

	if !state.uninstallDone {
		t.Error("uninstallDone should be true after drain")
	}
	if state.uninstallExitCode != 0 {
		t.Errorf("uninstallExitCode = %d, want 0", state.uninstallExitCode)
	}
}

// TestUninstallConfirm_ExitCode1Reported verifies that a failing RunUninstall
// (exit code 1) is reflected in the model.
func TestUninstallConfirm_ExitCode1Reported(t *testing.T) {
	deps := ModelDeps{
		RunUninstall: func(_ headless.ParsedUninstallFlags, _ io.Writer) int {
			return 1
		},
	}
	m := newModel(deps)
	m.Screen = ScreenUninstallConfirm

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	// Drain.
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		updated, cmd = state.Update(msg)
		state = updated.(Model)
	}

	if !state.uninstallDone {
		t.Error("uninstallDone should be true")
	}
	if state.uninstallExitCode != 1 {
		t.Errorf("uninstallExitCode = %d, want 1", state.uninstallExitCode)
	}

	// View should show failure message.
	view := state.viewUninstalling()
	if !contains(view, "failed") && !contains(view, "Failed") {
		t.Errorf("view should indicate failure:\n%s", view)
	}
}
