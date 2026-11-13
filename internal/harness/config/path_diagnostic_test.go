package config

import (
	"strings"
	"testing"
)

func TestManagedPathDiagnosticPreservesLiteralPath(t *testing.T) {
	for _, tc := range []struct{ name, path string }{
		{"windows", `C:\Users\Mati Work\skills\sdd-orchestrator`},
		{"unix", "/outside/Mati Work/skills/sdd-orchestrator"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkRegularParents(tc.path, t.TempDir())
			if err == nil || !strings.Contains(err.Error(), `"`+tc.path+`"`) {
				t.Fatalf("diagnostic does not preserve literal path %s: %v", tc.path, err)
			}
		})
	}
}
