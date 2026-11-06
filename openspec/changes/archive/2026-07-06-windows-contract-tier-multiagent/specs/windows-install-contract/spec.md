## ADDED Requirements

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
