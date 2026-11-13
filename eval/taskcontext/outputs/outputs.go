// Package outputs inventaría las salidas de comando del ciclo OPSX sobre este
// repositorio (change opsx-task-context, tarea 1.8) y las contrasta con el par
// de umbrales de design.md D3: más de 300 líneas o más de 10.000 caracteres.
// Vive fuera del binario (CHANGES.md:49).
package outputs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Límites medidos de los clientes (design.md D3, entrega-03-04 §5):
// Claude conserva ~10.000 caracteres de un comando fallido (C1); Codex pierde
// el medio desde ~32 KB (X4).
const (
	claudeFailedKeeps = 10000
	codexLosesMiddle  = 32000
)

// Entry es una salida de comando inventariada.
type Entry struct {
	Action      string `json:"action"`
	Source      string `json:"source"`
	Command     string `json:"command"`
	Outcome     string `json:"outcome"` // success | failure
	ExitCode    int    `json:"exit_code"`
	Lines       int    `json:"lines"`
	Chars       int    `json:"chars"`
	ExceedsPair bool   `json:"exceeds_pair"`
	ClaudeTrims bool   `json:"claude_trims"`
	CodexTrims  bool   `json:"codex_trims"`
	Provoked    bool   `json:"provoked_failure,omitempty"`
}

// Inventory es el conjunto de entradas con la calibración resultante.
type Inventory struct {
	Tool        string  `json:"tool"`
	Calibration string  `json:"calibration"`
	Entries     []Entry `json:"entries"`
}

// Measure devuelve las líneas y los caracteres de una salida. Una última línea
// sin salto de línea cuenta como línea.
func Measure(out string) (lines, chars int) {
	if out == "" {
		return 0, 0
	}
	lines = strings.Count(out, "\n")
	if !strings.HasSuffix(out, "\n") {
		lines++
	}
	return lines, utf8.RuneCountInString(out)
}

// ExceedsPair dice si la salida supera el par de D3: más de las líneas o más
// de los caracteres declarados, lo que ocurra primero.
func ExceedsPair(lines, chars int, th catalog.OutputThreshold) bool {
	return lines > th.Lines || chars > th.Characters
}

// ClientTrims dice si cada cliente recortaría la salida antes de que el modelo
// la vea, según lo medido. Con éxito, Claude deja el original en un archivo del
// cliente (C6) en lugar de recortarlo sin ruta.
func ClientTrims(outcome string, chars int) (claude, codex bool) {
	claude = outcome == "failure" && chars > claudeFailedKeeps
	codex = chars > codexLosesMiddle
	return claude, codex
}

// Validate comprueba que cada entrada declare su acción y su fuente y que sus
// banderas coincidan con el cálculo.
func Validate(inv Inventory, th catalog.OutputThreshold) error {
	if len(inv.Entries) == 0 {
		return fmt.Errorf("outputs: el inventario no tiene entradas")
	}
	for _, e := range inv.Entries {
		if !slices.Contains(catalog.Actions, e.Action) {
			return fmt.Errorf("outputs: %q declara la acción %q, fuera del ciclo OPSX", e.Command, e.Action)
		}
		if e.Source == "" || e.Command == "" {
			return fmt.Errorf("outputs: entrada sin fuente o sin comando: %+v", e)
		}
		if e.Outcome != "success" && e.Outcome != "failure" {
			return fmt.Errorf("outputs: %q tiene un resultado desconocido %q", e.Command, e.Outcome)
		}
		if want := ExceedsPair(e.Lines, e.Chars, th); e.ExceedsPair != want {
			return fmt.Errorf("outputs: %q declara exceeds_pair=%v pero %d líneas y %d caracteres dan %v", e.Command, e.ExceedsPair, e.Lines, e.Chars, want)
		}
		c, x := ClientTrims(e.Outcome, e.Chars)
		if e.ClaudeTrims != c || e.CodexTrims != x {
			return fmt.Errorf("outputs: %q declara recortes (%v, %v) pero el cálculo da (%v, %v)", e.Command, e.ClaudeTrims, e.CodexTrims, c, x)
		}
	}
	return nil
}

// Load lee un inventario.
func Load(path string) (Inventory, error) {
	var inv Inventory
	data, err := os.ReadFile(path)
	if err != nil {
		return inv, err
	}
	return inv, json.Unmarshal(data, &inv)
}

// Save escribe un inventario.
func Save(path string, inv Inventory) error {
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

type spec struct {
	action, source string
	args           []string
	dir            string // vacío = raíz del repositorio
	provoked       string // nombre del fixture de fallo provocado
}

// Collect corre los comandos del ciclo sobre el repositorio y arma el
// inventario. Los fallos se provocan en un módulo temporal: no tocan el repo.
func Collect(root string, th catalog.OutputThreshold) (Inventory, error) {
	tmp, err := os.MkdirTemp("", "taskcontext-outputs-*")
	if err != nil {
		return Inventory{}, err
	}
	defer os.RemoveAll(tmp)
	failDir, err := writeFailureModule(tmp)
	if err != nil {
		return Inventory{}, err
	}

	specs := []spec{
		{action: "explore", source: "git log", args: []string{"git", "log", "--oneline", "-30"}},
		{action: "explore", source: "git status", args: []string{"git", "status", "--short"}},
		{action: "explore", source: "openspec", args: []string{"openspec", "list", "--json"}},
		{action: "propose", source: "openspec", args: []string{"openspec", "status", "--change", "opsx-task-context", "--json"}},
		{action: "propose", source: "openspec", args: []string{"openspec", "validate", "opsx-task-context"}},
		{action: "apply", source: "go build", args: []string{"go", "build", "./..."}},
		{action: "apply", source: "go vet", args: []string{"go", "vet", "./internal/catalog/"}},
		{action: "apply", source: "go test", args: []string{"go", "test", "-count=1", "./internal/catalog/"}},
		{action: "apply", source: "go test", args: []string{"go", "test", "-count=1", "-v", "./internal/catalog/"}},
		{action: "apply", source: "go test", args: []string{"go", "test", "-count=1", "./internal/..."}},
		{action: "apply", source: "go test", args: []string{"go", "test", "-count=1", "-run", "TestShortFail", "./short/"}, dir: failDir, provoked: "falla corta"},
		{action: "apply", source: "go test", args: []string{"go", "test", "-count=1", "-v", "-run", "TestVerboseFail", "./verbose/"}, dir: failDir, provoked: "falla al medio de 3.000 líneas"},
		{action: "apply", source: "go build", args: []string{"go", "build", "./broken/"}, dir: failDir, provoked: "error de compilación"},
		{action: "apply", source: "go vet", args: []string{"go", "vet", "./vetbad/"}, dir: failDir, provoked: "hallazgo de vet"},
		{action: "verify", source: "go test", args: []string{"go", "test", "-count=1", "-v", "./internal/catalog/", "./internal/model/", "./internal/harness/config/..."}},
		{action: "verify", source: "openspec", args: []string{"openspec", "validate", "opsx-task-context"}},
		{action: "verify", source: "go test", args: []string{"go", "test", "-count=1", "-v", "-run", "TestVerboseFail", "./verbose/"}, dir: failDir, provoked: "falla al medio de 3.000 líneas"},
		{action: "archive", source: "openspec", args: []string{"openspec", "list", "--json"}},
	}

	inv := Inventory{Tool: "eval/taskcontext/outputs (go test -update)"}
	for _, s := range specs {
		dir := s.dir
		if dir == "" {
			dir = root
		}
		cmd := exec.Command(s.args[0], s.args[1:]...)
		cmd.Dir = dir
		out, runErr := cmd.CombinedOutput()
		code := 0
		if runErr != nil {
			if ee, ok := runErr.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else {
				return inv, fmt.Errorf("outputs: no se pudo correr %v: %w", s.args, runErr)
			}
		}
		lines, chars := Measure(string(out))
		outcome := "success"
		if code != 0 {
			outcome = "failure"
		}
		if s.provoked == "" && code != 0 {
			return inv, fmt.Errorf("outputs: %v falló sin haberlo provocado (código %d)", s.args, code)
		}
		c, x := ClientTrims(outcome, chars)
		inv.Entries = append(inv.Entries, Entry{
			Action: s.action, Source: s.source, Command: strings.Join(s.args, " "),
			Outcome: outcome, ExitCode: code, Lines: lines, Chars: chars,
			ExceedsPair: ExceedsPair(lines, chars, th), ClaudeTrims: c, CodexTrims: x,
			Provoked: s.provoked != "",
		})
	}
	inv.Calibration = calibrate(inv, th)
	return inv, nil
}

// calibrate resume qué dice el inventario sobre el par de D3.
func calibrate(inv Inventory, th catalog.OutputThreshold) string {
	var byLines, byChars, both, missed int
	for _, e := range inv.Entries {
		l, c := e.Lines > th.Lines, e.Chars > th.Characters
		switch {
		case l && c:
			both++
		case l:
			byLines++
		case c:
			byChars++
		}
		if (e.ClaudeTrims || e.CodexTrims) && !e.ExceedsPair {
			missed++
		}
	}
	return fmt.Sprintf("%d salidas; superan el par por líneas y caracteres: %d, solo por líneas: %d, solo por caracteres: %d; recortadas por un cliente y no resumidas por el par: %d",
		len(inv.Entries), both, byLines, byChars, missed)
}

// writeFailureModule crea un módulo Go temporal con los fallos provocados.
func writeFailureModule(dir string) (string, error) {
	files := map[string]string{
		"go.mod": "module taskcontextfix\n\ngo 1.22\n",
		"short/short_test.go": `package short

import "testing"

func TestShortFail(t *testing.T) { t.Fatal("expected 4, got 5") }
`,
		"verbose/verbose_test.go": `package verbose

import (
	"fmt"
	"testing"
)

func TestVerboseFail(t *testing.T) {
	for i := 1; i <= 3000; i++ {
		if i == 1500 {
			t.Errorf("needle_middle_5521 at line %d", i)
			continue
		}
		t.Logf("line %d padding padding padding", i)
	}
	fmt.Println("done")
}
`,
		"broken/broken.go": `package broken

func F() int { return "no es un entero" }
`,
		"vetbad/vetbad.go": `package vetbad

import "fmt"

func F() { fmt.Printf("%d\n", "texto") }
`,
	}
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return "", err
		}
	}
	return dir, nil
}
