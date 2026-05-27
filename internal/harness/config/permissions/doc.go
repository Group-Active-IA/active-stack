// Package permissions implements the permissions harness for JR Stack.
// It injects safe-by-default permission overlays into each supported agent's
// settings.json file, with mandatory backup before any write.
//
// Governance: ALTO — writes user-level agent configuration.
// Backup is MANDATORY and cannot be disabled. If backup fails, the settings
// file is NOT touched.
//
// Supported agents and their overlay behavior:
//   - claude:      acceptEdits mode + deny-list (.env, rm -rf /, etc.)
//   - opencode:    bash/read granularity; git destructive = ask
//   - gemini:      auto_edit mode
//   - vscode:      chat.tools.autoApprove = true
//   - cursor:      no-op (permissions via cli-config.json)
//   - codex:       no-op (no known settings.json)
//   - antigravity: no-op (permissions via IDE UI / Artifact Review Policy)
package permissions
