package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// permTUIHarness mirrors the real security-first permissions harness for TUI tests.
func permTUIHarness() model.Harness {
	return model.Harness{
		ID:           "permissions",
		Name:         "Permissions (security-first)",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode, model.AgentGemini, model.AgentVSCode},
	}
}

func otherTUIHarness() model.Harness {
	return model.Harness{
		ID:           "h1",
		Name:         "H1",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
}

// enterCustomPicker drives the model from ScreenMode (Custom selected) into the
// custom picker, returning the resulting model.
func enterCustomPicker(t *testing.T, deps ModelDeps) Model {
	t.Helper()
	m := newModel(deps)
	m.Screen = ScreenMode
	m.Selection.Agents = []model.Agent{model.AgentClaude}
	m.Cursor = 2 // ModeCustom is index 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	state := updated.(Model)
	if state.Screen != ScreenCustomPicker {
		t.Fatalf("expected ScreenCustomPicker, got %v", state.Screen)
	}
	return state
}

// TestCustomPickerPreSelectsPermissions verifies that entering the custom picker
// starts with permissions already selected (C-21).
func TestCustomPickerPreSelectsPermissions(t *testing.T) {
	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{otherTUIHarness(), permTUIHarness()}},
	}

	state := enterCustomPicker(t, deps)

	if !isStringSelected(state.Selection.CustomHarnesses, "permissions") {
		t.Errorf("permissions must be pre-selected on picker entry, got %v", state.Selection.CustomHarnesses)
	}
}

// TestCustomPickerCannotDeselectPermissions verifies that pressing Space on the
// permissions row keeps it selected (C-21 — non-deselectable).
func TestCustomPickerCannotDeselectPermissions(t *testing.T) {
	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{otherTUIHarness(), permTUIHarness()}},
	}

	state := enterCustomPicker(t, deps)

	// Move the cursor onto the permissions row.
	permIdx := -1
	for i, h := range state.AvailableHarnesses {
		if h.ID == "permissions" {
			permIdx = i
			break
		}
	}
	if permIdx == -1 {
		t.Fatal("permissions not present in AvailableHarnesses")
	}
	state.Cursor = permIdx

	// Press Space — must NOT remove permissions.
	updated, _ := state.Update(tea.KeyMsg{Type: tea.KeySpace})
	state = updated.(Model)

	if !isStringSelected(state.Selection.CustomHarnesses, "permissions") {
		t.Errorf("Space on permissions must NOT deselect it, got %v", state.Selection.CustomHarnesses)
	}
}

// TestCustomPickerRendersPermissionsAsRequired verifies the picker view shows
// permissions as forced ([x] + a "requerido" / security-first marker).
func TestCustomPickerRendersPermissionsAsRequired(t *testing.T) {
	deps := ModelDeps{
		Catalog: &fakeTUICatalog{harnesses: []model.Harness{otherTUIHarness(), permTUIHarness()}},
	}

	state := enterCustomPicker(t, deps)
	out := state.View()

	if !strings.Contains(out, "requerido") {
		t.Errorf("picker view must mark permissions as 'requerido', got:\n%s", out)
	}
}

// TestSelectTUIHarnessesForcesPermissions verifies the preflight mirror forces
// permissions in Custom mode even when not listed (C-21 consistency).
func TestSelectTUIHarnessesForcesPermissions(t *testing.T) {
	cat := &fakeTUICatalog{harnesses: []model.Harness{otherTUIHarness(), permTUIHarness()}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"h1"}, // permissions omitted
	}

	got := selectTUIHarnesses(cat, intent)

	found := false
	for _, h := range got {
		if h.ID == "permissions" {
			found = true
		}
	}
	if !found {
		t.Errorf("selectTUIHarnesses must force permissions in Custom mode, got %v", got)
	}
}
