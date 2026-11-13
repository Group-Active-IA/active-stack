## Directed Code Reading

Read only what the task needs, and say where every claim about code comes from.

- Read a file whole only when one of these five conditions holds:
  1. the file has 300 lines or fewer;
  2. the task will rewrite the whole file;
  3. the file is the artifact the active action produces or edits;
  4. a bounded read of that file came back empty;
  5. no range covers what the task needs. Then write in your report:
     `full read: <path> — <reason>`
  Any other full read is unjustified.
- If you do not know how long a file is, do not open it whole to find out: search for what you
  need, keep each match as `path:line`, and read that range.
- Otherwise read by path and line range, or by symbol. A symbol read returns the whole symbol,
  from its first line to its last, plus the signature of the declaration that contains it when it
  is nested. Nothing more: not the rest of the enclosing body, not the neighbouring symbols.
- To learn where a symbol is used, search and keep each match as `path:line`. Do not open every
  file that matches.
- To learn where a capability lives, run one bounded search over the repository. List each
  candidate path with its match type and count. A definition ranks above any number of
  references, and a reference above any number of plain-text matches; the count orders paths
  only within the same type. Do not open the files the search discards. Never walk the
  repository file by file to orient yourself.
- Every claim about code carries its provenance: `path:line` or `path#symbol`. A claim without
  provenance is unverified and does not count as evidence for accepting a change.
- If you read the same file and the same range again with no write to that file in between, do
  it if you need to, and write this line, even when the user asked for the repeat:
  `unnecessary call: <tool> <input>`
