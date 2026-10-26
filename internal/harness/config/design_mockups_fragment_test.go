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
		{"foundation before change design", "propose → **design system** → **design** → apply"},
		{"full flow with verify", "apply → verify → archive"},
		{"mockups location", "openspec/changes/<name>/mockups/"},
		{"visual changes trigger design automatically", "Run the foundation and mockup gates automatically"},
		{"user can explicitly skip", "explicit request from the user"},
		{"decision record section", "## UI Mockups"},
		{"approval record", "Approval: approved"},
		{"declared gate exception", "never block on phase gates"},
		{"project design system", "design-system/README.md"},
		{"portable html review", "local responsive HTML/CSS"},
		{"render before approval", "Render each mockup"},
		{"captures are presented", "present images"},
		{"preview location", "mockups/previews/"},
		{"conditional design skills", "emil-design-eng"},
		{"soft dependency on design skills", "soft dependencies"},
		// visual-verify: step 7 (spec: design-mockups-harness).
		{"visual verify step", "Visual verify"},
		{"playwright tools", "Playwright"},
		{"dev server", "dev server"},
		{"mobile capture", "mobile"},
		{"desktop capture", "desktop"},
		{"console errors", "console"},
		{"failed requests", "failed network requests"},
		{"captures location", "mockups/verify/"},
		{"gated on approved UI mockups", "approved mockup"},
		{"gated on decision", "Decision: yes | provided"},
		{"priority levels", "blocker"},
		{"reproduction steps", "reproduction steps"},
		{"honesty rule", "Never claim"},
		{"explicit degradation", "NOT automated"},
		{"manual checklist fallback", "manual checklist"},
		{"suggests playwright harness", "`playwright` harness"},
		{"soft dependency on playwright", "design-mockups still works"},
		{"outcome record values", "passed | differences | manual"},
	}
	normalized := strings.Join(strings.Fields(got), " ")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(normalized, tc.want) {
				t.Errorf("fragment missing %q", tc.want)
			}
		})
	}
}

// TestDesignMockupsFragment_DoesNotHardcodeTypeURL guards against pinning a
// per-account or remote design service URL into the portable fragment.
func TestDesignMockupsFragment_DoesNotHardcodeTypeURL(t *testing.T) {
	got, err := loadFragment("design-mockups.md")
	if err != nil {
		t.Fatalf("loadFragment: %v", err)
	}
	if strings.Contains(got, "claude.ai/") || strings.Contains(got, "https://") {
		t.Error("fragment must not hardcode remote design service URLs")
	}
}
