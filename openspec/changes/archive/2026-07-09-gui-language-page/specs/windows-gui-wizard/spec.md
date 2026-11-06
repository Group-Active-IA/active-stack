## MODIFIED Requirements

### Requirement: Wizard shell hosts logical page navigation

The Windows GUI SHALL present a multi-page wizard hosted by a thin shell (`MainWindow` + `ShellViewModel`) composed of a header (the current page's title and subtitle), a `ContentControl` bound to the current page view-model, and a footer with `Back`, a primary action (`Next` or `Install`), and `Cancel`. Page views SHALL be selected by `DataTemplate` keyed on the page view-model's type (not `Frame`/`NavigationService`). The window SHALL open at 760×560, be resizable (`ResizeMode=CanResize`), and MUST NOT be maximized. The single-screen `MainWindowViewModel` dashboard SHALL be replaced by `ShellViewModel` plus per-page view-models. The shell SHALL open on the Language page (the wizard's first page); the `Back` action SHALL be disabled only on the Language page — the Hub is reachable Back to Language and therefore has `Back` enabled.

#### Scenario: Shell renders the current page and footer

- **WHEN** the wizard is showing a page whose view-model exposes a title, subtitle, and advance-eligibility
- **THEN** the shell displays that title and subtitle in the header, renders the page's view via its type's `DataTemplate` in the `ContentControl`, and enables the primary footer action only when the page reports it can advance

#### Scenario: Back is unavailable on the Language page (the wizard's first page)

- **WHEN** the current page is the Language page (the wizard entry)
- **THEN** the `Back` action is disabled

#### Scenario: Back is available on the Hub because Language precedes it

- **WHEN** the current page is the Hub
- **THEN** the `Back` action is enabled and returns to the Language page

## ADDED Requirements

### Requirement: Wizard opens on the language selection page

The wizard SHALL present a Language page (`WizardPageId.Language`) as its first page, before the Hub. The page SHALL show a fixed bilingual title ("Select your language / Elegí tu idioma") and an extensible single-select list of language choices (English, Español). The initial selection SHALL follow the resolved preselection (persisted preference > OS UI culture > English). `WizardFlow` SHALL advance Language → Hub and go Back Hub → Language. On advancing from the Language page, the `ShellViewModel` SHALL set its `_language`, persist it via `LanguagePreference.Save`, set the engine client's `Language`, and — when the language changed from the previously active one — invalidate the cached `_session`, `_uninstallOptions`, and `_starters` so subsequent flows re-fetch engine data in the new language. The `ShellViewModel` `CanGoBack` SHALL exclude the Language page (not the Hub).

#### Scenario: Language is the first page and advances to the Hub

- **WHEN** the wizard starts
- **THEN** the current page is the Language page with the preselected language chosen, and advancing goes to the Hub

#### Scenario: Advancing from Language sets and persists the language and configures the client

- **WHEN** the user selects a language on the Language page and advances
- **THEN** the shell's active language is set to that choice, `LanguagePreference.Save` is invoked with it, and the engine client's `Language` is set to the same value

#### Scenario: Changing the language invalidates cached engine data

- **WHEN** the user goes Back to the Language page, selects a different language than before, and advances
- **THEN** the previously cached `_session`, `_uninstallOptions`, and `_starters` are invalidated so the next flow re-fetches them in the new language

#### Scenario: Re-selecting the same language does not invalidate caches

- **WHEN** the user returns to the Language page and advances again with the same language selected
- **THEN** the cached engine data is retained (no needless re-fetch)

#### Scenario: Auto-install traverses the preselected Language page

- **WHEN** the bootstrapper runs an automated/auto-install loop
- **THEN** it traverses the preselected Language page without stalling, using the resolved (or `ACTIVE_STACK_UI_LANG`-overridden) language

### Requirement: Shell and pages render their literals from the language string table

The shell footer labels, every page title/subtitle, the XAML section headers, the backups and uninstall warnings, the complete-state label, and the progress step messages SHALL be rendered from `UiStrings` scoped to the shell's active language rather than hardcoded English. Specifically: page titles SHALL be produced via `base(UiStrings.Get(lang, …))`; `MainWindow.xaml` Back/Cancel SHALL bind to `ShellViewModel.BackLabel`/`CancelLabel`; the primary action label (Next / Install / Uninstall overrides) SHALL come from `UiStrings`; the `PageTemplates.xaml` section headers SHALL bind to view-model properties (so they are unit-testable) fed from `UiStrings`; `ProgressPageViewModel.ToFriendlyMessage` SHALL source its own strings from `UiStrings` while engine-provided `phase_started`/`install_finished` messages (already localized by the engine) are shown verbatim; and `CompletePageViewModel` SHALL expose a localized `StateLabel`. The backups Restore/Delete warnings and the uninstall modification warning SHALL come from `UiStrings`.

#### Scenario: Footer and page titles render in the active language

- **WHEN** the active language is Spanish
- **THEN** the footer primary/Back/Cancel labels and the current page's title/subtitle render their Spanish `UiStrings` values, and when the active language is English they render the English values

#### Scenario: Progress step messages render in the active language

- **WHEN** the active language is Spanish and the progress page maps an install step to a friendly message it owns
- **THEN** the message renders in Spanish (e.g. "Instalando OpenSpec."), while a `phase_started`/`install_finished` message supplied by the engine is shown verbatim as received

#### Scenario: Section headers and warnings are testable view-model properties

- **WHEN** the backups page surfaces its Restore/Delete confirmation warnings, or a page surfaces a `PageTemplates` section header
- **THEN** those texts come from view-model properties fed by `UiStrings` (assertable without rendering XAML) and reflect the active language
