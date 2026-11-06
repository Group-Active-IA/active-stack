# Tasks — windows-gui-wizard-install (C4)

> TDD-ordered where `dotnet test` allows: for each behavior write the failing test first (RED),
> minimum code to pass (GREEN), triangulate a second case, then refactor. All work is confined to
> `installer/windows/ui/ActiveStack.Bootstrapper.Host` and its `.Tests` project. Consume the C3
> Core/client contract as-is (except the additive `RunInstallAsync` extension in group 7).

## 1. Navigation substrate (contracts, no behavior yet)

- [x] 1.1 Add `Navigation/IWizardPage.cs` — interface: `Title`, `Subtitle`, `CanAdvance` (bool + change notification); marker for pages that own footer nav.
- [x] 1.2 Add `Navigation/InstallSelection.cs` — mutable shared model: `Agents` (list), `Mode` (string), `CustomIds` (list), `Tier` (string). Threaded through the Install pages; never mutated by `WizardFlow`.
- [x] 1.3 Add `Navigation/WizardPageId` enum (Hub, Assistants, InstallType, Components, Permissions, Review, Installing, Complete) so flow functions are typed.

## 2. WizardFlow pure functions (test-first)

- [x] 2.1 RED: `Host.Tests/WizardFlowTests.cs` — `NextPage(InstallType, {mode:full, agents:[claude]}, tierCapable:[claude])` → Permissions (Components skipped).
- [x] 2.2 RED: `NextPage(InstallType, {mode:custom, ...})` → Components.
- [x] 2.3 RED: `NextPage(InstallType, {mode:full, agents:[gemini]}, tierCapable:[claude])` → Review (Permissions skipped); and `NextPage(Components, {gemini-only})` → Review.
- [x] 2.4 RED: `NextPage(Components, {mode:custom, agents:[claude]}, tierCapable:[claude])` → Permissions; `NextPage(Permissions, ...)` → Review.
- [x] 2.5 RED: `PreviousPage(Review, {mode:full, agents:[gemini]}, tierCapable:[claude])` → InstallType (skips Permissions + Components); `PreviousPage(Review, {mode:custom, agents:[claude]})` → Permissions.
- [x] 2.6 GREEN: implement `Navigation/WizardFlow.cs` as pure `NextPage`/`PreviousPage` over `(WizardPageId, InstallSelection, tierCapableAgents)`; tier-capability = intersection of `tierCapableAgents` with selected agents. Make all group-2 tests pass; refactor.

## 3. Assistants page (test-first)

- [x] 3.1 RED: `AssistantsPageViewModelTests` with a fake `IInstallerEngineClient` session — detected `claude` is pre-checked and `CanAdvance` is true.
- [x] 3.2 RED: zero selected agents → `CanAdvance` is false (min-1 gate).
- [x] 3.3 GREEN: `Pages/Install/AssistantsPageViewModel.cs` — multi-select checkboxes from available agents, prechecked from detect, recommended marked, writes `InstallSelection.Agents`; make tests pass; refactor.

## 4. InstallType + Components pages (test-first)

- [x] 4.1 RED: `InstallTypePageViewModelTests` — `full` preselected by default; selecting a card writes `InstallSelection.Mode`.
- [x] 4.2 GREEN: `Pages/Install/InstallTypePageViewModel.cs` — lite/full/custom cards with end-user labels (Quick/Complete/Custom), default full.
- [x] 4.3 RED: `ComponentsPageViewModelTests` — recommended components prechecked; `permissions` rendered checked+disabled; toggling `permissions` off leaves it checked and keeps it in `CustomIds`.
- [x] 4.4 GREEN: `Pages/Install/ComponentsPageViewModel.cs` — checkboxes from session components, forced-and-locked `permissions` (SecurityFirstHarnessID) with engine copy, writes `InstallSelection.CustomIds`; make tests pass; refactor.

## 5. Permissions + Review pages (test-first)

- [x] 5.1 RED: `PermissionsPageViewModelTests` — from `PermissionTierChoices`, balanceado (default) preselected; selecting bypass surfaces its warning text; writes `InstallSelection.Tier`.
- [x] 5.2 GREEN: `Pages/Install/PermissionsPageViewModel.cs` — radio group from `session.PermissionTierChoices`, default tier, warning binding.
- [x] 5.3 RED: `ReviewPageViewModelTests` — summary reflects agents/mode/components/tier; primary label is "Install"; confirming invokes the client install stream with the full selection (agents CSV, mode, custom ids, tier).
- [x] 5.4 GREEN: `Pages/Install/ReviewPageViewModel.cs` — read-only summary from `InstallSelection`; triggers install on confirm.

## 6. Progress page migration (move, tests-first)

- [x] 6.1 RED/setup: create `Pages/ProgressPageViewModel.cs` and copy `Host.Tests/ProgressPageViewModelTests.cs` from the progress assertions in `MainWindowViewModelTests.cs` (retarget to the new VM, keep assertions byte-identical).
- [x] 6.2 GREEN: MOVE `ApplySnapshot`/`CalculateProgress`/`ToFriendlyMessage`/`ToDisplayLabel` verbatim from `MainWindowViewModel.cs:314-425` into `ProgressPageViewModel`; consume a generic `IAsyncEnumerable<InstallProgressSnapshot>`; migrated tests pass unchanged.
- [x] 6.3 Add `Pages/CompletePageViewModel.cs` — success / degraded / rollback / error states surfaced from the terminal snapshot; minimal test for state mapping.

## 7. Engine client: carry the full install selection (additive)

- [x] 7.1 RED: extend the `IInstallerEngineClient` fake and add a `ProcessInstallerEngineClientTests` case — install invocation routes agents/mode/custom ids/tier through `InstallArguments.BuildInstallArgs`.
- [x] 7.2 GREEN: change `IInstallerEngineClient.RunInstallAsync` to accept the full selection (agents, mode, custom ids, tier); implement in `ProcessInstallerEngineClient` via `BuildInstallArgs` (already using `ArgumentList`). Update the fake used by page tests.

> Note: group 7 was implemented ahead of group 5 (out of file order) because `ReviewPageViewModel` (5.4) needs the extended `RunInstallAsync` signature to compile against. Sequencing recorded here; scope/behavior unchanged from the plan.

## 8. Hub page (test-first)

- [x] 8.1 RED: `HubPageViewModelTests` — Install enabled and starts the flow; Starters/Manage Backups/Uninstall/Update Stack present but disabled (with tooltip), do not navigate.
- [x] 8.2 GREEN: `Pages/HubPageViewModel.cs` per design D8 (disabled entries + tooltip, only Install actionable).

## 9. Shell view-model + WizardFlow wiring (test-first)

- [x] 9.1 RED: `ShellViewModelTests` — advancing/back uses `WizardFlow` to compute the next/previous page and instantiates the right page VM; `CanGoBack`, `PrimaryLabel` ("Next" vs "Install" on Review), and `PrimaryEnabled` (delegates to `CurrentPage.CanAdvance`) update correctly.
- [x] 9.2 GREEN: `Navigation/ShellViewModel.cs` — owns `CurrentPage`, footer state, and the shared `InstallSelection`; drives navigation via `WizardFlow`.

> Note: while implementing, found that reconstructing page VMs fresh on every Back/Next navigation (the simplest design, consistent with D9's "pages read/write the shared InstallSelection") would silently reset user choices to defaults when navigating Back then Forward again (e.g. Assistants would re-check every agent). Fixed by having Assistants/InstallType/Components/Permissions page VMs honor an already-populated `InstallSelection` field instead of always seeding from session defaults; added `InstallSelection.Mode` defaulting to `""` (not `"full"`) so "not yet chosen" is distinguishable from "user explicitly chose full". Covered by one new test per page VM.

## 10. Shell view + templates + Burn wiring

- [x] 10.1 Add `Themes/PageTemplates.xaml` — one implicit `DataTemplate` per page-VM type; merge into `App.xaml` resources.
- [x] 10.2 Rewrite `MainWindow.xaml` as the thin shell (header + `ContentControl Content={Binding CurrentPage}` + footer Back/Next|Install/Cancel); window 760×560, `ResizeMode=CanResize`, not maximized; long lists in `ScrollViewer`.
- [x] 10.3 Update `MainWindow.xaml.cs` and the Burn entry (`App.xaml.cs` / `ActiveStackBootstrapperApplication`) to construct `ShellViewModel` instead of `MainWindowViewModel` (Burn spawn/Process path unchanged).

> Deviation: `ActiveStackBootstrapperApplication`'s `DispatchStatus`/`DispatchFailure` previously called `MainWindowViewModel.ReportExternalProgress`/`FailExternalInstall` — dead-ish wiring already unused for real installs (the actual install always ran via `ProcessInstallerEngineClient`, not Burn's own apply). There is no single free-text status slot on the wizard shell (each page owns its state), so these now just trace via `BootstrapperTrace` instead of touching a view-model. `ShellViewModel.InstallSucceeded` (true on Complete in Success/Degraded state) replaces `MainWindowViewModel.InstallSucceeded` for the Burn exit-code decision. `MainWindowXaml_UsesOneWayBindingForProgressValue` (Host.Tests/BootstrapperLaunchOptionsTests.cs) was retargeted to `Themes/PageTemplates.xaml`, where the ProgressBar now lives. The old dashboard's one-shot `ACTIVE_STACK_AUTO_INSTALL` auto-run was reimplemented in `MainWindow.xaml.cs` (`RunAutoInstallAsync`) to drive the new `ShellViewModel` end to end instead of calling a single removed method, since there is no longer one "start install" call to preserve verbatim.

## 11. Cleanup

- [x] 11.1 Delete `MainWindowViewModel.cs` and remove the obsolete (already-migrated) tests from `MainWindowViewModelTests.cs`; delete the file if fully emptied.

  > All of `MainWindowViewModelTests.cs`'s tests either exercised behavior already migrated to `ProgressPageViewModelTests.cs` (D6) or exercised `MainWindowViewModel`-only mechanics (single-select `SelectedAssistantId`/`SelectedInstallTypeId`, the vestigial `ConfigureExternalInstallRunner` seam) that has no equivalent in the wizard. The file was fully obsolete, so it — and `MainWindowViewModel.cs` — were deleted outright rather than partially pruned.

- [x] 11.2 Run `dotnet test` on `Host.Tests` (and `Core.Tests` unaffected) — all green; verify no dangling references to the deleted dashboard VM.
- [x] 11.3 Verify each page-VM type resolves a `DataTemplate` (no blank pages) — small smoke assertion or manual launch note for C6.

  > Manual-review note (chosen over a coded smoke test — see rationale below): `Themes/PageTemplates.xaml` declares exactly one `DataTemplate DataType="{x:Type ...}"` for each of the 8 concrete page-VM classes `ShellViewModel` ever assigns to `CurrentPage` (`HubPageViewModel`, `AssistantsPageViewModel`, `InstallTypePageViewModel`, `ComponentsPageViewModel`, `PermissionsPageViewModel`, `ReviewPageViewModel`, `ProgressPageViewModel`, `CompletePageViewModel`), cross-checked 1:1 against `ShellViewModel.CreatePage`/`RunInstallAndAdvanceToCompleteAsync`. The dictionary is merged into `App.xaml`'s `Application.Resources`, and `MainWindow.xaml`'s `ContentControl` has no explicit `ContentTemplate`, so WPF's implicit per-type template lookup applies uniformly. A coded test was skipped because resolving `DataTemplate`s from a `ResourceDictionary` pack-URI requires a live STA `Application` — this test project has no STA-aware xunit host (no `Xunit.StaFact`) and none of the existing Host.Tests bootstrap one, so a hand-rolled STA test would be the first of its kind here and a real flake/hang risk for comparatively low signal over the static cross-check above. Recommended follow-up for C5/C6: if `Xunit.StaFact` (or similar) gets adopted for other WPF-surface tests, add a proper automated smoke test at that point.
