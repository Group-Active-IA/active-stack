# Spec Delta

## Purpose

Declares the third-party design skills (`emil-design-eng` and the v1 taste skill) as opt-in, Full-only, Claude-only skill harnesses pinned to an exact upstream commit, so the soft dependency described by `design-mockups` can be satisfied by the installer.

## ADDED Requirements

### Requirement: The catalog declares the `emil-design-eng` skill harness

The embedded catalog SHALL contain a global-picker harness with id `emil-design-eng`, type `skill`, sourced by `clone` from repository `emilkowalski/skills` at path `skills/emil-design-eng`, pinned to commit `e8a175de22ae1e49370fc144c1f3bb9aeedf988d`. It MUST be marked third-party and best-effort, and its `description` and `long_description` MUST be present in both `es` and `en`. The catalog MUST still load under the existing bilingual validation.

#### Scenario: Catalog resolves emil-design-eng

- **WHEN** the catalog is loaded and `emil-design-eng` is resolved by ID
- **THEN** loading succeeds, its type is `skill`, its source repository is `emilkowalski/skills`, its path is `skills/emil-design-eng`, its method is `clone`, its ref is `e8a175de22ae1e49370fc144c1f3bb9aeedf988d`, and it is third-party and best-effort

#### Scenario: Bilingual texts present

- **WHEN** `emil-design-eng` is resolved
- **THEN** its description and long description both have non-empty `es` and `en` texts

### Requirement: The catalog declares the v1 taste skill harness

The embedded catalog SHALL contain a global-picker harness with id `design-taste-frontend-v1` (equal to the upstream skill's declared name), type `skill`, sourced by `clone` from repository `Leonxlnx/taste-skill` at path `skills/taste-skill-v1`, pinned to commit `b482f7a970abb98c4108d4a9f761e458c64cefc8`. It MUST be marked third-party and best-effort, with bilingual `description` and `long_description`. The catalog MUST NOT contain any harness sourced from `senlindesign/taste-skill` or from the experimental `skills/taste-skill` path of `Leonxlnx/taste-skill`.

#### Scenario: Catalog resolves the taste skill

- **WHEN** the catalog is loaded and `design-taste-frontend-v1` is resolved by ID
- **THEN** its type is `skill`, its source repository is `Leonxlnx/taste-skill`, its path is `skills/taste-skill-v1`, its method is `clone`, its ref is `b482f7a970abb98c4108d4a9f761e458c64cefc8`, and it is third-party and best-effort

#### Scenario: Rejected taste sources are absent

- **WHEN** every skill harness in the catalog is inspected
- **THEN** none has source repository `senlindesign/taste-skill`, and none has repository `Leonxlnx/taste-skill` with path `skills/taste-skill`

### Requirement: Design skills are pinned to a full commit SHA

Both design skill harnesses SHALL declare their `ref` as a full 40-character lowercase hexadecimal commit SHA. A branch name, tag, `latest` or empty ref MUST NOT be used for them.

#### Scenario: Refs are full SHAs

- **WHEN** the refs of `emil-design-eng` and `design-taste-frontend-v1` are inspected
- **THEN** each matches exactly 40 lowercase hexadecimal characters

### Requirement: Design skills are Full-only and opt-in

Both design skill harnesses SHALL be part of the Full and Custom modes and MUST NOT be part of Lite. Adding them MUST raise the Full and Custom global harness counts from 16 to 18 and leave Lite at 6. They MUST NOT be starter-only.

#### Scenario: Mode membership

- **WHEN** the harnesses for each mode are listed
- **THEN** Full and Custom contain both design skills and Lite contains neither

#### Scenario: Mode counts

- **WHEN** the global (non-starter-only) harnesses are counted per mode
- **THEN** Lite has 6, Full has 18 and Custom has 18

### Requirement: Design skills are offered only for Claude

Both design skill harnesses SHALL support only the `claude` agent. They MUST be offered when the `claude` agent is selected and MUST NOT be offered when only non-Claude agents are selected.

#### Scenario: Offered for Claude

- **WHEN** the harness list is built for the `claude` agent
- **THEN** it contains `emil-design-eng` and `design-taste-frontend-v1`

#### Scenario: Not offered for other agents

- **WHEN** the harness list is built for `opencode` only (or `codex` only)
- **THEN** it contains neither design skill

### Requirement: Design skills and design-mockups stay independent

Neither design skill harness SHALL depend on any other harness, and `design-mockups` MUST NOT depend on either design skill. Selecting a design skill MUST NOT pull `design-mockups` or `sdd-orchestrator` into the plan, and selecting `design-mockups` MUST NOT pull in a design skill.

#### Scenario: No dependencies declared

- **WHEN** `emil-design-eng`, `design-taste-frontend-v1` and `design-mockups` are resolved
- **THEN** the two design skills have no dependencies and the dependencies of `design-mockups` contain neither design skill

#### Scenario: Selecting design-mockups alone

- **WHEN** a Custom install for `claude` selects only `design-mockups`
- **THEN** the resolved plan contains neither `emil-design-eng` nor `design-taste-frontend-v1`
