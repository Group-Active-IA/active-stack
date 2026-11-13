package catalog

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Tarea 4.2 (opsx-task-context, D11): el umbral de lectura completa tiene que
// quedar entre el p85 y el p92 (rango más cercano) de la distribución de
// líneas de los .go versionados bajo internal/ y cmd/ que no son _test.go.

// nearestRank devuelve el percentil p (0-100) por rango más cercano sobre un
// arreglo ordenado: el valor en la posición ceil(p/100*N), contada desde 1.
func nearestRank(sorted []int, p float64) int {
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	return sorted[rank-1]
}

func TestNearestRank(t *testing.T) {
	hundred := make([]int, 100)
	for i := range hundred {
		hundred[i] = i + 1
	}
	tests := []struct {
		name string
		data []int
		p    float64
		want int
	}{
		{"p85 de 1..100", hundred, 85, 85},
		{"p92 de 1..100", hundred, 92, 92},
		{"p50 de cuatro valores", []int{10, 20, 30, 40}, 50, 20},
		{"p85 de diez valores redondea hacia arriba", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 85, 9},
		{"p0 cae en el primero", []int{7, 8}, 0, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nearestRank(tt.data, tt.p); got != tt.want {
				t.Errorf("nearestRank(p=%v) = %d, want %d", tt.p, got, tt.want)
			}
		})
	}
}

// versionedGoLines cuenta las líneas (saltos de línea, como wc -l) de cada
// archivo .go versionado bajo internal/ y cmd/ que no termina en _test.go.
func versionedGoLines(t *testing.T) []int {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "ls-files", "--", "internal/*.go", "cmd/*.go")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("git ls-files no disponible (%v): la distribución se mide sobre el árbol versionado", err)
	}
	var lines []int
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f == "" || strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil {
			continue // versionado pero ausente en el árbol de trabajo
		}
		lines = append(lines, bytes.Count(data, []byte("\n")))
	}
	sort.Ints(lines)
	return lines
}

func TestReadThreshold_BetweenP85AndP92(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	lines := versionedGoLines(t)
	if len(lines) < 50 {
		t.Fatalf("la muestra es demasiado chica (%d archivos): el cálculo no es significativo", len(lines))
	}
	p85, p92 := nearestRank(lines, 85), nearestRank(lines, 92)
	thr := c.TaskContext.ReadThreshold.Value
	if thr < p85 || thr > p92 {
		t.Errorf("umbral de lectura completa = %d líneas, fuera de [p85=%d, p92=%d] sobre %d archivos .go: elegí el valor nuevo con el mismo criterio (design.md D11)",
			thr, p85, p92, len(lines))
	}
	t.Logf("muestra: %d archivos; p85=%d p92=%d; umbral=%d", len(lines), p85, p92, thr)
}
