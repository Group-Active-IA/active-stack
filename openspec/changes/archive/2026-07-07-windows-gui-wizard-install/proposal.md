## Why

Today the Windows GUI installer (`installer/windows/ui/`, WPF + WiX Burn) is a single confusing dashboard (`MainWindow.xaml`) where every option lives on one screen: agents are single-select (a `ComboBox`, while the TUI and CLI are multi-select), Custom mode shows components read-only with no checkboxes and never sends `--custom`, and there is no permission-tier configuration at all. This is THE user-facing gap of the whole 6-change effort: C1–C3 landed the Go engine contract and the C# Core parsers/builders, but no user can reach them through a coherent flow.

C4 rewrites the WPF UI as a traditional multi-page wizard (next → next → finish) that mirrors the terminal installer (TUI) exactly, consuming the extended `IInstallerEngineClient` and `InstallArguments` builders that C3 already shipped and tested (see `openspec/changes/windows-gui-core-parsers/specs/windows-gui-engine-client/spec.md`). It materializes the approved but never-implemented 8-screen design in `docs/superpowers/specs/2026-07-02-windows-native-installer-design.md:226-346`, with documented deviations.

## What Changes

- **BREAKING (internal UI contract)**: `MainWindowViewModel` (the single dashboard VM) is replaced by a `ShellViewModel` plus per-page view-models. Its progress-rendering logic is *moved* (not rewritten) and its tests migrated before it is deleted.
- **New navigation substrate** under `installer/windows/ui/ActiveStack.Bootstrapper.Host/Navigation/`: `IWizardPage`, `ShellViewModel`, and a `WizardFlow` type exposing **pure** `NextPage`/`PreviousPage` functions. Navigation is `ContentControl` + `DataTemplate`-by-ViewModel-type (NOT `Frame`/`NavigationService`); wizard Back is *logical*, not chronological.
- **`MainWindow.xaml` rewritten** as a thin shell: header (per-page title/subtitle) + `<ContentControl Content="{Binding CurrentPage}"/>` + footer (`Back` / `Next|Install` / `Cancel`). Window becomes 760×560, `ResizeMode=CanResize`, no longer maximized (was 1320×820 maximized).
- **Install flow pages** under `Pages/` mirroring the TUI (`internal/tui/model.go:278-313`): Hub → Assistants (multi-select checkboxes, min 1, prechecked from detect) → InstallType (cards lite/full/custom, default full) → [custom → Components: checkboxes, `permissions` harness checked+disabled with engine copy, recommended prechecked] → [tier-capable selected agents → Permissions: 3-tier radio from `permission_tiers`, default balanceado, warning shown on bypass] → Review → Installing (`ProgressPage`) → Complete (success/degraded/rollback/error).
- **Hub page** with entries: Install, Starters, Manage Backups, Uninstall, and a disabled "Update Stack — Coming soon" (TUI parity). Starters / Backups / Uninstall entries are present but disabled with an explanatory tooltip in C4 (C5 wires them) — decision recorded in design.md.
- **`Themes/PageTemplates.xaml`**: `DataTemplate`s keyed by each page ViewModel type.
- **Tier-capability** is evaluated from `tier_capable_agents` (in the options response) against the *selected* agents, with no re-query on selection change.

## Capabilities

### New Capabilities
- `windows-gui-wizard`: The WPF wizard shell and Install-flow behavior — logical page navigation (pure Next/Previous functions), the ordered Install page sequence with its TUI-parity gates (min-1 agents, custom→Components, tier-capable→Permissions, forced-and-locked permissions harness), the Hub entry set with the disabled Update/placeholder entries, and progress/complete rendering migrated from the old dashboard VM.

### Modified Capabilities
<!-- No existing spec's REQUIREMENTS change. windows-gui-engine-client (C3, still in-flight change) and the windows-install/hub contracts are consumed as-is, not modified. -->

## Impact

- **Rewritten**: `installer/windows/ui/ActiveStack.Bootstrapper.Host/MainWindow.xaml` + `MainWindow.xaml.cs`.
- **New**: `Navigation/{IWizardPage,ShellViewModel,WizardFlow}.cs`; `Pages/HubPageViewModel.cs`, `Pages/Install/{AssistantsPageViewModel,InstallTypePageViewModel,ComponentsPageViewModel,PermissionsPageViewModel,ReviewPageViewModel}.cs`, `Pages/ProgressPageViewModel.cs`, `Pages/CompletePageViewModel.cs`; `Themes/PageTemplates.xaml`.
- **Moved**: `ApplySnapshot`/`CalculateProgress`/`ToFriendlyMessage`/`ToDisplayLabel` from `MainWindowViewModel.cs:314-425` into `ProgressPageViewModel`, with their tests migrated from `MainWindowViewModelTests.cs`.
- **Deleted (after migration)**: `MainWindowViewModel.cs` and the obsolete parts of `MainWindowViewModelTests.cs`.
- **Consumed (unchanged)**: `IInstallerEngineClient` (extended in C3), `InstallArguments`, `InstallerSessionState`/`PermissionTierChoice`, the parsers, and `InstallProgressSnapshot`.
- **Not touched**: `internal/install`, `internal/backup`, `internal/uninstall`, the Go engine, Burn wiring (`ActiveStackBootstrapperApplication`, `Process` spawn), `build.ps1`, `Bundle.wxs`.
- **Note**: `IInstallerEngineClient.RunInstallAsync` currently accepts only `(assistantId, installTypeId)`; C4 must extend it to carry the full selection (agents CSV, mode, custom ids, tier) via the C3 `InstallArguments.BuildInstallArgs` builder — this is additive wiring, addressed in design.md.
- **Tests**: new `Host.Tests/WizardFlowTests.cs` (pure-function flow), page-VM tests with a fake `IInstallerEngineClient`, `ProgressPageViewModelTests.cs` migrated from `MainWindowViewModelTests.cs`.
