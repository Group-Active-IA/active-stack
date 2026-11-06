## Why

After a successful install, the Windows GUI wizard's Complete page shows a primary footer button that visibly does nothing when clicked. The Complete page's terminal "Close/Finish" action — intended by `CompletePageViewModel`'s own doc-comment and by the `MainWindow.xaml` footer comment — was never actually implemented: the primary button falls through to the Install flow's navigation, self-loops on `WizardPageId.Complete`, and hits `ShellViewModel.CreatePage`'s `_ =>` default which throws `InvalidOperationException`. That exception is thrown from the `async void` click handler, escapes to the WPF dispatcher, and is silently swallowed by the global `DispatcherUnhandledException` net (added in the prior smoke-bugfix change), so every click throws and is absorbed with zero UI feedback. The user is left on a "Instalación completada" screen with no working way to finish.

## What Changes

- The Complete page's primary footer action becomes a real **Finish/Close** action: advancing from the Complete page closes the wizard window instead of routing into install navigation and throwing.
- `ShellViewModel` gains a `CloseRequested` event it raises when the primary action is invoked on the Complete page; `AdvanceAsync` short-circuits the Complete page to raise that event and return, before ever reaching the `switch (_operation)` block that throws.
- `ShellViewModel.PrimaryLabel` returns a localized "Finish" / "Finalizar" label when the current page is the `CompletePageViewModel`, instead of the generic "Next" / "Siguiente".
- A new `shell.finish` string key is added to `UiStrings` in both `en` and `es` tables (keeping en/es parity intact).
- `MainWindow.xaml.cs` subscribes to `ShellViewModel.CloseRequested` and calls `Close()` — keeping WPF `Window` ownership out of `ShellViewModel`, consistent with the existing `IFolderPicker` / injected-`persistLanguage` seam pattern.

No Go code is touched. No new navigation architecture — the wizard's flow functions and page graph are unchanged; only the Complete page's terminal action is wired.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `windows-gui-wizard`: adds a requirement specifying that the Complete page's primary footer action finishes and closes the wizard (a localized Finish/Close action that signals the host to close the window), rather than self-looping into a thrown navigation exception.

## Impact

- **Affected code** (all under `installer/windows/ui/**`, C# only):
  - `ActiveStack.Bootstrapper.Host/Navigation/ShellViewModel.cs` — new `CloseRequested` event, Complete-page branch in `AdvanceAsync`, `PrimaryLabel` special-case for `CompletePageViewModel`.
  - `ActiveStack.Bootstrapper.Core/Localization/UiStrings.cs` — new `shell.finish` key in `En` and `Es`.
  - `ActiveStack.Bootstrapper.Host/MainWindow.xaml.cs` — subscribe `CloseRequested` → `Close()`.
- **Tests**: `ShellViewModelTests` (new cases for the Complete-page Finish behavior and label), `UiStrings` en/es parity test (already covers the new key automatically).
- **No impact** on the engine, Go code, the install/uninstall/starter flow graphs, or the existing stream-failure Error-page behavior.
