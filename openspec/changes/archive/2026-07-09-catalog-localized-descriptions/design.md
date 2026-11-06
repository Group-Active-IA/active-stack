## Context

L1 (`i18n-engine-locales`, implemented and green but not yet archived) delivered `internal/i18n` and threaded `--lang` (default `en`) through every `windows` subcommand. It also added `long_description` fields to `windowsComponentOption`, `windowsPermissionTier`, `windowsModeOption`, `windowsUninstallStrategy`, and `windowsStarterEntry`. The tier/mode/strategy long descriptions are populated from the i18n tables; the **component** and **starter** `long_description` fields were deliberately left EMPTY, and `windowsComponentDescription(lang, h)` / `RunWindowsStartersList(cat, lang, w)` already receive `lang` but ignore it for the catalog-sourced text — they read the single-string `h.Description` / `s.Description`.

The catalog (`internal/catalog/harnesses.yaml`) is Spanish-only. `model.Harness.Description` and `model.Starter.Description` are plain `string`. There is no long-form catalog field. This change makes the catalog bilingual and connects it to the already-wired contract fields, finishing the bilingual Windows engine.

Constraint: `catalog.Load()` MUST stay green after every step — an invalid catalog is a build/release failure (hard project rule). The migration must therefore keep the catalog parseable and valid throughout.

## Goals / Non-Goals

**Goals:**
- A reusable `model.LocalizedText` type with deterministic fallback and strict bilingual validation.
- `Harness`/`Starter` carry localized `Description` + new `LongDescription`.
- `harnesses.yaml` migrated to `{es:, en:}`; short EN translations written for all harnesses and starters.
- `catalog.Load()` fails loudly, naming harness/starter + missing language, when a translation is absent.
- Windows contract emits localized component/starter `description` and fills their `long_description` from the catalog.
- TUI output is byte-for-byte unchanged.

**Non-Goals:**
- Reviewed long-form description CONTENT (the 3–6 sentence copy) — that is L3. This change seeds mandatory `long_description` with brief placeholder text only.
- Any new JSON contract field (L1 already added them) or any ID change.
- TUI localization (call sites are pinned to `"es"`).
- Touching `internal/backup`, mode/tier/strategy i18n text, or the C# GUI.

## Decisions

### D1: `LocalizedText` as `map[string]string` with custom UnmarshalYAML
`type LocalizedText map[string]string` in `internal/model/localized.go`.
- `Localized(lang string) string`: return `m[lang]` if non-empty, else `m["en"]`, else `m["es"]`, else `""`. Deterministic, never panics on nil map.
- `Validate(field, owner string) error`: returns an error unless BOTH `m["es"]` and `m["en"]` are non-empty. Message style: `catalog: <owner> <field> missing language "en"`.
- `UnmarshalYAML(value *yaml.Node)`: accept a scalar node (assign to `m["es"]`, parse-time retro-compat so a half-migrated YAML still parses) OR a mapping node (decode into the map). This lets migration proceed incrementally while final validation still demands both languages — a leftover scalar parses but then fails `Validate` with a clear message (intended loud failure).

*Alternative considered:* a struct `{ES, EN string}`. Rejected — a map keeps the door open for a third language without touching the type, and matches the plan's `map[string]string` decision. The map's "missing key" is the exact condition validation checks.

### D2: Field changes on Harness/Starter
- `Harness.Description string` → `Description LocalizedText` (`yaml:"description,omitempty"`); add `LongDescription LocalizedText yaml:"long_description,omitempty"` (harness.go ~:132).
- `Starter.Description string` → `Description LocalizedText`; add `LongDescription LocalizedText yaml:"long_description,omitempty"` (starter.go ~:44).
- All call sites that read `.Description` as a string are updated mechanically to `.Description.Localized(lang)`. Where no lang is in scope (TUI), pin to `.Description.Localized("es")` with a `// TODO(i18n): localize TUI` comment so behavior is unchanged.

### D3: Validation rules in catalog.validate()
Per harness, in the existing loop:
- `h.Description.Validate("description", fmt.Sprintf("harness %q", h.ID))` — mandatory for ALL harnesses.
- If `!h.IsStarterOnly()`: `h.LongDescription.Validate("long_description", …)` — mandatory (global-picker harnesses feed the GUI detail panel).
- If `h.IsStarterOnly()` AND `h.LongDescription != nil` (present): validate both languages anyway (no half-translated optional field).

Per starter, in `validateStarters()`:
- `s.Description.Validate(...)` — mandatory.
- `s.LongDescription.Validate(...)` — mandatory for ALL starters.

Because `UnmarshalYAML` maps a bare scalar to `es` only, any un-migrated entry fails `Validate` on the missing `en` — the migration's own safety net.

### D4: Windows contract propagation
- `windowsComponentDescription(lang, h)`: the `h.Description != ""` branch becomes `h.Description.Localized(string(lang)) != ""` returning that localized value; the fallback branches (security/external/config/skill) are unchanged (still i18n-sourced).
- In `RunWindowsOptions`, set `item.LongDescription = h.LongDescription.Localized(string(lang))` on each component option.
- `RunWindowsStartersList`: drop the `_ = lang` no-op; set `Description: s.Description.Localized(string(lang))` and `LongDescription: s.LongDescription.Localized(string(lang))`.

### D5: Placeholder long descriptions
Mandatory `long_description` entries (all global-picker harnesses + all starters) get brief ES+EN placeholder text in this change so `Load()` passes. L3 replaces them with reviewed copy. Placeholders must still be real bilingual text (both keys non-empty) — the point is a green catalog, not final prose.

## Risks / Trade-offs

- [A half-migrated YAML scalar parses silently and only fails at Validate] → That is the intended design: the parse-time scalar acceptance keeps migration incremental, and the bilingual `Validate` is the gate. Tests assert the loud failure explicitly (fixtures: legacy scalar, missing `en`, missing long).
- [Changing `Description` from `string` to `LocalizedText` breaks every call site] → Compile errors are the safety net; a repo-wide sweep of `.Description` usages is part of the tasks, and the TUI pin keeps its output identical (asserted by existing TUI tests staying green).
- [Placeholder long descriptions could leak to users if L3 slips] → Acceptable: the GUI detail panel falls back to the short description anyway, and the placeholders are bilingual and coherent, not lorem ipsum.
- [MODIFIED requirements target L1-added requirements not yet in main specs] → L2 depends on L1; L1 archives first (dependency order), so the modified requirement headers exist in main specs by the time L2 archives.

## Migration Plan

1. Land `LocalizedText` + field changes with unit tests (RED first).
2. Add catalog validation with invalid-catalog fixtures (missing `en`, legacy scalar, missing long) proving loud failure.
3. Migrate `harnesses.yaml` incrementally (ES stays, EN added, placeholders for mandatory long), keeping `catalog.Load()` green.
4. Update Windows propagation + TUI pins.
5. Full `go test ./...` sweep.

Rollback: revert the commit; no persisted state or external side effects.
