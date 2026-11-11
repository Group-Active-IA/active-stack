// Package run implementa el ciclo preparacion -> ejecucion -> intento ->
// evaluacion de una corrida en un checkout descartable (tarea 2.1 de
// opsx-baseline). La ejecucion del cliente real queda detras de la interfaz
// Runner: aqui solo se prueba con runners falsos. Reutiliza manifest (Plan y
// estados de corrida) y fixture (Prepare/Reset/Close, Evaluate).
package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/timing"
	"github.com/Group-Active-IA/active-stack/eval/baseline/trace"
)

// Termination es como termino la ejecucion segun el Runner; el arnes decide el
// estado de la corrida a partir de ella y del veredicto.
type Termination string

const (
	Completed         Termination = "completado"
	TimedOut          Termination = "timeout"
	PermissionBlocked Termination = "bloqueada_por_permisos"
	ClientFailed      Termination = "fallo_del_cliente"
)

// StatusPendingReview no es un estado final de manifest.RunStatus: la
// evaluacion espera la decision del revisor.
const StatusPendingReview manifest.RunStatus = "pendiente_de_revisor"

// StatusInterrupted es el estado del intento en curso cuando el llamador cancela
// durante la corrida; no es un estado final de manifest.RunStatus.
const StatusInterrupted manifest.RunStatus = "interrumpida_por_llamador"

// Event es un evento bruto del cliente. Raw se conserva tal cual; el arnes no
// le asigna semantica.
type Event struct {
	Type string
	Raw  json.RawMessage
}

// ErrorEventTypes son los tipos de evento que se copian a ErrorEvents.
var ErrorEventTypes = []string{"turn.failed", "error"}

// Result es lo que devuelve un Runner.
type Result struct {
	Termination Termination
	ExitCode    *int // nil: el cliente no devolvio codigo de salida
	Detail      string
	Answer      []byte
	Decisions   map[string]bool
	Review      map[string]fixture.ReviewCall
	Events      []Event
}

// WaitTimer es lo que el Runner usa para marcar esperas de aprobacion.
type WaitTimer interface {
	WaitBegin(id string) error
	WaitEnd(id string) error
}

// Request es lo que recibe un Runner.
type Request struct {
	CohortID     string
	TaskID       string
	Repetition   int
	Attempt      int
	Dir          string
	Prompt       []byte
	AllowedTools []string
	Timer        WaitTimer
	Trace        *trace.Log
}

// Runner ejecuta el cliente en req.Dir.
type Runner interface {
	Run(ctx context.Context, req Request) (Result, error)
}

// RunnerFunc adapta una funcion a Runner.
type RunnerFunc func(ctx context.Context, req Request) (Result, error)

// Run implementa Runner.
func (f RunnerFunc) Run(ctx context.Context, req Request) (Result, error) { return f(ctx, req) }

// Policy son los limites de una corrida. Cero significa no fijado (D3).
type Policy struct {
	MaxAttempts int
	Timeout     time.Duration
	RetryOn     []manifest.RunStatus
}

// Config configura el Executor.
type Config struct {
	Clock           timing.Clock
	WaitsObservable bool
	Reads, Tools    trace.Signal
	TraceReason     string
	Policy          Policy
}

// Attempt es un intento con su estado, causa y datos brutos.
type Attempt struct {
	N           int
	Cause       string
	Status      manifest.RunStatus
	Reason      string
	ExitCode    *int
	ErrorEvents []Event
	Events      int
	Verdict     *fixture.Verdict
	Timing      timing.Measurement
	Trace       trace.Summary
}

// Run es una corrida: sus intentos y el estado final.
type Run struct {
	CohortID   string
	TaskID     string
	Repetition int
	Status     manifest.RunStatus
	Attempts   []Attempt
}

// Executor ejecuta corridas.
type Executor struct {
	runner Runner
	cfg    Config
}

// NewExecutor crea un Executor.
func NewExecutor(r Runner, cfg Config) *Executor { return &Executor{runner: r, cfg: cfg} }

func (p Policy) retries(s manifest.RunStatus) bool {
	for _, r := range p.RetryOn {
		if r == s {
			return true
		}
	}
	return false
}

// Execute corre una repeticion de f: prepara un checkout descartable, ejecuta
// el Runner, evalua y, si la politica lo permite, reintenta desde el fixture
// reiniciado. Cada intento queda registrado. Los errores son del arnes
// (preparacion, evaluacion, terminacion desconocida) o una cancelacion del
// contexto del llamador; en ese caso se devuelve la corrida parcial con el
// intento en curso (estado StatusInterrupted).
func (e *Executor) Execute(ctx context.Context, f *fixture.Fixture, cohortID string, rep int) (Run, error) {
	r := Run{CohortID: cohortID, TaskID: f.Task.ID, Repetition: rep}
	co, err := f.Prepare()
	if err != nil {
		return r, err
	}
	defer co.Close()
	max := e.cfg.Policy.MaxAttempts
	if max < 1 {
		max = 1
	}
	for n := 1; n <= max; n++ {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		cause := "inicial"
		if n > 1 {
			cause = "reintento tras " + string(r.Attempts[n-2].Status)
			if err := co.Reset(); err != nil {
				return r, err
			}
		}
		a, err := e.attempt(ctx, f, co, cohortID, rep, n, cause)
		if a != nil {
			r.Attempts = append(r.Attempts, *a)
			r.Status = a.Status
		}
		if err != nil {
			return r, err
		}
		if !e.cfg.Policy.retries(a.Status) {
			break
		}
	}
	return r, nil
}

func (e *Executor) attempt(parent context.Context, f *fixture.Fixture, co *fixture.Checkout, cohortID string, rep, n int, cause string) (*Attempt, error) {
	ctx := parent
	if e.cfg.Policy.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(parent, e.cfg.Policy.Timeout)
		defer cancel()
	}
	timer := timing.NewRecorder(e.cfg.Clock, e.cfg.WaitsObservable)
	log := trace.NewLog(e.cfg.Reads, e.cfg.Tools, e.cfg.TraceReason)
	if err := timer.Start(); err != nil {
		return nil, err
	}
	res, runErr := e.runner.Run(ctx, Request{
		CohortID: cohortID, TaskID: f.Task.ID, Repetition: rep, Attempt: n,
		Dir: co.Dir, Prompt: f.Prompt, AllowedTools: f.Task.AllowedTools, Timer: timer, Trace: log,
	})
	m, err := timer.Stop()
	if err != nil {
		return nil, err
	}
	a := &Attempt{N: n, Cause: cause, ExitCode: res.ExitCode, Events: len(res.Events), Timing: m, Trace: log.Summary(), Reason: res.Detail}
	for _, ev := range res.Events {
		for _, t := range ErrorEventTypes {
			if ev.Type == t {
				a.ErrorEvents = append(a.ErrorEvents, ev)
			}
		}
	}
	term := res.Termination
	if runErr != nil {
		switch {
		case parent.Err() != nil:
			// Interrumpida por el llamador: se conserva el intento en curso
			// con su estado, causa y datos parciales, y se devuelve el error.
			a.Status = StatusInterrupted
			a.Reason = "cancelación del llamador: " + parent.Err().Error()
			return a, parent.Err()
		case errors.Is(runErr, context.DeadlineExceeded):
			term = TimedOut
		default:
			term = ClientFailed
			a.Reason = runErr.Error()
		}
	}
	switch term {
	case TimedOut:
		a.Status = manifest.RunTimeout
		if a.Reason == "" {
			a.Reason = "tiempo agotado"
		}
	case PermissionBlocked:
		a.Status = manifest.RunBlockedPermissions
	case ClientFailed:
		a.Status = manifest.RunClientFailure
	case Completed:
		v, err := f.Evaluate(fixture.Output{Dir: co.Dir, Answer: res.Answer, Decisions: res.Decisions, Review: res.Review})
		if err != nil {
			return nil, err
		}
		a.Verdict = &v
		switch v.Status {
		case fixture.StatusAccepted:
			a.Status = manifest.RunAccepted
		case fixture.StatusRejected:
			a.Status = manifest.RunRejectedQuality
			a.Reason = "criterios obligatorios incumplidos: " + strings.Join(v.Failed, ", ")
		case fixture.StatusPending:
			a.Status = StatusPendingReview
			a.Reason = "pendiente de revisor: " + strings.Join(v.Pending, ", ")
		default:
			return nil, fmt.Errorf("veredicto desconocido %q", v.Status)
		}
	default:
		return nil, fmt.Errorf("terminación desconocida %q", term)
	}
	return a, nil
}

// PolicyFromLimits convierte los limites del manifiesto; cero es no fijado.
func PolicyFromLimits(l manifest.Limits) Policy {
	return Policy{MaxAttempts: l.MaxAttempts, Timeout: time.Duration(l.TimeoutSeconds) * time.Second}
}

// ExecutePlan corre, en el orden del manifiesto, las cohortes ejecutables del
// plan (las incompatibles no corren) con reps repeticiones por tarea. El orden
// entre repeticiones y su alternancia dependen de D3 (TBD): aqui es secuencial.
func (e *Executor) ExecutePlan(ctx context.Context, p manifest.Plan, m manifest.Manifest, fx map[string]*fixture.Fixture, reps int) ([]Run, error) {
	if reps < 1 {
		return nil, fmt.Errorf("repeticiones debe ser al menos 1")
	}
	byID := map[string]manifest.Cohort{}
	for _, c := range m.Cohorts {
		byID[c.ID] = c
	}
	var runs []Run
	for _, pc := range p.Executable {
		c, ok := byID[pc.CohortID]
		if !ok {
			return runs, fmt.Errorf("cohorte %q del plan no está en el manifiesto", pc.CohortID)
		}
		for _, tw := range c.Tasks {
			f, ok := fx[tw.TaskID]
			if !ok {
				return runs, fmt.Errorf("cohorte %q: no hay fixture para la tarea %q", c.ID, tw.TaskID)
			}
			for rep := 1; rep <= reps; rep++ {
				r, err := e.Execute(ctx, f, c.ID, rep)
				runs = append(runs, r)
				if err != nil {
					return runs, err
				}
			}
		}
	}
	return runs, nil
}
