## Purpose

Covers the catalog entries that let `active-stack` offer and install the Discovery/competitor-research phase of the JR Stack foundation flow: `discovery-research` and its `web-scraper` dependency.

## ADDED Requirements

### Requirement: Catalog includes a `discovery-research` harness

The embedded catalog SHALL define a harness with `id: discovery-research`, `type: skill`, `source.repo: Group-Active-IA/discovery-research`, `source.method: clone`, `install_modes: [full]`, and `depends_on: [web-scraper]`. It MUST NOT be `scope: starter-only` and MUST NOT be referenced by any starter's `harnesses` list.

#### Scenario: discovery-research resolves via ByID

- **WHEN** `catalog.ByID("discovery-research")` is called after `Load()`
- **THEN** it returns the harness with `Type == "skill"`, `Source.Repo == "Group-Active-IA/discovery-research"`, and `DependsOn` containing `"web-scraper"`

#### Scenario: discovery-research appears in Full mode, not Lite

- **WHEN** `catalog.ForMode(ModeFull)` and `catalog.ForMode(ModeLite)` are called after `Load()`
- **THEN** `discovery-research` is present in the Full-mode result and absent from the Lite-mode result

### Requirement: Catalog includes a `web-scraper` harness

The embedded catalog SHALL define a harness with `id: web-scraper`, `type: skill`, `source.repo: Group-Active-IA/web-scraper`, `source.method: clone`, `install_modes: [full]`. It MUST NOT be `scope: starter-only` and MUST NOT be referenced by any starter's `harnesses` list.

#### Scenario: web-scraper resolves via ByID

- **WHEN** `catalog.ByID("web-scraper")` is called after `Load()`
- **THEN** it returns the harness with `Type == "skill"` and `Source.Repo == "Group-Active-IA/web-scraper"`

#### Scenario: web-scraper appears in Full mode, not Lite

- **WHEN** `catalog.ForMode(ModeFull)` and `catalog.ForMode(ModeLite)` are called after `Load()`
- **THEN** `web-scraper` is present in the Full-mode result and absent from the Lite-mode result

### Requirement: The catalog loads without error with both entries present

`catalog.Load()` SHALL succeed with `discovery-research` and `web-scraper` present — both provide a non-empty `Source.Repo`, a valid `Source.Method`, at least one valid `InstallModes` entry, and complete bilingual `description` and `long_description` fields (both harnesses are not starter-only, so `long_description` is mandatory per the existing bilingual-completeness requirement).

#### Scenario: Load succeeds with both new entries

- **WHEN** the embedded `harnesses.yaml` includes the `discovery-research` and `web-scraper` entries as specified above
- **THEN** `catalog.Load()` returns no error

#### Scenario: discovery-research's dependency resolves

- **WHEN** `catalog.Load()` validates `DependsOn` references for every harness
- **THEN** `discovery-research`'s `depends_on: [web-scraper]` resolves to an existing harness ID, so validation does not fail on a dangling dependency
