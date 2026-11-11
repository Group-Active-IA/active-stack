package review

import (
	"reflect"
	"testing"
)

func key() []Defect {
	return []Defect{
		{ID: "D1", File: "pkg/a.go", StartLine: 10, EndLine: 20, Mechanism: "nil-deref"},
		{ID: "D2", File: "pkg/b.go", StartLine: 5, EndLine: 8, Mechanism: "off-by-one"},
	}
}

func classOf(t *testing.T, r Result, findingID string) Classified {
	t.Helper()
	for _, c := range r.Findings {
		if c.FindingID == findingID {
			return c
		}
	}
	t.Fatalf("hallazgo %q sin clasificar en %+v", findingID, r.Findings)
	return Classified{}
}

func TestMatch_TruePositive(t *testing.T) {
	r, err := Match(key(), []Finding{
		{ID: "F1", File: "pkg/a.go", StartLine: 15, EndLine: 30, Mechanism: "nil-deref", Evidence: "linea 17 desreferencia p"},
	})
	if err != nil {
		t.Fatal(err)
	}
	c := classOf(t, r, "F1")
	if c.Class != TruePositive || c.DefectID != "D1" || c.Evidence == "" {
		t.Errorf("F1 = %+v, want verdadero positivo de D1 con evidencia", c)
	}
	if r.Counts.TruePositives != 1 {
		t.Errorf("TruePositives = %d, want 1", r.Counts.TruePositives)
	}
}

func TestMatch_ThreeConditionsAreAllRequired(t *testing.T) {
	tests := []struct {
		name string
		f    Finding
	}{
		{"otro archivo", Finding{File: "pkg/c.go", StartLine: 10, EndLine: 20, Mechanism: "nil-deref"}},
		{"rango sin solape (adyacente)", Finding{File: "pkg/a.go", StartLine: 21, EndLine: 25, Mechanism: "nil-deref"}},
		{"otro mecanismo", Finding{File: "pkg/a.go", StartLine: 10, EndLine: 20, Mechanism: "race"}},
		{"mecanismo vacio", Finding{File: "pkg/a.go", StartLine: 10, EndLine: 20, Mechanism: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.ID = "F1"
			tt.f.Disposition = DispositionRejected
			r, err := Match(key(), []Finding{tt.f})
			if err != nil {
				t.Fatal(err)
			}
			if c := classOf(t, r, "F1"); c.Class != FalsePositive {
				t.Errorf("clase = %q, want falso positivo", c.Class)
			}
			if r.Counts.TruePositives != 0 || r.Counts.Omissions != 2 {
				t.Errorf("counts = %+v, want 0 TP y 2 omisiones", r.Counts)
			}
		})
	}
}

func TestMatch_OverlapEdges(t *testing.T) {
	// Solape inclusivo en los extremos, contenido y contenedor.
	for _, tc := range []struct{ s, e int }{{20, 25}, {1, 10}, {12, 14}, {1, 99}} {
		r, err := Match(key(), []Finding{{ID: "F", File: "pkg/a.go", StartLine: tc.s, EndLine: tc.e, Mechanism: "nil-deref"}})
		if err != nil {
			t.Fatal(err)
		}
		if classOf(t, r, "F").Class != TruePositive {
			t.Errorf("rango %d-%d debería solapar con 10-20", tc.s, tc.e)
		}
	}
}

func TestMatch_DuplicateDoesNotAdd(t *testing.T) {
	r, err := Match(key(), []Finding{
		{ID: "F1", File: "pkg/a.go", StartLine: 12, EndLine: 12, Mechanism: "nil-deref"},
		{ID: "F2", File: "pkg/a.go", StartLine: 18, EndLine: 19, Mechanism: "NIL-DEREF "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c := classOf(t, r, "F1"); c.Class != TruePositive || c.DefectID != "D1" {
		t.Errorf("F1 = %+v", c)
	}
	if c := classOf(t, r, "F2"); c.Class != Duplicate || c.DefectID != "D1" {
		t.Errorf("F2 = %+v, want duplicado de D1", c)
	}
	if r.Counts.TruePositives != 1 || r.Counts.Duplicates != 1 {
		t.Errorf("counts = %+v, want 1 TP y 1 duplicado", r.Counts)
	}
}

func TestMatch_Omission(t *testing.T) {
	r, err := Match(key(), []Finding{
		{ID: "F1", File: "pkg/a.go", StartLine: 12, EndLine: 12, Mechanism: "nil-deref"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Omissions, []string{"D2"}) || r.Counts.Omissions != 1 {
		t.Errorf("omissions = %v / %+v, want [D2]", r.Omissions, r.Counts)
	}
}

func TestMatch_UnmatchedFindings(t *testing.T) {
	r, err := Match(key(), []Finding{
		{ID: "FP", File: "pkg/z.go", StartLine: 1, EndLine: 2, Mechanism: "x", Disposition: DispositionRejected},
		{ID: "NEW", File: "pkg/z.go", StartLine: 3, EndLine: 4, Mechanism: "y", Disposition: DispositionRealDefect},
		{ID: "PEND", File: "pkg/z.go", StartLine: 5, EndLine: 6, Mechanism: "w"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if classOf(t, r, "FP").Class != FalsePositive {
		t.Error("FP debería ser falso positivo")
	}
	if classOf(t, r, "NEW").Class != NewFinding {
		t.Error("NEW debería derivarse a adjudicación como hallazgo nuevo")
	}
	if classOf(t, r, "PEND").Class != PendingReview {
		t.Error("PEND sin decisión del revisor debería quedar pendiente")
	}
	want := Counts{FalsePositives: 1, NewFindings: 1, Pending: 1, Omissions: 2}
	if r.Counts != want {
		t.Errorf("counts = %+v, want %+v (el hallazgo nuevo no cambia TP/FP/omisiones)", r.Counts, want)
	}
}

func TestMatch_FindingPointingAtTwoDefects(t *testing.T) {
	k := []Defect{
		{ID: "D1", File: "a.go", StartLine: 1, EndLine: 10, Mechanism: "m"},
		{ID: "D2", File: "a.go", StartLine: 5, EndLine: 15, Mechanism: "m"},
	}
	r, err := Match(k, []Finding{{ID: "F1", File: "a.go", StartLine: 6, EndLine: 9, Mechanism: "m"}})
	if err != nil {
		t.Fatal(err)
	}
	c := classOf(t, r, "F1")
	if c.Class != TruePositive || c.DefectID != "D1" || !reflect.DeepEqual(c.Candidates, []string{"D1", "D2"}) {
		t.Errorf("F1 = %+v, want TP de D1 (primero de la clave) con candidatos D1,D2", c)
	}
	if r.Counts.TruePositives != 1 || !reflect.DeepEqual(r.Omissions, []string{"D2"}) {
		t.Errorf("un hallazgo no puede cubrir dos defectos: counts=%+v omissions=%v", r.Counts, r.Omissions)
	}
}

func TestMatch_FindingMatchingTakenDefectTakesFreeSecondDefect(t *testing.T) {
	k := []Defect{
		{ID: "D1", File: "a.go", StartLine: 1, EndLine: 10, Mechanism: "m"},
		{ID: "D2", File: "a.go", StartLine: 5, EndLine: 15, Mechanism: "m"},
	}
	r, err := Match(k, []Finding{
		{ID: "F1", File: "a.go", StartLine: 1, EndLine: 2, Mechanism: "m"},
		{ID: "F2", File: "a.go", StartLine: 6, EndLine: 9, Mechanism: "m"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c := classOf(t, r, "F2"); c.Class != TruePositive || c.DefectID != "D2" {
		t.Errorf("F2 = %+v, want TP de D2 (emparejamiento maximo, apply-2 F1)", c)
	}
	if r.Counts.TruePositives != 2 || r.Counts.Omissions != 0 {
		t.Errorf("counts = %+v", r.Counts)
	}
}

func TestMatch_EmptyInputs(t *testing.T) {
	r, err := Match(nil, []Finding{{ID: "F1", File: "a.go", StartLine: 1, EndLine: 1, Mechanism: "m", Disposition: DispositionRejected}})
	if err != nil {
		t.Fatal(err)
	}
	if classOf(t, r, "F1").Class != FalsePositive || r.Counts.Omissions != 0 {
		t.Errorf("clave vacía: %+v", r)
	}

	r, err = Match(key(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Omissions, []string{"D1", "D2"}) || len(r.Findings) != 0 || r.Counts.Omissions != 2 {
		t.Errorf("hallazgos vacíos: %+v", r)
	}

	r, err = Match(nil, nil)
	if err != nil || len(r.Findings) != 0 || len(r.Omissions) != 0 {
		t.Errorf("ambos vacíos: %+v err=%v", r, err)
	}
}

func TestMatch_PathNormalization(t *testing.T) {
	r, err := Match(key(), []Finding{{ID: "F1", File: `.\pkg\a.go`, StartLine: 10, EndLine: 10, Mechanism: "nil-deref"}})
	if err != nil {
		t.Fatal(err)
	}
	if classOf(t, r, "F1").Class != TruePositive {
		t.Error("separadores y ./ inicial no deben impedir el emparejamiento")
	}
}

func TestMatch_InvalidInputErrors(t *testing.T) {
	good := Finding{ID: "F", File: "a.go", StartLine: 1, EndLine: 1, Mechanism: "m"}
	tests := []struct {
		name string
		key  []Defect
		f    []Finding
	}{
		{"rango invertido en defecto", []Defect{{ID: "D", File: "a.go", StartLine: 5, EndLine: 1, Mechanism: "m"}}, nil},
		{"defecto sin id", []Defect{{File: "a.go", StartLine: 1, EndLine: 1, Mechanism: "m"}}, nil},
		{"ids de defecto duplicados", []Defect{{ID: "D", File: "a.go", StartLine: 1, EndLine: 1, Mechanism: "m"}, {ID: "D", File: "b.go", StartLine: 1, EndLine: 1, Mechanism: "m"}}, nil},
		{"hallazgo con rango invertido", nil, []Finding{{ID: "F", File: "a.go", StartLine: 3, EndLine: 2, Mechanism: "m"}}},
		{"hallazgo sin id", nil, []Finding{{File: "a.go", StartLine: 1, EndLine: 1}}},
		{"ids de hallazgo duplicados", nil, []Finding{good, good}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Match(tt.key, tt.f); err == nil {
				t.Error("Match() debería devolver error")
			}
		})
	}
}
