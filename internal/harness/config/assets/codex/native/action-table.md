
## Action to Modules

Context module source: change 05, e4ec0f54536760f7fca6692e27bafcb8cd380df1,
entrega-03-04.md sections 2 and 3. Module bodies are copied verbatim; the table uses
Codex-relative paths and the active toggles. Model assignments are omitted until 09.

Before an action, read the modules listed for it. Read every mandatory module. Read an optional
module when its condition applies. When you delegate, paste the modules the specialist needs
into the assignment: the specialist does not see this skill. Pass its enabled tool list too.

| Action | Mandatory | Optional |
|--------|-----------|----------|
| explore | — | `modules/directed-code-reading.md`; `modules/bounded-command-output.md` if you run commands; `modules/task-tool-selection.md`; `modules/concise-agent-reporting.md` |
| propose | `modules/artifacts.md` | `modules/directed-code-reading.md`; `modules/task-tool-selection.md`; `modules/concise-agent-reporting.md` |
| apply | `modules/artifacts.md`; `modules/bounded-command-output.md`; %s | `modules/directed-code-reading.md`; `modules/task-tool-selection.md`; `modules/concise-agent-reporting.md` |
| verify | %s; `modules/bounded-command-output.md` | `modules/directed-code-reading.md`; `modules/task-tool-selection.md`; `modules/concise-agent-reporting.md` |
| archive | `modules/artifacts.md` | `modules/task-tool-selection.md`; `modules/concise-agent-reporting.md` |
| delegate (any action) | — | %s |
| session start and close | — | %s |

### Missing or empty modules

A module file that does not exist, or that exists with no text (zero bytes or only
whitespace), is missing.

- Missing mandatory module: stop only the action that needs it and write
  `missing module: <id> — <action> stopped`. Finish the parts of the request that do not need
  it and report them apart from the stopped action. Resume when the module is available or a
  surface that delivers it is used, and say which one.
- Missing optional module: write `missing optional module: <id> — continuing without it` and
  resolve the action directly. Without delegation, do not delegate.
- On a surface without subagents, verify stops: write
  `missing module: delegation — verify stopped (surface without subagents)`.
- Never replace a missing module with your own version of its rules.

### After compaction or resume

Compaction removes this skill and its modules from your context. Invoke the skill again and
re-read the modules of the current action before you continue it.
