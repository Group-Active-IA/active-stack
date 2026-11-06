# Spec Delta

## MODIFIED Requirements

### Requirement: The catalog declares an opt-in `design-mockups` harness

The `design-mockups` config harness SHALL support Claude, Codex and OpenCode. It SHALL be selected automatically by Full, available in Custom, and absent from Lite. Its instructions SHALL use local HTML/CSS and SHALL NOT require or direct Claude Artifact Design.

#### Scenario: Three-agent Full install

- **WHEN** Full is selected for Claude, Codex and OpenCode
- **THEN** the plan includes `design-mockups` and the installed design section appears once in each agent's instructions file

#### Scenario: OpenCode section is shared

- **WHEN** `design-mockups` installs for OpenCode
- **THEN** its shared `AGENTS.md` receives the design section, while `opencode.json` does not gain a `design-mockups` primary agent

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog is loaded and `design-mockups` is resolved by ID
- **THEN** it is a `config` harness for Claude, Codex and OpenCode and depends on `sdd-orchestrator`

#### Scenario: Offered in Custom for Claude

- **WHEN** the Custom picker is built for Claude, Codex or OpenCode
- **THEN** it contains `design-mockups`

#### Scenario: Not offered for non-Claude agents

- **WHEN** the picker is built only for an agent outside the three supported agents
- **THEN** it does not offer `design-mockups`

#### Scenario: Selecting it pulls in the orchestrator

- **WHEN** Custom selects `design-mockups` without selecting `sdd-orchestrator`
- **THEN** the orchestrator is resolved and ordered before the design harness

### Requirement: Installation injects a dedicated section for each supported agent

Installing `design-mockups` SHALL inject the same portable visual section into Claude, Codex and OpenCode instructions files using the existing idempotent marker and backup path. OpenCode's section SHALL be shared `AGENTS.md` guidance and SHALL NOT create a `design-mockups` primary agent in `opencode.json`. Reinstallation SHALL leave exactly one section.

#### Scenario: OpenCode install follows SDD install

- **WHEN** Full installs `sdd-orchestrator` and then `design-mockups` for OpenCode
- **THEN** the primary SDD agent remains in `opencode.json` and one visual section appears in shared `AGENTS.md`

#### Scenario: Claude receives the section

- **WHEN** installed for `claude`
- **THEN** Claude's instructions file contains exactly one section with the local HTML/CSS review flow, and a backup is taken before modifying an existing file

#### Scenario: Codex receives its section

- **WHEN** installed for `codex`
- **THEN** Codex's `AGENTS.md` contains exactly one section with the local HTML/CSS flow, without Claude-only Artifact or model instructions

#### Scenario: Other selected agents are untouched

- **WHEN** installed with `claude`, `codex`, `opencode` and an unsupported agent
- **THEN** the unsupported agent's instructions and settings are unchanged

#### Scenario: Existing orchestrator section is preserved

- **WHEN** installed into a supported agent's instructions file containing `sdd-orchestrator` and user content
- **THEN** the existing section and user content are unchanged

## ADDED Requirements

### Requirement: A project-wide design system precedes visual mockups

For a visual change, the agent SHALL inspect project-root `design-system/` before producing change mockups. If no approved design-system revision exists, the agent SHALL create or audit the project-wide visual foundation, present its HTML/CSS preview and wait for explicit user approval. It SHALL NOT start per-change mockups until the foundation approval is recorded. Nonvisual changes do not require it.

#### Scenario: New project

- **WHEN** the first visual change begins without `design-system/`
- **THEN** the agent creates and presents the foundation, records `Approval: pending`, and stops before change mockups

#### Scenario: Existing app

- **WHEN** the first visual change begins in an app with existing brand and UI conventions
- **THEN** the agent audits those conventions and proposes a system that preserves them for review

#### Scenario: Later visual change

- **WHEN** an approved design-system revision already exists
- **THEN** the agent reuses it and records its path and revision in the change's `design.md` before creating mockups

### Requirement: The design system is portable and cheap to revise

The project SHALL use `design-system/README.md`, `brand.md`, `brand.js`, `tokens.css`, `components.html`, `assets/` and `previews/` as the default structure. The README SHALL record revision, approval, sources and open `(TBD)` decisions. Unresolved name/logo/imagery MAY remain `(TBD)` with clearly provisional, stable local assets or variables. The displayed brand name SHALL come from the shared `brand.js` in HTML previews so it can be replaced in one place. Mockups SHALL reuse shared tokens and assets where practical. A change to the approved system SHALL identify affected mockups and require renewed review if their appearance changes.

#### Scenario: Logo remains unresolved

- **WHEN** the user approves the foundation with `logo: (TBD)`
- **THEN** the preview marks the logo provisional and a later logo can replace the stable asset path without regenerating unrelated tokens or components

#### Scenario: Shared token changes

- **WHEN** an approved color or typography token changes
- **THEN** affected mockup previews are regenerated and presented for review before their implementation continues

### Requirement: Visual verification is available to all three agents

The Playwright MCP harness and the two pinned optional design skills SHALL be offered to OpenCode as well as Claude and Codex. An unavailable browser SHALL cause truthful manual preview instructions, not a false claim of rendered verification.

#### Scenario: OpenCode has browser and design skills

- **WHEN** Full is selected for OpenCode
- **THEN** Playwright MCP and both optional design skills are included in its install plan
