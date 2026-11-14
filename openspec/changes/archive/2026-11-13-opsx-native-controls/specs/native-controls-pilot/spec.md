# Spec Delta

## Purpose

Define how native client controls (permissions and hooks) are piloted and documented, how they behave when they fail, and how user configuration is written with review.

## ADDED Requirements

### Requirement: Classification of controls

Each rule in the pilot SHALL be classified as exactly one of: guides the agent (text), restricts an action (client permissions or rules), or verifies an event (client hook). The documentation MUST distinguish client hooks from the installer's own post-install verification hook.

#### Scenario: Rule classified once

- **WHEN** a rule is added to the pilot matrix
- **THEN** it is recorded with exactly one class and the client, surface, tool and shell it was evaluated on

#### Scenario: Installer hook not counted as a control

- **WHEN** the coverage table lists the controls of a client
- **THEN** the installer's post-install verification hook does not appear as a client control

### Requirement: Pre-action blocking of disallowed actions

The system SHALL block, through a native pre-action control, the actions classified as "must be blocked".

#### Scenario: Allowed action

- **WHEN** the agent attempts an action that is not listed as blocked
- **THEN** it runs without intervention from the control

#### Scenario: Denied action

- **WHEN** the agent attempts an action classified as "must be blocked"
- **THEN** the pre-action control blocks it and the agent receives the reason

### Requirement: Post-action validation of results

The system SHALL be able to validate the result of an action through a native post-action control.

#### Scenario: Conforming result

- **WHEN** an action's result complies with the contract defined by proposal 06 and the post-action control runs
- **THEN** the result is accepted

#### Scenario: Non-conforming result

- **WHEN** an action's result violates the contract and the post-action control runs
- **THEN** the result is marked invalid and the reason is reported

### Requirement: Behavior when a control fails

The system SHALL apply the agreed failure policy (TBD, group decision) when a control times out, returns invalid data, or has no handler, and the pilot MUST record each client's actual behavior. What each client documents (design D6) is a hypothesis to measure, not a result; any behavior the documentation does not state MUST be marked (TBD) until measured.

#### Scenario: Timeout

- **WHEN** a control exceeds its time limit while an action is evaluated
- **THEN** the failure policy for that kind of action is applied and reported

#### Scenario: Invalid JSON

- **WHEN** a control returns a response that is not valid JSON
- **THEN** the failure policy is applied and reported

#### Scenario: Missing handler

- **WHEN** a configured hook points to a command that does not exist on the system
- **THEN** the failure policy is applied and reported

### Requirement: Missing mandatory or optional control

When a control that an action requires is absent, the system SHALL stop only that action, name the missing protection, and allow resuming once an equivalent protection is verified. An absent optional control SHALL produce a warning. Which controls are mandatory is (TBD, group). This does not redefine the missing-module rules of proposals 03 and 04; it defines the message and what counts as effective protection.

#### Scenario: Mandatory control present

- **WHEN** an action requires a mandatory control and the control is installed and verified by the coverage method
- **THEN** the action proceeds under the control

#### Scenario: Mandatory control absent

- **WHEN** an action requires a mandatory control that is absent
- **THEN** only that action is stopped, the message names the missing protection and how to restore it, and the actions that do not depend on it continue

#### Scenario: Optional control absent

- **WHEN** an optional control is absent
- **THEN** a warning is shown and the action continues

#### Scenario: Equivalent protection verified

- **WHEN** a stopped action is retried and an equivalent protection has been verified by the same method as the coverage matrix
- **THEN** the action resumes, and the mere presence of a configuration key is not accepted as verification

### Requirement: Coverage against alternative paths and real shells

The pilot SHALL test alternative execution paths in real shells and record which remain uncovered.

#### Scenario: Alternative path

- **WHEN** a denied action is attempted through an equivalent path in each supported shell
- **THEN** the control blocks it or the path is documented as a limit of that combination

### Requirement: Documented limits per surface

Each client and each declared surface SHALL have the limits of the native control documented, and an action that could not be tested MUST be listed as uncontrolled on that surface.

#### Scenario: Untested coverage

- **WHEN** an action could not be verified on a surface
- **THEN** it appears as "uncontrolled" for that surface and no protection is claimed

### Requirement: Configuration writes with review, backup and reversal

The system SHALL show the exact permission and hook changes and obtain review before writing them, create a backup first, and be able to revert. The write MUST NOT proceed when the backup cannot be created.

#### Scenario: Reviewed change

- **WHEN** a configuration change is prepared
- **THEN** the diff is shown, the backup is created, and the file is written only after approval

#### Scenario: Backup failure

- **WHEN** the backup cannot be created
- **THEN** the configuration file is not modified

#### Scenario: Reversal

- **WHEN** a revert is requested for a configuration written by this feature
- **THEN** the snapshot is restored without altering entries that are not owned by the stack

#### Scenario: Edits made after installation

- **WHEN** the user edited a file written by this feature after the installation and a revert is requested
- **THEN** the current file is preserved, the conflict is reported with the backup location, and the revert uses the existing rollback guard of proposal 01 (`Manifest.GuardRollback`) instead of a new mechanism

### Requirement: Unreadable configuration is never overwritten

The system MUST NOT write over a configuration file it cannot read or interpret. An existing file that cannot be parsed SHALL abort the write, remain byte-identical, and be reported with its path and the reason. It MUST NOT be treated as an empty object.

#### Scenario: Malformed settings.json

- **WHEN** the user's `settings.json` exists but is not valid JSON (even after removing comments and trailing commas) and the stack installs or updates its permissions or hooks
- **THEN** the write is aborted, the file is byte-identical to before, and the report names the path and the parse error

#### Scenario: Invalid config.toml

- **WHEN** the user's Codex `config.toml` exists but is not valid TOML
- **THEN** the write is aborted, the file is byte-identical to before, and the diagnostic names the backup and the repair action

### Requirement: Preservation of user entries

Writing permissions or hooks SHALL preserve the user's own entries: scalar values the user set, entries in the allow and deny lists, and hooks of the user. The system MUST NOT replace a user list or scalar without showing it in the diff first.

#### Scenario: User allow entry

- **WHEN** a user's settings file has an allow entry that the stack does not own and the stack installs or updates its permissions
- **THEN** the entry is still present after the write

#### Scenario: User scalar setting

- **WHEN** a user's settings file sets `defaultMode` to a value different from the one the stack would write
- **THEN** the diff shows the change before it is applied and the write needs approval

#### Scenario: User hooks

- **WHEN** a settings file contains hooks of the user and the stack installs, updates or uninstalls its controls
- **THEN** the user's hooks remain unchanged

### Requirement: Codex cells use the policy and method of proposal 01

The Codex cells of the pilot matrix SHALL use the Strict policy that proposal 01 delivered to `main` (`sandbox_mode` set to read-only with `approval_policy` set to on-request) and its acceptance method, and the pilot MUST NOT define a separate Strict policy for Codex.

#### Scenario: Effective permissions in an isolated profile

- **WHEN** a Codex permission cell is tested
- **THEN** it runs with the real client in an isolated profile, attempts an allowed read and a write whose escalation is denied, and records the surface, version, result and final state of the test file

#### Scenario: First verifiable case

- **WHEN** the Codex matrix is first filled
- **THEN** Windows with Codex CLI 0.166.0 is recorded first, and the remaining environments are listed as not tested until proposal 01 agrees them

#### Scenario: Policy in force on main

- **WHEN** a Codex Strict cell is configured
- **THEN** it uses the policy written by `internal/harness/config/permissions/install.go` on `main`, and a later change to that policy changes only those cells

### Requirement: No collision with the explicit coordinator skill

The coordinator of proposal 03 is activated by an explicit skill (`/sdd-orchestrator` in Claude, `$sdd-orchestrator` in Codex), not by `--agent` or the `agent` setting. Writing permissions or hooks to the Claude settings file SHALL leave the `agent` key and any user hooks unchanged, and the pilot MUST NOT write the `agent` key.

#### Scenario: Coordinator skill installed

- **WHEN** the coordinator skill of proposal 03 is installed in the target and the stack writes its permissions overlay
- **THEN** the `agent` key that the user set, if any, is identical after the write, and the stack's own keys are limited to permissions (and hooks, if the pilot installs them)

#### Scenario: Parent directory and file mode

- **WHEN** the stack writes a settings file whose parent directory or file has a restrictive mode
- **THEN** the pilot records whether `WriteFileAtomic` widened the directory to `0755` or replaced the file mode with `0644`, on Windows and on Linux, and does not claim the mode is preserved unless measured

### Requirement: Idempotency, update and uninstall

The controls SHALL be idempotent on reinstall, take part in update, and be fully removed by uninstall.

#### Scenario: Reinstall

- **WHEN** controls are already installed and the installer runs again
- **THEN** entries are not duplicated and the file is unchanged if the content is identical

#### Scenario: Uninstall

- **WHEN** controls installed by the stack are uninstalled
- **THEN** the stack's own entries are removed, including the permissions overlay entries the stack wrote, and no orphan configuration remains

### Requirement: Inventory of writes and coverage table

The pilot SHALL list every file the change would write, with the `filemerge` function used and the governance level, and SHALL publish the coverage table in a closed, machine-readable format (one row per matrix cell with the states covered, not covered or not tested) so that proposal 10 can consume it. The inventory SHALL include the CRLF-to-LF conversion of `UpsertTOMLTopLevel` and the parent-directory chmod of `WriteFileAtomic` as risks measured on Windows and on Linux.

#### Scenario: Write inventory

- **WHEN** the design is reviewed
- **THEN** each file to be written is listed with its function and level, and files at ALTO level are marked

#### Scenario: Coverage row without evidence

- **WHEN** a matrix cell has no recorded run
- **THEN** its state is "not tested" and never "covered"

#### Scenario: Line endings of config.toml

- **WHEN** a Codex `config.toml` with CRLF line endings is written through `UpsertTOMLTopLevel`
- **THEN** the pilot records whether the whole file was converted to LF, on Windows and on Linux

### Requirement: Entries installed have owner, reversal and precedence

For each entry the change installs, the design SHALL state the owner, how it is reverted, its precedence against the user's configuration, its behavior under Bypass, its behavior in an unattended update, and how it is restored without overwriting later user edits. The pilot (measured in a temporary home) is distinct from permanent installation, which needs prior human review.

#### Scenario: Unattended update

- **WHEN** `update` replays a registered selection and no diff can be approved
- **THEN** the behavior is the one agreed for unattended updates (TBD) and no entry is written without the approval the design requires

### Requirement: Human review of filemerge changes

The pull request of this change touches `filemerge` (ALTO). Any change to `filemerge` SHALL require review by the designated human reviewer (Neyén, per the roadmap assignment), and the fixes of proposals 03 (`WriteFileAtomic`) and 08 (`MergeJSONObjects`) SHALL be integrated serialized with that single reviewer.

#### Scenario: Change to filemerge

- **WHEN** a change modifies a file under `internal/filemerge`
- **THEN** it is not integrated without the designated human reviewer's approval

### Requirement: Platform matrix

The pilot SHALL treat Windows and Linux as the minimum platforms and publish a matrix that distinguishes distributed (`.github/workflows/release.yml:71-76`), declared (`CLAUDE.md:13`) and tested platforms, where "tested" means what was actually run.

#### Scenario: Platform matrix

- **WHEN** the matrix is published
- **THEN** each platform is listed as distributed, declared and tested separately, and a platform not run is not listed as tested
