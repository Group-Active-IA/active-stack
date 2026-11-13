# eval/baseline

Arnés de medición de referencia del change `opsx-baseline` (ficha 02).

- **Externo al binario; no importado por `cmd/` ni `internal/`.** Decisión D6: directorio de evaluación separado, en Go, solo biblioteca estándar, sin tocar `go.mod`.
- **Nombre del directorio: `eval/baseline/` (TBD · renombrable).** No está decidido; renombrarlo solo exige mover la carpeta, porque nada fuera de `eval/` lo importa.

## Paquetes

| Paquete | Tarea | Contenido |
|---|---|---|
| `manifest` | 1.1 | Manifiesto versionado (JSON), matriz de cohortes con tareas y pesos, validación previa por celda (`ejecutable` / `incompatible` con motivo), plan con celdas fuera del denominador y enum de estados de corrida. |
| `review` | 1.4 | Regla de emparejamiento de la revisión de un diff sembrado (verdadero positivo, duplicado, falso positivo, hallazgo nuevo a adjudicación, omisiones). Desde la tanda 2 rechaza con error los defectos con archivo o mecanismo vacío y las disposiciones desconocidas (DEUDA 6). |
| `fixture` | 1.3 | Carga y validación de fixtures (huella del prompt, huella del estado inicial, criterios obligatorios con chequeo conocido), checkout descartable en un temporal con `Reset`/`Close`, evaluación por criterio y veredicto (`aceptada`, `rechazada_por_calidad`, `pendiente_de_revisor`). |
| `profile` | 1.2 | Perfiles de medición por cliente e interfaz (equivalente para edición y Estricto, 8 en total) con modo, flags y fuentes de la documentación; `Validate` (rechaza bypass, acceso total, allow sin alcance, directorios extra), `EditInsideCheckoutNeedsApproval`, `Compare` (permisos observados contra el perfil, reusando `manifest.CheckCell`), `CellMatches` (celda del manifiesto contra el perfil) y `Pending` (verificación con el cliente real en Windows y Linux, pendiente). |
| `capture` | 2.2, 2.3 | Adaptadores de captura: `ParseClaudeOTel`, `ParseClaudeResult`, `ParseClaudeStream`, `ParseCodexExec`, atribución por rol (`Attribute`, `ReconcileClaude`, `ReconcileCodex`), registros de traza por interfaz y `SignalsFor` (estado de cada señal por cliente, interfaz y versión). |

## Fixtures (`fixtures/<familia>/`)

Cada familia tiene `task.json` (id, familia, versión de rúbrica, fuente congelada con huella, prompt con huella sha256, herramientas permitidas, límites, criterios), `prompt.md`, la clave cuando la familia se evalúa por hechos, y una referencia y una variante defectuosa.

| Familia | Fuente | Referencia / defectuosa |
|---|---|---|
| `exploracion` | `git archive` de `1b4639b537a4d0c04dc8c25c0220732095a772d3` | respuestas JSON |
| `cambio-pequeno` | mini-proyecto Go sintético `project/` | parches (`git apply`) |
| `cambio-multiarchivo` | mini-proyecto Go sintético `project/` | parches |
| `diagnostico` | mini-proyecto Go sintético con falla sembrada | respuestas JSON |
| `revision` | la misma revisión + `seeded.patch` | respuestas JSON con hallazgos |

- Los mini-proyectos tienen su propio `go.mod`, así que `go test ./...` del repo no los incluye (el de diagnóstico falla a propósito). Las pruebas ocultas viven en `hidden/` con sufijo `.go.txt` y el evaluador las copia sin el sufijo.
- La huella de árbol es el sha256 de las líneas `sha256sum` ordenadas por ruta: `find . -type f | sed 's|^\./||' | LC_ALL=C sort | xargs sha256sum --text | sha256sum`.
- Cambiar un prompt o un proyecto exige recalcular su huella en `task.json`: es congelar una versión nueva del fixture.
- Los límites figuran `(TBD · D3)` hasta que se acuerden.

## Pruebas

```
go test ./eval/baseline/...
```

Los datos de los tests son sintéticos. Los adaptadores de captura de Claude Code y Codex (tarea 2.2, tanda 4) parsean muestras derivadas de la documentación oficial (`fixtures/samples/`, ver su README); ningún cliente real se ejecutó.

## Arnés (tanda 3, parte A)

- `run/` (2.1): `Executor.Execute` prepara un checkout descartable con `fixture`, llama a un `Runner` (interfaz; el cliente real llega con 2.2), evalúa y reintenta desde el fixture reiniciado según `Policy`. Cada intento guarda estado (`manifest.RunStatus`; `pendiente_de_revisor` no es final), causa, código de salida bruto y los eventos `turn.failed`/`error` como JSON bruto, sin semántica asumida. `ExecutePlan` corre solo las cohortes ejecutables de `manifest.Plan`.
- `trace/` (2.4): lecturas y llamadas a herramientas por tipo y rol; una señal que la interfaz no expone es `no disponible`, no cero.
- `timing/` (2.5): reloj monotónico inyectable, esperas de aprobación aparte, unión de intervalos solapados; el ajustado es no disponible si alguna espera no se puede delimitar.
- `checkout_intacto` compara la huella de archivos y además rechaza directorios vacíos agregados y entradas no regulares (symlinks): la huella solo cuenta archivos regulares, así que esas adiciones se detectan aparte.
- Duración: el paquete `fixture` tarda unos 100 s en Windows; el tiempo es casi todo `os.Open` de ~650 archivos por cada `Reset`/`TreeHash` (antivirus) y `go test` de los mini-proyectos. Las pruebas pesadas corren en paralelo (`t.Parallel`); no se recortó la verificación de la huella.

## Consumo, registros y resumen (tanda 3, parte B)

- `run/`: si el llamador cancela durante un intento, la corrida parcial conserva ese intento con estado `interrumpida_por_llamador` (no final), causa, código y eventos; `Execute` devuelve además el error del contexto.
- `usage/` (2.6): `FromClaude` (entrada lógica = entrada + lectura + creación de caché) y `FromCodex` (la entrada cacheada ya está en la entrada; sin caché por diferencia). `Probe` separa la entrada de una solicitud del total del turno; el total solo se usa como aproximación identificada si se comprobó que fue una única solicitud, y el estado de caché observado nunca se reescribe para coincidir con el previsto. `Reconcile` aplica la precedencia de design.md (60.000 en el ejemplo de Claude, 48.000 en el de Codex): con el agregado comprobado como inclusivo no se suman padre e hijos; detalle menor = sin atribuir y parcial; detalle mayor = ambas cifras y fuera de cobertura completa.
- `record/` (2.7): `Builder.Records` emite un registro por intento y uno por corrida (`WriteJSONL`), con cobertura por señal y conteo de redacciones por categoría. Lista sensible mínima aplicada (la lista completa es de 05): correo, UUID de cuenta, id de usuario, rutas del host, detalle de herramientas y cuerpos de solicitud. Los eventos de error se conservan como JSON con esas claves y textos redactados; un evento que no es JSON válido se reduce a su tipo. Huellas: `sha256` con el algoritmo en el campo (definición de 06). La referencia a evidencia local se conserva (una ruta absoluta se reduce a `local:<archivo>`) o se registra su ausencia con motivo. Los UUID de sesión y agente son identificadores opacos y se conservan.
- `summary/` (3.1): `Compute` y `Markdown`. El cálculo está documentado en el comentario del paquete y en la sección "Cálculo" del Markdown: el costo total suma todos los intentos de todas las corridas iniciadas; el cociente por aceptada queda sin valor con cero aceptadas o cobertura incompleta; las incompatibles quedan fuera y se listan aparte; las pendientes de revisor figuran aparte y dejan el resumen provisional; cada cohorte (cliente, versión, interfaz real, modalidad, variante) es una fila propia.

## Fuentes y límites de atribución por versión e interfaz (tareas 2.2 y 2.3)

Documentación consultada el 2026-11-11. "Documentada" = la página oficial lo afirma; "no comprobada" = la página no lo dice o no se probó con el cliente real (se conoce por una corrida de calibración, 3.2); "no disponible" = la documentación indica que la señal no existe en esa fuente. `capture.SignalsFor` expone esta misma tabla en código.

| Cliente · interfaz · versión | Uso de tokens | Atribución por worker | Llamadas a herramientas | Lecturas de archivos | Tiempo y permisos |
|---|---|---|---|---|---|
| Claude Code · terminal · >= 2.1.310 | Documentada: eventos `api_request` de OpenTelemetry (`input_tokens` sin caché, `cache_read_tokens`, `cache_creation_tokens`, `output_tokens`) y resultado de `-p` (`usage`, `modelUsage`). | Documentada: `agent_id` y `parent_agent_id` en los eventos (ausentes en los del principal). Solo `agent_id` atribuye. | Documentada: `tool_result.tool_name`, rol por `agent_id`. | No disponible sin `OTEL_LOG_TOOL_DETAILS=1` (opt-in de contenido prohibido en corridas medidas). | Documentada: `duration_ms` y `duration_api_ms` del resultado; `tool_decision.permission_mode` y `decision`. |
| Claude Code · terminal · < 2.1.310 | Documentada (igual). | No disponible: sin `agent_id`, todo lo que no es del principal queda sin atribuir y la atribución es parcial. | No disponible por rol (no se separa principal de worker). | No disponible. | Documentada, salvo `permission_mode` de `tool_decision`, que la página fija desde 2.1.301. |
| Claude Code · IDE (VS Code) | No comprobada: la página de la extensión no menciona telemetría; comparte `~/.claude/settings.json` con la CLI y trae su propia CLI. | No comprobada. | No comprobada. | No comprobada. | No comprobada; la extensión muestra solicitudes de permiso en la interfaz. |
| Codex · terminal (`codex exec --json`) | Documentada: `turn.completed.usage` (entrada, entrada cacheada, salida, salida de razonamiento); la entrada cacheada es parte de la entrada. | No comprobada: la documentación no dice cómo aparecen los hilos hijos en `exec --json` ni si el agregado del padre los incluye. | Documentada: items `command_execution`, `file_change`, `mcp_tool_call`, `web_search`. | No disponible como items propios (se hacen con `command_execution`). | Tiempo: no disponible en exec (los eventos de turno no documentan marcas de tiempo); permisos: no comprobada en exec (las solicitudes y resoluciones de aprobación son de App Server). |
| Codex · IDE | No comprobada: la extensión comparte `config.toml` con la CLI, sin fuentes de eventos documentadas. | No comprobada. | No comprobada. | No comprobada. | No comprobada. |

Reglas de atribución implementadas y probadas:

- **Claude, `agent_id`.** Con versión >= 2.1.310 una solicitud con `agent_id` es de un worker, identificado solo por ese id. Sin `agent_id` y con `query_source` `repl_main_thread` es del principal; con otro `query_source` es consumo auxiliar (los subagentes sí llevan `agent_id`). Antes de 2.1.310 lo que no es del principal queda sin atribuir y la atribución es parcial. La versión sale del manifiesto (`manifest.VersionAtLeast`).
- **Claude, `agent.name`.** Los agentes propios se informan como `custom` (cita: "Other user-defined agent names are replaced with `custom`"): dos workers distintos con `agent.name = custom` siguen separados por `agent_id`, y un evento `custom` sin `agent_id` no se atribuye a ningún worker.
- **Claude, cierre del subagente.** El `total_tokens` de `subagent_completed` es "the token footprint of the subagent's final API request ... Not a sum across the run": se parsea (`SubagentClose`) pero nunca entra en el consumo.
- **Claude, alcance del resultado.** `usage` excluye a los subagentes ("Counts only the top-level agent loop"); `total_cost_usd` y `modelUsage` los incluyen. `ReconcileClaude` usa la suma de `modelUsage` como agregado con alcance `incluye_hijos` y, si solo hay `usage`, como `solo_padre` (no es total del árbol). El "resultado del SDK que informa 60.000" de `design.md` corresponde a `modelUsage`, no a `usage`.
- **Claude, solicitudes repetidas.** `api_request` se deduplica por `request_id`, `tool_result` por `tool_use_id`; en `stream-json`, los mensajes con el mismo `message.id` cuentan una vez y la salida por paso es un marcador (se lee del resultado).
- **Codex, agregado del padre.** La documentación consultada no dice si el agregado del padre incluye a los hijos: el alcance queda `no_comprobado` (`CodexParentAggregateScope`). `ReconcileCodex` nunca suma el agregado a los hilos; solo lo acepta como total si coincide exactamente con la suma por hilo (48.000 en el ejemplo), y si difiere conserva ambas cifras y marca la corrida parcial. Un hilo hijo sin padre correlacionable queda sin atribuir (parcial).
- **Datos sensibles.** Los adaptadores no copian correo, UUID de cuenta ni id de usuario; los eventos crudos que se conservan (`run.Event`) los redacta `record` (2.7). Las `permission_denials` solo se cuentan.
