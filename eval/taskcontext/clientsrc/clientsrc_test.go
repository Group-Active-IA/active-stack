package clientsrc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/clientsrc"
)

// Tarea 3.6 (opsx-task-context, D8 y D22): adaptadores de fuente del cliente
// sobre muestras derivadas de la documentación y el comparador de cobertura.

const samples = "../fixtures/samples"

// Referencia del arnés para las muestras (el mini fixture de seis líneas).
const miniRef = "=== RUN   TestMini\nline 02 padding\nline 03 padding\n--- FAIL: TestMini (0.00s)\nline 05 padding\nFAIL\tmini\t0.100s\n"

const okRef = "=== RUN   TestOk\nline 02 padding\nok  \tmini\t0.100s\n"

func sample(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(samples, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("falta la muestra %s: %v", rel, err)
	}
	return data
}

func lines(data []byte) [][]byte {
	var out [][]byte
	for _, l := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		out = append(out, l)
	}
	return out
}

// Una muestra de cada caso: el veredicto es el esperado.
func TestCompare_EachCoverageVerdict(t *testing.T) {
	tests := []struct {
		name  string
		get   func(t *testing.T) (clientsrc.Observed, error)
		ref   string
		want  clientsrc.Coverage
		exit  int
		check func(t *testing.T, o clientsrc.Observed)
	}{
		{"hook PostToolUse completo", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromClaudePostToolUse(sample(t, "claude/posttooluse-complete.json"))
		}, okRef, clientsrc.Complete, 0, nil},
		{"hook PostToolUseFailure truncado en el medio", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromClaudePostToolUseFailure(sample(t, "claude/posttoolusefailure-truncated.json"))
		}, miniRef, clientsrc.Truncated, 1, func(t *testing.T, o clientsrc.Observed) {
			if o.TruncatedChars != 59 {
				t.Errorf("caracteres omitidos según la marca = %d, want 59", o.TruncatedChars)
			}
		}},
		{"hook PostToolUseFailure con la forma medida (cabeza y cola cortadas a mitad de línea, cola interior)", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromClaudePostToolUseFailure(sample(t, "claude/posttoolusefailure-real-shape.json"))
		}, miniRef, clientsrc.Truncated, 1, func(t *testing.T, o clientsrc.Observed) {
			if o.TruncatedChars != 40 {
				t.Errorf("caracteres omitidos según la marca = %d, want 40", o.TruncatedChars)
			}
		}},
		{"App Server completo (deltas y aggregatedOutput)", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromCodexAppServer(lines(sample(t, "codex/appserver-complete.jsonl")))
		}, miniRef, clientsrc.Complete, 1, nil},
		{"App Server sin salida", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromCodexAppServer(lines(sample(t, "codex/appserver-no-output.jsonl")))
		}, miniRef, clientsrc.Absent, 1, nil},
		{"codex exec --json completo (forma medida en X7)", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromCodexExecJSON(lines(sample(t, "codex/exec-command-execution.jsonl")))
		}, miniRef, clientsrc.Complete, 1, nil},
		{"codex exec --json recodificado", func(t *testing.T) (clientsrc.Observed, error) {
			return clientsrc.FromCodexExecJSON(lines(sample(t, "codex/exec-reencoded.jsonl")))
		}, miniRef, clientsrc.NotChecked, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs, err := tt.get(t)
			if err != nil {
				t.Fatal(err)
			}
			if got := clientsrc.Compare(tt.ref, obs); got != tt.want {
				t.Errorf("cobertura = %q, want %q (salida observada %q)", got, tt.want, obs.Output)
			}
			if obs.ExitCode != nil && *obs.ExitCode != tt.exit {
				t.Errorf("código de salida = %d, want %d", *obs.ExitCode, tt.exit)
			}
			if tt.check != nil {
				tt.check(t, obs)
			}
		})
	}
}

func TestCompare_Rules(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		obs  clientsrc.Observed
		want clientsrc.Coverage
	}{
		{"idéntica", "a\nb\nc\n", clientsrc.Observed{Present: true, Output: "a\nb\nc\n"}, clientsrc.Complete},
		{"solo la cabeza", "a\nb\nc\nd\n", clientsrc.Observed{Present: true, Output: "a\nb\n"}, clientsrc.Truncated},
		{"solo la cola", "a\nb\nc\nd\n", clientsrc.Observed{Present: true, Output: "c\nd\n"}, clientsrc.Truncated},
		{"cabeza y cola con marca del cliente", "a\nb\nc\nd\ne\n", clientsrc.Observed{Present: true, Output: "a\n... [6 characters truncated] ...\ne\n"}, clientsrc.Truncated},
		// Forma medida en Claude Code 2.1.323: Bash conserva primero unos 30.000
		// caracteres y después recorta el medio, así que la cola es un tramo interior.
		{"marca entre líneas en blanco y cola interior", "aaa\nbbb\nccc\nddd\neee\n", clientsrc.Observed{Present: true, Output: "aaa\nb\n\n... [5 characters truncated] ...\n\ncc\nddd\nee"}, clientsrc.Truncated},
		{"cola interior que no sigue a la cabeza", "aaa\nbbb\nccc\n", clientsrc.Observed{Present: true, Output: "bbb\n\n... [3 characters truncated] ...\n\naaa"}, clientsrc.NotChecked},
		{"cola que no está en la referencia", "aaa\nbbb\nccc\n", clientsrc.Observed{Present: true, Output: "aaa\n\n... [3 characters truncated] ...\n\nzzz"}, clientsrc.NotChecked},
		{"no presente", "a\n", clientsrc.Observed{}, clientsrc.Absent},
		{"presente pero vacío", "a\n", clientsrc.Observed{Present: true, Output: ""}, clientsrc.Absent},
		{"otro contenido sin marca", "a\nb\n", clientsrc.Observed{Present: true, Output: "x\ny\nz\n"}, clientsrc.NotChecked},
		{"sin referencia contra qué comparar", "", clientsrc.Observed{Present: true, Output: "a\n"}, clientsrc.NotChecked},
		{"final de línea CRLF equivale", "a\nb\n", clientsrc.Observed{Present: true, Output: "a\r\nb\r\n"}, clientsrc.Complete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientsrc.Compare(tt.ref, tt.obs); got != tt.want {
				t.Errorf("Compare = %q, want %q", got, tt.want)
			}
		})
	}
}

// Medido en Claude Code 2.1.319 en Windows: el cliente agrega a stderr su propia
// nota "Shell cwd was reset to <dir>", que no es salida del comando.
func TestClaudePostToolUse_DropsClientCwdNote(t *testing.T) {
	payload := []byte(`{"hook_event_name":"PostToolUse","tool_response":{"stdout":"=== RUN   TestOk\nline 02 padding\nok  \tmini\t0.100s\n","stderr":"\nShell cwd was reset to C:\\Users\\FICTICIO\\proj"}}`)
	obs, err := clientsrc.FromClaudePostToolUse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got := clientsrc.Compare(okRef, obs); got != clientsrc.Complete {
		t.Errorf("cobertura = %q, want %q (salida %q)", got, clientsrc.Complete, obs.Output)
	}
	// stderr real del comando se conserva.
	payload = []byte(`{"hook_event_name":"PostToolUse","tool_response":{"stdout":"a\n","stderr":"warning: x\n"}}`)
	obs, _ = clientsrc.FromClaudePostToolUse(payload)
	if obs.Output != "a\nwarning: x\n" {
		t.Errorf("salida = %q, want stdout más stderr del comando", obs.Output)
	}
}

func TestAdapters_RejectMalformedInput(t *testing.T) {
	if _, err := clientsrc.FromClaudePostToolUse([]byte("{no es json")); err == nil {
		t.Error("PostToolUse con JSON inválido debe dar error")
	}
	if _, err := clientsrc.FromClaudePostToolUseFailure([]byte(`{"hook_event_name":"PostToolUseFailure"}`)); err != nil {
		t.Errorf("un fallo sin campo error es una fuente ausente, no un error: %v", err)
	}
	obs, _ := clientsrc.FromClaudePostToolUseFailure([]byte(`{"hook_event_name":"PostToolUseFailure"}`))
	if obs.Present {
		t.Error("sin error no hay salida presente")
	}
	if _, err := clientsrc.FromCodexExecJSON([][]byte{[]byte("no json")}); err == nil {
		t.Error("exec --json con una línea inválida debe dar error")
	}
	// Una línea que no es un command_execution se ignora.
	obs, err := clientsrc.FromCodexExecJSON([][]byte{[]byte(`{"type":"thread.started","thread_id":"x"}`)})
	if err != nil || obs.Present {
		t.Errorf("sin command_execution la fuente está ausente: %+v %v", obs, err)
	}
}

// La salida del hook de fallo trae el código de salida en su primera línea.
func TestClaudeFailure_ParsesExitCodePrefix(t *testing.T) {
	obs, err := clientsrc.FromClaudePostToolUseFailure([]byte(`{"hook_event_name":"PostToolUseFailure","tool_name":"Bash","error":"Exit code 127\ncommand not found\n"}`))
	if err != nil {
		t.Fatal(err)
	}
	if obs.ExitCode == nil || *obs.ExitCode != 127 {
		t.Errorf("código = %v, want 127", obs.ExitCode)
	}
	if !strings.HasPrefix(obs.Output, "command not found") {
		t.Errorf("la salida no debe repetir la línea de código: %q", obs.Output)
	}
}
