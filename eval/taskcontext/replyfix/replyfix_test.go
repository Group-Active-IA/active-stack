package replyfix_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/replyfix"
)

// Tareas 7.1, 7.2 y 7.3 (opsx-task-context, D16): fixtures de respuesta, la
// rúbrica y el contrato documentado. La conducta real se evalúa en la 8.10.

var provenanceRe = regexp.MustCompile(`[\w./-]+(?::\d+|#[\w.]+)`)

// 7.1: el par describe el mismo hallazgo con la misma procedencia.
func TestPairs_DescribeTheSameFindingWithTheSameProvenance(t *testing.T) {
	for _, c := range replyfix.Cases() {
		if c.PathFirst == "" {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			a := provenanceRe.FindAllString(c.ResultFirst, -1)
			b := provenanceRe.FindAllString(c.PathFirst, -1)
			if len(a) == 0 {
				t.Fatalf("la respuesta resultado-primero no trae procedencia:\n%s", c.ResultFirst)
			}
			if strings.Join(a, ",") != strings.Join(b, ",") {
				t.Errorf("la procedencia difiere entre el par: %v vs %v", a, b)
			}
			// El par comparte el contenido del hallazgo: toda línea de resultado de la primera está en la segunda.
			for _, l := range strings.Split(strings.TrimSpace(c.ResultFirst), "\n") {
				if !strings.Contains(c.PathFirst, strings.TrimSpace(l)) {
					t.Errorf("la línea %q de resultado-primero no figura en recorrido-primero", l)
				}
			}
			if len(c.PathFirst) <= len(c.ResultFirst) {
				t.Error("recorrido-primero debe ser más larga: lleva el relato")
			}
		})
	}
}

func TestCases_CoverTheSpec(t *testing.T) {
	have := map[replyfix.Kind]bool{}
	for _, c := range replyfix.Cases() {
		have[c.Kind] = true
	}
	for _, k := range []replyfix.Kind{replyfix.SingleFinding, replyfix.MultipleFindings, replyfix.NoFinding, replyfix.InArtifact} {
		if !have[k] {
			t.Errorf("falta un caso de tipo %v (specs/concise-agent-reporting)", k)
		}
	}
}

// 7.2: la rúbrica distingue resultado primero de recorrido primero.
func TestRubric_DistinguishesTheFixturePair(t *testing.T) {
	for _, c := range replyfix.Cases() {
		t.Run(c.ID, func(t *testing.T) {
			good := replyfix.Evaluate(c, c.ResultFirst)
			if !good.Pass {
				t.Errorf("la respuesta conforme no pasa: %v\n%s", good.Violations, c.ResultFirst)
			}
			if c.PathFirst == "" {
				return
			}
			bad := replyfix.Evaluate(c, c.PathFirst)
			if bad.Pass {
				t.Errorf("recorrido-primero no puede pasar la rúbrica:\n%s", c.PathFirst)
			}
			if !strings.Contains(strings.Join(bad.Violations, " | "), "primera línea") {
				t.Errorf("la violación debe decir que la primera línea no es el resultado: %v", bad.Violations)
			}
		})
	}
}

func TestRubric_EachRuleFailsOnItsOwn(t *testing.T) {
	cases := map[string]replyfix.Case{}
	for _, c := range replyfix.Cases() {
		cases[c.ID] = c
	}
	single := cases["single-finding"]
	multi := cases["multiple-findings"]
	none := cases["no-finding"]
	art := cases["already-in-artifact"]
	tests := []struct {
		name  string
		c     replyfix.Case
		reply string
		want  string
	}{
		{"sin procedencia en la primera línea", single, "Compose rejects an unknown toggle\nNext: add it to knownToggles.\n", "procedencia"},
		{"sin segunda línea de consecuencia", single, strings.Split(single.ResultFirst, "\n")[0] + "\n", "segunda línea"},
		{"desarrollo no pedido después del resultado", single, single.ResultFirst + "I started by listing the package and then read compose.go in full.\n", "relato"},
		{"lista sin procedencia", multi, "Three findings:\n- toggle map is duplicated\n- error text is untranslated\n- missing test\nNext: fix the first.\n", "procedencia"},
		{"explicación antes de la lista", multi, "These issues come from the way the package grew over several rounds.\n" + multi.ResultFirst, "lista"},
		{"copia del artefacto en lugar de señalarlo", art, art.ArtifactText + "\nNext: apply it.\n", "artefacto"},
		{"sin hallazgo con relato", none, "I searched knownToggles in compose.go and then in install.go and finally in the tests.\nNothing found.\n", "una línea"},
		{"sin hallazgo en dos líneas", none, none.ResultFirst + "Also checked the tests.\n", "una línea"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replyfix.Evaluate(tt.c, tt.reply)
			if got.Pass || !strings.Contains(strings.Join(got.Violations, " | "), tt.want) {
				t.Errorf("= %+v, want violación con %q", got, tt.want)
			}
		})
	}
}

// El detalle pedido después se produce a pedido: una respuesta que desarrolla porque se lo pidieron pasa.
func TestRubric_DetailOnRequestIsAllowed(t *testing.T) {
	c := replyfix.Cases()[0]
	reply := c.ResultFirst + "How I got there: listed the package, opened compose.go:35 and followed the toggle loop.\n"
	if got := replyfix.EvaluateWithDetailRequested(c, reply); !got.Pass {
		t.Errorf("con el detalle pedido el desarrollo posterior es válido: %v", got.Violations)
	}
	if got := replyfix.Evaluate(c, reply); got.Pass {
		t.Error("sin pedido, el mismo desarrollo viola la rúbrica")
	}
}

// 7.3: el documento trae el par de 7.1 y el ejemplo coincide con el fixture.
func TestDoc_ResponseContractMatchesFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "contrato-de-respuesta.md"))
	if err != nil {
		t.Fatalf("falta docs/contrato-de-respuesta.md: %v", err)
	}
	doc := string(data)
	c := replyfix.Cases()[0]
	for _, want := range []string{
		strings.TrimSpace(c.ResultFirst),
		strings.TrimSpace(c.PathFirst),
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("el documento no trae el ejemplo del fixture:\n%s", want)
		}
	}
	for _, rule := range []string{"primera línea", "segunda línea", "procedencia", "artefacto", "ningún hallazgo"} {
		if !strings.Contains(strings.ToLower(doc), rule) {
			t.Errorf("el documento no explica la regla %q", rule)
		}
	}
}

// 8.10 (corridas del 13/11): las líneas fijas de los otros módulos no son
// relato; el módulo pide conservarlas. Una nota de método después del par sí lo es.
func TestRubric_FixedLinesOfOtherModulesAreNotNarrative(t *testing.T) {
	c := replyfix.Cases()[0]
	base := "Compose returns an error for an unknown toggle: src/compose.go:36\n" +
		"Next: add the toggle to knownToggles before using it.\n"
	for _, fixed := range []string{
		"full read: src/compose.go — 107 lines, condition 1",
		"unnecessary call: Read src/compose.go:1-40",
		"unavailable: engram — uncovered: memory; alternative: none",
		"out-of-class use: context7 in explore",
		"ref: cmd=go test ./... run=r1 path=p sha256:00 expires=2026-11-20",
		"partial — unverified: never captured",
	} {
		if r := replyfix.Evaluate(c, base+fixed+"\n"); !r.Pass {
			t.Errorf("%q: %v", fixed, r.Violations)
		}
	}
	// Lo que dejaron las corridas reales después de las dos líneas.
	for _, extra := range []string{
		"Esto es una lectura inline: no delegué, no ejecuté comandos ni cambié archivos.",
		"Módulo de reporte aplicado: `concise-agent-reporting.md`.",
		"full read: and then I looked at the callers too",
	} {
		if r := replyfix.Evaluate(c, base+extra+"\n"); r.Pass {
			t.Errorf("%q después del par debe fallar", extra)
		}
	}
}
