package shadow

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func mustWrite(t *testing.T, p, c string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(c), 0o600); err != nil {
		t.Fatal(err)
	}
}

var (
	when      = time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	okTerms   = ProviderTerms{Provider: "prov-a", Retention: "cero días (doc v3)", Training: "no entrena (doc v3)"}
	cleanScan = ScanResult{Send: true}
)

func testPackage(t *testing.T) Package {
	t.Helper()
	p, err := FromFiles("c1", []File{{Path: "diff.patch", Data: []byte("+a\n")}, {Path: "proposal.md", Data: []byte("# p\n")}})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// Spec «Auditoría del piloto»: el registro permite reconstruir cada paquete enviado por su huella.
func TestSendLog_RecordsReconstructablePackage(t *testing.T) {
	l := NewSendLog(filepath.Join(t.TempDir(), "envios.jsonl"))
	p := testPackage(t)
	rec, err := l.Authorize(p, cleanScan, "juez-base", okTerms, when)
	if err != nil {
		t.Fatal(err)
	}
	got, err := l.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("registros = %d", len(got))
	}
	r := got[0]
	if r.Case != "c1" || r.Candidate != "juez-base" || r.Provider != "prov-a" || !r.Date.Equal(when) || r.Package != p.Fingerprint || !reflect.DeepEqual(r, rec) {
		t.Fatalf("registro incompleto: %+v", r)
	}
	// Reconstrucción: las huellas por archivo vuelven a dar la huella del paquete.
	if ReconstructFingerprint(r.Files) != p.Fingerprint {
		t.Fatal("las huellas por archivo no reconstruyen la huella del paquete")
	}
}

// Un positivo del examen previo cancela el envío y no se registra como enviado.
func TestSendLog_ScanHitBlocks(t *testing.T) {
	l := NewSendLog(filepath.Join(t.TempDir(), "envios.jsonl"))
	hit := ScanResult{Send: false, Hits: []Hit{{Path: "diff.patch", Rule: "env-assignment"}}}
	if _, err := l.Authorize(testPackage(t), hit, "juez-base", okTerms, when); err == nil {
		t.Fatal("se autorizó un paquete con positivo del examen")
	}
	if recs, _ := l.Records(); len(recs) != 0 {
		t.Fatalf("se registró un envío cancelado: %+v", recs)
	}
}

// 3.4 y spec «Proveedor que declara retención pero no entrenamiento»: condiciones separadas.
func TestSendLog_ProviderTermsCheckedSeparately(t *testing.T) {
	l := NewSendLog(filepath.Join(t.TempDir(), "envios.jsonl"))
	for name, terms := range map[string]ProviderTerms{
		"sin entrenamiento": {Provider: "prov-a", Retention: "cero días"},
		"sin retención":     {Provider: "prov-a", Training: "no entrena"},
		"sin proveedor":     {Retention: "cero días", Training: "no entrena"},
	} {
		if _, err := l.Authorize(testPackage(t), cleanScan, "juez-base", terms, when); err == nil {
			t.Fatalf("%s: se autorizó el envío", name)
		}
	}
}

// Un paquete cuya huella no coincide con sus archivos no se registra.
func TestSendLog_FingerprintMismatchBlocks(t *testing.T) {
	l := NewSendLog(filepath.Join(t.TempDir(), "envios.jsonl"))
	p := testPackage(t)
	p.Files[0].Data = []byte("+cambiado\n")
	if _, err := l.Authorize(p, cleanScan, "juez-base", okTerms, when); err == nil {
		t.Fatal("se autorizó un paquete con huella desactualizada")
	}
}
