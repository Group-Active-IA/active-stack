# Tasks — windows-gui-hub-flows (C5)

> TDD-ordered where `dotnet test` allows: for each behavior write the failing test first (RED),
> minimum code to pass (GREEN), triangulate a second case, then refactor. All work is confined to
> `installer/windows/ui/ActiveStack.Bootstrapper.Host` and its `.Tests` project. Consume the C3
> Core/client contract and parsers as-is; touch no Go, `internal/backup`, or `internal/uninstall`.
> Keep the C4 Install-flow test suite green throughout (safety net for the shell generalization).

## 1. Navigation substrate (contracts, no behavior yet)

- [x] 1.1 Extend `Navigation/WizardPageId` with `UninstallAgents`, `UninstallMode`, `UninstallStrategy`, `UninstallConfirm`, `StarterCatalog`, `StarterTarget`, `StarterReview`, `Backups` (reuse `Installing`/`Complete` as the generic progress/complete pages).
- [x] 1.2 Add `Navigation/WizardOperation` enum (`None`, `Install`, `Uninstall`, `Starter`, `Backups`).
- [x] 1.3 Add `Navigation/UninstallSelection.cs` — mutable model: `Agents` (list), `Mode` (string), `Strategy` (string), `RequiresManifest` (bool), `SelectedBackup` (`BackupEntry?`). Never mutated by flow functions.
- [x] 1.4 Add `Navigation/StarterSelection.cs` — mutable model: `StarterId` (string), `ProjectPath` (string), `Agents` (list, default claude+opencode).

## 2. Uninstall flow pure functions (test-first)

- [x] 2.1 RED: `Host.Tests/UninstallFlowTests.cs` — `NextPage` walks Hub → UninstallAgents → UninstallMode → UninstallStrategy → UninstallConfirm → Installing → Complete.
- [x] 2.2 RED: `PreviousPage` is the inverse (UninstallConfirm → UninstallStrategy, …, UninstallAgents → Hub).
- [x] 2.3 GREEN: implement `Navigation/UninstallFlow.cs` as pure `NextPage`/`PreviousPage` over `(WizardPageId, UninstallSelection)`; make group-2 tests pass; refactor.

## 3. Starter flow pure functions (test-first)

- [x] 3.1 RED: `Host.Tests/StarterFlowTests.cs` — `NextPage` walks Hub → StarterCatalog → StarterTarget → StarterReview → Installing → Complete; `PreviousPage` is the inverse.
- [x] 3.2 GREEN: implement `Navigation/StarterFlow.cs`; make tests pass; refactor.

## 4. Shell generalization: operation dispatch + generic stream run (test-first)

- [x] 4.1 RED: `ShellViewModelTests` — selecting the Uninstall Hub entry and advancing sets active operation Uninstall and moves to UninstallAgents (not Assistants); selecting Starters → StarterCatalog; selecting Manage Backups → Backups page.
- [x] 4.2 RED: `ShellViewModelTests` — Back from a flow's first page returns to Hub and resets active operation to `None`; the C4 Install-flow transition tests still pass unchanged.
- [x] 4.3 GREEN: add `WizardOperation` field to `ShellViewModel`, set from `HubPageViewModel.SelectedEntryId`; dispatch `NextPage`/`PreviousPage` to the active operation's flow functions (D2). Keep the C4 Install path behavior identical.
- [x] 4.4 RED: `ShellViewModelTests` — advancing from a stream-trigger page (Uninstall Confirm / Starter Review, mirroring Install Review) runs the returned stream through the shared `ProgressPageViewModel` and lands on `CompletePageViewModel`.
- [x] 4.5 GREEN: generalize `RunInstallAndAdvanceToCompleteAsync` into an operation-agnostic run over any trigger page that yields an `IAsyncEnumerable<InstallProgressSnapshot>` (D9); keep Install green; refactor.

## 5. Uninstall pages (test-first)

- [x] 5.1 RED: `UninstallAgentsPageViewModelTests` with a fake `IInstallerEngineClient` returning `UninstallOptions` — all `detected_agents` pre-checked, `CanAdvance` true; unchecking all → `CanAdvance` false (min-1 gate, D8).
- [x] 5.2 GREEN: `Pages/Uninstall/UninstallAgentsPageViewModel.cs` — multi-select checkboxes from `detected_agents`, writes `UninstallSelection.Agents`.
- [x] 5.3 RED: `UninstallModePageViewModelTests` — lite/full/custom radio; default and selection write `UninstallSelection.Mode`.
- [x] 5.4 GREEN: `Pages/Uninstall/UninstallModePageViewModel.cs`.
- [x] 5.5 RED: `UninstallStrategyPageViewModelTests` — `targeted` (default, requires_manifest false) → no backup selector, `CanAdvance` true; selecting `restore` (requires_manifest true) reveals a selector fed by `ListBackupsAsync` and `CanAdvance` is false until a backup is chosen, then true. Writes `Strategy`, `RequiresManifest`, `SelectedBackup`.
- [x] 5.6 GREEN: `Pages/Uninstall/UninstallStrategyPageViewModel.cs` — radio from strategies, conditional backup selector by `requires_manifest`.
- [x] 5.7 RED: `UninstallConfirmPageViewModelTests` — summary shows agents/mode/strategy (+ chosen backup on restore) and the warning "This will modify your agent configuration."; advancing yields the uninstall stream carrying agents, mode, strategy, and — for `restore` — the selected `BackupEntry.ManifestPath` (D5); a `targeted` confirm passes no manifest path.
- [x] 5.8 GREEN: `Pages/Uninstall/UninstallConfirmPageViewModel.cs` — invokes `RunUninstallAsync(agents, mode, strategy, SelectedBackup?.ManifestPath)`; make tests pass; refactor.

## 6. Starter pages + folder picker (test-first)

- [x] 6.1 RED: `IFolderPicker` fake test — `StarterTargetPageViewModel` uses an injected `IFolderPicker`; a canned returned path (including one containing a space) sets `ProjectPath` and, with default agents checked, `CanAdvance` true; no path chosen → `CanAdvance` false.
- [x] 6.2 GREEN: add `IFolderPicker` (`string? PickFolder()`) + a WPF `OpenFolderDialog` implementation; `Pages/Starters/StarterTargetPageViewModel.cs` with the injected picker and agent checkboxes defaulting to claude+opencode; writes `StarterSelection.ProjectPath`/`Agents`.
- [x] 6.3 RED: `StarterCatalogPageViewModelTests` with a fake client returning `StarterChoice`s — single-select list showing name/description/harness count; nothing selected → `CanAdvance` false; selecting → writes `StarterSelection.StarterId`, `CanAdvance` true.
- [x] 6.4 GREEN: `Pages/Starters/StarterCatalogPageViewModel.cs` (loads via `ListStartersAsync`).
- [x] 6.5 RED: `StarterReviewPageViewModelTests` — summary reflects starter/target/agents; primary label "Install"; advancing yields the starter install stream via `RunStarterInstallAsync(starterId, projectPath, agents)` with the path intact.
- [x] 6.6 GREEN: `Pages/Starters/StarterReviewPageViewModel.cs`.

## 7. Backups page + confirmation state machine (test-first)

- [x] 7.1 RED: `BackupsPageViewModelTests` with a fake client — loads and lists backups by `display_label`; an empty store shows the empty-state and no actions; `CanAdvance` is always false (D7).
- [x] 7.2 GREEN: `Pages/Backups/BackupsPageViewModel.cs` — loads via `ListBackupsAsync`, exposes items and an `Action` state (`None`/`Restore`/`Delete`/`Rename`) + `RenameInput` + `LastMessage`.
- [x] 7.3 RED: restore confirmation text is "This will OVERWRITE your current configuration." and confirming invokes `RunBackupActionAsync("restore", id, null)`; delete text is "This will PERMANENTLY DELETE this backup." → `RunBackupActionAsync("delete", id, null)`.
- [x] 7.4 RED: rename captures a description and confirming invokes `RunBackupActionAsync("rename", id, description)`; cancelling any action invokes nothing and returns to the list; only one action is active at a time.
- [x] 7.5 RED: after any action the page re-invokes `ListBackupsAsync` and shows a result message (delete-then-relist yields the reduced list).
- [x] 7.6 GREEN: implement the action state machine + re-list-after-action; make group-7 tests pass; refactor.

## 8. Hub enablement + view wiring

- [x] 8.1 RED: `HubPageViewModelTests` — Install, Starters, Manage Backups, and Uninstall are enabled and selectable; Update Stack stays disabled with a "Coming soon" tooltip.
- [x] 8.2 GREEN: flip `IsEnabled` on the three entries in `HubPageViewModel`; map each `SelectedEntryId` to its `WizardOperation` in `ShellViewModel`.
- [x] 8.3 Add one implicit `DataTemplate` per new page VM to `Themes/PageTemplates.xaml` (uninstall pages, starter pages, backups page); ensure the shell suppresses the primary "Next" on the Backups page.
- [x] 8.4 Wire `IFolderPicker` (production `OpenFolderDialog` wrapper) into the composition root (`App.xaml.cs`/`ActiveStackBootstrapperApplication`); no Burn/`Process` wiring change.

## 9. Full suite green

- [x] 9.1 Run `dotnet test` on `Host.Tests` (and `Core.Tests`) — all C4 + C5 tests green; confirm no Install-flow regression.
