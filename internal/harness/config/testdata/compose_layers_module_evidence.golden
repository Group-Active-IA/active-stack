## Evidence Contract

Load this module when the action runs commands whose results support a claim (`apply`, `verify`),
whether you work directly or delegate. What a worker or you declare stays **unverified** until the
checks below say otherwise.

### Verification inputs

Resolve the change with `openspec status --change NAME --json` and read its
artifacts. Run `active-stack evidence-state` in the same checkout to read
the current commit and fingerprint for comparison with a receipt. If the OpenSpec verify skill is installed, use it to check completeness,
correctness, and coherence. Otherwise apply those checks directly to the artifacts
and implementation, then review any execution claims below.

### Capture command

From the change's checkout, capture each relevant test or check with:

```text
active-stack evidence-capture -client CLIENT -interface SURFACE -version VERSION -change CHANGE -requirement ID -- COMMAND ARG...
```

Pass the actual client (`claude` or `codex`) and command arguments. The result
contains a receipt and the path to its redacted copy. Read the command exit
code from the receipt. Use this same capture for direct work and include the
instruction in delegated assignments. If capture fails or is unavailable,
report the affected requirement as unverified. A receipt alone is not a verdict.

### Execution receipt
One receipt per command run, written from the capture, not from memory:
- `run`: id assigned by the capture source
- `tree`: directory, base commit and, if there were uncommitted changes, the fingerprint of the full
  state (staged, unstaged and new files). HEAD plus "dirty" is not enough
- `command`: the complete command, as run
- `result`: exit code and timestamp
- `requirements`: ids of the requirements or scenarios it claims to cover
- `output`: reference to the **redacted** copy, one sha256 fingerprint of its exact bytes (algorithm
  named; LF and CRLF give different values), whether it was truncated, and the capture record

Never keep or send the unredacted output. A truncated output must have a reference to the full
redacted copy.

### Three checks, reported separately
- **Integrity**: the redacted copy exists and its recomputed fingerprint matches. It does not prove
  the command ran.
- **Provenance**: an independent source or a re-run by another party confirms the command, the exit
  code (the capture record's, not the worker's) and the exact tree state. A re-run counts only if it
  has its own capture from an independent source; your own re-run without that capture does not
  confirm anything. No independent source means unverified.
- **Sufficiency**: the command exercises what the requirement asks and the result supports it.
  Green tests with an unmet requirement are rejected. If verifying needs a detail the redaction
  discarded, the receipt stays unverified; do not store the original to avoid that.

### Statuses
- **verified**: all three checks pass for every declared requirement and at least one receipt exists.
  No receipts is never verified.
- **rejected**: invented or stale receipt (the copy does not exist or does not match, the source
  contradicts it, or the tree state differs), insufficient command, truncated output with no
  original, or an unmet requirement with green tests.
- **unverified**: everything else, including an incomplete receipt, no independent confirmation, an
  expired copy, or a check that could not be evaluated. Doubt goes here, never to verified.

### Hooks
A hook trace is a hint, not a record: its absence neither verifies nor rejects a receipt.

### Review step
Run this for every requirement an action claims to have covered, in this order:

1. **Completeness**: all receipt fields present. A missing field → unverified.
2. **Integrity**: recompute the fingerprint of the redacted copy.
3. **Provenance**: the capture record's exit code and the tree state match what the receipt says and
   the current tree; an independent source or a re-run confirms it. A re-run counts only with its
   own capture from an independent source; the coordinator's re-run without it does not corroborate.
4. **Sufficiency**: the command and the result support the requirement. When the action requires
   independent review under governance or the change requirements, record review by an agent or human
   who did not implement the change.
5. **Status** per requirement: rejected if any step shows a contradiction, else unverified if any
   step could not confirm, else verified.

Direct execution: you ran the commands, so run this step on your own receipts before reporting.
Delegated execution: run it on the `receipts` of the return envelope. A requirement without
receipts is unverified. Report the status of each requirement with the reason.
