package summary_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/summary"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tarea 2.11: el ejemplo de docs/contrato-de-resumen.md coincide con el resumen
// esperado del fixture 2.1 (fallo al medio).
func TestDoc_SummaryContractExampleMatchesFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "contrato-de-resumen.md"))
	if err != nil {
		t.Fatalf("falta docs/contrato-de-resumen.md: %v", err)
	}
	doc := string(data)
	f := mustFixture(t, "fail-middle")
	want, err := summary.Expected(f, run)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc, want) {
		t.Errorf("el ejemplo del documento no coincide con el resumen esperado del fixture:\n%s", want)
	}
	// Las cifras del umbral del documento son las del catálogo.
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	th := cat.TaskContext.OutputThreshold
	re := regexp.MustCompile(`más de \*\*(\d+) líneas\*\* o de más de \*\*([\d.]+) caracteres\*\*`)
	m := re.FindStringSubmatch(doc)
	if m == nil {
		t.Fatal("el documento no declara el umbral con la forma «más de **N líneas** o de más de **M caracteres**»")
	}
	if m[1] != "300" || strings.ReplaceAll(m[2], ".", "") != "10000" || th.Lines != 300 || th.Characters != 10000 {
		t.Errorf("umbral del documento %s líneas / %s caracteres, catálogo %+v", m[1], m[2], th)
	}
	// Las líneas de forma fija del documento son las que el módulo prescribe.
	module, err := os.ReadFile(filepath.Join("..", "..", "..", "internal", "harness", "config", "assets", "modules", "bounded-command-output.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fixed := range []string{"exit: <code>", "failure: <line as captured>", "status: unverified — <what is missing>", "partial — unverified: <reason>"} {
		if !strings.Contains(doc, fixed) || !strings.Contains(string(module), fixed) {
			t.Errorf("la línea fija %q debe estar en el documento y en el módulo", fixed)
		}
	}
}
