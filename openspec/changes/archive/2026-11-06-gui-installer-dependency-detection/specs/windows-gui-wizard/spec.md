## MODIFIED Requirements

### Requirement: Install flow ordering mirrors the TUI

The Install flow SHALL be driven by pure `WizardFlow.NextPage`/`WizardFlow.PreviousPage` functions computed from the current page, the shared install selection (agents, mode, custom component ids, tier), and the session's tier-capable agents — with no re-query to the engine on selection change. The order SHALL be: Hub → Dependencies → Assistants → InstallType → (Components only when mode is `custom`) → (Permissions only when at least one selected agent is tier-capable) → Review → Installing → Complete. Tier-capability SHALL be the intersection of the session's `TierCapableAgents` with the selected agents.

#### Scenario: Dependencies page appears right after the Hub

- **WHEN** the Install flow is entered from the Hub and `NextPage` is evaluated from Hub
- **THEN** the next page is Dependencies, and evaluating `NextPage` from Dependencies yields Assistants

#### Scenario: Full install with a tier-capable agent shows Permissions, not Components

- **WHEN** the selection is mode `full` with a tier-capable agent (e.g. `claude`) and `NextPage` is evaluated from InstallType
- **THEN** the next page is Permissions (Components is skipped because mode is not custom)

#### Scenario: Custom install shows Components

- **WHEN** the selection is mode `custom` and `NextPage` is evaluated from InstallType
- **THEN** the next page is Components

#### Scenario: A non-tier-capable-only selection skips Permissions

- **WHEN** the only selected agent is not tier-capable (e.g. gemini-only) and `NextPage` is evaluated from InstallType (mode `full`) or from Components (mode `custom`)
- **THEN** the next page is Review (Permissions is skipped)

#### Scenario: Back from Review skips pages not in the current path

- **WHEN** the selection is mode `full` with a gemini-only (non-tier-capable) agent and `PreviousPage` is evaluated from Review
- **THEN** the previous page is InstallType (both Permissions and Components are skipped because neither is in the path)
