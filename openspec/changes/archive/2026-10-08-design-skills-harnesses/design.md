# Design

## Context

See proposal.md (Why). Current state verified in code on 2026-10-08:

- `model.Source.Ref` is documented as "tag/branch/commit; defaults to latest", but `internal/harness/skill/clone.go` only builds `git clone --depth 1 [--branch <ref>] <url> <tmp>`. `git clone --branch` accepts branches and tags only, so a commit SHA fails today. `catalog.Load()` does not validate `ref` at all. `Source.Ref` is read in exactly one place (`installer.go` → `cloneInstaller`).
- The skill step (`internal/install/plan.go` → `skillStep`) passes every selected adapter to `skill.Installer.Install`, which does not check `Harness.SupportsAgent`. The config installer (`internal/harness/config/install.go`) and `verify.ChecksForHarness` already skip unsupported agents. Plan selection (`filterByAgents`) and uninstall (removal of a missing dir is a no-op) need no change.
- Global third-party skills (`find-skill`, `skill-creator`) use `third_party: true`, `best_effort: true`, a `path`, no `agents` list, and a bilingual `long_description` (mandatory for non-starter-only harnesses, enforced by `catalog.Load()` and `localized_validation_test.go`). `interface-design` and `code-review-excellence` are `scope: starter-only` and are not the right template.
- The installed directory is `<skillsDir>/<harness ID>`; Claude skills expect the directory name to match the `SKILL.md` frontmatter `name`.
- Upstream facts (via `gh api`, 2026-10-08):

| Harness ID | Repo / path | License | Pinned SHA (HEAD of `main`) | Commit date | Files |
|---|---|---|---|---|---|
| `emil-design-eng` | `emilkowalski/skills` / `skills/emil-design-eng` | MIT | `e8a175de22ae1e49370fc144c1f3bb9aeedf988d` | 2026-10-02T10:43:31Z | `SKILL.md` (27,123 B), frontmatter `name: emil-design-eng` |
| `design-taste-frontend-v1` | `Leonxlnx/taste-skill` / `skills/taste-skill-v1` | MIT | `b482f7a970abb98c4108d4a9f761e458c64cefc8` | 2026-10-07T06:20:58Z | `SKILL.md` (21,195 B), frontmatter `name: design-taste-frontend-v1` |

  `Leonxlnx/taste-skill` is ~36 MB on GitHub; `emilkowalski/skills` is ~111 KB.

## Goals / Non-Goals

**Goals:**
- Reproducible installs of the two design skills at the pinned commits.
- Minimal, backwards-compatible extension of the clone installer (SHA refs only).
- Claude-only scope that is actually honoured at install time, not just in the picker.

**Non-Goals:**
- Changing the `design-mockups` fragment text (it already names `emil-design-eng` and "a taste skill"; its golden files stay untouched).
- Pinning existing `ref: latest` harnesses, or adding catalog-level ref validation for other harnesses.
- An automated SHA-bump mechanism (bumping is a manual catalog edit with a new `gh api` lookup).
- Filtering skill uninstall by agent or the broader adapter-filtering debt of `sdd-orchestrator`.
- Any change to `backup`, `filemerge` or `pipeline`.

## Decisions

### D1 — Fetch a SHA ref with `init` + shallow `fetch` + detached `checkout`
When `ref` matches `^[0-9a-fA-F]{40}$`, the installer runs, through the existing `Runner`:
1. `git init <tmp>`
2. `git -C <tmp> fetch --depth 1 <repoURL> <sha>`
3. `git -C <tmp> checkout --detach FETCH_HEAD`

then continues with the existing path resolution / idempotence / backup / copy logic unchanged. Any other ref keeps the current `git clone --depth 1 [--branch <ref>]` command.

- *Alternative: full `git clone` (no depth) + `git checkout <sha>`.* Rejected: downloads the full history of `Leonxlnx/taste-skill` (~36 MB) to copy a 21 KB file.
- *Alternative: download a tarball from `codeload.github.com/<repo>/tar.gz/<sha>`.* Rejected: new code path (HTTP + tar extraction), GitHub-specific, breaks the `file://` hermetic-test convention of `cloneInstaller`.
- *Alternative: keep `ref: latest` and document the risk.* Rejected by the user: repos change often; pinning is the decision.
- GitHub serves reachable commits by SHA to `git fetch` (protocol v2), so step 2 works against github.com. Only full 40-char SHAs trigger this path, so short hex strings that could be tag names keep their current meaning.
- No automatic fallback to a full clone on fetch failure: it would silently pull ~36 MB and hide a broken pin. Both harnesses are `best_effort`, so a failure is reported and the install continues.

### D2 — `agents: [claude]` and enforce it in the skill installer
Both harnesses declare `agents: [claude]`, for coherence with `design-mockups` (the only consumer of these skills is a Claude-only section) and to keep third-party prompt content out of agents that would never be told to use it. Because the skill installer currently ignores `agents`, `skill.Installer.Install` gains the same guard as the config installer: skip adapters where `!h.SupportsAgent(adapter.Agent())`. Every existing skill has an empty `agents` list (= all agents), so behaviour for them is unchanged.

- *Alternative: no `agents` list, like `find-skill`/`skill-creator`.* Rejected: those are generic tools; these skills only matter through the Claude-only design stage.
- *Alternative: `agents: [claude]` without the installer guard.* Rejected: with Claude + OpenCode selected the skill would still land in OpenCode's skills dir, making the catalog declaration misleading.

### D3 — Harness ID equals upstream skill name
`emil-design-eng` and `design-taste-frontend-v1` match the frontmatter `name` of each `SKILL.md`, because the installer names the directory after the ID. An ID such as `taste-skill` would produce a directory whose name differs from the skill's declared name. Display `name` fields: `Emil Design Engineering` and `Taste Skill v1 (frontend)`.

### D4 — Third-party, best-effort, Full-only, no dependencies
`third_party: true`, `best_effort: true` (same as other upstream skills: a clone failure does not abort), `install_modes: [full]`, no `depends_on`, not starter-only. They appear in the Custom picker for Claude and are installed by Full when Claude is selected. Full/Custom counts move 16 → 18.

### D5 — Catalog entries (target shape)
Placed in the "Skills de terceros" block next to `find-skill`/`skill-creator`, with a comment recording the pin date and the `gh api repos/<owner>/<repo>/commits/main` command used to bump it:

```yaml
- id: emil-design-eng
  type: skill
  third_party: true
  best_effort: true
  agents: [claude]
  source: { repo: emilkowalski/skills, ref: e8a175de22ae1e49370fc144c1f3bb9aeedf988d, method: clone, path: skills/emil-design-eng }
  install_modes: [full]
- id: design-taste-frontend-v1
  type: skill
  third_party: true
  best_effort: true
  agents: [claude]
  source: { repo: Leonxlnx/taste-skill, ref: b482f7a970abb98c4108d4a9f761e458c64cefc8, method: clone, path: skills/taste-skill-v1 }
  install_modes: [full]
```

`long_description` (es/en) must state: what the skill does, where it is cloned from and pinned, that it is optional and used by `design-mockups` when present, Claude-only, and (for Emil) that it answers a fixed greeting when invoked without a question.

## Risks / Trade-offs

- [Third-party content is injected into the agent's context (prompt-injection surface)] → pinned SHA reviewed at proposal time (both are plain-text `SKILL.md`, no scripts, no tool requirements); opt-in (Full/Custom only, Claude only); bumping the SHA is a deliberate catalog edit that should include re-reading the diff.
- [`emil-design-eng` forces a fixed first reply ("I'm ready to help you build interfaces that feel right…") when invoked without a specific question] → documented in its `long_description`; the `design-mockups` stage loads it with a concrete design task, so the greeting is not expected in that flow.
- [`design-taste-frontend-v1` imposes fixed baseline variables (DESIGN_VARIANCE 8, MOTION_INTENSITY 6, VISUAL_DENSITY 4) and opinionated conventions; upstream labels v1 "preserved for backward compatibility"] → pin protects against removal/rewrite; v2 explicitly rejected as experimental.
- [Upstream repos are active (pushed 2026-10-02 and 2026-10-07); pins go stale] → accepted; stale is safer than drifting. Manual bump procedure documented in the YAML comment and ARCHITECTURE.md.
- [Upstream force-push or history rewrite makes the pinned commit unreachable] → fetch fails with an error naming harness/repo/ref; best-effort keeps the install going; fix is a SHA bump.
- [Fetch-by-SHA depends on the server allowing SHA wants] → github.com supports it; unit tests use the stub `Runner`; a one-off manual smoke of the three git commands against both repos is part of the tasks (no `go build`).
- [Agent guard in the skill installer changes a shared code path] → no-op for every existing skill (empty `agents`); covered by an explicit unscoped-skill test.

## Migration Plan

Additive. Existing installs are unaffected; users get the new harnesses only by selecting them (Full re-run or Custom). Rollback: remove the two catalog entries; the SHA branch and agent guard are inert without scoped/pinned skills.

## Decisiones tomadas en apply (delegadas: "elegí lo mejor", criterio conservador)

- **A1. Checkpoint 3.3 (MEDIO, `harness/skill`) resuelto sin frenar**: el diff es acotado: `clone.go` suma `isCommitSHA` y `gitCommands` (rama SHA = `init` + `fetch --depth 1` + `checkout --detach FETCH_HEAD`; resto de los refs sin cambios) e `installer.go` suma el guard `!h.SupportsAgent(adapter.Agent())`. Ninguna skill existente declara `agents`, así que su comportamiento no cambia (cubierto por `TestInstaller_Unscoped_InstallsForAllAgents`).
- **A2. `gitCommands` extraída como función pura** (refactor 2.4) y el mensaje de error de la rama SHA dice `git fetch <url> at commit <sha>` (nombra harness, repo y ref).
- **A3. Tests de conteo expresados por modo** (Lite 6, Full 18, Custom 18) con historial en los comentarios, para que un change posterior (p. ej. `visual-verify`) solo tenga que bumpear números.
- **A4. Smoke 6.1 ejecutado** contra GitHub real en un directorio temporal nuevo (borrado al final): ambos SHAs se obtienen con `fetch --depth 1` y `checkout --detach`, y los `SKILL.md` existen en `skills/emil-design-eng/` y `skills/taste-skill-v1/`. No se ejecutó nada del contenido descargado.
