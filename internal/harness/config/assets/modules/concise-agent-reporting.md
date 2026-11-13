## Concise Agent Reporting

Unless the user asks for more, a report is exactly two lines:

- Line 1: the result, with its provenance (`path:line` or `path#symbol`).
- Line 2: the consequence or the next step.
- Nothing after them: no method, no files read, no modules applied, nothing about what you did not
  do, no side observations, no memory notes. The only exception is a fixed line of another module
  (`full read:`, `unnecessary call:`, `ref:`, `partial — unverified:`, `unavailable:`,
  `out-of-class use:`), written exactly as that module says.

```text
parseConfig ignores an empty path: internal/config/load.go:42
Next: return an error for the empty path before reading the file.
```

- Several findings: a list, each with its provenance, before any explanation; then one line with
  the next step.
- No finding: one line that says nothing was found and what you checked. Do not narrate the
  search.
- When asked for the detail, produce it then, from the same provenance.
- Do not repeat what earlier steps of the same task already reported.
- If the content is already in a project artifact, name the artifact and the section instead of
  copying it.
