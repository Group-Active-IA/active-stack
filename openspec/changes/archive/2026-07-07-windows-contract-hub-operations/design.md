## Context

The native Windows GUI is a thin WPF client over the Go engine, invoked as `active-stack.exe windows <subcmd>` with JSON in/out. The TUI hub (`internal/tui`) offers Starters, Manage Backups, and a guided Uninstall; the GUI cannot reach these because the `windows` contract stops at `detect | options | install | uninstall`. This change (C2 of the approved 6-change plan) adds the missing subcommands as **callers only** over the existing engine.

Current relevant surface:
- `RunWindowsInstall` (`cmd/active-stack/headless/windows.go:109`) owns the JSON event-stream plumbing (event file tailer, `emit`, progress mapping, `install_finished`).
- `RunWindowsDetect` (`windows.go:53`) scans agent configs via `system.ScanConfigs` and maps them with `configStateAgent`.
- `runStarterAdd` (`cmd/active-stack/starter_add.go:43`) builds an `install.Intent` + `headless.ParsedFlags` for a starter and runs `headless.RunHeadless`.
- Catalog starter API: `AllStarters()`, `StarterByID(id)`, `ResolveStarter(id)`, `ResolveStarterMCPs(id)` (`internal/catalog/catalog.go:270,310,315,328`).
- Backup public API: `ListManifests(dir)` (`internal/backup/retention.go:175`), `Manifest.DisplayLabel()` (`manifest.go:81`), `RestoreService{}.Restore` (`restore.go:13`), `DeleteBackup` (`manifest.go:136`), `RenameBackup` (`manifest.go:145`).
- Backup store layout: install snapshots live under `<home>/.active-stack/backups/install`, uninstall snapshots under `<home>/.active-stack/backups/uninstall` (`internal/install/plan.go:110`, `internal/uninstall/plan.go:74`). `ListManifests` scans the immediate subdirectories of a given root for `manifest.json`.
- Uninstall strategies: `StrategyTargeted = "targeted"`, `StrategyRestore = "restore"` (`internal/uninstall/types.go:18,21`); `StrategyRestore` needs a `RestoreManifest`.

## Goals / Non-Goals

**Goals:**
- Give the GUI TUI-parity hub operations through additive `windows` subcommands.
- Reuse existing plumbing: one shared `runWindowsPipeline` for install-style event-streams, one shared `detectAgents` for detect + uninstall-options, one shared `BuildStarterInstallParams` for CLI `starter add` + `windows starters install`.
- Keep every JSON field additive and defaulted so old/new GUI and old/new engine stay compatible.

**Non-Goals:**
- No changes to `internal/backup`, `internal/install`, or `internal/uninstall` — callers only. (`internal/backup` is governance ALTO.)
- No `--tier` / multi-agent options work (that is C1, `windows-contract-tier-multiagent`).
- No C# or WPF work (C3–C5).
- No new backup capabilities (pin/unpin toggle, compression policy) beyond restore/delete/rename.

## Decisions

### D1 — Pre-refactor `runStarterAdd` before adding `windows starters install`

Extract the intent/params construction (steps 1–8 of `runStarterAdd`) into a pure `headless.BuildStarterInstallParams` in a new `cmd/active-stack/headless/starter_params.go`. `runStarterAdd` becomes a wrapper that calls it and then runs `RunHeadless`. Rationale: the `windows starters install` handler needs the identical `ParsedFlags`; duplicating the resolution logic would drift. Alternative considered — inline duplication in the windows handler — rejected because it violates single-source-of-truth and risks the two paths diverging on MCP aggregation / effective-starter construction.

### D2 — `windows starters list` schema

```json
{"starters":[{id,name,description,includes,harnesses,mcp_count}]}
```

`harnesses` is the resolved id set from `ResolveStarter(id)`; `mcp_count` is `len(ResolveStarterMCPs(id))`; `includes` is the raw `Starter.Includes`. Sourced from `cat.AllStarters()` in catalog order. Empty catalog → `{"starters":[]}` (never null; pre-size slices as in `RunWindowsOptions`).

### D3 — `windows starters install` reuses the install event-stream

Extract a shared helper `runWindowsPipeline` from `RunWindowsInstall` that owns the event-file tailer, `emit`, progress/download wiring, and the plain-output replay — parameterized by the terminal event `type`. `RunWindowsInstall` passes `install_finished`; `windows starters install` passes `starter_finished`. The handler builds its `ParsedFlags` via `BuildStarterInstallParams` (D1), parses `--agent` as CSV, and honors `--dry-run` / `--yes`.

`windows starters install --starter <id> --project <dir> --agent <csv> [--dry-run] [--yes]` → same event contract as `windows install`, final event:

```json
{"type":"starter_finished","phase":"install","success":true,"message":"...","timestamp":"..."}
```

### D4 — `windows backups list` schema and store resolution

```json
{"backups":[{id,created_at,source,description,file_count,pinned,compressed,display_label,manifest_path}]}
```

Resolve the backup store root from the resolved home (`--home` override, else `os.UserHomeDir()`), then aggregate `backup.ListManifests` over the `install` and `uninstall` subdirectories under `<home>/.active-stack/backups`. For each `Manifest`: `id`←`ID`, `created_at`←`CreatedAt` (RFC3339), `source`←`Source.Label()` (or the raw string), `description`←`Description`, `file_count`←`FileCount`, `pinned`←`Pinned`, `compressed`←`Compressed`, `display_label`←`DisplayLabel()`, `manifest_path`←`filepath.Join(RootDir, ManifestFilename)`. `manifest_path` is included because the later uninstall-restore flow (C5) needs to load the manifest to pass as `RestoreManifest`. Read-only. Empty/missing store → `{"backups":[]}`.

### D5 — `windows backups restore|delete|rename` schema, public API only

```json
{"success":true,"message":"...","id":"<id>"}
```

Handler flow: `ListManifests` over the store → find the manifest whose `ID == --id`. If not found → `{"success":false,"message":"backup <id> not found","id":"<id>"}`, exit non-zero. Otherwise dispatch by action using ONLY the public API:
- `restore` → `backup.RestoreService{}.Restore(manifest)`
- `delete` → `backup.DeleteBackup(manifest)`
- `rename` → `backup.RenameBackup(manifest, --description)`

On error from the public API → `success:false` with the error message, exit non-zero. **`internal/backup` is not modified** (governance ALTO). Alternative considered — adding a lookup-by-id helper inside `internal/backup` — rejected: it would modify the ALTO package; the caller does the id match over `ListManifests` instead.

### D6 — `windows uninstall-options` schema and shared `detectAgents`

```json
{"detected_agents":[...],"modes":[{id,label,description}],"strategies":[{id,label,description,default,requires_manifest}]}
```

Extract `detectAgents(homeDir) []string` from `RunWindowsDetect` (the `ScanConfigs` + `configStateAgent` loop); `RunWindowsDetect` then calls it so its output is unchanged. `modes` reuses the same three mode options as `RunWindowsOptions` (`lite/full/custom`). `strategies`:

```json
[
  {"id":"targeted","label":"...","description":"...","default":true,"requires_manifest":false},
  {"id":"restore","label":"...","description":"...","default":false,"requires_manifest":true}
]
```

`targeted` is the default and needs no manifest; `restore` requires a manifest (mirrors `internal/uninstall` semantics where `StrategyRestore` needs `RestoreManifest`).

### D7 — Dispatch routing

Add cases in `runWindowsDispatch` (`windows_dispatch.go`) for `starters` (subcommand `list|install`), `backups` (subcommand `list|restore|delete|rename`), and `uninstall-options`. Each parses its own `flag.FlagSet` (mirroring the existing `detect`/`options` cases) with `--home`, `--starter`, `--project`, `--agent`, `--id`, `--description`, `--dry-run`, `--yes` as applicable. Update the top-level usage string.

## Risks / Trade-offs

- **[Backup actions touch governance ALTO domain]** → This change adds callers only; `internal/backup` is untouched, and per governance ALTO, apply waits for explicit user review/approval of this proposal before any code is written. A destructive action (`restore` overwrites config, `delete` removes a backup) is only ever triggered by an explicit `--id` the caller selected.
- **[Backup store layout assumption]** → the code assumes `<home>/.active-stack/backups/{install,uninstall}`. If a future change relocates snapshots, `windows backups list` must follow. Mitigation: resolve the root from the same constants the install/uninstall plans use; cover with tests that fabricate manifests at that exact path.
- **[Refactor regression in `runStarterAdd`]** → extracting `BuildStarterInstallParams` could alter behavior. Mitigation: the existing `starter add` tests are the safety net (run them RED-free before refactor, green after); the refactor is behavior-preserving.
- **[Extracting `runWindowsPipeline` could regress `windows install`]** → the 5 existing `windows_test.go` install assertions are the safety net and must stay green; the terminal-event `type` is the only parameterized difference.
- **[`Manifest.DisplayLabel()` uses local time]** → `display_label` is locale/timezone-dependent while `created_at` is RFC3339 UTC. Acceptable: `display_label` is for display, `created_at` is the machine-readable field; tests assert `display_label` equals `DisplayLabel()` rather than a hardcoded string.
