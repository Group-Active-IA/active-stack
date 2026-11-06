## Context

This is change C1 of an approved 6-change plan to rebuild the native Windows GUI installer as a traditional wizard with full TUI parity. The GUI is a thin WPF client over the Go engine, driven through `active-stack.exe windows <subcmd>` with JSON I/O. Before the WPF wizard (C4/C5) and its C# parsers (C3) can be built, the engine contract must expose three things the TUI already has but the headless/Windows path does not:

1. A permission tier on headless install. `install.Intent.Tier` (`internal/install/types.go:44`) exists and the engine already normalizes it, but no CLI flag reaches it.
2. Multi-agent `windows options`. The TUI is multi-select and `active-stack install --agent claude,codex` already accepts a CSV, but `windows options --agent` and `RunWindowsOptions` are single-agent.
3. Tier metadata in the options JSON so the GUI can render the permission screen without a second round trip.

Current tier-capability logic is duplicated and TUI-local (`internal/tui/permissions_screen.go:18-33`). Governance for this change is MEDIO (only `cmd/active-stack/*` and `internal/model` are touched); `internal/backup`, `internal/install`, and `internal/uninstall` are NOT modified.

## Goals / Non-Goals

**Goals:**
- Expose `--tier estricto|balanceado|bypass` on the headless install path, mapped to `install.Intent.Tier`, with invalid-value rejection and empty→balanceado normalization.
- Make `windows options` multi-agent (CSV `--agent`, `RunWindowsOptions([]model.Agent)`), with a deduplicated harness union.
- Extend the `windows options` JSON additively with `tier_capable`, `tier_capable_agents`, and `permission_tiers`.
- Establish `model.TierCapable(agents) bool` as the single source of truth consumed by both TUI and Windows options.

**Non-Goals:**
- Any WPF/C# work (that is C3–C5).
- Modifying `internal/install`, `internal/backup`, or `internal/uninstall` (the tier field and slice-based selection already exist there).
- Changing or removing any existing JSON field (extension is strictly additive).
- The `windows starters|backups|uninstall-options` contract (that is C2).

## Decisions

### D1 — `--tier` is a string flag validated in `ParseInstallFlags`

Add `--tier` to the flag set in `ParseInstallFlags` (`cmd/active-stack/headless/flags.go`). Reuse the existing switch-style validation pattern already used for `--mode`: map `estricto|balanceado|bypass` to the corresponding `model.PermissionTier`, return a descriptive error on any other non-empty value, and leave the zero value when empty (the engine's `PermissionTier.Normalize()` already turns that into `balanceado`). Fold `tier != ""` into the existing `isHeadless` predicate so `--tier` alone triggers headless — mirroring exactly how `--custom` already does it.

*Alternative considered:* validate downstream in the engine. Rejected — the CLI is where the other flags (`--mode`, `--custom`) are validated, and failing fast at parse time keeps the error path consistent and unit-testable without a plan build. `model.PermissionTier.IsValid()` already exists and can back the check.

### D2 — `RunWindowsOptions` takes `[]model.Agent`; dispatch parses CSV

Change `RunWindowsOptions(cat, agent model.Agent, w)` to `RunWindowsOptions(cat, agents []model.Agent, w)` (`cmd/active-stack/headless/windows.go`). It builds `install.Intent{Agents: agents, Mode: model.ModeCustom}` — `install.SelectHarnesses` already accepts a slice and the existing `seen` map in `RunWindowsOptions` already dedupes by harness ID, so the union is deduplicated with no new logic. In `windows_dispatch.go` the `options` case parses `--agent` as a comma-separated list (same trim/skip-empty loop already used in `ParseInstallFlags`) and errors when the resulting slice is empty.

*Compatibility note:* the existing `TestRunWindowsOptions_JSON` call site (`windows_test.go:62`) must change from `RunWindowsOptions(cat, model.AgentClaude, &out)` to pass `[]model.Agent{model.AgentClaude}`. This is a mechanical call-site adaptation; its assertions (3 modes, forced `permissions`, no `permissions` duplicate) are unaffected and stay green.

### D3 — Additive JSON schema extension

Extend `windowsOptionsResponse` with three fields, all with `omitempty`-free explicit encoding so the client always sees them:

```go
type windowsOptionsResponse struct {
    Modes            []windowsModeOption      `json:"modes"`
    ForcedComponents []windowsComponentOption `json:"forced_components"`
    CustomComponents []windowsComponentOption `json:"custom_components"`
    TierCapable       bool                    `json:"tier_capable"`
    TierCapableAgents []string                `json:"tier_capable_agents"`
    PermissionTiers   []windowsPermissionTier `json:"permission_tiers"`
}

type windowsPermissionTier struct {
    ID          string `json:"id"`
    Label       string `json:"label"`
    Description string `json:"description"`
    Default     bool   `json:"default"`
    Warning     string `json:"warning,omitempty"`
}
```

`TierCapable = model.TierCapable(agents)`. `TierCapableAgents` is the subset of the requested agents that are tier-capable (lets the GUI recompute capability locally when the user toggles agents, avoiding a re-query). `PermissionTiers` lists the three tiers in display order (`estricto`, `balanceado`, `bypass`) with `Default: true` on `balanceado` and the bypass warning text on the bypass entry. Because every field is new, no existing consumer breaks.

*Alternative considered:* a separate `windows tiers` subcommand. Rejected — it would force the GUI into an extra round trip and split tier data from the agent context that determines capability.

### D4 — `model.TierCapable` as the single source of truth

Move `tierCapableAgents` (the map) and the `anyTierCapable` logic from `internal/tui/permissions_screen.go` into `internal/model` (e.g. alongside `permission_tier.go`) as an unexported `tierCapableAgents` set plus an exported `func TierCapable(agents []model.Agent) bool`. The TUI's `anyTierCapable` becomes a thin call to `model.TierCapable` (or is deleted and callers updated), and `windows.go` uses the same function. This guarantees the TUI's "skip the permission screen" decision and the GUI's `tier_capable` flag can never diverge. The tier display order / default / warning strings that today live in the TUI can also be exposed from the model (or kept in the windows layer) so `permission_tiers` and the TUI radio list read from one place — implementer's call, but capability membership MUST be shared.

*Alternative considered:* leave the set in TUI and have windows import `internal/tui`. Rejected — the CLI/headless layer importing the TUI package is a dependency inversion; the model is the correct shared home.

### D5 — `custom_components` sources from a shared `install.CustomPickerHarnesses`, not `SelectHarnesses`

**Problem (found during apply, confirmed with a debug probe).** `RunWindowsOptions` builds its component list from `install.SelectHarnesses(cat, install.Intent{Agents, Mode: model.ModeCustom})` with an empty `Custom` slice. In Custom mode `SelectHarnesses` iterates `intent.Custom` (empty) and then force-adds only `SecurityFirstHarnessID` (`internal/install/plan.go:173-199`). So the returned set is always exactly `[permissions]`, and after windows.go splits forced vs. custom, `custom_components` is **structurally always `[]`**. `SelectHarnesses` answers *"what will be installed for this intent"* — with an empty custom intent, correctly nothing but the forced overlay. It is the wrong question for *"what can the user pick"*. The TUI does not hit this because its Custom picker populates from the catalog directly: `all := cat.ForMode(model.ModeCustom); filterHarnessesByAgents(all, agents)` (`internal/tui/model.go:265-266`). `ForMode(ModeCustom)` returns every non-starter-only harness (`Harness.InMode` returns true for `ModeCustom` unconditionally — `internal/model/harness.go:154-166`), then the agent filter narrows it.

**Decision.** Introduce an exported, read-only helper in `internal/install`:

```go
// CustomPickerHarnesses returns the harnesses a user may choose in Custom mode
// for the given agents: every catalog harness available in Custom mode, filtered
// to those supporting at least one requested agent (union, deduped by the
// catalog's own ordering). It answers "what can be picked", NOT "what will be
// installed" (that is SelectHarnesses). Read-only; it forces nothing.
func CustomPickerHarnesses(cat Catalog, agents []model.Agent) []model.Harness {
    return filterByAgents(cat.ForMode(model.ModeCustom), agents)
}
```

It reuses the already-existing unexported `filterByAgents` (`internal/install/plan.go:208`), which is the exact agent-union filter the TUI duplicates as `filterHarnessesByAgents`. `RunWindowsOptions` then:

1. calls `install.CustomPickerHarnesses(cat, agents)` to get the picker universe,
2. keeps the existing `seen`-map dedup by ID (D2's union dedup is unchanged),
3. keeps the existing split: `h.ID == install.SecurityFirstHarnessID` → `forced_components`; everything else → `custom_components`,
4. keeps `Recommended = h.InMode(model.ModeLite) || h.InMode(model.ModeFull)` (unchanged semantics).

Because `ForMode(ModeCustom)` includes `permissions` (it is non-starter-only and supports claude/codex/opencode/gemini/vscode), the forced/custom split still yields a populated `forced_components` with `permissions` AND a populated `custom_components` with the real selectable harnesses.

**Single source of truth.** To guarantee the Windows Custom picker and the TUI Custom picker can never diverge (the same D4 principle applied to the picker list), refactor `internal/tui/model.go:265-266` to consume `install.CustomPickerHarnesses` as well, and delete the now-redundant `filterHarnessesByAgents` in `internal/tui/intent.go` (or leave it if other call sites remain — implementer's call; the picker population MUST route through the shared helper).

**Governance justification (required — `internal/install` is above LOW).** The change to `internal/install` is a single **additive, read-only** function that composes two existing read primitives (`ForMode`, `filterByAgents`). It does not touch `SelectHarnesses`, `BuildPlan`, or any snapshot/inject/rollback path, so it cannot destroy user config — it stays clear of the ALTO-risk surface. The agent-union filter (`filterByAgents`) already lives in `internal/install`, making it the natural home; putting it in `catalog` would require widening the `install.Catalog` interface with an agent-aware method, a larger change. Net: the cleanest option is the small helper in `internal/install`.

*Alternatives considered:*
- *Inline `cat.ForMode(ModeCustom)` + a local agent filter in `windows.go`.* Rejected — re-duplicates `filterByAgents` a third time (install, tui, windows) and leaves windows and the TUI as independent implementations that can drift.
- *Add `ForModeAndAgents` to the catalog.* Rejected — widens the `install.Catalog` interface and every fake in the test suite, for logic that already exists in `install.filterByAgents`.
- *Keep `SelectHarnesses` and pass the full catalog IDs as `Custom`.* Rejected — abuses the "what will be installed" selector to answer a "what can be picked" question, and would force-carry strict unknown-id semantics into a display path.

## Risks / Trade-offs

- **[Existing 5 `windows_test.go` assertions must stay green]** → The only source-level break is the `RunWindowsOptions` signature; adapt that one call site to a slice. All field assertions are additive-safe. This is verified by keeping the 5 tests and running them green after the change.
- **[Tier metadata strings duplicated between TUI and windows layer]** → Mitigate by sharing at least capability membership via `model.TierCapable`; optionally lift the tier order/labels to the model too. Accept minor duplication of display copy if lifting labels bloats the model, but never duplicate the capability set.
- **[GUI (C3) depends on exact JSON field names]** → Lock the field names here (`tier_capable`, `tier_capable_agents`, `permission_tiers`) and require C3's C# fixtures to be copied byte-for-byte from these Go test outputs (schema-sync rule from the plan).

## Migration Plan

Additive and backward compatible — no data migration. The old GUI binary keeps working against the new engine (it ignores the new fields); a new GUI against an old engine sees the fields absent and falls back to defaults (handled in C3). Rollback is a straight revert of the C1 commit; no persisted state is involved.

## Open Questions

- Whether the tier display order, default marker, and bypass warning text should be fully lifted into `internal/model` or stay in the windows layer. Non-blocking: capability membership sharing is mandatory; label placement is an implementer decision resolved during TDD.
