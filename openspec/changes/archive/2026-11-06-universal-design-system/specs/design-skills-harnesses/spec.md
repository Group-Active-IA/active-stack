# Spec Delta

## MODIFIED Requirements

### Requirement: Design skills are offered only for Claude

The pinned optional `emil-design-eng` and `design-taste-frontend-v1` skills SHALL be offered in Full and Custom for Claude, Codex and OpenCode, using each agent's skill directory. They SHALL remain soft dependencies of `design-mockups` and SHALL NOT appear in Lite.

#### Scenario: OpenCode Full selection

- **WHEN** Full harnesses are selected for OpenCode
- **THEN** both pinned skills are in the selection, independent of whether an individual best-effort clone later succeeds

#### Scenario: Offered for Claude

- **WHEN** the harness list is built for Claude
- **THEN** it contains `emil-design-eng` and `design-taste-frontend-v1`

#### Scenario: Not offered for other agents

- **WHEN** the harness list is built for OpenCode or Codex
- **THEN** it contains both design skills; unsupported agents receive neither

#### Scenario: Offered for Codex

- **WHEN** the harness list is built for `codex`
- **THEN** it contains both design skills
