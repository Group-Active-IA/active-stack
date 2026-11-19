# Team mode contract

Shared contract for projects founded in team mode. Foundation (`active-orchestrator`
and `agent-instruction`) writes these files, `active-delivery` reads them at the
start of work, and the roadmap verifier (`roadmap-generator`) checks fichas against
them. Every reader and writer validates with `team.py` instead of reimplementing
the rules. A project without team mode behaves exactly as before.

## Files

| File | Versioned | Written by | Content |
|---|---|---|---|
| `.active-stack/team.json` | Yes | Foundation and "update team" | Team size, tracking mode and member ids; names only when the team approved versioning them |
| `.active-stack/team.local.json` | No (ignored) | Foundation, when names are not versioned | Names and account links for members |
| `.active-stack/me.local.json` | No (ignored) | `active-delivery`, on each person's first session | The member `id` of the person using this checkout |

The founder defines the team. The agent never invents names, accounts or ids.
Foundation adds the ignore rules for both local files to the project `.gitignore`
inside idempotent markers.

Where the roadmap link and observation state (`.active-stack/roadmap.json` and
`.active-stack/roadmap-state.json`) live and which parts are versioned is decision
DA-01 `(TBD)`, owned by the project lead. Team mode requires `team.json` to be
versioned: if it stays local, a fresh clone of a team project cannot be told apart
from a project with a missing file.

## `team.json` (version 1)

```json
{
  "version": 1,
  "size": 4,
  "tracking": "clickup",
  "names_versioned": true,
  "members": [
    { "id": "m1", "name": "Ana", "github": "ana-dev", "clickup_id": 12345 },
    { "id": "m2", "name": null }
  ]
}
```

| Field | Required | Rule |
|---|---|---|
| `version` | No (defaults to 1) | Integer. Only version 1 is supported; another value is a clear error, never a guess |
| `size` | Yes | Integer ≥ 2. The smallest valid file is `{"size": 2}` |
| `tracking` | No (defaults to `local`) | `local` or `clickup` |
| `names_versioned` | No (defaults to `true`) | Boolean. With `false`, members in this file carry only `id` |
| `members` | No | List of at most `size` members. Missing seats stay pending as "Member N" |

Member fields:

| Field | Required | Rule |
|---|---|---|
| `id` | Yes | Lowercase letters, digits and hyphens, unique (`m1`, `m2`…). Fichas use it as `Responsable` |
| `name` | No | Text or `null`. A pending member has no name |
| `github` | No | GitHub handle, used to suggest identity and for `CODEOWNERS` |
| `clickup_id` | No | Integer, only with `tracking: clickup`, confirmed with the person |

Unknown fields are rejected. No field accepts credentials: values that look like
GitHub, ClickUp, OpenAI or Slack tokens, or `Bearer` headers, are rejected in every
file of this contract, fichas included.

## Local files

`.active-stack/team.local.json` has the shape `{"members": [...]}` with the same
member fields. Its ids must exist in `team.json`, and `clickup_id` still requires
`tracking: clickup` there.

`.active-stack/me.local.json` has the shape `{"id": "m2"}`. The id must belong to a
known member. Without this file the person has not chosen an identity yet:
`active-delivery` suggests one from the GitHub account (`gh api user`) or asks.

## Team project signals

A missing `team.json` does not prove that a project is individual. The validator
looks for signals that the project was founded in team mode:

1. `.active-orchestrator-state.json` records the foundation phase `team-mode` with
   status `done` (`foundation.phases["team-mode"].status`). `skipped`, `deferred`
   and `pending` are not signals.
2. The project `CLAUDE.md` or `AGENTS.md` contains the team rules section, opened
   by `<!-- active-stack:team-mode -->` and closed by
   `<!-- /active-stack:team-mode -->`.
3. `me.local.json` or `team.local.json` exist in this checkout.
4. Git, read-only and offline: the file exists in the last commit (`HEAD`) or in
   the default branch of the local copy.

Writers must make these signals reliable: foundation writes `team.json`, the
`team-mode` phase record and the rules section together, and proposes committing
them in the same commit. If the founder commits none of them, the project cannot
be told apart from an individual one.

## Validator

Agents always call the launcher, which picks a Python 3.8+ interpreter:

- POSIX shells (Linux, macOS, Termux, Git Bash): `sh <skill-dir>/team.sh --root <project> validate`
- PowerShell or cmd on Windows: `<skill-dir>\team.bat --root <project> validate`

Commands:

- `validate` checks the project: `team.json`, both local files and the signals.
- `validate --file <path>` checks a draft as `team.json` without reading the
  project. Foundation validates its draft before writing.
- `ficha --file <path>` checks one ficha, given as a JSON object of its fields,
  against the project team.

Output is always one JSON object on stdout:

| Result | Exit | Output |
|---|---|---|
| Individual project: no `team.json` and no signal | 0 | `{"ok": true, "team_mode": false}` |
| Team project, valid | 0 | `{"ok": true, "team_mode": true, "size": 4, "tracking": "local", "members": ["m1", "m2"], "me": "m2"}` (`me` is `null` until chosen) |
| Team project, invalid | 1 | `{"ok": false, "team_mode": true, "errors": [...]}` |
| Team project, file missing | 1 | `{"ok": false, "team_mode": "missing", "errors": [...], "signals": [...], "hint": "..."}` |
| Wrong usage | 2 | Usage message |
| No Python 3.8+ available | 3 | `{"ok": false, "errors": ["falta Python 3.8 o superior..."]}` |

Each error names the field it refers to, for example `members[1].github`. Show
errors and hints to the person as they are and never repair, invent or guess team
data. With `missing`, the hint gives the fix for the strongest signal: restore the
file from the last commit, update the branch from the default branch (after
`git fetch`), or ask the founder to commit it and use "update team".

Recommended reader behavior, to be confirmed where each reader is implemented
`(TBD)`: with exit 1 or 3, continue without team features and say why, instead of
stopping all work.

Run the contract tests from the skill folder with `python -B -m unittest test_team`.
Always pass `-B`: the skill folder is embedded in the installer, and a
`__pycache__` folder there would ship with it.

## Ficha fields

Fichas in `CHANGES.md` stay plain Markdown and work for every agent. Team mode adds
these fields:

| Field | Rule |
|---|---|
| `id` | Stable ficha ID. It keys the branch, the PR, ClickUp and the handoff |
| `Resultado` | Verifiable outcome |
| `Depende de` | Ficha IDs, `#PR`, or `#PR@SHA` only for an open PR, comma separated, or `ninguna`. The legacy `Dependencias` field is read the same way |
| `Zona` | Paths or packages it touches. Recommended, not blocking |
| `Requiere` | Agent and system the task needs, e.g. `claude o codex · windows y linux` |
| `Modo` | `odd` or `sdd` |
| `Responsable` | Member `id` from `team.json`, assigned by the approved split |

Existing fichas without these fields remain valid: `ficha` returns warnings that
name the missing fields and no errors. Errors are reserved for malformed values,
an unknown `Responsable` and credentials. Waves are computed from `Depende de`;
cross-ficha checks such as cycles, overlapping zones and PR chains belong to the
roadmap verifier.

## Handoff in team mode

In team mode every ficha keeps one handoff at `knowledge-base/entrega/<id>.md`,
also when nobody continues it, using the template in `handoff.md`. It always
records six facts: result, PR and SHA, how to verify, what is left, who continues,
and what blocks it.

## Team rules section

Foundation writes a short section in the project `CLAUDE.md` or `AGENTS.md`, never
in the global file, between the `active-stack:team-mode` markers:

- When `.active-stack/team.json` exists or the validator reports `missing`, use
  `active-delivery` to identify the person and show their work before anything else.
- One branch and one worktree per ficha, named `<id>-slug`. Never two agents on the
  same branch or folder.
- Nothing directly to `main`; small PRs; commit, push and PR only with authorization.
- If `active-delivery` is unavailable in this client, apply the same convention by
  hand: branch per ficha, PR, no push to `main`, handoff in
  `knowledge-base/entrega/<id>.md`.
