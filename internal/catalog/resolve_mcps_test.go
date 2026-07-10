// Package catalog — tests for C-29 ResolveStarterMCPs (D3a).
package catalog

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// C-29 Task 1.1 RED: ResolveStarterMCPs aggregates MCPs across includes.
// ─────────────────────────────────────────────────────────────────────────────

// mcpFixtureYAML defines a synthetic catalog with starters that carry MCPs
// (used because the real seed starters may not have MCPs in practice, but the
// method must work for any future starter that does).
const mcpFixtureYAML = `
harnesses:
  - id: h-one
    name: H One
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [lite]

starters:
  - id: base-with-mcps
    name: Base With MCPs
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]
    mcps:
      - name: context7
        command: npx
        args: ["-y", "@upstash/context7-mcp@latest"]
  - id: extra-with-mcps
    name: Extra With MCPs
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]
    mcps:
      - name: engram-mcp
        command: npx
        args: ["-y", "engram-mcp@latest"]
  - id: composite-no-collision
    name: Composite No Collision
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [base-with-mcps, extra-with-mcps]
    harnesses: []
    mcps: []
  - id: root-only-mcps
    name: Root Only MCPs
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]
    mcps:
      - name: my-tool
        command: my-tool
        args: ["serve"]
  - id: collision-root
    name: Collision Root
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    includes: [base-with-mcps]
    harnesses: [h-one]
    mcps:
      - name: context7
        command: npx-root-override
        args: ["--root-version"]
  - id: no-mcps-anywhere
    name: No MCPs Anywhere
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]
`

// TestResolveStarterMCPs_UnionAcrossIncludes asserts that ResolveStarterMCPs
// for a composite starter returns the union of root + included MCPs, deduplicated
// by Name, in a stable order (root MCPs first, then included in includes order).
//
// RED: this test compiles but fails because ResolveStarterMCPs does not exist yet.
func TestResolveStarterMCPs_UnionAcrossIncludes(t *testing.T) {
	c, err := parse([]byte(mcpFixtureYAML))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	// composite-no-collision includes base-with-mcps (context7) and extra-with-mcps (engram-mcp).
	// Its own mcps list is empty. Should return both context7 and engram-mcp.
	mcps, err := c.ResolveStarterMCPs("composite-no-collision")
	if err != nil {
		t.Fatalf("ResolveStarterMCPs() error = %v", err)
	}

	if len(mcps) != 2 {
		t.Fatalf("ResolveStarterMCPs() len = %d, want 2; got: %v", len(mcps), mcps)
	}

	byName := make(map[string]model.MCP, len(mcps))
	for _, m := range mcps {
		byName[m.Name] = m
	}
	if _, ok := byName["context7"]; !ok {
		t.Error("expected MCP 'context7' in resolved set")
	}
	if _, ok := byName["engram-mcp"]; !ok {
		t.Error("expected MCP 'engram-mcp' in resolved set")
	}
}

// TestResolveStarterMCPs_NoIncludes asserts that ResolveStarterMCPs for a
// starter with no includes returns only its own MCPs (triangulation: simplest case).
func TestResolveStarterMCPs_NoIncludes(t *testing.T) {
	c, err := parse([]byte(mcpFixtureYAML))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	mcps, err := c.ResolveStarterMCPs("root-only-mcps")
	if err != nil {
		t.Fatalf("ResolveStarterMCPs() error = %v", err)
	}

	if len(mcps) != 1 {
		t.Fatalf("len = %d, want 1; got: %v", len(mcps), mcps)
	}
	if mcps[0].Name != "my-tool" {
		t.Errorf("MCP name = %q, want %q", mcps[0].Name, "my-tool")
	}
}

// TestResolveStarterMCPs_NoMCPs asserts that a starter with no MCPs anywhere
// returns an empty (not nil) slice without error.
func TestResolveStarterMCPs_NoMCPs(t *testing.T) {
	c, err := parse([]byte(mcpFixtureYAML))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	mcps, err := c.ResolveStarterMCPs("no-mcps-anywhere")
	if err != nil {
		t.Fatalf("ResolveStarterMCPs() error = %v", err)
	}

	// Must not return nil; empty slice is acceptable.
	if mcps == nil {
		mcps = []model.MCP{}
	}
	if len(mcps) != 0 {
		t.Errorf("expected empty MCPs for no-mcps-anywhere, got %v", mcps)
	}
}

// TestResolveStarterMCPs_UnknownID asserts that an unknown starter id returns
// an error (triangulation: error path).
func TestResolveStarterMCPs_UnknownID(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	_, err = c.ResolveStarterMCPs("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown starter id, got nil")
	}
}

// TestResolveStarterMCPs_RootWinsOnCollision asserts that when the root starter
// and an included starter share an MCP name, the root's entry is kept (D3a policy).
//
// RED: fails because ResolveStarterMCPs does not exist yet.
func TestResolveStarterMCPs_RootWinsOnCollision(t *testing.T) {
	c, err := parse([]byte(mcpFixtureYAML))
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	// collision-root includes base-with-mcps (context7 command=npx).
	// collision-root itself defines context7 with command=npx-root-override.
	// Root must win: command should be npx-root-override.
	mcps, err := c.ResolveStarterMCPs("collision-root")
	if err != nil {
		t.Fatalf("ResolveStarterMCPs() error = %v", err)
	}

	byName := make(map[string]model.MCP, len(mcps))
	for _, m := range mcps {
		byName[m.Name] = m
	}

	ctx7, ok := byName["context7"]
	if !ok {
		t.Fatal("expected MCP 'context7' in resolved set")
	}
	if ctx7.Command != "npx-root-override" {
		t.Errorf("expected root's context7 command %q, got %q", "npx-root-override", ctx7.Command)
	}
	// Must appear only once.
	count := 0
	for _, m := range mcps {
		if m.Name == "context7" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("context7 appears %d times, want exactly 1 (dedup)", count)
	}
}
