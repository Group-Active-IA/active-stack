# catalog-localization Specification

## Purpose
TBD - created by archiving change catalog-localized-descriptions. Update Purpose after archive.
## Requirements
### Requirement: LocalizedText resolves a language with deterministic fallback

The model SHALL provide a `LocalizedText` type (`map[string]string`) with a `Localized(lang string) string` method that returns the value for `lang` when non-empty, otherwise the value for `"en"`, otherwise the value for `"es"`, otherwise the empty string. The method MUST be safe on a nil or empty map (returning `""`, never panicking).

#### Scenario: Requested language present

- **WHEN** `Localized("es")` is called on a `LocalizedText` holding both `es` and `en`
- **THEN** the `es` value is returned

#### Scenario: Requested language missing falls back to en then es

- **WHEN** `Localized("fr")` is called on a `LocalizedText` holding only `en` and `es`
- **THEN** the `en` value is returned
- **AND WHEN** the map holds only `es`
- **THEN** the `es` value is returned

#### Scenario: Empty or nil map returns empty string

- **WHEN** `Localized("en")` is called on a nil or empty `LocalizedText`
- **THEN** the empty string is returned and no panic occurs

### Requirement: LocalizedText accepts scalar and map YAML forms

`LocalizedText` SHALL implement a custom YAML unmarshaler that accepts a scalar node — assigning its value to the `"es"` key for parse-time backward compatibility during migration — as well as a mapping node decoded key-for-key into the map. Accepting the scalar form MUST NOT bypass bilingual validation: a scalar-sourced value populates only `"es"`, so a later `Validate` call still fails on the missing `"en"`.

#### Scenario: Scalar YAML maps to the es key

- **WHEN** a `LocalizedText` field is unmarshaled from a bare scalar string
- **THEN** the resulting map has that string under key `"es"` and no `"en"` key

#### Scenario: Mapping YAML decodes each language key

- **WHEN** a `LocalizedText` field is unmarshaled from a `{es: ..., en: ...}` mapping
- **THEN** the resulting map holds both the `es` and `en` values under their respective keys

### Requirement: LocalizedText validates that both languages are present

`LocalizedText` SHALL provide `Validate(field, owner string) error` that returns an error unless BOTH `"es"` and `"en"` are present and non-empty. The error message MUST name the owner, the field, and the missing language, in the style `catalog: <owner> <field> missing language "en"`.

#### Scenario: Both languages present passes

- **WHEN** `Validate("description", "harness \"openspec\"")` is called on a map with non-empty `es` and `en`
- **THEN** it returns nil

#### Scenario: Missing language fails with a naming message

- **WHEN** `Validate("description", "harness \"openspec\"")` is called on a map holding only `es`
- **THEN** it returns an error whose message names the harness, the `description` field, and the missing language `"en"`

### Requirement: Harness and Starter carry localized description and long description

`model.Harness` and `model.Starter` SHALL type their `Description` field as `LocalizedText` (`yaml:"description,omitempty"`) and SHALL each gain a `LongDescription LocalizedText` field (`yaml:"long_description,omitempty"`). These fields carry per-language catalog text; the display name (`Name`) MUST remain a non-localized string (product names are not translated).

#### Scenario: Localized description round-trips from YAML

- **WHEN** a harness entry with `description: {es: ..., en: ...}` is parsed
- **THEN** `harness.Description.Localized("en")` returns the English value and `Localized("es")` returns the Spanish value

#### Scenario: Long description field is present and localizable

- **WHEN** a harness or starter entry declares `long_description: {es: ..., en: ...}`
- **THEN** the parsed `LongDescription` resolves each language via `Localized`

### Requirement: Catalog load enforces bilingual completeness

`catalog.Load()` SHALL validate localized text during load and fail loudly on any gap: every harness and every starter `Description` MUST have both `es` and `en`; `LongDescription` MUST have both languages for every global-picker harness (`!IsStarterOnly()`) and for every starter; for starter-only harnesses `LongDescription` is optional but, when present, MUST have both languages. A missing translation MUST cause `Load()` to return an error naming the offending harness/starter, the field, and the missing language, so a build/release cannot ship a one-language or un-migrated catalog. A legacy scalar description (parsed into `es` only) therefore fails load on the missing `en`.

#### Scenario: Fully bilingual catalog loads

- **WHEN** the embedded `harnesses.yaml` provides `es` + `en` for every mandatory `description` and `long_description`
- **THEN** `catalog.Load()` succeeds

#### Scenario: Harness description missing a language fails load

- **WHEN** a harness `description` provides only `es`
- **THEN** `catalog.Load()` returns an error naming that harness, the `description` field, and the missing `en`

#### Scenario: Legacy scalar description fails load

- **WHEN** a harness `description` is a bare scalar string (legacy form)
- **THEN** it parses into `es` only AND `catalog.Load()` fails on the missing `en`

#### Scenario: Global-picker harness missing long_description fails load

- **WHEN** a non-starter-only harness omits `long_description` or provides only one language
- **THEN** `catalog.Load()` returns an error naming that harness and the `long_description` field

#### Scenario: Starter-only harness may omit long_description

- **WHEN** a starter-only harness has a valid bilingual `description` and no `long_description`
- **THEN** `catalog.Load()` succeeds for that harness

#### Scenario: Starter missing long_description fails load

- **WHEN** a starter omits `long_description` or provides only one language
- **THEN** `catalog.Load()` returns an error naming that starter and the `long_description` field

### Requirement: Long-description content follows a three-part structure

For every catalog entry whose `long_description` is required (global-picker harnesses, i.e. `!IsStarterOnly()`, and all starters) and for the install-mode, permission-tier, and uninstall-strategy `*.long` keys in `internal/i18n`, the authored content SHALL, in each language (ES and EN), state (1) what the item is, (2) what it concretely installs or configures (binaries, files, MCPs, where it lands), and (3) when it makes sense to choose it, and when it does not; each language's text SHALL be 3-6 sentences. Harnesses with `scope: starter-only` are exempt: their `long_description` remains optional, and this requirement does not apply to them.

#### Scenario: Global-picker harness long_description follows the structure

- **WHEN** a `long_description` is authored for a harness that is not
  starter-only
- **THEN** its ES and EN text each state what the harness is, what it
  concretely installs/configures, and when to choose it (or not), in 3-6
  sentences per language

#### Scenario: Starter long_description follows the structure

- **WHEN** a `long_description` is authored for a starter (`base`, `backend`,
  `ux-ui`, `active-ia`)
- **THEN** its ES and EN text follow the same three-part structure

#### Scenario: Mode, tier, and strategy long text follows the structure

- **WHEN** a `mode.*.long`, `tier.*.long`, or `strategy.*.long` key is
  authored in `internal/i18n/table_es.go` and `table_en.go`
- **THEN** its ES and EN text follow the same three-part structure

#### Scenario: Starter-only harness is exempt

- **WHEN** a harness has `scope: starter-only`
- **THEN** it is not required to carry a `long_description`, and its absence
  does not violate this requirement

### Requirement: Drafted content is approved before being written to source

The complete table of proposed ES+EN long-description and `*.long` content SHALL be presented to the maintainer for explicit approval before any of it is written into `internal/catalog/harnesses.yaml` or `internal/i18n/table_{es,en}.go`. No file in either location may be edited with this content until that approval is given.

#### Scenario: Full table is presented before any file write

- **WHEN** the ES+EN content for all in-scope harnesses, starters, modes,
  tiers, and strategies has been drafted
- **THEN** the complete table is presented to the maintainer for review before
  `internal/catalog/harnesses.yaml` or `internal/i18n/table_{es,en}.go` is
  touched

#### Scenario: No write occurs without explicit approval

- **WHEN** the maintainer has not yet explicitly approved the drafted content
- **THEN** `internal/catalog/harnesses.yaml` and
  `internal/i18n/table_{es,en}.go` remain unmodified

