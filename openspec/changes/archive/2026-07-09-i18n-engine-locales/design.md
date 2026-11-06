## Context

The Windows installer engine (`cmd/active-stack/headless/windows*.go`) is the JSON-emitting backend the WPF GUI drives. Today it hardcodes ~50 English literals across phase messages, step messages, mode/tier/strategy labels and descriptions, and component-description fallbacks. There is no i18n layer in any package. The permission-tier screen is already language-inconsistent: `windowsPermissionTiers()` returns Spanish labels (`"Estricto"`, `"Balanceado"`, `"Bypass"`) paired with English descriptions.

A localized GUI (ES + EN) is planned as change L4; it needs the engine to return every user-facing string in a caller-selected language. This change (L1 of an approved 5-change plan) builds the Go-side foundation only: a reusable `internal/i18n` package, a `--lang` flag on the `windows` command surface, and the literal migration. The catalog-side localization (`model.LocalizedText`, `harnesses.yaml`) is change L2; the GUI page is L4; long-description content authoring is L3.

Constraints (from project CLAUDE.md §3 and the approved plan):
- JSON contract changes MUST be additive (`long_description,omitempty`). Default `en` keeps existing output byte-compatible so all current tests stay green.
- Contract IDs (`estricto`/`balanceado`/`bypass`, `lite`/`full`/`custom`, `targeted`/`restore`) NEVER change — only labels/descriptions localize.
- `internal/backup` MUST NOT be modified; the backup source label is localized at the engine boundary with a raw fallback.
- Strict TDD. A key-parity test between ES and EN is a first-class requirement, not an afterthought.

## Goals / Non-Goals

**Goals:**
- A standalone `internal/i18n` package with `Lang`, `Parse`, `T`, `Tf`, `LoadPreferred`, and two string tables, reusable by the future TUI.
- `--lang es|en` on every `windows` subcommand; `en` default; invalid value → non-zero exit.
- Every user-facing literal in `headless/windows.go`, `windows_uninstall_options.go`, and the backup source label in `windows_backups.go` sourced from the tables.
- The tier language-mix fixed: labels and descriptions consistent within each language.
- Additive `long_description` plumbed onto the mode/tier/strategy structs (placeholder content) and present-but-empty on component/starter structs.
- A key-parity test guaranteeing ES and EN tables have identical key sets — the anti-drift seatbelt.

**Non-Goals:**
- Catalog localization (`model.LocalizedText`, `harnesses.yaml`) — change L2.
- The GUI language page, `UiStrings.cs`, persistence UI — change L4.
- Authoring the final long-description prose — change L3 (placeholders here).
- A `--lang auto` / OS-detection mode in the engine — the GUI owns preselection; the engine default is a deterministic `en`.
- Touching `internal/tui`, `internal/backup`, `internal/install`, `internal/uninstall`, `internal/model`, `internal/catalog`.

## Decisions

### D1 — Table layout: flat `map[string]string` per language, namespaced keys

`table_en.go` and `table_es.go` each expose a package-level `map[string]string` (e.g. `tableEN`, `tableES`) registered into an internal `tables map[Lang]map[string]string`. Keys are dot-namespaced strings grouped by surface:

| Namespace | Covers | Example keys |
|---|---|---|
| `windows.install.*` | install stream literals | `windows.install.starting`, `windows.install.finished_ok`, `windows.install.finished_fail` |
| `windows.uninstall.*` | uninstall stream literals | `windows.uninstall.starting`, `windows.uninstall.finished_ok`, `windows.uninstall.finished_fail`, `windows.uninstall.plan_failed` |
| `windows.phase.*` | install phase messages | `windows.phase.prepare`, `windows.phase.apply`, `windows.phase.rollback`, `windows.phase.default` |
| `windows.uninstall.phase.*` | uninstall phase messages | `windows.uninstall.phase.prepare`, `.apply`, `.rollback`, `.default` |
| `windows.step.*` | per-step status messages | `windows.step.started`, `.succeeded`, `.failed`, `.rolled_back`, `.degraded` |
| `windows.dryrun.*` | dry-run step messages | `windows.dryrun.planned`, `windows.dryrun.listed` |
| `component.fallback.*` | component-description fallbacks | `component.fallback.security`, `.external`, `.config`, `.skill`, `.generic_fmt` (the Sprintf `"Installs %s."`) |
| `mode.{lite,full,custom}.{label,desc,long}` | install-mode options | `mode.lite.label`, `mode.lite.desc`, `mode.lite.long` |
| `tier.{estricto,balanceado,bypass}.{label,desc,long}` + `tier.bypass.warning` | permission tiers | `tier.estricto.label`, `tier.bypass.warning` |
| `strategy.{targeted,restore}.{label,desc,long}` | uninstall strategies | `strategy.targeted.label`, `strategy.restore.long` |
| `backup.source.*` | localized backup source labels | `backup.source.install`, `backup.source.sync`, `backup.source.upgrade` |

Rationale: a flat map is the simplest thing that supports both a compile-time-free key set and a runtime parity check. `.long` values here are placeholders (2-3 sentences); L3 refines them. Alternative considered — Go structs per surface — rejected because it makes the parity test structural instead of a simple set-equality over keys, and complicates the L3 content pass.

### D2 — `T` / `Tf` fallback chain, `Parse`, and `Default`

- `type Lang string`; `const LangEN Lang = "en"`, `LangES Lang = "es"`; `const Default = LangEN`.
- `Parse(s string) (Lang, error)`: accepts `"en"`/`"es"` (lower-cased, trimmed); anything else returns a non-nil error naming the value. Empty string is NOT valid at the flag boundary (the flag defaults to `"en"` before parse), but callers that hold a zero-value `Lang` treat it as `Default` (see D3).
- `T(lang, key) string`: look up `key` in `tables[lang]`; if the language is unknown or the key is missing, fall back to `tables[Default]`; if still missing, return `key` itself (so a stray key is visible, never a blank). A zero-value `lang` resolves through `Default`.
- `Tf(lang, key, args ...any) string`: `fmt.Sprintf(T(lang, key), args...)` — used for `component.fallback.generic_fmt` (`"Installs %s."`).
- `LoadPreferred(homeDir string) Lang`: reads `<homeDir>/.active-stack/config.json`, decodes `{"language": "..."}`, `Parse`s it, returns `Default` on any error/absence. Provided for the future TUI; the engine itself does NOT call it (the GUI passes `--lang` explicitly). Rationale: keeps the persistence contract co-located with the language type so L4/TUI share it.

### D3 — Threading `Lang` through the call graph

- `headless.ParsedFlags` and `ParsedUninstallFlags` each gain `Lang i18n.Lang`. A zero value is treated as `i18n.Default` at every use site (helpers call a tiny `langOrDefault(l)` or rely on `T`'s zero-value handling), so existing tests that construct these structs without setting `Lang` keep producing EN output.
- Every localized helper takes `lang i18n.Lang` as a parameter rather than reading a global: `windowsPhaseMessage(lang, stage)`, `windowsUninstallPhaseMessage(lang, stage)`, `mapProgressEvent(lang, e)`, `windowsComponentDescription(lang, h)`, `windowsPermissionTiers(lang)`, `windowsModeOptions(lang)`, `installFinishedMessage(lang, exitCode)`, and the strategy list in `RunWindowsUninstallOptions`. The inline literals in `runWindowsPipeline` / `RunWindowsUninstall` (e.g. `"Starting installation."`, dry-run messages) read from `params.Lang`.
- `RunWindowsOptions`, `RunWindowsUninstallOptions`, `RunWindowsStartersList`, `RunWindowsBackupsList` gain a `lang i18n.Lang` parameter, passed by the dispatch.
- Dispatch wiring: `windows_dispatch.go` registers `--lang` on each subcommand's `FlagSet`, defaulting to `"en"`, `Parse`s it, and on error prints to `w` and returns non-zero. For `install`/`uninstall`/`starters install` the parsed lang is assigned onto the `ParsedFlags`/`ParsedUninstallFlags`/`ParsedStarterAddFlags` (starters install threads it into the pipeline params). No global state; `i18n` is a pure function table.

### D4 — Backup source label localization without touching `internal/backup`

`windows_backups.go` currently sets `Source: m.Source.Label()`. The label is one of `"install"`, `"sync"`, `"upgrade"`, or `"unknown source"`. The engine maps it through `i18n.T(lang, "backup.source."+m.Source.Label())`; because `T` falls back to the key itself when absent, an unmapped label (e.g. `"unknown source"`, for which we deliberately register no key) renders as the raw string — the required raw fallback. `internal/backup` is untouched.

### D5 — Tier language-mix fix

`windowsPermissionTiers(lang)` reads label/desc/long/warning from the `tier.*` keys. In EN: `Strict`/`Balanced`/`Bypass` with EN descriptions. In ES: `Estricto`/`Balanceado`/`Bypass` with ES descriptions. IDs remain `estricto`/`balanceado`/`bypass`. The default-EN output changes the two mislabeled labels (`"Estricto"`→`"Strict"`, `"Balanceado"`→`"Balanced"`, `"Bypass"` stays) — this is the intended correctness fix; the current EN descriptions are preserved as the EN `.desc` values so only the labels move under default EN.

### D6 — Parity test as a first-class requirement

A test in `internal/i18n` asserts `keys(tableEN) == keys(tableES)` as sets, failing with the symmetric difference (missing-in-ES and missing-in-EN) named. This is the anti-drift seatbelt: any string added in one language but not the other breaks the build. Complemented by `Parse` round-trip/invalid tests, `T` fallback tests (missing key → EN → key), and `Tf` formatting tests.

## Risks / Trade-offs

- **[Default-EN output shifts for tier labels]** → The tier fix intentionally changes `"Estricto"`/`"Balanceado"` to `"Strict"`/`"Balanced"` under default EN. Mitigation: this is the correctness goal, not a regression; any GUI test fixture asserting the old ES labels under EN belongs to L4 and is updated there. The `windows_test.go:85` assertion (first mode label `"Quick"`) is unaffected because mode labels were already EN.
- **[Key drift between tables]** → Mitigated by the D6 parity test, which is a hard build gate.
- **[Placeholder `.long` content ships now]** → Acceptable by design; L3 refines with user review. The structs and keys exist so L4/L5 can wire the detail panel against a stable shape.
- **[Zero-value `Lang` ambiguity]** → Mitigated by treating zero as `Default` everywhere (`T` handles it), so struct literals in existing tests need no change.
- **[Sprintf key misuse]** → `component.fallback.generic_fmt` is the only format string; it is exercised by a dedicated `Tf` test so a malformed verb is caught.

## Migration Plan

Additive and backward-compatible; no data migration. Rollout is a normal build. Rollback is reverting the change — callers not passing `--lang` are unaffected because default `en` reproduces prior output (modulo the intended tier-label fix). No feature flag needed.

## Open Questions

- None blocking. Final `.long` prose is deferred to L3 by design; component/starter `long_description` population is deferred to L2.
