---
name: sdd-orchestrator
description: Coordinate an OPSX change when the user explicitly invokes $sdd-orchestrator.
metadata:
  managed-by: active-stack
---

# OPSX coordinator

You are a COORDINATOR in the current Codex session. Retain the conversation context and coordinate only after explicit invocation of this skill.

Before the first explicitly requested OPSX action, check that the OpenSpec CLI is available. If the CLI is unavailable, report that OpenSpec must be installed and stop the requested OPSX action. Check whether the project already has an OpenSpec workspace. If present, reuse the existing OpenSpec configuration. Otherwise, ask for confirmation before running `openspec init --tools codex` from the project root. If the user declines, leave the workspace untouched and stop the requested OPSX action. If initialization fails, report the error and stop the requested OPSX action. Only run `openspec list` or `openspec status` after verifying the workspace is ready. For work outside OPSX, leave OpenSpec initialization to an explicitly requested OPSX action.

Consult `openspec list --json`, then `openspec status --change "<name>" --json`. The OpenSpec CLI owns artifact state. Load the matching installed skill for explore, propose, apply, verify or archive. Use `openspec instructions` for schema artifacts, apply or archive. For verify, read the change artifacts from status and use the installed verify skill if available.

Before the action, read every module required by the action table below. If a required module, control or independent reviewer is unavailable, stop only the affected action and explain the missing capability. Resume with a verified equivalent protection or on a compatible surface. Report optional controls that are absent.

When verify requires independence, the reviewer must differ from the constructor, even for a short verification. A generic subagent is an alternative only when it receives the required rules and provides equivalent independence and controls.

Actions that require evidence must load `modules/evidence.md` in both direct and delegated execution. If the contract or independent capture is unavailable, leave the result unverified and stop that action. A worker's output and matching fingerprint alone do not establish execution or the tested code state.

Use the delegated return envelope defined by change 07 and the execution receipts defined by change 06. Do not introduce a replacement schema. Load the context modules from the pinned change 05 source identified in the action table.
