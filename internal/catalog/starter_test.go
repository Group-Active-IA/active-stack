package catalog

import (
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// C-26: Starter section in the catalog
// ─────────────────────────────────────────────────────────────────────────────

// minimalHarnessYAML is a valid harness-only YAML block reused across starter
// tests that need a populated harness catalog without loading the embedded one.
const minimalHarnessYAML = `
harnesses:
  - id: h-one
    name: H One
    description: { es: Descripción de H One., en: H One description. }
    long_description: { es: Larga de H One., en: H One long description. }
    type: config
    install_modes: [lite]
  - id: h-two
    name: H Two
    description: { es: Descripción de H Two., en: H Two description. }
    long_description: { es: Larga de H Two., en: H Two long description. }
    type: config
    install_modes: [full]
`

// ─── 2.1: parse + index ───────────────────────────────────────────────────────

// TestStarters_EmbeddedCatalogHasSeeds asserts that catalog.Load() parses the
// embedded harnesses.yaml and exposes the three seed starters: active-ia,
// ux-ui, and backend. It also asserts that active-ia.includes == [ux-ui, backend].
func TestStarters_EmbeddedCatalogHasSeeds(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, id := range []string{"active-ia", "ux-ui", "backend"} {
		s, ok := c.StarterByID(id)
		if !ok {
			t.Errorf("expected starter %q in catalog, not found", id)
			continue
		}
		if s.ID != id {
			t.Errorf("starter %q: ID mismatch, got %q", id, s.ID)
		}
	}

	activeIA, ok := c.StarterByID("active-ia")
	if !ok {
		t.Fatal("active-ia starter not found")
	}
	if len(activeIA.Includes) != 2 {
		t.Fatalf("active-ia.includes len = %d, want 2", len(activeIA.Includes))
	}
	wantIncludes := map[string]bool{"ux-ui": true, "backend": true}
	for _, inc := range activeIA.Includes {
		if !wantIncludes[inc] {
			t.Errorf("active-ia.includes: unexpected value %q", inc)
		}
	}
}

// TestStarters_ParseFromYAML asserts that parse() correctly reads a starters:
// section from raw YAML into Catalog.Starters and builds the index.
func TestStarters_ParseFromYAML(t *testing.T) {
	raw := minimalHarnessYAML + `
starters:
  - id: my-starter
    name: My Starter
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]
`
	c, err := parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	if len(c.Starters) != 1 {
		t.Fatalf("Starters len = %d, want 1", len(c.Starters))
	}

	s, ok := c.StarterByID("my-starter")
	if !ok {
		t.Fatal("my-starter not found via StarterByID")
	}
	if s.ID != "my-starter" {
		t.Errorf("starter ID = %q, want %q", s.ID, "my-starter")
	}
	if len(s.Harnesses) != 1 || s.Harnesses[0] != "h-one" {
		t.Errorf("starter.Harnesses = %v, want [h-one]", s.Harnesses)
	}
}

// ─── 3.1: validation — invalid starters ──────────────────────────────────────

// TestStarterValidation_RejectsInvalidCatalogs runs a table of in-memory YAML
// catalogs each containing a different kind of invalid starter and asserts that
// parse() returns an error naming the offending starter/field.
func TestStarterValidation_RejectsInvalidCatalogs(t *testing.T) {
	tests := map[string]struct {
		yaml string
		want string // substring expected in error
	}{
		"empty starter id": {
			yaml: minimalHarnessYAML + `
starters:
  - id: ""
    name: Some Starter
`,
			want: "id",
		},
		"duplicate starter id": {
			yaml: minimalHarnessYAML + `
starters:
  - id: dup
    name: Dup A
  - id: dup
    name: Dup B
`,
			want: "dup",
		},
		"harness reference does not exist": {
			yaml: minimalHarnessYAML + `
starters:
  - id: bad-harness
    name: Bad Harness Ref
    harnesses: [ghost-harness]
`,
			want: "ghost-harness",
		},
		"include reference does not exist": {
			yaml: minimalHarnessYAML + `
starters:
  - id: bad-include
    name: Bad Include Ref
    includes: [ghost-starter]
`,
			want: "ghost-starter",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

// ─── 3.2: cycle detection ─────────────────────────────────────────────────────

// TestStarterValidation_DetectsCycles asserts that self-reference and indirect
// cycles in includes are caught by catalog.Load() / parse().
func TestStarterValidation_DetectsCycles(t *testing.T) {
	tests := map[string]struct {
		yaml string
		want string
	}{
		"self-reference a→a": {
			yaml: minimalHarnessYAML + `
starters:
  - id: self
    name: Self
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [self]
`,
			want: "cycle",
		},
		"indirect cycle a→b→a": {
			yaml: minimalHarnessYAML + `
starters:
  - id: alpha
    name: Alpha
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [beta]
  - id: beta
    name: Beta
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [alpha]
`,
			want: "cycle",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected cycle error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

// ─── 4.1: StarterByID ─────────────────────────────────────────────────────────

// TestStarterByID_Known asserts that StarterByID returns the starter and true
// for a known starter id.
func TestStarterByID_Known(t *testing.T) {
	raw := minimalHarnessYAML + `
starters:
  - id: known
    name: Known Starter
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
`
	c, err := parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	s, ok := c.StarterByID("known")
	if !ok {
		t.Fatal("expected ok=true for known starter, got false")
	}
	if s.ID != "known" {
		t.Errorf("StarterByID returned starter with ID %q, want %q", s.ID, "known")
	}
}

// TestStarterByID_Unknown asserts that StarterByID returns false for an unknown
// starter id.
func TestStarterByID_Unknown(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	_, ok := c.StarterByID("does-not-exist")
	if ok {
		t.Fatal("expected ok=false for unknown starter, got true")
	}
}

// ─── 4.2: ResolveStarter ─────────────────────────────────────────────────────

// TestResolveStarter_CompositExpands asserts that ResolveStarter for a
// composite starter returns the direct harnesses plus all transitively included
// harnesses, without duplicates, in a deterministic order.
func TestResolveStarter_CompositeExpands(t *testing.T) {
	raw := `
harnesses:
  - id: h-ux
    name: H UX
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [full]
  - id: h-be
    name: H BE
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [full]
  - id: h-shared
    name: H Shared
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [full]
starters:
  - id: ux-ui
    name: UX/UI
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-ux, h-shared]
  - id: backend
    name: Backend
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-be, h-shared]
  - id: full-stack
    name: Full Stack
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [ux-ui, backend]
`
	c, err := parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	harnesses, err := c.ResolveStarter("full-stack")
	if err != nil {
		t.Fatalf("ResolveStarter() error = %v", err)
	}

	// Should include h-ux, h-be, h-shared — but h-shared only ONCE.
	ids := make(map[string]int)
	for _, h := range harnesses {
		ids[h.ID]++
	}

	for _, want := range []string{"h-ux", "h-be", "h-shared"} {
		if ids[want] == 0 {
			t.Errorf("ResolveStarter: missing harness %q in result", want)
		}
	}
	if ids["h-shared"] > 1 {
		t.Errorf("h-shared appears %d times, want exactly 1 (deduplication)", ids["h-shared"])
	}
}

// TestResolveStarter_SharedHarnessDeduped asserts explicitly that a harness
// referenced by two included starters appears exactly once in the resolved set.
func TestResolveStarter_SharedHarnessDeduped(t *testing.T) {
	raw := `
harnesses:
  - id: common
    name: Common
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [full]
starters:
  - id: a
    name: A
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [common]
  - id: b
    name: B
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [common]
  - id: parent
    name: Parent
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [a, b]
`
	c, err := parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	harnesses, err := c.ResolveStarter("parent")
	if err != nil {
		t.Fatalf("ResolveStarter() error = %v", err)
	}

	count := 0
	for _, h := range harnesses {
		if h.ID == "common" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("'common' harness appears %d times, want exactly 1", count)
	}
}

// TestResolveStarter_UnknownErrors asserts that ResolveStarter returns an error
// (not a partial result) when called with an unknown starter id.
func TestResolveStarter_UnknownErrors(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	_, err = c.ResolveStarter("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown starter id, got nil")
	}
}
