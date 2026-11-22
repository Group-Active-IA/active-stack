package main

import (
	"reflect"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/update"
)

// [EN] Fase 9 (R-20) T-F9-5: el intent Custom con engram-cloud hace ida y
// vuelta por installation.json y el comando headless lo incluye en --custom.
func TestUpdateRoundTripsCustomEngramCloud(t *testing.T) {
	home := t.TempDir()
	err := update.SaveIntent(home, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeCustom,
		Custom: []string{"engram-cloud"},
		Tier:   model.TierBalanceado,
	})
	if err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	sel := state.Agents[model.AgentClaude]
	if sel.Mode != model.ModeCustom || !reflect.DeepEqual(sel.Custom, []string{"engram-cloud"}) {
		t.Fatalf("selection = %+v", sel)
	}
	got := installArgs(model.AgentClaude, sel, home)
	want := []string{"install", "--headless", "--yes", "--no-self-install", "--home", home, "--agent", "claude", "--mode", "custom", "--tier", "balanceado", "--custom", "engram-cloud"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v\nwant   %v", got, want)
	}
}
