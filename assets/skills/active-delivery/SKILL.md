---
name: active-delivery
description: >-
  Choose ODD or SDD for one unit of project work. Resume ODD from project evidence,
  or carry that evidence into the existing OpenSpec flow with user approval.
license: Apache-2.0
metadata:
  version: "1.0"
---

# Active delivery

Use this skill at the start or resumption of a unit of work, or when its scope
changes enough to reconsider the delivery mode. Foundation defines the project
setup. This skill chooses the delivery mode for the current work. Follow project
governance, authorization, TDD and verification requirements in either mode.

## Choose the delivery mode

Read the relevant project instructions, roadmap item and any existing handoff.
Use the roadmap ID if available. Start with read-only inspection and explain the
recommendation in one or two sentences.

| Recommend | Signals |
|---|---|
| ODD (Organic Driven Development) | Bounded, understood, reversible work with clear acceptance and verification. Examples include a contained bug or a small internal adjustment. |
| SDD (Spec-Driven Development) | Changes to public contracts, migrations, changes spanning components or ownership boundaries, or work needing substantial coordination. A quick change can still have these consequences. |

Duration informs continuity. A bounded fix that takes two sessions can stay ODD
with a durable handoff, and a short public contract change can merit SDD. If the
signals are otherwise balanced, use an existing explicit preference recorded by
the user for this project. Read that preference without changing foundation
state. Honor an explicit mode choice subject to governance. When the choice or
scope is materially unclear, ask and wait before writing.

When SDD is selected, an OPSX action is explicit, or a known existing SDD change
is requested, follow the existing SDD workflow.

### Workspace readiness for SDD

Check CLI availability without installing anything. If the CLI is unavailable,
explain the blocker and continue in ODD within the authorized scope. Preserve
any existing SDD artifacts and pause actions that cannot safely run in ODD or
still need governance approval.

When the CLI is available, use the existing workspace if present. Otherwise
follow the existing SDD onboarding procedure and ask for confirmation before
running `openspec init --tools codex` from the project root. If declined, leave
the workspace untouched and continue ODD only within the approved scope. If
initialization fails, report the error and pause the requested SDD action.

Only after the workspace is ready, run `openspec list --json` and, for an
identified existing change, `openspec status --change "<name>" --json`. Use the
installed SDD action for the request. Ask for the target if ambiguous and resume
an existing change instead of duplicating it.

## Continue or resume ODD

1. Agree with the user on scope, acceptance and how to verify, using existing
   decisions when available. For self-contained work, keep decisions and evidence
   in the conversation. If another session or person needs to continue, create or
   update the single handoff at `knowledge-base/entrega/<id>.md` with the
   [handoff reference](references/handoff.md). Reuse the roadmap ID or agree on
   a stable local ID if none exists. Confirm the identity if several records match.
2. To resume, read the handoff for that ID. Inspect the current Git worktree and
   compare it with the recorded base, paths and verification. Preserve existing
   user edits and test evidence. Re-run checks when relevant inputs or files have
   changed and mark stale or unrun checks explicitly.
3. Implement and verify within the agreed scope. Update the existing record when
   continuity still depends on it. Keep OpenSpec out of ODD, including when
   resuming. Reassess the mode if impact or coordination changes, and surface any
   newly applicable governance gate before extending the scope.

## Reconcile the roadmap

For ODD and SDD, reconcile the roadmap item on resume, delivery and closure
using the [reconciliation reference](references/roadmap-reconciliation.md).
On resume, check before making decisions based on cached status or assignment.
On delivery and closure, check before reporting the current remote status.
Keep the work's local verification separate from the roadmap observation.

## Promote ODD to SDD

1. Pause work at the scope boundary. Explain what changed and why it merits SDD.
   Inspect the current diff, working tree and evidence without moving or undoing
   work. Prepare a handoff with the same ID and present the proposed destination
   and intended OpenSpec action. Ask the user to confirm promotion before running
   any OpenSpec action. If they decline, preserve work and continue ODD only
   within the scope they approve.
2. If confirmed, persist the handoff at `knowledge-base/entrega/<id>.md` as
   described in the reference. Follow workspace readiness above, including
   separate confirmation to initialize when needed. Keep the handoff and report
   that the promotion is pending if the workspace cannot be made ready.
3. With the workspace ready, run `openspec list --json`. For an existing matching
   change, run `openspec status --change "<name>" --json` and resume the
   appropriate installed OpenSpec action. If no change matches, use the existing
   proposal action for the confirmed destination when appropriate. Confirm the
   target when ambiguous. Supply the handoff path and ID, decisions, acceptance,
   open issues, current diff context and verification evidence to the action.
4. Let the existing SDD workflow own its artifacts and gates. Record the
   confirmed change name in the handoff once it exists. Carry valid evidence
   forward and identify what needs to be rerun or verified. Promotion never
   commits, stashes, resets, switches branches or rewrites project files on its
   own.

## ClickUp actions

When the user asks what they can do with ClickUp or makes an open-ended request about the project roadmap, invoke the action selector in `active-roadmap` if installed. For a concrete request, use the corresponding roadmap workflow.

## Documentation diagrams

Include a fenced Mermaid block in the delivered Markdown when it clarifies an architecture flow or another documented relationship. Use the facts in the deliverable for nodes and connections. On request, read the document and show the saved block as Mermaid. If the source is unavailable, identify what is missing. For ClickUp, follow the active-roadmap read and publication instructions.
