# Tasks — windows-gui-core-parsers

TDD applies (xUnit + `dotnet test`). For each behavior: write the failing test with a fixture copied byte-for-byte from the matching Go test JSON, then implement to green, then triangulate an edge case. Fixture sources are named per task. All work is C# only under `installer/windows/ui/`; no Go/`internal/*` edits.

## 1. Options parser: additive tier metadata (Core)

- [x] 1.1 Baseline: run existing `Core.Tests` (`dotnet test ActiveStack.Bootstrapper.Core.Tests`); capture "N passing". Existing `InstallerSessionStateBuilderTests` MUST stay green throughout (tier fields are additive).
- [x] 1.2 RED: add `InstallerSessionStateBuilderTests` cases for tier metadata — fixture mirrors `windows_test.go` `TestRunWindowsOptions_TierMetadata` (claude → `tier_capable:true`, `tier_capable_agents:["claude"]`, three `permission_tiers` with `balanceado` default and `bypass` warning).
- [x] 1.3 GREEN: add public record `PermissionTierChoice(string Id, string Label, string Description, bool IsDefault, string? Warning)`; extend `InstallerSessionState` with `TierCapable`, `TierCapableAgents`, `PermissionTierChoices`; add internal DTO `WindowsPermissionTierOption` + the three new `WindowsOptionsResponse` fields; parse them in `BuildFromJson`.
- [x] 1.4 TRIANGULATE: add a case with the tier fields ABSENT (pre-tier engine shape) asserting `TierCapable==false`, empty `TierCapableAgents`, empty `PermissionTierChoices`, and unchanged existing fields (absent-JSON safe defaults).
- [x] 1.5 REFACTOR: keep the `PropertyNameCaseInsensitive` shared options; confirm no duplicate deserialization. Re-run `Core.Tests` green.

## 2. Pure argument builders `InstallArguments` (Core)

- [x] 2.1 RED: `InstallArgumentsTests` for `BuildInstallArgs` — assert exact list `["windows","install","--agent","claude,codex","--mode","custom","--custom","a,b","--tier","balanceado"]`.
- [x] 2.2 GREEN: implement `InstallArguments.BuildInstallArgs(IEnumerable<string> agents, string mode, IEnumerable<string> customIds, string? tier)` returning `IReadOnlyList<string>`; join agents with `,`.
- [x] 2.3 TRIANGULATE `BuildInstallArgs`: (a) `mode:"full"`, empty customIds, `tier:null` → no `--custom`, no `--tier`; (b) `mode:"lite"` with non-empty customIds → still no `--custom` (custom only when `mode=="custom"`).
- [x] 2.4 RED+GREEN `BuildStarterInstallArgs(starterId, projectPath, agents, dryRun, yes)` — assert `--project` is followed by the single element `C:/My Projects/app` (space preserved), `--agent` is CSV, and `--dry-run`/`--yes` appear when flagged.
- [x] 2.5 RED+GREEN `BuildBackupActionArgs(action, id, description)` — `rename` with non-empty description includes `--description <d>` as one element; `delete`/`restore` and empty/null description omit `--description`.
- [x] 2.6 RED+GREEN `BuildUninstallArgs(agents, mode, strategy, backupId)` — includes `--strategy`; for `strategy=="restore"` includes the backup-selection flag. **O1 RESOLVED**: read `cmd/active-stack/headless/uninstall_flags.go` (`ParseUninstallFlags`) — the engine flag is `--restore-manifest <path>` (a manifest path, not a bare backup id). Builder param renamed `restoreManifestPath` accordingly; test asserts `--restore-manifest` followed by the manifest path.

## 3. Starter catalog parser (Core)

- [x] 3.1 RED: `StarterCatalogParserTests` — fixture from `windows_starters_test.go` `TestRunWindowsStartersList_CatalogOrder` (alpha includes beta, 2 harnesses, `mcp_count:2`; beta `mcp_count:1`).
- [x] 3.2 GREEN: implement `StarterCatalogParser.BuildFromJson(string)` → `IReadOnlyList<StarterChoice>` in catalog order; `StarterChoice(Id,Name,Description,Includes,Harnesses,McpCount)`; `Includes`/`Harnesses` never null.
- [x] 3.3 TRIANGULATE: empty `{"starters":[]}` (from `TestRunWindowsStartersList_EmptyCatalog`) → empty list, not null.

## 4. Backup list + action parsers (Core)

- [x] 4.1 RED: `BackupListParserTests` — fixture from `windows_backups_test.go` `TestRunWindowsBackupsList_ListsFabricatedManifests` (all D4 fields incl. `manifest_path`, `created_at` RFC3339).
- [x] 4.2 GREEN: implement `BackupListParser.BuildFromJson(string)` → `IReadOnlyList<BackupEntry>`; `BackupEntry(Id,CreatedAt,Source,Description,FileCount,Pinned,Compressed,DisplayLabel,ManifestPath)`; `CreatedAt` kept verbatim as string.
- [x] 4.3 TRIANGULATE: empty `{"backups":[]}` (from `TestRunWindowsBackupsList_EmptyStore`) → empty list, not null.
- [x] 4.4 RED+GREEN: `BackupActionResultParserTests` — fixtures from `windows_backups_action_test.go` (`{success:true,id:"backup-1"}` and unknown-id `{success:false,message,id}`); implement `BackupActionResultParser.BuildFromJson(string)` → `BackupActionResult(Success,Message,Id)`.

## 5. Uninstall options parser (Core)

- [x] 5.1 RED: `UninstallOptionsParserTests` — fixture from `windows_uninstall_options_test.go` (`detected_agents:["claude"]`, non-empty modes, strategies `targeted` default/`requires_manifest:false` + `restore` `requires_manifest:true`).
- [x] 5.2 GREEN: implement `UninstallOptionsParser.BuildFromJson(string)` → `UninstallOptions(DetectedAgents, Modes, Strategies)` with `UninstallStrategyChoice(Id,Label,Description,IsDefault,RequiresManifest)`; reuse `InstallTypeChoice` for modes; collections never null.
- [x] 5.3 Re-run `Core.Tests` fully green.

## 6. Engine client: ArgumentList migration + multi-agent session (Host)

- [x] 6.1 Baseline: run `Host.Tests` (`dotnet test ActiveStack.Bootstrapper.Host.Tests`); capture "N passing". Existing tests (tailer, `MainWindowViewModel`, launch options) MUST stay green.
- [x] 6.2 Migrate `ProcessInstallerEngineClient` to `ProcessStartInfo.ArgumentList`: shared private helper builds `ProcessStartInfo` and appends tokens via `ArgumentList.Add`; route all argv through `InstallArguments`. Convert `RunJsonCommandAsync` and `RunInstallAsync` (remove the string-concat at the old line 33).
- [x] 6.3 `LoadSessionAsync`: parse `detected_agents`, invoke `windows options --agent <csv-of-all-detected>` (comma-joined), fall back to `claude` when empty; delete `PickFirstAssistantId`. Added `ProcessInstallerEngineClient.ResolveDetectedAgentsCsv` (internal, pure) plus a `JsonCommandRunnerOverride` test seam; `Host.Tests` (`ProcessInstallerEngineClientTests`) verifies `LoadSessionAsync` invokes `windows options` with `--agent claude,codex` for two detected agents.

## 7. Engine client: extended contract methods (Host)

- [x] 7.1 Extend `IInstallerEngineClient` with `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, `RunUninstallAsync` (signatures per design D4).
- [x] 7.2 Implement the list/options/action methods on `ProcessInstallerEngineClient` (run one-shot JSON command via `ArgumentList`, delegate parsing to the Core parsers).
- [x] 7.3 Implement the streaming methods (`RunStarterInstallAsync`, `RunUninstallAsync`) reusing the existing line-reading loop (`RunStreamingCommandAsync`) + `InstallProgressSnapshotParser`; args from `InstallArguments`.
- [x] 7.4 Host.Tests where testable without live processes: updated `FakeInstallerEngineClient`/`FailingInstallerEngineClient` (in `MainWindowViewModelTests.cs`) to the new surface; added `JsonCommandRunnerOverride` test seam on `ProcessInstallerEngineClient` and `ProcessInstallerEngineClientTests` asserting `ListBackupsAsync`/`ListStartersAsync`/`LoadUninstallOptionsAsync`/`RunBackupActionAsync` return the Core-parsed results for fed JSON. Solution builds (net8.0/net8.0-windows) and all `Host.Tests` pass.

## 8. Verify

- [x] 8.1 `dotnet test` green for both `Core.Tests` and `Host.Tests`; existing tests unchanged. Final: Core.Tests 23/23, Host.Tests 20/20.
- [x] 8.2 Confirmed no Go / `internal/*` files were modified during this apply session (all Go diffs present in `git status` predate this change, from the already-implemented C1/C2 changes); no new NuGet dependencies added — both csproj files unchanged aside from `ProjectReference`/existing packages.
