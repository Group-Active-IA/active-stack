## 1. Wire detection data into the model

- [x] 1.1 Add `Detection system.DetectionResult` field to `ModelDeps` in `internal/tui/model.go`
- [x] 1.2 In `cmd/active-stack/main.go`, call `system.Detect(ctx)` alongside the existing `tui.DetectInstalledAgents(homeDir)` call (line ~218) and pass the result into `tui.ModelDeps{Detection: ...}`
- [x] 1.3 Handle the `system.Detect` error path in `main.go` (log/degrade gracefully — an empty `DetectionResult` should not crash the TUI; `viewDetection` must render sanely with zero values)

## 2. Render the detection screen

- [x] 2.1 Write a test (RED) asserting `viewDetection()` output contains OS/arch/shell/supported status when `m.deps.Detection` is populated
- [x] 2.2 Implement `viewDetection()` in `internal/tui/model.go`: OS/Shell/Supported section, using existing `titleStyle`/`selectedStyle`/`errorStyle`/`dimStyle` (GREEN)
- [x] 2.3 Write a test (RED) asserting the Tools section renders found/not-found per `m.deps.Detection.Tools`
- [x] 2.4 Implement the Tools section (GREEN)
- [x] 2.5 Write a test (RED) asserting the Dependencies section renders each dependency's name, required/optional label, and status (version when installed, "not found"/"NOT FOUND (required)" when missing)
- [x] 2.6 Implement the Dependencies section using `m.deps.Detection.Dependencies` (GREEN)
- [x] 2.7 Write a test (RED) — second case with different dependency data (TRIANGULATE): a run where all deps are present vs. a run with a mix of missing required/optional
- [x] 2.8 Write a test (RED) asserting the Configs section renders per `m.deps.Detection.Configs`
- [x] 2.9 Implement the Configs section (GREEN)

## 3. Missing-dependency install guidance

- [x] 3.1 Write a test (RED) asserting a missing dependency's `InstallHint` text appears in the rendered view
- [x] 3.2 Implement rendering of `InstallHint` per missing dependency (GREEN)
- [x] 3.3 Write a test (RED) asserting `system.InstallCommandsForDep(dep.Name, profile)` output appears as copy-only text (joined argv, one command per line) when non-nil
- [x] 3.4 Implement the command-rendering branch (GREEN)
- [x] 3.5 Write a test (RED) asserting that when `InstallCommandsForDep` returns nil, only `InstallHint` renders (no empty command block)
- [x] 3.6 Implement the nil-fallback branch (GREEN) — already held by construction (ranging over a nil `[][]string` is a no-op), test passed on first run with no code change needed
- [x] 3.7 Confirm (by reading the implementation, not just tests) that no code path in `viewDetection()` or its key handler calls `exec.Command`/`os/exec` — this screen only renders text (grep confirms zero matches in `detection_screen.go`; the `ScreenDetection` key handler in `model.go` only sets `m.Screen`)

## 4. Navigation stays non-blocking

- [x] 4.1 Write a test (RED) asserting Enter on `ScreenDetection` advances to `ScreenAgents` even when `m.deps.Detection.Dependencies.MissingRequired` is non-empty
- [x] 4.2 Confirm existing key handler in `internal/tui/model.go` (~line 207) already satisfies this without changes; adjust only if the test fails — held with no changes, confirmed by reading (handler only sets `m.Screen`, never inspects `m.deps.Detection`)
- [x] 4.3 Run the full `internal/tui` test suite — confirm no regression in `router.go` navigation tests or `hub_test.go`

## 5. Refactor pass

- [x] 5.1 Extract repeated section-rendering helpers if `viewDetection()` grows past ~60 lines (mirror the legacy screen's structure, not its styles package) — not triggered: `viewDetection()` stayed at 13 lines because each section was already factored into its own `write*Section` helper as it was built
- [x] 5.2 Re-run full suite after refactor — confirm still green
