package uninstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

func TestManagedRollbackDiagnosticPreservesLiteralPaths(t *testing.T) {
	for _, name := range []string{`nested\Mati Work`, "Mati Work"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			parent := filepath.Join(home, name)
			if err := os.MkdirAll(parent, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(parent, "AGENTS.md")
			point := filepath.Join(home, `backup\Point One`)
			r := newManagedRollback()
			r.owned[path], r.changed[path], r.parents[parent] = true, true, nil
			err := r.restore(backup.Manifest{RootDir: point, Entries: []backup.ManifestEntry{{OriginalPath: path}}})
			if err == nil {
				t.Fatal("expected replaced-parent conflict")
			}
			for _, expected := range []string{path, parent, point} {
				if !strings.Contains(err.Error(), `"`+expected+`"`) {
					t.Fatalf("diagnostic does not preserve literal path %s: %v", expected, err)
				}
			}
		})
	}
}
