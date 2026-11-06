## ADDED Requirements

### Requirement: Options parser exposes additive tier metadata

`InstallerSessionStateBuilder.BuildFromJson` SHALL parse the additive tier fields of the `windows options` response — `tier_capable` (bool), `tier_capable_agents` (array of agent id strings), and `permission_tiers` (array of `{id,label,description,default,warning?}`) — into new `InstallerSessionState` fields (`TierCapable`, `TierCapableAgents`, and `PermissionTierChoices` of type `PermissionTierChoice`). When any of these fields is absent from the JSON, the parser MUST resolve it to a safe default: `TierCapable=false`, `TierCapableAgents=[]`, `PermissionTierChoices=[]`. The existing detect/options parsing behavior MUST be unchanged.

#### Scenario: Tier metadata is parsed when present

- **WHEN** `BuildFromJson` receives an `options` JSON with `tier_capable: true`, `tier_capable_agents: ["claude"]`, and three `permission_tiers` where `balanceado` has `default: true` and `bypass` carries a `warning`
- **THEN** the resulting state has `TierCapable == true`, `TierCapableAgents == ["claude"]`, three `PermissionTierChoices` in the given order, the `balanceado` choice `IsDefault == true`, and the `bypass` choice's `Warning` populated

#### Scenario: Absent tier fields default safely

- **WHEN** `BuildFromJson` receives an `options` JSON that omits `tier_capable`, `tier_capable_agents`, and `permission_tiers` (the pre-tier engine shape)
- **THEN** the resulting state has `TierCapable == false`, an empty `TierCapableAgents`, and an empty `PermissionTierChoices`, and all existing fields parse as before

### Requirement: Starter catalog parser

A `StarterCatalogParser.BuildFromJson` SHALL parse the `windows starters list` response `{"starters":[{id,name,description,includes,harnesses,mcp_count}]}` into a list of `StarterChoice` results, preserving catalog order. `includes` and `harnesses` MUST always be non-null lists (empty when absent). An empty or `{"starters":[]}` response MUST yield an empty list, never null.

#### Scenario: Starters are parsed in order with resolved fields

- **WHEN** the parser receives a response with two starters `alpha` (includes `beta`, two harnesses, `mcp_count: 2`) and `beta`
- **THEN** it returns two `StarterChoice` entries in the order `[alpha, beta]`, with `alpha.Includes == ["beta"]`, `alpha.Harnesses` of length 2, and `alpha.McpCount == 2`

#### Scenario: Empty starters list yields an empty result

- **WHEN** the parser receives `{"starters":[]}`
- **THEN** it returns an empty list (not null)

### Requirement: Backup list parser

A `BackupListParser.BuildFromJson` SHALL parse the `windows backups list` response `{"backups":[...]}` into a list of `BackupEntry` results carrying `id`, `created_at`, `source`, `description`, `file_count`, `pinned`, `compressed`, `display_label`, and `manifest_path`. `created_at` MUST be preserved verbatim as the engine's RFC3339 string. An empty store response `{"backups":[]}` MUST yield an empty list, never null.

#### Scenario: Backups are parsed with all fields

- **WHEN** the parser receives a `backups` response with an entry `{id:"backup-1", source:"install", description:"before upgrade", file_count:3, pinned:true, compressed:false, display_label:"...", manifest_path:"C:/.../manifest.json", created_at:"2026-03-15T10:30:00Z"}`
- **THEN** the returned `BackupEntry` carries each field unchanged, with `CreatedAt == "2026-03-15T10:30:00Z"` and `ManifestPath` equal to the given path

#### Scenario: Empty backup store yields an empty result

- **WHEN** the parser receives `{"backups":[]}`
- **THEN** it returns an empty list (not null)

### Requirement: Backup action result parser

A `BackupActionResultParser.BuildFromJson` SHALL parse the `windows backups restore|delete|rename` response `{success,message,id}` into a `BackupActionResult` result. A failure response (`success:false`) MUST be represented as `BackupActionResult` with `Success == false`, the message preserved, and the requested id echoed.

#### Scenario: Successful action is parsed

- **WHEN** the parser receives `{"success":true,"message":"...","id":"backup-1"}`
- **THEN** it returns a result with `Success == true`, `Id == "backup-1"`, and the message preserved

#### Scenario: Failure action preserves message and id

- **WHEN** the parser receives `{"success":false,"message":"backup not found","id":"does-not-exist"}`
- **THEN** it returns a result with `Success == false`, `Id == "does-not-exist"`, and `Message == "backup not found"`

### Requirement: Uninstall options parser

An `UninstallOptionsParser.BuildFromJson` SHALL parse the `windows uninstall-options` response into an `UninstallOptions` result with `detected_agents`, `modes` (as install-type choices), and `strategies` (as strategy choices carrying `id`, `label`, `description`, `default`, `requires_manifest`). Collections MUST never be null.

#### Scenario: Uninstall options are parsed with strategies

- **WHEN** the parser receives a response with `detected_agents:["claude"]`, non-empty `modes`, and strategies including `targeted` (`default:true`, `requires_manifest:false`) and `restore` (`requires_manifest:true`)
- **THEN** the result carries `DetectedAgents == ["claude"]`, the modes list, and two strategy choices where `targeted.IsDefault == true`/`targeted.RequiresManifest == false` and `restore.RequiresManifest == true`

### Requirement: Pure argument builders for windows subcommands

An `InstallArguments` type SHALL provide pure, side-effect-free builders that return an ordered argument list (`IReadOnlyList<string>`) for each `windows` subcommand, suitable for `ProcessStartInfo.ArgumentList`. Each path or free-text value MUST be a single, unquoted list element so values containing spaces are preserved. `BuildInstallArgs` MUST join agents with commas into a single `--agent` value; MUST omit `--custom` when the custom id set is empty OR the mode is not `custom`; and MUST omit `--tier` when the tier is null or empty. `BuildBackupActionArgs` MUST include `--description` only for the `rename` action with a non-empty description.

#### Scenario: Install args with custom components and tier

- **WHEN** `BuildInstallArgs(agents:["claude","codex"], mode:"custom", customIds:["a","b"], tier:"balanceado")` is called
- **THEN** it returns exactly `["windows","install","--agent","claude,codex","--mode","custom","--custom","a,b","--tier","balanceado"]`

#### Scenario: Install args omit custom and tier when not applicable

- **WHEN** `BuildInstallArgs(agents:["claude"], mode:"full", customIds:[], tier:null)` is called
- **THEN** the returned list contains `--agent claude` and `--mode full` but contains neither `--custom` nor `--tier`

#### Scenario: Custom ids are ignored when mode is not custom

- **WHEN** `BuildInstallArgs(agents:["claude"], mode:"lite", customIds:["a","b"], tier:null)` is called
- **THEN** the returned list does not contain `--custom`

#### Scenario: Starter install path with spaces is a single argument

- **WHEN** `BuildStarterInstallArgs(starterId:"alpha", projectPath:"C:/My Projects/app", agents:["claude","opencode"], dryRun:true, yes:true)` is called
- **THEN** the returned list contains the elements `--project` followed by the single element `C:/My Projects/app`, `--agent` followed by `claude,opencode`, and includes `--dry-run` and `--yes`

#### Scenario: Backup rename includes description, other actions do not

- **WHEN** `BuildBackupActionArgs(action:"rename", id:"backup-1", description:"before upgrade")` is called
- **THEN** the list contains `--description` followed by the single element `before upgrade`
- **WHEN** `BuildBackupActionArgs(action:"delete", id:"backup-1", description:null)` is called
- **THEN** the list contains `--id backup-1` and does not contain `--description`

### Requirement: Engine client uses ArgumentList and passes all detected agents

`ProcessInstallerEngineClient` SHALL construct every child-process invocation via `ProcessStartInfo.ArgumentList` (one token per element) rather than a concatenated argument string, routing all argument construction through `InstallArguments` where a builder exists. `LoadSessionAsync` MUST pass ALL detected agents to `windows options` as a comma-separated `--agent` value (the `PickFirstAssistantId` single-agent collapse is removed), falling back to `claude` only when no agents are detected.

#### Scenario: Session load queries options with all detected agents

- **WHEN** `LoadSessionAsync` runs against an engine whose `windows detect` reports `["claude","codex"]`
- **THEN** the client invokes `windows options` with `--agent` equal to `claude,codex` (all detected agents), and builds the session state from that response

#### Scenario: Arguments carry spaces intact

- **WHEN** the client runs a starter install whose project path contains a space
- **THEN** the path is passed as a single `ArgumentList` element and reaches the engine uncorrupted (no split on the space)

### Requirement: Engine client methods for the extended contract

`IInstallerEngineClient` SHALL be extended with async methods that invoke the new `windows` subcommands: `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, and `RunUninstallAsync`. List/options/action methods MUST return the corresponding parsed result types; the install/uninstall streaming methods MUST stream `InstallProgressSnapshot` values parsed from the JSONL event stream, consistent with `RunInstallAsync`.

#### Scenario: Interface exposes the extended surface

- **WHEN** the `IInstallerEngineClient` contract is inspected
- **THEN** it declares `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, and `RunUninstallAsync`, and `ProcessInstallerEngineClient` implements all of them

#### Scenario: List backups parses the engine response

- **WHEN** `ListBackupsAsync` is invoked against an engine that returns a non-empty `windows backups list` response
- **THEN** it returns the parsed `BackupEntry` list produced by `BackupListParser`
