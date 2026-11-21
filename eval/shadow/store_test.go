package shadow

import (
	"path/filepath"
	"strings"
	"testing"
)

var closed = Closure{ReferenceDecision: "ref/c1.json", FrozenLabels: "etiquetas@abc123"}

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	repo := t.TempDir()
	dir := t.TempDir()
	s, err := NewStore(dir, repo, "custodio")
	if err != nil {
		t.Fatal(err)
	}
	return s, dir
}

// 4.2: el almacén vive fuera del árbol del repositorio.
func TestNewStore_RejectsDirInsideRepo(t *testing.T) {
	repo := t.TempDir()
	if _, err := NewStore(filepath.Join(repo, "hallazgos"), repo, "custodio"); err == nil {
		t.Fatal("un almacén dentro del árbol del repo debe rechazarse")
	}
	if _, err := NewStore(repo, repo, "custodio"); err == nil {
		t.Fatal("la raíz del repo como almacén debe rechazarse")
	}
}

func TestNewStore_RequiresCustodian(t *testing.T) {
	if _, err := NewStore(t.TempDir(), t.TempDir(), " "); err == nil {
		t.Fatal("sin custodio designado no hay almacén")
	}
}

// Spec «Acceso anticipado»: alguien del flujo de referencia o del etiquetado intenta leer los
// hallazgos de un caso todavía abierto → se deniega y queda registrado.
func TestStore_EarlyAccessDeniedAndLogged(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Save("c1", []byte(`[]`)); err != nil {
		t.Fatal(err)
	}
	for _, who := range []string{"referencia", "etiquetador-1", "tercero"} {
		if _, err := s.Open(who, "c1", Closure{}); err == nil {
			t.Fatalf("%s leyó un caso abierto", who)
		}
	}
	// Ni siquiera el custodio abre un caso sin cerrar, ni otro rol uno cerrado.
	if _, err := s.Open("custodio", "c1", Closure{ReferenceDecision: "ref/c1.json"}); err == nil {
		t.Fatal("el custodio abrió un caso sin etiquetas congeladas")
	}
	if _, err := s.Open("referencia", "c1", closed); err == nil {
		t.Fatal("un rol distinto del custodio abrió un caso cerrado")
	}
	log, err := s.Accesses()
	if err != nil {
		t.Fatal(err)
	}
	if len(log) != 5 {
		t.Fatalf("intentos registrados = %d, quiero 5", len(log))
	}
	for _, a := range log {
		if a.Granted || a.Case != "c1" || a.Reason == "" || a.Time.IsZero() {
			t.Fatalf("registro de intento incompleto o concedido: %+v", a)
		}
	}
}

// Spec «Apertura al cierre del caso»: el custodio abre y la huella coincide.
func TestStore_CustodianOpensClosedCase(t *testing.T) {
	s, _ := newTestStore(t)
	want := []byte(`[{"severidad":"mayor"}]`)
	fp, err := s.Save("c1", want)
	if err != nil {
		t.Fatal(err)
	}
	if fp.Algorithm != "sha256" || len(fp.Value) != 64 {
		t.Fatalf("huella inválida: %+v", fp)
	}
	got, err := s.Open("custodio", "c1", closed)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("contenido = %q", got)
	}
	log, _ := s.Accesses()
	if len(log) != 1 || !log[0].Granted || log[0].Who != "custodio" {
		t.Fatalf("apertura no registrada: %+v", log)
	}
}

// La huella prueba integridad: un archivo alterado no se entrega.
func TestStore_TamperedFindingsRejected(t *testing.T) {
	s, dir := newTestStore(t)
	if _, err := s.Save("c1", []byte(`[]`)); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "c1", findingsFile), `[{"agregado":true}]`)
	_, err := s.Open("custodio", "c1", closed)
	if err == nil || !strings.Contains(err.Error(), "huella") {
		t.Fatalf("hallazgos alterados entregados: %v", err)
	}
}

// Guardar dos veces el mismo caso no pisa lo registrado.
func TestStore_SaveIsWriteOnce(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Save("c1", []byte(`[]`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save("c1", []byte(`[1]`)); err == nil {
		t.Fatal("se sobrescribieron hallazgos ya guardados")
	}
	if _, err := s.Save("../fuera", []byte(`[]`)); err == nil {
		t.Fatal("id de caso con ruta aceptado")
	}
}
