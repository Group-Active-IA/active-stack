package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tareas 4.4, 5.3 y 7.2 (opsx-task-context): los textos de módulo viven como
// archivos en assets/modules/, sin cablear al compositor (lo hacen 03 y 04
// cuando integren su modo de entrega). La 6.3 vive en eval/taskcontext porque
// compara con las recetas de captura del arnés.

var taskContextModules = []string{
	"directed-code-reading",
	"bounded-command-output",
	"task-tool-selection",
	"concise-agent-reporting",
}

func readEmbeddedModule(t *testing.T, id string) string {
	t.Helper()
	data, err := fs.ReadFile(embeddedAssets, "assets/modules/"+id+".md")
	if err != nil {
		t.Fatalf("el módulo %q no está embebido en assets/modules/: %v", id, err)
	}
	return string(data)
}

func TestModules_ExistWithoutSkillFrontmatter(t *testing.T) {
	for _, id := range taskContextModules {
		t.Run(id, func(t *testing.T) {
			text := readEmbeddedModule(t, id)
			if strings.TrimSpace(text) == "" {
				t.Fatal("un módulo vacío cuenta como ausente")
			}
			if strings.HasPrefix(text, "---") {
				t.Error("el módulo lleva frontmatter: el cliente lo listaría como skill")
			}
			if !strings.HasPrefix(text, "## ") {
				t.Errorf("el módulo debe empezar con `## <Título>`, empieza con %q", firstLine(text))
			}
			for _, banned := range []string{"\n## Core Principles", "\n## Rules"} {
				if strings.Contains(text, banned) {
					t.Errorf("el módulo usa el encabezado de agrupación %q que la matriz común reserva", strings.TrimSpace(banned))
				}
			}
		})
	}
}

// El texto es el de la entrega a 03 y 04, byte a byte. Si el change ya se
// archivó, la comparación se saltea (la fuente deja de estar en el árbol).
func TestModules_MatchEntregaByteForByte(t *testing.T) {
	entrega := readEntrega(t)
	for _, id := range taskContextModules {
		t.Run(id, func(t *testing.T) {
			re := regexp.MustCompile("(?s)### `modules/" + regexp.QuoteMeta(id) + "\\.md`\n\n````markdown\n(.*?)\n````")
			m := re.FindStringSubmatch(entrega)
			if m == nil {
				t.Fatalf("la entrega no trae el texto de %s", id)
			}
			if got := readEmbeddedModule(t, id); got != m[1]+"\n" {
				t.Errorf("el módulo %s difiere del texto de entrega-03-04.md §2:\n--- entrega\n%s\n--- archivo\n%s", id, m[1], got)
			}
		})
	}
}

func TestModule_DirectedCodeReadingFixedForms(t *testing.T) {
	text := readEmbeddedModule(t, "directed-code-reading")
	for _, want := range []string{
		"`full read: <path> — <reason>`",
		"`unnecessary call: <tool> <input>`",
		"300 lines or fewer",
		"path:line",
		// 8.6: sin conocer el tamaño, el agente abría el archivo entero (3/3) y,
		// con la relectura pedida por el usuario, omitía la línea fija (2/2).
		"do not open it whole to find out",
		"even when the user asked for the repeat",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("directed-code-reading no contiene %q", want)
		}
	}
	// Las cinco condiciones de D15, numeradas.
	for i := 1; i <= 5; i++ {
		if !strings.Contains(text, "  "+string(rune('0'+i))+". ") {
			t.Errorf("falta la condición %d de lectura completa", i)
		}
	}
	if strings.Contains(text, "  6. ") {
		t.Error("la lista de condiciones es cerrada: no puede haber una sexta")
	}
}

func TestModule_TaskToolSelection(t *testing.T) {
	text := readEmbeddedModule(t, "task-tool-selection")
	for _, want := range []string{
		"`unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`",
		"`unnecessary call: <tool> <input>`",
		"`out-of-class use: <integration> in <action>`",
		// 8.7: con el MCP caído el cliente no lista sus herramientas y el agente
		// no lo declaraba (3/3); con la repetición pedida omitía la línea (2/2).
		"its tools are not listed in this session",
		"even when the user asked for the repeat",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("task-tool-selection no contiene la forma fija %q", want)
		}
	}
	// El nivel (a) no se presenta como bloqueo.
	if !strings.Contains(strings.Join(strings.Fields(text), " "), "Never describe a rule that is only recommended as a block") {
		t.Error("el módulo debe prohibir presentar una regla solo recomendada como bloqueo")
	}
	for _, bad := range []string{"is blocked", "cannot call", "the client blocks"} {
		if strings.Contains(strings.ToLower(text), bad) {
			t.Errorf("el módulo presenta un control (a) como bloqueo: %q", bad)
		}
	}
	// La tabla es la del catálogo (5.1).
	c, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	if want := c.TaskContext.RenderToolTable(); !strings.Contains(text, want) {
		t.Errorf("la tabla del módulo no coincide con la del catálogo:\n%s", want)
	}
}

func TestModule_ConciseAgentReporting(t *testing.T) {
	text := readEmbeddedModule(t, "concise-agent-reporting")
	for _, want := range []string{
		"Line 1: the result", "Line 2: the consequence or the next step",
		"`full read:`", "`unnecessary call:`", "`ref:`", "`partial — unverified:`",
		"`unavailable:`", "`out-of-class use:`",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("concise-agent-reporting no contiene %q", want)
		}
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// readEntrega lee entrega-03-04.md del change (o de su archivo). Si no está
// en el árbol, el test se saltea: el contraste con la entrega deja de existir.
func readEntrega(t *testing.T) string {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "opsx-task-context", "entrega-03-04.md"))
	if err != nil {
		matches, _ := filepath.Glob(filepath.Join(root, "openspec", "changes", "archive", "*-opsx-task-context", "entrega-03-04.md"))
		if len(matches) == 0 {
			t.Skipf("entrega-03-04.md no está en el árbol: %v", err)
		}
		if data, err = os.ReadFile(matches[0]); err != nil {
			t.Fatal(err)
		}
	}
	return string(data)
}
