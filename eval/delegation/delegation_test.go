package delegation

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Cada fixture de testdata/ es un caso de las tareas 4.1 (retornos), 4.2 (decisión) y 4.3
// (combinación) de 07 con el veredicto que exigen la spec selective-delegation y el contrato de 06.
// Agregar un caso es agregar un JSON.

func TestReturns_VerdictMatchesSpec(t *testing.T) {
	for _, f := range glob(t, "retornos") {
		c, err := LoadReturnCase(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		t.Run(c.ID, func(t *testing.T) {
			got := JudgeReturn(c.Assignment, c.Raw, c.Observed)
			if got.Status != c.Want.Status {
				t.Errorf("estado = %q, esperado %q (motivos: %v)", got.Status, c.Want.Status, got.Reasons)
			}
			if got.Step != c.Want.Step {
				t.Errorf("paso = %q, esperado %q (motivos: %v)", got.Step, c.Want.Step, got.Reasons)
			}
			if got.Action != c.Want.Action {
				t.Errorf("acción = %q, esperada %q", got.Action, c.Want.Action)
			}
			if !slices.Equal(got.Escalate, c.Want.Escalate) {
				t.Errorf("escala con %q, esperado %q", got.Escalate, c.Want.Escalate)
			}
		})
	}
}

func TestDecisions_FollowTheRule(t *testing.T) {
	for _, f := range glob(t, "decisiones") {
		c, err := LoadDecisionCase(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		t.Run(c.ID, func(t *testing.T) {
			if got := Decide(c.Task); got != c.Want {
				t.Errorf("decisión = %+v, esperada %+v", got, c.Want)
			}
		})
	}
}

func TestCombinations_VerdictMatchesSpec(t *testing.T) {
	for _, f := range glob(t, "combinacion") {
		c, err := LoadCombinationCase(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		t.Run(c.ID, func(t *testing.T) {
			got := JudgeCombination(c.Assignments, c.Integration)
			if got.Status != c.Want.Status {
				t.Errorf("estado = %q, esperado %q (motivos: %v)", got.Status, c.Want.Status, got.Reasons)
			}
			if got.Integrate != c.Want.Integrate {
				t.Errorf("integrar = %v, esperado %v", got.Integrate, c.Want.Integrate)
			}
			if !slices.Equal(got.FailedFor, c.Want.FailedFor) {
				t.Errorf("fallo contra %q, esperado %q", got.FailedFor, c.Want.FailedFor)
			}
		})
	}
}

// Los fixtures que piden las tareas 4.1, 4.2 y 4.3 deben existir.
func TestFixtures_CoverTasks(t *testing.T) {
	need := map[string][]string{
		"retornos": {
			"completo-corroborado", "completo-sin-corroborar", "reejecutado-coordinador-sin-captura",
			"prosa-libre", "partida-distinta", "ajeno-en-partida", "confirmado-fuera-de-alcance",
			"renombre-fuera-de-alcance", "nuevo-fuera-de-alcance", "modificados-incompleto",
			"tope-agotado", "sin-escritura",
			// Revisión del PR #13: bordes de la propiedad contra la partida y del juicio de recibos.
			"revierte-ajeno-en-partida", "revierte-propio-en-partida", "revierte-propio-sin-declarar",
			"borra-ajeno-en-partida",
			"ajeno-en-partida-sin-huella", "partida-sin-huella", "ruta-no-normalizada",
			"recibo-corroborado-falla",
		},
		"decisiones": {
			"verificacion-breve-independencia", "independencia-delegation-apagado", "modulo-ausente",
			"archivo-caliente", "tercer-worker-escritura",
		},
		"combinacion": {"rompe-al-combinar"},
	}
	for group, ids := range need {
		for _, id := range ids {
			if m, err := filepath.Glob(filepath.Join("testdata", group, id+".json")); err != nil || len(m) == 0 {
				t.Errorf("falta el fixture %s/%s", group, id)
			}
		}
	}
}

// Sin el sobre, ningún retorno queda verificado, sea cual sea lo observado.
func TestJudgeReturn_NoEnvelopeNeverVerified(t *testing.T) {
	for _, raw := range []string{"", "Listo, tests verdes.", `{"status":"completed"}`, `null`} {
		got := JudgeReturn(Assignment{Attempt: 1}, []byte(raw), Observed{})
		if got.Status != Unverified || got.Step != StepEnvelope {
			t.Errorf("%q: estado %q paso %q, esperado %q en %q", raw, got.Status, got.Step, Unverified, StepEnvelope)
		}
	}
}

func glob(t *testing.T, group string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", group, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("sin fixtures en testdata/%s: %v", group, err)
	}
	return files
}

// HotFiles es la lista del módulo instalado (assets/opsx/delegation-module.md), no una copia que
// pueda divergir.
func TestHotFiles_MatchTheModule(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "harness", "config", "assets", "opsx", "delegation-module.md"))
	if err != nil {
		t.Fatal(err)
	}
	const lead = "Never split these across concurrent workers:"
	_, after, ok := strings.Cut(string(raw), lead)
	if !ok {
		t.Fatalf("módulo sin %q", lead)
	}
	block, _, _ := strings.Cut(after, "\n\n")
	parts := strings.Split(block, "`")
	var listed []string
	for j := 1; j < len(parts); j += 2 {
		listed = append(listed, parts[j])
	}
	if !slices.Equal(listed, HotFiles) {
		t.Errorf("HotFiles = %q, el módulo lista %q", HotFiles, listed)
	}
}
