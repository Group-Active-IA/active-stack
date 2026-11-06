## ADDED Requirements

### Requirement: README documents build-time prerequisites
The README SHALL list Go 1.26+ with its official download link (https://go.dev/dl/) as required only to compile `active-stack` from source via `build.bat`/`build.sh`.

#### Scenario: Reader wants to build from source
- **WHEN** a reader intends to run `build.bat` or `build.sh` without a prebuilt binary
- **THEN** the README's prerequisites section names Go 1.26+ as required and links to https://go.dev/dl/

### Requirement: README documents runtime prerequisites
The README SHALL list Git and Node.js as required for `active-stack install` to materialize harnesses at runtime, each with its official download link (Git: https://git-scm.com/install/windows, Node.js: https://nodejs.org/es/download), and SHALL state that these apply regardless of whether the binary was built from source or downloaded from Releases.

#### Scenario: Reader downloads a prebuilt release binary
- **WHEN** a reader downloads `active-stack_windows_amd64.exe` or `ActiveStack-Setup.exe` from Releases without building from source
- **THEN** the README's prerequisites section still names Git and Node.js as required, with their official links, and does not list Go as required for this path

### Requirement: README documents the agent prerequisite
The README SHALL state that a supported AI agent (Claude Code, Codex, or OpenCode) must already be installed for the installer to have something to configure.

#### Scenario: Reader has no AI agent installed
- **WHEN** a reader reads the prerequisites section before installing any AI agent
- **THEN** the section tells them a supported agent must be installed first, naming at least Claude Code, Codex, and OpenCode
