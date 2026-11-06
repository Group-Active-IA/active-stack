## Why

The embedded catalog (`internal/catalog/harnesses.yaml`) is Spanish-only: every harness and starter `description` is a single ES string, and there is no long-form field. After L1 (`i18n-engine-locales`) made the Windows engine bilingual, the component and starter descriptions are the last hardcoded-Spanish surface — the GUI still shows Spanish component/starter text even under `--lang en`, and the `long_description` fields L1 added to the component and starter contract entries are wired but permanently empty. To finish the bilingual engine, the catalog itself must become bilingual and its localized text must flow through the Windows contract.

## What Changes

- **New `model.LocalizedText`** (`map[string]string`) with `Localized(lang)` (fallback `lang → en → es → ""`), `Validate(field, owner)` (requires non-empty `es` AND `en`), and a custom `UnmarshalYAML` that accepts both the scalar form (assigned to `es`, parse-time retro-compat during migration) and the `{es:, en:}` map form.
- **`model.Harness.Description` and `model.Starter.Description` become `LocalizedText`**, and both gain a new `LongDescription LocalizedText` field (`yaml:"long_description,omitempty"`).
- **`harnesses.yaml` migrates** every harness (~30) and starter (5) description to `{es:, en:}` blocks: existing ES text stays as `es:`, and a new short EN translation is written for each. Short-description translation is IN SCOPE. `long_description` gets a brief ES+EN placeholder wherever it is mandatory (reviewed long-form content is deferred to L3).
- **`catalog.validate()` enforces bilingual completeness**: `Description` mandatory (both languages) for ALL harnesses and starters; `LongDescription` mandatory (both languages) for global-picker harnesses (`!IsStarterOnly()`) and ALL starters, optional for starter-only harnesses (but if present, both languages). A missing translation is a loud build/release failure with a message like `catalog: harness "openspec" description missing language "en"`. A leftover legacy scalar (only `es` after parse) therefore breaks `Load()` — intended.
- **Windows contract propagation**: `windowsComponentDescription` resolves `h.Description.Localized(lang)`; the component-option `long_description` is filled from `h.LongDescription.Localized(lang)`; `windows starters list` emits localized `description` + `long_description`. No new JSON fields (L1 already added them) — this change only FILLS them.
- **TUI stays byte-identical**: its call sites are pinned to `Localized("es")` with a comment marking future TUI localization.

## Capabilities

### New Capabilities
- `catalog-localization`: the `LocalizedText` type contract (fallback + bilingual validation), the localized `Description`/`LongDescription` fields on `Harness`/`Starter`, and the catalog-load validation rules that make a missing/one-language translation a build failure.

### Modified Capabilities
- `windows-install-contract`: the component-option `long_description` (added-but-empty by L1) is now populated from the catalog's localized `LongDescription`, and the component `description` is resolved in the requested language from the catalog's localized `Description`.
- `windows-hub-contract`: the `windows starters list` entry `description` is resolved in the requested language and its `long_description` (added-but-empty by L1) is populated from the catalog.

## Impact

- Code: `internal/model/localized.go` (new), `internal/model/harness.go`, `internal/model/starter.go`, `internal/catalog/catalog.go` (validation), `internal/catalog/harnesses.yaml` (migration + EN translations), `cmd/active-stack/headless/windows.go` (`windowsComponentDescription`), `cmd/active-stack/headless/windows_starters.go` (`RunWindowsStartersList`), and TUI call sites (pinned to `es`).
- Depends on L1 (`i18n-engine-locales`) — the `long_description` contract fields and the `--lang` threading already exist.
- No JSON contract fields added or removed (additive fields landed in L1); IDs unchanged.
- Governance: `catalog`/`model` are LOW, but touching `harnesses.yaml` requires `catalog.Load()` to stay green at every step (invalid catalog = broken release).
