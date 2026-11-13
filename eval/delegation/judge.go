package delegation

import (
	"fmt"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/evidence"
)

// Estados del retorno. Los tres primeros son los de 06; out_of_scope es el rechazo por propiedad
// de D4, que se informa aparte porque el cambio no se integra aunque los tests pasen.
const (
	Verified   = evidence.Verified
	Unverified = evidence.Unverified
	Rejected   = evidence.Rejected
	OutOfScope = evidence.Status("out_of_scope")
)

// Pasos de la validación de D6, en orden.
const (
	StepEnvelope = "sobre"
	StepStart    = "partida"
	StepOwner    = "propiedad"
	StepReceipts = "recibos"
	StepCriteria = "criterios"
)

// Acciones del coordinador después del veredicto.
const (
	Integrate   = "integrar"
	AskEnvelope = "pedir sobre o reintentar"
	Retry       = "reintentar"
	Corroborate = "corroborar"
	Escalate    = "escalar"
)

// MaxAttempts es el tope de intentos por encargo (D7).
const MaxAttempts = 3

// Assignment es lo que el coordinador registró al delegar: partida, conjunto de escritura,
// requisitos que el encargo debe cubrir, intento actual y motivos de rechazo previos.
type Assignment struct {
	ID           string   `json:"id"`
	Start        Start    `json:"partida"`
	WriteSet     []string `json:"escritura"`
	Requirements []string `json:"requisitos"`
	Attempt      int      `json:"intento"`
	History      []string `json:"historial"`
}

// Observed es lo que el coordinador comprueba por su cuenta al recibir el retorno.
type Observed struct {
	// AtStart son los cambios que ya existían en la partida (no son del worker).
	AtStart []Change
	// Final son los cambios del árbol del worker contra el commit base: commits nuevos,
	// preparados, sin preparar, nuevos, borrados y renombrados.
	Final []Change
	// CriteriaMet: nil = no comprobados.
	CriteriaMet *bool
	// Evidence es lo observado para juzgar los recibos con el contrato de 06.
	Evidence evidence.Env
}

// ReturnVerdict es el veredicto sobre un retorno.
type ReturnVerdict struct {
	Status   evidence.Status
	Step     string // paso que cerró el veredicto
	Action   string
	Reasons  []string
	Escalate []string // historial con el que se escala, cuando corresponde
}

// JudgeReturn valida un retorno crudo contra su encargo, en el orden de D6: sobre, partida,
// propiedad, recibos (con el paso de revisión de 06) y criterios de salida. El primer paso que no
// pasa cierra el veredicto.
func JudgeReturn(a Assignment, raw []byte, obs Observed) ReturnVerdict {
	e, err := ParseEnvelope(raw)
	if err != nil {
		return a.close(Unverified, StepEnvelope, err.Error())
	}

	// 1. La partida declarada es la registrada (commit y huella del estado, 06 D2).
	if e.Start.Commit != a.Start.Commit || !sameFP(e.Start.Fingerprint, a.Start.Fingerprint) {
		return a.close(Unverified, StepStart, "partida declarada distinta de la registrada")
	}

	// 2. Propiedad (D4): lo atribuible es lo que cambió desde la partida; fuera del conjunto de
	// escritura es out_of_scope aunque esté confirmado y los tests pasen.
	attr := Attributable(obs.AtStart, obs.Final)
	if out := Outside(Touched(attr), a.WriteSet); len(out) > 0 {
		var reasons []string
		for _, p := range out {
			reasons = append(reasons, "ruta fuera del conjunto de escritura: "+p)
		}
		return a.close(OutOfScope, StepOwner, reasons...)
	}
	if diff := modifiedMismatch(attr, e.Modified); len(diff) > 0 {
		return a.close(Unverified, StepOwner, diff...)
	}

	// 3. Recibos: cada requisito del encargo con el juez de 06. Manda el peor estado.
	worst, reasons := Verified, []string(nil)
	for _, req := range a.Requirements {
		v := evidence.EvaluateRequirement(req, e.Receipts, obs.Evidence)
		if rank(v.Status) > rank(worst) {
			worst = v.Status
		}
		if v.Status != Verified {
			for _, r := range v.Reasons {
				reasons = append(reasons, req+": "+r)
			}
		}
	}
	if worst != Verified {
		return a.close(worst, StepReceipts, reasons...)
	}

	// 4. Criterios de salida comprobados por el coordinador y estado declarado por el worker.
	switch {
	case obs.CriteriaMet == nil:
		return a.close(Unverified, StepCriteria, "criterios de salida sin comprobar")
	case !*obs.CriteriaMet:
		return a.close(Rejected, StepCriteria, "criterios de salida no cumplidos")
	case e.Status != "completed":
		return a.close(Rejected, StepCriteria, "el worker informa "+e.Status)
	}
	return a.close(Verified, StepCriteria)
}

// close arma el veredicto y la acción siguiente. Con el tope de intentos agotado no hay otro
// intento: se escala con el historial (D7).
func (a Assignment) close(s evidence.Status, step string, reasons ...string) ReturnVerdict {
	v := ReturnVerdict{Status: s, Step: step, Reasons: reasons}
	switch {
	case s == Verified:
		v.Action = Integrate
	case s == Unverified && step == StepReceipts:
		// La ejecución puede ser real: falta corroborarla con captura independiente.
		v.Action = Corroborate
	case a.Attempt >= MaxAttempts:
		v.Action = Escalate
		v.Escalate = append(append([]string(nil), a.History...),
			fmt.Sprintf("intento %d: %s", a.Attempt, strings.Join(reasons, "; ")))
	case step == StepEnvelope:
		v.Action = AskEnvelope
	default:
		v.Action = Retry
	}
	return v
}

// Attributable devuelve los cambios finales que no estaban, idénticos, en la partida.
func Attributable(atStart, final []Change) []Change {
	pre := map[string]bool{}
	for _, c := range atStart {
		pre[key(c, true)] = true
	}
	var out []Change
	for _, c := range final {
		if !pre[key(c, true)] {
			out = append(out, c)
		}
	}
	return out
}

// Touched devuelve las rutas que tocan los cambios: en un renombre cuentan origen y destino.
func Touched(cs []Change) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Path)
		if c.Kind == "renamed" && c.From != "" {
			out = append(out, c.From)
		}
	}
	return out
}

// Outside devuelve las rutas que no cubre el conjunto de escritura. Sin conjunto, ninguna.
func Outside(paths, writeSet []string) []string {
	var out []string
	for _, p := range paths {
		if !covered(p, writeSet) {
			out = append(out, p)
		}
	}
	return out
}

func covered(p string, writeSet []string) bool {
	for _, w := range writeSet {
		if overlap(w, p) {
			return true
		}
	}
	return false
}

// modifiedMismatch compara `modified` con lo atribuible (ruta, tipo y origen). Una omisión o una
// ruta de más deja el retorno sin verificar: no se acepta por inferencia.
func modifiedMismatch(attr, declared []Change) []string {
	want, got := map[string]bool{}, map[string]bool{}
	for _, c := range attr {
		want[key(c, false)] = true
	}
	for _, c := range declared {
		got[key(c, false)] = true
	}
	var diff []string
	for _, c := range attr {
		if !got[key(c, false)] {
			diff = append(diff, "modified omite "+c.Path)
		}
	}
	for _, c := range declared {
		if !want[key(c, false)] {
			diff = append(diff, "modified declara de más "+c.Path)
		}
	}
	return diff
}

// key identifica un cambio por ruta, tipo y origen; con withFP, también por la huella del
// contenido (para separar un archivo que ya estaba modificado en la partida).
func key(c Change, withFP bool) string {
	k := c.Path + "|" + c.Kind + "|" + c.From
	if withFP && c.Fingerprint != nil {
		k += "|" + c.Fingerprint.Algorithm + ":" + c.Fingerprint.Value
	}
	return k
}

func rank(s evidence.Status) int {
	switch s {
	case Rejected:
		return 2
	case Unverified:
		return 1
	}
	return 0
}

func sameFP(a, b *Fingerprint) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
