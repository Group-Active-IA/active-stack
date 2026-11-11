// Package timing mide el tiempo de una corrida (tarea 2.5 de opsx-baseline):
// reloj monotonico, esperas de aprobacion observables aparte y union de
// intervalos solapados. El tiempo ajustado (sin esperas) es "no disponible"
// si alguna espera no se puede delimitar (design.md, "Medicion del tiempo").
package timing

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Clock es un reloj monotonico: Now devuelve el tiempo transcurrido desde un
// origen propio. Nunca depende de la hora de pared.
type Clock interface{ Now() time.Duration }

// Interval es [Start, End) medido por el reloj de la corrida.
type Interval struct{ Start, End time.Duration }

// Union une los intervalos solapados o contiguos. Devuelve intervalos
// ordenados y disjuntos; ignora los vacios o invertidos.
func Union(ivs []Interval) []Interval {
	var in []Interval
	for _, x := range ivs {
		if x.End > x.Start {
			in = append(in, x)
		}
	}
	sort.Slice(in, func(i, j int) bool { return in[i].Start < in[j].Start })
	var out []Interval
	for _, x := range in {
		if n := len(out); n > 0 && x.Start <= out[n-1].End {
			if x.End > out[n-1].End {
				out[n-1].End = x.End
			}
			continue
		}
		out = append(out, x)
	}
	return out
}

// Total es la duracion cubierta por la union de los intervalos.
func Total(ivs []Interval) time.Duration {
	var d time.Duration
	for _, x := range Union(ivs) {
		d += x.End - x.Start
	}
	return d
}

// Metric es un valor que puede no estar disponible. Un valor ausente no es
// cero: Available false lleva el motivo.
type Metric struct {
	Available bool
	Value     time.Duration
	Reason    string
}

// Measurement es el tiempo de una corrida: transcurrido siempre disponible;
// esperas y ajustado solo si se pudieron delimitar.
type Measurement struct {
	Elapsed   time.Duration
	Waits     []Interval
	WaitTotal Metric
	Adjusted  Metric
}

// Recorder mide una corrida con un Clock inyectado.
type Recorder struct {
	clock      Clock
	observable bool
	started    bool
	stopped    bool
	start      time.Duration
	open       map[string]time.Duration
	waits      []Interval
	problems   []string
}

// NewRecorder crea un Recorder. waitsObservable indica si la interfaz expone
// solicitud y resolucion de aprobaciones; si es false, esperas y ajustado
// quedan no disponibles.
func NewRecorder(c Clock, waitsObservable bool) *Recorder {
	return &Recorder{clock: c, observable: waitsObservable, open: map[string]time.Duration{}}
}

// Start marca el inicio de la tarea.
func (r *Recorder) Start() error {
	if r.started {
		return errors.New("timing: Start repetido")
	}
	r.started = true
	r.start = r.clock.Now()
	return nil
}

func (r *Recorder) running() error {
	switch {
	case !r.started:
		return errors.New("timing: la tarea no inició")
	case r.stopped:
		return errors.New("timing: la tarea ya terminó")
	}
	return nil
}

// WaitBegin marca la solicitud de una espera (por ejemplo una aprobacion). Una
// solicitud duplicada de un id abierto es una espera no delimitable: se
// registra y no es un error de uso.
func (r *Recorder) WaitBegin(id string) error {
	if err := r.running(); err != nil {
		return err
	}
	now := r.clock.Now()
	if _, dup := r.open[id]; dup {
		r.problems = append(r.problems, fmt.Sprintf("solicitud duplicada de la espera %q", id))
		return nil
	}
	r.open[id] = now
	return nil
}

// WaitEnd marca la resolucion de la espera id. Una resolucion sin solicitud
// es una espera no delimitable: se registra y no es un error de uso.
func (r *Recorder) WaitEnd(id string) error {
	if err := r.running(); err != nil {
		return err
	}
	now := r.clock.Now()
	begin, ok := r.open[id]
	if !ok {
		r.problems = append(r.problems, fmt.Sprintf("resolución de la espera %q sin solicitud", id))
		return nil
	}
	delete(r.open, id)
	if now < begin {
		r.problems = append(r.problems, fmt.Sprintf("la espera %q terminó antes de empezar (reloj retrocedió)", id))
		return nil
	}
	r.waits = append(r.waits, Interval{begin, now})
	return nil
}

// Stop marca el fin de la tarea y calcula la medicion.
func (r *Recorder) Stop() (Measurement, error) {
	if err := r.running(); err != nil {
		return Measurement{}, err
	}
	end := r.clock.Now()
	if end < r.start {
		return Measurement{}, errors.New("timing: el reloj retrocedió")
	}
	r.stopped = true
	m := Measurement{Elapsed: end - r.start}
	problems := append([]string(nil), r.problems...)
	for id := range r.open {
		problems = append(problems, fmt.Sprintf("espera %q sin resolución", id))
	}
	sort.Strings(problems)
	var reason string
	switch {
	case !r.observable:
		reason = "la interfaz no expone esperas de aprobación"
	case len(problems) > 0:
		reason = "espera no delimitable: " + strings.Join(problems, "; ")
	}
	if reason != "" {
		m.WaitTotal = Metric{Reason: reason}
		m.Adjusted = Metric{Reason: reason}
		return m, nil
	}
	m.Waits = Union(r.waits)
	w := Total(r.waits)
	m.WaitTotal = Metric{Available: true, Value: w}
	m.Adjusted = Metric{Available: true, Value: m.Elapsed - w}
	return m, nil
}

type systemClock struct{ origin time.Time }

// Now usa time.Since, que lee la parte monotonica de origin.
func (c systemClock) Now() time.Duration { return time.Since(c.origin) }

// SystemClock es el reloj monotonico real: time.Since de un origen tomado con
// time.Now, que no se ve afectado por cambios de la hora de pared.
func SystemClock() Clock { return systemClock{origin: time.Now()} }
