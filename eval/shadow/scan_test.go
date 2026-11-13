package shadow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

func newScanner(t *testing.T) *Scanner {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewScanner(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Cada directorio de testdata/examen es un caso: pkg/ es la raíz, caso.json dice qué incluir, si
// se envía y qué reglas deben aparecer. Valores ficticios.
type scanCase struct {
	Include []string `json:"incluir"`
	Send    bool     `json:"envia"`
	Rules   []string `json:"reglas"`
}

// 3.2: secreto en un diff público, ruta del home, variable de entorno, `.env` por nombre y
// settings.json cancelan el envío; un paquete limpio se envía.
func TestScan_Cases(t *testing.T) {
	s := newScanner(t)
	dirs, err := filepath.Glob(filepath.Join("testdata", "examen", "*"))
	if err != nil || len(dirs) == 0 {
		t.Fatalf("sin casos en testdata/examen: %v", err)
	}
	for _, d := range dirs {
		t.Run(filepath.Base(d), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(d, "caso.json"))
			if err != nil {
				t.Fatal(err)
			}
			var c scanCase
			if err := json.Unmarshal(raw, &c); err != nil {
				t.Fatal(err)
			}
			p, err := Build(filepath.Base(d), filepath.Join(d, "pkg"), c.Include)
			if err != nil {
				t.Fatal(err)
			}
			got := s.Scan(p)
			if got.Send != c.Send {
				t.Errorf("envía = %v, esperado %v (positivos: %+v)", got.Send, c.Send, got.Hits)
			}
			var rules []string
			for _, h := range got.Hits {
				rules = append(rules, h.Rule)
			}
			for _, r := range c.Rules {
				if !slices.Contains(rules, r) {
					t.Errorf("falta la regla %q (positivos: %+v)", r, got.Hits)
				}
			}
			if c.Send && len(got.Hits) > 0 {
				t.Errorf("paquete limpio con positivos: %+v", got.Hits)
			}
		})
	}
}

// Los casos que pide la tarea 3.2 existen.
func TestScan_CoverTask(t *testing.T) {
	for _, id := range []string{"limpio", "secreto-en-diff", "ruta-home", "variable-de-entorno", "settings-json", "dotenv-por-nombre"} {
		if _, err := os.Stat(filepath.Join("testdata", "examen", id, "caso.json")); err != nil {
			t.Errorf("falta el caso %s: %v", id, err)
		}
	}
}

// Sin patrones no hay examen: el escáner no se arma con una lista vacía.
func TestNewScanner_RequiresPatterns(t *testing.T) {
	if _, err := NewScanner(nil); err == nil {
		t.Error("NewScanner(nil) debía fallar")
	}
}
