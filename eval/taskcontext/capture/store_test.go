package capture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
)

// Tarea 3.5 (opsx-task-context, D9): ubicación y retención de la copia
// redactada: <os.UserCacheDir>/active-stack/runs/<change>/<id de corrida>/,
// `_sin-change` sin change activo, purga de los vencidos al iniciar cada
// corrida. Reloj inyectable.

type clock struct{ now time.Time }

func (c *clock) Now() time.Time          { return c.now }
func (c *clock) advance(d time.Duration) { c.now = c.now.Add(d) }

var t0 = time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)

const day = 24 * time.Hour

func newStore(t *testing.T) (*capture.Store, *clock, string) {
	t.Helper()
	cache := t.TempDir()
	c := &clock{now: t0}
	return capture.NewStore(cache, c.Now), c, cache
}

func goodRecord() capture.Record {
	return capture.Record{Source: capture.SourceHarness, Redactor: capture.SourceHarness, Coverage: capture.CoverageDemonstrated,
		Client: "harness", Interface: "cli", System: "test", Version: "0", ExitCode: 1, Command: "go test", RunID: "r1"}
}

func TestStore_LocationFollowsDesign(t *testing.T) {
	s, _, cache := newStore(t)
	want := filepath.Join(cache, "active-stack", "runs", "opsx-task-context", "r1")
	if got := s.RunDir("opsx-task-context", "r1"); got != want {
		t.Errorf("RunDir = %q, want %q", got, want)
	}
	// Sin change activo (p. ej. explore) se usa _sin-change.
	wantNo := filepath.Join(cache, "active-stack", "runs", "_sin-change", "r2")
	if got := s.RunDir("", "r2"); got != wantNo {
		t.Errorf("RunDir sin change = %q, want %q", got, wantNo)
	}
}

func TestStore_NothingLandsInTheRepositoryTree(t *testing.T) {
	// La ubicación por defecto sale de os.UserCacheDir: nunca dentro del árbol del repositorio.
	dir, err := capture.DefaultCacheDir()
	if err != nil {
		t.Skipf("sin directorio de caché del usuario: %v", err)
	}
	repo, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(repo, dir)
	if err == nil && !strings.HasPrefix(rel, "..") {
		t.Errorf("el directorio de caché %q está dentro del repositorio %q", dir, repo)
	}
	// Y el store solo escribe bajo su raíz.
	s, _, cache := newStore(t)
	if _, err := s.Save("", "r1", goodRecord(), []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(s.Root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !strings.HasPrefix(p, cache) {
			t.Errorf("archivo fuera de la caché: %s", p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStore_WithinRetentionIsRecoverable(t *testing.T) {
	s, c, _ := newStore(t)
	path, err := s.Save("", "r1", goodRecord(), []byte("fallo\n"))
	if err != nil {
		t.Fatal(err)
	}
	c.advance(6 * day)
	av := s.Lookup("", "r1")
	if av.State != capture.Available || av.Path != path {
		t.Fatalf("a los 6 días debe estar disponible en %q: %+v", path, av)
	}
	if data, err := os.ReadFile(av.Path); err != nil || string(data) != "fallo\n" {
		t.Errorf("la copia no se recupera: %v %q", err, data)
	}
}

func TestStore_ExpiredIsReportedWithDate(t *testing.T) {
	s, c, _ := newStore(t)
	if _, err := s.Save("", "r1", goodRecord(), []byte("fallo\n")); err != nil {
		t.Fatal(err)
	}
	c.advance(8 * day)
	// Todavía no se purgó: el vencido no se ofrece como disponible.
	if av := s.Lookup("", "r1"); av.State != capture.Expired {
		t.Fatalf("a los 8 días el original está vencido: %+v", av)
	}
	if _, err := s.BeginRun("", "r2"); err != nil { // la corrida siguiente purga
		t.Fatal(err)
	}
	av := s.Lookup("", "r1")
	if av.State != capture.Expired {
		t.Fatalf("estado = %v, want vencido", av.State)
	}
	if got := av.ExpiredOn.Format("2006-01-02"); got != "2026-11-20" {
		t.Errorf("vencido desde %s, want 2026-11-20 (captura + 7 días)", got)
	}
	if _, err := os.Stat(s.RunDir("", "r1")); !os.IsNotExist(err) {
		t.Errorf("la purga debe borrar el directorio vencido: %v", err)
	}
	if !strings.Contains(av.Message(), "2026-11-20") || strings.Contains(av.Message(), "nunca capturado") {
		t.Errorf("mensaje de vencido = %q", av.Message())
	}
}

func TestStore_NeverCapturedIsDistinctFromExpired(t *testing.T) {
	s, _, _ := newStore(t)
	av := s.Lookup("", "nunca")
	if av.State != capture.NeverCaptured {
		t.Fatalf("estado = %v, want nunca capturado", av.State)
	}
	if !strings.Contains(av.Message(), "nunca capturado") || strings.Contains(av.Message(), "vencido") {
		t.Errorf("mensaje = %q", av.Message())
	}
	if av.Path != "" {
		t.Error("un original nunca capturado no tiene ruta")
	}
}

func TestStore_ChangeRetentionStartsAtArchive(t *testing.T) {
	s, c, _ := newStore(t)
	if _, err := s.Save("opsx-task-context", "r1", goodRecord(), []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	// Con el change activo no vence, ni a los 30 días.
	c.advance(30 * day)
	if _, err := s.BeginRun("opsx-task-context", "r2"); err != nil {
		t.Fatal(err)
	}
	if av := s.Lookup("opsx-task-context", "r1"); av.State != capture.Available {
		t.Fatalf("con el change activo debe seguir disponible: %+v", av)
	}
	// Al archivar, siete días más.
	if err := s.MarkArchived("opsx-task-context"); err != nil {
		t.Fatal(err)
	}
	c.advance(6 * day)
	if av := s.Lookup("opsx-task-context", "r1"); av.State != capture.Available {
		t.Errorf("a los 6 días del archivo debe estar disponible: %+v", av)
	}
	c.advance(2 * day)
	if _, err := s.BeginRun("", "r3"); err != nil {
		t.Fatal(err)
	}
	av := s.Lookup("opsx-task-context", "r1")
	if av.State != capture.Expired {
		t.Fatalf("a los 8 días del archivo debe estar vencido: %+v", av)
	}
	if got := av.ExpiredOn.Format("2006-01-02"); got != "2026-12-20" {
		t.Errorf("vencido desde %s, want 2026-12-20 (archivo 2026-12-13 + 7 días)", got)
	}
}

func TestStore_SaveRejectsInvalidRecord(t *testing.T) {
	s, _, _ := newStore(t)
	bad := goodRecord()
	bad.Source = "otro"
	if _, err := s.Save("", "r1", bad, []byte("x")); err == nil {
		t.Error("Save debe validar el registro")
	}
}
