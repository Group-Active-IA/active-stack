package outputs_test

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/outputs"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tarea 1.8 (opsx-task-context): inventario de las salidas de comando del
// ciclo OPSX sobre este repositorio, para calibrar el umbral de D3.

var update = flag.Bool("update", false, "vuelve a correr los comandos y reescribe outputs.json")

func TestMeasure(t *testing.T) {
	tests := []struct {
		name      string
		out       string
		wantLines int
		wantChars int
	}{
		{"vacía", "", 0, 0},
		{"dos líneas con salto final", "a\nb\n", 2, 4},
		{"dos líneas sin salto final", "a\nb", 2, 3},
		{"cuenta caracteres, no bytes", "ñé\n", 1, 3},
		{"líneas en blanco cuentan", "\n\n\n", 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, c := outputs.Measure(tt.out)
			if l != tt.wantLines || c != tt.wantChars {
				t.Errorf("Measure(%q) = %d líneas, %d caracteres; want %d, %d", tt.out, l, c, tt.wantLines, tt.wantChars)
			}
		})
	}
}

func TestExceedsPair(t *testing.T) {
	th := catalog.OutputThreshold{Lines: 300, Characters: 10000}
	tests := []struct {
		name         string
		lines, chars int
		want         bool
	}{
		{"justo en el umbral de líneas", 300, 100, false},
		{"justo en el umbral de caracteres", 10, 10000, false},
		{"una línea de más", 301, 100, true},
		{"un carácter de más", 10, 10001, true},
		{"las dos", 5000, 400000, true},
		{"corta", 1, 20, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := outputs.ExceedsPair(tt.lines, tt.chars, th); got != tt.want {
				t.Errorf("ExceedsPair(%d, %d) = %v, want %v", tt.lines, tt.chars, got, tt.want)
			}
		})
	}
}

func TestClientTrims(t *testing.T) {
	// Medido en la ficha (design.md D3): Claude conserva ~10.000 caracteres de
	// un comando fallido; Codex pierde el medio desde ~32 KB.
	tests := []struct {
		name                  string
		outcome               string
		chars                 int
		wantClaude, wantCodex bool
	}{
		{"fallo corto", "failure", 500, false, false},
		{"fallo de 12.000 caracteres", "failure", 12000, true, false},
		{"fallo de 40.000 caracteres", "failure", 40000, true, true},
		// Con éxito Claude guarda el original en un archivo del cliente, no lo recorta sin ruta.
		{"éxito de 40.000 caracteres", "success", 40000, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, x := outputs.ClientTrims(tt.outcome, tt.chars)
			if c != tt.wantClaude || x != tt.wantCodex {
				t.Errorf("ClientTrims(%s, %d) = claude %v, codex %v; want %v, %v", tt.outcome, tt.chars, c, x, tt.wantClaude, tt.wantCodex)
			}
		})
	}
}

func TestInventory(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	th := cat.TaskContext.OutputThreshold
	path := filepath.Join("outputs.json")
	if *update {
		inv, err := outputs.Collect(filepath.Join("..", "..", ".."), th)
		if err != nil {
			t.Fatal(err)
		}
		if err := outputs.Save(path, inv); err != nil {
			t.Fatal(err)
		}
	}
	inv, err := outputs.Load(path)
	if err != nil {
		t.Fatalf("falta outputs.json (generalo con -update): %v", err)
	}
	if err := outputs.Validate(inv, th); err != nil {
		t.Fatal(err)
	}
	// Cada acción del ciclo tiene al menos una salida de éxito; apply y verify,
	// además, una de fallo provocado.
	success, failure := map[string]bool{}, map[string]bool{}
	for _, e := range inv.Entries {
		if e.Outcome == "success" {
			success[e.Action] = true
		} else {
			failure[e.Action] = true
		}
	}
	for _, a := range catalog.Actions {
		if !success[a] {
			t.Errorf("la acción %q no tiene ninguna salida de éxito inventariada", a)
		}
	}
	for _, a := range []string{"apply", "verify"} {
		if !failure[a] {
			t.Errorf("la acción %q no tiene un fallo provocado inventariado", a)
		}
	}
	// Ninguna salida que un cliente recortaría queda afuera del par.
	for _, e := range inv.Entries {
		if (e.ClaudeTrims || e.CodexTrims) && !e.ExceedsPair {
			t.Errorf("%s: un cliente la recortaría pero el par (300 líneas o 10.000 caracteres) no la resume", e.Command)
		}
	}
	t.Logf("%d salidas; %d superan el par; calibración: %s", len(inv.Entries), countExceeding(inv), inv.Calibration)
}

func countExceeding(inv outputs.Inventory) int {
	n := 0
	for _, e := range inv.Entries {
		if e.ExceedsPair {
			n++
		}
	}
	return n
}

func TestValidate_RejectsIncompleteEntries(t *testing.T) {
	th := catalog.OutputThreshold{Lines: 300, Characters: 10000}
	good := outputs.Entry{Action: "apply", Source: "go test", Command: "go test ./x", Outcome: "success", Lines: 2, Chars: 30}
	good.ExceedsPair = outputs.ExceedsPair(good.Lines, good.Chars, th)
	tests := []struct {
		name   string
		mutate func(*outputs.Entry)
	}{
		{"sin acción", func(e *outputs.Entry) { e.Action = "" }},
		{"acción fuera del ciclo", func(e *outputs.Entry) { e.Action = "deploy" }},
		{"sin fuente", func(e *outputs.Entry) { e.Source = "" }},
		{"resultado desconocido", func(e *outputs.Entry) { e.Outcome = "maybe" }},
		{"bandera del par inconsistente", func(e *outputs.Entry) { e.ExceedsPair = true }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := good
			tt.mutate(&e)
			if err := outputs.Validate(outputs.Inventory{Entries: []outputs.Entry{e}}, th); err == nil {
				t.Error("Validate aceptó una entrada inválida")
			}
		})
	}
	if err := outputs.Validate(outputs.Inventory{Entries: []outputs.Entry{good}}, th); err != nil {
		t.Errorf("Validate rechazó una entrada válida: %v", err)
	}
}
