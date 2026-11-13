package summary_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/summary"
)

// Regresión de la revisión del PR #10: la captura independiente debe
// corresponder a la ejecución que el resumen afirma (mismo comando, o una
// re-ejecución del mismo comando); una salida igual no alcanza.

func TestReviewUnrelatedRunCannotCorroborateReportedRun(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	copy, err := summary.RedactedOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	text := agentSummary(t, f, copy)
	base := harnessRecord(t, f, copy)

	tests := []struct {
		name   string
		mutate func(r *capture.Record)
		text   string
		want   summary.State
		reason string
	}{
		{"comando ajeno con la misma salida", func(r *capture.Record) {
			r.Command, r.RunID, r.CopyPath = "printf ok", "unrelated-run", "/another"
		}, text, summary.StateUnverified, "comando"},
		{"registro sin comando", func(r *capture.Record) { r.Command = "" }, text, summary.StateUnverified, "comando"},
		{"misma ejecución", func(r *capture.Record) {}, text, summary.StateVerified, ""},
		{"re-ejecución del mismo comando con otra corrida", func(r *capture.Record) {
			r.RunID, r.CopyPath = "rerun-by-other", "/elsewhere"
		}, text, summary.StateVerified, ""},
		{"referencia sin comando", func(r *capture.Record) {}, strings.Replace(text, "cmd="+f.Command, "cmd=", 1), summary.StateUnverified, "comando"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := base
			tt.mutate(&rec)
			got := corroborate(t, f, tt.text, summary.Evidence{Copy: []byte(copy), Independent: &summary.Independent{Record: rec, Copy: []byte(copy)}})
			if got.State != tt.want {
				t.Fatalf("estado = %s (%v), want %s", got.State, got.Reasons, tt.want)
			}
			if tt.reason != "" && !strings.Contains(strings.Join(got.Reasons, " | "), tt.reason) {
				t.Errorf("las razones %v no mencionan %q", got.Reasons, tt.reason)
			}
		})
	}

	// Segunda revisión: la igualdad tras redactar no prueba la misma ejecución
	// (la redacción borra justo lo que distingue credencial o cuenta). Antes este
	// caso daba verificado; ahora queda sin verificar.
	t.Run("comando con secreto redactado no se da por vinculado", func(t *testing.T) {
		const secret = "Bearer FICTICIO0123456789abcdefghij"
		withSecret := strings.Replace(text, "cmd="+f.Command, "cmd="+f.Command+" -H "+secret, 1)
		rec := base
		cat, err := catalog.Load()
		if err != nil {
			t.Fatal(err)
		}
		r, err := redact.New(cat.TaskContext.SensitivePatterns)
		if err != nil {
			t.Fatal(err)
		}
		rec.Command, _ = r.RedactText(f.Command + " -H " + secret) // lo que persiste el capturador
		got := corroborate(t, f, withSecret, summary.Evidence{Copy: []byte(copy), Independent: &summary.Independent{Record: rec, Copy: []byte(copy)}})
		if got.State != summary.StateUnverified {
			t.Errorf("estado = %s (%v), want sin verificar", got.State, got.Reasons)
		}
	})
}
