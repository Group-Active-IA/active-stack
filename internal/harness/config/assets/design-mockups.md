## Design Stage — UI Mockups

For visual changes, run: explore → propose → **design system** → **design** → apply → verify → archive.
Visual work includes a new or changed page, screen, component, layout, style or interaction.
Run the foundation and mockup gates automatically when the project policy enables them.
Before creating change mockups, read `visual_design.mode` in project-root
`.active-orchestrator-state.json` and `Mode:` in `design-system/README.md`.
The `design-system` skill owns that policy and the system creation workflow:

- `system-and-mockups`: require an approved system, then create and present a
  mockup for each visual change before apply.
- `system-only`: require an approved system for UI work, but do not ask again
  about mockups on each change. If change `design.md` exists, record
  `Decision: no (project policy)` there.
- `off`: continue visual changes without either gate and do not ask again
  about mockups on each change. If change `design.md` exists, record
  `Decision: no (project policy)` there.
- `deferred` or no recorded mode: invoke the installed design-system skill at
  the first visual change to ask once for the project-wide policy. Preserve
  older foundation state and completed phases.

If the JSON and README modes disagree, stop and reconcile with the user.
Only an explicit request from the user may revise the project policy or skip one mockup without
changing it. Do not treat a per-change skip as a permanent opt-out.
For nonvisual work, record `Decision: no (no UI)` and continue.

### Project-wide design system gate

1. For `system-and-mockups` or `system-only`, read project-root `design-system/README.md` and
   check its `Revision:` and `Approval:` fields. An absent or unapproved system
   blocks UI apply in these two modes. A Full install provides this workflow; it does not
   create a brand for every project at install time. Active Orchestrator
   invokes the design-system skill after discovery/KB and before planning
   frontend changes. If that point has passed, invoke it at the first visual change.
2. For a new project, make a brief from the product, audience, tasks, platform,
   references and constraints. For an existing app, first audit its actual
   logo, colors, typography, components and motion; preserve established
   choices unless the user requests a rebrand. Mark unknown facts `(TBD)`.
   Load relevant installed design skills such as `emil-design-eng` and
   `design-taste-frontend-v1` when available; these are soft dependencies.
3. Create the portable foundation at project root:
   - `design-system/README.md`: purpose, source references, `Revision: 1`,
     `Approval: pending`, open `(TBD)` decisions and replacement instructions.
   - `design-system/brand.md`: product name, logo, imagery, visual tone and
     usage rules, each with source or `(TBD)`.
   - `design-system/brand.js`: one local brand-name constant and a tiny binding
     for `[data-brand-name]` in HTML previews, so renaming does not require
     editing every mockup. Keep the script self-contained and accessible.
   - `design-system/tokens.css`: named CSS variables for color, typography,
     spacing, radii, shadows, focus and motion, with reduced-motion behavior.
   - `design-system/components.html`: local HTML/CSS reference for representative
     components and states, importing `tokens.css` and local assets.
   - `design-system/assets/`: stable local paths such as `logo.svg`. A provisional
     logo must visibly say it is a placeholder, not imply an approved brand.
   - `design-system/previews/`: rendered mobile and desktop captures.
   Add other files only when the project needs them. Do not require a design
   service, remote font, third-party account or Claude-only tool.
4. Make unresolved name, logo and imagery replaceable at a single stable path
   or named token (`brand.js` for displayed name, `assets/logo.svg` for the
   logo, `tokens.css` for visual values). Record exact replacement steps in
   `README.md`. The user may
   approve a usable foundation with `(TBD)` items; keep them visibly provisional
   in previews and later mockups. Never invent brand facts as if final.
5. Render `components.html` in an available browser at 390x844 and 1440x900.
   If Playwright MCP is installed, use it to open the local file (or serve the
   folder on localhost if file navigation is blocked), resize and call
   `browser_take_screenshot` with filenames in `design-system/previews/` (full
   page when supported). Stop a preview server you started. Inspect the images
   and fix obvious defects. Show the actual captures and a clickable local HTML
   link to the user. If rendering is unavailable, give the HTML and exact manual
   opening steps; say that visual inspection was not automated.
6. Wait for explicit user approval of this foundation before creating change
   mockups. Record `Approval: approved (<date>, by user)` in `README.md`;
   otherwise keep `Approval: pending`. A requested revision resets it to pending.
   Reuse the approved `Revision:` for later visual changes; do not regenerate
   the foundation for each screen. When a shared token or asset changes, bump
   the revision, identify affected mockups, rerender them and ask for renewed
   review wherever their appearance changes.

### Mockup gate for each visual change

Run steps 7–12 only in `system-and-mockups` mode. In `system-only` or `off`,
record `Decision: no (project policy)` in the change design document when it
exists and continue without a mockup gate. Do not create an otherwise
unnecessary design document solely to repeat the opt-out.

7. Read the approved design-system revision and the change proposal, specs,
   design and tasks. Record `## UI Mockups` in
   `openspec/changes/<name>/design.md` with:
   - `Design system revision: <revision from design-system/README.md>`.
   - `Decision: yes | provided | no` and date. In `system-and-mockups`, visual
     `no` requires the user's explicit per-change skip request; a user-supplied
     design uses `provided`.
   - `Mockups: <path to mockups/>` and supplied source references.
   - `Approval: pending | approved (<date>, by user)`.
   - `Preview: rendered | manual (<date>, paths)`.
   - `Visual verify: passed | differences | manual (<date>)` after verify.
   Respect a recorded approval for the same revision; do not ask again.
8. Write a short design brief covering audience, task, affected screens and
   states, responsive sizes, approved tokens/assets, constraints and supplied
   references. Use an installed design skill when relevant. If its advice
   conflicts with the approved foundation, follow the foundation or propose
   a foundation revision for user review.
9. Under `openspec/changes/<name>/mockups/`, create local responsive HTML/CSS
   for each affected screen and state, local assets only when needed, and a
   README listing screens, sources, preview steps and `Approval: pending`.
   Import the shared `design-system/tokens.css` and `brand.js`, and reference
   stable assets rather than copying values when practical. Clearly mark
   `(TBD)` placeholders.
   A supplied Figma or other design can be a reference; the local HTML remains
   the common review format for Claude, Codex, OpenCode and Antigravity.
10. Render each mockup at 390x844 and 1440x900, using Playwright MCP when
    installed. Save inspected captures in `mockups/previews/`, present images
    and clickable HTML files, and record `Preview: rendered`. If no browser is
    available, give manual opening steps, record `Preview: manual`, and do not
    claim rendered inspection. A URL alone is not proof the user saw the design.
11. Do NOT begin apply until the user explicitly approves the shown mockup
    revision and `Approval: approved` is recorded. This is the deliberate
    exception to "never block on phase gates". A revision resets approval to
    pending. Link approved mockups and captures from `design.md` and `tasks.md`
    as acceptance criteria. Never silently change an approved mockup.
12. During verify, compare every implemented screen with its approved mockup
    and the recorded design-system revision. With a browser, start the dev
    server, capture mobile and desktop, inspect console errors and failed
    network requests, save captures under `mockups/verify/`, then stop the
    server. Report differences by priority (blocker / major / minor), screen,
    viewport, expected mockup, observed result and reproduction steps. List
    checked and unchecked screens. Never claim a match without opening it.
    Without a browser, report that visual verification was NOT automated,
    provide a screen-by-screen manual checklist, suggest the `playwright`
    harness, and record `Visual verify: manual`. The browser is a soft
    dependency; design-mockups still works with manual review.
