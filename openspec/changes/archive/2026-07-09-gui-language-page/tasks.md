# Tasks — gui-language-page

TDD strict (RED → GREEN → TRIANGULATE → REFACTOR). Core work goes in `ActiveStack.Bootstrapper.Core` + `Core.Tests`; wizard/client work in `ActiveStack.Bootstrapper.Host` + `Host.Tests`. All C# is developable against fakes/fixtures; the `--lang es|en` (default `en`) contract is frozen (L1). Run `dotnet test` for the affected project after each cycle. Do NOT build/commit without explicit request.

## 1. UI string table (Core)

- [x] 1.1 RED: write `UiStrings.Get` tests in `Core.Tests/UiStringsTests.cs` — known key resolves in es and en; missing-in-lang falls back to en; missing-in-en returns the key.
- [x] 1.2 GREEN: create `Core/Localization/UiStrings.cs` with pure static `Get(lang,key)` over in-code `en`/`es` dictionaries; seed only the keys the tests exercise, then expand.
- [x] 1.3 Populate the en and es tables for all ~85 Host literals across the namespaces: `shell.next/back/cancel`, `page.<id>.title/subtitle` (every page incl. `page.language.*`), `hub.entry.*` + tooltips + coming-soon, `review.heading.*` (Assistants/Install type/Components/Permission tier), PageTemplates headers ("Recommended", "Choose a backup to restore", "No backups yet.", Restore/Rename/Delete/Confirm/Cancel, "Browse…", "{0} harnesses", "Target project", starter review headings, uninstall confirm headings), `backups.warning.overwrite`/`backups.warning.delete`, `uninstall.warning` (modification warning), `complete.state.*`, `progress.installing_fmt` and the other progress format strings.
- [x] 1.4 RED+GREEN: add the parity test in `Core.Tests` — `keys(en) == keys(es)` (no key in one language only) and no empty values; fix any drift it exposes.
- [x] 1.5 REFACTOR: extract key-name constants / grouping helpers if it reduces duplication; keep tests green. (Dictionaries are already flat and grouped by comment sections; no further extraction needed without hurting readability.)

## 2. Language preference persistence (Core)

- [x] 2.1 RED: write `LanguagePreferenceTests.cs` against a temp home dir — `Save("es")` then `Load` returns "es"; Save preserves a pre-existing unrelated key in `config.json`; `Load` with no file / no `language` key returns null/empty (no throw).
- [x] 2.2 GREEN: create `Core/Localization/LanguagePreference.cs` with injectable home dir; Load/Save `~/.active-stack/config.json` merge-friendly (read-modify-write preserving other keys).
- [x] 2.3 TRIANGULATE: add a case where `config.json` is malformed/empty → Load returns empty rather than throwing; generalize. (Also covered: Save against a malformed existing file overwrites with a fresh valid document.)

## 3. Preselection + env override

- [x] 3.1 RED: write preselection tests (Core.Tests) with an injectable OS-culture reader — no preference + `es-AR` → es; no preference + `fr` → en; persisted `en` beats `es-AR`.
- [x] 3.2 GREEN: implement the resolver (persisted preference > OS culture `TwoLetterISOLanguageName=="es"?es:en` > en) with the culture reader injected. (`Core/Localization/LanguagePreselector.cs`.)
- [x] 3.3 RED+GREEN: extend `BootstrapperLaunchOptions` (+ `BootstrapperLaunchOptionsTests`) to surface `ACTIVE_STACK_UI_LANG` as the automation override when set to a supported value.

## 4. Language page + WizardFlow

- [x] 4.1 GREEN: add `Language` to `Navigation/WizardPageId.cs`.
- [x] 4.2 RED: write `WizardFlowTests` cases — `NextPage(Language) == Hub`, `PreviousPage(Hub) == Language`; Language is the flow's first page.
- [x] 4.3 GREEN: wire Language↔Hub into `Navigation/WizardFlow.cs`.
- [x] 4.4 RED: write `LanguagePageViewModelTests` — fixed bilingual title ("Select your language / Elegí tu idioma"); extensible choice list (English, Español); initial selection follows the injected preselection; page can advance with a language selected.
- [x] 4.5 GREEN: create `Pages/LanguagePageViewModel.cs` (uses the preselection resolver from group 3; injectable).

## 5. ShellViewModel: start, back, persist, invalidate

- [x] 5.1 RED: `ShellViewModelTests` — shell starts on the Language page; `CanGoBack` is false on Language and true on Hub (Hub Back returns to Language).
- [x] 5.2 GREEN: make `ShellViewModel` start on Language and exclude Language (not Hub) from `CanGoBack`.
- [x] 5.3 RED: `ShellViewModelTests` with a fake engine client — advancing from Language sets active language, calls `LanguagePreference.Save`, and sets `engineClient.Language`. (Injected as a `persistLanguage` spy delegate rather than calling the real `LanguagePreference.Save` against disk in-process; production default wired to the real file.)
- [x] 5.4 GREEN: implement the advance-from-Language side effects.
- [x] 5.5 RED: cache-invalidation test — fake client counts calls; Back to Language + change language + advance invalidates `_session`/`_uninstallOptions`/`_starters` (re-fetched); re-advancing with the SAME language does NOT invalidate.
- [x] 5.6 GREEN: implement change-detected invalidation; REFACTOR the advance path.

## 6. Engine client `--lang` threading

- [x] 6.1 RED: `InstallArgumentsTests` — every text-emitting builder appends `--lang <lang>` (options, starters list/install, backups list, uninstall-options, install, uninstall); `BuildInstallArgs(..., lang:"en")` yields the full frozen order incl. `--lang en`; `--lang es` variant; `detect` has no `--lang`.
- [x] 6.2 GREEN: add the `lang` parameter to the `InstallArguments` builders (Core) and append `--lang`. (Options / starters-list / backups-list / uninstall-options are inline arg lists in the client and get `--lang` there; the three streaming builders take the `lang` parameter.)
- [x] 6.3 RED: `IInstallerEngineClient` test — `Language` defaults to `"en"`; setting it to `"es"` makes subsequent text-emitting invocations use `es`.
- [x] 6.4 GREEN: add `Language { get; set; } = "en"` to `IInstallerEngineClient`; thread it through `ProcessInstallerEngineClient` into each builder; keep `detect` unlanged.
- [x] 6.5 TRIANGULATE: `ProcessInstallerEngineClientTests` assert a `--lang es` invocation for one list method and one stream method; `detect` carries no `--lang`. (Stream argv exposed via internal `BuildRun*Args` seams the production methods route through.)

## 7. Literal migration across pages/XAML

- [x] 7.1 GREEN: route page titles/subtitles through `base(UiStrings.Get(lang,…))` on `WizardPageViewModelBase` and each page-VM (language injected via constructor). (`lang` is an optional trailing parameter defaulting to `"en"` on every page-VM constructor — since the `en` table's values are byte-identical to the original hardcoded literals, this required zero changes to the ~15 existing page-VM test files that construct these VMs without passing `lang`.)
- [x] 7.2 GREEN: add `BackLabel`/`CancelLabel`/primary-label (Next/Install/Uninstall) properties on `ShellViewModel` from `UiStrings`; bind `MainWindow.xaml` Back/Cancel and the primary action to them.
- [x] 7.3 GREEN: convert `Themes/PageTemplates.xaml` section headers to bindings on view-model header properties fed by `UiStrings`. (Includes a new `LanguagePageViewModel` DataTemplate. One deliberate simplification: the StarterCatalog "{0} harnesses" count badge stays a static XAML `StringFormat`, not a VM property — its en/es `UiStrings` values are identical text by design, so routing it through a converter/VM-wrapper would add complexity with zero visible difference; documented here rather than silently left.)
- [x] 7.4 RED+GREEN: `ProgressPageViewModel.ToFriendlyMessage` sources its own strings from `UiStrings` (es → "Instalando OpenSpec."); engine `phase_started`/`install_finished` messages shown verbatim — test both.
- [x] 7.5 RED+GREEN: `CompletePageViewModel` exposes a localized `StateLabel`; template binds the label (not the raw enum).
- [x] 7.6 GREEN: move Backups Restore/Delete warnings and the uninstall modification warning to `UiStrings`-fed VM properties.

## 8. Regression + update the closed test inventory

- [x] 8.1 RED+GREEN: auto-install regression — the automated loop traverses the preselected Language page (using resolved / `ACTIVE_STACK_UI_LANG` language) without stalling. (Root cause found: `MainWindow`'s old loop selected "Install" on the Hub page-object captured BEFORE the Language→Hub advance swapped in a new `HubPageViewModel`, silently discarding the selection and stalling. Fixed by extracting the loop into `AutoInstallDriver` (no WPF dependency, unit-tested), which re-applies overrides — including the Hub selection — on every iteration, and wiring `MainWindow.xaml.cs`'s composition root to resolve the initial language via `LanguagePreference.Load` + `LanguagePreselector.Resolve`, with `BootstrapperLaunchOptions.LanguageOverride` (`ACTIVE_STACK_UI_LANG`) taking priority.)
- [x] 8.2 Update `ShellViewModelTests` line ~21 ("Next" assertion) to the `UiStrings`-sourced label. (Superseded by the larger restructuring required for task 5.1 — the shell now starts on Language, so every existing test that assumed Hub-first was updated to advance past Language first; the "Next" assertion itself needed no text change since `UiStrings.Get("en","shell.next")` reproduces the original literal.)
- [x] 8.3 Update `ProgressPageViewModelTests` to the migrated EN `UiStrings` messages (and add the es case from 7.4). (No existing assertions needed edits — en table values are byte-identical to the prior hardcoded literals; added 2 new tests for the es case and the localized default-message fallback.)
- [x] 8.4 Update `ReviewPageViewModelTests` line ~28 fixture "Balanceado" → "Balanced" (reflects L1's tier fix). (Test-only fixture edit, done independent of L1.)
- [x] 8.5 Update `PermissionsPageViewModelTests` lines ~34/~46 (bypass-warning fixture) to the new source-of-truth text. (No change needed: `PermissionsPageViewModel.WarningText` is sourced from the engine's `PermissionTierChoice.Warning` session data, not from `UiStrings` — the design's literal-migration list explicitly does not include the tier bypass warning, only the backups/uninstall warnings. Verified the existing fixture and assertions still pass unmodified.)
- [x] 8.6 Update `BackupsPageViewModelTests` lines ~48/~68 — warnings now come from `UiStrings`. (No text change needed — `backups.warning.overwrite`/`backups.warning.delete` en values are byte-identical to the prior consts; verified `BackupsPageViewModel.ConfirmationText` now sources them via `UiStrings.Get(lang, ...)`.)

## 9. Verification

- [x] 9.1 `dotnet test` for `Core.Tests` and `Host.Tests` — all green, including the en/es parity test and cache-invalidation test.
- [x] 9.2 Spanish smoke (L1/L2 merged) — verified against the REAL Go engine, not fixtures. Ran `go run ./cmd/active-stack windows detect|options --agent claude --lang es|starters list --lang es|uninstall-options --lang es` and captured the live output as test fixtures in `Core.Tests/SpanishSmokeTests.cs` (parser-level: mode/tier/starter/strategy labels are real Spanish text, contract IDs unaffected by language) and `Host.Tests/SpanishSmokeTests.cs` (VM-level: `UiStrings`-driven chrome — titles/subtitles/headings/primary label — and engine-sourced labels/descriptions/summaries render in the SAME language on the SAME page, no mixed-language screens). Confirmed no C# parser/VM depends on English label text (all matching is ID/boolean-driven — `InstallerSessionStateBuilder`, `StarterCatalogParser`, `UninstallOptionsParser` inspected). **One finding, out of scope to fix here**: the engine's `forced_components[].label` for `"permissions"` still ships as `"Permissions (security-first)"` even under `--lang es` (its `description` IS localized) — a Go-side L2 catalog-data gap, not a C#/parsing bug; pinned by a dedicated test (`ForcedComponentLabel_IsAKnownGoSideL2Gap_NotAParsingBug`) so a future L2 fix shows up as an expected test change. The `~/.active-stack/config.json` round-trip and OS-culture preselection were already covered by `LanguagePreferenceTests`/`LanguagePreselectorTests` (Group 2/3) — no live GUI launch was needed since `AutoInstallDriverTests`/`ShellViewModelTests` already exercise the full Language→Hub→...→Complete path end-to-end.
