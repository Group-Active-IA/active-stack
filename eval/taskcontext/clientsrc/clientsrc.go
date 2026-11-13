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
	// Interrupted: el hook marcó la ejecución como interrumpida; no hay código de salida.
	Interrupted bool
}

var truncMarker = regexp.MustCompile(`\.\.\. \[(\d+) characters truncated\] \.\.\.`)

func intp(n int) *int { return &n }

// FromClaudePostToolUse adapta el hook PostToolUse de Bash: tool_response con
// stdout, stderr, interrupted e isImage. Un PostToolUse es un comando que no
// falló, así que el código es 0, salvo que venga interrumpido: entonces
// Interrupted=true y ExitCode queda nil.
func FromClaudePostToolUse(payload []byte) (Observed, error) {
	var ev struct {
		ToolResponse *struct {
			Stdout      string `json:"stdout"`
			Stderr      string `json:"stderr"`
			Interrupted bool   `json:"interrupted"`
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
	obs.Output = ev.ToolResponse.Stdout + stripCwdNote(ev.ToolResponse.Stderr)
	// Una ejecución interrumpida no terminó: el hook no entrega código y no se inventa 0.
	if ev.ToolResponse.Interrupted {
		obs.Interrupted = true
		return obs, nil
	}
	obs.ExitCode = intp(0)
	return obs, nil
}

var exitPrefix = regexp.MustCompile(`^Exit code (-?\d+)\n`)

// cwdNote es la nota de Claude Code en Windows (medida en 2.1.319): la última
// línea completa de stderr. Anclada a inicio de línea para no tocar una línea
// real del comando que solo contenga la frase.
var cwdNote = regexp.MustCompile(`(?m)^Shell cwd was reset to [^\n]*\n?\z`)

// stripCwdNote quita la nota del cliente y conserva el salto de línea que cierra
// el contenido real. Si no queda contenido real, el salto era solo el separador.
func stripCwdNote(stderr string) string {
	loc := cwdNote.FindStringIndex(stderr)
	if loc == nil {
		return stderr
	}
	rest := stderr[:loc[0]]
	if rest == "\n" {
		return ""
	}
	return rest
}

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

// errAmbiguous: la fuente trae más de un comando y no se puede decir cuál se pide.
func errAmbiguous(source string, ids []string) error {
	return fmt.Errorf("clientsrc: %s: la entrada trae más de un comando (%s); pasar los eventos de uno solo", source, strings.Join(ids, ", "))
}

// errConflict: el mismo comando trae más de un final y no son copias idénticas.
func errConflict(source, id string) error {
	if id == "" {
		return fmt.Errorf("clientsrc: %s: más de un final de comando sin identidad; no se puede decir si son uno o varios", source)
	}
	return fmt.Errorf("clientsrc: %s: el comando %s trae finales contradictorios (comando, salida o código distintos)", source, id)
}

// final es lo que afirma el final de un comando. Un final repetido solo se
// tolera si es una copia idéntica (replay); si difiere, la entrada es ambigua.
type final struct {
	command, output string
	exit            *int
}

func (a final) same(b final) bool {
	if a.command != b.command || a.output != b.output || (a.exit == nil) != (b.exit == nil) {
		return false
	}
	return a.exit == nil || *a.exit == *b.exit
}

// FromCodexAppServer adapta las notificaciones del App Server:
// item/commandExecution/outputDelta (en orden) y el item final con
// aggregatedOutput y exitCode, ambos opcionales. Salida y código se correlacionan
// por la identidad completa threadId/turnId/itemId: la entrada debe ser de un
// solo comando, y se rechaza en vez de mezclar o descartar en silencio si trae
// más de una identidad (el mismo itemId en otro thread o turn es otro comando),
// más de un final sin identidad, o finales del mismo comando que no son copias
// idénticas.
func FromCodexAppServer(lines [][]byte) (Observed, error) {
	obs := Observed{Source: "cliente:app-server"}
	var deltas strings.Builder
	var ids []string
	var done *final
	seen := func(id string) {
		for _, s := range ids {
			if s == id {
				return
			}
		}
		ids = append(ids, id)
	}
	for _, line := range lines {
		var ev struct {
			Method string `json:"method"`
			Params struct {
				ThreadID string `json:"threadId"`
				TurnID   string `json:"turnId"`
				ItemID   string `json:"itemId"`
				Delta    string `json:"delta"`
				Item     struct {
					ID               string `json:"id"`
					Type             string `json:"type"`
					Command          string `json:"command"`
					AggregatedOutput string `json:"aggregatedOutput"`
					ExitCode         *int   `json:"exitCode"`
				} `json:"item"`
			} `json:"params"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return obs, fmt.Errorf("clientsrc: App Server: %w", err)
		}
		scope := func(item string) string {
			if ev.Params.ThreadID == "" && ev.Params.TurnID == "" {
				return item
			}
			return ev.Params.ThreadID + "/" + ev.Params.TurnID + "/" + item
		}
		switch ev.Method {
		case "item/commandExecution/outputDelta":
			seen(scope(ev.Params.ItemID))
			deltas.WriteString(ev.Params.Delta)
		case "item/completed":
			if ev.Params.Item.Type != "commandExecution" {
				continue
			}
			id := scope(ev.Params.Item.ID)
			seen(id)
			if len(ids) > 1 {
				return Observed{Source: obs.Source}, errAmbiguous("App Server", ids)
			}
			f := final{ev.Params.Item.Command, ev.Params.Item.AggregatedOutput, ev.Params.Item.ExitCode}
			if done != nil && (ev.Params.Item.ID == "" || !done.same(f)) {
				return Observed{Source: obs.Source}, errConflict("App Server", id)
			}
			done = &f
		}
	}
	if len(ids) > 1 {
		return Observed{Source: obs.Source}, errAmbiguous("App Server", ids)
	}
	var aggregated string
	if done != nil {
		aggregated, obs.ExitCode = done.output, done.exit
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
// salida completa. La entrada debe tener un solo command_execution: con ids
// distintos, con más de uno sin id, o con finales del mismo id que no son copias
// idénticas, devuelve error en vez de quedarse con el último.
func FromCodexExecJSON(lines [][]byte) (Observed, error) {
	obs := Observed{Source: "cliente:exec-json"}
	var ids []string
	var done *final
	for _, line := range lines {
		var ev struct {
			Type string `json:"type"`
			Item struct {
				ID               string `json:"id"`
				Type             string `json:"type"`
				Command          string `json:"command"`
				AggregatedOutput string `json:"aggregated_output"`
				ExitCode         *int   `json:"exit_code"`
			} `json:"item"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return obs, fmt.Errorf("clientsrc: exec --json: %w", err)
		}
		if ev.Type != "item.completed" || ev.Item.Type != "command_execution" {
			continue
		}
		if len(ids) > 0 && (ev.Item.ID == "" || ev.Item.ID != ids[0]) {
			return Observed{Source: obs.Source}, errAmbiguous("exec --json", append(ids, ev.Item.ID))
		}
		f := final{ev.Item.Command, ev.Item.AggregatedOutput, ev.Item.ExitCode}
		if done != nil && !done.same(f) {
			return Observed{Source: obs.Source}, errConflict("exec --json", ev.Item.ID)
		}
		done = &f
		if len(ids) == 0 {
			ids = append(ids, ev.Item.ID)
		}
		obs.Present = true
		obs.Output = ev.Item.AggregatedOutput
		obs.ExitCode = ev.Item.ExitCode
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
