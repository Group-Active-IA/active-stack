# Tasks

> Strict TDD: in every group the RED test is written and seen failing before production code. Governance: `harness/skill` MEDIUM (checkpoint after group 2), `catalog` LOW. No `go build` (operator rule); run tests with `go test` scoped to the touched packages. No commits.

## 1. Safety net

- [x] 1.1 Run `go test ./internal/catalog/... ./internal/harness/skill/... ./internal/install/... ./internal/verify/...` and record the baseline pass count; if anything fails, stop and report it as pre-existing (do not fix)

## 2. Skill installer: commit-SHA refs (spec `skill-harness-install`)

- [x] 2.1 RED: in `internal/harness/skill/skill_test.go` add `TestClone_WithSHARef_FetchesCommit` (stub `Runner`; ref = 40-hex SHA) asserting the runner receives `git init <tmp>`, `git -C <tmp> fetch --depth 1 <url> <sha>`, `git -C <tmp> checkout --detach FETCH_HEAD`, that no call contains `--branch`, and that `SKILL.md` from the checked-out tree is copied to the skills dir (the stub side effect writes files on the `checkout` call); verify it fails
- [x] 2.2 RED: add `TestClone_WithSHARef_FetchError_ReturnsError` (runner errors on `fetch`) asserting the error names harness ID, repo URL and ref and the skills dir has no `<id>` directory; add `TestClone_ShortHexRef_UsesBranchFlag` (ref `e8a175d` → `git clone --depth 1 --branch e8a175d`); verify the new tests fail (short-hex may already pass — note it as characterization)
- [x] 2.3 GREEN: in `internal/harness/skill/clone.go` add an `isCommitSHA(ref)` helper (`^[0-9a-fA-F]{40}$`) and the init/fetch/checkout branch from design D1, leaving path resolution, idempotence, backup and copy unchanged; verify 2.1–2.2 pass and existing `TestClone_UsesDepth1AndHTTPS` / `TestClone_WithRef_UsesBranchFlag` still pass
- [x] 2.4 REFACTOR: extract command building into a small pure function if it improves readability; update the `model.Source.Ref` comment in `internal/model/harness.go` to state "branch/tag/latest, or a full 40-char commit SHA (fetched by commit)"; verify `go test ./internal/harness/skill/... ./internal/model/...` is green

## 3. Skill installer: agent scoping (spec `skill-harness-install`)

- [x] 3.1 RED: add `TestInstaller_AgentScoped_SkipsUnsupportedAgent` (harness `agents: [claude]`, adapters claude + opencode with distinct temp skills dirs) asserting only Claude's dir gets `<id>/SKILL.md` and exactly one `Result` is returned; add `TestInstaller_Unscoped_InstallsForAllAgents` (no `agents`) asserting both dirs get it; verify the scoped test fails
- [x] 3.2 GREEN: in `internal/harness/skill/installer.go` skip adapters where `!h.SupportsAgent(adapter.Agent())` (same guard as `internal/harness/config/install.go`); verify 3.1 passes and the full `./internal/harness/skill/...` suite is green
- [x] 3.3 Checkpoint (MEDIUM governance): summarize the `harness/skill` diff (SHA branch + agent guard) to the orchestrator before continuing

## 4. Catalog entries (spec `design-skills-harnesses`)

- [x] 4.1 RED: in `internal/catalog/catalog_test.go` add a table-driven `TestDesignSkills_InCatalog` covering both IDs: exists, type `skill`, repo/path/method `clone`, ref equal to the pinned SHA and matching `^[0-9a-f]{40}$`, `ThirdParty` and `BestEffort` true, not starter-only, `Agents == [claude]`, empty `DependsOn`, in Full and Custom but not Lite, non-empty `es`/`en` description and long_description; verify it fails
- [x] 4.2 RED: add `TestDesignSkills_RejectedTasteSourcesAbsent` (no skill with repo `senlindesign/taste-skill`, none with `Leonxlnx/taste-skill` + path `skills/taste-skill`) and assert `design-mockups` `DependsOn` contains neither design skill; bump `TestForMode_CustomReturnsOnlyGlobal` (`wantFoundationGlobal` 16 → 18) and `TestForMode_ExcludesStarterOnlyHarnesses` (Full 16 → 18, Custom 16 → 18, Lite stays 6), updating their history comments; verify the count tests fail
- [x] 4.3 RED: in `internal/install/design_mockups_test.go` extend the Custom picker test so `emil-design-eng` and `design-taste-frontend-v1` appear for `claude` and not for `opencode`-only / `codex`-only, and assert `BuildPlan` Custom with only `design-mockups` contains neither design skill; verify the picker assertion fails
- [x] 4.4 GREEN: add both entries to `internal/catalog/harnesses.yaml` in the "Skills de terceros" block per design D5 (display names `Emil Design Engineering` / `Taste Skill v1 (frontend)`, bilingual description + long_description covering source, pin, Claude-only, optional use by `design-mockups`, and the Emil fixed-greeting note), with a YAML comment recording pin date 2026-10-08 and the bump command `gh api repos/<owner>/<repo>/commits/main`; verify `go test ./internal/catalog/... ./internal/install/...` is green (this exercises `catalog.Load()` validation, including `localized_validation_test.go`)
- [x] 4.5 Run `go test ./internal/verify/... ./internal/uninstall/... ./cmd/...` to confirm no other suite hard-codes the harness counts or IDs; fix only count/ID expectations directly caused by the two new entries, and report anything else

## 5. Documentation

- [x] 5.1 `README.md`: add both skills to the harness table (type "skill (terceros)", mode full, Claude-only, short description) and mention them in the Full row; verify by reading the rendered table rows
- [x] 5.2 `ARCHITECTURE.md`: add both to the third-party list in the Skill row (§ harness types) and document the pinning convention (full SHA in `ref` → shallow fetch by commit; manual bump via `gh api`) next to the skill installer notes; verify the section reads consistently with design D1
- [x] 5.3 `CLAUDE.md` and `AGENTS.md` (mirror — edit both identically): in the `design-mockups` paragraph add one sentence that `emil-design-eng` and `design-taste-frontend-v1` are optional Claude-only skill harnesses (Full, pinned SHA) loaded by the design stage when installed, with no hard dependency; verify both files contain the same sentence
- [x] 5.4 `CHANGES.md`: update the `design-skills-harnesses` card (Estado: IMPLEMENTADO (pend. archive), chosen sources, SHAs + date, rejected sources and why, agents decision, SHA-ref installer extension), mark the row in the summary table, and update the dependency tree line (remove "futuro"); verify no `senlindesign` "o" ambiguity remains in the card

## 6. Integration checks

- [x] 6.1 Manual smoke of design D1 without `go build`: in a scratch temp dir run `git init`, `git fetch --depth 1 https://github.com/emilkowalski/skills e8a175de22ae1e49370fc144c1f3bb9aeedf988d`, `git checkout --detach FETCH_HEAD`, confirm `skills/emil-design-eng/SKILL.md` exists; repeat for `Leonxlnx/taste-skill` @ `b482f7a970abb98c4108d4a9f761e458c64cefc8` and `skills/taste-skill-v1/SKILL.md`; delete the scratch dirs
- [x] 6.2 Run the full suite `go test ./...` and confirm it is green relative to the 1.1 baseline; run `openspec validate design-skills-harnesses --strict`
