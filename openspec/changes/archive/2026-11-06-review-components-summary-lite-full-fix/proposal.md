## Why

In Lite and Full install modes the Review page's components summary
(`ReviewPageViewModel.ComponentsSummary`) always renders **blank**, so the user
confirms an install without ever seeing what it will actually install — even
though a real, fixed component set is installed for those modes. This is the
symptom the prior `installer-post-smoke-bugfixes` change filed as an
out-of-scope "Go-data" follow-up (its task 4.4); that attribution was wrong. The
integration test it relied on only exercised the **Custom** path, where the
Components page is visited — it never covered Lite/Full, where the page is
skipped. The defect is 100% in the C# GUI layer.

## What Changes

- Make `ReviewPageViewModel.ComponentsSummary` **mode-aware** so it reflects
  what will actually be installed regardless of install mode:
  - **Custom** mode: unchanged — the summary is built from
    `InstallSelection.CustomIds` (populated by the Components page's
    `SyncSelection`), which is correct today.
  - **Lite / Full** mode (Components page intentionally skipped, so `CustomIds`
    is legitimately empty): the summary is built from the session's forced
    components plus the components the session marks as included in the default
    ("recommended") set, instead of the always-empty `CustomIds`.
- Add regression coverage that exercises the **Lite/Full** path end to end
  (`WizardFlow` skips Components → `CustomIds` stays empty → Review still shows a
  non-empty, correct component summary), closing the gap the prior change's
  Custom-only integration test left open.
- No change to `InstallSelection.CustomIds` semantics or to what is sent to the
  engine: `CustomIds` stays engine-faithful (empty for Lite/Full, since the
  engine selects the mode's set itself). The fix lives purely in the read-only
  display layer.

## Capabilities

### New Capabilities

<!-- None. This is a bugfix to existing wizard behavior. -->

### Modified Capabilities

- `windows-gui-wizard`: the "Review confirms the selection and triggers install"
  requirement gains mode-aware component-summary behavior — the components
  summary MUST be non-empty and reflect the mode's implied component set for
  Lite/Full, not only for Custom.

## Impact

- **Affected code (C# only):**
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host/Pages/Install/ReviewPageViewModel.cs`
    — mode-aware `ComponentsSummary` computation.
  - `installer/windows/ui/ActiveStack.Bootstrapper.Host.Tests/**` — new/updated
    `ReviewPageViewModelTests` and a Lite/Full flow integration test.
- **No Go code** (`cmd/`, `internal/`) is touched. The engine's per-mode install
  behavior is already correct; only the GUI's summary rendering is wrong.
- **Known limitation (documented in design.md), not fixed here:** the
  `windows options` payload exposes a single `recommended` flag per component
  equal to `InMode(lite) || InMode(full)`. That is exact for **Full** (the
  default mode) but **over-inclusive for Lite** (Full-only skills also carry
  `recommended=true`). Making the Lite summary component-exact would require the
  Go `windows options` payload to expose per-mode membership — an explicit
  **out-of-scope follow-up** (it touches Go). This change fixes the blank-summary
  bug (Full exactly, Lite non-empty and mode-scoped).

## Out-of-scope follow-ups (logged, not fixed here)

- **Stale agent list in `ProcessInstallerEngineClient.LoadSessionAsync`**
  (`ProcessInstallerEngineClient.cs`, ~lines 32-38): the session is loaded once,
  lazily, calling `windows options --agent <csv>` with the agents from
  `windows detect` (what is already installed on the machine) rather than the
  user's actual selection on the Assistants page — because the load happens
  before that page is shown. This can make the Custom picker's component list
  stale relative to the agents the user later selects. It is a **distinct** bug
  from the blank-summary defect fixed here, does not gate this fix, and is not a
  one-line change (it needs a session re-fetch/cache-invalidation after the
  Assistants page). Explicitly **out of scope** for this change.
- **Per-mode component granularity in the Go `windows options` payload** (see
  Known limitation above) — required only to make the **Lite** summary
  component-exact; out of scope because it touches Go.
