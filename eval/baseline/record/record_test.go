package record

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/timing"
	"github.com/Group-Active-IA/active-stack/eval/baseline/trace"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

func has(cs []Category, want Category) bool {
	for _, c := range cs {
		if c == want {
			return true
		}
	}
	return false
}

// Un campo sensible por vez (lista mínima de 05): ninguno llega al resultado.
func TestRedactJSON_EachSensitiveFieldIsRemoved(t *testing.T) {
	for _, c := range []struct {
		name     string
		in       string
		sentinel string
		cat      Category
	}{
		{"correo por clave", `{"type":"error","user_email":"ana.perez@example.com"}`, "ana.perez", Email},
		{"correo en el texto", `{"message":"fallo de ana.perez@example.com al iniciar"}`, "ana.perez", Email},
		{"uuid de cuenta por clave", `{"account_uuid":"x-not-a-uuid-SENTINEL"}`, "SENTINEL", AccountUUID},
		{"uuid de cuenta anidado", `{"user":{"account_uuid":"SENTINEL"}}`, "SENTINEL", AccountUUID},
		{"uuid en el texto", `{"message":"cuenta 123e4567-e89b-12d3-a456-426614174000 sin cuota"}`, "123e4567", AccountUUID},
		{"id de usuario por clave", `{"user_id":"SENTINEL-user"}`, "SENTINEL", UserID},
		{"id de usuario con punto", `{"user.id":"SENTINEL-user"}`, "SENTINEL", UserID},
		{"id anidado bajo user", `{"user":{"id":"SENTINEL-user"}}`, "SENTINEL", UserID},
		{"ruta de Windows", `{"message":"no se pudo abrir C:\\Users\\ana\\proj\\a.go ahora"}`, "ana", HostPath},
		{"ruta POSIX home", `{"message":"leyendo /home/ana/work/a.go"}`, "/home/ana", HostPath},
		{"ruta POSIX Users", `{"message":"leyendo /Users/ana/work/a.go"}`, "/Users/ana", HostPath},
		{"ruta UNC", `{"message":"abrir \\\\srv\\share\\SENTINEL-unc"}`, "SENTINEL", HostPath},
		{"ruta con tilde", `{"message":"leer ~/SENTINEL-home/x"}`, "SENTINEL", HostPath},
		{"correo por clave sin arroba", `{"email":"SENTINEL-mail"}`, "SENTINEL", Email},
		{"cuerpo como texto", `{"request_body":"SENTINEL-body"}`, "SENTINEL", RequestBody},
		{"ruta tras comillas", `{"cwd":"/home/ana"}`, "/home/ana", HostPath},
		{"detalle de herramienta tool_input", `{"tool_input":{"command":"SENTINEL-cmd"}}`, "SENTINEL", ToolDetail},
		{"detalle de herramienta arguments", `{"arguments":"SENTINEL-args"}`, "SENTINEL", ToolDetail},
		{"salida de herramienta", `{"tool_output":"SENTINEL-out"}`, "SENTINEL", ToolDetail},
		{"cuerpo de solicitud", `{"request_body":{"messages":[{"content":"SENTINEL-body"}]}}`, "SENTINEL", RequestBody},
		{"mensajes", `{"messages":[{"role":"user","content":"SENTINEL-body"}]}`, "SENTINEL", RequestBody},
		{"prompt", `{"prompt":"SENTINEL-body"}`, "SENTINEL", RequestBody},
	} {
		out, cats := RedactJSON(json.RawMessage(c.in))
		if strings.Contains(string(out), c.sentinel) {
			t.Errorf("%s: la salida conserva %q: %s", c.name, c.sentinel, out)
		}
		if !has(cats, c.cat) {
			t.Errorf("%s: categorías = %v, want incluir %q", c.name, cats, c.cat)
		}
		if !json.Valid(out) {
			t.Errorf("%s: la salida no es JSON válido: %s", c.name, out)
		}
	}
}

// Lección del auditor: los miembros benignos no se tapan.
func TestRedactJSON_BenignMembersAreKept(t *testing.T) {
	in := `{"type":"turn.failed","code":500,"message":"rate limit, ver https://example.com/home/docs y internal/foo/bar.go","model":"claude-x","attempt":2,"duration_ms":1200,"id":"evt_1","version":"2.1.310","ok":false,"items":[1,2,{"name":"a"}],"note":"sesión https://x.io/Users/guide"}`
	out, cats := RedactJSON(json.RawMessage(in))
	var a, b any
	if err := json.Unmarshal([]byte(in), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("se alteró contenido benigno:\n in  %s\n out %s", in, out)
	}
	if len(cats) != 0 {
		t.Errorf("categorías = %v, want ninguna", cats)
	}
}

func TestRedactText_BenignAndSensitive(t *testing.T) {
	got, cats := RedactText("fallo en C:\\Users\\ana\\x.go y /home/ana/y.go, mail ana@example.com")
	for _, s := range []string{"ana", "C:\\", "/home"} {
		if strings.Contains(got, s) {
			t.Errorf("texto conserva %q: %s", s, got)
		}
	}
	if !has(cats, HostPath) || !has(cats, Email) {
		t.Errorf("categorías = %v", cats)
	}
	ben := "sin datos sensibles: 3 archivos en internal/x.go, versión 2.1.310, ratio a:b"
	if g, c := RedactText(ben); g != ben || len(c) != 0 {
		t.Errorf("texto benigno alterado: %q %v", g, c)
	}
}

func TestRedactJSON_MalformedEventIsDroppedNotPartiallyScrubbed(t *testing.T) {
	// JSON inválido: no se puede saber qué claves son sensibles, así que no se conserva nada.
	out, cats := RedactJSON(json.RawMessage(`{"user_id":"SENTINEL-u","tool_input":"SENTINEL-t" ana@example.com`))
	if !json.Valid(out) || strings.Contains(string(out), "SENTINEL") || strings.Contains(string(out), "ana@") || !has(cats, Unparsed) {
		t.Errorf("= %s %v", out, cats)
	}
	out, cats = RedactJSON(json.RawMessage(`{"a":1} {"user_id":"SENTINEL-u"}`))
	if strings.Contains(string(out), "SENTINEL") || !has(cats, Unparsed) {
		t.Errorf("segundo documento filtrado: %s", out)
	}
}

func TestRecords_MalformedErrorEventKeepsOnlyItsType(t *testing.T) {
	a := syntheticAttempt()
	a.ErrorEvents = []run.Event{{Type: "turn.failed", Raw: json.RawMessage(`{"user_id":"SENTINEL-u" broken`)}, {Type: "error"}}
	ev, _ := NewEvidence("", nil, "n/a")
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Attempts: []run.Attempt{a}}
	_, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a, Evidence: ev}})
	if err != nil || len(ars[0].ErrorEvents) != 2 {
		t.Fatalf("= %v %v", ars, err)
	}
	b0, b1 := string(ars[0].ErrorEvents[0]), string(ars[0].ErrorEvents[1])
	if strings.Contains(b0, "SENTINEL") || !strings.Contains(b0, `"type":"turn.failed"`) || b1 != `{"type":"error"}` {
		t.Errorf("eventos = %s %s", b0, b1)
	}
}

func TestSHA256_NamesTheAlgorithm(t *testing.T) {
	fp := SHA256([]byte("abc"))
	if fp.Algorithm != "sha256" || fp.Value != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("= %+v", fp)
	}
	if SHA256([]byte("abd")).Value == fp.Value {
		t.Error("contenidos distintos, misma huella")
	}
}

func TestNewEvidence_KeepsReferenceAndFingerprintOrRecordsAbsence(t *testing.T) {
	ev, err := NewEvidence("local://traces/run-1.jsonl", []byte("abc"), "")
	if err != nil || ev.Ref != "local://traces/run-1.jsonl" || ev.Fingerprint == nil || ev.Fingerprint.Algorithm != "sha256" || ev.Absent != "" {
		t.Errorf("= %+v, %v", ev, err)
	}
	// una ruta del host no llega: se conserva solo el nombre del archivo
	ev, _ = NewEvidence(`C:\Users\ana\traces\run-1.jsonl`, []byte("abc"), "")
	if strings.Contains(ev.Ref, "ana") || strings.Contains(ev.Ref, "C:") || !strings.Contains(ev.Ref, "run-1.jsonl") || ev.Fingerprint == nil {
		t.Errorf("ref con ruta del host = %+v", ev)
	}
	ev, _ = NewEvidence("/home/ana/t/run-2.jsonl", []byte("x"), "")
	if strings.Contains(ev.Ref, "/home") || !strings.Contains(ev.Ref, "run-2.jsonl") {
		t.Errorf("ref POSIX = %+v", ev)
	}
	ev, _ = NewEvidence("/data/run/trace-3.jsonl", []byte("x"), "")
	if strings.Contains(ev.Ref, "/data") || !strings.Contains(ev.Ref, "trace-3.jsonl") {
		t.Errorf("ref absoluta cualquiera = %+v", ev)
	}
	// ausencia registrada con motivo
	ev, err = NewEvidence("", nil, "la interfaz no exporta trazas")
	if err != nil || ev.Fingerprint != nil || ev.Absent != "la interfaz no exporta trazas" || ev.Ref != "" {
		t.Errorf("ausente = %+v, %v", ev, err)
	}
	if _, err := NewEvidence("", nil, " "); err == nil {
		t.Error("sin evidencia ni motivo: error")
	}
	if _, err := NewEvidence("ref", nil, ""); err == nil {
		t.Error("referencia sin contenido ni motivo de ausencia: error")
	}
}

// --- Registros por intento y corrida ---

const (
	secretEmail = "ana.perez@example.com"
	secretAcct  = "123e4567-e89b-12d3-a456-426614174000"
	secretUser  = "usr-SECRET-77"
	secretPath  = "/home/ana/work"
	secretWin   = `C:\Users\ana\proj`
	secretTool  = "rm -rf SECRET-TOOL"
	secretBody  = "SECRET PROMPT BODY"
	sessionID   = "9a1b2c3d-0000-4000-8000-00000000abcd"
)

func syntheticAttempt() run.Attempt {
	ev := `{"type":"error","message":"rate limit para ` + secretEmail + ` en ` + strings.ReplaceAll(secretWin, `\`, `\\`) + `\\a.go","code":429,` +
		`"user":{"email":"` + secretEmail + `","id":"` + secretUser + `"},"user_id":"` + secretUser + `",` +
		`"account_uuid":"` + secretAcct + `","tool_input":{"command":"` + secretTool + `"},` +
		`"request_body":{"messages":[{"role":"user","content":"` + secretBody + `"}]},"cwd":"` + secretPath + `"}`
	code := 1
	return run.Attempt{
		N: 1, Cause: "inicial", Status: manifest.RunClientFailure,
		Reason:      "falló al escribir " + secretPath + "/a.go para " + secretEmail,
		ExitCode:    &code,
		ErrorEvents: []run.Event{{Type: "error", Raw: json.RawMessage(ev)}},
		Events:      3,
		Timing: timing.Measurement{Elapsed: 1500 * time.Millisecond,
			WaitTotal: timing.Metric{Available: true, Value: 200 * time.Millisecond},
			Adjusted:  timing.Metric{Available: true, Value: 1300 * time.Millisecond}},
		Trace: trace.Summary{
			Reads:     trace.Counts{Available: true, Total: 2, ByType: map[string]int{"Read": 2}, ByRole: map[string]int{"principal": 2}, Unique: 2},
			ToolCalls: trace.Counts{Reason: "la interfaz no expone esta señal"},
		},
	}
}

func buildSynthetic(t *testing.T) (RunRecord, []AttemptRecord) {
	t.Helper()
	a := syntheticAttempt()
	agg := int64(60000)
	tot, err := usage.Reconcile(usage.Input{Aggregate: &agg, Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{
		{Role: "principal", Tokens: 40000}, {Role: "worker", Tokens: 20000}}})
	if err != nil {
		t.Fatal(err)
	}
	ev, _ := NewEvidence(secretWin+`\trace.jsonl`, []byte("traza original"), "")
	cache := usage.CacheRecord{Expected: usage.CacheCold, Observed: usage.CacheRead, Diverges: true}
	r := run.Run{CohortID: "c1", TaskID: "t1", Repetition: 2, Status: a.Status, Attempts: []run.Attempt{a}}
	rr, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{
		Attempt: a, SessionID: sessionID, AgentID: "agent-1", Totals: &tot, Cache: &cache, Evidence: ev}})
	if err != nil {
		t.Fatal(err)
	}
	return rr, ars
}

func TestRecords_SyntheticTraceNeverLeaksSensitiveData(t *testing.T) {
	rr, ars := buildSynthetic(t)
	var buf bytes.Buffer
	if err := WriteJSONL(&buf, ars[0], rr); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, s := range []string{secretEmail, "ana.perez", secretAcct, secretUser, secretPath, "/home", secretWin, `C:\\Users`, "ana", secretTool, secretBody} {
		if strings.Contains(out, s) {
			t.Errorf("el resultado compartido contiene %q:\n%s", s, out)
		}
	}
	// lo benigno sigue ahí
	for _, s := range []string{`"type":"error"`, `"code":429`, "rate limit", sessionID, `"exit_code":1`, "run-", `"status":"fallo_del_cliente"`, "trace.jsonl"} {
		if s == "run-" {
			continue
		}
		if !strings.Contains(out, s) {
			t.Errorf("falta lo benigno %q:\n%s", s, out)
		}
	}
}

func TestRecords_CoverageMetadataAndRedactionCounts(t *testing.T) {
	rr, ars := buildSynthetic(t)
	a := ars[0]
	if a.Kind != "attempt" || rr.Kind != "run" || a.ProtocolVersion != "1" || rr.ProtocolVersion != "1" {
		t.Errorf("kind/versión: %+v %+v", a, rr)
	}
	if a.Coverage.Reads != "disponible" || a.Coverage.ToolCalls != "no_disponible" || a.Coverage.Usage != "completa" || a.Coverage.Timing != "disponible" {
		t.Errorf("cobertura = %+v", a.Coverage)
	}
	for _, c := range []Category{Email, AccountUUID, UserID, HostPath, ToolDetail, RequestBody} {
		if a.Coverage.Redactions[string(c)] == 0 {
			t.Errorf("redactions[%s] = 0, want > 0: %+v", c, a.Coverage.Redactions)
		}
	}
	if a.Consumption.Total != 60000 || a.Consumption.ByRole["worker"] != 20000 || !a.Consumption.FullCoverage {
		t.Errorf("consumo = %+v", a.Consumption)
	}
	if a.Cache == nil || a.Cache.Observed != usage.CacheRead || !a.Cache.Diverges {
		t.Errorf("caché = %+v", a.Cache)
	}
	if a.ToolCalls.Available || a.ToolCalls.Reason == "" || !a.Reads.Available || a.Reads.Total != 2 {
		t.Errorf("lecturas/herramientas = %+v %+v", a.Reads, a.ToolCalls)
	}
	if a.Timing.ElapsedMs != 1500 || a.Timing.WaitsMs == nil || *a.Timing.WaitsMs != 200 || a.Timing.AdjustedMs == nil || *a.Timing.AdjustedMs != 1300 {
		t.Errorf("tiempo = %+v", a.Timing)
	}
	if a.Evidence.Fingerprint == nil || a.Evidence.Fingerprint.Algorithm != "sha256" || a.Evidence.Ref == "" {
		t.Errorf("evidencia = %+v", a.Evidence)
	}
	if rr.Attempts != 1 || rr.RunID != a.RunID || rr.CohortID != "c1" || rr.TaskID != "t1" || rr.Repetition != 2 || rr.Status != "fallo_del_cliente" {
		t.Errorf("corrida = %+v", rr)
	}
	if rr.Coverage.Usage != "completa" {
		t.Errorf("cobertura de corrida = %+v", rr.Coverage)
	}
}

func TestRecords_MissingSignalsAreUnavailableNotZero(t *testing.T) {
	a := syntheticAttempt()
	a.Timing = timing.Measurement{Elapsed: time.Second,
		WaitTotal: timing.Metric{Reason: "espera abierta"}, Adjusted: timing.Metric{Reason: "espera abierta"}}
	ev, _ := NewEvidence("", nil, "sin traza exportada")
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Status: a.Status, Attempts: []run.Attempt{a}}
	rr, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a, Evidence: ev}})
	if err != nil {
		t.Fatal(err)
	}
	x := ars[0]
	if x.Consumption.Available || x.Consumption.Reason == "" || x.Coverage.Usage != "no_disponible" || rr.Coverage.Usage != "no_disponible" {
		t.Errorf("consumo ausente = %+v %+v", x.Consumption, x.Coverage)
	}
	if x.Timing.WaitsMs != nil || x.Timing.AdjustedMs != nil || x.Timing.AdjustedReason != "espera abierta" || x.Coverage.Timing != "no_disponible" {
		t.Errorf("tiempo ausente = %+v", x.Timing)
	}
	if x.Evidence.Absent != "sin traza exportada" || x.Evidence.Fingerprint != nil {
		t.Errorf("evidencia ausente = %+v", x.Evidence)
	}
	var buf bytes.Buffer
	_ = WriteJSONL(&buf, x)
	if !strings.Contains(buf.String(), `"waits_ms":null`) || !strings.Contains(buf.String(), `"adjusted_ms":null`) {
		t.Errorf("null esperado para no disponible: %s", buf.String())
	}
}

func TestRecords_PartialUsageIsMarkedPartial(t *testing.T) {
	a := syntheticAttempt()
	agg := int64(100)
	tot, _ := usage.Reconcile(usage.Input{Aggregate: &agg, Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{{Role: "p", Tokens: 60}}})
	ev, _ := NewEvidence("", nil, "n/a")
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Attempts: []run.Attempt{a}}
	rr, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a, Totals: &tot, Evidence: ev}})
	if err != nil || ars[0].Coverage.Usage != "parcial" || rr.Coverage.Usage != "parcial" || ars[0].Consumption.Unattributed != 40 {
		t.Errorf("= %+v %+v %v", ars[0].Coverage, ars[0].Consumption, err)
	}
}

func TestRecords_RunCoverageIsTheWeakestAttempt(t *testing.T) {
	a1, a2 := syntheticAttempt(), syntheticAttempt()
	a2.N = 2
	agg := int64(10)
	full, _ := usage.Reconcile(usage.Input{Aggregate: &agg, Scope: usage.ScopeIncludesChildren, Entries: []usage.Entry{{Role: "p", Tokens: 10}}})
	ev, _ := NewEvidence("", nil, "n/a")
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Attempts: []run.Attempt{a1, a2}}
	rr, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a1, Totals: &full, Evidence: ev}, {Attempt: a2, Evidence: ev}})
	if err != nil || len(ars) != 2 || ars[0].RunID != ars[1].RunID || ars[1].Attempt != 2 || rr.Attempts != 2 || rr.Coverage.Usage != "no_disponible" {
		t.Errorf("= %+v %+v %v", rr, ars, err)
	}
}

func TestRecords_InputsMustMatchAttempts(t *testing.T) {
	a := syntheticAttempt()
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Attempts: []run.Attempt{a}}
	ev, _ := NewEvidence("", nil, "n/a")
	if _, _, err := (Builder{ProtocolVersion: "1"}).Records(r, nil); err == nil {
		t.Error("entradas de menos: error")
	}
	b := a
	b.N = 9
	if _, _, err := (Builder{ProtocolVersion: "1"}).Records(r, []AttemptInput{{Attempt: b, Evidence: ev}}); err == nil {
		t.Error("intento distinto: error")
	}
	if _, _, err := (Builder{}).Records(r, []AttemptInput{{Attempt: a, Evidence: ev}}); err == nil {
		t.Error("sin versión de protocolo: error")
	}
}

func TestRecords_InterruptedAttemptIsKept(t *testing.T) {
	a := run.Attempt{N: 1, Cause: "inicial", Status: run.StatusInterrupted, Reason: "cancelación del llamador: context canceled"}
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Status: a.Status, Attempts: []run.Attempt{a}}
	ev, _ := NewEvidence("", nil, "n/a")
	rr, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a, Evidence: ev}})
	if err != nil || ars[0].Status != "interrumpida_por_llamador" || rr.Status != "interrumpida_por_llamador" || !strings.Contains(ars[0].Reason, "cancelación") {
		t.Errorf("= %+v %+v %v", rr, ars, err)
	}
}

func TestWriteJSONL_OneValidJSONObjectPerLine(t *testing.T) {
	rr, ars := buildSynthetic(t)
	var buf bytes.Buffer
	if err := WriteJSONL(&buf, ars[0], rr); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("líneas = %d, want 2: %q", len(lines), buf.String())
	}
	for i, l := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Errorf("línea %d no es JSON: %v", i, err)
		}
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("falta salto de línea final")
	}
	var empty bytes.Buffer
	if err := WriteJSONL(&empty); err != nil || empty.Len() != 0 {
		t.Errorf("sin registros: %q %v", empty.String(), err)
	}
}

func TestRecords_CauseIsRedactedToo(t *testing.T) {
	a := syntheticAttempt()
	a.Cause = "reintento por " + secretEmail
	ev, _ := NewEvidence("", nil, "n/a")
	r := run.Run{CohortID: "c", TaskID: "t", Repetition: 1, Attempts: []run.Attempt{a}}
	_, ars, err := Builder{ProtocolVersion: "1"}.Records(r, []AttemptInput{{Attempt: a, Evidence: ev}})
	if err != nil || strings.Contains(ars[0].Cause, "ana.perez") || !strings.Contains(ars[0].Cause, "reintento por") {
		t.Errorf("= %q %v", ars[0].Cause, err)
	}
}
