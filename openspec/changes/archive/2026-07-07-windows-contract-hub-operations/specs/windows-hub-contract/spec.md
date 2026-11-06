## ADDED Requirements

### Requirement: Starter install params are built by a shared helper

The intent/params construction currently inline in `runStarterAdd` (`cmd/active-stack/starter_add.go`) SHALL be extracted into a pure exported helper `headless.BuildStarterInstallParams`, which resolves a starter's harnesses and MCPs from the catalog and returns the `headless.ParsedFlags` needed to run the pipeline. `runStarterAdd` MUST become a thin wrapper that delegates to this helper, and the new `windows starters install` path MUST reuse the same helper. The refactor MUST NOT change the observable behavior of `starter add`.

#### Scenario: runStarterAdd delegates to the shared helper

- **WHEN** `runStarterAdd` is invoked with a known starter id and a target project
- **THEN** it builds its `ParsedFlags` via `headless.BuildStarterInstallParams` AND produces the same plan/exit code as before the refactor (existing `starter add` tests stay green)

#### Scenario: Unknown starter is rejected by the helper path

- **WHEN** `BuildStarterInstallParams` is asked to build params for a starter id absent from the catalog
- **THEN** it returns an error identifying the unknown starter AND no pipeline runs

### Requirement: windows starters list emits the starter catalog

The `active-stack windows starters list` command SHALL emit a JSON object `{"starters":[...]}` where each entry has the fields `id`, `name`, `description`, `includes` (array of included starter ids), `harnesses` (array of resolved harness ids from `ResolveStarter`), and `mcp_count` (integer count from `ResolveStarterMCPs`). The list MUST reflect `cat.AllStarters()` in catalog order.

#### Scenario: All starters are listed with resolved counts

- **WHEN** `windows starters list` runs against a catalog with starters
- **THEN** the response `starters` array contains one entry per starter from `AllStarters()` in catalog order, each carrying its `id`, `name`, `description`, `includes`, resolved `harnesses`, and `mcp_count`

#### Scenario: Empty catalog yields an empty array

- **WHEN** `windows starters list` runs against a catalog with no starters
- **THEN** the response is `{"starters":[]}` (an empty array, never null)

### Requirement: windows starters install streams the install event contract

The `active-stack windows starters install --starter <id> --project <dir> --agent <csv> [--dry-run] [--yes]` command SHALL run the starter installation through the shared `runWindowsPipeline` helper (extracted from `RunWindowsInstall`) and emit the SAME JSON event-stream as `windows install`, except that the terminal event's `type` MUST be `starter_finished` (rather than `install_finished`). The `--agent` value MUST be parsed as a comma-separated list. A `--dry-run` run MUST NOT mutate the target project.

#### Scenario: Successful starter install ends with starter_finished

- **WHEN** `windows starters install --starter <id> --project <dir> --agent claude` completes successfully
- **THEN** the emitted event-stream matches the install event contract AND its final event has `type: "starter_finished"` with `success: true`

#### Scenario: Dry-run does not mutate the project

- **WHEN** `windows starters install ... --dry-run` is run
- **THEN** the event-stream is produced AND no files in the target project are created or modified

#### Scenario: Unknown starter fails the stream

- **WHEN** `windows starters install --starter <unknown> ...` is run
- **THEN** a failure is reported in the stream AND the command exits non-zero

### Requirement: windows backups list reports backups from the home backup store

The `active-stack windows backups list [--home]` command SHALL read backups via the public `backup.ListManifests` API rooted at the user's backup store under the resolved home directory, and emit a JSON object `{"backups":[...]}`. Each entry MUST carry `id`, `created_at`, `source`, `description`, `file_count`, `pinned`, `compressed`, `display_label` (from `Manifest.DisplayLabel()`), and `manifest_path` (the absolute path to the backup's `manifest.json`, needed later by uninstall-restore). The command MUST be read-only and MUST NOT modify `internal/backup`.

#### Scenario: Backups are listed from a fabricated home store

- **WHEN** `windows backups list --home <tmp>` runs against a home whose backup store contains fabricated manifests
- **THEN** the response `backups` array has one entry per manifest, each with `id`, `created_at`, `source`, `description`, `file_count`, `pinned`, `compressed`, `display_label`, and a valid `manifest_path`

#### Scenario: display_label mirrors the manifest label

- **WHEN** a listed backup has a known source, timestamp, and file count
- **THEN** its `display_label` equals the value returned by `Manifest.DisplayLabel()` for that manifest

#### Scenario: No backups yields an empty array

- **WHEN** `windows backups list --home <tmp>` runs against a home with no backup store or an empty one
- **THEN** the response is `{"backups":[]}` (an empty array, never null) AND the exit code is zero

### Requirement: windows backups actions restore, delete, and rename via the public API

The `active-stack windows backups restore|delete|rename --id <id> [--description <text>]` commands SHALL locate the manifest by `--id` from the home backup store and act on it using ONLY the public `internal/backup` API: `RestoreService{}.Restore` for `restore`, `DeleteBackup` for `delete`, and `RenameBackup` (with `--description`) for `rename`. Each command MUST emit `{"success":<bool>,"message":<string>,"id":<string>}`. An unknown `--id` MUST yield `success:false` with a descriptive message and a non-zero exit code. `internal/backup` MUST NOT be modified by this change.

#### Scenario: Restore invokes the public restore service

- **WHEN** `windows backups restore --id <known>` is run
- **THEN** the matching manifest is restored via `RestoreService{}.Restore` AND the response is `{"success":true,"message":...,"id":"<known>"}`

#### Scenario: Delete removes the backup

- **WHEN** `windows backups delete --id <known>` is run
- **THEN** the matching backup is removed via `DeleteBackup` AND the response reports `success:true` with the id

#### Scenario: Rename updates the description

- **WHEN** `windows backups rename --id <known> --description "before upgrade"` is run
- **THEN** the manifest description is updated via `RenameBackup` AND the response reports `success:true` with the id

#### Scenario: Unknown id fails gracefully

- **WHEN** any backup action is run with an `--id` not present in the store
- **THEN** the response is `{"success":false,"message":...,"id":"<id>"}` AND the command exits non-zero

### Requirement: windows uninstall-options exposes the guided-uninstall option set

The `active-stack windows uninstall-options` command SHALL emit a JSON object with `detected_agents` (from the shared `detectAgents(homeDir)` helper, also used by `RunWindowsDetect`), `modes` (the install-mode options), and `strategies` — an array of `{id, label, description, default, requires_manifest}` objects. The strategy whose `id` is `targeted` MUST be marked `default: true` and `requires_manifest: false`; the `restore` strategy MUST be marked `requires_manifest: true`. Extracting `detectAgents` MUST NOT change `RunWindowsDetect`'s observable output.

#### Scenario: Options carry detected agents, modes, and strategies

- **WHEN** `windows uninstall-options --home <tmp>` runs against a home with installed agent configs
- **THEN** the response `detected_agents` equals what `windows detect` reports for the same home AND `modes` is non-empty AND `strategies` lists at least `targeted` and `restore`

#### Scenario: Targeted is the default strategy

- **WHEN** the uninstall options are served
- **THEN** the `targeted` strategy has `default: true` and `requires_manifest: false`

#### Scenario: Restore strategy requires a manifest

- **WHEN** the uninstall options are served
- **THEN** the `restore` strategy has `requires_manifest: true`

### Requirement: windows dispatch routes the new hub subcommands

The `cmd/active-stack/windows_dispatch.go` router SHALL recognize and dispatch the new subcommand paths: `starters list`, `starters install`, `backups list`, `backups restore`, `backups delete`, `backups rename`, and `uninstall-options`. An unknown subcommand or missing required flag MUST print an error and return a non-zero exit code. The existing `detect|options|install|uninstall` routes MUST keep working unchanged.

#### Scenario: New subcommands reach their handlers

- **WHEN** the dispatcher receives `starters list`, `starters install`, `backups list`, `backups <action>`, or `uninstall-options`
- **THEN** it routes to the corresponding handler and returns that handler's exit code

#### Scenario: Missing required flag is rejected

- **WHEN** a backup action subcommand is dispatched without `--id`
- **THEN** the dispatcher prints an error AND returns a non-zero exit code

#### Scenario: Existing routes are preserved

- **WHEN** the dispatcher receives `detect`, `options`, `install`, or `uninstall`
- **THEN** it routes exactly as before this change (the existing `windows_test.go` assertions stay green)
