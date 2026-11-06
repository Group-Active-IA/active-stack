## ADDED Requirements

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
