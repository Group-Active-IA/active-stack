// Package usage normaliza el consumo de tokens de Claude y Codex, implementa la
// sonda de sesión normal y el registro de caché, y reconcilia padre e hijos
// (tarea 2.6 de opsx-baseline; design.md, "Señales y fuentes" y "Reconciliación
// de padre e hijos"). Trabaja con eventos ya extraídos; la captura real es 2.2.
package usage

import (
	"errors"
	"fmt"
	"strings"
)

// ClaudeRequest es el uso de una solicitud de Claude Code: la entrada
// documentada excluye las lecturas y creaciones de caché.
type ClaudeRequest struct{ Input, CacheRead, CacheCreation, Output int64 }

// CodexTurn es el uso de un turno de Codex: la entrada cacheada forma parte de
// la entrada y el razonamiento se informa aparte.
type CodexTurn struct{ Input, CachedInput, Output, Reasoning int64 }

// Logical es el uso normalizado. Input es la entrada lógica (incluye lecturas y
// creaciones de caché); Uncached, la parte sin caché.
type Logical struct{ Input, Uncached, CacheRead, CacheCreation, Output, Reasoning int64 }

// Total es la entrada lógica más la salida. El razonamiento se asume incluido
// en la salida (suposición a confirmar con la versión fijada de Codex).
func (l Logical) Total() int64 { return l.Input + l.Output }

func anyNegative(vs ...int64) bool {
	for _, v := range vs {
		if v < 0 {
			return true
		}
	}
	return false
}

// FromClaude suma lectura y creación de caché a la entrada documentada.
func FromClaude(r ClaudeRequest) (Logical, error) {
	if anyNegative(r.Input, r.CacheRead, r.CacheCreation, r.Output) {
		return Logical{}, errors.New("tokens negativos")
	}
	return Logical{Input: r.Input + r.CacheRead + r.CacheCreation, Uncached: r.Input,
		CacheRead: r.CacheRead, CacheCreation: r.CacheCreation, Output: r.Output}, nil
}

// FromCodex deriva la entrada sin caché por diferencia.
func FromCodex(t CodexTurn) (Logical, error) {
	if anyNegative(t.Input, t.CachedInput, t.Output, t.Reasoning) {
		return Logical{}, errors.New("tokens negativos")
	}
	if t.CachedInput > t.Input {
		return Logical{}, fmt.Errorf("entrada cacheada %d mayor que la entrada %d", t.CachedInput, t.Input)
	}
	return Logical{Input: t.Input, Uncached: t.Input - t.CachedInput, CacheRead: t.CachedInput,
		Output: t.Output, Reasoning: t.Reasoning}, nil
}

// Metric es un valor que puede no estar disponible; ausente no es cero.
type Metric struct {
	Available bool
	Value     int64
	Basis     string
	Reason    string
}

// Bases de una métrica.
const (
	BasisRequest    = "solicitud"
	BasisTurn       = "turno"
	BasisApproxTurn = "aproximacion_total_turno"
)

// CacheState es el estado de caché previsto u observado.
type CacheState string

// Estados de caché.
const (
	CacheCold    CacheState = "fria"
	CacheWarm    CacheState = "con_lectura_prevista"
	CacheRead    CacheState = "con_lectura"
	CacheCreate  CacheState = "solo_creacion"
	CacheNone    CacheState = "sin_cache"
	CacheUnknown CacheState = "no_disponible"
)

// CacheRecord conserva el estado previsto y el observado; este último nunca se
// reescribe para coincidir con la previsión.
type CacheRecord struct {
	Expected, Observed CacheState
	Diverges           bool
}

// Sample es lo que la sonda captura de una sesión nueva.
type Sample struct {
	Request             *Logical // desglose nativo de la primera solicitud
	TurnTotal           *int64   // total del turno
	TurnIsSingleRequest bool     // comprobado: el turno fue una única solicitud
	Expected            CacheState
}

// Reading es el resultado de la sonda.
type Reading struct {
	RequestInput, TurnTotal Metric
	Cache                   CacheRecord
}

// Probe distingue la entrada de una solicitud del total del turno. Con solo el
// total del turno, lo usa como aproximación identificada únicamente si se
// comprobó que representa una sola solicitud.
func Probe(s Sample) Reading {
	var r Reading
	switch {
	case s.Request != nil:
		r.RequestInput = Metric{Available: true, Value: s.Request.Input, Basis: BasisRequest}
	case s.TurnTotal != nil && s.TurnIsSingleRequest:
		r.RequestInput = Metric{Available: true, Value: *s.TurnTotal, Basis: BasisApproxTurn}
	case s.TurnTotal != nil:
		r.RequestInput = Metric{Reason: "solo hay total de turno y no se comprobó que sea una única solicitud"}
	default:
		r.RequestInput = Metric{Reason: "la interfaz no expone la entrada de la solicitud"}
	}
	if s.TurnTotal != nil {
		r.TurnTotal = Metric{Available: true, Value: *s.TurnTotal, Basis: BasisTurn}
	} else {
		r.TurnTotal = Metric{Reason: "la interfaz no expone el total del turno"}
	}
	r.Cache = CacheRecord{Expected: s.Expected, Observed: CacheUnknown}
	if s.Request != nil {
		switch {
		case s.Request.CacheRead > 0:
			r.Cache.Observed = CacheRead
		case s.Request.CacheCreation > 0:
			r.Cache.Observed = CacheCreate
		default:
			r.Cache.Observed = CacheNone
		}
		r.Cache.Diverges = (s.Expected == CacheCold && r.Cache.Observed == CacheRead) ||
			(s.Expected == CacheWarm && r.Cache.Observed != CacheRead)
	}
	return r
}

// Scope es el alcance comprobado de un agregado.
type Scope string

// Alcances de un agregado.
const (
	ScopeIncludesChildren Scope = "incluye_hijos"
	ScopeParentOnly       Scope = "solo_padre"
	ScopeUnknown          Scope = "no_comprobado"
)

// Entry es el consumo de un evento por solicitud o turno con su rol.
type Entry struct {
	Role, ID string
	Tokens   int64
}

// Input es lo que se reconcilia: un agregado opcional con su alcance y el
// detalle por rol. El `total_tokens` del cierre de un subagente no es una
// Entry: solo cubre su última solicitud.
type Input struct {
	Aggregate *int64
	Scope     Scope
	Entries   []Entry
}

// Totals es el resultado de la precedencia de la reconciliación.
type Totals struct {
	Total                         int64
	ByRole                        map[string]int64
	Unattributed                  int64
	Partial, FullCoverage         bool
	AggregateFigure, DetailFigure *int64
}

// Reconcile aplica la regla de precedencia del diseño.
func Reconcile(in Input) (Totals, error) {
	if in.Aggregate != nil {
		if *in.Aggregate < 0 {
			return Totals{}, errors.New("agregado negativo")
		}
		switch in.Scope {
		case ScopeIncludesChildren, ScopeParentOnly, ScopeUnknown:
		default:
			return Totals{}, fmt.Errorf("alcance de agregado inválido %q", in.Scope)
		}
	}
	t := Totals{ByRole: map[string]int64{}}
	var detail int64
	for _, e := range in.Entries {
		if e.Tokens < 0 {
			return Totals{}, errors.New("entrada con tokens negativos")
		}
		if strings.TrimSpace(e.Role) == "" {
			return Totals{}, errors.New("entrada sin rol")
		}
		t.ByRole[e.Role] += e.Tokens
		detail += e.Tokens
	}
	if in.Aggregate == nil || in.Scope == ScopeParentOnly {
		// Sin agregado de mayor alcance (o uno que no incluye a los hijos),
		// el total es la suma del detalle.
		t.Total, t.FullCoverage = detail, true
		if in.Aggregate != nil {
			t.AggregateFigure = in.Aggregate
		}
		return t, nil
	}
	agg := *in.Aggregate
	t.Total = agg
	switch {
	case detail == agg:
		t.FullCoverage = true
	case detail < agg && in.Scope == ScopeIncludesChildren:
		t.Unattributed, t.Partial = agg-detail, true
	default: // detalle mayor, o diferencia con alcance no comprobado
		d := detail
		t.Partial, t.AggregateFigure, t.DetailFigure = true, in.Aggregate, &d
	}
	return t, nil
}
