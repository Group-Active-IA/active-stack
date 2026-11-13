package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/Group-Active-IA/active-stack/eval/baseline/review"
	"github.com/Group-Active-IA/active-stack/eval/evidence"
)

// Fuentes de la muestra preparada (D2).
const (
	Source02 = "02" // familia "Revisión de un diff sembrado" (eval/baseline/fixtures/revision)
	Source06 = "06" // casos de rechazo de los recibos (eval/evidence/testdata/casos)
)

// material02 es lo que ve el juez de un fixture de revisión de 02: la consigna y el diff. key.json,
// reference.json y defective.json son la respuesta; task.json describe los criterios de la rúbrica.
var material02 = []string{"prompt.md", "seeded.patch"}

// PreparedEntry es una fila del manifiesto: referencia a un fixture de otro change, con la huella
// de cada archivo de origen fijada para detectar que alguien lo cambió.
type PreparedEntry struct {
	ID      string            `json:"id"`
	Source  string            `json:"fuente"`
	Class   string            `json:"clase"`
	Fixture string            `json:"fixture"` // directorio (02) o archivo de caso (06), relativo a la raíz
	Pinned  map[string]string `json:"huellas"` // ruta relativa a la raíz -> sha256
}

// Label es la verdad del caso, leída de la fuente sin reescribirla.
type Label struct {
	Defects []review.Defect // 02: key.json
	Status  evidence.Status // 06: "esperado.estado" del caso
}

// PreparedCase es un caso de la muestra preparada: su etiqueta y el material que vería el juez.
type PreparedCase struct {
	Entry    PreparedEntry
	Label    Label
	Material Package
}

// LoadPrepared lee el manifiesto, comprueba las huellas fijadas y arma cada caso. El material
// excluye todo lo que revela la respuesta (D2, paquetes sin pistas). root es la raíz del repo y
// manifest la ruta del manifiesto tal como se pasa.
func LoadPrepared(root, manifest string) ([]PreparedCase, error) {
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}
	var m struct {
		Cases []PreparedEntry `json:"casos"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", manifest, err)
	}
	var out []PreparedCase
	for _, e := range m.Cases {
		if err := checkPins(root, e); err != nil {
			return nil, err
		}
		var c PreparedCase
		switch e.Source {
		case Source02:
			c, err = load02(root, e)
		case Source06:
			c, err = load06(root, e)
		default:
			err = fmt.Errorf("%s: fuente desconocida %q", e.ID, e.Source)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func checkPins(root string, e PreparedEntry) error {
	if len(e.Pinned) == 0 {
		return fmt.Errorf("%s: sin huellas fijadas", e.ID)
	}
	for rel, want := range e.Pinned {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("%s: %w", e.ID, err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != want {
			return fmt.Errorf("%s: %s cambió (sha256 %s, fijado %s)", e.ID, rel, got, want)
		}
	}
	return nil
}

func load02(root string, e PreparedEntry) (PreparedCase, error) {
	dir := filepath.Join(root, filepath.FromSlash(e.Fixture))
	var key struct {
		Defects []review.Defect `json:"defects"`
	}
	if err := readJSON(filepath.Join(dir, "key.json"), &key); err != nil {
		return PreparedCase{}, err
	}
	if len(key.Defects) == 0 {
		return PreparedCase{}, fmt.Errorf("%s: key.json sin defectos", e.ID)
	}
	p, err := Build(e.ID, dir, material02)
	if err != nil {
		return PreparedCase{}, err
	}
	return PreparedCase{Entry: e, Label: Label{Defects: key.Defects}, Material: p}, nil
}

// load06 separa un caso de 06 en etiqueta (esperado) y material: requisito, recibos, entorno y las
// copias redactadas que referencia. id, descripcion y esperado no viajan.
func load06(root string, e PreparedEntry) (PreparedCase, error) {
	file := filepath.Join(root, filepath.FromSlash(e.Fixture))
	var src struct {
		Requirement string          `json:"requisito"`
		Receipts    json.RawMessage `json:"recibos"`
		Env         json.RawMessage `json:"entorno"`
		Want        struct {
			Status evidence.Status `json:"estado"`
		} `json:"esperado"`
	}
	if err := readJSON(file, &src); err != nil {
		return PreparedCase{}, err
	}
	if src.Want.Status == "" {
		return PreparedCase{}, fmt.Errorf("%s: el caso de 06 no tiene estado esperado", e.ID)
	}
	body, err := json.MarshalIndent(struct {
		Requirement string          `json:"requisito"`
		Receipts    json.RawMessage `json:"recibos"`
		Env         json.RawMessage `json:"entorno"`
	}{src.Requirement, src.Receipts, src.Env}, "", "  ")
	if err != nil {
		return PreparedCase{}, err
	}
	files := []File{{Path: "caso.json", Data: append(body, '\n')}}

	var env struct {
		Copies map[string]string `json:"copias"`
	}
	if len(src.Env) > 0 {
		if err := json.Unmarshal(src.Env, &env); err != nil {
			return PreparedCase{}, err
		}
	}
	runs := make([]string, 0, len(env.Copies))
	for run := range env.Copies {
		runs = append(runs, run)
	}
	sort.Strings(runs)
	testdata := filepath.Dir(filepath.Dir(file))
	for _, run := range runs {
		rel := env.Copies[run]
		data, err := os.ReadFile(filepath.Join(testdata, filepath.FromSlash(rel)))
		if err != nil {
			return PreparedCase{}, fmt.Errorf("%s: copia %s: %w", e.ID, rel, err)
		}
		files = append(files, File{Path: path.Clean(rel), Data: data})
	}
	p, err := FromFiles(e.ID, files)
	if err != nil {
		return PreparedCase{}, err
	}
	return PreparedCase{Entry: e, Label: Label{Status: src.Want.Status}, Material: p}, nil
}

func readJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
