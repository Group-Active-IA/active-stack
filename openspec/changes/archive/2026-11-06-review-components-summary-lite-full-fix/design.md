## Context

The WPF Windows bootstrapper GUI (`installer/windows/ui/**`) drives the install
wizard through pure `WizardFlow.NextPage`/`PreviousPage` functions over a shared
`InstallSelection`. The Components page is only routed to when the user picks
**Custom** mode (`WizardFlow.cs`, ~lines 23-26); for Lite/Full it is skipped by
design (the mode implies a fixed set — there is nothing to pick).

`ReviewPageViewModel` (`Host/Pages/Install/ReviewPageViewModel.cs`, ~lines 41-44)
builds `ComponentsSummary` purely from `InstallSelection.CustomIds`:

```csharp
var allComponents = session.ForcedComponents.Concat(session.CustomComponents);
ComponentsSummary = string.Join(", ", allComponents
    .Where(c => _selection.CustomIds.Contains(c.Id, StringComparer.OrdinalIgnoreCase))
    .Select(c => c.Label));
```

`InstallSelection.CustomIds` is only ever populated by
`ComponentsPageViewModel.SyncSelection()`. Because that page never runs in
Lite/Full, `CustomIds` stays empty and the summary renders blank for every
Lite/Full install — even though the engine installs a real, mode-determined set.

**Why the prior change missed it.** `installer-post-smoke-bugfixes` added a
`ComponentsToReviewIntegrationTests` and concluded (task 4.4) the C# wiring was
correct and the defect must be in Go data. That test only drove the **Custom**
path (Components page present → `SyncSelection` runs → `CustomIds` populated). It
never covered the Lite/Full path where the page is absent. So the misattribution
was a test-coverage gap, not a Go bug.

**Engine ground truth (read-only reference, not modified).** For Lite/Full the Go
engine selects `cat.ForMode(mode)` intersected with the selected agents
(`internal/install/plan.go` `SelectHarnesses`, default branch). The GUI runs in a
separate process and cannot call that; it only has the `windows options` JSON,
which exposes `forced_components`, `custom_components`, and a per-component
`recommended` flag computed as `InMode(lite) || InMode(full)`
(`cmd/active-stack/headless/windows.go`, ~line 123).

## Goals / Non-Goals

**Goals:**

- `ReviewPageViewModel.ComponentsSummary` is non-empty and reflects the mode's
  implied component set for **Lite/Full**, matching existing correct behavior for
  Custom.
- Close the test-coverage gap: pin the Lite/Full path (Components skipped →
  summary still correct) with a durable regression test.
- Keep the fix in the read-only display layer; do not change what is sent to the
  engine.

**Non-Goals:**

- No Go changes (`cmd/`, `internal/`). The engine's per-mode install behavior is
  already correct.
- No change to `InstallSelection.CustomIds` semantics or to the install argv.
- Not fixing the stale-agent-list issue in
  `ProcessInstallerEngineClient.LoadSessionAsync` (separate bug, out of scope —
  see proposal.md).
- Not achieving component-exact Lite precision (requires Go payload change — see
  D3 and Risks).

## Decisions

### D1: Fix in the display layer (`ReviewPageViewModel`), not in `InstallSelection`

The defect is a **display** defect: the selection state itself is already
engine-faithful (empty `CustomIds` for Lite/Full is correct, because the engine
picks the mode's set itself and ignores custom ids in non-custom modes —
`SelectHarnesses` default branch). Therefore the summary must be computed in the
read-only page from the mode + session, leaving `InstallSelection` untouched.

**Alternative considered — populate `CustomIds` for Lite/Full** (e.g. seed it in
`InstallTypePageViewModel` from forced+recommended): rejected. `CustomIds` is
handed to `RunInstallAsync` and sent to the engine; populating it for a
non-custom mode makes the selection state lie about the user's choice, is only
"harmless" because the engine happens to ignore it today, and couples display
concerns to the engine payload. The display layer is the correct, tightest home.

### D2: Mode-aware `ComponentsSummary`

`ComponentsSummary` branches on `InstallSelection.Mode`:

- **Custom** (`Mode == "custom"`, case-insensitive): unchanged — the set is the
  session components whose id is in `CustomIds` (which already includes the
  forced `permissions` id and the user's picks). This preserves every existing
  Custom scenario.
- **Non-custom** (Lite/Full): the set is the session's `ForcedComponents` plus
  the `CustomComponents` whose `Recommended` flag is true — i.e. the components
  the engine marks as part of the mode's default set — rendered by `Label`, in
  the session's order (forced first, then recommended custom), deduped.

Both branches reuse the existing `session.ForcedComponents` /
`session.CustomComponents` data already on the page; no new session field is
introduced.

**Alternative considered — always union `CustomIds` with forced+recommended**:
rejected. In Custom mode the user may deliberately deselect a recommended
component; unioning would re-add it and misreport the Custom summary. Branching on
mode keeps Custom authoritative from `CustomIds` and Lite/Full derived from the
mode's implied set.

### D3: Accept Full-exact / Lite-approximate, log Lite-exact as a Go follow-up

`recommended = InMode(lite) || InMode(full)`. Because the project convention is
Full ⊇ Lite and Full-only harnesses are `InMode(full) == true`, the recommended
set equals **exactly the Full default set** (forced + recommended custom is
component-exact for Full, the wizard's default mode). For **Lite** it is
**over-inclusive**: Full-only skills (e.g. `kb-creator`, `roadmap-generator`,
`find-skill`, `active-orchestrator`) carry `recommended = true` and would appear
even though Lite does not install them.

The current `windows options` payload has **no per-mode granularity** to separate
lite from full — one boolean per component is all the GUI receives. Making the
Lite summary component-exact requires the Go payload to expose per-mode
membership (e.g. a `modes: ["lite","full"]` array per component, or separate
`lite`/`full` inclusion flags). That touches Go and is therefore an explicit
**out-of-scope follow-up**. This change deliberately ships: blank → fixed;
**Full exact**; **Lite non-empty and mode-scoped** (over-inclusive by at most the
Full-only skills). This is a strict improvement over today's always-blank summary
and unblocks the default (Full) path completely.

### D4: Regression coverage drives the previously-uncovered Lite/Full path

Tests (TDD, `dotnet test`) MUST:

- Construct a `ReviewPageViewModel` over an `InstallSelection` with `Mode` = full
  (and lite) and **empty** `CustomIds`, over an `InstallerSessionState` with
  forced + recommended + non-recommended custom components, and assert
  `ComponentsSummary` lists forced + recommended labels (and excludes
  non-recommended custom) — this fails today (blank) and passes after D2.
- Preserve the Custom-mode behavior: a Custom selection still summarizes exactly
  from `CustomIds` (the prior `ComponentsToReviewIntegrationTests` shape stays
  green).

## Risks / Trade-offs

- **[Lite over-reports Full-only skills]** → Documented limitation (D3); the
  summary is honest for Full (default) and non-empty/mode-scoped for Lite. Exact
  Lite is a logged Go follow-up. Chosen over shipping a still-blank Lite summary
  or a Go change that the change's constraints forbid.
- **[Mode string mismatch]** (`Mode` compared to `"custom"`): use
  `StringComparison.OrdinalIgnoreCase`, consistent with the existing comparisons
  in `ReviewPageViewModel` (agents, mode, tier). At Review `Mode` is always set;
  treat any non-"custom" value as the mode-set branch.
- **[Divergence from engine truth]** → The GUI cannot compute `ForMode ∩ agents`
  out of process; forced + recommended is the closest faithful proxy from the
  available payload and is exact for the common (Full) case.

## Migration Plan

Not applicable — internal GUI bugfix, no data migration, no external contract
change. Rollback is reverting the `ReviewPageViewModel` change; the added tests
document the intended behavior.

## Open Questions

None. The Lite-exactness question is resolved as an out-of-scope Go follow-up
(D3); the stale-agent-list issue is resolved as out of scope (proposal.md).
