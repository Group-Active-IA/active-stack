# windows-gui-localization Specification

## Purpose
TBD - created by archiving change gui-language-page. Update Purpose after archive.
## Requirements
### Requirement: Pure UI string table resolves localized Host literals

A static `UiStrings` type SHALL expose a PURE `Get(lang, key)` function that returns the Host UI literal for the given language and key, with NO mutable global state — the caller (the shell) owns the current language and passes it in. The table SHALL cover the ~85 Host literals: the shell footer actions (`shell.next`/`shell.back`/`shell.cancel`), every page's title/subtitle (`page.*.title`/`page.*.subtitle`), hub entries plus their tooltips and the coming-soon label, review headings (Assistants / Install type / Components / Permission tier), PageTemplates headers ("Recommended", "Choose a backup to restore", "No backups yet.", Restore/Rename/Delete/Confirm/Cancel, "Browse…", the "{0} harnesses" format, "Target project", starter review headings, uninstall confirm headings), the backups warnings, the uninstall modification warning, the complete-state labels, and the progress format strings (e.g. `progress.installing_fmt`). When a key is missing for the requested language the function SHALL fall back to the English value, and when missing in English SHALL return the key itself. Localization MUST NOT be implemented with `.resx`/satellite assemblies (they break `PublishSingleFile`); the tables are plain in-code dictionaries.

#### Scenario: A known key resolves in the requested language

- **WHEN** `UiStrings.Get("es", "shell.next")` is called and the Spanish table defines that key
- **THEN** it returns the Spanish value, and `UiStrings.Get("en", "shell.next")` returns the English value

#### Scenario: Missing translation falls back to English then to the key

- **WHEN** `Get` is called for a language that lacks a key that English defines
- **THEN** it returns the English value; and when neither language defines the key, it returns the key string verbatim

### Requirement: English and Spanish tables have identical, non-empty key sets

The `en` and `es` string tables SHALL define exactly the same set of keys, and every value SHALL be non-empty. A parity test (in Core.Tests) SHALL assert `keys(en) == keys(es)` and that no value is empty, so a literal added in only one language fails the test.

#### Scenario: Key parity is enforced

- **WHEN** the parity test compares the `en` and `es` key sets
- **THEN** the sets are equal, with no key present in one language but absent in the other, and no empty value in either table

### Requirement: Language preference persists to the user config, merge-friendly

A `LanguagePreference` type SHALL Load and Save the selected language from `~/.active-stack/config.json` as `{"language":"<lang>"}`. Save SHALL preserve any other keys already present in that file (merge, not overwrite the whole document), and Load SHALL return the stored language when present and a null/empty result when the file or key is absent. The home directory SHALL be injectable so the round-trip is testable against a temp directory.

#### Scenario: Round-trips the language while preserving other keys

- **WHEN** `Save("es")` runs against a `config.json` that already contains an unrelated key
- **THEN** the file afterwards contains `"language":"es"` AND the pre-existing unrelated key is preserved, and a subsequent `Load` returns `"es"`

#### Scenario: Absent preference loads as empty

- **WHEN** `Load` runs against a home directory with no `config.json` (or a `config.json` without a `language` key)
- **THEN** it returns a null/empty result rather than throwing

### Requirement: Preselected language follows persisted preference, then OS culture, then English

The initial language SHALL be resolved in priority order: the persisted `LanguagePreference` when present; otherwise the OS UI culture (`CultureInfo.CurrentUICulture.TwoLetterISOLanguageName == "es"` → Spanish, any other value → English); otherwise English. The OS-culture reader SHALL be injectable so the detection is deterministic in tests. The `ACTIVE_STACK_UI_LANG` environment variable, when set to a supported value, SHALL be honored via `BootstrapperLaunchOptions` as an automation override.

#### Scenario: OS culture selects Spanish when no preference is stored

- **WHEN** no preference is persisted and the injected OS culture reports `es-AR`
- **THEN** the preselected language is Spanish; and when the injected culture reports `fr`, the preselected language is English

#### Scenario: Persisted preference wins over OS culture

- **WHEN** a preference of `en` is persisted and the injected OS culture reports `es-AR`
- **THEN** the preselected language is English (the persisted preference overrides the OS culture)

#### Scenario: Environment override is honored for automation

- **WHEN** `ACTIVE_STACK_UI_LANG` is set to a supported value
- **THEN** `BootstrapperLaunchOptions` surfaces that language as the override for the automated run

