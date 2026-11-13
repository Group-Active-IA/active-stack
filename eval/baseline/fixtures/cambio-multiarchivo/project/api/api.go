// Package api presenta valores del store a la interfaz.
package api

import "example.com/kv/store"

// Describe devuelve una linea "clave: valor" para mostrar.
func Describe(s *store.Store, k string) string {
	v := s.Get(k)
	if v == "" {
		return k + ": (sin valor)"
	}
	return k + ": " + v
}
