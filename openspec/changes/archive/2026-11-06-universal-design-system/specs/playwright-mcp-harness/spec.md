# Spec Delta

## MODIFIED Requirements

### Requirement: The catalog declares a Full-only, Claude-only `playwright` harness

The catalog SHALL offer the pinned local stdio `playwright` harness in Full and Custom for Claude, Codex and OpenCode. Lite SHALL exclude it. The harness remains a soft dependency of `design-mockups`.

#### Scenario: OpenCode Full selection

- **WHEN** Full harnesses are selected for OpenCode
- **THEN** `playwright` is included with an MCP command of `npx` and a pinned package version

#### Scenario: Catalog resolves the harness and loads

- **WHEN** the catalog loads `playwright`
- **THEN** it is a Full-only external MCP harness supporting Claude, Codex and OpenCode with no hard dependency

#### Scenario: Included in Full, excluded from Lite

- **WHEN** the Full and Lite lists are built
- **THEN** Full contains `playwright` and Lite does not

#### Scenario: Not offered for non-Claude agents

- **WHEN** the agent-specific list is built for OpenCode or Codex
- **THEN** it offers `playwright`, while unsupported agents do not receive it

### Requirement: Installation registers the server only for Claude

Installation SHALL register Playwright as a local stdio server for all three supported agents. OpenCode's `opencode.json` entry SHALL use `type: local` with a combined command array, preserving unrelated MCP entries. It SHALL NOT register an empty remote URL.

#### Scenario: OpenCode local MCP

- **WHEN** Playwright is installed for OpenCode
- **THEN** its `mcp.playwright` entry is enabled, local and runs pinned `npx @playwright/mcp`, while other MCP entries remain present

#### Scenario: Claude receives the server entry

- **WHEN** Playwright is installed for Claude
- **THEN** its MCP config contains the pinned local stdio entry after backing up an existing config

#### Scenario: Other selected agents are untouched

- **WHEN** Playwright is installed with Claude, Codex, OpenCode and an unsupported agent
- **THEN** only the first three receive the server entry

#### Scenario: Existing MCP servers are preserved

- **WHEN** Playwright is installed into a supported agent's config containing other MCP servers
- **THEN** those entries remain and Playwright is added

#### Scenario: Codex receives the server entry

- **WHEN** installed with `codex`
- **THEN** Codex's TOML MCP config contains a single pinned `playwright` stdio entry and reinstalling reports no change

### Requirement: Post-install verification covers only Claude and skips endpoint checks

Post-install verification SHALL check a local MCP configuration entry for Claude, Codex and OpenCode and SHALL NOT run a remote endpoint check for Playwright.

#### Scenario: Three-agent verification plan

- **WHEN** verification checks are built for Claude, Codex and OpenCode
- **THEN** there is one hard MCP config check per agent and no endpoint check

#### Scenario: Verify checks for a Claude and OpenCode install

- **WHEN** verification checks are built for Playwright with Claude and OpenCode
- **THEN** each receives a hard local config check and neither receives a remote endpoint check
