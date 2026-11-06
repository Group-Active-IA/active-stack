# Spec Delta

## MODIFIED Requirements

### Requirement: Design skills are offered only for Claude

Both design skill harnesses SHALL support `claude` and `codex` only. They MUST be offered when either agent is selected and MUST NOT be offered when only unsupported agents are selected.

#### Scenario: Offered for Claude

- **WHEN** the harness list is built for `claude`
- **THEN** it contains `emil-design-eng` and `design-taste-frontend-v1`

#### Scenario: Offered for Codex

- **WHEN** the harness list is built for `codex`
- **THEN** it contains both design skills

#### Scenario: Not offered for other agents

- **WHEN** the harness list is built for `opencode` only
- **THEN** it contains neither design skill
