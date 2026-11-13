package summary_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/fixtures"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/summary"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tareas 2.9 y 2.10 (opsx-task-context, D3 y D20): el evaluador del contrato
// del resumen y los patrones de fallo por herramienta.

var run = summary.RunInfo{
	RunID:   "20261113-example",
	Path:    "<cache>/active-stack/runs/_sin-change/20261113-example/output.redacted.txt",
	Expires: "2026-11-20",
}

func mustFixture(t *testing.T, id string) fixtures.Fixture {
	t.Helper()
	f, ok := fixtures.ByID(id)
	if !ok {
		t.Fatalf("falta el fixture %q", id)
	}
	return f
}

func copySHA(t *testing.T, f fixtures.Fixture) string {
	t.Helper()
	sha, err := summary.CopySHA(f)
	if err != nil {
		t.Fatal(err)
	}
	return sha
}

// El resumen conforme que produce Expected se acepta para cada fixture.
func TestEvaluate_ExpectedSummaryIsAcceptedForEveryFixture(t *testing.T) {
	for _, f := range fixtures.All() {
		t.Run(f.ID, func(t *testing.T) {
			text, err := summary.Expected(f, run)
			if err != nil {
				t.Fatal(err)
			}
			res := summary.Evaluate(f, text, copySHA(t, f))
			if res.Verdict != summary.Accepted {
				t.Errorf("veredicto = %s (%v), want aceptado\n%s", res.Verdict, res.Reasons, text)
			}
			if f.NeverCaptured != res.Partial {
				t.Errorf("Partial = %v, want %v", res.Partial, f.NeverCaptured)
			}
		})
	}
}

// Un resumen escrito a mano, conforme, para el fallo al medio (es el ejemplo
// que documenta docs/contrato-de-resumen.md).
func TestEvaluate_HandWrittenConformingSummary(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	text := strings.Join([]string{
		"exit: 1",
		"failure: --- FAIL: TestNeedle_fail_middle (0.00s)",
		"failure: FAIL\tfixture/pkg\t0.512s",
		"ref: cmd=go test -v ./fixture/pkg/ run=20261113-example path=" + run.Path + " sha256:" + copySHA(t, f) + " expires=2026-11-20",
		"capture: harness · redactor: harness · producer: agent",
	}, "\n") + "\n"
	if res := summary.Evaluate(f, text, copySHA(t, f)); res.Verdict != summary.Accepted {
		t.Errorf("veredicto = %s (%v)", res.Verdict, res.Reasons)
	}
}

// Un incumplimiento por vez: cada uno recibe el veredicto esperado.
func TestEvaluate_Violations(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	conforming, err := summary.Expected(f, run)
	if err != nil {
		t.Fatal(err)
	}
	sha := copySHA(t, f)
	tests := []struct {
		name         string
		mutate       func(string) string
		want         summary.Verdict
		wantReason   string
		wantProducer string
	}{
		{"fallo omitido", func(s string) string {
			return strings.Replace(s, "failure: --- FAIL: TestNeedle_fail_middle (0.00s)\n", "", 1)
		}, summary.Rejected, "omitido", "agent"},
		{"fallo inventado", func(s string) string {
			return strings.Replace(s, "exit: 1\n", "exit: 1\nfailure: --- FAIL: TestNoExiste (0.00s)\n", 1)
		}, summary.Rejected, "inventado", "agent"},
		{"código de salida distinto", func(s string) string {
			return strings.Replace(s, "exit: 1", "exit: 0", 1)
		}, summary.Rejected, "código de salida", "agent"},
		{"huella distinta", func(s string) string {
			return strings.Replace(s, "sha256:"+sha, "sha256:"+strings.Repeat("0", 64), 1)
		}, summary.Rejected, "huella", "agent"},
		{"referencia incompleta (sin vencimiento)", func(s string) string {
			return strings.Replace(s, " expires=2026-11-20", "", 1)
		}, summary.Insufficient, "referencia", "agent"},
		{"referencia incompleta (sin ruta)", func(s string) string {
			return strings.Replace(s, " path="+run.Path, "", 1)
		}, summary.Insufficient, "referencia", "agent"},
		{"sin referencia", func(s string) string {
			var out []string
			for _, l := range strings.Split(s, "\n") {
				if !strings.HasPrefix(l, "ref:") {
					out = append(out, l)
				}
			}
			return strings.Join(out, "\n")
		}, summary.Insufficient, "referencia", "agent"},
		{"capturador sin declarar", func(s string) string {
			return strings.Replace(s, "capture: harness · redactor: harness · producer: agent", "capture: · redactor: harness · producer: agent", 1)
		}, summary.Insufficient, "capturador", "agent"},
		{"productor sin declarar", func(s string) string {
			return strings.Replace(s, " · producer: agent", "", 1)
		}, summary.Insufficient, "productor", "no declarado"},
		{"sin línea de captura", func(s string) string {
			return strings.Replace(s, "capture: harness · redactor: harness · producer: agent\n", "", 1)
		}, summary.Insufficient, "capturador", "no declarado"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutated := tt.mutate(conforming)
			if mutated == conforming {
				t.Fatal("la mutación no cambió el resumen")
			}
			res := summary.Evaluate(f, mutated, sha)
			if res.Verdict != tt.want {
				t.Errorf("veredicto = %s (%v), want %s", res.Verdict, res.Reasons, tt.want)
			}
			if !strings.Contains(strings.Join(res.Reasons, " | "), tt.wantReason) {
				t.Errorf("las razones %v no mencionan %q", res.Reasons, tt.wantReason)
			}
			if res.Producer != tt.wantProducer {
				t.Errorf("productor registrado = %q, want %q", res.Producer, tt.wantProducer)
			}
		})
	}
}

func TestEvaluate_RejectionRecordsDeclaredProducer(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	text, _ := summary.Expected(f, run)
	text = strings.Replace(text, "producer: agent", "producer: subagent-explore", 1)
	text = strings.Replace(text, "failure: FAIL\tfixture/pkg\t0.512s\n", "", 1)
	res := summary.Evaluate(f, text, copySHA(t, f))
	if res.Verdict != summary.Rejected || res.Producer != "subagent-explore" {
		t.Errorf("rechazo = %+v, want rechazado con productor subagent-explore", res)
	}
}

// Un capturador del agente nunca supera «sin verificar» (D20): el resumen debe decirlo.
func TestEvaluate_AgentCaptureMustDeclareUnverified(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	text, _ := summary.Expected(f, run)
	agent := strings.Replace(text, "capture: harness · redactor: harness", "capture: agent · redactor: agent", 1)
	if res := summary.Evaluate(f, agent, copySHA(t, f)); res.Verdict != summary.Insufficient || !strings.Contains(strings.Join(res.Reasons, " "), "sin verificar") {
		t.Errorf("captura del agente sin `status: unverified`: %+v", res)
	}
	agent += "status: unverified — no independent capture\n"
	if res := summary.Evaluate(f, agent, copySHA(t, f)); res.Verdict != summary.Accepted {
		t.Errorf("captura del agente que declara unverified debe aceptarse como contenido: %+v", res)
	}
}

// 8.5 se prepara acá: un valor sensible sin redactar en el resumen lo rechaza.
func TestEvaluate_RawSensitiveValueIsRejected(t *testing.T) {
	f := mustFixture(t, "sensitive")
	text, _ := summary.Expected(f, run)
	leaked := text + "note: DB_PASSWORD=FICTICIO-clave-0001\n"
	res := summary.Evaluate(f, leaked, copySHA(t, f))
	if res.Verdict != summary.Rejected || !strings.Contains(strings.Join(res.Reasons, " "), "sensible") {
		t.Errorf("resumen con valor sin redactar: %+v", res)
	}
}

// 2.8: original nunca capturado.
func TestEvaluate_NeverCaptured(t *testing.T) {
	f := mustFixture(t, "never-captured")
	good, _ := summary.Expected(f, run)
	for _, want := range []string{"partial — unverified: never captured", "exit: 1"} {
		if !strings.Contains(good, want) {
			t.Fatalf("el resumen esperado no contiene %q:\n%s", want, good)
		}
	}
	if strings.Contains(good, "ref:") || strings.Contains(good, "sha256") || strings.Contains(good, "path=") {
		t.Errorf("un original nunca capturado no lleva ruta ni huella:\n%s", good)
	}
	tests := []struct {
		name   string
		text   string
		want   summary.Verdict
		reason string
	}{
		{"sin marca de parcial", "exit: 1\n", summary.Insufficient, "parcial"},
		{"parcial sin motivo", "exit: 1\npartial — unverified:\n", summary.Insufficient, "motivo"},
		{"parcial con una referencia inventada", good + "ref: cmd=x run=y path=z sha256:00 expires=2026-01-01\n", summary.Insufficient, "ruta ni huella"},
		{"motivo fuera de la lista", "exit: 1\npartial — unverified: me olvidé\n", summary.Insufficient, "motivo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := summary.Evaluate(f, tt.text, "")
			if res.Verdict != tt.want || !strings.Contains(strings.Join(res.Reasons, " "), tt.reason) {
				t.Errorf("= %+v, want %s con %q", res, tt.want, tt.reason)
			}
		})
	}
	// Cada motivo permitido se acepta.
	for _, reason := range []string{"never captured", "copy missing", "expired on 2026-11-20", "no persistent disk"} {
		text := "exit: 1\npartial — unverified: " + reason + "\n"
		if res := summary.Evaluate(f, text, ""); res.Verdict != summary.Accepted || !res.Partial {
			t.Errorf("motivo %q: %+v", reason, res)
		}
	}
}

// 2.10: los patrones de fallo por herramienta salen del catálogo.
func TestFindFailures_PerToolPatternsFromCatalog(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"fail-start", "fail-middle", "fail-end", "scattered", "control-negative", "sensitive", "noise-exit0", "real-among-noise", "stderr-only"} {
		t.Run(id, func(t *testing.T) {
			f := mustFixture(t, id)
			got, byExit := summary.FindFailures(cat.TaskContext, f.Tool, f.Output(), f.ExitCode)
			if byExit {
				t.Error("estas herramientas tienen patrones declarados")
			}
			if strings.Join(got, "\n") != strings.Join(f.Failures, "\n") {
				t.Errorf("fallos identificados %q, esperados %q", got, f.Failures)
			}
		})
	}
}

func TestFindFailures_ToolWithoutPatternsUsesExitCodeOnly(t *testing.T) {
	cat, _ := catalog.Load()
	out := "building...\nerror: something odd\nFAIL\tnot-a-go-pattern\n"
	got, byExit := summary.FindFailures(cat.TaskContext, "make", out, 2)
	if !byExit || len(got) != 0 {
		t.Errorf("una herramienta sin patrones se evalúa solo por código de salida: %q %v", got, byExit)
	}
	// El resumen debe decirlo.
	f := fixtures.Fixture{ID: "make-fixture", Command: "make all", Tool: "make", ExitCode: 2,
		Lines: []fixtures.Line{{Text: "building..."}, {Text: "error: something odd"}}}
	text, err := summary.Expected(f, run)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "failures identified by exit code only") {
		t.Errorf("el resumen esperado debe decir que se identifica solo por código de salida:\n%s", text)
	}
	sha, _ := summary.CopySHA(f)
	if res := summary.Evaluate(f, text, sha); res.Verdict != summary.Accepted {
		t.Errorf("= %+v", res)
	}
	without := strings.Replace(text, "failures identified by exit code only\n", "", 1)
	if res := summary.Evaluate(f, without, sha); res.Verdict != summary.Insufficient {
		t.Errorf("sin la frase debe ser insuficiente: %+v", res)
	}
}

// Con código cero, una línea con forma de error no es un fallo.
func TestFindFailures_ExitZeroIsNeverAFailure(t *testing.T) {
	cat, _ := catalog.Load()
	out := "logger.go:42:7: error: simulated retry (recovered)\n--- FAIL: TestLooksReal (0.00s)\n"
	got, _ := summary.FindFailures(cat.TaskContext, "go build", out, 0)
	if len(got) != 0 {
		t.Errorf("con exit 0 no hay fallos: %q", got)
	}
	got, _ = summary.FindFailures(cat.TaskContext, "go build", out, 1)
	if len(got) != 1 || !strings.HasPrefix(got[0], "logger.go:42:7:") {
		t.Errorf("con exit 1 el patrón de go build identifica la línea archivo:línea:columna: %q", got)
	}
}

// --- Corridas reales del grupo 8 (13/11): casos que el evaluador aceptaba mal ---

// leakedCopy es la copia del arnés con el valor de un patrón sin redactar, como
// la que dejó el agente en t3-sensitive-a (account-uuid).
func leakedCopy(t *testing.T, f fixtures.Fixture, patternID string) string {
	t.Helper()
	redacted, err := summary.RedactedOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range f.Sensitive {
		if s.PatternID == patternID {
			out := strings.Replace(redacted, s.Redacted, s.Line, 1)
			if out == redacted {
				t.Fatalf("no se pudo deshacer la redacción de %s", patternID)
			}
			return out
		}
	}
	t.Fatalf("el fixture no tiene el patrón %s", patternID)
	return ""
}

func shaOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// agentSummary es el resumen conforme de una captura del agente cuya huella es la de copy.
func agentSummary(t *testing.T, f fixtures.Fixture, copy string) string {
	t.Helper()
	text, err := summary.Expected(f, run)
	if err != nil {
		t.Fatal(err)
	}
	harness := copySHA(t, f)
	text = strings.Replace(text, "sha256:"+harness, "sha256:"+shaOf(copy), 1)
	text = strings.Replace(text, "capture: harness · redactor: harness", "capture: agent · redactor: agent", 1)
	return text + "status: unverified — no independent capture\n"
}

// 8.5 / 8.2 (t3-sensitive-a): la copia del agente conserva un valor ficticio sin
// redactar; resumen y huella son coherentes con esa copia. Antes se aceptaba.
func TestEvaluateCopy_RejectsRawSensitiveValueInTheCopy(t *testing.T) {
	f := mustFixture(t, "sensitive")
	copy := leakedCopy(t, f, "account-uuid")
	text := agentSummary(t, f, copy)
	res := summary.EvaluateCopy(f, text, []byte(copy))
	if res.Verdict != summary.Rejected {
		t.Fatalf("veredicto = %s (%v), want rechazado", res.Verdict, res.Reasons)
	}
	reasons := strings.Join(res.Reasons, " | ")
	if !strings.Contains(reasons, "account-uuid") || !strings.Contains(reasons, "copia") {
		t.Errorf("las razones deben nombrar el patrón filtrado en la copia: %v", res.Reasons)
	}
}

// 8.2 (t3-sensitive-b): la copia del agente no es la del arnés (otra redacción);
// la huella del resumen coincide con esa copia. Antes se aceptaba.
func TestEvaluateCopy_RejectsCopyThatDiffersFromTheHarnessCopy(t *testing.T) {
	f := mustFixture(t, "sensitive")
	harnessCopy, err := summary.RedactedOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	copy := strings.ReplaceAll(harnessCopy, "[REDACTED:", "[OMITIDO:")
	text := agentSummary(t, f, copy)
	res := summary.EvaluateCopy(f, text, []byte(copy))
	if res.Verdict != summary.Rejected || !strings.Contains(strings.Join(res.Reasons, " "), "huella") {
		t.Errorf("copia distinta de la del arnés: %+v, want rechazado por huella", res)
	}
}

// La copia del agente idéntica a la del arnés se evalúa como antes.
func TestEvaluateCopy_AcceptsTheHarnessCopy(t *testing.T) {
	for _, id := range []string{"fail-middle", "sensitive", "noise-exit0"} {
		t.Run(id, func(t *testing.T) {
			f := mustFixture(t, id)
			copy, err := summary.RedactedOutput(f)
			if err != nil {
				t.Fatal(err)
			}
			text := agentSummary(t, f, copy)
			if res := summary.EvaluateCopy(f, text, []byte(copy)); res.Verdict != summary.Accepted {
				t.Errorf("= %+v", res)
			}
			// La huella del resumen debe ser la de la copia entregada.
			bad := strings.Replace(text, "sha256:"+shaOf(copy), "sha256:"+strings.Repeat("a", 64), 1)
			if res := summary.EvaluateCopy(f, bad, []byte(copy)); res.Verdict != summary.Rejected {
				t.Errorf("huella del resumen distinta de la copia: %+v", res)
			}
		})
	}
}

// 8.3 (t5-ps-*): PowerShell escribe el archivo con BOM y CRLF. Las líneas fijas
// son las mismas; el fin de línea no es contenido.
func TestEvaluate_SummaryFileWithBOMAndCRLF(t *testing.T) {
	for _, id := range []string{"fail-middle", "stderr-only"} {
		t.Run(id, func(t *testing.T) {
			f := mustFixture(t, id)
			text, _ := summary.Expected(f, run)
			ps := "\ufeff" + strings.ReplaceAll(text, "\n", "\r\n")
			if res := summary.Evaluate(f, ps, copySHA(t, f)); res.Verdict != summary.Accepted {
				t.Errorf("= %+v", res)
			}
		})
	}
}

// 8.2 (noise-exit0, control-negative): sin fallos, la línea fija es exactamente
// `failure: none` (spec «Ruido que imita el formato de un fallo»).
func TestEvaluate_NoFailureFixedLine(t *testing.T) {
	for _, id := range []string{"noise-exit0", "control-negative"} {
		t.Run(id, func(t *testing.T) {
			f := mustFixture(t, id)
			text, _ := summary.Expected(f, run)
			if !strings.Contains(text, "\n"+summary.NoFailureLine+"\n") {
				t.Fatalf("el resumen esperado sin fallos debe llevar %q:\n%s", summary.NoFailureLine, text)
			}
			if res := summary.Evaluate(f, text, copySHA(t, f)); res.Verdict != summary.Accepted {
				t.Errorf("con `failure: none`: %+v", res)
			}
			// Sin ninguna línea failure: tampoco declara un fallo.
			without := strings.Replace(text, summary.NoFailureLine+"\n", "", 1)
			if res := summary.Evaluate(f, without, copySHA(t, f)); res.Verdict != summary.Accepted {
				t.Errorf("sin línea failure: %+v", res)
			}
			// Lo que dejaron las corridas reales: `none` con un comentario es un fallo inventado.
			noted := strings.Replace(text, summary.NoFailureLine, `failure: none (no "--- FAIL" lines in the capture)`, 1)
			if res := summary.Evaluate(f, noted, copySHA(t, f)); res.Verdict != summary.Rejected {
				t.Errorf("`failure: none (…)`: %+v, want rechazado", res)
			}
		})
	}
	// `failure: none` en una salida con fallos los omite.
	f := mustFixture(t, "fail-middle")
	text, _ := summary.Expected(f, run)
	var keep []string
	for _, l := range strings.Split(text, "\n") {
		if !strings.HasPrefix(l, "failure: ") {
			keep = append(keep, l)
		}
	}
	none := strings.Replace(strings.Join(keep, "\n"), "exit: 1\n", "exit: 1\n"+summary.NoFailureLine+"\n", 1)
	if res := summary.Evaluate(f, none, copySHA(t, f)); res.Verdict != summary.Rejected || !strings.Contains(strings.Join(res.Reasons, " "), "omitido") {
		t.Errorf("`failure: none` con fallos: %+v", res)
	}
	// Junto a un fallo, `failure: none` se contradice.
	both := strings.Replace(text, "exit: 1\n", "exit: 1\n"+summary.NoFailureLine+"\n", 1)
	if res := summary.Evaluate(f, both, copySHA(t, f)); res.Verdict != summary.Rejected {
		t.Errorf("`failure: none` junto a fallos: %+v", res)
	}
}

// 8.2 (t3-scattered-b): la línea failure: es la capturada, sin agregados.
func TestEvaluate_FailureLineWithAppendedNoteIsNotTheCapturedLine(t *testing.T) {
	f := mustFixture(t, "scattered")
	text, _ := summary.Expected(f, run)
	noted := strings.Replace(text, "failure: --- FAIL: TestAlpha (0.00s)\n", "failure: --- FAIL: TestAlpha (0.00s)   [output line 100]\n", 1)
	if res := summary.Evaluate(f, noted, copySHA(t, f)); res.Verdict != summary.Rejected {
		t.Errorf("= %+v, want rechazado", res)
	}
}

// --- 8.4: corroboración de un resultado declarado (D20) ---

func harnessRecord(t *testing.T, f fixtures.Fixture, copy string) capture.Record {
	t.Helper()
	return capture.Record{
		Command: f.Command, RunID: "h-" + f.ID, Source: capture.SourceHarness, Redactor: capture.SourceHarness,
		Client: "arnés", Interface: "cli", System: "windows", Version: "test",
		ExitCode: f.ExitCode, Coverage: capture.CoverageDemonstrated, SHA256: shaOf(copy),
	}
}

func corroborate(t *testing.T, f fixtures.Fixture, text string, ev summary.Evidence) summary.Corroboration {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	return summary.Corroborate(cat.TaskContext, f.Tool, text, ev)
}

func TestCorroborate_D20States(t *testing.T) {
	f := mustFixture(t, "fail-middle")
	copy, err := summary.RedactedOutput(f)
	if err != nil {
		t.Fatal(err)
	}
	agentText := agentSummary(t, f, copy)
	harnessText, _ := summary.Expected(f, run)
	harnessText = strings.Replace(harnessText, "producer: agent", "producer: worker-3", 1)
	omitted := strings.Replace(agentText, "failure: --- FAIL: TestNeedle_fail_middle (0.00s)\n", "", 1)
	omitted = strings.Replace(omitted, "failure: FAIL\tfixture/pkg\t0.512s\n", "", 1)
	rec := harnessRecord(t, f, copy)
	notProven := rec
	notProven.Source, notProven.Coverage = "cliente:PostToolUse", capture.CoverageNotDemonstrated
	agentRec := rec
	agentRec.Source, agentRec.Redactor = capture.SourceAgent, capture.SourceAgent
	wrongRec := rec
	wrongRec.SHA256 = strings.Repeat("f", 64)
	indep := func(r capture.Record, c string) *summary.Independent {
		return &summary.Independent{Record: r, Copy: []byte(c)}
	}
	rerunWithoutNeedle := strings.Replace(copy, "--- FAIL: TestNeedle_fail_middle (0.00s)", "", 1)

	tests := []struct {
		name   string
		text   string
		ev     summary.Evidence
		want   summary.State
		reason string
	}{
		// A: archivo, huella y resumen del mismo agente, coherentes.
		{"A fabricado y coherente", agentText, summary.Evidence{Copy: []byte(copy)}, summary.StateUnverified, "procedencia"},
		{"A con registro del propio agente", agentText, summary.Evidence{Copy: []byte(copy), Independent: indep(agentRec, copy)}, summary.StateUnverified, "procedencia"},
		{"fuente del cliente sin cobertura demostrada", agentText, summary.Evidence{Copy: []byte(copy), Independent: indep(notProven, copy)}, summary.StateUnverified, "procedencia"},
		// B: el resumen omite un fallo que la copia registra.
		{"B fallo omitido, sin captura independiente", omitted, summary.Evidence{Copy: []byte(copy)}, summary.StateRejected, "omitido"},
		{"B fallo omitido, con captura del arnés", omitted, summary.Evidence{Copy: []byte(copy), Independent: indep(rec, copy)}, summary.StateRejected, "omitido"},
		{"B fallo omitido que aparece al re-ejecutar", strings.Replace(omitted, "sha256:"+shaOf(copy), "sha256:"+shaOf(rerunWithoutNeedle), 1) + "failure: FAIL\tfixture/pkg\t0.512s\n", summary.Evidence{Copy: []byte(rerunWithoutNeedle), Independent: indep(rec, copy)}, summary.StateRejected, "omitido"},
		// C: captura independiente con cobertura demostrada, huella y fallos coinciden.
		{"C captura del arnés", harnessText, summary.Evidence{Copy: []byte(copy), Independent: indep(rec, copy)}, summary.StateVerified, ""},
		{"re-ejecución por otra parte", agentText, summary.Evidence{Copy: []byte(copy), Independent: indep(rec, copy)}, summary.StateVerified, ""},
		// Huella o código distintos.
		{"huella distinta", strings.Replace(harnessText, "sha256:"+shaOf(copy), "sha256:"+strings.Repeat("0", 64), 1), summary.Evidence{Copy: []byte(copy), Independent: indep(rec, copy)}, summary.StateRejected, "huella"},
		{"código distinto al del registro", strings.Replace(harnessText, "exit: 1", "exit: 0", 1), summary.Evidence{Copy: []byte(copy), Independent: indep(rec, copy)}, summary.StateRejected, "código"},
		{"registro que no es de esa copia", harnessText, summary.Evidence{Copy: []byte(copy), Independent: indep(wrongRec, copy)}, summary.StateUnverified, "registro"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := corroborate(t, f, tt.text, tt.ev)
			if got.State != tt.want {
				t.Fatalf("estado = %s (%v), want %s", got.State, got.Reasons, tt.want)
			}
			if tt.reason != "" && !strings.Contains(strings.Join(got.Reasons, " | "), tt.reason) {
				t.Errorf("las razones %v no mencionan %q", got.Reasons, tt.reason)
			}
			if got.State == summary.StateRejected && got.Producer == "" {
				t.Error("el rechazo registra el productor declarado")
			}
		})
	}
}
