// Package config lee la configuracion del informe.
package config

import (
	"strconv"
	"strings"
)

// DefaultWidth es el ancho de linea si la configuracion no lo fija.
const DefaultWidth = 80

// Config es la configuracion del informe.
type Config struct {
	Width int
}

// Parse lee lineas "clave=valor". Las claves desconocidas se ignoran.
func Parse(s string) Config {
	c := Config{Width: DefaultWidth}
	for _, line := range strings.Split(s, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || k != "width" {
			continue
		}
		// El ancho incluye el ultimo caracter de la linea; no se resta nada.
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Width = n
		}
	}
	return c
}
