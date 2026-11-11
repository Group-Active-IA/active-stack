// Package summary genera el resumen por cohorte (tarea 3.1 de opsx-baseline;
// design.md, "Denominadores", "API y cuota" y "Estructura de resultados").
//
// Cálculo, por cohorte ejecutable del plan:
//
//   - Iniciadas: toda corrida de la cohorte, sea cual sea su estado. Las celdas
//     incompatibles no tienen corridas, quedan fuera del denominador y se listan
//     aparte con su motivo (una corrida en una celda incompatible es un error).
//   - Aceptadas, rechazadas por calidad, bloqueadas por permisos, timeout, fallo
//     del cliente e interrumpidas se cuentan por separado. Las corridas
//     pendientes de revisor no son aceptadas ni rechazadas: figuran aparte y
//     dejan el resumen como provisional.
//   - Costo total, por medida (tokens como consumo técnico, dinero en API, cuota
//     en suscripción): suma de todos los intentos de todas las corridas
//     iniciadas, incluidos rechazos, bloqueos, fallos y reintentos.
//   - Costo por tarea aceptada: costo total / aceptadas. Sin aceptadas, el
//     total se publica y el cociente queda sin valor. Si algún intento no tiene
//     la medida o es parcial, el cociente queda sin valor y se publica la
//     cobertura (esperados, contados, parciales).
//   - Cada cohorte es una fila propia (cliente, versión, interfaz real,
//     modalidad, variante): API, suscripción e interfaces nunca se mezclan.
//
// Las tareas y pesos del manifiesto son los mismos para ambos brazos y no
// entran en el cociente de esta tarea. El resultado no depende del orden de las
// corridas: las sumas se hacen sobre valores ordenados.
package summary

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
	"github.com/Group-Active-IA/active-stack/eval/baseline/run"
	"github.com/Group-Active-IA/active-stack/eval/baseline/usage"
)

// Amount es una cantidad que puede no estar disponible (con motivo). Partial
// marca una cifra cuyo alcance no está comprobado.
type Amount struct {
	Available bool
	Value     float64
	Partial   bool
	Reason    string
}

// Cost es el consumo de un intento por medida.
type Cost struct{ Tokens, Money, Quota Amount }

// RunResult es una corrida iniciada con su estado final y el costo de cada intento.
type RunResult struct {
	CohortID   string
	TaskID     string
	Repetition int
	Status     manifest.RunStatus
	Attempts   []Cost
}

// Measure es el agregado de una medida en una cohorte.
type Measure struct {
	Total       Amount
	Counted     int // intentos con la medida disponible
	Partial     int // de ellos, con cifra parcial
	Expected    int // intentos de la cohorte
	Complete    bool
	PerAccepted Amount
}

// CohortSummary es una fila del resumen.
type CohortSummary struct {
	CohortID, Client, Version, Interface, Modality, Variant string
	Started, Accepted, Rejected, Blocked                    int
	TimedOut, ClientFailed, Interrupted, Pending            int
	Attempts                                                int
	Provisional                                             bool
	Tokens, Money, Quota                                    Measure
}

// Report es el resumen de una medición.
type Report struct {
	ProtocolVersion string
	Cohorts         []CohortSummary
	Incompatible    []manifest.Incompatible
}

// TokensFromTotals convierte un consumo reconciliado. La cifra es parcial solo
// cuando el detalle excede el total o su alcance no está comprobado; un detalle
// menor deja la atribución parcial pero el total es confiable.
func TokensFromTotals(t usage.Totals) Amount {
	return Amount{Available: true, Value: float64(t.Total), Partial: t.DetailFigure != nil}
}

// FromRun arma el resultado de una corrida con un costo por intento.
func FromRun(r run.Run, costs []Cost) (RunResult, error) {
	if len(costs) != len(r.Attempts) {
		return RunResult{}, fmt.Errorf("%d costos para %d intentos", len(costs), len(r.Attempts))
	}
	return RunResult{CohortID: r.CohortID, TaskID: r.TaskID, Repetition: r.Repetition, Status: r.Status,
		Attempts: append([]Cost(nil), costs...)}, nil
}

func knownStatus(s manifest.RunStatus) bool {
	return s.Valid() || s == run.StatusPendingReview || s == run.StatusInterrupted
}

func validAmount(a Amount) error {
	if a.Available && (math.IsNaN(a.Value) || math.IsInf(a.Value, 0) || a.Value < 0) {
		return fmt.Errorf("cantidad inválida %v", a.Value)
	}
	return nil
}

// Compute arma el resumen. Valida que cada corrida pertenezca a una cohorte
// ejecutable, que no haya corridas repetidas y que los estados y cantidades
// sean válidos.
func Compute(m manifest.Manifest, p manifest.Plan, runs []RunResult) (Report, error) {
	meta := map[string]manifest.Cohort{}
	for _, c := range m.Cohorts {
		meta[c.ID] = c
	}
	exec := map[string]manifest.Planned{}
	for _, pc := range p.Executable {
		exec[pc.CohortID] = pc
	}
	incompatible := map[string]bool{}
	for _, in := range p.Incompatible {
		incompatible[in.CohortID] = true
	}
	byCohort := map[string][]RunResult{}
	seen := map[string]bool{}
	for _, r := range runs {
		if incompatible[r.CohortID] {
			return Report{}, fmt.Errorf("la cohorte %q es incompatible: no puede tener corridas", r.CohortID)
		}
		if _, ok := exec[r.CohortID]; !ok {
			return Report{}, fmt.Errorf("la cohorte %q no está en el plan ejecutable", r.CohortID)
		}
		key := fmt.Sprintf("%s\x00%s\x00%d", r.CohortID, r.TaskID, r.Repetition)
		if seen[key] {
			return Report{}, fmt.Errorf("corrida repetida: %s/%s#%d", r.CohortID, r.TaskID, r.Repetition)
		}
		seen[key] = true
		if !knownStatus(r.Status) {
			return Report{}, fmt.Errorf("estado de corrida desconocido %q", r.Status)
		}
		if len(r.Attempts) == 0 {
			return Report{}, fmt.Errorf("corrida %s/%s#%d iniciada sin intentos", r.CohortID, r.TaskID, r.Repetition)
		}
		for _, a := range r.Attempts {
			for _, am := range []Amount{a.Tokens, a.Money, a.Quota} {
				if err := validAmount(am); err != nil {
					return Report{}, fmt.Errorf("corrida %s/%s#%d: %w", r.CohortID, r.TaskID, r.Repetition, err)
				}
			}
		}
		byCohort[r.CohortID] = append(byCohort[r.CohortID], r)
	}
	rep := Report{ProtocolVersion: m.ProtocolVersion}
	for _, pc := range p.Executable {
		c := meta[pc.CohortID]
		s := CohortSummary{CohortID: pc.CohortID, Client: string(c.Cell.Client), Version: c.Cell.Version,
			Interface: string(pc.Interface), Modality: string(c.Cell.Modality), Variant: m.Variant}
		var tk, mo, qu []Amount
		for _, r := range byCohort[pc.CohortID] {
			s.Started++
			switch r.Status {
			case manifest.RunAccepted:
				s.Accepted++
			case manifest.RunRejectedQuality:
				s.Rejected++
			case manifest.RunBlockedPermissions:
				s.Blocked++
			case manifest.RunTimeout:
				s.TimedOut++
			case manifest.RunClientFailure:
				s.ClientFailed++
			case run.StatusInterrupted:
				s.Interrupted++
			case run.StatusPendingReview:
				s.Pending++
			}
			for _, a := range r.Attempts {
				s.Attempts++
				tk, mo, qu = append(tk, a.Tokens), append(mo, a.Money), append(qu, a.Quota)
			}
		}
		s.Provisional = s.Pending > 0
		s.Tokens, s.Money, s.Quota = measure(tk, s.Accepted), measure(mo, s.Accepted), measure(qu, s.Accepted)
		rep.Cohorts = append(rep.Cohorts, s)
	}
	sort.Slice(rep.Cohorts, func(i, j int) bool {
		a, b := rep.Cohorts[i], rep.Cohorts[j]
		ka := []string{a.Client, a.Interface, a.Modality, a.Variant, a.Version, a.CohortID}
		kb := []string{b.Client, b.Interface, b.Modality, b.Variant, b.Version, b.CohortID}
		for x := range ka {
			if ka[x] != kb[x] {
				return ka[x] < kb[x]
			}
		}
		return false
	})
	rep.Incompatible = append([]manifest.Incompatible(nil), p.Incompatible...)
	sort.Slice(rep.Incompatible, func(i, j int) bool { return rep.Incompatible[i].CohortID < rep.Incompatible[j].CohortID })
	return rep, nil
}

const noReason = "no disponible (sin motivo registrado)"

// measure agrega una medida sobre todos los intentos de la cohorte.
func measure(as []Amount, accepted int) Measure {
	m := Measure{Expected: len(as)}
	var vals []float64
	var reasons []string
	for _, a := range as {
		if !a.Available {
			r := a.Reason
			if strings.TrimSpace(r) == "" {
				r = noReason
			}
			reasons = append(reasons, r)
			continue
		}
		m.Counted++
		if a.Partial {
			m.Partial++
		}
		vals = append(vals, a.Value)
	}
	m.Complete = m.Expected > 0 && m.Counted == m.Expected && m.Partial == 0
	switch {
	case m.Counted > 0:
		sort.Float64s(vals) // suma independiente del orden
		var sum float64
		for _, v := range vals {
			sum += v
		}
		m.Total = Amount{Available: true, Value: sum, Partial: !m.Complete}
	case m.Expected == 0:
		m.Total = Amount{Reason: "sin corridas iniciadas"}
	default:
		sort.Strings(reasons)
		m.Total = Amount{Reason: reasons[0]}
	}
	switch {
	case m.Expected == 0:
		m.PerAccepted = Amount{Reason: "sin corridas iniciadas"}
	case accepted == 0:
		m.PerAccepted = Amount{Reason: "sin tareas aceptadas"}
	case !m.Complete && m.Counted == 0:
		m.PerAccepted = Amount{Reason: "cobertura incompleta: " + m.Total.Reason}
	case !m.Complete:
		m.PerAccepted = Amount{Reason: fmt.Sprintf("cobertura incompleta: %d de %d intentos con la medida, %d parciales", m.Counted, m.Expected, m.Partial)}
	default:
		m.PerAccepted = Amount{Available: true, Value: m.Total.Value / float64(accepted)}
	}
	return m
}

func amountText(a Amount) string {
	if !a.Available {
		return "sin valor (" + a.Reason + ")"
	}
	s := fmt.Sprintf("%g", a.Value)
	if a.Partial {
		s += " (parcial)"
	}
	return s
}

func measureLine(name string, m Measure) string {
	return fmt.Sprintf("- %s: total %s; por tarea aceptada %s; cobertura %d de %d intentos, %d parciales",
		name, amountText(m.Total), amountText(m.PerAccepted), m.Counted, m.Expected, m.Partial)
}

// Markdown presenta el resumen: una sección por cohorte, las celdas
// incompatibles aparte y la descripción del cálculo.
func Markdown(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Resumen por cohorte\n\nVersión de protocolo: %s\n", r.ProtocolVersion)
	for _, c := range r.Cohorts {
		fmt.Fprintf(&b, "\n## %s · %s %s · %s · %s · variante %s\n\n", c.CohortID, c.Client, c.Version, c.Interface, c.Modality, c.Variant)
		fmt.Fprintf(&b, "- Iniciadas %d: aceptadas %d, rechazadas por calidad %d, bloqueadas por permisos %d, timeout %d, fallo del cliente %d, interrumpidas %d\n",
			c.Started, c.Accepted, c.Rejected, c.Blocked, c.TimedOut, c.ClientFailed, c.Interrupted)
		fmt.Fprintf(&b, "- Pendientes de revisor (ni aceptadas ni rechazadas): %d", c.Pending)
		if c.Provisional {
			b.WriteString(" · resumen provisional")
		}
		fmt.Fprintf(&b, "\n- Intentos: %d\n", c.Attempts)
		b.WriteString(measureLine("Tokens (consumo técnico)", c.Tokens) + "\n")
		b.WriteString(measureLine("Dinero (API)", c.Money) + "\n")
		b.WriteString(measureLine("Cuota (suscripción)", c.Quota) + "\n")
	}
	b.WriteString("\n## Celdas incompatibles\n\n")
	if len(r.Incompatible) == 0 {
		b.WriteString("Ninguna.\n")
	}
	for _, in := range r.Incompatible {
		fmt.Fprintf(&b, "- %s: %s\n", in.CohortID, strings.Join(in.Reasons, "; "))
	}
	b.WriteString("\n## Cálculo\n\n" +
		"El costo total suma todos los intentos de todas las corridas iniciadas de la cohorte, incluidos rechazos, bloqueos, fallos y reintentos. " +
		"El costo por tarea aceptada es ese total dividido por las aceptadas; sin aceptadas o con cobertura incompleta el cociente queda sin valor. " +
		"Las celdas incompatibles quedan fuera del denominador. API, suscripción e interfaces se publican por separado.\n")
	return b.String()
}
