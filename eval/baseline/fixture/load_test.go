package fixture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixturesDir = "../fixtures"

var allFamilies = []Family{FamilyExploration, FamilySmallChange, FamilyMultiFile, FamilyDiagnosis, FamilyReview}

func mustLoad(t *testing.T, fam Family) *Fixture {
	t.Helper()
	f, err := Load(filepath.Join(fixturesDir, string(fam)))
	if err != nil {
		t.Fatalf("Load(%s): %v", fam, err)
	}
	return f
}

func TestLoad_FiveFamiliesValid(t *testing.T) {
	for _, fam := range allFamilies {
		f := mustLoad(t, fam)
		if f.Task.Family != fam {
			t.Errorf("%s: family = %q", fam, f.Task.Family)
		}
		if len(f.Prompt) == 0 {
			t.Errorf("%s: prompt vacio", fam)
		}
		if len(f.Task.Criteria) == 0 {
			t.Errorf("%s: sin criterios", fam)
		}
	}
}

// copia el fixture a un temporal y aplica edit sobre task.json.
func tamper(t *testing.T, fam Family, edit func(m map[string]any), files map[string]string) string {
	t.Helper()
	dst := t.TempDir()
	if err := copyTree(filepath.Join(fixturesDir, string(fam)), dst); err != nil {
		t.Fatal(err)
	}
	if edit != nil {
		b, _ := os.ReadFile(filepath.Join(dst, "task.json"))
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		edit(m)
		b, _ = json.Marshal(m)
		os.WriteFile(filepath.Join(dst, "task.json"), b, 0o644)
	}
	for p, c := range files {
		os.WriteFile(filepath.Join(dst, filepath.FromSlash(p)), []byte(c), 0o644)
	}
	return dst
}

func TestLoad_RejectsInvalid(t *testing.T) {
	crit := func(m map[string]any) []any { return m["criteria"].([]any) }
	cases := []struct {
		name string
		fam  Family
		edit func(m map[string]any)
		file map[string]string
		want string
	}{
		{"prompt alterado", FamilySmallChange, nil, map[string]string{"prompt.md": "otro prompt\n"}, "huella"},
		{"huella de prompt alterada", FamilySmallChange, func(m map[string]any) { m["prompt"].(map[string]any)["sha256"] = strings.Repeat("0", 64) }, nil, "huella"},
		{"fuente embebida alterada", FamilySmallChange, nil, map[string]string{"project/pad.go": "package texto\n"}, "árbol"},
		{"sin criterios", FamilySmallChange, func(m map[string]any) { m["criteria"] = []any{} }, nil, "criterio"},
		{"sin obligatorios", FamilySmallChange, func(m map[string]any) {
			for _, c := range crit(m) {
				c.(map[string]any)["mandatory"] = false
			}
		}, nil, "obligatorio"},
		{"chequeo desconocido", FamilySmallChange, func(m map[string]any) { crit(m)[0].(map[string]any)["check"] = "go_tset" }, nil, "chequeo"},
		{"id de criterio duplicado", FamilySmallChange, func(m map[string]any) { crit(m)[1].(map[string]any)["id"] = "S1" }, nil, "duplicado"},
		{"familia desconocida", FamilySmallChange, func(m map[string]any) { m["family"] = "otra" }, nil, "familia desconocida"},
		{"revision corta", FamilyExploration, func(m map[string]any) { m["source"].(map[string]any)["revision"] = "1b4639b" }, nil, "revisión"},
		{"revision no hex", FamilyExploration, func(m map[string]any) {
			m["source"].(map[string]any)["revision"] = strings.Repeat("g", 40)
		}, nil, "revisión"},
		{"familia con clave sin clave", FamilyReview, func(m map[string]any) { delete(m, "key") }, nil, "necesita clave"},
		{"sin herramientas", FamilyReview, func(m map[string]any) { m["allowed_tools"] = []any{} }, nil, "herramientas"},
		{"sin limites", FamilyReview, func(m map[string]any) { m["limits"] = map[string]any{} }, nil, "límites"},
		{"campo desconocido", FamilyReview, func(m map[string]any) { m["extra"] = 1 }, nil, "extra"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := tamper(t, tc.fam, tc.edit, tc.file)
			_, err := Load(dir)
			if err == nil {
				t.Fatalf("Load = nil, want error con %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want mencion de %q", err, tc.want)
			}
		})
	}
}

// Las copias sin tocar siguen cargando: el test de rechazo no pasa por
// romper la copia en si.
func TestLoad_UntamperedCopyLoads(t *testing.T) {
	for _, fam := range allFamilies {
		if _, err := Load(tamper(t, fam, func(map[string]any) {}, nil)); err != nil {
			t.Errorf("%s copia intacta: %v", fam, err)
		}
	}
}
