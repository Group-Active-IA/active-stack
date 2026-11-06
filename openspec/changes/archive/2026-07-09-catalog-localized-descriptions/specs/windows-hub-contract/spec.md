## MODIFIED Requirements

### Requirement: Starter list entries carry an additive long description

The entries emitted by `windows starters list` SHALL resolve their `description` in the requested language from the starter's localized `Description` (`s.Description.Localized(lang)`) and SHALL populate their `long_description` field (`json:"long_description,omitempty"`) from the starter's localized `LongDescription` resolved in the requested language (`s.LongDescription.Localized(lang)`). Populating these fields MUST NOT change any existing starter-entry field name (`id`, `name`, `description`, `includes`, `harnesses`, `mcp_count`, `long_description`), so a client reading only the pre-existing fields keeps working. The starter `name` MUST remain non-localized.

#### Scenario: Starter description is localized from the catalog

- **WHEN** `windows starters list --lang es` emits its `starters` array
- **THEN** each entry's `description` is the Spanish catalog value for that starter
- **AND WHEN** the same command runs under default `en`
- **THEN** each entry's `description` is the English catalog value

#### Scenario: Starter long_description is populated from the catalog

- **WHEN** `windows starters list` emits its `starters` array
- **THEN** each entry's `long_description` is the starter's localized `LongDescription` resolved in the requested language AND all pre-existing starter fields retain their names and values

#### Scenario: Starter name is not localized

- **WHEN** `windows starters list` emits an entry
- **THEN** the `name` field is the catalog display name unchanged in every language
