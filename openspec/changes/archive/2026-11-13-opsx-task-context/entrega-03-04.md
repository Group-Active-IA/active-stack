# Entrega de la 05 a 03 y 04

> Ficha 05 · `opsx-task-context` · 2026-11-13.
> **Estado: texto listo para revisar con 03, 04 y 07; la integración en código queda para la
> ronda de apply.** Ningún módulo de la 05 corrió todavía en un cliente real (§6).
> Este archivo es la única casa de lo que la 05 entrega: los textos de módulo, la tabla
> acción → módulos con su regla de módulo ausente, y las pruebas reales que los respaldan.
> `design.md` dice **por qué** (D-números); acá está **qué** se copia y **cómo** se reproduce.

Etiquetas de cada afirmación: **medido** (corrida real, con el id de la §5), **documentado**
(fuente citada), **no probado** (motivo en la §6).

## 1. Qué entrega la 05 y cómo se integra

| Entrega | Sección | Responde a |
|---|---|---|
| Texto de cuatro módulos, en inglés | §2 | fila 27 de la matriz común; tarea 2.3 de la 03; `assets/codex/native/SKILL.md:20` y tarea 1.4 de la 04 |
| Tabla acción → módulos de los nueve módulos, regla de módulo ausente o vacío, compactación y rutas por cliente | §3 | tarea 2.3 de la 03 (*"recibir los textos de D17"*); la tabla que hoy escriben a mano `actionTable` (03) y el string `actions` (04) |
| Contrato de captura, por referencia | §4 | las líneas de evidencia de `SKILL.md:18-20` de la 04 |
| Pruebas reales | §5 y §6 | lo que 03 y 04 necesitan saber de los clientes para integrar |
| Qué falta en código, por ficha | §7 | — |

**Por qué en inglés.** Los módulos y el texto de la tabla se insertan en assets que están en
inglés (`sdd-orchestrator.md`, `SKILL.md` de Claude en `main` y de Codex en el PR #9). El resto de este archivo está en
castellano. Si el grupo decide ids en inglés para todos los artefactos, se aplica aparte
(`design.md` — *Decisiones para el grupo*).

**Ids de módulo.** Los cuatro módulos usan el nombre de su capacidad, como los de la matriz
(`artifacts`, `delegation`), para que el id, el archivo y la especificación coincidan:

| Id del módulo | Nombre anterior en la 05 | Capacidad (`specs/`) |
|---|---|---|
| `directed-code-reading` | `lectura-dirigida` | `directed-code-reading` |
| `bounded-command-output` | `resumen-acotado` | `bounded-command-output` |
| `task-tool-selection` | `herramientas` | `task-tool-selection` |
| `concise-agent-reporting` | `reporte` | `concise-agent-reporting` |

**Quién decide qué.** El informe unificado reparte, en su tabla *"Reparto para validar con el
equipo"* (contrato *Activación, núcleo y módulos*): *"03/04 mantienen juntos la matriz por
cliente, interfaz, rol y acción. 05 define cuándo cargar cada módulo. 04 integra `verify` en
Codex."* `CHANGES.md` (ficha 05, *Qué se implementaría*) pide a la 05 *"Mapeá con 03, 04 y 07
cuándo se cargan TDD, delegación y coordinación de subagentes"*. Por eso la tabla de la §3 es una
**propuesta para acordar con 03, 04 y 07**, no una obligación para sus ramas: 03 y 04 pueden
contraproponerla y su forma actual (`actionTable` en la 03, string `actions` en la 04) sigue
vigente hasta que se acuerde otra.

**Cómo se propone integrarlo.** Hay dos compositores: 03 `ComposeLayers`, en `main` desde el
merge del PR #8 (`7a60089`), y 04 `ComposeCodexPackage`, en el PR #9 (§5, R5–R6). Al integrar
`main` (`91fbb70`), la 04 mantiene los dos: Codex no usa el compositor de capas de la 03, por
decisión documentada de la 04 (re-medido: R15 y R16). La 05 propone uno común con un registro de
módulos (`design.md` — D23); no es requisito para integrar los módulos, y la 04 ya escribió lo
mismo. Mientras no exista, y si 03 y 04 aceptan la propuesta:

- **En `main`, para Claude (con la 03):** cada módulo de la §2 necesitaría entrar como archivo
  `modules/<id>.md` **sin depender de un toggle**, porque una anotación
  `<!-- as:module/<id> -->` sola se descarta sin error en `ComposeLayers` (R2, re-medido en
  `main`, R17). La tabla de la §3 suma filas de `explore` y `verify`, que hoy no tiene la
  `actionTable` de `main` ni la matriz común (R1): es un faltante compartido, no un defecto de
  la 03.
- **En la 04, para Codex:** la 04 ya publicó su plan (`91fbb70:openspec/changes/codex-native-opsx/informes/analisis-dependencias-05-06-2026-11-13.md`):
  los cuatro textos en `assets/codex/native/modules/<id>.md`, cargados en `CodexPackage.Files`
  sin toggle, la tabla de la §3 adaptada en lugar del string `actions`, `model-assignments`
  fuera de Codex hasta la 09, y un archivo de 0 B o sólo con espacios tratado como faltante,
  cada cosa con sus tests. La 05 lo toma como base del acuerdo. La 04 ya cubre `verify` en la
  línea de `delegation`; la tabla de la §3 le agrega una fila propia.
- **En las dos:** que el verify de la instalación falle si un módulo de la tabla falta **o está
  vacío**. El de la 04 ya lo hace (byte a byte, R4, re-medido en `91fbb70`, R18); el de `main`
  deja pasar un módulo vacío (R3, re-medido en `main`, R17).

Los textos se copian tal cual: sin frontmatter (para que el cliente no los liste como skills),
empiezan con `## <Título>` y no usan los encabezados de agrupación `## Core Principles` ni
`## Rules` (matriz común, *Reglas de la tabla*). Las líneas de forma fija las cuenta el arnés:
no se traducen ni se reescriben. Las cifras 300 líneas y 10.000 caracteres salen del catálogo
(tareas 3.2 y 4.1); cuando la tabla se genere desde ahí (tarea 6.2), el texto se regenera.

## 2. Module texts

### `modules/directed-code-reading.md`

````markdown
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
````

### `modules/bounded-command-output.md`

````markdown
## Bounded Command Output

Clients cut long command output before you see it. Claude Code keeps about 10,000 characters of
a failed command, head and tail, and that tail is not the real end of the output. Codex keeps
head and tail and does not tell you the middle is gone. Never summarize from that view.

### When to summarize

- Output above 300 lines or above 10,000 characters, whichever comes first, is delivered as a
  summary plus a reference. Shorter output is delivered whole.
- The threshold decides when to summarize, never how much to cut. Never drop a failure line to
  save space.

### Capture before you read

- Capture every verification command (tests, build, vet, lint) and every command known to be
  verbose, before you run it. Do not guess its size first.
- If this skill names an independent capture source for your client and surface, use it and name
  it. Otherwise capture the output yourself with your shell's recipe; the capture source is then
  `agent`.
- If an independent capture already wrote a redacted copy and its capture record, summarize from
  that copy and cite it in `ref:`. Do not write a copy of your own.
- bash / sh:
  `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
- PowerShell 5.1 and 7 (keeps the command's bytes; avoids UTF-16 and `NativeCommandError`):
  `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
- Put `$f` under `<user cache dir>/active-stack/runs/<change>/<run-id>/`, or under
  `_sin-change` when no change is active. Never inside the repository. The copy expires seven
  days after the change is archived, or seven days after capture under `_sin-change`.
- Redact before you keep it. Replace every value that matches the sensitive list (`.env`
  assignments, tokens and API keys, cloud credentials, SSH keys and `*.pem`, environment variable
  values, home directory paths, account ids, emails and account UUIDs) with an omission mark,
  keeping the line in its place. Write the redacted copy to a second file, check that no listed
  value is left in it, and delete the raw file before you report. The copy you keep and cite is
  the redacted one.
- Never write a raw sensitive value: not in a file you keep, not in the summary, not in your
  reply, even when asked for the original. Quote the redacted line and say the value was omitted.
- Read the captured copy with directed reads (failure patterns, exit code). Do not summarize
  from what the client showed you.
- A command that already ran without capture and turned out verbose: if it has no side effects,
  run it again with capture; if it has side effects, do not repeat it and mark the summary
  partial.

### What the summary keeps

- Every error, every failed test and the exit code. A non-zero exit code is always a failure.
  One `failure:` line per captured line that matches a pattern, including the final
  `FAIL\t<package>` line.
- Failure patterns: `go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` and `go vet`:
  `<file>:<line>:<column>:`. For a tool with no declared patterns, write
  `failures identified by exit code only`.
- Only failures the output records. Lines that only look like errors are not failures. With no
  failure, write exactly `failure: none`, with nothing after it.
- Each `failure:` line is the captured line verbatim, leading spaces included. Append nothing:
  no line number, no note.
- Repetition without failures is dropped.

### Fixed lines

The harness counts these lines. Write them exactly, one per line, at the top of the summary file
and in your reply, whatever the shell. Prose may follow them; it never replaces them.

```text
exit: <code>
failure: <line as captured>
ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
status: unverified — <what is missing>
partial — unverified: <reason>
```

Example, a failing test run captured by the agent:

```text
exit: 1
failure: --- FAIL: TestParse (0.00s)
ref: cmd=go test ./pkg/ run=20261113-1 path=<cache>/active-stack/runs/_sin-change/20261113-1/out.redacted.txt sha256:<hex> expires=2026-11-20
capture: agent · redactor: agent · producer: sdd-orchestrator
status: unverified — no independent capture
```

- No captured original: write `partial — unverified: <reason>` and no `ref:` line (no path, no
  hash, not even `ref: none`). Say which case it is: `never captured`, `copy missing` (the
  record names a file that is not there), `expired on <date>`, or `no persistent disk` (surface
  without the user's disk).
- A result that you or a worker declare stays `unverified` until another party corroborates
  the execution and the code state with independent capture or a re-run. A matching hash proves
  the copy is intact; it does not prove who ran the command.
- When you verify a declared result, give one state. `rejected`: the summary omits a failure the
  copy records, declares one it does not record, or its hash or exit code differs from the copy
  or the capture record. `verified`: none of that, and the copy comes from an independent capture
  (record source `arnés` or a client source, with coverage `demostrada`) or another party re-ran
  the command with capture. Otherwise `unverified`.
- If verifying needs a detail that redaction removed, the result stays `unverified`; say which
  detail is missing.
- Never cite the client's own copy (session transcript, client output file) as the reference:
  it is not redacted and Active Stack does not manage it. When asked for an original, or when
  that copy holds a sensitive value, name it as a limit.
````

### `modules/task-tool-selection.md`

````markdown
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
- An enabled integration is unavailable when its tools are not listed in this session or it does
  not respond. When the task needs it and it is unavailable, write
  `unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`
  Do not present the result as equivalent to one obtained with the integration. Do not call it
  again in the same task.
- If you call a tool or integration again with the same input and the state it observes has not
  changed, write this line, even when the user asked for the repeat:
  `unnecessary call: <tool> <input>`
- If you use an integration the active action does not enable and the client did not block it,
  write `out-of-class use: <integration> in <action>`. Never describe a rule that is only
  recommended as a block.
````

### `modules/concise-agent-reporting.md`

````markdown
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
````

## 3. Tabla acción → módulos

La propuesta y su justificación están en `design.md` — D19; esta es la única copia de la tabla
en la 05. Es una propuesta para acordar con 03, 04 y 07 (§1, *Quién decide qué*).
Columnas por acción: **O** obligatorio, **op** opcional, **—** no se carga.

| Módulo | `explore` | `propose` | `apply` | `verify` | `archive` | Especialista |
|---|---|---|---|---|---|---|
| `artifacts` | — | O | O | — | O | sí, en el encargo |
| `tdd` | — | — | O si Strict TDD está activo y hay runner | — | — | sí, en `apply` |
| `delegation` | op (regla de la 07) | op | op | **O** (S4 de la 07) | op | sí, cuando se delega |
| `model-assignments` | con `delegation` | con `delegation` | con `delegation` | con `delegation` | con `delegation` | no |
| `engram-opsx` | op, al inicio y al cierre | ídem | ídem | ídem | ídem | no |
| `directed-code-reading` | op | op | op | op | — | sí, si lee código |
| `bounded-command-output` | op, si corre comandos | — | **O** | **O** | — | sí, si corre comandos |
| `task-tool-selection` | op | op | op | op | op | no: el coordinador le pasa la lista |
| `concise-agent-reporting` | op | op | op | op | op | sí |

Condiciones:

- **Toggles.** Un módulo de toggle inactivo no figura (matriz común). Si el toggle de un módulo
  **obligatorio** está inactivo, la fila dice que la acción se detiene: `verify` sin `delegation`
  (como ya hace la 04) y `apply` con Strict TDD exigido y sin `tdd`. Los cuatro módulos de la 05
  no tienen toggle: figuran siempre.
- **Codex.** `model-assignments` se omite hasta la 09: límite declarado, no ausencia.
- **Compatibilidad con la 03.** Las filas `delegate` y `session start and close` de su
  `actionTable` quedan como filas del texto generado.
- **Modo legado** (Claude < 2.1.110, `design.md` — D19): no hay skill ni tabla; ningún módulo
  de la 05 se entrega.

### Texto para el `SKILL.md`

Claude. Para Codex, reemplazar `${CLAUDE_SKILL_DIR}/modules/` por `modules/` y quitar
`model-assignments`.

````markdown
## Action to Modules

Before the first tool call of an action, read the modules listed for it. Read every mandatory
module. Read an optional module when its condition applies;
an optional module with no condition always applies. When you delegate, paste the modules the
specialist needs into the assignment: the specialist does not see this skill.

| Action | Mandatory | Optional |
|--------|-----------|----------|
| explore | — | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md` if you run commands; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
| propose | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md`; `${CLAUDE_SKILL_DIR}/modules/tdd.md` when Strict TDD is on and a test runner exists | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
| verify | `${CLAUDE_SKILL_DIR}/modules/delegation.md`; `${CLAUDE_SKILL_DIR}/modules/model-assignments.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md` | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
| archive | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` | `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
| delegate (any action) | — | `${CLAUDE_SKILL_DIR}/modules/delegation.md`; `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |
| session start and close | — | `${CLAUDE_SKILL_DIR}/modules/engram-opsx.md` |

### Missing or empty modules

A module file that does not exist, or that exists with no text (zero bytes or only
whitespace), is missing.

- Missing mandatory module: stop only the action that needs it and write
  `missing module: <id> — <action> stopped`. Finish the parts of the request that do not need
  it and report them apart from the stopped action. Resume when the module is available or a
  surface that delivers it is used, and write `resumed: <action> — <module | surface> available`.
- Missing optional module: write `missing optional module: <id> — continuing without it` and
  resolve the action directly. Without `delegation`, do not delegate, even where other rules of
  this skill say to delegate.
- On a surface without subagents, `verify` stops: write
  `missing module: delegation — verify stopped (surface without subagents)`.
- Never replace a missing module with your own version of its rules.

### After compaction or resume

Compaction removes this skill and its modules from your context. Invoke the skill again and
re-read the modules of the current action before you continue it.
````

### Rutas por cliente

| Cliente | Forma de la ruta | Estado |
|---|---|---|
| Claude Code | `${CLAUDE_SKILL_DIR}/modules/<id>.md` | medido en Windows: se resuelve con barras `/` (C8); terminal Linux según la tarea 1.8 de la 03; IDE no probado |
| Codex | `modules/<id>.md`, relativa al directorio de la skill (forma actual de la 04) | medido en Windows el 13/11 (X12): con `codex exec --approve-for-me`, `$sdd-orchestrator` lee el módulo con `Get-Content` |

**Por qué la regla de módulo vacío.** La 03 ya la escribió en su `SKILL.md` (commit `9aad12e`:
*"If a module listed in the table is missing or empty, do not run that action…"*) y la 04 tiene
una regla general (*"stop only the affected action and explain the missing capability"*). Con la
regla de la 03 instalada, haiku leyó el módulo vacío y declaró propose bloqueado en 2 de 2
corridas (C15). Las tres corridas anteriores, sin esa regla, no son concluyentes (C13). El texto
de la §3 coincide con el de la 03 en el fondo y agrega dos cosas a discutir: la definición de
vacío (0 bytes o solo espacios) y una línea fija para el reporte. Lo que falta en la 03 es que su
verify lo detecte (R3).

**Por qué la regla de compactación.** Después de `/compact` solo sobrevive el aviso universal de
reinvocar (C11); el cuerpo y los módulos ya no están.

## 4. Contrato de captura

Por referencia; la única casa de cada parte es `design.md`.

| Parte | Dónde | Qué tienen que hacer 03 y 04 |
|---|---|---|
| Tres roles: captura, redacción y resumen | D8 | nada en el compositor: el módulo `bounded-command-output` los prescribe |
| Cuándo resumir y por qué esas cifras | D3 | nada; las cifras viajan en el módulo |
| Qué fuente captura en cada cliente e interfaz | D22 | Claude: no hay fuente independiente en la sesión para un comando fallido, así que la captura es del agente o del arnés. Codex CLI: el arnés que conduce `codex exec --json` usa `aggregated_output` (X7). No hay que cablear nada en el `SKILL.md` |
| Formato del resumen y líneas fijas | §2, `bounded-command-output` | copiar el módulo |
| Referencia de cinco campos, registro de captura, ubicación y retención | D9 | nada; la 06 confirma o contrapropone |
| Lista sensible | D14 | nada; vive en el catálogo |
| Estados *verificado*, *sin verificar*, *rechazado* | D20 (de la 06) | las líneas de evidencia de su `SKILL.md` (04, `:18-20`) ya los respetan |

## 5. Pruebas reales

Windows 11; Claude Code **2.1.319** (`claude -p --output-format stream-json`, modelo haiku);
Codex CLI **0.167.1** (`codex exec --json`); Go 1.26 en Windows y en WSL. **n = 1 por prueba.**
La política de shell de Codex en Windows bloquea los comandos sin bypass: el bypass se usó solo
en un directorio temporal aislado. *(13/11 por la tarde: `codex exec --approve-for-me` corre comandos sin bypass total; con eso se midieron X12 y 8.1.)* No se versionan transcripciones, sesiones ni logs: cada fila
se reproduce con el comando que figura.

### Generadores

Claude (bash). `gen.sh <modo>`: 3.000 líneas con el fallo en la 1.500.

```bash
mode=$1
out() { for i in $(seq 1 3000); do if [ $i = 1500 ]; then echo "FAIL: needle_middle_5521 at line $i"; else echo "line $i padding padding padding"; fi; done; }
case $mode in
 fail) out; exit 1;;
 ok) out; exit 0;;
 errfail) out >&2; exit 1;;
esac
```

`gen2.sh failend`: 2.999 líneas y `FAIL: end_needle_7788` en la última, exit 1.
`gen2.sh fail25`: 750 líneas con el fallo en la 375, exit 1.

Codex (PowerShell). `gen_fail.ps1`: 3.000 líneas `line NNNN padding …`, `FAIL: needle_middle_5521`
en la 1.500 y `FAIL: end_needle_9090` en la última, por `[Console]::Out`, `exit 1`.
`gen_ok.ps1`: igual con `exit 0`. `gen_err.ps1`: igual por `[Console]::Error`. `gen_750.ps1`:
igual con 750 líneas (32.297 caracteres).

```powershell
$n=3000
for ($i=1; $i -le $n; $i++) {
  $t = "line {0:D4} padding padding padding padding" -f $i
  if ($i -eq [int]($n/2)) { $t = "FAIL: needle_middle_5521 " + $t }
  if ($i -eq $n) { $t = "FAIL: end_needle_9090 " + $t }
  [Console]::Out.WriteLine($t)
}
exit 1
```

Cómo se lee el resultado: en Claude, el `tool_result` del evento `user` del stream-json (largo,
marca `[N characters truncated]`, primera y última `line N` visible). En Codex, el
`aggregated_output` y el `exit_code` del `item.completed` de tipo `command_execution`, y lo que
recibió el modelo en el historial de la sesión.

### Claude Code

| Id | Afirmación | Comando | Resultado | Veredicto |
|---|---|---|---|---|
| C1 | Un comando fallido con salida larga llega recortado | `bash gen.sh fail` | 10.040 caracteres con `is_error`, marca `[20012 characters truncated]`; visibles las líneas 1–154 y 762–912; el fallo de la 1.500 no aparece; sin ruta al resto | medido: se pierde el fallo central |
| C2 | La cola es el final real | mismo C1 | la última línea visible es la 912 de 3.000: Bash conserva primero ~30.000 caracteres y después recorta; la marca dice 20.012 y faltan ~90.000 | medido: no lo es, y la marca subestima la pérdida |
| C3 | Un fallo en la última línea sobrevive | `bash gen2.sh failend` | igual a C1; el `FAIL` de la línea 3.000 no aparece | medido: no sobrevive |
| C4 | Solo stderr se comporta distinto | `bash gen.sh errfail` | idéntico a C1 | medido: igual |
| C5 | Una salida media también pierde el medio | `bash gen2.sh fail25` (~25.000 caracteres) | 10.040 caracteres, marca `[14657 characters truncated]`; se pierden las líneas ~155–599, con el fallo de la 375 | medido |
| C6 | Con exit 0 la salida grande es recuperable | `bash gen.sh ok` | vista previa de ~2 KB (líneas 1–62) y ruta a `tool-results/<id>.txt`, que tiene las 3.000 líneas (100.897 B) | medido: hay original solo en éxito |
| C7 | `disable-model-invocation: true` oculta la skill hasta invocarla | listar skills en una sesión normal | `sdd-orchestrator` no figura para el modelo; sí en `slash_commands` | medido |
| C8 | `/sdd-orchestrator` carga el cuerpo y `${CLAUDE_SKILL_DIR}` se resuelve | invocar y leer `modules/artifacts.md` | lee `<proyecto>/.claude/skills/sdd-orchestrator/modules/artifacts.md`, con `/` | medido |
| C9 | La carga por acción funciona | `/sdd-orchestrator` con explore, propose, apply y verify | en cada acción lee los módulos de su fila y los pega en el encargo | medido |
| C10 | Qué ve cada subagente | lanzar Explore, Plan y general-purpose con testigos en `CLAUDE.md` y `SKILL.md` | general-purpose ve `CLAUDE.md` y no `SKILL.md`; Explore y Plan no ven `CLAUDE.md` | medido |
| C11 | `/compact` conserva el cuerpo de la skill | invocar, `/compact`, preguntar por el testigo con su fuente | solo sobrevive el aviso universal de reinvocar; el modelo primero dijo que veía el cuerpo y se corrigió al pedirle la fuente | medido: no lo conserva |
| C12 | Módulo obligatorio ausente | borrar `modules/artifacts.md` y pedir propose | no empieza propose e informa qué módulo falta | medido |
| C13 | Módulo obligatorio vacío, sin regla en el `SKILL.md` (03 en `ed37113`) | `modules/artifacts.md` de 0 B y pedir propose, tres veces | corrida 1: no leyó el módulo (el pedido prohibía ejecutar) y describió un plan; corrida 2: la skill no se cargó, porque Git Bash convirtió `/sdd-orchestrator` en una ruta; corrida 3: leyó el módulo, el cliente avisó *"the file exists but the contents are empty"*, haiku lo informó, sugirió revisar el archivo y se detuvo porque el pedido decía parar tras el primer paso, sin declarar la acción bloqueada | no concluyente: 1 corrida válida, con la parada pedida por el prompt |
| C14 | Modo legado | install de la PR #8 (`01f944c`) con un `claude` que reporta 2.1.109 | `degraded` con aviso *"older than the minimum 2.1.110 … (legacy mode)"*; un bloque `sdd-orchestrator` en `CLAUDE.md`, sin skill | medido, con versión simulada |
| C15 | Módulo obligatorio vacío, con la regla de la 03 (`01f944c`) | skill instalada por la PR #8 en un proyecto, `modules/artifacts.md` de 0 B, `/sdd-orchestrator propose a change named foo-change that adds a hello command`; solo `Read`, `Glob` y `Grep` | 2 de 2: lee el módulo, dice que está vacío, cita la regla (*"the skill says not to run an action whose module is missing or empty"*), nombra la ruta y pide `active-stack install` | medido (n = 2). Límite: la sesión tampoco tenía `Bash` ni `Agent`, y el proyecto no tenía `openspec/`; haiku los nombró como bloqueos aparte y citó la regla del módulo vacío como motivo propio |

### Codex

| Id | Afirmación | Comando | Resultado | Veredicto |
|---|---|---|---|---|
| X1 | El modelo recibe la salida completa | `codex exec --json` con `gen_fail.ps1` | cabeza y cola, mitad y mitad aproximada, con `Warning: truncated output (original token count: N)` y `…N tokens truncated…`; el final real se conserva | medido: recortada |
| X2 | Exit 0 y solo stderr cambian el recorte | `gen_ok.ps1`, `gen_err.ps1` | mismo recorte | medido: igual |
| X3 | Quién fija el tamaño | inspección de las llamadas en el historial | el modelo, con `max_output_tokens` por llamada; tope externo de ~10.000 tokens (~40 KB) | medido |
| X4 | Dónde empieza el recorte | `gen_750.ps1` (32.297 caracteres) | se pierde la línea 375 | medido: por debajo de 32 KB |
| X5 | `tool_output_token_limit` cambia el recorte | `codex exec -c tool_output_token_limit=2000` y `=50000` | sin efecto observable | no concluyente |
| X6 | El modelo avisa que falta el medio | X1–X4 | no; informa la última línea como si la salida estuviera completa | medido |
| X7 | `codex exec --json` es fuente completa | `aggregated_output` del `item.completed` | 129.047 caracteres, líneas 1–3.000, fallo central y final, `exit_code` 1, en stdout y en solo stderr; con exit 0 también completo | medido: fuente independiente para Codex CLI en Windows |
| X8 | El historial guarda lo completo | historial de la sesión y bases internas de `CODEX_HOME` | el historial guarda la versión recortada; la sqlite interna conserva más, con tope, y no se le ofrece al modelo | medido: no es fuente |
| X9 | Las skills se listan sin invocarlas | `skills/list` del App Server (`codex app-server --stdio`) | nombre y descripción listados; el cuerpo no se carga | medido |
| X10 | `$sdd-orchestrator` inyecta el cuerpo | `codex exec` con `$sdd-orchestrator` | el cuerpo entra al contexto; `AGENTS.md` también se ve | medido |
| X11 | `allow_implicit_invocation: false` bloquea la invocación implícita | pedido OPSX sin `$` | la skill no se invoca | medido |
| X12 | Codex lee los módulos | `$sdd-orchestrator` y leer `modules/artifacts.md` | la política de shell rechaza `Get-Content` y `type` (`blocked by policy`) | no probado |
| X12b | Codex lee los módulos (13/11, por la tarde) | skill instalada con `active-stack install --agent codex` (rama de la 05 con `main` `d5e95bd`), copiada a `.agents/skills/` de un proyecto temporal; `codex exec --approve-for-me --json '$sdd-orchestrator … read modules/directed-code-reading.md … reply only with its first line'` | ejecuta `Get-Content -LiteralPath '.agents/skills/sdd-orchestrator/modules/directed-code-reading.md' -TotalCount 1` (exit 0) y responde `## Directed Code Reading` | medido (Codex 0.167.1, Windows) |

### PR #8 (03) y PR #9 (04)

Re-medido el 2026-11-13 sobre **`01f944c`** (03, `origin/claude-native-opsx`), **`e4c7e48`**
(04, `origin/propose/codex-native-opsx`) y **`35a89f8`** (`main`), después de `git fetch`. Después
de medir, el PR #8 se mergeó a `main` (`7a60089`) y el PR #9 avanzó a `91fbb70`, que integra
`main`: R1–R14 valen para las revisiones de arriba, y lo que cambió se re-midió en R15–R20
(subsección siguiente). Las
instalaciones corrieron con `HOME`, `USERPROFILE`, `APPDATA`, `LOCALAPPDATA` y `CODEX_HOME`
apuntando a un directorio temporal (`install --mode custom --custom sdd-orchestrator --agent
<claude|codex> --home <tmp> --no-self-install -y`). Las sondas de Go son tests temporales en un
worktree de cada rama, borrados después; la columna *Comando* describe cada sonda y su código
está en *Sondas de Go*, al final de esta sección. Las dos ramas se miden con las mismas sondas y
el mismo hecho lleva la misma etiqueta; los clientes, los modelos y la cantidad de corridas no son
los mismos (Claude con haiku; Codex sin modelo registrado).

| Id | Afirmación | Comando | Resultado | Veredicto |
|---|---|---|---|---|
| R1 | Las tablas de carga cubren las cinco acciones | 03: `SKILL.md` instalado; `compose_layers.go:238-282`. 04: `codex_package.go:48-107` | 03: filas propose, apply, archive, delegate e inicio y cierre (`SKILL.md` 8.117 B, módulos 8.916 B), **sin explore ni verify**. 04: líneas para propose, apply y archive (`artifacts`), apply (`tdd`), *"before delegating and for verify"* (`delegation`) y `engram-opsx` sin acción; **sin explore**. La matriz común tampoco tiene filas de explore ni verify | medido: ninguna cubre las cinco; faltante compartido con la matriz y con la 05 |
| R2 | Un sexto fragmento `module/*` sin toggle aparece en la 03 | agregar `<!-- as:module/directed-code-reading -->` con un texto testigo al asset y llamar `ComposeLayers` con los cinco toggles | `err=nil`, los mismos cinco módulos, el testigo no aparece en ninguna capa (`compose_layers.go:168-204` arma los módulos por clave fija) | medido: se descarta sin error |
| R3 | El verify de la 03 detecta un módulo vacío | sonda en `internal/verify` con el fixture `installManaged` de la rama: `modules/tdd.md` intacto, de 0 B, solo espacios y borrado | intacto, 0 B y espacios: 0 fallas; borrado: 1 falla (`module "tdd" missing`). `harness_checks.go:200` sólo hace `os.Stat` | medido: el vacío pasa |
| R4 | El verify de la 04 detecta un módulo vacío | la misma sonda sobre `config.InstallManagedSkill` y `verify.ChecksForHarness` con los cinco toggles | intacto: 0 fallas; 0 B y espacios: *"managed skill content differs"*; borrado: falla al abrir. `managed_skill.go:405-439` compara byte a byte y contra el inventario | medido: lo detecta |
| R5 | 03 y 04 se integran sin conflicto | `git merge-tree --write-tree --name-only 01f944c e4c7e48` | conflicto en 11 archivos: `AGENTS.md`, `ARCHITECTURE.md`, `CLAUDE.md`, `harness/config/install.go`, `harness/config/types.go`, `install/plan.go`, `install/steps.go`, `model/harness.go`, `uninstall/steps.go`, `verify/harness_checks.go`, `openspec/matriz-reglas-opsx.md`. Coincide con la lista de la respuesta de la 04 a la revisión del PR #9 | medido en esas revisiones: no. La 04 los concilió al integrar `main` en `91fbb70`: ver R15 |
| R6 | Hay un solo compositor | lectura de `compose_layers.go` (03) y `codex_package.go:20-107` (04) | 03: `ComposeLayers` reubica fragmentos anotados y falla si falta uno; la tabla está en `actionTable`, `exists` y `ExpectedModules`. 04: `ComposeCodexPackage` corta con `splitCodexAsset`, falla si falta un límite, y arma la tabla como string. Las dos agregan `ConfigDeliveryManagedSkill` con el mismo nombre y valor; la respuesta de la 04 propone discriminarla por agente y target | documentado: hay dos |
| R7 | Tests de Go | `go test ./...` en Windows 11 (Go 1.26.5) y en WSL Ubuntu (Go 1.26.5 portable, copia sin `.git`) | Windows: **`main` verde; 03 verde** (suite completa; en `ed37113` fallaban 8 y en `55f66de`, 3); **04: 5 tests en 4 paquetes**, los mismos que en `cd4aa37`: dos por permisos (*"widened to 777"*), dos por privilegio para crear symlinks y `TestManagedSkillRollbackPreservesInstructionBytes` (*"new migration path already exists"*, causa no investigada); los tests de rutas de `e4c7e48` pasan. WSL: las tres ramas fallan los mismos 18 tests en 3 paquetes, todos de entorno (sin `node`, copia sin historia de Git para `eval/baseline/fixture`, ruta Windows del tema del instalador): ninguna diferencia atribuible a 03 ni a 04 | medido |
| R8 | El `SKILL.md` dice qué hacer si falta un módulo | lectura del `SKILL.md` instalado | 03: *"If a module listed in the table is missing or empty, do not run that action: name the module and its path, tell the user to reinstall…"* (`compose_layers.go:279`, commit `9aad12e`). 04: *"If a required module, control or independent reviewer is unavailable, stop only the affected action and explain the missing capability"* (`assets/codex/native/SKILL.md:14`); no dice si un vacío cuenta como faltante | documentado: las dos tienen regla |
| R9 | Reinstalar no cambia nada | install dos veces y `sha256sum` de los archivos del agente | 03: idénticos; un par de markers `universal-rules`. 04: skill y `AGENTS.md` idénticos; cambian líneas en blanco de `config.toml`, que la 04 informa que también pasa en `main` (no re-medido acá) | medido: idempotente |
| R10 | Uninstall deja el home como estaba | install en un home sin archivo de instrucciones, uninstall, listar | 03: queda `.claude/CLAUDE.md` de **0 B**. 04: queda `.codex/AGENTS.md` de **1 B** (`\n`). En las dos el archivo no existía antes | medido: residuo menor, el mismo en las dos |
| R11 | Un lock huérfano se recupera solo (04) | crear a mano `.active-stack/locks/sdd-orchestrator-codex` e instalar y desinstalar | install y uninstall salen con 1: *"managed skill operation busy or lock unavailable at …"*; el mensaje no dice cómo liberarlo (`managed_skill.go:75-87`). La 03 no usa lock | medido; el lock se creó a mano, no se simuló un crash |
| R12 | Un archivo de la skill editado por el usuario | 04: agregar una línea a `modules/tdd.md`, instalar y desinstalar. 03: `uninstall --help` | 04: install sale con 1 (*"managed file changed"*, `managed_skill.go:200`) y uninstall con 1 (*"preserved edited managed file: modules\tdd.md"*, `:498`), sin flag ni instrucción de salida. 03: documenta `--overwrite-edited` para `--strategy restore` | 04 medido; 03 documentado. Preservar la edición es decisión de diseño; lo que falta en la 04 es una salida documentada |
| R13 | Tareas | `grep` de `- [ ]` y `- [x]` | 03 (archivada): 48 cerradas, 13 abiertas (1.1–1.3, 1.5, 1.6, 1.9, 1.12, 1.13, 2.3, 2.6, 8.1–8.3), todas con su motivo en la sección *DEUDA aceptada* de su `tasks.md`. 04 (borrador): 11 de 21 | medido |
| R14 | El inventario de la 04 usa `/` en Windows | mensaje de R12 | la clave se guarda como `modules\tdd.md` | medido; sin falla observada: riesgo si el mismo home se usa desde WSL |

### Después del merge del PR #8: `main` en `7a60089` y la 04 en `91fbb70`

Re-medido el 2026-11-13, después de `git fetch` y de una auditoría independiente, en worktrees
temporales de `origin/main` (**`7a60089`**, merge del PR #8, 12:48:56Z) y de
`origin/propose/codex-native-opsx` (**`91fbb70`**, *"reconcile native OPSX delivery with merged
main"*), borrados después. Windows 11, Go 1.26.5. No se re-midió en WSL, ni install, uninstall,
lock, archivo editado ni corridas con clientes: R9–R12, C14 y C15 siguen valiendo para las
revisiones de arriba.

| Id | Afirmación | Comando | Resultado | Veredicto |
|---|---|---|---|---|
| R15 | `main` y la 04 se integran sin conflicto | `git merge-tree --write-tree --name-only 7a60089 91fbb70`; `git merge-base --is-ancestor` | sin archivos en conflicto (sale con 0); `7a60089` es ancestro de `91fbb70`, que lleva 6 commits propios | medido: sí. Antes, 11 conflictos (R5) |
| R16 | Con `main` integrado, la 04 deja un solo modo de entrega y su propio compositor | `git grep` en `91fbb70`; `git diff --stat e4c7e48 91fbb70` | un solo `ConfigDeliveryManagedSkill` (`internal/model/harness.go:90`); `ComposeLayers` sólo se llama para Claude (`internal/harness/config/install.go:216`) y Codex usa `ComposeCodexPackage` (`managed_skill.go:95`). `codex_package.go`, `managed_skill.go` y `assets/codex/` no cambian desde `e4c7e48`: R1 (string `actions` en `codex_package.go:48`), R8 (`SKILL.md:14`) y la línea de `SKILL.md:20` siguen iguales. Con los cinco toggles, el paquete de Codex no trae ninguna anotación `<!-- as:` (sonda, 0 en 6 archivos) | medido: dos compositores, uno por cliente, por decisión documentada de la 04 |
| R17 | Lo de la 03 medido en `01f944c` vale en `main` | `git rev-parse 01f944c^{tree} 7a60089^{tree}`; sondas de R2 y R3 en `main` | el árbol de `main` es el de `01f944c` (`9e70a83`). Sonda de verify: intacto, 0 B y espacios, 0 fallas; faltante, 1. Sexto fragmento: `err=nil`, los mismos cinco módulos, testigo ausente. `actionTable` sigue sin explore ni verify (`compose_layers.go:238-282`), y la regla de vacío en `:279` | medido: R1, R2, R3 y R8 de la 03 siguen en `main` |
| R18 | El verify de la 04 sigue detectando un módulo vacío | sonda de R4 en `91fbb70` | intacto: 0 fallas; 0 B y espacios: *"managed skill content differs"*; faltante: falla al abrir | medido: lo detecta |
| R19 | Tests de Go en Windows | `go test ./...` en cada worktree | `main`: 38 paquetes `ok`, sale con 0. 04: 6 tests fallan en 4 paquetes: los 5 de R7 y `TestSharedManagedDeliveryRollbackPreservesConcurrentCodexEdits`, nuevo, en los dos órdenes. Este último falla porque el error de rollback cita la ruta con `%q` y duplica las `\` (`internal/backup/restore.go:65`), y el test busca la ruta literal. La misma línea con `%q` está en `main` (`restore.go:39`); no se midió si ahí llega al usuario. La 04 informa la suite verde en macOS (`integracion-main-pr8-2026-11-13.md`) | medido |
| R20 | Tareas | `grep` de `- [ ]` y `- [x]` | 03, archivada en `main` (`openspec/changes/archive/2026-11-13-claude-native-opsx/`): 48 cerradas y 13 abiertas, con *DEUDA aceptada*. 04 (`91fbb70`): 11 de 21; abiertas 1.1–1.4, 2.4, 2.5, 4.2–4.4 y 4.6 | medido: sin cambios desde R13 |

### Grupo 8 con Claude Code: corridas de aceptación

Medido el 2026-11-13 con Claude Code 2.1.319 en Windows 11 y modelo haiku:
`claude -p … --model haiku --setting-sources project --strict-mcp-config`, `MSYS_NO_PATHCONV=1`,
la skill colocada a mano en un proyecto temporal y los evaluadores del repositorio. Fueron 50
llamadas a Claude. Los textos de módulo son los de la §2; G1 y G4 a G6 corrieron con la versión
anterior a la última iteración de `bounded-command-output`, que solo agregó la línea final
`FAIL\t<package>`, la indentación de cada `failure:` y la regla de no escribir `ref:` sin
original. Linux, Codex (la política de shell de Windows bloquea su shell) e IDE quedaron sin
correr en este entorno. *(13/11 por la tarde: Codex y Linux ya corren; ver §5, 8.1. Las corridas de Codex de 8.6 y 8.7 son deuda D05-0806 y D05-0807.)*

| Id | Tarea | Comando | n | Resultado | Veredicto |
|---|---|---|---|---|---|
| G1 | 8.2 | `/sdd-orchestrator` apply con `./emit.exe <fixture>` en Bash, captura del agente; `EvaluateCopy` con la copia citada | 3 por fixture, 10 fixtures | 1.ª iteración: 26/30; rechazados `fail-middle` (omitió `FAIL\t…`), `scattered` (perdió la indentación) y `never-captured` ×2 (`ref: none`). 2.ª: 11/11 en `fail-middle`, `scattered`, `never-captured` ×3, `fail-start` y `fail-end` | aceptado 30/30 con el texto final |
| G2 | 8.2 | lo mismo con `noise-exit0` y `control-negative` | 3 + 3 | `failure: none` sin agregados; antes, 0/4 por `failure: none (…)` | aceptado 6/6 |
| G3 | 8.3 | lo mismo con la herramienta PowerShell | 2 por fixture (`fail-middle`, `stderr-only`) | `cmd /c` y `Set-Content -Encoding ascii` en 4/4; la copia es byte a byte la del arnés. El rechazo anterior venía del BOM y del CRLF del archivo, que el evaluador ahora normaliza solo en el resumen | aceptado 4/4 |
| G4 | 8.4 | verify sobre A (agente, coherente), B (arnés, fallo omitido) y C (arnés, cobertura `demostrada`), en directa y delegada | 1 + 1 | A *sin verificar*, B *rechazado*, C *verificado* en las dos, igual que `summary.Corroborate` | aceptado 6/6 |
| G5 | 8.5 | captura del arnés de `sensitive`; apply y luego verify pide el original | 3 | `EvaluateCopy` aceptado 3/3; 0 valores crudos en copias, resumen, respuesta y transcripción; niega el original, da ruta y vencimiento a 7 días y declara el límite de la copia del cliente 3/3 | aceptado 3/3 |
| G6 | 8.10 | explore: «¿qué pasa si `Compose` recibe un toggle desconocido?» (caso de 7.1) | 3 | dos líneas, con `src/compose.go:36-40` en la primera; antes, 0/3 por el relato agregado | aceptado 3/3 |
| G7 | 8.6 | fixtures de 4.6 en corridas reales: símbolo en archivo de 422 líneas, quinta condición (361), relectura sin y con escritura, orientación; `readfix.Evaluate` sobre el registro de la transcripción. Skill generada con `ComposeLayers`+`InstallLayered` (integración propuesta de 6.5, pendiente de que la 03 la adopte) | 3/3/2/2/3; iter. 2: 2+1 | quinta condición 3/3 (`FifthCondition`=1, motivo); escritura 2/2 y orientación 3/3 sin registros; símbolo 0/3 (lee el archivo entero antes del módulo) y relectura 0/2 sin `unnecessary call:`. Iteración 2 del módulo y del texto del `SKILL.md`: relectura 1/1; símbolo 0/2 (sigue abriendo el archivo entero) | parcial |
| G8 | 8.7 | fixtures de 5.4 con `toolfix.Evaluate`: Context7 presente, caído (`--mcp-config` con comando inexistente), no habilitado en verify, repetición, fuera de clase (Context7 en verify, Playwright en explore); nivel b con skill testigo `disable-model-invocation` | 3/3/2/2/3+1/3+1; iter. 2: 2+1+1 | presente 3/3 y no habilitada 2/2 sin registros; caído 0/3 y 0/2 sin `unavailable:`; fuera de clase 0/4 y 0/1 sin `out-of-class use:`; repetición: escribe la línea en la iteración 2, el registro no la empareja por el formato del argumento. Nivel efectivo en Claude: (a) para MCP (las llamadas fuera de clase pasan, 0 denegaciones); (b) para skills: testigo invisible 3/3, visible en el control 1/1 | parcial |
| G9 | 8.8 | módulo desde la skill: `delegation` vacío, ausente en verify y en apply, dos pedidos, sin subagentes (`--disallowedTools`), retomar con `--resume`, `/compact` con `--resume` | 2 por caso; iter. 2: 1+1 | vacío 2/2, sin subagentes 2/2 y partes separadas 2/2 aceptados; apply sin `delegation`: 0/2 (se detuvo o delegó igual), iter. 2 resuelve directo sin delegar 1/1 pero sin la línea fija; retomar: retoma 2/2 sin decir por qué, iter. 2 anuncia la línea pero no la escribe; tras `/compact` 0/2 relee módulos: el modelo no puede invocar la skill (`disable-model-invocation: true`) y el texto universal de la 03 le pide invocarla | parcial; propuesta a la 03 |
| G10 | 9.1 | explore «limpieza de `useEffect`» con y sin Context7, mismo prompt (sha256 `80f19766…`) y misma aceptación (`acc91.py`, sha256 `82c40bce…`) | 3 + 3 | prompt y aceptación idénticos byte a byte en los dos brazos | aceptado |
| G11 | 9.2 | `eval/baseline` (`manifest`, `capture.ParseClaudeStream`, `summary.Compute`) por brazo | 3 + 3 | mismos prompt, rúbrica y fixture; cambian solo `variant` y la huella de configuración. Con Context7: +2 herramientas MCP, entrada de la 1.ª solicitud 40.800 contra 40.530, 519.884 contra 377.124 tokens, 2/3 aceptadas en ambos. La 02 rechaza `system/permission_denied` con `message` string (se saneó en 3 líneas; propuesta a la 02) | aceptado |
| G12 | 9.3 | apply «renombrar `available` en `reserve`» (mismo prompt que G7) con todos los módulos y sin cada uno (toggle apagado o archivo quitado, árbol que difiere solo en ese módulo); `acc93.py` sha256 `7adb0dac…` | 2 + 8×1 | aceptado en todos salvo sin `bounded-command-output` (obligatorio: se detuvo, `missing module`). Medido sobre la integración propuesta de 6.5, pendiente de la 03 | medido, no cierra |

### 8.1: cobertura por fuente con los fixtures del arnés (13/11, por la tarde)

Fixtures `fail-middle` (2.1, 218.919 caracteres) y `stderr-only` (2.7, 87.706 caracteres) y, para exit 0, `control-negative` (65.991 caracteres). Cada uno se emite con un ejecutable que llama a `fixtures.ByID(...).Emit` y sale con el código del fixture. Las fuentes se comparan con `clientsrc.Compare` (3.6), ya corregido. n = 1 por celda. Linux es Ubuntu 26.04 sobre WSL2, con los clientes nativos en `~/.local`. En Windows, la shell de Codex corre con `codex exec --approve-for-me`.

| Id | Cliente · interfaz · sistema | Fuente | Comando | Resultado | Veredicto |
|---|---|---|---|---|---|
| F1 | Codex 0.167.1 · CLI · Windows | `aggregated_output` de `codex exec --json` | `codex exec --approve-for-me --skip-git-repo-check --json "Run exactly … .\emit.exe <fixture> …"` | 218.919 y 87.706 caracteres, exit 1, el fallo presente | **completa** |
| F2 | Codex 0.169.0 · CLI · Linux | igual | igual, con `./emit` | igual | **completa** |
| F3 | Claude 2.1.323 · CLI · Linux | `tool_result` del modelo (`stream-json`) | `claude -p "Run exactly … ./emit <fixture> …" --output-format stream-json --verbose --model haiku --allowedTools Bash` | ~10.040 caracteres con `Exit code 1`; el fallo central se pierde | **truncada** |
| F4 | Claude 2.1.323 · CLI · Linux | hook `PostToolUseFailure` (`error`) | mismos pedidos, con hooks en `.claude/settings.json` que vuelcan la entrada a un archivo | ~10.028 caracteres, marca `[20012 characters truncated]` entre líneas en blanco, cabeza y cola cortadas a mitad de línea; la cola es un tramo interior | **truncada** (con 3.6 corregido) |
| F5 | Claude 2.1.323 · CLI · Linux | hook `PostToolUse` con exit 0 | `control-negative` | `stdout` cortado en 30.000; `persistedOutputPath` con 65.991 B completos | **truncada**; archivo **completo** solo con exit 0 |
| F6 | Claude 2.1.319 · CLI · Windows | F3 a F5 | mismos pedidos con `./emit.exe` | igual que en Linux; además el cliente agrega a `stderr` la nota `Shell cwd was reset to …`, que el adaptador ahora descarta | **truncada**, igual que Linux |
| F7 | Codex 0.167.1 · App Server · Windows | `item/commandExecution/outputDelta` y `aggregatedOutput` de `item/completed` | `codex app-server` conducido por JSON-RPC: `initialize`, `thread/start` (`approvalPolicy: never`, `sandbox: danger-full-access`, en un directorio temporal) y `turn/start` | deltas y `aggregatedOutput` completos e idénticos, exit 1 | **completa** |
| F8 | Codex 0.169.0 · App Server · Linux | igual | igual | sin deltas; `aggregatedOutput` completo en `fail-middle`; en `stderr-only` trae todas las líneas pero la de stderr aparece en la 1.175 y no al final | **reordenada** |
| F9 | Linux | tests 3.3 y 3.8 | `go test ./eval/taskcontext/capture/ ./eval/taskcontext/recipes/` con Go 1.26.5 | 17 pasan y 1 se omite (PowerShell) | verde |

Lo que queda abierto: la extensión de VS Code y la app de escritorio de Claude Code, que según la documentación disparan los mismos hooks, y el IDE en Linux (deuda D05-0801). Registrado en `eval/taskcontext/clientsrc/matrix.go` y en `design.md` (D22).

### Sondas de Go

Las sondas son tests temporales: se copian al paquete indicado del worktree, se corren con
`go test -count=1 -run <Nombre> -v ./<paquete>/` y se borran. Leen el resultado de las líneas
`PROBE…` de la salida.

R3 y R17, verify de la 03 (`internal/verify/zz_remedicion03_verify_test.go`; usa el fixture
`installManaged` de la rama):

```go
package verify_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
)

func TestZZRemedicionVerifyEmptyModule(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(f managedFixture)
	}{
		{"intacto", func(f managedFixture) {}},
		{"vacio-0-bytes", func(f managedFixture) { _ = os.WriteFile(f.skillFile("modules/tdd.md"), nil, 0o644) }},
		{"solo-espacios", func(f managedFixture) { _ = os.WriteFile(f.skillFile("modules/tdd.md"), []byte("  \n\n\t\n"), 0o644) }},
		{"faltante", func(f managedFixture) { _ = os.Remove(f.skillFile("modules/tdd.md")) }},
	}
	for _, c := range cases {
		f := installManaged(t)
		c.mutate(f)
		failed, warned := statuses(f.verify(t, claude.VersionCompatible))
		fmt.Printf("PROBE03 %s failed=%d warned=%d %v\n", c.name, len(failed), len(warned), failed)
	}
}
```

R4 y R18, verify de la 04 (`internal/verify/zz_remedicion04_verify_test.go`; usa
`managedVerifyAdapter` y `newFakeAdapter` de la rama):

```go
package verify_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

func TestZZRemedicionVerifyEmptyModule04(t *testing.T) {
	for _, c := range []struct {
		name    string
		content *string
		remove  bool
	}{
		{name: "intacto"},
		{name: "vacio-0-bytes", content: ptr("")},
		{name: "solo-espacios", content: ptr("  \n\n\t\n")},
		{name: "faltante", remove: true},
	} {
		home := t.TempDir()
		a := managedVerifyAdapter{newFakeAdapter(t, home, model.AgentCodex)}
		a.delivery = model.ConfigDeliveryManagedSkill
		h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation", "model-routing", "engram", "tdd", "governance"}}
		if _, err := config.InstallManagedSkill(h, a, home); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(a.SkillsDir(home), h.ID, "modules", "tdd.md")
		if c.remove {
			_ = os.Remove(p)
		} else if c.content != nil {
			_ = os.WriteFile(p, []byte(*c.content), 0o600)
		}
		checks := verify.ChecksForHarness(h, []verify.Adapter{a}, home)
		var errs []string
		for _, ch := range checks {
			if err := ch.Run(context.Background()); err != nil {
				errs = append(errs, err.Error())
			}
		}
		fmt.Printf("PROBE04 %s checks=%d failed=%d %v\n", c.name, len(checks), len(errs), errs)
	}
}

func ptr(s string) *string { return &s }
```

R2 y R17, sexto fragmento en la 03 (`internal/harness/config/zz_remedicion_sixth_test.go`).
Antes de correrla, se insertan estas líneas en `internal/harness/config/assets/claude/sdd-orchestrator.md`
justo antes de `<!-- as:module/model-assignments -->`, y después se restaura el asset:

```markdown
<!-- as:module/directed-code-reading -->
## Probe

PROBE_SIXTH_MODULE_TEXT

```

```go
package config

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestZZRemedicionSixthModule(t *testing.T) {
	l, err := ComposeLayers([]string{"delegation", "model-routing", "engram", "tdd", "governance"})
	fmt.Printf("PROBE_SIXTH err=%v\n", err)
	var names []string
	for n := range l.Modules {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Printf("PROBE_SIXTH modules=%v\n", names)
	all := l.Universal + l.Skill.Body
	for _, m := range l.Modules {
		all += m
	}
	fmt.Printf("PROBE_SIXTH text_present=%v\n", strings.Contains(all, "PROBE_SIXTH_MODULE_TEXT"))
}
```

R16, anotaciones en el paquete de Codex de la 04 (`internal/harness/config/zz_remedicion04_annot_test.go`):

```go
package config

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestZZRemedicionCodexAnnotations(t *testing.T) {
	pkg, err := ComposeCodexPackage([]string{"delegation", "model-routing", "engram", "tdd", "governance"})
	fmt.Printf("PROBE_ANNOT err=%v\n", err)
	var names []string
	leak := strings.Count(pkg.Universal, "<!-- as:")
	for n, c := range pkg.Files {
		names = append(names, n)
		leak += strings.Count(c, "<!-- as:")
	}
	sort.Strings(names)
	fmt.Printf("PROBE_ANNOT files=%v annotations=%d\n", names, leak)
}
```

## 6. No probado y por qué

| Qué | Por qué |
|---|---|
| ~~Lectura real de módulos por Codex~~ | **medido el 13/11 (X12b)** con `codex exec --approve-for-me` |
| Módulo ausente y módulo vacío en Codex | mismo bloqueo: el coordinador no puede leer el directorio de la skill |
| Si un subagente de Codex lee `AGENTS.md` | lanzar y observar subagentes necesita shell (mismo bloqueo) |
| Descubrimiento en `%USERPROFILE%\.agents\skills` con un HOME temporal | Codex ignora `HOME` y `USERPROFILE` para esa ruta; probarlo exigía escribir en el perfil real |
| `/skills` en Codex sin interfaz | `codex exec` no lo expone |
| IDE y app de escritorio, en los dos clientes; Codex en la nube | la extensión y la app no envían un pedido sin una persona (el link `prompt`/`q` solo lo precarga, según la documentación oficial); la 04 no promete escritorio ni nube (`codex-native-opsx/design.md:59-63`). Deuda D05-0801 |
| ~~Clientes en Linux~~ | **medido el 13/11 (8.1):** Claude 2.1.323 y Codex 0.169.0 nativos en Ubuntu 26.04 sobre WSL2 |
| Claude 2.1.109 real | se simuló la versión ante el instalador (C14) |
| ~~Hooks `PostToolUse` y `PostToolUseFailure` de Claude como fuente~~ | **medido el 13/11 (8.1)** en Windows y Linux: truncados |
| Conteo de tokens | sin API key no hay contador oficial; las cifras son caracteres y bytes |
| Efecto de `tool_output_token_limit` | X5 sin diferencia observable |
| ~~Instalación real del árbol integrado con los dos clientes a la vez~~ | **medido el 13/11** sobre `main` `d5e95bd` y sobre la 07 en Linux, en la revisión del PR #13: install, reinstall y uninstall cruzado sin afectar al otro cliente byte a byte. Sobre esta rama: install de Codex para X12b |
| Módulo vacío en Claude como conducta estable | sin regla, una sola corrida válida (C13); con la regla de la 03, 2 de 2 (C15), con `Bash` y `Agent` deshabilitados: alcanza para ver que la regla se sigue, no para medir una tasa |
| Representatividad | n = 1 por prueba (n = 2 en C15), un modelo (haiku en Claude); Codex sin modelo registrado |
| `-race` | no corrió en ninguna plataforma (Windows sin cgo; en WSL no se pidió) |

### Lo que la 05 no probó de lo suyo

La 05 exige a 03 y 04 evidencia medida; con el mismo criterio, su propia entrega es la más débil
de las tres en estos puntos:

| Qué | Estado |
|---|---|
| Los cuatro módulos de la §2 | **nunca corrieron en un cliente real**: no hay sonda que los instale, los cargue y observe la conducta. Su efecto (menos contexto, fallos conservados, formas fijas respetadas) es hipótesis hasta las tareas 8.x y 9.x |
| La tabla de la §3 | no se compuso en ninguna de las dos ramas ni se probó con un modelo; las filas nuevas de `explore` y `verify` no tienen corrida propia |
| El registro común de D23 | es un diseño en texto; no hay prototipo ni test, a diferencia de los compositores de 03 y 04, que tienen tests en sus ramas |
| Evidencia compartible | las sondas de Go de R2–R4 y R17–R18 están en la §5 (*Sondas de Go*), pero los logs y las transcripciones de las corridas con clientes quedaron en la máquina de quien las corrió. La 04 versiona sus informes en `openspec/changes/codex-native-opsx/informes/` y la 03 su investigación en su change |
| Tareas | 0 de 64 cerradas; 03 y 04 entregan código funcionando (§5, R7–R14) |

## 7. Pendientes por ficha

Primero lo que debe la 05; después, lo que la 05 **propone** a las demás fichas. Las propuestas
llevan su razón y quedan abiertas a contrapropuesta: no son obligaciones para sus ramas. Como la
03 ya está en `main`, lo que se le propone es un cambio sobre `main` (issue o PR aparte), no
sobre el PR #8. A la 04 se le propone alinearse con el plan de consumo que ella misma publicó.

### Lo que debe la 05

| Qué | Tareas | Contra qué |
|---|---|---|
| Probar sus propios módulos en un cliente real antes de pedir que se adopten: instalarlos, cargarlos por acción y medir con el protocolo de 02 | 8.x, 9.3, 9.4 | §6, *Lo que la 05 no probó de lo suyo* |
| Pasar los textos de la §2 a `internal/harness/config/assets/modules/` con sus tests | 4.4, 5.3, 6.3, 7.2 | §2 |
| Declarar la tabla en el catálogo y generarla, si 03, 04 y 07 la acuerdan | 6.1, 6.2 | §3 |
| El adaptador de `codex exec --json` en el arnés | 3.6 | X7 |
| Versionar los logs y las transcripciones de sus corridas en un lugar que el grupo pueda abrir (el código de las sondas de Go ya está en la §5) | (TBD: dónde) | §5 |

### Lo que la 05 propone (a acordar)

| Ficha | Propuesta | Razón | Ya resuelto en su rama (re-medido) |
|---|---|---|---|
| **03** (Juani), sobre `main` | Que verify trate como faltante un módulo de 0 bytes o solo con espacios. Que `ComposeLayers` falle si sobra un fragmento `module/*`, o que acepte módulos sin toggle (los cuatro de la 05). Que la tabla de la §3 sea la fuente única, con `explore` y `verify`. Con la 04, si se acuerda, un compositor común con registro de módulos (`design.md` — D23). Tarea 2.3 (en su *DEUDA aceptada*): revisar los textos de la §2 | R1 (sin filas de `explore` ni `verify`, igual que la matriz), R2 y R3, re-medidos en `main` (R17) | PR mergeado a `main` (`7a60089`); suite completa verde en Windows (R7, R19); regla de módulo faltante o vacío en el `SKILL.md` (R8); rutas literales en los avisos; modo legado (C14); deuda de las 13 tareas abiertas documentada (R13) |
| **04** (Matías, según el informe unificado) | Ejecutar su plan de consumo de esta entrega (`informes/analisis-dependencias-05-06-2026-11-13.md`), que la 05 toma como base: los cuatro módulos en `Files`, la tabla de la §3 adaptada en lugar del string `actions`, con una fila de `explore` y una propia de `verify`. Fijar la revisión de este archivo en la tarea 1.4 (`SKILL.md:20`); como esta revisión agrega filas a la §5, fijar la que se publique en el PR de la 05. Probar lectura de módulos, módulo ausente y vacío y subagentes cuando la política de shell lo permita (§6). Si se acuerda, el compositor común; si no, su decisión de mantener uno propio para Codex sigue en pie | R1 (sin fila de `explore`; `verify` cubierto en la línea de `delegation`, sin fila propia), R6, R16, X12 | verify byte a byte que detecta vacío, espacios y faltante (R4, R18); regla general de módulo faltante (R8); rutas literales en sus diagnósticos; conciliación con `main` ejecutada en `91fbb70` (un solo `ConfigDeliveryManagedSkill`, sin conflictos, R15 y R16); plan para consumir esta entrega en Codex (`informes/analisis-dependencias-05-06-2026-11-13.md`) |
| **03 y 04** | Mismo criterio para el residuo de uninstall (R10): borrar el archivo si queda vacío y antes no existía. Para archivos editados (R12), sólo la 04: la 03 ya documenta `--overwrite-edited`. Revisión conjunta del árbol integrado (tarea 4.6 de la 04) antes de sacar la 04 de borrador | R10: el mismo hecho en las dos ramas, la misma etiqueta: menor. Las dos son dominio ALTO con aprobación humana registrada: la 03, de Juani (`archive/2026-11-13-claude-native-opsx/proposal.md`, *Aprobación*); la 04, que además toca `internal/backup` y `filemerge/writer.go`, de Matías (`codex-native-opsx/proposal.md:39`) | 03: `--overwrite-edited` (R12) |
| **06** (Gianfranco) | Acordar con la 05, y con la 04 como consumidora (deudas D04-0601 a D04-0604 de `codex-native-opsx/informes/deudas-aceptadas-05-06-07-2026-11-13.md`): (1) ubicación, acceso y vencimiento de la copia redactada y el registro de captura (D9 propone caché del usuario fuera del repo, acceso local y retención hasta archivar el change más siete días; los siete días son propuesta, no acuerdo); (2) la lista de patrones sensibles (D14); (3) la identidad exacta del estado probado (árbol y huella) que usan los recibos; (4) los criterios de suficiencia: qué detalle debe conservar la copia redactada para que un requisito no quede sin verificar; (5) que su contrato de evidencia se cargue con `bounded-command-output` en `apply` y `verify`, en ejecución directa y delegada. Hasta ese acuerdo, un resultado sin captura independiente queda *sin verificar* | §4, D9, D14, D20 | — |
| **07** (Neyen) | Acordar con la 05, y con la 04 como consumidora (deudas D04-0701 y D04-0702): (1) las filas de delegación de la §3 (S4: `delegation` obligatorio en `verify`); (2) el umbral de 300 líneas o 10.000 caracteres de la 05 frente a su señal S1 y la profundidad D10; (3) que el sobre de retorno referencie la copia redactada y el registro de captura de la 05 | §3, D3, `CHANGES.md` pide mapear la carga con la 07 | — |
| **02** (Giuliano) | Usar la lista de D14 en `eval/baseline/record` y el plazo de D9 en su D6; que el arnés de la 05 reutilice su manifiesto | §4 | — |

## 8. Propuesta de integración para la 03 (no aplicada)

La 05 no modifica código de otro change. Este diff es la conexión que hace falta en el compositor de la 03 para que `active-stack install` entregue los cuatro módulos y la tabla de la 05 en Claude, con verify que detecta módulos vacíos y error ante un módulo anotado sin toggle. Se probó en una copia local (suite verde y las corridas G de §5 hechas sobre esta integración) y se revirtió antes de commitear. Integrarlo, adaptarlo o rechazarlo es decisión de la 03 (tareas 6.5, 6.6, 6.7 y 9.3 de la 05 dependen de eso).

```diff
diff --git a/internal/harness/config/compose_layers.go b/internal/harness/config/compose_layers.go
index ee3f25f..47da56c 100644
--- a/internal/harness/config/compose_layers.go
+++ b/internal/harness/config/compose_layers.go
@@ -4,7 +4,10 @@ import (
 	"fmt"
 	"io/fs"
 	"regexp"
+	"sort"
 	"strings"
+
+	"github.com/Group-Active-IA/active-stack/internal/catalog"
 )
 
 // Layers is the layered Claude composition (claude-native-opsx 4.1): the
@@ -59,6 +62,9 @@ func ComposeLayers(toggles []string) (Layers, error) {
 		}
 		byName[key] = f.text
 	}
+	if err := checkWiredModules(byName); err != nil {
+		return Layers{}, err
+	}
 	frag := func(key string) (string, error) {
 		s, ok := byName[key]
 		if !ok {
@@ -143,10 +149,14 @@ func ComposeLayers(toggles []string) (Layers, error) {
 	if err != nil {
 		return Layers{}, err
 	}
+	table, err := coordinatorTable(toggles)
+	if err != nil {
+		return Layers{}, err
+	}
 	head := append(roleParts,
 		renumbered(strings.TrimSpace(coreHeading), []string{pDelegate}),
 		strings.TrimSpace(rulesHeading)+"\n\n"+strings.TrimSpace(cRules),
-		actionTable(active))
+		table)
 	restKeys := []string{"coordinator/starting-work", "coordinator/workflow", "coordinator/launch-pattern"}
 	if active["governance"] {
 		restKeys = append(restKeys, "coordinator/governance-apply")
@@ -202,6 +212,15 @@ func ComposeLayers(toggles []string) (Layers, error) {
 		modules["tdd"] = joinBlocks("Strict TDD Mode: enabled\n\n" + strict)
 	}
 
+	// Task-context modules (opsx-task-context, D23): no toggle, always delivered.
+	for _, id := range TaskContextModules {
+		text, err := fs.ReadFile(embeddedAssets, "assets/modules/"+id+".md")
+		if err != nil {
+			return Layers{}, fmt.Errorf("config compose layers: %w", err)
+		}
+		modules[id] = string(text)
+	}
+
 	return Layers{
 		Universal: universal,
 		Skill: SkillSpec{
@@ -232,53 +251,51 @@ func renumbered(heading string, items []string) string {
 	return strings.Join(out, "\n\n")
 }
 
-// actionTable is the action -> modules table of the coordinator body (matrix,
-// "Reglas de la tabla"). A module whose toggle is inactive is not listed, and a
-// row left without modules is dropped.
-func actionTable(active map[string]bool) string {
-	exists := func(n string) bool {
-		switch n {
-		case "artifacts":
-			return true
-		case "tdd":
-			return active["tdd"]
-		case "delegation":
-			return active["delegation"]
-		case "model-assignments":
-			return active["model-routing"]
-		case "engram-opsx":
-			return active["engram"]
-		}
-		return false
-	}
-	rows := []struct {
-		action  string
-		modules []string
-	}{
-		{"propose", []string{"artifacts"}},
-		{"apply", []string{"artifacts", "tdd"}},
-		{"archive", []string{"artifacts"}},
-		{"delegate", []string{"delegation", "model-assignments"}},
-		{"session start and close", []string{"engram-opsx"}},
-	}
-	var b strings.Builder
-	b.WriteString("## Action to Modules\n\n")
-	b.WriteString("Before an action, read ONLY the modules listed for it. When you delegate, paste their content into the sub-agent prompt: the sub-agent does not see this skill.\n\n")
-	b.WriteString("| Action | Modules to read |\n|--------|-----------------|\n")
-	for _, r := range rows {
-		var refs []string
-		for _, m := range r.modules {
-			if exists(m) {
-				refs = append(refs, fmt.Sprintf("`${CLAUDE_SKILL_DIR}/modules/%s.md`", m))
-			}
-		}
-		if len(refs) > 0 {
-			b.WriteString("| " + r.action + " | " + strings.Join(refs, ", ") + " |\n")
+// TaskContextModules are the modules of opsx-task-context (ficha 05). They have
+// no toggle: they are delivered always, from assets/modules/.
+var TaskContextModules = []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"}
+
+// wiredModuleFragments are the annotated module fragments ComposeLayers consumes.
+var wiredModuleFragments = map[string]bool{
+	"module/artifacts": true, "module/delegation": true, "module/model-assignments": true,
+	"module/engram-opsx-wiring": true, "module/engram-opsx-git-sync": true,
+}
+
+// checkWiredModules (R2): an annotated module fragment that ComposeLayers does
+// not wire would be dropped silently. Fail loudly instead; D23 registers modules
+// explicitly, so an unregistered one is an integration bug, not a module to guess.
+func checkWiredModules(byName map[string]string) error {
+	var bad []string
+	for key := range byName {
+		if strings.HasPrefix(key, "module/") && !wiredModuleFragments[key] {
+			bad = append(bad, key)
 		}
 	}
-	b.WriteString("\nIf a module listed in the table is missing or empty, do not run that action: name the module and its path, tell the user to reinstall with `active-stack install`, and continue with the actions that do not depend on it.\n\n")
-	b.WriteString("NEVER delegate the coordinator role to a sub-agent or ask one to act as `sdd-orchestrator`.\n")
-	return b.String()
+	sort.Strings(bad)
+	if len(bad) > 0 {
+		return fmt.Errorf("config compose layers: annotated module fragment %q is not registered in ComposeLayers; register it or remove the annotation", bad[0])
+	}
+	return nil
+}
+
+// coordinatorTable is the action -> modules text of the SKILL.md. It is
+// generated from the catalog (opsx-task-context 6.2, Claude path form) for the
+// active toggles, followed by the D21 assignment rule and the coordinator rule.
+func coordinatorTable(toggles []string) (string, error) {
+	cat, err := catalog.Load()
+	if err != nil {
+		return "", fmt.Errorf("config compose layers: %w", err)
+	}
+	tbl, err := cat.TaskContext.RenderActionTableFor("claude", toggles)
+	if err != nil {
+		return "", err
+	}
+	rule, err := cat.TaskContext.AssignmentRule("claude")
+	if err != nil {
+		return "", err
+	}
+	const never = "NEVER delegate the coordinator role to a sub-agent or ask one to act as `sdd-orchestrator`."
+	return strings.TrimRight(tbl, "\n") + "\n\n" + strings.TrimRight(rule, "\n") + "\n\n" + never + "\n", nil
 }
 
 // joinBlocks trims each block, drops empty ones, joins with a blank line and
@@ -296,7 +313,7 @@ func joinBlocks(blocks ...string) string {
 // ExpectedModules lists the module names ComposeLayers produces for the given
 // toggles (artifacts always), so verify can check that none is missing.
 func ExpectedModules(toggles []string) []string {
-	names := []string{"artifacts"}
+	names := append([]string{"artifacts"}, TaskContextModules...)
 	for _, t := range toggles {
 		switch t {
 		case "delegation":
diff --git a/internal/harness/config/compose_layers_test.go b/internal/harness/config/compose_layers_test.go
index 3519cf5..da02b62 100644
--- a/internal/harness/config/compose_layers_test.go
+++ b/internal/harness/config/compose_layers_test.go
@@ -88,16 +88,22 @@ func TestComposeLayers_SkillMDHasRoleTableAndNoModuleText(t *testing.T) {
 
 // 4.2 (M): every module exists only with its toggle; artifacts always.
 func TestComposeLayers_ModulesFollowToggles(t *testing.T) {
-	if got := moduleNames(mustLayers(t, nil)); strings.Join(got, ",") != "artifacts" {
-		t.Errorf("no toggles: modules = %v, want [artifacts]", got)
+	if got := moduleNames(mustLayers(t, nil)); strings.Join(got, ",") != "artifacts,bounded-command-output,concise-agent-reporting,directed-code-reading,task-tool-selection" {
+		t.Errorf("no toggles: modules = %v, want artifacts + the four task-context modules (no toggle)", got)
 	}
-	if got := moduleNames(mustLayers(t, allToggles)); strings.Join(got, ",") != "artifacts,delegation,engram-opsx,model-assignments,tdd" {
+	if got := moduleNames(mustLayers(t, allToggles)); strings.Join(got, ",") != "artifacts,bounded-command-output,concise-agent-reporting,delegation,directed-code-reading,engram-opsx,model-assignments,task-tool-selection,tdd" {
 		t.Errorf("all toggles: modules = %v", got)
 	}
 	for toggle, module := range map[string]string{"tdd": "tdd", "delegation": "delegation", "model-routing": "model-assignments", "engram": "engram-opsx"} {
 		got := moduleNames(mustLayers(t, []string{toggle}))
-		if strings.Join(got, ",") != "artifacts,"+module {
-			t.Errorf("toggle %s: modules = %v, want artifacts + %s", toggle, got, module)
+		var base []string
+		for _, n := range got {
+			if n != module {
+				base = append(base, n)
+			}
+		}
+		if len(got) != len(base)+1 || strings.Join(base, ",") != "artifacts,bounded-command-output,concise-agent-reporting,directed-code-reading,task-tool-selection" {
+			t.Errorf("toggle %s: modules = %v, want the base set + %s", toggle, got, module)
 		}
 	}
 	l := mustLayers(t, []string{"governance"})
@@ -243,18 +249,21 @@ func tableOf(t *testing.T, toggles []string) string {
 	return md[i:j]
 }
 
-// Action -> modules table: fixed rows, and a module of an inactive toggle is not listed.
+// Action -> modules table (generated from the catalog, 6.5): module-less toggles
+// are not listed, a row left without modules is dropped, and the 05 modules
+// (no toggle) are always there.
 func TestComposeLayers_ActionTableFollowsToggles(t *testing.T) {
+	p := func(id string) string { return "`${CLAUDE_SKILL_DIR}/modules/" + id + ".md`" }
 	all := tableOf(t, allToggles)
 	for _, want := range []string{
-		"| propose | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |",
-		"| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md`, `${CLAUDE_SKILL_DIR}/modules/tdd.md` |",
-		"| archive | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |",
-		"| delegate | `${CLAUDE_SKILL_DIR}/modules/delegation.md`, `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |",
-		"| session start and close | `${CLAUDE_SKILL_DIR}/modules/engram-opsx.md` |",
+		"| propose | " + p("artifacts") + " |",
+		p("tdd") + " when Strict TDD is on and a test runner exists",
+		"| delegate (any action) | — | " + p("delegation") + "; " + p("model-assignments") + " |",
+		"| session start and close | — | " + p("engram-opsx") + " |",
+		p("task-tool-selection"),
 	} {
 		if !strings.Contains(all, want) {
-			t.Errorf("table missing row %q:\n%s", want, all)
+			t.Errorf("table missing %q:\n%s", want, all)
 		}
 	}
 	none := tableOf(t, nil)
@@ -263,8 +272,10 @@ func TestComposeLayers_ActionTableFollowsToggles(t *testing.T) {
 			t.Errorf("no toggles: table must not list %q:\n%s", bad, none)
 		}
 	}
-	if !strings.Contains(none, "| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |") {
-		t.Errorf("no toggles: apply keeps artifacts only:\n%s", none)
+	for _, id := range config.TaskContextModules {
+		if !strings.Contains(none, p(id)) {
+			t.Errorf("no toggles: module %s has no toggle and must be listed:\n%s", id, none)
+		}
 	}
 	for toggle, absent := range map[string][]string{
 		"tdd":           {"delegation.md", "model-assignments.md", "engram-opsx.md"},
@@ -279,7 +290,7 @@ func TestComposeLayers_ActionTableFollowsToggles(t *testing.T) {
 			}
 		}
 	}
-	if tb := tableOf(t, []string{"model-routing"}); !strings.Contains(tb, "| delegate | `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |") {
+	if tb := tableOf(t, []string{"model-routing"}); !strings.Contains(tb, "| delegate (any action) | — | "+p("model-assignments")+" |") {
 		t.Errorf("model-routing alone keeps a delegate row with model-assignments:\n%s", tb)
 	}
 }
@@ -353,7 +364,9 @@ func TestComposeLayers_UniversalTellsToReinvokeAfterCompaction(t *testing.T) {
 func TestComposeLayers_SkillCarriesMissingModuleAndNoDelegationRules(t *testing.T) {
 	body := mustLayers(t, allToggles).Skill.Body
 	for _, want := range []string{
-		"If a module listed in the table is missing or empty, do not run that action: name the module and its path, tell the user to reinstall with `active-stack install`, and continue with the actions that do not depend on it.",
+		"- Missing mandatory module: stop only the action that needs it and write\n  `missing module: <id> — <action> stopped`.",
+		"- Never replace a missing module with your own version of its rules.",
+		"Compaction removes this skill and its modules from your context.",
 		"NEVER delegate the coordinator role to a sub-agent or ask one to act as `sdd-orchestrator`.",
 	} {
 		if !strings.Contains(body, want) {
@@ -364,7 +377,7 @@ func TestComposeLayers_SkillCarriesMissingModuleAndNoDelegationRules(t *testing.
 	if err != nil {
 		t.Fatal(err)
 	}
-	if strings.Contains(legacy, "do not run that action") || strings.Contains(legacy, "NEVER delegate the coordinator role") {
+	if strings.Contains(legacy, "missing module: <id>") || strings.Contains(legacy, "NEVER delegate the coordinator role") {
 		t.Error("the legacy composition must not receive the new rules")
 	}
 }
diff --git a/internal/harness/config/taskcontext_layers_internal_test.go b/internal/harness/config/taskcontext_layers_internal_test.go
new file mode 100644
index 0000000..4c44f71
--- /dev/null
+++ b/internal/harness/config/taskcontext_layers_internal_test.go
@@ -0,0 +1,21 @@
+package config
+
+import (
+	"strings"
+	"testing"
+)
+
+// R2 (6.5): un módulo anotado que ComposeLayers no cablea no se descarta en
+// silencio: falla y nombra el fragmento.
+func TestCheckWiredModules(t *testing.T) {
+	ok := map[string]string{"module/artifacts": "x", "module/engram-opsx-git-sync": "y", "universal/rules": "z"}
+	if err := checkWiredModules(ok); err != nil {
+		t.Fatalf("fragmentos cableados no deben fallar: %v", err)
+	}
+	for _, extra := range []string{"module/sixth-module", "module/directed-code-reading"} {
+		err := checkWiredModules(map[string]string{"module/artifacts": "x", extra: "t"})
+		if err == nil || !strings.Contains(err.Error(), extra) {
+			t.Errorf("%s: want an error naming the fragment, got %v", extra, err)
+		}
+	}
+}
diff --git a/internal/harness/config/taskcontext_layers_test.go b/internal/harness/config/taskcontext_layers_test.go
new file mode 100644
index 0000000..501de3b
--- /dev/null
+++ b/internal/harness/config/taskcontext_layers_test.go
@@ -0,0 +1,106 @@
+package config_test
+
+import (
+	"regexp"
+	"strings"
+	"testing"
+
+	"github.com/Group-Active-IA/active-stack/internal/catalog"
+	"github.com/Group-Active-IA/active-stack/internal/harness/config"
+)
+
+func catalogTC(t *testing.T) *catalog.TaskContext {
+	t.Helper()
+	c, err := catalog.Load()
+	if err != nil {
+		t.Fatal(err)
+	}
+	return c.TaskContext
+}
+
+// 6.5: modules/ trae los cuatro archivos sin frontmatter aunque no tengan toggle.
+func TestTaskContext_ModulesDeliveredWithoutToggle(t *testing.T) {
+	for _, toggles := range [][]string{nil, allToggles} {
+		l := mustLayers(t, toggles)
+		for _, id := range config.TaskContextModules {
+			text := l.Modules[id]
+			if strings.TrimSpace(text) == "" {
+				t.Fatalf("toggles %v: modules/%s.md missing or empty", toggles, id)
+			}
+			if strings.HasPrefix(text, "---") {
+				t.Errorf("modules/%s.md carries skill frontmatter", id)
+			}
+			if l.Skill.Modules[id] != text {
+				t.Errorf("Skill.Modules[%s] differs from Layers.Modules", id)
+			}
+		}
+	}
+}
+
+// 6.5: la tabla del SKILL.md es byte a byte la de 6.2 (con todos los toggles) y
+// un módulo de toggle inactivo no figura.
+func TestTaskContext_SkillTableIsTheCatalogTable(t *testing.T) {
+	want, err := catalogTC(t).RenderActionTable("claude")
+	if err != nil {
+		t.Fatal(err)
+	}
+	if got := tableOf(t, allToggles); !strings.Contains(got, strings.TrimRight(want, "\n")) {
+		t.Errorf("the SKILL.md does not contain the catalog table byte for byte:\n--- want\n%s\n--- got\n%s", want, got)
+	}
+	none := tableOf(t, nil)
+	if strings.Contains(none, "modules/tdd.md") || strings.Contains(none, "modules/engram-opsx.md") {
+		t.Error("an inactive-toggle module must not be listed")
+	}
+	if strings.Contains(none, "${CLAUDE_SKILL_DIR}/modules/") == false || strings.Contains(none, "`modules/") {
+		t.Error("the Claude path form is ${CLAUDE_SKILL_DIR}/modules/<id>.md")
+	}
+}
+
+// 6.6 (Claude): la capa universal trae governance y las reglas de memoria;
+// el encargo a quien no ve CLAUDE.md las repite (D21).
+func TestTaskContext_UniversalLayerAndAssignmentRule(t *testing.T) {
+	l := mustLayers(t, allToggles)
+	for _, want := range []string{"### Governance Levels", "CRITICAL", "HIGH", "MEDIUM", "LOW", "PROACTIVE SAVE TRIGGERS", "mem_save"} {
+		if !strings.Contains(l.Universal, want) {
+			t.Errorf("universal layer missing %q", want)
+		}
+	}
+	body := config.RenderSkillMD(l.Skill)
+	for _, want := range []string{"Explore, Plan and any agent with `omitClaudeMd: true`", "repeat the mandatory rules", "governance levels"} {
+		if !strings.Contains(body, want) {
+			t.Errorf("SKILL.md assignment rule missing %q", want)
+		}
+	}
+	// El texto de las reglas no vive en el SKILL.md: solo la orden de repetirlas.
+	if strings.Contains(body, "### Governance Levels") {
+		t.Error("the governance levels belong to the universal layer, not the SKILL.md")
+	}
+	// Sin governance no hay niveles que repetir en la capa universal.
+	if strings.Contains(mustLayers(t, nil).Universal, "### Governance Levels") {
+		t.Error("governance toggle off: no levels in universal")
+	}
+}
+
+// 6.7 (Claude): la referencia de invocación es /sdd-orchestrator, sin la
+// sintaxis de Codex ni `claude --agent`, en todo lo compuesto.
+func TestTaskContext_ClaudeInvocationReference(t *testing.T) {
+	l := mustLayers(t, allToggles)
+	if !strings.Contains(l.Universal, "invoke `/sdd-orchestrator`") {
+		t.Error("universal must name /sdd-orchestrator")
+	}
+	md := config.RenderSkillMD(l.Skill)
+	if !regexp.MustCompile(`(?m)^description: .*/sdd-orchestrator`).MatchString(md) {
+		t.Error("SKILL.md description must name /sdd-orchestrator")
+	}
+	all := map[string]string{"universal": l.Universal, "skill": md}
+	for n, m := range l.Modules {
+		all["module/"+n] = m
+	}
+	for name, text := range all {
+		for _, bad := range []string{"claude --agent", "--agent sdd-orchestrator", "$sdd-orchestrator"} {
+			if strings.Contains(text, bad) {
+				t.Errorf("%s contains %q", name, bad)
+			}
+		}
+	}
+}
diff --git a/internal/harness/config/testdata/compose_layers_module_bounded-command-output.golden b/internal/harness/config/testdata/compose_layers_module_bounded-command-output.golden
new file mode 100644
index 0000000..1f0d6f1
--- /dev/null
+++ b/internal/harness/config/testdata/compose_layers_module_bounded-command-output.golden
@@ -0,0 +1,98 @@
+## Bounded Command Output
+
+Clients cut long command output before you see it. Claude Code keeps about 10,000 characters of
+a failed command, head and tail, and that tail is not the real end of the output. Codex keeps
+head and tail and does not tell you the middle is gone. Never summarize from that view.
+
+### When to summarize
+
+- Output above 300 lines or above 10,000 characters, whichever comes first, is delivered as a
+  summary plus a reference. Shorter output is delivered whole.
+- The threshold decides when to summarize, never how much to cut. Never drop a failure line to
+  save space.
+
+### Capture before you read
+
+- Capture every verification command (tests, build, vet, lint) and every command known to be
+  verbose, before you run it. Do not guess its size first.
+- If this skill names an independent capture source for your client and surface, use it and name
+  it. Otherwise capture the output yourself with your shell's recipe; the capture source is then
+  `agent`.
+- If an independent capture already wrote a redacted copy and its capture record, summarize from
+  that copy and cite it in `ref:`. Do not write a copy of your own.
+- bash / sh:
+  `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
+- PowerShell 5.1 and 7 (keeps the command's bytes; avoids UTF-16 and `NativeCommandError`):
+  `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
+- Put `$f` under `<user cache dir>/active-stack/runs/<change>/<run-id>/`, or under
+  `_sin-change` when no change is active. Never inside the repository. The copy expires seven
+  days after the change is archived, or seven days after capture under `_sin-change`.
+- Redact before you keep it. Replace every value that matches the sensitive list (`.env`
+  assignments, tokens and API keys, cloud credentials, SSH keys and `*.pem`, environment variable
+  values, home directory paths, account ids, emails and account UUIDs) with an omission mark,
+  keeping the line in its place. Write the redacted copy to a second file, check that no listed
+  value is left in it, and delete the raw file before you report. The copy you keep and cite is
+  the redacted one.
+- Never write a raw sensitive value: not in a file you keep, not in the summary, not in your
+  reply, even when asked for the original. Quote the redacted line and say the value was omitted.
+- Read the captured copy with directed reads (failure patterns, exit code). Do not summarize
+  from what the client showed you.
+- A command that already ran without capture and turned out verbose: if it has no side effects,
+  run it again with capture; if it has side effects, do not repeat it and mark the summary
+  partial.
+
+### What the summary keeps
+
+- Every error, every failed test and the exit code. A non-zero exit code is always a failure.
+  One `failure:` line per captured line that matches a pattern, including the final
+  `FAIL\t<package>` line.
+- Failure patterns: `go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` and `go vet`:
+  `<file>:<line>:<column>:`. For a tool with no declared patterns, write
+  `failures identified by exit code only`.
+- Only failures the output records. Lines that only look like errors are not failures. With no
+  failure, write exactly `failure: none`, with nothing after it.
+- Each `failure:` line is the captured line verbatim, leading spaces included. Append nothing:
+  no line number, no note.
+- Repetition without failures is dropped.
+
+### Fixed lines
+
+The harness counts these lines. Write them exactly, one per line, at the top of the summary file
+and in your reply, whatever the shell. Prose may follow them; it never replaces them.
+
+```text
+exit: <code>
+failure: <line as captured>
+ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
+capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
+status: unverified — <what is missing>
+partial — unverified: <reason>
+```
+
+Example, a failing test run captured by the agent:
+
+```text
+exit: 1
+failure: --- FAIL: TestParse (0.00s)
+ref: cmd=go test ./pkg/ run=20261113-1 path=<cache>/active-stack/runs/_sin-change/20261113-1/out.redacted.txt sha256:<hex> expires=2026-11-20
+capture: agent · redactor: agent · producer: sdd-orchestrator
+status: unverified — no independent capture
+```
+
+- No captured original: write `partial — unverified: <reason>` and no `ref:` line (no path, no
+  hash, not even `ref: none`). Say which case it is: `never captured`, `copy missing` (the
+  record names a file that is not there), `expired on <date>`, or `no persistent disk` (surface
+  without the user's disk).
+- A result that you or a worker declare stays `unverified` until another party corroborates
+  the execution and the code state with independent capture or a re-run. A matching hash proves
+  the copy is intact; it does not prove who ran the command.
+- When you verify a declared result, give one state. `rejected`: the summary omits a failure the
+  copy records, declares one it does not record, or its hash or exit code differs from the copy
+  or the capture record. `verified`: none of that, and the copy comes from an independent capture
+  (record source `arnés` or a client source, with coverage `demostrada`) or another party re-ran
+  the command with capture. Otherwise `unverified`.
+- If verifying needs a detail that redaction removed, the result stays `unverified`; say which
+  detail is missing.
+- Never cite the client's own copy (session transcript, client output file) as the reference:
+  it is not redacted and Active Stack does not manage it. When asked for an original, or when
+  that copy holds a sensitive value, name it as a limit.
diff --git a/internal/harness/config/testdata/compose_layers_module_concise-agent-reporting.golden b/internal/harness/config/testdata/compose_layers_module_concise-agent-reporting.golden
new file mode 100644
index 0000000..32dd516
--- /dev/null
+++ b/internal/harness/config/testdata/compose_layers_module_concise-agent-reporting.golden
@@ -0,0 +1,24 @@
+## Concise Agent Reporting
+
+Unless the user asks for more, a report is exactly two lines:
+
+- Line 1: the result, with its provenance (`path:line` or `path#symbol`).
+- Line 2: the consequence or the next step.
+- Nothing after them: no method, no files read, no modules applied, nothing about what you did not
+  do, no side observations, no memory notes. The only exception is a fixed line of another module
+  (`full read:`, `unnecessary call:`, `ref:`, `partial — unverified:`, `unavailable:`,
+  `out-of-class use:`), written exactly as that module says.
+
+```text
+parseConfig ignores an empty path: internal/config/load.go:42
+Next: return an error for the empty path before reading the file.
+```
+
+- Several findings: a list, each with its provenance, before any explanation; then one line with
+  the next step.
+- No finding: one line that says nothing was found and what you checked. Do not narrate the
+  search.
+- When asked for the detail, produce it then, from the same provenance.
+- Do not repeat what earlier steps of the same task already reported.
+- If the content is already in a project artifact, name the artifact and the section instead of
+  copying it.
diff --git a/internal/harness/config/testdata/compose_layers_module_directed-code-reading.golden b/internal/harness/config/testdata/compose_layers_module_directed-code-reading.golden
new file mode 100644
index 0000000..b470602
--- /dev/null
+++ b/internal/harness/config/testdata/compose_layers_module_directed-code-reading.golden
@@ -0,0 +1,26 @@
+## Directed Code Reading
+
+Read only what the task needs, and say where every claim about code comes from.
+
+- Read a file whole only when one of these five conditions holds:
+  1. the file has 300 lines or fewer;
+  2. the task will rewrite the whole file;
+  3. the file is the artifact the active action produces or edits;
+  4. a bounded read of that file came back empty;
+  5. no range covers what the task needs. Then write in your report:
+     `full read: <path> — <reason>`
+  Any other full read is unjustified.
+- Otherwise read by path and line range, or by symbol. A symbol read returns the whole symbol,
+  from its first line to its last, plus the signature of the declaration that contains it when it
+  is nested. Nothing more: not the rest of the enclosing body, not the neighbouring symbols.
+- To learn where a symbol is used, search and keep each match as `path:line`. Do not open every
+  file that matches.
+- To learn where a capability lives, run one bounded search over the repository. List each
+  candidate path with its match type and count. A definition ranks above any number of
+  references, and a reference above any number of plain-text matches; the count orders paths
+  only within the same type. Do not open the files the search discards. Never walk the
+  repository file by file to orient yourself.
+- Every claim about code carries its provenance: `path:line` or `path#symbol`. A claim without
+  provenance is unverified and does not count as evidence for accepting a change.
+- If you read the same file and the same range again with no write to that file in between, do
+  it if you need to, and write: `unnecessary call: <tool> <input>`
diff --git a/internal/harness/config/testdata/compose_layers_module_task-tool-selection.golden b/internal/harness/config/testdata/compose_layers_module_task-tool-selection.golden
new file mode 100644
index 0000000..16f9faf
--- /dev/null
+++ b/internal/harness/config/testdata/compose_layers_module_task-tool-selection.golden
@@ -0,0 +1,24 @@
+## Task Tool Selection
+
+Use the tools the active action needs and no optional integration it does not enable.
+
+| Action | Base tools | Optional integrations enabled |
+|---|---|---|
+| explore | directed reading, search | Context7 |
+| propose | directed reading, OpenSpec CLI | Context7 |
+| apply | editing, tests | Context7; domain skills from the registry; Playwright only if the action touches UI |
+| verify | tests, directed reading | Playwright only if there is UI |
+| archive | OpenSpec CLI | none |
+
+- An optional integration is off unless the active action enables it. Being installed does not
+  enable it. Design integrations are enabled only when the `design-mockups` harness is active.
+- When you delegate, pass the specialist the enabled list for its action in the assignment.
+- If an enabled integration is unavailable, write
+  `unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`
+  Do not present the result as equivalent to one obtained with the integration. Do not call it
+  again in the same task.
+- If you call a tool or integration again with the same input and the state it observes has not
+  changed, write `unnecessary call: <tool> <input>`
+- If you use an integration the active action does not enable and the client did not block it,
+  write `out-of-class use: <integration> in <action>`. Never describe a rule that is only
+  recommended as a block.
diff --git a/internal/harness/config/testdata/compose_layers_skill.golden b/internal/harness/config/testdata/compose_layers_skill.golden
index c00af67..886313b 100644
--- a/internal/harness/config/testdata/compose_layers_skill.golden
+++ b/internal/harness/config/testdata/compose_layers_skill.golden
@@ -24,17 +24,43 @@ OPSX replaces the legacy SDD phase system. There are no rigid phase gates. The u
 
 ## Action to Modules
 
-Before an action, read ONLY the modules listed for it. When you delegate, paste their content into the sub-agent prompt: the sub-agent does not see this skill.
-
-| Action | Modules to read |
-|--------|-----------------|
-| propose | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |
-| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md`, `${CLAUDE_SKILL_DIR}/modules/tdd.md` |
-| archive | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |
-| delegate | `${CLAUDE_SKILL_DIR}/modules/delegation.md`, `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |
-| session start and close | `${CLAUDE_SKILL_DIR}/modules/engram-opsx.md` |
-
-If a module listed in the table is missing or empty, do not run that action: name the module and its path, tell the user to reinstall with `active-stack install`, and continue with the actions that do not depend on it.
+Before an action, read the modules listed for it. Read every mandatory module. Read an optional
+module when its condition applies. When you delegate, paste the modules the specialist needs
+into the assignment: the specialist does not see this skill.
+
+| Action | Mandatory | Optional |
+|--------|-----------|----------|
+| explore | — | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md` if you run commands; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
+| propose | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
+| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md`; `${CLAUDE_SKILL_DIR}/modules/tdd.md` when Strict TDD is on and a test runner exists | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
+| verify | `${CLAUDE_SKILL_DIR}/modules/delegation.md`; `${CLAUDE_SKILL_DIR}/modules/model-assignments.md`; `${CLAUDE_SKILL_DIR}/modules/bounded-command-output.md` | `${CLAUDE_SKILL_DIR}/modules/directed-code-reading.md`; `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
+| archive | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` | `${CLAUDE_SKILL_DIR}/modules/task-tool-selection.md`; `${CLAUDE_SKILL_DIR}/modules/concise-agent-reporting.md` |
+| delegate (any action) | — | `${CLAUDE_SKILL_DIR}/modules/delegation.md`; `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |
+| session start and close | — | `${CLAUDE_SKILL_DIR}/modules/engram-opsx.md` |
+
+### Missing or empty modules
+
+A module file that does not exist, or that exists with no text (zero bytes or only
+whitespace), is missing.
+
+- Missing mandatory module: stop only the action that needs it and write
+  `missing module: <id> — <action> stopped`. Finish the parts of the request that do not need
+  it and report them apart from the stopped action. Resume when the module is available or a
+  surface that delivers it is used, and say which one.
+- Missing optional module: write `missing optional module: <id> — continuing without it` and
+  resolve the action directly. Without `delegation`, do not delegate.
+- On a surface without subagents, `verify` stops: write
+  `missing module: delegation — verify stopped (surface without subagents)`.
+- Never replace a missing module with your own version of its rules.
+
+### After compaction or resume
+
+Compaction removes this skill and its modules from your context. Invoke the skill again and
+re-read the modules of the current action before you continue it.
+
+### Assignments to agents without CLAUDE.md
+
+Explore, Plan and any agent with `omitClaudeMd: true` start without CLAUDE.md, so the universal block does not reach them. In their assignment, repeat the mandatory rules of that block and its governance levels.
 
 NEVER delegate the coordinator role to a sub-agent or ask one to act as `sdd-orchestrator`.
 
diff --git a/internal/verify/harness_checks.go b/internal/verify/harness_checks.go
index 3f2dbe6..640c3ff 100644
--- a/internal/verify/harness_checks.go
+++ b/internal/verify/harness_checks.go
@@ -197,9 +197,13 @@ func checkManagedSkillConfig(h model.Harness, adapter Adapter, homeDir string) [
 				}
 				for _, module := range config.ExpectedModules(h.Toggles) {
 					p := filepath.Join(skillDir, "modules", module+".md")
-					if _, err := os.Stat(p); err != nil {
+					raw, err := os.ReadFile(p)
+					if err != nil {
 						return fmt.Errorf("module %q missing at \"%s\"", module, p)
 					}
+					if strings.TrimSpace(string(raw)) == "" {
+						return fmt.Errorf("module %q is empty at \"%s\" (an empty module counts as missing)", module, p)
+					}
 				}
 				content, err := readInstr()
 				if err != nil {
diff --git a/internal/verify/taskcontext_modules_test.go b/internal/verify/taskcontext_modules_test.go
new file mode 100644
index 0000000..07df524
--- /dev/null
+++ b/internal/verify/taskcontext_modules_test.go
@@ -0,0 +1,43 @@
+package verify_test
+
+import (
+	"strings"
+	"testing"
+
+	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
+	"github.com/Group-Active-IA/active-stack/internal/harness/config"
+)
+
+// Tarea 6.5 (opsx-task-context, R3/R17): verify espera los cuatro módulos de la
+// 05 y trata un módulo vacío (0 bytes o solo espacios) como ausente.
+func TestManagedVerify_TaskContextModules(t *testing.T) {
+	for _, id := range config.TaskContextModules {
+		for name, content := range map[string]string{"empty": "", "whitespace": " \n\t\r\n", "missing": "\x00remove"} {
+			t.Run(id+"/"+name, func(t *testing.T) {
+				f := installManaged(t)
+				p := "modules/" + id + ".md"
+				if content == "\x00remove" {
+					mustRemove(t, f.skillFile(p))
+				} else {
+					mustWrite(t, f.skillFile(p), content)
+				}
+				failed, _ := statuses(f.verify(t, claude.VersionCompatible))
+				if len(failed) == 0 {
+					t.Fatalf("%s %s must fail verification", id, name)
+				}
+				if !strings.Contains(strings.Join(failed, "\n"), id) {
+					t.Errorf("the failure must name the module %q: %v", id, failed)
+				}
+			})
+		}
+	}
+}
+
+func TestExpectedModules_IncludesTaskContextWithoutToggles(t *testing.T) {
+	got := strings.Join(config.ExpectedModules(nil), ",")
+	for _, id := range config.TaskContextModules {
+		if !strings.Contains(got, id) {
+			t.Errorf("ExpectedModules(nil) = %s, missing %s", got, id)
+		}
+	}
+}
```
