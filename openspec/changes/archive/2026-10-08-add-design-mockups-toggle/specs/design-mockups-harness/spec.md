# Spec Delta

## Purpose

Defines the `design-mockups` config harness: an opt-in, Claude-only harness that injects into the agent's instructions file its own section describing a human-approved UI mockup design stage between `propose` and `apply`, using Claude's Artifact "Design" capability. It also fixes the Claude orchestrator workflow diagram to include `verify`.

## ADDED Requirements

### Requirement: The catalog declares an opt-in `design-mockups` harness

The embedded catalog SHALL contain a global-picker harness with id `design-mockups`, type `config`, supported agents limited to `claude`, and a dependency on `sdd-orchestrator`. Its `description` and `long_description` MUST be present in both `es` and `en`, and the catalog MUST still load under the existing bilingual validation. The harness MUST be selectable in Custom mode whenever the `claude` agent is selected, and MUST NOT be offered when only non-Claude agents are selected.

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog is loaded and `design-mockups` is resolved by ID
- **THEN** loading succeeds, its type is `config`, it supports `claude` only, and it depends on `sdd-orchestrator`

#### Scenario: Offered in Custom for Claude

- **WHEN** the Custom harness list is built for the `claude` agent
- **THEN** the list contains `design-mockups`

#### Scenario: Not offered for non-Claude agents

- **WHEN** the Custom harness list is built for `opencode` only (or `codex` only)
- **THEN** the list does not contain `design-mockups`

#### Scenario: Selecting it pulls in the orchestrator

- **WHEN** a Custom install selects `design-mockups` without `sdd-orchestrator`
- **THEN** the resolved plan includes `sdd-orchestrator` as a dependency, ordered before `design-mockups`

### Requirement: Installation injects a dedicated section only for Claude

Installing `design-mockups` SHALL write the design stage content into the Claude agent's instructions file inside its own idempotent section delimited by `<!-- active-stack:design-mockups -->` markers, separate from the `sdd-orchestrator` section. A backup of the existing file MUST be taken before writing. Agents the harness does not support MUST NOT be modified, even when they are selected in the same install: their instructions files and settings files stay byte-for-byte unchanged, and no primary agent named `design-mockups` is registered anywhere.

#### Scenario: Claude receives the section

- **WHEN** `design-mockups` is installed with the `claude` agent
- **THEN** Claude's instructions file contains exactly one `design-mockups` section with the design stage heading, and a backup of the previous file exists

#### Scenario: Other selected agents are untouched

- **WHEN** `design-mockups` is installed with agents `claude` and `opencode` (or `claude` and `codex`)
- **THEN** the opencode/codex instructions and settings files are identical to their state before the install

#### Scenario: Existing orchestrator section is preserved

- **WHEN** `design-mockups` is installed into a Claude instructions file that already contains the `sdd-orchestrator` section and user content
- **THEN** the orchestrator section and the user content are unchanged and the `design-mockups` section is added

### Requirement: Installation is idempotent

Re-installing `design-mockups` SHALL NOT duplicate its section. When the content is unchanged, the instructions file MUST NOT be rewritten.

#### Scenario: Re-install does not duplicate

- **WHEN** `design-mockups` is installed twice for `claude`
- **THEN** the instructions file contains exactly one `design-mockups` section and the second install reports no change

### Requirement: The section is not purged by other config harnesses

The `design-mockups` section SHALL be treated as an owned section: installing, re-installing or uninstalling `sdd-orchestrator` (or any other config harness) MUST NOT remove it.

#### Scenario: Orchestrator re-install keeps the section

- **WHEN** `sdd-orchestrator` is installed into a Claude instructions file that already contains the `design-mockups` section
- **THEN** the `design-mockups` section is still present and unchanged

#### Scenario: Orchestrator uninstall keeps the section

- **WHEN** only `sdd-orchestrator` is uninstalled from a Claude instructions file containing both sections
- **THEN** the orchestrator section is removed and the `design-mockups` section remains

### Requirement: Uninstall removes the section cleanly

Uninstalling `design-mockups` SHALL remove its section (markers and content) from the Claude instructions file, leaving the `sdd-orchestrator` section and user content intact, with a snapshot taken before the change. Uninstalling when the section is absent MUST be a no-op.

#### Scenario: Clean removal

- **WHEN** `design-mockups` is uninstalled from a Claude instructions file containing it
- **THEN** no `design-mockups` marker remains and the rest of the file is unchanged

#### Scenario: Removal when absent is a no-op

- **WHEN** `design-mockups` is uninstalled and the section is not present
- **THEN** the file is not modified and no error is returned

### Requirement: Post-install verification checks the harness's own section

Post-install verification for `design-mockups` SHALL check that the `design-mockups` marker appears exactly once in the Claude instructions file, and MUST NOT produce checks for agents the harness does not support. Verification of `sdd-orchestrator` MUST keep checking its own marker.

#### Scenario: Verify passes for a correct install

- **WHEN** verification runs after installing `design-mockups` for `claude` and `opencode`
- **THEN** there is one passing check for Claude's `design-mockups` marker and no check for opencode

#### Scenario: Verify fails when the section is missing

- **WHEN** verification runs for `design-mockups` and Claude's instructions file contains only the orchestrator section
- **THEN** the `design-mockups` check fails naming the missing marker

### Requirement: Orchestrator output for other agents is unchanged

Adding this harness SHALL NOT change the composed `sdd-orchestrator` block for `opencode`, `codex` or the generic fallback.

#### Scenario: OpenCode golden unchanged

- **WHEN** the orchestrator block is composed for `opencode` with all toggles
- **THEN** the output equals the existing opencode all-toggles golden

### Requirement: The design section places the stage between propose and apply

The `design-mockups` section SHALL instruct the orchestrator that the design stage runs after `propose` and before `apply`, giving the flow `explore → propose → design → apply → verify → archive`. It MUST NOT be placed before `propose` or during `apply`.

#### Scenario: Stage ordering is stated

- **WHEN** a reader inspects the injected `design-mockups` section
- **THEN** it describes the stage as running after `propose` and before `apply`

### Requirement: The orchestrator asks about mockups only for UI changes

The section SHALL instruct the orchestrator to ask "does this change need mockups?" only when the change's artifacts (proposal, specs, design, tasks) indicate user-interface work, such as pages, screens, components or forms. It MUST offer three answers: yes, no, and "I already have them". It MUST instruct that changes with no UI signal are not asked, that the answer is recorded in the change's `design.md`, and that a recorded answer is not asked again.

#### Scenario: Backend-only change is not asked

- **WHEN** the orchestrator follows the section for a change whose artifacts contain no UI signal
- **THEN** the section's instructions direct it to skip the mockups question

#### Scenario: Recorded decision is reused

- **WHEN** the change's `design.md` already records a mockups decision
- **THEN** the section's instructions direct the orchestrator not to ask again and to honour the recorded decision

### Requirement: Mockups are generated with the Artifact Design type, with a fallback

For a "yes" decision, the section SHALL instruct generating the mockups with the Artifact tool by first calling `quickstart` with intent `design` and creating the artifact from the returned type, following the instructions returned at creation time. It MUST NOT hardcode a type URL. When the Artifact tool is unavailable, it MUST instruct writing self-contained static HTML mockups instead.

#### Scenario: Artifact flow is described

- **WHEN** a reader inspects the injected section
- **THEN** it names `quickstart` with intent `design` and the static HTML fallback

### Requirement: Installed design skills are used when available

The section SHALL contain a single conditional instruction: if design skills such as `emil-design-eng` (Emil Kowalski) or a taste skill are installed, the orchestrator loads them for the design step. This MUST be a soft dependency: the stage MUST work without those skills, and the harness MUST NOT require or install them.

#### Scenario: Soft dependency is stated

- **WHEN** a reader inspects the injected section
- **THEN** it names `emil-design-eng` and states the skills are used only if installed

### Requirement: Mockups live inside the change and act as visual acceptance criteria

The section SHALL instruct that mockups (or references to them, such as the Artifact URL and any exported copy) are stored under `openspec/changes/<name>/mockups/`. They MUST be referenced from the change's `design.md` (under a `## UI Mockups` section) and `tasks.md`, and they MUST be treated as visual acceptance criteria during `apply` and `verify`.

#### Scenario: Location and references are stated

- **WHEN** a reader inspects the injected section
- **THEN** it names `openspec/changes/<name>/mockups/` as the location, the `## UI Mockups` record in `design.md`, and requires references from `design.md` and `tasks.md`

### Requirement: Explicit human approval gates apply

When the recorded decision is "yes" or "I already have them", the section SHALL instruct the orchestrator not to start `apply` until the user has explicitly approved the mockups and that approval is recorded in the change. The section MUST state that this gate is a deliberate exception to the general "never block on phase gates" rule.

#### Scenario: Apply blocked without approval

- **WHEN** a UI change has decision "yes" and no recorded approval
- **THEN** the section's instructions direct the orchestrator to refuse to start `apply` and to request approval first

#### Scenario: Decision "no" does not gate

- **WHEN** a change has the recorded decision "no"
- **THEN** the section's instructions allow `apply` to proceed with no mockup approval

### Requirement: Claude OPSX workflow diagram includes verify

The Claude orchestrator block SHALL show the OPSX workflow diagram with a `verify` step between `apply` and `archive`, regardless of whether `design-mockups` is installed.

#### Scenario: Diagram shows verify

- **WHEN** the orchestrator block is composed for variant `claude` with no toggles
- **THEN** the OPSX workflow diagram lists a verify step between apply and archive
