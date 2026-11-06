## ADDED Requirements

### Requirement: windows install commands accept a language flag

The `active-stack windows options` and `active-stack windows install` commands SHALL accept a `--lang` flag whose value is parsed by `i18n.Parse`. Valid values are `en` and `es`. When `--lang` is absent, the language MUST default to `i18n.Default` (`en`), and the emitted JSON MUST be byte-compatible with the pre-change output for these commands (aside from the tier-label fix specified below). An unrecognized `--lang` value MUST cause the dispatch to print an error and return a non-zero exit code before any options/install work runs. The resolved language MUST be threaded into every user-facing string these commands emit, sourced from the `internal/i18n` tables rather than hardcoded literals.

#### Scenario: Default language is en and output stays byte-compatible

- **WHEN** `windows options --agent claude` is run without `--lang`
- **THEN** the response is produced in English AND every pre-existing field name and value (except the corrected tier labels) is identical to the pre-change output, so the existing `windows_test.go` assertions stay green (first mode label `"Quick"`, `permissions` in `forced_components`)

#### Scenario: Spanish language localizes user-facing strings

- **WHEN** `windows options --agent claude --lang es` is run
- **THEN** the `modes`, `permission_tiers`, and component-description strings are emitted in Spanish, while the `id` fields (`lite`/`full`/`custom`, `estricto`/`balanceado`/`bypass`) remain unchanged

#### Scenario: Invalid language is rejected

- **WHEN** any `windows options` or `windows install` command receives `--lang xx`
- **THEN** the dispatch prints an error naming the invalid value AND returns a non-zero exit code AND no options/install work proceeds

### Requirement: Permission tiers are language-consistent and carry a long description

The `permission_tiers` array emitted by `windows options` SHALL, for each tier, source its `label`, `description`, and `warning` from the `internal/i18n` tables in the requested language, so that labels and descriptions are consistent within one language (fixing the prior Spanish-label / English-description mix). Under the default `en` language the tier labels MUST be `Strict`, `Balanced`, and `Bypass`. Under `es` they MUST be `Estricto`, `Balanceado`, and `Bypass`. The tier `id` values MUST remain `estricto`, `balanceado`, and `bypass` in every language (contract-stable). Each `permission_tiers` entry SHALL additionally carry an additive `long_description` field (`json:"long_description,omitempty"`) sourced from the tier `.long` key; its content MAY be placeholder-quality in this change.

#### Scenario: English tier labels are consistent

- **WHEN** `windows options --agent claude` is served under default `en`
- **THEN** the three tiers have labels `Strict`, `Balanced`, `Bypass` with English descriptions AND `balanceado` is still marked `default: true` AND `bypass` still carries its warning

#### Scenario: Spanish tier labels are consistent

- **WHEN** `windows options --agent claude --lang es` is served
- **THEN** the three tiers have labels `Estricto`, `Balanceado`, `Bypass` with Spanish descriptions AND the `id` values remain `estricto`/`balanceado`/`bypass`

#### Scenario: Tiers carry an additive long_description

- **WHEN** the `permission_tiers` array is emitted
- **THEN** each entry includes a `long_description` field populated from the i18n `tier.*.long` key AND clients reading only the pre-existing tier fields keep working

### Requirement: Mode and component options carry an additive long description

The `modes` entries emitted by `windows options` SHALL each carry an additive `long_description` field (`json:"long_description,omitempty"`) sourced from the `mode.*.long` i18n key in the requested language. The `forced_components` and `custom_components` entries SHALL each expose a `long_description` field of the same additive shape; in this change that field is left EMPTY (it is populated from the catalog by a later change). Adding these fields MUST NOT change any existing field name or value, so a client reading only the pre-existing fields keeps working.

#### Scenario: Modes carry a populated long_description

- **WHEN** `windows options` emits its `modes` array
- **THEN** each mode entry includes a `long_description` populated from the `mode.*.long` i18n key for the requested language

#### Scenario: Component long_description is present but empty

- **WHEN** `windows options` emits `forced_components` or `custom_components`
- **THEN** each entry includes a `long_description` field that is empty in this change AND all pre-existing component fields (`id`, `label`, `description`, `recommended`) retain their names and values
