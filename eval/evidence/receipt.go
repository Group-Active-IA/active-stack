// Package evidence es la forma ejecutable del contrato de recibos de ejecución del change
// opsx-verifiable-evidence (ficha 06): el esquema del recibo (design.md D2), los tres controles
// (D3) y los tres estados (D4).
//
// Igual que eval/baseline, es externo al binario: no lo importan cmd/ ni internal/, usa solo la
// biblioteca estándar y los paquetes de eval/. No es el verificador del producto; sirve para que
// los fixtures de data fijen, con casos, qué veredicto exige el contrato.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/eval/baseline/record"
)

// Fingerprint es una huella con el algoritmo nombrado ({algoritmo, valor}), la misma de 02.
type Fingerprint = record.Fingerprint

// Fingerprint256 es el sha256 de los bytes exactos, sin conversión de fines de línea.
func Fingerprint256(data []byte) Fingerprint {
	h := sha256.Sum256(data)
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(h[:])}
}

// Tree es el árbol donde corrió el comando y el estado exacto probado (D2 `arbol`). Fingerprint es
// la huella del estado completo cuando había cambios sin confirmar; vacía con el árbol limpio.
type Tree struct {
	Dir         string       `json:"directorio"`
	Commit      string       `json:"commit"`
	Fingerprint *Fingerprint `json:"huella,omitempty"`
}

// Result es el resultado declarado: código de salida y marca de tiempo (D2 `resultado`).
type Result struct {
	ExitCode *int   `json:"codigo_de_salida"`
	At       string `json:"marca_de_tiempo"`
}

// Capture es el registro de captura de 05 (D9). 06 lo referencia y no define sus campos; el
// código de salida lo escribe quien captura, no el worker.
type Capture struct {
	Source     string         `json:"fuente"`
	Redactor   string         `json:"redactor"`
	Client     string         `json:"cliente"`
	Interface  string         `json:"interfaz"`
	System     string         `json:"sistema"`
	Version    string         `json:"version"`
	ExitCode   *int           `json:"codigo_de_salida"`
	Coverage   string         `json:"cobertura"`
	Redactions map[string]int `json:"redacciones"`
}

// Output es la referencia a la copia redactada: una sola huella, si fue truncada y el registro
// de captura (D2 `salida`).
type Output struct {
	Fingerprint *Fingerprint `json:"huella,omitempty"`
	Truncated   bool         `json:"truncada"`
	Capture     *Capture     `json:"captura,omitempty"`
}

// Receipt es un recibo de ejecución: una ejecución de un comando (D2).
type Receipt struct {
	Run          string   `json:"corrida"`
	Tree         Tree     `json:"arbol"`
	Command      string   `json:"comando"`
	Result       Result   `json:"resultado"`
	Requirements []string `json:"requisitos"`
	Output       Output   `json:"salida"`
}

// Missing lista los campos obligatorios de D2 que faltan. Un recibo con faltantes es incompleto y
// queda sin verificar.
func (r Receipt) Missing() []string {
	var m []string
	add := func(cond bool, name string) {
		if cond {
			m = append(m, name)
		}
	}
	add(r.Run == "", "corrida")
	add(r.Tree.Dir == "" || r.Tree.Commit == "", "arbol")
	add(r.Command == "", "comando")
	add(r.Result.ExitCode == nil || r.Result.At == "", "resultado")
	add(len(r.Requirements) == 0, "requisitos")
	add(r.Output.Fingerprint == nil, "salida.huella")
	add(r.Output.Capture == nil || r.Output.Capture.Source == "" || r.Output.Capture.ExitCode == nil, "salida.captura")
	return m
}

func (r Receipt) covers(req string) bool {
	for _, x := range r.Requirements {
		if x == req {
			return true
		}
	}
	return false
}

// Case es un fixture: recibos, lo observado y el veredicto esperado.
type Case struct {
	ID          string
	Description string
	Requirement string
	Receipts    []Receipt
	Env         Env
	Want        Verdict
}

type caseFile struct {
	ID          string    `json:"id"`
	Description string    `json:"descripcion"`
	Requirement string    `json:"requisito"`
	Receipts    []Receipt `json:"recibos"`
	Env         envFile   `json:"entorno"`
	Want        struct {
		Status      Status `json:"estado"`
		Integrity   string `json:"integridad"`
		Provenance  string `json:"procedencia"`
		Sufficiency string `json:"suficiencia"`
	} `json:"esperado"`
}

type envFile struct {
	Dir           string                   `json:"directorio"`
	Commit        string                   `json:"commit"`
	StateFP       *Fingerprint             `json:"huella_estado,omitempty"`
	Copies        map[string]string        `json:"copias"`
	Expired       []string                 `json:"vencidas"`
	Corroboration map[string]Corroboration `json:"corroboracion"`
	Sufficiency   map[string]Sufficiency   `json:"suficiencia"`
	DetailNeeded  []string                 `json:"detalle_descartado_necesario"`
}

// LoadCase lee un caso; las copias redactadas se leen como bytes, relativas al directorio del
// caso padre (testdata/).
func LoadCase(path string) (Case, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Case{}, err
	}
	var f caseFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return Case{}, fmt.Errorf("%s: %w", path, err)
	}
	base := filepath.Dir(filepath.Dir(path))
	env := Env{
		Dir:           f.Env.Dir,
		Commit:        f.Env.Commit,
		StateFP:       f.Env.StateFP,
		Copies:        map[string][]byte{},
		Expired:       set(f.Env.Expired),
		Corroboration: f.Env.Corroboration,
		Sufficiency:   f.Env.Sufficiency,
		DetailNeeded:  set(f.Env.DetailNeeded),
	}
	for run, rel := range f.Env.Copies {
		b, err := os.ReadFile(filepath.Join(base, rel))
		if err != nil {
			return Case{}, fmt.Errorf("%s: copia de %s: %w", path, run, err)
		}
		env.Copies[run] = b
	}
	return Case{
		ID: f.ID, Description: f.Description, Requirement: f.Requirement,
		Receipts: f.Receipts, Env: env,
		Want: Verdict{Status: f.Want.Status, Integrity: f.Want.Integrity,
			Provenance: f.Want.Provenance, Sufficiency: f.Want.Sufficiency},
	}, nil
}

func set(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}
