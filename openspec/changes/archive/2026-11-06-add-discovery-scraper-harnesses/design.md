## Context

See `proposal.md` - Why. The two skills (`discovery-research`, `web-scraper`) already exist, are published under `Group-Active-IA`, and the two external repos that dispatch/consume them (`active-orchestrator`, `kb-creator`) are already edited and pushed. This change only adds the missing catalog entries in `internal/catalog/harnesses.yaml` so `active-stack` can offer/install them — see `specs/catalog-discovery-harnesses/spec.md` for the exact requirements.

## Goals / Non-Goals

**Goals:**
- Both harnesses resolve via `catalog.ByID`, appear in `ForMode(full)`, and are absent from `ForMode(lite)`.
- `catalog.Load()` keeps succeeding (bilingual completeness, valid `depends_on`).

**Non-Goals:**
- No Go code changes — `catalog.Load()`/`ByID`/`ForMode`/`ForAgent` already work generically off the YAML; this is data-only.
- No starter wiring — neither harness is added to `base`/`backend`/`ux-ui`/`active-ia`'s `harnesses` list; they stay selectable only as global Full-mode harnesses.

## Decisions

- **Two separate harness entries, not one bundled entry.** `web-scraper` is independently useful (any skill that wants to scrape a URL can depend on it), and the catalog already models this exact relationship via `depends_on` — same pattern used nowhere else in the catalog yet for two first-party skills, but it's the same field `openspec`/`kb-creator` use for cross-harness dependencies today, just applied between two skill-type harnesses instead of a skill depending on an external tool.
- **`discovery-research` depends_on `web-scraper`, not the reverse.** `discovery-research` invokes `web-scraper` at runtime as a sub-skill; `web-scraper` has no knowledge of `discovery-research`. The catalog dependency direction mirrors the runtime call direction, so installing `discovery-research` always brings `web-scraper` along.
- **`install_modes: [full]` only, no starter membership.** Every existing skill-type harness in this catalog (`kb-creator`, `roadmap-generator`, `agent-instruction`, `skill-registry`, `active-orchestrator`) follows this same pattern — Full is the foundation-guided mode, Lite is bare substrate. Adding these two to a starter's list was considered and rejected: Discovery is optional even within Full (the user is asked at runtime), so forcing it into a starter bundle would contradict that opt-in design.

## Risks / Trade-offs

- **Both source repos started private** (`Group-Active-IA/discovery-research`, `Group-Active-IA/web-scraper`) — `skill-forge`'s publish default — unlike every other first-party skill in this table, which is public. Resolved before this change closed: the maintainer chose to make both public for consistency with the rest of the table, so `method: clone` works for anyone installing `active-stack` in Full mode without needing repo access.
- **`ForMode`'s foundation-global count was hardcoded in two tests** (`TestForMode_CustomReturnsOnlyGlobal`, `TestForMode_ExcludesStarterOnlyHarnesses`, both at 13). → Discovered while running the validation tasks, not anticipated here: both bumped to 15 with a comment explaining why, rather than left as a silent regression. Every other catalog test (bilingual completeness, `depends_on` resolution, the `c30_curation_test.go` fixed-ID lists) exercises new entries generically and needed no changes.
