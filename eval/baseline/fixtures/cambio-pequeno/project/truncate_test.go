package texto

import "testing"

func TestTruncateASCII(t *testing.T) {
	cases := []struct {
		s    string
		n    int
		want string
	}{
		{"hola", 10, "hola"},
		{"hola mundo", 4, "hola"},
		{"hola", 0, ""},
		{"", 3, ""},
	}
	for _, c := range cases {
		if got := Truncate(c.s, c.n); got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.s, c.n, got, c.want)
		}
	}
}
