package trace

import (
	"reflect"
	"strings"
	"testing"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// Traza conocida: 5 lecturas (principal 3, una repetida; worker 2) y 4
// llamadas (principal Bash x2 y Edit; worker Bash).
func knownLog(t *testing.T) *Log {
	l := NewLog(true, true, "")
	must(t, l.RecordRead("principal", "Read", "a.go"))
	must(t, l.RecordRead("principal", "Read", "b.go"))
	must(t, l.RecordRead("principal", "Read", "a.go"))
	must(t, l.RecordRead("worker", "Read", "c.go"))
	must(t, l.RecordRead("worker", "Grep", "a.go"))
	must(t, l.RecordToolCall("principal", "Bash"))
	must(t, l.RecordToolCall("principal", "Edit"))
	must(t, l.RecordToolCall("principal", "Bash"))
	must(t, l.RecordToolCall("worker", "Bash"))
	return l
}

func TestSummary_KnownTraceExactCounts(t *testing.T) {
	s := knownLog(t).Summary()
	wantReads := Counts{
		Available: true, Total: 5, Unique: 3,
		ByType:     map[string]int{"Read": 4, "Grep": 1},
		ByRole:     map[string]int{"principal": 3, "worker": 2},
		ByTypeRole: map[string]map[string]int{"Read": {"principal": 3, "worker": 1}, "Grep": {"worker": 1}},
	}
	wantTools := Counts{
		Available: true, Total: 4,
		ByType:     map[string]int{"Bash": 3, "Edit": 1},
		ByRole:     map[string]int{"principal": 3, "worker": 1},
		ByTypeRole: map[string]map[string]int{"Bash": {"principal": 2, "worker": 1}, "Edit": {"principal": 1}},
	}
	if !reflect.DeepEqual(s.Reads, wantReads) {
		t.Errorf("Reads = %+v\nwant   %+v", s.Reads, wantReads)
	}
	if !reflect.DeepEqual(s.ToolCalls, wantTools) {
		t.Errorf("ToolCalls = %+v\nwant   %+v", s.ToolCalls, wantTools)
	}
}

func TestSummary_ReadsDoNotCountAsToolCalls(t *testing.T) {
	l := NewLog(true, true, "")
	must(t, l.RecordRead("principal", "Read", "a.go"))
	s := l.Summary()
	if s.Reads.Total != 1 || s.ToolCalls.Total != 0 {
		t.Errorf("lecturas %d, llamadas %d, want 1 y 0", s.Reads.Total, s.ToolCalls.Total)
	}
}

func TestSummary_EmptyWithSignalIsRealZero(t *testing.T) {
	s := NewLog(true, true, "").Summary()
	if !s.Reads.Available || !s.ToolCalls.Available || s.Reads.Total != 0 || s.ToolCalls.Total != 0 {
		t.Errorf("s = %+v, want cero real disponible", s)
	}
}

func TestSummary_NoSignalIsUnavailableNotZero(t *testing.T) {
	s := NewLog(false, false, "la interfaz no expone eventos de herramientas").Summary()
	for name, c := range map[string]Counts{"lecturas": s.Reads, "llamadas": s.ToolCalls} {
		if c.Available || c.Total != 0 || !strings.Contains(c.Reason, "no expone") {
			t.Errorf("%s = %+v, want no disponible con motivo", name, c)
		}
	}
}

func TestSummary_SignalsAreIndependent(t *testing.T) {
	l := NewLog(true, false, "sin eventos de herramientas")
	must(t, l.RecordRead("principal", "Read", "a.go"))
	s := l.Summary()
	if !s.Reads.Available || s.Reads.Total != 1 || s.ToolCalls.Available {
		t.Errorf("s = %+v, want lecturas 1 y llamadas no disponibles", s)
	}
}

func TestSummary_NoSignalWithoutReasonStillHasOne(t *testing.T) {
	s := NewLog(false, true, "").Summary()
	if s.Reads.Available || s.Reads.Reason == "" {
		t.Errorf("Reads = %+v, want no disponible con motivo por defecto", s.Reads)
	}
}

func TestRecord_Errors(t *testing.T) {
	l := NewLog(false, false, "x")
	if err := l.RecordRead("principal", "Read", "a.go"); err == nil {
		t.Error("lectura sin senal = nil, want error")
	}
	if err := l.RecordToolCall("principal", "Bash"); err == nil {
		t.Error("llamada sin senal = nil, want error")
	}
	l = NewLog(true, true, "")
	for _, tc := range []struct{ role, typ, path string }{
		{"", "Read", "a"}, {"p", "", "a"}, {"p", "Read", ""}, {" ", "Read", "a"},
	} {
		if err := l.RecordRead(tc.role, tc.typ, tc.path); err == nil {
			t.Errorf("RecordRead(%q,%q,%q) = nil, want error", tc.role, tc.typ, tc.path)
		}
	}
	for _, tc := range []struct{ role, typ string }{{"", "Bash"}, {"p", ""}, {"p", "  "}} {
		if err := l.RecordToolCall(tc.role, tc.typ); err == nil {
			t.Errorf("RecordToolCall(%q,%q) = nil, want error", tc.role, tc.typ)
		}
	}
	if s := l.Summary(); s.Reads.Total != 0 || s.ToolCalls.Total != 0 {
		t.Errorf("los registros invalidos se contaron: %+v", s)
	}
}

func TestSummary_IsACopy(t *testing.T) {
	l := knownLog(t)
	s := l.Summary()
	s.Reads.ByRole["principal"] = 99
	s.ToolCalls.ByTypeRole["Bash"]["principal"] = 99
	if s2 := l.Summary(); s2.Reads.ByRole["principal"] != 3 || s2.ToolCalls.ByTypeRole["Bash"]["principal"] != 2 {
		t.Errorf("modificar el resumen altero el registro: %+v", s2)
	}
}
