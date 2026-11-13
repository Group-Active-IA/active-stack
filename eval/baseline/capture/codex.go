package capture

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/trace"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

// CodexParentAggregateScope es el alcance del agregado del hilo padre. La
// documentacion consultada (non-interactive-mode, app-server y subagents de
// learn.chatgpt.com, 2026-11-11) no dice si incluye a los hilos hijos: queda
// "no comprobado" y Reconcile solo confia en una coincidencia exacta con el
// detalle por hilo. Nunca se suman el agregado del padre y los hijos.
const CodexParentAggregateScope = usage.ScopeUnknown

// CodexTurn es un turno del flujo `codex exec --json`.
type CodexTurn struct {
	Completed, Failed bool
	Usage             *usage.Logical
}

// CodexToolCall es una llamada a herramienta (item) identificada por id.
type CodexToolCall struct{ ID, Type string }

// CodexThread es lo capturado de un flujo `codex exec --json` (un hilo).
type CodexThread struct {
	ID                       string
	Turns                    []CodexTurn
	ToolCalls                []CodexToolCall
	Unfinished               int
	DuplicateItems           int
	DuplicateTurnCompletions int
	Errors                   []run.Event
	Events                   []run.Event
}

// codexToolTypes son los items que cuentan como llamada a herramienta. Los
// agent_message, reasoning y plan_update no lo son.
var codexToolTypes = map[string]bool{"command_execution": true, "file_change": true, "mcp_tool_call": true, "web_search": true}

type codexLine struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id"`
	Item     *struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"item"`
	Usage *struct {
		InputTokens           *int64 `json:"input_tokens"`
		CachedInputTokens     *int64 `json:"cached_input_tokens"`
		OutputTokens          *int64 `json:"output_tokens"`
		ReasoningOutputTokens *int64 `json:"reasoning_output_tokens"`
	} `json:"usage"`
}

// ParseCodexExec lee un flujo JSONL de un hilo. Los eventos se conservan en
// crudo. Un turn.completed que no cierra ningun turno abierto se cuenta como
// repetido y no suma; un item.completed con un id ya visto, tambien.
func ParseCodexExec(r io.Reader) (CodexThread, error) {
	var th CodexThread
	var cur CodexTurn
	open := false
	seen := map[string]bool{}
	started := map[string]bool{}
	completed := map[string]bool{}
	err := forEachLine(r, func(n int, line []byte) error {
		var l codexLine
		if err := json.Unmarshal(line, &l); err != nil {
			return fmt.Errorf("codex exec: línea %d: %w", n, err)
		}
		ev := run.Event{Type: l.Type, Raw: append(json.RawMessage(nil), line...)}
		th.Events = append(th.Events, ev)
		switch l.Type {
		case "thread.started":
			switch {
			case th.ID == "":
				th.ID = l.ThreadID
			case th.ID != l.ThreadID:
				return fmt.Errorf("codex exec: línea %d: el flujo trae dos hilos (%s y %s); se espera un flujo por hilo", n, th.ID, l.ThreadID)
			}
		case "turn.started":
			if open {
				th.Turns = append(th.Turns, cur)
			}
			cur, open = CodexTurn{}, true
		case "turn.completed":
			if !open {
				th.DuplicateTurnCompletions++
				return nil
			}
			cur.Completed = true
			if u := l.Usage; u != nil && u.InputTokens != nil && u.CachedInputTokens != nil && u.OutputTokens != nil && u.ReasoningOutputTokens != nil {
				lg, err := usage.FromCodex(usage.CodexTurn{Input: *u.InputTokens, CachedInput: *u.CachedInputTokens, Output: *u.OutputTokens, Reasoning: *u.ReasoningOutputTokens})
				if err != nil {
					return fmt.Errorf("codex exec: línea %d: %w", n, err)
				}
				cur.Usage = &lg
			}
			th.Turns = append(th.Turns, cur)
			open = false
		case "turn.failed":
			th.Errors = append(th.Errors, ev)
			if open {
				cur.Failed = true
				th.Turns = append(th.Turns, cur)
				open = false
			}
		case "error":
			th.Errors = append(th.Errors, ev)
		case "item.started", "item.updated", "item.completed":
			if l.Item == nil || l.Item.ID == "" || !codexToolTypes[l.Item.Type] {
				return nil
			}
			id := l.Item.ID
			if !seen[id] {
				seen[id] = true
				th.ToolCalls = append(th.ToolCalls, CodexToolCall{ID: id, Type: l.Item.Type})
			}
			switch l.Type {
			case "item.started":
				started[id] = true
			case "item.completed":
				if completed[id] {
					th.DuplicateItems++
				}
				completed[id] = true
			}
		}
		return nil
	})
	if err != nil {
		return CodexThread{}, err
	}
	if open {
		th.Turns = append(th.Turns, cur)
	}
	for id := range started {
		if !completed[id] {
			th.Unfinished++
		}
	}
	return th, nil
}

// Tokens suma el uso de los turnos del hilo; no esta disponible si no hay
// turnos o si algun turno no informo uso (no se inventa un cero).
func (t CodexThread) Tokens() (usage.Logical, Availability) {
	if len(t.Turns) == 0 {
		return usage.Logical{}, Availability{Reason: "el flujo no trae turnos"}
	}
	var sum usage.Logical
	for i, tr := range t.Turns {
		if !tr.Completed || tr.Usage == nil {
			return usage.Logical{}, Availability{Reason: fmt.Sprintf("el turno %d no informó uso (sin turn.completed con usage)", i+1)}
		}
		sum.Input += tr.Usage.Input
		sum.Uncached += tr.Usage.Uncached
		sum.CacheRead += tr.Usage.CacheRead
		sum.Output += tr.Usage.Output
		sum.Reasoning += tr.Usage.Reasoning
	}
	return sum, Availability{Available: true}
}

// ReconcileCodex reconcilia el uso de los hilos con un agregado opcional del
// hilo padre. threads[0] es el hilo principal; parents mapea hilo hijo -> hilo
// padre. Un hilo que no se puede correlacionar con un padre conocido queda sin
// atribuir y la atribucion se marca parcial; un hilo sin uso disponible no
// aporta cifras. El agregado nunca se suma a los hilos.
func ReconcileCodex(threads []CodexThread, parents map[string]string, aggregate *int64) (usage.Totals, Attribution, error) {
	if len(threads) == 0 {
		return usage.Totals{}, Attribution{}, errors.New("codex: no hay hilos que reconciliar")
	}
	known := map[string]bool{}
	for _, t := range threads {
		known[t.ID] = true
	}
	att := Attribution{ByAgent: map[string]int64{}, PerWorker: Availability{Available: true}}
	for i, t := range threads {
		role := RoleUnattributed
		p, ok := parents[t.ID]
		switch {
		case i == 0:
			role = RolePrincipal
		case ok && known[p]:
			role = RoleWorker
		default:
			att.Partial = true
			att.Reasons = append(att.Reasons, fmt.Sprintf("hilo %s sin hilo padre correlacionable: su consumo queda sin atribuir", t.ID))
		}
		tok, av := t.Tokens()
		if !av.Available {
			att.Partial = true
			att.Reasons = append(att.Reasons, fmt.Sprintf("hilo %s: %s", t.ID, av.Reason))
			continue
		}
		if role == RoleWorker {
			att.ByAgent[t.ID] += tok.Total()
		}
		att.Entries = append(att.Entries, usage.Entry{Role: role, ID: t.ID, Tokens: tok.Total()})
	}
	if att.Partial {
		att.PerWorker = Availability{Reason: "hay hilos sin correlacionar o sin uso"}
	}
	in := usage.Input{Entries: att.Entries}
	if aggregate != nil {
		in.Aggregate, in.Scope = aggregate, CodexParentAggregateScope
	}
	tot, err := usage.Reconcile(in)
	if err != nil {
		return usage.Totals{}, Attribution{}, fmt.Errorf("codex: %w", err)
	}
	return tot, att, nil
}

// NewCodexTraceLog crea el registro de trazas con las senales de la celda. Las
// lecturas de archivos no son items propios (se hacen con command_execution).
func NewCodexTraceLog(i manifest.Interface) (*trace.Log, error) {
	switch i {
	case manifest.InterfaceTerminal:
		return trace.NewLog(false, true, "las lecturas de archivos no se exponen como items propios: se hacen mediante command_execution"), nil
	case manifest.InterfaceIDE:
		return trace.NewLog(false, false, "las señales de la extensión de IDE no están comprobadas con la extensión real"), nil
	}
	return nil, fmt.Errorf("interfaz %q fuera de la matriz", i)
}

// FillCodexTrace registra las llamadas a herramientas del hilo con el rol dado.
func FillCodexTrace(log *trace.Log, t CodexThread, role string) error {
	for _, c := range t.ToolCalls {
		if err := log.RecordToolCall(role, c.Type); err != nil {
			return err
		}
	}
	return nil
}
