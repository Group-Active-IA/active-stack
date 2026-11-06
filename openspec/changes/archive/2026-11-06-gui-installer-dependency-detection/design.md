## Context

Confirmed by reading the code:

- **Go side**: `RunWindowsDetect` (`cmd/active-stack/headless/windows.go:71-74`) encodes `windowsDetectResponse{DetectedAgents: []string}` — nothing else. `system.DetectDependencies(ctx, profile)` (`internal/system/deps.go`) already returns the full report; it's currently only consumed by the TUI (`gate.go`, and after `tui-dependency-detection`, also `viewDetection()`).
- **.NET side**: `InstallerSessionStateBuilder.BuildFromJson` (`installer/windows/ui/ActiveStack.Bootstrapper.Core/InstallerSessionStateBuilder.cs`) deserializes an internal `WindowsDetectResponse` (mirrors the Go struct 1:1 via `System.Text.Json`) and an internal `WindowsOptionsResponse`, then builds the immutable `InstallerSessionState` record consumed by view-models. This is the exact seam to extend.
- **Wizard navigation**: `WizardPageId` enum (`Navigation/WizardPageId.cs`) lists every page; `WizardFlow.cs` expresses `NextPage`/`PreviousPage` as pure switch expressions. Confirmed: `Hub → Assistants` is the current first hop of the Install flow (`WizardFlow.cs:20-21` forward, `:43-44` backward) — the direct analog of the TUI's `Welcome → Detection → Agents`.
- **Existing spec coverage**: `openspec/specs/windows-gui-wizard/spec.md` already has a requirement "Install flow ordering mirrors the TUI" stating the order as `Hub → Assistants → InstallType → ...` — inserting a page here is a genuine change to that requirement's stated behavior, not just an addition, so `specs/windows-gui-wizard/spec.md` in this change needs a `## MODIFIED Requirements` block (full requirement text, per the schema's MODIFIED workflow), not just ADDED. `openspec/specs/windows-hub-contract/spec.md` has no existing requirement for the `windows detect` command's own contract (it's referenced by other requirements but never itself specified) — the `dependencies` field addition there is a fresh `## ADDED Requirements` entry.
- Pages follow one consistent shape confirmed across the existing spec (`AssistantsPageViewModel`, etc.): a page view-model implementing `IWizardPage`-family behavior, a `DataTemplate` entry in `Themes/PageTemplates.xaml`, pre-checked/default state derived from session data, and a detail panel pattern (long-description-falls-back-to-description) used by every selection-style page in this codebase.

## Goals / Non-Goals

**Goals:**
- Give the graphical installer the same upfront visibility the TUI now has (after `tui-dependency-detection`): dependency status before Assistants, install hints and copyable commands for anything missing.
- Reuse `system.DetectDependencies` as the sole source of truth — no dependency-detection logic duplicated in C#.
- Match this codebase's established page pattern (view-model + `DataTemplate` + session-derived state) rather than inventing a new one.

**Non-Goals:**
- No auto-install button, no `Process.Start` of winget/apt/brew from the .NET side — same "show it, don't run it" stance as `tui-dependency-detection`.
- No change to `system.DetectDependencies`/`defineDependencies` — Go stays `Required: false`, unchanged from the other two changes in this series.
- No change to the Install flow steps after Assistants (InstallType, Components, Permissions, Review, ...) — only the Hub→Assistants hop gains a stop.

## Decisions

- **JSON contract shape**: extend `windowsDetectResponse` with `Dependencies system.DependencyReport `json:"dependencies"`` (Go) so the wire format is `{"detected_agents":[...], "dependencies": {"dependencies":[...], "all_present":bool, "missing_required":[...], "missing_optional":[...]}}` — additive field, existing consumers of `detected_agents` alone are unaffected (confirmed: no other Go or .NET code reads `windowsDetectResponse` besides this one struct/response pair).
- **Where the install command text comes from**: reuse `system.InstallCommandsForDep(name, profile)` server-side (Go), pre-joined into a single display string per dependency (e.g. `"winget install --id Git.Git -e --accept-source-agreements --accept-package-agreements"`), rather than shipping the raw `[][]string` and joining in C#. Rationale: keeps "how a command renders as text" a single Go-owned concern instead of duplicating argv-joining logic in two languages; the .NET side only ever needs display strings, never argv it could accidentally execute.
- **Page placement**: `WizardPageId.Dependencies` inserted between `Hub` and `Assistants`. `WizardFlow.NextPage`: `Hub => Dependencies`, `Dependencies => Assistants`. `PreviousPage`: `Assistants => Dependencies`, `Dependencies => Hub`. Only the Install flow gains this stop (Uninstall/Starters/Backups flows are unaffected — they don't install harnesses that need git/node).
- **Non-blocking navigation**: mirrors `tui-dependency-detection`'s decision — `DependenciesPageViewModel.CanAdvance` is always `true` regardless of missing required deps. The actual install still runs through the existing engine pipeline, which will fail loudly if a required tool truly is absent (unchanged failure path); this page's job is upfront visibility, not gating.
- **Rendering pattern**: follow the existing per-row list pattern (like Components/Backups pages) — one row per dependency showing name, required/optional badge, status, and for missing ones an expandable install-hint + copyable command block (`TextBox IsReadOnly="True"` with select-all-on-focus, the standard WPF "copyable text" idiom — no custom clipboard button needed for v1).

## Risks / Trade-offs

- [Risk] Two toolchains (Go JSON contract + .NET parser/view-model/XAML) must stay in lockstep or the page silently shows nothing → Mitigation: task list requires a Core-side parser test asserting the new `Dependencies` field round-trips from a fixed JSON fixture, mirroring the existing `BackupListParserTests.cs`/`UninstallOptionsParserTests.cs` pattern in `ActiveStack.Bootstrapper.Core.Tests`.
- [Risk] Modifying "Install flow ordering mirrors the TUI" touches a requirement with 4 existing scenarios covering Components/Permissions skip logic unrelated to this change → Mitigation: only the order-statement sentence and the Hub/Assistants scenario need updating; the Components/Permissions/tier-capability scenarios are copied verbatim into the MODIFIED block, unedited.
- [Risk] This change requires a .NET 8 SDK + WiX toolchain to build/verify, which isn't part of the Go `go test ./...` loop → Mitigation: task list calls out running the `.Tests` projects (`dotnet test`) separately, matching how `installer/windows/README.md` already documents the two-toolchain build.
