package shadow

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// FileHash es la huella de un archivo enviado. Con ellas se reconstruye la huella del paquete.
type FileHash struct {
	Path   string `json:"ruta"`
	SHA256 string `json:"sha256"`
}

// SendRecord es una entrada del registro de envío (tarea 3.3): caso, huella, candidato,
// proveedor y fecha, más la huella de cada archivo.
type SendRecord struct {
	Case      string      `json:"caso"`
	Candidate string      `json:"candidato"`
	Provider  string      `json:"proveedor"`
	Package   Fingerprint `json:"huella"`
	Date      time.Time   `json:"fecha"`
	Files     []FileHash  `json:"archivos"`
}

// SendLog es el registro de envío, append-only en JSONL. No envía nada: autoriza y registra.
type SendLog struct{ path string }

func NewSendLog(path string) *SendLog { return &SendLog{path: path} }

// Authorize registra un envío sólo si el examen previo no tuvo positivos, el proveedor documenta
// retención y entrenamiento por separado (3.4) y la huella coincide con los archivos.
func (l *SendLog) Authorize(p Package, scan ScanResult, candidate string, terms ProviderTerms, now time.Time) (SendRecord, error) {
	if strings.TrimSpace(p.Case) == "" || strings.TrimSpace(candidate) == "" {
		return SendRecord{}, errors.New("envío sin caso o sin candidato")
	}
	if !scan.Send || len(scan.Hits) > 0 {
		return SendRecord{}, fmt.Errorf("envío de %s cancelado por el examen previo", p.Case)
	}
	if m := terms.missing(); len(m) > 0 {
		return SendRecord{}, fmt.Errorf("envío de %s a %q bloqueado: %s", p.Case, terms.Provider, strings.Join(m, "; "))
	}
	files := fileHashes(p.Files)
	if ReconstructFingerprint(files) != p.Fingerprint {
		return SendRecord{}, fmt.Errorf("envío de %s: la huella del paquete no coincide con sus archivos", p.Case)
	}
	rec := SendRecord{Case: p.Case, Candidate: candidate, Provider: terms.Provider, Package: p.Fingerprint, Date: now.UTC(), Files: files}
	line, err := json.Marshal(rec)
	if err != nil {
		return SendRecord{}, err
	}
	if err := appendLine(l.path, line); err != nil {
		return SendRecord{}, err
	}
	return rec, nil
}

// Records lee el registro completo.
func (l *SendLog) Records() ([]SendRecord, error) {
	var out []SendRecord
	err := readLines(l.path, func(b []byte) error {
		var r SendRecord
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}
		out = append(out, r)
		return nil
	})
	return out, err
}

func fileHashes(files []File) []FileHash {
	out := make([]FileHash, 0, len(files))
	for _, f := range files {
		sum := sha256.Sum256(f.Data)
		out = append(out, FileHash{Path: f.Path, SHA256: hex.EncodeToString(sum[:])})
	}
	return out
}

// ReconstructFingerprint aplica la convención de TreeHash de 02 sobre huellas ya calculadas:
// sha256 sobre las líneas "<sha256 del archivo>  <ruta>" ordenadas por ruta.
func ReconstructFingerprint(files []FileHash) Fingerprint {
	sorted := append([]FileHash(nil), files...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	h := sha256.New()
	for _, f := range sorted {
		fmt.Fprintf(h, "%s  %s\n", f.SHA256, f.Path)
	}
	return Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(h.Sum(nil))}
}

func appendLine(path string, line []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func readLines(path string, fn func([]byte) error) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		if err := fn(sc.Bytes()); err != nil {
			return err
		}
	}
	return sc.Err()
}
