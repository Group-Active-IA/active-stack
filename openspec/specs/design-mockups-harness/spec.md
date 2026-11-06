# design-mockups-harness Specification

## Purpose
Defines the `design-mockups` config harness: an opt-in, Claude-only harness that injects into the agent's instructions file its own section describing a human-approved UI mockup design stage between `propose` and `apply`, using Claude's Artifact "Design" capability. It also fixes the Claude orchestrator workflow diagram to include `verify`.

## Requirements

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

Post-install verification SHALL check that the `design-mockups` marker appears exactly once in each selected supported agent's instructions file and MUST NOT produce checks for unsupported agents. Verification of `sdd-orchestrator` MUST keep checking its own marker.

#### Scenario: Verify passes for a correct install

- **WHEN** verification runs after installing for `claude`, `codex` and `opencode`
- **THEN** there is one passing design marker check for Claude, one for Codex, and none for OpenCode

#### Scenario: Verify fails when the section is missing

- **WHEN** verification runs for a supported agent whose file contains only the orchestrator section
- **THEN** the design check fails naming the missing marker

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

### Requirement: Verify includes a visual check against approved mockups

The `design-mockups` section SHALL instruct that, during `verify`, a visual verification step runs when the change's `design.md` records under `## UI Mockups` a decision of `yes` or `provided` together with `Approval: approved`. When Playwright MCP tools are available, the step MUST direct the agent to:

- start the project's development server;
- open the local URL of each screen covered by the mockups;
- capture each screen at a mobile viewport and at a desktop viewport;
- inspect the browser console errors and the failed network requests;
- compare the captures against the mockups in `openspec/changes/<name>/mockups/`;
- stop the development server it started.

#### Scenario: Visual step is described with both viewports

- **WHEN** a reader inspects the injected `design-mockups` section
- **THEN** it describes a visual verification step during `verify` that names Playwright, the development server, a mobile and a desktop capture, console errors, failed network requests, and comparison against `mockups/`

#### Scenario: Step applies only to approved UI mockups

- **WHEN** a change records `Decision: no`, or records `yes`/`provided` without `Approval: approved`
- **THEN** the section's instructions direct the agent not to run the visual verification step for that change

### Requirement: Visual differences are reported by priority with reproduction steps

The visual verification step SHALL instruct reporting each difference between the built UI and the approved mockups with a priority level, the screen and viewport, the expected result (mockup reference), the observed result, and reproduction steps. It MUST state that the agent never claims a screen matches or works unless it actually opened and checked it, and that the report lists which screens and viewports were checked and which were not.

#### Scenario: Report format and honesty rule are stated

- **WHEN** a reader inspects the injected section
- **THEN** it requires prioritized differences with reproduction steps and states that nothing is claimed to work without having been checked

### Requirement: Without Playwright the step degrades to an explicit manual checklist

When Playwright MCP tools are not available, the visual verification step SHALL instruct producing a manual checklist (each mockup screen × mobile and desktop) for the user, stating explicitly that the visual verification was not automated. It MUST also suggest installing the `playwright` harness. The `design-mockups` harness MUST NOT depend on the `playwright` harness: it SHALL install and work without it.

#### Scenario: Degradation is explicit

- **WHEN** a reader inspects the injected section
- **THEN** it describes the manual checklist fallback, requires saying explicitly that the check was not automated, and suggests the `playwright` harness

#### Scenario: No hard dependency on Playwright

- **WHEN** the catalog is loaded and `design-mockups` is resolved by ID
- **THEN** its dependencies contain `sdd-orchestrator` and do not contain `playwright`

### Requirement: The visual verification outcome is recorded in the change

The visual verification step SHALL instruct recording its outcome in the change's `design.md` under `## UI Mockups` as a `Visual verify:` line with one of `passed`, `differences` or `manual` and the date. Captures MUST be stored or referenced under `openspec/changes/<name>/mockups/verify/`.

#### Scenario: Outcome record is stated

- **WHEN** a reader inspects the injected section
- **THEN** it names the `Visual verify:` record with the values `passed | differences | manual` and the `mockups/verify/` location for captures

### Requirement: The extended section keeps a single owned section

Adding the visual verification step SHALL NOT introduce a new instructions-file section or marker: the content lives inside the existing `active-stack:design-mockups` section. Re-installing `design-mockups` over a file that has the previous section content MUST replace it in place, leaving exactly one `design-mockups` section, and the content MUST NOT contain absolute web URLs.

#### Scenario: Upgrade replaces the section in place

- **WHEN** `design-mockups` is installed into a Claude instructions file that contains an older `design-mockups` section
- **THEN** the file contains exactly one `design-mockups` section whose content includes the visual verification step, and the orchestrator section and user content are unchanged

#### Scenario: No new marker is introduced

- **WHEN** a reader inspects the injected section after install
- **THEN** no marker other than `active-stack:design-mockups` was added by `design-mockups`, and the section contains no `https://` URL

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
