package catalog

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
)

// Tarea 3.1 (opsx-task-context, D14): la lista única de patrones sensibles
// vive en el catálogo y cubre todo lo que bloquea el piso de Claude.

func TestLoad_SensitivePatternsDeclared(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := map[string]bool{}
	for _, p := range c.TaskContext.SensitivePatterns {
		got[p.ID] = true
	}
	// Un patrón por cada categoría de D14.
	for _, id := range []string{
		"env-assignment", "env-var-value", "token", "api-key", "cloud-credential",
		"ssh-private-key", "home-path", "account-id", "account-uuid", "email",
	} {
		if !got[id] {
			t.Errorf("falta el patrón sensible %q (D14)", id)
		}
	}
}

func TestParseTaskContext_RejectsMalformedSensitivePatterns(t *testing.T) {
	good := mustTaskContextYAML(t)
	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{"regex inválida", func(s string) string {
			return strings.Replace(s, `regex: '(?P<secret>[A-Za-z0-9._%+\-]+@`, `regex: '(?P<secret>((`, 1)
		}, "regex"},
		{"id duplicado", func(s string) string {
			return strings.Replace(s, "id: email", "id: token", 1)
		}, "duplicate"},
		{"tipo desconocido", func(s string) string {
			return strings.Replace(s, "kind: block", "kind: multiline", 1)
		}, "kind"},
		{"bloque sin fin", func(s string) string {
			return strings.Replace(s, "    end: '-----END [A-Z0-9 ]*PRIVATE KEY-----'\n", "", 1)
		}, "begin and end"},
		{"patrón de línea sin regex", func(s string) string {
			return strings.Replace(s, `    regex: '(?P<secret>[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,})'`+"\n", "", 1)
		}, "regex"},
		{"patrón de fallo inválido", func(s string) string {
			return strings.Replace(s, `'^panic:'`, `'^panic:(('`, 1)
		}, "failure_patterns"},
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

// El piso de Claude (overlays.go) bloquea rutas con Read/Edit y comandos
// destructivos con Bash. Lo que apunta a un archivo sensible debe estar
// cubierto por algún patrón; los comandos destructivos no tienen valor que
// redactar y quedan fuera, nombrados.
func TestSensitivePatterns_CoverClaudeSecurityFloor(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	covered := coveredFloorTargets(c.TaskContext)
	for _, rule := range permissions.SecurityFloorDeny() {
		tool, target := splitRule(t, rule)
		if tool == "Bash" {
			continue // comando destructivo: no hay valor sensible que redactar
		}
		if !covered[target] {
			t.Errorf("la regla del piso %q no está cubierta por ningún patrón sensible (target %q): agregá un patrón con `covers: [%q]`", rule, target, target)
		}
	}
}

// Si el piso agrega una regla de archivo no cubierta, el test falla nombrándola.
func TestFloorCoverage_DetectsUncoveredRule(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	covered := coveredFloorTargets(c.TaskContext)
	for _, rule := range []string{"Read(.env)", "Edit(.env.*)", "Read(secrets/*.key)"} {
		_, target := splitRule(t, rule)
		want := rule != "Read(secrets/*.key)"
		if covered[target] != want {
			t.Errorf("cobertura de %q = %v, want %v", rule, covered[target], want)
		}
	}
}

func coveredFloorTargets(tc *TaskContext) map[string]bool {
	out := map[string]bool{}
	for _, p := range tc.SensitivePatterns {
		for _, c := range p.Covers {
			out[c] = true
		}
	}
	return out
}

var ruleRe = regexp.MustCompile(`^([A-Za-z]+)\((.*)\)$`)

func splitRule(t *testing.T, rule string) (tool, target string) {
	t.Helper()
	m := ruleRe.FindStringSubmatch(rule)
	if m == nil {
		t.Fatalf("regla del piso con forma inesperada: %q", rule)
	}
	return m[1], m[2]
}
