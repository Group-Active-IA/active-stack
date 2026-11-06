## 1. Go: extend the windows detect JSON contract

- [x] 1.1 Write a test (RED) in `cmd/active-stack/headless/windows_test.go` asserting `RunWindowsDetect`'s output includes a `dependencies` field shaped from `system.DetectDependencies`
- [x] 1.2 Add `Dependencies` to `windowsDetectResponse` in `cmd/active-stack/headless/windows.go`, sourced from `system.DetectDependencies(ctx, profile)` (GREEN) — sourced via `system.Detect(ctx).Dependencies`/`.System.Profile` rather than calling `DetectDependencies` directly, so the profile used matches the real host
- [x] 1.3 Write a test (RED) asserting each dependency entry carries a pre-joined `install_command` string built from `system.InstallCommandsForDep`, empty when nil
- [x] 1.4 Implement the `install_command` join (GREEN) — implemented together with 1.2 in `toWindowsDependenciesResponse`
- [x] 1.5 Write a second test (TRIANGULATE) with a different profile (e.g. a platform where a dependency has no automatic command) confirming `install_command` is empty and `install_hint` is still populated — covered by `TestToWindowsDependenciesResponse_NoCommandFallsBackToHint` and `TestToWindowsDependenciesResponse_InstalledDependencyHasNoCommand` in the new `windows_dependencies_test.go`
- [x] 1.6 Write a test (RED) asserting `detected_agents`'s value and shape are byte-identical to before this change (regression guard for the additive-field claim)
- [x] 1.7 Run `go test ./cmd/... ./internal/system/...` — confirm green

## 2. .NET: parse the new field

- [x] 2.1 Write a test (RED) in `ActiveStack.Bootstrapper.Core.Tests` (mirrors `BackupListParserTests.cs`) asserting `InstallerSessionStateBuilder.BuildFromJson` parses a `dependencies` fixture into a `DependencyChoice` list on `InstallerSessionState`
- [x] 2.2 Add `WindowsDetectResponse.Dependencies` (internal DTO) and the public `DependencyChoice` record + `InstallerSessionState.Dependencies` property in `InstallerSessionStateBuilder.cs` (GREEN)
- [x] 2.3 Write a test (RED) asserting a missing dependency's `install_hint` and `install_command` round-trip into `DependencyChoice`
- [x] 2.4 Implement that mapping (GREEN) — implemented together with 2.2; this test held on first run, confirming the mapping
- [x] 2.5 Write a test (RED) — second fixture (TRIANGULATE): all dependencies present vs. a mix of missing required/optional
- [x] 2.6 Run `dotnet test` on `ActiveStack.Bootstrapper.Core.Tests` — confirm green (52/52)

## 3. .NET: wire the Dependencies page into navigation

- [x] 3.1 Add `Dependencies` to `WizardPageId` enum
- [x] 3.2 Update `WizardFlow.NextPage`/`PreviousPage`: `Hub => Dependencies`, `Dependencies => Assistants` (forward); `Assistants => Dependencies`, `Dependencies => Hub` (backward)
- [x] 3.3 Write a test (RED) in `ActiveStack.Bootstrapper.Host.Tests` asserting `WizardFlow.NextPage(Hub, ...)` returns `Dependencies` and `NextPage(Dependencies, ...)` returns `Assistants`
- [x] 3.4 Confirm the test passes (GREEN) — should already hold from 3.2
- [x] 3.5 Write a test (RED) asserting `PreviousPage(Dependencies, ...)` returns `Hub` and `PreviousPage(Assistants, ...)` returns `Dependencies`
- [x] 3.6 Confirm GREEN
- [x] 3.7 Run the full existing `WizardFlow` test suite — confirm no regression in the other flow orderings (Uninstall/Starters/Backups untouched) — 17/17 green in `WizardFlowTests.cs`; two pre-existing tests (`NextPage_FromHub_GoesToAssistants`, `PreviousPage_FromAssistants_GoesToHub`) asserted the old direct Hub↔Assistants hop and were updated to assert the new two-hop chain through Dependencies, per the MODIFIED requirement in `specs/windows-gui-wizard`. Note: the broader `Host.Tests` project now has 20 failing `ShellViewModelTests` because `ShellViewModel.CreatePage` doesn't yet construct a page for `Dependencies` — expected, fixed by group 4

## 4. .NET: Dependencies page view-model and view

- [x] 4.1 Write a test (RED) asserting a new `DependenciesPageViewModel` exposes the session's dependency list and `CanAdvance` is always `true`
- [x] 4.2 Implement `DependenciesPageViewModel` in `Pages/Install/` (GREEN) — plus new `DependencyRow` record, `WindowsDetectResponse`/`WindowsDependenciesResponse`/`WindowsDependencyEntry` DTOs already existed from group 2; added `UiStrings` keys (en+es, parity-checked) for the page title/subtitle and row status/label text
- [x] 4.3 Write a test (RED) asserting missing dependencies expose their install hint and install command as displayable strings on the view-model (no execution capability anywhere on the type)
- [x] 4.4 Implement that mapping (GREEN) — implemented together with 4.2; also verified installed dependencies actively suppress hint/command (not just pass-through)
- [x] 4.5 Add the Dependencies page view + `DataTemplate` entry in `Themes/PageTemplates.xaml`, following the existing row-list + detail-panel pattern (e.g. Components page) — used the plain `ItemsControl` row-list pattern (like Assistants) rather than the detail-panel variant, since there's no "long description" concept for a dependency; install hint/command render inline per-row as read-only text
- [x] 4.6 Wire `ShellViewModel` to instantiate `DependenciesPageViewModel` when the current page is `WizardPageId.Dependencies` — this also required fixing 15 pre-existing `ShellViewModelTests` that hardcoded the old direct Hub→Assistants hop (now Hub→Dependencies→Assistants): inserted one `AdvanceAsync()`/`GoBack()` step at each Install-flow Hub-selection and Assistants-return point
- [~] 4.7 Manual smoke: build and run the bundle locally (`installer/windows/build.ps1 -SkipBundleBuild` per `installer/windows/README.md`), walk Hub → Install → confirm the Dependencies page renders before Assistants and Next always advances — build confirmed green end-to-end (Go payload + self-contained WPF publish, both succeeded). The actual visual walkthrough was NOT run by me: this is an interactive desktop GUI app, and launching/killing it blind from here would pop a window on your screen with no way for me to observe it — not a real check. **Needs you to run it once**: launch `installer\windows\payload\ba\ActiveStack.Bootstrapper.Host.exe` and confirm the Dependencies page appears between Hub and Assistants, shows real detected dependencies, and Next always advances

## 5. Spec sync check

- [x] 5.1 Re-read `specs/windows-gui-wizard/spec.md`'s MODIFIED requirement against the implemented `WizardFlow` — confirm the stated order (`Hub → Dependencies → Assistants → ...`) matches exactly — confirmed by the 17/17 `WizardFlowTests` assertions, which encode the order literally
- [x] 5.2 Re-read `specs/windows-hub-contract/spec.md`'s ADDED requirement against the actual `windows detect` JSON output — confirm field names match exactly (`dependencies`, `install_command`, etc.) — ran `go run ./cmd/active-stack windows detect` for real and diffed field names against the spec text by hand; exact match

## 6. Refactor pass

- [x] 6.1 Re-run `go test ./...` and `dotnet test` (both `.Tests` projects) together — confirm both toolchains green after any cleanup — Go: build+vet+all tests green; Core.Tests 52/52; Host.Tests 193/193
