# Entrega de la 05 a 03 y 04

> Ficha 05 · `opsx-task-context` · 2026-11-13.
> **Estado: texto listo para integrar; la integración en código queda para la ronda de apply.**
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
inglés (`sdd-orchestrator.md`, `SKILL.md` de las dos ramas). El resto de este archivo está en
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

**Cómo se integra.** Hoy hay dos compositores (03 `ComposeLayers`, 04 `ComposeCodexPackage`) y
ninguno está en `main` (§5, R5–R6). La 05 propone uno común con un registro de módulos
(`design.md` — D23). Mientras no exista:

- **En la 03:** agregar cada módulo de la §2 como archivo `modules/<id>.md` de la skill **sin
  depender de un toggle**: con la anotación `<!-- as:module/<id> -->` sola, un módulo sin toggle
  se descarta en silencio (R2). Reemplazar `actionTable`, `exists` y `ExpectedModules` por la
  tabla de la §3, que suma `explore` y `verify` (R1).
- **En la 04:** agregar los mismos archivos a `Files` y reemplazar el string `actions` por la
  tabla de la §3 con la ruta relativa de Codex. La línea *"Consume context and capture modules
  from change 05 when available, and identify their versions"* (`SKILL.md:20`) se cumple citando
  la revisión de este archivo.
- **En las dos:** el verify de la instalación falla si un módulo de la tabla falta **o está
  vacío**. El de la 04 ya compara byte a byte (R4); el de la 03 deja pasar un módulo vacío (R3).

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
  it if you need to, and write: `unnecessary call: <tool> <input>`
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
- bash / sh:
  `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"`
- PowerShell 5.1 and 7 (keeps the command's bytes; avoids UTF-16 and `NativeCommandError`):
  `cmd /c "<command> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" | Set-Content -Encoding ascii "$f.exit"`
- Put `$f` under `<user cache dir>/active-stack/runs/<change>/<run-id>/`, or under
  `_sin-change` when no change is active. Never inside the repository.
- Redact before you keep it. Replace every value that matches the sensitive list (`.env`
  assignments, tokens and API keys, cloud credentials, SSH keys and `*.pem`, environment variable
  values, home directory paths, account ids, emails and account UUIDs) with an omission mark,
  keeping the line in its place. Write the redacted copy to a second file and delete the raw file
  before you report.
- Read the captured copy with directed reads (failure patterns, exit code). Do not summarize
  from what the client showed you.
- A command that already ran without capture and turned out verbose: if it has no side effects,
  run it again with capture; if it has side effects, do not repeat it and mark the summary
  partial.

### What the summary keeps

- Every error, every failed test and the exit code. A non-zero exit code is always a failure.
- Failure patterns: `go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` and `go vet`:
  `<file>:<line>:<column>:`. For a tool with no declared patterns, write
  `failures identified by exit code only`.
- Only failures the output records. Lines that only look like errors are not failures.
- Repetition without failures is dropped.

### Fixed lines

The harness counts these lines. Write them exactly:

```text
exit: <code>
failure: <line as captured>
ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>
capture: <harness | client:<source> | agent> · redactor: <same set> · producer: <who wrote this summary>
status: unverified — <what is missing>
partial — unverified: <reason>
```

- No captured original: write `partial — unverified: <reason>`, with no path and no hash. Say
  which case it is: `never captured`, `copy missing` (the record names a file that is not there),
  `expired on <date>`, or `no persistent disk` (surface without the user's disk).
- A result that you or a worker declare stays `unverified` until another party corroborates
  the execution and the code state with independent capture or a re-run. A matching hash proves
  the copy is intact; it does not prove who ran the command.
- If verifying needs a detail that redaction removed, the result stays `unverified`; say which
  detail is missing.
- Never cite the client's own copy (session transcript, client output file) as the reference:
  it is not redacted. Name it as a limit when it holds a sensitive value.
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
- If an enabled integration is unavailable, write
  `unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`
  Do not present the result as equivalent to one obtained with the integration. Do not call it
  again in the same task.
- If you call a tool or integration again with the same input and the state it observes has not
  changed, write `unnecessary call: <tool> <input>`
- If you use an integration the active action does not enable and the client did not block it,
  write `out-of-class use: <integration> in <action>`. Never describe a rule that is only
  recommended as a block.
````

### `modules/concise-agent-reporting.md`

````markdown
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
````

## 3. Tabla acción → módulos

La decisión y su justificación están en `design.md` — D19; esta es la única copia de la tabla.
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

Before an action, read the modules listed for it. Read every mandatory module. Read an optional
module when its condition applies. When you delegate, paste the modules the specialist needs
into the assignment: the specialist does not see this skill.

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
  surface that delivers it is used, and say which one.
- Missing optional module: write `missing optional module: <id> — continuing without it` and
  resolve the action directly. Without `delegation`, do not delegate.
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
| Codex | `modules/<id>.md`, relativa al directorio de la skill (forma actual de la 04) | no probado: la política de shell de Windows bloqueó la lectura (X12) |

**Por qué la regla de módulo vacío.** Con un módulo vacío Claude se comportó distinto en tres
corridas (C13), y el verify de la 03 lo deja pasar (R3). Tratar vacío como ausente da una sola
conducta y la vuelve verificable.

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
en un directorio temporal aislado. No se versionan transcripciones, sesiones ni logs: cada fila
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
| C13 | Módulo obligatorio vacío | `modules/artifacts.md` de 0 B y pedir propose, tres veces | planifica sin leerlo; se desvía; lo lee, lo informa vacío y se detiene | medido: inconsistente |
| C14 | Modo legado | install de la PR #8 con un `claude` que reporta 2.1.109 | `degraded`: coordinador como bloque único en `CLAUDE.md`, sin skill ni módulos | medido, con versión simulada |

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

### PR #8 (03) y PR #9 (04)

| Id | Afirmación | Comando | Resultado | Veredicto |
|---|---|---|---|---|
| R1 | La tabla de la 03 cubre las cinco acciones | sonda que llama `ComposeLayers` con los cinco toggles | `SKILL.md` 7.804 B, universal 4.305 B, módulos 8.916 B; filas propose, apply, archive, delegate e inicio y cierre; **sin explore ni verify** | medido: no las cubre |
| R2 | Un sexto módulo anotado aparece | anotar `<!-- as:module/<id> -->` sin toggle y componer | salen los mismos cinco módulos, sin error | medido: se descarta en silencio |
| R3 | El verify de la 03 detecta un módulo vacío | sonda de verify con módulo íntegro, vacío y borrado | íntegro y vacío pasan; borrado falla | medido: el vacío pasa |
| R4 | El verify de la 04 detecta un módulo vacío | lectura de `internal/verify/harness_checks.go` en la PR #9 | compara byte a byte con lo esperado | documentado (código de la rama) |
| R5 | 03 y 04 se integran sin conflicto | `git merge-tree --write-tree origin/claude-native-opsx origin/propose/codex-native-opsx` | conflicto en 9 archivos: `install.go`, `model/harness.go`, `install/plan.go`, `install/steps.go`, `uninstall/steps.go`, `verify/harness_checks.go`, `AGENTS.md`, `ARCHITECTURE.md`, `CLAUDE.md` | medido: no |
| R6 | Hay un solo compositor | lectura de `compose_layers.go:235-311` (03) y `codex_package.go:20-107` (04) | 03: `ComposeLayers` con anotaciones y tabla en `actionTable`, `exists` y `ExpectedModules`. 04: `ComposeCodexPackage` con `splitCodexAsset` y la tabla como string. Las dos usan `ConfigDeliveryManagedSkill` con semánticas distintas | documentado: hay dos |
| R7 | Tests de Go | `go test ./...` en Windows y en WSL sobre cada rama | PR #9 verde en WSL; PR #8 en WSL falla lo mismo que `main`; sus 8 fallas adicionales son solo de Windows | medido |

## 6. No probado y por qué

| Qué | Por qué |
|---|---|
| Lectura real de módulos por Codex | la política de shell de Codex en Windows rechaza los comandos sin bypass (X12) |
| Módulo ausente y módulo vacío en Codex | mismo bloqueo: el coordinador no puede leer el directorio de la skill |
| Si un subagente de Codex lee `AGENTS.md` | lanzar y observar subagentes necesita shell (mismo bloqueo) |
| Descubrimiento en `%USERPROFILE%\.agents\skills` con un HOME temporal | Codex ignora `HOME` y `USERPROFILE` para esa ruta; probarlo exigía escribir en el perfil real |
| `/skills` en Codex sin interfaz | `codex exec` no lo expone |
| IDE y app de escritorio, en los dos clientes; Codex en la nube | sin extensión ni app operables desde la sesión de prueba; la 04 no promete escritorio ni nube (`codex-native-opsx/design.md:59-63`) |
| Clientes en Linux | no hay `claude` ni `codex` nativos en WSL; ahí solo corrieron los tests de Go (R7) |
| Claude 2.1.109 real | se simuló la versión ante el instalador (C14) |
| Hooks `PostToolUse` y `PostToolUseFailure` de Claude como fuente | no se instalaron hooks en esta ronda |
| Conteo de tokens | sin API key no hay contador oficial; las cifras son caracteres y bytes |
| Efecto de `tool_output_token_limit` | X5 sin diferencia observable |
| Mezcla real de 03 y 04 y si quedan anotaciones `<!-- as:` en la salida de Codex | la sonda aplicó los assets anotados de la 03 sin su `assets.go`, que es donde se quitan; requiere el merge real |
| Módulo vacío en Claude como conducta estable | tres corridas, tres conductas (C13): muestra el riesgo, no lo mide |
| Representatividad | n = 1 por prueba, un modelo (haiku en Claude) |

## 7. Pendientes por ficha

| Ficha | Qué falta en código | Contra qué de este archivo |
|---|---|---|
| **03** (Juani) | Con la 04, compositor común con registro de módulos (`design.md` — D23). Mientras tanto: agregar los cuatro módulos sin toggle (R2), reemplazar `actionTable`/`exists`/`ExpectedModules` por la §3 con `explore` y `verify` (R1), y que verify falle con un módulo vacío (R3). Tarea 2.3: recibe los textos (§2) | §2, §3 |
| **04** (Matías según el informe unificado) | Mismo compositor común. Agregar los cuatro módulos a `Files`, reemplazar el string `actions` por la §3 con la ruta de Codex, y fijar la revisión de este archivo (tarea 1.4; `SKILL.md:20`). Pendiente de su tarea 2.5: `verify` y el contrato de evidencia por acción. Probar lectura de módulos, módulo ausente y vacío y subagentes cuando la política de shell lo permita (§6) | §2, §3, §6 |
| **06** (Gianfranco) | Confirmar o contraproponer D9 y D14; que su contrato de evidencia se cargue con `bounded-command-output` en `apply` y `verify` | §4 |
| **02** (Giuliano) | Usar la lista de D14 en `eval/baseline/record` y el plazo de D9 en su D6; que el arnés de la 05 reutilice su manifiesto | §4 |
| **05** | Pasar los textos de la §2 a `internal/harness/config/assets/modules/` con sus tests (tareas 4.4, 5.3, 6.3 y 7.2), declarar la tabla en el catálogo y generarla (6.1 y 6.2), y el adaptador de `codex exec --json` en el arnés (3.6) | §2, §3, X7 |
