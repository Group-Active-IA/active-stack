package capture_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
)

// Regresiones de la segunda revisión del PR #10: la raíz administrada del store
// se ancla, y si después de registrar capturas se la sustituye por otra (otro
// directorio con el mismo nombre o un symlink a otro store), el store no lee
// el índice ajeno ni borra lo que está debajo. Solo datos ficticios bajo t.TempDir().

// twoStores arma dos stores reales con una corrida de _sin-change ya guardada en
// cada uno y una nota de usuario en la corrida del ajeno.
func twoStores(t *testing.T) (own *capture.Store, ownCache string, foreign *capture.Store, note string, c *clock) {
	t.Helper()
	c = &clock{now: t0}
	ownCache, foreignCache := t.TempDir(), t.TempDir()
	own = capture.NewStore(ownCache, c.Now)
	foreign = capture.NewStore(foreignCache, c.Now)
	for _, s := range []*capture.Store{own, foreign} {
		if _, err := s.Save("", "r1", goodRecord(), []byte("x\n")); err != nil {
			t.Fatal(err)
		}
	}
	note = filepath.Join(foreign.RunDir("", "r1"), "foreign-user-note.txt")
	if err := os.WriteFile(note, []byte("ajeno"), 0o600); err != nil {
		t.Fatal(err)
	}
	return own, ownCache, foreign, note, c
}

func assertForeignIntact(t *testing.T, own *capture.Store, note string) {
	t.Helper()
	removed, err := own.Purge()
	if err == nil {
		t.Errorf("Purge no detectó la raíz sustituida (borró %v)", removed)
	}
	if _, statErr := os.Stat(note); statErr != nil {
		t.Errorf("Purge borró la nota del otro store: %v", statErr)
	}
}

func TestReview2RootReplacedByOtherDirCannotPurgeForeignStore(t *testing.T) {
	own, _, foreign, note, c := twoStores(t)
	// Sin symlinks (corre también en Windows): la raíz propia se aparta y el
	// directorio del otro store ocupa su nombre.
	if err := os.Rename(own.Root, own.Root+".bak"); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(foreign.Root, own.Root); err != nil {
		t.Fatal(err)
	}
	note = filepath.Join(own.Root, capture.NoChange, "r1", filepath.Base(note))
	c.advance(capture.Retention + day)

	assertForeignIntact(t, own, note)
	t.Run("otro proceso con la misma configuración tampoco", func(t *testing.T) {
		fresh := capture.NewStore(filepath.Dir(filepath.Dir(own.Root)), c.Now)
		assertForeignIntact(t, fresh, note)
	})
	t.Run("Save tampoco escribe en la raíz ajena", func(t *testing.T) {
		if _, err := own.Save("", "r2", goodRecord(), []byte("y\n")); err == nil {
			t.Error("Save aceptó escribir en la raíz sustituida")
		}
	})
}

func TestReview2RootReplacedBySymlinkCannotPurgeForeignStore(t *testing.T) {
	own, _, foreign, note, c := twoStores(t)
	if err := os.Rename(own.Root, own.Root+".bak"); err != nil {
		t.Fatal(err)
	}
	mustSymlink(t, foreign.Root, own.Root)
	c.advance(capture.Retention + day)

	assertForeignIntact(t, own, note)
}

func TestReview2IntentionalRootAliasStillWorks(t *testing.T) {
	// Un alias elegido al configurar el cache (antes de la primera captura) no
	// es una sustitución: el store sigue funcionando a través de él.
	real := t.TempDir()
	aliasParent := t.TempDir()
	alias := filepath.Join(aliasParent, "cache")
	mustSymlink(t, real, alias)
	c := &clock{now: t0}
	s := capture.NewStore(alias, c.Now)
	if _, err := s.Save("", "r1", goodRecord(), []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	c.advance(capture.Retention + day)
	removed, err := s.Purge()
	if err != nil || len(removed) != 1 {
		t.Fatalf("Purge con alias configurado: removed=%v err=%v", removed, err)
	}
}

func TestReview2LegitimateStoreAcrossInstances(t *testing.T) {
	// Triangulación: otro Store con la misma configuración sobre la raíz intacta
	// purga normalmente (el ancla no rompe el uso entre procesos).
	s, c, cache := newStore(t)
	if _, err := s.Save("", "r1", goodRecord(), []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	c.advance(capture.Retention + day)
	removed, err := capture.NewStore(cache, c.Now).Purge()
	if err != nil || len(removed) != 1 {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
}
