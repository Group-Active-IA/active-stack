package readfix_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/readfix"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tareas 4.5 y 4.6 (opsx-task-context, D11, D13 y D15): fixtures de lectura y
// el evaluador que cuenta lecturas completas no justificadas, la quinta
// condición, relecturas sin registrar y recorridos archivo por archivo.

func threshold(t *testing.T) int {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cat.TaskContext.ReadThreshold.Value
}

// 4.5: una definición pesa más que cualquier cantidad de referencias; la
// cantidad solo ordena dentro del mismo tipo.
func TestRank_TypeBeforeCount(t *testing.T) {
	cands := []readfix.Candidate{
		{Path: "internal/a/refs.go", Type: readfix.Reference, Count: 40},
		{Path: "internal/b/grep.go", Type: readfix.Text, Count: 200},
		{Path: "internal/c/def.go", Type: readfix.Definition, Count: 1},
		{Path: "internal/d/refs.go", Type: readfix.Reference, Count: 3},
		{Path: "internal/e/def2.go", Type: readfix.Definition, Count: 2},
	}
	got := readfix.Rank(cands)
	var paths []string
	for _, c := range got {
		paths = append(paths, c.Path)
	}
	want := "internal/e/def2.go,internal/c/def.go,internal/a/refs.go,internal/d/refs.go,internal/b/grep.go"
	if strings.Join(paths, ",") != want {
		t.Errorf("orden = %v\nwant   %v", paths, want)
	}
	// Triangulación: el orden no depende del orden de entrada.
	rev := make([]readfix.Candidate, len(cands))
	for i, c := range cands {
		rev[len(cands)-1-i] = c
	}
	var paths2 []string
	for _, c := range readfix.Rank(rev) {
		paths2 = append(paths2, c.Path)
	}
	if strings.Join(paths2, ",") != want {
		t.Errorf("el orden cambió con la entrada invertida: %v", paths2)
	}
	// Cada candidato se explica en una línea con su tipo y su cantidad.
	line := readfix.Explain(got[0])
	if !strings.Contains(line, "internal/e/def2.go") || !strings.Contains(line, "definition") || !strings.Contains(line, "2") || strings.Contains(line, "\n") {
		t.Errorf("explicación = %q", line)
	}
}

func TestDiscardedFilesAreNotOpened(t *testing.T) {
	cands := readfix.Rank([]readfix.Candidate{
		{Path: "a.go", Type: readfix.Definition, Count: 1},
		{Path: "b.go", Type: readfix.Reference, Count: 9},
		{Path: "c.go", Type: readfix.Text, Count: 4},
	})
	opened := readfix.DiscardedOpened(cands, []string{"a.go"}, []string{"a.go", "c.go"})
	if len(opened) != 1 || opened[0] != "c.go" {
		t.Errorf("archivos descartados que se abrieron = %v, want [c.go]", opened)
	}
	if got := readfix.DiscardedOpened(cands, []string{"a.go"}, []string{"a.go"}); len(got) != 0 {
		t.Errorf("sin lecturas de descartados: %v", got)
	}
}

// 4.6: lectura por símbolo anidado con la firma de la declaración que lo contiene.
func TestSymbolRead_NestedIncludesContainerSignature(t *testing.T) {
	f := readfix.NestedSymbol()
	want := readfix.ExpectedSymbolRead(f)
	if !strings.HasPrefix(want, f.ContainerSignature+"\n") {
		t.Fatalf("la lectura esperada debe empezar con la firma del contenedor:\n%s", want)
	}
	if strings.Contains(want, "def other_method") {
		t.Errorf("la lectura no debe traer los símbolos vecinos:\n%s", want)
	}
	if v := readfix.CheckSymbolRead(f, want); len(v) != 0 {
		t.Errorf("la lectura conforme da violaciones: %v", v)
	}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"sin la firma del contenedor", f.SymbolText(), "firma"},
		{"solo la primera línea del símbolo", f.ContainerSignature + "\n" + strings.Split(f.SymbolText(), "\n")[0], "completo"},
		{"con el resto del cuerpo del contenedor", want + "\n" + f.NeighbourText(), "vecino"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := readfix.CheckSymbolRead(f, tt.got)
			if len(v) == 0 || !strings.Contains(strings.Join(v, " "), tt.want) {
				t.Errorf("violaciones = %v, want mención de %q", v, tt.want)
			}
		})
	}
}

func TestEvaluate_Scenarios(t *testing.T) {
	th := threshold(t)
	for _, sc := range readfix.Scenarios(th) {
		t.Run(sc.ID, func(t *testing.T) {
			got := readfix.Evaluate(sc.Log, sc.Report, th)
			if len(got.Unjustified) != sc.Want.Unjustified {
				t.Errorf("lecturas completas no justificadas = %v, want %d", got.Unjustified, sc.Want.Unjustified)
			}
			if len(got.FifthCondition) != sc.Want.Fifth {
				t.Errorf("quinta condición = %v, want %d", got.FifthCondition, sc.Want.Fifth)
			}
			if len(got.UnrecordedRereads) != sc.Want.UnrecordedRereads {
				t.Errorf("relecturas sin registrar = %v, want %d", got.UnrecordedRereads, sc.Want.UnrecordedRereads)
			}
			if got.Walk != sc.Want.Walk {
				t.Errorf("recorrido archivo por archivo = %v, want %v", got.Walk, sc.Want.Walk)
			}
		})
	}
}

func TestScenarios_CoverTheSpecCases(t *testing.T) {
	have := map[string]bool{}
	for _, sc := range readfix.Scenarios(300) {
		have[sc.ID] = true
	}
	for _, id := range []string{
		"full-read-unjustified", "full-read-under-threshold", "full-read-rewrite", "full-read-artifact", "full-read-after-empty-bounded",
		"fifth-condition-with-reason", "fifth-condition-without-reason",
		"reread-without-write-recorded", "reread-without-write-unrecorded", "reread-after-write",
		"walk-file-by-file", "orientation-by-search",
	} {
		if !have[id] {
			t.Errorf("falta el escenario %q (specs/directed-code-reading)", id)
		}
	}
}

// La cuenta de la quinta condición separa los motivos: insumo para recalibrar D11.
func TestEvaluate_FifthConditionCountsReasonsSeparately(t *testing.T) {
	log := []readfix.Event{
		{Kind: readfix.KindRead, Path: "internal/tui/model.go", Whole: true, FileLines: 809},
		{Kind: readfix.KindRead, Path: "internal/big/generated.go", Whole: true, FileLines: 2400},
	}
	report := "full read: internal/tui/model.go — the update loop is interleaved across the file\n" +
		"full read: internal/big/generated.go — no range covers the registry init\n"
	got := readfix.Evaluate(log, report, 300)
	if len(got.Unjustified) != 0 || len(got.FifthCondition) != 2 {
		t.Fatalf("resultado = %+v", got)
	}
	if got.FifthCondition[0].Reason == got.FifthCondition[1].Reason || got.FifthCondition[0].Reason == "" {
		t.Errorf("los motivos deben quedar separados y no vacíos: %+v", got.FifthCondition)
	}
}

// 4.7: cada ejemplo documentado corresponde a un fixture y cada escenario de
// lectura queda documentado.
func TestDoc_DirectedReadingExamplesMatchFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "lectura-dirigida.md"))
	if err != nil {
		t.Fatalf("falta docs/lectura-dirigida.md: %v", err)
	}
	doc := string(data)
	re := regexp.MustCompile("fixture: ([a-z0-9-]+)")
	documented := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(doc, -1) {
		documented[m[1]] = true
	}
	known := map[string]bool{"nested-symbol": true}
	for _, sc := range readfix.Scenarios(300) {
		known[sc.ID] = true
		if !documented[sc.ID] {
			t.Errorf("el escenario %q no está documentado", sc.ID)
		}
	}
	for id := range documented {
		if !known[id] {
			t.Errorf("el documento cita el fixture %q, que no existe", id)
		}
	}
	// El umbral documentado es el del catálogo.
	if !strings.Contains(doc, "valor 300") || threshold(t) != 300 {
		t.Errorf("el documento debe declarar el umbral del catálogo (%d)", threshold(t))
	}
	// Las cinco condiciones del documento son las del catálogo.
	cat, _ := catalog.Load()
	if len(cat.TaskContext.FullReadConditions) != 5 {
		t.Fatalf("el catálogo declara %d condiciones", len(cat.TaskContext.FullReadConditions))
	}
	for i := 1; i <= 5; i++ {
		if !strings.Contains(doc, "\n"+string(rune('0'+i))+". ") {
			t.Errorf("falta la condición %d en el documento", i)
		}
	}
}
