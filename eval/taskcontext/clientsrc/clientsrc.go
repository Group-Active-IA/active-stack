// Package clientsrc adapta las fuentes de captura del cliente (change
// opsx-task-context, D8 y D22) y compara lo que cada una entrega con la
// referencia del arnés. Hasta que la tarea 8.1 las corra con un cliente real,
// las muestras derivan de la documentación y de lo medido (fixtures/samples).
// Vive fuera del binario (CHANGES.md:49).
package clientsrc

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Coverage es el veredicto del comparador (tarea 3.6).
type Coverage string

const (
	Complete   Coverage = "completa"
	Truncated  Coverage = "truncada"
	Absent     Coverage = "ausente"
	NotChecked Coverage = "no comprobada"
)

// Observed es lo que una fuente del cliente entregó.
type Observed struct {
	Source string
	// Present dice si la fuente trajo alguna salida (los campos pueden ser opcionales).
	Present        bool
	Output         string
	ExitCode       *int
	TruncatedChars int
}

var truncMarker = regexp.MustCompile(`\.\.\. \[(\d+) characters truncated\] \.\.\.`)

func intp(n int) *int { return &n }

// FromClaudePostToolUse adapta el hook PostToolUse de Bash: tool_response con
// stdout, stderr, interrupted e isImage. Un PostToolUse es un comando que no
// falló, así que el código es 0.
func FromClaudePostToolUse(payload []byte) (Observed, error) {
	var ev struct {
		ToolResponse *struct {
			Stdout string `json:"stdout"`
			Stderr string `json:"stderr"`
		} `json:"tool_response"`
	}
	if err := json.Unmarshal(payload, &ev); err != nil {
		return Observed{}, fmt.Errorf("clientsrc: PostToolUse: %w", err)
	}
	obs := Observed{Source: "cliente:PostToolUse"}
	if ev.ToolResponse == nil {
		return obs, nil
	}
	// stdout y stderr llegan separados: se concatenan y el orden entre ambos se pierde.
	// La nota que el cliente agrega a stderr en Windows no es salida del comando.
	obs.Present = true
	obs.Output = ev.ToolResponse.Stdout + cwdNote.ReplaceAllString(ev.ToolResponse.Stderr, "")
	obs.ExitCode = intp(0)
	return obs, nil
}

var exitPrefix = regexp.MustCompile(`^Exit code (-?\d+)\n`)

// cwdNote es la nota de Claude Code en Windows (medida en 2.1.319) al final de stderr.
var cwdNote = regexp.MustCompile(`\n?Shell cwd was reset to [^\n]*\n?$`)

// FromClaudePostToolUseFailure adapta el hook PostToolUseFailure: el campo
// error empieza con `Exit code N` y se recorta por el medio con la marca
// `... [N characters truncated] ...`.
func FromClaudePostToolUseFailure(payload []byte) (Observed, error) {
	var ev struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(payload, &ev); err != nil {
		return Observed{}, fmt.Errorf("clientsrc: PostToolUseFailure: %w", err)
	}
	obs := Observed{Source: "cliente:PostToolUseFailure"}
	if ev.Error == "" {
		return obs, nil
	}
	obs.Present = true
	body := ev.Error
	if m := exitPrefix.FindStringSubmatch(body); m != nil {
		n, _ := strconv.Atoi(m[1])
		obs.ExitCode = &n
		body = body[len(m[0]):]
	}
	obs.Output = body
	if m := truncMarker.FindStringSubmatch(body); m != nil {
		obs.TruncatedChars, _ = strconv.Atoi(m[1])
	}
	return obs, nil
}

// FromCodexAppServer adapta las notificaciones del App Server:
// item/commandExecution/outputDelta (en orden) y el item final con
// aggregatedOutput y exitCode, ambos opcionales.
func FromCodexAppServer(lines [][]byte) (Observed, error) {
	obs := Observed{Source: "cliente:app-server"}
	var deltas strings.Builder
	var aggregated string
	for _, line := range lines {
		var ev struct {
			Method string `json:"method"`
			Params struct {
				Delta string `json:"delta"`
				Item  struct {
					Type             string `json:"type"`
					AggregatedOutput string `json:"aggregatedOutput"`
					ExitCode         *int   `json:"exitCode"`
				} `json:"item"`
			} `json:"params"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return obs, fmt.Errorf("clientsrc: App Server: %w", err)
		}
		switch ev.Method {
		case "item/commandExecution/outputDelta":
			deltas.WriteString(ev.Params.Delta)
		case "item/completed":
			if ev.Params.Item.Type == "commandExecution" {
				aggregated = ev.Params.Item.AggregatedOutput
				obs.ExitCode = ev.Params.Item.ExitCode
			}
		}
	}
	switch {
	case aggregated != "":
		obs.Present, obs.Output = true, aggregated
	case deltas.Len() > 0:
		obs.Present, obs.Output = true, deltas.String()
	}
	return obs, nil
}

// FromCodexExecJSON adapta `codex exec --json`: el item command_execution con
// aggregated_output y exit_code. Medido en X7 (Codex 0.167.1, Windows): trae la
// salida completa.
func FromCodexExecJSON(lines [][]byte) (Observed, error) {
	obs := Observed{Source: "cliente:exec-json"}
	for _, line := range lines {
		var ev struct {
			Type string `json:"type"`
			Item struct {
				Type             string `json:"type"`
				AggregatedOutput string `json:"aggregated_output"`
				ExitCode         *int   `json:"exit_code"`
			} `json:"item"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return obs, fmt.Errorf("clientsrc: exec --json: %w", err)
		}
		if ev.Type == "item.completed" && ev.Item.Type == "command_execution" {
			obs.Present = true
			obs.Output = ev.Item.AggregatedOutput
			obs.ExitCode = ev.Item.ExitCode
		}
	}
	return obs, nil
}

func normalizeEOL(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

// Compare contrasta lo que entregó la fuente con la referencia del arnés.
//
//   - ausente: la fuente no trajo salida;
//   - completa: coincide con la referencia (los finales de línea CRLF equivalen a LF);
//   - truncada: es una parte de la referencia: cabeza, cola, o cabeza y cola
//     separadas por la marca `... [N characters truncated] ...`;
//   - no comprobada: no se puede relacionar con la referencia.
func Compare(reference string, obs Observed) Coverage {
	if !obs.Present || obs.Output == "" {
		return Absent
	}
	ref, out := normalizeEOL(reference), normalizeEOL(obs.Output)
	if ref == "" {
		return NotChecked
	}
	if out == ref {
		return Complete
	}
	if loc := truncMarker.FindStringIndex(out); loc != nil {
		// La marca va entre saltos de línea propios del cliente. La cabeza puede
		// cortarse a mitad de línea, y la cola es un tramo de la referencia posterior a
		// la cabeza: el sufijo, o un tramo interior cuando el cliente ya había
		// descartado el final (Bash de Claude Code conserva unos 30.000 caracteres).
		head := strings.TrimRight(out[:loc[0]], "\n")
		tail := strings.TrimLeft(out[loc[1]:], "\n")
		if strings.HasPrefix(ref, head) && strings.Contains(ref[len(head):], tail) {
			return Truncated
		}
		return NotChecked
	}
	if len(out) < len(ref) && (strings.HasPrefix(ref, out) || strings.HasSuffix(ref, out)) {
		return Truncated
	}
	return NotChecked
}
