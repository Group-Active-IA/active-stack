## ADDED Requirements

### Requirement: Wizard shell hosts logical page navigation

The Windows GUI SHALL present a multi-page wizard hosted by a thin shell (`MainWindow` + `ShellViewModel`) composed of a header (the current page's title and subtitle), a `ContentControl` bound to the current page view-model, and a footer with `Back`, a primary action (`Next` or `Install`), and `Cancel`. Page views SHALL be selected by `DataTemplate` keyed on the page view-model's type (not `Frame`/`NavigationService`). The window SHALL open at 760×560, be resizable (`ResizeMode=CanResize`), and MUST NOT be maximized. The single-screen `MainWindowViewModel` dashboard SHALL be replaced by `ShellViewModel` plus per-page view-models.

#### Scenario: Shell renders the current page and footer

- **WHEN** the wizard is showing a page whose view-model exposes a title, subtitle, and advance-eligibility
- **THEN** the shell displays that title and subtitle in the header, renders the page's view via its type's `DataTemplate` in the `ContentControl`, and enables the primary footer action only when the page reports it can advance

#### Scenario: Back is unavailable on the first page of a flow

- **WHEN** the current page is the Hub (the flow entry)
- **THEN** the `Back` action is disabled

### Requirement: Install flow ordering mirrors the TUI

The Install flow SHALL be driven by pure `WizardFlow.NextPage`/`WizardFlow.PreviousPage` functions computed from the current page, the shared install selection (agents, mode, custom component ids, tier), and the session's tier-capable agents — with no re-query to the engine on selection change. The order SHALL be: Hub → Assistants → InstallType → (Components only when mode is `custom`) → (Permissions only when at least one selected agent is tier-capable) → Review → Installing → Complete. Tier-capability SHALL be the intersection of the session's `TierCapableAgents` with the selected agents.

#### Scenario: Full install with a tier-capable agent shows Permissions, not Components

- **WHEN** the selection is mode `full` with a tier-capable agent (e.g. `claude`) and `NextPage` is evaluated from InstallType
- **THEN** the next page is Permissions (Components is skipped because mode is not custom)

#### Scenario: Custom install shows Components

- **WHEN** the selection is mode `custom` and `NextPage` is evaluated from InstallType
- **THEN** the next page is Components

#### Scenario: A non-tier-capable-only selection skips Permissions

- **WHEN** the only selected agent is not tier-capable (e.g. gemini-only) and `NextPage` is evaluated from InstallType (mode `full`) or from Components (mode `custom`)
- **THEN** the next page is Review (Permissions is skipped)

#### Scenario: Back from Review skips pages not in the current path

- **WHEN** the selection is mode `full` with a gemini-only (non-tier-capable) agent and `PreviousPage` is evaluated from Review
- **THEN** the previous page is InstallType (both Permissions and Components are skipped because neither is in the path)

### Requirement: Assistants page requires at least one agent

The Assistants page SHALL present the available agents as multi-select checkboxes, pre-checking those reported as detected by the session and marking recommended agents. The page SHALL report it cannot advance while zero agents are selected, and the shell's primary action MUST be disabled in that state.

#### Scenario: Advancing is blocked with no agents selected

- **WHEN** the Assistants page has zero selected agents
- **THEN** the page reports it cannot advance and the primary footer action is disabled

#### Scenario: Detected agents are pre-checked

- **WHEN** the Assistants page is shown for a session that detected `claude`
- **THEN** `claude` is checked by default and the page can advance

### Requirement: Install type defaults to full with end-user labels

The InstallType page SHALL present lite, full, and custom as selectable cards using end-user labels (Quick / Complete / Custom) with the internal mode mapping behind them, defaulting the selection to `full`.

#### Scenario: Full is preselected

- **WHEN** the InstallType page is first shown
- **THEN** the `full` (Complete) card is selected by default and the page can advance

### Requirement: Components page forces and locks the permissions harness

For `custom` mode, the Components page SHALL list the available components as checkboxes, pre-checking recommended items, and SHALL render the security-first `permissions` harness checked and disabled (it cannot be unchecked), showing the engine-provided descriptive copy. The `permissions` id SHALL always be present in the custom component ids sent to the engine.

#### Scenario: Permissions cannot be unchecked

- **WHEN** the user attempts to toggle the `permissions` component off on the Components page
- **THEN** the component remains checked and disabled, and the resulting custom ids still include `permissions`

### Requirement: Permissions page offers the engine's tiers with a default and bypass warning

When shown, the Permissions page SHALL render the tiers from the session's `permission_tiers` as a single-select (radio) group, preselecting the tier marked default (balanceado), and SHALL surface the tier's warning text when a tier carrying a warning (bypass) is selected.

#### Scenario: Balanceado is preselected and bypass shows a warning

- **WHEN** the Permissions page is shown with three tiers where balanceado is default and bypass carries a warning
- **THEN** balanceado is selected initially, and selecting bypass surfaces its warning text

### Requirement: Review confirms the selection and triggers install

The Review page SHALL present a read-only summary of the selection (agents, install type, chosen components, tier) and the shell's primary action on Review SHALL read "Install". Advancing from Review SHALL start the install by invoking the engine client's install stream with the full selection (agents, mode, custom ids, tier).

#### Scenario: Review primary action starts the install with the full selection

- **WHEN** the user confirms on the Review page with agents, mode, custom ids, and tier set
- **THEN** the install stream is started carrying all of those values (not a single agent and not without tier/custom)

### Requirement: Progress page renders the install event stream

A generic `ProgressPageViewModel` SHALL consume an `IAsyncEnumerable<InstallProgressSnapshot>` and render phase-based progress, friendly step messages, and an activity log, reaching a terminal success/degraded/rollback/error state. The snapshot-to-UI logic (progress calculation, friendly-message and display-label mapping, terminal success detection) SHALL be the behavior migrated from the former `MainWindowViewModel` with its tests preserved.

#### Scenario: Terminal install_finished sets completion

- **WHEN** the progress stream emits an `install_finished` snapshot with success true
- **THEN** progress reaches 100 and the page reports the install succeeded

#### Scenario: Migrated progress behavior is unchanged

- **WHEN** the migrated progress tests (previously in `MainWindowViewModelTests`) run against `ProgressPageViewModel`
- **THEN** they pass with unchanged assertions

### Requirement: Hub surfaces all destinations with unimplemented ones disabled

The Hub page SHALL present the entries Install, Starters, Manage Backups, Uninstall, and a disabled "Update Stack — Coming soon". In this change only Install SHALL be enabled; Starters, Manage Backups, and Uninstall SHALL be present but disabled with an explanatory tooltip (no placeholder navigation), matching the TUI's "coming soon" treatment.

#### Scenario: Only Install is actionable on the Hub

- **WHEN** the Hub page is shown
- **THEN** the Install entry is enabled and navigates into the Install flow, while Starters, Manage Backups, Uninstall, and Update Stack are disabled with a tooltip and do not navigate
