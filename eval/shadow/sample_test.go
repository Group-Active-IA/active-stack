package shadow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/review"
)

var repoRoot = filepath.Join("..", "..")

const manifest = "muestra/preparados.json"

func loadPrepared(t *testing.T) []PreparedCase {
	t.Helper()
	cs, err := LoadPrepared(repoRoot, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) == 0 {
		t.Fatal("muestra preparada vacía")
	}
	return cs
}

// 1.2: la muestra preparada tiene el diff sembrado de 02 y los cuatro rechazos de 06 que nombra D2.
func TestPrepared_CoversD2Classes(t *testing.T) {
	var classes []string
	for _, c := range loadPrepared(t) {
		classes = append(classes, c.Entry.Class)
	}
	for _, want := range []string{
		"diff sembrado",
		"recibo falso con hash correcto",
		"mismo HEAD con diff distinto",
		"salida truncada sin original",
		"tests verdes sin cubrir el requisito",
	} {
		if !slices.Contains(classes, want) {
			t.Errorf("falta la clase %q (hay %q)", want, classes)
		}
	}
}

// 1.2: la etiqueta se lee de la fuente sin reescribirla.
func TestPrepared_LabelsComeFromSource(t *testing.T) {
	for _, c := range loadPrepared(t) {
		switch c.Entry.Source {
		case Source02:
			var key struct {
				Defects []review.Defect `json:"defects"`
			}
			readJSONFile(t, filepath.Join(repoRoot, c.Entry.Fixture, "key.json"), &key)
			if len(key.Defects) == 0 || !slices.Equal(c.Label.Defects, key.Defects) {
				t.Errorf("%s: defectos %+v, key.json %+v", c.Entry.ID, c.Label.Defects, key.Defects)
			}
		case Source06:
			var src struct {
				Want struct {
					Status string `json:"estado"`
				} `json:"esperado"`
			}
			readJSONFile(t, filepath.Join(repoRoot, c.Entry.Fixture), &src)
			if src.Want.Status == "" || string(c.Label.Status) != src.Want.Status {
				t.Errorf("%s: estado %q, el caso de 06 dice %q", c.Entry.ID, c.Label.Status, src.Want.Status)
			}
		default:
			t.Errorf("%s: fuente desconocida %q", c.Entry.ID, c.Entry.Source)
		}
	}
}

// D2, paquetes sin pistas: el material no contiene la respuesta ni nada que la nombre.
func TestPrepared_MaterialHasNoClues(t *testing.T) {
	for _, c := range loadPrepared(t) {
		clues := []string{c.Entry.ID, `"esperado"`, `"descripcion"`, "key.json", "reference.json", "defective.json"}
		if c.Entry.Source == Source06 {
			var src struct {
				ID string `json:"id"`
			}
			readJSONFile(t, filepath.Join(repoRoot, c.Entry.Fixture), &src)
			clues = append(clues, src.ID)
		}
		for _, d := range c.Label.Defects {
			clues = append(clues, d.Mechanism)
		}
		if len(c.Material.Files) == 0 {
			t.Errorf("%s: material vacío", c.Entry.ID)
		}
		for _, f := range c.Material.Files {
			for _, clue := range clues {
				if clue != "" && (strings.Contains(string(f.Data), clue) || strings.Contains(f.Path, clue)) {
					t.Errorf("%s: %s contiene la pista %q", c.Entry.ID, f.Path, clue)
				}
			}
		}
	}
}

// 1.2 × 3.2: el material de la primera ronda es público o ficticio y pasa el examen previo.
func TestPrepared_MaterialPassesPreSendScan(t *testing.T) {
	s := newScanner(t)
	for _, c := range loadPrepared(t) {
		if r := s.Scan(c.Material); !r.Send {
			t.Errorf("%s: el examen previo cancela el envío: %+v", c.Entry.ID, r.Hits)
		}
	}
}

// Si un fixture de 02 o 06 cambia, la huella fijada falla: la muestra no se mueve sola.
func TestPrepared_PinnedFingerprintDetectsChange(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash("eval/shadow/"+manifest)))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Cases []PreparedEntry `json:"casos"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for p := range m.Cases[0].Pinned {
		m.Cases[0].Pinned[p] = strings.Repeat("0", 64)
		break
	}
	tampered := filepath.Join(t.TempDir(), "preparados.json")
	out, _ := json.Marshal(m)
	if err := os.WriteFile(tampered, out, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPrepared(repoRoot, tampered); err == nil {
		t.Error("una huella fijada que no coincide debía fallar")
	}
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
