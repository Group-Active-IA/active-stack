package planner

import (
	"reflect"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestBuildReviewPayloadLabelsSelectedAndAutoDependency(t *testing.T) {
	// User selected "skills"; "engram" and "sdd" were auto-added.
	selected := []model.Harness{
		{ID: "skills", Name: "Skills", Type: model.HarnessSkill, InstallModes: []model.InstallMode{model.ModeFull}},
	}

	resolved := ResolvedPlan{
		OrderedIDs: []string{"engram", "sdd", "skills"},
		AddedIDs:   []string{"engram", "sdd"},
	}

	payload := BuildReviewPayload(selected, resolved, nil, model.ModeFull)

	if len(payload.HarnessActions) != 3 {
		t.Fatalf("HarnessActions len = %d, want 3", len(payload.HarnessActions))
	}

	want := []HarnessAction{
		{ID: "engram", Action: "auto-dependency"},
		{ID: "sdd", Action: "auto-dependency"},
		{ID: "skills", Action: "selected"},
	}
	if !reflect.DeepEqual(payload.HarnessActions, want) {
		t.Fatalf("HarnessActions = %v, want %v", payload.HarnessActions, want)
	}
}

func TestBuildReviewPayloadAllSelectedNoAutoDeps(t *testing.T) {
	selected := []model.Harness{
		{ID: "engram", Name: "Engram", Type: model.HarnessExternal, InstallModes: []model.InstallMode{model.ModeFull}},
		{ID: "sdd", Name: "SDD", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeFull}},
	}

	resolved := ResolvedPlan{
		OrderedIDs: []string{"engram", "sdd"},
		AddedIDs:   nil,
	}

	payload := BuildReviewPayload(selected, resolved, nil, model.ModeFull)

	for _, a := range payload.HarnessActions {
		if a.Action != "selected" {
			t.Errorf("HarnessAction %q has action %q, want %q", a.ID, a.Action, "selected")
		}
	}

	if len(payload.AddedIDs) != 0 {
		t.Fatalf("AddedIDs = %v, want empty", payload.AddedIDs)
	}
}

func TestBuildReviewPayloadPropagatesAgentsAndMode(t *testing.T) {
	agents := []model.Agent{model.AgentClaude, model.AgentOpenCode}
	mode := model.ModeLite

	resolved := ResolvedPlan{
		OrderedIDs: []string{"engram"},
		AddedIDs:   nil,
	}

	payload := BuildReviewPayload(nil, resolved, agents, mode)

	if !reflect.DeepEqual(payload.Agents, agents) {
		t.Fatalf("Agents = %v, want %v", payload.Agents, agents)
	}
	if payload.Mode != mode {
		t.Fatalf("Mode = %q, want %q", payload.Mode, mode)
	}
}

func TestBuildReviewPayloadAddedIDsMirroredFromPlan(t *testing.T) {
	resolved := ResolvedPlan{
		OrderedIDs: []string{"engram", "sdd", "skills"},
		AddedIDs:   []string{"engram", "sdd"},
	}

	payload := BuildReviewPayload(nil, resolved, nil, model.ModeFull)

	if !reflect.DeepEqual(payload.AddedIDs, resolved.AddedIDs) {
		t.Fatalf("AddedIDs = %v, want %v", payload.AddedIDs, resolved.AddedIDs)
	}
}

func TestBuildReviewPayloadEmptyPlan(t *testing.T) {
	resolved := ResolvedPlan{}
	payload := BuildReviewPayload(nil, resolved, nil, model.ModeCustom)

	if len(payload.HarnessActions) != 0 {
		t.Fatalf("HarnessActions = %v, want empty for empty plan", payload.HarnessActions)
	}
}
