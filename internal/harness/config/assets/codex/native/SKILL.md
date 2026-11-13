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

Use the delegated return envelope defined by change 07 and the execution receipts defined by change 06. Do not introduce a replacement schema. Load the context modules from the pinned change 05 source identified in the action table.

## Partial evidence requirements

For actions requiring evidence, the following constraints come from change 06,
f62cf00c09dc8a3aa96ddb4648801250e1a67a8c, design decisions D2–D4. They are
not a complete receipt contract: the literal evidence module and assignment extract,
the tested-tree mechanism and the remaining capture agreements are still pending.

- Apply the same review in direct and delegated execution. An installation record is
  ownership information, not proof that a task command ran.
- A receipt describes one execution of one command and ties it to the run, tested tree,
  command, result, requirement IDs and redacted-output reference defined by 06. Do not
  invent a serialization schema for these fields.
- Record the exact tested state, including staged, unstaged and new untracked files.
  The same HEAD with different diffs represents different tested states; HEAD plus a
  dirty flag is insufficient. Do not invent the pending tree-fingerprint mechanism.
- Integrity, provenance and sufficiency are separate checks. A matching hash alone
  proves neither execution nor the tested state. An independent capture source or a
  re-run by another party must corroborate the execution, exit code and exact code state.
  Only mark a result verified when all three checks pass for every declared requirement
  and at least one execution receipt exists for it.
- Assess whether each command actually exercises its declared requirement. Passing
  tests do not establish an uncovered requirement. An unknown or unassessable result,
  a coherent but uncorroborated receipt, or an empty receipt set stays unverified.
  Independently contradicted execution or an unmet requirement is rejected.
  A corroborated receipt for a different tested state is stale and rejected for the
  current tree. A missing copy or an undecided expiry case cannot become verified.
- Fingerprint the exact redacted-file bytes with named SHA-256, without line-ending
  conversion: LF and CRLF are different bytes and produce different fingerprints.
  If redaction discarded a required detail, the result stays unverified; do not persist
  a secret-bearing original to compensate.
- Without the complete required contract, sufficient independent capture or an
  obligatory reviewer, stop only the action requiring evidence. Keep unrelated reading
  and actions available. Do not present these partial constraints as verified execution,
  a runtime validator, or a delivered `evidence` module.
- Location, access, retention, disputed receipt states and the final return envelope
  remain with 05/06/07. Report those limits instead of filling them with local defaults.
