package inventory_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/inventory"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Tareas 1.5, 1.6 y 1.7 (opsx-task-context): inventario de cuerpos de skill y
// tabla única de fuentes de instrucción ordenada por tamaño.

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadAll(t *testing.T) (*catalog.Catalog, inventory.ThirdParty, string) {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	root := repoRoot(t)
	tp, err := inventory.LoadThirdParty(filepath.Join(root, "eval", "taskcontext", "inventory", "skills-terceros.json"))
	if err != nil {
		t.Fatal(err)
	}
	return cat, tp, root
}

func TestSkills_CoverEverySkillTheCatalogDeclares(t *testing.T) {
	cat, tp, root := loadAll(t)
	got, err := inventory.Skills(root, cat, tp)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]inventory.Skill{}
	for _, s := range got {
		if _, dup := byID[s.ID]; dup {
			t.Errorf("la skill %q figura dos veces", s.ID)
		}
		byID[s.ID] = s
	}
	declared := 0
	for _, h := range cat.Harnesses {
		if h.Type != "skill" {
			continue
		}
		declared++
		s, ok := byID[h.ID]
		if !ok {
			t.Errorf("la skill %q del catálogo no está en el inventario", h.ID)
			continue
		}
		// Sin medir, la entrada dice por qué; nunca queda en silencio.
		if s.Bytes == 0 && s.Status == "" {
			t.Errorf("la skill %q figura sin tamaño ni motivo", h.ID)
		}
	}
	if declared < 40 {
		t.Fatalf("el catálogo declara %d skills: se esperaban más de 40", declared)
	}
	// La skill propia embebida se mide del archivo real.
	ds, ok := byID["design-system"]
	if !ok || ds.Source != inventory.SourceEmbedded || ds.Bytes < 1000 {
		t.Errorf("design-system debe medirse como embebida: %+v", ds)
	}
	// Hay de terceros medidas y las que cambiaron de layout se informan con motivo.
	var measured, unmeasured int
	for _, s := range got {
		if s.Source == inventory.SourceThirdParty {
			if s.Bytes > 0 {
				measured++
			} else {
				unmeasured++
				if !strings.Contains(s.Status, "SKILL.md") {
					t.Errorf("la skill de terceros %q sin medir debe decir por qué: %q", s.ID, s.Status)
				}
			}
		}
	}
	if measured < 30 {
		t.Errorf("se esperaban al menos 30 skills de terceros medidas, hay %d", measured)
	}
	t.Logf("skills declaradas=%d; terceros medidas=%d, sin medir=%d", declared, measured, unmeasured)
}

func TestSkills_ErrorsWhenThirdPartyMeasurementMissing(t *testing.T) {
	cat, tp, root := loadAll(t)
	// Se quita una entrada del inventario de terceros: el error nombra la skill.
	trimmed := inventory.ThirdParty{MeasuredOn: tp.MeasuredOn}
	for _, s := range tp.Skills {
		if s.ID != "kb-creator" {
			trimmed.Skills = append(trimmed.Skills, s)
		}
	}
	_, err := inventory.Skills(root, cat, trimmed)
	if err == nil || !strings.Contains(err.Error(), "kb-creator") {
		t.Fatalf("err = %v, want error que nombre kb-creator", err)
	}
}

func TestBuild_SortedBySizeDescending(t *testing.T) {
	cat, tp, root := loadAll(t)
	srcs, err := inventory.Build(root, cat, tp)
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) < 50 {
		t.Fatalf("la tabla única tiene %d fuentes, se esperaban más de 50", len(srcs))
	}
	for i := 1; i < len(srcs); i++ {
		if srcs[i-1].Bytes < srcs[i].Bytes {
			t.Fatalf("no está ordenada por tamaño descendente en %d: %s (%d) antes de %s (%d)",
				i, srcs[i-1].Name, srcs[i-1].Bytes, srcs[i].Name, srcs[i].Bytes)
		}
	}
	// El bloque inyectado de Claude y el de Codex figuran, con el total real inyectado.
	names := map[string]int{}
	for _, s := range srcs {
		names[s.Name] = s.Bytes
	}
	if names["Bloque sdd-orchestrator inyectado — Claude"] != 4382 {
		t.Errorf("bloque de Claude = %d, want 4382 (bloque universal de la entrega en capas, con marcadores)", names["Bloque sdd-orchestrator inyectado — Claude"])
	}
	if names["Sección design-mockups inyectada — Claude"] != 8940 {
		t.Errorf("sección design-mockups = %d, want 8940", names["Sección design-mockups inyectada — Claude"])
	}
}

func TestTop_ReturnsFiveLargest(t *testing.T) {
	cat, tp, root := loadAll(t)
	srcs, err := inventory.Build(root, cat, tp)
	if err != nil {
		t.Fatal(err)
	}
	top := inventory.Top(srcs, 5)
	if len(top) != 5 {
		t.Fatalf("Top(5) devolvió %d", len(top))
	}
	for _, s := range srcs[5:] {
		if s.Bytes > top[4].Bytes {
			t.Errorf("%s (%d B) es más grande que la quinta (%d B) y quedó afuera", s.Name, s.Bytes, top[4].Bytes)
		}
	}
	// Sintético: el orden es por tamaño y desempata por nombre.
	synth := []inventory.Source{{Name: "b", Bytes: 5}, {Name: "a", Bytes: 5}, {Name: "c", Bytes: 9}, {Name: "d", Bytes: 1}}
	got := inventory.Top(synth, 3)
	if got[0].Name != "c" || got[1].Name != "a" || got[2].Name != "b" {
		t.Errorf("Top sintético = %v", got)
	}
}

// docs/inventario-contexto.md se genera desde los datos; el documento
// reproduce las cifras del golden de 1.1 y registra las cinco más grandes.
func TestDoc_InventoryReproducesGolden(t *testing.T) {
	cat, tp, root := loadAll(t)
	doc, err := inventory.Markdown(root, cat, tp)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "docs", "inventario-contexto.md")
	if updateRequested() {
		if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falta docs/inventario-contexto.md (generalo con -update): %v", err)
	}
	if string(onDisk) != doc {
		t.Error("docs/inventario-contexto.md no coincide con la generación (regeneralo con -update)")
	}
	// Cifras de la tabla de proposal.md — Impact, tal como las produce Compose.
	for _, want := range []string{
		"20.065 B · 448 L", "18.010 B · 403 L", // bloque completo
		"7.116 B · 188 L", "5.543 B · 147 L", // sin toggles
		"| `delegation` | 1.005 B |", "| `engram` | 6.093 B |", "| `tdd` | 3.797 B |", // fragmentos
		"8.865 B", // design-mockups
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("el documento no reproduce la cifra %q", want)
		}
	}
	srcs, _ := inventory.Build(root, cat, tp)
	for _, s := range inventory.Top(srcs, 5) {
		if !strings.Contains(doc, "**"+s.Name+"**") {
			t.Errorf("el documento no marca entre las cinco más grandes a %q", s.Name)
		}
	}
}
