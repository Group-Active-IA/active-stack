package delegation

import "github.com/Group-Active-IA/active-stack/eval/evidence"

// Integration es la comprobación del resultado combinado de varios encargos (D4 paso 4): los
// cambios del árbol combinado contra la base, los escenarios que dependen de la combinación y la
// evidencia de su reejecución.
type Integration struct {
	Final     []Change
	Scenarios []string
	Receipts  []evidence.Receipt
	Evidence  evidence.Env
}

// CombinedVerdict es el veredicto sobre el resultado combinado.
type CombinedVerdict struct {
	Status    evidence.Status
	Integrate bool
	FailedFor []string // encargos contra los que se registra el fallo
	Reasons   []string
}

// JudgeCombination juzga la integración de encargos cuyos retornos ya fueron aceptados: vuelve a
// comprobar el alcance sobre el árbol combinado y juzga la reejecución de cada escenario combinado
// con el contrato de 06. Sólo corrobora una reejecución con captura propia de una fuente
// independiente. Si un escenario corroborado falla, no se integra y el fallo se registra contra
// todos los encargos combinados.
func JudgeCombination(assignments []Assignment, in Integration) CombinedVerdict {
	var writeSet, ids []string
	for _, a := range assignments {
		writeSet = append(writeSet, a.WriteSet...)
		ids = append(ids, a.ID)
	}
	if out := Outside(Touched(in.Final), writeSet); len(out) > 0 {
		var reasons []string
		for _, p := range out {
			reasons = append(reasons, "ruta del resultado combinado fuera de todo conjunto de escritura: "+p)
		}
		return CombinedVerdict{Status: OutOfScope, Reasons: reasons}
	}

	worst, failed, reasons := Verified, false, []string(nil)
	for _, sc := range in.Scenarios {
		v := evidence.EvaluateRequirement(sc, in.Receipts, in.Evidence)
		if v.Status == Verified && fails(sc, in.Receipts) {
			failed = true
			reasons = append(reasons, "el escenario "+sc+" falla en el resultado combinado")
			continue
		}
		if rank(v.Status) > rank(worst) {
			worst = v.Status
		}
		for _, r := range v.Reasons {
			reasons = append(reasons, sc+": "+r)
		}
	}
	switch {
	case failed:
		return CombinedVerdict{Status: Rejected, FailedFor: ids, Reasons: reasons}
	case worst != Verified:
		return CombinedVerdict{Status: worst, Reasons: reasons}
	}
	return CombinedVerdict{Status: Verified, Integrate: true}
}

// fails: algún recibo del escenario terminó con código distinto de cero. Sólo se consulta cuando
// el recibo ya quedó verificado, es decir, cuando el resultado es real.
func fails(scenario string, receipts []evidence.Receipt) bool {
	for _, r := range receipts {
		for _, req := range r.Requirements {
			if req == scenario && r.Result.ExitCode != nil && *r.Result.ExitCode != 0 {
				return true
			}
		}
	}
	return false
}
