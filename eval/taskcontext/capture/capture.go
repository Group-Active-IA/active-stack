// Package capture es el capturador independiente del arnés (change
// opsx-task-context, D8, D9 y D20): ejecuta el comando de un fixture, recibe
// stdout y stderr en un mismo flujo ordenado, registra el código de salida,
// redacta en memoria con la lista del catálogo y escribe solo la copia
// redactada, su huella y el registro de captura. Vive fuera del binario
// (CHANGES.md:49).
package capture

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
)

// Archivos que deja una captura en su ruta de corrida.
const (
	CopyFile   = "output.redacted.txt"
	RecordFile = "capture.json"
)

// Fuentes de captura y de redacción (D9).
const (
	SourceHarness = "arnés"
	SourceAgent   = "agente"
	// clientSourcePrefix abre `cliente:<fuente>`.
	clientSourcePrefix = "cliente:"
)

// Cobertura antes del truncamiento (D9, D22).
const (
	CoverageDemonstrated    = "demostrada"
	CoverageNotDemonstrated = "no demostrada"
)

// Status es el estado máximo al que puede llegar un resultado según su captura.
type Status string

const (
	StatusUnverified Status = "sin verificar"
	// StatusVerifiable: la procedencia es independiente; que llegue a
	// «verificado» lo decide la suficiencia (06).
	StatusVerifiable Status = "verificable"
)

// Meta identifica dónde corrió la captura.
type Meta struct {
	Client, Interface, System, Version string
}

// Record es el registro de captura (D9), junto a la copia.
type Record struct {
	Command    string         `json:"command"`
	RunID      string         `json:"run_id"`
	Source     string         `json:"source"`
	Redactor   string         `json:"redactor"`
	Client     string         `json:"client"`
	Interface  string         `json:"interface"`
	System     string         `json:"system"`
	Version    string         `json:"version"`
	ExitCode   int            `json:"exit_code"`
	Coverage   string         `json:"coverage"`
	Redactions map[string]int `json:"redactions"`
	SHA256     string         `json:"sha256"`
	CopyPath   string         `json:"copy_path,omitempty"`
}

// Validate comprueba que el registro declare lo que D9 exige.
func (r Record) Validate() error {
	if !validSource(r.Source) {
		return fmt.Errorf("capture: fuente de captura inválida %q (want %s, %s<fuente> o %s)", r.Source, SourceHarness, clientSourcePrefix, SourceAgent)
	}
	if r.Redactor == "" {
		return errors.New("capture: el registro debe declarar el redactor")
	}
	if r.Coverage != CoverageDemonstrated && r.Coverage != CoverageNotDemonstrated {
		return fmt.Errorf("capture: cobertura inválida %q", r.Coverage)
	}
	if r.Client == "" || r.Interface == "" || r.System == "" || r.Version == "" {
		return errors.New("capture: el registro debe declarar cliente, interfaz, sistema y versión")
	}
	return nil
}

func validSource(s string) bool {
	switch {
	case s == SourceHarness, s == SourceAgent:
		return true
	case strings.HasPrefix(s, clientSourcePrefix):
		return len(s) > len(clientSourcePrefix)
	}
	return false
}

// MaxStatus es el estado más alto que el resultado puede alcanzar (D20): una
// captura del agente, o con cobertura no demostrada, nunca supera «sin
// verificar» aunque la huella coincida.
func (r Record) MaxStatus() Status {
	if r.Source == SourceAgent || r.Coverage != CoverageDemonstrated {
		return StatusUnverified
	}
	return StatusVerifiable
}

// Spec describe una captura.
type Spec struct {
	// Command es el comando que figura en la referencia.
	Command string
	// Cmd es lo que se ejecuta de verdad (el proceso del fixture).
	Cmd      *exec.Cmd
	RunDir   string
	RunID    string
	Redactor *redact.Redactor
	Meta     Meta
}

// Run ejecuta el comando, redacta en memoria y escribe la copia y el registro.
// Nunca escribe el original sin redactar.
func Run(spec Spec) (Record, error) {
	if spec.Cmd == nil || spec.Redactor == nil {
		return Record{}, errors.New("capture: Spec necesita Cmd y Redactor")
	}
	// Un único flujo para stdout y stderr: el orden de emisión se conserva.
	var buf bytes.Buffer
	spec.Cmd.Stdout = &buf
	spec.Cmd.Stderr = &buf
	exit := 0
	if err := spec.Cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			return Record{}, fmt.Errorf("capture: no se pudo ejecutar el comando: %w", err)
		}
		exit = ee.ExitCode()
	}
	redacted, byID := spec.Redactor.RedactText(buf.String())
	buf.Reset() // el original no sale de memoria
	sum := sha256.Sum256([]byte(redacted))
	// El comando de la referencia también es texto libre que se persiste: pasa
	// por el mismo redactor que la salida (la identidad de la corrida son el
	// id, la huella y el código, no la credencial).
	command, _ := spec.Redactor.RedactText(spec.Command)
	rec := Record{
		Command: command, RunID: spec.RunID,
		Source: SourceHarness, Redactor: SourceHarness,
		Client: spec.Meta.Client, Interface: spec.Meta.Interface, System: spec.Meta.System, Version: spec.Meta.Version,
		ExitCode: exit, Coverage: CoverageDemonstrated,
		Redactions: spec.Redactor.ByCategory(byID),
		SHA256:     hex.EncodeToString(sum[:]),
		CopyPath:   filepath.Join(spec.RunDir, CopyFile),
	}
	if err := writeRun(spec.RunDir, rec, []byte(redacted)); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func writeRun(dir string, rec Record, copy []byte) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := requireOwnDir(dir); err != nil {
		return err
	}
	if err := writeFileAtomic(filepath.Join(dir, CopyFile), copy); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(dir, RecordFile), append(data, '\n'))
}

// requireOwnDir exige que dir sea un directorio real, no un symlink.
func requireOwnDir(dir string) error {
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("capture: %s no es un directorio propio (symlink o archivo)", dir)
	}
	return nil
}

// writeFileAtomic publica path sin adoptar destinos ajenos: si ya existe y no
// es un archivo regular (symlink, directorio) falla; escribe en un temporal
// 0600 del mismo directorio y lo renombra, de modo que el rename reemplaza el
// nombre final en lugar de seguir cualquier enlace.
func writeFileAtomic(path string, data []byte) error {
	if fi, err := os.Lstat(path); err == nil {
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("capture: %s existe y no es un archivo regular: no se sobrescribe", path)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, werr := tmp.Write(data)
	if werr == nil {
		werr = tmp.Sync()
	}
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Chmod(name, 0o600)
	}
	if werr == nil {
		werr = os.Rename(name, path)
	}
	if werr != nil {
		_ = os.Remove(name)
	}
	return werr
}
