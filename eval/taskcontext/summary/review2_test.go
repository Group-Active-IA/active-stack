package summary_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/summary"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Regresión de la segunda revisión del PR #10: dos comandos distintos que la
// redacción vuelve idénticos no vinculan la captura independiente con la
// ejecución afirmada. Sin material seguro que pruebe la equivalencia, queda sin
// verificar. Credenciales ficticias; no se hacen llamadas HTTP.

func TestReview2RedactionCollisionIsNotExecutionEquivalence(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	copy, err := summary.RedactedOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	const (
		tokenA = "FICTICIOtokenAAAA0123456789abcdef"
		tokenB = "FICTICIOtokenBBBB0123456789abcdef"
	)
	cmdA := "curl -H 'Authorization: Bearer " + tokenA + "' https://example.invalid/data"
	cmdB := "curl -H 'Authorization: Bearer " + tokenB + "' https://example.invalid/data"
	redA, _ := r.RedactText(cmdA)
	redB, _ := r.RedactText(cmdB)
	if redA != redB {
		t.Fatalf("precondición: la redacción debería volver iguales los comandos (%q vs %q)", redA, redB)
	}
	base := harnessRecord(t, f, copy)
	text := agentSummary(t, f, copy)

	tests := []struct {
		name   string
		refCmd string // lo que cita la referencia del resumen
		recCmd string // lo que persiste el capturador (redactado)
		want   summary.State
	}{
		{"credenciales distintas, mismo texto redactado (referencia en claro)", cmdA, redB, summary.StateUnverified},
		{"credenciales distintas, referencia ya redactada", redA, redB, summary.StateUnverified},
		{"comando sin datos redactados sigue vinculando", f.Command, f.Command, summary.StateVerified},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := base
			rec.Command = tt.recCmd
			body := strings.Replace(text, "cmd="+f.Command, "cmd="+tt.refCmd, 1)
			got := corroborate(t, f, body, summary.Evidence{Copy: []byte(copy), Independent: &summary.Independent{Record: rec, Copy: []byte(copy)}})
			if got.State != tt.want {
				t.Fatalf("estado = %s (%v), want %s", got.State, got.Reasons, tt.want)
			}
			if tt.want == summary.StateUnverified && !strings.Contains(strings.Join(got.Reasons, " | "), "redact") {
				t.Errorf("las razones %v no explican que la redacción impide el vínculo", got.Reasons)
			}
		})
	}
}
