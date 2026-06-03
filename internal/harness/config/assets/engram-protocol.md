## Engram Persistent Memory — Protocol

You have access to Engram, a persistent memory system that survives across sessions and compactions.
This protocol is MANDATORY and ALWAYS ACTIVE — not something you activate on demand.

### OPSX wiring (orchestrator-specific)

The rest of this protocol applies as-is. This adds ONLY the OPSX-specific wiring:

- **On start**: after recovering memory (`mem_context` / `mem_search "opsx"`), use that context to BRIEF your sub-agents accurately — each sub-agent starts with NO context of its own.
- **After every completed OPSX action** (explore/propose/apply/archive): save with `topic_key: "opsx/{change-name}/{phase}"` and a title like `"OPSX: {action} completed for {change-name}"`, noting the recommended next action. This keeps each change's progress reconstructable across sessions and compactions.

### PROACTIVE SAVE TRIGGERS (mandatory — do NOT wait for user to ask)

Call `mem_save` IMMEDIATELY and WITHOUT BEING ASKED after any of these:
- Architecture or design decision made
- Team convention documented or established
- Workflow change agreed upon
- Tool or library choice made with tradeoffs
- Bug fix completed (include root cause)
- Feature implemented with non-obvious approach
- Notion/Jira/GitHub artifact created or updated with significant content
- Configuration change or environment setup done
- Non-obvious discovery about the codebase
- Gotcha, edge case, or unexpected behavior found
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Self-check after EVERY task: "Did I make a decision, fix a bug, learn something non-obvious, or establish a convention? If yes, call mem_save NOW."

Format for `mem_save`:
- **title**: Verb + what — short, searchable (e.g. "Fixed N+1 query in UserList")
- **type**: bugfix | decision | architecture | discovery | pattern | config | preference
- **scope**: `project` (default) | `personal`
- **topic_key** (recommended for evolving topics): stable key like `architecture/auth-model`
- **content**:
  - **What**: One sentence — what was done
  - **Why**: What motivated it (user request, bug, performance, etc.)
  - **Where**: Files or paths affected
  - **Learned**: Gotchas, edge cases, things that surprised you (omit if none)

Topic update rules:
- Different topics MUST NOT overwrite each other
- Same topic evolving → use same `topic_key` (upsert)
- Unsure about key → call `mem_suggest_topic_key` first
- Know exact ID to fix → use `mem_update`

### WHEN TO SEARCH MEMORY

On any variation of "remember", "recall", "what did we do", "how did we solve", or references to past work:
1. Call `mem_context` — checks recent session history (fast, cheap)
2. If not found, call `mem_search` with relevant keywords
3. If found, use `mem_get_observation` for full untruncated content

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- User mentions a topic you have no context on
- User's FIRST message references the project, a feature, or a problem — call `mem_search` with keywords from their message to check for prior work before responding

### SESSION CLOSE PROTOCOL (mandatory)

Before ending a session or saying "done" / "listo" / "that's it", call `mem_session_summary`:

## Goal
[What we were working on this session]

## Instructions
[User preferences or constraints discovered — skip if none]

## Discoveries
- [Technical findings, gotchas, non-obvious learnings]

## Accomplished
- [Completed items with key details]

## Next Steps
- [What remains to be done — for the next session]

## Relevant Files
- path/to/file — [what it does or what changed]

This is NOT optional. If you skip this, the next session starts blind.

### AFTER COMPACTION

If you see a compaction message or "FIRST ACTION REQUIRED":
1. IMMEDIATELY call `mem_session_summary` with the compacted summary content — this persists what was done before compaction
2. Call `mem_context` to recover additional context from previous sessions
3. Only THEN continue working

Do not skip step 1. Without it, everything done before compaction is lost from memory.

### GIT SYNC — version project memory in its repo

Engram can version **this project's** memory in its git repo, so it travels with the code and is recoverable on another machine or by the team. The CLI already ships the mechanism:

- `engram sync` → exports NEW observations **for the project of this directory** to `.engram/chunks/*.jsonl.gz` (compressed). Filters by the `project_name` in `.engram/config.json`.
- `engram sync --import` → imports the chunks in `.engram/` into the local DB.

**Workflow (do it as part of the push, without being asked):**

- **When pushing** a repo that versions `.engram/`: BEFORE `git push`, run `engram sync`, then `git add .engram` and commit the chunks (own commit `chore(engram): sync project memory`, or folded into the push commit).
- **When cloning or `git pull`** of a repo with `.engram/`: run `engram sync --import` to recover that memory into your local DB.

**⚠️ Privacy guardrail — NON-negotiable:**

- **NEVER `engram sync --all`** in this flow. Engram's DB is global and cross-project — `--all` would leak EVERY project's memory into a single repo. Always use the per-directory filter (plain `engram sync`, or `--project <name>`).
- **Before versioning `.engram/` in a PUBLIC repo**, confirm with the user: project memory becomes visible to anyone. If it is sensitive, do NOT version it.
- `.engram/` must not be in `.gitignore` if you want to version it (check: `git check-ignore .engram`).

**Optional shortcut** — a git alias that joins both steps into one command:

```bash
git config alias.sync '!f() { if [ "$1" = "--import" ]; then git pull --ff-only && engram sync --import; else engram sync && git add .engram && { git diff --cached --quiet || git commit -m "chore(engram): sync project memory"; } && git push; fi; }; f'
```

With that: `git sync` exports + commits + pushes the project memory; `git sync --import` pulls changes + imports into your local DB.
