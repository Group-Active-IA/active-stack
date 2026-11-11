package store

import "testing"

func TestSetGet(t *testing.T) {
	s := New()
	s.Set("a", "1")
	if got := s.Get("a"); got != "1" {
		t.Errorf("Get(a) = %q, want %q", got, "1")
	}
}
