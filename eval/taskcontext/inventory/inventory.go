// Package inventory mide el contexto de partida del instalador (change
// opsx-task-context, grupo 1): bytes de los bloques de instrucción que llegan
// al proyecto, tamaño de cada cuerpo de skill del catálogo y la tabla única
// ordenada por tamaño. Vive fuera del binario (CHANGES.md:49): solo biblioteca
// estándar e internal/catalog; los goldens de Compose se leen como datos.
package inventory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// Origen de una skill.
const (
	SourceEmbedded           = "embebida"
	SourceThirdParty         = "terceros"
	SourceEmbeddedUndeclared = "embebida sin harness en el catálogo"
)

// ThirdPartySkill es la medición de una skill clonada de un repo de terceros,
// hecha una vez contra el commit resuelto (las refs `latest` no son estables).
type ThirdPartySkill struct {
	ID          string `json:"id"`
	Repo        string `json:"repo"`
	DeclaredRef string `json:"declared_ref"`
	Path        string `json:"path"`
	ResolvedSHA string `json:"resolved_sha"`
	Status      string `json:"status"`
	Bytes       int    `json:"skill_md_bytes"`
	Lines       int    `json:"skill_md_lines"`
}

// ThirdParty es el inventario de skills de terceros.
type ThirdParty struct {
	MeasuredOn string            `json:"measured_on"`
	Unit       string            `json:"unit"`
	Skills     []ThirdPartySkill `json:"skills"`
}

// LoadThirdParty lee el inventario de terceros.
func LoadThirdParty(path string) (ThirdParty, error) {
	var tp ThirdParty
	data, err := os.ReadFile(path)
	if err != nil {
		return tp, err
	}
	if err := json.Unmarshal(data, &tp); err != nil {
		return tp, fmt.Errorf("inventory: %s: %w", path, err)
	}
	return tp, nil
}

// Skill es el tamaño del cuerpo (SKILL.md) de una skill.
type Skill struct {
	ID     string
	Source string
	Bytes  int
	Lines  int
	// Status explica por qué no hay medición; vacío si se midió.
	Status string
	Ref    string
}

// Skills mide todas las skills que declara el catálogo: las embebidas desde su
// archivo y las de terceros desde el inventario. Una skill declarada sin
// medición ni motivo es un error: el inventario no puede quedar incompleto.
func Skills(root string, cat *catalog.Catalog, tp ThirdParty) ([]Skill, error) {
	third := make(map[string]ThirdPartySkill, len(tp.Skills))
	for _, s := range tp.Skills {
		third[s.ID] = s
	}
	var out []Skill
	declared := map[string]bool{}
	for _, h := range cat.Harnesses {
		if h.Type != "skill" {
			continue
		}
		declared[h.ID] = true
		if h.Source != nil && h.Source.Method == "embed" {
			by, li, err := fileSize(filepath.Join(root, "assets", "skills", h.ID, "SKILL.md"))
			if err != nil {
				return nil, fmt.Errorf("inventory: skill embebida %q: %w", h.ID, err)
			}
			out = append(out, Skill{ID: h.ID, Source: SourceEmbedded, Bytes: by, Lines: li})
			continue
		}
		m, ok := third[h.ID]
		if !ok {
			return nil, fmt.Errorf("inventory: la skill de terceros %q del catálogo no tiene medición en el inventario", h.ID)
		}
		s := Skill{ID: h.ID, Source: SourceThirdParty, Ref: shortSHA(m.ResolvedSHA)}
		if m.Status == "medido" {
			s.Bytes, s.Lines = m.Bytes, m.Lines
		} else {
			s.Status = fmt.Sprintf("sin SKILL.md en %s del commit %s (el layout del repo cambió: %s)", m.Path, shortSHA(m.ResolvedSHA), m.Status)
		}
		out = append(out, s)
	}
	// Skills embebidas que el catálogo no declara como harness propio.
	dirs, err := os.ReadDir(filepath.Join(root, "assets", "skills"))
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		if !d.IsDir() || declared[d.Name()] {
			continue
		}
		by, li, err := fileSize(filepath.Join(root, "assets", "skills", d.Name(), "SKILL.md"))
		if err != nil {
			continue
		}
		out = append(out, Skill{ID: d.Name(), Source: SourceEmbeddedUndeclared, Bytes: by, Lines: li})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func shortSHA(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func fileSize(path string) (bytes, lines int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	return len(data), strings.Count(string(data), "\n"), nil
}

// Source es una fuente de instrucción en la tabla única.
type Source struct {
	Name  string
	Kind  string
	Bytes int
	Note  string
}

// Top devuelve las n fuentes más grandes (por tamaño, desempata por nombre).
func Top(srcs []Source, n int) []Source {
	cp := append([]Source(nil), srcs...)
	sortSources(cp)
	if n > len(cp) {
		n = len(cp)
	}
	return cp[:n]
}

func sortSources(s []Source) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].Bytes != s[j].Bytes {
			return s[i].Bytes > s[j].Bytes
		}
		return s[i].Name < s[j].Name
	})
}

// ComposeRow es una línea del golden de Compose (tareas 1.1 y 1.2).
type ComposeRow struct {
	Variant string
	Scope   string
	Bytes   int
	Lines   int
}

// InjectedRow es una línea del golden de fuentes inyectadas (tarea 1.6).
type InjectedRow struct {
	Client       string
	Orchestrator int
	DesignMock   int
	Separator    int
	Total        int
}

func readGoldenRows(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows, sc.Err()
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// LoadCompose lee internal/harness/config/testdata/inventory_compose.golden.
func LoadCompose(root string) ([]ComposeRow, error) {
	rows, err := readGoldenRows(filepath.Join(root, "internal", "harness", "config", "testdata", "inventory_compose.golden"))
	if err != nil {
		return nil, err
	}
	var out []ComposeRow
	for _, r := range rows {
		if len(r) != 4 {
			return nil, fmt.Errorf("inventory: línea de Compose con %d columnas: %q", len(r), r)
		}
		out = append(out, ComposeRow{Variant: r[0], Scope: r[1], Bytes: atoi(r[2]), Lines: atoi(r[3])})
	}
	return out, nil
}

// LoadInjected lee internal/harness/config/testdata/inventory_sources.golden.
func LoadInjected(root string) ([]InjectedRow, error) {
	rows, err := readGoldenRows(filepath.Join(root, "internal", "harness", "config", "testdata", "inventory_sources.golden"))
	if err != nil {
		return nil, err
	}
	var out []InjectedRow
	for _, r := range rows {
		if len(r) != 5 {
			return nil, fmt.Errorf("inventory: línea de fuentes inyectadas con %d columnas: %q", len(r), r)
		}
		out = append(out, InjectedRow{Client: r[0], Orchestrator: atoi(r[1]), DesignMock: atoi(r[2]), Separator: atoi(r[3]), Total: atoi(r[4])})
	}
	return out, nil
}

func clientTitle(c string) string {
	switch c {
	case "claude":
		return "Claude"
	case "codex":
		return "Codex"
	}
	return c
}

// commandsSources mide los archivos de comandos que el instalador escribe.
func commandsSources(root string) ([]Source, error) {
	var out []Source
	base := filepath.Join(root, "assets", "commands")
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(base, p)
		out = append(out, Source{Name: "Comando " + filepath.ToSlash(rel), Kind: "comando", Bytes: len(data)})
		return nil
	})
	return out, err
}

// Build arma la tabla única de fuentes de instrucción (tareas 1.1 a 1.6),
// ordenada por tamaño descendente.
func Build(root string, cat *catalog.Catalog, tp ThirdParty) ([]Source, error) {
	var out []Source
	injected, err := LoadInjected(root)
	if err != nil {
		return nil, err
	}
	for _, r := range injected {
		t := clientTitle(r.Client)
		orchNote := "cinco toggles; sección con marcadores"
		if r.Client == "claude" {
			orchNote = "bloque universal de la entrega en capas; sección con marcadores"
		}
		out = append(out,
			Source{Name: "Bloque sdd-orchestrator inyectado — " + t, Kind: "instrucción inyectada", Bytes: r.Orchestrator,
				Note: orchNote},
			Source{Name: "Sección design-mockups inyectada — " + t, Kind: "instrucción inyectada", Bytes: r.DesignMock,
				Note: "fragmento de 8.865 B más marcadores; solo con el harness design-mockups"},
		)
	}
	skills, err := Skills(root, cat, tp)
	if err != nil {
		return nil, err
	}
	for _, s := range skills {
		if s.Bytes == 0 {
			continue // sin medición: se lista aparte con su motivo
		}
		note := s.Source
		if s.Ref != "" {
			note += " @" + s.Ref
		}
		out = append(out, Source{Name: "Skill " + s.ID, Kind: "skill", Bytes: s.Bytes, Note: note})
	}
	cmds, err := commandsSources(root)
	if err != nil {
		return nil, err
	}
	out = append(out, cmds...)
	sortSources(out)
	return out, nil
}

// Unmeasured lista las skills sin medición, con su motivo.
func Unmeasured(root string, cat *catalog.Catalog, tp ThirdParty) ([]Skill, error) {
	skills, err := Skills(root, cat, tp)
	if err != nil {
		return nil, err
	}
	var out []Skill
	for _, s := range skills {
		if s.Bytes == 0 {
			out = append(out, s)
		}
	}
	return out, nil
}
