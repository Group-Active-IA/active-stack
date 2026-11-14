package shadow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/fixture"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, c := range files {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// 3.1: sólo entra lo que nombra la lista de inclusión.
func TestBuild_OnlyIncludedFiles(t *testing.T) {
	root := writeTree(t, map[string]string{
		"diff.patch":         "+a\n",
		"docs/proposal.md":   "# p\n",
		"no-incluido.txt":    "fuera\n",
		"docs/no-tampoco.md": "fuera\n",
	})
	p, err := Build("c1", root, []string{"docs/proposal.md", "diff.patch"})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range p.Files {
		got = append(got, f.Path)
	}
	want := []string{"diff.patch", "docs/proposal.md"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("archivos = %q, esperado %q (ordenados)", got, want)
	}
	if p.Case != "c1" || p.Fingerprint.Algorithm != "sha256" || p.Fingerprint.Value == "" {
		t.Errorf("paquete sin caso o sin huella: %+v", p)
	}
}

// 3.1: la huella del paquete usa la convención de TreeHash de 02: un directorio con exactamente
// los mismos archivos da el mismo valor.
func TestPackageFingerprint_MatchesTreeHashOf02(t *testing.T) {
	files := map[string]string{"diff.patch": "+a\r\n", "docs/proposal.md": "# p\n", "z/y/x.txt": "x"}
	dir := writeTree(t, files)
	var inc []string
	for p := range files {
		inc = append(inc, p)
	}
	p, err := Build("c1", dir, inc)
	if err != nil {
		t.Fatal(err)
	}
	want, err := fixture.TreeHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Fingerprint.Value != want {
		t.Errorf("huella del paquete %s, TreeHash de 02 %s", p.Fingerprint.Value, want)
	}
}

// 3.1: la lista es explícita; lo que no es un archivo nombrado bajo la raíz es un error.
func TestBuild_RejectsBadEntries(t *testing.T) {
	// fuera.txt existe al lado de la raíz: sólo el chequeo de ".." puede impedir que entre.
	parent := writeTree(t, map[string]string{"fuera.txt": "f", "raiz/a.txt": "a", "raiz/dir/b.txt": "b"})
	root := filepath.Join(parent, "raiz")
	abs, _ := filepath.Abs(filepath.Join(root, "a.txt"))
	for name, inc := range map[string][]string{
		"sube con ..":           {"../fuera.txt"},
		"sube con .. (Windows)": {`..\fuera.txt`},
		"absoluta":              {abs},
		"repetida":              {"a.txt", "a.txt"},
		"inexistente":           {"falta.txt"},
		"directorio":            {"dir"},
		"vacía":                 {},
		"sube anidada":          {"dir/../../fuera.txt"},
	} {
		if _, err := Build("c", root, inc); err == nil {
			t.Errorf("%s: %q debía fallar", name, inc)
		}
	}
}
