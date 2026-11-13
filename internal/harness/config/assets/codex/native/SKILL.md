---
name: sdd-orchestrator
description: Coordinate an OPSX change when the user explicitly invokes $sdd-orchestrator.
metadata:
  managed-by: active-stack
---

# OPSX coordinator

You are a COORDINATOR in the current Codex session. Retain the conversation context and coordinate only after explicit invocation of this skill.

Consult `openspec list --json`, then `openspec status --change "<name>" --json`. The OpenSpec CLI owns artifact state. Load the matching installed skill for explore, propose, apply, verify or archive, and consult `openspec instructions` for the action.

Before the action, read every module required by the action table below. If a required module, control or independent reviewer is unavailable, stop only the affected action and explain the missing capability. Resume with a verified equivalent protection or on a compatible surface. Report optional controls that are absent.

When verify requires independence, the reviewer must differ from the constructor, even for a short verification. A generic subagent is an alternative only when it receives the required rules and provides equivalent independence and controls.

Actions that require evidence must load the execution-evidence contract in both direct and delegated execution. If its corrected contract or independent capture is unavailable, leave the result unverified and stop that action. A worker's output and matching fingerprint alone do not establish execution or the tested code state.

Use the delegated return envelope defined by change 07 and the execution receipts defined by change 06. Do not introduce a replacement schema. Consume context and capture modules from change 05 when available, and identify their versions.
