package model

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// ─────────────────────────────────────────────────────────────────────────────
// C-32: harness-scope-model — ScopeKind / IsStarterOnly()
// ─────────────────────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────────────────
// External.MCP field — round-trip and validation
// ─────────────────────────────────────────────────────────────────────────────

// TestExternalMCP_YAMLRoundTrip asserts that an External struct with a nested
// MCP field correctly round-trips through YAML marshal/unmarshal.
// The MCP field uses model.MCP (Name, Command, Args) and the yaml tag "mcp".
func TestExternalMCP_YAMLRoundTrip(t *testing.T) {
	raw := `
method: homebrew
pkg: gentleman-programming/tap/engram
repo: Gentleman-Programming/engram
mcp:
  name: engram
  command: engram
  args:
    - mcp
`
	var ext External
	if err := yaml.Unmarshal([]byte(raw), &ext); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if ext.MCP == nil {
		t.Fatal("External.MCP is nil after unmarshal; field may be missing or mis-tagged")
	}
	if ext.MCP.Name != "engram" {
		t.Errorf("MCP.Name = %q, want %q", ext.MCP.Name, "engram")
	}
	if ext.MCP.Command != "engram" {
		t.Errorf("MCP.Command = %q, want %q", ext.MCP.Command, "engram")
	}
	if len(ext.MCP.Args) != 1 || ext.MCP.Args[0] != "mcp" {
		t.Errorf("MCP.Args = %v, want [mcp]", ext.MCP.Args)
	}
}

// TestExternalMCP_OmitEmptyWhenNil asserts that when External.MCP is nil the
// "mcp" key is absent from the marshaled YAML (omitempty behaviour).
func TestExternalMCP_OmitEmptyWhenNil(t *testing.T) {
	ext := External{Method: "homebrew", Pkg: "gentleman-programming/tap/engram"}
	out, err := yaml.Marshal(ext)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	if contains(string(out), "mcp:") {
		t.Errorf("marshaled YAML should not contain 'mcp:' when MCP is nil, got:\n%s", out)
	}
}

// TestExternalMCP_ValidateReused asserts that the existing MCP.Validate() is
// reused: an External.MCP with a missing Command fails validation via Validate().
func TestExternalMCP_ValidateReused(t *testing.T) {
	ext := External{
		Method: "homebrew",
		Pkg:    "gentleman-programming/tap/engram",
		MCP:    &MCP{Name: "engram", Command: ""},
	}
	if err := ext.MCP.Validate(); err == nil {
		t.Error("expected validation error for MCP with empty Command, got nil")
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

// ─────────────────────────────────────────────────────────────────────────────
// L2 (catalog-localized-descriptions): Harness.Description / LongDescription
// become LocalizedText.
// ─────────────────────────────────────────────────────────────────────────────

// TestHarness_LocalizedDescriptionAndLongDescription_ParseAndResolve asserts
// that a harness entry with description/long_description as {es, en} maps
// parses into LocalizedText fields that resolve per language via Localized.
func TestHarness_LocalizedDescriptionAndLongDescription_ParseAndResolve(t *testing.T) {
	raw := `
id: openspec
name: OpenSpec CLI
description:
  es: CLI de Spec-Driven Development.
  en: Spec-Driven Development CLI.
long_description:
  es: Fuente de verdad del estado del proyecto.
  en: Single source of truth for project state.
type: external
install_modes: [lite, full]
`
	var h Harness
	if err := yaml.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if got := h.Description.Localized("en"); got != "Spec-Driven Development CLI." {
		t.Errorf("Description.Localized(en) = %q, want %q", got, "Spec-Driven Development CLI.")
	}
	if got := h.Description.Localized("es"); got != "CLI de Spec-Driven Development." {
		t.Errorf("Description.Localized(es) = %q, want %q", got, "CLI de Spec-Driven Development.")
	}
	if got := h.LongDescription.Localized("en"); got != "Single source of truth for project state." {
		t.Errorf("LongDescription.Localized(en) = %q, want %q", got, "Single source of truth for project state.")
	}
	if got := h.LongDescription.Localized("es"); got != "Fuente de verdad del estado del proyecto." {
		t.Errorf("LongDescription.Localized(es) = %q, want %q", got, "Fuente de verdad del estado del proyecto.")
	}
}

// TestModel_ScopeKind_ZeroValueIsGlobal asserts that:
//  1. A Harness with no Scope set (zero value) returns IsStarterOnly()==false.
//  2. A Harness with Scope==ScopeStarterOnly returns IsStarterOnly()==true.
//
// RED: fails to compile/run against current code — type ScopeKind and helper
// IsStarterOnly() do not exist yet.
func TestModel_ScopeKind_ZeroValueIsGlobal(t *testing.T) {
	// Case 1: zero value (field omitted) → global, not starter-only.
	zero := Harness{}
	if zero.IsStarterOnly() {
		t.Error("Harness{} (zero Scope): IsStarterOnly() = true, want false")
	}

	// Case 2: explicit ScopeGlobal → also not starter-only.
	explicit := Harness{Scope: ScopeGlobal}
	if explicit.IsStarterOnly() {
		t.Errorf("Harness{Scope: ScopeGlobal}: IsStarterOnly() = true, want false")
	}

	// Case 3 (triangulation): ScopeStarterOnly → must be starter-only.
	starterOnly := Harness{Scope: ScopeStarterOnly}
	if !starterOnly.IsStarterOnly() {
		t.Error("Harness{Scope: ScopeStarterOnly}: IsStarterOnly() = false, want true")
	}
}
