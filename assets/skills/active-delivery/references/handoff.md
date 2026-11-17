# Delivery handoff

Use one Markdown record for each unit that needs ODD continuity or a confirmed
promotion. Save it at `knowledge-base/entrega/<id>.md`. Use the stable roadmap
ID when available. If the ID is unsuitable as a filename, agree on a safe,
unique basename and retain the original ID inside the record. Keep the file
under the project root. Check ownership and identity before updating an
existing record, and merge additions without discarding earlier decisions.

Capture current Git evidence with `git rev-parse HEAD`, `git status --short`,
and staged and unstaged diff summaries (`git diff --cached --stat` and
`git diff --stat`). Include relevant untracked paths separately. Capture the
branch or detached state. If no Git base is available, record why and include
the relevant local file paths. Distinguish pre-existing edits from work for
this ID. Keep sensitive content out of the handoff and refer to safe project
paths and results. Use Engram as an additional pointer if available and
appropriate. The project record remains sufficient to resume without Engram.

## Template

```markdown
# <ID> · <work title>

- Roadmap reference or local ID
- Current mode and status
- OpenSpec change name or pending promotion, when relevant

## Scope and acceptance

- Confirmed outcome and boundaries
- Acceptance criteria and verification plan

## Decisions

- Confirmed decision, rationale and source
- Open decision or governing approval pending

## Work and Git context

- Branch, worktree and HEAD at the time of capture
- Paths changed for this ID and relevant staged, unstaged or untracked changes
- Pre-existing edits to preserve
- Remaining actions and dependencies

## Verification

- Command or manual check, checked scope and outcome
- Evidence location and Git or working tree context at execution
- Stale, failed or unrun checks and what needs to be repeated

## Promotion

- Trigger, approval and intended OpenSpec action
- Existing change reference and authoritative status, once checked
```

On resumption, compare current work with the recorded evidence. An unchanged
HEAD alone does not validate tests against edited files. Preserve historical
verification results while marking stale evidence and rerun checks affected by
the current diff.
