## Purpose

Define the parsers that turn `active-stack windows ...` engine JSON responses (options, starters, backups, streams) into the typed session state and choice models consumed by the Windows GUI, with safe defaults when fields are absent.

## Requirements

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

An `InstallArguments` type SHALL provide pure, side-effect-free builders that return an ordered argument list (`IReadOnlyList<string>`) for each `windows` subcommand, suitable for `ProcessStartInfo.ArgumentList`. Each path or free-text value MUST be a single, unquoted list element so values containing spaces are preserved. `BuildInstallArgs` MUST join agents with commas into a single `--agent` value; MUST omit `--custom` when the custom id set is empty OR the mode is not `custom`; and MUST omit `--tier` when the tier is null or empty. `BuildBackupActionArgs` MUST include `--description` only for the `rename` action with a non-empty description. Every builder for a subcommand that emits user-facing text (options, starters list/install, backups list, uninstall-options, install, uninstall) SHALL accept a `lang` parameter and append `--lang <lang>` to the argument list; the `detect` subcommand SHALL NOT take `--lang`. The `--lang` value is one of `es|en` and defaults to `en`.

#### Scenario: Install args with custom components and tier

- **WHEN** `BuildInstallArgs(agents:["claude","codex"], mode:"custom", customIds:["a","b"], tier:"balanceado", lang:"en")` is called
- **THEN** it returns `["windows","install","--agent","claude,codex","--mode","custom","--custom","a,b","--tier","balanceado","--lang","en"]`

#### Scenario: Install args omit custom and tier when not applicable

- **WHEN** `BuildInstallArgs(agents:["claude"], mode:"full", customIds:[], tier:null, lang:"es")` is called
- **THEN** the returned list contains `--agent claude` and `--mode full` and `--lang es`, but contains neither `--custom` nor `--tier`

#### Scenario: Custom ids are ignored when mode is not custom

- **WHEN** `BuildInstallArgs(agents:["claude"], mode:"lite", customIds:["a","b"], tier:null, lang:"en")` is called
- **THEN** the returned list does not contain `--custom`

#### Scenario: Starter install path with spaces is a single argument

- **WHEN** `BuildStarterInstallArgs(starterId:"alpha", projectPath:"C:/My Projects/app", agents:["claude","opencode"], dryRun:true, yes:true, lang:"es")` is called
- **THEN** the returned list contains the elements `--project` followed by the single element `C:/My Projects/app`, `--agent` followed by `claude,opencode`, `--lang es`, and includes `--dry-run` and `--yes`

#### Scenario: Backup rename includes description, other actions do not

- **WHEN** `BuildBackupActionArgs(action:"rename", id:"backup-1", description:"before upgrade")` is called
- **THEN** the list contains `--description` followed by the single element `before upgrade`
- **WHEN** `BuildBackupActionArgs(action:"delete", id:"backup-1", description:null)` is called
- **THEN** the list contains `--id backup-1` and does not contain `--description`

### Requirement: Engine client uses ArgumentList and passes all detected agents

`ProcessInstallerEngineClient` SHALL construct every child-process invocation via `ProcessStartInfo.ArgumentList` (one token per element) rather than a concatenated argument string, routing all argument construction through `InstallArguments` where a builder exists. `LoadSessionAsync` MUST pass ALL detected agents to `windows options` as a comma-separated `--agent` value (the `PickFirstAssistantId` single-agent collapse is removed), falling back to `claude` only when no agents are detected. The client SHALL expose a mutable `Language` property (default `"en"`), and SHALL thread that language as `--lang <Language>` into every text-emitting subcommand invocation — `windows options`, starters list/install, backups list, uninstall-options, install, and uninstall — while the `windows detect` invocation MUST NOT carry `--lang`.

#### Scenario: Session load queries options with all detected agents

- **WHEN** `LoadSessionAsync` runs against an engine whose `windows detect` reports `["claude","codex"]`
- **THEN** the client invokes `windows options` with `--agent` equal to `claude,codex` (all detected agents), and builds the session state from that response

#### Scenario: Arguments carry spaces intact

- **WHEN** the client runs a starter install whose project path contains a space
- **THEN** the path is passed as a single `ArgumentList` element and reaches the engine uncorrupted (no split on the space)

#### Scenario: The active language is threaded to text-emitting subcommands

- **WHEN** the client's `Language` is set to `"es"` and it invokes `windows options` (or starters list/install, backups list, uninstall-options, install, or uninstall)
- **THEN** that invocation includes `--lang es`, while a `windows detect` invocation carries no `--lang` argument

### Requirement: Engine client methods for the extended contract

`IInstallerEngineClient` SHALL be extended with async methods that invoke the new `windows` subcommands: `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, and `RunUninstallAsync`. List/options/action methods MUST return the corresponding parsed result types; the install/uninstall streaming methods MUST stream `InstallProgressSnapshot` values parsed from the JSONL event stream, consistent with `RunInstallAsync`.

#### Scenario: Interface exposes the extended surface

- **WHEN** the `IInstallerEngineClient` contract is inspected
- **THEN** it declares `ListStartersAsync`, `RunStarterInstallAsync`, `ListBackupsAsync`, `RunBackupActionAsync`, `LoadUninstallOptionsAsync`, and `RunUninstallAsync`, and `ProcessInstallerEngineClient` implements all of them

#### Scenario: List backups parses the engine response

- **WHEN** `ListBackupsAsync` is invoked against an engine that returns a non-empty `windows backups list` response
- **THEN** it returns the parsed `BackupEntry` list produced by `BackupListParser`

### Requirement: Engine client exposes a settable language defaulting to English

`IInstallerEngineClient` SHALL declare a `Language { get; set; }` property whose default value is `"en"`. Setting it SHALL change the language threaded into subsequent text-emitting subcommand invocations without otherwise altering the contract, so the shell can switch languages between flows.

#### Scenario: Interface exposes a settable Language defaulting to en

- **WHEN** an `IInstallerEngineClient` is inspected before any language is chosen
- **THEN** its `Language` reads `"en"`, and after `Language` is set to `"es"` subsequent text-emitting invocations use `es`

### Requirement: Engine client isolates the GUI from child-process decoding and stream faults

`ProcessInstallerEngineClient` SHALL decode the engine child process's standard output and standard error as UTF-8 by setting `StandardOutputEncoding` and `StandardErrorEncoding` to UTF-8 on the `ProcessStartInfo` used to launch the engine, so accented and non-ASCII text (e.g. Spanish copy) is read without corruption. The progress-stream parser (`InstallProgressSnapshotParser.Parse`) SHALL tolerate a stdout line that is not valid JSON by treating it as an ignorable/log-type snapshot rather than throwing an uncaught exception, so a single malformed line cannot abort the whole stream. This tolerance applies to individual malformed lines only: a genuine non-zero engine exit MUST still be surfaced as a thrown failure (so the shell can present it as a terminal Error state), never silently swallowed.

#### Scenario: Child-process stdout and stderr are decoded as UTF-8

- **WHEN** the client builds the `ProcessStartInfo` to launch the engine
- **THEN** both `StandardOutputEncoding` and `StandardErrorEncoding` are UTF-8

#### Scenario: A non-JSON stdout line does not throw

- **WHEN** `InstallProgressSnapshotParser.Parse` receives a stdout line that is not valid JSON
- **THEN** it returns a snapshot (an ignorable/log-type entry carrying the raw line) instead of throwing an uncaught `JsonException`

#### Scenario: A non-zero engine exit is still surfaced

- **WHEN** the engine process exits with a non-zero code
- **THEN** the streaming/one-shot client still raises a failure (it is not swallowed), so the caller can present it as an error rather than the failure passing unnoticed
