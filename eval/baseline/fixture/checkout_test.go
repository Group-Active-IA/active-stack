package fixture

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitStatus(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// Escenario "Repeticiones del mismo caso": cada repeticion recibe un checkout
// nuevo con el mismo contenido aunque la anterior lo haya modificado, Reset
// restaura el estado congelado y el fixture fuente no cambia.
func TestCheckout_RepetitionsStartIdenticalAndRestore(t *testing.T) {
	for _, fam := range []Family{FamilySmallChange, FamilyExploration, FamilyReview} {
		t.Run(string(fam), func(t *testing.T) {
			t.Parallel()
			f := mustLoad(t, fam)
			srcBefore, err := TreeHash(f.Dir)
			if err != nil {
				t.Fatal(err)
			}
			repoBefore := gitStatus(t)

			c1, err := f.Prepare()
			if err != nil {
				t.Fatal(err)
			}
			h1, err := TreeHash(c1.Dir)
			if err != nil {
				t.Fatal(err)
			}
			if h1 != f.Task.Source.SHA256 {
				t.Fatalf("checkout 1 = %s, want la huella congelada %s", h1, f.Task.Source.SHA256)
			}
			// La repeticion 1 modifica, agrega y borra.
			files, _ := fileHashes(c1.Dir)
			var victim string
			for p := range files {
				victim = p
				break
			}
			os.WriteFile(filepath.Join(c1.Dir, filepath.FromSlash(victim)), []byte("modificado"), 0o644)
			os.WriteFile(filepath.Join(c1.Dir, "nuevo.txt"), []byte("x"), 0o644)
			os.MkdirAll(filepath.Join(c1.Dir, "dir_nuevo"), 0o755)
			if h, _ := TreeHash(c1.Dir); h == h1 {
				t.Fatal("la modificacion no cambio la huella: test sin sensibilidad")
			}

			c2, err := f.Prepare()
			if err != nil {
				t.Fatal(err)
			}
			if c2.Dir == c1.Dir {
				t.Fatal("dos repeticiones comparten checkout")
			}
			if h2, _ := TreeHash(c2.Dir); h2 != h1 {
				t.Errorf("repeticion 2 = %s, want %s (igual a la 1 antes de modificar)", h2, h1)
			}
			if err := c1.Reset(); err != nil {
				t.Fatal(err)
			}
			if h, _ := TreeHash(c1.Dir); h != h1 {
				t.Errorf("tras Reset = %s, want %s", h, h1)
			}
			if _, err := os.Stat(filepath.Join(c1.Dir, "dir_nuevo")); !os.IsNotExist(err) {
				t.Errorf("Reset dejo un directorio agregado: %v", err)
			}
			for _, c := range []*Checkout{c1, c2} {
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(c.Dir); !os.IsNotExist(err) {
					t.Errorf("Close no borro %s", c.Dir)
				}
			}
			if srcAfter, _ := TreeHash(f.Dir); srcAfter != srcBefore {
				t.Errorf("el fixture fuente cambio: %s -> %s", srcBefore, srcAfter)
			}
			if repoAfter := gitStatus(t); repoAfter != repoBefore {
				t.Errorf("el arbol de trabajo del repo cambio:\nantes:\n%s\ndespues:\n%s", repoBefore, repoAfter)
			}
		})
	}
}

// Una fuente embebida alterada despues de Load no se materializa.
func TestCheckout_PrepareRejectsSourceChangedAfterLoad(t *testing.T) {
	dir := tamper(t, FamilySmallChange, nil, nil)
	f, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "project", "pad.go"), []byte("package texto\n"), 0o644)
	if c, err := f.Prepare(); err == nil {
		c.Close()
		t.Fatal("Prepare con fuente alterada = nil, want error")
	}
}
