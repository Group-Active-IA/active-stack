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

The system SHALL apply the agreed failure policy when a control does not respond in time, returns invalid data, or its handler does not exist, and the pilot MUST record the actual behavior of each client in those cases.

#### Scenario: Timeout

- **WHEN** a control exceeds its time limit while an action is evaluated
- **THEN** the failure policy for that kind of action is applied and reported

#### Scenario: Invalid JSON

- **WHEN** a control returns a response that is not valid JSON
- **THEN** the failure policy is applied and reported

#### Scenario: Missing handler

- **WHEN** a configured hook points to a command that does not exist on the system
- **THEN** the failure policy is applied and reported

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

### Requirement: Idempotency, update and uninstall

The controls SHALL be idempotent on reinstall, take part in update, and be fully removed by uninstall.

#### Scenario: Reinstall

- **WHEN** controls are already installed and the installer runs again
- **THEN** entries are not duplicated and the file is unchanged if the content is identical

#### Scenario: Uninstall

- **WHEN** controls installed by the stack are uninstalled
- **THEN** the stack's own entries are removed and no orphan configuration remains
