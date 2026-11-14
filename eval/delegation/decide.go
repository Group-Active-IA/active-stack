package delegation

import (
	"path"
	"strings"
)

// Modalidades de la decisión.
const (
	Direct     = "directo"
	Delegated  = "delegado"
	Stopped    = "detenido"
	Serialized = "serializado"
)

// HotFiles son los archivos que nunca van a dos workers concurrentes (D5). Es la lista del módulo
// instalado; TestHotFiles_MatchTheModule impide que diverjan.
var HotFiles = []string{
	"go.mod",
	"go.sum",
	"internal/catalog/harnesses.yaml",
	"internal/harness/config/assets/*/sdd-orchestrator.md",
	"CLAUDE.md",
	"AGENTS.md",
}

// MaxWriters es el tope de workers con escritura en paralelo (D5).
const MaxWriters = 2

// Umbral de salida verbosa de 05 D3 (módulo bounded-command-output).
const (
	VerboseLines = 300
	VerboseChars = 10000
)

// Task describe una acción a decidir con la regla de assets/opsx/delegation-rule.md.
type Task struct {
	Action            string     `json:"accion"`
	IndependentReview bool       `json:"revision_independiente"`
	DelegationEnabled bool       `json:"delegation_activo"`
	ModuleAvailable   bool       `json:"modulo_disponible"`
	WriteSet          []string   `json:"escritura"`
	OpenWriteSets     [][]string `json:"encargos_abiertos"` // conjuntos de escritura en curso
	Short             bool       `json:"breve"`
	DependsOnLive     bool       `json:"depende_del_contexto_vivo"`
	ReadsFullResult   bool       `json:"relee_el_resultado"`
	OutputLines       int        `json:"lineas_de_salida"`
	OutputChars       int        `json:"caracteres_de_salida"`
	ConclusionOnly    bool       `json:"solo_conclusion"`
	Parallelizable    bool       `json:"paralelizable"`
}

// Decision es el registro `{action, step, signal, direct|delegated|stopped}` de la regla, más lo
// que falta cuando la acción se detiene.
type Decision struct {
	Action  string `json:"accion"`
	Step    int    `json:"paso"`
	Signal  string `json:"senal"`
	Mode    string `json:"modalidad"`
	Missing string `json:"falta,omitempty"`
}

// Decide aplica la regla en orden; el primer paso que aplica cierra la decisión.
func Decide(t Task) Decision {
	d := Decision{Action: t.Action}
	set := func(step int, signal, mode string) Decision {
		d.Step, d.Signal, d.Mode = step, signal, mode
		return d
	}

	// 1. Independencia obligatoria: nada de lo que sigue la anula.
	if t.IndependentReview {
		switch {
		case !t.DelegationEnabled:
			d.Missing = "delegation"
			return set(1, "independencia obligatoria", Stopped)
		case !t.ModuleAvailable:
			d.Missing = "modules/delegation.md"
			return set(1, "independencia obligatoria", Stopped)
		}
		return set(1, "independencia obligatoria", Delegated)
	}

	// Con delegation apagado la regla instalada (delegation-rule-disabled.md) sólo tiene el paso 1:
	// todo lo demás se resuelve directo.
	if !t.DelegationEnabled {
		return set(5, "delegation desactivado", Direct)
	}

	// 2. Conflicto de archivos con un encargo abierto, archivos calientes repartidos y tope de
	// workers con escritura (D5): se serializa.
	if len(t.WriteSet) > 0 {
		for _, open := range t.OpenWriteSets {
			if setsOverlap(t.WriteSet, open) || hotConflict(t.WriteSet, open) {
				return set(2, "conflicto de archivos", Serialized)
			}
		}
		if writers(t.OpenWriteSets) >= MaxWriters {
			return set(2, "tope de dos workers con escritura", Serialized)
		}
	}

	// 3. Contraseñales: directo.
	switch {
	case t.Short:
		return set(3, "tarea breve", Direct)
	case t.DependsOnLive:
		return set(3, "depende del contexto vivo", Direct)
	case t.ReadsFullResult:
		return set(3, "relectura completa", Direct)
	}

	// 4. Señales: delegar. El umbral de salida verbosa es el de 05 D3: más de 300 líneas o más de
	// 10.000 caracteres, lo que ocurra primero.
	switch {
	case t.OutputLines > VerboseLines || t.OutputChars > VerboseChars:
		return set(4, "salida verbosa", Delegated)
	case t.ConclusionOnly:
		return set(4, "exploración con conclusión", Delegated)
	case t.Parallelizable:
		return set(4, "trabajo paralelizable", Delegated)
	}

	// 5. Por defecto, directo.
	return set(5, "por defecto", Direct)
}

func writers(open [][]string) int {
	n := 0
	for _, ws := range open {
		if len(ws) > 0 {
			n++
		}
	}
	return n
}

func setsOverlap(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if overlap(x, y) {
				return true
			}
		}
	}
	return false
}

// hotConflict: los dos conjuntos tocan el mismo archivo caliente (o el mismo patrón, como
// internal/harness/config/assets/*/sdd-orchestrator.md), aunque no compartan una ruta exacta.
func hotConflict(a, b []string) bool {
	for _, h := range HotFiles {
		if setsOverlap(a, []string{h}) && setsOverlap(b, []string{h}) {
			return true
		}
	}
	return false
}

// overlap dice si dos entradas de conjunto de escritura (o una entrada y una ruta) pueden nombrar
// el mismo archivo. "dir/**" cubre todo lo que está debajo de dir/; el resto usa path.Match.
func overlap(a, b string) bool {
	if a == b {
		return true
	}
	pa, aw := strings.CutSuffix(a, "**")
	pb, bw := strings.CutSuffix(b, "**")
	switch {
	case aw && bw:
		return strings.HasPrefix(pa, pb) || strings.HasPrefix(pb, pa)
	case aw:
		return strings.HasPrefix(b, pa)
	case bw:
		return strings.HasPrefix(a, pb)
	}
	m1, _ := path.Match(a, b)
	m2, _ := path.Match(b, a)
	return m1 || m2
}
