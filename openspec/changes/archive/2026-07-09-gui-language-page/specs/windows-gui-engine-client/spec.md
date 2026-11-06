## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Engine client exposes a settable language defaulting to English

`IInstallerEngineClient` SHALL declare a `Language { get; set; }` property whose default value is `"en"`. Setting it SHALL change the language threaded into subsequent text-emitting subcommand invocations without otherwise altering the contract, so the shell can switch languages between flows.

#### Scenario: Interface exposes a settable Language defaulting to en

- **WHEN** an `IInstallerEngineClient` is inspected before any language is chosen
- **THEN** its `Language` reads `"en"`, and after `Language` is set to `"es"` subsequent text-emitting invocations use `es`
