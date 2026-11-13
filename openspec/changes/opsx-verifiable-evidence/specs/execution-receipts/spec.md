# Spec Delta

## Purpose

Define the execution receipt used to back what is claimed to have been run, and how the coordinator checks its integrity, provenance and sufficiency. In this capability, "execution receipt" is the only term used. It is distinct from the "installation record" (registro de instalación) of proposal 04, which tracks files owned by the installer and is not evidence of executed commands. Where proposals 03, 04 and 07 write "receipts" (for example the `receipts` field of the return envelope), they mean execution receipts.

## ADDED Requirements

### Requirement: Execution receipt contract

Every result that claims a command was executed SHALL include one execution receipt per executed command with these fields: run identifier; tree and exact state tested (the commit and, when there are uncommitted changes, a fingerprint of the full state, because HEAD plus a "dirty" flag is not enough); command; result (exit code and timestamp); covered requirements; and a reference to the redacted copy of the output with exactly one fingerprint of that copy (algorithm named in the field), whether it was truncated, and a capture record. Only the redacted copy, with one fingerprint, and the capture record are persisted (agreed). The independent capture mechanism per interface (defined by proposal 05) and the location, access and expiry of the redacted copy (to be agreed between proposals 05 and 06) are (TBD) and are not fixed by this capability.

#### Scenario: Complete execution receipt

- **WHEN** a command ran on the change's tree and the execution receipt has all fields and the redacted copy is available
- **THEN** the execution receipt is eligible for integrity and provenance checks

#### Scenario: Incomplete execution receipt

- **WHEN** an execution receipt lacks one or more fields
- **THEN** it is evaluated as "unverified" and the coordinator MUST NOT infer the missing field

#### Scenario: One execution receipt per command

- **WHEN** a requirement is covered by more than one command
- **THEN** each command has its own execution receipt and each references the requirement

#### Scenario: Same HEAD, different uncommitted state

- **WHEN** two runs happen on the same HEAD with two different uncommitted diffs (including a different new untracked file)
- **THEN** their state fingerprints differ, and the execution receipt of one MUST NOT be accepted for the other

### Requirement: Contract loaded when the action requires evidence

The orchestrator SHALL load the evidence contract whenever the action requires evidence (at least `apply` that runs tests, and `verify`), both when the coordinator executes directly and when it delegates, in Claude and in Codex, and both paths SHALL go through the same execution receipt review step. When the action is delegated, the brief of the delegation that executes commands SHALL also carry an excerpt of the contract (`## Evidence Contract`). The contract MUST NOT be a universal rule in `CLAUDE.md` or `AGENTS.md`, because everything there also reaches the specialists (proposal 03, D0). Loading MUST NOT depend on the `delegation` toggle: direct execution needs the contract even when delegation is off. This capability PROPOSES a dedicated module (`evidence`) loaded by the `apply` and `verify` actions, with the excerpt pasted in the brief; the proposal is pending agreement with proposals 03 and 04 and lists the exact changes it would imply in the shared rules matrix (design D5). Where the module file lives follows the mechanism chosen by 03 (support files under `sdd-orchestrator/modules/`, pasted into the brief by the coordinator). When the required module is not available in the session, the coordinator MUST report the limitation and stop the action that requires it.

#### Scenario: Direct apply

- **WHEN** the coordinator runs the tests of an `apply` itself, without delegating
- **THEN** execution receipts are produced and go through the same review step as in a delegation

#### Scenario: Delegation that runs commands

- **WHEN** the orchestrator builds the brief for a delegation that runs tests or a build
- **THEN** the brief contains the evidence contract excerpt with the execution receipt fields, and the worker returns the receipts in the `receipts` field of the return envelope

#### Scenario: Direct apply with delegation off

- **WHEN** the delegation toggle is off and the coordinator runs the tests of an `apply` itself
- **THEN** the evidence contract is still loaded and the execution receipts go through the review step

#### Scenario: Contract not placed in the shared instructions file

- **WHEN** the orchestrator is installed for Claude or Codex
- **THEN** the evidence contract is not written as a universal rule in `CLAUDE.md` or `AGENTS.md`

#### Scenario: Contract module unavailable

- **WHEN** the module that carries the evidence contract is not loaded in the session
- **THEN** the coordinator reports the limitation and does not run or delegate the action that requires evidence

### Requirement: Return envelope ownership

The return envelope of a delegation is owned by proposal 07 (`opsx-selective-delegation`, D6). This capability SHALL define only the content of the `receipts` field: a list of execution receipts, one per executed command. It MUST NOT define another envelope or state that the execution receipt is a section of the return format of proposal 04.

#### Scenario: Receipts field

- **WHEN** a worker returns after running commands
- **THEN** the envelope's `receipts` field holds one execution receipt per command, and nothing else in the envelope is defined here

### Requirement: Integrity, provenance and sufficiency are separate

The coordinator SHALL evaluate integrity, provenance and sufficiency separately and report the three results independently.

- Integrity: the stored redacted copy exists and its fingerprint, recomputed over its exact bytes, matches the one in the execution receipt.
- Provenance: the execution is real and corresponds to the state of the code. A result declared by the worker MUST remain "unverified" until an independent source or a re-execution corroborates the execution and the state of the code. A purely mechanical check (fingerprint plus the worker's account) MUST NOT be sufficient. The capture mechanism (harness or client source, before truncation) is (TBD) until proposal 05 shows which source covers each interface.
- Sufficiency: what was executed is enough to cover each declared requirement.

#### Scenario: Integrity alone is not enough

- **WHEN** an execution receipt passes integrity and no independent source or re-execution corroborates it
- **THEN** its status is "unverified"

#### Scenario: False execution receipt with correct fingerprint

- **WHEN** a worker fabricates an execution receipt whose fingerprint matches a stored copy it also fabricated
- **THEN** integrity passes, and the status is "unverified" without corroboration, or "rejected" if an independent source or a re-execution contradicts it

#### Scenario: Re-execution without independent capture

- **WHEN** the only corroboration is a re-execution by the coordinator without its own capture from an independent source
- **THEN** it does not corroborate (and does not contradict): the execution receipt stays "unverified", and a re-execution that has its own independent capture (proposal 05) does count

#### Scenario: Stale state

- **WHEN** corroboration shows the execution ran on a state of the tree different from the current one
- **THEN** the execution receipt is rejected as outdated

#### Scenario: Foreign directory

- **WHEN** an execution receipt was produced outside the repository or worktree of the change
- **THEN** its status is (TBD) between "rejected" and "unverified"; the starting proposal is "rejected", to be agreed with proposal 05, because the roadmap card does not list this case

#### Scenario: Invented execution receipt

- **WHEN** the stored copy does not exist or its fingerprint does not match, or an independent source or re-execution shows the command did not run or gave another result
- **THEN** the execution receipt is rejected as invented

#### Scenario: Insufficient command

- **WHEN** an execution receipt passes integrity and provenance but its command does not exercise the declared requirement
- **THEN** the execution receipt is rejected as insufficient and the requirement is not considered covered

#### Scenario: Green tests with unmet requirement

- **WHEN** an execution receipt reports success, passes integrity and provenance, and a declared requirement is not met by the code
- **THEN** the requirement is rejected even though the tests passed

### Requirement: Fingerprint of the stored copy

The fingerprint SHALL be the sha256 of the exact bytes of the redacted copy, with no line-ending conversion or content filters, and the field SHALL name the algorithm (for example `{algorithm: "sha256", value: "<hex>"}`), as proposals 02 and 05 do. A Git blob hash (`git hash-object`) MUST NOT be used as the definition: it hashes a Git object with a header, not the file digest. Identical bytes MUST produce the same fingerprint on every platform; files that differ in any byte, including line endings, MUST produce different fingerprints. The fingerprint of the tested state, when there are uncommitted changes, SHALL be aligned with `TreeHash(dir)` of `eval/baseline/fixture/load.go` (sha256 over the lines `<file sha256>  <path>` sorted by path); a temporary Git index was measured and is not used as the definition (it depends on `core.autocrlf` and can map LF and CRLF files to the same tree); the choice is a proposal of 06 pending confirmation (design D2, `informes/medicion-huella-2026-11-13.md`). Windows and Linux are the minimum platforms; a platform matrix SHALL distinguish distributed, declared and tested platforms.

#### Scenario: Same fingerprint across platforms

- **WHEN** the same file (identical bytes) is fingerprinted on Windows and on Linux, with and without `core.autocrlf`
- **THEN** both fingerprints are equal

#### Scenario: Line endings change the fingerprint

- **WHEN** two files differ only in line endings (LF versus CRLF)
- **THEN** their fingerprints differ

#### Scenario: Different bytes, different state fingerprint

- **WHEN** two working trees on the same HEAD differ in one file's bytes (including LF versus CRLF) or by one new untracked file
- **THEN** their state fingerprints differ

#### Scenario: Platform matrix

- **WHEN** the matrix is published
- **THEN** it lists each platform as distributed (`release.yml:71-76`), declared (`CLAUDE.md:21`) and tested separately, and Windows and Linux are tested

### Requirement: Status and default to unverified

Each requirement SHALL end in exactly one status: verified, rejected or unverified. Rejected covers, as the roadmap card (06, "Aceptación futura") says: invented or stale receipts, insufficient commands, truncated output without original, and unmet requirements with green tests. An incomplete receipt is unverified. Anything that cannot be evaluated MUST be unverified, and an execution with no execution receipts MUST NEVER be verified. Verified requires integrity, corroborated provenance and sufficiency.

#### Scenario: Missing field defaults to unverified

- **WHEN** a field is missing or the redacted copy cannot be read
- **THEN** the status is "unverified", not "verified"

#### Scenario: Zero execution receipts

- **WHEN** a report covers a requirement and contains no execution receipts
- **THEN** the requirement is "unverified" and the report is not presented as verified

### Requirement: Stored output and sensitive data

A truncated or bounded output SHALL keep a reference to the complete stored copy. Only the redacted copy MAY be persisted, and a full output MUST NOT appear in the working tree's `git status` after a run. Redaction, retention, access and format of the reference are defined by proposal 05 (redaction at capture); this capability MUST NOT define a second redaction policy. The execution receipt MUST record the reference and one fingerprint (of the redacted copy) and a capture record, and MUST NOT record a second fingerprint.

#### Scenario: Truncated output without original

- **WHEN** an execution receipt's output was truncated and it never had a reference to a redacted copy
- **THEN** it is rejected (roadmap card: "salidas truncadas sin original"); if the reference existed and the copy expired, see "Retention expired" (status (TBD), starting proposal: "unverified")

#### Scenario: Redacted output with a valid reference

- **WHEN** a command output contains a value of a sensitive class and is captured
- **THEN** the stored copy is already redacted, and the execution receipt records its reference, one fingerprint of that copy and the capture record, and integrity passes over that copy

#### Scenario: Redacted detail is not enough to verify

- **WHEN** verifying a requirement needs a detail that the redaction of proposal 05 discarded (requirement "Un detalle redactado no alcanza para verificar" of proposal 05, referenced and not redefined here)
- **THEN** the execution receipt is "unverified" and the unredacted original is not persisted to avoid that outcome

#### Scenario: Retention expired

- **WHEN** the stored copy is no longer available because the expiry (to be agreed with proposal 05) has passed
- **THEN** the execution receipt is never "verified"; whether it is "unverified" or "rejected" is (TBD), and the starting proposal is "unverified"

#### Scenario: No full output in git status

- **WHEN** a command with voluminous output runs and is captured
- **THEN** `git status` shows no file containing the full output, tracked or untracked

#### Scenario: Versioning of execution receipts

- **WHEN** an execution receipt text is saved
- **THEN** its location and whether it is versioned are (TBD) pending the group decision (design D6), and a full output is never versioned by default

### Requirement: Hooks complement evidence

Client hooks (proposal 08) MAY be used as a corroboration source, but they SHALL NOT be treated as an exhaustive record of executions. The absence of a hook trace MUST NOT by itself verify or reject an execution receipt.

#### Scenario: No hook trace

- **WHEN** an execution receipt has no matching hook trace
- **THEN** its status is decided by the other sources and controls, not by the missing trace alone

### Requirement: Parity between Claude and Codex

The Claude and Codex assets SHALL produce the same contract and the same verdicts over the same set of test execution receipts, in direct and delegated execution. Codex `verify` with a reviewer different from the builder is covered by proposal 04 (`specs/codex-opsx-delivery/spec.md`, requirement "Verify con independencia obligatoria", and `internal/harness/config/assets/codex/native/SKILL.md`, on `origin/propose/codex-native-opsx`, pull request pending); this capability does not add that action.

#### Scenario: Stable composition

- **WHEN** the orchestrator is composed with all toggles for each agent and compared with its golden file
- **THEN** the evidence section matches and any divergence fails the test

#### Scenario: Same verdicts

- **WHEN** the same battery (valid, incomplete, invented, stale, foreign-directory, false-with-correct-fingerprint, same-HEAD-different-diff) is evaluated with the Claude flow and the Codex flow
- **THEN** both flows return the same verdict for every execution receipt

#### Scenario: Codex verify pending

- **WHEN** the pull request of proposal 04 with the Codex `verify` action is not yet merged into `main`
- **THEN** the sufficiency review in Codex stays "unverified" and the dependency is reported
