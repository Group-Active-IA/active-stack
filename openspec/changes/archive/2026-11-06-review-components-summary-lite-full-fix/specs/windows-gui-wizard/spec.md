## MODIFIED Requirements

### Requirement: Review confirms the selection and triggers install

The Review page SHALL present a read-only summary of the selection (agents,
install type, chosen components, tier) and the shell's primary action on Review
SHALL read "Install". Advancing from Review SHALL start the install by invoking
the engine client's install stream with the full selection (agents, mode, custom
ids, tier).

The components summary SHALL be mode-aware so it reflects what will actually be
installed regardless of install mode, and SHALL NOT be blank for any mode that
installs components:

- In `custom` mode the components summary SHALL be derived from the selection's
  custom component ids (which include the forced `permissions` id and the user's
  picks).
- In a non-custom mode (`lite` or `full`), where the Components page is skipped
  and the custom component ids are legitimately empty, the components summary
  SHALL instead be derived from the session's forced components together with the
  session's custom components marked as included in the mode's default
  (recommended) set — never from the empty custom ids.

#### Scenario: Review primary action starts the install with the full selection

- **WHEN** the user confirms on the Review page with agents, mode, custom ids, and tier set
- **THEN** the install stream is started carrying all of those values (not a single agent and not without tier/custom)

#### Scenario: Custom mode summarizes the chosen component ids

- **WHEN** the Review page is shown for `custom` mode with a non-empty set of custom component ids
- **THEN** the components summary lists exactly the labels of the session components whose ids are in that custom-id set (and no others)

#### Scenario: Custom components selected on Components appear in the Review summary

- **WHEN** the user toggles a set of components on the Components page (writing them into the shared `InstallSelection`) and reaches the Review page built over that same selection
- **THEN** the Review page's components summary lists exactly the selected components (including the forced `permissions` harness), matching what will be sent to the engine

#### Scenario: Full mode summary reflects the mode's implied set, not the empty custom ids

- **WHEN** the Review page is shown for `full` mode with empty custom component ids (the Components page was skipped)
- **THEN** the components summary is non-empty and lists the session's forced components plus the custom components marked recommended, and excludes custom components not marked recommended

#### Scenario: Lite mode summary is non-empty and mode-scoped

- **WHEN** the Review page is shown for `lite` mode with empty custom component ids (the Components page was skipped)
- **THEN** the components summary is non-empty and lists the session's forced components plus the custom components marked recommended (rather than rendering blank)
