# Tasks

## 1. Codex design stage

- [x] 1.1 Add a failing install test for Codex-specific instructions, then implement the variant fragment and verify idempotent installation.
- [x] 1.2 Document Claude/Codex behavior in AGENTS.md and CLAUDE.md and verify the paragraphs match.

## 2. Companion harnesses

- [x] 2.1 Add failing catalog/picker tests for Codex, extend the four harness agent lists, and verify `catalog.Load()` and the tests pass.
- [x] 2.2 Verify Playwright installs idempotently in Codex's TOML MCP config with an integration test.
- [x] 2.3 Update README.md, ARCHITECTURE.md and CHANGES.md and verify the descriptions match the catalog.

## 3. Integration

- [x] 3.1 Run `go test ./...` and report the three Windows-specific `internal/system` test failures by name.
- [ ] 3.2 Exercise the complete design approval and visual verification flow in a real destination UI project; record the observed result before archiving this change.

## 4. Portable visual review

- [x] 4.1 Define one HTML/CSS review bundle and mobile/desktop preview contract for Claude and Codex; make Artifact/Figma/Stitch optional.
- [x] 4.2 Require visible presentation of previews, explicit approval and reapproval after revisions before apply.
- [x] 4.3 Update the catalog and project documentation for the shared behavior.
- [ ] 4.4 Smoke-test the preview, presentation and approval gate with both agents in a real UI project.
