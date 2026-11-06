## ADDED Requirements

### Requirement: windows hub subcommands accept a language flag

The `active-stack windows uninstall-options`, `windows starters list`, `windows starters install`, `windows backups list`, and `windows backups restore|delete|rename` commands SHALL each accept a `--lang` flag parsed by `i18n.Parse` with valid values `en` and `es`. When `--lang` is absent the language MUST default to `i18n.Default` (`en`), and the emitted JSON MUST stay byte-compatible with the pre-change output for that command. An unrecognized `--lang` value MUST cause the dispatch to print an error and return a non-zero exit code before the command's work runs. The resolved language MUST be threaded into every user-facing string the command emits, sourced from the `internal/i18n` tables. `internal/backup` MUST NOT be modified by this change.

#### Scenario: Default language keeps hub output byte-compatible

- **WHEN** `windows uninstall-options --home <tmp>` is run without `--lang`
- **THEN** the response is produced in English AND its pre-existing fields (`detected_agents`, `modes`, `strategies` with `targeted` default and `restore` requiring a manifest) retain their names and values so the existing hub-contract assertions stay green

#### Scenario: Spanish localizes hub strings

- **WHEN** `windows uninstall-options --home <tmp> --lang es` is run
- **THEN** the `modes` and `strategies` labels/descriptions are emitted in Spanish while the `id` values (`lite`/`full`/`custom`, `targeted`/`restore`) remain unchanged

#### Scenario: Invalid language is rejected on a hub subcommand

- **WHEN** any of the listed hub subcommands receives `--lang xx`
- **THEN** the dispatch prints an error naming the invalid value AND returns a non-zero exit code AND no work proceeds

### Requirement: Uninstall strategies and modes carry an additive long description

The `strategies` entries emitted by `windows uninstall-options` SHALL each source their `label` and `description` from the `strategy.*` i18n keys in the requested language and SHALL carry an additive `long_description` field (`json:"long_description,omitempty"`) sourced from the `strategy.*.long` key; its content MAY be placeholder-quality in this change. The `modes` entries emitted by `windows uninstall-options` (and shared with `windows options`) SHALL carry the same additive `long_description` from the `mode.*.long` key. The `targeted` strategy MUST remain `default: true` / `requires_manifest: false` and `restore` MUST remain `requires_manifest: true` in every language. Adding these fields MUST NOT change any existing field name or value.

#### Scenario: Strategies carry a populated long_description

- **WHEN** `windows uninstall-options` emits its `strategies` array
- **THEN** each strategy entry includes a `long_description` populated from the `strategy.*.long` i18n key AND `targeted` is still `default:true`/`requires_manifest:false` AND `restore` is still `requires_manifest:true`

#### Scenario: Localized strategy labels

- **WHEN** `windows uninstall-options --lang es` is served
- **THEN** the `targeted` and `restore` strategy labels/descriptions are Spanish while their `id` and `requires_manifest` values are unchanged

### Requirement: Starter list entries carry an additive long description

The entries emitted by `windows starters list` SHALL each expose a `long_description` field (`json:"long_description,omitempty"`) of additive shape; in this change that field is left EMPTY (it is populated from the catalog by a later change). Adding the field MUST NOT change any existing starter-entry field name or value (`id`, `name`, `description`, `includes`, `harnesses`, `mcp_count`), so a client reading only the pre-existing fields keeps working.

#### Scenario: Starter long_description is present but empty

- **WHEN** `windows starters list` emits its `starters` array
- **THEN** each entry includes a `long_description` field that is empty in this change AND all pre-existing starter fields retain their names and values

### Requirement: Backup source labels are localized with a raw fallback

The `source` field emitted per entry by `windows backups list` SHALL be localized by mapping the backup's `Source.Label()` value through `i18n.T(lang, "backup.source."+label)` in the requested language. When no i18n key exists for a given source label (for example `"unknown source"`), the value MUST fall back to the raw label string exactly as `internal/backup` returns it. `internal/backup` MUST NOT be modified — the localization happens only at the engine boundary in `windows_backups.go`.

#### Scenario: Known source is localized

- **WHEN** `windows backups list --home <tmp> --lang es` lists a backup whose source label is `install`
- **THEN** the entry's `source` value is the Spanish string mapped from `backup.source.install`

#### Scenario: Unknown source falls back to the raw label

- **WHEN** a listed backup has a source label with no registered i18n key
- **THEN** the entry's `source` value equals the raw label string returned by `Source.Label()` unchanged
