package timing

import (
	"strings"
	"testing"
	"time"
)

type manual struct{ t time.Duration }

func (m *manual) Now() time.Duration { return m.t }

func at(c *manual, t int) { c.t = time.Duration(t) }

func TestRecorder_NoWaitsAdjustedEqualsElapsed(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	at(c, 5)
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	at(c, 15)
	m, err := r.Stop()
	if err != nil {
		t.Fatal(err)
	}
	if m.Elapsed != 10 || !m.WaitTotal.Available || m.WaitTotal.Value != 0 || !m.Adjusted.Available || m.Adjusted.Value != 10 {
		t.Errorf("m = %+v, want elapsed 10, esperas 0 disponible, ajustado 10", m)
	}
}

func TestRecorder_OverlappingWaitsAreUnitedNotSummed(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	at(c, 0)
	r.Start()
	for _, ev := range []struct {
		t    int
		op   string
		id   string
		want error
	}{
		{2, "b", "a1", nil}, {4, "b", "a2", nil}, {6, "e", "a1", nil}, {9, "e", "a2", nil},
		{12, "b", "a3", nil}, {13, "e", "a3", nil},
	} {
		at(c, ev.t)
		var err error
		if ev.op == "b" {
			err = r.WaitBegin(ev.id)
		} else {
			err = r.WaitEnd(ev.id)
		}
		if err != nil {
			t.Fatalf("%s %s: %v", ev.op, ev.id, err)
		}
	}
	at(c, 20)
	m, err := r.Stop()
	if err != nil {
		t.Fatal(err)
	}
	// [2,9) solapadas = 7 y [12,13) = 1: union 8; la suma ingenua daria 10.
	if !m.WaitTotal.Available || m.WaitTotal.Value != 8 {
		t.Errorf("WaitTotal = %+v, want 8", m.WaitTotal)
	}
	if !m.Adjusted.Available || m.Adjusted.Value != 12 || m.Elapsed != 20 {
		t.Errorf("Adjusted = %+v elapsed %v, want 12 de 20", m.Adjusted, m.Elapsed)
	}
	if len(m.Waits) != 2 || m.Waits[0] != iv(2, 9) || m.Waits[1] != iv(12, 13) {
		t.Errorf("Waits = %v, want [2,9) y [12,13)", m.Waits)
	}
}

func TestRecorder_UndelimitableWaitMakesAdjustedUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		events func(r *Recorder, c *manual)
		reason string
	}{
		{"espera abierta al terminar", func(r *Recorder, c *manual) { at(c, 3); r.WaitBegin("x") }, "sin resolución"},
		{"resolucion sin solicitud", func(r *Recorder, c *manual) { at(c, 3); r.WaitEnd("x") }, "sin solicitud"},
		{"solicitud duplicada", func(r *Recorder, c *manual) {
			at(c, 3)
			r.WaitBegin("x")
			at(c, 4)
			r.WaitBegin("x")
			at(c, 5)
			r.WaitEnd("x")
		}, "duplicada"},
	} {
		c := &manual{}
		r := NewRecorder(c, true)
		r.Start()
		tc.events(r, c)
		at(c, 10)
		m, err := r.Stop()
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if m.Adjusted.Available || m.Adjusted.Value != 0 || !strings.Contains(m.Adjusted.Reason, tc.reason) {
			t.Errorf("%s: Adjusted = %+v, want no disponible con %q", tc.name, m.Adjusted, tc.reason)
		}
		if m.WaitTotal.Available {
			t.Errorf("%s: WaitTotal = %+v, want no disponible", tc.name, m.WaitTotal)
		}
		if m.Elapsed != 10 {
			t.Errorf("%s: Elapsed = %v, want 10 (siempre disponible)", tc.name, m.Elapsed)
		}
	}
}

func TestRecorder_IDCanBeReusedAfterItClosed(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	r.Start()
	at(c, 1)
	r.WaitBegin("x")
	at(c, 2)
	r.WaitEnd("x")
	at(c, 3)
	r.WaitBegin("x")
	at(c, 5)
	r.WaitEnd("x")
	at(c, 6)
	m, _ := r.Stop()
	if !m.Adjusted.Available || m.WaitTotal.Value != 3 || m.Adjusted.Value != 3 {
		t.Errorf("m = %+v, want esperas 3 y ajustado 3", m)
	}
}

func TestRecorder_WaitsNotObservableIsUnavailableNotZero(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, false)
	r.Start()
	at(c, 7)
	m, _ := r.Stop()
	if m.Adjusted.Available || m.WaitTotal.Available || m.Adjusted.Reason == "" {
		t.Errorf("m = %+v, want esperas y ajustado no disponibles con motivo", m)
	}
	if m.Elapsed != 7 {
		t.Errorf("Elapsed = %v, want 7", m.Elapsed)
	}
}

func TestRecorder_MisuseIsError(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	if _, err := r.Stop(); err == nil {
		t.Error("Stop sin Start = nil, want error")
	}
	if err := r.WaitBegin("x"); err == nil {
		t.Error("WaitBegin sin Start = nil, want error")
	}
	if err := r.WaitEnd("x"); err == nil {
		t.Error("WaitEnd sin Start = nil, want error")
	}
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	if err := r.Start(); err == nil {
		t.Error("Start doble = nil, want error")
	}
	at(c, 1)
	if _, err := r.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Stop(); err == nil {
		t.Error("Stop doble = nil, want error")
	}
	if err := r.WaitBegin("x"); err == nil {
		t.Error("WaitBegin tras Stop = nil, want error")
	}
}

func TestRecorder_ClockGoingBackwardsIsError(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	at(c, 10)
	r.Start()
	at(c, 4)
	if _, err := r.Stop(); err == nil {
		t.Error("Stop con reloj retrocedido = nil, want error")
	}
}

func TestSystemClock_MonotonicAndMeasuresRealTime(t *testing.T) {
	c := SystemClock()
	a := c.Now()
	time.Sleep(20 * time.Millisecond)
	b := c.Now()
	if b < a+15*time.Millisecond || b > a+5*time.Second {
		t.Errorf("Now() paso de %v a %v tras dormir 20ms", a, b)
	}
	prev := c.Now()
	for i := 0; i < 1000; i++ {
		n := c.Now()
		if n < prev {
			t.Fatalf("reloj retrocedio: %v -> %v", prev, n)
		}
		prev = n
	}
}

func TestRecorder_WaitEndingBeforeItBeganIsUndelimitable(t *testing.T) {
	c := &manual{}
	r := NewRecorder(c, true)
	r.Start()
	at(c, 8)
	r.WaitBegin("x")
	at(c, 5)
	r.WaitEnd("x")
	at(c, 10)
	m, _ := r.Stop()
	if m.Adjusted.Available || !strings.Contains(m.Adjusted.Reason, "retroced") {
		t.Errorf("Adjusted = %+v, want no disponible por reloj retrocedido", m.Adjusted)
	}
}

// El origen del reloj real conserva la lectura monotonica (String() muestra
// "m=+..."); sin ella time.Since mediria con la hora de pared. Que Now()
// use la lectura monotonica es verdad por construccion y no se puede forzar
// un salto de la hora de pared desde un test: limite declarado.
func TestSystemClock_OriginKeepsMonotonicReading(t *testing.T) {
	sc, ok := SystemClock().(systemClock)
	if !ok || !strings.Contains(sc.origin.String(), "m=") {
		t.Errorf("origen = %v, want con lectura monotonica", sc.origin)
	}
}
