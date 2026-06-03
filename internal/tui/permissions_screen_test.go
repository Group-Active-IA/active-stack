package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ── Task 5.2: anyTierCapable ─────────────────────────────────────────────────

// TestAnyTierCapable verifies the helper that determines whether a set of agents
// contains at least one tier-capable agent (claude or opencode).
// Spec: "ScreenPermissions es condicional según haya agentes tier-capaces".
func TestAnyTierCapable(t *testing.T) {
	tests := []struct {
		name   string
		agents []model.Agent
		want   bool
	}{
		// Tier-capable only → true.
		{"claude only → true", []model.Agent{model.AgentClaude}, true},
		{"opencode only → true", []model.Agent{model.AgentOpenCode}, true},
		// Non tier-capable only → false.
		{"gemini only → false", []model.Agent{model.AgentGemini}, false},
		{"vscode only → false", []model.Agent{model.AgentVSCode}, false},
		{"gemini + vscode → false", []model.Agent{model.AgentGemini, model.AgentVSCode}, false},
		// Empty → false.
		{"empty → false", []model.Agent{}, false},
		// Mixed: at least one tier-capable → true.
		{"gemini + claude → true", []model.Agent{model.AgentGemini, model.AgentClaude}, true},
		{"gemini + opencode → true", []model.Agent{model.AgentGemini, model.AgentOpenCode}, true},
		{"claude + opencode → true", []model.Agent{model.AgentClaude, model.AgentOpenCode}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := anyTierCapable(tt.agents)
			if got != tt.want {
				t.Errorf("anyTierCapable(%v) = %v, want %v", tt.agents, got, tt.want)
			}
		})
	}
}

// ── Task 5.3: ScreenPermissions appears in flow with tier-capable agents ─────

// TestScreenPermissionsAppearsForTierCapableAgentsLite verifies that with
// claude selected in Lite mode, the flow goes Mode → ScreenPermissions → Review.
// Spec: "Lite y Full con un agente tier-capaz pasan por la pantalla de tier".
func TestScreenPermissionsAppearsForTierCapableAgentsLite(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Cursor = 0 // Lite

	// ScreenMode → Enter → should land on ScreenPermissions.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenPermissions {
		t.Fatalf("Lite+claude: Screen = %v, want ScreenPermissions", state.Screen)
	}

	// ScreenPermissions → Enter → should advance to ScreenReview.
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenReview {
		t.Errorf("After confirming tier: Screen = %v, want ScreenReview", state.Screen)
	}
}

// TestScreenPermissionsAppearsForOpencode verifies opencode is tier-capable
// and triggers the permissions screen.
// Spec: "opencode es tier-capaz y dispara la pantalla".
func TestScreenPermissionsAppearsForOpencode(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentOpenCode}
	m.Cursor = 1 // Full

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenPermissions {
		t.Errorf("Full+opencode: Screen = %v, want ScreenPermissions", state.Screen)
	}
}

// TestScreenPermissionsAppearsForCustomWithTierCapableAgent verifies that
// Custom mode with a tier-capable agent shows ScreenPermissions after the picker.
// Spec: "Custom con un agente tier-capaz pasa por la pantalla de tier después del picker".
func TestScreenPermissionsAppearsForCustomWithTierCapableAgent(t *testing.T) {
	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{
			{ID: "permissions", Name: "Permissions", Type: model.HarnessConfig,
				InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
				Agents:       []model.Agent{model.AgentClaude}},
		}},
	}
	m := newModel(deps)
	m.Screen = ScreenCustomPicker
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Selection.Mode = model.ModeCustom
	m.AvailableHarnesses = []model.Harness{
		{ID: "h1", Name: "H1", Type: model.HarnessExternal,
			External: &model.External{Method: "npm"},
			InstallModes: []model.InstallMode{model.ModeFull}},
	}

	// CustomPicker → Enter → ScreenPermissions (not Review).
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen != ScreenPermissions {
		t.Errorf("Custom+claude after picker: Screen = %v, want ScreenPermissions", state.Screen)
	}
}

// ── Task 5.4: ScreenPermissions is SKIPPED for non-tier-capable agents ────

// TestScreenPermissionsSkippedForGeminiOnly verifies that with only gemini
// selected, the flow goes Mode → Review (skipping ScreenPermissions).
// The intent still carries TierBalanceado by default.
// Spec: "Solo agentes no tier-capaces saltean la pantalla de tier".
func TestScreenPermissionsSkippedForGeminiOnly(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentGemini}
	m.Cursor = 0 // Lite

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenPermissions {
		t.Errorf("gemini-only: ScreenPermissions should be skipped, but got it")
	}
	if state.Screen != ScreenReview {
		t.Errorf("gemini-only: Screen = %v, want ScreenReview (direct skip)", state.Screen)
	}
}

// TestScreenPermissionsSkippedForGeminiAndVSCode is the multi-agent non-tier-capable case.
func TestScreenPermissionsSkippedForGeminiAndVSCode(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentGemini, model.AgentVSCode}
	m.Cursor = 1 // Full

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)

	if state.Screen == ScreenPermissions {
		t.Errorf("gemini+vscode: ScreenPermissions should be skipped, but got it")
	}
	if state.Screen != ScreenReview {
		t.Errorf("gemini+vscode: Screen = %v, want ScreenReview", state.Screen)
	}

	// The intent must carry TierBalanceado by default.
	intent := state.Selection.BuildIntent()
	if intent.Tier != model.TierBalanceado {
		t.Errorf("intent.Tier = %q, want %q (default for skipped screen)", intent.Tier, model.TierBalanceado)
	}
}

// ── Bugfix: Space selects the tier under the cursor without advancing ────────

// TestPermissionsSpaceSelectsTierWithoutAdvancing verifies that pressing Space
// on ScreenPermissions sets Selection.Tier to the tier under the cursor and
// STAYS on the screen (does not advance to Review). This mirrors the Space
// toggle used by the multi-select screens; previously Space was a dead key here
// and only Enter (which also advances) could set the tier.
func TestPermissionsSpaceSelectsTierWithoutAdvancing(t *testing.T) {
	for cursor, want := range tierOrder {
		t.Run(string(want), func(t *testing.T) {
			m := newModel(ModelDeps{})
			m.Screen = ScreenPermissions
			m.Cursor = cursor

			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
			state := updated.(Model)

			if state.Selection.Tier != want {
				t.Errorf("after Space: Selection.Tier = %q, want %q", state.Selection.Tier, want)
			}
			// Space must NOT advance — that is Enter's job.
			if state.Screen != ScreenPermissions {
				t.Errorf("after Space: Screen = %v, want ScreenPermissions (must not advance)", state.Screen)
			}
		})
	}
}

// TestPermissionsSpaceThenEnterKeepsSelection verifies the two-step flow:
// Space picks a non-default tier, Enter confirms it and advances to Review
// carrying the picked tier.
func TestPermissionsSpaceThenEnterKeepsSelection(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenPermissions
	m.Cursor = 0 // estricto (a non-default tier)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	state := updated.(Model)
	updated, _ = state.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state = updated.(Model)

	if state.Screen != ScreenReview {
		t.Errorf("after Space+Enter: Screen = %v, want ScreenReview", state.Screen)
	}
	if state.Selection.Tier != model.TierEstricto {
		t.Errorf("after Space+Enter: Selection.Tier = %q, want %q", state.Selection.Tier, model.TierEstricto)
	}
}

// ── Task 5.5: Render balanceado preselected ──────────────────────────────────

// TestPermissionsScreenPreSelectsBalanceado verifies that the initial cursor
// is on balanceado and the view renders all three tiers.
// Spec: "La pantalla preselecciona balanceado".
func TestPermissionsScreenPreSelectsBalanceado(t *testing.T) {
	m := newModel(ModelDeps{})
	m.Screen = ScreenPermissions
	m.Cursor = defaultTierCursor() // should be balanceado

	view := m.View()

	// All three tiers must be visible.
	for _, tier := range []string{"estricto", "balanceado", "bypass"} {
		if !strings.Contains(view, string(tier)) {
			t.Errorf("view missing tier %q", tier)
		}
	}

	// The initial cursor position should correspond to balanceado.
	wantCursor := defaultTierCursor()
	if tierOrder[wantCursor] != model.TierBalanceado {
		t.Errorf("defaultTierCursor() points to %q, want balanceado", tierOrder[wantCursor])
	}
}

// ── Task 5.6: Bypass warning ────────────────────────────────────────────────

// TestPermissionsScreenBypassShowsWarning verifies that moving the cursor to
// bypass shows a warning, while other tiers do not show it.
// Spec: "Bypass muestra advertencia".
func TestPermissionsScreenBypassShowsWarning(t *testing.T) {
	// Find bypass index in tierOrder.
	bypassIdx := -1
	balanceadoIdx := -1
	for i, tier := range tierOrder {
		if tier == model.TierBypass {
			bypassIdx = i
		}
		if tier == model.TierBalanceado {
			balanceadoIdx = i
		}
	}
	if bypassIdx == -1 {
		t.Fatal("bypass not found in tierOrder")
	}

	// Cursor on bypass — warning should appear.
	m := newModel(ModelDeps{})
	m.Screen = ScreenPermissions
	m.Cursor = bypassIdx
	bypassView := m.View()

	if !strings.Contains(bypassView, "Bypass") {
		t.Errorf("cursor on bypass: expected warning text containing 'Bypass', got:\n%s", bypassView)
	}

	// Cursor on balanceado — no warning.
	m.Cursor = balanceadoIdx
	balanceadoView := m.View()

	if strings.Contains(balanceadoView, bypassWarning) {
		t.Errorf("cursor on balanceado: unexpected warning text, got:\n%s", balanceadoView)
	}
}

// ── Task 5.7: Tier feeds Intent, harnesses unchanged ─────────────────────────

// TestPermissionsSelectionFeedsIntent verifies that choosing a tier on
// ScreenPermissions sets Selection.Tier and BuildIntent() includes it.
// Also verifies the harnesses are NOT changed by the tier choice.
// Spec: "El intent lleva el tier seleccionado", "El tier no cambia la selección de harnesses".
func TestPermissionsSelectionFeedsIntent(t *testing.T) {
	tests := []struct {
		name     string
		wantTier model.PermissionTier
		cursor   int
	}{
		{"estricto", model.TierEstricto, 0},
		{"bypass", model.TierBypass, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel(ModelDeps{})
			m.Screen = ScreenPermissions
			m.Cursor = tt.cursor
			// Pre-set some harnesses to verify they don't change.
			m.Selection.CustomHarnesses = []string{"permissions", "engram"}

			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			state := updated.(Model)

			if state.Selection.Tier != tt.wantTier {
				t.Errorf("Selection.Tier = %q, want %q", state.Selection.Tier, tt.wantTier)
			}

			intent := state.Selection.BuildIntent()
			if intent.Tier != tt.wantTier {
				t.Errorf("Intent.Tier = %q, want %q", intent.Tier, tt.wantTier)
			}

			// Harness selection unchanged.
			for _, id := range []string{"permissions", "engram"} {
				if !isStringSelected(state.Selection.CustomHarnesses, id) {
					t.Errorf("harness %q was removed from selection after tier choice", id)
				}
			}
		})
	}
}
