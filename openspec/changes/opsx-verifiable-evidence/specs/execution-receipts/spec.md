# Spec Delta

## Purpose

Define the execution receipt a sub-agent uses to back what it claims to have run, and how the coordinator verifies its provenance and sufficiency.

## ADDED Requirements

### Requirement: Execution receipt contract

Every return from a sub-agent that claims to have executed a command SHALL include one receipt per executed command with these fields: revision (commit and whether the working tree had uncommitted changes), directory, command, result (exit code and timestamp), covered requirements, and a reference to the original output with its fingerprint and whether it was truncated.

#### Scenario: Complete receipt

- **WHEN** a sub-agent ran tests on the change's revision and returns a receipt with all six fields and the original output available
- **THEN** the receipt is eligible for provenance verification

#### Scenario: Incomplete receipt

- **WHEN** a receipt lacks one or more fields
- **THEN** it is evaluated as "unverified" and the coordinator MUST NOT infer the missing field

#### Scenario: One receipt per command

- **WHEN** a requirement is covered by more than one command
- **THEN** each command has its own receipt and each receipt references the requirement

### Requirement: Delivery of the contract to the sub-agent

The orchestrator SHALL include the evidence contract in the prompt of every delegation that executes commands, in both Claude and Codex. When the module that carries the contract is not available in the session, the coordinator MUST report the limitation and stop the action that requires it.

#### Scenario: Delegation that runs commands

- **WHEN** the orchestrator builds the prompt for a delegation that runs tests or a build
- **THEN** the prompt contains the evidence contract with the six receipt fields

#### Scenario: Contract module unavailable

- **WHEN** the module that carries the evidence contract is not loaded in the session
- **THEN** the coordinator reports the limitation and does not delegate the command execution

### Requirement: Provenance verification

The coordinator SHALL verify that each receipt corresponds to a real execution on the change's revision and directory before accepting it. The check MUST compare the revision and directory, confirm the original output exists, recompute its fingerprint without content filters, and check that the declared exit code is consistent with that output.

#### Scenario: Invented receipt

- **WHEN** a receipt's original output does not exist or its fingerprint does not match
- **THEN** the receipt is rejected

#### Scenario: Stale receipt

- **WHEN** a receipt was issued on a revision different from the current one
- **THEN** the receipt is rejected as outdated

#### Scenario: Foreign directory

- **WHEN** a receipt was executed outside the repository or worktree of the change
- **THEN** the receipt is rejected

#### Scenario: Same fingerprint across platforms

- **WHEN** the same output file is fingerprinted on Windows and on Linux
- **THEN** both fingerprints are equal, regardless of line-ending conversion settings

### Requirement: Sufficiency review

The system SHALL evaluate separately whether what was executed is enough to cover each declared requirement, and MUST report provenance and sufficiency results independently.

#### Scenario: Insufficient command

- **WHEN** a receipt passes provenance but its command does not exercise the declared requirement
- **THEN** the requirement is not considered covered

#### Scenario: Green tests with unmet requirement

- **WHEN** a receipt reports a successful result and a declared requirement is not met by the code
- **THEN** the receipt is rejected even though the tests passed

### Requirement: Status and default to unverified

Each requirement SHALL end in exactly one status: verified, rejected or unverified. Anything that cannot be evaluated MUST be unverified, and a requirement with no receipts MUST NOT be verified.

#### Scenario: Missing field defaults to unverified

- **WHEN** a field is missing or the original output cannot be read
- **THEN** the status is "unverified", not "verified"

#### Scenario: Zero receipts

- **WHEN** a report covers a requirement and contains no receipts
- **THEN** the requirement is "unverified" and the report is not presented as verified

### Requirement: Original output access and sensitive data

A truncated or bounded output SHALL keep a reference to the complete original. Full original outputs MUST NOT be committed to the repository by default, and any output that is shared outside the local machine MUST be redacted first. The receipt MUST record the fingerprint of the original and, when redacted, the fingerprint of the redacted version.

#### Scenario: Truncated output without original

- **WHEN** a receipt's output was truncated and has no reference to the original
- **THEN** it is evaluated as "unverified"

#### Scenario: Output shared with the group

- **WHEN** an original output is about to be shared outside the local machine
- **THEN** secrets, environment variable values and home-directory paths are redacted first and both fingerprints are recorded

#### Scenario: Output not versioned by default

- **WHEN** a receipt is saved under the change's evidence folder
- **THEN** the receipt text is versionable and the full output stays out of version control

### Requirement: Parity between Claude and Codex

The Claude and Codex assets SHALL produce the same contract and the same verdicts over the same set of test receipts, and Codex MUST include a verification step for receipts in its flow.

#### Scenario: Stable composition

- **WHEN** the orchestrator is composed with all toggles for each agent and compared with its golden file
- **THEN** the evidence section matches and any divergence fails the test

#### Scenario: Same verdicts

- **WHEN** the same battery of valid, incomplete, invented, stale and foreign-directory receipts is evaluated with the Claude flow and the Codex flow
- **THEN** both flows return the same verdict for every receipt
