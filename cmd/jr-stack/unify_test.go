package main

import (
	"reflect"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// cmdFakeCatalog is a minimal install.Catalog for cmd-level tests.
type cmdFakeCatalog struct{ harnesses []model.Harness }

func (f cmdFakeCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f cmdFakeCatalog) ForMode(mode model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(mode) {
			out = append(out, h)
		}
	}
	return out
}

func (f cmdFakeCatalog) ForAgent(agent model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(agent) {
			out = append(out, h)
		}
	}
	return out
}

// TestCollectSelectedHarnessesMatchesCanonical verifies the C-24 unification:
// the verify-hook selection path (collectSelectedHarnesses) resolves the SAME
// set as the canonical install.SelectHarnesses, including the forced
// security-first harness in Custom mode.
func TestCollectSelectedHarnessesMatchesCanonical(t *testing.T) {
	cat := cmdFakeCatalog{harnesses: []model.Harness{
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

	got := collectSelectedHarnesses(cat, intent)

	want, err := install.SelectHarnesses(cat, intent)
	if err != nil {
		t.Fatalf("install.SelectHarnesses() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("verify-hook selection diverged from canonical:\n got=%v\nwant=%v",
			cmdIDsOf(got), cmdIDsOf(want))
	}

	if !cmdContains(got, install.SecurityFirstHarnessID) {
		t.Errorf("expected security-first harness forced, got %v", cmdIDsOf(got))
	}
}

func cmdContains(harnesses []model.Harness, id string) bool {
	for _, h := range harnesses {
		if h.ID == id {
			return true
		}
	}
	return false
}

func cmdIDsOf(harnesses []model.Harness) []string {
	ids := make([]string, 0, len(harnesses))
	for _, h := range harnesses {
		ids = append(ids, h.ID)
	}
	return ids
}
