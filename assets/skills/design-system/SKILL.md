---
name: design-system
description: >-
  Define or audit a project's visual identity during Active Orchestrator foundation,
  or when an older project first needs UI work. Ask once whether the project wants
  a design system with per-change mockups, a design system alone, or neither;
  preserve that choice. Use for the shared foundation, not for an individual
  screen mockup (design-mockups handles those).
license: Apache-2.0
metadata:
  version: "1.0"
---

# Project design system

Create one portable visual foundation for the project. The user's choice to
decline mockups is a durable project policy, not an invitation to ask again at
every change. An approved system is reused by later mockups and UI work.

## When to run

- Active Orchestrator invokes this after KB/discovery and before the frontend
  roadmap for a new UI project.
- In an already founded project, invoke it once before the first visual change
  if no visual policy is recorded. Do not rerun completed foundation phases.
- A direct user request to revise the visual policy or system also invokes it.
- Per-change HTML mockups belong to `design-mockups`, not this skill.

Read `.active-orchestrator-state.json` when present, `knowledge-base/`, existing
UI assets and `design-system/README.md`. Do not overwrite an existing identity.

## 1. Resolve the project policy once

If `visual_design.mode` is already recorded in the state JSON, or `Mode:` is
recorded in `design-system/README.md`, honor it. Do not ask again unless the user
explicitly wants to change it. The exception is `deferred`: at the **first
visual change**, replace it with one of the three durable choices below, after
asking the user once. `deferred` is not a permanent opt-out. If both records
disagree, stop, show the disagreement and ask which policy is current before
writing.

If the project has no UI plans yet, record `deferred` and continue the other
foundation phases. At its first visual change, offer the choice below once.

Otherwise present these three project-wide choices, recommending the first:

1. `system-and-mockups`: define the design system now; review a mockup for each
   later visual change before implementation.
2. `system-only`: define the design system now; implement later visual changes
   without routine mockup requests.
3. `off`: skip both visual foundation and routine mockup requests.

Wait for the user's choice. Do not treat silence as approval. `system-only` and
`off` are valid outcomes. In either case, write the choice so a future agent
does not ask again at each propose. An explicit later request can change it.

## 2. Preserve foundation state

The project-root `.active-orchestrator-state.json` has an additive section
owned by this skill:

```json
"visual_design": {
  "created_by": "design-system",
  "mode": "system-and-mockups|system-only|off|deferred",
  "status": "pending|approved|skipped|deferred",
  "revision": 1,
  "path": "design-system/README.md"
}
```

Record `revision` only for an existing design system. A user-approved system
uses `status: approved`; `off` uses `skipped`; no planned UI uses `deferred`.
For older state schemas (versions 2, 3 or 4), preserve `kb`, `roadmap`, `skills`,
`registry`, `agents`, `discovery`, `step`, and unknown fields. Set top-level
`version: 5` when adding `visual_design`; this is the one migration exception
to the orchestrator's normal version ownership, needed when this skill runs
standalone from an older project's first visual change. Do not reset
`step: done`, rerun finished phases, or label an old project approved without
the user's approval. Active Orchestrator normally owns `version` and `step`;
this skill owns `visual_design` and may upgrade `version` only as stated above.
Before changing an existing JSON file,
save a backup; write valid JSON atomically and reread it. If no state file
exists, record the policy in `design-system/README.md` without inventing a
completed foundation state.

For `off` or `deferred`, create or update only `design-system/README.md` with
`Mode: off|deferred`, `Approval: skipped|deferred`, the decision source and date.
Do not create provisional branding for a project that opted out.

## 3. Build or audit the visual foundation

For `system-and-mockups` and `system-only`, derive a brief from the KB, audience,
tasks, platforms and supplied references. In an existing app, inspect its actual
logo, palette, typography, components and motion first. Keep established choices
unless the user asks for a rebrand. Apply installed design skills such as
`emil-design-eng` and `design-taste-frontend-v1` when useful; their absence
does not block this phase.

Create or audit these project-root files:

- `design-system/README.md`: `Mode:`, `Revision:`, `Approval: pending`, sources,
  open `(TBD)` decisions and exact replacement instructions.
- `design-system/brand.md`: product name, tone, logo and imagery rules with
  source citations or `(TBD)`.
- `design-system/brand.js`: one name constant and a binding for
  `[data-brand-name]` in local HTML previews.
- `design-system/tokens.css`: named color, type, spacing, radii, shadow, focus
  and motion tokens, including reduced-motion behavior.
- `design-system/components.html`: representative components and states that
  import the shared tokens and local assets.
- `design-system/assets/logo.svg`: a stable logo path. A placeholder must be
  visibly labeled provisional.
- `design-system/previews/`: mobile and desktop captures when rendering works.

Keep name, logo and visual values at those stable paths so replacing `(TBD)`
details does not require editing every mockup. Do not require Figma, a remote
font, a paid account or a Claude-only tool.

## 4. Preview and obtain approval

Render `components.html` at 390×844 and 1440×900 in an available browser.
Playwright MCP may take screenshots; if file navigation is blocked, serve the
folder on localhost and stop the server afterward. Inspect both captures and
fix obvious defects. Show the actual images and link the local HTML file. If
rendering is unavailable, provide opening steps and say inspection was manual.

Wait for explicit approval before setting `Approval: approved (<date>, by user)`
in the README and `visual_design.status: approved` in state. A usable system
may be approved with visibly provisional `(TBD)` name or assets. If the user
requests changes, keep approval pending, revise and show the previews again.
Changing a shared token or asset later bumps `Revision:`, resets approval to
pending and identifies affected mockups for renewed review.

## Handoff

Return the selected mode, README path, revision/status, preview paths and open
decisions to Active Orchestrator. It advances its `step` only through its normal
checkpoint. For `system-and-mockups`, the later `design-mockups` gate reuses this
approved revision. For `system-only` and `off`, that gate silently respects the
recorded policy until the user explicitly changes it.
