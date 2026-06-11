---
description: "Apply an Active Stack starter to the current project (thin wrapper over active-stack starter add)."
---

Run the following bash command to apply the requested starter to the current project:

```bash
active-stack starter add $ARGUMENTS
```

This command delegates entirely to the `active-stack` binary on your PATH. It does not reimplement any starter resolution or install logic — it is a thin wrapper over the `active-stack starter add` CLI subcommand introduced in C-29.

**Arguments**: pass the starter id and any optional flags (e.g. `--project <root>`, `--dry-run`) directly. They are forwarded verbatim via `$ARGUMENTS`.

If the binary is not found, confirm that `active-stack` was installed via `active-stack install` and that your PATH is configured correctly.
