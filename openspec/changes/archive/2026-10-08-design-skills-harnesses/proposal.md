# Proposal

## Why

The `design-mockups` harness (archived change `add-design-mockups-toggle`) already tells the orchestrator to load `emil-design-eng` or a taste skill "if installed", but the installer offers no way to install them, so that soft dependency is never satisfied by Active Stack itself. Both skills are third-party repositories that change frequently (pushed 2026-10-02 and 2026-10-07), so they must enter the catalog pinned to an exact commit, and the skill installer currently cannot clone a commit SHA (`git clone --depth 1 --branch <sha>` fails).

## What Changes

- Add two global-picker skill harnesses to `internal/catalog/harnesses.yaml`, Full-only (`install_modes: [full]`), Claude-only (`agents: [claude]`), `third_party: true`, `best_effort: true`, with bilingual `description` and `long_description`:
  - `emil-design-eng` — repo `emilkowalski/skills`, path `skills/emil-design-eng`, MIT, pinned to commit `e8a175de22ae1e49370fc144c1f3bb9aeedf988d` (HEAD of `main`, 2026-10-02T10:43:31Z).
  - `design-taste-frontend-v1` — repo `Leonxlnx/taste-skill`, path `skills/taste-skill-v1`, MIT, pinned to commit `b482f7a970abb98c4108d4a9f761e458c64cefc8` (HEAD of `main`, 2026-10-07T06:20:58Z). The harness ID equals the skill's frontmatter `name`, because the installer names the installed directory after the harness ID.
- Extend the skill clone installer so that a `source.ref` that is a full 40-character hexadecimal commit SHA is fetched by SHA (shallow fetch + detached checkout) instead of being passed to `--branch`. Branch, tag, `latest` and empty refs keep the current behaviour.
- Make the skill installer skip agents that the harness does not support (`Harness.SupportsAgent`), mirroring what the config installer and `verify` already do. Harnesses with no `agents` list (every existing skill) are unaffected.
- No `depends_on` between the new skills and `design-mockups`: the dependency stays soft in both directions.
- Update docs: `README.md` harness table, `ARCHITECTURE.md` (skill row + pinning convention), `CHANGES.md` (mark the change done), and the `CLAUDE.md`/`AGENTS.md` mirror paragraph about `design-mockups`.
- Rejected sources: `Leonxlnx/taste-skill` v2 (`skills/taste-skill`, experimental, 87 KB) and `senlindesign/taste-skill` (no LICENSE, requires Playwright MCP, runs JS on external pages, `SKILL.md` at repo root).

## Capabilities

### New Capabilities
- `design-skills-harnesses`: the catalog declares the two pinned, Full-only, Claude-only third-party design skills and keeps them independent of `design-mockups`.
- `skill-harness-install`: the skill installer honours a commit-SHA `source.ref` and installs a skill only for the agents the harness supports.

### Modified Capabilities
<!-- None: design-mockups-harness already states the soft dependency ("used only if installed"); its requirements do not change. -->

## Impact

- Code: `internal/catalog/harnesses.yaml`, `internal/harness/skill/clone.go`, `internal/harness/skill/installer.go`.
- Tests: `internal/catalog/catalog_test.go` (new entry tests; Full/Custom counts 16 → 18), `internal/harness/skill/skill_test.go` (SHA ref, agent scoping).
- Docs: `README.md`, `ARCHITECTURE.md`, `CHANGES.md`, `CLAUDE.md`, `AGENTS.md`.
- Runtime: third-party prompt content is copied into `~/.claude/skills/` when selected in Full/Custom; install is best-effort (a failed clone does not abort the install).
- No change to `backup`, `filemerge` or `pipeline` (HIGH governance packages). Governance: MEDIUM (`harness/skill`), LOW (`catalog`).
