// Package texto reune utilidades de texto para mensajes de la interfaz.
package texto

// Truncate devuelve como maximo los primeros n caracteres de s.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}
