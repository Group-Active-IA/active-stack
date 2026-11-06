## Why

C4 landed the Windows GUI wizard shell and the full Install flow, but left the other three TUI destinations — Uninstall, Starters, Manage Backups — present on the Hub yet **disabled with a tooltip** (C4 decision D8). The Go engine contract (C2) and the C# client/parsers (C3) for all three already exist and are archived. C5 is the user-facing wiring that flips those Hub entries on and delivers each flow with TUI parity, closing the gap between the GUI and the terminal installer.

## What Changes

- **Hub**: flip `IsEnabled` on the Install, Starters, Manage Backups, and Uninstall entries so each navigates to its flow. "Update Stack — Coming soon" stays permanently disabled (TUI parity).
- **Uninstall wizard** (mirrors `internal/tui/uninstall_screen.go`): Hub → UninstallAgents (multi-select checkboxes from `uninstall-options.detected_agents`, min 1) → UninstallMode (lite/full/custom radio) → UninstallStrategy (targeted/restore radio; `restore` reveals a backup selector fed by `ListBackupsAsync` and gated by `requires_manifest`) → UninstallConfirm (red warning + summary) → Progress (`RunUninstallAsync`) → Complete.
- **Starters flow**: Hub → StarterCatalog (single-select list with name/description/harness count from `ListStartersAsync`) → StarterTarget (folder picker behind an injectable `IFolderPicker` + agent checkboxes defaulting to claude+opencode) → Review → Progress (`RunStarterInstallAsync`) → Complete.
- **Backups page** (single page, not a wizard): list by `display_label` with per-item Restore / Delete / Rename actions using inline confirmations whose texts match the TUI ("This will OVERWRITE your current configuration.", "This will PERMANENTLY DELETE this backup."), rename via text input, re-listing after every action via `ListBackupsAsync`.
- **Navigation**: extend the flow layer with the new page identities and pure per-operation flow functions (Install / Uninstall / Starter), and generalize the shell so the Hub dispatches to the selected flow instead of always starting Install. The generic `ProgressPageViewModel` and `CompletePageViewModel` from C4 are reused for the uninstall and starter streams.
- **No engine, `internal/backup`, `internal/uninstall`, or C3 contract changes** — C5 consumes the existing client (`IInstallerEngineClient` already declares every method the flows need) and parsers as-is.

## Capabilities

### New Capabilities
<!-- None. C5 extends the existing wizard capability rather than introducing a new one. -->

### Modified Capabilities
- `windows-gui-wizard`: the Hub requirement changes from "all destinations disabled except Install" to "Install, Starters, Manage Backups, and Uninstall enabled; Update Stack disabled". New requirements are ADDED for the Uninstall wizard flow, the Starters flow (with an injectable folder picker), the Backups management page and its confirmation state machine, and the per-operation pure flow functions. (Base spec is C4's `windows-gui-wizard` delta, archived ahead of C5 per the C1–C6 plan order.)

## Impact

- **Code (all in the `Host` WPF project — governance MEDIUM, UI only)**: `Navigation/` (new `WizardPageId` values; per-operation flow functions; `ShellViewModel` Hub dispatch + uninstall/starter selection models), `Pages/Uninstall/*`, `Pages/Starters/*`, `Pages/Backups/*`, `Themes/PageTemplates.xaml` (one `DataTemplate` per new page VM), `HubPageViewModel` (enablement), and a new `IFolderPicker` abstraction with a WPF `OpenFolderDialog` implementation.
- **Consumes unchanged**: the archived `windows-gui-engine-client` (client methods + parsers) and `windows-hub-contract` (Go `windows starters|backups|uninstall-options`) capabilities.
- **Tests (`Host.Tests`)**: uninstall flow transition tests (incl. the restore-requires-a-selected-backup gate), page VM tests with a fake `IInstallerEngineClient` + fake `IFolderPicker`, backups confirmation state-machine tests, and Hub enablement tests. TDD-ordered.
- **Not touched**: the Go engine, `internal/backup`, `internal/uninstall`, `internal/install`, Burn/`Process` wiring, `build.ps1`, `Bundle.wxs`.
