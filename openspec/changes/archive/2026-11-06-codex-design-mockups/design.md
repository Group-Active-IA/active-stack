# Design

## Context

See proposal.md. The config installer already injects owned sections with backup and idempotent markers. The Codex adapter resolves `AGENTS.md`, `.agents/skills` and `.codex/config.toml`. The Claude fragment calls Artifact Design and names `sonnet`, so it cannot be sent unchanged to Codex.

## Goals / Non-Goals

**Goals:** Give Claude and Codex the same portable, visibly reviewable local mockup flow, with explicit approval before implementation and visual verification afterward.

**Non-Goals:** Installing Impeccable, changing the two pinned skill sources, changing the backup/merge/pipeline code, or publishing mockups to a remote design service.

## Decisions

1. Keep one `design-mockups` harness and marker, with a fragment override keyed by adapter variant. This keeps selection, uninstall and verification behavior shared while allowing Claude-specific Artifact instructions. Duplicating the harness would split one capability and its approval state.
2. Both agents write self-contained responsive HTML/CSS under the change's `mockups/` directory and render mobile and desktop captures under `mockups/previews/`. The response presents the captures and HTML paths before requesting approval. Figma, Stitch and Artifact are optional sources or secondary editors; no external account or tool is required to inspect the local bundle. When browser capture is unavailable, the agent must give exact manual-open instructions and report the missing visual check.
3. When the harness is installed, visual work automatically enters design after propose. Only an explicit user request skips it. A supplied design enters the same approval and verification flow. Revisions reset approval to pending; approved artifacts are acceptance criteria and cannot be silently changed.
4. Expand the existing two pinned design skills and Playwright MCP to Codex through their `agents` lists. Their installers already use adapter paths and agent scoping. The design skills and browser remain soft dependencies; Full installs them together.
5. Treat Impeccable as a separate candidate `(TBD)`. The archived proposal chose Emil and Taste v1 independently and does not mention Impeccable. Adding its current multi-command package would require a separate source, pin and install-contract review.

## Risks / Trade-offs

- [HTML may not capture every interaction] → list interaction states in the brief and README; use browser preview to show representative states.
- [Browser MCP unavailable] → the fragment provides local HTML with exact opening instructions; it must not claim visual inspection. `verify` records `Visual verify: manual`.
- [Pinned third-party skills may not match later upstream versions] → keep existing SHAs; bump only after reviewing upstream changes.

## Migration Plan

Reinstall Full or select the harnesses in Custom for Codex. The existing snapshot and marker flow updates an installed section in place. Rollback uses the existing uninstall/snapshot path. Claude's fragment and source pins stay unchanged.
