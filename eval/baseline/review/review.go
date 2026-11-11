// Package review implementa la regla de emparejamiento de la revision de un
// diff sembrado (change opsx-baseline, tarea 1.4; design.md, "Rubrica de
// revision de un diff sembrado"). 10 la reutiliza.
package review

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Defect es un defecto sembrado de la clave de referencia.
type Defect struct {
	ID        string
	File      string
	StartLine int
	EndLine   int
	Mechanism string
}

// Disposition es la decision del revisor sobre un hallazgo.
type Disposition string

const (
	// DispositionUndecided: el revisor todavia no decidio.
	DispositionUndecided Disposition = ""
	// DispositionRejected: el revisor rechaza el hallazgo.
	DispositionRejected Disposition = "rejected"
	// DispositionRealDefect: el revisor lo considera un defecto real no sembrado.
	DispositionRealDefect Disposition = "real_defect"
)

// Finding es un hallazgo de la revision. Mechanism es el mecanismo de falla
// que decidio el revisor; Evidence la evidencia que registro.
type Finding struct {
	ID          string
	File        string
	StartLine   int
	EndLine     int
	Mechanism   string
	Evidence    string
	Disposition Disposition
}

// Class es la clasificacion de un hallazgo.
type Class string

const (
	TruePositive  Class = "verdadero_positivo"
	Duplicate     Class = "duplicado"
	FalsePositive Class = "falso_positivo"
	// NewFinding es un hallazgo nuevo derivado a adjudicacion separada.
	NewFinding Class = "hallazgo_nuevo"
	// PendingReview es un hallazgo sin emparejar sin decision del revisor.
	PendingReview Class = "pendiente_de_revisor"
)

// Classified es la clasificacion de un hallazgo.
type Classified struct {
	FindingID string
	Class     Class
	// DefectID es el defecto emparejado (verdadero positivo) o ya cubierto (duplicado).
	DefectID string
	// Candidates lista todos los defectos con los que el hallazgo coincide.
	Candidates []string
	Evidence   string
}

// Counts son los conteos del caso. NewFindings y Pending no alteran
// TruePositives, FalsePositives ni Omissions.
type Counts struct {
	TruePositives  int
	Duplicates     int
	FalsePositives int
	Omissions      int
	NewFindings    int
	Pending        int
}

// Result es el resultado de emparejar hallazgos con la clave.
type Result struct {
	Findings  []Classified
	Omissions []string
	Counts    Counts
}

// Match clasifica los hallazgos contra la clave y reporta las omisiones.
//
// Emparejamiento (corregido en apply-2, F1 de verify-1-breaker):
//   - Uno a uno y maximo: se busca la mayor cantidad de pares hallazgo-defecto
//     (bipartito por caminos de aumento), asi que el resultado no depende del
//     orden de la clave.
//   - Desempate estable: los hallazgos se procesan en el orden de la lista;
//     cada uno toma primero el candidato libre de menor ID y solo si no hay
//     libre desplaza a otro hallazgo por un camino de aumento.
//   - Un hallazgo con candidatos que queda sin par es duplicado del primer
//     candidato por ID; no suma.
//   - El mecanismo coincide si es igual sin distinguir mayusculas ni espacios
//     de los extremos; un mecanismo vacio no empareja.
func Match(key []Defect, findings []Finding) (Result, error) {
	ids := map[string]bool{}
	for _, d := range key {
		if d.ID == "" {
			return Result{}, fmt.Errorf("clave: defecto sin id")
		}
		if ids[d.ID] {
			return Result{}, fmt.Errorf("clave: id de defecto duplicado %q", d.ID)
		}
		ids[d.ID] = true
		if d.StartLine < 1 || d.EndLine < d.StartLine {
			return Result{}, fmt.Errorf("clave: rango inválido %d-%d en el defecto %q", d.StartLine, d.EndLine, d.ID)
		}
		// DEUDA 6 (S2): un defecto sin archivo o sin mecanismo seria una
		// omision garantizada.
		if strings.TrimSpace(d.File) == "" {
			return Result{}, fmt.Errorf("clave: archivo vacío en el defecto %q", d.ID)
		}
		if normMech(d.Mechanism) == "" {
			return Result{}, fmt.Errorf("clave: mecanismo vacío en el defecto %q", d.ID)
		}
	}
	fids := map[string]bool{}
	for _, f := range findings {
		if f.ID == "" {
			return Result{}, fmt.Errorf("hallazgo sin id")
		}
		if fids[f.ID] {
			return Result{}, fmt.Errorf("id de hallazgo duplicado %q", f.ID)
		}
		fids[f.ID] = true
		if f.StartLine < 1 || f.EndLine < f.StartLine {
			return Result{}, fmt.Errorf("rango inválido %d-%d en el hallazgo %q", f.StartLine, f.EndLine, f.ID)
		}
		// DEUDA 6 (S4): una disposicion desconocida es error, no pendiente.
		switch f.Disposition {
		case DispositionUndecided, DispositionRejected, DispositionRealDefect:
		default:
			return Result{}, fmt.Errorf("disposición desconocida %q en el hallazgo %q", f.Disposition, f.ID)
		}
	}

	byID := make([]Defect, len(key))
	copy(byID, key)
	sort.Slice(byID, func(i, j int) bool { return byID[i].ID < byID[j].ID })

	cands := make([][]string, len(findings))
	for i, f := range findings {
		for _, d := range byID {
			if matches(d, f) {
				cands[i] = append(cands[i], d.ID)
			}
		}
	}
	owner := maxMatching(cands)
	assigned := map[int]string{}
	for defectID, fi := range owner {
		assigned[fi] = defectID
	}

	var res Result
	for i, f := range findings {
		c := Classified{FindingID: f.ID, Candidates: cands[i], Evidence: f.Evidence}
		switch {
		case len(cands[i]) == 0:
			switch f.Disposition {
			case DispositionRejected:
				c.Class = FalsePositive
				res.Counts.FalsePositives++
			case DispositionRealDefect:
				c.Class = NewFinding
				res.Counts.NewFindings++
			default:
				c.Class = PendingReview
				res.Counts.Pending++
			}
		case assigned[i] != "":
			c.Class, c.DefectID = TruePositive, assigned[i]
			res.Counts.TruePositives++
		default:
			c.Class, c.DefectID = Duplicate, cands[i][0]
			res.Counts.Duplicates++
		}
		res.Findings = append(res.Findings, c)
	}
	for _, d := range key {
		if _, ok := owner[d.ID]; !ok {
			res.Omissions = append(res.Omissions, d.ID)
		}
	}
	res.Counts.Omissions = len(res.Omissions)
	return res, nil
}

// maxMatching devuelve defecto -> indice de hallazgo para un emparejamiento
// maximo (algoritmo de Kuhn). cands[i] son los defectos candidatos del
// hallazgo i, ya ordenados por ID.
func maxMatching(cands [][]string) map[string]int {
	owner := map[string]int{}
	var try func(i int, seen map[string]bool) bool
	try = func(i int, seen map[string]bool) bool {
		// Primero un candidato libre (desempate estable: no desplaza pares ya
		// hechos si no hace falta); luego caminos de aumento.
		for _, d := range cands[i] {
			if _, taken := owner[d]; !taken && !seen[d] {
				seen[d] = true
				owner[d] = i
				return true
			}
		}
		for _, d := range cands[i] {
			if seen[d] {
				continue
			}
			seen[d] = true
			prev, taken := owner[d]
			if !taken || try(prev, seen) {
				owner[d] = i
				return true
			}
		}
		return false
	}
	for i := range cands {
		try(i, map[string]bool{})
	}
	return owner
}
func matches(d Defect, f Finding) bool {
	return normPath(d.File) == normPath(f.File) &&
		f.StartLine <= d.EndLine && d.StartLine <= f.EndLine &&
		normMech(f.Mechanism) != "" && normMech(f.Mechanism) == normMech(d.Mechanism)
}

func normPath(p string) string {
	return path.Clean(strings.ReplaceAll(p, `\`, "/"))
}

func normMech(m string) string {
	return strings.ToLower(strings.TrimSpace(m))
}
