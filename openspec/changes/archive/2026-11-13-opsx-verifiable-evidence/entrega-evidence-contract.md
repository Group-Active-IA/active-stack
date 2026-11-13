# Entrega de 06 a 03/04/07: texto del `## Evidence Contract`

> Borrador para la ronda `apply`, sobre `origin/main` `d5e95bd`. Llena la fila 28 de
> `openspec/matriz-reglas-opsx.md` y el contenido de `modules/evidence.md`, que
> `entrega-delegation.md` de 07 (`origin/propose/opsx-selective-delegation`, revisión consultada
> en `design.md`) ya referencia en sus secciones «Return envelope» y «Evidence». El texto
> de los módulos va en inglés, como los assets de 03. **Nada de esto está instalado ni edita
> archivos de 03, 04 ni 07.** La ubicación del módulo (`evidence` propio o dentro de `delegation`) es
> **propuesta de 06, pendiente de acuerdo con 03 (Juani) y 04 (Matías)** (design D5). Las decisiones
> de 06 que dependen de Gianfranco figuran como propuesta, pendiente de confirmación.
> Contrato completo y fixtures: `contrato-recibo-de-ejecucion.md`.

## (a) Extracto `## Evidence Contract`, listo para pegar

Estilo de los módulos de `main` (`compose_layers_module_*.golden`): encabezado `##`, frases cortas,
sin markers `active-stack:`.

````markdown
## Evidence Contract

Load this module when the action runs commands whose results support a claim (`apply`, `verify`),
whether you work directly or delegate. What a worker or you declare stays **unverified** until the
checks below say otherwise.

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
````

El cuerpo del módulo es (a) seguido de (b): la sección `### Review step` de (b) cierra el módulo.

## (b) Paso de revisión de recibos, único para ejecución directa y delegada

Texto para el cuerpo del módulo (en `evidence` si se acuerda, o en `delegation`), inglés:

````markdown
### Review step
Run this for every requirement an action claims to have covered, in this order:

1. **Completeness**: all receipt fields present. A missing field → unverified.
2. **Integrity**: recompute the fingerprint of the redacted copy.
3. **Provenance**: the capture record's exit code and the tree state match what the receipt says and
   the current tree; an independent source or a re-run confirms it. A re-run counts only with its
   own capture from an independent source; the coordinator's re-run without it does not corroborate.
4. **Sufficiency**: the command and the result support the requirement. When the action requires
   independent review (for example, verify after an apply), the reviewer is a different agent than
   the one that built it.
5. **Status** per requirement: rejected if any step shows a contradiction, else unverified if any
   step could not confirm, else verified.

Direct execution: you ran the commands, so run this step on your own receipts before reporting.
Delegated execution: run it on the `receipts` of the return envelope. A requirement without
receipts is unverified. Report the status of each requirement with the reason.
````

Esto responde a "mismo paso de revisión": 04 («Matriz común de reglas» y requisito «Evidencia directa y delegada»), 07 D6 (validación y veredicto).

## (c) Contenido del campo `recibos` (`receipts`) del sobre de 07

07 es dueña del sobre (`design.md` D6: `estado`, `partida`, `modificados`, `hallazgos`,
`recibos`, `pendiente`; en `entrega-delegation.md`: `status`, `start`, `modified`, `findings`,
`receipts`, `pending`). **06 define solo el contenido de `recibos`.**

- Una **lista de recibos de ejecución**, uno por comando ejecutado. Vacía si no se ejecutó ningún
  comando.
- Cada elemento tiene los seis campos de `contrato-recibo-de-ejecucion.md` §2 (`run`, `tree`,
  `command`, `result`, `requirements`, `output`), en español en el contrato y en inglés en el módulo.
- Lleva **referencias**, no salidas: la huella y el registro de captura apuntan a la copia redactada.
- Nada declarado por el worker en `recibos` cuenta como verificado hasta que el paso (b) lo revise.
  Un `completed` con `recibos` vacío no verifica ningún requisito.

**Hueco que 06 cierra para 07.** Su campo `start` ("base commit + declared fingerprint") necesita la
huella del estado de trabajo: la de 06 es **TreeHash sobre los bytes de la copia de trabajo**
(propuesta de 06, pendiente de confirmación; `informes/medicion-huella-2026-11-13.md`), la misma que
`arbol.huella` del recibo.

Hueco que **no** cierra 06: "qué evidencia esperada puede exigir 07 a un worker" (tarea 5.2) depende
de quién evalúa la suficiencia, pregunta abierta (design, Open Questions 3).

## (d) Qué cambia en la fila 28 si el grupo acepta el módulo `evidence`

Resumen de los ocho cambios de `design.md` D5 (los aplican sus dueños; **no se editó la matriz**):

| Dónde | Hoy (`origin/main`) | Con el módulo `evidence` |
|---|---|---|
| Fila 28, capa | módulo `delegation` y encargo | módulo `evidence` |
| Fila 28, especialista | sí, si ejecuta comandos | sí, extracto pegado en el encargo cuando ejecuta comandos |
| Fila 28, Claude / Codex | pendiente de texto de 06 | `modules/evidence.md` / módulo `evidence` |
| Tabla acción → módulos | `apply` → `artifacts`, `tdd`; no hay fila `verify` | `apply` → `artifacts`, `tdd`, `evidence`; `verify` → `evidence` (más `delegation` si exige revisor independiente); delegar → `delegation`, `model-assignments` |
| Regla "módulo solo con toggle activo" | sin excepciones | excepción: `evidence` es obligatorio sin toggle propio (su ausencia detiene la acción, D11 de 03) |
| 03 D4 / D12 / tarea 2.3 (archivados en `main`) | el extracto vive en `delegation.md` | vive en `evidence.md`; requiere un seguimiento de 03 sobre `main` |
| 04 `openspec/changes/archive/2026-11-13-codex-native-opsx/design.md:33` y `:43` (`main`) | ajuste coordinado pendiente | se cierra con la fila nueva; 04 está archivada: seguimiento de Matías sobre `main` |
| 04 `assets/codex/native/SKILL.md`, sección «Partial evidence requirements» (`main`) | requisitos parciales inline de D2–D4 de `f62cf00`, sin `modules/evidence.md` | se alinea o se reemplaza por el módulo cuando exista |

**Si el grupo no acepta el módulo**, el texto (a) y (b) se pega igual dentro de `delegation.md` como
prevé la fila 28 actual, y queda un límite declarado: con el toggle `delegation` apagado la
ejecución directa no tiene contrato y la acción se detiene (design D5, riesgo "Módulo `evidence` sin
acuerdo"). No hay cambio en 07 ni en 04 en ese caso, salvo el límite.

**Quién decide:** el grupo con Juani (03) y Matías (04); 07 (Neyen) y 05 (Emiliano) se informan.
Ningún archivo de 03, 04 ni 07 fue modificado por esta entrega.
