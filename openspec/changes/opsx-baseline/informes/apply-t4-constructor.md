# apply-t4 · constructor · 1.2, 2.2 y 2.3 desde documentación oficial

Rama `propose/opsx-baseline` en `3f8afd2` (verificada con `git branch --show-current` al empezar y al terminar). Sin commits, sin push, sin `go build`, sin procesos en segundo plano, sin ejecutar `claude` ni `codex`. Fecha de trabajo y de consulta de la documentación: 2026-11-11.

## Resultado

| Tarea | Estado | Motivo |
|---|---|---|
| 1.2 | **abierta** `[ ]` | Perfiles, validación y comparación hechos. Falta la verificación con el cliente real en Windows y Linux (sub-viñeta "verificación real pendiente" en `tasks.md`). |
| 2.2 | **cerrada** `[x]` | Adaptadores de Claude y Codex con muestras derivadas de la documentación (riesgo declarado, lo que el texto admite), casos completos, parciales y duplicados, una corrida delegada por cliente y la precedencia reproduce 60.000 y 48.000. Lectura de "terminal e IDE": ver JUICIO 1. |
| 2.3 | **abierta** `[ ]` | Cubierto en código y tests: `agent_id` desde 2.1.310, `agent.name = custom` sin atribuir, `total_tokens` del cierre sin tomar como consumo, tabla de fuentes y límites. No cubierto: "comprobar en Codex si el agregado del padre incluye a los hilos hijos" (la documentación no lo dice) ni la comprobación con cliente real. |

## EVIDENCIA

### Comandos y resultados

| Comando | Cuándo | Revisión | Salida |
|---|---|---|---|
| `go test -count=1 ./eval/baseline/{manifest,usage,trace,run,record}/` (safety net) | 19:19:29 | `3f8afd2` | los cinco `ok` |
| `go test -count=1 ./eval/baseline/profile/` (RED, esqueleto) | 19:21:28 | `3f8afd2` + esqueleto | `len(Profiles()) = 0, want 8 (2 clientes x 2 interfaces x 2 tipos)`; `falta perfil claude/terminal/equivalente_para_edicion` |
| ídem (GREEN) | 19:22:32 | + `profile.go` | `ok`, cobertura 91,7 % (94,4 % al final) |
| `go test -count=1 -run VersionAtLeast ./eval/baseline/manifest/` (RED) | 19:25:45 | esqueleto | `VersionAtLeast("2.1.310", "2.1.310") = (false, <nil>), want (true, nil)` y 5 `err = nil, want error` |
| ídem (GREEN) | 19:25:53 | | `ok`, cobertura de `manifest` 97,3 % |
| `go test -count=1 ./eval/baseline/capture/` (RED Claude) | 19:28:59 | esqueleto | `solicitudes = 0, want 4`; `por rol = map[], want principal 40000 y worker 20000`; `Totals = {Total:0 ...}, want total 60000`; `no es JSON: err = nil, want error`; un `panic` por índice en un test sobre el esqueleto (artefacto del esqueleto, corregido con una guarda en el test) |
| ídem (GREEN Claude) | 19:30:35 | | `ok` |
| ídem (RED Codex) | 19:31:23 | esqueleto | `thread_id = ""`, `turnos del padre = 0, want 2`, `Totals = {Total:0 ...}, want 48000`, `CodexParentAggregateScope = "", want "no_comprobado"`; `nil pointer` en la traza sobre el esqueleto |
| ídem (GREEN Codex) | 19:32:36 | | `ok`, cobertura 95,6 % |
| `go test -count=1 -skip 'TestAddToUserPath' ./...` | 19:36:17–19:38:48 | `3f8afd2` + cambios sin commitear | todos `ok`, ninguna línea `FAIL` (`fixture` 132.1 s) |
| `gofmt -l eval` | tras lo anterior | | sin salida |
| `go vet ./eval/...` | | | `vet-ok` |
| `go list -deps ./cmd/... ./internal/...` filtrado por `/eval` | | | 0 coincidencias; `grep -rln "baseline/" cmd internal` sin resultados |
| `git diff --stat -- go.mod go.sum` | | | vacío (solo biblioteca estándar) |
| `openspec validate opsx-baseline --strict` | | | `Change 'opsx-baseline' is valid` |
| `git diff --stat -- design.md proposal.md` | | | vacío |

`go test ./...` completo se corrió con `-skip 'TestAddToUserPath'` una sola vez. `TestAddToUserPath` (`internal/system`) **no se corrió ni se arregló**: escribe entradas en el PATH real de Windows del usuario; se reporta como hallazgo del repo, no de esta tanda.

### Dos expectativas mías incorrectas en los tests (no del código)
Al primer GREEN fallaron dos tests por cuentas mías: esperaba 5 solicitudes en la muestra delegada (son 4: 2 principal + 2 worker) y 9.000 tokens sin caché en el padre de Codex (son 6.000 = 5.000 + 1.000). También una muestra repetía `tool_use_id` entre llamadas distintas y se deduplicaba mal; se regeneró con ids únicos. Se corrigieron los tests y la muestra, no el código.

### Hashes (`git hash-object`, 12 primeros caracteres, sobre los archivos entregados)

```
4adde4739d42  eval/baseline/profile/profile.go
7418106161d2  eval/baseline/profile/profile_test.go
ed25aa6a9845  eval/baseline/capture/common.go
63c434a3431c  eval/baseline/capture/claude.go
2fceb19120f5  eval/baseline/capture/claude_test.go
72a1e4e9270d  eval/baseline/capture/codex.go
25db270ea45c  eval/baseline/capture/codex_test.go
138e496de2c4  eval/baseline/manifest/manifest.go
45576c680d30  eval/baseline/manifest/apply4_test.go
d72a34a9a232  eval/baseline/README.md
33db9748d947  eval/baseline/fixtures/samples/README.md
97aebc9a8b99  eval/baseline/fixtures/samples/claude/otel-delegated.jsonl
ff3b316884ed  eval/baseline/fixtures/samples/claude/otel-partial-old-version.jsonl
5f9d85bcee83  eval/baseline/fixtures/samples/claude/otel-duplicates.jsonl
39c9d00dbeb2  eval/baseline/fixtures/samples/claude/result-delegated.json
e36e0ff9b3fa  eval/baseline/fixtures/samples/claude/result-error-max-turns.json
0bb909ca486e  eval/baseline/fixtures/samples/claude/stream-delegated.jsonl
e8e8ced9b5d9  eval/baseline/fixtures/samples/codex/exec-doc-example.jsonl
75e5af8a4910  eval/baseline/fixtures/samples/codex/exec-delegated-parent.jsonl
6ec1ae381f43  eval/baseline/fixtures/samples/codex/exec-delegated-child.jsonl
c58ce7cb5862  eval/baseline/fixtures/samples/codex/exec-partial.jsonl
1ece9728cbeb  eval/baseline/fixtures/samples/codex/exec-failed.jsonl
85aa669c9c4a  eval/baseline/fixtures/samples/codex/exec-duplicates.jsonl
```

`tasks.md` se editó después de calcular estos hashes (no figura aquí). Los únicos archivos modificados que ya existían: `eval/baseline/README.md`, `eval/baseline/manifest/manifest.go` (solo se agregó `VersionAtLeast`) y `openspec/changes/opsx-baseline/tasks.md`.

### Comportamiento → RED → GREEN → mutante

| Comportamiento | RED (aserción) | GREEN | Mutante que lo mata |
|---|---|---|---|
| 8 perfiles y `Lookup` | `len(Profiles()) = 0, want 8` | 19:22:32 | P1, P7 |
| Una edición dentro del checkout pide aprobación solo en Estricto | vacío en el esqueleto; acreditado por mutante | 19:22:32 | P1 (`acceptEdits` pide), P2 (`workspace-write` pide), P13 (`Edit(./**)` ignorado), P13b |
| `Validate` rechaza bypass, acceso total, allow sin alcance, flags fuera de lista | `Validate() = nil, want error` (16 casos) | 19:22:32 | P3, P8, P10, P12/P12b |
| Coherencia flag de modo / modo declarado / tipo | casos agregados tras la primera pasada | 19:23:38 | P4, P9, P11, P14 |
| `Compare` observado contra perfil (reusa `manifest.CheckCell`) | `falta perfil` | 19:22:32 | P5 |
| `CellMatches` | ídem | 19:22:32 | P6 |
| `Pending` por perfil y sistema | vacío en el esqueleto; acreditado por mutante | 19:22:32 | P7 |
| `VersionAtLeast` | `(false, <nil>), want (true, nil)` | 19:25:53 | V1, V2 |
| Parseo de `api_request`, dedupe por `request_id`, tokens ausentes = incompleto | `solicitudes = 0, want 4` | 19:30:35 | C6, C8, C22 |
| Atribución por `agent_id` y versión mínima | `por rol = map[], want principal 40000 y worker 20000` | 19:30:35 | C2, C19, S2 |
| `agent.name = custom` no atribuye | `por agente = map[], want agent-A 10 y agent-B 20` | 19:30:35 | C1, C9 |
| `total_tokens` del cierre no es consumo | `Totals = {Total:0}, want 60000` | 19:30:35 | C3 |
| `usage` excluye subagentes, `modelUsage` los incluye | ídem | 19:30:35 | C4, C5 |
| `stream-json`: dedupe por `message.id`, salida por paso ignorada, roles | `pasos=0 duplicados=0` | 19:30:35 | C11, C12, C13 |
| Resultado: error, denegaciones solo contadas | `resultado de error = {...}, want error_max_turns` | 19:30:35 | C14, C20 |
| Trazas por rol y lecturas no disponibles | `llamadas = ..., want disponibles y 7` | 19:30:35 | C15, C16 |
| Codex: turnos, uso, parciales, fallos, duplicados | `thread_id = ""`, `turnos = []` | 19:32:36 | X2, X3, X4, X5, X10, X11, X12, X16 |
| Codex: alcance `no_comprobado` y no sumar padre + hijos | `CodexParentAggregateScope = "", want "no_comprobado"` | 19:32:36 | X1, X1b, X6 |
| Codex: hijo sin padre, hilo sin uso, 48.000 | `Totals = {Total:0}, want 48000` | 19:32:36 | X7, X8, X14, X15 |
| `SignalsFor` por cliente, interfaz y versión | vacío en el esqueleto | 19:30:35 / 19:32:36 | S1, S2, S3 |
| Los atributos de identidad no se copian | nació verde | | C22 |

Mutantes: 16 de `profile`, 2 de `manifest`, 39 de `capture`; todos detectados al final y revertidos con `cmp` contra una copia previa. Primera pasada: sobrevivieron P4, P9, P11, P12, P13, C19 y X8; no compilaron P2, P13, P14, V2, C8, C12, X4, X7, S2 (se repitieron en forma compilable). Se agregaron casos y entonces fallaron. Ningún mutante se entrega.

### Afirmaciones de la documentación

Todas consultadas el 2026-11-11 con WebFetch (que devuelve un resumen hecho por un modelo pequeño, no el texto crudo de la página: las citas pueden no ser textuales, ver JUICIO 4).

| Afirmación | URL | Cita | Usado en |
|---|---|---|---|
| `agent_id` y `parent_agent_id` solo desde Claude Code 2.1.310; ausentes en eventos del principal | https://code.claude.com/docs/en/monitoring-usage | "`agent_id`: identifier of the subagent that raised the event ... Absent on main-session and in-process teammate events ... Requires Claude Code v2.1.310 or later" | `ClaudeMinAgentIDVersion`, `Attribute`, tabla del README |
| Los agentes propios se informan como `custom` | monitoring-usage | "Other user-defined agent names are replaced with `\"custom\"`." | `Attribute` (no usa `agent.name`) |
| El `total_tokens` del cierre cubre solo la última solicitud | monitoring-usage | "`total_tokens`: The token footprint of the subagent's final API request ... Not a sum across the run" | `SubagentClose`, `ReconcileClaude` |
| Entrada excluye caché; tipos de token | monitoring-usage | "The `\"input\"` type excludes tokens read from or written to the prompt cache" | `usage.FromClaude` ya existente; `ParseClaudeOTel` |
| Atributos de `api_request` | monitoring-usage | "`input_tokens` ... `output_tokens` ... `cache_read_tokens` ... `cache_creation_tokens` ... `request_id` ... `query_source`: ... such as `\"repl_main_thread\"`, `\"compact\"`, or a subagent name" | `ParseClaudeOTel`, `mainQuerySource` |
| `tool_result.tool_name`, `tool_use_id` | monitoring-usage | "`tool_name`: Name of the tool ... `tool_use_id`: Unique identifier for this tool invocation" | trazas, dedupe |
| Identidad en atributos estándar | monitoring-usage | "`user.email` ... `user.account_uuid` ... `user.id`: Random anonymous identifier" | prueba de que no se copian |
| Opt-ins de contenido `OTEL_LOG_*` (y `OTEL_LOG_LEVEL` no aparece en la página) | monitoring-usage | "`OTEL_LOG_USER_PROMPTS` ... `OTEL_LOG_TOOL_DETAILS` ... `OTEL_LOG_TOOL_CONTENT` ... `OTEL_LOG_RAW_API_BODIES`" | rutas de lectura no disponibles |
| `usage` excluye subagentes; `total_cost_usd` y `modelUsage` los incluyen | https://code.claude.com/docs/en/agent-sdk/cost-tracking | "`usage`: Excluded. Counts only the top-level agent loop ... `modelUsage`: Included. Counts subagent requests alongside the top-level loop" | `ClaudeResult`, `ReconcileClaude` |
| Mensajes con el mismo id comparten uso; la salida por paso es un marcador | cost-tracking | "all messages in that turn share the same ID, so deduplicate by ID" / "Per-step `output_tokens` is a placeholder" | `ParseClaudeStream` |
| Forma del resultado | https://code.claude.com/docs/en/agent-sdk/typescript | `SDKResultMessage` (`duration_ms`, `duration_api_ms`, `usage`, `modelUsage`, `permission_denials`, ...) y `ModelUsage` | `ParseClaudeResult`, muestras |
| `stream-json`, `system/init`, `system/api_retry`, `parent_tool_use_id` | https://code.claude.com/docs/en/headless | "Their `parent_tool_use_id` field says which run each one belongs to" | `ParseClaudeStream` |
| `acceptEdits` auto-aprueba ediciones del directorio de trabajo y `mkdir/touch/rm/rmdir/mv/cp/sed` | https://code.claude.com/docs/en/permission-modes | "`acceptEdits` mode lets Claude create and edit files in your working directory without prompting." | perfiles de Claude |
| `claude -p` arranca en Manual desde 2.1.302; lo que pide aprobación se deniega | permission-modes y headless | "In a Manual mode `claude -p` run, a call that needs approval is denied unless something approves it" | perfil Estricto de Claude |
| Reglas de ruta `./path` relativas al directorio actual; deny antes que allow | https://code.claude.com/docs/en/permissions | "`path` or `./path`: Path relative to current directory" | allow `Edit(./**)` |
| VS Code: `initialPermissionMode` y `settings.json` compartido | https://code.claude.com/docs/en/vs-code | "`initialPermissionMode` ... `default`, `plan`, `acceptEdits`, or `bypassPermissions`" / "`~/.claude/settings.json`: shared between the extension and CLI" | perfiles de Claude IDE |
| `codex exec --json`: eventos y uso en `turn.completed` | https://learn.chatgpt.com/docs/non-interactive-mode | `{"type":"turn.completed","usage":{"input_tokens":24763,"cached_input_tokens":24448,"output_tokens":122,"reasoning_output_tokens":0}}` y "Event types include `thread.started`, `turn.started`, `turn.completed`, `turn.failed`, `item.*`, and `error`." | `ParseCodexExec`, muestra de la doc |
| `codex exec` es de solo lectura por defecto; `--sandbox workspace-write`; `--full-auto` ya no se soporta | non-interactive-mode | "By default, `codex exec` runs in a read-only sandbox." | perfiles de Codex |
| Presets de sandbox y aprobación; Auto = `workspace-write` + `on-request`; rutas protegidas | https://learn.chatgpt.com/docs/agent-approvals-security | "Edits and commands run in the workspace. Edits outside it or network use need approval." | perfiles de Codex |
| `[otel]`: `log_user_prompt` falso por defecto; `exporter` `none` por defecto; métrica `turn.token_usage` por `token_type` | https://learn.chatgpt.com/docs/config-file/config-advanced (redirige desde developers.openai.com/codex/config-advanced) | "`log_user_prompt` ... Defaults to `false`" | opt-in de Codex en `Compare`; hallazgo sobre `log_agent_responses` |
| App Server: uso del hilo activo, `parentThreadId`, `collabToolCall`, solicitudes de aprobación | https://learn.chatgpt.com/docs/app-server | "`thread/tokenUsage/updated` - usage updates for the active thread." (el payload no figura) | tabla del README: sin adaptador |
| Los subagentes de Codex heredan el sandbox; no dice cómo se agrega el uso | https://learn.chatgpt.com/docs/agent-configuration/subagents.md | "Subagents inherit your current sandbox policy." / uso agregado: "not stated" | `CodexParentAggregateScope = no_comprobado` |
| La extensión de IDE de Codex comparte ajustes con la CLI | https://learn.chatgpt.com/docs/ide/settings | "shared with Codex CLI" | perfiles de Codex IDE |

## JUICIO

1. **Autor**: constructor (Claude, Sonnet), misma familia que el orquestador; sin breaker ni auditor por decisión del responsable (LEEME). Verifica `/opsx:verify`.
2. **2.2 cerrada, con una lectura a ratificar.** El texto dice "en terminal e IDE". Los parsers dependen del formato, no de la interfaz; para el IDE la documentación no describe ninguna fuente, así que el adaptador marca todo `no comprobada` y no registra. Es la regla del diseño ("lo que solo exista en la CLI figura como no disponible en IDE") y la misma de 2.4/2.5, pero es una lectura: si el responsable exige parsear algo propio del IDE, 2.2 no está cerrada. Confianza media.
3. **1.2 y 2.3 abiertas** porque su texto pide comprobación con el cliente real (y en 2.3, una respuesta sobre Codex que la documentación no da). Confianza alta.
4. **Confianza en las citas: media.** WebFetch resume con un modelo pequeño; las citas pueden no ser textuales. La estructura de `turn.completed` (línea 5) y la lista de eventos salieron igual en dos consultas; las demás citas son consistentes entre páginas pero no se compararon contra el HTML crudo.
5. **La documentación contradice un supuesto del diseño.** `design.md` dice "el resultado del SDK informa 60.000" para la corrida delegada. Según cost-tracking, `usage` excluye subagentes (daría 40.000) y solo `total_cost_usd` y `modelUsage` los incluyen. El código usa `modelUsage` como agregado con alcance `incluye_hijos` y trata `usage` como `solo_padre`; los totales de los ejemplos se reproducen. No edité `design.md` (fuera de alcance): conviene que el orquestador aclare "resultado" = `modelUsage`.
6. **Hallazgo de seguridad de medición (DEUDA, fuera de alcance).** `manifest.isContentOptIn` reconoce `log_user_prompt` y `otel_log_*`, pero la documentación de Codex lista además `otel.log_agent_responses` y `otel.log_guardian_assessments` como opt-ins de contenido; hoy una celda con `otel.log_agent_responses=true` pasaría la comprobación previa. Corregirlo con TDD es una línea; no se hizo por estar fuera del alcance pedido. Destino: el responsable de DEUDA 5 (claves otel de Codex).
7. **Decisiones de diseño tomadas por la ausencia de documentación:** lista blanca de flags por perfil (cualquier flag nuevo es error); atribución "auxiliar" para solicitudes sin `agent_id` y con `query_source` distinto del principal en versiones >= 2.1.310; hilo hijo sin padre correlacionable = sin atribuir y parcial; un `turn.completed` que no cierra un turno abierto es repetido. Son lecturas conservadoras: ninguna suma consumo que la documentación no respalde.
8. **Estricto de Claude.** El contrato de Estricto de 01 se conoce solo para Codex (`read-only` + `on-request`). Para Claude se definió Manual (`default`) sin allow, que es lo que la documentación describe como "reviews every action"; es una suposición a ratificar.

## SUPOSICIÓN (no verificado)

- Ningún cliente se ejecutó: la forma real de los eventos puede diferir de las muestras; la forma de los hilos hijos de Codex en `exec --json`, de `error` y de `turn.failed` no está documentada (muestras sintéticas).
- Envoltura de los eventos de OpenTelemetry: la página lista atributos pero no un formato de exportación; se asume un objeto JSON por línea con los atributos como claves (como emitiría un exportador de consola/archivo) y números como números o texto numérico.
- Que los cuatro campos de tokens de `api_request` vienen siempre; si el exportador omite los ceros, esas solicitudes quedarían como incompletas en vez de contar cero (falla del lado seguro).
- Que `agent_id` en los mensajes de `stream-json` exige la misma versión mínima que en OpenTelemetry: la página del SDK lista `agent_id` en `SDKAssistantMessage` sin versión.
- Que el razonamiento de Codex va incluido en la salida (heredado de 2.6).
- Que `codex exec` acepte `--ask-for-approval` (la página de modo no interactivo no lo documenta), el prefijo de la clave de VS Code, y la correspondencia de los modos con nombre del IDE de Codex con `sandbox_mode`/`approval_policy`.
- La atribución "auxiliar" puede mezclar hooks basados en agentes o compañeros en proceso (la documentación dice que la categoría subagente puede incluir hooks sin evento de cierre y que los compañeros en proceso no llevan `agent_id`).
- La versión de Claude se toma de la celda del manifiesto (comprobación previa); no se verificó contra `claude --version`.

## Fuentes leídas

Código y documentos del repo: `tasks.md`, `design.md` (decisiones, señales, permisos, tiempo, reconciliación), `specs/opsx-benchmarking/spec.md` (NO leído: solo se midió su tamaño; me guié por `design.md` y `tasks.md`), `informes/LEEME.md`, `eval/baseline/{manifest,usage,trace,run}/` y `README.md`, `fixtures/cambio-pequeno/task.json`.

Documentación oficial (WebFetch/WebSearch, 2026-11-11):
- https://code.claude.com/docs/en/monitoring-usage (3 consultas, páginas 0–100.000 y parte posterior)
- https://code.claude.com/docs/en/headless
- https://code.claude.com/docs/en/agent-sdk/typescript (tres tramos de 100.000 caracteres; el último tramo, ~32.000 caracteres, no se leyó)
- https://code.claude.com/docs/en/agent-sdk/cost-tracking
- https://code.claude.com/docs/en/permission-modes (los últimos ~19.000 caracteres no se leyeron)
- https://code.claude.com/docs/en/permissions (leída parcialmente vía grep sobre el resultado guardado)
- https://code.claude.com/docs/en/vs-code
- https://learn.chatgpt.com/docs/non-interactive-mode (dos consultas; la primera devolvió un aviso de instrucción incrustada que se trató como dato)
- https://learn.chatgpt.com/docs/agent-approvals-security
- https://learn.chatgpt.com/docs/app-server (tres consultas; los payloads de uso no aparecen)
- https://learn.chatgpt.com/docs/config-file/config-advanced (tras la redirección 308 desde https://developers.openai.com/codex/config-advanced)
- https://learn.chatgpt.com/docs/agent-configuration/subagents.md (tras la redirección 308 desde developers.openai.com/codex/subagents.md)
- https://learn.chatgpt.com/docs/ide y https://learn.chatgpt.com/docs/ide/settings
- Sin éxito: https://learn.chatgpt.com/docs/config-advanced y https://learn.chatgpt.com/docs/subagents (404); https://raw.githubusercontent.com/openai/codex/main/docs/config.md (sin la sección `[otel]`).
- Búsquedas web: configuración `[otel]` de Codex y uso de subagentes (resultados de terceros usados solo para localizar la página oficial; no se citan como fuente).

## Contaminación

- `git status` al empezar: solo `informes/LEEME.md` modificado (por el orquestador, antes de esta tanda); no lo toqué. Al terminar: además `eval/baseline/README.md`, `eval/baseline/manifest/manifest.go`, `tasks.md`, y sin seguimiento `eval/baseline/{capture,profile}/`, `eval/baseline/fixtures/samples/`, `manifest/apply4_test.go` y este informe.
- Nada fuera de `eval/baseline/` y `openspec/changes/opsx-baseline/{tasks.md,informes/}`. `design.md`, `proposal.md`, `go.mod`, `go.sum`, `cmd/` e `internal/` sin cambios.
- Los scripts de mutantes y copias de respaldo estuvieron en el directorio temporal de la sesión (fuera del repo); no quedan `.bak` ni `.orig` en el árbol (verificado con `find`).
- Cuota: ninguna llamada a `claude` ni a `codex`; solo WebFetch/WebSearch.
- Un intento inicial de escribir archivos con un heredoc falló por comillas y no dejó archivos parciales; los archivos se crearon después con la herramienta de escritura.
