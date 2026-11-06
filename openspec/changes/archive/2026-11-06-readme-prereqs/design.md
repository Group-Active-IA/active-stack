## Context

`README.md` currently has a "Quick start" section (`git clone` + `go build` + `./active-stack install`) with no prerequisites section before it. `internal/system/deps.go` already models the runtime dependency set (git, curl, node, npm required; go optional — no catalog harness uses a `go-install` method today, confirmed against `internal/catalog/harnesses.yaml`). This change only needs to reflect that reality in prose; it does not touch detection logic.

## Goals / Non-Goals

**Goals:**
- Give a first-time reader, before they run any command, the three official download links and which ones apply to their situation (building from source vs. running a released binary).
- Keep the build-time-only requirement (Go) visually and textually separate from the runtime requirements (Git, Node), so the section doesn't claim the released `.exe`/`ActiveStack-Setup.exe` needs Go.

**Non-Goals:**
- No change to `internal/system` detection logic — that's `tui-dependency-detection` and `gui-installer-dependency-detection`.
- No auto-install tooling or scripted dependency checks in this change — pure prose.

## Decisions

- **Placement**: new "## Requisitos previos" section immediately before "## Quick start", so it's the first thing a reader hits after the "¿Qué es esto?" pitch, not buried after the catalog tables.
- **Split by scope, not by tool**: two sub-lists ("Para compilar desde fuente" / "Para que el instalador funcione") rather than one flat list, because conflating them is exactly the confusion this change fixes (see proposal's Why).
- **Agent note stays qualitative**: "un agente soportado ya instalado" with a link to the existing "Agentes y plataformas" section anchor, rather than duplicating the agent table.

## Risks / Trade-offs

- [Risk] Wording drifts from `internal/system/defineDependencies()` if that list changes later (e.g. a future `go-install` harness makes Go a runtime requirement) → Mitigation: `tui-dependency-detection`'s design explicitly revisits this; this change's prose is scoped to "as of today" and doesn't claim to be machine-derived.
