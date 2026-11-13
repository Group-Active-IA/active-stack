# Entrega de 07 a 03/04: texto de delegación

> Borrador para la ronda `apply`, sobre `main` en `d5e95bd` (03 y 04 integradas). Llena las filas 10 y 29 de
> `openspec/matriz-reglas-opsx.md`. El texto de los módulos va en inglés, como el resto de los
> assets. Nada de esto está instalado: sale de `design.md` (D1–D10) y cambia si el grupo cambia
> una decisión pendiente.
> Actualizado el 13/11 con 06 archivada (`3c7d908`): huella de la partida, módulo `evidence` y regla de reejecución con captura independiente.

## 1. Inventario de obligaciones de delegar (tarea 1.1, borrador)

Fuente: goldens de `main` (`compose_layers_*.golden` para Claude; `assets/codex/native/` de 04 y, para la composición legacy, `assets/codex/sdd-orchestrator.md` para Codex).
Falta confirmarlo componiendo **sin** el toggle `delegation`, porque hoy no hay golden de ese caso.

| Cliente | Capa | Ubicación | Texto | ¿Sobrevive sin `delegation`? | Destino según 07 |
|---|---|---|---|---|---|
| Claude | skill (frontmatter) | `compose_layers_skill.golden:3` | "Delegates real work to sub-agents" | sí | reescribir: "Decides per action whether to work directly or delegate" |
| Claude | skill | `:10` (`coordinator/role`) | "delegate real work to sub-agents via the Agent tool" | sí | reemplazar por la regla de §2 |
| Claude | skill | `:16` (`coordinator/principle-delegate`) | "Delegate, don't inflate" | sí, se agrega siempre | reemplazar por la regla de §2 |
| Claude | skill | `:21` (`coordinator/rules`) | "NEVER do apply or propose work inline — ALWAYS delegate" | sí | quitar; S4 cubre la independencia |
| Claude | skill | `:100`, `:109` (`coordinator/workflow`) | "For each action, delegate to a sub-agent" | sí | "For each action, apply the delegation rule" |
| Claude | módulo | `compose_layers_module_delegation.golden` | tabla 1-3 / 4+ | no (el módulo no existe sin toggle) | reemplazar por §3 |
| Codex | nativo, módulo | `native/delegation.md:3-7` | remite a la política de 07 y al sobre de D6, con un resumen propio | carga opcional (`action-table.md:19`) | reemplazar por §3 |
| Codex | nativo, tabla | `native/action-table.md:32-34` | sin delegación no delega; verify se detiene sin subagentes | sí | conservar: coincide con D1 paso 1 |
| Codex | legacy, bloque | `sdd-orchestrator.md:14` | "Delegate, don't inflate" | sí, fuera de markers | reemplazar por la regla de §2 |
| Codex | legacy, bloque | `:20` (dentro de `sdd-delegation`) | "You are a COORDINATOR — delegate real work" | no | reemplazar por la regla de §2 |
| Codex | legacy, bloque | `:85` | "For each action, delegate to the matching skill" | sí | "For each action, apply the delegation rule" |

## 2. Regla para el cuerpo de `SKILL.md` (D1, D8)

Reemplaza `coordinator/principle-delegate`. Va siempre, con o sin toggle, porque la decisión se toma
antes de cargar módulos.

```markdown
## Direct or delegated

Decide per action, in this order. The first step that applies closes the decision.

1. **Independent review required** (for example, verify after an apply): the review goes to a
   different agent than the one that built it. Nothing below overrides this. If delegation is not
   available, stop that verification, say what is missing, and continue with the rest.
2. **File conflict**: the write set overlaps an open assignment → serialize or narrow the scope.
3. **Work directly** when the task is short, depends on decisions made in this conversation, or you
   will read the full result anyway.
4. **Delegate** when the task produces more than 300 lines of output, when the answer is a
   conclusion and not the content read, or when two subtasks can run in parallel.
5. Otherwise, work directly.

Record each decision in the round report as `{action, step, signal, direct|delegated|stopped}`.
When you delegate, read `modules/delegation.md` first.
```

Sin el toggle `delegation`, la skill agrega en lugar de los pasos 2–5:

```markdown
Delegation is disabled in this installation: work directly. When an action requires independent
review, write `missing module: delegation — <action> stopped` and continue with the rest.
```

## 3. Texto de `modules/delegation.md` (D3–D7, D10)

```markdown
## Delegation

Load this module only when the delegation rule decided to delegate.

### Assignment
Every assignment includes:
- goal: one verifiable sentence
- action: explore | propose | apply | verify | archive
- start: base commit and working-state fingerprint you hand to the worker, computed as the
  tested-state fingerprint of `modules/evidence.md`
- write set: exclusive paths or globs the worker may create, modify, rename or delete (empty = read-only)
- read pointers: paths or symbols, never pasted content
- limits: what is not touched and what is not decided
- exit criteria: observable "done"
- return contract: the envelope below
- attempts: at most 3

Never split these across concurrent workers: `go.mod`, `go.sum`,
`internal/catalog/harnesses.yaml`, `assets/*/sdd-orchestrator.md`, `CLAUDE.md`, `AGENTS.md`.

### Ownership
A worker with a write set works in its own worktree created from the recorded start. When it
returns, compute what changed since the start: new commits, staged, unstaged, untracked, deleted and
renamed files (both source and destination count). Any path outside the write set → `out_of_scope`,
even if it is committed and tests pass. Changes that existed at the start are not the worker's.
Before integrating, check scope again on the combined result and re-run the scenarios that depend on
more than one assignment, with the capture of the harness: a re-run without its own independent
capture does not corroborate anything.

### Parallelism
At most two workers with a write set at the same time, each in its own worktree, and only when
their write sets are disjoint, neither depends on what the other writes, and each has its own exit
criteria and tests. Otherwise serialize. Read-only workers may share a tree.

### Depth
Only you launch workers. An assignment never asks a worker to delegate. Never delegate the
coordinator role.

### Return envelope
    status:    completed | blocked | out_of_scope | failed
    start:     base commit + declared fingerprint
    modified:  [path, kind: added | modified | deleted | renamed(from→to)]
    findings:  [{severity, path:line, description}]
    receipts:  [one execution receipt per command run, as defined in `modules/evidence.md`]
    pending:   short text | —

Validate in order: start matches the recorded one; `modified` matches the attributable changes and
stays inside the write set; receipts, with the review step of `modules/evidence.md`; exit criteria. What the worker declares stays **unverified**
until an independent source, or a re-run with its own independent capture, confirms it. "Tests green"
without that is still unverified.

### Attempts
Up to 3 attempts per assignment. Each retry gets the rejection reason and a new recorded start.
When the limit is reached, escalate to the user with the history.

### Evidence
When the assignment runs commands, paste the `## Evidence Contract` excerpt from
`modules/evidence.md` into it. Review the returned receipts with the same step you use when you
run the commands yourself.
```

## 4. Huecos que no son de 07

| Hueco | Dueño | Qué bloquea |
|---|---|---|
| Módulo `evidence` separado de `delegation` | 06 D5, pendiente de acuerdo con 03 y 04 (tarea 1.1 de 06) | si no se acuerda, el extracto vuelve a `delegation.md` y la ejecución directa sin el toggle queda sin contrato |
| Huella en Linux nativo o CI (medida sólo en Windows y WSL2) | 06 (deuda de 1.3a) | la prueba de `start` en Linux nativo |
| Quién evalúa la suficiencia | 06 (Open Questions 3) | qué evidencia esperada puede exigir el encargo |
| Captura independiente por interfaz | 05 (D20, D22) | que la reejecución de integración corrobore |
| Integración de 05 y 06 en `main` | PR #10 y PR #12 | la ronda `apply` de 07 |
| Umbral de 300 líneas | 05 D19 | el paso 4 de §2; supuesto a calibrar |
| Bloqueo preventivo de escritura fuera de alcance | 08 | hoy sólo hay comprobación posterior |
| Profundidad de delegación | grupo (D10) | la sección Depth de §3 |
