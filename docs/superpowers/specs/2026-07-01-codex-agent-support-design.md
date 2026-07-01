# Codex Agent Support Design

Date: 2026-07-01
Status: approved

## Goal

Add Codex as a first-class Active Stack installation target with the same
operational coverage as Claude and OpenCode: machine and project installation,
skills, instructions, MCP configuration, permission tiers, detection,
verification, rollback, and uninstall.

## Scope

Included:

- A native Codex adapter registered in the default agent registry.
- Machine- and project-aware path resolution.
- Codex detection in headless and TUI flows.
- Harness catalog compatibility for Codex.
- Marker-based injection into `AGENTS.md`.
- Skill installation in Codex-supported `.agents/skills` directories.
- MCP registration in Codex TOML configuration.
- Active Stack permission tiers mapped to Codex settings.
- Backup, atomic writes, rollback, verification, and uninstall coverage.
- Correction of stale Codex paths in bundled assets.

Excluded:

- Installing or authenticating the Codex product itself.
- Selecting or changing the user's Codex model.
- Installing Codex plugins, marketplaces, hooks, or unrelated MCP servers.
- Automatically trusting a project.
- Custom slash-command files. Codex exposes built-in slash commands and no
  supported custom command directory is part of this design.

## Official Codex Layout

### Machine target

| Concern | Path |
|---|---|
| Instructions | `~/.codex/AGENTS.md` |
| Skills | `$HOME/.agents/skills` |
| Settings and MCP | `~/.codex/config.toml` |
| Commands | unsupported; empty path |

### Project target

| Concern | Path |
|---|---|
| Instructions | `<root>/AGENTS.md` |
| Skills | `<root>/.agents/skills` |
| Settings and MCP | `<root>/.codex/config.toml` |
| Commands | unsupported; empty path |

Project-local Codex configuration is effective only for trusted projects.
Active Stack writes the requested project configuration but never changes the
project's trust state.

## Architecture

### Codex adapter

Create `internal/agents/codex` implementing the complete `agents.Adapter`
contract:

- `Agent()` returns `model.AgentCodex`.
- `InstructionsPath`, `SkillsDir`, `SettingsPath`, and `MCPConfigPath` resolve
  the machine layout.
- `PathsFor` resolves machine and project layouts without caller-side path
  conditionals.
- `CommandsDir` and project `CommandsDir` are empty.
- `VariantKey()` returns `codex`.
- `ConfigDelivery()` returns `ConfigDeliveryInstructions`.
- `MCPStrategy()` and target-aware paths select a new TOML single-file merge
  strategy.

Register the adapter in `NewAdapter` and `NewDefaultRegistry`. Update registry
assertions and supported-agent tests so Codex cannot silently disappear.

### Detection and catalog

Detection recognizes Codex from its global configuration directory and the
`codex` executable where tool detection is available. TUI detection includes
Codex beside Claude and OpenCode.

Every applicable harness entry in `internal/catalog/harnesses.yaml` declares
Codex support. Catalog edits must pass through `catalog.Load()` validation.
Harnesses without a valid Codex materialization remain excluded explicitly.

### Instructions and skills

The existing config installer composes the Codex
`sdd-orchestrator.md` variant and injects it into `AGENTS.md` using Active Stack
markers. Reinstallation replaces the owned block and never duplicates it.

The skill installer receives the Codex skill directory exclusively through the
adapter. The bundled Codex orchestrator asset must reference
`$HOME/.agents/skills`, not the obsolete `~/.codex/skills` path.

### TOML merge support

Add a target-aware TOML MCP strategy rather than invoking `codex mcp add`.
Keeping the write inside Active Stack preserves dry-run planning, backup,
rollback, atomic writes, and deterministic tests.

The TOML merge layer owns only:

- Top-level permission keys managed by Active Stack.
- `sandbox_workspace_write` when required by the selected tier.
- Individual `[mcp_servers.<id>]` tables and their nested tables.

It preserves unrelated keys and tables, including models, providers, projects,
plugins, marketplaces, hooks, and MCP servers not selected by Active Stack.
The merge is idempotent and preserves valid Windows paths. Existing malformed
TOML aborts the operation before any write.

Use `github.com/pelletier/go-toml/v2` to validate input and merged output. The
actual mutation remains a section-aware textual upsert so comments, ordering,
and unknown Codex settings are not lost through full-file reserialization.

MCP shapes:

- Stdio server: `[mcp_servers.<id>]` with `command`, `args`, and optional
  `[mcp_servers.<id>.env]`.
- HTTP server: `[mcp_servers.<id>]` with `url` and supported optional fields.

Replacing a managed MCP table removes stale keys belonging to its previous
shape, such as an old `url` after switching to stdio.

### Permission tiers

| Active Stack tier | `approval_policy` | `sandbox_mode` | Additional configuration |
|---|---|---|---|
| Estricto | `untrusted` | `read-only` | none |
| Balanceado | `on-request` | `workspace-write` | network access remains disabled |
| Bypass | `never` | `danger-full-access` | none |

Only permission-related Codex keys are changed. Model, reasoning effort,
reviewer, feature flags, and other user choices remain untouched.

## Install Data Flow

1. Detect Codex and expose it as a selectable registered agent.
2. Resolve compatible harnesses from the validated catalog.
3. Resolve all write paths through the Codex adapter.
4. Snapshot every existing path before the first mutation.
5. Install skills.
6. Inject config harnesses into `AGENTS.md` with markers.
7. Merge permission keys and MCP tables into `config.toml`.
8. Write atomically.
9. Run Codex-specific verification checks.
10. If an apply or verify step fails, restore the installation snapshot.

For a project target that is not trusted, installation succeeds with a warning
that Codex will ignore the project-local config until the user trusts it.

## Verification

Verification checks:

- Expected `AGENTS.md` exists and contains exactly one owned section.
- Selected skills exist under the resolved `.agents/skills` path.
- `config.toml` is syntactically valid.
- Permission keys match the selected tier.
- Each selected MCP server has one valid managed table.
- Unrelated TOML keys and tables remain present.
- Project-target verification reports a warning, not a destructive fix, when
  trust cannot be established.

## Uninstall

Targeted uninstall:

- Removes Active Stack marker sections from `AGENTS.md`.
- Removes installed skill directories owned by the selected harnesses.
- Removes only the selected Active Stack-managed MCP tables.
- Restores permission values from the installation manifest when that manifest
  is available.
- If the original permission values cannot be recovered safely, leaves current
  values unchanged and emits a warning.

Restore uninstall continues to replay the selected installation snapshot in
full. Every uninstall mutation gets its own pre-uninstall snapshot so a failed
uninstall can roll back.

## Error Handling

- Unsupported or unregistered Codex: fail during planning, before writes.
- Invalid catalog: fail through `catalog.Load()`.
- Invalid existing TOML: fail before snapshot apply steps mutate files.
- Backup failure: abort before writing.
- Atomic write failure: return an error and restore the snapshot.
- Verification failure: mark the install failed and roll back.
- Missing project trust: warn; never auto-trust.
- Missing safe uninstall provenance: preserve permission values and warn.

## Testing Strategy

Use test-first development. No build command is required.

Adapter tests:

- Machine and project paths on Windows and POSIX-style inputs.
- Empty command directories.
- Variant, delivery mode, and MCP strategy.
- Interface compile-time assertions.

Registry and detection tests:

- `NewAdapter(AgentCodex)` succeeds.
- Default registry returns Claude, Codex, and OpenCode.
- TUI and configuration scanning detect `~/.codex`.

TOML tests:

- Empty file, existing file, CRLF, comments, Windows paths, and nested tables.
- Stdio and HTTP MCP shapes.
- Replacement of stale managed MCP keys.
- Preservation of unrelated plugins, hooks, projects, and MCP servers.
- All tier mappings.
- Idempotent second run.
- Invalid TOML produces no write.

Pipeline tests:

- Snapshot includes all Codex paths before apply.
- Injected failures restore the original files.
- Machine and project installs route through adapter paths.
- Untrusted project produces a warning.

Verify and uninstall tests:

- Healthy, missing, duplicate, and malformed states.
- Targeted removal preserves unrelated TOML.
- Permission restoration with a manifest.
- Missing provenance preserves user settings and warns.
- Restore strategy replays the complete snapshot.

Catalog tests:

- `catalog.Load()` accepts the updated catalog.
- Mode and agent filters include Codex only where supported.

The relevant Go test suites are run after implementation. Per project policy,
no `go build` is run unless the operator explicitly requests it.

## Governance and Checkpoints

- Adapter, catalog, detection, and TUI changes are implemented first.
- TOML merge and permissions are governance ALTO and require focused tests
  before integration.
- Pipeline and uninstall changes are governance ALTO and are reviewed after the
  low-level TOML behavior is proven.
- Existing untracked `.active-stack/` content is user-owned and must not be
  modified.
- No commit is created without an explicit commit request.
