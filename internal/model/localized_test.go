package model

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// ─────────────────────────────────────────────────────────────────────────────
// L2 (catalog-localized-descriptions): LocalizedText — fallback resolution
// ─────────────────────────────────────────────────────────────────────────────

// TestLocalizedText_Localized_RequestedLanguagePresent asserts that Localized
// returns the value for the requested language when present.
func TestLocalizedText_Localized_RequestedLanguagePresent(t *testing.T) {
	lt := LocalizedText{"es": "Hola", "en": "Hello"}
	if got := lt.Localized("es"); got != "Hola" {
		t.Errorf("Localized(\"es\") = %q, want %q", got, "Hola")
	}
	if got := lt.Localized("en"); got != "Hello" {
		t.Errorf("Localized(\"en\") = %q, want %q", got, "Hello")
	}
}

// TestLocalizedText_Localized_FallsBackToEnThenEs triangulates the fallback
// chain: requested language missing falls back to "en", and if "en" is also
// missing falls back to "es".
func TestLocalizedText_Localized_FallsBackToEnThenEs(t *testing.T) {
	both := LocalizedText{"es": "Hola", "en": "Hello"}
	if got := both.Localized("fr"); got != "Hello" {
		t.Errorf("Localized(\"fr\") on {es,en} = %q, want %q (en fallback)", got, "Hello")
	}

	onlyES := LocalizedText{"es": "Hola"}
	if got := onlyES.Localized("fr"); got != "Hola" {
		t.Errorf("Localized(\"fr\") on {es} = %q, want %q (es fallback)", got, "Hola")
	}
}

// TestLocalizedText_Localized_NilOrEmptyMapReturnsEmptyString asserts that a
// nil or empty LocalizedText never panics and returns "".
func TestLocalizedText_Localized_NilOrEmptyMapReturnsEmptyString(t *testing.T) {
	var nilMap LocalizedText
	if got := nilMap.Localized("en"); got != "" {
		t.Errorf("Localized(\"en\") on nil map = %q, want \"\"", got)
	}

	empty := LocalizedText{}
	if got := empty.Localized("en"); got != "" {
		t.Errorf("Localized(\"en\") on empty map = %q, want \"\"", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// L2: LocalizedText.Validate — bilingual completeness
// ─────────────────────────────────────────────────────────────────────────────

// TestLocalizedText_Validate_BothLanguagesPresent asserts that Validate
// returns nil when both "es" and "en" are non-empty.
func TestLocalizedText_Validate_BothLanguagesPresent(t *testing.T) {
	lt := LocalizedText{"es": "Hola", "en": "Hello"}
	if err := lt.Validate("description", `harness "openspec"`); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestLocalizedText_Validate_MissingEnglishNamesTheGap asserts that Validate
// fails naming the owner, field, and missing "en" when only "es" is present.
func TestLocalizedText_Validate_MissingEnglishNamesTheGap(t *testing.T) {
	lt := LocalizedText{"es": "Hola"}
	err := lt.Validate("description", `harness "openspec"`)
	if err == nil {
		t.Fatal("expected error for missing en, got nil")
	}
	for _, want := range []string{`harness "openspec"`, "description", `"en"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
}

// TestLocalizedText_Validate_MissingSpanishNamesTheGap triangulates the
// opposite gap: only "en" present fails naming "es".
func TestLocalizedText_Validate_MissingSpanishNamesTheGap(t *testing.T) {
	lt := LocalizedText{"en": "Hello"}
	err := lt.Validate("long_description", `starter "backend"`)
	if err == nil {
		t.Fatal("expected error for missing es, got nil")
	}
	for _, want := range []string{`starter "backend"`, "long_description", `"es"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
}

// TestLocalizedText_Validate_EmptyMapFails asserts that an empty/nil map
// fails validation (missing both languages).
func TestLocalizedText_Validate_EmptyMapFails(t *testing.T) {
	var lt LocalizedText
	if err := lt.Validate("description", `harness "x"`); err == nil {
		t.Fatal("expected error for empty LocalizedText, got nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// L2: LocalizedText.UnmarshalYAML — scalar and mapping forms
// ─────────────────────────────────────────────────────────────────────────────

// TestLocalizedText_UnmarshalYAML_ScalarMapsToES asserts that a bare scalar
// string is assigned to the "es" key only (parse-time retro-compat).
func TestLocalizedText_UnmarshalYAML_ScalarMapsToES(t *testing.T) {
	raw := `description: Configura el orquestador SDD.`
	var holder struct {
		Description LocalizedText `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(raw), &holder); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if got := holder.Description["es"]; got != "Configura el orquestador SDD." {
		t.Errorf("Description[es] = %q, want %q", got, "Configura el orquestador SDD.")
	}
	if _, ok := holder.Description["en"]; ok {
		t.Errorf("Description should not have an en key from a scalar, got: %v", holder.Description)
	}
}

// TestLocalizedText_UnmarshalYAML_MappingDecodesEachKey asserts that a
// {es:, en:} mapping decodes both keys into the map.
func TestLocalizedText_UnmarshalYAML_MappingDecodesEachKey(t *testing.T) {
	raw := `
description:
  es: Configura el orquestador SDD.
  en: Configures the SDD orchestrator.
`
	var holder struct {
		Description LocalizedText `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(raw), &holder); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if got := holder.Description["es"]; got != "Configura el orquestador SDD." {
		t.Errorf("Description[es] = %q, want %q", got, "Configura el orquestador SDD.")
	}
	if got := holder.Description["en"]; got != "Configures the SDD orchestrator." {
		t.Errorf("Description[en] = %q, want %q", got, "Configures the SDD orchestrator.")
	}
}
