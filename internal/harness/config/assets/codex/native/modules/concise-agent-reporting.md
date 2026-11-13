## Concise Agent Reporting

- Line 1: the result, with its provenance (`path:line` or `path#symbol`).
- Line 2: the consequence or the next step.
- Several findings: a list, each with its provenance, before any explanation.
- Do not tell how you got to a result unless asked. When asked, produce the detail then, from the
  same provenance.
- Do not repeat what earlier steps of the same task already reported.
- If the content is already in a project artifact, name the artifact and the section instead of
  copying it.
- No finding: one line that says nothing was found and what you checked. Do not narrate the
  search.
- Keep the fixed lines of the other modules (`full read:`, `unnecessary call:`, `ref:`,
  `partial — unverified:`, `unavailable:`, `out-of-class use:`) exactly as written.
