package backup

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPersistentRestoreDiagnosticPreservesLiteralPath(t *testing.T) {
	for _, tc := range []struct{ name, path string }{
		{"windows", `C:\Users\Mati Work\AGENTS.md`},
		{"unix", "/home/Mati Work/AGENTS.md"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := Manifest{RestoreGuardVersion: 1, Entries: []ManifestEntry{{OriginalPath: tc.path}}}
			err := (RestoreService{}).Restore(m)
			if err == nil || !strings.Contains(err.Error(), `"`+filepath.Clean(tc.path)+`"`) {
				t.Fatalf("diagnostic does not preserve literal path %s: %v", tc.path, err)
			}
		})
	}
}
