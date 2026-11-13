package catalog

import (
	"strings"
	"testing"
)

// Tareas 3.2 y 4.1 (opsx-task-context): los umbrales viven en el catálogo y
// catalog.Load() los valida.

func TestLoad_TaskContextThresholds(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	tc := c.TaskContext
	if tc == nil {
		t.Fatal("Load() no cargó la declaración de task-context")
	}
	if tc.OutputThreshold.Lines != 300 || tc.OutputThreshold.Characters != 10000 {
		t.Errorf("umbral de salidas = %+v, want 300 líneas y 10.000 caracteres (D3)", tc.OutputThreshold)
	}
	if tc.ReadThreshold.Unit != "lines" || tc.ReadThreshold.Value != 300 {
		t.Errorf("umbral de lectura = %+v, want 300 líneas (D11)", tc.ReadThreshold)
	}
}

func TestParseTaskContext_RejectsInvalidThresholds(t *testing.T) {
	good := mustTaskContextYAML(t)
	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{"salidas sin líneas", func(s string) string {
			return strings.Replace(s, "  lines: 300\n  characters: 10000", "  characters: 10000", 1)
		}, "output_threshold"},
		{"salidas sin caracteres", func(s string) string {
			return strings.Replace(s, "  lines: 300\n  characters: 10000", "  lines: 300", 1)
		}, "output_threshold"},
		{"lectura en caracteres", func(s string) string {
			return strings.Replace(s, "unit: lines", "unit: characters", 1)
		}, "read_threshold"},
		{"lectura en bytes", func(s string) string {
			return strings.Replace(s, "unit: lines", "unit: bytes", 1)
		}, "read_threshold"},
		{"lectura sin valor", func(s string) string {
			return strings.Replace(s, "  value: 300", "  value: 0", 1)
		}, "read_threshold"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTaskContext([]byte(tt.mutate(good)))
			if err == nil {
				t.Fatalf("ParseTaskContext aceptó %q", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want mención de %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseTaskContext_AcceptsDeclaredCatalog(t *testing.T) {
	if _, err := ParseTaskContext([]byte(mustTaskContextYAML(t))); err != nil {
		t.Fatalf("el catálogo embebido debe validar: %v", err)
	}
}

// Tarea 4.3: la lista de condiciones de lectura completa es cerrada.
func TestParseTaskContext_FullReadConditionsClosed(t *testing.T) {
	good := mustTaskContextYAML(t)
	tc, err := ParseTaskContext([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"within-threshold", "rewrite-whole-file", "action-artifact", "empty-bounded-read", "no-range-covers"}
	if strings.Join(tc.FullReadConditions, ",") != strings.Join(want, ",") {
		t.Errorf("condiciones = %v, want %v", tc.FullReadConditions, want)
	}
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{"condición fuera de la lista", func(s string) string {
			return strings.Replace(s, "  - no-range-covers", "  - no-range-covers\n  - file-looks-small", 1)
		}},
		{"falta una de las cinco", func(s string) string {
			return strings.Replace(s, "  - empty-bounded-read\n", "", 1)
		}},
		{"repetida", func(s string) string {
			return strings.Replace(s, "  - action-artifact", "  - action-artifact\n  - action-artifact", 1)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTaskContext([]byte(tt.mutate(good)))
			if err == nil || !strings.Contains(err.Error(), "full_read_conditions") {
				t.Fatalf("err = %v, want rechazo de full_read_conditions", err)
			}
		})
	}
}

func mustTaskContextYAML(t *testing.T) string {
	t.Helper()
	return string(rawTaskContext)
}
