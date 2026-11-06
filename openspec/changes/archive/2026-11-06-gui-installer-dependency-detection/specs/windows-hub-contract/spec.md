## ADDED Requirements

### Requirement: windows detect reports the full dependency set alongside detected agents
The `active-stack windows detect` command SHALL emit a JSON object carrying `detected_agents` (unchanged) and an additive `dependencies` field built from `system.DetectDependencies`, containing one entry per dependency with `name`, `required`, `installed`, `version` (when detected), `install_hint`, and an `install_command` string (the platform install command from `system.InstallCommandsForDep`, pre-joined into a single display string; empty when no automatic command exists for the current platform), plus top-level `all_present`, `missing_required`, and `missing_optional`. Adding this field MUST NOT change the existing `detected_agents` field's name, type, or values.

#### Scenario: Detect reports dependencies alongside agents
- **WHEN** `windows detect` runs on a machine missing Node.js
- **THEN** the response's `dependencies` array includes a `node` entry with `installed:false`, `required:true`, its `install_hint`, and an `install_command` string when one exists for the platform, AND `missing_required` includes `"node"`

#### Scenario: Existing detected_agents output is unaffected
- **WHEN** `windows detect` runs before and after this change on the same machine
- **THEN** the `detected_agents` field's value is identical in both runs
