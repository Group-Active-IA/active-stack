## Why

A manual smoke test of the just-shipped Windows GUI installer (ES/EN i18n + component detail panel, L1–L5) surfaced four real defects, all confirmed by direct code reading and all confined to the C# WPF layer under `installer/windows/ui/**` (the Go engine was checked in the two suspected spots — stdout encoding and `--lang` passing — and is clean). The most severe (P0) makes the installer window close silently on any install failure or malformed engine output, leaving the user with zero signal about success or failure. This change fixes the four bugs and closes the test-coverage gaps that let them ship.

## What Changes

- **Bug A+B (P0) — surface engine/stream failures instead of dying silently.** Wrap the shell's stream-consumption in `ShellViewModel.RunStreamAndAdvanceToCompleteAsync` so that any exception thrown while consuming the install/uninstall/starter stream navigates to the existing `CompletePageViewModel` with `CompleteState.Error` (a terminal state today only reachable via `install_finished`, never via an exception). Harden `ProcessInstallerEngineClient.CreateProcess` to set UTF-8 `StandardOutputEncoding`/`StandardErrorEncoding`, and harden `InstallProgressSnapshotParser.Parse` so a non-JSON stdout line is treated as an ignorable/log snapshot rather than throwing an uncaught `JsonException`. Add a global `DispatcherUnhandledException` handler in `App.xaml.cs` as a secondary safety net only — the try/catch is the primary fix.
- **Bug C (visual) — list pages wrap instead of scrolling horizontally.** Set `ScrollViewer.HorizontalScrollBarVisibility="Disabled"` on every `ListBox` in `Themes/PageTemplates.xaml` (a transversal defect the user noticed on the Permissions page but that affects all list pages).
- **Bug D (data/wiring, diagnostic) — Custom component selection must reach the Review summary.** Add the missing end-to-end Components→Review integration test that static reading could not stand in for. The test is written first as a DIAGNOSTIC: if it fails, the same RED→GREEN cycle fixes the C# wiring gap; if it passes, the bug lives in real Go-engine data and is recorded as an explicit out-of-scope follow-up (this change does NOT touch Go code).
- **Test-seam hardening.** New tests must exercise the real process-invocation / stream-consumption paths (not the `JsonCommandRunnerOverride` seam that let the original bugs slip through).

## Capabilities

### New Capabilities
<!-- none — all four bugs are fixes to already-specified capabilities -->

### Modified Capabilities
- `windows-gui-wizard`: adds requirements for surfacing stream/engine failures on a terminal Error page (never a silent process exit) and for disabling horizontal scroll on list pages; adds a Review scenario guaranteeing chosen custom components appear in the Review summary.
- `windows-gui-engine-client`: adds a requirement that the client isolates the GUI from child-process failures — UTF-8 stdout/stderr decoding and tolerance of non-JSON stream lines, with process/stream failures surfaced (not swallowed, not fatal-uncaught).

## Impact

- **Affected code (C# only):**
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/Navigation/ShellViewModel.cs` — primary fix (try/catch → `CompleteState.Error`).
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/ProcessInstallerEngineClient.cs` — UTF-8 encoding on `CreateProcess`.
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/InstallProgressSnapshotParser.cs` — non-JSON line tolerance.
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/App.xaml.cs` — secondary global handler.
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/Themes/PageTemplates.xaml` — `HorizontalScrollBarVisibility="Disabled"` on every `ListBox`.
- **Tests:** `ShellViewModelTests`, `ProcessInstallerEngineClientTests`, and a new Components→Review integration test in the Host test project.
- **Explicitly out of scope:** any Go code under `cmd/` or `internal/`; committing; building/bundling the installer. Bug D may reveal a future Go-side JSON-contract investigation — recorded as a follow-up, not a task here.
- **No breaking changes.**
