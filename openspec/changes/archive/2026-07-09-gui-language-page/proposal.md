## Why

The Windows GUI installer wizard is English-only with ~85 hardcoded UI literals and no way to pick a language, and today the screen is even language-mixed (engine labels in English, catalog copy in Spanish). Spanish-speaking users — the primary audience — get a wizard they cannot read in their language. This change adds a first-class language selection (ES/EN) as the very first wizard page, with OS detection and persistence, so the whole GUI (including the engine-provided mode/tier/component text, once L1 lands) renders in one chosen language.

## What Changes

- **New LanguagePage as the FIRST wizard page** (before Hub): a fixed bilingual title ("Select your language / Elegí tu idioma") and an extensible choice list (English / Español). Preselection order: persisted preference > OS UI culture (`es` → Español, else English) > English. The OS detector is injectable for tests.
- **Pure C# UI string table** (`Core/Localization/UiStrings.cs`): a static, PURE `Get(lang, key)` covering the ~85 Host literals (shell footer, every page title/subtitle, hub entries + tooltips + coming-soon, review headings, PageTemplates headers, backups warnings, uninstall modification warning, complete-state labels, progress format strings). Fallback: `en` → key. NO `.resx` (satellite assemblies break `PublishSingleFile`). `en`/`es` tables must have identical key sets and no empty values (parity test).
- **Language persistence** (`Core/Localization/LanguagePreference.cs`): merge-friendly Load/Save of `~/.active-stack/config.json` (`{"language":"es"}`), preserving any other keys, temp-dir testable.
- **`--lang` threading to the Go engine client**: `IInstallerEngineClient.Language { get; set; }` (default `"en"`); `InstallArguments` builders gain a `lang` parameter; `ProcessInstallerEngineClient` appends `--lang <lang>` to options, starters list/install, backups list, uninstall-options, install, and uninstall (the `windows detect` subcommand does NOT take `--lang`). The flag name `--lang` with values `es|en` (default `en`) is FROZEN and shared with change L1 (`i18n-engine-locales`); all C# work here is developable against fakes/fixtures and needs L1 only for live end-to-end integration.
- **ShellViewModel owns `_language`**: starts on the Language page, injects the language into every page-VM constructor, excludes Language (not Hub) from `CanGoBack`, and on advancing from Language sets `_language`, persists via `LanguagePreference.Save`, sets the engine client's `Language`, and INVALIDATES cached `_session`/`_uninstallOptions`/`_starters` when the language changed (so going Back and changing language re-fetches localized engine data).
- **Migrate the ~85 literals** to the string table: page titles via `base(UiStrings.Get(lang, …))`; `MainWindow.xaml` Back/Cancel → `ShellViewModel.BackLabel`/`CancelLabel` bindings; `PrimaryLabel` + Install/Uninstall overrides; XAML section headers in `PageTemplates.xaml` → testable VM properties; `ProgressPageViewModel.ToFriendlyMessage` own strings via `UiStrings` (engine `phase_started`/`install_finished` messages arrive already localized and are shown verbatim); `CompletePageViewModel` gains a localized `StateLabel` (the template binds the enum today).
- **`ACTIVE_STACK_UI_LANG` env override** in `BootstrapperLaunchOptions` for automation; the auto-install loop must traverse the preselected LanguagePage without breaking (regression test).

## Capabilities

### New Capabilities
- `windows-gui-localization`: the pure UI string table (`UiStrings.Get`), en/es key parity, language-preference persistence to `~/.active-stack/config.json`, OS-culture-based preselection, and the `ACTIVE_STACK_UI_LANG` env override.

### Modified Capabilities
- `windows-gui-wizard`: adds the Language page as the flow's first page (Language → Hub forward, Hub → Language back), moves the first-page/Back-disabled rule from Hub to Language, and requires the shell/pages to render their literals from the language-scoped string table (footer labels, page titles/subtitles, section headers, warnings, complete-state label, progress messages) instead of hardcoded English.
- `windows-gui-engine-client`: adds a `Language` property on the client and a `lang` parameter to the argument builders, so every windows subcommand that emits user-facing text (options, starters list/install, backups list, uninstall-options, install, uninstall) is invoked with `--lang <lang>`; `detect` remains unlanged.

## Impact

- **Depends on** change L1 (`i18n-engine-locales`, proposed in parallel) for live integration only; the `--lang es|en` (default `en`) contract is frozen so all C# here is testable against fakes/fixtures.
- **New files**: `ActiveStack.Bootstrapper.Core/Localization/UiStrings.cs`, `.../Localization/LanguagePreference.cs`, `.../Navigation/WizardPageId.cs` (add `Language`), `Pages/LanguagePageViewModel.cs`, plus Core.Tests + Host.Tests for each.
- **Modified files**: `Navigation/WizardFlow.cs`, `Navigation/ShellViewModel.cs`, `IInstallerEngineClient.cs`, `InstallArguments.cs` (Core), `ProcessInstallerEngineClient.cs`, `BootstrapperLaunchOptions.cs`, `MainWindow.xaml`, `Themes/PageTemplates.xaml`, `Pages/ProgressPageViewModel.cs`, `Pages/CompletePageViewModel.cs`, and every page-VM whose title/subtitle/labels move to the string table.
- **Existing literal-asserting tests to update** (closed inventory): `ShellViewModelTests` (line 21, "Next"), `ProgressPageViewModelTests` (EN messages), `ReviewPageViewModelTests` (line 28, "Balanceado" → "Balanced" fixture, reflecting L1's tier fix), `PermissionsPageViewModelTests` (lines 34/46, bypass-warning fixture), `BackupsPageViewModelTests` (lines 48/68, warnings move to `UiStrings`).
- **Governance**: MEDIO (`Pages/*` view-models and the engine-client adapter). No `internal/backup`-equivalent destructive surface is touched; `~/.active-stack/config.json` writes are merge-friendly and backed by round-trip tests.
