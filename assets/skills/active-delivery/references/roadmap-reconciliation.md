# Roadmap reconciliation during delivery

The project is local by default. Read `.active-stack/roadmap.json` from the
project root. With no configuration or `mode: local`, use `CHANGES.md` and make
no remote lookup or write. If the configuration is unreadable or invalid,
report it and continue locally without consulting a provider.

## ClickUp mode

Use the stable ID from `CHANGES.md`. That file owns the scope and dependencies.
ClickUp owns the status and assignee of the linked task. Treat values in the
repo as observations, and surface discrepancies without rewriting scope,
dependencies or remote fields.

1. On resume, delivery and closure, ask `active-roadmap` to reconcile that ID
   in read-only mode against the configured ClickUp list. Resolve links through
   `.active-stack/roadmap-state.json`. Use the installed skill's contract for
   looking up tasks and updating that file. A missing or ambiguous link requires
   resolution in that workflow. Never infer a match from a similar title or
   create a task during reconciliation.
2. On a successful read, persist the linked task's last confirmed status,
   assignee, link and timestamp in the repo snapshot. Store it once per item in
   `.active-stack/roadmap-state.json`. Treat the timestamp as the time of the
   successful observation. If the read fails or the connector or link is
   unavailable, preserve the last confirmed observation and its timestamp.
   Record reconciliation as pending with the reason and continue local work
   within its agreed scope. Report status or assignment as unconfirmed if no
   successful observation exists.
3. A message such as “ready for delivery” or a local closure does not change
   ClickUp. If the user wants a remote change, tell `active-roadmap` to prepare
   a read-only preview for the exact destinations and field changes. Wait for
   explicit confirmation of that preview before any remote write, including a
   status change, assignment or closure. Keep read reconciliation separate from
   publication. Report local results and the confirmed remote observation or
   pending reconciliation in the delivery summary.
