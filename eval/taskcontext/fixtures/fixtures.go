// Package fixtures define los comandos deterministas cuya salida el arnés
// conoce de antemano (change opsx-task-context, D4). Esa salida es la
// referencia contra la que se comparan la captura (D22) y el resumen (D3).
// Vive fuera del binario (CHANGES.md:49).
package fixtures

import (
	"fmt"
	"io"
	"strings"
)

// Stream es el flujo por el que se escribe una línea.
type Stream int

const (
	Stdout Stream = iota
	Stderr
)

// Line es una línea de la salida del fixture, con su flujo.
type Line struct {
	Text   string
	Stream Stream
}

// Sensitive es un valor ficticio de un patrón de D14 dentro de la salida.
type Sensitive struct {
	PatternID string
	// Raw es el valor ficticio que no puede sobrevivir.
	Raw string
	// Line es el texto original (una o varias líneas separadas por \n) y
	// Redacted el texto esperado tras la redacción, con la misma cantidad de líneas.
	Line     string
	Redacted string
	// At es el índice (desde 0) de la primera línea en Lines.
	At int
}

// Fixture es un comando determinista con su salida conocida.
type Fixture struct {
	ID    string
	Title string
	// Command es el comando que el fixture representa; el capturador lo ejecuta
	// a través del proceso auxiliar del arnés.
	Command string
	// Tool nombra los patrones de fallo del catálogo; vacío = sin patrones declarados.
	Tool     string
	Lines    []Line
	ExitCode int
	// Failures son los fallos esperados, cada uno como la línea capturada.
	Failures []string
	// MustSurvive es el fallo que el resumen no puede perder (2.1).
	MustSurvive string
	Sensitive   []Sensitive
	// NeverCaptured marca un comando ya corrido sin captura y sin archivo (2.8).
	NeverCaptured bool
}

// Output es la salida completa en orden de emisión, stdout y stderr juntos.
func (f Fixture) Output() string {
	var b strings.Builder
	for _, l := range f.Lines {
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

// ExpectedRedacted es el original redactado esperado: cada línea sensible
// reemplazada por la suya redactada, en la misma posición.
func (f Fixture) ExpectedRedacted() string {
	lines := make([]string, len(f.Lines))
	for i, l := range f.Lines {
		lines[i] = l.Text
	}
	for _, s := range f.Sensitive {
		red := strings.Split(s.Redacted, "\n")
		for k, r := range red {
			lines[s.At+k] = r
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// Emit escribe las líneas del fixture en su flujo, en orden. Lo usa el proceso
// auxiliar que el capturador ejecuta como si fuera el comando del fixture.
func (f Fixture) Emit(stdout, stderr io.Writer) error {
	for _, l := range f.Lines {
		w := stdout
		if l.Stream == Stderr {
			w = stderr
		}
		if _, err := io.WriteString(w, l.Text+"\n"); err != nil {
			return err
		}
	}
	return nil
}

// All devuelve todos los fixtures en orden.
func All() []Fixture {
	return []Fixture{
		largeFailure("fail-start", "Fallo al inicio de 3.000 líneas", 3, 3000),
		largeFailure("fail-middle", "Fallo al medio de 3.000 líneas", 1500, 3000),
		largeFailure("fail-end", "Fallo al final de 3.000 líneas", 2998, 3000),
		scattered(),
		controlNegative(),
		sensitive(),
		noiseExitZero(),
		realAmongNoise(),
		stderrOnly(),
		neverCaptured(),
	}
}

// ByID busca un fixture.
func ByID(id string) (Fixture, bool) {
	for _, f := range All() {
		if f.ID == id {
			return f, true
		}
	}
	return Fixture{}, false
}

func pad(i int) string {
	return fmt.Sprintf("    case_%04d_test.go:%d: step completed padding padding padding padding", i, 10+i%50)
}

func stdout(text string) Line { return Line{Text: text, Stream: Stdout} }

const failPkgLine = "FAIL\tfixture/pkg\t0.512s"

// largeFailure arma una salida de go test de total líneas con la aguja en pos
// (1-based) y la línea FAIL final del paquete.
func largeFailure(id, title string, pos, total int) Fixture {
	needle := fmt.Sprintf("--- FAIL: TestNeedle_%s (0.00s)", strings.ReplaceAll(id, "-", "_"))
	lines := make([]Line, total)
	for i := range lines {
		lines[i] = stdout(pad(i + 1))
	}
	lines[pos-1] = stdout(needle)
	lines[total-1] = stdout(failPkgLine)
	return Fixture{
		ID: id, Title: title, Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 1,
		Failures:    []string{needle, failPkgLine},
		MustSurvive: needle,
	}
}

func scattered() Fixture {
	total := 3000
	lines := make([]Line, total)
	for i := range lines {
		lines[i] = stdout(pad(i + 1))
	}
	alpha := "--- FAIL: TestAlpha (0.00s)"
	beta := "    --- FAIL: TestBeta/sub_case (0.00s)"
	gamma := "panic: boom in TestGamma [recovered]"
	lines[99] = stdout(alpha)
	lines[1199] = stdout(beta)
	lines[2399] = stdout(gamma)
	lines[total-1] = stdout(failPkgLine)
	return Fixture{
		ID: "scattered", Title: "Fallos dispersos en tres posiciones", Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 1,
		Failures:    []string{alpha, beta, gamma, failPkgLine},
		MustSurvive: beta,
	}
}

func controlNegative() Fixture {
	total := 2000
	lines := make([]Line, total)
	for i := range lines {
		lines[i] = stdout(fmt.Sprintf("=== RUN   TestRepeated/case_%04d", i+1))
	}
	lines[total-1] = stdout("ok  \tfixture/pkg\t0.512s")
	return Fixture{
		ID: "control-negative", Title: "Salida verbosa y repetitiva con código cero", Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 0,
	}
}

// sensitive arma una salida con un valor ficticio por cada patrón de D14 y un
// fallo real en la misma salida.
func sensitive() Fixture {
	total := 400
	lines := make([]Line, total)
	for i := range lines {
		lines[i] = stdout(pad(i + 1))
	}
	type entry struct {
		pattern, raw, line, redacted string
		at                           int
	}
	entries := []entry{
		{"env-assignment", "FICTICIO-clave-0001", "DB_PASSWORD=FICTICIO-clave-0001", "DB_PASSWORD=[REDACTED:env-assignment]", 20},
		{"env-var-value", "FICTICIO-region-9", `$env:API_REGION = "FICTICIO-region-9"`, `$env:API_REGION = "[REDACTED:env-var-value]"`, 40},
		{"token", "FICTICIOtoken0123456789abcdef", "Authorization: Bearer FICTICIOtoken0123456789abcdef", "Authorization: Bearer [REDACTED:token]", 60},
		{"api-key", "sk-FICTICIO0123456789abcd", "api_key=sk-FICTICIO0123456789abcd", "api_key=[REDACTED:api-key]", 80},
		{"cloud-credential", "AKIAFICTICIO01234567", "credentials: AKIAFICTICIO01234567 (profile fixture)", "credentials: [REDACTED:cloud-credential] (profile fixture)", 100},
		{"ssh-private-key", "FICTICIOb3BlbnNzaC1rZXktdjEAAAAABG5vbmU",
			"-----BEGIN OPENSSH PRIVATE KEY-----\nFICTICIOb3BlbnNzaC1rZXktdjEAAAAABG5vbmU\nFICTICIOAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n-----END OPENSSH PRIVATE KEY-----",
			"-----BEGIN OPENSSH PRIVATE KEY-----\n[REDACTED:ssh-private-key]\n[REDACTED:ssh-private-key]\n-----END OPENSSH PRIVATE KEY-----", 120},
		{"home-path", "FICTICIO_usuario", `open C:\Users\FICTICIO_usuario\proyecto\x.go: access denied`, `open C:\Users\[REDACTED:home-path]\proyecto\x.go: access denied`, 140},
		{"account-id", "FICTICIO-acct-4455", "account_id: FICTICIO-acct-4455", "account_id: [REDACTED:account-id]", 160},
		{"account-uuid", "00000000-0000-4000-8000-0000F1C71C10", "account_uuid=00000000-0000-4000-8000-0000F1C71C10", "account_uuid=[REDACTED:account-uuid]", 180},
		{"email", "FICTICIO.persona@ejemplo.invalid", "owner FICTICIO.persona@ejemplo.invalid requested access", "owner [REDACTED:email] requested access", 200},
	}
	var sens []Sensitive
	for _, e := range entries {
		for k, l := range strings.Split(e.line, "\n") {
			lines[e.at+k] = stdout(l)
		}
		sens = append(sens, Sensitive{PatternID: e.pattern, Raw: e.raw, Line: e.line, Redacted: e.redacted, At: e.at})
	}
	fail := "--- FAIL: TestLeakyConfig (0.00s)"
	lines[300] = stdout(fail)
	lines[total-1] = stdout(failPkgLine)
	return Fixture{
		ID: "sensitive", Title: "Un valor ficticio por patrón sensible, con un fallo real", Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 1,
		Failures:    []string{fail, failPkgLine},
		MustSurvive: fail,
		Sensitive:   sens,
	}
}

func noiseExitZero() Fixture {
	total := 1500
	lines := make([]Line, total)
	for i := range lines {
		lines[i] = stdout(pad(i + 1))
	}
	noise := []string{
		"    handler_test.go:44:9: error: injected timeout (expected by the test)",
		"    logger.go:42:7: error: simulated retry (recovered)",
		"WARN: FAIL-safe mode enabled, not a failure",
		"--- PASS: TestErrorPathsExercised (0.00s)",
	}
	for i := 100; i < total-1; i += 150 {
		for k, n := range noise {
			lines[i+k] = stdout(n)
		}
	}
	lines[total-1] = stdout("ok  \tfixture/pkg\t0.512s")
	return Fixture{
		ID: "noise-exit0", Title: "Ruido con forma de error y código cero", Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 0,
	}
}

func realAmongNoise() Fixture {
	var lines []Line
	lines = append(lines, stdout("# fixture/pkg"))
	noise := []string{
		"logger.go:42: error: simulated retry (recovered)",
		"retry.go:17: error budget consumed 3/5 (informational)",
		"WARN: errors.Is(err, ErrTransient) path exercised",
	}
	for i := 0; i < 15; i++ {
		lines = append(lines, stdout(noise[i%len(noise)]))
	}
	real := "broken.go:12:5: undefined: widgetCount"
	lines = append(lines, stdout(real))
	for i := 0; i < 15; i++ {
		lines = append(lines, stdout(noise[(i+1)%len(noise)]))
	}
	return Fixture{
		ID: "real-among-noise", Title: "Un fallo real rodeado de ruido que lo imita", Command: "go build ./fixture/pkg/", Tool: "go build",
		Lines: lines, ExitCode: 1,
		Failures:    []string{real},
		MustSurvive: real,
	}
}

func stderrOnly() Fixture {
	total := 1200
	lines := make([]Line, 0, total+3)
	for i := 1; i <= total; i++ {
		lines = append(lines, stdout(pad(i)))
	}
	panicLine := "panic: runtime error: index out of range [5] with length 3"
	lines = append(lines,
		Line{Text: panicLine, Stream: Stderr},
		Line{Text: "goroutine 1 [running]:", Stream: Stderr},
		Line{Text: failPkgLine, Stream: Stderr},
	)
	return Fixture{
		ID: "stderr-only", Title: "El fallo se escribe solo por stderr", Command: "go test -v ./fixture/pkg/", Tool: "go test",
		Lines: lines, ExitCode: 1,
		Failures:    []string{panicLine, failPkgLine},
		MustSurvive: panicLine,
	}
}

func neverCaptured() Fixture {
	return Fixture{
		ID: "never-captured", Title: "Comando ya corrido sin captura y sin archivo", Command: "go test -v ./fixture/pkg/ -run TestNeverCaptured", Tool: "go test",
		ExitCode: 1, NeverCaptured: true,
	}
}
