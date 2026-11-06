## ADDED Requirements

### Requirement: Stream failures surface a terminal Error page, never a silent exit

When the shell consumes an install, uninstall, or starter progress stream (`ShellViewModel.RunStreamAndAdvanceToCompleteAsync`), any exception thrown while consuming that stream — a non-zero engine exit, a parse failure, or any other exception — SHALL be caught and converted into navigation to the shared `CompletePageViewModel` in the `CompleteState.Error` terminal state, carrying the failure's message. The exception MUST NOT propagate past the shell (so it can never reach the WPF dispatcher and terminate the process silently). This targeted catch is the primary mechanism; a global unhandled-exception handler is a secondary net and MUST NOT be relied on as the primary path.

#### Scenario: A stream that throws mid-enumeration lands on the Error page

- **WHEN** the current page is a stream-trigger page whose `StartStream` yields an `IAsyncEnumerable` that throws partway through enumeration, and the shell advances
- **THEN** `AdvanceAsync` does not propagate the exception, the current page becomes a `CompletePageViewModel` whose `State` is `CompleteState.Error`, and the caught failure's message is surfaced on that page

#### Scenario: A non-zero engine exit reaches the Error page instead of closing the window

- **WHEN** an install/uninstall/starter stream fails because the engine process exits non-zero (surfaced as a thrown exception from the streaming client)
- **THEN** the wizard shows the terminal Error page rather than the process terminating with no user-facing signal

### Requirement: List pages wrap content instead of scrolling horizontally

Every `ListBox` in the wizard page templates (`Themes/PageTemplates.xaml`) SHALL disable its horizontal scrollbar (`ScrollViewer.HorizontalScrollBarVisibility="Disabled"`) so long item content wraps within the available width rather than exposing a horizontal scrollbar. This applies uniformly to all list-bearing pages (Language, Hub, InstallType, Components, Permissions, UninstallMode, UninstallStrategy strategy and backup lists, and StarterCatalog), not only the page where the defect was first observed.

#### Scenario: A long list item wraps rather than scrolling sideways

- **WHEN** a list page (e.g. Permissions) renders an item whose text is wider than the list's available width
- **THEN** the item's text wraps within the list and no horizontal scrollbar is shown

## MODIFIED Requirements

### Requirement: Review confirms the selection and triggers install

The Review page SHALL present a read-only summary of the selection (agents, install type, chosen components, tier) and the shell's primary action on Review SHALL read "Install". Advancing from Review SHALL start the install by invoking the engine client's install stream with the full selection (agents, mode, custom ids, tier). The chosen custom components summarized on Review SHALL be derived from the same `InstallSelection` the Components page writes, so the components the user selected in `custom` mode are exactly those shown on Review and sent to the engine.

#### Scenario: Review primary action starts the install with the full selection

- **WHEN** the user confirms on the Review page with agents, mode, custom ids, and tier set
- **THEN** the install stream is started carrying all of those values (not a single agent and not without tier/custom)

#### Scenario: Custom components selected on Components appear in the Review summary

- **WHEN** the user toggles a set of components on the Components page (writing them into the shared `InstallSelection`) and reaches the Review page built over that same selection
- **THEN** the Review page's components summary lists exactly the selected components (including the forced `permissions` harness), matching what will be sent to the engine
