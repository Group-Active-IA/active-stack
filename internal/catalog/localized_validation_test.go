// Package catalog — L2 (catalog-localized-descriptions) validation fixtures.
//
// These tests exercise catalog.validate() bilingual-completeness rules (D3,
// design.md) via the internal parse([]byte) entry point, independent of the
// real embedded harnesses.yaml.
package catalog

import (
	"strings"
	"testing"
)

// TestCatalogLoad_BilingualCatalog_LoadsClean asserts that a minimal but
// fully bilingual catalog (harness description+long_description, starter
// description+long_description) loads without error.
func TestCatalogLoad_BilingualCatalog_LoadsClean(t *testing.T) {
	valid := `harnesses:
  - id: good-harness
    name: Good Harness
    description: { es: Descripción buena., en: Good description. }
    long_description: { es: Descripción larga buena., en: Good long description. }
    type: config
    install_modes: [lite]
starters:
  - id: good-starter
    name: Good Starter
    description: { es: Descripción de starter., en: Starter description. }
    long_description: { es: Descripción larga de starter., en: Starter long description. }
    harnesses: [good-harness]`

	if _, err := parse([]byte(valid)); err != nil {
		t.Errorf("unexpected error for fully bilingual catalog: %v", err)
	}
}

// TestCatalogLoad_HarnessDescription_OnlyEsFailsNamingEn asserts that a
// harness description with only "es" fails Load, naming the harness, the
// "description" field, and the missing "en".
func TestCatalogLoad_HarnessDescription_OnlyEsFailsNamingEn(t *testing.T) {
	invalid := `harnesses:
  - id: openspec
    name: OpenSpec
    description: { es: Solo español. }
    long_description: { es: Larga solo español., en: Long only English placeholder. }
    type: config
    install_modes: [lite]`

	_, err := parse([]byte(invalid))
	if err == nil {
		t.Fatal("expected error for harness description missing en, got nil")
	}
	for _, want := range []string{`"openspec"`, "description", `"en"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}

// TestCatalogLoad_LegacyScalarDescription_FailsOnMissingEn asserts that a
// bare scalar description (legacy form) parses into "es" only and fails
// Load on the missing "en" — the migration's own safety net.
func TestCatalogLoad_LegacyScalarDescription_FailsOnMissingEn(t *testing.T) {
	legacy := `harnesses:
  - id: openspec
    name: OpenSpec
    description: CLI de Spec-Driven Development.
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [lite]`

	_, err := parse([]byte(legacy))
	if err == nil {
		t.Fatal("expected error for legacy scalar description (missing en), got nil")
	}
	if !strings.Contains(err.Error(), `"en"`) {
		t.Errorf("error %q does not name the missing language \"en\"", err.Error())
	}
}

// TestCatalogLoad_GlobalPickerHarness_MissingLongDescription_Fails asserts
// that a non-starter-only harness (the default scope) omitting
// long_description fails Load, naming the harness and the long_description
// field.
func TestCatalogLoad_GlobalPickerHarness_MissingLongDescription_Fails(t *testing.T) {
	invalid := `harnesses:
  - id: global-picker
    name: Global Picker
    description: { es: Descripción., en: Description. }
    type: config
    install_modes: [lite]`

	_, err := parse([]byte(invalid))
	if err == nil {
		t.Fatal("expected error for global-picker harness missing long_description, got nil")
	}
	for _, want := range []string{`"global-picker"`, "long_description"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}

// TestCatalogLoad_StarterOnlyHarness_NoLongDescription_LoadsClean asserts
// that a starter-only harness with a valid bilingual description and no
// long_description at all loads successfully (long_description is optional
// for starter-only harnesses).
func TestCatalogLoad_StarterOnlyHarness_NoLongDescription_LoadsClean(t *testing.T) {
	valid := `harnesses:
  - id: starter-only-skill
    name: Starter Only Skill
    description: { es: Descripción., en: Description. }
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [full]
starters:
  - id: holder
    name: Holder
    description: { es: Descripción de starter., en: Starter description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [starter-only-skill]`

	if _, err := parse([]byte(valid)); err != nil {
		t.Errorf("unexpected error for starter-only harness without long_description: %v", err)
	}
}

// TestCatalogLoad_StarterOnlyHarness_OneLangLongDescription_Fails
// triangulates: a starter-only harness that DOES provide a long_description
// but only in one language must still fail (optional-but-both-if-present).
func TestCatalogLoad_StarterOnlyHarness_OneLangLongDescription_Fails(t *testing.T) {
	invalid := `harnesses:
  - id: starter-only-skill
    name: Starter Only Skill
    description: { es: Descripción., en: Description. }
    long_description: { es: Solo español. }
    type: skill
    scope: starter-only
    source: { repo: some/repo, method: clone }
    install_modes: [full]
starters:
  - id: holder
    name: Holder
    description: { es: Descripción de starter., en: Starter description. }
    long_description: { es: Larga., en: Long. }
    harnesses: [starter-only-skill]`

	_, err := parse([]byte(invalid))
	if err == nil {
		t.Fatal("expected error for starter-only harness with one-language long_description, got nil")
	}
	for _, want := range []string{`"starter-only-skill"`, "long_description", `"en"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}

// TestCatalogLoad_Starter_MissingLongDescription_Fails asserts that a
// starter omitting long_description fails Load, naming the starter and the
// long_description field.
func TestCatalogLoad_Starter_MissingLongDescription_Fails(t *testing.T) {
	invalid := `harnesses:
  - id: h-one
    name: H One
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [lite]
starters:
  - id: bad-starter
    name: Bad Starter
    description: { es: Descripción de starter., en: Starter description. }
    harnesses: [h-one]`

	_, err := parse([]byte(invalid))
	if err == nil {
		t.Fatal("expected error for starter missing long_description, got nil")
	}
	for _, want := range []string{`"bad-starter"`, "long_description"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}

// TestCatalogLoad_Starter_MissingDescription_Fails asserts that a starter
// entirely omitting description fails Load, naming the starter and the
// description field.
func TestCatalogLoad_Starter_MissingDescription_Fails(t *testing.T) {
	invalid := `harnesses:
  - id: h-one
    name: H One
    description: { es: Descripción., en: Description. }
    long_description: { es: Larga., en: Long. }
    type: config
    install_modes: [lite]
starters:
  - id: bad-starter
    name: Bad Starter
    long_description: { es: Larga., en: Long. }
    harnesses: [h-one]`

	_, err := parse([]byte(invalid))
	if err == nil {
		t.Fatal("expected error for starter missing description, got nil")
	}
	for _, want := range []string{`"bad-starter"`, "description"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}
