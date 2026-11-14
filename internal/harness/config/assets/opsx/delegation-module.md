## Delegation

Load this module only when the delegation rule decided to delegate.

### Assignment
Every assignment includes:
- goal: one verifiable sentence
- action: explore | propose | apply | verify | archive
- start: base commit and working-state fingerprint you hand to the worker, computed as the
  tested-state fingerprint of the execution receipt contract (change 06)
- write set: exclusive paths or globs the worker may create, modify, rename or delete (empty = read-only)
- read pointers: paths or symbols, never pasted content
- limits: what is not touched and what is not decided
- exit criteria: observable "done"
- return contract: the envelope below
- attempts: at most 3

Never split these across concurrent workers: `go.mod`, `go.sum`,
`internal/catalog/harnesses.yaml`, `internal/harness/config/assets/*/sdd-orchestrator.md`,
`CLAUDE.md`, `AGENTS.md`.

### Ownership
A worker with a write set works in its own worktree created from the recorded start. When it
returns, compute what changed since the start: new commits, staged, unstaged, untracked, deleted and
renamed files (both source and destination count). Any path outside the write set → `out_of_scope`,
even if it is committed and tests pass. Changes that existed at the start are not the worker's.
Before integrating, check scope again on the combined result and re-run the scenarios that depend on
more than one assignment, with the capture of the harness: a re-run without its own independent
capture does not corroborate anything.

### Parallelism
At most two workers with a write set at the same time, each in its own worktree, and only when
their write sets are disjoint, neither depends on what the other writes, and each has its own exit
criteria and tests. Otherwise serialize. Read-only workers may share a tree.

### Depth
Only you launch workers. An assignment never asks a worker to delegate. Never delegate the
coordinator role.

### Return envelope
    status:    completed | blocked | out_of_scope | failed
    start:     base commit + declared fingerprint
    modified:  [path, kind: added | modified | deleted | renamed(from→to)]
    findings:  [{severity, path:line, description}]
    receipts:  [one execution receipt per command run, as defined by change 06]
    pending:   short text | —

Validate in order: start matches the recorded one; `modified` matches the attributable changes and
stays inside the write set; receipts, with the review step of the execution receipt contract; exit
criteria. A free-text return without these fields stays unverified: ask for the envelope or retry.
What the worker declares stays **unverified** until an independent source, or a re-run with its own
independent capture, confirms it. "Tests green" without that is still unverified.
Corroboration must name the command, tree and result it observed: a judgment about another
execution is not reused, and a capture the worker declares about itself is not independent.

### Attempts
Up to 3 attempts per assignment. Each retry gets the rejection reason and a new recorded start.
When the limit is reached, escalate to the user with the history.

### Evidence
When the assignment runs commands, paste the `## Evidence Contract` excerpt of the execution
receipt contract into it, and review the returned receipts with the same step you use when you run
the commands yourself. If that contract is not installed, every receipt stays unverified: say so in
the report.
