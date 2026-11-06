<!-- Strict TDD (C#, `dotnet test`): every group follows RED → GREEN → TRIANGULATE → REFACTOR.
     Before modifying a file, run its existing tests as a safety-net baseline; a pre-existing
     failure is reported, not fixed. Tests MUST drive the real path (raise CloseRequested via
     ShellViewModel.AdvanceAsync on a real CompletePageViewModel; assert PrimaryLabel off the
     real UiStrings table), never a bypass seam. No Go code (`cmd/`, `internal/`). No commit.
     No build/bundle (operator-only, later). All edits under `installer/windows/ui/**`. -->

## 1. Localized Finish label (`shell.finish`)

- [x] 1.1 Safety net: run `ActiveStack.Bootstrapper.Core.Tests` (includes the en/es `UiStrings` parity test) and capture the passing baseline (report, do not fix, any pre-existing failure).
- [x] 1.2 RED: add/extend a `UiStrings` test asserting `UiStrings.Get("en", "shell.finish") == "Finish"` and `UiStrings.Get("es", "shell.finish") == "Finalizar"`. Confirm it fails today (key absent → `Get` returns the raw key `"shell.finish"`).
- [x] 1.3 GREEN: add `["shell.finish"] = "Finish"` to the `En` table and `["shell.finish"] = "Finalizar"` to the `Es` table in `ActiveStack.Bootstrapper.Core/Localization/UiStrings.cs`. Run until 1.2 passes AND the existing en/es parity test stays green (both tables gain the key together).

## 2. Complete-page Finish label on the shell

- [x] 2.1 Safety net: run `ShellViewModelTests` and capture the passing baseline (report, do not fix, any pre-existing failure).
- [x] 2.2 RED: add a `ShellViewModelTests` case that drives the shell to the `CompletePageViewModel` (reuse the existing `ConfirmingReview_RunsInstallAndReachesComplete` path: Language → Hub(install) → Assistants → InstallType → Review → advance to Complete) and asserts `shell.PrimaryLabel == "Finish"` (English). Confirm it fails today (label falls through to `shell.next` → "Next").
- [x] 2.3 GREEN: in `Navigation/ShellViewModel.cs`, change `PrimaryLabel` to return `UiStrings.Get(_language, "shell.finish")` when `CurrentPage is CompletePageViewModel`, before the existing `IStreamTriggerPage`/`shell.next` fallback. Run until 2.2 passes.
- [x] 2.4 TRIANGULATE: add a Spanish case — construct the shell with `initialLanguage: "es"` (or select `es` on the Language page), reach Complete, and assert `shell.PrimaryLabel == "Finalizar"`; also assert a non-Complete page (e.g. Hub) still returns "Next"/"Siguiente" so the special-case didn't leak. Confirm all green.

## 3. Complete-page primary action closes the wizard (`CloseRequested`)

- [x] 3.1 RED: add a `ShellViewModelTests` case that reaches the `CompletePageViewModel`, subscribes to a new `shell.CloseRequested` event (capturing a bool `raised`), then `await shell.AdvanceAsync()`. Assert: `AdvanceAsync` did not throw (`Record.ExceptionAsync` returns null), `raised == true`, `client.RunInstallCallCount` is unchanged (advancing did NOT re-enter any flow), and `shell.CurrentPage` is still the `CompletePageViewModel`. Confirm it fails to compile/run today (`CloseRequested` does not exist, and — once stubbed — the current code throws `InvalidOperationException` via `CreatePage`).
- [x] 3.2 GREEN: in `ShellViewModel`, declare `public event EventHandler? CloseRequested;` and, in `AdvanceAsync`, add — after the `IStreamTriggerPage` branch and before the `if (_currentPageId == WizardPageId.Hub)` / `switch (_operation)` block — an early branch: `if (CurrentPage is CompletePageViewModel) { CloseRequested?.Invoke(this, EventArgs.Empty); return; }`. Run until 3.1 passes. (Do not change `CompletePageViewModel.CanAdvance`; it stays `true` so the top-of-method `if (!CurrentPage.CanAdvance) return;` guard does not block the Finish action.)
- [x] 3.3 TRIANGULATE: add a case proving Back is still disabled on Complete (`shell.CanGoBack == false`) and that invoking the primary action twice raises `CloseRequested` each time without throwing or navigating (idempotent terminal behavior). Confirm green.
- [x] 3.4 REFACTOR: tidy the new branch and `PrimaryLabel` (e.g. a private `bool OnCompletePage => CurrentPage is CompletePageViewModel` if it clarifies both call sites), with `ShellViewModelTests` green after each step.

## 4. Host wiring (production close)

- [x] 4.1 In `MainWindow.xaml.cs` constructor (after `DataContext = _viewModel;`), subscribe `_viewModel.CloseRequested += (_, _) => Close();` so the real window closes when the shell requests it. Keep `PrimaryButton_Click` unchanged (it still just calls `AdvanceAsync()`). Deviation note: no dedicated automated test — `MainWindow`/`Close()` require a live STA WPF `Window`, impractical to unit-test in xUnit; the testable behavior (the `CloseRequested` event firing) is fully covered by task 3, and this line is the minimal host subscription, consistent with the codebase's untested WPF code-behind seams.

## 5. Full suite + smoke verification

- [x] 5.1 Run the full `dotnet test` for `ActiveStack.Bootstrapper.Core.Tests` and `ActiveStack.Bootstrapper.Host.Tests`; confirm 0 failures including every pre-existing test (safety net). Record the pass counts and the number of new tests added. Result: Core.Tests 48/48 (47 baseline + 1 new), Host.Tests 161/161 (157 baseline + 4 new ShellViewModelTests cases; the 20/24 count difference from earlier ShellViewModelTests-only runs is because the Host.Tests suite also includes other page/view-model test files beyond ShellViewModelTests). 0 failures, 0 skipped.
- [ ] 5.2 Manual smoke checklist (operator, not automated here): run an install to completion; confirm the Complete page's primary button reads "Finish" (or "Finalizar" in Spanish), that clicking it closes the wizard window cleanly with no exception in the trace log, and that Back remains disabled on the Complete page. **Left unchecked — requires a live installer build (`installer\windows\build.bat`), which this apply run was explicitly instructed not to run; the operator should run this smoke pass separately.**
