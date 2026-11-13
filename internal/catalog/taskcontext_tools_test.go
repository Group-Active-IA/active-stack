package catalog

import (
	"strings"
	"testing"
)

// Tareas 5.1 y 5.2 (opsx-task-context, D18): tabla de herramientas por acción,
// integraciones opcionales apagadas por defecto y nivel alcanzable por cliente.

func TestLoad_ToolSelectionTable(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	ts := c.TaskContext.ToolSelection
	var ids []string
	for _, a := range ts.Actions {
		ids = append(ids, a.ID)
	}
	if strings.Join(ids, ",") != "explore,propose,apply,verify,archive" {
		t.Errorf("acciones = %v, want las cinco del ciclo OPSX en orden", ids)
	}
	for _, in := range ts.Integrations {
		if in.Default != "disabled" {
			t.Errorf("integración %q con default %q, want disabled", in.ID, in.Default)
		}
		if len(in.Actions) == 0 {
			t.Errorf("integración %q sin acción declarada", in.ID)
		}
	}
	// Las integraciones de diseño exigen el harness design-mockups.
	for _, id := range []string{"stitch", "emil-design-eng", "design-taste-frontend-v1"} {
		in := findIntegration(ts, id)
		if in == nil || in.RequiresHarness != "design-mockups" {
			t.Errorf("integración de diseño %q debe requerir design-mockups, got %+v", id, in)
		}
	}
}

func TestLoad_ToolLevelsPerClient(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range c.TaskContext.ToolSelection.Integrations {
		for _, client := range []string{"claude", "codex"} {
			lv, ok := in.Levels[client]
			if !ok || lv.Mechanism == "" {
				t.Errorf("integración %q sin nivel o mecanismo para %s", in.ID, client)
			}
		}
	}
	// D18: el control por acción de un MCP en Codex es solo (a).
	if lv := findIntegration(c.TaskContext.ToolSelection, "context7").Levels["codex"]; lv.Level != "a" {
		t.Errorf("context7 en Codex = nivel %q, want a (solo recomendada)", lv.Level)
	}
	// Las skills en Claude se ocultan con disable-model-invocation: nivel (b).
	if lv := findIntegration(c.TaskContext.ToolSelection, "domain-skills").Levels["claude"]; lv.Level != "b" {
		t.Errorf("domain-skills en Claude = nivel %q, want b (fuera del contexto)", lv.Level)
	}
}

func TestLevelName(t *testing.T) {
	tests := map[string]string{"a": "solo recomendada", "b": "fuera del contexto", "c": "restringida por el cliente"}
	for level, want := range tests {
		if got := LevelName(level); got != want {
			t.Errorf("LevelName(%q) = %q, want %q", level, got, want)
		}
	}
	if got := LevelName("z"); got != "" {
		t.Errorf("LevelName(z) = %q, want vacío", got)
	}
}

func TestParseTaskContext_RejectsInvalidToolSelection(t *testing.T) {
	good := mustTaskContextYAML(t)
	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{"integración sin acción declarada", func(s string) string {
			return strings.Replace(s, "      actions: [explore, propose, apply]\n", "      actions: []\n", 1)
		}, "no action"},
		{"integración activada por defecto", func(s string) string {
			return strings.Replace(s, "    - id: context7\n      default: disabled", "    - id: context7\n      default: enabled", 1)
		}, "default"},
		{"sin nivel para Codex", func(s string) string {
			i := strings.Index(s, "    - id: context7")
			j := strings.Index(s[i:], "        codex:")
			k := strings.Index(s[i+j:], "    - id: playwright")
			return s[:i+j] + s[i+j+k:]
		}, "codex"},
		{"nivel inválido", func(s string) string {
			return strings.Replace(s, "          level: a", "          level: d", 1)
		}, "level"},
		{"allowed-tools como nivel (c)", func(s string) string {
			return strings.Replace(s, "          level: a\n          mechanism: \"CLAUDE.md o cuerpo de la skill (as context", "          level: c\n          mechanism: \"allowed-tools de la skill (as context", 1)
		}, "allowed-tools"},
		{"acción de la tabla sin integración declarada", func(s string) string {
			return strings.Replace(s, "integration: playwright, label: Playwright only if there is UI", "integration: playwrite, label: Playwright only if there is UI", 1)
		}, "playwrite"},
		{"acción que la integración no lista", func(s string) string {
			return strings.Replace(s, "    - id: archive\n      base: [OpenSpec CLI]\n      optional: []", "    - id: archive\n      base: [OpenSpec CLI]\n      optional:\n        - { integration: context7, label: Context7 }", 1)
		}, "archive"},
		{"falta una acción del ciclo", func(s string) string {
			i := strings.Index(s, "    - id: archive\n      base:")
			j := strings.Index(s[i:], "  # Nivel alcanzable")
			return s[:i] + s[i+j:]
		}, "actions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutated := tt.mutate(good)
			if mutated == good {
				t.Fatal("la mutación no cambió el YAML: el test no prueba nada")
			}
			_, err := ParseTaskContext([]byte(mutated))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want mención de %q", err, tt.wantErr)
			}
		})
	}
}

func findIntegration(ts ToolSelection, id string) *Integration {
	for i := range ts.Integrations {
		if ts.Integrations[i].ID == id {
			return &ts.Integrations[i]
		}
	}
	return nil
}
