// Package redact es el redactor determinista del arnés (design.md D8 y D14).
// Aplica la lista única de patrones de valor sensibles del catálogo, línea por
// línea, y reemplaza el valor por una marca de omisión conservando la línea en
// su posición. Vive fuera del binario (CHANGES.md:49).
package redact

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

const markPrefix = "[REDACTED:"

// Mark es la marca de omisión de un patrón.
func Mark(patternID string) string { return markPrefix + patternID + "]" }

type linePattern struct {
	id       string
	re       *regexp.Regexp
	secretIx []int // índices de los grupos `secret*`
}

type blockPattern struct {
	id         string
	begin, end *regexp.Regexp
}

// Redactor aplica los patrones del catálogo.
type Redactor struct {
	lines      []linePattern
	blocks     []blockPattern
	categoryOf map[string]string
}

// New compila los patrones del catálogo.
func New(patterns []catalog.SensitivePattern) (*Redactor, error) {
	r := &Redactor{categoryOf: map[string]string{}}
	for _, p := range patterns {
		r.categoryOf[p.ID] = p.Category
		switch p.Kind {
		case "line":
			re, err := regexp.Compile(p.Regex)
			if err != nil {
				return nil, fmt.Errorf("redact: patrón %q: %w", p.ID, err)
			}
			lp := linePattern{id: p.ID, re: re}
			for i, name := range re.SubexpNames() {
				if strings.HasPrefix(name, "secret") {
					lp.secretIx = append(lp.secretIx, i)
				}
			}
			r.lines = append(r.lines, lp)
		case "block":
			b, err := regexp.Compile(p.Begin)
			if err != nil {
				return nil, fmt.Errorf("redact: patrón %q (begin): %w", p.ID, err)
			}
			e, err := regexp.Compile(p.End)
			if err != nil {
				return nil, fmt.Errorf("redact: patrón %q (end): %w", p.ID, err)
			}
			r.blocks = append(r.blocks, blockPattern{id: p.ID, begin: b, end: e})
		default:
			return nil, fmt.Errorf("redact: patrón %q con tipo desconocido %q", p.ID, p.Kind)
		}
	}
	return r, nil
}

// ByCategory agrupa un conteo por id de patrón en un conteo por categoría.
func (r *Redactor) ByCategory(byID map[string]int) map[string]int {
	out := map[string]int{}
	for id, n := range byID {
		out[r.categoryOf[id]] += n
	}
	return out
}

// RedactText redacta un texto completo. Devuelve el texto y la cantidad de
// redacciones por id de patrón. La cantidad de líneas no cambia.
func (r *Redactor) RedactText(text string) (string, map[string]int) {
	counts := map[string]int{}
	lines := strings.Split(text, "\n")
	var inBlock *blockPattern
	for i, line := range lines {
		if inBlock != nil {
			if inBlock.end.MatchString(line) {
				inBlock = nil
				continue
			}
			lines[i] = Mark(inBlock.id)
			counts[blockID(r, line, inBlock)]++
			continue
		}
		if b := r.beginsBlock(line); b != nil {
			inBlock = b
			continue
		}
		lines[i] = r.redactLine(line, counts)
	}
	return strings.Join(lines, "\n"), counts
}

func blockID(_ *Redactor, _ string, b *blockPattern) string { return b.id }

func (r *Redactor) beginsBlock(line string) *blockPattern {
	for i := range r.blocks {
		if r.blocks[i].begin.MatchString(line) {
			return &r.blocks[i]
		}
	}
	return nil
}

type span struct{ start, end int }

func (r *Redactor) redactLine(line string, counts map[string]int) string {
	for _, p := range r.lines {
		matches := p.re.FindAllStringSubmatchIndex(line, -1)
		if len(matches) == 0 {
			continue
		}
		var spans []span
		for _, m := range matches {
			found := false
			for _, ix := range p.secretIx {
				if s, e := m[2*ix], m[2*ix+1]; s >= 0 && e > s {
					spans = append(spans, span{s, e})
					found = true
				}
			}
			if !found && len(p.secretIx) == 0 {
				spans = append(spans, span{m[0], m[1]})
			}
		}
		// De derecha a izquierda para no correr los índices; lo ya redactado se saltea.
		for k := len(spans) - 1; k >= 0; k-- {
			s := spans[k]
			if strings.HasPrefix(line[s.start:s.end], markPrefix) {
				continue
			}
			line = line[:s.start] + Mark(p.id) + line[s.end:]
			counts[p.id]++
		}
	}
	return line
}
