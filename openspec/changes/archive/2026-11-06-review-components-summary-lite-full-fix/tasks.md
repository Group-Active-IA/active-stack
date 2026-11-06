<!-- Strict TDD (C#, `dotnet test`): every group follows RED → GREEN → TRIANGULATE → REFACTOR.
     Before modifying a file, run its existing tests as a safety-net baseline; a pre-existing
     failure is REPORTED, not fixed. Tests MUST drive the real ComponentsSummary path over a
     real InstallSelection + InstallerSessionState (never a mock of the summary itself).
     No Go code (`cmd/`, `internal/`). No commit. No build/bundle (operator-only, later). -->

## 1. Safety-net baseline

- [x] 1.1 Run `ReviewPageViewModelTests` (and, if present, `ComponentsToReviewIntegrationTests`) in `ActiveStack.Bootstrapper.Host.Tests` and capture the passing baseline counts. A pre-existing failure is reported to the orchestrator, not fixed. **Baseline: 6/6 passing, 0 failures.**

## 2. RED — Lite/Full summary is blank today (drives the fix)

- [x] 2.1 RED: add a `ReviewPageViewModelTests` case that builds an `InstallSelection` with `Mode = "full"` and **empty** `CustomIds`, over an `InstallerSessionState` whose `ForcedComponents` include `permissions` and whose `CustomComponents` include at least one recommended (`Recommended: true`) and one non-recommended (`Recommended: false`) entry; assert `ComponentsSummary` is non-empty, contains the forced label and the recommended custom label, and does NOT contain the non-recommended custom label. Confirm it FAILS today (summary renders blank because `CustomIds` is empty). **Confirmed RED: `Assert.NotEmpty() Failure: Collection was empty`.**

## 3. GREEN — mode-aware ComponentsSummary (D2)

- [x] 3.1 GREEN: in `Host/Pages/Install/ReviewPageViewModel.cs`, make `ComponentsSummary` branch on `_selection.Mode` (`StringComparison.OrdinalIgnoreCase`):
  - `custom` → current behavior: labels of `ForcedComponents.Concat(CustomComponents)` whose `Id` is in `_selection.CustomIds`.
  - non-custom (lite/full) → labels of `ForcedComponents` plus `CustomComponents.Where(c => c.Recommended)`, in session order (forced first), deduped.
  Do NOT touch `InstallSelection`, `RunInstallAsync`, or the custom ids sent to the engine. Run until 2.1 passes. **Confirmed GREEN: 7/7 passing.**

## 4. TRIANGULATE — cover lite, custom-preserved, and the recommended boundary (D4)

- [x] 4.1 TRIANGULATE: add a `Mode = "lite"` case (empty `CustomIds`) asserting the summary is non-empty and lists forced + recommended custom labels (mirrors 2.1's full case; guards against a full-only hardcode).
- [x] 4.2 TRIANGULATE: add/keep a `Mode = "custom"` case with a non-empty `CustomIds` set (e.g. `["permissions", "openspec"]`) asserting the summary lists exactly those labels — including asserting a recommended custom component that is NOT in `CustomIds` is EXCLUDED (proves Custom stays authoritative from `CustomIds`, not unioned with recommended). The existing `Constructor_SummaryReflectsAgentsModeComponentsAndTier` custom case must stay green.
- [x] 4.3 Confirm all `ReviewPageViewModelTests` cases (and `ComponentsToReviewIntegrationTests` if present) are green. **Confirmed: 9/9 passing.**

## 5. REFACTOR — tidy without changing behavior

- [x] 5.1 REFACTOR: if it clarifies, extract the two summary branches into a small private helper (e.g. `BuildComponentsSummary(session, selection)`); keep the existing `StringComparison.OrdinalIgnoreCase` convention used by the agents/mode/tier summaries. Run tests after each step — green throughout. **Extracted `BuildComponentsSummary`; 9/9 stayed green.**

## 6. Suite confirmation

- [x] 6.1 Run the full `ActiveStack.Bootstrapper.Host.Tests` suite; confirm 0 failures including every pre-existing test (safety net), and record the pass count (baseline + new cases). No build/bundle and no commit unless the operator explicitly asks. **Confirmed: Host.Tests 157/157 passing (baseline 6 + 3 new cases + all pre-existing), Core.Tests 47/47 passing. 0 failures across both suites.**
