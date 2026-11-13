package texto

import "strings"

// PadRight completa s con espacios hasta n bytes.
func PadRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
