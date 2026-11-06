## Why

The Windows GUI wizard shows only a one-line short description per option, so users cannot tell what each Component / install type / permission tier / starter / uninstall mode / uninstall strategy actually does before choosing it. Changes L1–L4 already built the whole substrate for a fix: the Go engine now emits `long_description` for every option (L1), the catalog carries bilingual long descriptions (L2), and the GUI has a working language system with a pure `UiStrings` table threaded through every page view-model (L4). This change (L5, the last of the plan) is the payoff — it wires the C# side so the wizard renders those long descriptions in a side detail panel.

## What Changes

- Core `record` types gain an additive `string LongDescription = ""`: `ComponentChoice`, `InstallTypeChoice` (which backs both InstallType and UninstallMode), `PermissionTierChoice`, `StarterChoice`, `UninstallStrategyChoice`. The three parsers (`InstallerSessionStateBuilder`, `StarterCatalogParser`, `UninstallOptionsParser`) read the `long_description` JSON field with default `""` (pure consumption — the engine already emits it).
- Each of the 6 affected page view-models exposes derived `DetailTitle` / `DetailBody` reflecting the currently selected/focused item, re-raised on selection change; `DetailBody` falls back to the item's short `Description` when `LongDescription` is empty (never a blank panel).
- A single shared detail-panel resource in `Themes/PageTemplates.xaml` (defined once, referenced by all 6 page templates — not duplicated) laid out as a right-side column (Grid `3*`/`2*`, `MinWidth≈230`, `ScrollViewer`).
- The Components page gains a row-level `SelectedComponentId` independent of the per-row `IsSelected` checkbox: its template moves from `ItemsControl` to `ListBox` with the `CheckBox` inside the item template, so row selection feeds the panel while the checkbox still only toggles inclusion.
- One new localized `UiStrings` key for the panel header (e.g. `detail.header`), added to both the `En` and `Es` tables keeping parity.

No Go changes. No new capabilities. Not a breaking change (all additions are additive; JSON parsing tolerates absent `long_description`).

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `windows-gui-wizard`: The Components, InstallType, Permissions, StarterCatalog, and Uninstall-mode-and-strategy requirements each gain a detail-panel behavior — the wizard SHALL render the selected/focused item's long description (falling back to its short description) in a shared side panel, and the Components page SHALL track a row selection independent of the inclusion checkbox.

## Impact

- **C# (Core)**: `InstallerSessionStateBuilder.cs`, `StarterCatalogParser.cs`, `UninstallOptionsParser.cs` (records + JSON DTOs + parser mapping); one new `UiStrings` key in `Localization/UiStrings.cs`.
- **C# (Host)**: `Pages/Install/{ComponentsPageViewModel,ComponentOption,InstallTypePageViewModel,PermissionsPageViewModel}.cs`, `Pages/Starters/StarterCatalogPageViewModel.cs`, `Pages/Uninstall/{UninstallModePageViewModel,UninstallStrategyPageViewModel}.cs`, and `Themes/PageTemplates.xaml`.
- **Tests**: `ActiveStack.Bootstrapper.Core.Tests` (parser long_description defaulting) and `ActiveStack.Bootstrapper.Host.Tests` (per-page `DetailBody`/fallback/notification + Components selection-vs-checkbox independence).
- **Governance**: MEDIO (`tui`/view-model layer; no `backup`/`filemerge`/`pipeline` touched).
- Depends on L2 (catalog long descriptions in JSON) and L4 (`UiStrings` + `_language` threading), both landed.
