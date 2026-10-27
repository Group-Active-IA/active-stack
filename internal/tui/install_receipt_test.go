package tui

import (
	"errors"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

func TestInstallReceiptOnlyAfterSuccessfulCompletion(t *testing.T) {
	calls := 0
	m := newModel(ModelDeps{OnInstallComplete: func(intent install.Intent) error {
		calls++
		if intent.Mode != model.ModeCustom || len(intent.Custom) != 1 || intent.Custom[0] != "design-system" {
			t.Fatalf("intent=%+v", intent)
		}
		return nil
	}})
	m.Selection = Selection{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeCustom, CustomHarnesses: []string{"design-system"}}
	updated, _ := m.Update(doneMsg{result: pipeline.ExecutionResult{Err: errors.New("failed")}})
	if calls != 0 {
		t.Fatal("receipt written after failed install")
	}
	updated, _ = updated.(Model).Update(doneMsg{result: pipeline.ExecutionResult{}})
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}
