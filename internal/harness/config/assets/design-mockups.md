## Design Stage — UI Mockups (Claude)

Extends the OPSX workflow above. Flow with this stage:
explore → propose → **design** → apply → verify → archive.
The design stage runs AFTER propose (the specs define the screens) and BEFORE apply. Never during apply.

1. Detect UI: if the change's proposal/specs/design/tasks mention pages, screens,
   components, forms or other user-interface work → ask the user
   "Does this change need mockups?" (yes / no / I already have them).
   No UI signal → do not ask; record `Decision: no`.
   If `design.md` already has a `## UI Mockups` decision → honour it, never re-ask.
2. Record the outcome in `openspec/changes/<name>/design.md` under `## UI Mockups`:
   - `Decision: yes | no | provided` (+ date)
   - `Mockups: <path to mockups/ and/or Artifact URL>`
   - `Approval: pending | approved (<date>, by user)`
3. yes → delegate to a general-purpose sub-agent (it has the Artifact tool): call
   Artifact with action "quickstart" and intent "design"; create the artifact from
   the returned `type_url` (never hardcode it; add a title and auto_open
   "after_first_write") and follow the instructions the creation result returns.
   Write `mockups/README.md` with the artifact URL, the screen list and the approval
   state, and save an exported copy into `mockups/` when the tool allows it.
   Artifact unavailable → write self-contained static HTML mockups into `mockups/`.
   provided → copy or reference the user's mockups into `mockups/`.
4. Approval gate (a deliberate exception to "never block on phase gates"): when the
   decision is yes or provided, do NOT start apply until the user explicitly approves
   the mockups and `Approval: approved` is recorded. Present them, ask, then record.
   If the user asks for changes, iterate on the same artifact (republish by its url)
   and set the approval back to `pending`. Decision `no` never gates apply.
5. Store mockups under `openspec/changes/<name>/mockups/` and reference them from
   `design.md` and `tasks.md`. In apply and verify they are visual acceptance criteria.
6. If design skills such as `emil-design-eng` (Emil Kowalski) or a taste skill are
   installed (if installed; this is a soft dependency, the stage works without them),
   load them for the design step.
Model: if the Model Assignments table is present, use design → sonnet for this stage.
Artifact URLs are private until the user shares them.
