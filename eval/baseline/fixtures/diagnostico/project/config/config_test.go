package config

import "testing"

func TestParseWidth(t *testing.T) {
	if got := Parse("width=5").Width; got != 5 {
		t.Errorf("Parse(width=5).Width = %d, want 5", got)
	}
	if got := Parse("").Width; got != DefaultWidth {
		t.Errorf("Parse(\"\").Width = %d, want %d", got, DefaultWidth)
	}
}
