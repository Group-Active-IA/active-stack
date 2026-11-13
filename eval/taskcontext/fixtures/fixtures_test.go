package fixtures_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/fixtures"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tareas 2.1 a 2.8 (opsx-task-context, D4): comandos deterministas cuya salida
// el arnés conoce de antemano. Esa salida es la referencia contra la que se
// comparan la captura y el resumen.

func threshold(t *testing.T) catalog.OutputThreshold {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cat.TaskContext.OutputThreshold
}

func countLines(s string) int { return strings.Count(s, "\n") }

// 2.1: tres fixtures de salida grande con el fallo al inicio, al medio y al
// final. Tienen miles de líneas y superan 10.000 caracteres: el cliente los
// recortaría sin captura.
func TestLargeFailureFixtures(t *testing.T) {
	th := threshold(t)
	tests := []struct {
		id       string
		wantLine func(total int) (lo, hi int) // rango (1-based) donde debe estar la aguja
	}{
		{"fail-start", func(n int) (int, int) { return 1, 10 }},
		{"fail-middle", func(n int) (int, int) { return n/2 - 20, n/2 + 20 }},
		{"fail-end", func(n int) (int, int) { return n - 5, n }},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			f, ok := fixtures.ByID(tt.id)
			if !ok {
				t.Fatalf("falta el fixture %q", tt.id)
			}
			out := f.Output()
			lines, chars := countLines(out), utf8.RuneCountInString(out)
			if lines < 2000 {
				t.Errorf("tiene %d líneas, deben ser miles", lines)
			}
			if lines <= th.Lines || chars <= th.Characters {
				t.Errorf("no supera el par de D3 (%d líneas, %d caracteres)", lines, chars)
			}
			if f.ExitCode == 0 || f.Command == "" || f.MustSurvive == "" || len(f.Failures) == 0 {
				t.Errorf("el fixture debe declarar comando, código de salida distinto de cero, fallos y el que sobrevive: %+v", f.ID)
			}
			// El fallo que debe sobrevivir está en la posición declarada.
			lo, hi := tt.wantLine(lines)
			pos := 0
			for i, l := range strings.Split(out, "\n") {
				if l == f.MustSurvive {
					pos = i + 1
				}
			}
			if pos < lo || pos > hi {
				t.Errorf("la línea %q está en la %d, esperada entre %d y %d", f.MustSurvive, pos, lo, hi)
			}
		})
	}
}

// 2.2: fallos dispersos, cada uno declarado por separado.
func TestScatteredFailures(t *testing.T) {
	f, ok := fixtures.ByID("scattered")
	if !ok {
		t.Fatal("falta el fixture scattered")
	}
	if len(f.Failures) < 3 {
		t.Fatalf("declara %d fallos, deben ser al menos tres por separado", len(f.Failures))
	}
	out := strings.Split(f.Output(), "\n")
	var positions []int
	for _, want := range f.Failures {
		found := -1
		for i, l := range out {
			if l == want {
				found = i + 1
				break
			}
		}
		if found < 0 {
			t.Fatalf("el fallo declarado %q no está en la salida", want)
		}
		positions = append(positions, found)
	}
	// Al menos tres posiciones distintas y separadas por cientos de líneas.
	distant := 0
	for i := 1; i < len(positions); i++ {
		if d := positions[i] - positions[i-1]; d > 100 || d < -100 {
			distant++
		}
	}
	if distant < 2 {
		t.Errorf("los fallos no están dispersos en más de una posición: %v", positions)
	}
}

// 2.3: control negativo: verbosa, repetitiva y con código cero.
func TestNegativeControl(t *testing.T) {
	f, ok := fixtures.ByID("control-negative")
	if !ok {
		t.Fatal("falta el fixture control-negative")
	}
	if f.ExitCode != 0 || len(f.Failures) != 0 || f.MustSurvive != "" {
		t.Errorf("el control negativo no declara fallos y sale con 0: %+v", f.ID)
	}
	out := f.Output()
	if countLines(out) <= threshold(t).Lines {
		t.Error("debe ser verboso: más líneas que el umbral")
	}
	if !strings.Contains(out, "ok  \t") {
		t.Error("debe terminar como una corrida exitosa")
	}
}

// 2.4: datos sensibles ficticios, uno por patrón del catálogo.
func TestSensitiveFixture(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := fixtures.ByID("sensitive")
	if !ok {
		t.Fatal("falta el fixture sensitive")
	}
	// Un valor por cada patrón de la lista.
	have := map[string]bool{}
	for _, s := range f.Sensitive {
		have[s.PatternID] = true
	}
	for _, p := range cat.TaskContext.SensitivePatterns {
		if !have[p.ID] {
			t.Errorf("el fixture no trae un valor ficticio para el patrón %q", p.ID)
		}
	}
	out := f.Output()
	for _, s := range f.Sensitive {
		// Ficticio: lleva FICTICIO, o es el UUID nulo reservado cuando el formato no admite letras.
		if !strings.Contains(s.Raw, "FICTICIO") && !strings.HasPrefix(s.Raw, "00000000-0000-4000-8000-") {
			t.Errorf("%s: el valor %q no está marcado como ficticio", s.PatternID, s.Raw)
		}
		if !strings.Contains(out, s.Raw) {
			t.Errorf("%s: el valor no figura en la salida del fixture", s.PatternID)
		}
		// Cada valor coincide con su patrón: redactando solo esa línea, el valor desaparece y el conteo del patrón sube.
		got, counts := r.RedactText(s.Line + "\n")
		if strings.Contains(got, s.Raw) || counts[s.PatternID] == 0 {
			t.Errorf("%s: el valor no coincide con su patrón (salida %q, conteos %v)", s.PatternID, got, counts)
		}
	}
	// El original redactado esperado conserva cada línea en su posición.
	expected := f.ExpectedRedacted()
	if countLines(expected) != countLines(out) {
		t.Fatalf("el redactado esperado cambió la cantidad de líneas: %d vs %d", countLines(expected), countLines(out))
	}
	got, _ := r.RedactText(out)
	if got != expected {
		t.Errorf("el redactor no produce el original redactado esperado:\n%s", firstDiff(expected, got))
	}
	for _, s := range f.Sensitive {
		if strings.Contains(expected, s.Raw) {
			t.Errorf("%s: el valor sobrevive en el redactado esperado", s.PatternID)
		}
	}
	// Con un fallo real en la misma salida.
	if len(f.Failures) == 0 || !strings.Contains(expected, f.Failures[0]) {
		t.Error("la salida debe traer un fallo real que sobreviva a la redacción")
	}
}

func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) && i < len(g); i++ {
		if w[i] != g[i] {
			return "línea " + strconv.Itoa(i+1) + "\n  esperada: " + w[i] + "\n  obtenida: " + g[i]
		}
	}
	return "difieren en la cantidad de líneas"
}

// 2.5: ruido con forma de error y código cero: sin fallos.
func TestNoiseWithExitZero(t *testing.T) {
	f, ok := fixtures.ByID("noise-exit0")
	if !ok {
		t.Fatal("falta el fixture noise-exit0")
	}
	if f.ExitCode != 0 || len(f.Failures) != 0 {
		t.Fatalf("el ruido sale con 0 y no declara fallos: %+v", f.ID)
	}
	out := f.Output()
	// Imita el formato de un error aunque no lo sea.
	if !strings.Contains(out, "error") || !strings.Contains(out, "FAIL") {
		t.Error("el ruido debe contener la forma de un error")
	}
}

// 2.6: un fallo real rodeado de ruido que lo imita: exactamente uno.
func TestRealFailureAmongNoise(t *testing.T) {
	f, ok := fixtures.ByID("real-among-noise")
	if !ok {
		t.Fatal("falta el fixture real-among-noise")
	}
	if len(f.Failures) != 1 || f.ExitCode == 0 {
		t.Fatalf("declara exactamente un fallo y código distinto de cero: %+v", f.ID)
	}
	out := f.Output()
	if strings.Count(out, "error") < 3 {
		t.Error("debe haber ruido que imita errores alrededor")
	}
	if !strings.Contains(out, f.Failures[0]) {
		t.Error("el fallo real no está en la salida")
	}
}

// 2.7: el fallo se escribe solo por stderr, con stdout verboso sin fallos.
func TestStderrOnlyFailure(t *testing.T) {
	f, ok := fixtures.ByID("stderr-only")
	if !ok {
		t.Fatal("falta el fixture stderr-only")
	}
	if f.ExitCode == 0 || len(f.Failures) == 0 {
		t.Fatalf("falla con código distinto de cero: %+v", f.ID)
	}
	var stdoutLines int
	stderr := map[string]bool{}
	for _, l := range f.Lines {
		if l.Stream == fixtures.Stderr {
			stderr[l.Text] = true
		} else {
			stdoutLines++
			for _, want := range f.Failures {
				if l.Text == want {
					t.Errorf("el fallo %q aparece por stdout: debe ir solo por stderr", want)
				}
			}
		}
	}
	if stdoutLines < 500 {
		t.Errorf("stdout debe ser verboso, tiene %d líneas", stdoutLines)
	}
	for _, want := range f.Failures {
		if !stderr[want] {
			t.Errorf("el fallo %q no está en stderr", want)
		}
	}
}

// 2.8: original nunca capturado.
func TestNeverCapturedFixture(t *testing.T) {
	f, ok := fixtures.ByID("never-captured")
	if !ok {
		t.Fatal("falta el fixture never-captured")
	}
	if !f.NeverCaptured || f.Command == "" {
		t.Errorf("debe declarar un comando ya corrido sin captura: %+v", f.ID)
	}
}

func TestAllHaveUniqueIDsAndCommands(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range fixtures.All() {
		if f.ID == "" || seen[f.ID] {
			t.Errorf("id vacío o repetido: %q", f.ID)
		}
		seen[f.ID] = true
		if f.Command == "" {
			t.Errorf("%s: sin comando declarado", f.ID)
		}
	}
	if len(seen) != 10 {
		t.Errorf("hay %d fixtures, se esperaban 10 (2.1 x3, 2.2 a 2.8)", len(seen))
	}
}
