# Spec Delta

## MODIFIED Requirements

### Requirement: The catalog declares a Full-only, Claude-only `playwright` harness

The embedded catalog SHALL contain a global-picker harness with id `playwright`, type `external`, install method `mcp`, supported agents limited to `claude` and `codex`, and install modes limited to `full`. It MUST NOT depend on another harness. Its descriptions MUST be present in `es` and `en`, and the catalog MUST load under existing validation.

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog is loaded and `playwright` is resolved by ID
- **THEN** loading succeeds, its type is `external` with method `mcp`, it supports Claude and Codex only, is Full-only and has no dependencies

#### Scenario: Included in Full, excluded from Lite

- **WHEN** harnesses are listed for Full and Lite
- **THEN** Full contains `playwright` and Lite does not

#### Scenario: Not offered for non-Claude agents

- **WHEN** the harness list is built for `opencode` only
- **THEN** it does not contain `playwright`

### Requirement: Installation registers the server only for Claude

Installing `playwright` SHALL register a stdio server entry named `playwright` with `npx` and pinned arguments in the selected supported agent's MCP configuration, taking a backup before changing an existing file. Unsupported agents' MCP configuration files MUST remain unchanged.

#### Scenario: Claude receives the server entry

- **WHEN** installed with `claude`
- **THEN** Claude's MCP config contains the pinned `playwright` stdio entry and existing config was backed up

#### Scenario: Codex receives the server entry

- **WHEN** installed with `codex`
- **THEN** Codex's TOML MCP config contains a single pinned `playwright` stdio entry and reinstalling reports no change

#### Scenario: Other selected agents are untouched

- **WHEN** installed with `claude`, `codex` and `opencode`
- **THEN** OpenCode's MCP config stays unchanged

#### Scenario: Existing MCP servers are preserved

- **WHEN** installed into an agent MCP config containing another server
- **THEN** the other server entry is unchanged

### Requirement: Post-install verification covers only Claude and skips endpoint checks

Post-install verification SHALL produce one hard MCP configuration check per selected supported agent. It MUST NOT produce a remote endpoint reachability check or checks for unsupported agents.

#### Scenario: Verify checks for a Claude and OpenCode install

- **WHEN** verification checks are built for `claude`, `codex` and `opencode`
- **THEN** there are two hard MCP configuration checks, no endpoint check and no check for OpenCode
