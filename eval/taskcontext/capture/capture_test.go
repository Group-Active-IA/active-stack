package capture_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/fixtures"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tarea 3.3 (opsx-task-context, D8): el capturador del arnés ejecuta el
// comando del fixture, recibe stdout y stderr en un mismo flujo ordenado,
// registra el código de salida, redacta en memoria y escribe solo la copia
// redactada, su huella y el registro de captura.

const helperEnv = "TASKCONTEXT_HELPER_FIXTURE"

// TestMain convierte el binario de test en el «comando» del fixture: cuando
// helperEnv está definido, emite la salida del fixture por stdout y stderr y
// sale con su código. Así el capturador ejecuta un proceso real.
func TestMain(m *testing.M) {
	if id := os.Getenv(helperEnv); id != "" {
		f, ok := fixtures.ByID(id)
		if !ok {
			os.Exit(97)
		}
		if err := f.Emit(os.Stdout, os.Stderr); err != nil {
			os.Exit(98)
		}
		os.Exit(f.ExitCode)
	}
	os.Exit(m.Run())
}

func fixtureCmd(t *testing.T, id string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), helperEnv+"="+id)
	return cmd
}

func newRedactor(t *testing.T) *redact.Redactor {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func capturefix(t *testing.T, id string) (capture.Record, string, fixtures.Fixture) {
	t.Helper()
	f, _ := fixtures.ByID(id)
	dir := t.TempDir()
	rec, err := capture.Run(capture.Spec{
		Command: f.Command, Cmd: fixtureCmd(t, id), RunDir: dir, RunID: "run-" + id,
		Redactor: newRedactor(t),
		Meta:     capture.Meta{Client: "harness", Interface: "cli", System: "test", Version: "0"},
	})
	if err != nil {
		t.Fatalf("Run(%s): %v", id, err)
	}
	return rec, dir, f
}

func readCopy(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, capture.CopyFile))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Fallo al medio: la copia coincide línea por línea con la referencia.
func TestRun_LargeFailureCopyMatchesReference(t *testing.T) {
	rec, dir, f := capturefix(t, "fail-middle")
	got := readCopy(t, dir)
	if got != f.Output() {
		t.Fatalf("la copia difiere de la referencia: %d vs %d bytes", len(got), len(f.Output()))
	}
	if rec.ExitCode != 1 {
		t.Errorf("código registrado = %d, want 1", rec.ExitCode)
	}
	if !strings.Contains(got, f.MustSurvive) {
		t.Errorf("el fallo central %q no sobrevivió", f.MustSurvive)
	}
	sum := sha256.Sum256([]byte(got))
	if rec.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("huella del registro = %s, la de la copia es %x", rec.SHA256, sum)
	}
	if rec.Source != capture.SourceHarness || rec.Redactor != capture.SourceHarness {
		t.Errorf("fuente y redactor deben ser del arnés: %+v", rec)
	}
}

// Datos ficticios: ningún archivo de la ruta de corrida contiene un valor sin redactar.
func TestRun_SensitiveValuesNeverReachDisk(t *testing.T) {
	rec, dir, f := capturefix(t, "sensitive")
	got := readCopy(t, dir)
	if got != f.ExpectedRedacted() {
		t.Errorf("la copia no coincide con el original redactado esperado")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range f.Sensitive {
			if bytes.Contains(data, []byte(s.Raw)) {
				t.Errorf("%s contiene el valor sin redactar de %s", e.Name(), s.PatternID)
			}
		}
	}
	if len(entries) != 2 {
		t.Errorf("solo deben quedar la copia redactada y el registro, hay %d archivos", len(entries))
	}
	// El registro cuenta redacciones por categoría y las líneas se conservan.
	if len(rec.Redactions) == 0 {
		t.Error("el registro debe contar las redacciones por categoría")
	}
	if strings.Count(got, "\n") != strings.Count(f.Output(), "\n") {
		t.Error("la redacción cambió la cantidad de líneas")
	}
	total := 0
	for _, n := range rec.Redactions {
		total += n
	}
	if total < len(f.Sensitive) {
		t.Errorf("redacciones contadas = %d, hay al menos %d valores ficticios", total, len(f.Sensitive))
	}
}

// Solo stderr: la línea de stderr está y el código es el del comando.
func TestRun_StderrLinePresentAndExitCodeIsTheCommands(t *testing.T) {
	rec, dir, f := capturefix(t, "stderr-only")
	got := readCopy(t, dir)
	for _, want := range f.Failures {
		if !strings.Contains(got, want+"\n") {
			t.Errorf("la línea de stderr %q no está en la copia", want)
		}
	}
	if rec.ExitCode != f.ExitCode {
		t.Errorf("código = %d, want %d", rec.ExitCode, f.ExitCode)
	}
	// El orden stdout/stderr es el de emisión: stderr va después de las 1.200 líneas de stdout.
	if got != f.Output() {
		t.Error("el flujo combinado no respeta el orden de emisión")
	}
}

func TestRun_ExitCodeZero(t *testing.T) {
	rec, _, _ := capturefix(t, "control-negative")
	if rec.ExitCode != 0 {
		t.Errorf("código = %d, want 0", rec.ExitCode)
	}
}

func TestRun_CommandThatCannotStartIsAnError(t *testing.T) {
	_, err := capture.Run(capture.Spec{
		Command: "no existe", Cmd: exec.Command(filepath.Join(t.TempDir(), "no-such-binary")),
		RunDir: t.TempDir(), RunID: "x", Redactor: newRedactor(t),
	})
	if err == nil {
		t.Error("un comando que no arranca no es un código de salida: debe dar error")
	}
}

// El registro (tarea 3.4) y su regla de estado máximo (D20).
func TestRecord_MaxStatus(t *testing.T) {
	base := capture.Record{Source: capture.SourceHarness, Redactor: capture.SourceHarness, Coverage: capture.CoverageDemonstrated,
		Client: "claude", Interface: "cli", System: "windows", Version: "2.1.319", ExitCode: 1}
	tests := []struct {
		name   string
		mutate func(*capture.Record)
		want   capture.Status
	}{
		{"arnés con cobertura demostrada", func(r *capture.Record) {}, capture.StatusVerifiable},
		{"cliente con cobertura demostrada", func(r *capture.Record) { r.Source = "cliente:aggregated_output" }, capture.StatusVerifiable},
		{"cliente con cobertura no demostrada", func(r *capture.Record) {
			r.Source = "cliente:hook"
			r.Coverage = capture.CoverageNotDemonstrated
		}, capture.StatusUnverified},
		{"agente aunque diga demostrada", func(r *capture.Record) { r.Source = capture.SourceAgent }, capture.StatusUnverified},
		{"arnés con cobertura no demostrada", func(r *capture.Record) { r.Coverage = capture.CoverageNotDemonstrated }, capture.StatusUnverified},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := base
			tt.mutate(&r)
			if err := r.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if got := r.MaxStatus(); got != tt.want {
				t.Errorf("MaxStatus = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRecord_ValidateRejectsMalformed(t *testing.T) {
	good := capture.Record{Source: capture.SourceHarness, Redactor: capture.SourceHarness, Coverage: capture.CoverageDemonstrated,
		Client: "claude", Interface: "cli", System: "windows", Version: "2.1.319"}
	tests := []struct {
		name   string
		mutate func(*capture.Record)
	}{
		{"fuente desconocida", func(r *capture.Record) { r.Source = "alguien" }},
		{"cliente sin fuente nombrada", func(r *capture.Record) { r.Source = "cliente:" }},
		{"cobertura desconocida", func(r *capture.Record) { r.Coverage = "quizás" }},
		{"sin redactor", func(r *capture.Record) { r.Redactor = "" }},
		{"sin cliente", func(r *capture.Record) { r.Client = "" }},
		{"sin sistema", func(r *capture.Record) { r.System = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := good
			tt.mutate(&r)
			if err := r.Validate(); err == nil {
				t.Error("Validate aceptó un registro malformado")
			}
		})
	}
	if err := good.Validate(); err != nil {
		t.Errorf("registro válido rechazado: %v", err)
	}
}
