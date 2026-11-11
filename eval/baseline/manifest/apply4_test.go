package manifest

import "testing"

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		v, min string
		want   bool
	}{
		{"2.1.310", "2.1.310", true},
		{"2.1.311", "2.1.310", true},
		{"2.1.1000", "2.1.310", true},
		{"2.2.0", "2.1.310", true},
		{"3", "2.1.310", true},
		{"2.1.309", "2.1.310", false},
		{"2.1.99", "2.1.310", false},
		{"2.0.999", "2.1.310", false},
		{"2.1", "2.1.310", false},
	}
	for _, tc := range tests {
		got, err := VersionAtLeast(tc.v, tc.min)
		if err != nil || got != tc.want {
			t.Errorf("VersionAtLeast(%q, %q) = (%v, %v), want (%v, nil)", tc.v, tc.min, got, err, tc.want)
		}
	}
	for _, bad := range [][2]string{{"", "2.1.310"}, {"2.1.x", "2.1.310"}, {"2.1.310", ""}, {"2.1.310", "v2"}, {"-1.0", "1.0"}} {
		if _, err := VersionAtLeast(bad[0], bad[1]); err == nil {
			t.Errorf("VersionAtLeast(%q, %q): err = nil, want error de versión no interpretable", bad[0], bad[1])
		}
	}
}
