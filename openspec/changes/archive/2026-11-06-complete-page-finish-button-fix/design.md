## Context

The Windows GUI wizard's terminal Complete page (`CompletePageViewModel`) shows a primary footer button, but nobody ever wired an action for it. The failure chain, root-caused against the live smoke build:

1. `CompletePageViewModel.CanAdvance` is hardcoded `true` (doc-comment: "the primary action here is Close/Finish"), so the footer primary button is enabled.
2. `ShellViewModel.PrimaryLabel` (`Navigation/ShellViewModel.cs:98`) has no case for the Complete page, so it falls through to the generic `shell.next` label ("Next" / "Siguiente").
3. Clicking calls `MainWindow.xaml.cs` `PrimaryButton_Click` → `ShellViewModel.AdvanceAsync()`. On Complete, `AdvanceAsync` is not on Language, the page is not an `IStreamTriggerPage`, and it is not the Hub — so it falls into `switch (_operation)`. `_operation` is still `WizardOperation.Install`, so it calls `MoveTo(WizardFlow.NextPage(WizardPageId.Complete, …))`.
4. `WizardFlow.NextPage` self-loops: `WizardPageId.Complete => WizardPageId.Complete` (`WizardFlow.cs:30`).
5. `MoveTo` → `CreatePage(WizardPageId.Complete)`, which has **no** case for Complete (its `_ =>` default throws `InvalidOperationException`, because Complete is only ever reached via the dedicated async stream path, never rebuilt by `CreatePage`).
6. The exception is thrown inside the `async void` click handler, crosses into the WPF dispatcher, and is caught by the global `DispatcherUnhandledException` net in `App.xaml.cs` (a secondary safety net from the prior smoke-bugfix change), which logs and marks it handled — silently. Net effect: the button "does nothing", every click.

The prior change (`installer-post-smoke-bugfixes`) added a targeted try/catch around **stream consumption** in `RunStreamAndAdvanceToCompleteAsync`. That catch does **not** cover this path: the throw here happens in the synchronous `MoveTo`/`CreatePage` branch of `AdvanceAsync`, not during stream enumeration. So this is a distinct, still-open defect.

Constraints: C# / WPF layer only (`installer/windows/ui/**`); no Go code; keep the fix minimal and scoped to this defect; do not redesign the navigation architecture. `ShellViewModel` is deliberately WPF-free and fully unit-tested (it does not own the `Window`); the codebase resolves Window-owned concerns through seams (`IFolderPicker`, an injected `persistLanguage` `Action`, `AutoInstallDriver`).

## Goals / Non-Goals

**Goals:**
- Clicking the primary button on the Complete page performs a real, visible **Finish/Close** action: it closes the wizard window.
- The primary button on the Complete page is labeled Finish/Close (localized), not "Next".
- `ShellViewModel` stays free of any WPF `Window` dependency and fully unit-testable.
- No exception is ever thrown from the Complete-page advance path.

**Non-Goals:**
- Redesigning the wizard navigation graph, the flow functions, or `CreatePage`.
- Changing the always-present "Cancel" footer button's label or behavior (it already closes the window; relabeling it context-aware is a separate, out-of-scope UX polish).
- Touching the existing stream-failure Error-page behavior or the `App.xaml.cs` global net.
- Any Go engine change.

## Decisions

### D1: Wire the primary action as a real Finish/Close (Option A), not disable it (Option B)

Two candidates were considered:

- **Option A — wire Primary as a real Finish/Close on Complete.** Give the Complete page a working terminal action that closes the wizard, labeled "Finish".
- **Option B — disable Primary on Complete** (mirror `ProgressPageViewModel.CanAdvance => false`), and let the user close via the existing "Cancel" button.

**Chosen: Option A.** It matches `CompletePageViewModel`'s own doc-comment intent ("the primary action here is Close/Finish") and the standard installer convention (a "Finish" button that dismisses a successful install). Option B leaves a confusingly-labeled "Cancel" as the *only* way to dismiss a screen that reads "Instalación completada" — a UX dead end, exactly what we're asked to avoid. Option A costs only a small event + label; it is not a redesign.

### D2: `ShellViewModel` raises a `CloseRequested` event; `MainWindow.xaml.cs` closes the window

`ShellViewModel` does not (and must not) own the WPF `Window`. Two sub-options for letting it trigger a close:

- **D2a — `CloseRequested` event on `ShellViewModel`.** `AdvanceAsync`, when on the Complete page, raises `CloseRequested`; `MainWindow`'s constructor subscribes and calls `Close()`.
- **D2b — code-behind checks the page type.** `PrimaryButton_Click` inspects `_viewModel.CurrentPage is CompletePageViewModel` and calls `Close()` directly instead of `AdvanceAsync()`.

**Chosen: D2a (event).** It keeps navigation ownership inside `ShellViewModel` (the code-behind stays a dumb forwarder — `PrimaryButton_Click` still just calls `AdvanceAsync()`), and it is unit-testable: a test subscribes to `CloseRequested` and asserts it fires when advancing on Complete, with no STA UI thread. D2b leaks page-type knowledge into the WPF code-behind (which is not unit-tested — it requires a live `Window`), splitting the "what does advancing do" decision across two files. D2a is consistent with the existing seam philosophy (`IFolderPicker`, injected `persistLanguage`): side effects the VM can't own are expressed as an abstraction/event the host wires up.

Mechanically, `AdvanceAsync` gets an early branch, placed **before** the `switch (_operation)` block (and after the existing `IStreamTriggerPage` check, which the Complete page is not):

```
if (CurrentPage is CompletePageViewModel)
{
    CloseRequested?.Invoke(this, EventArgs.Empty);
    return;
}
```

This short-circuits the broken `MoveTo(NextPage(Complete,…))` → `CreatePage(Complete)` throw path entirely. `CanAdvance` on the Complete page stays `true` (the Finish button is enabled), so the existing `if (!CurrentPage.CanAdvance) return;` guard at the top of `AdvanceAsync` does not block it.

### D3: `PrimaryLabel` returns a localized Finish label on Complete; new `shell.finish` key

`PrimaryLabel` currently is: stream-trigger page's own label, else `shell.next`. Add a Complete-page case ahead of the `shell.next` fallback:

```
public string PrimaryLabel =>
    CurrentPage is CompletePageViewModel
        ? UiStrings.Get(_language, "shell.finish")
        : (CurrentPage as IStreamTriggerPage)?.PrimaryLabel ?? UiStrings.Get(_language, "shell.next");
```

Add `shell.finish` to both `UiStrings` tables: `En` = `"Finish"`, `Es` = `"Finalizar"`. The existing Core en/es parity test enforces that every `En` key has an `Es` counterpart, so both must be added together — this is enforced, not optional.

`CurrentPage` already raises `PrimaryLabel` change notification whenever the page changes (`ShellViewModel.CurrentPage` setter calls `RaisePropertyChanged(nameof(PrimaryLabel))`), so the footer relabels to "Finish" automatically when the wizard reaches Complete. No extra notification wiring needed.

### D4: Leave `CanGoBack`, the Cancel button, and `CanAdvance` as-is

`CanGoBack` already excludes the Complete page (`ShellViewModel.cs:90-91`), so Back stays disabled there — correct, unchanged. The always-present "Cancel" button already closes the window; it remains as a secondary way out. `CompletePageViewModel.CanAdvance` stays `true` so the Finish button is enabled. Minimal surface: no change to any of these.

## Risks / Trade-offs

- **[Two footer buttons both close the window on Complete — "Finish" (primary) and "Cancel" (secondary).]** → Acceptable and low-harm: "Finish" is now the affirmative primary action a user naturally clicks. Relabeling/hiding "Cancel" contextually is explicitly out of scope (D4) to keep the fix minimal; it can be a later UX-polish change.
- **[The `CloseRequested` event is only meaningful when a real `Window` is wired (production); in unit tests it is observed but no window closes.]** → This is the intended seam boundary and is exactly what makes it testable. `MainWindow.xaml.cs` is the single production subscriber.
- **[If a future flow reaches Complete with a different operation active, the early `is CompletePageViewModel` branch still closes correctly regardless of `_operation`.]** → Desired: Complete is a shared terminal page across install/uninstall/starter flows; closing is operation-independent, and the branch is placed before the `_operation` switch so it applies uniformly.

## Migration Plan

Pure additive bug-fix in the C# UI layer; no data, no schema, no engine contract change. Rollback is a straight revert of the three touched files. Verification is unit tests (`ShellViewModelTests`, Core parity test) plus the operator's manual smoke pass (install to completion, click Finish, confirm the window closes).

## Open Questions

None.
