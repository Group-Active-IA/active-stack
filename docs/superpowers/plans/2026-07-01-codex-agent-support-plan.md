# Codex Agent Support Implementation Plan

## 1. Adapter and routing

- Add failing tests for Codex machine/project paths and adapter contracts.
- Add `internal/agents/codex`.
- Register Codex in the adapter factory and default registry.
- Extend TUI/config detection and associated tests.
- Update the Codex asset skill path.

## 2. Catalog compatibility

- Add failing catalog assertions for Codex-compatible harnesses.
- Add Codex only to harnesses with a valid materialization.
- Run `catalog.Load()` tests after every catalog edit.

## 3. TOML merge foundation

- Add `go-toml/v2` for syntax validation.
- Add failing tests for section-aware MCP table upsert/removal and top-level
  permission mutations.
- Preserve unrelated content, comments, CRLF tolerance, nested tables, and
  Windows paths.
- Make every merge idempotent and validate the result before returning it.

## 4. MCP and permissions integration

- Add a TOML MCP strategy to model/external adapter contracts.
- Route machine and project Codex MCP writes through the TOML merge path.
- Add Codex permission overlays for strict, balanced, and bypass tiers.
- Ensure snapshots precede writes and rollback restores originals.

## 5. Verify and uninstall

- Verify Codex marker sections, skills, TOML permissions, and MCP tables.
- Remove owned MCP tables during targeted uninstall.
- Restore permission values only when safe provenance exists; otherwise
  preserve them and report a warning.
- Keep complete snapshot restore behavior unchanged.

## 6. Validation

- Run focused package tests while implementing.
- Run `go test ./...`.
- Run `git diff --check`.
- Inspect the final diff for unrelated changes and branding leftovers.
- Do not run `go build` and do not commit.

