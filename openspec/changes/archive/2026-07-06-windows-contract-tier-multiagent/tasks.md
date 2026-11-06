# Tasks — windows-contract-tier-multiagent

> Strict TDD: for each behavior write the failing test FIRST (RED), minimal code to pass (GREEN),
> then TRIANGULATE with a second case, then REFACTOR. Do not build unless explicitly asked.
> Governance: MEDIO — implement with checkpoints. Do NOT modify internal/backup, internal/install, internal/uninstall.

## 1. Safety net (baseline)

- [x] 1.1 Run `go test ./cmd/active-stack/headless/... ./internal/model/... ./internal/tui/...` and record the passing baseline (the 5 existing `windows_test.go` tests must be green before any change).

## 2. model.TierCapable — single source of truth (RED → GREEN → TRIANGULATE)

- [x] 2.1 RED: add `internal/model/tier_capable_test.go` — `TestTierCapable` asserting `TierCapable([]Agent{AgentClaude})` is true and `TierCapable([]Agent{AgentGemini})` is false (references `model.TierCapable`, which does not exist yet).
- [x] 2.2 GREEN: add `model.TierCapable(agents []model.Agent) bool` (with the unexported `tierCapableAgents` set: claude, opencode, codex) in `internal/model` (e.g. alongside `permission_tier.go`). Run 2.1 green.
- [x] 2.3 TRIANGULATE: extend the test with a mixed set (`{AgentGemini, AgentClaude}` → true), an empty set (→ false), and each tier-capable agent individually.
- [x] 2.4 Refactor `internal/tui/permissions_screen.go`: replace the local `tierCapableAgents` map and `anyTierCapable` with calls to `model.TierCapable` (delete the duplicate; update call sites). Run `go test ./internal/tui/...` green — the TUI keeps its existing behavior.

## 3. `--tier` flag on headless install (RED → GREEN → TRIANGULATE)

- [x] 3.1 RED: in `cmd/active-stack/headless/flags_test.go` add cases: `--tier estricto` → `Intent.Tier == model.TierEstricto` and `TUI == false`; `--tier balanceado` alone (no other flags) → `TUI == false`; `--tier loose` → non-nil error; no `--tier` → `Intent.Tier == ""` (zero value).
- [x] 3.2 GREEN: in `ParseInstallFlags` (`flags.go`) register `--tier`, validate via a switch (estricto|balanceado|bypass → `model.PermissionTier`, else error using the invalid value), set `Intent.Tier`, and fold `tier != ""` into the `isHeadless` predicate. Run 3.1 green.
- [x] 3.3 TRIANGULATE: add `--tier bypass` → `TierBypass`, and confirm empty `--tier` combined with `--mode lite` leaves `Intent.Tier == ""` (engine normalizes downstream). Run green.

## 4. Multi-agent `windows options` (RED → GREEN → TRIANGULATE)

- [x] 4.1 RED: update `cmd/active-stack/headless/windows_test.go` — change the existing `TestRunWindowsOptions_JSON` call site to `RunWindowsOptions(cat, []model.Agent{model.AgentClaude}, &out)` (mechanical signature adaptation; keep its 3-modes / forced-`permissions` / no-duplicate assertions). Add a new `TestRunWindowsOptions_MultiAgentUnionDeduped` using a fake catalog where a harness is shared by two agents, asserting it appears exactly once.
- [x] 4.2 GREEN: change `RunWindowsOptions` signature to `(cat install.Catalog, agents []model.Agent, w io.Writer)`, building `install.Intent{Agents: agents, Mode: model.ModeCustom}`. The existing `seen` dedup map already collapses the union. Run 4.1 green.
- [x] 4.3 RED: in `windows_test.go` add `TestRunWindowsOptions_TierMetadata` — `[]model.Agent{model.AgentClaude}` → `tier_capable: true`, `permission_tiers` has 3 entries with `balanceado` marked `default: true`, `tier_capable_agents` contains `claude`; and a gemini-only case → `tier_capable: false`.
- [x] 4.4 GREEN: extend `windowsOptionsResponse` with `tier_capable`, `tier_capable_agents`, `permission_tiers` (add `windowsPermissionTier` struct per design D3). Populate `TierCapable` from `model.TierCapable(agents)`, build the tier list (estricto/balanceado/bypass, default on balanceado, warning on bypass), and the tier-capable subset. Run 4.3 green.
- [x] 4.5 TRIANGULATE: add a multi-agent tier case (`{claude, gemini}` → `tier_capable: true`, `tier_capable_agents` == [claude]) and confirm the 5 pre-existing assertions remain green.

## 5. Dispatch: CSV `--agent` for `windows options` (RED → GREEN → TRIANGULATE)

- [x] 5.1 RED: add/extend `cmd/active-stack/windows_dispatch_test.go` — `windows options --agent claude,opencode` parses two agents and exits 0 with valid JSON; `windows options` with no `--agent` exits non-zero with an error message.
- [x] 5.2 GREEN: in `windows_dispatch.go` `options` case, parse `--agent` as a comma-separated list (trim, skip empties), error when the slice is empty, and pass `[]model.Agent` to `RunWindowsOptions`. Run 5.1 green.
- [x] 5.3 TRIANGULATE: add a whitespace/empty-entry case (`--agent " , claude "` → one agent `claude`) and a single-agent case (backward-compatible) → green.

## 6. Verification

- [x] 6.1 Run `go test ./cmd/active-stack/... ./internal/model/... ./internal/tui/...` — all green, including the 5 original `windows_test.go` tests.
- [ ] 6.2 Manual contract check (only if asked to run): `active-stack windows options --agent claude,codex` returns JSON with the three new additive fields populated and `tier_capable: true`. (Not run — no build was requested; see apply summary.)
- [x] 6.3 Fill in the TDD Cycle Evidence table (task, test file, layer, safety net, RED/GREEN/TRIANGULATE/REFACTOR) in the apply return summary.

## 7. Fix empty `custom_components` — real selectable picker list (RED → GREEN → TRIANGULATE)

> Discovered during apply: `RunWindowsOptions` sourced `custom_components` from `install.SelectHarnesses(cat, Intent{Agents, Mode: ModeCustom})` with an empty `Custom`, which returns ONLY the forced `permissions` harness — so `custom_components` was structurally always `[]`. Fix per design D5: source from a shared `install.CustomPickerHarnesses`, mirroring the TUI. Governance: the new `internal/install` helper is additive and read-only (justified in D5); `internal/backup`/`internal/uninstall` untouched.

- [x] 7.1 Safety net: run `go test ./cmd/active-stack/headless/... ./internal/install/... ./internal/tui/...` and record the passing baseline before touching anything.
- [x] 7.2 RED: in `cmd/active-stack/headless/windows_test.go` add `TestRunWindowsOptions_CustomComponentsNonEmpty` using a fake catalog with at least one non-`permissions` harness available to `claude` plus the `permissions` harness — assert `custom_components` is non-empty, contains the non-`permissions` harness, and does NOT contain `permissions`; assert `forced_components` contains `permissions`. This test MUST FAIL on the current tree (empty `custom_components`).
- [x] 7.3 RED: add `internal/install/custom_picker_test.go` — `TestCustomPickerHarnesses` asserting that for a fake catalog it returns the `ForMode(ModeCustom)` set filtered by agents (includes `permissions` since it is non-starter-only and supports the agent), referencing `install.CustomPickerHarnesses` which does not exist yet.
- [x] 7.4 GREEN: add exported `install.CustomPickerHarnesses(cat Catalog, agents []model.Agent) []model.Harness` in `internal/install` (e.g. `plan.go` or a new `custom_picker.go`) returning `filterByAgents(cat.ForMode(model.ModeCustom), agents)`. Run 7.3 green.
- [x] 7.5 GREEN: in `RunWindowsOptions` (`cmd/active-stack/headless/windows.go`) replace the `install.SelectHarnesses(...)` call with `install.CustomPickerHarnesses(cat, agents)`; keep the existing `seen`-map dedup, the forced-vs-custom split on `install.SecurityFirstHarnessID`, and `Recommended = h.InMode(ModeLite) || h.InMode(ModeFull)`. Run 7.2 green.
- [x] 7.6 TRIANGULATE: add `TestRunWindowsOptions_CustomComponentsUnionDeduped` for `[]model.Agent{claude, opencode}` where a harness supports both — assert it appears exactly once in `custom_components`; add a `recommended`-flag case (a lite/full harness → `recommended: true`; a custom-only harness → `recommended: false`). Run green.
- [x] 7.7 REFACTOR (single source of truth): change `internal/tui/model.go:265-266` custom-picker population to consume `install.CustomPickerHarnesses(cat, agents)`; remove the now-redundant `filterHarnessesByAgents` in `internal/tui/intent.go` if it has no other call sites (otherwise leave it). Run `go test ./internal/tui/...` green — the TUI Custom picker keeps its existing behavior.
- [x] 7.8 Verification: run `go test ./cmd/active-stack/... ./internal/install/... ./internal/tui/... ./internal/model/...` — all green, including the 5 original `windows_test.go` tests and the new custom-components tests.
