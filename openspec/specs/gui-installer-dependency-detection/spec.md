# gui-installer-dependency-detection Specification

## Purpose
TBD - created by archiving change gui-installer-dependency-detection. Update Purpose after archive.

## Requirements

### Requirement: Dependencies page shows the full dependency report
The Dependencies page SHALL present every dependency from the session's `dependencies` data (name, required/optional, installed status, detected version when present) as a read-only list, shown between the Hub and the Assistants page in the Install flow.

#### Scenario: User enters the Install flow from the Hub
- **WHEN** the user selects Install on the Hub and advances
- **THEN** the Dependencies page is shown next, listing every dependency with its required/optional status and installed/version state, before the Assistants page is reached

### Requirement: Missing dependencies show install guidance without executing anything
For each dependency where `installed` is false, the Dependencies page SHALL display its install hint text and, when the session data includes a platform install command for it, that command as read-only, selectable text. The GUI SHALL NOT execute any install command.

#### Scenario: Git is not installed
- **WHEN** the session's dependency data reports `git` as not installed and required
- **THEN** the Dependencies page shows git as missing/required, its install hint, and the platform install command as read-only selectable text — no process is started by the GUI

#### Scenario: A dependency has no platform install command
- **WHEN** a missing dependency's session data carries no install command for the current platform
- **THEN** the Dependencies page still shows that dependency's install hint text, with no command block

### Requirement: Dependencies page never blocks advancing
The Dependencies page SHALL report it can always advance to the Assistants page, regardless of missing required dependencies. Any failure caused by a genuinely missing dependency SHALL surface later, through the existing install engine pipeline's own failure reporting, unchanged by this page.

#### Scenario: Required dependency missing, user advances
- **WHEN** a required dependency (e.g. Node.js) is missing and the user advances from the Dependencies page
- **THEN** the wizard proceeds to the Assistants page without being blocked by this page
