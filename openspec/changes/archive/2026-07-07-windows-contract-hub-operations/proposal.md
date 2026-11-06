## Why

The Windows GUI wizard rebuild (planned changes C4–C5) needs engine parity with the TUI hub, which offers Starters, Manage Backups, and a guided Uninstall. Today the `windows` JSON contract only exposes `detect`, `options`, `install`, and `uninstall` — there is no way for the GUI to list/install starters, list/restore/delete/rename backups, or fetch the uninstall option set (detected agents, modes, strategies). This change adds those subcommands so the GUI reaches full hub parity through the JSON contract, by adding **callers only** over the existing engine — without touching the governance-sensitive `internal/backup`, `internal/install`, or `internal/uninstall` packages.

## What Changes

- **Pre-refactor (enabler):** extract the intent/params construction inside `runStarterAdd` (`cmd/active-stack/starter_add.go`) into a new pure helper `headless.BuildStarterInstallParams`; `runStarterAdd` becomes a thin wrapper over it. This lets both the CLI `starter add` path and the new `windows starters install` path share one construction routine. Existing starter tests stay green.
- Add `windows starters list` → `{"starters":[{id,name,description,includes,harnesses,mcp_count}]}`, sourced from `cat.AllStarters()` / `ResolveStarter` / `ResolveStarterMCPs`.
- Add `windows starters install --starter <id> --project <dir> --agent <csv> [--dry-run] [--yes]` → the same JSON event-stream as `windows install`, with a final event `starter_finished`. Extract a shared helper `runWindowsPipeline` out of `RunWindowsInstall` so the event-stream plumbing is written once.
- Add `windows backups list [--home]` → `{"backups":[{id,created_at,source,description,file_count,pinned,compressed,display_label,manifest_path}]}`, sourced from `backup.ListManifests` and `Manifest.DisplayLabel()`.
- Add `windows backups restore|delete|rename --id <id> [--description]` → `{"success","message","id"}`, calling ONLY the public `internal/backup` API (`RestoreService{}.Restore`, `DeleteBackup`, `RenameBackup`).
- Add `windows uninstall-options` → `{detected_agents, modes, strategies:[{id,label,description,default,requires_manifest}]}`. Extract a shared `detectAgents(homeDir)` helper reused by `RunWindowsDetect`.
- Route all new subcommands in `cmd/active-stack/windows_dispatch.go`.

No breaking changes: all additions are new subcommands and new helpers; the existing `windows` subcommands and their `windows_test.go` assertions remain valid.

## Capabilities

### New Capabilities
- `windows-hub-contract`: the JSON/CLI contract between the active-stack engine and the native Windows GUI for the hub operations at parity with the TUI — starter catalog + install, backup listing + restore/delete/rename actions, and the uninstall option set (detected agents, modes, strategies).

### Modified Capabilities
<!-- The C1 capability `windows-install-contract` is not yet in openspec/specs/ (that change is unarchived) and its requirements do not change here. Nothing to modify. -->

## Impact

- **Code (callers only):** `cmd/active-stack/starter_add.go` (extract `BuildStarterInstallParams`, wrapper), `cmd/active-stack/headless/starter_params.go` (new pure helper), `cmd/active-stack/headless/windows.go` (new `RunWindowsStarters*`, `RunWindowsBackups*`, `RunWindowsUninstallOptions`; extract `runWindowsPipeline` + `detectAgents`), `cmd/active-stack/windows_dispatch.go` (new subcommand routing).
- **Contract:** the `windows` command gains `starters list`, `starters install`, `backups list`, `backups restore|delete|rename`, and `uninstall-options`. Consumed later by C3 (C# parsers) and C5 (WPF hub pages).
- **Tests:** `cmd/active-stack/headless/windows_starters_test.go`, `cmd/active-stack/headless/windows_backups_test.go` (manifests fabricated in a temporary `--home`), `cmd/active-stack/windows_dispatch_test.go` (routing), and the pre-refactor guard on `starter_add` existing tests.
- **Governance:** the backups subcommands read/act on the `internal/backup` package, which is governance **ALTO** (it can destroy user config). This change adds **callers only** — `internal/backup`, `internal/install`, and `internal/uninstall` are NOT modified. Per governance ALTO, implementation (apply) MUST wait for explicit user review and approval of this proposal before any code is written.
