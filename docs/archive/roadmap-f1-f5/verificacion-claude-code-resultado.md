# Resultado de la verificación en Claude Code · F1–F5

Fecha: 2026-11-18. Encargo: [verificacion-claude-code.md](verificacion-claude-code.md).

## Entorno

| Dato | Valor |
|---|---|
| Base | `origin/main` 59685b1 (v0.8.0 + documentación), rama `verify/claude-f1-f5` en worktree aislado |
| Plataforma | Windows 11 Enterprise 10.0.26200, Git Bash |
| Active Stack | binario global v0.8.0 (release, SHA256 verificado); instalación de prueba con el build de la rama (`install --agent claude --mode full --no-self-install`) |
| Claude Code | 2.1.319, sesiones `claude -p` nuevas por caso, modelo `claude-opus-5-5` |
| OpenSpec | 2.1.0 |
| Otras | Go 1.26.5, Node 26.7.0, Engram 3.7.2, conector ClickUp de claude.ai (MCP oficial `https://mcp.clickup.com/mcp`) |

Los proyectos de prueba se crearon fuera del checkout con datos sintéticos. Las escrituras en ClickUp se hicieron en la lista descartable «Pruebas Active Stack (descartable)» (`1600230000004473`, espacio Grupo Desarrollo, workspace `90171440963`).

## Defectos corregidos en esta rama

| # | Defecto | Corrección |
|---|---|---|
| 1 | `active-delivery` y `active-roadmap` declaraban `agents: [codex]`; Claude no recibía F3 ni F4 y la fundación mostraba ClickUp «no disponible». | Catálogo con `agents: [codex, claude]`; tests de catálogo e instalación para ambos clientes. |
| 2 | `roadmap.py setup` solo leía `codex mcp list --json`; en una máquina con Claude cortaba con «No se pudo consultar la configuración MCP de Codex». | `setup --client claude` interpreta `claude mcp list` (incluido el conector de claude.ai); `--client codex` sigue siendo el valor por omisión. |
| 3 | En Windows, `subprocess.run(["codex"…])` y `["claude"…]` no encuentran los shims `.cmd` de npm, así que el diagnóstico fallaba también en Codex. | `client_command` resuelve la ruta con `shutil.which`. |
| 4 | La skill solo explicaba la conexión con `codex mcp add/login` y `openspec init --tools codex`. | Instrucciones por cliente: `claude mcp add --transport http --scope user clickup …`, `/mcp` para login, prefijos de herramientas y `openspec init --tools claude`. |
| 5 | El bloque universal de Claude no nombraba `active-delivery`, a diferencia de Codex (`$active-delivery`). | Línea de entrada en `ComposeLayers`; goldens e inventario regenerados (bloque de Claude 5.902 → 6.046 B). |
| 6 | Los tests de `internal/install` y `cmd/active-stack/headless` agregaban carpetas temporales al PATH de usuario en el registro de Windows aunque `LOCALAPPDATA` estuviera aislado (punto 1 de `correcciones-stack-main-2026-11-17`). | `TestMain` de ambos paquetes reemplaza la escritura persistente del PATH. Verificado: suite completa sin cambios en `HKCU\Environment\Path`. |

| 7 | Con el mismo pedido, Claude a veces no invocaba `active-delivery`, porque la línea de entrada competía con «Handle routine tasks directly». | La línea del bloque universal pide invocar la skill antes de empezar o retomar trabajo y limita la guía de tareas rutinarias al caso sin skill. |
| 8 | Tras recomendar ODD, la skill implementaba sin esperar; al retomar, volvía a preguntar un modo ya confirmado. | `active-delivery` espera la confirmación del modo la primera vez y la reutiliza cuando el traspaso la registra con el mismo alcance. |
| 9 | La configuración de ClickUp se guardaba antes del diagnóstico, con `project_id` tomado del nombre de la carpeta. | `setup --list-id` diagnostica antes de escribir y `roadmap.py configure` guarda el destino con un UUID estable, copia previa y responsables conservados. |
| 10 | El diagnóstico de Claude descartaba los servidores stdio (un `clickup` stdio daba `missing`), tomaba SSE como oficial, podía elegir una conexión caída y no permitía elegir entre dos; `--client` era opcional. | Se conservan stdio y SSE con su transporte, se prefiere la conexión alcanzable, `--name` elige una conexión y `--client` es obligatorio. |
| 11 | Tras publicar o actualizar se sincronizaba el estado sin pedido, y las consultas de avance o diagrama no cargaban la skill. | La reconciliación queda limitada al pedido de sincronizar o a retomar, entregar o cerrar; la descripción de la skill cubre avance y diagramas. |

Pruebas: `go test ./...` con `USERPROFILE`, `HOME`, `LOCALAPPDATA` y `APPDATA` aislados (53 paquetes OK) y `python -B -m unittest` en `active-roadmap` (27 OK).

## Resultados por ficha

### F1 · Fundación

| Escenario | Esperado | Observado | Sesión | Resultado |
|---|---|---|---|---|
| Ruta directa | Recomienda directa, espera elección, sin escrituras de fundación | Recomendó directa con justificación y esperó; al confirmar corrigió y testeó sin estado ni OpenSpec | 8a3bdaf3 | OK |
| Checklist | Fases, obligatorias, variantes, reutilización, bloqueos, sin escrituras | Checklist completa con escrituras previstas; nada escrito antes de aprobar | dac0b6e3 | OK |
| Aprobación | Resumen final y OK explícito antes de escribir | Mostró plan final y esperó | dac0b6e3 | OK |
| Pausa y reanudación | Estado persistido; sesión nueva ofrece retomar sin repetir fases | `kb: done`, `step: visual-design`; la sesión nueva mostró progreso, verificó la KB y completó el plan. KB idéntica por hash | 00ada0d5 | OK |

Bloqueo por permisos: dos defectos de F1 están en la skill externa `active-orchestrator` (Group-Active-IA/active-orchestrator@4318a75), donde la cuenta del responsable tiene solo lectura.

1. **Entrada en Claude.** La skill anuncia `/active-orchestrator:init` y entradas como `:design` o `:find-skill`. En Claude Code `/active-orchestrator:init` no existe como comando y la fundación no se activa; funciona `/active-orchestrator init` o el pedido en lenguaje natural. Las sugerencias del propio modelo alternaron entre `:find-skill` y ` :find-skill`.
2. **Plan modificado sin aprobación.** Un subagente respondió la checklist solo con variantes (`1=skip, 2=import, 3=off, …`) para un plan distinto del recomendado. La skill escribió el estado con `approved_at`, despachó `kb-creator` y no mostró el resumen final ni pidió aprobación, aunque su contrato dice «A new or changed plan requires approval».

Corrección propuesta en la rama `fix/claude-entrypoints-plan-approval` de ese repositorio (commits 3223696 y 1518ef9, 16 pruebas de contrato OK). La skill documenta la forma con espacio para Claude, una respuesta que solo elige fases o variantes cambia el plan sin aprobarlo y el resumen lista el registro de política visual en `design-system/README.md`. Con esa versión instalada a mano, los subagentes verificaron F1 y F2 completos. El cambio se publicó desde el fork `cachi001/active-orchestrator` en el [PR #6](https://github.com/Group-Active-IA/active-orchestrator/pull/6). El SHA del catálogo se actualiza cuando ese PR se integre.

### F2 · Adaptación de arneses

| Escenario | Esperado | Observado | Sesión | Resultado |
|---|---|---|---|---|
| Generación | KB, roadmap e instrucciones se ofrecen y se guardan | `knowledge-base/`, `CHANGES.md`, `.active-stack/roadmap.json` y `CLAUDE.md` creados por sus fases | dac0b6e3 / 00ada0d5 | OK |
| Reutilización | Proyecto existente sin estado: ofrece `reuse` y no regenera | Ofreció reuse/update, detectó referencias a archivos inexistentes y no las dio por válidas; con `reuse` los artefactos quedaron idénticos por hash y solo se crearon estado y `roadmap.json` | 4e543cb2 | OK |

### F3 · ODD junto a SDD

| Escenario | Esperado | Observado | Sesión | Resultado |
|---|---|---|---|---|
| Elección | Invoca `active-delivery`, inspecciona sin escribir y recomienda | Invocó la skill con el ID C-07, detectó el contrato público y ofreció ODD o SDD | 24e456b1 | OK |
| ODD con traspaso | Alcance y verificación acordados; traspaso al dejar el trabajo | `knowledge-base/entrega/C-07.md` con base Git, decisiones, verificación y pendientes; sin OpenSpec | 24e456b1 | OK |
| Reanudación ODD | Sesión nueva compara Git, revalida y continúa | Comparó HEAD y árbol, reejecutó los tests, implementó dentro del alcance y actualizó el traspaso | 0c33da65 | OK |
| Promoción | Pausa, explica, pide confirmación antes de OpenSpec y del init | Pausó sin tocar el diff, propuso `openspec init --tools claude`, esperó; con OK creó `cotizar-v2` con la evidencia del traspaso | 0c33da65 | OK |
| Reanudación SDD | Sesión nueva retoma el change existente | Identificó `cotizar-v2` y los pendientes sin duplicarlo ni escribir | 2c7821cd | OK |

Observación: `openspec init` 2.1.0 reescribe además las skills `openspec-*` de `~/.agents/skills/`. Es comportamiento del CLI externo; conviene mencionarlo al pedir la confirmación.

### F4 · ClickUp

| Escenario | Esperado | Observado | Sesión | Resultado |
|---|---|---|---|---|
| Configuración | Diagnóstico y lectura de la lista antes de escribir la configuración | `setup --client claude` → `check_list` con «claude.ai ClickUp»; `clickup_get_list` validó la lista; `roadmap.json` escrito tras aprobación con UUID | c3a79da1 | OK |
| Menú y detalle | Acciones con requisitos, bloqueadas explicadas, sin escrituras | Menú, bloqueos y «Ver detalle»; lista vacía y sin `roadmap-state.json` después | 413d9bda | OK |
| Publicar una de dos | Vista previa, selección y escritura solo de la fila aprobada | Vista previa con marcadores y Mermaid; con «solo T-01» se creó `1dfcuabdctr` y T-02 no | 413d9bda | OK |
| Actualizar vinculada | Inspección, diff, confirmación; preserva texto externo, estado y responsables | Cambió solo el alcance; la nota externa, el diagrama, el estado y la asignación quedaron iguales (lectura independiente) | 9c88f551 | OK |
| Ver y sincronizar | Ver no escribe; sincronizar guarda el estado observado | Ver dejó el estado local idéntico por hash; sincronizar guardó `in progress` sin escribir en ClickUp | 229d1118 | OK |

Observaciones: el primer intento de sincronizar pasó al helper una ruta temporal de Git Bash que Python no resuelve y se recuperó solo. La primera creación de la lista de pruebas devolvió «ClickUp server error»; se comprobó que no existía antes de reintentar y la misma descripción se aceptó después, así que fue un fallo transitorio del servidor.

### F5 · Visualizaciones

| Escenario | Esperado | Observado | Sesión | Resultado |
|---|---|---|---|---|
| Varios bloques | Una opción por diagrama y espera | Lista numerada Intake, Approval y Handoff (modo no interactivo) | 8be131eb | OK |
| Selección | Solo el bloque elegido, exacto | Solo `Confirmation --> Handoff` | 8be131eb | OK |
| Coincidencia directa | Bloque sin pregunta adicional | Devolvió el bloque Handoff | e8e0b0b5 | OK |
| Fuente ausente | Indica qué falta | Informó que `ARQUITECTURA.md` no existe y listó los bloques disponibles | 1fd34091 | OK |
| Discovery | Mermaid fiel sin inferencias | Dibujó el ciclo de aclaración y dejó fuera del diagrama lo que la fuente no define | f6af9bb6 | OK |
| ClickUp | Bloque exacto de `markdown_description` con contexto | Devolvió el bloque de `1dfcuabdctr` con estado, watchers y nota | 85e08a2b | OK |
| ClickUp sin acceso | Indica qué falta | Tarea inexistente: explicó las causas y no inventó un bloque | 53ba7d86 | OK |

| Selector nativo (CLI interactivo) | Selector de opción única, solo el bloque elegido | El responsable abrió `claude` en el proyecto de prueba: apareció el selector con Intake, Approval y Handoff; al elegir Handoff mostró solo `Confirmation --> Handoff` | manual, 2026-11-18 | OK |

Render: `claude -p` no expone el selector y devuelve una lista numerada. En el CLI interactivo la terminal muestra el bloque Mermaid como código editable, sin dibujarlo, que es el comportamiento esperado en una superficie sin render. En la app de escritorio, con la misma carpeta y el mismo pedido, el responsable observó el diagrama dibujado, con un control (ícono de ojo) que alterna entre el dibujo y el código fuente Mermaid (manual, 2026-11-18).

## Verificación independiente con subagentes

Tres subagentes repitieron la matriz en proyectos nuevos (`sb1/`, `sb2/`, `sb3/`) con sesiones nuevas de `claude -p`, y un cuarto revisó el diff. Una primera tanda quedó incompleta por el límite de uso y sus hallazgos originaron las correcciones 7 a 11. La segunda tanda corrió completa sobre la instalación corregida.

| Ficha | Casos | Resultado | Evidencia principal |
|---|---|---|---|
| F1 · F2 | Ruta directa, checklist, respuesta solo con variantes, aprobación con pausa tras kb, reanudación y reutilización | OK con la corrección propuesta de `active-orchestrator` | Sin escrituras antes del «apruebo», KB y artefactos idénticos por hash, sugerencias `/active-orchestrator <acción>` |
| F3 | Elección repetida tres veces, traspaso, reanudación, promoción y reanudación SDD | OK | Skill invocada y espera en las tres corridas, `git diff --stat` idéntico tras la pausa, un único change en `openspec list` |
| F4 | Configuración, menú y detalle, publicación de una fila de dos, actualización vinculada, avance y sincronización | OK | Lecturas independientes de la lista: una sola tarea nueva (V-01, `1dfcuabdcyn`), nota externa, Mermaid, estado y responsables conservados, `1dfcuabdctr` intacta |
| F5 | Varias opciones, coincidencia directa, fuente ausente, Discovery y ClickUp | OK | Bloques idénticos a la fuente por comparación exacta |
| Revisión | Diff completo | Sin defectos bloqueantes; hallazgos del parser corregidos (punto 10) | 27 pruebas Python del helper |

Desvíos menores observados: la confirmación de `openspec init` se pidió en la misma pregunta que la promoción, antes de ejecutar nada; algunas respuestas mezclaron tuteo y voseo; las sesiones guardaron entradas en Engram por las instrucciones globales del usuario, no por las skills del stack.

## Cierre de integración · 2026-11-18

Los [PR #23](https://github.com/Group-Active-IA/active-stack/pull/23) y [#24](https://github.com/Group-Active-IA/active-stack/pull/24) quedaron integrados en main de Active Stack. El pendiente de `active-orchestrator` descrito arriba queda resuelto: se fusionó el [PR #6](https://github.com/Group-Active-IA/active-orchestrator/pull/6), con merge `09c841e33c33744abeafe600e88fe1c36fa0f599`. El catálogo reemplaza el pin `4318a75008e72d4820bb9800bb16d5a413c4e80b` por ese merge de upstream para que Full instale la corrección en Claude, Codex, OpenCode y Antigravity.

La revisión de integración ejecutó `python -m pytest -q` en un clon temporal de la skill: 14 pruebas pasan en la base y 16 en la versión corregida. En Active Stack, `go test ./internal/catalog` pasó antes del cambio; la nueva prueba `TestActiveOrchestratorFullUsesApprovedClaudeFix` falló primero para los cuatro clientes con el pin antiguo y pasó al actualizarlo. También pasó `go test ./internal/catalog -count=1`, que valida el YAML embebido mediante `catalog.Load()`.

Esta comprobación verifica el contrato publicado y la selección de la versión corregida por el catálogo. No repite ni atribuye nuevos pilotos de Claude o escrituras en ClickUp; los recorridos y sesiones de las secciones anteriores conservan su evidencia original.

### Evidencia del ciclo TDD

| Tarea | Prueba | Capa | Red de seguridad | RED | GREEN | Triangulación | Refactor |
|---|---|---|---|---|---|---|---|
| Actualizar pin de active-orchestrator | `internal/catalog/orchestrator_release_test.go` | Catálogo | `go test ./internal/catalog` OK | Cuatro clientes fallan con el SHA anterior | Prueba focalizada y paquete completo OK | Claude, Codex, OpenCode y Antigravity | No necesario |

## Observación previa fuera del cierre de Claude

- `starter add` sigue haciendo `self-install` en proyectos (punto 2 de `correcciones-stack-main-2026-11-17`); no se tocó.
