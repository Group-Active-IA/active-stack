# Proposal

## Why

The approved mockup stage, its two design skills and browser verification are restricted to Claude. Codex already has instruction, skill and MCP adapters, so UI changes made there should have the same review and approval workflow.

## What Changes

- Offer `design-mockups`, `emil-design-eng`, `design-taste-frontend-v1` and `playwright` to Codex as well as Claude.
- Inject a Codex-specific design section into its `AGENTS.md`. It creates local HTML/CSS mockups, records the decision and approval, and verifies the implemented UI against approved mockups.
- Use local responsive HTML/CSS as the portable review bundle for both agents. Claude may additionally use Artifact Design when available; Figma or Stitch may provide references or editable copies. Render mobile and desktop previews, show the captures and local files to the user, and require explicit approval of that revision before `apply`.
- Trigger the design step automatically for visual changes while the harness is installed. Keep design skills and Playwright as soft dependencies; an unavailable browser requires an honest manual-preview path.
- Update the catalog, tests and documentation. Impeccable was not selected in the original proposal; evaluating it as a separate optional harness remains `(TBD)`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `design-mockups-harness`: extend supported agents and define the shared review, presentation, approval and verification contract.
- `design-skills-harnesses`: offer the two existing pinned skills to Codex.
- `playwright-mcp-harness`: register the existing pinned MCP for Codex.

## Impact

`internal/catalog`, `internal/harness/config`, existing install and verification tests, and project documentation. No changes to backup, filemerge or pipeline.
