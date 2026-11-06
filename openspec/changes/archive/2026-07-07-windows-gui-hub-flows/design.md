## Context

C4 (`windows-gui-wizard-install`, implemented 35/35, not yet archived) delivered the wizard shell and the full Install flow: a thin `ShellViewModel` hosting an `IWizardPage` in a `ContentControl`, pure `WizardFlow.NextPage`/`PreviousPage` functions over an `InstallSelection`, a generic `ProgressPageViewModel`, a `CompletePageViewModel`, and a Hub whose non-Install entries are disabled-with-tooltip (C4 decision D8). C3 (archived) already shipped everything the remaining flows consume: `IInstallerEngineClient` declares `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, `RunUninstallAsync`; `InstallArguments` has `BuildStarterInstallArgs`, `BuildUninstallArgs` (`--restore-manifest <path>`), `BuildBackupActionArgs`; and the parsers (`StarterCatalogParser`, `BackupListParser`, `BackupActionResultParser`, `UninstallOptionsParser`) produce `StarterChoice`, `BackupEntry` (with `Id`, `ManifestPath`, `DisplayLabel`), `BackupActionResult`, `UninstallOptions`.

C5 is pure `Host`-project UI wiring (governance MEDIUM). The Go engine, `internal/backup`, `internal/uninstall`, the C3 contract, and Burn/`Process` wiring are untouched. TUI parity references: `internal/tui/uninstall_screen.go`, `internal/tui/starters_screen.go`, `internal/tui/backups_screen.go`.

The shell as C4 shipped it hardcodes the Install flow (Hub always advances to Assistants; the Review→Installing→Complete stream run is Install-specific). The core work of C5 is to **generalize the shell to dispatch multiple operations** while keeping every navigation decision a pure function, then add the three flows' pages.

## Goals / Non-Goals

**Goals:**
- Flip the Hub's Starters / Manage Backups / Uninstall entries on and route each to its flow; keep Update Stack permanently disabled.
- Deliver the Uninstall wizard, the Starters flow, and the Backups management page with TUI-parity behavior and confirmation texts.
- Keep all navigation as pure, unit-testable per-operation flow functions; reuse the generic progress/complete pages.
- Make the folder picker and the engine client injectable so every new VM is testable without a UI thread or a real process.

**Non-Goals:**
- Any change to the Go engine, `internal/backup`, `internal/uninstall`, or the C3 Core/client contract (consumed as-is).
- Implementing Update Stack (permanent TUI parity: disabled "Coming soon").
- Backup pinning, backup creation UI, or uninstall dry-run toggles (not in the TUI hub scope).
- Visual theming beyond the functional wizard already established in C4.

## Decisions

### D1 — Separate pure flow-function sets per operation (not one monolithic switch)
Install keeps its existing `WizardFlow`. Uninstall and Starter each get their own pure `NextPage`/`PreviousPage` set (`UninstallFlow`, `StarterFlow`) over their own selection type. **Why over extending the single `WizardFlow` switch**: the three flows share no branching predicates (Install branches on mode/tier-capability; Uninstall is linear; Starter is linear), so one switch would interleave unrelated logic and force every flow's tests to construct an `InstallSelection`. Separate sets keep each function trivially pure and let each flow's transition table be asserted in isolation, mirroring the distinct TUI screen groups. This is the recorded answer to the plan's "separate flow functions per operation are fine — record the decision."

### D2 — Shell tracks an active `WizardOperation`; the Hub sets it
`ShellViewModel` gains a `WizardOperation` (Install | Uninstall | Starter | Backups | None) set when the user advances from the Hub, derived from `HubPageViewModel.SelectedEntryId`. Back/Next dispatch to the active operation's flow functions. Advancing from the Hub begins the selected operation instead of assuming Install; returning Back to the Hub resets the operation to `None` so another entry can be chosen. **Why**: the shell must host four destinations from one entry menu; a single explicit operation field keeps the dispatch a small, testable switch and avoids leaking flow choice into the page VMs.

### D3 — Extend the single flat `WizardPageId` enum
Add `UninstallAgents`, `UninstallMode`, `UninstallStrategy`, `UninstallConfirm`, `StarterCatalog`, `StarterTarget`, `StarterReview`, `Backups` to the existing enum; reuse `Installing` (generic Progress) and `Complete` for all streaming flows. **Why over per-operation enums**: the shell's `CurrentPageId`, `CanGoBack`, and page-construction switch already key off one enum; a flat enum keeps those in one place. The active-operation field (D2) disambiguates the shared terminal pages.

### D4 — One selection model per operation
`UninstallSelection` (agents, mode, strategy id, `RequiresManifest`, selected `BackupEntry?`) and `StarterSelection` (starter id, project path, agents) are separate mutable models threaded through their pages, exactly as `InstallSelection` is for Install. Flow functions read them and never mutate. **Why**: each operation's fields are disjoint; a shared bag would carry mostly-null fields and blur which flow reads what.

### D5 — Restore threads the chosen backup's manifest path through the existing `RunUninstallAsync` `backupId` parameter
C3's `RunUninstallAsync(agents, mode, strategy, backupId, ...)` and `InstallArguments.BuildUninstallArgs(..., restoreManifestPath)` disagree in naming, but the engine flag is `--restore-manifest <path>` (confirmed in engram/C3 learnings, not `--id`). C5 passes the selected `BackupEntry.ManifestPath` as the `backupId` argument; `ProcessInstallerEngineClient` forwards it to `BuildUninstallArgs`' `restoreManifestPath`. **Why over changing the signature**: the C3 client contract is archived; the value semantics (a string identifying the backup to restore) are unchanged and the path is exactly what the engine needs. The mismatch is documented here and pinned by a test asserting the manifest path reaches the args. Alternative (rename the C3 param) rejected — it reopens an archived contract for no behavioral gain.

### D6 — `IFolderPicker` abstraction for the starter target
A new `IFolderPicker` (`string? PickFolder()`) is injected into `StarterTargetPageViewModel`; the production implementation wraps .NET 8's `Microsoft.Win32.OpenFolderDialog`, and tests supply a fake returning a canned path (including one with a space). **Why**: `OpenFolderDialog` needs an STA UI thread and blocks — untestable inline. The seam keeps the VM's advance-gate logic (non-empty path + ≥1 agent) fully unit-testable. The picked path is stored and passed as a single value so `BuildStarterInstallArgs` preserves spaces (the C3 `ArgumentList` guarantee).

### D7 — Backups is a standalone page with its own action state machine, not a wizard step
`BackupsPageViewModel` implements `IWizardPage` so the shell can host it, but it has no forward wizard step: `CanAdvance` is always false (the shell's primary "Next" is suppressed on this page), Back returns to the Hub, and the per-item Restore/Delete/Rename actions are internal commands. It holds an `Action` state (`None | Restore | Delete | Rename`) mirroring `internal/tui/backups_screen.go`'s `backupAction`, with a `RenameInput` string for the rename case and a `LastMessage` for results. Only one action confirmation is active at a time; confirming calls `RunBackupActionAsync` then re-lists via `ListBackupsAsync`; cancelling makes no engine call. **Why**: the TUI models backups as a single screen with inline confirmations, not a linear wizard; forcing it into Next/Back pages would misrepresent it and complicate the flow functions.

### D8 — Uninstall agents require an explicit non-empty selection (divergence from TUI)
The TUI treats an empty agent set as "all"; the GUI pre-checks all `detected_agents` and gates advance on ≥1 selected, matching the Install flow's Assistants min-1 pattern. **Why**: the wizard has no dedicated "all" affordance, and for a destructive operation an explicit selection is safer and clearer. Recorded as an intentional divergence.

### D9 — Generalize the shell's "trigger page → stream → progress → complete" run
C4's `RunInstallAndAdvanceToCompleteAsync` is Install-specific (it casts to `ReviewPageViewModel` and calls `StartInstall`). C5 generalizes it: the page that triggers a stream (Install Review, Uninstall Confirm, Starter Review) exposes a common capability — return an `IAsyncEnumerable<InstallProgressSnapshot>` when advanced — and the shell runs any such stream through the shared `ProgressPageViewModel` and then `CompletePageViewModel`. The terminal event type differs by flow (`install_finished` / `uninstall_finished` / `starter_finished`) but the generic progress/complete pages already key off the snapshot's success/degraded/rollback fields, so no per-flow renderer is needed. **Why**: avoids three near-duplicate run methods and keeps the streaming/terminal behavior in one place (matching C4 D4's "ProgressPageViewModel is generic" intent).

## Risks / Trade-offs

- **[MODIFIED Hub requirement can't resolve until C4 is archived]** → C5's delta MODIFIES `windows-gui-wizard`, whose base spec only materializes when C4 archives. The C1–C6 plan archives C4 before C5 (both at C6 after visual review). Mitigation: keep the MODIFIED requirement header byte-identical to C4's ("Hub surfaces all destinations with unimplemented ones disabled") so archive-time matching succeeds; archive C4 first.
- **[Shell generalization (D2/D9) could regress the Install flow]** → the Install flow's existing `WizardFlowTests`, `ShellViewModelTests`, and page tests stay green as the pinning safety net; generalize by extraction (introduce the operation field and generic run method without changing Install's observed transitions), running the C4 suite after each step.
- **[Restore manifest-path mismatch (D5) ships silently wrong]** → a test asserts that confirming a restore uninstall passes the selected `BackupEntry.ManifestPath` all the way to `BuildUninstallArgs`' `--restore-manifest`, and that a `targeted` uninstall omits it.
- **[Backups re-list after delete could desync the selection/cursor]** → re-list from `ListBackupsAsync` after every action (source of truth is the store, per the hub-contract spec) rather than mutating the in-memory list, so the view can't drift from the store; a test covers delete-then-relist yielding the reduced list.
- **[Folder picker STA/threading in production]** → confined behind `IFolderPicker`; the production wrapper is thin and exercised only in the C6 manual smoke, never in unit tests.
- **[Implicit `DataTemplate` for a new page VM resolves to a blank page]** → add each new page-VM `DataTemplate` to `Themes/PageTemplates.xaml` in the same merged dictionary C4 uses; the C6 smoke walks every Hub destination.

## Migration Plan

1. Land the navigation extensions (new `WizardPageId` values, `WizardOperation`, `UninstallFlow`/`StarterFlow`, `UninstallSelection`/`StarterSelection`) alongside the C4 code — Install flow untouched and green.
2. Generalize `ShellViewModel` (operation dispatch D2 + generic stream run D9) with the C4 suite as the safety net.
3. Add the Uninstall pages, then the Starter pages (+ `IFolderPicker`), then the Backups page, each test-first.
4. Flip `HubPageViewModel` enablement and wire each entry to its operation; add the new `DataTemplate`s.
5. **Rollback**: all changes are confined to the `Host` project; revert the C5 commits to return to C4's shell (Install-only, other entries disabled). No engine/installer state is touched.
6. **Archive order**: archive C4 (`windows-gui-wizard-install`) before C5 so C5's MODIFIED Hub requirement resolves.

## Open Questions

- None blocking. The confirmation and warning strings are fixed by TUI parity (`OVERWRITE` / `PERMANENTLY DELETE` / "This will modify your agent configuration."). The default starter agents (claude, opencode) come from the plan. The restore flag semantics (D5) are settled from the C3 learnings.
