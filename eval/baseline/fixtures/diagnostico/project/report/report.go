// Package report arma las lineas del informe.
package report

import "strings"

// Wrap reparte las palabras de text en lineas de como maximo width
// caracteres. Una palabra mas larga que width queda sola en su linea.
func Wrap(text string, width int) []string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(text) {
		switch {
		case line == "":
			line = w
		case len(line)+1+len(w) >= width:
			lines = append(lines, line)
			line = w
		default:
			line += " " + w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
