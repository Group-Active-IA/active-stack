# windows-install-contract Specification

## Purpose
TBD - created by archiving change windows-contract-tier-multiagent. Update Purpose after archive.
## Requirements
### Requirement: Headless install accepts a permission tier flag

The `active-stack install` and `active-stack windows install` commands SHALL accept a `--tier` flag whose value maps to `install.Intent.Tier`. Valid values are `estricto`, `balanceado`, and `bypass`. An unrecognized value MUST cause `ParseInstallFlags` to return an error (non-nil), and the caller MUST exit non-zero. An empty `--tier` MUST leave `Intent.Tier` at its zero value so the engine normalizes it to `balanceado`. Passing `--tier` MUST trigger headless mode even when `--headless`, `--mode`, and `--agent` are all absent (same criterion as `--custom`).

#### Scenario: Valid tier maps to intent

- **WHEN** `ParseInstallFlags` receives `--tier estricto`
- **THEN** the returned `ParsedFlags.TUI` is false AND `Intent.Tier` equals `model.TierEstricto`

#### Scenario: Invalid tier is rejected

- **WHEN** `ParseInstallFlags` receives `--tier loose`
- **THEN** it returns a non-nil error naming the invalid value AND no install proceeds

#### Scenario: Empty tier normalizes to balanceado

- **WHEN** `ParseInstallFlags` is called without `--tier` but with other headless intent
- **THEN** `Intent.Tier` is the zero value, which the engine normalizes to `model.TierBalanceado`

#### Scenario: Tier flag alone implies headless

- **WHEN** `ParseInstallFlags` receives only `--tier balanceado` (no `--mode`, `--agent`, or `--headless`)
- **THEN** `ParsedFlags.TUI` is false (headless mode is triggered)

### Requirement: windows options accepts multiple agents

The `active-stack windows options` command SHALL accept a comma-separated `--agent` list, and `RunWindowsOptions` SHALL take `[]model.Agent`. Component selection MUST be the deduplicated union of the harnesses available to all requested agents. Requesting `--agent` with zero non-empty entries MUST be rejected by the dispatch as an error.

#### Scenario: Multi-agent union is deduplicated

- **WHEN** `windows options --agent claude,opencode` is run and a harness is available to both agents
- **THEN** the component appears exactly once in the options response (no duplicate entries)

#### Scenario: Empty agent list is rejected

- **WHEN** `windows options` is run with no `--agent` value (or only empty/whitespace entries)
- **THEN** the dispatch prints an error and returns a non-zero exit code

### Requirement: windows options exposes tier metadata additively

The `windows options` JSON response SHALL include, in addition to its existing `modes`, `forced_components`, and `custom_components` fields, three new fields: `tier_capable` (bool), `tier_capable_agents` (array of agent id strings), and `permission_tiers` (array of `{id, label, description, default, warning}` objects). `tier_capable` MUST be true when at least one requested agent is tier-capable and false otherwise. Existing field names and semantics MUST NOT change (the extension is purely additive), so a client reading only the pre-existing fields keeps working.

#### Scenario: Tier-capable when a capable agent is selected

- **WHEN** `windows options --agent claude` is run
- **THEN** `tier_capable` is true AND `permission_tiers` lists the three tiers with `balanceado` marked `default: true`

#### Scenario: Not tier-capable when no capable agent is selected

- **WHEN** `windows options --agent gemini` is run (gemini is not tier-capable)
- **THEN** `tier_capable` is false

#### Scenario: Existing fields remain unchanged

- **WHEN** any `windows options` request is served
- **THEN** the `modes`, `forced_components`, and `custom_components` fields retain their existing names and shape (the five existing contract assertions stay green)

### Requirement: Tier capability has a single source of truth in the model

The knowledge of which agents are tier-capable SHALL live in `internal/model` as an exported function `model.TierCapable(agents []model.Agent) bool`. Both the TUI permission screen and the `windows options` command MUST consume this function rather than defining their own agent set, so the two paths cannot diverge.

#### Scenario: Model function drives both consumers

- **WHEN** `model.TierCapable` is called with an agent set containing at least one tier-capable agent
- **THEN** it returns true, AND both `internal/tui` and `windows options` produce their tier-capable behavior from this same result

#### Scenario: No tier-capable agent

- **WHEN** `model.TierCapable` is called with a set containing only non-tier-capable agents
- **THEN** it returns false

### Requirement: windows options populates custom_components with the selectable harness list

The `custom_components` field of the `windows options` JSON response SHALL contain the real list of harnesses the user may choose in Custom mode: every catalog harness available in Custom mode, filtered to those supporting at least one of the requested agents, as a union deduplicated by harness ID. The list MUST NOT be sourced from `install.SelectHarnesses` with an empty custom selection (which returns only the forced `permissions` harness and leaves `custom_components` structurally empty). The forced `permissions` harness MUST NOT appear in `custom_components`; it stays in `forced_components`. Each entry's `recommended` flag MUST be true exactly when the harness belongs to lite or full mode. The Windows picker list MUST derive from the same source of truth as the TUI Custom picker so the two cannot diverge.

#### Scenario: Custom components is non-empty and mirrors the catalog

- **WHEN** `windows options --agent claude` is run
- **THEN** `custom_components` is non-empty AND contains the catalog harnesses available to `claude` in Custom mode, EXCLUDING `permissions`

#### Scenario: Forced permissions is excluded from custom components

- **WHEN** any `windows options` request is served
- **THEN** the `permissions` harness appears in `forced_components` AND does NOT appear in `custom_components`

#### Scenario: Recommended flag reflects lite/full membership

- **WHEN** `windows options` returns a `custom_components` entry
- **THEN** its `recommended` flag is true if and only if that harness is present in lite mode or full mode

#### Scenario: Multi-agent picker list is the deduplicated union

- **WHEN** `windows options --agent claude,opencode` is run and a harness supports both agents
- **THEN** that harness appears exactly once in `custom_components` (union deduplicated by ID)

#### Scenario: Picker list shares its source with the TUI

- **WHEN** the Windows `custom_components` list and the TUI Custom picker are both built for the same agent set
- **THEN** both derive from the same shared helper (`install.CustomPickerHarnesses`) so their selectable harness sets are identical

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

