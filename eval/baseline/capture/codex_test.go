package capture

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

func codexThread(t *testing.T, rel string) CodexThread {
	t.Helper()
	th, err := ParseCodexExec(sampleReader(t, rel))
	if err != nil {
		t.Fatalf("ParseCodexExec(%s): %v", rel, err)
	}
	return th
}

func callsByType(th CodexThread) map[string]int {
	m := map[string]int{}
	for _, c := range th.ToolCalls {
		m[c.Type]++
	}
	return m
}

func TestParseCodexExec_DocExample(t *testing.T) {
	th := codexThread(t, "codex/exec-doc-example.jsonl")
	if th.ID != "0199a213-81c0-7800-8aa1-bbab2a035a53" {
		t.Errorf("thread_id = %q", th.ID)
	}
	if len(th.Turns) != 1 || !th.Turns[0].Completed || th.Turns[0].Failed || th.Turns[0].Usage == nil {
		t.Fatalf("turnos = %+v, want un turno completado con uso", th.Turns)
	}
	want := usage.Logical{Input: 24763, Uncached: 315, CacheRead: 24448, Output: 122}
	if *th.Turns[0].Usage != want {
		t.Errorf("uso = %+v, want %+v (entrada 24763 incluye 24448 cacheados; sin cache por diferencia 315)", *th.Turns[0].Usage, want)
	}
	tok, av := th.Tokens()
	if !av.Available || tok.Total() != 24885 {
		t.Errorf("Tokens = %+v %+v, want 24885 (entrada 24763 + salida 122; la cacheada no se suma)", tok, av)
	}
	if got := callsByType(th); got["command_execution"] != 1 || len(th.ToolCalls) != 1 {
		t.Errorf("llamadas = %v, want una command_execution (el agent_message no es una llamada)", got)
	}
	if th.Unfinished != 1 {
		t.Errorf("sin terminar = %d, want 1 (item_1 empezo y la muestra no lo cierra)", th.Unfinished)
	}
	if len(th.Events) != 5 || th.Events[0].Type != "thread.started" || th.Events[4].Type != "turn.completed" {
		t.Errorf("eventos = %d, want 5 en orden, con el crudo conservado", len(th.Events))
	}
	if !strings.Contains(string(th.Events[4].Raw), `"cached_input_tokens":24448`) {
		t.Errorf("el evento crudo no se conservo byte a byte: %s", th.Events[4].Raw)
	}
}

func TestParseCodexExec_DelegatedParentAndChild(t *testing.T) {
	p := codexThread(t, "codex/exec-delegated-parent.jsonl")
	if len(p.Turns) != 2 {
		t.Fatalf("turnos del padre = %d, want 2", len(p.Turns))
	}
	tok, av := p.Tokens()
	if !av.Available || tok.Total() != 30000 || tok.CacheRead != 21000 || tok.Uncached != 6000 {
		t.Errorf("padre = %+v %+v, want total 30000 (22000 + 8000), cacheada 21000 y sin cache 6000 (5000 + 1000)", tok, av)
	}
	if got := callsByType(p); got["command_execution"] != 2 || got["file_change"] != 1 || len(p.ToolCalls) != 3 || p.Unfinished != 0 {
		t.Errorf("llamadas del padre = %v (%d), sin terminar %d, want 2 command_execution + 1 file_change", got, len(p.ToolCalls), p.Unfinished)
	}
	c := codexThread(t, "codex/exec-delegated-child.jsonl")
	if c.ID == p.ID || c.ID != "0199a213-81c0-7800-8aa1-bbab2a035a54" {
		t.Errorf("id del hijo = %q", c.ID)
	}
	tok, av = c.Tokens()
	if !av.Available || tok.Total() != 18000 {
		t.Errorf("hijo = %+v %+v, want 18000", tok, av)
	}
}

func TestParseCodexExec_PartialFailedAndDuplicates(t *testing.T) {
	// parcial: el turno nunca se completo
	p := codexThread(t, "codex/exec-partial.jsonl")
	if len(p.Turns) != 1 || p.Turns[0].Completed || p.Turns[0].Usage != nil {
		t.Errorf("turnos = %+v, want un turno sin completar y sin uso", p.Turns)
	}
	if tok, av := p.Tokens(); av.Available || av.Reason == "" || tok.Total() != 0 {
		t.Errorf("Tokens = %+v %+v, want no disponible con motivo (no es cero)", tok, av)
	}
	if len(p.ToolCalls) != 2 || p.Unfinished != 1 {
		t.Errorf("llamadas=%d sin terminar=%d, want 2 llamadas y 1 sin terminar (command_execution item_1)", len(p.ToolCalls), p.Unfinished)
	}
	// fallo: error y turn.failed conservados crudos
	f := codexThread(t, "codex/exec-failed.jsonl")
	if len(f.Turns) != 1 || !f.Turns[0].Failed || f.Turns[0].Completed || f.Turns[0].Usage != nil {
		t.Errorf("turno fallido = %+v", f.Turns)
	}
	if len(f.Errors) != 2 || f.Errors[0].Type != "error" || f.Errors[1].Type != "turn.failed" ||
		!strings.Contains(string(f.Errors[1].Raw), "stream disconnected before completion") {
		t.Errorf("errores = %+v, want error y turn.failed con su crudo", f.Errors)
	}
	if _, av := f.Tokens(); av.Available {
		t.Error("un turno fallido sin uso no debe dar tokens disponibles")
	}
	// duplicados
	d := codexThread(t, "codex/exec-duplicates.jsonl")
	if d.ID == "" || len(d.Turns) != 1 || d.DuplicateTurnCompletions != 1 || d.DuplicateItems != 1 || len(d.ToolCalls) != 1 {
		t.Errorf("duplicados: id=%q turnos=%d turn.completed repetidos=%d items repetidos=%d llamadas=%d, want 1 turno, 1 y 1 repetidos, 1 llamada",
			d.ID, len(d.Turns), d.DuplicateTurnCompletions, d.DuplicateItems, len(d.ToolCalls))
	}
	if tok, av := d.Tokens(); !av.Available || tok.Total() != 1100 {
		t.Errorf("uso con turn.completed repetido = %+v, want 1100 (no 2200)", tok)
	}
}

func TestParseCodexExec_EdgeCases(t *testing.T) {
	ts := `{"type":"thread.started","thread_id":"T1"}`
	cases := []struct {
		name, in string
		check    func(t *testing.T, th CodexThread, err error)
	}{
		{"vacio", "", func(t *testing.T, th CodexThread, err error) {
			if err != nil || len(th.Turns) != 0 || th.ID != "" {
				t.Errorf("%+v %v", th, err)
			}
			if _, av := th.Tokens(); av.Available {
				t.Error("sin turnos no hay tokens disponibles")
			}
		}},
		{"JSON invalido", ts + "\n{no\n", func(t *testing.T, th CodexThread, err error) {
			if err == nil || !strings.Contains(err.Error(), "línea 2") {
				t.Errorf("err = %v, want 'línea 2'", err)
			}
		}},
		{"dos hilos en un flujo", ts + "\n" + `{"type":"thread.started","thread_id":"T2"}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err == nil {
				t.Error("un flujo con dos thread.started distintos debe ser error: es un flujo por hilo")
			}
		}},
		{"evento desconocido se conserva y no cuenta", ts + "\n" + `{"type":"future.event","x":1}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err != nil || len(th.Events) != 2 || len(th.Turns) != 0 || len(th.ToolCalls) != 0 {
				t.Errorf("%+v %v", th, err)
			}
		}},
		{"turn.completed sin usage", ts + "\n" + `{"type":"turn.started"}` + "\n" + `{"type":"turn.completed"}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err != nil || len(th.Turns) != 1 || !th.Turns[0].Completed || th.Turns[0].Usage != nil {
				t.Errorf("%+v %v", th, err)
			}
			if _, av := th.Tokens(); av.Available {
				t.Error("turno completado sin usage: no disponible, no cero")
			}
		}},
		{"uso con cacheada mayor que la entrada", ts + "\n" + `{"type":"turn.started"}` + "\n" + `{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":20,"output_tokens":1,"reasoning_output_tokens":0}}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err == nil {
				t.Error("entrada cacheada mayor que la entrada debe ser error")
			}
		}},
		{"turn.started dentro de un turno abierto deja el anterior incompleto", ts + "\n" + `{"type":"turn.started"}` + "\n" + `{"type":"turn.started"}` + "\n" + `{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":0,"output_tokens":1,"reasoning_output_tokens":0}}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err != nil || len(th.Turns) != 2 || th.Turns[0].Completed || !th.Turns[1].Completed {
				t.Errorf("%+v %v", th.Turns, err)
			}
			if _, av := th.Tokens(); av.Available {
				t.Error("un turno incompleto vuelve no disponible la suma")
			}
		}},
		{"turn.failed sin turno abierto se conserva como error", ts + "\n" + `{"type":"turn.failed","error":{"message":"x"}}` + "\n", func(t *testing.T, th CodexThread, err error) {
			if err != nil || len(th.Errors) != 1 || len(th.Turns) != 0 {
				t.Errorf("%+v %v", th, err)
			}
		}},
		{"herramientas por tipo", ts + "\n" +
			`{"type":"item.completed","item":{"id":"a","type":"mcp_tool_call"}}` + "\n" +
			`{"type":"item.completed","item":{"id":"b","type":"web_search"}}` + "\n" +
			`{"type":"item.completed","item":{"id":"c","type":"reasoning"}}` + "\n" +
			`{"type":"item.completed","item":{"id":"d","type":"agent_message"}}` + "\n" +
			`{"type":"item.completed","item":{"type":"command_execution"}}` + "\n", func(t *testing.T, th CodexThread, err error) {
			got := callsByType(th)
			if err != nil || got["mcp_tool_call"] != 1 || got["web_search"] != 1 || len(th.ToolCalls) != 2 {
				t.Errorf("llamadas = %v (%d) err=%v, want mcp_tool_call y web_search; reasoning, agent_message y un item sin id no cuentan", got, len(th.ToolCalls), err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			th, err := ParseCodexExec(strings.NewReader(tc.in))
			tc.check(t, th, err)
		})
	}
}

func TestCodexParentAggregateScope_DocumentationDoesNotConfirmChildren(t *testing.T) {
	if CodexParentAggregateScope != usage.ScopeUnknown {
		t.Errorf("CodexParentAggregateScope = %q, want %q: la documentación consultada no dice si el agregado del padre incluye a los hijos", CodexParentAggregateScope, usage.ScopeUnknown)
	}
}

func TestReconcileCodex_DelegatedExampleReproduces48000(t *testing.T) {
	parent := codexThread(t, "codex/exec-delegated-parent.jsonl")
	child := codexThread(t, "codex/exec-delegated-child.jsonl")
	parents := map[string]string{child.ID: parent.ID}
	agg := int64(48000)

	tot, att, err := ReconcileCodex([]CodexThread{parent, child}, parents, &agg)
	if err != nil {
		t.Fatal(err)
	}
	if tot.Total != 48000 || tot.ByRole[RolePrincipal] != 30000 || tot.ByRole[RoleWorker] != 18000 || tot.Unattributed != 0 || tot.Partial || !tot.FullCoverage {
		t.Errorf("Totals = %+v, want 48000 (padre 30000, hijo 18000, sin atribuir 0, cobertura completa por coincidencia exacta)", tot)
	}
	if tot.Total == 66000 {
		t.Error("66000 es el doble conteo del diseno (agregado + hijo)")
	}
	if att.Partial || att.ByAgent[child.ID] != 18000 || len(att.ByAgent) != 1 {
		t.Errorf("atribucion = %+v, want el hijo con 18000 y completa", att)
	}

	// sin agregado: suma de los hilos, que son eventos distintos
	tot, _, err = ReconcileCodex([]CodexThread{parent, child}, parents, nil)
	if err != nil || tot.Total != 48000 || !tot.FullCoverage {
		t.Errorf("sin agregado: %+v err=%v, want 48000", tot, err)
	}
	// agregado que parece incluir solo al padre (30000) pero el detalle suma 48000: ambas cifras, parcial
	pOnly := int64(30000)
	tot, _, err = ReconcileCodex([]CodexThread{parent, child}, parents, &pOnly)
	if err != nil || !tot.Partial || tot.FullCoverage || tot.AggregateFigure == nil || *tot.AggregateFigure != 30000 || tot.DetailFigure == nil || *tot.DetailFigure != 48000 {
		t.Errorf("agregado de 30000: %+v err=%v, want parcial con ambas cifras (el alcance no esta comprobado, no se suman)", tot, err)
	}
	// un agregado mayor que el detalle con el alcance no comprobado tambien es parcial con ambas cifras
	big := int64(60000)
	tot, _, _ = ReconcileCodex([]CodexThread{parent, child}, parents, &big)
	if !tot.Partial || tot.Unattributed != 0 || tot.AggregateFigure == nil || *tot.AggregateFigure != 60000 {
		t.Errorf("agregado de 60000: %+v, want parcial sin atribuir 0 y ambas cifras (alcance no comprobado)", tot)
	}
}

func TestReconcileCodex_UncorrelatedAndUnavailableThreads(t *testing.T) {
	parent := codexThread(t, "codex/exec-delegated-parent.jsonl")
	child := codexThread(t, "codex/exec-delegated-child.jsonl")
	// hijo sin correlacion con el padre: su consumo queda sin atribuir y la atribucion es parcial
	tot, att, err := ReconcileCodex([]CodexThread{parent, child}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tot.Total != 48000 || tot.ByRole[RolePrincipal] != 30000 || tot.ByRole[RoleUnattributed] != 18000 || tot.ByRole[RoleWorker] != 0 {
		t.Errorf("Totals = %+v, want principal 30000 y sin_atribuir 18000", tot)
	}
	if !att.Partial || len(att.ByAgent) != 0 || len(att.Reasons) == 0 {
		t.Errorf("atribucion = %+v, want parcial, sin worker y con motivo", att)
	}
	// padre declarado que no esta entre los hilos: tampoco se correlaciona
	_, att, _ = ReconcileCodex([]CodexThread{parent, child}, map[string]string{child.ID: "otro"}, nil)
	if !att.Partial {
		t.Error("padre desconocido: la atribucion debe ser parcial")
	}
	// un hilo parcial no aporta cifras inventadas
	partial := codexThread(t, "codex/exec-partial.jsonl")
	tot, att, err = ReconcileCodex([]CodexThread{parent, partial}, map[string]string{partial.ID: parent.ID}, nil)
	if err != nil || tot.Total != 30000 || !att.Partial || len(att.Reasons) == 0 {
		t.Errorf("hilo parcial: %+v %+v err=%v, want 30000 del padre y atribucion parcial con motivo", tot, att, err)
	}
	if len(att.Entries) != 1 || len(att.ByAgent) != 0 {
		t.Errorf("hilo parcial: entradas=%v por agente=%v, want solo la del padre y ningun worker con cifra inventada", att.Entries, att.ByAgent)
	}
	// sin hilos
	if _, _, err := ReconcileCodex(nil, nil, nil); err == nil {
		t.Error("sin hilos: err = nil, want error")
	}
	neg := int64(-1)
	if _, _, err := ReconcileCodex([]CodexThread{parent}, nil, &neg); err == nil {
		t.Error("agregado negativo: err = nil, want error")
	}
}

func TestCodexTrace_ToolCallsByRoleAndUnavailableReads(t *testing.T) {
	parent := codexThread(t, "codex/exec-delegated-parent.jsonl")
	child := codexThread(t, "codex/exec-delegated-child.jsonl")
	log, err := NewCodexTraceLog(manifest.InterfaceTerminal)
	if err != nil {
		t.Fatal(err)
	}
	if err := FillCodexTrace(log, parent, RolePrincipal); err != nil {
		t.Fatal(err)
	}
	if err := FillCodexTrace(log, child, RoleWorker); err != nil {
		t.Fatal(err)
	}
	s := log.Summary()
	if !s.ToolCalls.Available || s.ToolCalls.Total != 6 || s.ToolCalls.ByRole[RolePrincipal] != 3 || s.ToolCalls.ByRole[RoleWorker] != 3 {
		t.Errorf("llamadas = %+v, want 6 (3 + 3)", s.ToolCalls)
	}
	if s.ToolCalls.ByType["command_execution"] != 4 || s.ToolCalls.ByType["file_change"] != 2 {
		t.Errorf("por tipo = %v", s.ToolCalls.ByType)
	}
	if s.Reads.Available || s.Reads.Reason == "" {
		t.Errorf("lecturas = %+v, want no disponibles con motivo (se hacen por command_execution)", s.Reads)
	}
	ide, err := NewCodexTraceLog(manifest.InterfaceIDE)
	if err != nil {
		t.Fatal(err)
	}
	if s := ide.Summary(); s.ToolCalls.Available || s.Reads.Available {
		t.Errorf("IDE: %+v, want todo no disponible", s)
	}
	if err := FillCodexTrace(ide, parent, RolePrincipal); err == nil {
		t.Error("registrar en un log sin senal debe ser error")
	}
	if _, err := NewCodexTraceLog("desktop"); err == nil {
		t.Error("interfaz fuera de la matriz: err = nil")
	}
}

func TestSignalsFor_Codex(t *testing.T) {
	term, err := SignalsFor(manifest.ClientCodex, manifest.InterfaceTerminal, "0.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if term.TokenUsage != StatusDocumented || term.ToolCalls != StatusDocumented || term.FileReads != StatusUnavailable ||
		term.PerAgentAttribution != StatusNotVerified || term.Timing != StatusUnavailable || term.Permissions != StatusNotVerified {
		t.Errorf("terminal = %+v", term)
	}
	if len(term.Notes) == 0 {
		t.Error("las notas de limites no pueden estar vacias")
	}
	ide, err := SignalsFor(manifest.ClientCodex, manifest.InterfaceIDE, "0.0.0")
	if err != nil {
		t.Fatal(err)
	}
	for name, st := range map[string]Status{"uso": ide.TokenUsage, "atribucion": ide.PerAgentAttribution, "herramientas": ide.ToolCalls, "lecturas": ide.FileReads, "permisos": ide.Permissions, "tiempo": ide.Timing} {
		if st != StatusNotVerified {
			t.Errorf("IDE %s = %q, want %q", name, st, StatusNotVerified)
		}
	}
}
