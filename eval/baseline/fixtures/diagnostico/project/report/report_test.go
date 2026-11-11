package report

import (
	"reflect"
	"testing"

	"example.com/informe/config"
)

func TestWrapShortWidth(t *testing.T) {
	got := Wrap("aa bb cc", 4)
	want := []string{"aa", "bb", "cc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
}

// Falla reportada: con el ancho de la configuracion, una linea que entra
// justo se corta antes de tiempo.
func TestWrapExactWidth(t *testing.T) {
	width := config.Parse("width=5").Width
	got := Wrap("aa bb", width)
	want := []string{"aa bb"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Wrap(%q, %d) = %q, want %q", "aa bb", width, got, want)
	}
}
