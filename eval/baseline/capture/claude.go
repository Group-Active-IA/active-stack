package capture

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/trace"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

// ClaudeMinAgentIDVersion es la primera version de Claude Code que informa
// agent_id y parent_agent_id en los eventos de OpenTelemetry
// (https://code.claude.com/docs/en/monitoring-usage, consultada 2026-11-11:
// "Requires Claude Code v2.1.310 or later").
const ClaudeMinAgentIDVersion = "2.1.310"

// mainQuerySource es el query_source de un evento api_request de la
// conversacion principal ("repl_main_thread", segun la misma pagina).
const mainQuerySource = "repl_main_thread"

// Attribution es el resultado de repartir el consumo por rol.
type Attribution struct {
	Entries   []usage.Entry
	ByAgent   map[string]int64
	Partial   bool
	PerWorker Availability
	Reasons   []string
}

// ClaudeRequest es una solicitud (evento api_request de OpenTelemetry).
type ClaudeRequest struct {
	RequestID, QuerySource, AgentID, ParentAgentID, AgentName, Model string
	Usage                                                            usage.Logical
}

// ClaudeToolResult es un evento tool_result.
type ClaudeToolResult struct{ ToolName, AgentID string }

// SubagentClose es el evento subagent_completed. Su TotalTokens cubre solo la
// ultima solicitud del subagente y nunca se suma al consumo.
type SubagentClose struct {
	AgentID, AgentType    string
	TotalTokens, ToolUses int64
}

// ClaudeOTel es lo capturado de los eventos de OpenTelemetry.
type ClaudeOTel struct {
	Requests       []ClaudeRequest
	ToolResults    []ClaudeToolResult
	SubagentCloses []SubagentClose
	Duplicates     int
	Incomplete     []string
	Ignored        int
}

type attrs map[string]any

func decodeAttrs(line []byte) (attrs, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber()
	var m attrs
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, errors.New("contenido extra después del objeto JSON")
	}
	return m, nil
}

func (a attrs) str(key string) string {
	if s, ok := a[key].(string); ok {
		return s
	}
	return ""
}

// count lee un entero no negativo, como numero o como texto numerico.
func (a attrs) count(key string) (int64, bool) {
	switch v := a[key].(type) {
	case json.Number:
		n, err := v.Int64()
		return n, err == nil && n >= 0
	case string:
		var n int64
		if _, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &n); err != nil || strings.TrimSpace(v) != fmt.Sprint(n) {
			return 0, false
		}
		return n, n >= 0
	}
	return 0, false
}

// forEachLine recorre las lineas no vacias; el numero de linea parte de 1.
func forEachLine(r io.Reader, fn func(n int, line []byte) error) error {
	br := bufio.NewReader(r)
	for n := 1; ; n++ {
		line, err := br.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			if ferr := fn(n, bytes.TrimSpace(line)); ferr != nil {
				return ferr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// ParseClaudeOTel lee eventos de OpenTelemetry, uno por linea JSON con los
// atributos como claves. Deduplica api_request por request_id y tool_result por
// tool_use_id. Una solicitud sin los cuatro campos de tokens se informa en
// Incomplete y no suma (ausente no es cero).
func ParseClaudeOTel(r io.Reader) (ClaudeOTel, error) {
	var out ClaudeOTel
	seenReq := map[string]bool{}
	seenTool := map[string]bool{}
	err := forEachLine(r, func(n int, line []byte) error {
		a, err := decodeAttrs(line)
		if err != nil {
			return fmt.Errorf("claude otel: línea %d: %w", n, err)
		}
		switch strings.TrimPrefix(a.str("event.name"), "claude_code.") {
		case "api_request":
			id := a.str("request_id")
			if id != "" {
				if seenReq[id] {
					out.Duplicates++
					return nil
				}
				seenReq[id] = true
			}
			in, ok1 := a.count("input_tokens")
			o, ok2 := a.count("output_tokens")
			cr, ok3 := a.count("cache_read_tokens")
			cc, ok4 := a.count("cache_creation_tokens")
			if !ok1 || !ok2 || !ok3 || !ok4 {
				out.Incomplete = append(out.Incomplete, fmt.Sprintf("solicitud %q (línea %d): tokens ausentes o inválidos", id, n))
				return nil
			}
			l, err := usage.FromClaude(usage.ClaudeRequest{Input: in, CacheRead: cr, CacheCreation: cc, Output: o})
			if err != nil {
				return fmt.Errorf("claude otel: línea %d: %w", n, err)
			}
			out.Requests = append(out.Requests, ClaudeRequest{
				RequestID: id, QuerySource: a.str("query_source"), AgentID: a.str("agent_id"),
				ParentAgentID: a.str("parent_agent_id"), AgentName: a.str("agent.name"), Model: a.str("model"), Usage: l,
			})
		case "tool_result":
			if id := a.str("tool_use_id"); id != "" {
				if seenTool[id] {
					out.Duplicates++
					return nil
				}
				seenTool[id] = true
			}
			if a.str("tool_name") == "" {
				out.Incomplete = append(out.Incomplete, fmt.Sprintf("tool_result (línea %d): sin tool_name", n))
				return nil
			}
			out.ToolResults = append(out.ToolResults, ClaudeToolResult{ToolName: a.str("tool_name"), AgentID: a.str("agent_id")})
		case "subagent_completed":
			tt, _ := a.count("total_tokens")
			tu, _ := a.count("total_tool_uses")
			out.SubagentCloses = append(out.SubagentCloses, SubagentClose{AgentID: a.str("agent_id"), AgentType: a.str("agent_type"), TotalTokens: tt, ToolUses: tu})
		default:
			out.Ignored++
		}
		return nil
	})
	return out, err
}

// Attribute reparte las solicitudes por rol segun la version de Claude Code.
//
//   - Con agent_id (desde 2.1.310) la solicitud es de un worker, identificado solo
//     por agent_id. agent.name nunca atribuye: un agente propio se informa como
//     "custom".
//   - Sin agent_id y con query_source principal, es del principal.
//   - Sin agent_id y con otro query_source (version >= 2.1.310), no es un
//     subagente: consumo auxiliar.
//   - Antes de 2.1.310 no hay agent_id: lo que no es del principal queda sin
//     atribuir y la atribucion es parcial.
//
// El total_tokens del cierre de un subagente no entra en ningun caso.
func (c ClaudeOTel) Attribute(version string) (Attribution, error) {
	ok, err := manifest.VersionAtLeast(version, ClaudeMinAgentIDVersion)
	if err != nil {
		return Attribution{}, fmt.Errorf("claude: %w", err)
	}
	att := Attribution{ByAgent: map[string]int64{}, PerWorker: Availability{Available: ok}}
	if !ok {
		att.PerWorker.Reason = fmt.Sprintf("agent_id solo desde Claude Code %s (versión %s): no hay atribución por worker", ClaudeMinAgentIDVersion, version)
		att.Reasons = append(att.Reasons, att.PerWorker.Reason)
	}
	for _, r := range c.Requests {
		role := RoleUnattributed
		switch {
		case ok && r.AgentID != "":
			role = RoleWorker
			att.ByAgent[r.AgentID] += r.Usage.Total()
		case r.AgentID == "" && r.QuerySource == mainQuerySource:
			role = RolePrincipal
		case ok:
			role = RoleAuxiliary
		default:
			att.Partial = true
		}
		id := r.RequestID
		if r.AgentID != "" && role == RoleWorker {
			id = r.AgentID
		}
		att.Entries = append(att.Entries, usage.Entry{Role: role, ID: id, Tokens: r.Usage.Total()})
	}
	if len(c.Incomplete) > 0 {
		att.Partial = true
		att.Reasons = append(att.Reasons, fmt.Sprintf("%d evento(s) sin tokens válidos", len(c.Incomplete)))
	}
	return att, nil
}

// ClaudeResult es el mensaje result de `claude -p --output-format json` o del
// ultimo evento de stream-json.
type ClaudeResult struct {
	Subtype       string
	IsError       bool
	SessionID     string
	DurationMS    int64
	DurationAPIMS int64
	NumTurns      int
	// TotalCostUSD es una estimacion del cliente, no un dato de facturacion.
	TotalCostUSD float64
	// MainUsage es `usage`: solo el bucle principal, sin subagentes.
	MainUsage    usage.Logical
	HasMainUsage bool
	// ModelUsage es `modelUsage`: incluye las solicitudes de los subagentes.
	ModelUsage        map[string]usage.Logical
	PermissionDenials int
}

// TreeTotal es la suma de modelUsage, que la documentacion dice que incluye a
// los subagentes. No esta disponible si el resultado no trae modelUsage.
func (r ClaudeResult) TreeTotal() (int64, bool) {
	if len(r.ModelUsage) == 0 {
		return 0, false
	}
	var sum int64
	for _, l := range r.ModelUsage {
		sum += l.Total()
	}
	return sum, true
}

type rawUsage struct {
	InputTokens              *int64 `json:"input_tokens"`
	OutputTokens             *int64 `json:"output_tokens"`
	CacheCreationInputTokens *int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     *int64 `json:"cache_read_input_tokens"`
}

type rawModelUsage struct {
	InputTokens              *int64 `json:"inputTokens"`
	OutputTokens             *int64 `json:"outputTokens"`
	CacheReadInputTokens     *int64 `json:"cacheReadInputTokens"`
	CacheCreationInputTokens *int64 `json:"cacheCreationInputTokens"`
}

func logicalOf(in, out, cr, cc *int64) (usage.Logical, bool, error) {
	if in == nil || out == nil || cr == nil || cc == nil {
		return usage.Logical{}, false, nil
	}
	l, err := usage.FromClaude(usage.ClaudeRequest{Input: *in, Output: *out, CacheRead: *cr, CacheCreation: *cc})
	return l, err == nil, err
}

type rawResult struct {
	Type              string                   `json:"type"`
	Subtype           string                   `json:"subtype"`
	IsError           bool                     `json:"is_error"`
	SessionID         string                   `json:"session_id"`
	DurationMS        int64                    `json:"duration_ms"`
	DurationAPIMS     int64                    `json:"duration_api_ms"`
	NumTurns          int                      `json:"num_turns"`
	TotalCostUSD      float64                  `json:"total_cost_usd"`
	Usage             *rawUsage                `json:"usage"`
	ModelUsage        map[string]rawModelUsage `json:"modelUsage"`
	PermissionDenials []json.RawMessage        `json:"permission_denials"`
}

// ParseClaudeResult interpreta el resultado JSON. Solo cuenta las denegaciones
// de permisos: su contenido (entrada de la herramienta) no se conserva.
func ParseClaudeResult(data []byte) (ClaudeResult, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var raw rawResult
	if err := dec.Decode(&raw); err != nil {
		return ClaudeResult{}, fmt.Errorf("claude result: %w", err)
	}
	if dec.More() {
		return ClaudeResult{}, errors.New("claude result: contenido extra después del primer objeto JSON")
	}
	if raw.Type != "result" {
		return ClaudeResult{}, fmt.Errorf("claude result: type %q, se esperaba result", raw.Type)
	}
	res := ClaudeResult{
		Subtype: raw.Subtype, IsError: raw.IsError, SessionID: raw.SessionID,
		DurationMS: raw.DurationMS, DurationAPIMS: raw.DurationAPIMS, NumTurns: raw.NumTurns,
		TotalCostUSD: raw.TotalCostUSD, PermissionDenials: len(raw.PermissionDenials),
	}
	if raw.Usage != nil {
		l, ok, err := logicalOf(raw.Usage.InputTokens, raw.Usage.OutputTokens, raw.Usage.CacheReadInputTokens, raw.Usage.CacheCreationInputTokens)
		if err != nil {
			return ClaudeResult{}, fmt.Errorf("claude result: usage: %w", err)
		}
		res.MainUsage, res.HasMainUsage = l, ok
	}
	if len(raw.ModelUsage) > 0 {
		res.ModelUsage = map[string]usage.Logical{}
		for model, mu := range raw.ModelUsage {
			l, ok, err := logicalOf(mu.InputTokens, mu.OutputTokens, mu.CacheReadInputTokens, mu.CacheCreationInputTokens)
			if err != nil {
				return ClaudeResult{}, fmt.Errorf("claude result: modelUsage %q: %w", model, err)
			}
			if !ok {
				// un modelo sin todos los campos: el total del arbol no es fiable
				res.ModelUsage = nil
				break
			}
			res.ModelUsage[model] = l
		}
	}
	return res, nil
}

// ReconcileClaude aplica la regla de precedencia con el detalle de
// OpenTelemetry y, si existe, el agregado del resultado. modelUsage incluye a
// los subagentes (alcance comprobado en la documentacion); usage no.
func ReconcileClaude(att Attribution, res *ClaudeResult) (usage.Totals, error) {
	in := usage.Input{Entries: att.Entries}
	if res != nil {
		if tree, ok := res.TreeTotal(); ok {
			in.Aggregate, in.Scope = &tree, usage.ScopeIncludesChildren
		} else if res.HasMainUsage {
			t := res.MainUsage.Total()
			in.Aggregate, in.Scope = &t, usage.ScopeParentOnly
		}
	}
	return usage.Reconcile(in)
}

// StreamStep es una solicitud del flujo stream-json, deduplicada por id de
// mensaje. La salida por paso es un marcador (la documentacion lo dice): Usage
// lleva solo la entrada y la caché.
type StreamStep struct {
	MessageID, Role, AgentID, ParentToolUseID string
	Usage                                     usage.Logical
}

// ClaudeStream es lo capturado de `--output-format stream-json`.
type ClaudeStream struct {
	Version, SessionID, PermissionMode, Model string
	Steps                                     []StreamStep
	Duplicates                                int
	Incomplete                                []string
	Retries                                   int
	Result                                    *ClaudeResult
	Events                                    []run.Event
}

type streamLine struct {
	Type              string  `json:"type"`
	Subtype           string  `json:"subtype"`
	ClaudeCodeVersion string  `json:"claude_code_version"`
	SessionID         string  `json:"session_id"`
	PermissionMode    string  `json:"permissionMode"`
	Model             string  `json:"model"`
	AgentID           string  `json:"agent_id"`
	ParentToolUseID   *string `json:"parent_tool_use_id"`
	Message           *struct {
		ID    string    `json:"id"`
		Usage *rawUsage `json:"usage"`
	} `json:"message"`
}

// ParseClaudeStream lee el flujo stream-json. Los mensajes de asistente que
// comparten message.id (llamadas a herramientas en paralelo) cuentan una vez.
func ParseClaudeStream(r io.Reader) (ClaudeStream, error) {
	var s ClaudeStream
	seen := map[string]bool{}
	err := forEachLine(r, func(n int, line []byte) error {
		var l streamLine
		if err := json.Unmarshal(line, &l); err != nil {
			return fmt.Errorf("claude stream: línea %d: %w", n, err)
		}
		typ := l.Type
		if l.Type == "system" && l.Subtype != "" {
			typ = "system/" + l.Subtype
		}
		s.Events = append(s.Events, run.Event{Type: typ, Raw: append(json.RawMessage(nil), line...)})
		switch typ {
		case "system/init":
			s.Version, s.SessionID, s.PermissionMode, s.Model = l.ClaudeCodeVersion, l.SessionID, l.PermissionMode, l.Model
		case "system/api_retry":
			s.Retries++
		case "assistant":
			if l.Message == nil || l.Message.ID == "" {
				return nil
			}
			if seen[l.Message.ID] {
				s.Duplicates++
				return nil
			}
			seen[l.Message.ID] = true
			if l.Message.Usage == nil || l.Message.Usage.InputTokens == nil || l.Message.Usage.CacheReadInputTokens == nil || l.Message.Usage.CacheCreationInputTokens == nil {
				s.Incomplete = append(s.Incomplete, fmt.Sprintf("mensaje %s (línea %d): uso incompleto", l.Message.ID, n))
				return nil
			}
			zero := int64(0)
			lg, ok, err := logicalOf(l.Message.Usage.InputTokens, &zero, l.Message.Usage.CacheReadInputTokens, l.Message.Usage.CacheCreationInputTokens)
			if err != nil || !ok {
				s.Incomplete = append(s.Incomplete, fmt.Sprintf("mensaje %s (línea %d): uso inválido", l.Message.ID, n))
				return nil
			}
			step := StreamStep{MessageID: l.Message.ID, AgentID: l.AgentID, Usage: lg}
			switch {
			case l.ParentToolUseID == nil:
				step.Role = RolePrincipal
			case l.AgentID != "":
				step.Role, step.ParentToolUseID = RoleWorker, *l.ParentToolUseID
			default:
				step.Role, step.ParentToolUseID = RoleUnattributed, *l.ParentToolUseID
			}
			s.Steps = append(s.Steps, step)
		case "result":
			res, err := ParseClaudeResult(line)
			if err != nil {
				return fmt.Errorf("línea %d: %w", n, err)
			}
			s.Result = &res
		}
		return nil
	})
	return s, err
}

const readsUnavailableClaude = "las rutas de archivo solo se exponen con OTEL_LOG_TOOL_DETAILS=1, un opt-in de contenido que las corridas medidas no activan"

// NewClaudeTraceLog crea el registro de trazas con las senales de la celda:
// las lecturas de archivos no estan disponibles (sin rutas); las llamadas a
// herramientas por rol exigen agent_id (>= 2.1.310) y una interfaz con la
// senal comprobada.
func NewClaudeTraceLog(i manifest.Interface, version string) (*trace.Log, error) {
	ok, err := manifest.VersionAtLeast(version, ClaudeMinAgentIDVersion)
	if err != nil {
		return nil, fmt.Errorf("claude: %w", err)
	}
	if i == manifest.InterfaceIDE {
		return trace.NewLog(false, false, "las señales de la extensión de IDE no están comprobadas con la extensión real"), nil
	}
	reason := readsUnavailableClaude
	if !ok {
		reason += "; sin agent_id (Claude Code < " + ClaudeMinAgentIDVersion + ") no se separan principal y worker"
	}
	return trace.NewLog(false, trace.Signal(ok), reason), nil
}

// FillClaudeTrace registra las llamadas a herramientas por rol: con agent_id,
// worker; sin el, principal (los eventos del principal no lo llevan).
func FillClaudeTrace(log *trace.Log, o ClaudeOTel, version string) error {
	for _, tr := range o.ToolResults {
		role := RolePrincipal
		if tr.AgentID != "" {
			role = RoleWorker
		}
		if err := log.RecordToolCall(role, tr.ToolName); err != nil {
			return err
		}
	}
	return nil
}
