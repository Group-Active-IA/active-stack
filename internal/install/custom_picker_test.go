package install_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestCustomPickerHarnesses verifies that CustomPickerHarnesses returns the
// ForMode(ModeCustom) set filtered by agents — the "what can be picked"
// universe (design D5), as opposed to SelectHarnesses' "what will be
// installed" answer. permissions is included because it is non-starter-only
// and supports the requested agent; CustomPickerHarnesses does not force or
// split anything, it is a pure read.
func TestCustomPickerHarnesses(t *testing.T) {
	cat := &fakeCatalog{harnesses: []model.Harness{
		{ID: "openspec", Name: "OpenSpec", InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}, Agents: []model.Agent{model.AgentClaude}},
		{ID: "permissions", Name: "Permissions", InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}, Agents: []model.Agent{model.AgentClaude}},
		{ID: "gemini-only", Name: "Gemini Only", InstallModes: []model.InstallMode{model.ModeFull}, Agents: []model.Agent{model.AgentGemini}},
	}}

	got := install.CustomPickerHarnesses(cat, []model.Agent{model.AgentClaude})

	ids := make(map[string]bool, len(got))
	for _, h := range got {
		ids[h.ID] = true
	}

	if !ids["openspec"] {
		t.Errorf("CustomPickerHarnesses() = %+v, want to contain openspec", got)
	}
	if !ids["permissions"] {
		t.Errorf("CustomPickerHarnesses() = %+v, want to contain permissions (non-starter-only, supports claude)", got)
	}
	if ids["gemini-only"] {
		t.Errorf("CustomPickerHarnesses() = %+v, did not expect gemini-only harness for claude-only request", got)
	}
}
