## MODIFIED Requirements

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

### Requirement: Install type defaults to full with end-user labels

The InstallType page SHALL present lite, full, and custom as selectable cards using end-user labels (Quick / Complete / Custom) with the internal mode mapping behind them, defaulting the selection to `full`. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected install type.

#### Scenario: Full is preselected

- **WHEN** the InstallType page is first shown
- **THEN** the `full` (Complete) card is selected by default and the page can advance

#### Scenario: Detail panel reflects the selected install type

- **WHEN** the user selects a different install type card
- **THEN** the detail panel updates to that type's long description, falling back to its short description when no long description is present

### Requirement: Permissions page offers the engine's tiers with a default and bypass warning

When shown, the Permissions page SHALL render the tiers from the session's `permission_tiers` as a single-select (radio) group, preselecting the tier marked default (balanceado), and SHALL surface the tier's warning text when a tier carrying a warning (bypass) is selected. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected tier.

#### Scenario: Balanceado is preselected and bypass shows a warning

- **WHEN** the Permissions page is shown with three tiers where balanceado is default and bypass carries a warning
- **THEN** balanceado is selected initially, and selecting bypass surfaces its warning text

#### Scenario: Detail panel reflects the selected tier

- **WHEN** the user selects a different permission tier
- **THEN** the detail panel updates to that tier's long description, falling back to its short description when no long description is present

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

### Requirement: Starter catalog is a single-select list

The StarterCatalog page SHALL list the starters from `ListStartersAsync` as a single-select list, showing each starter's name, description, and harness count. The page SHALL report it cannot advance until a starter is selected. The page SHALL render a detail panel showing the long description (falling back to the short description when no long description is present) of the currently selected starter.

#### Scenario: Selecting a starter enables advancing

- **WHEN** the StarterCatalog page is shown with a non-empty starter list and the user selects one entry
- **THEN** that starter becomes the selection and the page can advance; with nothing selected the page cannot advance

#### Scenario: Detail panel reflects the selected starter

- **WHEN** the user selects a different starter
- **THEN** the detail panel updates to that starter's long description, falling back to its short description when no long description is present
