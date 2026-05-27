// Package config implements the installer for harnesses of type "config".
//
// It composes the sdd-orchestrator block from per-agent base assets and
// toggle-controlled additive/subtractive fragments, then injects the block
// into each agent's instructions file (e.g. ~/.claude/CLAUDE.md) using
// idempotent marker-based injection with an atomic backup-first guarantee.
//
// Core API:
//   - Compose(toggles, variantKey) — pure function; returns assembled string
//   - Inject(path, composed, snapshotDir) — backs up file, then injects section
//   - Install(h, adapters, homeDir) — orchestrates compose + inject per agent
//
// Governance: HIGH — this package writes to user dotfiles. A backup is always
// taken before any write, and injection is idempotent (reinstalling is safe).
package config
