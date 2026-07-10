package model

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// ─────────────────────────────────────────────────────────────────────────────
// C-26: Starter domain type — field-level validation
// ─────────────────────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────────────────
// C-28: model.MCP shape — fields and YAML round-trip
// ─────────────────────────────────────────────────────────────────────────────

// TestMCPShape_Fields asserts that model.MCP carries Name, Command, Args, Env
// with the expected YAML tags and that a well-formed local MCP round-trips.
func TestMCPShape_Fields(t *testing.T) {
	raw := `
name: context7
command: npx
args:
  - -y
  - "@upstash/context7-mcp"
env:
  DEBUG: "1"
`
	var m MCP
	if err := yaml.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if m.Name != "context7" {
		t.Errorf("Name = %q, want %q", m.Name, "context7")
	}
	if m.Command != "npx" {
		t.Errorf("Command = %q, want %q", m.Command, "npx")
	}
	if len(m.Args) != 2 || m.Args[0] != "-y" || m.Args[1] != "@upstash/context7-mcp" {
		t.Errorf("Args = %v, want [-y @upstash/context7-mcp]", m.Args)
	}
	if m.Env["DEBUG"] != "1" {
		t.Errorf("Env[DEBUG] = %q, want %q", m.Env["DEBUG"], "1")
	}
}

// TestMCPShape_NameFirst asserts that Name is the first field (C-26 back-compat).
// We verify that an MCP with only a name round-trips correctly — the other fields
// default to zero value (empty string, nil slice, nil map).
func TestMCPShape_NameFirst(t *testing.T) {
	m := MCP{Name: "my-server"}
	out, err := yaml.Marshal(m)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	// The only non-zero field should be name.
	if !strings.Contains(string(out), "name: my-server") {
		t.Errorf("marshaled yaml missing 'name: my-server': %s", out)
	}
	// Command, Args, Env are omitempty — should not appear.
	if strings.Contains(string(out), "command") {
		t.Errorf("marshaled yaml should omit empty 'command': %s", out)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// C-28: MCP.Validate()
// ─────────────────────────────────────────────────────────────────────────────

// TestMCPValidate_WellFormed asserts that an MCP with non-empty Name and Command
// passes field-level validation without error.
func TestMCPValidate_WellFormed(t *testing.T) {
	m := MCP{Name: "context7", Command: "npx"}
	if err := m.Validate(); err != nil {
		t.Errorf("expected no error for well-formed MCP, got: %v", err)
	}
}

// TestMCPValidate_EmptyName asserts that an MCP with an empty Name fails
// validation and the error identifies the missing field.
func TestMCPValidate_EmptyName(t *testing.T) {
	m := MCP{Name: "", Command: "npx"}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for MCP with empty Name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention 'name', got: %q", err.Error())
	}
}

// TestMCPValidate_EmptyCommand asserts that an MCP with an empty Command fails
// validation and the error identifies the missing field.
func TestMCPValidate_EmptyCommand(t *testing.T) {
	m := MCP{Name: "context7", Command: ""}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for MCP with empty Command, got nil")
	}
	if !strings.Contains(err.Error(), "command") {
		t.Errorf("expected error to mention 'command', got: %q", err.Error())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// L2 (catalog-localized-descriptions): Starter.Description / LongDescription
// become LocalizedText.
// ─────────────────────────────────────────────────────────────────────────────

// TestStarter_LocalizedDescriptionAndLongDescription_ParseAndResolve asserts
// that a starter entry with description/long_description as {es, en} maps
// parses into LocalizedText fields that resolve per language via Localized.
func TestStarter_LocalizedDescriptionAndLongDescription_ParseAndResolve(t *testing.T) {
	raw := `
id: backend
name: Backend
description:
  es: Stack de backend Python/FastAPI.
  en: Python/FastAPI backend stack.
long_description:
  es: Arquitectura limpia, base de datos, seguridad y contenedores.
  en: Clean architecture, database, security, and containers.
`
	var s Starter
	if err := yaml.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if got := s.Description.Localized("en"); got != "Python/FastAPI backend stack." {
		t.Errorf("Description.Localized(en) = %q, want %q", got, "Python/FastAPI backend stack.")
	}
	if got := s.Description.Localized("es"); got != "Stack de backend Python/FastAPI." {
		t.Errorf("Description.Localized(es) = %q, want %q", got, "Stack de backend Python/FastAPI.")
	}
	if got := s.LongDescription.Localized("en"); got != "Clean architecture, database, security, and containers." {
		t.Errorf("LongDescription.Localized(en) = %q, want %q", got, "Clean architecture, database, security, and containers.")
	}
	if got := s.LongDescription.Localized("es"); got != "Arquitectura limpia, base de datos, seguridad y contenedores." {
		t.Errorf("LongDescription.Localized(es) = %q, want %q", got, "Arquitectura limpia, base de datos, seguridad y contenedores.")
	}
}

// TestStarterValidate_WellFormed asserts that a Starter with a non-empty ID
// and Name passes field-level validation without error.
func TestStarterValidate_WellFormed(t *testing.T) {
	s := Starter{
		ID:   "active-ia",
		Name: "Active IA",
	}
	if err := s.Validate(); err != nil {
		t.Errorf("expected no error for well-formed Starter, got: %v", err)
	}
}

// TestStarterValidate_EmptyID asserts that a Starter with an empty ID fails
// validation and that the error names the missing field.
func TestStarterValidate_EmptyID(t *testing.T) {
	s := Starter{
		ID:   "",
		Name: "Some Starter",
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("expected error for Starter with empty ID, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("expected error to mention 'id', got: %q", err.Error())
	}
}

// TestStarterValidate_EmptyName asserts that a Starter with an empty Name
// fails validation and that the error names the missing field.
func TestStarterValidate_EmptyName(t *testing.T) {
	s := Starter{
		ID:   "active-ia",
		Name: "",
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("expected error for Starter with empty Name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention 'name', got: %q", err.Error())
	}
}
