package config

import (
	"strings"
	"testing"
)

// TestDesignMockupsFragment_Content pins the functional contract of the
// design-mockups fragment (spec: design-mockups-harness). Each case is a
// distinct instruction the orchestrator must receive; deleting that line from
// the asset makes exactly that case fail.
func TestDesignMockupsFragment_Content(t *testing.T) {
	got, err := loadFragment("design-mockups.md")
	if err != nil {
		t.Fatalf("loadFragment(design-mockups.md): %v", err)
	}

	cases := []struct {
		name string
		want string
	}{
		{"heading", "## Design Stage — UI Mockups"},
		{"stage between propose and apply", "propose → **design** → apply"},
		{"full flow with verify", "apply → verify → archive"},
		{"mockups location", "openspec/changes/<name>/mockups/"},
		{"three answers", "(yes / no / I already have them)"},
		{"no UI signal is not asked", "No UI signal"},
		{"decision record section", "## UI Mockups"},
		{"approval record", "Approval: approved"},
		{"declared gate exception", "never block on phase gates"},
		{"artifact quickstart", "action \"quickstart\""},
		{"design intent", "intent \"design\""},
		{"type comes from quickstart result", "type_url"},
		{"static html fallback", "static HTML"},
		{"model for design phase", "design → sonnet"},
		{"conditional design skills", "emil-design-eng"},
		{"soft dependency on design skills", "if installed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(got, tc.want) {
				t.Errorf("fragment missing %q", tc.want)
			}
		})
	}
}

// TestDesignMockupsFragment_DoesNotHardcodeTypeURL guards against pinning a
// per-account Artifact type URL into the fragment.
func TestDesignMockupsFragment_DoesNotHardcodeTypeURL(t *testing.T) {
	got, err := loadFragment("design-mockups.md")
	if err != nil {
		t.Fatalf("loadFragment: %v", err)
	}
	if strings.Contains(got, "claude.ai/") || strings.Contains(got, "https://") {
		t.Error("fragment must not hardcode URLs; type_url comes from quickstart at runtime")
	}
}
