// Package replyfix contiene los fixtures de respuesta del agente y la rúbrica
// que los evalúa (change opsx-task-context, D16): el resultado va primero con
// su procedencia, el detalle se produce a pedido, lo que ya está en un
// artefacto se señala y una revisión sin hallazgos ocupa una línea. Vive fuera
// del binario (CHANGES.md:49).
package replyfix

import (
	"regexp"
	"strings"
)

// Kind es el tipo de caso.
type Kind int

const (
	SingleFinding Kind = iota
	MultipleFindings
	NoFinding
	InArtifact
)

// Case es un caso de respuesta con la respuesta conforme y, para el par de 7.1,
// la misma con el recorrido primero.
type Case struct {
	ID   string
	Kind Kind
	// ResultFirst es la respuesta conforme.
	ResultFirst string
	// PathFirst describe el mismo hallazgo con el recorrido primero (solo el par).
	PathFirst string
	// ArtifactText es el contenido que ya vive en un artefacto (caso InArtifact).
	ArtifactText string
	// Artifact es la ruta del artefacto que debe señalarse.
	Artifact string
}

// Cases devuelve los fixtures de respuesta. El primero es el par de 7.1.
func Cases() []Case {
	return []Case{
		{
			ID: "single-finding", Kind: SingleFinding,
			ResultFirst: "Compose rejects an unknown toggle with an error: internal/harness/config/compose.go:35\n" +
				"Next: add the toggle to knownToggles before wiring a new module.\n",
			PathFirst: "I started by listing the config package and opened compose.go. Then I read the first 60 lines and followed the toggle loop.\n" +
				"After that I checked how the error is built and where the known toggles are declared.\n" +
				"Compose rejects an unknown toggle with an error: internal/harness/config/compose.go:35\n" +
				"Next: add the toggle to knownToggles before wiring a new module.\n",
		},
		{
			ID: "multiple-findings", Kind: MultipleFindings,
			ResultFirst: "Three findings:\n" +
				"- toggle map is duplicated: internal/harness/config/compose.go:12\n" +
				"- error text is untranslated: internal/harness/config/compose.go:35\n" +
				"- no test for the empty toggle list: internal/harness/config/compose_test.go:48\n" +
				"Next: fix the duplicated map first; the other two depend on it.\n",
		},
		{
			ID: "no-finding", Kind: NoFinding,
			ResultFirst: "No unchecked toggle found: searched knownToggles usages in internal/harness/config/.\n",
		},
		{
			ID: "already-in-artifact", Kind: InArtifact,
			ArtifactText: "El único número del contrato es cuándo se resume, no cuánto: una salida de más de 300 líneas o de más de 10.000 caracteres, lo que ocurra primero, se entrega como resumen más referencia.",
			Artifact:     "openspec/changes/archive/2026-11-13-opsx-task-context/design.md",
			ResultFirst: "The summary threshold is already decided: openspec/changes/archive/2026-11-13-opsx-task-context/design.md#D3\n" +
				"Next: read section D3 instead of restating it here.\n",
		},
	}
}

// Result es el veredicto de la rúbrica.
type Result struct {
	Pass       bool
	Violations []string
}

var (
	provenanceRe = regexp.MustCompile(`[\w./-]+(?::\d+|#[\w.]+)`)
	narrativeRe  = regexp.MustCompile(`(?i)^\s*(first|then|next,|after that|finally|i (started|looked|read|searched|opened|checked|listed|followed|went))\b`)
	listItemRe   = regexp.MustCompile(`^\s*[-*]\s+`)
	// fixedLineRes son las formas fijas de los otros módulos, que el módulo de
	// reporte manda conservar: no son relato (corridas 8.10 del 13/11).
	fixedLineRes = []*regexp.Regexp{
		regexp.MustCompile(`^full read: \S+ — \S.*$`),
		regexp.MustCompile(`^unnecessary call: \S+ \S.*$`),
		regexp.MustCompile(`^unavailable: \S+ — uncovered: .+; alternative: .+$`),
		regexp.MustCompile(`^out-of-class use: \S+ in \S+$`),
		regexp.MustCompile(`^ref: cmd=.+ run=\S+ path=.+ sha256:[0-9a-f]+ expires=\d{4}-\d{2}-\d{2}$`),
		regexp.MustCompile(`^partial — unverified: \S.*$`),
	}
)

func isFixedLine(l string) bool {
	l = strings.TrimSpace(strings.Trim(strings.TrimSpace(l), "`"))
	for _, re := range fixedLineRes {
		if re.MatchString(l) {
			return true
		}
	}
	return false
}

// Evaluate aplica la rúbrica: primera línea con el resultado y su procedencia,
// segunda línea con la consecuencia o el próximo paso, sin relato, lista con
// procedencia antes de cualquier explicación, señalar en lugar de reproducir lo
// que ya está en un artefacto, y una línea para una revisión sin hallazgos.
func Evaluate(c Case, reply string) Result { return evaluate(c, reply, false) }

// EvaluateWithDetailRequested acepta el desarrollo posterior al resultado
// cuando el usuario lo pidió (el detalle se produce a pedido).
func EvaluateWithDetailRequested(c Case, reply string) Result { return evaluate(c, reply, true) }

func evaluate(c Case, reply string, detailRequested bool) Result {
	var v []string
	lines := nonEmpty(reply)
	if len(lines) == 0 {
		return Result{Violations: []string{"respuesta vacía"}}
	}

	switch c.Kind {
	case NoFinding:
		if len(lines) != 1 {
			v = append(v, "una revisión sin hallazgos ocupa una línea: dice que no hay y qué se revisó")
		}
		if narrativeRe.MatchString(lines[0]) || !strings.Contains(strings.ToLower(lines[0]), "no ") {
			v = append(v, "la primera línea debe decir que no hay hallazgo y qué se revisó, sin narrar la búsqueda")
		}
	case MultipleFindings:
		items := 0
		for i, l := range lines {
			if listItemRe.MatchString(l) {
				items++
				if !provenanceRe.MatchString(l) {
					v = append(v, "cada ítem de la lista lleva procedencia: falta en «"+strings.TrimSpace(l)+"»")
				}
				continue
			}
			// Antes del primer ítem solo puede ir el encabezado corto de la lista.
			if items == 0 && i > 0 {
				v = append(v, "la lista con procedencia va antes de cualquier explicación")
			}
			if items == 0 && i == 0 && len(strings.Fields(l)) > 6 {
				v = append(v, "la lista con procedencia va antes de cualquier explicación")
			}
			if items > 0 && !strings.HasPrefix(strings.ToLower(l), "next") && !detailRequested {
				v = append(v, "después de la lista solo va la consecuencia o el próximo paso")
			}
		}
		if items < 2 {
			v = append(v, "varios hallazgos: la respuesta debe traer una lista")
		}
	default:
		first := lines[0]
		if narrativeRe.MatchString(first) {
			v = append(v, "la primera línea no es el resultado: es el relato del recorrido")
		} else if !provenanceRe.MatchString(first) {
			v = append(v, "la primera línea debe llevar el resultado con su procedencia (path:line o path#symbol)")
		}
		if len(lines) < 2 {
			v = append(v, "falta la segunda línea con la consecuencia o el próximo paso")
		} else if narrativeRe.MatchString(lines[1]) {
			v = append(v, "la segunda línea debe ser la consecuencia o el próximo paso, no relato")
		}
		if !detailRequested {
			for _, l := range lines[min(2, len(lines)):] {
				if isFixedLine(l) {
					continue
				}
				v = append(v, "no se cuenta cómo se llegó al resultado si no se pidió: sobra el relato «"+strings.TrimSpace(l)+"»")
			}
		}
		if c.Kind == InArtifact {
			if strings.Contains(reply, strings.TrimSpace(c.ArtifactText)) {
				v = append(v, "el contenido ya está en el artefacto: hay que señalarlo, no copiarlo")
			}
			if !strings.Contains(reply, c.Artifact) {
				v = append(v, "falta nombrar el artefacto y su sección")
			}
		}
	}
	// Relato en cualquier posición cuando no se pidió detalle.
	if !detailRequested && c.Kind != SingleFinding && c.Kind != InArtifact {
		for _, l := range lines {
			if narrativeRe.MatchString(l) {
				v = append(v, "relato del recorrido: «"+strings.TrimSpace(l)+"»")
			}
		}
	}
	return Result{Pass: len(v) == 0, Violations: v}
}

func nonEmpty(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
