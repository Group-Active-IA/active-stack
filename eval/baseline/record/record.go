// Package record emite los registros por intento y por corrida en JSONL
// (tarea 2.7 de opsx-baseline), con metadatos de cobertura, la lista sensible
// mínima y huellas. La lista sensible completa, la captura y la retención son
// de 05; este paquete aplica el mínimo que fija design.md ("Datos sensibles"):
// correo, UUID de cuenta, identificador de usuario, rutas del host, detalle de
// herramientas y cuerpos de solicitudes. Las huellas siguen la definición de
// 06: sha256 con el algoritmo nombrado en el campo.
package record

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/trace"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

// Category es una categoría de la lista sensible mínima.
type Category string

// Categorías sensibles.
const (
	Email       Category = "email"
	AccountUUID Category = "account_uuid"
	UserID      Category = "user_id"
	HostPath    Category = "host_path"
	ToolDetail  Category = "tool_detail"
	RequestBody Category = "request_body"
	// Unparsed marca un evento que no es JSON válido: no se conserva nada.
	Unparsed Category = "unparsed"
)

// Fingerprint es una huella con el algoritmo nombrado.
type Fingerprint struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// SHA256 calcula la huella sha256 de data.
func SHA256(data []byte) Fingerprint {
	h := sha256.Sum256(data)
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(h[:])}
}

// EvidenceRef referencia la evidencia original (local, no versionada): su
// referencia y huella, o el motivo de su ausencia.
type EvidenceRef struct {
	Ref         string       `json:"ref,omitempty"`
	Fingerprint *Fingerprint `json:"fingerprint,omitempty"`
	Absent      string       `json:"absent_reason,omitempty"`
}

var absPath = regexp.MustCompile(`^(?:/|\\\\|[A-Za-z]:[\\/])`)

// NewEvidence arma la referencia. Con data, calcula la huella; una referencia
// que sea ruta absoluta del host se reduce a "local:<nombre>". Sin data exige
// el motivo de ausencia.
func NewEvidence(ref string, data []byte, absentReason string) (EvidenceRef, error) {
	if data == nil {
		if strings.TrimSpace(absentReason) == "" {
			return EvidenceRef{}, errors.New("evidencia sin contenido: falta el motivo de ausencia")
		}
		return EvidenceRef{Ref: safeRef(ref), Absent: absentReason}, nil
	}
	fp := SHA256(data)
	return EvidenceRef{Ref: safeRef(ref), Fingerprint: &fp}, nil
}

func safeRef(ref string) string {
	if ref == "" {
		return ""
	}
	if red, _ := RedactText(ref); red != ref || absPath.MatchString(ref) {
		base := ref[strings.LastIndexAny(ref, `/\`)+1:]
		return "local:" + base
	}
	return ref
}

var (
	reEmail = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9\-]+(?:\.[A-Za-z0-9\-]+)*\.[A-Za-z]{2,}`)
	reUUID  = regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)
	reWin   = regexp.MustCompile(`(^|[^A-Za-z0-9])[A-Za-z]:[\\/][^\s"'<>|,;)\]]*`)
	reUNC   = regexp.MustCompile(`(^|[\s"'=(,\[])\\\\[A-Za-z0-9_.$\-]+\\[^\s"']*`)
	rePosix = regexp.MustCompile(`(^|[\s"'=(,\[])/(?:home|Users|root|tmp|var|etc|opt|mnt|usr|private|Volumes)/[^\s"'<>|,;)\]]*`)
	reTilde = regexp.MustCompile(`(^|[\s"'=(,\[])~/[^\s"']*`)
)

func hasCat(cs []Category, want Category) bool {
	for _, c := range cs {
		if c == want {
			return true
		}
	}
	return false
}

func marker(c Category) string { return "[redacted:" + string(c) + "]" }

func sub(s string, re *regexp.Regexp, c Category, prefix bool, cats *[]Category) string {
	return re.ReplaceAllStringFunc(s, func(m string) string {
		*cats = append(*cats, c)
		if !prefix {
			return marker(c)
		}
		loc := re.FindStringSubmatchIndex(m)
		return m[:loc[3]] + marker(c) // conserva el carácter previo
	})
}

func redactText(s string, keepUUID bool) (string, []Category) {
	var cats []Category
	s = sub(s, reEmail, Email, false, &cats)
	if !keepUUID {
		s = sub(s, reUUID, AccountUUID, false, &cats)
	}
	s = sub(s, reWin, HostPath, true, &cats)
	s = sub(s, reUNC, HostPath, true, &cats)
	s = sub(s, rePosix, HostPath, true, &cats)
	s = sub(s, reTilde, HostPath, true, &cats)
	return s, cats
}

// RedactText reemplaza correos, UUID y rutas del host de un texto libre por
// marcas [redacted:<categoría>]. Devuelve una entrada por redacción.
func RedactText(s string) (string, []Category) { return redactText(s, false) }

func normKey(k string) string {
	return strings.NewReplacer(".", "_", "-", "_").Replace(strings.ToLower(k))
}

var keyCategories = map[string]Category{
	"email": Email, "user_email": Email, "e_mail": Email,
	"account_uuid": AccountUUID, "user_account_uuid": AccountUUID, "account_id": AccountUUID,
	"organization_id": AccountUUID, "organization_uuid": AccountUUID,
	"user_id": UserID, "userid": UserID,
	"tool_input": ToolDetail, "tool_args": ToolDetail, "tool_arguments": ToolDetail, "arguments": ToolDetail,
	"tool_parameters": ToolDetail, "tool_output": ToolDetail, "tool_result": ToolDetail, "tool_response": ToolDetail,
	"request_body": RequestBody, "body": RequestBody, "messages": RequestBody, "prompt": RequestBody,
	"user_prompt": RequestBody, "system_prompt": RequestBody,
}

func keyCategory(key, parent string) (Category, bool) {
	n := normKey(key)
	if n == "id" && normKey(parent) == "user" {
		return UserID, true
	}
	c, ok := keyCategories[n]
	return c, ok
}

func redactValue(v any, parent string, cats *[]Category) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			if c, ok := keyCategory(k, parent); ok {
				*cats = append(*cats, c)
				out[k] = marker(c)
				continue
			}
			out[k] = redactValue(val, k, cats)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = redactValue(val, parent, cats)
		}
		return out
	case string:
		s, c := RedactText(x)
		*cats = append(*cats, c...)
		return s
	}
	return v
}

// RedactJSON quita de un documento JSON los valores de claves sensibles y
// redacta el texto de los demás. Los miembros benignos se conservan. Si raw no
// es JSON válido (o trae más de un documento) no se puede saber qué claves son
// sensibles: se descarta entero y se devuelve la marca [redacted:unparsed].
func RedactJSON(raw json.RawMessage) (json.RawMessage, []Category) {
	var cats []Category
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil || dec.More() {
		b, _ := json.Marshal(marker(Unparsed))
		return b, []Category{Unparsed}
	}
	b, _ := json.Marshal(redactValue(v, "", &cats))
	return b, cats
}

// Counts es el conteo de una señal (lecturas o llamadas a herramientas):
// solo cantidades, nunca el detalle de las llamadas.
type Counts struct {
	Available bool           `json:"available"`
	Reason    string         `json:"reason,omitempty"`
	Total     int            `json:"total"`
	ByType    map[string]int `json:"by_type,omitempty"`
	ByRole    map[string]int `json:"by_role,omitempty"`
	Unique    int            `json:"unique,omitempty"`
}

func toCounts(c trace.Counts) Counts {
	out := Counts{Available: c.Available, Reason: c.Reason, Total: c.Total, Unique: c.Unique}
	if c.Available {
		out.ByType = copyInts(c.ByType)
		out.ByRole = copyInts(c.ByRole)
	}
	return out
}

func copyInts(m map[string]int) map[string]int {
	if len(m) == 0 {
		return nil
	}
	o := make(map[string]int, len(m))
	for k, v := range m {
		o[k] = v
	}
	return o
}

// Consumption es el consumo de tokens de un intento.
type Consumption struct {
	Available       bool             `json:"available"`
	Reason          string           `json:"reason,omitempty"`
	Total           int64            `json:"total"`
	ByRole          map[string]int64 `json:"by_role,omitempty"`
	Unattributed    int64            `json:"unattributed"`
	Partial         bool             `json:"partial"`
	FullCoverage    bool             `json:"full_coverage"`
	AggregateFigure *int64           `json:"aggregate_figure,omitempty"`
	DetailFigure    *int64           `json:"detail_figure,omitempty"`
}

// Timing son los tiempos de un intento; null es no disponible.
type Timing struct {
	ElapsedMs      int64  `json:"elapsed_ms"`
	WaitsMs        *int64 `json:"waits_ms"`
	AdjustedMs     *int64 `json:"adjusted_ms"`
	AdjustedReason string `json:"adjusted_reason,omitempty"`
}

// Coverage son los metadatos de cobertura: por señal, disponible o no; el uso,
// completo, parcial o no disponible; y cuántas redacciones se aplicaron por
// categoría.
type Coverage struct {
	Reads      string         `json:"reads"`
	ToolCalls  string         `json:"tool_calls"`
	Usage      string         `json:"usage"`
	Timing     string         `json:"adjusted_time"`
	Redactions map[string]int `json:"redactions,omitempty"`
}

// Valores de cobertura.
const (
	covAvailable   = "disponible"
	covUnavailable = "no_disponible"
	covPartial     = "parcial"
	covComplete    = "completa"
)

// AttemptRecord es una línea JSONL por intento.
type AttemptRecord struct {
	Kind            string             `json:"kind"`
	ProtocolVersion string             `json:"protocol_version"`
	RunID           string             `json:"run_id"`
	SessionID       string             `json:"session_id,omitempty"`
	AgentID         string             `json:"agent_id,omitempty"`
	Attempt         int                `json:"attempt"`
	Cause           string             `json:"cause"`
	Status          string             `json:"status"`
	Reason          string             `json:"reason,omitempty"`
	ExitCode        *int               `json:"exit_code"`
	Events          int                `json:"events"`
	ErrorEvents     []json.RawMessage  `json:"error_events,omitempty"`
	Consumption     Consumption        `json:"consumption"`
	Reads           Counts             `json:"reads"`
	ToolCalls       Counts             `json:"tool_calls"`
	Cache           *usage.CacheRecord `json:"cache,omitempty"`
	Timing          Timing             `json:"timing"`
	Coverage        Coverage           `json:"coverage"`
	Evidence        EvidenceRef        `json:"evidence"`
}

// RunRecord es una línea JSONL por corrida.
type RunRecord struct {
	Kind            string        `json:"kind"`
	ProtocolVersion string        `json:"protocol_version"`
	RunID           string        `json:"run_id"`
	CohortID        string        `json:"cohort_id"`
	TaskID          string        `json:"task_id"`
	Repetition      int           `json:"repetition"`
	Status          string        `json:"status"`
	Attempts        int           `json:"attempts"`
	Coverage        Coverage      `json:"coverage"`
	Evidence        []EvidenceRef `json:"evidence"`
}

// AttemptInput es lo que se suma a un intento de run: identificadores opacos de
// sesión y agente (el adaptador no debe poner ahí datos de cuenta), el consumo
// reconciliado, el estado de caché y la evidencia original.
type AttemptInput struct {
	Attempt   run.Attempt
	SessionID string
	AgentID   string
	Totals    *usage.Totals
	Cache     *usage.CacheRecord
	Evidence  EvidenceRef
}

// Builder arma los registros de un protocolo.
type Builder struct{ ProtocolVersion string }

func ms(d interface{ Milliseconds() int64 }) int64 { return d.Milliseconds() }

// Records arma el registro de la corrida y uno por intento, redactando todo
// texto libre y los eventos de error.
func (b Builder) Records(r run.Run, in []AttemptInput) (RunRecord, []AttemptRecord, error) {
	if strings.TrimSpace(b.ProtocolVersion) == "" {
		return RunRecord{}, nil, errors.New("falta la versión de protocolo")
	}
	if len(in) != len(r.Attempts) {
		return RunRecord{}, nil, fmt.Errorf("%d entradas para %d intentos", len(in), len(r.Attempts))
	}
	runID := fmt.Sprintf("%s/%s#%d", r.CohortID, r.TaskID, r.Repetition)
	rr := RunRecord{Kind: "run", ProtocolVersion: b.ProtocolVersion, RunID: runID, CohortID: r.CohortID,
		TaskID: r.TaskID, Repetition: r.Repetition, Status: string(r.Status), Attempts: len(r.Attempts),
		Evidence: []EvidenceRef{}}
	rr.Coverage = Coverage{Reads: covAvailable, ToolCalls: covAvailable, Usage: covComplete, Timing: covAvailable, Redactions: map[string]int{}}
	var out []AttemptRecord
	for i, a := range r.Attempts {
		x := in[i]
		if x.Attempt.N != a.N {
			return RunRecord{}, nil, fmt.Errorf("la entrada %d es del intento %d, no del %d", i, x.Attempt.N, a.N)
		}
		red := map[string]int{}
		note := func(cs []Category) {
			for _, c := range cs {
				red[string(c)]++
			}
		}
		text := func(s string) string {
			t, cs := RedactText(s)
			note(cs)
			return t
		}
		id := func(s string) string { // los UUID de sesión y agente son opacos y se conservan
			t, cs := redactText(s, true)
			note(cs)
			return t
		}
		ar := AttemptRecord{Kind: "attempt", ProtocolVersion: b.ProtocolVersion, RunID: runID,
			SessionID: id(x.SessionID), AgentID: id(x.AgentID), Attempt: a.N, Cause: text(a.Cause),
			Status: string(a.Status), Reason: text(a.Reason), Events: a.Events,
			Reads: toCounts(a.Trace.Reads), ToolCalls: toCounts(a.Trace.ToolCalls), Evidence: x.Evidence}
		if a.ExitCode != nil {
			c := *a.ExitCode
			ar.ExitCode = &c
		}
		for _, ev := range a.ErrorEvents {
			raw := ev.Raw
			if len(raw) == 0 {
				raw, _ = json.Marshal(map[string]string{"type": ev.Type})
			}
			rd, cs := RedactJSON(raw)
			if hasCat(cs, Unparsed) { // conserva solo el tipo del evento
				rd, _ = json.Marshal(map[string]string{"type": text(ev.Type), "detail": marker(Unparsed)})
			}
			note(cs)
			ar.ErrorEvents = append(ar.ErrorEvents, rd)
		}
		if x.Cache != nil {
			c := *x.Cache
			ar.Cache = &c
		}
		cov := Coverage{Reads: availability(ar.Reads.Available), ToolCalls: availability(ar.ToolCalls.Available), Usage: covUnavailable, Timing: covUnavailable}
		if t := x.Totals; t != nil {
			ar.Consumption = Consumption{Available: true, Total: t.Total, ByRole: copyInt64s(t.ByRole), Unattributed: t.Unattributed,
				Partial: t.Partial, FullCoverage: t.FullCoverage, AggregateFigure: t.AggregateFigure, DetailFigure: t.DetailFigure}
			cov.Usage = covComplete
			if t.Partial || !t.FullCoverage {
				cov.Usage = covPartial
			}
		} else {
			ar.Consumption = Consumption{Reason: "sin fuente de consumo para este intento"}
		}
		ar.Timing = Timing{ElapsedMs: ms(a.Timing.Elapsed)}
		if a.Timing.WaitTotal.Available {
			w := ms(a.Timing.WaitTotal.Value)
			ar.Timing.WaitsMs = &w
		}
		if a.Timing.Adjusted.Available {
			j := ms(a.Timing.Adjusted.Value)
			ar.Timing.AdjustedMs = &j
			cov.Timing = covAvailable
		} else {
			ar.Timing.AdjustedReason = a.Timing.Adjusted.Reason
		}
		cov.Redactions = red
		ar.Coverage = cov
		out = append(out, ar)
		rr.Evidence = append(rr.Evidence, x.Evidence)
		rr.Coverage.Reads = weakest(rr.Coverage.Reads, cov.Reads)
		rr.Coverage.ToolCalls = weakest(rr.Coverage.ToolCalls, cov.ToolCalls)
		rr.Coverage.Usage = weakest(rr.Coverage.Usage, cov.Usage)
		rr.Coverage.Timing = weakest(rr.Coverage.Timing, cov.Timing)
		for k, v := range red {
			rr.Coverage.Redactions[k] += v
		}
	}
	return rr, out, nil
}

func availability(ok bool) string {
	if ok {
		return covAvailable
	}
	return covUnavailable
}

func copyInt64s(m map[string]int64) map[string]int64 {
	if len(m) == 0 {
		return nil
	}
	o := make(map[string]int64, len(m))
	for k, v := range m {
		o[k] = v
	}
	return o
}

func rank(s string) int {
	switch s {
	case covUnavailable:
		return 0
	case covPartial:
		return 1
	}
	return 2
}

// weakest devuelve la cobertura más débil de dos.
func weakest(a, b string) string {
	if rank(b) < rank(a) {
		return b
	}
	return a
}

// WriteJSONL escribe un objeto JSON por línea.
func WriteJSONL(w io.Writer, rec ...any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, r := range rec {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
