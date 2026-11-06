# Spec Delta

## Purpose

Defines the `playwright` external harness: it registers Microsoft's official Playwright MCP server, pinned to an exact version, as a local stdio MCP server for Claude so the agent can drive a real browser (for example, for visual verification against approved mockups).

## ADDED Requirements

### Requirement: The catalog declares a Full-only, Claude-only `playwright` harness

The embedded catalog SHALL contain a global-picker harness with id `playwright`, type `external`, install method `mcp`, supported agents limited to `claude`, and install modes limited to `full`. It MUST NOT declare a dependency on any other harness. Its `description` and `long_description` MUST be present in both `es` and `en`, and the catalog MUST still load under the existing validation.

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog is loaded and `playwright` is resolved by ID
- **THEN** loading succeeds, its type is `external` with method `mcp`, it supports `claude` only, it is in mode `full`, and it has no dependencies

#### Scenario: Included in Full, excluded from Lite

- **WHEN** the harness list is built for mode `full` and for mode `lite`
- **THEN** the Full list contains `playwright` and the Lite list does not

#### Scenario: Not offered for non-Claude agents

- **WHEN** the harness list is built for the `opencode` agent only (or `codex` only)
- **THEN** the list does not contain `playwright`

### Requirement: The MCP server spec is local stdio and pinned to an exact version

The `playwright` harness SHALL declare a local stdio MCP server named `playwright` launched with `npx` and the package `@playwright/mcp` at an exact semantic version (`MAJOR.MINOR.PATCH`). The arguments MUST NOT reference a floating tag such as `@latest` or `@next`. The harness MUST NOT declare a remote URL.

#### Scenario: Version is pinned

- **WHEN** the `playwright` harness MCP arguments are inspected
- **THEN** exactly one argument names `@playwright/mcp@` followed by an exact `MAJOR.MINOR.PATCH` version, and no argument contains `@latest` or `@next`

#### Scenario: No remote endpoint

- **WHEN** the `playwright` harness is inspected
- **THEN** it has no remote URL and its MCP server name is `playwright`

### Requirement: Installation registers the server only for Claude

Installing `playwright` SHALL register a stdio server entry named `playwright` (command `npx` and the pinned arguments) in Claude's MCP configuration file, taking a backup of the existing file before writing. Agents the harness does not support MUST NOT be modified, even when they are selected in the same install: their MCP configuration files stay byte-for-byte unchanged.

#### Scenario: Claude receives the server entry

- **WHEN** `playwright` is installed with the `claude` agent
- **THEN** Claude's MCP configuration contains a stdio server `playwright` with command `npx` and the pinned arguments, and a backup of the previous file exists when one existed

#### Scenario: Other selected agents are untouched

- **WHEN** `playwright` is installed with agents `claude` and `opencode`
- **THEN** the opencode MCP configuration is identical to its state before the install and contains no `playwright` entry

#### Scenario: Existing MCP servers are preserved

- **WHEN** `playwright` is installed into a Claude MCP configuration that already contains other servers (for example `context7`)
- **THEN** those entries are unchanged and the `playwright` entry is added

### Requirement: Installation is idempotent

Re-installing `playwright` SHALL NOT duplicate the server entry. When the entry is already up to date, the configuration file MUST NOT be rewritten.

#### Scenario: Re-install reports no change

- **WHEN** `playwright` is installed twice for `claude`
- **THEN** the configuration contains exactly one `playwright` entry and the second install reports it as already installed

### Requirement: Agent scoping does not change harnesses without an agent list

External harnesses that declare no supported-agent list SHALL keep being installed for every selected agent, exactly as before this change.

#### Scenario: Agent-agnostic external still installs for all selected agents

- **WHEN** an external MCP harness with no agent list (such as `context7`) is installed with agents `claude` and `opencode`
- **THEN** both agents' MCP configurations receive its entry

### Requirement: Post-install verification covers only Claude and skips endpoint checks

Post-install verification for `playwright` SHALL produce a hard check that Claude's MCP configuration exists and is valid. It MUST NOT produce a remote-endpoint reachability check, and it MUST NOT produce checks for agents the harness does not support.

#### Scenario: Verify checks for a Claude and OpenCode install

- **WHEN** verification checks are built for `playwright` with agents `claude` and `opencode`
- **THEN** there is exactly one MCP configuration check (for `claude`), no endpoint check, and no check for `opencode`
