package catalog

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Tareas 6.2, 6.4, 5.5 y el contraste de 5.3 (opsx-task-context): el texto que
// 03 y 04 ponen en el SKILL.md y los documentos de docs/ se GENERAN desde el
// catálogo; el test falla si el documento o el texto se desvía de la generación.

var update = flag.Bool("update", false, "regenera los golden y los documentos generados de docs/")

const entregaRel = "openspec/changes/opsx-task-context/entrega-03-04.md"

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func checkGolden(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falta %s (generalo con `go test ./internal/catalog -update`): %v", path, err)
	}
	if string(want) != got {
		t.Errorf("%s no coincide con la generación desde el catálogo (regeneralo con -update):\n--- want\n%s\n--- got\n%s", path, want, got)
	}
}

// entregaActionTable extrae de la entrega el texto del SKILL.md de Claude (§3).
// La entrega es un artefacto del change: si ya se archivó, el test se saltea
// y la verificación queda en el golden.
func entregaActionTable(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entregaRel)))
	if err != nil {
		matches, _ := filepath.Glob(filepath.Join(root, "openspec", "changes", "archive", "*-opsx-task-context", "entrega-03-04.md"))
		if len(matches) == 0 {
			t.Skipf("entrega-03-04.md no está en el árbol (change archivado o ausente): %v", err)
		}
		data, err = os.ReadFile(matches[0])
		if err != nil {
			t.Fatal(err)
		}
	}
	text := string(data)
	_, after, ok := strings.Cut(text, "### Texto para el `SKILL.md`")
	if !ok {
		t.Fatal("la entrega no tiene la sección «Texto para el `SKILL.md`»")
	}
	re := regexp.MustCompile("(?s)````markdown\n(.*?)\n````")
	m := re.FindStringSubmatch(after)
	if m == nil {
		t.Fatal("la entrega no tiene el bloque markdown del SKILL.md")
	}
	return m[1] + "\n"
}

func TestRenderActionTable_MatchesEntregaForClaude(t *testing.T) {
	got, err := loadTC(t).RenderActionTable("claude")
	if err != nil {
		t.Fatal(err)
	}
	if want := entregaActionTable(t); got != want {
		t.Errorf("la tabla generada difiere del texto de entrega-03-04.md §3:\n--- entrega\n%s\n--- generada\n%s", want, got)
	}
}

// 8.6 (corrida real, Claude Code 2.1.319): en 4 de 6 corridas el agente leyó el
// archivo de la tarea antes que el módulo opcional sin condición. El texto fija
// el orden y que un opcional sin condición siempre aplica.
func TestRenderActionTable_ModulesBeforeFirstToolCall(t *testing.T) {
	for _, client := range []string{"claude", "codex"} {
		got, err := loadTC(t).RenderActionTable(client)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			"Before the first tool call of an action, read the modules listed for it.",
			"an optional module with no condition always applies.",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("%s: la tabla no contiene %q", client, want)
			}
		}
	}
}

// 8.8 (corrida real, Claude Code 2.1.319): sin `delegation`, apply se detuvo como
// si fuera obligatorio o delegó igual (0/2), y al retomar no dijo por qué (0/2).
func TestRenderActionTable_MissingOptionalAndResume(t *testing.T) {
	got, err := loadTC(t).RenderActionTable("claude")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Without `delegation`, do not delegate, even where other rules of this skill say to delegate.",
		"`resumed: <action> — <module | surface> available`",
	} {
		if !strings.Contains(strings.Join(strings.Fields(got), " "), want) {
			t.Errorf("la tabla no contiene %q", want)
		}
	}
}

func TestRenderActionTable_CodexDiffersOnlyInPathsAndModelAssignments(t *testing.T) {
	tc := loadTC(t)
	claude, err := tc.RenderActionTable("claude")
	if err != nil {
		t.Fatal(err)
	}
	codex, err := tc.RenderActionTable("codex")
	if err != nil {
		t.Fatal(err)
	}
	// Regla de la entrega: reemplazar la ruta y quitar model-assignments.
	want := strings.ReplaceAll(claude, "${CLAUDE_SKILL_DIR}/modules/", "modules/")
	want = strings.ReplaceAll(want, "`modules/model-assignments.md`; ", "")
	want = strings.ReplaceAll(want, "; `modules/model-assignments.md`", "")
	if codex != want {
		t.Errorf("la variante de Codex no sigue la regla de la entrega:\n--- esperada\n%s\n--- generada\n%s", want, codex)
	}
	if strings.Contains(codex, "model-assignments") || strings.Contains(codex, "CLAUDE_SKILL_DIR") {
		t.Error("Codex no debe nombrar model-assignments ni la ruta de Claude")
	}
	if _, err := tc.RenderActionTable("gemini"); err == nil {
		t.Error("un cliente sin forma de ruta declarada debe dar error")
	}
}

func TestRenderActionTable_Golden(t *testing.T) {
	tc := loadTC(t)
	for _, client := range []string{"claude", "codex"} {
		t.Run(client, func(t *testing.T) {
			got, err := tc.RenderActionTable(client)
			if err != nil {
				t.Fatal(err)
			}
			checkGolden(t, filepath.Join("testdata", "action_table_"+client+".golden"), got)
		})
	}
}

// Tarea 5.3 (contraste): la tabla del módulo task-tool-selection es la del catálogo.
func TestRenderToolTable_MatchesModuleText(t *testing.T) {
	want := loadTC(t).RenderToolTable()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "internal", "harness", "config", "assets", "modules", "task-tool-selection.md"))
	if err != nil {
		t.Skipf("el módulo todavía no existe: %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("la tabla del módulo task-tool-selection no coincide con la del catálogo:\n--- catálogo\n%s", want)
	}
}

func TestRenderToolTable_Content(t *testing.T) {
	got := loadTC(t).RenderToolTable()
	for _, line := range []string{
		"| explore | directed reading, search | Context7 |",
		"| apply | editing, tests | Context7; domain skills from the registry; Playwright only if the action touches UI |",
		"| verify | tests, directed reading | Playwright only if there is UI |",
		"| archive | OpenSpec CLI | none |",
	} {
		if !strings.Contains(got, line+"\n") {
			t.Errorf("falta la fila %q en\n%s", line, got)
		}
	}
}

// Tarea 5.5: docs/herramientas-por-accion.md se genera desde el catálogo.
func TestDocsToolsByAction(t *testing.T) {
	got := loadTC(t).RenderToolsMarkdown()
	checkGolden(t, filepath.Join(repoRoot(t), "docs", "herramientas-por-accion.md"), got)
}

// Tarea 6.4: docs/carga-de-modulos-por-accion.md se genera desde el catálogo.
func TestDocsModuleLoading(t *testing.T) {
	got := loadTC(t).RenderLoadMarkdown()
	checkGolden(t, filepath.Join(repoRoot(t), "docs", "carga-de-modulos-por-accion.md"), got)
}

func TestRenderToolsMarkdown_StatesLevelsPerClient(t *testing.T) {
	got := loadTC(t).RenderToolsMarkdown()
	for _, want := range []string{
		"solo recomendada", "fuera del contexto", "restringida por el cliente",
		"desactivada", "context7", "design-mockups",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("el documento de herramientas no menciona %q", want)
		}
	}
	if strings.Contains(got, "bloqueada por allowed-tools") {
		t.Error("allowed-tools no restringe: el documento no puede presentarlo como bloqueo")
	}
}

func TestRenderLoadMarkdown_StatesLimits(t *testing.T) {
	got := loadTC(t).RenderLoadMarkdown()
	for _, want := range []string{
		"model-assignments", "omitido en Codex", "09", "bounded-command-output", "Modo legado",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("el documento de carga no menciona %q", want)
		}
	}
}
