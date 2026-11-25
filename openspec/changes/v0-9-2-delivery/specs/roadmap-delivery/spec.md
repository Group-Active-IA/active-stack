## ADDED Requirements

### Requirement: Explicit tracking choice during foundation

Project foundation SHALL obtain an explicit choice about ClickUp and preserve that choice when resumed. ClickUp setup SHALL validate the destination list and its development statuses before marking the phase complete.

#### Scenario: The user chooses local tracking

- **WHEN** the user answers that the project will use local tracking
- **THEN** the foundation records the decision and subsequent delivery proceeds without ClickUp writes

#### Scenario: The user chooses ClickUp

- **WHEN** the user chooses ClickUp
- **THEN** setup verifies the list and required statuses, and keeps the choice pending configuration if the list or statuses cannot be verified

### Requirement: Verified delivery milestones drive ClickUp status

Delivery SHALL propose the configured destination status for the linked ClickUp task based on confirmed branch, PR and archive evidence. It SHALL preserve the existing preview and confirmation rules for remote writes.

#### Scenario: The reviewer requests changes

- **WHEN** GitHub confirms that the designated reviewer has requested changes on the open PR
- **THEN** delivery previews the corrections status for the exact linked task

#### Scenario: The remote task is already in its target status

- **WHEN** the task already has the target status for the confirmed milestone
- **THEN** delivery performs no remote write

### Requirement: Sequential peer review

The team roadmap SHALL determine the reviewer from the assignee of the next eligible change in roadmap order. When that person is unavailable, is the current author or does not exist, the roadmap SHALL identify an approved independent fallback.

#### Scenario: The next peer reviews an open PR

- **WHEN** an author opens a PR for review
- **THEN** the next eligible reviewer checks the current head and either approves and merges after required checks or submits requested changes and returns the PR to the author
