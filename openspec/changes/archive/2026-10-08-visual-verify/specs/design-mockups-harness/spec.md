# Spec Delta

## ADDED Requirements

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
