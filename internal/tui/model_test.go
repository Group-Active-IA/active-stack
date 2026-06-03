package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// --- Navigation tests (Pattern 2 from go-testing skill: direct Model.Update) ---

func TestWelcomeAdvancesOnEnter(t *testing.T) {
	m := newModel(ModelDeps{})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenDetection {
		t.Errorf("Screen = %v, want %v", state.Screen, ScreenDetection)
	}
}

func TestDetectionAdvancesOnEnter(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenDetection

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenAgents {
		t.Errorf("Screen = %v, want %v", state.Screen, ScreenAgents)
	}
}

func TestAgentsGoesBackOnEsc(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenAgents

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenDetection {
		t.Errorf("Screen = %v, want %v", state.Screen, ScreenDetection)
	}
}

func TestModeGoesBackOnEsc(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenAgents {
		t.Errorf("Screen = %v, want %v", state.Screen, ScreenAgents)
	}
}

// TestAtLeastOneAgentRequired verifies the guard: Enter on ScreenAgents with
// no agent selected stays on ScreenAgents.
func TestAtLeastOneAgentRequired(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenAgents
	m.Selection.Agents = nil // no agents selected

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenAgents {
		t.Errorf("Screen = %v, want %v (guard must block)", state.Screen, ScreenAgents)
	}
}

// TestLiteSkipsCustomPickerGoesToPermissions verifies that Lite mode (cursor=0)
// goes from Mode → ScreenPermissions (not ScreenCustomPicker) when a tier-capable
// agent (claude) is selected. The permissions screen replaces the old direct-to-review.
func TestLiteSkipsCustomPickerGoesToPermissions(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Cursor = 0 // ModeLite is index 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenCustomPicker {
		t.Errorf("Lite should skip custom picker, but got ScreenCustomPicker")
	}
	// With claude selected, Lite goes to ScreenPermissions, not directly ScreenReview.
	if state.Screen != ScreenPermissions {
		t.Errorf("Screen = %v, want ScreenPermissions for Lite+claude", state.Screen)
	}
}

// TestLiteSkipsPermissionsWithNonTierCapableAgents verifies that Lite mode with
// only non-tier-capable agents (gemini) goes Mode → Review (skips ScreenPermissions).
func TestLiteSkipsPermissionsWithNonTierCapableAgents(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentGemini}
	m.Cursor = 0 // ModeLite is index 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenCustomPicker {
		t.Errorf("Lite should skip custom picker, but got ScreenCustomPicker")
	}
	if state.Screen == ScreenPermissions {
		t.Errorf("Lite with only gemini should skip ScreenPermissions, but got ScreenPermissions")
	}
	if state.Screen != ScreenReview {
		t.Errorf("Screen = %v, want ScreenReview for Lite+gemini (no tier-capable agents)", state.Screen)
	}
}

// TestFullSkipsCustomPickerGoesToPermissions is the Full-mode analogue for tier-capable agents.
func TestFullSkipsCustomPickerGoesToPermissions(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Cursor = 1 // ModeFull is index 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenCustomPicker {
		t.Errorf("Full should skip custom picker, but got ScreenCustomPicker")
	}
	if state.Screen != ScreenPermissions {
		t.Errorf("Screen = %v, want ScreenPermissions for Full+claude", state.Screen)
	}
}

// TestCustomShowsPicker verifies that Custom mode (cursor=2) goes from
// Mode → CustomPicker.
func TestCustomShowsPicker(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Cursor = 2 // ModeCustom is index 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenCustomPicker {
		t.Errorf("Screen = %v, want ScreenCustomPicker for Custom", state.Screen)
	}
}

// TestCustomPickerGoesBackToMode verifies that Esc on CustomPicker goes back to Mode.
func TestCustomPickerGoesBackToMode(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenCustomPicker

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	state := updated.(Model)

	if state.Screen != ScreenMode {
		t.Errorf("Screen = %v, want ScreenMode", state.Screen)
	}
}

// TestAgentCursorDownWraps verifies that Down key increments the cursor.
func TestAgentCursorDown(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
	})
	m.Screen = ScreenAgents
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	state := updated.(Model)

	if state.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", state.Cursor)
	}
}

// TestAgentCursorDoesNotGoNegative verifies Up at 0 stays at 0.
func TestAgentCursorDoesNotGoNegative(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
	})
	m.Screen = ScreenAgents
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	state := updated.(Model)

	if state.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0 (cannot go negative)", state.Cursor)
	}
}

// TestAgentSpaceTogglesSelection verifies Space toggles an agent's selection.
func TestAgentSpaceTogglesSelection(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
	})
	m.Screen = ScreenAgents
	m.Cursor = 0

	// First toggle ON.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	state := updated.(Model)
	if len(state.Selection.Agents) != 1 || state.Selection.Agents[0] != model.AgentClaude {
		t.Errorf("after first toggle: Agents = %v, want [claude]", state.Selection.Agents)
	}

	// Second toggle OFF.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeySpace})
	state = updated.(Model)
	if len(state.Selection.Agents) != 0 {
		t.Errorf("after second toggle: Agents = %v, want []", state.Selection.Agents)
	}
}

// TestAgentWindowsSpaceTogglesSelection reproduces the Windows console driver
// behavior: bubbletea's key_windows.go maps VK_SPACE to KeyRunes (rune ' ')
// instead of KeySpace. Before the normalization in handleKey, this space was
// swallowed by the navigation case (tea.KeyUp, tea.KeyRunes) and the toggle
// never fired. This test sends the space exactly as Windows delivers it.
func TestAgentWindowsSpaceTogglesSelection(t *testing.T) {
	m := newModel(ModelDeps{
		AvailableAgents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
	})
	m.Screen = ScreenAgents
	m.Cursor = 0

	// Space as delivered by the Windows console driver: KeyRunes with rune ' '.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	state := updated.(Model)
	if len(state.Selection.Agents) != 1 || state.Selection.Agents[0] != model.AgentClaude {
		t.Errorf("Windows-style space did not toggle: Agents = %v, want [claude]", state.Selection.Agents)
	}
}

// TestCustomPickerWindowsSpaceToggles verifies the same Windows space handling
// works on the custom picker (which also toggles with the space bar).
func TestCustomPickerWindowsSpaceToggles(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenCustomPicker
	m.AvailableHarnesses = []model.Harness{
		{ID: "engram"},
		{ID: "context7"},
	}
	m.Cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	state := updated.(Model)
	if !isStringSelected(state.Selection.CustomHarnesses, "engram") {
		t.Errorf("Windows-style space did not toggle harness: CustomHarnesses = %v, want [engram]", state.Selection.CustomHarnesses)
	}
}

// TestQuitKey verifies that 'q' from the complete screen quits.
func TestQuitKey(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenComplete

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if cmd == nil {
		t.Fatal("cmd should not be nil after 'q' on complete screen")
	}
	// Evaluate the cmd — it should return a tea.QuitMsg.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("cmd() returned %T, want tea.QuitMsg", msg)
	}
}

// --- Screen-string tests (smoke) ---

func TestScreenString(t *testing.T) {
	tests := []struct {
		s    Screen
		want string
	}{
		{ScreenWelcome, "welcome"},
		{ScreenDetection, "detection"},
		{ScreenAgents, "agents"},
		{ScreenMode, "mode"},
		{ScreenCustomPicker, "custom-picker"},
		{ScreenReview, "review"},
		{ScreenInstalling, "installing"},
		{ScreenComplete, "complete"},
		{ScreenUnknown, "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Screen(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}
