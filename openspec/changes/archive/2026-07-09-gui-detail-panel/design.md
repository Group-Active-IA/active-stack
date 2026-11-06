## Context

The Windows GUI wizard (WPF Host + Core) currently shows a single short `Description` per option on 6 pages: Components, InstallType, Permissions, StarterCatalog, UninstallMode, UninstallStrategy. The Go engine (L1) and catalog (L2) already emit a bilingual `long_description` on every relevant JSON payload (`windows options`, `windows uninstall-options`, `windows starters list`) — confirmed by reading `cmd/active-stack/headless/windows.go`, `windows_uninstall_options.go`, `windows_starters.go`: every option DTO carries `LongDescription string \`json:"long_description,omitempty"\``. The C# side (Core parsers + Host page view-models) does not read or render it yet. L4 landed `UiStrings.Get(lang, key)` (pure, static, en/es tables in `Core/Localization/UiStrings.cs`) and `_language` threading from `ShellViewModel` into every page-VM constructor via a trailing `string lang = "en"` parameter — this change follows that exact convention.

Verified current record/property shapes (source of truth, not the plan's guesses):
- `InstallerSessionStateBuilder.cs`: `InstallTypeChoice(string Id, string Label, string Description)`, `ComponentChoice(string Id, string Label, string Description, bool Recommended)`, `PermissionTierChoice(string Id, string Label, string Description, bool IsDefault, string? Warning)`.
- `StarterCatalogParser.cs`: `StarterChoice(string Id, string Name, string Description, IReadOnlyList<string> Includes, IReadOnlyList<string> Harnesses, int McpCount)`.
- `UninstallOptionsParser.cs`: `UninstallStrategyChoice(string Id, string Label, string Description, bool IsDefault, bool RequiresManifest)`. **There is no separate `UninstallModeChoice` record** — `UninstallModePageViewModel` reuses `InstallTypeChoice` for its `Choices` (`ObservableCollection<InstallTypeChoice>`), same as `InstallTypePageViewModel`. So the plan's "6 records" collapses to **5 distinct records**, one of which (`InstallTypeChoice`) backs two pages (InstallType and UninstallMode).
- `ComponentsPageViewModel` builds a Host-local `ComponentOption` (not `ComponentChoice`) per row from `ComponentChoice`; the panel-feeding data on Components must flow through `ComponentOption`, not the Core record, since that's what's bound in `ItemsControl`/`ListBox`.
- `Themes/PageTemplates.xaml` currently has one `DataTemplate` per page view-model type, each a bare `ScrollViewer` wrapping the page's primary control. There is no shared chrome/frame today beyond `PageHeadingStyle`/`PageBodyTextStyle`.

## Goals / Non-Goals

**Goals:**
- Render each option's long description in a side panel on the 6 named pages, updated live as the user changes selection/focus.
- Never show a blank panel: fall back to the short `Description` when `LongDescription` is empty.
- One shared panel template/style, not 6 copies.
- Give Components page a row-selection concept independent of the checkbox, without changing what gets submitted to the engine (still driven by `IsSelected`/checkbox state, never by row selection).
- Keep every change additive on the JSON/record boundary (backward compatible with a hypothetical engine payload missing `long_description`).

**Non-Goals:**
- No Go changes (L1/L2 already emit the field; out of scope here).
- No rewrite of unrelated pages (Assistants, Review, UninstallConfirm, StarterTarget, StarterReview, Backups, Progress, Complete, Hub, Language) — none of them show a single-item long description today, so none gain a panel.
- No change to window default size (760×560, `CanResize`) unless the reviewer finds the layout cramped during visual QA — if so, bumping width is a one-line follow-up, not part of this change's scope.
- No change to which ids get submitted to the engine on Components (still `IsSelected`-driven).

## Decisions

### D1 — Records gain `string LongDescription = ""` (additive default)
Add `LongDescription` as the last positional parameter with a default value on: `InstallTypeChoice`, `ComponentChoice`, `PermissionTierChoice`, `StarterChoice`, `UninstallStrategyChoice`. A C# positional record parameter with a default is source-compatible with existing call sites (tests constructing these records without a `LongDescription` argument keep compiling). Each of the 3 parsers' internal JSON DTOs gains `[JsonPropertyName("long_description")] public string? LongDescription { get; init; }`, mapped into the record as `item.LongDescription ?? string.Empty`. No `UninstallModeChoice` record is introduced — `InstallTypeChoice.LongDescription` is reused by both InstallType and UninstallMode pages, matching the existing reuse pattern.

Alternative considered: a wrapper `LocalizedOption` type shared across all 5 records. Rejected — over-engineering for 5 simple record additions; would touch every call site's construction pattern for no behavioral gain, and the plan explicitly favors additive minimal diffs.

### D2 — Per-page `DetailTitle`/`DetailBody` derived properties
Each of the 6 page-VMs (`ComponentsPageViewModel`, `InstallTypePageViewModel`, `PermissionsPageViewModel`, `StarterCatalogPageViewModel`, `UninstallModePageViewModel`, `UninstallStrategyPageViewModel`) gains two read-only properties:
- `DetailTitle` — the selected/focused item's `Label` (or `Name` for `StarterChoice`).
- `DetailBody` — `LongDescription` if non-empty, else `Description`.

Both are raised via `RaisePropertyChanged(nameof(DetailTitle))`/`RaisePropertyChanged(nameof(DetailBody))` from the existing selection-change setters (`SelectedId`, `SelectedTierId`, `SelectedStarterId`, `SelectedStrategyId`, and the new `SelectedComponentId` on Components — see D3). No new fields are needed for the 5 single-select pages since a selected id already exists; `DetailTitle`/`DetailBody` are computed getters over the existing `Choices`/`Tiers`/`Strategies` collection filtered by the current selected id, so they never go stale independent of the notification.

Naming: `DetailTitle`/`DetailBody` (not `LongDescription`/`SelectedDescription`) to keep panel semantics decoupled from the underlying field name, and to match the plan's own naming suggestion; every VM uses the identical pair so the shared XAML template can bind uniformly regardless of page type via `RelativeSource AncestorType` on a per-page grid (each page's own root grid holds `DetailTitle`/`DetailBody`, not a page-type-agnostic binding at the `ResourceDictionary` root — WPF `DataTemplate` content binds directly to the page VM's `DataContext`, so this is a plain property binding, no adapter needed).

### D3 — Components: `ListBox` with `CheckBox` in `ItemTemplate`, `SelectedComponentId` independent of `IsSelected`
Today `ComponentsPageViewModel.Choices` is an `ObservableCollection<ComponentOption>` rendered by an `ItemsControl` with no selection concept — only `IsSelected` (the checkbox). To feed the panel from row focus/click without touching checkbox semantics:
- XAML: swap `ItemsControl` → `ListBox` (`BorderThickness="0"`, `Background="Transparent"`, matching the styling convention already used for `InstallType`/`UninstallMode`/etc.), with `SelectedValuePath="Id"` and `SelectedValue="{Binding SelectedComponentId, Mode=TwoWay}"`; the `CheckBox` moves inside the `ListBox.ItemTemplate` (still bound to `IsSelected`/`IsToggleable` exactly as today). Row click or arrow-key navigation sets `ListBoxItem` selection → `SelectedComponentId`; clicking the `CheckBox` inside the row still only toggles `IsSelected` (WPF routes the `CheckBox` click as its own event, not a `ListBoxItem.IsSelected` change, so checkbox-click does not force row selection — but per standard `ListBox` behavior clicking anywhere in the row, including on the checkbox's row background, will still select the row; this is acceptable and intentional: it means checking an item also focuses its detail, which is a reasonable UX outcome, not a bug).
- `ComponentsPageViewModel` adds `string? SelectedComponentId { get; set; }` (defaults to the first option's `Id`, or empty if `Choices` is empty) with the standard `RaisePropertyChanged` + `RaisePropertyChanged(nameof(DetailTitle))`/`RaisePropertyChanged(nameof(DetailBody))` pattern used elsewhere in this file. `SyncSelection()` (which derives `_selection.CustomIds` from `IsSelected`) is untouched — `SelectedComponentId` never feeds the engine payload.
- `DetailTitle`/`DetailBody` on Components look up the option in `Choices` whose `Id == SelectedComponentId` (falling back to `Choices.FirstOrDefault()` when `SelectedComponentId` is null/not found, e.g. before first render or after the list is rebuilt with a different item set).

Alternative considered: keep `ItemsControl` and add a separate "focused index" tracked via `PreviewMouseDown`/`GotFocus` attached behavior on each row. Rejected — reinvents `ListBox`'s built-in single-selection and keyboard navigation (arrow keys, Home/End) for no benefit; `ListBox` gives that for free and every other single-select page in this codebase already uses it.

### D4 — One shared detail-panel template/style in `PageTemplates.xaml`
Add one `Style x:Key="DetailPanelBorderStyle"` (or a `ControlTemplate`/`DataTemplate` fragment) and a reusable structural pattern: for each of the 6 affected page templates, replace the current single-column `ScrollViewer` root with a `Grid` of two `ColumnDefinition`s (`Width="3*"` for the list/choices, `Width="2*"` for the panel, panel `MinWidth="230"`), each column wrapped in its own `ScrollViewer`. The panel column is a `StackPanel` containing:
- A `TextBlock` bound to the static localized header (`UiStrings.Get(lang, "detail.header")`, surfaced by each VM as a `DetailHeader` property, or a single shared resource — see below) styled `PageHeadingStyle`.
- `TextBlock Text="{Binding DetailTitle}"` (`FontWeight=SemiBold`).
- `TextBlock Text="{Binding DetailBody}"` (`Style="{StaticResource PageBodyTextStyle}"`, wrapped).

To avoid 6 near-identical `Grid`/`ScrollViewer`/`StackPanel` blocks, the shared piece is factored as a `DataTemplate` fragment invoked via `ContentPresenter` — concretely: define `<DataTemplate x:Key="DetailPanelTemplate">` once, binding `DetailTitle`/`DetailBody`/`DetailHeader` off whatever `DataContext` it's given, and each of the 6 page templates' right column is `<ContentPresenter Content="{Binding}" ContentTemplate="{StaticResource DetailPanelTemplate}" />` — since the page VM itself carries `DetailTitle`/`DetailBody`, `Content="{Binding}"` (the page VM) is enough; no extra wrapper object needed. This keeps the panel markup defined exactly once.

`DetailHeader`: rather than have each of the 6 VMs re-declare a `DetailHeader` property, `WizardPageViewModelBase` gains one shared `public string DetailHeader { get; }` set from the `lang` constructor parameter already threaded everywhere (`DetailHeader = UiStrings.Get(lang, "detail.header")`), so the 6 constructors need only pass `lang` through to the base constructor as they already do — no per-page duplication. `WizardPageViewModelBase`'s constructor signature changes from `(string title, string subtitle)` to `(string title, string subtitle, string lang)`; the 6 affected page VMs already receive `lang` and already call `base(...)`, so this is a call-site touch, not new plumbing. Pages that don't show a panel keep passing `lang` too (harmless — `DetailHeader` is simply unused by their `DataTemplate`), keeping `WizardPageViewModelBase` uniform rather than forking a subclass just for the 6 panel pages.

Alternative considered: a `DetailPanelViewModel` composed into each page VM (has-a instead of properties-on-self). Rejected for this change — would require either wrapping every page VM's `DataContext` for the panel binding (extra indirection in already-simple property paths) or exposing a nested `Detail` property and rebinding every panel `TextBlock` to `Detail.Title`/`Detail.Body`; the flat-properties-on-self approach matches this codebase's existing convention of flat computed properties (e.g. `WarningText` on `PermissionsPageViewModel`) and keeps the `ContentPresenter Content="{Binding}"` trick simple (no adapter object to construct/track per selection change).

### D5 — Fallback and empty-state handling
`DetailBody` getter: `!string.IsNullOrEmpty(item?.LongDescription) ? item.LongDescription : item?.Description ?? string.Empty`. If `Choices`/`Tiers`/`Strategies` is empty (defensive — shouldn't happen given upstream engine guarantees at least one entry), `DetailTitle`/`DetailBody` return `string.Empty`, and the panel simply renders blank text (not literally "no panel" — acceptable, matches "never blank panel" only in the sense of "never blank when there IS a selection"; an empty list is a pre-existing out-of-scope edge case for every one of these pages already, since e.g. `InstallTypePageViewModel` already assumes `Choices` is non-empty for `RecommendedModeId` fallback logic).

## Risks / Trade-offs

- **[Risk]** `WizardPageViewModelBase` constructor signature change (adds a `lang` parameter) touches all callers, including pages outside the 6 in this change (Hub, Language, Assistants, Review, Progress, Complete, UninstallAgents, UninstallConfirm, StarterTarget, StarterReview, Backups) — a wider blast radius than the 6 panel pages alone. → **Mitigation**: every page VM already receives and threads `lang` today (L4 landed that), so this is a mechanical one-line addition to each existing `base(...)` call, not new parameter plumbing; compile-time errors will catch any missed call site immediately, and the existing Host.Tests suite already exercises every page VM's constructor.
- **[Risk]** Changing Components from `ItemsControl` to `ListBox` could alter visual/keyboard behavior subtly (e.g. `ListBox` adds a selection highlight brush, focus rectangle) versus the current plain `CheckBox` stack. → **Mitigation**: reuse the same `BorderThickness="0" Background="Transparent"` pattern already applied to `InstallType`/`UninstallMode`/`StarterCatalog` `ListBox`es in this file, and cover the checkbox-vs-row-selection independence with a dedicated Host.Tests case (toggling a checkbox must not change `SelectedComponentId` unless that row was already the one clicked).
- **[Risk]** Reusing `InstallTypeChoice.LongDescription` for both InstallType and UninstallMode is a shortcut versus a hypothetical dedicated `UninstallModeChoice` — if a future change needs uninstall-mode-specific fields, this reuse would need to be undone. → **Mitigation**: acceptable now since the two pages already share the exact same record and DTO shape end-to-end (engine emits the same `modes` array shape for both endpoints); documented here for future readers.
- **[Trade-off]** Bumping window width is deferred rather than done preemptively, per the plan's explicit guidance — if the 3*/2* split feels cramped at 760px during visual QA, that's a fast one-line follow-up (`Width`/`MinWidth` in `MainWindow.xaml`), not blocking this change's merge.

## Migration Plan

Purely additive, no data migration. Rollout is the change itself: land records + parsers first (Core, TDD), then the shared XAML panel + `WizardPageViewModelBase` `lang`-parameter change, then each page VM's `DetailTitle`/`DetailBody`/(`SelectedComponentId` for Components) in sequence, keeping `dotnet test` green after each page. No rollback complexity beyond a normal revert — no persisted state, no schema.

## Open Questions

- None blocking. Exact panel pixel styling (padding, divider line vs. plain column gap) is left to implementation-time visual judgment within the Grid `3*/2*` + `MinWidth 230` constraint already fixed by the plan.
