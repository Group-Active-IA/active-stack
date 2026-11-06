## ADDED Requirements

### Requirement: Complete page's primary action finishes and closes the wizard

On the terminal Complete page, the shell's primary footer action SHALL be a Finish/Close action, not a navigation "Next". Its label SHALL be the localized `shell.finish` string ("Finish" / "Finalizar") — supplied for both `en` and `es` in `UiStrings` — rather than the generic `shell.next` label. Invoking the primary action while the current page is the `CompletePageViewModel` SHALL cause the wizard window to close, and SHALL NOT route into any flow's navigation. Concretely, `ShellViewModel` SHALL expose a `CloseRequested` event and, in `AdvanceAsync`, SHALL raise `CloseRequested` and return when the current page is the `CompletePageViewModel` — before reaching the operation switch that rebuilds pages via `CreatePage` (which has no Complete case and would otherwise throw). The host (`MainWindow`) SHALL subscribe to `CloseRequested` and close the window; `ShellViewModel` MUST NOT take a dependency on the WPF `Window`. Advancing from the Complete page MUST NOT throw and MUST NOT be silently swallowed by the global dispatcher handler. `CanGoBack` SHALL remain false on the Complete page (Back stays disabled), and the Complete page's `CanAdvance` SHALL remain true so the Finish action is enabled.

#### Scenario: Primary action on Complete raises CloseRequested instead of navigating

- **WHEN** the wizard has reached the `CompletePageViewModel` (after an install/uninstall/starter stream) and the shell's primary action is invoked
- **THEN** `ShellViewModel` raises its `CloseRequested` event, `AdvanceAsync` returns without throwing, and the current page remains the `CompletePageViewModel` (no navigation into any flow occurs)

#### Scenario: Primary label on Complete is the localized Finish label

- **WHEN** the current page is the `CompletePageViewModel` with the active language English (and, separately, Spanish)
- **THEN** the shell's `PrimaryLabel` is "Finish" in English and "Finalizar" in Spanish (the `shell.finish` string), not "Next"/"Siguiente"

#### Scenario: Advancing on Complete no longer throws through the dispatcher

- **WHEN** the primary action is invoked on the `CompletePageViewModel` while the active operation is still Install
- **THEN** no exception is thrown (the former `CreatePage` `InvalidOperationException` path is bypassed), so nothing reaches the global `DispatcherUnhandledException` net to be silently swallowed
