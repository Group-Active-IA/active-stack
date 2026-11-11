package review

import (
	"reflect"
	"testing"
)

// F1 (verify-1-breaker): el emparejamiento es uno a uno maximo e
// independiente del orden de la clave.
func TestMatch_MaximumMatchingIndependentOfKeyOrder(t *testing.T) {
	dA := Defect{ID: "A", File: "x.go", StartLine: 10, EndLine: 20, Mechanism: "nil deref"}
	dB := Defect{ID: "B", File: "x.go", StartLine: 15, EndLine: 25, Mechanism: "nil deref"}
	findings := []Finding{
		{ID: "f1", File: "x.go", StartLine: 10, EndLine: 12, Mechanism: "nil deref"},
		{ID: "f2", File: "x.go", StartLine: 16, EndLine: 18, Mechanism: "nil deref"},
	}
	want := Counts{TruePositives: 2}
	var pairs []map[string]string
	for _, k := range [][]Defect{{dA, dB}, {dB, dA}} {
		r, err := Match(k, findings)
		if err != nil {
			t.Fatal(err)
		}
		if r.Counts != want || len(r.Omissions) != 0 {
			t.Errorf("clave %s,%s: counts=%+v omissions=%v, want %+v sin omisiones", k[0].ID, k[1].ID, r.Counts, r.Omissions, want)
		}
		p := map[string]string{}
		for _, c := range r.Findings {
			p[c.FindingID] = c.DefectID
		}
		pairs = append(pairs, p)
	}
	if !reflect.DeepEqual(pairs[0], pairs[1]) {
		t.Errorf("asignacion depende del orden de la clave: %v vs %v", pairs[0], pairs[1])
	}
}

// F1: un hallazgo con varios candidatos sigue cubriendo un solo defecto
// y los repetidos sobre un mismo defecto siguen sin sumar.
func TestMatch_MaximumMatchingStillOneToOne(t *testing.T) {
	k := []Defect{
		{ID: "B", File: "a.go", StartLine: 5, EndLine: 15, Mechanism: "m"},
		{ID: "A", File: "a.go", StartLine: 1, EndLine: 10, Mechanism: "m"},
	}
	r, err := Match(k, []Finding{
		{ID: "F1", File: "a.go", StartLine: 6, EndLine: 9, Mechanism: "m"},
		{ID: "F2", File: "a.go", StartLine: 6, EndLine: 9, Mechanism: "m"},
		{ID: "F3", File: "a.go", StartLine: 7, EndLine: 7, Mechanism: "m"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Counts.TruePositives != 2 || r.Counts.Duplicates != 1 || r.Counts.Omissions != 0 {
		t.Errorf("counts = %+v, want 2 TP, 1 duplicado, 0 omisiones", r.Counts)
	}
	// Desempate estable documentado: hallazgos en orden de lista, candidatos por ID.
	if c := classOf(t, r, "F1"); c.Class != TruePositive || c.DefectID != "A" {
		t.Errorf("F1 = %+v, want TP de A (menor ID)", c)
	}
	if c := classOf(t, r, "F2"); c.Class != TruePositive || c.DefectID != "B" {
		t.Errorf("F2 = %+v, want TP de B", c)
	}
	if c := classOf(t, r, "F3"); c.Class != Duplicate || c.DefectID != "A" || !reflect.DeepEqual(c.Candidates, []string{"A", "B"}) {
		t.Errorf("F3 = %+v, want duplicado de A con candidatos ordenados A,B", c)
	}
}

// F7 (M_emptymech): mecanismo vacio en clave y hallazgo no empareja.
func TestMatch_EmptyMechanismOnBothSidesDoesNotMatch(t *testing.T) {
	k := []Defect{{ID: "D", File: "a.go", StartLine: 1, EndLine: 5, Mechanism: ""}}
	r, err := Match(k, []Finding{{ID: "F", File: "a.go", StartLine: 1, EndLine: 5, Mechanism: "  ", Disposition: DispositionRejected}})
	if err != nil {
		t.Fatal(err)
	}
	if c := classOf(t, r, "F"); c.Class != FalsePositive || r.Counts.TruePositives != 0 {
		t.Errorf("F = %+v counts=%+v, want falso positivo sin TP", c, r.Counts)
	}
}

// F1: hace falta un camino de aumento: el primer hallazgo toma A, el segundo
// solo puede ir a A, y el maximo exige mover el primero a B.
func TestMatch_MaximumMatchingNeedsAugmentingPath(t *testing.T) {
	k := []Defect{
		{ID: "A", File: "a.go", StartLine: 1, EndLine: 10, Mechanism: "m"},
		{ID: "B", File: "a.go", StartLine: 8, EndLine: 20, Mechanism: "m"},
	}
	r, err := Match(k, []Finding{
		{ID: "F1", File: "a.go", StartLine: 9, EndLine: 9, Mechanism: "m"},
		{ID: "F2", File: "a.go", StartLine: 2, EndLine: 3, Mechanism: "m"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Counts.TruePositives != 2 || r.Counts.Omissions != 0 {
		t.Errorf("counts = %+v, want 2 TP", r.Counts)
	}
	if classOf(t, r, "F1").DefectID != "B" || classOf(t, r, "F2").DefectID != "A" {
		t.Errorf("asignacion = %+v", r.Findings)
	}
}
