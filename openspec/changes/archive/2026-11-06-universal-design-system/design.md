# Design

## Context

Full currently installs the mockup harness automatically for Claude and Codex; Custom offers it explicitly. OpenCode supports global `AGENTS.md`, skills and local MCP servers, but the catalog excludes it. The OpenCode SDD orchestrator is a separate primary agent in `opencode.json`; the visual rule belongs to all OpenCode agents that might implement UI and therefore belongs in its shared `AGENTS.md` section.

## Decisions

1. Retain a single `design-mockups` config harness. The project-wide design-system gate and per-change mockup gate are one capability; Custom cannot select the latter without the former.
2. Use the same instructions for Claude, Codex and OpenCode, with no Artifact or model-specific command. The existing `active-stack:design-mockups` marker remains the upgrade path.
3. Store the source in project-root `design-system/`: `README.md` (status, revision, approvals, sources and TBD list), `brand.md` (name, logo and imagery decisions), `brand.js` (single displayed-name value for HTML previews), `tokens.css` (CSS custom properties), `components.html` (representative components and states), `assets/` (stable named local resources), `previews/` (mobile/desktop captures). Add only files needed by the project; keep paths stable for replacements. The bundle must work locally without external accounts.
4. During new UI project foundation, run the visual gate after discovery/KB and before frontend roadmap work. If foundation has already passed, run it at the first visual change. Inspect existing brand/UI assets and preserve them for an existing app; propose a new system for a new app. Present the foundation preview and await explicit approval before creating change mockups. Reuse the approved revision later. An approval may list `(TBD)` decisions if the temporary values are marked and replaceable.
5. The change `design.md` records the design-system path/revision and mockup approval separately. Mockups link shared `tokens.css` and stable asset paths where practical. A changed system invalidates only mockups whose appearance is affected; report them and request renewed review.
6. Playwright MCP and the two pinned skills are offered to all three agents in Full; they remain soft dependencies in Custom. If rendering is unavailable, the agent provides local HTML and exact manual steps without claiming a rendered preview.

## Risks / Trade-offs

- An instruction file cannot mechanically force an arbitrary agent to comply. The installed section defines a mandatory gate, and tests verify installation and its content; synthetic agent runs verify behavior. This is not a process-level blocker for an agent that ignores instructions.
- Browser setup varies by machine. Manual preview remains a truthful fallback.
- A complete logo is not always available during foundation. Stable assets and `(TBD)` placeholders allow replacement without remaking the system.

## Migration

Reinstall Full, or select `design-mockups` in Custom. The existing section is replaced idempotently after backup. Existing projects create or audit `design-system/` at the next visual change. Existing approved mockups are reviewed against the new foundation before further visual implementation.
