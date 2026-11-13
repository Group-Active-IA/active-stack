package evidence

import (
	"path"
	"strings"
)

// Status es el estado de un requisito (design.md D4).
type Status string

const (
	Verified   Status = "verificado"
	Rejected   Status = "rechazado"
	Unverified Status = "sin verificar"
)

// Corroboration es lo que dijo una fuente independiente o una reejecución sobre un recibo.
// Vacía = nadie corroboró.
//
// Una reejecución solo cuenta como fuente independiente si tiene captura propia de una fuente
// independiente (05 D20). La reejecución del coordinador sin esa captura no corrobora ni
// contradice: queda como si nadie hubiera corroborado.
type Corroboration string

const (
	Confirms          Corroboration = "confirma"                              // fuente independiente
	Contradicts       Corroboration = "contradice"                            // fuente independiente
	ReexecConfirms    Corroboration = "reejecucion confirma"                  // reejecución con captura independiente
	ReexecContradicts Corroboration = "reejecucion contradice"                // reejecución con captura independiente
	ReexecUncaptured  Corroboration = "reejecucion sin captura independiente" // no cuenta
)

// Sufficiency es el juicio del revisor sobre si el comando respalda el requisito (D3).
type Sufficiency string

const (
	Sufficient       Sufficiency = "suficiente"
	InsufficientCmd  Sufficiency = "comando insuficiente"
	UnmetRequirement Sufficiency = "requisito incumplido"
	NotEvaluable     Sufficiency = "no evaluable"
)

// Valores de los tres controles que se informan por separado (D3).
const (
	intPass      = "aprobada"
	intFail      = "fallida"
	intUnknown   = "no verificable"
	provOK       = "corroborada"
	provBad      = "contradicha"
	provNone     = "sin corroborar"
	suffOK       = "suficiente"
	suffBad      = "insuficiente"
	notEvaluable = "no evaluable"
	notApplied   = "no aplica"
)

// Env es lo observado fuera del recibo: el árbol actual, las copias guardadas y lo que dijeron la
// fuente independiente, la reejecución y el revisor. Se indexa por id de corrida.
type Env struct {
	Dir           string // worktree del change
	Commit        string // commit actual
	StateFP       *Fingerprint
	Copies        map[string][]byte // bytes de la copia redactada guardada; ausente = no existe
	Expired       map[string]bool   // la referencia existía y la copia venció
	Corroboration map[string]Corroboration
	Sufficiency   map[string]Sufficiency
	DetailNeeded  map[string]bool // verificar exige un detalle que la redacción descartó
}

// Verdict separa los tres controles y da el estado (D3, D4).
type Verdict struct {
	Status      Status
	Integrity   string
	Provenance  string
	Sufficiency string
	Reasons     []string
}

// EvaluateRequirement juzga un requisito con los recibos que lo declaran.
//
// Sin recibos nunca queda verificado. Con varios recibos, el peor estado manda
// (rechazado > sin verificar > verificado); esa regla de agregación es (TBD) en el diseño y aquí
// solo se ejercita con un recibo.
func EvaluateRequirement(req string, receipts []Receipt, env Env) Verdict {
	var mine []Receipt
	for _, r := range receipts {
		if r.covers(req) {
			mine = append(mine, r)
		}
	}
	if len(mine) == 0 {
		return Verdict{Status: Unverified, Integrity: notApplied, Provenance: notApplied,
			Sufficiency: notApplied, Reasons: []string{"sin recibos para " + req}}
	}
	worst := evaluateReceipt(mine[0], env)
	for _, r := range mine[1:] {
		if v := evaluateReceipt(r, env); rank(v.Status) > rank(worst.Status) {
			worst = v
		}
	}
	return worst
}

func rank(s Status) int {
	switch s {
	case Rejected:
		return 2
	case Unverified:
		return 1
	}
	return 0
}

func evaluateReceipt(r Receipt, env Env) Verdict {
	// Salida truncada que nunca tuvo referencia a una copia: caso 4 de la ficha, rechazado.
	if r.Output.Truncated && r.Output.Fingerprint == nil {
		return Verdict{Status: Rejected, Integrity: intFail, Provenance: notEvaluable, Sufficiency: notEvaluable,
			Reasons: []string{"salida truncada sin original"}}
	}
	// Recibo incompleto: queda sin verificar, no se juzga lo demás.
	if missing := r.Missing(); len(missing) > 0 {
		return Verdict{Status: Unverified, Integrity: notEvaluable, Provenance: notEvaluable, Sufficiency: notEvaluable,
			Reasons: []string{"recibo incompleto, faltan: " + strings.Join(missing, ", ")}}
	}

	v := Verdict{}
	var rejected, unverified []string

	// Integridad: la copia existe y su huella recomputada coincide. Mecánico: no prueba que la
	// ejecución haya ocurrido.
	switch copyBytes, has := env.Copies[r.Run]; {
	case env.Expired[r.Run]:
		v.Integrity = intUnknown
		unverified = append(unverified, "copia vencida")
	case !has:
		v.Integrity = intFail
		rejected = append(rejected, "la copia no existe")
	case Fingerprint256(copyBytes) != *r.Output.Fingerprint:
		v.Integrity = intFail
		rejected = append(rejected, "la huella no coincide con los bytes de la copia")
	default:
		v.Integrity = intPass
	}

	// Procedencia: la ejecución es real y corresponde al estado del código.
	switch {
	case env.Corroboration[r.Run] == Contradicts || env.Corroboration[r.Run] == ReexecContradicts:
		v.Provenance = provBad
		rejected = append(rejected, "la fuente independiente o la reejecución lo contradicen")
	case r.Result.ExitCode != nil && r.Output.Capture.ExitCode != nil && *r.Result.ExitCode != *r.Output.Capture.ExitCode:
		v.Provenance = provBad
		rejected = append(rejected, "el código de salida declarado difiere del registro de captura")
	case env.Corroboration[r.Run] == Confirms || env.Corroboration[r.Run] == ReexecConfirms:
		v.Provenance = provOK
	default:
		v.Provenance = provNone
		unverified = append(unverified, "sin fuente independiente ni reejecución")
	}
	if !sameDir(r.Tree.Dir, env.Dir) {
		rejected = append(rejected, "directorio ajeno")
	}
	if r.Tree.Commit != env.Commit || !sameFP(r.Tree.Fingerprint, env.StateFP) {
		rejected = append(rejected, "estado del árbol distinto del actual")
	}

	// Suficiencia: el comando ejercita lo que el requisito pide y el resultado lo respalda.
	switch {
	case env.DetailNeeded[r.Run]:
		v.Sufficiency = notEvaluable
		unverified = append(unverified, "verificar exige un detalle que la redacción descartó")
	default:
		switch env.Sufficiency[r.Run] {
		case Sufficient:
			v.Sufficiency = suffOK
		case InsufficientCmd:
			v.Sufficiency = suffBad
			rejected = append(rejected, "comando insuficiente")
		case UnmetRequirement:
			v.Sufficiency = suffBad
			rejected = append(rejected, "requisito incumplido con tests verdes")
		default:
			v.Sufficiency = notEvaluable
			unverified = append(unverified, "suficiencia no evaluada")
		}
	}

	switch {
	case len(rejected) > 0:
		v.Status, v.Reasons = Rejected, rejected
	case len(unverified) > 0:
		v.Status, v.Reasons = Unverified, unverified
	default:
		v.Status = Verified
	}
	return v
}

func sameFP(a, b *Fingerprint) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// sameDir compara directorios sin importar el estilo del separador ni la barra final. Se normaliza
// a mano (no con filepath) para que el veredicto sea el mismo en Windows y en Linux: en Linux
// filepath no trata la barra invertida como separador.
//
// Límite (TBD): la comparación distingue mayúsculas y minúsculas, y no resuelve enlaces
// simbólicos ni nombres cortos 8.3 de Windows.
func sameDir(a, b string) bool {
	return normDir(a) == normDir(b)
}

func normDir(p string) string {
	return path.Clean(strings.ReplaceAll(p, "\\", "/"))
}
