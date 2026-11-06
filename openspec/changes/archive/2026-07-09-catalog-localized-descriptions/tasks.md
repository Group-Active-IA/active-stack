## 1. Safety net & baseline

- [x] 1.1 Run `go test ./...` and record the green baseline (L1 just passed; note counts for model, catalog, headless packages).
- [x] 1.2 Grep the repo for every read of `Harness.Description` / `Starter.Description` as a string (TUI, planner, headless, install) to size the mechanical sweep before changing the type.

## 2. LocalizedText type (RED → GREEN → TRIANGULATE → REFACTOR)

- [x] 2.1 RED: write `internal/model/localized_test.go` for `Localized(lang)` fallback — requested-language hit, `en` fallback, `es` fallback, nil/empty map returns `""` (no panic). At least a happy path + edge case per behavior.
- [x] 2.2 GREEN: add `internal/model/localized.go` with `type LocalizedText map[string]string` and `Localized`.
- [x] 2.3 RED: add tests for `Validate(field, owner)` — both languages pass; only `es` fails naming the missing `en`; only `en` fails naming the missing `es`; empty map fails.
- [x] 2.4 GREEN: implement `Validate` with the message style `catalog: <owner> <field> missing language "<lang>"`.
- [x] 2.5 RED: add tests for `UnmarshalYAML` — scalar node populates only `es`; mapping node populates both keys.
- [x] 2.6 GREEN: implement `UnmarshalYAML` accepting scalar (→ `es`) and mapping nodes.
- [x] 2.7 REFACTOR: dedupe/clean; tests stay green.

## 3. Harness/Starter localized fields

- [x] 3.1 RED: add model tests parsing a harness and a starter with `description: {es,en}` + `long_description: {es,en}`, asserting `Localized` per language.
- [x] 3.2 GREEN: change `Harness.Description` to `LocalizedText` and add `LongDescription LocalizedText` (harness.go ~:132).
- [x] 3.3 GREEN: change `Starter.Description` to `LocalizedText` and add `LongDescription LocalizedText` (starter.go ~:44).
- [x] 3.4 Fix all non-Windows call sites the type change breaks: TUI reads pinned to `.Description.Localized("es")` with a `// TODO(i18n): localize TUI` comment; planner/install/any other reads updated so the package compiles. (Only `internal/tui/starters_screen.go` + its test fixture needed a fix; planner/install/verify/backup have no `Harness.Description`/`Starter.Description` reads. Windows call sites deferred to Group 6 as designed.)

## 4. Catalog validation (fixtures FIRST — loud failure)

- [x] 4.1 RED: in `internal/catalog`, add table-driven fixture tests via the internal `parse([]byte)` entry — a valid bilingual fixture loads; a harness `description` with only `es` fails naming harness + `description` + `en`; a legacy scalar description fails on the missing `en`; a global-picker harness missing `long_description` fails; a starter-only harness with no `long_description` loads; a starter missing `long_description` fails.
- [x] 4.2 GREEN: extend `catalog.validate()` — per-harness `Description.Validate` (all) and `LongDescription.Validate` (mandatory when `!IsStarterOnly()`, optional-but-both-if-present otherwise); per-starter `Description.Validate` + `LongDescription.Validate` in `validateStarters()`.
- [x] 4.3 TRIANGULATE: confirmed each invalid fixture reports the correct owner/field/language. Checks are placed AFTER all pre-existing structural checks (per-harness and per-starter) so an entry invalid for another reason still reports that error first; fixed ~15 pre-existing `parse()`-based fixtures across `catalog_test.go`/`starter_test.go`/`resolve_mcps_test.go` that omitted description/long_description and would otherwise have failed for the wrong reason (or the right reason with the wrong message). `Load()`-based tests intentionally still red — the embedded `harnesses.yaml` migration is Group 5.

## 5. Migrate harnesses.yaml (keep Load() green)

- [x] 5.1 Convert every harness `description` to `{es: <existing>, en: <new short translation>}` (all 43 harnesses — proposal estimated ~30; the actual catalog has 13 global-picker + 30 starter-only).
- [x] 5.2 Convert every starter `description` to `{es: <existing>, en: <new short translation>}` (4 starters — proposal estimated 5; the actual catalog has base/backend/ux-ui/active-ia).
- [x] 5.3 Add bilingual placeholder `long_description: {es,en}` for every global-picker harness (`!IsStarterOnly()`) and every starter (brief, coherent — L3 replaces with reviewed copy). Starter-only harnesses left without `long_description` (optional per D3).
- [x] 5.4 Run `go test ./internal/catalog/...` — real `catalog.Load()` passes; no scalar leftovers. (Fixed a YAML gotcha found along the way: unquoted flow-mapping values containing commas broke `{ es: ..., en: ... }` parsing — every ES/EN value is now double-quoted.)

## 6. Windows contract propagation

- [x] 6.1 RED: extend `windows_test.go` (or a sibling) — `windowsComponentDescription`/`RunWindowsOptions` under `--lang es` vs `en` emit the localized component `description` and a non-empty `long_description` from the catalog; a no-catalog-description component still hits the type fallback. (Also fixed ~9 pre-existing `model.Harness{Description: "..."}` string literals broken by the type change.)
- [x] 6.2 GREEN: update `windowsComponentDescription(lang, h)` to use `h.Description.Localized(string(lang))` and set `item.LongDescription = h.LongDescription.Localized(string(lang))` in `RunWindowsOptions` (windows.go).
- [x] 6.3 RED: extend starters-list test — `RunWindowsStartersList` under `--lang es` vs `en` emits localized `description` + populated `long_description`; `name` unchanged across languages. (Updated `twoStarterFakeCatalog` fixture + the pre-existing `TestRunWindowsStartersList_CatalogOrder` assertion that expected `long_description` empty — that was L1's documented interim state, now superseded by L2.)
- [x] 6.4 GREEN: in `RunWindowsStartersList` remove the `_ = lang` no-op and set `Description`/`LongDescription` via `Localized(string(lang))` (windows_starters.go).

## 7. Verification sweep

- [x] 7.1 Run `go test ./...` — full suite green, matching or exceeding the 1.1 baseline; TUI tests confirm byte-identical output. All 25 packages `ok` (baseline was also 25/25; `e2e/starter`'s in-memory fixture builder needed bilingual description/long_description added too — same D3 consequence as the catalog test fixtures).
- [x] 7.2 Manual check via `go run ./cmd/active-stack windows options --agent claude --lang es|en` and `windows starters list --lang es|en`: component/starter `description`+`long_description` localize correctly; `id`/`name`/`harnesses`/`mcp_count`/`includes` unchanged across languages.
- [x] 7.3 Confirmed no JSON contract field was added/removed and no harness/starter `id` changed: `git diff` on `harnesses.yaml` shows zero `- id:` line changes; the `long_description` struct fields already existed from L1 — L2 only populates them.
