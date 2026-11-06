## Purpose

Define the Windows GUI wizard shell (WPF bootstrapper UI) and its operation flows: a Hub that dispatches to Install, Uninstall, Starters, and Backups, each driven by pure per-operation flow functions and per-page view-models, mirroring the TUI's ordering and semantics.

## Requirements

### Requirement: Wizard shell hosts logical page navigation

The Windows GUI SHALL present a multi-page wizard hosted by a thin shell (`MainWindow` + `ShellViewModel`) composed of a header (the current page's title and subtitle), a `ContentControl` bound to the current page view-model, and a footer with `Back`, a primary action (`Next` or `Install`), and `Cancel`. Page views SHALL be selected by `DataTemplate` keyed on the page view-model's type (not `Frame`/`NavigationService`). The window SHALL open at 760×560, be resizable (`ResizeMode=CanResize`), and MUST NOT be maximized. The single-screen `MainWindowViewModel` dashboard SHALL be replaced by `ShellViewModel` plus per-page view-models. The shell SHALL open on the Language page (the wizard's first page); the `Back` action SHALL be disabled only on the Language page — the Hub is reachable Back to Language and therefore has `Back` enabled.

#### Scenario: Shell renders the current page and footer

- **WHEN** the wizard is showing a page whose view-model exposes a title, subtitle, and advance-eligibility
- **THEN** the shell displays that title and subtitle in the header, renders the page's view via its type's `DataTemplate` in the `ContentControl`, and enables the primary footer action only when the page reports it can advance

#### Scenario: Back is unavailable on the Language page (the wizard's first page)

- **WHEN** the current page is the Language page (the wizard entry)
- **THEN** the `Back` action is disabled

#### Scenario: Back is available on the Hub because Language precedes it

- **WHEN** the current page is the Hub
- **THEN** the `Back` action is enabled and returns to the Language page

### Requirement: Install flow ordering mirrors the TUI

The Install flow SHALL be driven by pure `WizardFlow.NextPage`/`WizardFlow.PreviousPage` functions computed from the current page, the shared install selection (agents, mode, custom component ids, tier), and the session's tier-capable agents — with no re-query to the engine on selection change. The order SHALL be: Hub → Dependencies → Assistants → InstallType → (Components only when mode is `custom`) → (Permissions only when at least one selected agent is tier-capable) → Review → Installing → Complete. Tier-capability SHALL be the intersection of the session's `TierCapableAgents` with the selected agents.

#### Scenario: Dependencies page appears right after the Hub

- **WHEN** the Install flow is entered from the Hub and `NextPage` is evaluated from Hub
- **THEN** the next page is Dependencies, and evaluating `NextPage` from Dependencies yields Assistants

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

The InstallType page SHALL present lite, full, and custom as selectable cards using end-user labels (Quick / Complete / Custom) with the internal mode mapping behind them, defaulting the selection to `full`. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected install type.

#### Scenario: Full is preselected

- **WHEN** the InstallType page is first shown
- **THEN** the `full` (Complete) card is selected by default and the page can advance

#### Scenario: Detail panel reflects the selected install type

- **WHEN** the user selects a different install type card
- **THEN** the detail panel updates to that type's long description, falling back to its short description when no long description is present

### Requirement: Components page forces and locks the permissions harness

For `custom` mode, the Components page SHALL list the available components as a single-select `ListBox` whose rows carry a checkbox each, pre-checking recommended items, and SHALL render the security-first `permissions` harness checked and disabled (it cannot be unchecked), showing the engine-provided descriptive copy. The `permissions` id SHALL always be present in the custom component ids sent to the engine. Row selection (which item is focused/highlighted) SHALL be tracked independently of each row's checkbox state: selecting a row SHALL NOT change its checkbox, and toggling a row's checkbox SHALL NOT be required to select that row. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected row.

#### Scenario: Permissions cannot be unchecked

- **WHEN** the user attempts to toggle the `permissions` component off on the Components page
- **THEN** the component remains checked and disabled, and the resulting custom ids still include `permissions`

#### Scenario: Row selection is independent of the checkbox

- **WHEN** the user selects a component row without toggling its checkbox
- **THEN** the row becomes the selection (feeding the detail panel) while the resulting custom component ids sent to the engine are unchanged

#### Scenario: Detail panel reflects the selected component

- **WHEN** the user selects a component row whose long description is present
- **THEN** the detail panel shows that component's long description; when the long description is empty, the panel shows the component's short description instead

### Requirement: Permissions page offers the engine's tiers with a default and bypass warning

When shown, the Permissions page SHALL render the tiers from the session's `permission_tiers` as a single-select (radio) group, preselecting the tier marked default (balanceado), and SHALL surface the tier's warning text when a tier carrying a warning (bypass) is selected. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected tier.

#### Scenario: Balanceado is preselected and bypass shows a warning

- **WHEN** the Permissions page is shown with three tiers where balanceado is default and bypass carries a warning
- **THEN** balanceado is selected initially, and selecting bypass surfaces its warning text

#### Scenario: Detail panel reflects the selected tier

- **WHEN** the user selects a different permission tier
- **THEN** the detail panel updates to that tier's long description, falling back to its short description when no long description is present

### Requirement: Review confirms the selection and triggers install

The Review page SHALL present a read-only summary of the selection (agents,
install type, chosen components, tier) and the shell's primary action on Review
SHALL read "Install". Advancing from Review SHALL start the install by invoking
the engine client's install stream with the full selection (agents, mode, custom
ids, tier).

The components summary SHALL be mode-aware so it reflects what will actually be
installed regardless of install mode, and SHALL NOT be blank for any mode that
installs components:

- In `custom` mode the components summary SHALL be derived from the selection's
  custom component ids (which include the forced `permissions` id and the user's
  picks).
- In a non-custom mode (`lite` or `full`), where the Components page is skipped
  and the custom component ids are legitimately empty, the components summary
  SHALL instead be derived from the session's forced components together with the
  session's custom components marked as included in the mode's default
  (recommended) set — never from the empty custom ids.

#### Scenario: Review primary action starts the install with the full selection

- **WHEN** the user confirms on the Review page with agents, mode, custom ids, and tier set
- **THEN** the install stream is started carrying all of those values (not a single agent and not without tier/custom)

#### Scenario: Custom mode summarizes the chosen component ids

- **WHEN** the Review page is shown for `custom` mode with a non-empty set of custom component ids
- **THEN** the components summary lists exactly the labels of the session components whose ids are in that custom-id set (and no others)

#### Scenario: Custom components selected on Components appear in the Review summary

- **WHEN** the user toggles a set of components on the Components page (writing them into the shared `InstallSelection`) and reaches the Review page built over that same selection
- **THEN** the Review page's components summary lists exactly the selected components (including the forced `permissions` harness), matching what will be sent to the engine

#### Scenario: Full mode summary reflects the mode's implied set, not the empty custom ids

- **WHEN** the Review page is shown for `full` mode with empty custom component ids (the Components page was skipped)
- **THEN** the components summary is non-empty and lists the session's forced components plus the custom components marked recommended, and excludes custom components not marked recommended

#### Scenario: Lite mode summary is non-empty and mode-scoped

- **WHEN** the Review page is shown for `lite` mode with empty custom component ids (the Components page was skipped)
- **THEN** the components summary is non-empty and lists the session's forced components plus the custom components marked recommended (rather than rendering blank)

### Requirement: Progress page renders the install event stream

A generic `ProgressPageViewModel` SHALL consume an `IAsyncEnumerable<InstallProgressSnapshot>` and render phase-based progress, friendly step messages, and an activity log, reaching a terminal success/degraded/rollback/error state. The snapshot-to-UI logic (progress calculation, friendly-message and display-label mapping, terminal success detection) SHALL be the behavior migrated from the former `MainWindowViewModel` with its tests preserved.

#### Scenario: Terminal install_finished sets completion

- **WHEN** the progress stream emits an `install_finished` snapshot with success true
- **THEN** progress reaches 100 and the page reports the install succeeded

#### Scenario: Migrated progress behavior is unchanged

- **WHEN** the migrated progress tests (previously in `MainWindowViewModelTests`) run against `ProgressPageViewModel`
- **THEN** they pass with unchanged assertions

### Requirement: Hub surfaces all destinations with unimplemented ones disabled

The Hub page SHALL present the entries Install, Starters, Manage Backups, Uninstall, and a disabled "Update Stack — Coming soon". Install, Starters, Manage Backups, and Uninstall SHALL each be enabled and, when chosen, SHALL start their respective flow (Install wizard, Starter flow, Backups page, or Uninstall wizard). Update Stack SHALL remain disabled with a "Coming soon" tooltip, matching the TUI's permanent placeholder treatment. Choosing a disabled entry SHALL remain a no-op.

#### Scenario: Every implemented destination is actionable on the Hub

- **WHEN** the Hub page is shown
- **THEN** Install, Starters, Manage Backups, and Uninstall are each enabled and can advance into their flow, while Update Stack is disabled with a tooltip and does not navigate

#### Scenario: Selecting a Hub entry starts that entry's flow

- **WHEN** the user selects the Uninstall entry (or Starters, or Manage Backups) and advances
- **THEN** the shell enters the corresponding flow's first page (Uninstall → UninstallAgents, Starters → StarterCatalog, Manage Backups → the Backups page) rather than the Install flow

### Requirement: Shell dispatches the Hub selection to the chosen operation's flow

The `ShellViewModel` SHALL track which operation (Install, Uninstall, Starter, or Backups) the active flow belongs to, set from the Hub selection, and SHALL drive Back/Next navigation through that operation's pure flow functions. Advancing from the Hub SHALL begin the selected operation; the Install flow MUST NOT be assumed as the only flow. Returning Back to the Hub SHALL clear the active operation so a different entry can be chosen.

#### Scenario: Uninstall selection routes through the uninstall flow functions

- **WHEN** the active operation is Uninstall and the shell computes the next or previous page
- **THEN** it uses the uninstall flow functions (not the install flow functions) to resolve the page

#### Scenario: Back to Hub resets the active operation

- **WHEN** the user is on the first page of a flow and goes Back to the Hub
- **THEN** the active operation is cleared and any other enabled Hub entry can be selected next

### Requirement: Navigation exposes pure per-operation flow functions

The navigation layer SHALL expose pure, side-effect-free `NextPage`/`PreviousPage` functions per operation (Install, Uninstall, Starter), each computed only from the current page identity plus that operation's selection state — no I/O and no mutation. Keeping a separate function set per operation (rather than one monolithic switch) is the recorded decision; each set MUST be independently unit-testable and MUST mirror the corresponding TUI transitions.

#### Scenario: Uninstall flow ordering mirrors the TUI

- **WHEN** the uninstall `NextPage` function is evaluated across the flow
- **THEN** the order is Hub → UninstallAgents → UninstallMode → UninstallStrategy → UninstallConfirm → Progress → Complete

#### Scenario: Starter flow ordering mirrors the intended sequence

- **WHEN** the starter `NextPage` function is evaluated across the flow
- **THEN** the order is Hub → StarterCatalog → StarterTarget → Review → Progress → Complete

#### Scenario: Back is the inverse of Next within an operation

- **WHEN** a `PreviousPage` function is evaluated for any non-entry page of its operation
- **THEN** it returns the page that `NextPage` would have advanced from, and returns the Hub from the operation's first page

### Requirement: Uninstall agents page requires at least one agent

The UninstallAgents page SHALL present the agents from the uninstall options' `detected_agents` as multi-select checkboxes, pre-checking all detected agents. The page SHALL report it cannot advance while zero agents are selected, and the shell's primary action MUST be disabled in that state. This is a deliberate divergence from the TUI's implicit "empty means all": the wizard requires an explicit, non-empty selection for a destructive operation.

#### Scenario: Detected agents are pre-checked and advancing is allowed

- **WHEN** the UninstallAgents page is shown for options reporting `detected_agents: ["claude"]`
- **THEN** `claude` is checked by default and the page can advance

#### Scenario: Advancing is blocked with no agents selected

- **WHEN** the user unchecks every agent on the UninstallAgents page
- **THEN** the page reports it cannot advance and the primary footer action is disabled

### Requirement: Uninstall mode and strategy selection

The UninstallMode page SHALL present lite, full, and custom as a single-select (radio) group, and SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected mode. The UninstallStrategy page SHALL present the strategies from the uninstall options as a single-select (radio) group, preselecting the strategy marked default (`targeted`), and SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected strategy. When the selected strategy is `restore` (a strategy whose `requires_manifest` is true), the page SHALL reveal a backup selector populated from `ListBackupsAsync` (rendering each backup by its `display_label`), and SHALL report it cannot advance until a backup is selected. When the strategy does not require a manifest, no backup selector is shown and the page can advance immediately.

#### Scenario: Targeted strategy needs no backup and can advance

- **WHEN** the UninstallStrategy page has `targeted` (requires_manifest false) selected
- **THEN** no backup selector is shown and the page can advance

#### Scenario: Restore reveals the backup selector and gates on a selection

- **WHEN** the user selects the `restore` strategy (requires_manifest true)
- **THEN** a backup selector fed by `ListBackupsAsync` appears and the page reports it cannot advance until a backup is chosen; once a backup is selected the page can advance

#### Scenario: Detail panel reflects the selected uninstall mode

- **WHEN** the user selects a different uninstall mode
- **THEN** the detail panel updates to that mode's long description, falling back to its short description when no long description is present

#### Scenario: Detail panel reflects the selected uninstall strategy

- **WHEN** the user selects a different uninstall strategy
- **THEN** the detail panel updates to that strategy's long description, falling back to its short description when no long description is present

### Requirement: Uninstall confirm warns and triggers the uninstall stream

The UninstallConfirm page SHALL present a read-only summary of the selection (agents, mode, strategy, and the chosen backup when restoring) together with a prominent warning matching the TUI ("This will modify your agent configuration."). Advancing from Confirm SHALL start the uninstall by invoking the engine client's uninstall stream with the selected agents, mode, and strategy — and, for the `restore` strategy, the chosen backup's manifest path (threaded through the client's uninstall call so the engine receives `--restore-manifest <path>`). Progress and completion SHALL be rendered by the shared generic progress and complete pages.

#### Scenario: Confirm shows the warning and summary

- **WHEN** the UninstallConfirm page is shown with agents, mode, and strategy selected
- **THEN** it displays the agents, mode, and strategy summary and the destructive warning text

#### Scenario: Restore uninstall carries the selected backup's manifest path

- **WHEN** the user confirms an uninstall whose strategy is `restore` with a backup selected
- **THEN** the uninstall stream is started carrying that backup's manifest path (so the engine is invoked with `--restore-manifest`), along with the selected agents, mode, and strategy

### Requirement: Starter catalog is a single-select list

The StarterCatalog page SHALL list the starters from `ListStartersAsync` as a single-select list, showing each starter's name, description, and harness count. The page SHALL report it cannot advance until a starter is selected. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected starter.

#### Scenario: Selecting a starter enables advancing

- **WHEN** the StarterCatalog page is shown with a non-empty starter list and the user selects one entry
- **THEN** that starter becomes the selection and the page can advance; with nothing selected the page cannot advance

#### Scenario: Detail panel reflects the selected starter

- **WHEN** the user selects a different starter
- **THEN** the detail panel updates to that starter's long description, falling back to its short description when no long description is present

### Requirement: Starter target uses an injectable folder picker and default agents

The StarterTarget page SHALL obtain the target project directory through an injectable `IFolderPicker` abstraction (implemented over the WPF `OpenFolderDialog` in production, faked in tests), and SHALL present the agents as checkboxes defaulting to claude and opencode. The page SHALL report it cannot advance until a non-empty project directory has been chosen and at least one agent is selected. The chosen path MUST be carried as a single value (never split) so paths containing spaces survive.

#### Scenario: Picking a folder via the injected picker enables advancing

- **WHEN** the user invokes the folder picker on the StarterTarget page and the injected `IFolderPicker` returns a directory
- **THEN** that directory becomes the target, the default agents (claude, opencode) are checked, and the page can advance

#### Scenario: No chosen directory blocks advancing

- **WHEN** the StarterTarget page has no directory chosen (the picker was cancelled or never invoked)
- **THEN** the page reports it cannot advance

### Requirement: Starter review triggers the starter install stream

The Starter Review page SHALL present a read-only summary of the chosen starter, target directory, and agents, and the shell's primary action SHALL read "Install". Advancing from Review SHALL start the install by invoking the engine client's starter install stream with the starter id, project path, and agents. Progress and completion SHALL be rendered by the shared generic progress and complete pages.

#### Scenario: Review starts the starter install with the full selection

- **WHEN** the user confirms on the Starter Review page with a starter, project path, and agents chosen
- **THEN** the starter install stream is started carrying the starter id, the project path (as a single value), and the selected agents

### Requirement: Backups page lists backups and re-lists after each action

The Backups page SHALL be a single page (not part of a wizard flow) that lists backups from `ListBackupsAsync`, rendering each by its `display_label`, and offering per-item Restore, Delete, and Rename actions. After any action completes (success or failure), the page SHALL re-list from `ListBackupsAsync` so the displayed state reflects the store, and SHALL surface a short result message. An empty store SHALL render an empty-state message and no actions.

#### Scenario: Backups are listed by display label

- **WHEN** the Backups page loads against a store with backups
- **THEN** each backup is shown by its `display_label` with Restore, Delete, and Rename actions available

#### Scenario: The list is refreshed after an action

- **WHEN** a Restore, Delete, or Rename action completes
- **THEN** the page re-invokes `ListBackupsAsync` and shows a result message reflecting the outcome

#### Scenario: Empty store shows an empty state

- **WHEN** the Backups page loads against an empty store
- **THEN** it shows an empty-state message and offers no per-item actions

### Requirement: Backups actions confirm inline with TUI-matching texts

Each Backups action SHALL require an inline confirmation before invoking the engine client's backup action. Restore SHALL show "This will OVERWRITE your current configuration."; Delete SHALL show "This will PERMANENTLY DELETE this backup."; Rename SHALL capture a new description via a text input. Confirming SHALL invoke `RunBackupActionAsync` with the matching action (`restore`, `delete`, `rename`) and the target backup's id (and, for rename, the entered description); cancelling SHALL return to the list without invoking the engine. Only one action confirmation SHALL be active at a time.

#### Scenario: Restore confirmation matches the TUI text and invokes restore

- **WHEN** the user triggers Restore on a backup and confirms
- **THEN** the confirmation text reads "This will OVERWRITE your current configuration." and `RunBackupActionAsync` is invoked with action `restore` and that backup's id

#### Scenario: Delete confirmation matches the TUI text and invokes delete

- **WHEN** the user triggers Delete on a backup and confirms
- **THEN** the confirmation text reads "This will PERMANENTLY DELETE this backup." and `RunBackupActionAsync` is invoked with action `delete` and that backup's id

#### Scenario: Rename captures a description and invokes rename

- **WHEN** the user triggers Rename, types a new description, and confirms
- **THEN** `RunBackupActionAsync` is invoked with action `rename`, that backup's id, and the entered description

#### Scenario: Cancelling an action invokes nothing

- **WHEN** the user triggers any action and cancels the confirmation
- **THEN** no engine call is made and the page returns to the list view

### Requirement: Generic progress and complete pages are reused for uninstall and starter streams

The uninstall and starter flows SHALL render their live progress through the same generic `ProgressPageViewModel` used by the Install flow (consuming an `IAsyncEnumerable<InstallProgressSnapshot>`), and SHALL render their terminal outcome through the shared `CompletePageViewModel`. No uninstall- or starter-specific progress renderer SHALL be introduced.

#### Scenario: Uninstall progress reuses the generic progress page

- **WHEN** an uninstall or starter install is running
- **THEN** its event stream is rendered by the shared `ProgressPageViewModel` and its terminal state by the shared `CompletePageViewModel`, identical to the Install flow's progress rendering

### Requirement: Wizard opens on the language selection page

The wizard SHALL present a Language page (`WizardPageId.Language`) as its first page, before the Hub. The page SHALL show a fixed bilingual title ("Select your language / Elegí tu idioma") and an extensible single-select list of language choices (English, Español). The initial selection SHALL follow the resolved preselection (persisted preference > OS UI culture > English). `WizardFlow` SHALL advance Language → Hub and go Back Hub → Language. On advancing from the Language page, the `ShellViewModel` SHALL set its `_language`, persist it via `LanguagePreference.Save`, set the engine client's `Language`, and — when the language changed from the previously active one — invalidate the cached `_session`, `_uninstallOptions`, and `_starters` so subsequent flows re-fetch engine data in the new language. The `ShellViewModel` `CanGoBack` SHALL exclude the Language page (not the Hub).

#### Scenario: Language is the first page and advances to the Hub

- **WHEN** the wizard starts
- **THEN** the current page is the Language page with the preselected language chosen, and advancing goes to the Hub

#### Scenario: Advancing from Language sets and persists the language and configures the client

- **WHEN** the user selects a language on the Language page and advances
- **THEN** the shell's active language is set to that choice, `LanguagePreference.Save` is invoked with it, and the engine client's `Language` is set to the same value

#### Scenario: Changing the language invalidates cached engine data

- **WHEN** the user goes Back to the Language page, selects a different language than before, and advances
- **THEN** the previously cached `_session`, `_uninstallOptions`, and `_starters` are invalidated so the next flow re-fetches them in the new language

#### Scenario: Re-selecting the same language does not invalidate caches

- **WHEN** the user returns to the Language page and advances again with the same language selected
- **THEN** the cached engine data is retained (no needless re-fetch)

#### Scenario: Auto-install traverses the preselected Language page

- **WHEN** the bootstrapper runs an automated/auto-install loop
- **THEN** it traverses the preselected Language page without stalling, using the resolved (or `ACTIVE_STACK_UI_LANG`-overridden) language

### Requirement: Shell and pages render their literals from the language string table

The shell footer labels, every page title/subtitle, the XAML section headers, the backups and uninstall warnings, the complete-state label, and the progress step messages SHALL be rendered from `UiStrings` scoped to the shell's active language rather than hardcoded English. Specifically: page titles SHALL be produced via `base(UiStrings.Get(lang, …))`; `MainWindow.xaml` Back/Cancel SHALL bind to `ShellViewModel.BackLabel`/`CancelLabel`; the primary action label (Next / Install / Uninstall overrides) SHALL come from `UiStrings`; the `PageTemplates.xaml` section headers SHALL bind to view-model properties (so they are unit-testable) fed from `UiStrings`; `ProgressPageViewModel.ToFriendlyMessage` SHALL source its own strings from `UiStrings` while engine-provided `phase_started`/`install_finished` messages (already localized by the engine) are shown verbatim; and `CompletePageViewModel` SHALL expose a localized `StateLabel`. The backups Restore/Delete warnings and the uninstall modification warning SHALL come from `UiStrings`.

#### Scenario: Footer and page titles render in the active language

- **WHEN** the active language is Spanish
- **THEN** the footer primary/Back/Cancel labels and the current page's title/subtitle render their Spanish `UiStrings` values, and when the active language is English they render the English values

#### Scenario: Progress step messages render in the active language

- **WHEN** the active language is Spanish and the progress page maps an install step to a friendly message it owns
- **THEN** the message renders in Spanish (e.g. "Instalando OpenSpec."), while a `phase_started`/`install_finished` message supplied by the engine is shown verbatim as received

#### Scenario: Section headers and warnings are testable view-model properties

- **WHEN** the backups page surfaces its Restore/Delete confirmation warnings, or a page surfaces a `PageTemplates` section header
- **THEN** those texts come from view-model properties fed by `UiStrings` (assertable without rendering XAML) and reflect the active language

### Requirement: Stream failures surface a terminal Error page, never a silent exit

When the shell consumes an install, uninstall, or starter progress stream (`ShellViewModel.RunStreamAndAdvanceToCompleteAsync`), any exception thrown while consuming that stream — a non-zero engine exit, a parse failure, or any other exception — SHALL be caught and converted into navigation to the shared `CompletePageViewModel` in the `CompleteState.Error` terminal state, carrying the failure's message. The exception MUST NOT propagate past the shell (so it can never reach the WPF dispatcher and terminate the process silently). This targeted catch is the primary mechanism; a global unhandled-exception handler is a secondary net and MUST NOT be relied on as the primary path.

#### Scenario: A stream that throws mid-enumeration lands on the Error page

- **WHEN** the current page is a stream-trigger page whose `StartStream` yields an `IAsyncEnumerable` that throws partway through enumeration, and the shell advances
- **THEN** `AdvanceAsync` does not propagate the exception, the current page becomes a `CompletePageViewModel` whose `State` is `CompleteState.Error`, and the caught failure's message is surfaced on that page

#### Scenario: A non-zero engine exit reaches the Error page instead of closing the window

- **WHEN** an install/uninstall/starter stream fails because the engine process exits non-zero (surfaced as a thrown exception from the streaming client)
- **THEN** the wizard shows the terminal Error page rather than the process terminating with no user-facing signal

### Requirement: List pages wrap content instead of scrolling horizontally

Every `ListBox` in the wizard page templates (`Themes/PageTemplates.xaml`) SHALL disable its horizontal scrollbar (`ScrollViewer.HorizontalScrollBarVisibility="Disabled"`) so long item content wraps within the available width rather than exposing a horizontal scrollbar. This applies uniformly to all list-bearing pages (Language, Hub, InstallType, Components, Permissions, UninstallMode, UninstallStrategy strategy and backup lists, and StarterCatalog), not only the page where the defect was first observed.

#### Scenario: A long list item wraps rather than scrolling sideways

- **WHEN** a list page (e.g. Permissions) renders an item whose text is wider than the list's available width
- **THEN** the item's text wraps within the list and no horizontal scrollbar is shown

### Requirement: Complete page's primary action finishes and closes the wizard

On the terminal Complete page, the shell's primary footer action SHALL be a Finish/Close action, not a navigation "Next". Its label SHALL be the localized `shell.finish` string ("Finish" / "Finalizar") — supplied for both `en` and `es` in `UiStrings` — rather than the generic `shell.next` label. Invoking the primary action while the current page is the `CompletePageViewModel` SHALL cause the wizard window to close, and SHALL NOT route into any flow's navigation. Concretely, `ShellViewModel` SHALL expose a `CloseRequested` event and, in `AdvanceAsync`, SHALL raise `CloseRequested` and return when the current page is the `CompletePageViewModel` — before reaching the operation switch that rebuilds pages via `CreatePage` (which has no Complete case and would otherwise throw). The host (`MainWindow`) SHALL subscribe to `CloseRequested` and close the window; `ShellViewModel` MUST NOT take a dependency on the WPF `Window`. Advancing from the Complete page MUST NOT throw and MUST NOT be silently swallowed by the global dispatcher handler. `CanGoBack` SHALL remain false on the Complete page (Back stays disabled), and the Complete page's `CanAdvance` SHALL remain true so the Finish action is enabled.

#### Scenario: Primary action on Complete raises CloseRequested instead of navigating

- **WHEN** the wizard has reached the `CompletePageViewModel` (after an install/uninstall/starter stream) and the shell's primary action is invoked
- **THEN** `ShellViewModel` raises its `CloseRequested` event, `AdvanceAsync` returns without throwing, and the current page remains the `CompletePageViewModel` (no navigation into any flow occurs)

#### Scenario: Primary label on Complete is the localized Finish label

- **WHEN** the current page is the `CompletePageViewModel` with the active language English (and, separately, Spanish)
- **THEN** the shell's `PrimaryLabel` is "Finish" in English and "Finalizar" in Spanish (the `shell.finish` string), not "Next"/"Siguiente"

#### Scenario: Advancing on Complete no longer throws through the dispatcher

- **WHEN** the primary action is invoked on the `CompletePageViewModel` while the active operation is still Install
- **THEN** no exception is thrown (the former `CreatePage` `InvalidOperationException` path is bypassed), so nothing reaches the global `DispatcherUnhandledException` net to be silently swallowed
