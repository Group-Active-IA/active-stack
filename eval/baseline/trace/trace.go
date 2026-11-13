// Package trace registra las lecturas de archivos y las llamadas a
// herramientas de un intento por tipo y rol (tarea 2.4 de opsx-baseline). Una
// interfaz que no expone la senal la marca no disponible: no es cero.
package trace

import (
	"errors"
	"strings"
)

// Signal indica si la interfaz expone una senal.
type Signal bool

// Counts es el conteo de una senal. Available false significa que la interfaz
// no la expone y Reason dice por que; Total 0 con Available true es un cero
// real.
type Counts struct {
	Available  bool
	Reason     string
	Total      int
	ByType     map[string]int
	ByRole     map[string]int
	ByTypeRole map[string]map[string]int // tipo -> rol -> n
	Unique     int                       // rutas distintas (solo lecturas)
}

// Summary es el conteo de un intento.
type Summary struct {
	Reads     Counts
	ToolCalls Counts
}

// Log acumula los eventos de un intento.
type Log struct {
	reads, tools       Signal
	readReason, toolsR string
	r, c               counter
}

type counter struct {
	total  int
	byType map[string]int
	byRole map[string]int
	byTR   map[string]map[string]int
	paths  map[string]bool
}

func newCounter() counter {
	return counter{byType: map[string]int{}, byRole: map[string]int{}, byTR: map[string]map[string]int{}, paths: map[string]bool{}}
}

func (c *counter) add(role, typ string) {
	c.total++
	c.byType[typ]++
	c.byRole[role]++
	if c.byTR[typ] == nil {
		c.byTR[typ] = map[string]int{}
	}
	c.byTR[typ][role]++
}

func (c counter) counts(avail Signal, reason string) Counts {
	if !avail {
		if strings.TrimSpace(reason) == "" {
			reason = "la interfaz no expone esta señal"
		}
		return Counts{Reason: reason}
	}
	out := Counts{Available: true, Total: c.total, Unique: len(c.paths),
		ByType: map[string]int{}, ByRole: map[string]int{}, ByTypeRole: map[string]map[string]int{}}
	for k, v := range c.byType {
		out.ByType[k] = v
	}
	for k, v := range c.byRole {
		out.ByRole[k] = v
	}
	for t, m := range c.byTR {
		out.ByTypeRole[t] = map[string]int{}
		for r, n := range m {
			out.ByTypeRole[t][r] = n
		}
	}
	return out
}

// NewLog crea un registro. reads y tools indican si la interfaz expone cada
// senal; reason explica la ausencia cuando alguna es false.
func NewLog(reads, tools Signal, reason string) *Log {
	return &Log{reads: reads, tools: tools, readReason: reason, toolsR: reason, r: newCounter(), c: newCounter()}
}

func blank(ss ...string) bool {
	for _, s := range ss {
		if strings.TrimSpace(s) == "" {
			return true
		}
	}
	return false
}

// RecordRead registra una lectura de archivo hecha por role con la
// herramienta typ sobre path. Una lectura no cuenta como llamada a
// herramienta: el adaptador registra ambas si la interfaz las expone por
// separado.
func (l *Log) RecordRead(role, typ, path string) error {
	if !l.reads {
		return errors.New("trace: la interfaz no expone lecturas y se registró una")
	}
	if blank(role, typ, path) {
		return errors.New("trace: lectura con rol, tipo o ruta vacíos")
	}
	l.r.add(role, typ)
	l.r.paths[path] = true
	return nil
}

// RecordToolCall registra una llamada a la herramienta typ hecha por role.
func (l *Log) RecordToolCall(role, typ string) error {
	if !l.tools {
		return errors.New("trace: la interfaz no expone llamadas a herramientas y se registró una")
	}
	if blank(role, typ) {
		return errors.New("trace: llamada con rol o tipo vacíos")
	}
	l.c.add(role, typ)
	return nil
}

// Summary devuelve los conteos (copias: modificarlos no altera el registro).
func (l *Log) Summary() Summary {
	return Summary{Reads: l.r.counts(l.reads, l.readReason), ToolCalls: l.c.counts(l.tools, l.toolsR)}
}
