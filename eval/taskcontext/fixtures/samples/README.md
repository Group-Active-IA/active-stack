# Muestras de fuentes del cliente (tarea 3.6)

Cada muestra es la forma en que **una fuente del cliente** entrega la salida de un comando, para que el comparador de cobertura (`eval/taskcontext/clientsrc`) devuelva `completa`, `truncada`, `ausente` o `no comprobada` contra la referencia del arnés.

**Derivadas de la documentación y de lo medido en la ficha; ninguna es una captura real de esta tanda.** Los valores (ids, hilos) son inventados y llevan `FICTICIO`. La salida de referencia es el mini fixture de seis líneas que usa el test (`=== RUN   TestMini` … `FAIL	mini	0.100s`). Si la forma real de un evento difiere de la documentada, el adaptador la trata como ausente o no comprobada; no la inventa. La tarea 8.1 repite esto con los fixtures del arnés y un cliente real.

| Muestra | Caso | Basada en |
|---|---|---|
| `claude/posttooluse-complete.json` | hook `PostToolUse` de Bash: `tool_response.stdout`/`stderr`, `interrupted`, `isImage` | https://code.claude.com/docs/en/hooks (documentado, consultado por la ficha el 2026-11-13) |
| `claude/posttoolusefailure-truncated.json` | hook `PostToolUseFailure`: `error` empieza con `Exit code N` y se *middle-truncates* con la marca `... [N characters truncated] ...` | https://code.claude.com/docs/en/hooks; la marca coincide con la medida en C1–C5 (`entrega-03-04.md` §5) |
| `claude/posttoolusefailure-real-shape.json` | hook `PostToolUseFailure` con la forma **medida** el 2026-11-13: la cabeza se corta a mitad de línea, la marca va entre `\n\n` y la cola es un tramo **interior** de la salida, también cortado a mitad de línea, porque Bash conserva primero unos 30.000 caracteres y recién después recorta | corridas reales de Claude Code 2.1.323 en Linux con los fixtures `fail-middle` y `stderr-only` (tarea 8.1); los valores son del mini fixture |
| `codex/appserver-complete.jsonl` | App Server: `item/commandExecution/outputDelta` en orden y `item/completed` con `aggregatedOutput` y `exitCode` | https://learn.chatgpt.com/docs/app-server (documentado; ambos campos del item final son opcionales) |
| `codex/appserver-no-output.jsonl` | App Server: `item/completed` sin `aggregatedOutput` y sin deltas | mismo documento; caso del campo opcional ausente |
| `codex/exec-command-execution.jsonl` | `codex exec --json`: `item.completed` de tipo `command_execution` con `aggregated_output` completo y `exit_code` | **forma medida en X7** (`entrega-03-04.md` §5: Codex 0.167.1, Windows, 129.047 caracteres completos); https://learn.chatgpt.com/docs/non-interactive-mode para los nombres de campo |
| `codex/exec-reencoded.jsonl` | salida con colores ANSI y recodificada: no relacionable con la referencia | caso sintético para `no comprobada` |
