## MODIFIED Requirements

### Requirement: Mode and component options carry an additive long description

The `modes` entries emitted by `windows options` SHALL each carry an additive `long_description` field (`json:"long_description,omitempty"`) sourced from the `mode.*.long` i18n key in the requested language. The `forced_components` and `custom_components` entries SHALL each carry a `long_description` field of the same additive shape, populated from the catalog harness's localized `LongDescription` resolved in the requested language (`h.LongDescription.Localized(lang)`). The component `description` field SHALL be resolved in the requested language from the catalog harness's localized `Description` (`h.Description.Localized(lang)`), falling back to the type-based i18n fallback strings when the harness has no localized description. Adding/populating these fields MUST NOT change any existing field name, and a client reading only the pre-existing fields keeps working.

#### Scenario: Modes carry a populated long_description

- **WHEN** `windows options` emits its `modes` array
- **THEN** each mode entry includes a `long_description` populated from the `mode.*.long` i18n key for the requested language

#### Scenario: Component description is localized from the catalog

- **WHEN** `windows options --agent claude --lang es` emits a component whose harness has a bilingual catalog description
- **THEN** the component `description` is the Spanish catalog value
- **AND WHEN** the same command runs under default `en`
- **THEN** the component `description` is the English catalog value

#### Scenario: Component long_description is populated from the catalog

- **WHEN** `windows options` emits `forced_components` or `custom_components`
- **THEN** each entry's `long_description` is the harness's localized `LongDescription` resolved in the requested language AND all pre-existing component fields (`id`, `label`, `description`, `recommended`) retain their names

#### Scenario: Component without a catalog description uses the type fallback

- **WHEN** a component harness has no localized `Description` value for the requested language and no fallback language
- **THEN** the component `description` falls back to the type-based i18n string (security/external/config/skill) in the requested language
