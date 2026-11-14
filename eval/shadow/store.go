package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	findingsFile    = "hallazgos.json"
	fingerprintFile = "huella.json"
	accessLogFile   = "accesos.jsonl"
)

var caseIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Closure registra el cierre definitivo de un caso (D3 paso 4): decisión de referencia y
// etiquetas congeladas. Con alguna vacía el caso sigue abierto.
type Closure struct {
	ReferenceDecision string
	FrozenLabels      string
}

func (c Closure) closed() bool {
	return strings.TrimSpace(c.ReferenceDecision) != "" && strings.TrimSpace(c.FrozenLabels) != ""
}

// Access es un intento de lectura del almacén, concedido o no.
type Access struct {
	Time    time.Time `json:"fecha"`
	Who     string    `json:"quien"`
	Case    string    `json:"caso"`
	Granted bool      `json:"concedido"`
	Reason  string    `json:"motivo"`
}

// Store es el almacén de hallazgos del juez (tarea 4.2): fuera del árbol del repo, con huella por
// caso, escritura única y lectura sólo del custodio con el caso cerrado. Cada intento queda en el
// registro de accesos. La huella prueba integridad, no confidencialidad: la confidencialidad
// depende de que el directorio sea del custodio (se crea 0700, archivos 0600).
type Store struct {
	dir       string
	custodian string
	now       func() time.Time
}

// NewStore abre el almacén en dir. Rechaza un dir dentro de repoRoot y un custodio vacío.
func NewStore(dir, repoRoot, custodian string) (*Store, error) {
	if strings.TrimSpace(custodian) == "" {
		return nil, errors.New("almacén sin custodio designado")
	}
	inside, err := within(dir, repoRoot)
	if err != nil {
		return nil, err
	}
	if inside {
		return nil, fmt.Errorf("el almacén %s está dentro del árbol del repo", dir)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir, custodian: custodian, now: time.Now}, nil
}

func within(dir, root string) (bool, error) {
	d, err := realAbs(dir)
	if err != nil {
		return false, err
	}
	r, err := realAbs(root)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(r, d)
	if err != nil {
		return false, nil // otra unidad: fuera del árbol
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))), nil
}

// realAbs resuelve los enlaces del tramo existente de la ruta, para que un enlace no esconda que
// el almacén cae dentro del repo.
func realAbs(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	rest := ""
	for cur := abs; ; cur = filepath.Dir(cur) {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(real, rest), nil
		}
		if filepath.Dir(cur) == cur {
			return abs, nil
		}
		rest = filepath.Join(filepath.Base(cur), rest)
	}
}

func validCase(id string) error {
	if !caseIDPattern.MatchString(id) || strings.Contains(id, "..") {
		return fmt.Errorf("id de caso inválido %q", id)
	}
	return nil
}

// Save guarda los hallazgos de un caso una sola vez y devuelve su huella.
func (s *Store) Save(caseID string, findings []byte) (Fingerprint, error) {
	if err := validCase(caseID); err != nil {
		return Fingerprint{}, err
	}
	dir := filepath.Join(s.dir, caseID)
	if err := os.Mkdir(dir, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return Fingerprint{}, fmt.Errorf("los hallazgos de %s ya están guardados", caseID)
		}
		return Fingerprint{}, err
	}
	sum := sha256.Sum256(findings)
	fp := Fingerprint{Algorithm: "sha256", Value: hex.EncodeToString(sum[:])}
	if err := os.WriteFile(filepath.Join(dir, findingsFile), findings, 0o600); err != nil {
		return Fingerprint{}, err
	}
	raw, err := json.Marshal(fp)
	if err != nil {
		return Fingerprint{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, fingerprintFile), raw, 0o600); err != nil {
		return Fingerprint{}, err
	}
	return fp, nil
}

// Open entrega los hallazgos sólo al custodio y sólo con el caso cerrado, después de verificar la
// huella. Todo intento, concedido o no, queda registrado.
func (s *Store) Open(who, caseID string, c Closure) ([]byte, error) {
	deny := func(reason string) ([]byte, error) {
		if err := s.log(Access{Who: who, Case: caseID, Reason: reason}); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("acceso a %s denegado a %q: %s", caseID, who, reason)
	}
	if err := validCase(caseID); err != nil {
		return deny(err.Error())
	}
	if who != s.custodian {
		return deny("sólo el custodio abre resultados")
	}
	if !c.closed() {
		return deny("caso abierto: falta decisión de referencia o etiquetas congeladas")
	}
	dir := filepath.Join(s.dir, caseID)
	data, err := os.ReadFile(filepath.Join(dir, findingsFile))
	if err != nil {
		return deny("hallazgos no disponibles")
	}
	raw, err := os.ReadFile(filepath.Join(dir, fingerprintFile))
	if err != nil {
		return deny("huella no disponible")
	}
	var want Fingerprint
	if err := json.Unmarshal(raw, &want); err != nil {
		return deny("huella ilegible")
	}
	sum := sha256.Sum256(data)
	if want.Algorithm != "sha256" || want.Value != hex.EncodeToString(sum[:]) {
		return deny("la huella no coincide con la registrada al guardar")
	}
	if err := s.log(Access{Who: who, Case: caseID, Granted: true, Reason: "caso cerrado; huella verificada"}); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) log(a Access) error {
	a.Time = s.now().UTC()
	line, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return appendLine(filepath.Join(s.dir, accessLogFile), line)
}

// Accesses lee el registro de accesos.
func (s *Store) Accesses() ([]Access, error) {
	var out []Access
	err := readLines(filepath.Join(s.dir, accessLogFile), func(b []byte) error {
		var a Access
		if err := json.Unmarshal(b, &a); err != nil {
			return err
		}
		out = append(out, a)
		return nil
	})
	return out, err
}
