## ADDED Requirements

### Requirement: ScreenDetection shows the full system report
`ScreenDetection` SHALL render OS, architecture, shell, platform-supported status, detected tools, the full dependency report (name, required/optional, installed status, detected version), and detected agent config presence — sourced from `system.DetectionResult` — before the user reaches `ScreenAgents`.

#### Scenario: User enters the install flow
- **WHEN** the user selects "Install" on the welcome hub and lands on `ScreenDetection`
- **THEN** the screen displays OS/arch/shell/supported, the tools section, the dependencies section (each dependency with its found/not-found status and version when found), and the detected-configs section, without requiring any prior selection

### Requirement: Missing dependencies show install guidance without executing anything
For each dependency where `Installed` is false, `ScreenDetection` SHALL display its `InstallHint` text and, when `system.InstallCommandsForDep` returns a non-nil command sequence for the current platform, the copyable command text. The TUI SHALL NOT execute any install command automatically.

#### Scenario: Git is not installed
- **WHEN** `system.DetectDependencies` reports `git` as not installed and required
- **THEN** `ScreenDetection` shows git as missing/required, its `InstallHint`, and the platform install command as copy-only text — no command is executed by the TUI

#### Scenario: A dependency has no automatic install command for the platform
- **WHEN** `system.InstallCommandsForDep` returns nil for a missing dependency on the current platform
- **THEN** `ScreenDetection` still shows that dependency's `InstallHint` text, with no command block

### Requirement: ScreenDetection never blocks navigation on missing dependencies
`ScreenDetection` SHALL allow the user to proceed to `ScreenAgents` via Enter regardless of missing dependencies; the existing reactive preflight gate (`internal/tui/gate.go`) remains the sole blocking check, evaluated later against the harness-scoped dependency set.

#### Scenario: Required dependency missing, user presses Enter on ScreenDetection
- **WHEN** a required dependency (e.g. Node.js) is missing and the user presses Enter on `ScreenDetection`
- **THEN** the TUI advances to `ScreenAgents` without blocking; any install-time block still happens later via the existing preflight gate before install starts
