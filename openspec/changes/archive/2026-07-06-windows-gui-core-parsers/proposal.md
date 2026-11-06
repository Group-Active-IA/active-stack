## Why

The Go engine now exposes a full `windows` JSON contract (changes `windows-contract-tier-multiagent` and `windows-contract-hub-operations`, both implemented and green): multi-agent options with tier metadata, starters list/install, backups list + actions, and uninstall-options. The WPF installer (`installer/windows/ui`) still only consumes the old two-command `detect` + `options` shape via a single parser and a string-concatenating process client. Before any wizard UI (C4/C5) can be built, the C# **Core** needs typed parsers for every new contract response and pure argv builders, and the **Host** engine client needs methods to invoke the new subcommands correctly (paths with spaces currently break the string-concat argument path). This change is the consumer-side foundation: it makes the Go contract usable from C# with tests, no UI yet.

## What Changes

- Extend `InstallerSessionStateBuilder` (`Core`) additively: parse `tier_capable`, `tier_capable_agents`, and `permission_tiers[]` from the `windows options` response into new `InstallerSessionState` fields (`TierCapable`, `TierCapableAgents`, `PermissionTierChoice[]`). Absent JSON fields resolve to safe defaults (compat with the pre-tier engine).
- Add new `Core` parsers in the existing `BuildFromJson` style: `StarterCatalogParser` (starters list), `BackupListParser` + `BackupActionResultParser` (backups list + action result), `UninstallOptionsParser` (uninstall-options).
- Add `InstallArguments` (`Core`): pure argv builders — `BuildInstallArgs`, `BuildStarterInstallArgs`, `BuildUninstallArgs`, `BuildBackupActionArgs` — returning `IReadOnlyList<string>` argument lists (no string concatenation), fully unit-testable without spawning processes.
- Migrate `ProcessInstallerEngineClient` (`Host`) from `ProcessStartInfo.Arguments` string concatenation to `ProcessStartInfo.ArgumentList`, so project paths and descriptions containing spaces no longer corrupt the command line.
- Extend the engine client with new methods: `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, `RunUninstallAsync`. `LoadSessionAsync` passes ALL detected agents as a CSV (removing `PickFirstAssistantId`, which collapsed detection to a single agent).
- Add the corresponding interface surface (`IInstallerEngineClient` extension or sibling interfaces — decided in design.md).
- Tests: `Core.Tests` parser + `InstallArguments` builder tests with JSON fixtures mirroring the Go test outputs byte-for-byte; `Host.Tests` for the client where testable without live processes.

No Go code, no `internal/*`, and no XAML/ViewModel/wizard UI change here.

## Capabilities

### New Capabilities
- `windows-gui-engine-client`: The C# consumer contract for the Go `windows` JSON engine — typed parsers for every `windows` response (options+tier, starters, backups, backup-action, uninstall-options), pure argv builders for every `windows` subcommand, and the async engine-client methods that invoke them (via `ArgumentList`, multi-agent CSV).

### Modified Capabilities
<!-- The Go-side windows-hub-contract spec is unchanged; this change only consumes it. No existing spec requirements change. -->

## Impact

- **Code (new/modified, C# only)**:
  - `installer/windows/ui/ActiveStack.Bootstrapper.Core/InstallerSessionStateBuilder.cs` (additive tier fields)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Core/StarterCatalogParser.cs` (new)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Core/BackupListParser.cs` + `BackupActionResultParser.cs` (new)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Core/UninstallOptionsParser.cs` (new)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Core/InstallArguments.cs` (new)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/ProcessInstallerEngineClient.cs` (ArgumentList migration + new methods)
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/IInstallerEngineClient.cs` (interface extension / siblings)
  - Test projects: `ActiveStack.Bootstrapper.Core.Tests`, `ActiveStack.Bootstrapper.Host.Tests`
- **Contract dependency**: consumes the Go `windows` JSON schemas defined by `cmd/active-stack/headless/windows*.go`. The C# parsers and fixtures MUST stay byte-for-byte synchronized with the Go test outputs (schema-sync rule).
- **Depends on**: C1 (`windows-contract-tier-multiagent`) + C2 (`windows-contract-hub-operations`), both implemented.
- **Unblocks**: C4 (wizard shell + Install flow) and C5 (Uninstall/Starters/Backups pages), which bind these parsers and client methods to ViewModels.
- **Governance**: MEDIO (`agents`/consumer adapters). No `internal/backup|filemerge|pipeline` touched.
