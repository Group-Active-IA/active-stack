## Why

`ActiveStack-Setup.exe` (the WiX Burn bundle published in Releases, the installer most end users will actually run) never checks for Git or Node.js. `RunWindowsDetect` in `cmd/active-stack/headless/windows.go` — the Go command the bundle's `.NET`/WPF UI calls to populate its first data-driven page — only returns `detected_agents`. If Git or Node is missing, the user finds out only when the underlying `active-stack windows install` invocation fails partway through, surfaced as a generic install-failed event in the progress page, with no upfront warning or link to fix it.

This is the third of three related changes (`readme-prereqs` documented the prerequisites in prose; `tui-dependency-detection` added the equivalent proactive screen to the CLI TUI, reusing `internal/system/deps.go`). This change closes the gap for the graphical installer, using the same underlying Go detection logic and the same "show it, don't auto-install it" stance already agreed for the TUI.

## What Changes

- Extend `windowsDetectResponse` (`cmd/active-stack/headless/windows.go`) with a `Dependencies` field carrying `system.DetectDependencies(ctx, profile)`'s report (name, required/optional, installed, version, install hint, and platform install command).
- Add a `WizardPageId.Dependencies` page to the .NET wizard (`installer/windows/ui/ActiveStack.Bootstrapper.Host/Navigation/WizardPageId.cs`), inserted between `Hub` and `Assistants` in `WizardFlow.cs` — the direct GUI analog of the TUI's `Welcome → Detection → Agents` ordering.
- Add the matching parser (mirrors `BackupListParser.cs`/`UninstallOptionsParser.cs` in `ActiveStack.Bootstrapper.Core`) and extend `InstallerSessionStateBuilder`/`WindowsDetectResponse` on the .NET side to carry the new `Dependencies` data into a new `DependenciesPageViewModel` and its page.
- For missing dependencies, the new page displays the official install hint and the platform install command as **selectable, copyable text** — no button that shells out and runs it. Same stance as `tui-dependency-detection`: facilitate, never auto-install.
- Go stays `Required: false` per the earlier decision (build-from-source-only, no catalog harness needs it at runtime today) — the GUI installer never claims Go is required either.

## Capabilities

### New Capabilities
- `gui-installer-dependency-detection`: the graphical installer's wizard shows a dependency-detection page (Git/curl/Node/npm/Go status, install hints, copyable install commands) before the Assistants page, without blocking navigation or auto-installing anything.

### Modified Capabilities
- `windows-hub-contract`: the `windows detect` JSON contract gains a `dependencies` field (additive — existing `detected_agents` field and its consumers are unaffected).
- `windows-gui-wizard`: the wizard page graph gains a `Dependencies` step between `Hub` and `Assistants`.

## Impact

- `cmd/active-stack/headless/windows.go` (`windowsDetectResponse`, `RunWindowsDetect`)
- `installer/windows/ui/ActiveStack.Bootstrapper.Core/InstallerSessionStateBuilder.cs` (new `WindowsDetectResponse.Dependencies`, new `DependencyChoice`-style record)
- `installer/windows/ui/ActiveStack.Bootstrapper.Host/Navigation/WizardPageId.cs`, `WizardFlow.cs`
- New `installer/windows/ui/ActiveStack.Bootstrapper.Host/Pages/Install/DependenciesPageViewModel.cs` + its XAML page + `Themes/PageTemplates.xaml` entry
- Reuses without modification: `internal/system/deps.go`, `install_deps.go` (already implemented)
- Two toolchains touched (Go + .NET/WiX) — governance MEDIO per this project's CLAUDE.md (`harness`/installer-adjacent code); implement with checkpoints before the .NET side lands
