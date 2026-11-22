package install_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// [EQ] Fase 7 (R-20) T-F7-2: una instalacion que no elige las funciones nuevas
// (Engram Cloud: hooks de memoria y AUTOSYNC) deja los cuatro agentes byte a
// byte como estaban. El golden se tomo ANTES de conectar esas funciones; si el
// cableado agrega una escritura a una instalacion que no las pidio, falla.
//
// Para regenerarlo (solo con una razon documentada):
//
//	go test ./internal/install -run TestInstallWithoutTeamFunctionsIsByteIdentical -update
var updateGolden = flag.Bool("update", false, "update golden files")

const goldenPath = "testdata/no_team_install.golden.json"

func scenarios() map[string]install.Intent {
	out := map[string]install.Intent{}
	for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode, model.AgentAntigravity} {
		out["lite/"+string(agent)] = install.Intent{Agents: []model.Agent{agent}, Mode: model.ModeLite, Tier: model.TierBalanceado}
		out["custom-sin-cloud/"+string(agent)] = install.Intent{
			Agents: []model.Agent{agent}, Mode: model.ModeCustom, Tier: model.TierBalanceado,
			Custom: []string{"engram", "sdd-orchestrator", "design-mockups"},
		}
	}
	return out
}

func TestInstallWithoutTeamFunctionsIsByteIdentical(t *testing.T) {
	stubNetwork(t)
	got := map[string]map[string]string{}
	for name, intent := range scenarios() {
		home := t.TempDir()
		seedHome(t, home, newRealRegistry(t))
		res := runRealInstall(t, home, intent)
		if res.Err != nil {
			t.Fatalf("%s: install failed: %v", name, res.Err)
		}
		got[name] = treeHashes(t, home)
	}

	if *updateGolden {
		raw, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, append(raw, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden missing (run with -update once, before wiring): %v", err)
	}
	var want map[string]map[string]string
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	for name, hashes := range got {
		if !reflect.DeepEqual(hashes, want[name]) {
			for _, k := range sortedKeys(hashes) {
				if hashes[k] != want[name][k] {
					t.Errorf("%s: %s differs from the golden", name, k)
				}
			}
			for _, k := range sortedKeys(want[name]) {
				if _, ok := hashes[k]; !ok {
					t.Errorf("%s: %s is in the golden but was not written", name, k)
				}
			}
		}
	}
}
