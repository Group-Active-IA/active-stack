package toolfix_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/toolfix"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tarea 5.4 (opsx-task-context, D18): fixtures de herramientas que corre la 8.7,
// con su registro esperado, y el evaluador que los cuenta.

func tc(t *testing.T) *catalog.TaskContext {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cat.TaskContext
}

func TestEnabledFor_FollowsTheCatalogTable(t *testing.T) {
	c := tc(t)
	tests := []struct {
		action     string
		designMock bool
		want       []string
	}{
		{"explore", false, []string{"context7"}},
		{"archive", false, nil},
		{"verify", false, []string{"playwright"}},
		{"apply", false, []string{"context7", "domain-skills", "playwright"}},
		// Las integraciones de diseño solo con el harness design-mockups activo.
		{"apply", true, []string{"context7", "design-taste-frontend-v1", "domain-skills", "emil-design-eng", "playwright", "stitch"}},
		{"explore", true, []string{"context7"}},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got := toolfix.EnabledFor(c, tt.action, tt.designMock)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("habilitadas en %s (diseño=%v) = %v, want %v", tt.action, tt.designMock, got, tt.want)
			}
		})
	}
}

func TestEvaluate_Scenarios(t *testing.T) {
	c := tc(t)
	for _, sc := range toolfix.Scenarios(c) {
		t.Run(sc.ID, func(t *testing.T) {
			got := toolfix.Evaluate(sc.Env, sc.Calls, sc.Report)
			w := sc.Want
			checks := []struct {
				name     string
				got, wnt int
			}{
				{"uso fuera de clase sin registrar", len(got.UnrecordedOutOfClass), w.UnrecordedOutOfClass},
				{"uso fuera de clase registrado", len(got.OutOfClass), w.OutOfClass},
				{"no disponible sin línea fija", len(got.MissingUnavailableLine), w.MissingUnavailableLine},
				{"repetidas sin registrar", len(got.UnrecordedRepeats), w.UnrecordedRepeats},
				{"sondeos de una integración ya no disponible", len(got.Reprobed), w.Reprobed},
			}
			for _, ck := range checks {
				if ck.got != ck.wnt {
					t.Errorf("%s = %d, want %d (%+v)", ck.name, ck.got, ck.wnt, got)
				}
			}
			if got.PresentedAsBlocked != w.PresentedAsBlocked {
				t.Errorf("presentada como bloqueada = %v, want %v", got.PresentedAsBlocked, w.PresentedAsBlocked)
			}
		})
	}
}

func TestScenarios_CoverTheSpecCases(t *testing.T) {
	have := map[string]bool{}
	for _, sc := range toolfix.Scenarios(tc(t)) {
		have[sc.ID] = true
	}
	for _, id := range []string{
		"installed-not-enabled", "enabled-absent-with-alternative", "enabled-absent-no-alternative",
		"repeated-call-no-state-change", "repeated-call-after-state-change", "probe-of-unavailable",
		"out-of-class-level-a-recorded", "out-of-class-level-a-unrecorded", "out-of-class-presented-as-blocked",
	} {
		if !have[id] {
			t.Errorf("falta el escenario %q (specs/task-tool-selection)", id)
		}
	}
}

// La línea fija de no disponible exige capacidad no cubierta y camino alternativo (o none).
func TestEvaluate_UnavailableLineShape(t *testing.T) {
	env := toolfix.Env{Action: "explore", Enabled: []string{"context7"}, Unavailable: []string{"context7"}}
	calls := []toolfix.Call{{Integration: "context7", Tool: "mcp__context7__query-docs", Input: "react hooks", Unavailable: true}}
	tests := []struct {
		name    string
		report  string
		missing int
	}{
		{"con alternativa", "unavailable: context7 — uncovered: current library docs; alternative: read the vendored docs in docs/\n", 0},
		{"sin alternativa, con none", "unavailable: context7 — uncovered: current library docs; alternative: none\n", 0},
		{"sin la capacidad no cubierta", "unavailable: context7 — alternative: none\n", 1},
		{"sin alternativa declarada", "unavailable: context7 — uncovered: current library docs;\n", 1},
		{"sin la línea", "context7 was not available\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toolfix.Evaluate(env, calls, tt.report)
			if len(got.MissingUnavailableLine) != tt.missing {
				t.Errorf("sin línea fija = %v, want %d", got.MissingUnavailableLine, tt.missing)
			}
		})
	}
}

// 8.7 (corrida real): un MCP que no arranca no expone herramientas, así que el
// agente nunca lo llama. La integración declarada no disponible en el entorno
// exige igual la línea fija.
func TestEvaluate_UnavailableWithoutAnyCall(t *testing.T) {
	env := toolfix.Env{Action: "explore", Enabled: []string{"context7"}, Unavailable: []string{"context7"}}
	got := toolfix.Evaluate(env, nil, "")
	if len(got.MissingUnavailableLine) != 1 || got.MissingUnavailableLine[0] != "context7" {
		t.Errorf("sin llamada ni línea: MissingUnavailableLine = %v, se esperaba [context7]", got.MissingUnavailableLine)
	}
	got = toolfix.Evaluate(env, nil, "unavailable: context7 — uncovered: current library docs; alternative: none\n")
	if len(got.MissingUnavailableLine) != 0 {
		t.Errorf("con la línea: MissingUnavailableLine = %v, se esperaba vacío", got.MissingUnavailableLine)
	}
	// Una integración no habilitada en la acción no exige la línea.
	got = toolfix.Evaluate(toolfix.Env{Action: "verify", Enabled: []string{"playwright"}, Unavailable: []string{"context7"}}, nil, "")
	if len(got.MissingUnavailableLine) != 0 {
		t.Errorf("no habilitada: MissingUnavailableLine = %v, se esperaba vacío", got.MissingUnavailableLine)
	}
}
