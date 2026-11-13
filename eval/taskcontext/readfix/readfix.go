// Package readfix contiene los fixtures de lectura dirigida y su evaluador
// (change opsx-task-context, D11, D13 y D15): orden de rutas candidatas por
// tipo de coincidencia, lectura por símbolo con la firma del contenedor,
// lecturas completas justificadas o no, relecturas y recorridos archivo por
// archivo. Vive fuera del binario (CHANGES.md:49).
package readfix

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// MatchType es el tipo de coincidencia de una ruta candidata (D13).
type MatchType int

const (
	Text       MatchType = iota // coincidencia literal sin estructura
	Reference                   // el símbolo se usa en ese archivo
	Definition                  // el símbolo se define en ese archivo
)

func (m MatchType) String() string {
	switch m {
	case Definition:
		return "definition"
	case Reference:
		return "reference"
	}
	return "text"
}

// Candidate es una ruta candidata con su tipo de coincidencia y su cantidad.
type Candidate struct {
	Path  string
	Type  MatchType
	Count int
}

// Rank ordena por tipo (una definición pesa más que cualquier cantidad de
// referencias, y una referencia más que cualquier cantidad de texto) y, dentro
// del mismo tipo, por cantidad descendente. Desempata por ruta.
func Rank(cands []Candidate) []Candidate {
	out := append([]Candidate(nil), cands...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Type != b.Type {
			return a.Type > b.Type
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Path < b.Path
	})
	return out
}

// Explain es la explicación de un candidato en una línea.
func Explain(c Candidate) string {
	return fmt.Sprintf("%s — %s ×%d", c.Path, c.Type, c.Count)
}

// DiscardedOpened devuelve los archivos descartados por la orientación que
// igual se abrieron: las tres categorías se obtienen sin leerlos.
func DiscardedOpened(cands []Candidate, chosen, opened []string) []string {
	keep := map[string]bool{}
	for _, p := range chosen {
		keep[p] = true
	}
	candidate := map[string]bool{}
	for _, c := range cands {
		candidate[c.Path] = true
	}
	var out []string
	for _, p := range opened {
		if candidate[p] && !keep[p] {
			out = append(out, p)
		}
	}
	return out
}

// SymbolFixture es una lectura por símbolo anidado sobre un archivo conocido.
type SymbolFixture struct {
	Source             string
	ContainerSignature string
	// Rangos 1-based inclusivos dentro de Source.
	SymbolStart, SymbolEnd       int
	NeighbourStart, NeighbourEnd int
}

func (f SymbolFixture) lines() []string { return strings.Split(f.Source, "\n") }

func (f SymbolFixture) slice(a, b int) string { return strings.Join(f.lines()[a-1:b], "\n") }

// SymbolText es el símbolo completo, de su primera a su última línea.
func (f SymbolFixture) SymbolText() string { return f.slice(f.SymbolStart, f.SymbolEnd) }

// NeighbourText es un símbolo vecino que la lectura no debe traer.
func (f SymbolFixture) NeighbourText() string { return f.slice(f.NeighbourStart, f.NeighbourEnd) }

// NestedSymbol es el fixture: el método `put` dentro de la clase `Store`.
func NestedSymbol() SymbolFixture {
	src := strings.Join([]string{
		"class Store:",
		"    def __init__(self):",
		"        self.items = {}",
		"",
		"    def put(self, key, value):",
		"        self.items[key] = value",
		"        return value",
		"",
		"    def other_method(self):",
		"        return len(self.items)",
	}, "\n")
	return SymbolFixture{
		Source: src, ContainerSignature: "class Store:",
		SymbolStart: 5, SymbolEnd: 7, NeighbourStart: 9, NeighbourEnd: 10,
	}
}

// ExpectedSymbolRead es lo que debe devolver una lectura por símbolo anidado:
// la firma de la declaración que lo contiene y el símbolo completo. Nada más.
func ExpectedSymbolRead(f SymbolFixture) string {
	return f.ContainerSignature + "\n" + f.SymbolText()
}

// CheckSymbolRead devuelve las violaciones de una lectura por símbolo.
func CheckSymbolRead(f SymbolFixture, got string) []string {
	var v []string
	if !strings.HasPrefix(got, f.ContainerSignature+"\n") {
		v = append(v, "falta la firma de la declaración que contiene al símbolo")
	}
	if !strings.Contains(got, f.SymbolText()) {
		v = append(v, "el símbolo no se devolvió completo, de su primera a su última línea")
	}
	if strings.Contains(got, f.NeighbourText()) {
		v = append(v, "trae un símbolo vecino o el resto del cuerpo del contenedor")
	}
	return v
}

// Kind es el tipo de evento del registro de lectura.
type Kind int

const (
	KindRead Kind = iota
	KindWrite
	KindSearch
)

// Event es una entrada del registro de lectura de una tarea.
type Event struct {
	Kind      Kind
	Path      string
	Whole     bool
	Start     int // lectura acotada: líneas 1-based
	End       int
	Symbol    string
	FileLines int
	// Condiciones declaradas por el escenario.
	Rewrite  bool // la tarea va a reescribir el archivo entero (condición 2)
	Artifact bool // es el artefacto que la acción produce o edita (condición 3)
	Empty    bool // una lectura acotada devolvió vacío (condición 4)
	Tool     string
	Input    string
}

// FifthRead es una lectura completa por la quinta condición, con su motivo.
type FifthRead struct{ Path, Reason string }

// Result es lo que cuenta el evaluador.
type Result struct {
	Unjustified       []string
	FifthCondition    []FifthRead
	UnrecordedRereads []string
	Walk              bool
}

var (
	fullReadRe    = regexp.MustCompile(`(?m)^full read: (\S+) — (.+)$`)
	unnecessaryRe = regexp.MustCompile(`(?m)^unnecessary call: (\S+) (.+)$`)
)

// walkMin es la cantidad de lecturas completas seguidas de archivos distintos,
// sin una búsqueda en el medio, que se considera un recorrido para orientarse.
const walkMin = 4

// Evaluate cuenta, sobre el registro y el reporte de una tarea, las lecturas
// completas sin condición, las que entran por la quinta, las relecturas sin
// registrar y el recorrido archivo por archivo. threshold es el umbral de líneas
// del catálogo (D11).
func Evaluate(log []Event, report string, threshold int) Result {
	var res Result
	declared := map[string]string{}
	for _, m := range fullReadRe.FindAllStringSubmatch(report, -1) {
		declared[m[1]] = strings.TrimSpace(m[2])
	}
	recorded := unnecessaryRe.FindAllStringSubmatch(report, -1)

	version := map[string]int{}       // por ruta: cuenta de escrituras
	seen := map[string]int{}          // clave de lectura -> versión en que se leyó
	boundedEmpty := map[string]bool{} // ruta con una lectura acotada vacía
	var streak []string               // lecturas completas seguidas sin búsqueda
	streakSet := map[string]bool{}

	for _, e := range log {
		switch e.Kind {
		case KindWrite:
			version[e.Path]++
			streak, streakSet = nil, map[string]bool{}
		case KindSearch:
			streak, streakSet = nil, map[string]bool{}
		case KindRead:
			key := readKey(e)
			if v, ok := seen[key]; ok && v == version[e.Path] && !hasRecord(recorded, e) {
				res.UnrecordedRereads = append(res.UnrecordedRereads, key)
			}
			seen[key] = version[e.Path]

			if !e.Whole {
				if e.Empty {
					boundedEmpty[e.Path] = true
				}
				continue
			}
			if !streakSet[e.Path] {
				streakSet[e.Path] = true
				streak = append(streak, e.Path)
				if len(streak) >= walkMin {
					res.Walk = true
				}
			}
			switch {
			case e.FileLines <= threshold, e.Rewrite, e.Artifact, boundedEmpty[e.Path]:
				// condiciones 1 a 4
			default:
				if reason, ok := declared[e.Path]; ok && reason != "" {
					res.FifthCondition = append(res.FifthCondition, FifthRead{Path: e.Path, Reason: reason})
				} else {
					res.Unjustified = append(res.Unjustified, e.Path)
				}
			}
		}
	}
	return res
}

func readKey(e Event) string {
	switch {
	case e.Whole:
		return e.Path + "|whole"
	case e.Symbol != "":
		return e.Path + "|sym:" + e.Symbol
	}
	return fmt.Sprintf("%s|%d-%d", e.Path, e.Start, e.End)
}

func hasRecord(recorded [][]string, e Event) bool {
	for _, r := range recorded {
		if strings.Contains(r[2], e.Path) || (e.Input != "" && strings.Contains(r[2], e.Input)) {
			return true
		}
	}
	return false
}

// Scenario es un fixture de lectura con su registro esperado (lo corre la 8.6).
type Scenario struct {
	ID     string
	Log    []Event
	Report string
	Want   Want
}

// Want es el registro esperado de un escenario.
type Want struct {
	Unjustified, Fifth, UnrecordedRereads int
	Walk                                  bool
}

// Scenarios devuelve los fixtures de lectura para el umbral dado.
func Scenarios(th int) []Scenario {
	big := th + 500
	whole := func(p string, lines int) Event {
		return Event{Kind: KindRead, Path: p, Whole: true, FileLines: lines, Tool: "Read", Input: p}
	}
	rng := func(p string, a, b int) Event {
		return Event{Kind: KindRead, Path: p, Start: a, End: b, FileLines: big, Tool: "Read", Input: fmt.Sprintf("%s:%d-%d", p, a, b)}
	}
	return []Scenario{
		{ID: "full-read-unjustified", Log: []Event{whole("internal/x/big.go", big)}, Want: Want{Unjustified: 1}},
		{ID: "full-read-under-threshold", Log: []Event{whole("internal/x/small.go", th)}, Want: Want{}},
		{ID: "full-read-rewrite", Log: []Event{func() Event { e := whole("internal/x/big.go", big); e.Rewrite = true; return e }()}, Want: Want{}},
		{ID: "full-read-artifact", Log: []Event{func() Event { e := whole("openspec/changes/c/tasks.md", big); e.Artifact = true; return e }()}, Want: Want{}},
		{ID: "full-read-after-empty-bounded", Log: []Event{
			func() Event { e := rng("internal/x/big.go", 1, 40); e.Empty = true; return e }(),
			whole("internal/x/big.go", big),
		}, Want: Want{}},
		{ID: "fifth-condition-with-reason", Log: []Event{whole("internal/tui/model.go", big)},
			Report: "full read: internal/tui/model.go — the update loop is interleaved across the whole file\n", Want: Want{Fifth: 1}},
		{ID: "fifth-condition-without-reason", Log: []Event{whole("internal/tui/model.go", big)},
			Report: "full read: internal/tui/model.go —\n", Want: Want{Unjustified: 1}},
		{ID: "reread-without-write-recorded", Log: []Event{rng("internal/x/a.go", 10, 40), rng("internal/x/a.go", 10, 40)},
			Report: "unnecessary call: Read internal/x/a.go:10-40\n", Want: Want{}},
		{ID: "reread-without-write-unrecorded", Log: []Event{rng("internal/x/a.go", 10, 40), rng("internal/x/a.go", 10, 40)},
			Want: Want{UnrecordedRereads: 1}},
		{ID: "reread-after-write", Log: []Event{rng("internal/x/a.go", 10, 40), {Kind: KindWrite, Path: "internal/x/a.go"}, rng("internal/x/a.go", 10, 40)},
			Want: Want{}},
		{ID: "walk-file-by-file", Log: []Event{
			whole("internal/a/a.go", 80), whole("internal/b/b.go", 90), whole("internal/c/c.go", 70), whole("internal/d/d.go", 60),
		}, Want: Want{Walk: true}},
		{ID: "orientation-by-search", Log: []Event{
			whole("internal/a/a.go", 80), whole("internal/b/b.go", 90), {Kind: KindSearch, Input: "rg Resolve"},
			whole("internal/c/c.go", 70), whole("internal/d/d.go", 60),
		}, Want: Want{}},
	}
}
