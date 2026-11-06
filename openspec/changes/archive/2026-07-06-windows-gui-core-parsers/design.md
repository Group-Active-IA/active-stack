## Context

The Go engine emits a JSON contract over `active-stack windows <subcmd>`. The WPF installer (`installer/windows/ui`) is a thin client that shells out to the engine and parses its JSON. Today it consumes only two commands:

- `windows detect` → `{"detected_agents":[...]}`, parsed by `InstallerSessionStateBuilder.BuildFromJson` (`Core`).
- `windows options --agent <id>` → modes + forced/custom components, parsed by the same builder.

The engine now has a much wider contract (C1+C2, both implemented and green). This change makes it consumable from C#. Current constraints observed in the codebase:

- `InstallerSessionStateBuilder` already models `detect` + `options`; the tier fields (`tier_capable`, `tier_capable_agents`, `permission_tiers`) are new and MUST be additive.
- `ProcessInstallerEngineClient.RunInstallAsync` builds arguments with string interpolation (`Arguments = $"windows install --agent {assistantId} --mode {installTypeId}"`, line 33). Any project path or backup description with a space breaks argument tokenization.
- `LoadSessionAsync` calls `PickFirstAssistantId` (line 97), collapsing multi-agent detection to one agent — the engine already accepts `--agent claude,codex` and returns a deduped union, so the GUI should pass all detected agents.
- Tests use xUnit with C# raw-string-literal JSON fixtures (`InstallerSessionStateBuilderTests.cs`). Runner: `dotnet test`, target `net8.0` / `net8.0-windows`.

Governance: MEDIO (consumer adapters). No Go code and no `internal/*` are touched.

### Contract schemas consumed (source of truth: `cmd/active-stack/headless/windows*.go` and their tests)

The C# parsers and every test fixture MUST match these byte-for-byte with the Go test outputs (schema-sync rule). Field names below are the exact JSON keys.

**`windows options --agent <csv>`** (`windowsOptionsResponse`, `windows.go:24`) — existing fields plus additive tier metadata:

```json
{
  "modes": [
    { "id": "lite",   "label": "Quick",    "description": "Fast setup to start working right away." },
    { "id": "full",   "label": "Complete", "description": "Full recommended setup with all key tools." },
    { "id": "custom", "label": "Custom",   "description": "Choose exactly what to install." }
  ],
  "forced_components": [ { "id": "permissions", "label": "...", "description": "...", "recommended": true } ],
  "custom_components": [ { "id": "openspec",    "label": "...", "description": "...", "recommended": true } ],
  "tier_capable": true,
  "tier_capable_agents": ["claude"],
  "permission_tiers": [
    { "id": "estricto",   "label": "Estricto",   "description": "...", "default": false },
    { "id": "balanceado", "label": "Balanceado", "description": "...", "default": true },
    { "id": "bypass",     "label": "Bypass",     "description": "...", "warning": "Bypass: autonomous mode ..." }
  ]
}
```

Notes: `recommended` and `warning` are `omitempty` in Go → may be **absent** in JSON; parser MUST default them (false / null). `permission_tiers` is always the same three entries in display order; only `tier_capable` / `tier_capable_agents` vary with the requested agents. `tier_capable_agents` is the requested subset that is tier-capable, in request order.

**`windows starters list`** (`windowsStarterEntry`, `windows_starters.go:12`):

```json
{ "starters": [
  { "id": "alpha", "name": "Alpha", "description": "...", "includes": ["beta"], "harnesses": ["harness-a","harness-b"], "mcp_count": 2 }
] }
```

`includes`/`harnesses` are always arrays (never null); empty catalog → `{"starters":[]}`.

**`windows starters install`** streams JSONL install events; terminal event `{ "type": "starter_finished", "success": bool, ... }` (event shape = `windowsInstallEvent`, already parsed by `InstallProgressSnapshotParser`).

**`windows backups list [--home]`** (`windowsBackupEntry`, `windows_backups.go:21`):

```json
{ "backups": [
  { "id": "backup-1", "created_at": "2026-03-15T10:30:00Z", "source": "install", "description": "before upgrade",
    "file_count": 3, "pinned": true, "compressed": false,
    "display_label": "...", "manifest_path": "C:\\...\\manifest.json" }
] }
```

Empty store → `{"backups":[]}`. `created_at` is RFC3339. `manifest_path` is the absolute path needed by uninstall-restore.

**`windows backups restore|delete|rename --id <id> [--description <d>]`** (`backupActionResponse`, `windows_backups_action_test.go:18`):

```json
{ "success": true, "message": "...", "id": "backup-1" }
```

Unknown id → `{ "success": false, "message": "<why>", "id": "<requested-id>" }` and non-zero exit.

**`windows uninstall-options`** (`windowsUninstallOptionsResponse`, `windows_uninstall_options.go:20`):

```json
{
  "detected_agents": ["claude"],
  "modes": [ { "id": "lite", "label": "Quick", "description": "..." }, ... ],
  "strategies": [
    { "id": "targeted", "label": "Targeted",           "description": "...", "default": true,  "requires_manifest": false },
    { "id": "restore",  "label": "Restore from backup", "description": "...", "default": false, "requires_manifest": true }
  ]
}
```

`detected_agents` equals `windows detect`'s output for the same home.

## Goals / Non-Goals

**Goals:**
- Typed, defensively-parsed C# models for every new `windows` response, in the existing `BuildFromJson` static-parser style.
- Pure argv builders (`IReadOnlyList<string>`) for every `windows` subcommand — testable with zero process spawns.
- An engine client that invokes the new subcommands correctly (`ArgumentList`, multi-agent CSV) and exposes async methods for each.
- Full parser + builder test coverage with fixtures byte-for-byte mirroring Go test JSON.

**Non-Goals:**
- No Go / `internal/*` changes (contract is frozen upstream).
- No XAML, ViewModels, navigation, or wizard flow (that is C4/C5).
- No behavioral change to the existing `detect`/`options`/install event flow beyond the additive tier fields and multi-agent CSV.
- No new NuGet dependencies (`System.Text.Json` only, as today).

## Decisions

### D1 — Tier fields are additive on `InstallerSessionState`
Add `TierCapable` (bool), `TierCapableAgents` (`IReadOnlyList<string>`), and `PermissionTierChoices` (`IReadOnlyList<PermissionTierChoice>`) to the `InstallerSessionState` record, parsed inside the existing `BuildFromJson`. New internal DTOs: `WindowsOptionsResponse` gains `tier_capable`, `tier_capable_agents`, `permission_tiers`; a `WindowsPermissionTierOption` DTO mirrors `{id,label,description,default,warning}`. New public record `PermissionTierChoice(string Id, string Label, string Description, bool IsDefault, string? Warning)`.
- **Absent-JSON defaults**: missing `tier_capable` → `false`; missing `tier_capable_agents` → empty list; missing `permission_tiers` → empty list. This keeps the parser compatible with a pre-tier engine (old binary) and satisfies the "absent → safe defaults" requirement.
- **Alternative rejected**: a separate `TierMetadataParser` reading the same `options` JSON twice — redundant deserialization and a second source of truth for the same document. The tier data lives in the `options` response, so it belongs in the one builder that already owns it.

### D2 — New parsers are static `BuildFromJson`-style classes, one per response
`StarterCatalogParser`, `BackupListParser`, `BackupActionResultParser`, `UninstallOptionsParser` each expose `public static <Result> BuildFromJson(string json)` using the shared `PropertyNameCaseInsensitive` options, internal `[JsonPropertyName]` DTOs, and public record results. This mirrors `InstallerSessionStateBuilder` exactly (consistency, zero new patterns). Each returns collection-typed results that never expose null collections (empty list on absent/null arrays).
- Public result records: `StarterChoice(Id, Name, Description, IReadOnlyList<string> Includes, IReadOnlyList<string> Harnesses, int McpCount)`; `BackupEntry(Id, CreatedAt, Source, Description, FileCount, Pinned, Compressed, DisplayLabel, ManifestPath)`; `BackupActionResult(Success, Message, Id)`; `UninstallOptions(IReadOnlyList<string> DetectedAgents, IReadOnlyList<InstallTypeChoice> Modes, IReadOnlyList<UninstallStrategyChoice> Strategies)` with `UninstallStrategyChoice(Id, Label, Description, IsDefault, RequiresManifest)`.
- `created_at` is parsed as `string` (kept verbatim, RFC3339) to avoid timezone/formatting drift; the UI layer decides display formatting. **Alternative rejected**: parse to `DateTimeOffset` — invites round-trip/format mismatch against the Go RFC3339 string and complicates fixtures.

### D3 — `InstallArguments` returns argument LISTS, not strings
New static class `InstallArguments` in `Core` with pure methods returning `IReadOnlyList<string>` (or `string[]`):
- `BuildInstallArgs(IEnumerable<string> agents, string mode, IEnumerable<string> customIds, string? tier)` →
  `["windows","install","--agent","claude,codex","--mode","custom","--custom","a,b","--tier","balanceado"]`.
  Rules: `--custom` is omitted when `customIds` is empty **or** `mode != "custom"`; `--tier` is omitted when `tier` is null/empty; `--agent` joins agents with `,`.
- `BuildStarterInstallArgs(string starterId, string projectPath, IEnumerable<string> agents, bool dryRun, bool yes)` → `["windows","starters","install","--starter",id,"--project",path,"--agent",csv,("--dry-run"),("--yes")]`. The project path is a **single list element** (never quoted/concatenated) so spaces survive.
- `BuildUninstallArgs(IEnumerable<string> agents, string mode, string strategy, string? backupId)` → includes `--strategy` and, when strategy is `restore`, `--id <backupId>` (or `--manifest`, matched to the engine's uninstall flag — verified against C2's uninstall flags during apply; captured as Open Question O1).
- `BuildBackupActionArgs(string action, string id, string? description)` → `["windows","backups",action,"--id",id]` plus `--description <d>` only for `rename` with a non-empty description.
- **Rationale**: `ProcessStartInfo.ArgumentList` consumes a token list and does OS-correct quoting; producing a list (not a string) is the only robust way to carry paths with spaces, and makes builders trivially unit-testable (assert on the list) with no process.
- **Alternative rejected**: keep building a single argument string and rely on manual quoting — the current bug. Quoting rules differ across Win32 argv parsing and are error-prone.

### D4 — Interface shape: extend `IInstallerEngineClient` (single interface)
Add the new async methods directly to `IInstallerEngineClient` rather than introducing sibling interfaces (`IStarterClient`, `IBackupClient`, …):
- `Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken)`
- `IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken)`
- `Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken)`
- `Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken)`
- `Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken)`
- `IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken)`
- **Rationale**: there is exactly one production implementation (`ProcessInstallerEngineClient`) and one test-fake surface. Splitting into role interfaces adds indirection with no reuse benefit and no distinct lifetime. C4/C5 ViewModels each take the one client. A single interface keeps the fake in `Host.Tests` to one class.
- **Alternative considered**: sibling interfaces per hub area (segregation). Rejected for this codebase size; can be revisited if the client grows independent lifetimes. This decision is the one the task flagged to capture explicitly.

### D5 — `LoadSessionAsync` passes all detected agents; drop `PickFirstAssistantId`
`LoadSessionAsync` runs `windows detect`, parses `detected_agents`, and calls `windows options --agent <csv-of-all-detected>` (comma-joined), then `BuildFromJson`. `PickFirstAssistantId` is deleted. The engine dedupes the union, so passing all detected agents yields the correct combined component/tier picture. When detection is empty, fall back to `claude` (preserves current safe default).

### D6 — `ProcessInstallerEngineClient` uses `ArgumentList` everywhere
Both the streaming (`RunInstallAsync` + new stream methods) and the one-shot JSON helper (`RunJsonCommandAsync`) switch to a shared private helper that builds `ProcessStartInfo` and appends each token via `StartInfo.ArgumentList.Add(...)`. All argument construction routes through `InstallArguments`. The one-shot commands (`detect`, `options`, `starters list`, `backups list`, `uninstall-options`) also use `ArgumentList` (each token separate). The event-stream methods reuse the existing line-reading loop, differing only in the argument list and the fact that the terminal event `type` differs (`starter_finished` / `uninstall_finished`) — parsing is unchanged (`InstallProgressSnapshotParser`).

## Risks / Trade-offs

- **[Fixture drift vs. Go]** C# fixtures could silently diverge from the real Go output (e.g., `omitempty` fields, RFC3339 formatting, key casing). → Copy fixtures byte-for-byte from the Go test JSON during apply; keep `PropertyNameCaseInsensitive`; treat absent optional fields (`recommended`, `warning`) as defaults. A later smoke step (C6) runs the real engine and diffs.
- **[Uninstall restore flag name]** The exact engine flag for selecting a backup on `windows uninstall` (`--id` vs `--manifest <manifest_path>`) is not pinned by the tests read here. → Verify against C2's uninstall flag parser during apply before finalizing `BuildUninstallArgs` (Open Question O1). Low blast radius: one builder method + its test.
- **[Single interface growth]** `IInstallerEngineClient` becomes larger. → Acceptable at this size; segregation is a cheap later refactor if lifetimes diverge.
- **[ArgumentList behavior parity]** Switching from `Arguments` to `ArgumentList` changes how the child process receives argv. → Existing install-stream tests (`MainWindowViewModelTests`, tailer tests) guard the event flow; builder unit tests assert exact token lists.

## Migration Plan

Additive and consumer-only; no runtime migration. Order: (1) additive tier fields + their tests; (2) `InstallArguments` + builder tests; (3) new parsers + fixture tests; (4) engine-client `ArgumentList` migration + `LoadSessionAsync` multi-agent + new methods + Host tests. Rollback = revert the C# files; the Go contract and existing GUI behavior are unaffected. No UI is wired, so nothing user-visible changes until C4/C5.

## Open Questions

- **O1**: Exact `windows uninstall` flag for restore-from-backup (`--id <backupId>` vs `--manifest <manifest_path>`). Resolve by reading C2's uninstall flag parser during apply; adjust `BuildUninstallArgs` + its test accordingly.
- **O2**: Whether `RunStarterInstallAsync`/`RunUninstallAsync` should surface the terminal event's `success`/`message` as a typed result in addition to streaming snapshots. Default: keep streaming-only for parity with `RunInstallAsync`; revisit if C5 ViewModels need the terminal payload directly.
