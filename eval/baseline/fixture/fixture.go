// Package fixture carga, valida, materializa y evalua los fixtures de las
// cinco familias de tareas del change opsx-baseline (tarea 1.3).
package fixture

import "fmt"

// Family es la familia de tareas de un fixture.
type Family string

const (
	FamilyExploration Family = "exploracion"
	FamilySmallChange Family = "cambio-pequeno"
	FamilyMultiFile   Family = "cambio-multiarchivo"
	FamilyDiagnosis   Family = "diagnostico"
	FamilyReview      Family = "revision"
)

// Source es el estado inicial congelado del checkout.
type Source struct {
	Kind     string `json:"kind"`
	Revision string `json:"revision,omitempty"`
	Patch    string `json:"patch,omitempty"`
	Dir      string `json:"dir,omitempty"`
	SHA256   string `json:"sha256"`
}

// PromptRef es el prompt congelado y su huella.
type PromptRef struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// Criterion es un criterio de la rubrica.
type Criterion struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Mandatory   bool     `json:"mandatory"`
	Check       string   `json:"check"`
	Args        []string `json:"args,omitempty"`
}

// Task es el contenido de task.json.
type Task struct {
	ID            string            `json:"id"`
	Family        Family            `json:"family"`
	RubricVersion string            `json:"rubric_version"`
	Source        Source            `json:"source"`
	Prompt        PromptRef         `json:"prompt"`
	AllowedTools  []string          `json:"allowed_tools"`
	Limits        map[string]string `json:"limits"`
	Key           string            `json:"key,omitempty"`
	Reference     string            `json:"reference"`
	Defective     string            `json:"defective"`
	Criteria      []Criterion       `json:"criteria"`
}

// Fixture es un fixture cargado y validado.
type Fixture struct {
	Dir    string
	Task   Task
	Prompt []byte
}

// Outcome es el resultado de un criterio.
type Outcome string

const (
	Met     Outcome = "cumplido"
	NotMet  Outcome = "incumplido"
	Pending Outcome = "pendiente"
)

// CriterionResult es el resultado de un criterio con su detalle.
type CriterionResult struct {
	ID        string
	Mandatory bool
	Outcome   Outcome
	Detail    string
}

// Status es el veredicto de la corrida.
type Status string

const (
	StatusAccepted Status = "aceptada"
	StatusRejected Status = "rechazada_por_calidad"
	StatusPending  Status = "pendiente_de_revisor"
)

// Verdict es el veredicto con el resultado por criterio.
type Verdict struct {
	Status   Status
	Criteria []CriterionResult
	Failed   []string
	Pending  []string
}

// Decide aplica la regla de aceptacion (design.md, "Tareas y calidad"): un
// criterio obligatorio incumplido da rechazo por calidad, aunque otros sigan
// pendientes; sin incumplidos, un obligatorio pendiente deja el veredicto
// pendiente de revisor; aceptada solo con todos los obligatorios cumplidos.
// Los no obligatorios se reportan y no cambian el estado. Sin criterios
// obligatorios no hay veredicto posible: error.
func Decide(rs []CriterionResult) (Verdict, error) {
	v := Verdict{Criteria: rs}
	mandatory := 0
	for _, r := range rs {
		switch r.Outcome {
		case Met, NotMet, Pending:
		default:
			return Verdict{}, fmt.Errorf("criterio %q: resultado desconocido %q", r.ID, r.Outcome)
		}
		if !r.Mandatory {
			continue
		}
		mandatory++
		switch r.Outcome {
		case NotMet:
			v.Failed = append(v.Failed, r.ID)
		case Pending:
			v.Pending = append(v.Pending, r.ID)
		}
	}
	switch {
	case mandatory == 0:
		return Verdict{}, fmt.Errorf("sin criterios obligatorios no hay veredicto")
	case len(v.Failed) > 0:
		v.Status = StatusRejected
	case len(v.Pending) > 0:
		v.Status = StatusPending
	default:
		v.Status = StatusAccepted
	}
	return v, nil
}

// ReviewCall es la decision del revisor sobre un hallazgo.
type ReviewCall struct {
	Mechanism   string
	Disposition string
}

// Output es lo producido por una corrida.
type Output struct {
	Dir       string
	Answer    []byte
	Decisions map[string]bool
	Review    map[string]ReviewCall
}
