# Windows Native Installer Design

Date: 2026-07-02
Status: approved

## Goal

Replace the current Bubble Tea TUI entrypoint for Windows end users with a
native graphical installer delivered as a single public GitHub Release `.exe`.
The new installer must preserve the same installation logic Active Stack
already has today while presenting it through a minimal, polished, commercial
Windows experience.

## Scope

Included:

- A single-file Windows release artifact for public download:
  `ActiveStack-Setup-x.y.z.exe`.
- WiX Burn bootstrapper architecture with a custom minimal UI.
- Online installation flow that downloads external tools during installation so
  they stay current.
- Reuse of `active-stack.exe` as the single installation engine.
- New or hardened headless commands for Windows GUI orchestration.
- Structured progress and error reporting from the Go engine to the UI.
- Copywriting for each screen and selectable item in simple, commercial
  language.
- Support for the existing functional flow: agent detection, install modes,
  custom selection, dependency resolution, backup, rollback, verification, and
  uninstall entrypoints.
- GitHub Release build pipeline for the Windows installer artifact.

Excluded:

- Keeping the Bubble Tea TUI as the primary Windows end-user experience.
- Reimplementing install logic inside MSI custom actions.
- Shipping a raw MSI as the public landing-page download.
- Offline-first packaging of third-party dependencies.
- Requiring the user to clone the repository or run `build.bat`.
- A redesign of the underlying installation rules, harness catalog, or Active
  Stack methodology.

## Product Requirements

- The installer must be usable from a single downloaded `.exe`.
- The installer must not require the repository checkout.
- The installer must preserve current Active Stack behavior rather than
  simplifying or replacing it.
- The installer UI must explain each choice clearly for non-technical users.
- The installer copy must be simple and commercial, not internal-methodology
  jargon-first.
- The installer may download third-party components at install time and should
  prefer current compatible versions.

## User Experience Principles

- Lead with benefits, not internal implementation terms.
- Explain what each choice does and why the user would pick it.
- Hide technical detail by default, but keep a secondary "more details" path.
- Use calm, minimal layouts with strong hierarchy and low cognitive load.
- Keep progress transparent: the user should know what is happening, what is
  being downloaded, and whether the installer is safe to retry.

The presentation at `https://jr-stack-inicio-a-fin.vercel.app/` is the content
baseline for explanations, but the installer rewrites that material in simpler,
more commercial language.

## Distribution Model

### Public artifact

The public Windows release artifact is a single bootstrapper executable:

- `ActiveStack-Setup-x.y.z.exe`

This is the file linked from the landing page and attached to GitHub Releases.
The bootstrapper is the only artifact the end user needs.

### Online bootstrapper

The release is intentionally online:

- The bootstrapper installs the local Active Stack payload.
- The Go engine downloads external tools during the install flow.
- Dependency downloads should use current compatible versions rather than
  stale embedded copies.

This keeps the installer smaller and avoids rebuilding the full release every
time an upstream dependency updates.

## Architecture

### Recommended architecture

Use WiX Burn with a custom native Windows UI, but keep `active-stack.exe` as
the single source of truth for installation behavior.

Responsibilities split:

| Layer | Owns |
|---|---|
| WiX Burn bundle | Windows bootstrap lifecycle, install footprint, Add/Remove Programs registration, launch, uninstall entrypoint, bundle logging |
| Custom UI | Native graphical flow, validation display, simplified copy, user choices, progress rendering, access to details |
| `active-stack.exe` | Agent detection, mode resolution, harness selection rules, dependency graph, backup, rollback, verify, external tool installation, structured status events |

The UI must orchestrate the engine. It must not duplicate installer rules.

### Why this architecture

This keeps one functional engine:

- no duplicated dependency logic
- no second rollback implementation
- no drift between CLI/TUI and Windows GUI behavior
- testable separation between presentation and execution

Any design that moves business logic into Burn custom actions or a second
Windows-only installer engine is out of scope because it would create long-term
behavior drift.

## Windows Installer Composition

### Bundle contents

The Burn bundle contains or installs:

- the custom bootstrapper shell
- the custom UI assets
- `active-stack.exe`
- version and branding metadata
- local configuration needed to launch the engine

The bundle does not need to carry third-party tools that can be resolved online
during installation.

### Install footprint

The installer is expected to:

- copy the required local Active Stack payload
- register the application in Windows installed programs
- create any agreed shortcuts or launch affordances
- hand execution to `active-stack.exe` for the actual Active Stack install flow

Final install directories and shortcut policy can be specified during planning,
but are not the architectural risk in this design.

## Engine Contract for GUI Orchestration

The current engine needs an explicit Windows GUI-facing headless contract.
The UI should consume structured output only.

### Commands

#### `active-stack windows detect --json`

Returns:

- detected supported agents
- machine readiness
- missing prerequisites
- writable target locations
- warnings the UI should surface before proceeding

#### `active-stack windows options --agent <id> --json`

Returns:

- available install modes
- simple descriptions for each mode
- included components per mode
- rules and forced selections
- custom-selectable components and dependencies

#### `active-stack windows install --agent <id> --mode <mode> [component flags] --json-stream`

Executes the install flow and emits structured events for:

- phase changes
- dependency resolution
- downloads
- backup
- installation steps
- config injection
- verification
- rollback
- completion

#### `active-stack windows uninstall --json-stream`

Exposes the uninstall flow with the same event model so the Windows shell can
surface uninstall status consistently.

### Event model

The JSON stream should emit stable event types:

- `phase_started`
- `step_started`
- `step_output`
- `step_succeeded`
- `step_failed`
- `download_started`
- `download_progress`
- `download_finished`
- `rollback_started`
- `rollback_finished`
- `install_finished`

Recommended event fields:

- `type`
- `phase`
- `step_id`
- `title`
- `message`
- `percent`
- `severity`
- `details`
- `retryable`
- `timestamp`

The custom UI must not parse human-readable logs to infer state.

## Screen Flow

The installer flow is a Windows-native version of the current TUI behavior,
not a different product flow.

### 1. Welcome

Purpose:

- explain what Active Stack does
- set expectation that the installer may download components
- make it clear the installer prepares the user's AI coding environment

Primary copy style:

- "Set up your AI coding workspace in a few steps."
- "Active Stack installs the tools and configuration needed to get started."

### 2. Environment Check

Purpose:

- show what the installer found
- show what is missing
- reassure the user that missing items will be downloaded automatically when
  possible

Examples of information:

- supported agents detected
- internet requirement
- external tools that will be installed
- warnings requiring user attention

### 3. Choose Your Assistant

Purpose:

- let the user select the AI agent target when multiple are available
- explain the practical impact of the choice

Copy style:

- simple agent names
- one-line descriptions
- recommendation marker where appropriate

### 4. Choose Installation Type

Modes should be presented with end-user labels first, with internal mapping
behind the scenes:

| UI label | Internal mode | Positioning |
|---|---|---|
| Quick | `lite` | Fast setup to start working right away |
| Complete | `full` | Full recommended setup with all key tools |
| Custom | `custom` | Choose exactly what to install |

The copy should explain outcomes, not implementation jargon.

### 5. Customize Components

Purpose:

- available only for Custom mode
- let the user opt into individual components
- explain each item in simple language

Requirements:

- every item has a plain-language description
- recommended items are visually marked
- forced protections are clearly explained

Example:

- "Basic protection" instead of foregrounding `permissions`
- supporting text explains that it helps avoid unsafe changes and is always on

### 6. Review

Purpose:

- confirm what will be installed
- show which downloads will happen
- summarize where the setup will apply

This is the user's final confirmation screen before writes begin.

### 7. Installing

Purpose:

- show progress clearly by phase
- surface download progress
- reassure the user during longer-running steps

Requirements:

- simple progress state by default
- expandable technical details/log view
- explicit rollback messaging if recovery is triggered

### 8. Complete / Error

Purpose:

- tell the user whether setup finished successfully
- explain next steps
- surface recovery guidance if something failed

Success view should offer:

- launch or open next-step guidance
- shortcut/open actions if supported
- very short explanation of what to do next

Failure view should offer:

- plain-language explanation
- details/logs
- retry guidance where safe

## Copy Strategy

The UI uses a dual-layer explanation model:

- primary layer: simple, commercial, user-facing copy
- secondary layer: technical detail available on demand

Examples:

| Internal concept | Primary copy |
|---|---|
| Lite | Quick setup |
| Full | Complete setup |
| Custom | Custom setup |
| permissions | Basic protection |
| idempotent marker injection | Reinstalling will not duplicate your settings |
| rollback by stage | If something fails, the installer reverts the changes it made |

Internal terms like `harness`, `OPSX`, and `orchestrator` should be hidden from
top-level screens unless the user opens more information.

## Functional Parity Requirements

The native Windows installer must preserve the same logic already available in
Active Stack today:

- supported agent detection
- mode selection
- custom component selection
- dependency resolution
- backup before mutation
- idempotent config injection
- staged rollback
- verification after install
- uninstall entrypoints consistent with the Windows shell experience

The Windows GUI is a new presentation layer, not a new installer core.

## Networking and Dependency Strategy

### Online behavior

The installer is allowed and expected to download external components during
installation.

The engine should:

- resolve current compatible dependency versions
- validate downloads before use
- report download progress
- report actionable failures when the network or upstream endpoints fail

### Failure handling

Expected network failure scenarios:

- no internet connection
- timeout
- upstream package host unavailable
- incompatible upstream change

Mitigations:

- bounded retry logic
- clear retry-safe messaging
- detailed logs
- compatibility pinning where required by Active Stack policy
- rollback when a partially applied install fails

## Release and CI/CD Strategy

The repository should gain a Windows installer release pipeline that:

1. builds `active-stack.exe`
2. runs the relevant test suite
3. builds the WiX Burn bootstrapper
4. versions the installer artifact from tag or release metadata
5. publishes `ActiveStack-Setup-x.y.z.exe` to GitHub Releases

The Windows release artifact becomes the supported public distribution path for
end users.

## Repository Changes Required

### Go engine

- add or harden the Windows GUI command surface
- add structured JSON and JSON-stream outputs
- separate presentation text from internal engine execution state where needed
- ensure existing TUI-only assumptions are not required for installation logic

### Windows installer project

- add WiX Burn bundle sources
- add custom UI implementation and assets
- add branding, icons, license, and version resources
- add release packaging configuration

### Tests

- contract tests for `windows detect`
- contract tests for `windows options`
- streaming event tests for `windows install`
- failure and rollback tests for GUI-driven execution
- packaging smoke test in CI for the bundle artifact

## Risks

### Primary risk

The main technical risk is not WiX itself. It is exposing the existing install
engine as a stable GUI-ready API without leaking TUI assumptions or duplicating
rules in the Windows layer.

### Secondary risks

- drift between UI copy and actual engine behavior
- incomplete event model forcing log parsing
- upstream download instability
- ambiguous ownership between Burn and the Go engine

## Implementation Phases

### Phase 1 - GUI-ready engine contract

- define and implement the Windows headless contract
- emit structured JSON outputs
- cover the contract with tests

### Phase 2 - Minimal WiX Burn shell

- create bundle skeleton
- install local payload
- invoke the engine
- validate launch and lifecycle wiring

### Phase 3 - Custom UI and screen flow

- implement the approved screens
- connect to structured engine events
- add simple/commercial copy

### Phase 4 - Online dependency handling hardening

- improve progress reporting
- improve retry/recovery messaging
- test failure paths

### Phase 5 - Release automation

- wire GitHub Release artifact generation
- attach signed/versioned bootstrapper if signing is adopted later

## Testing Strategy

Use test-first development for engine-facing work.

Required coverage:

- command contract tests for each new Windows headless command
- event schema stability tests
- rollback behavior under download and apply failure
- mode/component parity tests against existing install logic
- UI integration tests around selection flow and progress rendering where
  practical
- CI artifact build verification for the bootstrapper

No part of the Windows UI should require duplicating install logic just to make
tests pass.

## Governance

- Engine contract work is MEDIUM governance: it touches the installer's public
  behavior and should be implemented in checkpoints.
- Windows packaging and UI shell work is LOW to MEDIUM governance.
- Any change that weakens backup, rollback, verification, or config safety is
  HIGH governance and requires focused review.

## Approval Notes

Approved design decisions captured during brainstorming:

- native Windows graphical installer replaces the current TUI for Windows users
- same functional logic as today must be preserved
- WiX Burn with custom minimal UI
- one public `.exe` release artifact
- online dependency installation
- simple, commercial copy
- the existing explanatory website is the reference content source for item
  descriptions and guidance
