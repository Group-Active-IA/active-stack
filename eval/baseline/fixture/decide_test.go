package fixture

import (
	"math/rand"
	"reflect"
	"testing"
)

// Regla esperada, escrita aparte de la produccion: rechazo si algun
// obligatorio esta incumplido; si no, pendiente si algun obligatorio esta
// pendiente; si no, aceptada. Los no obligatorios no cambian el estado.
func oracle(rs []CriterionResult) (Status, []string, []string) {
	var failed, pending []string
	for _, r := range rs {
		if !r.Mandatory {
			continue
		}
		switch r.Outcome {
		case NotMet:
			failed = append(failed, r.ID)
		case Pending:
			pending = append(pending, r.ID)
		}
	}
	switch {
	case len(failed) > 0:
		return StatusRejected, failed, pending
	case len(pending) > 0:
		return StatusPending, failed, pending
	}
	return StatusAccepted, failed, pending
}

// Propiedad exhaustiva: todas las combinaciones de resultado y obligatoriedad
// para 1..4 criterios, y la misma combinacion permutada da el mismo estado.
func TestDecide_PropertyExhaustiveAndPermutation(t *testing.T) {
	outcomes := []Outcome{Met, NotMet, Pending}
	ids := []string{"A", "B", "C", "D"}
	rng := rand.New(rand.NewSource(20261111))
	checked := 0
	for n := 1; n <= 4; n++ {
		total := 1
		for i := 0; i < n; i++ {
			total *= 6
		}
		for code := 0; code < total; code++ {
			rs := make([]CriterionResult, n)
			c := code
			anyMandatory := false
			for i := 0; i < n; i++ {
				rs[i] = CriterionResult{ID: ids[i], Outcome: outcomes[c%3], Mandatory: (c/3)%2 == 1}
				anyMandatory = anyMandatory || rs[i].Mandatory
				c /= 6
			}
			v, err := Decide(rs)
			if !anyMandatory {
				if err == nil {
					t.Fatalf("%+v sin obligatorios: err = nil, want error", rs)
				}
				continue
			}
			if err != nil {
				t.Fatalf("%+v: err = %v", rs, err)
			}
			st, failed, pending := oracle(rs)
			if v.Status != st || !reflect.DeepEqual(v.Failed, failed) || !reflect.DeepEqual(v.Pending, pending) {
				t.Fatalf("%+v: Decide = {%s %v %v}, want {%s %v %v}", rs, v.Status, v.Failed, v.Pending, st, failed, pending)
			}
			if !reflect.DeepEqual(v.Criteria, rs) {
				t.Fatalf("Decide no conserva el resultado por criterio: %+v", v.Criteria)
			}
			perm := make([]CriterionResult, n)
			for i, j := range rng.Perm(n) {
				perm[i] = rs[j]
			}
			if pv, _ := Decide(perm); pv.Status != v.Status {
				t.Fatalf("estado depende del orden: %+v -> %s, %+v -> %s", rs, v.Status, perm, pv.Status)
			}
			checked++
		}
	}
	if checked < 1000 {
		t.Fatalf("solo %d combinaciones con obligatorios", checked)
	}
}

func TestDecide_Examples(t *testing.T) {
	v, err := Decide([]CriterionResult{
		{ID: "S1", Mandatory: true, Outcome: Met},
		{ID: "S2", Mandatory: true, Outcome: NotMet},
		{ID: "X", Mandatory: false, Outcome: NotMet},
	})
	if err != nil || v.Status != StatusRejected || !reflect.DeepEqual(v.Failed, []string{"S2"}) {
		t.Errorf("una prueba pasa y un obligatorio queda sin evidencia: %+v, %v; want rechazada [S2]", v, err)
	}
	v, _ = Decide([]CriterionResult{{ID: "A", Mandatory: true, Outcome: Met}, {ID: "X", Mandatory: false, Outcome: NotMet}})
	if v.Status != StatusAccepted {
		t.Errorf("no obligatorio incumplido: status = %s, want aceptada", v.Status)
	}
	if _, err := Decide([]CriterionResult{{ID: "A", Mandatory: true, Outcome: "cumplio"}}); err == nil {
		t.Error("resultado desconocido: err = nil, want error")
	}
	if _, err := Decide(nil); err == nil {
		t.Error("sin criterios: err = nil, want error")
	}
}
