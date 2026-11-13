package shadow

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

// UserConfigFiles son los archivos de configuración de usuario que D5 rechaza como clase de
// archivo (no son un patrón de valor, así que no están en D14).
var UserConfigFiles = []string{"settings.json", "config.toml"}

// Hit es un motivo de cancelación encontrado en un archivo del paquete.
type Hit struct {
	Path string
	Rule string // id de patrón de D14, "cubre:<id>" o "config-de-usuario"
}

// ScanResult es el resultado del examen previo. Send es falso si hubo al menos un positivo: un
// positivo cancela el envío del caso.
type ScanResult struct {
	Send bool
	Hits []Hit
}

// Scanner examina un paquete con la lista única de patrones de 05 (D14), leída del catálogo y
// aplicada con el redactor de 05: 10 no tiene una lista propia.
type Scanner struct {
	r      *redact.Redactor
	covers []cover
}

type cover struct{ id, glob string }

// NewScanner compila los patrones de D14. Sin patrones no hay examen.
func NewScanner(patterns []catalog.SensitivePattern) (*Scanner, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("examen previo sin patrones de D14")
	}
	r, err := redact.New(patterns)
	if err != nil {
		return nil, err
	}
	s := &Scanner{r: r}
	for _, p := range patterns {
		for _, g := range p.Covers {
			s.covers = append(s.covers, cover{id: p.ID, glob: g})
		}
	}
	return s, nil
}

// Scan examina todos los archivos del paquete, diffs públicos incluidos: el contenido con los
// patrones de valor de D14, el nombre con los archivos que D14 cubre (`.env`, `*.pem`) y con los
// archivos de configuración de usuario de D5.
func (s *Scanner) Scan(p Package) ScanResult {
	var hits []Hit
	for _, f := range p.Files {
		base := path.Base(f.Path)
		for _, name := range UserConfigFiles {
			if base == name {
				hits = append(hits, Hit{Path: f.Path, Rule: "config-de-usuario"})
			}
		}
		for _, c := range s.covers {
			if ok, _ := path.Match(c.glob, base); ok {
				hits = append(hits, Hit{Path: f.Path, Rule: "cubre:" + c.id})
			}
		}
		found := map[string]bool{}
		for _, text := range views(f) {
			_, counts := s.r.RedactText(text)
			for id, n := range counts {
				if n > 0 {
					found[id] = true
				}
			}
		}
		ids := make([]string, 0, len(found))
		for id := range found {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			hits = append(hits, Hit{Path: f.Path, Rule: id})
		}
	}
	return ScanResult{Send: len(hits) == 0, Hits: hits}
}

// views devuelve los textos a examinar de un archivo. En un diff, además del texto tal cual, cada
// línea de contenido sin su marca (+, - o espacio): los patrones de D14 anclados al inicio de
// línea, como la asignación de `.env`, no ven "+SECRETO=valor". Los encabezados +++ y --- quedan
// como están. No cambia D14: normaliza el formato del diff antes de aplicarla.
func views(f File) []string {
	text := string(f.Data)
	if !isDiff(f) {
		return []string{text}
	}
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "+++") || strings.HasPrefix(l, "---") {
			continue
		}
		if l != "" && (l[0] == '+' || l[0] == '-' || l[0] == ' ') {
			lines[i] = l[1:]
		}
	}
	return []string{text, strings.Join(lines, "\n")}
}

func isDiff(f File) bool {
	switch path.Ext(f.Path) {
	case ".patch", ".diff":
		return true
	}
	return strings.HasPrefix(string(f.Data), "diff --git ") || strings.Contains(string(f.Data), "\ndiff --git ")
}
