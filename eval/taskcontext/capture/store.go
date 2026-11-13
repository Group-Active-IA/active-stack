package capture

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Retención (design.md D9): hasta que el change se archive y siete días más; en
// `_sin-change`, siete días desde la captura. Los siete días son un supuesto que
// se recalibra con el primer uso real.
const (
	Retention = 7 * 24 * time.Hour
	// NoChange es el nombre de carpeta cuando la acción no tiene change activo.
	NoChange = "_sin-change"
)

// AvailState es el estado de un original pedido por referencia.
type AvailState int

const (
	// NeverCaptured: no existe registro de esa corrida; no tiene ruta ni huella.
	NeverCaptured AvailState = iota
	// Available: dentro del plazo, recuperable.
	Available
	// Expired: tuvo referencia completa y ya se borró o venció.
	Expired
)

// Availability es la respuesta de Lookup.
type Availability struct {
	State     AvailState
	Path      string
	ExpiredOn time.Time
}

// Message es el texto que se informa al pedir el original.
func (a Availability) Message() string {
	switch a.State {
	case Available:
		return "disponible en " + a.Path
	case Expired:
		return "no disponible: vencido desde " + a.ExpiredOn.Format("2006-01-02")
	default:
		return "nunca capturado"
	}
}

// Store es la ubicación y la retención de las copias redactadas.
type Store struct {
	Root string
	Now  func() time.Time
}

type indexEntry struct {
	CapturedAt time.Time  `json:"captured_at"`
	PurgedOn   *time.Time `json:"purged_on,omitempty"`
}

type index struct {
	Runs     map[string]indexEntry `json:"runs"`
	Archived map[string]time.Time  `json:"archived"`
}

// DefaultCacheDir es <os.UserCacheDir>.
func DefaultCacheDir() (string, error) { return os.UserCacheDir() }

// NewStore arma un store bajo <cacheDir>/active-stack/runs.
func NewStore(cacheDir string, now func() time.Time) *Store {
	return &Store{Root: filepath.Join(cacheDir, "active-stack", "runs"), Now: now}
}

func key(change, runID string) string { return folder(change) + "/" + runID }

func folder(change string) string {
	if change == "" {
		return NoChange
	}
	return change
}

// RunDir es <raíz>/<change>/<id de corrida>/. No valida: es la ruta nominal;
// Save, Purge, BeginRun y Lookup validan antes de tocar disco.
func (s *Store) RunDir(change, runID string) string {
	return filepath.Join(s.Root, folder(change), runID)
}

// idRe es el alcance de un componente de ruta (change o id de corrida): sin
// separadores, sin `.`/`..`, sin volumen, sin vacío.
var idRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func validComponent(c string) bool { return idRe.MatchString(c) }

// checkIdentity valida change (vacío = _sin-change) y runID.
func checkIdentity(change, runID string) error {
	if change != "" && !validComponent(change) {
		return fmt.Errorf("capture: change inválido %q", change)
	}
	if !validComponent(runID) {
		return fmt.Errorf("capture: id de corrida inválido %q", runID)
	}
	return nil
}

// safeRunDir valida la identidad, confina el resultado dentro de Root y rechaza
// que el directorio del change o el de la corrida sean symlinks.
func (s *Store) safeRunDir(change, runID string) (string, error) {
	if err := checkIdentity(change, runID); err != nil {
		return "", err
	}
	dir := s.RunDir(change, runID)
	rel, err := filepath.Rel(s.Root, dir)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("capture: la corrida %q/%q escapa de la raíz", change, runID)
	}
	for _, p := range []string{filepath.Join(s.Root, folder(change)), dir} {
		fi, err := os.Lstat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !fi.IsDir() {
			return "", fmt.Errorf("capture: %s no es un directorio propio (symlink o archivo)", p)
		}
	}
	return dir, nil
}

func (s *Store) indexPath() string { return filepath.Join(s.Root, "index.json") }

func (s *Store) load() index {
	idx := index{Runs: map[string]indexEntry{}, Archived: map[string]time.Time{}}
	if data, err := os.ReadFile(s.indexPath()); err == nil {
		_ = json.Unmarshal(data, &idx)
		if idx.Runs == nil {
			idx.Runs = map[string]indexEntry{}
		}
		if idx.Archived == nil {
			idx.Archived = map[string]time.Time{}
		}
	}
	return idx
}

func (s *Store) save(idx index) error {
	if err := os.MkdirAll(s.Root, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.indexPath(), data)
}

// expiry devuelve cuándo vence una corrida, o false si todavía no tiene plazo
// (change activo, sin archivar).
func (idx index) expiry(change string, e indexEntry) (time.Time, bool) {
	if change == "" {
		return e.CapturedAt.Add(Retention), true
	}
	if at, ok := idx.Archived[change]; ok {
		return at.Add(Retention), true
	}
	return time.Time{}, false
}

// Save guarda la copia redactada y su registro bajo la ruta de corrida.
func (s *Store) Save(change, runID string, rec Record, copy []byte) (string, error) {
	dir, err := s.safeRunDir(change, runID)
	if err != nil {
		return "", err
	}
	if err := writeRun(dir, rec, copy); err != nil {
		return "", err
	}
	idx := s.load()
	idx.Runs[key(change, runID)] = indexEntry{CapturedAt: s.Now()}
	if err := s.save(idx); err != nil {
		return "", err
	}
	return filepath.Join(dir, CopyFile), nil
}

// MarkArchived registra que el change se archivó ahora: desde ahí corren los siete días.
func (s *Store) MarkArchived(change string) error {
	idx := s.load()
	idx.Archived[change] = s.Now()
	return s.save(idx)
}

// Purge borra las corridas vencidas y deja constancia de desde cuándo.
func (s *Store) Purge() ([]string, error) {
	idx := s.load()
	var removed []string
	var bad []error
	now := s.Now()
	for k, e := range idx.Runs {
		if e.PurgedOn != nil {
			continue
		}
		change, runID := splitKey(k)
		// El índice es entrada no confiable: se valida antes de cualquier RemoveAll.
		dir, err := s.safeKeyDir(change, runID)
		if err != nil {
			bad = append(bad, fmt.Errorf("capture: clave de índice %q rechazada: %w", k, err))
			continue
		}
		exp, ok := idx.expiry(unfolder(change), e)
		if !ok || now.Before(exp) {
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return removed, err
		}
		purged := exp
		e.PurgedOn = &purged
		idx.Runs[k] = e
		removed = append(removed, k)
	}
	if len(removed) == 0 {
		return nil, errors.Join(bad...)
	}
	return removed, errors.Join(append(bad, s.save(idx))...)
}

// safeKeyDir es safeRunDir para una clave de índice, cuyo change llega como
// nombre de carpeta (`_sin-change` incluido).
func (s *Store) safeKeyDir(folderName, runID string) (string, error) {
	return s.safeRunDir(unfolder(folderName), runID)
}

// BeginRun purga los vencidos al iniciar cada corrida y devuelve su ruta.
func (s *Store) BeginRun(change, runID string) (string, error) {
	if _, err := s.Purge(); err != nil {
		return "", fmt.Errorf("capture: purga de vencidos: %w", err)
	}
	return s.RunDir(change, runID), nil
}

// Lookup informa si el original de una corrida se recupera, venció o nunca existió.
func (s *Store) Lookup(change, runID string) Availability {
	if checkIdentity(change, runID) != nil {
		return Availability{State: NeverCaptured}
	}
	idx := s.load()
	e, ok := idx.Runs[key(change, runID)]
	if !ok {
		return Availability{State: NeverCaptured}
	}
	exp, hasExpiry := idx.expiry(change, e)
	if e.PurgedOn != nil {
		return Availability{State: Expired, ExpiredOn: *e.PurgedOn}
	}
	if hasExpiry && !s.Now().Before(exp) {
		return Availability{State: Expired, ExpiredOn: exp}
	}
	return Availability{State: Available, Path: filepath.Join(s.RunDir(change, runID), CopyFile)}
}

func splitKey(k string) (change, runID string) {
	for i := len(k) - 1; i >= 0; i-- {
		if k[i] == '/' {
			return k[:i], k[i+1:]
		}
	}
	return "", k
}

func unfolder(f string) string {
	if f == NoChange {
		return ""
	}
	return f
}
