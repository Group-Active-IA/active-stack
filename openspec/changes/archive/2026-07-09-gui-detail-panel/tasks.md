## 1. Core: `LongDescription` on records + parsers

- [x] 1.1 Add failing Core.Tests cases: `InstallerSessionStateBuilder.BuildFromJson` maps `long_description` into `InstallTypeChoice`/`ComponentChoice`/`PermissionTierChoice`, and defaults to `""` when the field is absent from the JSON.
- [x] 1.2 Add failing Core.Tests cases: `StarterCatalogParser.BuildFromJson` maps `long_description` into `StarterChoice`, defaulting to `""` when absent.
- [x] 1.3 Add failing Core.Tests cases: `UninstallOptionsParser.BuildFromJson` maps `long_description` into the `modes` array's `InstallTypeChoice` entries and into `UninstallStrategyChoice`, defaulting to `""` when absent.
- [x] 1.4 Implement: add `string LongDescription = ""` to `InstallTypeChoice`, `ComponentChoice`, `PermissionTierChoice`, `StarterChoice`, `UninstallStrategyChoice` records; add `[JsonPropertyName("long_description")]` to each internal JSON DTO in `InstallerSessionStateBuilder.cs`, `StarterCatalogParser.cs`, `UninstallOptionsParser.cs`; wire the mapping in each `Select` projection. Run Core.Tests until 1.1-1.3 pass.
- [x] 1.5 Triangulate: add a second fixture per parser with a different id/long_description pair to rule out a hardcoded return; confirm both pass.

## 2. Core: localized panel header string

- [x] 2.1 Add a failing Core.Tests assertion that `UiStrings.EnglishKeys`/`UiStrings.SpanishKeys` both contain a new `detail.header` key (parity test already exists — extend it or confirm it already catches a missing key on one side).
- [x] 2.2 Implement: add `["detail.header"] = "Details"` to the `En` table and `["detail.header"] = "Detalle"` to the `Es` table in `Core/Localization/UiStrings.cs`. Run Core.Tests until green.

## 3. Host: thread `DetailHeader` through `WizardPageViewModelBase`

- [x] 3.1 Add a failing Host.Tests case: constructing any existing page view-model with `lang: "es"` exposes `DetailHeader == "Detalle"`, and with `lang: "en"` (or default) exposes `DetailHeader == "Details"`.
- [x] 3.2 Implement: change `WizardPageViewModelBase` constructor to `(string title, string subtitle, string lang)`, set `DetailHeader = UiStrings.Get(lang, "detail.header")`; update every derived page-VM's `base(...)` call to pass `lang` (mechanical, compiler-enforced). Run full Host.Tests suite — fix any call site the compiler flags. Confirm the existing suite (all previously-passing tests) is still green (safety net).

## 4. Host: InstallType detail panel

- [x] 4.1 Add failing Host.Tests: `InstallTypePageViewModel.DetailBody` returns the selected choice's `LongDescription` when non-empty; falls back to `Description` when `LongDescription` is empty; changing `SelectedId` raises `PropertyChanged` for `DetailBody`/`DetailTitle`.
- [x] 4.2 Implement `DetailTitle`/`DetailBody` computed properties on `InstallTypePageViewModel`, raised from the `SelectedId` setter. Run tests until green.

## 5. Host: Permissions detail panel

- [x] 5.1 Add failing Host.Tests: `PermissionsPageViewModel.DetailBody` reflects the selected tier's `LongDescription`/fallback to `Description`; changing `SelectedTierId` raises `PropertyChanged` for `DetailBody`/`DetailTitle`.
- [x] 5.2 Implement `DetailTitle`/`DetailBody` on `PermissionsPageViewModel`, raised from `SelectedTierId`. Run tests until green.

## 6. Host: UninstallMode detail panel

- [x] 6.1 Add failing Host.Tests: `UninstallModePageViewModel.DetailBody` reflects the selected mode's `LongDescription`/fallback; changing `SelectedId` raises `PropertyChanged` for `DetailBody`/`DetailTitle`.
- [x] 6.2 Implement `DetailTitle`/`DetailBody` on `UninstallModePageViewModel`, raised from `SelectedId`. Run tests until green.

## 7. Host: UninstallStrategy detail panel

- [x] 7.1 Add failing Host.Tests: `UninstallStrategyPageViewModel.DetailBody` reflects the selected strategy's `LongDescription`/fallback; changing strategy (via `SelectedStrategyId`/`SelectStrategyAsync`) raises `PropertyChanged` for `DetailBody`/`DetailTitle`.
- [x] 7.2 Implement `DetailTitle`/`DetailBody` on `UninstallStrategyPageViewModel`, raised from `SelectStrategyAsync`. Run tests until green.

## 8. Host: StarterCatalog detail panel

- [x] 8.1 Add failing Host.Tests: `StarterCatalogPageViewModel.DetailBody` reflects the selected starter's `LongDescription`/fallback to `Description`; `DetailTitle` reflects the starter's `Name`; changing `SelectedStarterId` raises `PropertyChanged` for both.
- [x] 8.2 Implement `DetailTitle`/`DetailBody` on `StarterCatalogPageViewModel`, raised from `SelectedStarterId`. Run tests until green.

## 9. Host: Components row selection + detail panel

- [x] 9.1 Add failing Host.Tests: `ComponentsPageViewModel` exposes `SelectedComponentId`; toggling a row's `IsSelected` (checkbox) does NOT change `SelectedComponentId` of a different, already-selected row; setting `SelectedComponentId` does NOT change any `IsSelected` value; `DetailBody` reflects the selected component's `LongDescription`/fallback to `Description`.
- [x] 9.2 Add failing Host.Tests: `SelectedComponentId` defaults to the first rendered option's `Id` when the page is constructed with a non-empty `Choices`.
- [x] 9.3 Implement `SelectedComponentId` property + `DetailTitle`/`DetailBody` computed properties on `ComponentsPageViewModel`, defaulted in the constructor, raised on selection change; confirm `SyncSelection()`/`_selection.CustomIds` derivation is untouched (still driven solely by `IsSelected`). Run tests until 9.1-9.2 pass.
- [x] 9.4 Triangulate: add a fixture with 3 components (one forced/locked, two custom) and assert selecting the forced row still updates `SelectedComponentId`/panel while its checkbox stays locked checked; run tests, confirm green.

## 10. XAML: shared detail panel + Components ListBox migration

- [x] 10.1 In `Themes/PageTemplates.xaml`, add one shared `DataTemplate x:Key="DetailPanelTemplate"` (or equivalent single-definition fragment) rendering `DetailHeader` (heading style), `DetailTitle` (semibold), `DetailBody` (wrapped body style) inside a `ScrollViewer`.
- [x] 10.2 Update the `InstallType`, `Permissions`, `UninstallMode`, `UninstallStrategy`, `StarterCatalog` page `DataTemplate`s: wrap each page's existing content in a two-column `Grid` (`3*`/`2*`, right column `MinWidth="230"`), left column keeps the existing list/choices content, right column is `<ContentPresenter Content="{Binding}" ContentTemplate="{StaticResource DetailPanelTemplate}" />`.
- [x] 10.3 Update the `Components` page `DataTemplate`: replace `ItemsControl` with `ListBox` (`SelectedValuePath="Id"`, `SelectedValue="{Binding SelectedComponentId, Mode=TwoWay}"`, `BorderThickness="0"`, `Background="Transparent"`), move the existing `CheckBox`+label/description stack into the `ListBox.ItemTemplate`, and add the same two-column `Grid` + `DetailPanelTemplate` right column as the other 5 pages.
- [ ] 10.4 Manual/visual verification (not automated): build and run the Host app (`installer\windows\build.bat` or `dotnet run` on the Host project, per operator's own build cadence — do not build unless explicitly requested), step through Components/InstallType/Permissions/StarterCatalog/UninstallMode/UninstallStrategy in both `en` and `es`, confirm the panel updates on selection and never renders blank text when a selection exists. **DEFERRED**: not run in this session per the "NUNCA build después de cambios salvo pedido explícito" project rule (CLAUDE.md §3) and the apply-change hard rule "NO bundle build" — `dotnet build ActiveStack.Bootstrapper.Host` confirms the XAML compiles cleanly (0 errors), but visual/keyboard-interaction QA needs an operator to actually run the app.

## 11. Final verification

- [x] 11.1 Run the full `dotnet test` suite for `ActiveStack.Bootstrapper.Core.Tests` and `ActiveStack.Bootstrapper.Host.Tests` — confirm 0 failures, including every pre-existing test untouched by this change (safety net). Result: Core.Tests 47/47, Host.Tests 145/145, 0 failures.
- [x] 11.2 Re-read `openspec/changes/gui-detail-panel/specs/windows-gui-wizard/spec.md` against the final implementation and confirm every scenario is covered by at least one automated test from groups 1-9. All scenarios verified covered (Components row-selection independence + panel; InstallType/Permissions/UninstallMode/UninstallStrategy/StarterCatalog panel + fallback scenarios).
