# Spec Delta

## MODIFIED Requirements

### Requirement: The catalog declares an opt-in `design-mockups` harness

The embedded catalog SHALL contain a global-picker harness with id `design-mockups`, type `config`, supported agents limited to `claude` and `codex`, and a dependency on `sdd-orchestrator`. Its `description` and `long_description` MUST be present in both `es` and `en`, and the catalog MUST still load under the existing bilingual validation. The harness MUST be selectable in Custom mode when either supported agent is selected and MUST NOT be offered when only unsupported agents are selected.

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog is loaded and `design-mockups` is resolved by ID
- **THEN** loading succeeds, its type is `config`, it supports `claude` and `codex` only, and it depends on `sdd-orchestrator`

#### Scenario: Offered in Custom for Claude

- **WHEN** the Custom harness list is built for `claude` or `codex`
- **THEN** the list contains `design-mockups`

#### Scenario: Not offered for non-Claude agents

- **WHEN** the Custom harness list is built for `opencode` only
- **THEN** the list does not contain `design-mockups`

#### Scenario: Selecting it pulls in the orchestrator

- **WHEN** a Custom install selects `design-mockups` without `sdd-orchestrator`
- **THEN** the resolved plan includes `sdd-orchestrator` as a dependency, ordered before `design-mockups`

### Requirement: Installation injects a dedicated section for each supported agent

Installing `design-mockups` SHALL write the agent-specific design stage into the supported agent's instructions file inside its own idempotent `<!-- active-stack:design-mockups -->` section, separate from `sdd-orchestrator`. A backup of the existing file MUST be taken before writing. Unsupported agents' instructions and settings MUST remain unchanged, and no primary agent named `design-mockups` is registered.

#### Scenario: Claude receives the section

- **WHEN** installed for `claude`
- **THEN** Claude's instructions file contains exactly one section with the local HTML review flow and optional Artifact Design instructions, and a backup is taken before modifying an existing file

#### Scenario: Codex receives its section

- **WHEN** installed for `codex`
- **THEN** Codex's `AGENTS.md` contains exactly one section with the local HTML/CSS flow, without Claude-only Artifact or model instructions

#### Scenario: Other selected agents are untouched

- **WHEN** installed with `claude`, `codex` and `opencode`
- **THEN** OpenCode's instructions and settings are unchanged

#### Scenario: Existing orchestrator section is preserved

- **WHEN** installed into a supported agent's instructions file containing `sdd-orchestrator` and user content
- **THEN** the existing section and user content are unchanged

### Requirement: Post-install verification checks the harness's own section

Post-install verification SHALL check that the `design-mockups` marker appears exactly once in each selected supported agent's instructions file and MUST NOT produce checks for unsupported agents. Verification of `sdd-orchestrator` MUST keep checking its own marker.

#### Scenario: Verify passes for a correct install

- **WHEN** verification runs after installing for `claude`, `codex` and `opencode`
- **THEN** there is one passing design marker check for Claude, one for Codex, and none for OpenCode

#### Scenario: Verify fails when the section is missing

- **WHEN** verification runs for a supported agent whose file contains only the orchestrator section
- **THEN** the design check fails naming the missing marker

### Requirement: Visual changes automatically enter design

When `design-mockups` is installed, the section SHALL direct the agent to enter design automatically for changes involving pages, screens, components, layout, styling or interactions. It SHALL use a supplied design as input, record the decision in the change's `design.md`, and reuse a recorded decision for the same revision. Nonvisual changes SHALL continue without a mockup. Only an explicit user request may skip mockups for visual work.

#### Scenario: Visual change starts design

- **WHEN** a change contains visual work and no mockup decision is recorded
- **THEN** the agent starts design without asking whether mockups are needed

#### Scenario: Backend-only change is not asked

- **WHEN** a change has no visual work
- **THEN** the agent records `Decision: no (no UI)` and continues

#### Scenario: Recorded decision is reused

- **WHEN** `design.md` records a decision for the same mockup revision
- **THEN** the agent honours it without repeating the question

### Requirement: Visual changes produce a portable review bundle before implementation

For visual changes with `design-mockups` installed, Claude and Codex SHALL automatically enter design after propose. Both SHALL create self-contained responsive HTML/CSS under the change's `mockups/` directory, covering affected screens and states before production UI implementation. They SHALL render mobile and desktop previews when a browser or preview tool exists, save the captures under `mockups/previews/`, record whether preview was rendered or manual, present the captures and local HTML to the user, and wait for explicit approval before apply. A user-supplied design SHALL enter the same review flow. Only the user's explicit skip request may omit mockups for visual work. A revision SHALL reset approval to pending. Artifact Design, Figma and Stitch MAY supplement the local bundle, but SHALL NOT be required for review.

#### Scenario: Artifact flow is described

- **WHEN** Claude's injected section is inspected
- **THEN** it describes local HTML as the review format and names `quickstart` with intent `design` only as an optional available tool, without hardcoding a type URL

#### Scenario: Codex has the same review and approval contract

- **WHEN** Codex's injected section is inspected
- **THEN** it requires local HTML/CSS, mobile and desktop previews, presentation to the user and approval before apply, without Claude-only Artifact or `sonnet` instruction

#### Scenario: Browser preview is unavailable

- **WHEN** no browser or screenshot tool can render the mockup
- **THEN** the agent supplies local HTML with exact opening instructions, does not claim visual inspection, and still waits for approval

## RENAMED Requirements

- FROM: `### Requirement: Installation injects a dedicated section only for Claude`
- TO: `### Requirement: Installation injects a dedicated section for each supported agent`
- FROM: `### Requirement: Mockups are generated with the Artifact Design type, with a fallback`
- TO: `### Requirement: Visual changes produce a portable review bundle before implementation`
- FROM: `### Requirement: The orchestrator asks about mockups only for UI changes`
- TO: `### Requirement: Visual changes automatically enter design`
