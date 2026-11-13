## Task Tool Selection

Use the tools the active action needs and no optional integration it does not enable.

| Action | Base tools | Optional integrations enabled |
|---|---|---|
| explore | directed reading, search | Context7 |
| propose | directed reading, OpenSpec CLI | Context7 |
| apply | editing, tests | Context7; domain skills from the registry; Playwright only if the action touches UI |
| verify | tests, directed reading | Playwright only if there is UI |
| archive | OpenSpec CLI | none |

- An optional integration is off unless the active action enables it. Being installed does not
  enable it. Design integrations are enabled only when the `design-mockups` harness is active.
- When you delegate, pass the specialist the enabled list for its action in the assignment.
- If an enabled integration is unavailable, write
  `unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`
  Do not present the result as equivalent to one obtained with the integration. Do not call it
  again in the same task.
- If you call a tool or integration again with the same input and the state it observes has not
  changed, write `unnecessary call: <tool> <input>`
- If you use an integration the active action does not enable and the client did not block it,
  write `out-of-class use: <integration> in <action>`. Never describe a rule that is only
  recommended as a block.
