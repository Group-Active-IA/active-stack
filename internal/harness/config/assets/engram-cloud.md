## Cloud Team Memory — extends the Engram Protocol above

Applies ONLY when the project has `.engram/active-stack.json` with `"memory_policy": "cloud-branch"`.
Otherwise ignore this whole section.

It extends the Engram Protocol block above and OVERRIDES: proactive-save routing, the
`opsx/{change-name}/{phase}` keys and the "GIT SYNC" section. Never use Git Sync (`.engram/chunks`).

### Saving
You save memories as usual. The author line (`Author: <name>`) is added for you, and so is the branch
tag when you are not on the default branch: do NOT add tags or the author yourself and do NOT save
under a different project.

### Reading
- Untagged result (no `branch/` key, no `[slug]` in the title) = merged team knowledge. Trust it.
- `branch/<current slug>/…` = valid for this branch. The current slug comes with the session context;
  if it is not there, treat every `branch/` memory as unverified.
- `branch/<other slug>/…` = UNVERIFIED work in progress of another branch. Never treat it as fact;
  ignore it unless the user asks.
- If an untagged and a tagged memory contradict, the untagged one wins until the branch is merged.

### Closing a unit of work
Only after git confirms the branch is merged into the default branch, and only for YOUR memories
(`Author:` = you). Promote ONLY durable ones (decision, architecture, pattern, bugfix with a
reusable root cause, config gotcha): search the target `topic_key` first; if free, `mem_update` its
`topic_key` and title to the untagged ones; if it exists, update the existing one. Do NOT delete
or touch ephemeral memories (opsx/*, session summaries, next steps).

### Forbidden
Writing tokens or secrets in a memory · `engram sync --all` · Git Sync · enrolling a project other
than the one in `.engram/config.json` · promoting before the merge.
