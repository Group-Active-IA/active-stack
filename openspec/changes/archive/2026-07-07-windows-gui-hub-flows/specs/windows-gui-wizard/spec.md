## MODIFIED Requirements

### Requirement: Hub surfaces all destinations with unimplemented ones disabled

The Hub page SHALL present the entries Install, Starters, Manage Backups, Uninstall, and a disabled "Update Stack — Coming soon". Install, Starters, Manage Backups, and Uninstall SHALL each be enabled and, when chosen, SHALL start their respective flow (Install wizard, Starter flow, Backups page, or Uninstall wizard). Update Stack SHALL remain disabled with a "Coming soon" tooltip, matching the TUI's permanent placeholder treatment. Choosing a disabled entry SHALL remain a no-op.

#### Scenario: Every implemented destination is actionable on the Hub

- **WHEN** the Hub page is shown
- **THEN** Install, Starters, Manage Backups, and Uninstall are each enabled and can advance into their flow, while Update Stack is disabled with a tooltip and does not navigate

#### Scenario: Selecting a Hub entry starts that entry's flow

- **WHEN** the user selects the Uninstall entry (or Starters, or Manage Backups) and advances
- **THEN** the shell enters the corresponding flow's first page (Uninstall → UninstallAgents, Starters → StarterCatalog, Manage Backups → the Backups page) rather than the Install flow

## ADDED Requirements

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

The UninstallMode page SHALL present lite, full, and custom as a single-select (radio) group. The UninstallStrategy page SHALL present the strategies from the uninstall options as a single-select (radio) group, preselecting the strategy marked default (`targeted`). When the selected strategy is `restore` (a strategy whose `requires_manifest` is true), the page SHALL reveal a backup selector populated from `ListBackupsAsync` (rendering each backup by its `display_label`), and SHALL report it cannot advance until a backup is selected. When the strategy does not require a manifest, no backup selector is shown and the page can advance immediately.

#### Scenario: Targeted strategy needs no backup and can advance

- **WHEN** the UninstallStrategy page has `targeted` (requires_manifest false) selected
- **THEN** no backup selector is shown and the page can advance

#### Scenario: Restore reveals the backup selector and gates on a selection

- **WHEN** the user selects the `restore` strategy (requires_manifest true)
- **THEN** a backup selector fed by `ListBackupsAsync` appears and the page reports it cannot advance until a backup is chosen; once a backup is selected the page can advance

### Requirement: Uninstall confirm warns and triggers the uninstall stream

The UninstallConfirm page SHALL present a read-only summary of the selection (agents, mode, strategy, and the chosen backup when restoring) together with a prominent warning matching the TUI ("This will modify your agent configuration."). Advancing from Confirm SHALL start the uninstall by invoking the engine client's uninstall stream with the selected agents, mode, and strategy — and, for the `restore` strategy, the chosen backup's manifest path (threaded through the client's uninstall call so the engine receives `--restore-manifest <path>`). Progress and completion SHALL be rendered by the shared generic progress and complete pages.

#### Scenario: Confirm shows the warning and summary

- **WHEN** the UninstallConfirm page is shown with agents, mode, and strategy selected
- **THEN** it displays the agents, mode, and strategy summary and the destructive warning text

#### Scenario: Restore uninstall carries the selected backup's manifest path

- **WHEN** the user confirms an uninstall whose strategy is `restore` with a backup selected
- **THEN** the uninstall stream is started carrying that backup's manifest path (so the engine is invoked with `--restore-manifest`), along with the selected agents, mode, and strategy

### Requirement: Starter catalog is a single-select list

The StarterCatalog page SHALL list the starters from `ListStartersAsync` as a single-select list, showing each starter's name, description, and harness count. The page SHALL report it cannot advance until a starter is selected.

#### Scenario: Selecting a starter enables advancing

- **WHEN** the StarterCatalog page is shown with a non-empty starter list and the user selects one entry
- **THEN** that starter becomes the selection and the page can advance; with nothing selected the page cannot advance

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
