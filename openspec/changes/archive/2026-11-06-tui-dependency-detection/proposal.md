## Why

`ScreenDetection` already exists in the TUI's state machine (`internal/tui/screen.go`, wired Welcome → Detection → Agents in `router.go`) and its doc comment promises "show OS/agent detection results" — but `viewDetection()` in `internal/tui/model.go` is a stub: it renders only "Detecting environment / Press Enter to continue." with no data. The only place dependencies are actually checked today is `internal/tui/gate.go`'s `checkPreflightDeps`, which runs reactively right before install starts, only checks the narrow set of deps required by the harnesses already selected, and surfaces missing deps as a raw error string — not as an upfront, browsable report. A user only finds out Git or Node is missing after clicking through Agents/Mode/Custom-picker/Review, deep into the flow.

This mirrors what `jr-stack-legacy/internal/tui/screens/detection.go` already did correctly: a dedicated screen, shown early, listing OS/Shell/Supported, Tools, the full Dependencies report (with per-dep found/not-found/version and required/optional), and Detected Configs — all built from `internal/system.Detect()`, which is already fully ported and unused for this purpose in the current TUI.

## What Changes

- Flesh out `viewDetection()` in `internal/tui/model.go` to render the full `system.DetectionResult` (OS/arch/shell/supported, tools, the complete dependency report from `system.DetectDependencies` — git/curl/node/npm/go with status, version, required/optional — and detected agent configs), replacing the current one-line stub.
- Wire real detection data into the model: add a `DetectFn func(ctx context.Context) (system.DetectionResult, error)` (or an equivalent pre-computed field, matching the existing `AvailableAgents` pre-computed-before-model-construction pattern) to `ModelDeps`, and call it from `cmd/active-stack/main.go` alongside the existing `tui.DetectInstalledAgents(homeDir)` call.
- For each missing dependency, render its `InstallHint` and — per the earlier "facilitar sin auto-instalar" decision — the ready-made per-platform command from `system.InstallCommandsForDep` (currently dead code, only exercised by its own test) as copyable text. Never execute it.
- Leave `gate.go`'s `checkPreflightDeps` as-is: it stays the last hard gate right before install starts, now backed by a user who already saw the full picture on `ScreenDetection`.
- Go stays `Required: false` in `system.defineDependencies` — per the `readme-prereqs` decision, Go is a build-from-source concern only, not a runtime requirement, since no current catalog harness uses a `go-install` method.

## Capabilities

### New Capabilities
- `tui-environment-detection`: the TUI's `ScreenDetection` step displays the full system/tools/dependency/config detection report before the user picks an agent, including install hints and copyable install commands for anything missing.

### Modified Capabilities
(none — `ScreenDetection` exists but has no prior spec describing its content; this is the first spec to cover it)

## Impact

- `internal/tui/model.go` (`viewDetection`, `ModelDeps`, `newModel`)
- `internal/tui/run.go` / `cmd/active-stack/main.go` (wiring `system.Detect` into `ModelDeps`)
- Reuses without modification: `internal/system/detect.go`, `deps.go`, `install_deps.go` (all already implemented)
- No changes to `internal/tui/gate.go`'s reactive preflight gate
