package review

import (
	"strings"
	"testing"
)

// DEUDA 6 (verify-1-auditor, S2): un defecto de la clave sin archivo o sin
// mecanismo seria una omision garantizada; Match lo rechaza con error.
func TestMatch_KeyDefectWithBlankFileOrMechanismIsError(t *testing.T) {
	cases := []struct {
		name string
		d    Defect
		want string
	}{
		{"archivo vacio", Defect{ID: "D", File: "", StartLine: 1, EndLine: 2, Mechanism: "m"}, "archivo"},
		{"archivo en blanco", Defect{ID: "D", File: " \t", StartLine: 1, EndLine: 2, Mechanism: "m"}, "archivo"},
		{"mecanismo vacio", Defect{ID: "D", File: "a.go", StartLine: 1, EndLine: 2, Mechanism: ""}, "mecanismo"},
		{"mecanismo en blanco", Defect{ID: "D", File: "a.go", StartLine: 1, EndLine: 2, Mechanism: "  "}, "mecanismo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Match([]Defect{tc.d}, nil)
			if err == nil {
				t.Fatalf("Match con defecto de %s = nil, want error", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), `"D"`) {
				t.Errorf("error = %q, want mencion de %q y del defecto \"D\"", err, tc.want)
			}
		})
	}
}

// DEUDA 6 (S4): una disposicion desconocida (typo del revisor) es error
// explicito, no un pendiente silencioso. Vale tambien para hallazgos
// emparejados, porque el typo sigue siendo un dato invalido.
func TestMatch_UnknownDispositionIsError(t *testing.T) {
	key := []Defect{{ID: "D", File: "a.go", StartLine: 1, EndLine: 5, Mechanism: "m"}}
	for _, disp := range []Disposition{"rejectd", "Rejected", "REAL_DEFECT", "real defect", " rejected", "pending"} {
		for _, matched := range []bool{false, true} {
			f := Finding{ID: "F", File: "z.go", StartLine: 1, EndLine: 1, Mechanism: "x", Disposition: disp}
			if matched {
				f.File, f.Mechanism = "a.go", "m"
			}
			_, err := Match(key, []Finding{f})
			if err == nil {
				t.Errorf("disposicion %q (emparejado=%v): err = nil, want error", disp, matched)
				continue
			}
			if !strings.Contains(err.Error(), "disposición") || !strings.Contains(err.Error(), `"F"`) {
				t.Errorf("disposicion %q: error = %q, want mencion de disposición y del hallazgo \"F\"", disp, err)
			}
		}
	}
}

// Negativo de miembros benignos (LECCION del auditor): las tres
// disposiciones validas siguen aceptadas y clasifican como antes.
func TestMatch_KnownDispositionsStillAccepted(t *testing.T) {
	key := []Defect{{ID: "D", File: "a.go", StartLine: 1, EndLine: 5, Mechanism: "m"}}
	want := map[Disposition]Class{
		DispositionUndecided:  PendingReview,
		DispositionRejected:   FalsePositive,
		DispositionRealDefect: NewFinding,
	}
	for disp, class := range want {
		r, err := Match(key, []Finding{{ID: "F", File: "z.go", StartLine: 1, EndLine: 1, Mechanism: "x", Disposition: disp}})
		if err != nil {
			t.Fatalf("disposicion %q: err = %v, want nil", disp, err)
		}
		if got := classOf(t, r, "F").Class; got != class {
			t.Errorf("disposicion %q: clase = %q, want %q", disp, got, class)
		}
	}
	// Un defecto con archivo y mecanismo con espacios alrededor sigue valido.
	if _, err := Match([]Defect{{ID: "D", File: " a.go ", StartLine: 1, EndLine: 1, Mechanism: " m "}}, nil); err != nil {
		t.Errorf("defecto con espacios alrededor: err = %v, want nil", err)
	}
}
