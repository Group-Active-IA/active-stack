// Package delegation es la forma ejecutable de los veredictos del change opsx-selective-delegation
// (ficha 07): el sobre de retorno de un worker (design.md D6), la propiedad contra la partida (D4),
// la regla de decisión directo o delegado (D1) y la comprobación del resultado combinado.
//
// Igual que eval/evidence, es externo al binario: no lo importan cmd/ ni internal/ y usa sólo la
// biblioteca estándar y los paquetes de eval/. El contenido del campo `receipts` lo define el
// change 06; este paquete no valida, renombra ni agrega campos del recibo: lo juzga con
// evidence.EvaluateRequirement.
package delegation

import (
	"encoding/json"
	"fmt"

	"github.com/Group-Active-IA/active-stack/eval/evidence"
)

// Fingerprint es la huella {algorithm, value} de 02 y 06.
type Fingerprint = evidence.Fingerprint

// Start es la partida: commit base y huella del estado de trabajo (D4 paso 1), con la definición
// de la huella del estado probado de 06 (TreeHash).
type Start struct {
	Commit      string       `json:"commit"`
	Fingerprint *Fingerprint `json:"huella,omitempty"`
}

// Change es una ruta tocada con su tipo de cambio. En un renombre, From es el origen.
type Change struct {
	Path        string       `json:"ruta"`
	Kind        string       `json:"tipo"` // added | modified | deleted | renamed | reverted (volvió al commit base)
	From        string       `json:"desde,omitempty"`
	Fingerprint *Fingerprint `json:"huella,omitempty"` // huella del contenido, para separar lo de la partida
}

// Finding es un hallazgo acotado del worker.
type Finding struct {
	Severity    string `json:"severity"`
	Where       string `json:"path_line"`
	Description string `json:"description"`
}

// Envelope es el sobre de retorno de D6. Receipts es un recibo de ejecución por comando ejecutado,
// según 06; vacío si no se ejecutó ningún comando.
type Envelope struct {
	Status   string             `json:"status"` // completed | blocked | out_of_scope | failed
	Start    Start              `json:"start"`
	Modified []Change           `json:"modified"`
	Findings []Finding          `json:"findings"`
	Receipts []evidence.Receipt `json:"receipts"`
	Pending  string             `json:"pending"`
}

var envelopeFields = []string{"status", "start", "modified", "findings", "receipts", "pending"}

var validStatus = map[string]bool{"completed": true, "blocked": true, "out_of_scope": true, "failed": true}

// ParseEnvelope lee un retorno crudo. Un retorno que no es el sobre (prosa libre, JSON sin alguno
// de los seis campos o con un estado desconocido) devuelve error: queda sin verificar.
func ParseEnvelope(raw []byte) (Envelope, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Envelope{}, fmt.Errorf("el retorno no es el sobre: %w", err)
	}
	for _, f := range envelopeFields {
		if v, ok := fields[f]; !ok || string(v) == "null" {
			return Envelope{}, fmt.Errorf("falta el campo %q del sobre", f)
		}
	}
	var e Envelope
	if err := json.Unmarshal(raw, &e); err != nil {
		return Envelope{}, fmt.Errorf("sobre ilegible: %w", err)
	}
	if !validStatus[e.Status] {
		return Envelope{}, fmt.Errorf("estado desconocido %q", e.Status)
	}
	if e.Start.Commit == "" {
		return Envelope{}, fmt.Errorf("la partida no tiene commit")
	}
	return e, nil
}
