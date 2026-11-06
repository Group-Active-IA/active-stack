## Context

The Windows GUI installer (WPF `ActiveStack.Bootstrapper.Host` + pure `ActiveStack.Bootstrapper.Core`) is a multi-page wizard: `ShellViewModel` drives navigation through pure `WizardFlow`/`UninstallFlow`/`StarterFlow` functions, page view-models expose title/subtitle/advance-eligibility, and `DataTemplate`s in `Themes/PageTemplates.xaml` render each VM. Today all ~85 UI literals are hardcoded English and the engine (`active-stack windows ...`, Go) returns some labels in English and some catalog copy in Spanish — a language-mixed screen with no selector.

This change (L4 of a 5-change plan) adds language selection ES/EN to the GUI. It is developable in parallel with L1 (`i18n-engine-locales`, the Go side) because the `--lang es|en` contract (default `en`) is FROZEN: the C# code threads the flag and is tested against fake clients/fixtures; only the final smoke test needs L1's real localized output. Governance is MEDIO (page view-models + the engine-client adapter).

Constraints that shape the design:
- `build.ps1` uses `PublishSingleFile`, which is incompatible with `.resx` satellite assemblies → localization must be plain in-code tables.
- Contract IDs (`estricto`/`balanceado`/`bypass`, `lite`/`full`/`custom`) are stable; only labels/descriptions localize.
- The wizard already has a closed inventory of literal-asserting tests that must migrate in lockstep (listed in the proposal Impact).

## Goals / Non-Goals

**Goals:**
- A first-class Language page as the wizard's first page, with OS detection, persistence, and an extensible language list.
- A pure, testable UI string table (`UiStrings.Get(lang,key)`) with en/es parity enforced by test.
- Thread `--lang` into every text-emitting engine subcommand; re-fetch engine data when the language changes.
- Migrate all ~85 Host literals to the string table without changing any behavior other than their language.

**Non-Goals:**
- Any Go/engine change — that is L1. This change only consumes the frozen `--lang` contract.
- The lateral detail panel and long descriptions — that is L5 (`gui-detail-panel`).
- Catalog/YAML localization — that is L2/L3.
- Adding languages beyond ES/EN now (the list is built to be extensible, but only two entries ship).
- Any `internal/backup`-style destructive change; the only file written is `~/.active-stack/config.json`, merge-friendly.

## Decisions

### Localization lives in Core as pure static tables, not `.resx`
`UiStrings` and `LanguagePreference` go in `ActiveStack.Bootstrapper.Core/Localization/` (no WPF dependency) so `Core.Tests` can cover them directly. `UiStrings.Get(lang, key)` is a pure function over two in-code `Dictionary<string,string>` tables (en, es) with fallback en→key. **Why not `.resx`:** satellite assemblies break `PublishSingleFile` in `build.ps1`. **Why not a mutable singleton:** a global "current language" is hidden state that makes VMs non-deterministic to test; instead the `ShellViewModel` owns `_language` and injects it into every page-VM constructor. Key namespacing mirrors usage: `shell.*`, `page.<id>.title|subtitle`, `hub.entry.*`, `review.heading.*`, `backups.warning.*`, `uninstall.warning`, `complete.state.*`, `progress.*_fmt`.

### en/es parity is a test, not a convention
A Core.Tests parity test asserts `keys(en) == keys(es)` and no empty values. This is the anti-drift seatbelt: a literal added in one language only fails the build, mirroring L1's Go key-parity test. Alternatives (manual review, runtime warnings) don't fail loudly and were rejected.

### Preselection order: persisted preference > OS culture > English, all injectable
`LanguagePreference.Load(homeDir)` reads `~/.active-stack/config.json`; the OS reader wraps `CultureInfo.CurrentUICulture.TwoLetterISOLanguageName` (`"es"`→Español, else English). Both are injected so tests are deterministic (`es-AR`→es, `fr`→en, persisted `en` beats `es-AR`). `ACTIVE_STACK_UI_LANG` in `BootstrapperLaunchOptions` is the automation override. **Why injectable:** the whole point of pure preselection logic is a table of unit tests without touching the real OS or filesystem.

### Persistence is merge-friendly
`LanguagePreference.Save(homeDir, lang)` reads the existing JSON, sets `language`, and rewrites — preserving other keys — so it composes with future config keys (the TUI reads the same file per the plan). Round-trip tested against a temp dir. **Why:** `config.json` is shared user state; clobbering it would be a HIGH-governance-style regression.

### Shell owns language and invalidates caches on change
`ShellViewModel` starts on `WizardPageId.Language`. `CanGoBack` excludes Language (not Hub — a divergence from today, where Hub was the entry). On advancing from Language: set `_language`, `LanguagePreference.Save`, set `engineClient.Language`, and if the language **changed** from the previously active value, invalidate `_session`/`_uninstallOptions`/`_starters` so the next flow re-fetches localized engine data. Re-selecting the same language must NOT invalidate (avoids needless re-query). **Why invalidate:** the engine's option/tier/starter text is localized server-side; cached English data would leak through after a Back-and-change.

### `--lang` threading centralized in `InstallArguments`
Every text-emitting builder gains a `lang` parameter appended as `--lang <lang>`; `detect` does not. `IInstallerEngineClient.Language { get; set; }` defaults `"en"`; `ProcessInstallerEngineClient` passes `Language` into each builder. **Why the client holds the property rather than each call passing it:** the shell sets it once per language change and every subsequent flow inherits it, matching how the shell drives flows.

### Literal migration keeps XAML dumb, VMs testable
Section headers in `PageTemplates.xaml` and footer Back/Cancel in `MainWindow.xaml` become bindings to VM/shell properties (`BackLabel`, `CancelLabel`, header properties) fed by `UiStrings`, so they are assertable in Host.Tests without rendering XAML. Page titles flow through `base(UiStrings.Get(lang,…))` on `WizardPageViewModelBase`. `ProgressPageViewModel.ToFriendlyMessage` owns its strings via `UiStrings`, but engine `phase_started`/`install_finished` messages arrive already localized (from L1) and are shown verbatim. `CompletePageViewModel` gains a localized `StateLabel` (the template binds the enum today).

## Risks / Trade-offs

- **[L1 not yet merged when L4 implements]** → All C# is tested against fakes/fixtures; the frozen `--lang es|en`/default-`en` contract makes this safe. Only the final end-to-end smoke (real Spanish engine output) waits on L1. The `ReviewPageViewModelTests` fixture change ("Balanceado"→"Balanced") anticipates L1's tier fix but is a test-only fixture edit.
- **[Stale cache leaks old-language engine text]** → The change-detected cache invalidation of `_session`/`_uninstallOptions`/`_starters` on language change is specced with a dedicated test (fake client counts calls) covering the Back-and-change path.
- **[Clobbering `config.json`]** → Merge-friendly Save + round-trip test that asserts a pre-existing unrelated key survives.
- **[Missing/партial migration leaves English literals]** → The en/es parity test plus the closed inventory of literal-asserting tests (updated to the new source-of-truth strings) catch stragglers; any literal still hardcoded won't be covered by a `UiStrings` key and will surface in the Spanish smoke test.
- **[Auto-install loop stalls on the new first page]** → A regression test asserts the automated loop traverses the preselected Language page using the resolved/overridden language.

## Migration Plan

Pure additive, no data migration. Order (TDD, per tasks.md): (1) `UiStrings` + parity test; (2) `LanguagePreference` + round-trip; (3) preselection logic + `ACTIVE_STACK_UI_LANG`; (4) `WizardPageId.Language` + `LanguagePageViewModel` + `WizardFlow` Language↔Hub; (5) `ShellViewModel` start/CanGoBack/persist/invalidate; (6) `IInstallerEngineClient.Language` + `InstallArguments` lang param + `ProcessInstallerEngineClient` threading; (7) literal migration across pages/XAML + update the closed inventory of literal tests. No rollback concern beyond reverting the branch; the only persisted side effect is a merge-friendly `config.json` key. Final verification: `dotnet test` (Core.Tests + Host.Tests), then the Spanish smoke once L1 is available.

## Open Questions

- None blocking. The window stays 760×560 (L5 may revisit sizing when the detail panel lands). Exact `UiStrings` key strings are an implementation detail bounded by the namespaces above and the parity test.
