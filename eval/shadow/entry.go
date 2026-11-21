package shadow

import "strings"

// Candidate declara lo que un juez candidato documenta frente a los criterios de entrada de D6.
type Candidate struct {
	ID                string
	PackageOnly       bool // acepta el paquete de D5 sin datos adicionales
	LocatableFindings bool // devuelve hallazgos ubicables (D4)
	ObservableCost    bool // expone costo observable
	Terms             ProviderTerms
}

// EntryDecision es el resultado de la tarea 4.4. Un candidato no admitido no recibe casos y sus
// motivos van al informe.
type EntryDecision struct {
	Candidate string
	Admitted  bool
	Reasons   []string
}

// CheckEntry aplica los criterios de entrada de D6. Lista todos los motivos, no sólo el primero.
func CheckEntry(c Candidate) EntryDecision {
	var r []string
	if !c.PackageOnly {
		r = append(r, "pide datos fuera del paquete de D5")
	}
	if !c.LocatableFindings {
		r = append(r, "no devuelve hallazgos ubicables (D4)")
	}
	if !c.ObservableCost {
		r = append(r, "no expone costo observable")
	}
	r = append(r, c.Terms.missing()...)
	return EntryDecision{Candidate: c.ID, Admitted: len(r) == 0, Reasons: r}
}

// ProviderTerms son las condiciones declaradas por el proveedor (D5), verificadas por separado:
// un campo vacío es una condición no documentada.
type ProviderTerms struct {
	Provider  string `json:"proveedor"`
	Retention string `json:"retencion"`
	Training  string `json:"entrenamiento"`
}

func (t ProviderTerms) missing() []string {
	var r []string
	if strings.TrimSpace(t.Provider) == "" {
		r = append(r, "proveedor sin identificar")
	}
	if strings.TrimSpace(t.Retention) == "" {
		r = append(r, "retención de datos no documentada")
	}
	if strings.TrimSpace(t.Training) == "" {
		r = append(r, "uso para entrenamiento no documentado")
	}
	return r
}
