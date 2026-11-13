package fixture

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	fullCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256Hex  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// checks son los chequeos conocidos por familia. requiere_revisor vale en
// todas: el criterio queda pendiente hasta la decision del revisor.
var checks = map[Family][]string{
	FamilyExploration: {"respuesta_entrada", "respuesta_archivos", "checkout_intacto"},
	FamilySmallChange: {"go_test", "pruebas_ocultas", "prueba_regresion", "archivos_permitidos"},
	FamilyMultiFile:   {"go_test", "pruebas_ocultas", "prueba_regresion", "archivos_permitidos"},
	FamilyDiagnosis:   {"diag_reproduccion", "diag_causa_raiz", "diag_descarte", "checkout_intacto"},
	FamilyReview:      {"rev_sin_omisiones", "rev_evidencia", "rev_sin_falsos_positivos", "checkout_intacto"},
}

const checkReviewer = "requiere_revisor"

// needsKey: familias que se evaluan por hechos contra una clave.
var needsKey = map[Family]bool{FamilyExploration: true, FamilyDiagnosis: true, FamilyReview: true}

// Load carga y valida el fixture de dir: task.json estricto, huella del
// prompt, criterios obligatorios no vacios con chequeos conocidos, revision
// fija como commit completo o huella del arbol embebido, y archivos
// referenciados presentes.
func Load(dir string) (*Fixture, error) {
	b, err := os.ReadFile(filepath.Join(dir, "task.json"))
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var t Task
	if err := dec.Decode(&t); err != nil {
		return nil, fmt.Errorf("task.json: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("task.json: contenido extra tras el documento")
	}
	f := &Fixture{Dir: dir, Task: t}
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("fixture %s: %w", dir, err)
	}
	return f, nil
}

func (f *Fixture) validate() error {
	t := f.Task
	if strings.TrimSpace(t.ID) == "" || strings.TrimSpace(t.RubricVersion) == "" {
		return fmt.Errorf("id y rubric_version son obligatorios")
	}
	known, ok := checks[t.Family]
	if !ok {
		return fmt.Errorf("familia desconocida %q", t.Family)
	}
	if len(t.AllowedTools) == 0 {
		return fmt.Errorf("sin herramientas permitidas")
	}
	if len(t.Limits) == 0 {
		return fmt.Errorf("sin límites (usar (TBD) si no están acordados)")
	}
	for k, v := range t.Limits {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("límite %q vacío", k)
		}
	}
	// Prompt congelado.
	if !sha256Hex.MatchString(t.Prompt.SHA256) {
		return fmt.Errorf("huella de prompt inválida %q", t.Prompt.SHA256)
	}
	p, err := f.read(t.Prompt.File)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(p)
	if got := hex.EncodeToString(sum[:]); got != t.Prompt.SHA256 {
		return fmt.Errorf("huella de prompt %s no coincide con la congelada %s", got, t.Prompt.SHA256)
	}
	if len(bytes.TrimSpace(p)) == 0 {
		return fmt.Errorf("prompt vacío")
	}
	f.Prompt = p
	// Criterios.
	if len(t.Criteria) == 0 {
		return fmt.Errorf("sin criterios")
	}
	ids := map[string]bool{}
	mandatory := 0
	for _, c := range t.Criteria {
		if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.Description) == "" {
			return fmt.Errorf("criterio sin id o sin descripción")
		}
		if ids[c.ID] {
			return fmt.Errorf("id de criterio duplicado %q", c.ID)
		}
		ids[c.ID] = true
		if c.Mandatory {
			mandatory++
		}
		if c.Check != checkReviewer && !contains(known, c.Check) {
			return fmt.Errorf("criterio %q: chequeo desconocido %q para la familia %s", c.ID, c.Check, t.Family)
		}
		if (c.Check == "pruebas_ocultas" || c.Check == "archivos_permitidos") && len(c.Args) == 0 {
			return fmt.Errorf("criterio %q: el chequeo %s necesita args", c.ID, c.Check)
		}
		if c.Check == "pruebas_ocultas" {
			if err := safeName(c.Args[0]); err != nil {
				return fmt.Errorf("criterio %q: %w", c.ID, err)
			}
			if st, err := os.Stat(f.path(c.Args[0])); err != nil || !st.IsDir() {
				return fmt.Errorf("criterio %q: %q no es un directorio del fixture", c.ID, c.Args[0])
			}
		}
	}
	if mandatory == 0 {
		return fmt.Errorf("sin criterios obligatorios")
	}
	// Clave, referencia y variante defectuosa.
	if needsKey[t.Family] {
		if t.Key == "" {
			return fmt.Errorf("la familia %s necesita clave", t.Family)
		}
		if _, err := f.read(t.Key); err != nil {
			return fmt.Errorf("clave: %w", err)
		}
	}
	for _, name := range []string{t.Reference, t.Defective} {
		if name == "" {
			return fmt.Errorf("faltan referencia o variante defectuosa")
		}
		if _, err := f.read(name); err != nil {
			return err
		}
	}
	// Estado inicial congelado.
	s := t.Source
	if !sha256Hex.MatchString(s.SHA256) {
		return fmt.Errorf("huella de árbol inválida %q", s.SHA256)
	}
	switch s.Kind {
	case "git":
		if !fullCommit.MatchString(s.Revision) {
			return fmt.Errorf("revisión %q: se exige un commit completo de 40 hex", s.Revision)
		}
		if s.Dir != "" {
			return fmt.Errorf("fuente git con dir")
		}
		if s.Patch != "" {
			if _, err := f.read(s.Patch); err != nil {
				return err
			}
		}
	case "embedded":
		if s.Dir == "" || s.Revision != "" || s.Patch != "" {
			return fmt.Errorf("fuente embebida: dir obligatorio, sin revision ni patch")
		}
		if err := safeName(s.Dir); err != nil {
			return err
		}
		got, err := TreeHash(f.path(s.Dir))
		if err != nil {
			return err
		}
		if got != s.SHA256 {
			return fmt.Errorf("huella de árbol %s no coincide con la congelada %s", got, s.SHA256)
		}
	default:
		return fmt.Errorf("tipo de fuente desconocido %q", s.Kind)
	}
	return nil
}

// path resuelve un nombre relativo del fixture sin salir de su directorio.
func (f *Fixture) path(name string) string {
	return filepath.Join(f.Dir, filepath.FromSlash(name))
}

// safeName rechaza rutas vacias, absolutas, con volumen o que salgan del
// directorio del fixture (incluidos los .. internos).
func safeName(name string) error {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	if name == "" || filepath.IsAbs(name) || filepath.VolumeName(name) != "" ||
		strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) ||
		clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("ruta de fixture inválida %q", name)
	}
	return nil
}

func (f *Fixture) read(name string) ([]byte, error) {
	if err := safeName(name); err != nil {
		return nil, err
	}
	return os.ReadFile(f.path(name))
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// TreeHash es la huella del arbol de archivos regulares de dir: sha256 de las
// lineas "<sha256 del contenido>  <ruta relativa con />\n" ordenadas por ruta
// (el mismo formato que `sha256sum`, para poder recalcularla con shell). Los
// directorios no cuentan; un archivo que no es regular es error.
func TreeHash(dir string) (string, error) {
	files, err := fileHashes(dir)
	if err != nil {
		return "", err
	}
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, p := range paths {
		fmt.Fprintf(h, "%s  %s\n", files[p], p)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// fileHashes devuelve ruta relativa -> sha256 del contenido.
func fileHashes(dir string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if err := entryKind(d, p); err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		fh, err := os.Open(p)
		if err != nil {
			return err
		}
		defer fh.Close()
		h := sha256.New()
		if _, err := io.Copy(h, fh); err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = hex.EncodeToString(h.Sum(nil))
		return nil
	})
	return out, err
}

// copyTree copia los archivos regulares de src en dst (que puede existir).
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("archivo no regular: %s", p)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

var errNonRegular = errors.New("archivo no regular en el árbol")

// entryKind acepta directorios y archivos regulares; cualquier otra entrada
// (symlink, pipe, socket) es errNonRegular.
func entryKind(d fs.DirEntry, p string) error {
	if d.IsDir() || d.Type().IsRegular() {
		return nil
	}
	return fmt.Errorf("%w: %s", errNonRegular, p)
}

// emptyDirs lista (con /, ordenados) los directorios de dir sin ningun
// archivo regular en su subarbol. TreeHash no los ve; un fixture congelado no
// los tiene (git no versiona directorios vacios), asi que un directorio vacio
// en un checkout es una adicion del agente.
func emptyDirs(dir string) ([]string, error) {
	dirs := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := entryKind(d, p); err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if rel != "." {
				dirs[filepath.ToSlash(rel)] = true
			}
			return nil
		}
		for up := filepath.ToSlash(filepath.Dir(rel)); up != "."; up = path.Dir(up) {
			delete(dirs, up)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Un directorio con un subdirectorio que si tiene archivos ya fue
	// borrado arriba (los ancestros se quitan); los restantes no tienen
	// archivos debajo.
	out := make([]string, 0, len(dirs))
	for d := range dirs {
		out = append(out, d)
	}
	sort.Strings(out)
	return out, nil
}

// hashVerdict traduce el resultado de TreeHash: una entrada no regular o una
// huella distinta son un checkout modificado, no un error del arnes.
func hashVerdict(h string, err error, want string) (bool, string, error) {
	switch {
	case errors.Is(err, errNonRegular):
		return false, err.Error(), nil
	case err != nil:
		return false, "", err
	case h != want:
		return false, "el checkout fue modificado", nil
	}
	return true, "", nil
}

// intact compara el arbol de dir con la huella congelada y exige que no haya
// directorios vacios ni entradas no regulares. Un symlink o un directorio
// vacio no son error del arnes: son un checkout modificado (ok=false con el
// motivo). Los errores de E/S si son error.
func intact(dir, want string) (ok bool, detail string, err error) {
	h, err := TreeHash(dir)
	if ok, detail, err := hashVerdict(h, err, want); !ok || err != nil {
		return ok, detail, err
	}
	empty, err := emptyDirs(dir)
	if err != nil {
		return false, "", err
	}
	if len(empty) > 0 {
		return false, "directorio vacío agregado: " + empty[0], nil
	}
	return true, "", nil
}
