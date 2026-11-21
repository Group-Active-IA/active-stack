package shadow

import "testing"

func fullCandidate() Candidate {
	return Candidate{ID: "juez-base", PackageOnly: true, LocatableFindings: true, ObservableCost: true, Terms: okTerms}
}

// 4.4: criterios de entrada de D6.
func TestCheckEntry_Admitted(t *testing.T) {
	d := CheckEntry(fullCandidate())
	if !d.Admitted || len(d.Reasons) != 0 {
		t.Fatalf("candidato completo rechazado: %+v", d)
	}
}

// Spec «Candidato sin política de datos declarada»: no se le envía nada y se registra el descarte.
func TestCheckEntry_EachMissingCriterionIsRecorded(t *testing.T) {
	for name, mut := range map[string]func(*Candidate){
		"pide datos adicionales":  func(c *Candidate) { c.PackageOnly = false },
		"hallazgos no ubicables":  func(c *Candidate) { c.LocatableFindings = false },
		"costo no observable":     func(c *Candidate) { c.ObservableCost = false },
		"sin retención declarada": func(c *Candidate) { c.Terms.Retention = "" },
		"sin entrenamiento":       func(c *Candidate) { c.Terms.Training = "" },
	} {
		c := fullCandidate()
		mut(&c)
		d := CheckEntry(c)
		if d.Admitted || len(d.Reasons) != 1 || d.Candidate != "juez-base" {
			t.Fatalf("%s: %+v", name, d)
		}
	}
}

func TestCheckEntry_AllReasonsListed(t *testing.T) {
	d := CheckEntry(Candidate{ID: "jev"})
	if d.Admitted || len(d.Reasons) != 6 {
		t.Fatalf("un candidato sin documentar debe listar todos los motivos: %+v", d)
	}
}
