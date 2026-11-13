// Package store guarda pares clave-valor en memoria.
package store

// Store es un almacen clave-valor en memoria.
type Store struct {
	m map[string]string
}

// New devuelve un Store vacio.
func New() *Store {
	return &Store{m: map[string]string{}}
}

// Set guarda v bajo la clave k.
func (s *Store) Set(k, v string) {
	s.m[k] = v
}

// Get devuelve el valor de k, o "" si k no existe.
func (s *Store) Get(k string) string {
	return s.m[k]
}
