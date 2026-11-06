## Why

The Windows installer engine emits its user-facing text as ~50 hardcoded English literals scattered across `cmd/active-stack/headless/windows*.go`, with zero i18n infrastructure. Worse, the permission-tier screen is already language-MIXED: it ships Spanish labels ("Estricto"/"Balanceado") with English descriptions. A localized GUI (Spanish + English) is planned, and it needs the engine to return every user-facing string in the caller's chosen language. This change lays the Go-side foundation: a reusable `internal/i18n` package, a `--lang` flag on every `windows` subcommand, and the migration of all hardcoded literals into ES/EN string tables — fixing the tier mix along the way.

## What Changes

- **New `internal/i18n` package**: `Lang` type with `LangEN`/`LangES` constants and `Default = LangEN`; `Parse(s) (Lang, error)`; `T(lang, key) string` (fallback EN → key itself); `Tf(lang, key, args...) string`; `LoadPreferred(homeDir) Lang` (reads `~/.active-stack/config.json` key `"language"`, provided for the future TUI, unused by the engine). Two string tables `table_en.go` / `table_es.go` keyed by namespaces: `windows.install.*`, `windows.uninstall.*`, `windows.phase.*`, `windows.step.*`, `windows.dryrun.*`, `component.fallback.*`, `mode.{lite,full,custom}.{label,desc,long}`, `tier.{estricto,balanceado,bypass}.{label,desc,long}` + `tier.bypass.warning`, `strategy.{targeted,restore}.{label,desc,long}`, `backup.source.*`.
- **`--lang es|en` flag on all `windows` subcommands** in `cmd/active-stack/windows_dispatch.go` (options, uninstall-options, starters list, starters install, backups list, backups actions, install, uninstall). Default `en`. An invalid value MUST exit non-zero.
- **`headless.ParsedFlags` and `ParsedUninstallFlags` gain `Lang i18n.Lang`** (zero value → `Default`), threaded into every localized helper.
- **Migrate the ~50 hardcoded EN literals** in `headless/windows.go`, `windows_uninstall_options.go`, and `windows_backups.go` (Source label) to the i18n tables. `internal/backup` is NOT touched.
- **Tier language-mix FIX**: `en` → Strict/Balanced/Bypass with EN descriptions; `es` → Estricto/Balanceado/Bypass with ES descriptions. Tier IDs (`estricto`/`balanceado`/`bypass`) NEVER change — they are contract-stable.
- **Additive `LongDescription string json:"long_description,omitempty"`** field on `windowsModeOption`, `windowsPermissionTier`, and `windowsUninstallStrategy` (filled from the i18n `.long` keys — placeholder-quality text is acceptable here; final content lands in change L3). The same field is added to `windowsComponentOption` and `windowsStarterEntry` but left EMPTY (filled by change L2 from the catalog).
- **Default `en` keeps JSON output byte-compatible** for callers not passing `--lang`, so every existing headless/windows test stays green (e.g. `windows_test.go` still asserts the first mode label is `"Quick"`).

## Capabilities

### New Capabilities
<!-- No new capability specs. internal/i18n is engine infrastructure exercised through the two existing contracts below; its behavior (Parse, fallback, key parity) is covered by tasks/tests, not by a standalone user-facing capability spec. -->

### Modified Capabilities
- `windows-install-contract`: `windows options` and `windows install` gain a `--lang es|en` flag (default `en`, invalid → non-zero exit); the `permission_tiers` array gains an additive `long_description` field and its labels/descriptions become language-consistent (fixing the ES-label/EN-description mix); `modes` entries gain additive `long_description`; `custom_components`/`forced_components` gain an additive `long_description` field (empty in this change).
- `windows-hub-contract`: `windows uninstall-options`, `windows starters list`, `windows starters install`, `windows backups list`, and `windows backups <action>` gain the `--lang es|en` flag; `strategies` and `modes` entries gain additive `long_description`; the backup `source` label is localized via the i18n tables with a raw fallback; starter entries gain an additive `long_description` field (empty in this change).

## Impact

- **New code**: `internal/i18n/` (`i18n.go`, `table_en.go`, `table_es.go`, tests).
- **Modified code**: `cmd/active-stack/windows_dispatch.go` (flag wiring on all windows subcommands), `cmd/active-stack/headless/flags.go` + `uninstall_flags.go` (`Lang` field), `cmd/active-stack/headless/windows.go` (literal migration, tier fix, `LongDescription` on structs), `windows_uninstall_options.go` (modes + strategies migration), `windows_backups.go` (localized source label).
- **JSON contract**: purely additive (`long_description,omitempty`); default `en` preserves byte-compatibility for existing callers.
- **NOT touched**: `internal/backup`, `internal/install`, `internal/uninstall`, `internal/tui`, `internal/catalog`, `internal/model` (the catalog-side `LocalizedText` work is change L2).
- **Downstream**: the GUI language page (L4) will pass `--lang` from a language-selection screen; the future TUI can reuse `internal/i18n` and `LoadPreferred`.
