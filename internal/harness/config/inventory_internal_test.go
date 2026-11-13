package config

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Tareas 1.1 y 1.2 (opsx-task-context): inventario de lo que produce Compose.
//
// El golden registra, por variante de assets, los bytes y las líneas (wc -l:
// cantidad de saltos de línea) del bloque compuesto. Un cambio de asset de otra
// ficha se resuelve regenerándolo con `go test ./internal/harness/config -update`
// (la bandera ya existe en compose_test.go).

// cumulativeToggles son las seis combinaciones acumulativas de la tabla de
// proposal.md — Impact, en el orden en que se suman.
var cumulativeToggles = []struct {
	label  string
	toggle string
}{
	{"none", ""},
	{"+delegation", "delegation"},
	{"+model-routing", "model-routing"},
	{"+governance", "governance"},
	{"+engram", "engram"},
	{"+tdd", "tdd"},
}

// parityVariants son los dos clientes dentro del alcance de paridad de la ficha.
var parityVariants = []string{"claude", "codex"}

func measure(s string) (bytes, lines int) {
	return len(s), strings.Count(s, "\n")
}

// assetVariants lista las variantes de assets/ (directorios con
// sdd-orchestrator.md), ordenadas.
func assetVariants(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(embeddedAssets, "assets")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := fs.Stat(embeddedAssets, "assets/"+e.Name()+"/sdd-orchestrator.md"); err == nil {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// composeInventory arma el texto del golden.
func composeInventory(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Inventario de Compose (opsx-task-context, tareas 1.1 y 1.2)\n")
	b.WriteString("# bytes y líneas (wc -l: cantidad de saltos de línea) del bloque compuesto.\n")
	b.WriteString("# Regenerar: go test ./internal/harness/config -run TestInventory_ComposeGolden -update\n")
	b.WriteString("\n## Paridad: combinaciones acumulativas de toggles\n")
	b.WriteString("# variant\ttoggles\tbytes\tlines\n")
	for _, v := range parityVariants {
		var active []string
		for _, step := range cumulativeToggles {
			if step.toggle != "" {
				active = append(active, step.toggle)
			}
			got, err := Compose(active, v)
			if err != nil {
				t.Fatalf("Compose(%v, %s): %v", active, v, err)
			}
			by, li := measure(got)
			fmt.Fprintf(&b, "%s\t%s\t%d\t%d\n", v, step.label, by, li)
		}
	}
	b.WriteString("\n## Referencia: demás variantes (fuera del alcance de paridad)\n")
	b.WriteString("# variant\tscope\tbytes\tlines\n")
	all := []string{"delegation", "model-routing", "governance", "engram", "tdd"}
	for _, v := range assetVariants(t) {
		if v == "claude" || v == "codex" {
			continue
		}
		base, err := Compose(nil, v)
		if err != nil {
			t.Fatal(err)
		}
		full, err := Compose(all, v)
		if err != nil {
			t.Fatal(err)
		}
		bb, bl := measure(base)
		fb, fl := measure(full)
		fmt.Fprintf(&b, "%s\tnone\t%d\t%d\n", v, bb, bl)
		fmt.Fprintf(&b, "%s\tall\t%d\t%d\n", v, fb, fl)
	}
	return b.String()
}

func updateRequested() bool {
	f := flag.Lookup("update")
	return f != nil && f.Value.String() == "true"
}

// checkInventoryGolden compara got con el golden; con -update lo reescribe.
func checkInventoryGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if updateRequested() {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falta el golden %s (generalo con -update): %v", path, err)
	}
	if d := diffLines(string(want), got); d != "" {
		t.Errorf("el inventario no coincide con %s (si un asset cambió a propósito, regeneralo con -update):\n%s", path, d)
	}
}

// diffLines devuelve las líneas que difieren, o "" si son idénticos.
func diffLines(want, got string) string {
	if want == got {
		return ""
	}
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	var b strings.Builder
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			fmt.Fprintf(&b, "línea %d:\n  golden: %q\n  actual: %q\n", i+1, wl, gl)
		}
	}
	return b.String()
}

func TestDiffLines(t *testing.T) {
	if d := diffLines("a\nb", "a\nb"); d != "" {
		t.Errorf("idénticos dieron diff: %q", d)
	}
	d := diffLines("a\nb\nc", "a\nX\nc\nd")
	if !strings.Contains(d, "línea 2") || !strings.Contains(d, "línea 4") || strings.Contains(d, "línea 1") {
		t.Errorf("diff inesperado: %q", d)
	}
}

func TestInventory_ComposeGolden(t *testing.T) {
	checkInventoryGolden(t, "inventory_compose", composeInventory(t))
}

// Ninguna variante de assets/ queda sin medición registrada (1.2).
func TestInventory_EveryVariantMeasured(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "inventory_compose.golden"))
	if err != nil {
		t.Fatalf("falta el golden: %v", err)
	}
	text := string(data)
	variants := assetVariants(t)
	if len(variants) < 8 {
		t.Fatalf("se esperaban al menos ocho variantes en assets/, hay %d: %v", len(variants), variants)
	}
	for _, v := range variants {
		if !strings.Contains(text, "\n"+v+"\t") {
			t.Errorf("la variante %q no tiene medición en el golden de inventario", v)
		}
	}
}

// El golden refleja los bytes reales de Compose para los dos clientes de
// paridad: los valores de partida de proposal.md (1b4639b) son el piso de
// referencia, y el bloque instalado hoy es la combinación con los cinco toggles.
func TestInventory_ParityFullBlockMatchesCompose(t *testing.T) {
	all := []string{"delegation", "model-routing", "governance", "engram", "tdd"}
	for _, v := range parityVariants {
		got, err := Compose(all, v)
		if err != nil {
			t.Fatal(err)
		}
		by, li := measure(got)
		data, _ := os.ReadFile(filepath.Join("testdata", "inventory_compose.golden"))
		want := fmt.Sprintf("%s\t+tdd\t%d\t%d\n", v, by, li)
		if !bytes.Contains(data, []byte(want)) {
			t.Errorf("el golden no registra %q para el bloque completo", strings.TrimSpace(want))
		}
	}
}
