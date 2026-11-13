package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

const samplesDir = "../fixtures/samples"

func sampleBytes(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(samplesDir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("muestra %s: %v", rel, err)
	}
	return b
}

func sampleReader(t *testing.T, rel string) *strings.Reader {
	t.Helper()
	return strings.NewReader(string(sampleBytes(t, rel)))
}

func parseOTel(t *testing.T, rel string) ClaudeOTel {
	t.Helper()
	o, err := ParseClaudeOTel(sampleReader(t, rel))
	if err != nil {
		t.Fatalf("ParseClaudeOTel(%s): %v", rel, err)
	}
	return o
}

// jsonl arma un flujo de eventos de OpenTelemetry desde mapas de atributos.
func jsonl(t *testing.T, events ...map[string]any) *strings.Reader {
	t.Helper()
	var sb strings.Builder
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return strings.NewReader(sb.String())
}

func apiEvent(id, qs string, in, out, cr, cc int, extra map[string]any) map[string]any {
	m := map[string]any{
		"event.name": "api_request", "request_id": id, "query_source": qs, "model": "claude-sonnet-5",
		"input_tokens": in, "output_tokens": out, "cache_read_tokens": cr, "cache_creation_tokens": cc,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func logicalTotal(l usage.Logical) int64 { return l.Total() }

func TestParseClaudeOTel_DelegatedSample(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	if len(o.Requests) != 4 {
		t.Fatalf("solicitudes = %d, want 4 (2 principal + 2 worker)", len(o.Requests))
	}
	if o.Duplicates != 0 || len(o.Incomplete) != 0 {
		t.Errorf("duplicados=%d incompletos=%v, want 0 y ninguno", o.Duplicates, o.Incomplete)
	}
	first := o.Requests[0]
	want := usage.Logical{Input: 9200, Uncached: 200, CacheRead: 0, CacheCreation: 9000, Output: 300}
	if first.Usage != want || first.RequestID != "req_011MAIN0001" || first.QuerySource != "repl_main_thread" || first.AgentID != "" {
		t.Errorf("primera solicitud = %+v, want entrada logica 9200 (200 + 9000 de creacion), salida 300, principal", first)
	}
	w := o.Requests[1]
	if w.AgentID != "a3f2c1b4d5e6f7a80" || w.AgentName != "custom" || logicalTotal(w.Usage) != 12000 {
		t.Errorf("solicitud del worker = %+v, want agent_id a3f2..., agent.name custom, total 12000", w)
	}
	if len(o.ToolResults) != 7 {
		t.Errorf("tool_result = %d, want 7 (4 principal + 3 worker)", len(o.ToolResults))
	}
	if len(o.SubagentCloses) != 1 || o.SubagentCloses[0].TotalTokens != 8000 || o.SubagentCloses[0].AgentID != "a3f2c1b4d5e6f7a80" {
		t.Errorf("cierres = %+v, want uno con total_tokens 8000", o.SubagentCloses)
	}
}

func TestParseClaudeOTel_DuplicatesAndPartials(t *testing.T) {
	d := parseOTel(t, "claude/otel-duplicates.jsonl")
	if len(d.Requests) != 2 || d.Duplicates != 1 {
		t.Fatalf("duplicados: solicitudes=%d duplicados=%d, want 2 y 1", len(d.Requests), d.Duplicates)
	}
	var sum int64
	for _, r := range d.Requests {
		sum += logicalTotal(r.Usage)
	}
	if sum != 40000 {
		t.Errorf("suma con la solicitud repetida = %d, want 40000 (9500 + 30500, la repetida no suma)", sum)
	}
	p := parseOTel(t, "claude/otel-partial-old-version.jsonl")
	if len(p.Requests) != 2 {
		t.Errorf("solicitudes validas = %d, want 2 (la que no trae tokens no se cuenta como cero)", len(p.Requests))
	}
	if len(p.Incomplete) != 1 || !strings.Contains(p.Incomplete[0], "req_011MAIN0002") {
		t.Errorf("incompletos = %v, want la solicitud req_011MAIN0002 sin tokens", p.Incomplete)
	}
}

func TestParseClaudeOTel_IncompleteAndMalformedInput(t *testing.T) {
	// cada campo de tokens es obligatorio; ausente no es cero
	for _, field := range []string{"input_tokens", "output_tokens", "cache_read_tokens", "cache_creation_tokens"} {
		e := apiEvent("req_1", "repl_main_thread", 1, 2, 3, 4, nil)
		delete(e, field)
		o, err := ParseClaudeOTel(jsonl(t, e))
		if err != nil || len(o.Requests) != 0 || len(o.Incomplete) != 1 {
			t.Errorf("sin %s: solicitudes=%d incompletos=%v err=%v, want 0 solicitudes y 1 incompleto", field, len(o.Requests), o.Incomplete, err)
		}
	}
	// tokens negativos o no numericos son incompletos, no error de lectura
	for _, bad := range []any{-5, "abc", nil} {
		e := apiEvent("req_2", "repl_main_thread", 1, 2, 3, 4, map[string]any{"output_tokens": bad})
		o, err := ParseClaudeOTel(jsonl(t, e))
		if err != nil || len(o.Requests) != 0 || len(o.Incomplete) != 1 {
			t.Errorf("output_tokens=%v: solicitudes=%d incompletos=%v err=%v, want incompleto", bad, len(o.Requests), o.Incomplete, err)
		}
	}
	// numeros como texto numerico se aceptan; prefijo claude_code. en el nombre tambien
	e := apiEvent("req_3", "repl_main_thread", 1, 2, 3, 4, map[string]any{"input_tokens": "10", "event.name": "claude_code.api_request"})
	o, err := ParseClaudeOTel(jsonl(t, e))
	if err != nil || len(o.Requests) != 1 || o.Requests[0].Usage.Uncached != 10 {
		t.Errorf("tokens como texto y nombre con prefijo: %+v err=%v, want 1 solicitud con 10 sin cache", o.Requests, err)
	}
	// lineas en blanco, eventos de otros tipos y vacio
	in := "\n" + `{"event.name":"user_prompt","prompt_length":5}` + "\n\n"
	o, err = ParseClaudeOTel(strings.NewReader(in))
	if err != nil || o.Ignored != 1 || len(o.Requests) != 0 {
		t.Errorf("evento de otro tipo: ignorados=%d err=%v, want 1 ignorado", o.Ignored, err)
	}
	if o, err := ParseClaudeOTel(strings.NewReader("")); err != nil || len(o.Requests) != 0 {
		t.Errorf("entrada vacia: %+v err=%v, want vacio sin error", o, err)
	}
	// JSON invalido: error con el numero de linea
	_, err = ParseClaudeOTel(strings.NewReader(`{"event.name":"api_request"}` + "\n{no json\n"))
	if err == nil || !strings.Contains(err.Error(), "línea 2") {
		t.Errorf("JSON invalido: err = %v, want error con 'línea 2'", err)
	}
}

func TestParseClaudeOTel_ToolResultDuplicatesByToolUseID(t *testing.T) {
	tr := map[string]any{"event.name": "tool_result", "tool_name": "Read", "tool_use_id": "toolu_1"}
	o, err := ParseClaudeOTel(jsonl(t, tr, tr))
	if err != nil || len(o.ToolResults) != 1 || o.Duplicates != 1 {
		t.Errorf("tool_result repetido: resultados=%d duplicados=%d err=%v, want 1 y 1", len(o.ToolResults), o.Duplicates, err)
	}
	// sin tool_use_id no se puede deduplicar: ambos cuentan
	a := map[string]any{"event.name": "tool_result", "tool_name": "Read"}
	o, _ = ParseClaudeOTel(jsonl(t, a, a))
	if len(o.ToolResults) != 2 {
		t.Errorf("sin tool_use_id: %d resultados, want 2", len(o.ToolResults))
	}
}

func roleTotals(a Attribution) map[string]int64 {
	m := map[string]int64{}
	for _, e := range a.Entries {
		m[e.Role] += e.Tokens
	}
	return m
}

func TestAttribute_DelegatedByAgentID(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	a, err := o.Attribute("2.1.310")
	if err != nil {
		t.Fatal(err)
	}
	got := roleTotals(a)
	if got[RolePrincipal] != 40000 || got[RoleWorker] != 20000 || len(got) != 2 {
		t.Errorf("por rol = %v, want principal 40000 y worker 20000", got)
	}
	if a.ByAgent["a3f2c1b4d5e6f7a80"] != 20000 || len(a.ByAgent) != 1 {
		t.Errorf("por agente = %v, want {a3f2c1b4d5e6f7a80: 20000}", a.ByAgent)
	}
	if !a.PerWorker.Available || a.Partial {
		t.Errorf("atribucion por worker %+v partial=%v, want disponible y completa", a.PerWorker, a.Partial)
	}
}

func TestAttribute_AgentNameCustomIsNeverUsedToAttribute(t *testing.T) {
	ev := []map[string]any{
		apiEvent("r1", "repl_main_thread", 0, 100, 0, 0, nil),
		apiEvent("r2", "worker-a", 0, 10, 0, 0, map[string]any{"agent.name": "custom", "agent_id": "agent-A"}),
		apiEvent("r3", "worker-b", 0, 20, 0, 0, map[string]any{"agent.name": "custom", "agent_id": "agent-B"}),
		// agent.name custom pero sin agent_id: no es un worker "custom"
		apiEvent("r4", "compact", 0, 5, 0, 0, map[string]any{"agent.name": "custom"}),
	}
	o, err := ParseClaudeOTel(jsonl(t, ev...))
	if err != nil {
		t.Fatal(err)
	}
	a, err := o.Attribute("2.1.310")
	if err != nil {
		t.Fatal(err)
	}
	if a.ByAgent["agent-A"] != 10 || a.ByAgent["agent-B"] != 20 || a.ByAgent["custom"] != 0 || len(a.ByAgent) != 2 {
		t.Errorf("por agente = %v, want agent-A 10 y agent-B 20 separados, sin 'custom'", a.ByAgent)
	}
	got := roleTotals(a)
	if got[RoleWorker] != 30 || got[RoleAuxiliary] != 5 || got[RolePrincipal] != 100 {
		t.Errorf("por rol = %v, want worker 30, auxiliar 5 (sin agent_id) y principal 100", got)
	}
	// agent.name distinto con el mismo agent_id sigue siendo el mismo worker
	ev2 := []map[string]any{
		apiEvent("s1", "w", 0, 7, 0, 0, map[string]any{"agent.name": "custom", "agent_id": "same"}),
		apiEvent("s2", "w", 0, 8, 0, 0, map[string]any{"agent.name": "Explore", "agent_id": "same"}),
	}
	o, _ = ParseClaudeOTel(jsonl(t, ev2...))
	a, _ = o.Attribute("2.1.310")
	if a.ByAgent["same"] != 15 || len(a.ByAgent) != 1 {
		t.Errorf("mismo agent_id con distinto agent.name: %v, want {same: 15}", a.ByAgent)
	}
}

func TestAttribute_BelowMinimumVersionHasNoPerWorkerAttribution(t *testing.T) {
	o := parseOTel(t, "claude/otel-partial-old-version.jsonl")
	a, err := o.Attribute("2.1.309")
	if err != nil {
		t.Fatal(err)
	}
	got := roleTotals(a)
	if got[RolePrincipal] != 9500 || got[RoleUnattributed] != 12000 || got[RoleWorker] != 0 {
		t.Errorf("por rol = %v, want principal 9500 y sin_atribuir 12000 (sin agent_id no hay worker)", got)
	}
	if a.PerWorker.Available || !strings.Contains(a.PerWorker.Reason, ClaudeMinAgentIDVersion) || !a.Partial {
		t.Errorf("atribucion %+v partial=%v, want no disponible con la version minima en el motivo y parcial", a.PerWorker, a.Partial)
	}
	// aunque el evento traiga agent_id, por debajo de la minima no se atribuye a un worker
	e := apiEvent("r", "general-purpose", 0, 50, 0, 0, map[string]any{"agent_id": "x"})
	o, _ = ParseClaudeOTel(jsonl(t, e))
	a, _ = o.Attribute("2.1.309")
	if roleTotals(a)[RoleWorker] != 0 || a.ByAgent["x"] != 0 {
		t.Errorf("version 2.1.309 con agent_id: %v %v, want sin worker", roleTotals(a), a.ByAgent)
	}
	// sin eventos incompletos: la atribucion es parcial solo por la version
	clean, _ := ParseClaudeOTel(jsonl(t,
		apiEvent("m", "repl_main_thread", 0, 100, 0, 0, nil),
		apiEvent("w", "general-purpose", 0, 50, 0, 0, nil)))
	ca, _ := clean.Attribute("2.1.309")
	if !ca.Partial || len(ca.Reasons) == 0 || roleTotals(ca)[RoleUnattributed] != 50 {
		t.Errorf("version vieja sin incompletos: %+v, want parcial con motivo y 50 sin atribuir", ca)
	}
	// con la version minima y sin subagentes la atribucion es completa
	ok, _ := ParseClaudeOTel(jsonl(t, apiEvent("m", "repl_main_thread", 0, 100, 0, 0, nil)))
	if oa, _ := ok.Attribute("2.1.310"); oa.Partial || len(oa.Reasons) != 0 {
		t.Errorf("version minima solo con principal: %+v, want completa", oa)
	}
	if _, err := o.Attribute("no-version"); err == nil {
		t.Error("version no interpretable: err = nil, want error")
	}
}

func claudeResultOf(t *testing.T, rel string) ClaudeResult {
	t.Helper()
	r, err := ParseClaudeResult(sampleBytes(t, rel))
	if err != nil {
		t.Fatalf("ParseClaudeResult(%s): %v", rel, err)
	}
	return r
}

func TestParseClaudeResult_DelegatedAndErrorSamples(t *testing.T) {
	r := claudeResultOf(t, "claude/result-delegated.json")
	if r.Subtype != "success" || r.IsError || r.SessionID != "00000000-0000-4000-8000-000000000001" ||
		r.DurationMS != 91234 || r.DurationAPIMS != 70211 || r.NumTurns != 7 || r.TotalCostUSD != 0.42 || r.PermissionDenials != 0 {
		t.Errorf("resultado = %+v", r)
	}
	if !r.HasMainUsage || r.MainUsage.Total() != 40000 || r.MainUsage.CacheRead != 28000 || r.MainUsage.CacheCreation != 10500 || r.MainUsage.Uncached != 300 {
		t.Errorf("usage (solo el bucle principal) = %+v, want total 40000", r.MainUsage)
	}
	if tree, ok := r.TreeTotal(); !ok || tree != 60000 {
		t.Errorf("TreeTotal = (%d, %v), want (60000, true) desde modelUsage", tree, ok)
	}

	e := claudeResultOf(t, "claude/result-error-max-turns.json")
	if e.Subtype != "error_max_turns" || !e.IsError || e.PermissionDenials != 1 {
		t.Errorf("resultado de error = %+v, want error_max_turns, is_error y 1 denegacion", e)
	}
	if dump := fmt.Sprintf("%+v", e); strings.Contains(dump, "SECRET") || strings.Contains(dump, "rm -rf") {
		t.Errorf("el resultado conserva el contenido de la herramienta denegada: %s", dump)
	}
}

func TestParseClaudeResult_Invalid(t *testing.T) {
	cases := map[string]string{
		"no es JSON":                     `{no`,
		"no es un resultado":             `{"type":"assistant"}`,
		"tokens negativos":               `{"type":"result","subtype":"success","usage":{"input_tokens":-1,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}`,
		"segundo documento":              `{"type":"result","subtype":"success"} {"type":"result"}`,
		"modelUsage negativo":            `{"type":"result","subtype":"success","modelUsage":{"m":{"inputTokens":-3,"outputTokens":0,"cacheReadInputTokens":0,"cacheCreationInputTokens":0}}}`,
		"permission_denials sin arreglo": `{"type":"result","subtype":"success","permission_denials":"x"}`,
	}
	for name, in := range cases {
		if _, err := ParseClaudeResult([]byte(in)); err == nil {
			t.Errorf("%s: err = nil, want error", name)
		}
	}
	// sin usage ni modelUsage: ni cero ni total disponible
	r, err := ParseClaudeResult([]byte(`{"type":"result","subtype":"success","session_id":"s"}`))
	if err != nil || r.HasMainUsage {
		t.Errorf("sin usage: %+v err=%v, want HasMainUsage=false", r, err)
	}
	if _, ok := r.TreeTotal(); ok {
		t.Error("sin modelUsage TreeTotal debe ser no disponible")
	}
}

func TestReconcileClaude_DelegatedExampleReproduces60000(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	att, err := o.Attribute("2.1.310")
	if err != nil {
		t.Fatal(err)
	}
	res := claudeResultOf(t, "claude/result-delegated.json")
	tot, err := ReconcileClaude(att, &res)
	if err != nil {
		t.Fatal(err)
	}
	if tot.Total != 60000 || tot.ByRole[RolePrincipal] != 40000 || tot.ByRole[RoleWorker] != 20000 || tot.Unattributed != 0 || tot.Partial || !tot.FullCoverage {
		t.Errorf("Totals = %+v, want total 60000 (principal 40000, worker 20000, sin atribuir 0, cobertura completa)", tot)
	}
	// los calculos erroneos del diseno no aparecen: 80000 (doble conteo) ni 48000 (total_tokens del cierre)
	if tot.Total == 80000 || tot.Total == 48000 {
		t.Errorf("total %d es uno de los calculos erroneos del diseno", tot.Total)
	}
	if len(o.SubagentCloses) != 1 {
		t.Fatalf("cierres = %d, want 1", len(o.SubagentCloses))
	}
	if got := o.SubagentCloses[0].TotalTokens; tot.ByRole[RolePrincipal]+got == tot.Total {
		t.Errorf("el total coincide con principal + total_tokens del cierre (%d): se uso el cierre como consumo", got)
	}
}

func TestReconcileClaude_AggregateScopes(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	att, _ := o.Attribute("2.1.310")
	res := claudeResultOf(t, "claude/result-delegated.json")

	// solo usage (sin modelUsage): excluye subagentes, asi que no es un total del arbol
	onlyUsage := res
	onlyUsage.ModelUsage = nil
	tot, err := ReconcileClaude(att, &onlyUsage)
	if err != nil || tot.Total != 60000 || tot.AggregateFigure == nil || *tot.AggregateFigure != 40000 {
		t.Errorf("solo usage: %+v err=%v, want total 60000 del detalle y la cifra de usage (40000) registrada", tot, err)
	}
	// sin resultado: el total es la suma del detalle
	tot, err = ReconcileClaude(att, nil)
	if err != nil || tot.Total != 60000 || !tot.FullCoverage {
		t.Errorf("sin resultado: %+v err=%v, want 60000", tot, err)
	}
	// el detalle no cubre al worker: la diferencia va a sin atribuir y es parcial
	short := Attribution{Entries: att.Entries[:0:0]}
	for _, e := range att.Entries {
		if e.Role == RolePrincipal {
			short.Entries = append(short.Entries, e)
		}
	}
	tot, err = ReconcileClaude(short, &res)
	if err != nil || tot.Total != 60000 || tot.Unattributed != 20000 || !tot.Partial {
		t.Errorf("detalle menor: %+v err=%v, want sin atribuir 20000 y parcial", tot, err)
	}
	// el detalle supera al arbol: ambas cifras y parcial
	small := res
	small.ModelUsage = map[string]usage.Logical{"m": {Input: 1000, Output: 1000}}
	tot, err = ReconcileClaude(att, &small)
	if err != nil || !tot.Partial || tot.AggregateFigure == nil || *tot.AggregateFigure != 2000 || tot.DetailFigure == nil || *tot.DetailFigure != 60000 || tot.FullCoverage {
		t.Errorf("detalle mayor: %+v err=%v, want parcial con ambas cifras y sin cobertura completa", tot, err)
	}
}

func TestParseClaudeStream_DelegatedSample(t *testing.T) {
	s, err := ParseClaudeStream(sampleReader(t, "claude/stream-delegated.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != "2.1.310" || s.PermissionMode != "acceptEdits" || s.SessionID != "00000000-0000-4000-8000-000000000001" || s.Model != "claude-sonnet-5" {
		t.Errorf("init = %+v", s)
	}
	if len(s.Steps) != 4 || s.Duplicates != 4 {
		t.Fatalf("pasos=%d duplicados=%d, want 4 pasos y 4 mensajes repetidos (3 de msg_01A y 1 de msg_01W1)", len(s.Steps), s.Duplicates)
	}
	var main, worker int64
	for _, st := range s.Steps {
		if st.Usage.Output != 0 {
			t.Errorf("paso %s: salida %d, want 0 (la salida por paso es un marcador, se lee del resultado)", st.MessageID, st.Usage.Output)
		}
		switch st.Role {
		case RolePrincipal:
			main += st.Usage.Input
		case RoleWorker:
			worker += st.Usage.Input
			if st.AgentID != "a3f2c1b4d5e6f7a80" || st.ParentToolUseID != "toolu_01AGENT" {
				t.Errorf("paso del worker = %+v", st)
			}
		default:
			t.Errorf("rol inesperado %q en %s", st.Role, st.MessageID)
		}
	}
	if main != 38800 || worker != 18900 {
		t.Errorf("entrada por rol principal=%d worker=%d, want 38800 y 18900 (9200+29600 y 11400+7500)", main, worker)
	}
	if s.Retries != 1 {
		t.Errorf("reintentos = %d, want 1 (system/api_retry)", s.Retries)
	}
	if s.Result == nil || s.Result.Subtype != "success" {
		t.Fatalf("resultado = %+v, want success", s.Result)
	}
	if tree, _ := s.Result.TreeTotal(); tree != 60000 {
		t.Errorf("TreeTotal del resultado del flujo = %d, want 60000", tree)
	}
	// entrada de los pasos + salida del resultado = total del arbol
	if got := main + worker + 2300; got != 60000 {
		t.Errorf("pasos + salida de modelUsage = %d, want 60000", got)
	}
	types := map[string]int{}
	for _, e := range s.Events {
		types[e.Type]++
	}
	if types["system/init"] != 1 || types["system/api_retry"] != 1 || types["assistant"] != 8 || types["result"] != 1 {
		t.Errorf("eventos por tipo = %v", types)
	}
}

func TestParseClaudeStream_EdgeCases(t *testing.T) {
	s, err := ParseClaudeStream(strings.NewReader(""))
	if err != nil || len(s.Steps) != 0 || s.Result != nil {
		t.Errorf("flujo vacio: %+v err=%v", s, err)
	}
	// flujo cortado: sin result queda nil, no un resultado cero
	cut := `{"type":"system","subtype":"init","claude_code_version":"2.1.310","session_id":"s","permissionMode":"default","model":"m"}` + "\n" +
		`{"type":"assistant","message":{"id":"m1","usage":{"input_tokens":1,"output_tokens":1,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"parent_tool_use_id":null}` + "\n"
	s, err = ParseClaudeStream(strings.NewReader(cut))
	if err != nil || s.Result != nil || len(s.Steps) != 1 || s.Version != "2.1.310" {
		t.Errorf("flujo cortado: %+v err=%v, want un paso y sin resultado", s, err)
	}
	// subagente sin agent_id: no se atribuye a un worker concreto
	noID := `{"type":"assistant","message":{"id":"m2","usage":{"input_tokens":5,"output_tokens":1,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"parent_tool_use_id":"toolu_X"}` + "\n"
	s, _ = ParseClaudeStream(strings.NewReader(noID))
	if len(s.Steps) != 1 || s.Steps[0].Role != RoleUnattributed || s.Steps[0].AgentID != "" {
		t.Errorf("subagente sin agent_id: %+v, want rol sin_atribuir", s.Steps)
	}
	if _, err := ParseClaudeStream(strings.NewReader("{no\n")); err == nil {
		t.Error("JSON invalido: err = nil")
	}
}

func TestSignalsFor_Claude(t *testing.T) {
	tests := []struct {
		name    string
		iface   manifest.Interface
		version string
		want    Signals
	}{
		{"terminal desde 2.1.310", manifest.InterfaceTerminal, "2.1.310",
			Signals{TokenUsage: StatusDocumented, PerAgentAttribution: StatusDocumented, ToolCalls: StatusDocumented, FileReads: StatusUnavailable, Permissions: StatusDocumented, Timing: StatusDocumented}},
		{"terminal antes de 2.1.310", manifest.InterfaceTerminal, "2.1.309",
			Signals{TokenUsage: StatusDocumented, PerAgentAttribution: StatusUnavailable, ToolCalls: StatusDocumented, FileReads: StatusUnavailable, Permissions: StatusDocumented, Timing: StatusDocumented}},
		{"IDE", manifest.InterfaceIDE, "2.1.310",
			Signals{TokenUsage: StatusNotVerified, PerAgentAttribution: StatusNotVerified, ToolCalls: StatusNotVerified, FileReads: StatusNotVerified, Permissions: StatusNotVerified, Timing: StatusNotVerified}},
	}
	for _, tc := range tests {
		got, err := SignalsFor(manifest.ClientClaude, tc.iface, tc.version)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		got.Notes = nil
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %+v, want %+v", tc.name, got, tc.want)
		}
	}
	if _, err := SignalsFor(manifest.ClientClaude, manifest.InterfaceTerminal, "x"); err == nil {
		t.Error("version de Claude no interpretable: err = nil")
	}
	if _, err := SignalsFor("gemini", manifest.InterfaceTerminal, "1"); err == nil {
		t.Error("cliente fuera de la matriz: err = nil")
	}
	if _, err := SignalsFor(manifest.ClientClaude, "desktop", "2.1.310"); err == nil {
		t.Error("interfaz fuera de la matriz: err = nil")
	}
	s, _ := SignalsFor(manifest.ClientClaude, manifest.InterfaceTerminal, "2.1.310")
	if len(s.Notes) == 0 {
		t.Error("las notas de limites de atribucion no pueden estar vacias")
	}
}

func TestClaudeTrace_ToolCallsByRoleAndUnavailableReads(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	log, err := NewClaudeTraceLog(manifest.InterfaceTerminal, "2.1.310")
	if err != nil {
		t.Fatal(err)
	}
	if err := FillClaudeTrace(log, o, "2.1.310"); err != nil {
		t.Fatal(err)
	}
	s := log.Summary()
	if !s.ToolCalls.Available || s.ToolCalls.Total != 7 {
		t.Fatalf("llamadas = %+v, want disponibles y 7", s.ToolCalls)
	}
	if s.ToolCalls.ByRole[RolePrincipal] != 4 || s.ToolCalls.ByRole[RoleWorker] != 3 {
		t.Errorf("por rol = %v, want principal 4 y worker 3", s.ToolCalls.ByRole)
	}
	if s.ToolCalls.ByType["Read"] != 4 || s.ToolCalls.ByType["Grep"] != 1 || s.ToolCalls.ByType["Bash"] != 1 || s.ToolCalls.ByType["Edit"] != 1 {
		t.Errorf("por tipo = %v", s.ToolCalls.ByType)
	}
	if s.ToolCalls.ByTypeRole["Read"][RoleWorker] != 2 || s.ToolCalls.ByTypeRole["Read"][RolePrincipal] != 2 {
		t.Errorf("tipo x rol = %v", s.ToolCalls.ByTypeRole)
	}
	if s.Reads.Available || !strings.Contains(s.Reads.Reason, "OTEL_LOG_TOOL_DETAILS") {
		t.Errorf("lecturas = %+v, want no disponibles con el motivo de OTEL_LOG_TOOL_DETAILS", s.Reads)
	}

	// antes de 2.1.310 no se separa principal de worker: no disponible, no cero
	old, err := NewClaudeTraceLog(manifest.InterfaceTerminal, "2.1.309")
	if err != nil {
		t.Fatal(err)
	}
	if s := old.Summary(); s.ToolCalls.Available || !strings.Contains(s.ToolCalls.Reason, ClaudeMinAgentIDVersion) {
		t.Errorf("version vieja: %+v, want llamadas no disponibles con la version minima en el motivo", s.ToolCalls)
	}
	// IDE: ninguna senal comprobada
	ide, _ := NewClaudeTraceLog(manifest.InterfaceIDE, "2.1.310")
	if s := ide.Summary(); s.ToolCalls.Available || s.Reads.Available {
		t.Errorf("IDE: %+v, want todo no disponible", s)
	}
	if err := FillClaudeTrace(ide, o, "2.1.310"); err == nil {
		t.Error("registrar en un log sin senal debe ser error")
	}
	if _, err := NewClaudeTraceLog(manifest.InterfaceTerminal, "?"); err == nil {
		t.Error("version no interpretable: err = nil")
	}
}

// Los atributos estandar (correo, UUID de cuenta, id de usuario) viajan en cada
// evento de la muestra; el adaptador no los copia a lo capturado.
func TestParseClaudeOTel_DoesNotCopyIdentityAttributes(t *testing.T) {
	o := parseOTel(t, "claude/otel-delegated.jsonl")
	dump := fmt.Sprintf("%+v", o)
	for _, secret := range []string{"persona@example.invalid", "0000000000aa", "anon-0000"} {
		if strings.Contains(dump, secret) {
			t.Errorf("lo capturado contiene %q", secret)
		}
	}
}
