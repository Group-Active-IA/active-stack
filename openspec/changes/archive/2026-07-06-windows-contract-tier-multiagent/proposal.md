## Why

The Windows GUI wizard rebuild (planned changes C3–C5) needs an engine contract it does not yet have: the headless install path cannot select a permission tier, the `windows options` command is single-agent only, and the options JSON carries no tier metadata. `install.Intent.Tier` already exists in the engine but is unreachable from the CLI. This change closes that gap so the GUI (and any headless caller) reaches full TUI parity — multi-agent selection plus the estricto/balanceado/bypass permission tier — through the JSON contract.

## What Changes

- Add a `--tier estricto|balanceado|bypass` flag to `active-stack install` / `windows install`. It maps to `install.Intent.Tier`; an invalid value errors; an empty value lets the engine normalize to `balanceado`. `--tier` implies headless mode (same criterion as `--custom`).
- Make `windows options` multi-agent: `RunWindowsOptions` accepts `[]model.Agent`, and `windows_dispatch.go` parses `--agent` as a comma-separated list. Harness selection already dedupes across a slice.
- Extend the `windows options` JSON schema **additively** with tier metadata: `tier_capable` (bool), `tier_capable_agents` ([]string), and `permission_tiers` (array of `{id, label, description, default, warning}`). No existing field is renamed or removed.
- Move the tier-capability knowledge (`tierCapableAgents` / `anyTierCapable`) out of `internal/tui/permissions_screen.go` into `internal/model` as an exported `model.TierCapable(agents) bool`, making it the single source of truth consumed by both the TUI and the Windows options command.
- **Fix `custom_components` being structurally always empty (discovered during apply).** `RunWindowsOptions` sources its component list from `install.SelectHarnesses(cat, Intent{Agents, Mode: ModeCustom})` with an empty `Custom`, which in Custom mode returns ONLY the forced `permissions` harness — so `custom_components` in the `windows options` JSON is always `[]` and the future Windows GUI Custom picker (C4) would have nothing to show. Change the source to mirror the TUI custom picker exactly: the real selectable list is `cat.ForMode(ModeCustom)` filtered by the requested agents (union, deduped by ID). `permissions` stays in `forced_components` (excluded from `custom_components`); `recommended` keeps its current meaning (harness present in lite or full mode). To keep the TUI and the Windows path from diverging, extract the shared list into an exported `install.CustomPickerHarnesses(cat, agents)` consumed by both.

No breaking changes: the JSON contract grows additively, so the existing GUI client and the 5 current `windows_test.go` assertions remain valid.

## Capabilities

### New Capabilities
- `windows-install-contract`: the JSON/CLI contract between the active-stack engine and the native Windows GUI installer — headless install flags (including permission tier), multi-agent `windows options`, and the options metadata schema (modes, components, tier capability, permission tiers).

### Modified Capabilities
<!-- No existing spec files under openspec/specs/ yet; nothing to modify. -->

## Impact

- **Code**: `cmd/active-stack/headless/flags.go` (`ParseInstallFlags` — new `--tier` flag, headless trigger), `cmd/active-stack/headless/windows.go` (`RunWindowsOptions` signature → `[]model.Agent`, additive response fields, `custom_components` sourced from `install.CustomPickerHarnesses`), `cmd/active-stack/windows_dispatch.go` (CSV `--agent` parse for `options`), `internal/model` (new `TierCapable` + tier metadata source of truth), `internal/tui/permissions_screen.go` (consume `model.TierCapable`, drop local duplicate), `internal/install` (new read-only `CustomPickerHarnesses` helper — the shared picker-list source of truth), `internal/tui/model.go` (custom picker population consumes `install.CustomPickerHarnesses`).
- **Contract**: `windows options` JSON gains `tier_capable`, `tier_capable_agents`, `permission_tiers`. Consumed later by C3 (C# parsers) and C4/C5 (WPF wizard pages).
- **Tests**: `cmd/active-stack/headless/flags_test.go`, `cmd/active-stack/headless/windows_test.go` (existing Options call site adapts to the slice signature; assertions stay green), `cmd/active-stack/windows_dispatch_test.go`, plus a `model.TierCapable` unit test.
- **Governance**: MEDIO (`cmd/active-stack/*`, additive contract). `internal/backup` and `internal/uninstall` are NOT modified. `internal/install` gains ONE new read-only helper (`CustomPickerHarnesses`) — additive, no change to `SelectHarnesses`/`BuildPlan` or any destructive path (see design D5 for the governance justification).
