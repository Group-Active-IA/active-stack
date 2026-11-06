## Context

The Windows GUI installer is a WPF app (`installer/windows/ui/**`, C#/.NET) that drives the Go engine as a child process, streaming its JSONL progress events into a wizard. A manual smoke test of the recently-shipped i18n + detail-panel work (changes L1–L5, all archived) found four defects, all confirmed by direct code reading and all confined to the C# layer. The Go engine was verified clean in the two suspected spots (stdout encoding and `--lang` passing).

Current state, as read from source:

- `ShellViewModel.RunStreamAndAdvanceToCompleteAsync` (`Navigation/ShellViewModel.cs`) calls `progressPage.ConsumeAsync(triggerPage.StartStream(...))` with **no** exception handling around it.
- `ProcessInstallerEngineClient.RunStreamingCommandAsync` throws `InvalidOperationException` when the engine exits non-zero (line ~149); `RunJsonCommandViaProcessAsync` does the same (~183). `InstallProgressSnapshotParser.Parse` throws `JsonException` (uncaught) on any non-JSON stdout line.
- `CreateProcess` sets `RedirectStandardOutput/Error` and `UseShellExecute=false` but does **not** set `StandardOutputEncoding`/`StandardErrorEncoding`.
- There is **no** unhandled-exception handler anywhere: `App.xaml.cs` only calls `InitializeComponent()`; no `DispatcherUnhandledException`, no `AppDomain.UnhandledException`, no `TaskScheduler.UnobservedTaskException`. WPF's default is to terminate the process silently.
- `CompleteState.Error` already exists (`Pages/CompleteState.cs`) and `CompletePageViewModel.DetermineState` already maps a non-success, non-rollback snapshot to `Error` — but that path is only reachable from an engine-emitted terminal snapshot, never from a thrown exception.
- `PageTemplates.xaml` has 9 `ListBox` elements across 8 page `DataTemplate`s; none sets `ScrollViewer.HorizontalScrollBarVisibility`, so long content scrolls horizontally instead of wrapping.
- The Components→Review data flow (`ComponentsPageViewModel.SyncSelection` → `InstallSelection.CustomIds` → `ReviewPageViewModel.ComponentsSummary`) reads as correctly wired; no integration test exercises it end-to-end — each fixture fakes the `InstallSelection.CustomIds` midpoint independently.

Why no test caught A/B: the only test seam (`JsonCommandRunnerOverride`) bypasses `CreateProcess`/`RunJsonCommandViaProcessAsync` and the streaming path entirely — every existing test injects JSON as a C# string and never reads real child-process stdout.

Governance: this touches the `Host` view-model/adapter layer (MEDIUM — implement with checkpoints). It does NOT touch `backup`/`filemerge`/`pipeline` (the HIGH-risk Go packages) or any Go code at all.

## Goals / Non-Goals

**Goals:**
- On any failure while consuming an install/uninstall/starter stream (engine non-zero exit, thrown parse error, any other exception), the wizard lands on a visible terminal Error page — never a silent window close.
- The engine client decodes child-process stdout/stderr as UTF-8 and tolerates a non-JSON stdout line without throwing an uncaught exception.
- List pages wrap long content instead of exposing a horizontal scrollbar.
- Close the coverage gaps: new tests drive the real stream-consumption path and the real process-config path (not the `JsonCommandRunnerOverride` seam), and a real Components→Review integration test exists.
- Diagnose Bug D: determine whether the Custom-selection→Review defect is C# wiring or Go-engine data, via the integration test's outcome.

**Non-Goals:**
- Any Go code change (`cmd/`, `internal/`). If Bug D's test passes, the Go-contract investigation is a recorded follow-up, not part of this change.
- New features, new wizard pages, or restyling beyond the one-line horizontal-scroll fix.
- Committing or building/bundling the installer (operator does these on explicit request, in a later session).
- A full live-process integration test if it proves flaky to isolate — a `CreateProcess`-configuration assertion is the accepted floor (see D2).

## Decisions

### D1 — Primary fix: try/catch in the shell around stream consumption → `CompleteState.Error` (NOT the global handler)

`RunStreamAndAdvanceToCompleteAsync` wraps the `await progressPage.ConsumeAsync(triggerPage.StartStream(...))` call in a try/catch. On catch, it navigates to a `CompletePageViewModel` built from a synthesized failure snapshot (`Success = false`), with `hadRollback = false`, so `CompletePageViewModel.DetermineState` resolves to `CompleteState.Error`; the caught exception's `Message` becomes the page's `Message`. The exception is swallowed after being turned into UI state (it must not re-propagate).

- **Why here:** this is the exact seam where the engine/stream exception crosses from the streaming `IAsyncEnumerable` into the WPF dispatcher via the `async void` advance handler. Catching it here converts a fatal, invisible failure into a first-class terminal UI state that already exists and is already localized (`complete.state.error.title`/`.label` keys are already present — `CompletePageViewModel` uses them). No Core/localization change is needed.
- **Alternative rejected:** relying solely on a global `DispatcherUnhandledException` handler. That can only show a generic dialog and cannot cleanly resume the wizard into its own Error page; it also cannot distinguish a stream failure from any other dispatcher fault. It is a net, not a fix.
- **Ordering is explicit:** D1 is the primary fix; D5 (global handler) is a secondary safety net. This ordering is binding.

### D2 — Engine client: UTF-8 encoding + non-JSON-line tolerance, with failures still surfaced

- `CreateProcess` sets `StandardOutputEncoding = Encoding.UTF8` and `StandardErrorEncoding = Encoding.UTF8` on the returned `ProcessStartInfo` (fixes accented text corruption, Bug B). This is directly assertable: a `ProcessInstallerEngineClientTests` case builds the `ProcessStartInfo` via the client and asserts both encodings are UTF-8 — the accepted coverage floor when a live-process test is impractical to isolate reliably.
- `InstallProgressSnapshotParser.Parse` no longer lets a `JsonException` escape on a malformed/non-JSON line. It catches the parse failure and returns an ignorable/log-type snapshot (e.g. a snapshot whose type marks it as raw log output carrying the original line as its message) instead of throwing. Rationale: the engine may interleave a non-JSON diagnostic line; one bad line must not abort the whole stream. The streaming loop already skips blank lines — this extends the same tolerance to unparseable lines.
- **Failures are still surfaced, not hidden:** a genuine non-zero engine exit still throws from `RunStreamingCommandAsync`/`RunJsonCommandViaProcessAsync`; D1 catches it and shows the Error page. We harden the *parser* against a single bad line, not the *exit code* contract — a real failure must still reach the user via the Error page.

### D3 — Horizontal scroll disabled on every list ListBox

Add `ScrollViewer.HorizontalScrollBarVisibility="Disabled"` to each of the 9 `ListBox` elements in `PageTemplates.xaml` (Language, Hub, InstallType, Components, Permissions, UninstallMode, UninstallStrategy strategies, UninstallStrategy backup selector, StarterCatalog). Their inner `TextBlock`s already use the wrapping body style / `TextWrapping="Wrap"`, so disabling horizontal scroll makes long descriptions wrap. This is transversal: the user noticed it on Permissions, but the fix is applied uniformly so no list page regresses. No automated WPF-layout assertion is realistic; verification is the manual smoke pass.

### D4 — Bug D as TDD-diagnosis, not a predetermined fix

Write the missing integration test FIRST: construct a real `ComponentsPageViewModel` over a real `InstallSelection`, toggle component checkboxes, then construct a real `ReviewPageViewModel` over the SAME `InstallSelection` and `InstallerSessionState`, and assert `ComponentsSummary` reflects exactly the toggled selection. The test's outcome branches the work:

- **If it FAILS:** the bug is the suspected C# wiring gap; the same RED→GREEN cycle fixes it (in `ComponentsPageViewModel`/`ReviewPageViewModel`/`InstallSelection`) and the test locks it.
- **If it PASSES:** the C# layer is correct; the defect lives in real data the Go engine returns in production (candidate fields: `custom_components` / `forced_components`). This is recorded as an out-of-scope Go-contract follow-up. This change does NOT preemptively touch Go code.

`tasks.md` MUST encode this branch — the fix task is conditional on the diagnostic outcome.

### D5 — Global unhandled-exception handler as secondary net

`App.xaml.cs` subscribes `DispatcherUnhandledException` (and may add `AppDomain.CurrentDomain.UnhandledException` / `TaskScheduler.UnobservedTaskException`) to log and avoid a silent kill for faults that escape D1's targeted catch. Explicitly secondary: it exists so an *unanticipated* fault degrades to something visible, not to replace D1. Kept minimal.

## Risks / Trade-offs

- **[A live-process stream test is hard to isolate reliably (needs a real child process / cross-platform runner).]** → Primary coverage is at the seam that actually failed: a `ShellViewModelTests` case with a fake `IStreamTriggerPage` whose `StartStream` throws mid-`IAsyncEnumerable`, asserting `AdvanceAsync` does not propagate and lands on `CompleteState.Error`. The engine client gets a `CreateProcess`-encoding assertion plus parser-level unit tests. This drives the real shell path and the real parser without a flaky live process.
- **[Swallowing exceptions in the parser could hide a real engine failure.]** → We only downgrade a single unparseable *line* to a log snapshot; the non-zero *exit code* still throws and still routes to the Error page (D2). Real failures stay visible.
- **[Disabling horizontal scroll could clip genuinely wide, non-wrapping content.]** → All list content is text using wrapping styles; there is no fixed-width non-wrapping content in these templates. Manual smoke confirms.
- **[Bug D might not reproduce (test passes), leaving the user-reported symptom unexplained in this change.]** → Accepted by design: the test still adds durable coverage and pins the C# contract, and the outcome precisely localizes the remaining suspect (Go data) as a scoped follow-up rather than guesswork.

## Migration Plan

No deployment/runtime migration — these are behavioral bug fixes to an unreleased-to-general-users installer build. Rollback is a straight revert of the C# edits. No data or schema changes. Build/bundle is out of scope and happens only on explicit operator request later.

## Open Questions

- Bug D root cause (C# wiring vs. Go data) is resolved by running D4's diagnostic test during apply — not decidable by static reading, deliberately deferred to the RED step.
