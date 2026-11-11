package run

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

const (
	goodAnswer = `{"entry":{"file":"a.txt","symbol":"S"},"files":["a.txt"]}`
	badAnswer  = `{"entry":{"file":"b.txt","symbol":"S"},"files":["a.txt"]}`
)

// miniFixture crea un fixture sintetico de la familia exploracion (checkout de
// un archivo, clave, referencia y defectuosa). Con reviewer agrega un criterio
// que requiere revisor.
func miniFixture(t *testing.T, reviewer bool) *fixture.Fixture {
	t.Helper()
	dir := t.TempDir()
	w := func(name, body string) {
		p := filepath.Join(dir, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	prompt := "Explora el proyecto.\n"
	w("prompt.md", prompt)
	w("project/a.txt", "contenido\n")
	w("key.json", goodAnswer)
	w("reference.json", goodAnswer)
	w("defective.json", badAnswer)
	tree, err := fixture.TreeHash(filepath.Join(dir, "project"))
	if err != nil {
		t.Fatal(err)
	}
	ps := sha256.Sum256([]byte(prompt))
	crit := `{"id":"E1","description":"entrada","mandatory":true,"check":"respuesta_entrada"},
	 {"id":"E2","description":"archivos","mandatory":true,"check":"respuesta_archivos"},
	 {"id":"E4","description":"intacto","mandatory":true,"check":"checkout_intacto"}`
	if reviewer {
		crit += `,{"id":"E3","description":"revisor","mandatory":true,"check":"requiere_revisor"}`
	}
	w("task.json", `{"id":"mini","family":"exploracion","rubric_version":"1",
	"source":{"kind":"embedded","dir":"project","sha256":"`+tree+`"},
	"prompt":{"file":"prompt.md","sha256":"`+hex.EncodeToString(ps[:])+`"},
	"allowed_tools":["leer"],"limits":{"timeout":"(TBD)"},"key":"key.json",
	"reference":"reference.json","defective":"defective.json","criteria":[`+crit+`]}`)
	f, err := fixture.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func intp(n int) *int { return &n }

func raw(s string) json.RawMessage { return json.RawMessage(s) }

type script []func(req Request) (Result, error)

// runner devuelve el runner falso que consume un paso del guion por intento y
// guarda las requests.
func (s script) runner(reqs *[]Request) Runner {
	i := 0
	return RunnerFunc(func(ctx context.Context, req Request) (Result, error) {
		*reqs = append(*reqs, req)
		step := s[i]
		i++
		return step(req)
	})
}

func ok(ans string) func(Request) (Result, error) {
	return func(Request) (Result, error) {
		return Result{Termination: Completed, ExitCode: intp(0), Answer: []byte(ans)}, nil
	}
}

type fixedClock struct{ t time.Duration }

func (c *fixedClock) Now() time.Duration { c.t += time.Second; return c.t }

func newExec(r Runner, p Policy) *Executor {
	return NewExecutor(r, Config{Clock: &fixedClock{}, WaitsObservable: true, Reads: true, Tools: true, Policy: p})
}

func TestExecute_SixCasesHaveDistinctStatesCausesExitCodesAndErrorEvents(t *testing.T) {
	f := miniFixture(t, false)
	turnFailed := raw(`{"type":"turn.failed","error":{"message":"boom"}}`)
	errEv := raw(`{"type":"error","message":"fallo de red"}`)
	failedStep := func(Request) (Result, error) {
		return Result{Termination: ClientFailed, ExitCode: intp(137), Detail: "el cliente termino",
			Events: []Event{{"turn.started", raw(`{"type":"turn.started"}`)}, {"turn.failed", turnFailed}, {"error", errEv}}}, nil
	}
	cases := []struct {
		name       string
		step       func(Request) (Result, error)
		want       manifest.RunStatus
		exit       *int
		reasonHas  string
		errorTypes []string
		verdict    bool
	}{
		{"aceptacion", ok(goodAnswer), manifest.RunAccepted, intp(0), "", nil, true},
		{"rechazo por calidad", ok(badAnswer), manifest.RunRejectedQuality, intp(0), "E1", nil, true},
		{"bloqueo por permisos", func(Request) (Result, error) {
			return Result{Termination: PermissionBlocked, ExitCode: intp(1), Detail: "aprobacion requerida"}, nil
		}, manifest.RunBlockedPermissions, intp(1), "aprobacion requerida", nil, false},
		{"timeout", func(Request) (Result, error) {
			return Result{Termination: TimedOut}, nil
		}, manifest.RunTimeout, nil, "tiempo", nil, false},
		{"fallo del cliente", failedStep, manifest.RunClientFailure, intp(137), "el cliente termino", []string{"turn.failed", "error"}, false},
	}
	seen := map[manifest.RunStatus]string{}
	for _, tc := range cases {
		var reqs []Request
		r, err := newExec(script{tc.step}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c1", 1)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if r.Status != tc.want || len(r.Attempts) != 1 {
			t.Fatalf("%s: Status = %q con %d intentos, want %q con 1", tc.name, r.Status, len(r.Attempts), tc.want)
		}
		if prev, dup := seen[r.Status]; dup {
			t.Errorf("%s y %s comparten estado %q", tc.name, prev, r.Status)
		}
		seen[r.Status] = tc.name
		a := r.Attempts[0]
		if a.Status != tc.want || !strings.Contains(a.Reason, tc.reasonHas) || (tc.reasonHas != "" && a.Reason == "") {
			t.Errorf("%s: intento = %q / causa %q, want %q con %q", tc.name, a.Status, a.Reason, tc.want, tc.reasonHas)
		}
		if !reflect.DeepEqual(a.ExitCode, tc.exit) {
			t.Errorf("%s: ExitCode = %v, want %v", tc.name, a.ExitCode, tc.exit)
		}
		var gotTypes []string
		for _, e := range a.ErrorEvents {
			gotTypes = append(gotTypes, e.Type)
		}
		if !reflect.DeepEqual(gotTypes, tc.errorTypes) {
			t.Errorf("%s: ErrorEvents = %v, want %v", tc.name, gotTypes, tc.errorTypes)
		}
		if (a.Verdict != nil) != tc.verdict {
			t.Errorf("%s: Verdict presente = %v, want %v", tc.name, a.Verdict != nil, tc.verdict)
		}
		if r.CohortID != "c1" || r.TaskID != "mini" || r.Repetition != 1 || a.N != 1 || a.Cause != "inicial" {
			t.Errorf("%s: metadatos = %+v / %+v", tc.name, r, a)
		}
	}
	// Eventos de error: datos brutos sin cambios.
	var reqs []Request
	r, _ := newExec(script{failedStep}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c1", 1)
	if len(r.Attempts) != 1 {
		t.Fatalf("intentos = %d", len(r.Attempts))
	}
	if got := r.Attempts[0].ErrorEvents; len(got) != 2 || string(got[0].Raw) != string(turnFailed) || string(got[1].Raw) != string(errEv) {
		t.Errorf("ErrorEvents = %+v, want los dos JSON brutos intactos", got)
	}
	if r.Attempts[0].Events != 3 {
		t.Errorf("Events = %d, want 3 (todos los eventos contados)", r.Attempts[0].Events)
	}
}

func TestExecute_ErrorEventDoesNotChangeStatusOfCompletedRun(t *testing.T) {
	f := miniFixture(t, false)
	step := func(Request) (Result, error) {
		return Result{Termination: Completed, ExitCode: intp(0), Answer: []byte(goodAnswer),
			Events: []Event{{"error", raw(`{"type":"error"}`)}}}, nil
	}
	var reqs []Request
	r, err := newExec(script{step}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c", 1)
	if err != nil || r.Status != manifest.RunAccepted || len(r.Attempts) != 1 || len(r.Attempts[0].ErrorEvents) != 1 {
		t.Errorf("Run = %+v err %v, want aceptada conservando el evento de error", r, err)
	}
}

func TestExecute_RetryResetsCheckoutAndKeepsEveryAttempt(t *testing.T) {
	f := miniFixture(t, false)
	first := func(req Request) (Result, error) {
		os.WriteFile(filepath.Join(req.Dir, "a.txt"), []byte("roto"), 0o644)
		os.WriteFile(filepath.Join(req.Dir, "extra.txt"), []byte("x"), 0o644)
		return Result{Termination: ClientFailed, ExitCode: intp(2), Detail: "se cayo"}, nil
	}
	second := func(req Request) (Result, error) {
		if b, _ := os.ReadFile(filepath.Join(req.Dir, "a.txt")); string(b) != "contenido\n" {
			t.Errorf("el reintento no empezo desde el fixture: a.txt = %q", b)
		}
		if _, err := os.Stat(filepath.Join(req.Dir, "extra.txt")); err == nil {
			t.Error("el reintento conserva extra.txt del intento anterior")
		}
		return ok(goodAnswer)(req)
	}
	var reqs []Request
	pol := Policy{MaxAttempts: 2, RetryOn: []manifest.RunStatus{manifest.RunClientFailure}}
	r, err := newExec(script{first, second}.runner(&reqs), pol).Execute(context.Background(), f, "c", 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != manifest.RunAccepted || len(r.Attempts) != 2 {
		t.Fatalf("Run = %q con %d intentos, want aceptada con 2", r.Status, len(r.Attempts))
	}
	a1, a2 := r.Attempts[0], r.Attempts[1]
	if a1.Status != manifest.RunClientFailure || *a1.ExitCode != 2 || a1.Cause != "inicial" {
		t.Errorf("intento 1 = %+v", a1)
	}
	if a2.N != 2 || a2.Status != manifest.RunAccepted || a2.Cause != "reintento tras fallo_del_cliente" {
		t.Errorf("intento 2 = %+v", a2)
	}
	if len(reqs) != 2 || reqs[0].Attempt != 1 || reqs[1].Attempt != 2 || reqs[1].Dir != reqs[0].Dir {
		t.Errorf("requests = %+v (el reintento reutiliza el mismo checkout reiniciado)", reqs)
	}
}

func TestExecute_RetryPolicyBoundaries(t *testing.T) {
	f := miniFixture(t, false)
	timeout := func(Request) (Result, error) { return Result{Termination: TimedOut}, nil }
	blocked := func(Request) (Result, error) { return Result{Termination: PermissionBlocked}, nil }
	retryTimeout := []manifest.RunStatus{manifest.RunTimeout}
	for _, tc := range []struct {
		name     string
		pol      Policy
		steps    script
		want     manifest.RunStatus
		attempts int
	}{
		{"sin politica no reintenta", Policy{}, script{timeout}, manifest.RunTimeout, 1},
		{"MaxAttempts 1 no reintenta", Policy{MaxAttempts: 1, RetryOn: retryTimeout}, script{timeout}, manifest.RunTimeout, 1},
		{"agota los reintentos", Policy{MaxAttempts: 2, RetryOn: retryTimeout}, script{timeout, timeout}, manifest.RunTimeout, 2},
		{"no reintenta lo que no esta en RetryOn", Policy{MaxAttempts: 3, RetryOn: retryTimeout}, script{blocked}, manifest.RunBlockedPermissions, 1},
		{"no reintenta el rechazo por calidad", Policy{MaxAttempts: 3, RetryOn: retryTimeout}, script{ok(badAnswer)}, manifest.RunRejectedQuality, 1},
		{"reintenta dos veces y acepta", Policy{MaxAttempts: 3, RetryOn: retryTimeout}, script{timeout, timeout, ok(goodAnswer)}, manifest.RunAccepted, 3},
	} {
		var reqs []Request
		r, err := newExec(tc.steps.runner(&reqs), tc.pol).Execute(context.Background(), f, "c", 1)
		if err != nil || r.Status != tc.want || len(r.Attempts) != tc.attempts {
			t.Errorf("%s: %q con %d intentos (err %v), want %q con %d", tc.name, r.Status, len(r.Attempts), err, tc.want, tc.attempts)
		}
	}
}

func TestExecute_TimeoutFromContextDeadline(t *testing.T) {
	f := miniFixture(t, false)
	blockUntilDone := RunnerFunc(func(ctx context.Context, req Request) (Result, error) {
		<-ctx.Done()
		return Result{}, ctx.Err()
	})
	r, err := newExec(blockUntilDone, Policy{Timeout: 30 * time.Millisecond}).Execute(context.Background(), f, "c", 1)
	if err != nil || r.Status != manifest.RunTimeout || len(r.Attempts) != 1 || !strings.Contains(r.Attempts[0].Reason, "tiempo") {
		t.Errorf("Run = %+v err %v, want timeout por el plazo", r, err)
	}
}

func TestExecute_ClientErrorIsClientFailureWithCause(t *testing.T) {
	f := miniFixture(t, false)
	step := func(Request) (Result, error) {
		return Result{ExitCode: intp(3), Events: []Event{{"error", raw(`{"type":"error"}`)}}}, os.ErrPermission
	}
	var reqs []Request
	r, err := newExec(script{step}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c", 1)
	if err != nil || r.Status != manifest.RunClientFailure || len(r.Attempts) != 1 {
		t.Fatalf("Run = %+v err %v, want fallo del cliente", r, err)
	}
	a := r.Attempts[0]
	if !strings.Contains(a.Reason, os.ErrPermission.Error()) || *a.ExitCode != 3 || len(a.ErrorEvents) != 1 {
		t.Errorf("intento = %+v, want causa con el error, codigo 3 y el evento", a)
	}
}

func TestExecute_PendingReviewIsNotAFinalStatus(t *testing.T) {
	f := miniFixture(t, true)
	var reqs []Request
	r, err := newExec(script{ok(goodAnswer)}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c", 1)
	if err != nil || r.Status != StatusPendingReview || r.Status.Valid() {
		t.Errorf("Run = %q err %v, want pendiente_de_revisor (no valido como final)", r.Status, err)
	}
}

func TestExecute_HarnessErrors(t *testing.T) {
	f := miniFixture(t, false)
	var reqs []Request
	weird := func(Request) (Result, error) { return Result{Termination: "otra"}, nil }
	if _, err := newExec(script{weird}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c", 1); err == nil {
		t.Error("terminacion desconocida = nil, want error")
	}
	badDec := func(Request) (Result, error) {
		return Result{Termination: Completed, Answer: []byte(goodAnswer), Decisions: map[string]bool{"ZZ": true}}, nil
	}
	if _, err := newExec(script{badDec}.runner(&reqs), Policy{}).Execute(context.Background(), f, "c", 1); err == nil {
		t.Error("decision sobre criterio desconocido = nil, want error del arnes")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newExec(script{ok(goodAnswer)}.runner(&reqs), Policy{}).Execute(ctx, f, "c", 1); err == nil {
		t.Error("contexto cancelado = nil, want error")
	}
}

func TestExecute_RecordsTimingAndTrace(t *testing.T) {
	f := miniFixture(t, false)
	step := func(req Request) (Result, error) {
		req.Timer.WaitBegin("ap1")
		req.Timer.WaitEnd("ap1")
		req.Trace.RecordRead("principal", "Read", "a.txt")
		return ok(goodAnswer)(req)
	}
	var reqs []Request
	cfg := Config{Clock: &fixedClock{}, WaitsObservable: true, Reads: true, Tools: false, TraceReason: "sin herramientas"}
	r, err := NewExecutor(script{step}.runner(&reqs), cfg).Execute(context.Background(), f, "c", 1)
	if err != nil || len(r.Attempts) != 1 {
		t.Fatalf("Run = %+v err %v", r, err)
	}
	a := r.Attempts[0]
	if !a.Timing.Adjusted.Available || a.Timing.WaitTotal.Value != time.Second || a.Timing.Elapsed != 3*time.Second {
		t.Errorf("Timing = %+v, want espera 1s de 3s", a.Timing)
	}
	if a.Trace.Reads.Total != 1 || a.Trace.ToolCalls.Available {
		t.Errorf("Trace = %+v, want 1 lectura y llamadas no disponibles", a.Trace)
	}
}

func TestPolicyFromLimits(t *testing.T) {
	p := PolicyFromLimits(manifest.Limits{MaxAttempts: 3, TimeoutSeconds: 90})
	if p.MaxAttempts != 3 || p.Timeout != 90*time.Second {
		t.Errorf("Policy = %+v", p)
	}
	if z := PolicyFromLimits(manifest.Limits{}); z.MaxAttempts != 0 || z.Timeout != 0 {
		t.Errorf("limites sin fijar = %+v, want ceros", z)
	}
}

func TestExecutePlan_RunsOnlyExecutableCohortsAndRepetitions(t *testing.T) {
	f := miniFixture(t, false)
	perm := manifest.Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode", "acceptEdits"}}
	good := manifest.Cell{
		Client: manifest.ClientClaude, Version: "2.1.310", MinVersion: "2.1.310", OS: manifest.OSLinux,
		Interface: manifest.InterfaceTerminal, Modality: manifest.ModalityAPI,
		Permissions: perm, Observed: manifest.Observed{Permissions: perm},
	}
	old := good
	old.Version = "1.0.0"
	tasks := []manifest.TaskWeight{{TaskID: "mini", Weight: 1}}
	m := manifest.Manifest{
		ProtocolVersion: "1", RubricVersion: "1", FixtureRevision: "x", Variant: "actual",
		Cohorts: []manifest.Cohort{{ID: "ok-1", Cell: good, Tasks: tasks}, {ID: "vieja", Cell: old, Tasks: tasks}, {ID: "ok-2", Cell: good, Tasks: tasks}},
	}
	plan, err := m.Plan()
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	runner := RunnerFunc(func(ctx context.Context, req Request) (Result, error) {
		calls++
		return ok(goodAnswer)(req)
	})
	fx := map[string]*fixture.Fixture{"mini": f}
	runs, err := newExec(runner, Policy{}).ExecutePlan(context.Background(), plan, m, fx, 2)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range runs {
		got = append(got, r.CohortID+"#"+string(rune('0'+r.Repetition)))
	}
	want := []string{"ok-1#1", "ok-1#2", "ok-2#1", "ok-2#2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("corridas = %v, want %v (la incompatible no corre)", got, want)
	}
	if calls != 4 {
		t.Errorf("el runner se invoco %d veces, want 4", calls)
	}
	if _, err := newExec(runner, Policy{}).ExecutePlan(context.Background(), plan, m, map[string]*fixture.Fixture{}, 1); err == nil {
		t.Error("tarea sin fixture = nil, want error")
	}
	if _, err := newExec(runner, Policy{}).ExecutePlan(context.Background(), plan, m, fx, 0); err == nil {
		t.Error("0 repeticiones = nil, want error")
	}
}

func TestExecute_CancelDuringRunKeepsInterruptedAttempt(t *testing.T) {
	f := miniFixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	runner := RunnerFunc(func(ctx context.Context, req Request) (Result, error) {
		_ = req.Trace.RecordRead("principal", "Read", "a.txt")
		cancel()
		return Result{ExitCode: intp(1), Events: []Event{{Type: "turn.failed", Raw: json.RawMessage(`{"type":"turn.failed"}`)}}}, ctx.Err()
	})
	r, err := newExec(runner, Policy{MaxAttempts: 2, RetryOn: []manifest.RunStatus{manifest.RunClientFailure}}).Execute(ctx, f, "c", 1)
	if err == nil {
		t.Fatal("la cancelación del llamador debe devolver error")
	}
	if len(r.Attempts) != 1 || r.Status != StatusInterrupted {
		t.Fatalf("Run = %+v, want 1 intento conservado con estado %q", r, StatusInterrupted)
	}
	a := r.Attempts[0]
	if a.Status != StatusInterrupted || !strings.Contains(a.Reason, "cancelación del llamador") || a.Cause != "inicial" {
		t.Errorf("intento = %+v, want estado, causa de la interrupción y causa inicial", a)
	}
	if a.ExitCode == nil || *a.ExitCode != 1 || len(a.ErrorEvents) != 1 || a.Trace.Reads.Total != 1 {
		t.Errorf("intento = %+v, want código, evento de error y lecturas parciales conservados", a)
	}
	if StatusInterrupted.Valid() {
		t.Error("interrumpida no es un estado final válido de manifest")
	}
}

func TestExecute_CancelBetweenAttemptsKeepsPriorAttempts(t *testing.T) {
	f := miniFixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	runner := RunnerFunc(func(ctx context.Context, req Request) (Result, error) {
		cancel()
		return Result{Termination: ClientFailed}, nil
	})
	r, err := newExec(runner, Policy{MaxAttempts: 3, RetryOn: []manifest.RunStatus{manifest.RunClientFailure}}).Execute(ctx, f, "c", 1)
	if err == nil || len(r.Attempts) != 1 || r.Attempts[0].Status != manifest.RunClientFailure {
		t.Errorf("Run = %+v err %v, want el primer intento intacto y error", r, err)
	}
}
