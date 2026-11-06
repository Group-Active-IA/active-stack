## Context

The Windows GUI installer is a WPF app (`installer/windows/ui/ActiveStack.Bootstrapper.Host`) that runs as a WiX Burn bootstrapper UI and drives the Go engine (`active-stack.exe windows <subcmd>`) by spawning a child process and parsing its JSON. Today the whole UI is one screen: `MainWindow.xaml` bound to a single god-object `MainWindowViewModel`. It is single-agent (`ComboBox`), Custom mode is read-only (never sends `--custom`), and there is no permission-tier screen.

C1–C3 already landed the substrate this change consumes:
- **C1 (Go engine)**: `--tier`, multi-agent `windows options`, and additive `options` metadata (`tier_capable`, `tier_capable_agents`, `permission_tiers[]`).
- **C3 (C# Core)**: `InstallerSessionState` now carries `TierCapable`/`TierCapableAgents`/`PermissionTierChoices`; parsers (`StarterCatalogParser`, `BackupListParser`, `UninstallOptionsParser`); pure `InstallArguments` builders; the extended `IInstallerEngineClient` with `ProcessStartInfo.ArgumentList`. All landed and green — see `openspec/changes/windows-gui-core-parsers/specs/windows-gui-engine-client/spec.md`.

C4 is the user-facing materialization: it turns the dashboard into a wizard that mirrors the TUI (`internal/tui/model.go:278-313`) and the approved 8-screen design (`docs/superpowers/specs/2026-07-02-windows-native-installer-design.md:226-346`), with documented deviations. C5 later wires Starters/Backups/Uninstall; C4 leaves those Hub entries present-but-disabled.

**Constraints**: no changes to `internal/install|backup|uninstall`, the Go engine, Burn wiring, `build.ps1`, or `Bundle.wxs`. Governance: this is UI (`Host` project) — MEDIUM autonomy, implement with checkpoints. TDD where `dotnet test` allows (WizardFlow and page VMs are plain MVVM, fully testable without a UI thread).

## Goals / Non-Goals

**Goals:**
- A traditional multi-page wizard (next → next → finish) implementing the full **Install** flow with exact TUI parity.
- Navigation as pure, unit-testable functions — the flow graph is data, not UI state.
- Migrate progress rendering from the old VM without behavioral change (its tests survive).
- Hub surfaces all TUI destinations; C4 delivers Install end-to-end and stubs the rest.

**Non-Goals:**
- Implementing Uninstall / Starters / Backups flows (C5).
- Implementing Update Stack (permanent TUI parity: "coming soon", disabled).
- Any change to the Go engine or the C3 Core/client contract (consume as-is; the one additive exception — extending `RunInstallAsync` to carry the full selection — is called out in Decisions).
- Visual theming beyond a clean functional wizard; no marketing chrome.
- Cosmetic re-copy of the approved design's commercial wording beyond what parity needs.

## Decisions

### D1 — Navigation: `ContentControl` + `DataTemplate`-by-VM-type, not `Frame`/`NavigationService`
The shell hosts `<ContentControl Content="{Binding CurrentPage}"/>`. `Themes/PageTemplates.xaml` declares one implicit `DataTemplate` per page-VM type (`DataType={x:Type ...PageViewModel}`), so setting `CurrentPage` to a VM instance swaps the view. **Why over Frame/NavigationService**: WPF's `NavigationService` keeps a *chronological* journal (Back = previous URI visited). The wizard Back must be *logical* — from Review, Back must return to Permissions (if it was shown) else Components (if custom) else InstallType, skipping pages not in the current path. A journal cannot express "skip pages not in path" cleanly. Modeling the flow as data (D2) makes Back deterministic and testable without a UI.

### D2 — `WizardFlow` as pure `NextPage`/`PreviousPage` functions
`Navigation/WizardFlow` exposes static/pure functions: given the current page identity + the current `InstallSelection` (agents, mode, custom ids, tier) + session state (`TierCapableAgents`), return the next (or previous) page identity. No I/O, no mutation of shell state. `ShellViewModel` calls these to compute the next VM to instantiate. This mirrors the TUI transitions in `internal/tui/model.go:278-318`:
- After InstallType: if mode == custom → Components; else if any selected agent is tier-capable → Permissions; else → Review.
- After Components: if any selected agent is tier-capable → Permissions; else → Review.
- After Permissions → Review.
- Back is the inverse over the *same* predicates, so pages not in the path are skipped (e.g. Back from Review with a gemini-only full install returns to InstallType, skipping Permissions and Components).

Tier-capability uses `session.TierCapableAgents` intersected with the *selected* agents — no re-query to the engine on selection change (the plan's explicit rule).

### D3 — Shell as thin host; `ShellViewModel` owns page lifecycle + footer state
`MainWindow.xaml` = header (`{Binding CurrentPage.Title}` / `.Subtitle`) + `ContentControl` + footer (`Back`, primary `Next`/`Install`, `Cancel`). Window 760×560, `ResizeMode=CanResize`, not maximized. `ShellViewModel` holds `CurrentPage` (an `IWizardPage`), `CanGoBack`, `PrimaryLabel` ("Next" vs "Install" on Review), `PrimaryEnabled` (delegates to `CurrentPage.CanAdvance`, e.g. Assistants min-1 gate), and a shared `InstallSelection` model threaded through pages. Each page VM implements `IWizardPage` (Title, Subtitle, `CanAdvance`, and hooks to read/write the shared selection).

### D4 — `IWizardPage` contract and page map
| Page VM | Role | Advance gate |
|---|---|---|
| `HubPageViewModel` | Entry menu; selecting "Install" starts the Install flow | one entry chosen |
| `AssistantsPageViewModel` | multi-select checkboxes; prechecked from `session.DetectedAssistants`; recommended marked | ≥1 agent selected |
| `InstallTypePageViewModel` | cards lite/full/custom (end-user labels Quick/Complete/Custom), default full | one selected (always true) |
| `ComponentsPageViewModel` | custom only; checkboxes from options components; `permissions` checked+**disabled** with engine copy; recommended prechecked | always (permissions guarantees ≥1) |
| `PermissionsPageViewModel` | 3-tier radio from `session.PermissionTierChoices`; default balanceado; bypass shows warning | one tier (default preselected) |
| `ReviewPageViewModel` | read-only summary of agents/mode/components/tier; primary button reads "Install" | always |
| `ProgressPageViewModel` | consumes `RunInstallAsync` stream; no footer nav | n/a (terminal-driven) |
| `CompletePageViewModel` | success / degraded / rollback / error states; offers close/next-steps | n/a |

`ProgressPageViewModel` is **generic** — it renders any `IAsyncEnumerable<InstallProgressSnapshot>` so C5 can reuse it for starter/uninstall streams.

### D5 — Forced-and-locked permissions harness in Components
The `permissions` harness (`install.SecurityFirstHarnessID`) is rendered checked and **disabled** (cannot be unchecked), with the engine-provided copy ("Basic protection … always on"). This mirrors `internal/tui/model.go:307-311` where the space toggle is ignored for that id. The item is always included in `custom` ids sent to the engine.

### D6 — Move (not rewrite) progress logic into `ProgressPageViewModel`
`ApplySnapshot`, `CalculateProgress`, `ToFriendlyMessage`, `ToDisplayLabel` (`MainWindowViewModel.cs:314-425`) move verbatim into `ProgressPageViewModel`. Their tests move from `MainWindowViewModelTests.cs` into `ProgressPageViewModelTests.cs` **first** (rename/retarget, keep assertions), stay green, then `MainWindowViewModel` and its dead tests are deleted. This is refactor-by-migration: behavior is pinned by the moved tests.

### D7 — Extend `RunInstallAsync` to carry the full selection
C3 shipped `RunInstallAsync(assistantId, installTypeId)` — single agent, no custom, no tier — but also shipped `InstallArguments.BuildInstallArgs(agents, mode, customIds, tier)`. C4 changes `RunInstallAsync` to accept the full selection (agents list, mode, custom ids, tier) and route through `BuildInstallArgs`. This is **additive** to the engine contract (the flags already exist from C1) and is the minimal client wiring the wizard needs. `ProcessInstallerEngineClient` already uses `ArgumentList`; only the method signature + arg construction change. The fake client in tests mirrors the new signature.

### D8 — Hub placeholder decision: **disabled entries with tooltip** (not placeholder pages)
Install / Starters / Manage Backups / Uninstall / "Update Stack — Coming soon" all appear on the Hub. In C4 **only Install is enabled**; Starters, Manage Backups, and Uninstall are rendered **disabled with an explanatory tooltip** ("Available in the next update"), exactly like Update Stack. **Why over placeholder pages**: a disabled entry is the cleanest, lowest-surface option — no dead navigation, nothing for a user to get stuck in, no throwaway XAML/VM to delete in C5. C5 flips `IsEnabled` and wires each entry's target VM. This keeps the Hub's visual completeness (TUI parity: the user sees every destination exists) without shipping non-functional flows. Recorded here as the binding C4 decision; C5 references it when enabling the entries.

### D9 — Selection model threaded through pages
A single mutable `InstallSelection` (agents, mode, custom ids, tier) is created when the Install flow starts and shared across the Install page VMs and the flow functions. Pages read/write it; `WizardFlow` reads it (never mutates). Review renders it; Progress passes it to `RunInstallAsync`. This keeps each page VM free of cross-page state and makes flow decisions a pure function of `(currentPage, selection, session)`.

## Risks / Trade-offs

- **[Implicit `DataTemplate` DataType resolution fails silently → blank page]** → keep templates in one merged `PageTemplates.xaml` referenced by `App.xaml`; add a smoke test/manual check that each page-VM type resolves a view; keep VMs in a stable namespace.
- **[Migrating progress logic could silently change behavior]** → D6 moves the *tests first* and keeps assertions byte-identical; the migration is done only when the moved tests pass unchanged.
- **[`RunInstallAsync` signature change (D7) ripples to existing callers/tests]** → the only caller is the old dashboard VM (being deleted) and its fake; update both in the same step. Additive on the engine side (no Go change).
- **[Back-skip logic drift vs TUI]** → `WizardFlowTests` encode the exact TUI transition table as pure-function assertions (full+claude → Permissions appears; custom → Components appears; gemini-only → Permissions skipped; Back from Review skips out-of-path pages), so drift fails a test.
- **[Window resize breaks fixed layouts]** → use `Grid`/`DockPanel` with proportional rows and `ScrollViewer` on long component lists; verify at min size in the C6 smoke.
- **[Disabled Hub entries look like bugs]** → tooltip copy ("Available in the next update") + consistent styling with the already-accepted "Update Stack — Coming soon" entry.

## Migration Plan

1. Land the new `Navigation/`, `Pages/`, `Themes/PageTemplates.xaml` alongside the existing `MainWindow`/`MainWindowViewModel` (both compile).
2. Migrate `ProgressPageViewModelTests` from `MainWindowViewModelTests` (D6) — green before deletion.
3. Rewrite `MainWindow.xaml`/`.xaml.cs` to the shell; wire `ShellViewModel` in `App.xaml.cs`/`ActiveStackBootstrapperApplication` (Burn entry point unchanged — it just constructs the shell instead of the dashboard).
4. Delete `MainWindowViewModel.cs` and its obsolete tests once nothing references them.
5. **Rollback**: revert the shell wiring commit; the old dashboard VM is recoverable from history until step 4 lands, and all changes are confined to the `Host` project (no engine/installer state touched).

## Open Questions

- None blocking. Copy strings for tier warnings and the permissions "always on" note come from the engine's `options` response (`permission_tiers[].warning`, component description), so no UI-authored copy needs sign-off. The disabled-entry tooltip text ("Available in the next update") is a C4 default and can be re-worded in C5 without contract impact.
