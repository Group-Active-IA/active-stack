## Why

The README never tells a new user what has to be installed before `active-stack` is useful. Nothing points at Git, Node.js, or Go, and nothing distinguishes "needed to compile the binary from source" from "needed at runtime for the installer to materialize harnesses." A user who clones the repo and runs `build.bat` with no Go on PATH gets a bare compiler error instead of a pointer to the right download page; a user who runs the installer with no Git/Node gets a late, unexplained failure instead of an upfront heads-up.

This is the first of three related changes (the other two — `tui-dependency-detection` and `gui-installer-dependency-detection` — add a proactive dependency-detection screen to the TUI and to the graphical Windows installer, respectively, reusing the detection logic already in `internal/system/deps.go`). This one is docs-only and unblocks the other two by establishing what gets said and where.

## What Changes

- Add a "Requisitos previos" section to `README.md`, placed before "Quick start", listing:
  - **Go 1.26+** (https://go.dev/dl/) — required only to compile from source via `build.bat`/`build.sh`.
  - **Git** (https://git-scm.com/install/windows) — required at runtime: the installer clones `skill`-type harnesses.
  - **Node.js LTS** (https://nodejs.org/es/download) — required at runtime: the installer installs `external` harnesses via `npm` (e.g. OpenSpec CLI).
  - A supported AI agent already installed (Claude Code, Codex, or OpenCode) — without one there is nothing for the installer to configure.
- Explicitly split the runtime list (Git, Node) from the build-from-source-only item (Go), so the section doesn't overstate what the released binaries need.

No code changes. No new or modified runtime behavior.

## Capabilities

### New Capabilities
- `installer-prerequisites-docs`: README documents the dependencies needed to build active-stack from source and to run the installer, with official download links and the build-time/runtime distinction.

### Modified Capabilities
(none — no existing spec covers README content)

## Impact

- `README.md` only. No Go code, no installer artifacts, no tests beyond the doc content itself.
