## Context

Current flow, confirmed by reading the code:

- `internal/tui/screen.go` declares `ScreenDetection` with comment "show OS/agent detection results" — aspirational, not yet true.
- `internal/tui/router.go`: `ScreenWelcome: {forward: ScreenDetection}`, `ScreenDetection: {forward: ScreenAgents, backward: ScreenWelcome}` — navigation is already wired.
- `internal/tui/model.go:568` — `viewDetection()` is a two-line stub with no data.
- `internal/tui/model.go:207` — the `ScreenDetection` key handler only handles Enter (→ ScreenAgents) and Esc (→ prevScreen); no data-dependent branching.
- `internal/tui/model.go` `ModelDeps` struct (line ~20) is the injection seam: fields like `AvailableAgents` are pre-computed by the caller (`cmd/active-stack/main.go:218-223`) before `newModel`/`NewProgram` runs. `system.Detect(ctx)` is not called anywhere in this path today.
- `internal/tui/gate.go` (`checkPreflightDeps`) already shows the pattern for turning a `system.DependencyReport` into a Go error string via `system.RenderDependencyReport` — reused as reference, not modified.
- `internal/system.Detect(ctx)` (in `detect.go`) already returns the full `DetectionResult{System, Tools, Configs, Dependencies}` — this is the exact shape `jr-stack-legacy/internal/tui/screens/detection.go` rendered. Nothing new needed on the `internal/system` side.
- `internal/system.InstallCommandsForDep(name, profile)` (in `install_deps.go`) returns `[][]string` — platform command argv sequences (e.g. `["winget","install","--id","Git.Git",...]`) — currently unused in production code, only exercised by `install_deps_test.go`.

## Goals / Non-Goals

**Goals:**
- Make `ScreenDetection` actually show what its doc comment claims: OS/Shell/Supported, Tools, full Dependencies report, detected agent configs.
- Surface, for each missing dependency, the official `InstallHint` text and a copyable command built from `InstallCommandsForDep` — text only, never executed.
- Keep the existing reactive gate in `gate.go` untouched; this screen is informational/proactive, the gate stays the hard stop.

**Non-Goals:**
- No auto-install, no shelling out to winget/apt/brew from the TUI.
- No change to `internal/system` detection logic itself (it's already correct and complete).
- No change to Go's `Required: false` status in `defineDependencies` (locked in during the planning conversation: Go is a build-from-source concern, not a runtime one, since no catalog harness uses `go-install`).

## Decisions

- **Where detection runs**: at `ModelDeps` construction time in `cmd/active-stack/main.go`, alongside the existing `tui.DetectInstalledAgents(homeDir)` call (line 218) — same lifecycle as `AvailableAgents`, so `ScreenDetection` has data the instant it's shown, no async/spinner state needed on first paint. Alternative considered: lazy-detect on screen entry via a `tea.Cmd` — rejected because it adds an async state (`ScreenDetectionLoading`) for a check that completes in well under a second (concurrent goroutines already in `detectDeps`), and duplicates the "pre-computed before model" pattern `AvailableAgents` already establishes.
- **Data shape in `ModelDeps`**: add `Detection system.DetectionResult` (a plain pre-computed value, not a func), matching `AvailableAgents []model.Agent`'s pattern exactly — consistent with the rest of the struct, and trivially fakeable in tests (just set the field).
- **Rendering**: a new `viewDetection()` body following `jr-stack-legacy/internal/tui/screens/detection.go`'s section order (OS/Shell/Supported → Tools → Dependencies → Configs) but using this repo's existing `titleStyle`/`selectedStyle`/`errorStyle`/`dimStyle` Lipgloss vars (`model.go` lines ~505-509) instead of the legacy `styles` package, since active-stack deliberately has no cosmetic theme package (CLAUDE.md §1: theme is out of scope).
- **Missing-dep copyable command**: render `InstallCommandsForDep(dep.Name, profile)` joined with spaces per command, one per line, under a "Run:" label — profile comes from `Detection.System.Profile` (already part of `DetectionResult`). If `InstallCommandsForDep` returns nil (no automatic command for this platform), fall back to just the `InstallHint` text, which already has a human-readable fallback (e.g. "install git from https://git-scm.com/").
- **Blocking vs informational**: `ScreenDetection`'s Enter key keeps advancing to `ScreenAgents` unconditionally, even with missing required deps — the existing `gate.go` preflight still blocks install later. Rationale: `ScreenDetection` runs before agent/mode selection, so `RequiredDependencies` (which depends on the *selected* harnesses) can't be computed yet; showing the full unscoped report here and gating later on the scoped report is the correct split, not a bug to fix.

## Risks / Trade-offs

- [Risk] `system.Detect(ctx)` in `main.go` adds a synchronous network-free but process-spawning call (git --version, node --version, etc.) before the TUI even starts rendering → Mitigation: `detectDeps` already runs all checks concurrently (`sync.WaitGroup` in `deps.go`); this was already true for `checkPreflightDeps`'s narrower call, just moved earlier and widened to the full set.
- [Risk] Duplicating dependency-rendering logic between this TUI screen and the future `gui-installer-dependency-detection` change (WPF side) → Mitigation: the WPF side consumes a JSON contract from the Go binary, not Go render code directly, so there's no actual duplication to dedupe — each side renders its own platform-native UI from the same `system.DetectDependencies` data.
