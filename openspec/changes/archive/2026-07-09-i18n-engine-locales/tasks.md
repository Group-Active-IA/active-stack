# Tasks — i18n-engine-locales

> Strict TDD: for every group, write the failing test FIRST, then the minimum code to pass, then triangulate + refactor. Do NOT build unless explicitly asked. `internal/backup` / `internal/install` / `internal/uninstall` / `internal/tui` / `internal/catalog` / `internal/model` stay untouched.

## 1. `internal/i18n` core — types, Parse, T/Tf, LoadPreferred

- [x] 1.1 RED: `internal/i18n/i18n_test.go` — `Parse` accepts `"en"`/`"es"` (trim + lower-case), returns the matching `Lang`; `Parse("xx")` and `Parse("")` return a non-nil error naming the value. Assert `Default == LangEN`.
- [x] 1.2 GREEN: `internal/i18n/i18n.go` — `type Lang string`; `LangEN`/`LangES` consts; `const Default = LangEN`; `Parse(s string) (Lang, error)`.
- [x] 1.3 RED: test `T(lang, key)` fallback chain — present key in `es` returns ES value; missing key in `es` falls back to EN value; key absent in both returns the key string itself; zero-value `Lang` resolves through `Default` (EN).
- [x] 1.4 GREEN: implement `T(lang, key) string` over an internal `tables map[Lang]map[string]string`, with the EN→key fallback and zero-value handling.
- [x] 1.5 RED: test `Tf(lang, key, args...)` formats via the stored format string (use `component.fallback.generic_fmt` = `"Installs %s."` → `Tf(LangEN, key, "Foo")` == `"Installs Foo."`).
- [x] 1.6 GREEN: implement `Tf(lang, key, args ...any) string` as `fmt.Sprintf(T(lang,key), args...)`.
- [x] 1.7 RED: test `LoadPreferred(homeDir)` — reads `<home>/.active-stack/config.json` `{"language":"es"}` → `LangES`; missing file / malformed JSON / unknown value → `Default`. Use a temp home dir.
- [x] 1.8 GREEN: implement `LoadPreferred(homeDir string) Lang` (JSON decode + `Parse`, `Default` on any error). Not called by the engine; provided for the future TUI.

## 2. String tables + key-parity seatbelt (first-class)

- [x] 2.1 RED: `internal/i18n/parity_test.go` — assert `keySet(tableEN) == keySet(tableES)`, failing with the symmetric difference (missing-in-ES and missing-in-EN) named. This is the anti-drift gate.
- [x] 2.2 GREEN: `internal/i18n/table_en.go` and `table_es.go` — flat `map[string]string` for every namespace in design D1: `windows.install.*`, `windows.uninstall.*` (incl. `.phase.*`, `.plan_failed`, `.finished_ok`, `.finished_fail`), `windows.phase.*`, `windows.step.*`, `windows.dryrun.*`, `component.fallback.*` (incl. `generic_fmt`), `mode.{lite,full,custom}.{label,desc,long}`, `tier.{estricto,balanceado,bypass}.{label,desc,long}` + `tier.bypass.warning`, `strategy.{targeted,restore}.{label,desc,long}`, `backup.source.{install,sync,upgrade}`. Register both into the `tables` map. Both tables MUST have identical key sets.
- [x] 2.3 Fill EN tier values as `Strict`/`Balanced`/`Bypass` and preserve the CURRENT English tier descriptions as the EN `.desc` values (so default-EN output only changes the two mislabeled labels). Fill ES tier values as `Estricto`/`Balanceado`/`Bypass` with Spanish descriptions.
- [x] 2.4 Fill `mode.*.long`, `tier.*.long`, `strategy.*.long` with reasonable 2-3 sentence PLACEHOLDER text in both languages (final prose is change L3). Confirm parity test (2.1) is green.

## 3. Thread `Lang` into the flag structs

- [x] 3.1 RED: `headless/flags_test.go` — `ParseInstallFlags` with `--lang es` sets `ParsedFlags.Lang == i18n.LangES`; no `--lang` leaves `Lang` at zero (treated as `Default`); `--lang xx` returns a non-nil error naming the value.
- [x] 3.2 GREEN: add `Lang i18n.Lang` to `ParsedFlags`; register `--lang` (default `"en"`) in `ParseInstallFlags`, `Parse` it, assign; invalid → error.
- [x] 3.3 RED: `headless/uninstall_flags_test.go` — same three assertions for `ParseUninstallFlags` / `ParsedUninstallFlags.Lang`.
- [x] 3.4 GREEN: add `Lang i18n.Lang` to `ParsedUninstallFlags`; wire `--lang` in `ParseUninstallFlags`.

## 4. Migrate `windows.go` literals + tier fix + LongDescription structs

- [x] 4.1 RED: extend `headless/windows_test.go` — under default (no `--lang`) `windows options` first mode label stays `"Quick"` and `permissions` stays in `forced_components` (byte-compat guard); tier labels are now `Strict`/`Balanced`/`Bypass` under EN; each `permission_tiers`, `modes` entry carries a populated `long_description`.
- [x] 4.2 RED: add an ES case — `RunWindowsOptions(..., i18n.LangES)` yields Spanish mode/tier labels while `id` values stay `lite/full/custom`, `estricto/balanceado/bypass`.
- [x] 4.3 GREEN: add `LongDescription string \`json:"long_description,omitempty"\`` to `windowsModeOption`, `windowsPermissionTier`, `windowsComponentOption`; thread `lang i18n.Lang` into `RunWindowsOptions`, `windowsPermissionTiers(lang)`, `windowsComponentDescription(lang, h)`, `mapProgressEvent(lang, e)`, `windowsPhaseMessage(lang, stage)`, `windowsUninstallPhaseMessage(lang, stage)`, `installFinishedMessage(lang, exitCode)`. Source all literals from `i18n.T`/`Tf`. Populate mode/tier `LongDescription` from `.long` keys; leave component `LongDescription` empty.
- [x] 4.4 GREEN: migrate inline literals in `runWindowsPipeline` and `RunWindowsUninstall` (e.g. `"Starting installation."`, `"Starting uninstall."`, dry-run `windows.dryrun.*`, `"Uninstall plan failed."`/`"Uninstall failed."`/`"Uninstall finished successfully."`) to read from `params.Lang` via `i18n.T`.
- [x] 4.5 TRIANGULATE: assert the full install and uninstall event streams under both `en` and `es` (phase + step + finished messages localized; IDs and event `type` values unchanged).

## 5. Migrate `windows_uninstall_options.go` (modes + strategies)

- [x] 5.1 RED: `headless/windows_uninstall_options_test.go` — default EN keeps `targeted` `default:true`/`requires_manifest:false` and `restore` `requires_manifest:true`; each strategy and mode carries a populated `long_description`; `--lang es` localizes labels/descriptions while `id`/`requires_manifest` are unchanged.
- [x] 5.2 GREEN: add `LongDescription` to `windowsUninstallStrategy`; thread `lang` into `RunWindowsUninstallOptions(homeDir, lang, w)` and `windowsModeOptions(lang)`; source strategy/mode label/desc/long from `strategy.*`/`mode.*` keys.

## 6. Migrate `windows_backups.go` source label + starter long_description

- [x] 6.1 RED: `headless/windows_backups_test.go` — a backup with source label `install` under `--lang es` yields the Spanish `backup.source.install` value; a source label with no registered key falls back to the raw string. `internal/backup` untouched.
- [x] 6.2 GREEN: thread `lang` into `RunWindowsBackupsList(homeDir, lang, w)`; map `Source` via `i18n.T(lang, "backup.source."+m.Source.Label())` (raw fallback comes free from `T`'s key fallback).
- [x] 6.3 RED+GREEN: `headless/windows_starters_test.go` — `windows starters list` entries carry an empty `long_description` field; add `LongDescription string \`json:"long_description,omitempty"\`` to `windowsStarterEntry` (left empty here; L2 fills it). Thread `lang` into `RunWindowsStartersList(sc, lang, w)`.

## 7. Dispatch wiring — `--lang` on every windows subcommand

- [x] 7.1 RED: `windows_dispatch_test.go` — each of `options`, `uninstall-options`, `starters list`, `starters install`, `backups list`, `backups <action>`, `install`, `uninstall` accepts `--lang es`; `--lang xx` prints an error and returns non-zero; omitting `--lang` behaves as EN. Existing routes stay green.
- [x] 7.2 GREEN: in `runWindowsDispatch`, `runWindowsStartersDispatch`, `runWindowsBackupsDispatch`, register `--lang` (default `"en"`) on each `FlagSet`, `i18n.Parse` it, on error print to `w` + return non-zero, and pass the parsed `Lang` to the handler / assign onto `ParsedFlags`/`ParsedUninstallFlags`/`ParsedStarterAddFlags` (starters install threads it into the pipeline params).
- [x] 7.3 REFACTOR: extract a small `parseLangFlag(fs, args) (i18n.Lang, error)` helper if the per-subcommand wiring duplicates; keep dispatch behavior identical.

## 8. Verify (no build unless asked)

- [x] 8.1 Run `go test ./internal/i18n/... ./cmd/active-stack/...`; confirm all green, including the parity test and the pre-existing `windows_test.go` byte-compat assertions.
- [x] 8.2 Sanity-diff default-EN JSON output for `windows options`/`uninstall-options`/`backups list` against the pre-change output — only the tier labels (`Estricto`→`Strict`, `Balanceado`→`Balanced`) and the new `long_description` fields should differ.
- [x] 8.3 Confirm no edits landed in `internal/backup`, `internal/install`, `internal/uninstall`, `internal/tui`, `internal/catalog`, `internal/model`.
