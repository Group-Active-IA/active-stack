# Delivery in team mode

Use this workflow on a new session in a team project, on resumption, and when
the user asks what to work on. Follow the same steps in Claude and Codex. The
project files and Git/GitHub are the shared evidence across clients.

## Identify the member

Run the skill's `team.sh --root <project> validate` (POSIX) or
`team.bat --root <project> validate` (Windows). Only use this reference on
`ok: true, team_mode: true`. For `team_mode: false`, continue with the original
individual workflow. If the validator detects missing or invalid team data, show
the error and any hint, keep team actions read-only, and ask the team to repair
its source. If a launcher is unavailable, report that team mode could not be
verified. Apply these limits to team actions while respecting project rules on
other work.

Use the same launcher with `--root <project> identity`. If it reports a member,
use that member until the user asks to change it. Otherwise run the read-only
`gh api user --jq .login` when GitHub is configured. Pass an observed login as
`identity --github <login>`. Pass `--clickup-id <id>` only when the configured
ClickUp connector has already supplied the viewer's own ID. These inputs only
suggest a known member. Ask the person to confirm the suggestion, choose from
the returned list, or stay in consultation mode. If providers disagree, ask
the person to select from the list. Persist only an explicitly confirmed
member with `identity --select <member-id>`. If Git does not ignore that file,
show the error and ask the founder to repair the ignore rules. Do not commit
either local identity file. An unassigned person can consult the roadmap and
ask the founder for a team seat or assignment.

## Show what is assigned

Read `CHANGES.md` and any ficha it links. Use the ficha's stable ID and fields
from [the team contract](team-mode.md#ficha-fields). Determine assignments,
status, dependencies still waiting, downstream fichas blocked by this member,
and pending PR reviews. Check GitHub with the configured repo, including
`gh pr view <number> --json state,headRefOid,mergeCommit,url` for known PR links
and `gh pr list --state open --search "review-requested:@me" --json number,url`
for the signed-in user. In ClickUp mode read the linked tasks through
[roadmap reconciliation](roadmap-reconciliation.md), where the authoritative
status and assignee live. Resolve an observed ClickUp assignee to a known
member ID using the validated team records. An explicitly unassigned task
overrules the local assignee. If the provider is unavailable, show the local
assignment with its source, and report remote status and assignment as
unconfirmed. Do not promote old observations to fresh provider reads.

Use `overview --file <temporary-json>` on the launcher to summarize evidence.
Keep the temporary input outside the repo and out of shared artifacts. Include
these fields:

- `sources` maps keys such as `roadmap`, `github`, and `clickup` to
  `{"source":"CHANGES.md","checked_at":"<UTC timestamp>","available":true}`.
  For an unavailable read, set `available` to false and include `reason`. Separate
  partial reads when a provider returns one fact and fails on another.
- `fichas` holds the roadmap's ficha objects as accepted by `team.py ficha`.
- `observations` maps a ficha ID or `#N` PR number to
  `{"source":"github","status":"merged"}` or the relevant confirmed status.
  Include `assignee` only with a confirmed member ID or explicit null for an
  unassigned task. Use `done` only for a completed ficha, and `omitted` only for
  a formally omitted dependency. Use `open`, `merged`, or `closed` for a PR.
- `reviews` holds the confirmed requested PR numbers or URLs, omitted when
  unconfirmed. A confirmed empty list means there are no matching reviews.

Report the source and check date next to each fact and name the unavailable
source and impact. Keep a closed PR without merge unresolved. If no
authoritative resolution exists, mark the dependency unconfirmed.

## Start or resume a ficha

Check ownership and workspace first. Review local and fetched branches, worktrees,
and open PRs for the stable ID. Where the user already has a clean, matching
workspace, confirm it is theirs and available before using it. Independently
verify no other agent is using that checkout. Git can report the path and
status of a worktree; ask the user or current owner to confirm session
ownership. Keep one agent per branch and directory.

If its dependency is open, show the known PR and current tip SHA and offer to
wait or start on that tip, subject to project policy. If the user chooses to
stack, document the unresolved dependency, pin the agreed SHA, target that
dependency's branch if an authorized draft is needed, and recheck the pin
before drafting or moving the base. Never silently change the base. Choose
ODD or SDD using the main skill's existing confirmation rules.

For a new workspace, propose `<id>-<slug>` and a sibling path beside the
current checkout. Preview with the launcher:

```text
workspace --id R-11 --branch R-11-delivery --base origin/main
```

Present the path, branch and base SHA. For `attach`, also present the observed
local tip SHA. After the user confirms, append `--apply --expected-base <sha>`,
and for `attach` also `--expected-tip <tip-sha>`. The helper resolves the base
again and refuses to act on stale evidence. Avoid making any edits before
confirming the resulting workspace. Continue inside the resulting path only
after ownership and availability are confirmed.

- `ready` identifies the created workspace. `occupied` identifies an existing
  checkout to inspect and resume with its owner. A retry never creates a
  second checkout for the branch.
- `dirty` gives the checkout to review with the person. Ask whether to resume
  there and preserve its changes, use a separate clean checkout subject to the
  ownership rule, or wait. Do not stash, reset, relocate or rewrite user changes.
- `collision`, `branch-conflict` and `remote-branch` mean the path or ficha
  already exists. Inspect the existing state and use it only with confirmation.
  A remote-only branch can be resumed from its verified remote tip after the
  owner confirms the local tracking branch and its base. Re-run the preview.
- `busy` means another helper is operating on the same ficha in the shared Git
  directory. Wait for that operation and then preview again.
- After a failure, read the current branch and worktree state before retrying.
  Confirm any changed base or branch tip. Never use `--force` or `-B` to recover.

Local identity and private team names stay in their original checkout when a
sibling is created. Use the original validated checkout to resolve a private
name or account before proposing the same stable member ID in the sibling.
Confirm that ID there with `identity --select`; the versioned team data is
sufficient for the same member IDs, assignments and handoff. If the original
private mapping is unavailable, use the team IDs and ask the founder or person
to confirm the mapping. Do not copy or commit private data as part of a task.

## Check five notices

Gather current observations at workspace start and again before an authorized
PR. Use Git and GitHub reads, configured ClickUp in read-only mode, and
`Requiere` from the ficha. Report notices and missing observations, then ask
what to do when action or coordination is needed. This check does not write
project code or mutate remote state.

1. **Existing work.** Report another assignee, a matching branch or a PR for
   the ficha. Match stable IDs in the roadmap, branches and linked tasks.
2. **Dependency base moved.** Compare the observed current tip with the
   agreed `#PR@SHA` pin. Record a failed read as unconfirmed.
3. **Open PR overlap.** List open PRs for this repo and use
   `gh pr diff <number> --name-only` to compare their changed paths with
   `Zona`. Probe enough pages for the scope or state which reads are incomplete.
4. **ODD overlaps an open SDD change.** From those PRs' changed paths find
   additions or updates below `openspec/changes/` outside the archive. Read
   their changed artifacts using the PR diff or its head and compare their
   scope with `Zona`. Include PRs visible only in GitHub. Treat missing scope
   as unconfirmed. Exclude the current ficha's own PR from overlap checks.
5. **Machine requirement unmet.** Check agent/platform/tools against `Requiere`
   and confirm ambiguous or untestable requirements with the person.

Use `notices --file <temporary-json>` with `assignment` (optional `assignee`,
`branch`, `pr`; empty object for a verified clear assignment, null for an
unavailable check), `dependency` (optional `expected` and `actual` SHAs),
`mode` (`odd` or `sdd`), `zone` (repo-relative paths with forward slashes),
`current_pr` (number if known), and `open_prs` (list of objects with
`number`, `paths`, and for an SDD artifact `sdd_zone`; use null when a read
fails). Set `requirements_met` to true, false, or null if unconfirmed. Report
the results with the time and source of the observations used as input.

## Record and deliver

Keep one [handoff](handoff.md) at `knowledge-base/entrega/<id>.md` from the start
of work, in ODD and SDD, even when nobody follows. Update it at resumption and
delivery. Each update includes all six team fields, with the evidence for PR,
SHA and checks. Preserve earlier decisions and evidence when changing clients.

Before a draft PR, repeat the notices and present the exact head, base,
repository, and verification evidence. Ask for explicit authorization to
commit, push, or create the draft when it has not already been given. Use
`gh pr create --draft` with explicit head and base and the repository's
template. Include the ficha ID, handoff link and verification. Only push the
authorized ficha branch to its corresponding remote ref. Never push directly
to `main` or any protected base, and never force-push (including with lease).
The same authorization and branch rules apply after ODD-to-SDD promotion.
Promotion itself never commits or moves the worktree.

After authorized delivery, look up the resulting PR and current head SHA,
then update the handoff. On closure verify the PR is merged and the actual
merge SHA is reachable from the agreed target. Record verification and offer
a separate update to the tracking system subject to its read/preview/approval
rules.

[Git worktree reference](https://git-scm.com/docs/git-worktree) ·
[GitHub CLI PR list](https://cli.github.com/manual/gh_pr_list) ·
[GitHub CLI PR diff](https://cli.github.com/manual/gh_pr_diff) ·
[GitHub CLI PR view](https://cli.github.com/manual/gh_pr_view)
