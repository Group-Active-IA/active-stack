<!-- Strict TDD (C#, `dotnet test`): every group follows RED → GREEN → TRIANGULATE → REFACTOR.
     Before modifying a file, run its existing tests as a safety-net baseline; a pre-existing
     failure is reported, not fixed. New tests MUST drive the REAL path (stream consumption /
     process configuration), never the `JsonCommandRunnerOverride` seam that let A/B ship.
     No Go code (`cmd/`, `internal/`). No commit. No build/bundle (operator-only, later). -->

## 1. Bug A — shell surfaces stream failures on the Error page (primary fix, D1)

- [x] 1.1 Safety net: run `ShellViewModelTests` and capture the passing baseline (do not fix any pre-existing failure — report it).
- [x] 1.2 RED: add a `ShellViewModelTests` case using a fake `IStreamTriggerPage` whose `StartStream` returns an `IAsyncEnumerable<InstallProgressSnapshot>` that throws mid-enumeration; drive the shell to that page and `await shell.AdvanceAsync()`; assert it does NOT propagate, the current page is `CompletePageViewModel`, and its `State == CompleteState.Error`. Confirm it fails against today's un-caught shell.
- [x] 1.3 GREEN: in `Navigation/ShellViewModel.cs` `RunStreamAndAdvanceToCompleteAsync`, wrap the `await progressPage.ConsumeAsync(triggerPage.StartStream(...))` call in try/catch; on catch, navigate to a `CompletePageViewModel` built from a synthesized failure snapshot (`Success = false`) with `hadRollback = false` and `hadDegradedSteps = false` (so `DetermineState` → `Error`), passing the caught exception's message as the snapshot message, and `_language`. Do not re-throw. Run until 1.2 passes.
- [x] 1.4 TRIANGULATE: add a second case where the failure message is a specific string and assert the Error page's `Message` carries it (rules out a hardcoded/blank message); confirm both cases green.
- [x] 1.5 REFACTOR: tidy the catch block (e.g. extract a small `BuildErrorSnapshot`/`ToErrorPage` helper if it clarifies) with tests green after each step.

## 2. Bug B — engine client UTF-8 encoding + parser tolerance (D2)

- [x] 2.1 Safety net: run `ProcessInstallerEngineClientTests` and capture the passing baseline.
- [x] 2.2 RED: add a `ProcessInstallerEngineClientTests` case that obtains the `ProcessStartInfo` the client configures (via a minimal internal `CreateProcess`/config seam exposed through `InternalsVisibleTo`, mirroring the existing `Build*Args` seams) and asserts `StandardOutputEncoding` and `StandardErrorEncoding` are both `Encoding.UTF8`. Confirm it fails today.
- [x] 2.3 GREEN: in `ProcessInstallerEngineClient.CreateProcess`, set `StandardOutputEncoding = Encoding.UTF8` and `StandardErrorEncoding = Encoding.UTF8` on the `ProcessStartInfo`. Run until 2.2 passes.
- [x] 2.4 RED: add `InstallProgressSnapshotParser` tests — (a) a non-JSON line (e.g. `"warning: something"`) does NOT throw and returns an ignorable/log-type snapshot carrying the raw line; (b) a valid JSON line still parses to the correct typed snapshot (regression guard). Confirm (a) fails today (it throws `JsonException`).
- [x] 2.5 GREEN: harden `InstallProgressSnapshotParser.Parse` to catch the JSON parse failure and return a log-type snapshot (raw line as message) instead of throwing. Run until 2.4 passes.
- [x] 2.6 TRIANGULATE: add a second malformed-line case (different non-JSON shape, e.g. a bare token) to rule out matching one literal; and assert a genuinely empty/whitespace line is still handled per existing behavior. Confirm green.
- [x] 2.7 REFACTOR: clean up the parser branch; keep the "non-zero exit still throws" contract in `RunStreamingCommandAsync`/`RunJsonCommandViaProcessAsync` untouched (only the per-line parser is softened). Tests green after each step.

## 3. Bug C — disable horizontal scroll on all list pages (D3)

- [x] 3.1 In `Themes/PageTemplates.xaml`, add `ScrollViewer.HorizontalScrollBarVisibility="Disabled"` to every `ListBox` element: Language, Hub, InstallType, Components, Permissions, UninstallMode, UninstallStrategy (both the strategy list and the backup-selector list), and StarterCatalog — 9 ListBox elements across 8 page templates.
- [x] 3.2 Confirm the XAML compiles cleanly (`dotnet build` of the Host project only if the operator requests a build; otherwise rely on the test build). Visual confirmation is deferred to the manual smoke pass (task 5). No automated WPF-layout assertion is realistic.

## 4. Bug D — Components → Review integration test as diagnosis (D4)

- [x] 4.1 Safety net: run `ComponentsPageViewModelTests` and `ReviewPageViewModelTests` and capture the passing baseline.
- [x] 4.2 RED/DIAGNOSTIC: add a new integration test (Host test project) that constructs a real `ComponentsPageViewModel` over a shared `InstallSelection` + `InstallerSessionState`, toggles a specific set of component checkboxes (including leaving `permissions` forced-on and toggling at least one custom component on and one off), then constructs a real `ReviewPageViewModel` over the SAME `InstallSelection` and session, and asserts `ComponentsSummary` lists exactly the expected selected components. Run it and record the outcome.
- [x] 4.3 BRANCH on 4.2's outcome:
  - If 4.2 FAILS → the bug is the C# wiring gap: complete the RED→GREEN cycle by fixing the flow (`ComponentsPageViewModel.SyncSelection` / `InstallSelection.CustomIds` / `ReviewPageViewModel.ComponentsSummary`) until 4.2 passes, then TRIANGULATE with a second selection shape.
  - If 4.2 PASSES → the C# layer is correct; keep the test as durable coverage, and record an OUT-OF-SCOPE follow-up (the defect lives in real Go-engine data — candidates `custom_components`/`forced_components`). Do NOT touch Go code in this change.
- [x] 4.4 **RESOLVED: 4.2 PASSED immediately (GREEN on first run, no fix needed).** `installer/windows/ui/ActiveStack.Bootstrapper.Host.Tests/ComponentsToReviewIntegrationTests.cs` constructs a real `ComponentsPageViewModel` and a real `ReviewPageViewModel` over the same `InstallerSessionState`/`InstallSelection`, toggles components (openspec on, engram on, extra off, permissions forced-on) and separately verifies the untouched recommended-defaults path — both green. **Conclusion: the C# wiring (`ComponentsPageViewModel.SyncSelection` → `InstallSelection.CustomIds` → `ReviewPageViewModel.ComponentsSummary`) is correct and is not the source of the smoke-tested Bug D symptom.** The test is kept as permanent regression coverage. The defect the user observed must live in real Go-engine response data for `windows options` (candidate fields: `custom_components` / `forced_components`) — e.g. the engine returning different/stale component ids than what the GUI displays and lets the user toggle. **This is an explicit OUT-OF-SCOPE follow-up for a future change; no Go code (`cmd/`, `internal/`) was touched in this change**, per the hard proposal/design constraint.

## 5. Bug A safety net + verification

- [x] 5.1 Add the secondary net (D5): in `App.xaml.cs`, subscribe `DispatcherUnhandledException` (optionally `AppDomain.CurrentDomain.UnhandledException` / `TaskScheduler.UnobservedTaskException`) to log and mark handled so an UNANTICIPATED fault degrades visibly instead of a silent kill. Keep it minimal — it is explicitly secondary to task 1's targeted catch, not a replacement. **Deviation note:** no dedicated automated test — `System.Windows.Application` is a per-process singleton requiring a live STA `Dispatcher`, impractical to unit-test in xUnit without a real WPF host process; consistent with D5's role as a minimal, secondary net (D1 already carries full TDD coverage as the primary fix).
- [x] 5.2 Run the full `dotnet test` suite for `ActiveStack.Bootstrapper.Core.Tests` and `ActiveStack.Bootstrapper.Host.Tests`; confirm 0 failures, including every pre-existing test (safety net). Record the pass counts. **Result: Core.Tests 47/47, Host.Tests 154/154** (145 baseline + 9 new: 2 `ShellViewModelTests` for Bug A, 1 `ProcessInstallerEngineClientTests` for the UTF-8 encoding seam, 4 `InstallProgressSnapshotParserTests` for Bug B parser tolerance, 2 `ComponentsToReviewIntegrationTests` for the Bug D diagnosis).
- [x] 5.3 Manual smoke checklist (to be run by the operator, not automated here):
  - Spanish language selected end to end.
  - Accented text correct everywhere, including long detail-panel descriptions (verifies the UTF-8 `StandardOutputEncoding`/`StandardErrorEncoding` fix, D2).
  - Permissions list — and every other list page (Language, Hub, InstallType, Components, UninstallMode, UninstallStrategy, StarterCatalog) — wraps long text with no horizontal scrollbar (D3).
  - Custom mode with components selected → Review shows the correct list (the C# wiring is now pinned by `ComponentsToReviewIntegrationTests`; if the summary is still wrong in a real run, the defect is the recorded Go-data follow-up from task 4.4, not a C# regression).
  - Install → Progress → a VISIBLE Complete/Error screen: force a real failure (e.g. disconnect network for an external harness, or point `ACTIVE_STACK_ENGINE_PATH` at a non-zero-exit binary) and confirm the wizard lands on the Error page (D1) instead of the window silently closing.
