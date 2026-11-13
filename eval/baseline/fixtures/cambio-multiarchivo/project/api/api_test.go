package api

import (
	"testing"

	"example.com/kv/store"
)

// Llamadas existentes: su comportamiento no debe cambiar.
func TestDescribeExisting(t *testing.T) {
	s := store.New()
	s.Set("a", "1")
	if got := Describe(s, "a"); got != "a: 1" {
		t.Errorf("Describe(a) = %q, want %q", got, "a: 1")
	}
	if got := Describe(s, "zz"); got != "zz: (sin valor)" {
		t.Errorf("Describe(zz) = %q, want %q", got, "zz: (sin valor)")
	}
}
