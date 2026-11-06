# Tasks — windows-contract-hub-operations

> Strict TDD: for each behavior, write the failing test FIRST (RED), minimum code to pass (GREEN), then triangulate (≥2 cases) and refactor. `internal/backup`, `internal/install`, `internal/uninstall` are NOT modified — callers only. Governance ALTO on backups: do not start apply until the proposal is reviewed/approved.

## 0. Safety net (baseline before refactor)

- [x] 0.1 Run `go test ./cmd/active-stack/...` and record the baseline for the existing `starter add` tests and the 5 `windows_test.go` install assertions (must be green before any change).

## 1. Pre-refactor: extract `BuildStarterInstallParams`

- [x] 1.1 RED: add a test in `cmd/active-stack/headless/starter_params_test.go` asserting `BuildStarterInstallParams` returns `ParsedFlags` with `Target=Project`, `Intent.Mode=Custom`, resolved harness ids in `Intent.Custom`, focal agents, and an effective `Starter` carrying aggregated MCPs; plus a case where an unknown starter id returns an error.
- [x] 1.2 GREEN: create `cmd/active-stack/headless/starter_params.go` with `BuildStarterInstallParams` extracted from `runStarterAdd` steps 1–8 (resolve starter, harnesses, MCPs, effective starter, intent, ParsedFlags).
- [x] 1.3 Refactor `runStarterAdd` (`cmd/active-stack/starter_add.go`) into a thin wrapper over `BuildStarterInstallParams` + `RunHeadless`; confirm existing `starter add` tests stay green (behavior preserved).

## 2. Shared helpers extraction

- [x] 2.1 RED: add a test asserting the extracted `runWindowsPipeline` emits the install event contract and that the terminal event `type` is parameterizable (install vs starter). Keep the existing `windows_test.go` install assertions as the regression guard.
- [x] 2.2 GREEN: extract `runWindowsPipeline` from `RunWindowsInstall` (`cmd/active-stack/headless/windows.go`), parameterized by terminal event `type`; `RunWindowsInstall` passes `install_finished`. Confirm the 5 existing install assertions stay green.
- [x] 2.3 RED: add a test asserting `detectAgents(homeDir)` returns the same agent set that `RunWindowsDetect` emits for a fabricated `--home`.
- [x] 2.4 GREEN: extract `detectAgents(homeDir) []string` from `RunWindowsDetect`; have `RunWindowsDetect` call it. Confirm detect output is unchanged.

## 3. windows starters list + install

- [x] 3.1 RED: in `cmd/active-stack/headless/windows_starters_test.go`, assert `windows starters list` emits `{"starters":[{id,name,description,includes,harnesses,mcp_count}]}` in catalog order (≥2 starters), and an empty catalog yields `{"starters":[]}` (not null).
- [x] 3.2 GREEN: implement `RunWindowsStartersList` sourcing `AllStarters()` / `ResolveStarter` / `ResolveStarterMCPs`.
- [x] 3.3 RED: assert `windows starters install --starter <id> --project <dir> --agent claude` streams the install event contract with a final `starter_finished` event (`success:true`); and a `--dry-run` case produces the stream without mutating the project; and an unknown starter fails the stream with non-zero exit.
- [x] 3.4 GREEN: implement `RunWindowsStartersInstall` using `BuildStarterInstallParams` (task 1) + `runWindowsPipeline` (task 2) with terminal type `starter_finished`; parse `--agent` as CSV.

## 4. windows backups list

- [x] 4.1 RED: in `cmd/active-stack/headless/windows_backups_test.go`, fabricate manifests under a temp `--home` backup store and assert `windows backups list` emits one entry per manifest with `id,created_at,source,description,file_count,pinned,compressed,display_label,manifest_path`; assert `display_label` equals `Manifest.DisplayLabel()`; assert an empty/missing store yields `{"backups":[]}` with exit 0.
- [x] 4.2 GREEN: implement `RunWindowsBackupsList` aggregating `backup.ListManifests` over `<home>/.active-stack/backups/{install,uninstall}`, mapping fields per design D4 (read-only).

## 5. windows backups restore | delete | rename

- [x] 5.1 RED: assert `restore --id <known>` returns `{"success":true,...,"id":<known>}` and invokes `RestoreService{}.Restore` on the matching manifest (verify via restored file/side effect in temp home).
- [x] 5.2 RED: assert `delete --id <known>` removes the backup via `DeleteBackup` and returns `success:true`; assert `rename --id <known> --description <text>` updates the manifest via `RenameBackup` and returns `success:true`.
- [x] 5.3 RED: assert an unknown `--id` for any action returns `{"success":false,...,"id":<id>}` and a non-zero exit.
- [x] 5.4 GREEN: implement `RunWindowsBackupsAction` (restore/delete/rename): match manifest by id over `ListManifests`, dispatch to the public backup API only, emit `{"success","message","id"}`. Do NOT modify `internal/backup`.

## 6. windows uninstall-options

- [x] 6.1 RED: assert `windows uninstall-options --home <tmp>` emits `detected_agents` equal to `windows detect`, non-empty `modes`, and `strategies` including `targeted` (`default:true`, `requires_manifest:false`) and `restore` (`requires_manifest:true`).
- [x] 6.2 GREEN: implement `RunWindowsUninstallOptions` using `detectAgents` (task 2), the shared mode options, and the two strategy entries per design D6.

## 7. Dispatch routing

- [x] 7.1 RED: in `cmd/active-stack/windows_dispatch_test.go`, assert routing for `starters list`, `starters install`, `backups list`, `backups restore|delete|rename`, and `uninstall-options` reaches the correct handler; assert a backup action without `--id` errors with non-zero exit; assert existing `detect|options|install|uninstall` routes are unchanged.
- [x] 7.2 GREEN: add the new cases to `runWindowsDispatch` (`cmd/active-stack/windows_dispatch.go`) with per-subcommand `flag.FlagSet` parsing (`--home,--starter,--project,--agent,--id,--description,--dry-run,--yes`); update the usage string.

## 8. Verification

- [x] 8.1 Run `go test ./cmd/active-stack/... ./internal/...`; all green, including the untouched `internal/backup` package.
- [x] 8.2 Manual contract smoke against a sandbox `--home`: `windows starters list`, `windows backups list`, `windows uninstall-options`; validate each JSON payload matches the schemas in design.md.
