# Proposal

## Why

The existing mockup stage can choose a new visual direction for every change because no approved, project-wide visual source exists. It is available to Claude and Codex, but OpenCode is excluded. A Full installation should give all three supported agents one consistent, reviewable design workflow.

## What Changes

- Extend `design-mockups`, Playwright MCP, and the existing optional design skills to Claude, Codex, and OpenCode.
- Use one portable HTML/CSS workflow for every agent. Claude Artifact Design is no longer part of the workflow.
- Before the first visual mockup in a project, create or audit `design-system/`, show a visual foundation preview, and record the user's explicit approval. A later change reuses that approved version.
- Keep brand decisions such as name and logo as `(TBD)` when unresolved, with a stable replaceable path or token and a visibly provisional preview. The foundation can be approved with recorded TBD items.
- Link each change's mockups to the approved design-system revision. A revision of the shared system must identify affected mockups and trigger review where their appearance changes.
- Keep Full automatic and Custom selectable. Lite remains unchanged.

## Capabilities

### Modified Capabilities

- `design-mockups-harness`: project-wide visual foundation and mandatory approval gate before per-change mockups; three supported agents; one HTML/CSS contract.
- `design-skills-harnesses`: offer the two pinned optional skills to OpenCode.
- `playwright-mcp-harness`: offer the browser MCP to OpenCode.

## Impact

Catalog, config assets, tests, verification, docs. No changes to backup, filemerge, or pipeline. The installer supplies instructions and tools; each project's agent run creates the project-specific design system.
