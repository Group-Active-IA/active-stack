package shadow

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// PromptVersion identifica el prompt fijo de revisión (tarea 4.3). Cambiar su texto exige una
// versión nueva: TestPrompt_FixedAndVersioned fija la huella.
const PromptVersion = "revision-v1"

//go:embed prompt/revision-v1.md
var promptV1 string

// Prompt devuelve el texto del prompt de revisión de PromptVersion.
func Prompt() string { return promptV1 }

// PromptFingerprint es la huella sha256 del prompt, para registrarla junto a cada corrida.
func PromptFingerprint() Fingerprint {
	sum := sha256.Sum256([]byte(promptV1))
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(sum[:])}
}

// Severities son las severidades de D4.
var Severities = []string{"bloqueante", "mayor", "menor"}

// Finding es un hallazgo en el formato de D4: severidad, ruta, rango de líneas, requisito o
// escenario y descripción.
type Finding struct {
	Severity    string `json:"severidad"`
	Path        string `json:"ruta"`
	From        int    `json:"linea_desde"`
	To          int    `json:"linea_hasta"`
	Requirement string `json:"requisito"`
	Description string `json:"descripcion"`
}

// ParseFindings lee la respuesta de un juez. Es estricta: una lista JSON, sin campos fuera de D4
// ni texto adicional. Un hallazgo que no se puede ubicar no entra al emparejamiento.
func ParseFindings(raw []byte) ([]Finding, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var out []Finding
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("hallazgos: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("hallazgos: texto después de la lista")
	}
	if out == nil {
		out = []Finding{}
	}
	for i, f := range out {
		if err := f.validate(); err != nil {
			return nil, fmt.Errorf("hallazgo %d: %w", i, err)
		}
	}
	return out, nil
}

func (f Finding) validate() error {
	known := false
	for _, s := range Severities {
		known = known || f.Severity == s
	}
	switch {
	case !known:
		return fmt.Errorf("severidad %q fuera de D4", f.Severity)
	case strings.TrimSpace(f.Path) == "":
		return errors.New("sin ruta")
	case f.From < 1 || f.To < f.From:
		return fmt.Errorf("rango de líneas inválido %d-%d", f.From, f.To)
	case strings.TrimSpace(f.Requirement) == "":
		return errors.New("sin requisito o escenario")
	case strings.TrimSpace(f.Description) == "":
		return errors.New("sin descripción")
	}
	return nil
}
