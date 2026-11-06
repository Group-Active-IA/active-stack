# Design: claude-native-opsx

Motivación en proposal.md (Why). Requisitos en `specs/claude-instruction-layers`,
`specs/claude-coordinator-agent` y `specs/claude-orchestrator-migration`.

## Context

### Recorrido de InstallSDD para Claude (estado en 97d7f6c)

1. `internal/harness/config/install.go:21` `Install` recibe el harness `sdd-orchestrator`
   (catálogo: `internal/catalog/harnesses.yaml:65-72`, toggles `tdd, engram, model-routing,
   delegation, governance`, modos `lite, full`, agentes `claude, codex, opencode, antigravity`).
2. `install.go:43` decide por `adapter.ConfigDelivery()`. Claude devuelve
   `ConfigDeliveryInstructions` (`internal/agents/claude/adapter.go:125-127`), así que va a la rama
   `default` (`install.go:62-76`).
3. `Compose` (`internal/harness/config/compose.go:20-34`, función pura) arma: base
   `assets/claude/sdd-orchestrator.md`; quita `sdd-delegation` y `sdd-model-assignments` si el toggle
   está inactivo; agrega `governance.md`, `engram-protocol.md` y `strict-tdd.md` si están activos.
4. `Inject` → `InjectSection(path, "sdd-orchestrator", ...)` (`inject.go:85`, `:101`): snapshot si
   existe, `PurgeStaleSections` (borra toda sección `active-stack:` no registrada como propia,
   `inject.go:66-73`), inyección por markers y `WriteFileAtomic`.
5. Ruta: Machine `~/.claude/CLAUDE.md` (`adapter.go:28-30`); Project `<root>/.claude/CLAUDE.md`
   (`adapter.go:91-98`, resuelto por `internal/install/plan.go:369-372`).
6. Verify: `checkConfig` cuenta el marker una vez (`internal/verify/harness_checks.go:98-130`);
   la variante primary-agent está en `:244`. Uninstall: `internal/uninstall/steps.go:150-170`
   quita la sección y, sólo para `ConfigDeliveryPrimaryAgent`, la entrada JSON.

### D0. Evidencia de contexto efectivo (EVIDENCIA, 2026-11-06)

Claude Code 2.1.318, modelo sonnet, proyecto temporal con `CLAUDE.md` que contenía el testigo
aleatorio `CANARIO-ceea159f1478`, más el `~/.claude/CLAUDE.md` real del usuario:

| Caso | Invocación | Vio CLAUDE.md de proyecto | Vio CLAUDE.md global |
|---|---|---|---|
| A | `claude -p --tools ""` | sí | sí |
| B | `--agents '{sonda...}' --agent sonda` (hilo principal) | sí | sí |
| C | subagente definido por `--agents` JSON, sin herramientas, lanzado con Agent | sí (transcripción `subagents/agent-*.jsonl`) | sí |

En C el prompt del subagente (355 caracteres) no contenía los testigos. Conclusión: todo lo que
queda en `CLAUDE.md` llega a sesión, coordinador y especialistas. Mover el coordinador a un agente
aísla SU prompt, no lo que quede en `CLAUDE.md`. Esto fija el diseño: `CLAUDE.md` sólo puede
contener reglas verdaderamente universales.

### Precedente en el repo

`InjectPrimaryAgent` (`internal/harness/config/primary_agent.go:25-89`) resolvió lo mismo para
OpenCode: registra el orquestador en `opencode.json` y purga la sección del archivo compartido
(`purgeInstructionsSection`, `:96`) con snapshot previo. Su comentario (`:29-34`) documenta la fuga
("leaks the orchestrator into every agent").

## Goals / Non-Goals

**Goals:**
- Tres capas con dueño: universal (en `CLAUDE.md`), coordinador (agente nativo), módulos por acción.
- Reutilizar el patrón snapshot + purga acotada + escritura atómica del precedente de OpenCode.
- Pruebas de contexto efectivo con testigo reproducibles, además de tests Go del instalador.

**Non-Goals:**
- Filtrar el contenido propio del usuario en `CLAUDE.md`: llega a todos (D0) y se documenta como
  límite, no se resuelve.
- Cambiar la entrega de OpenCode, Codex o Antigravity (Codex es Propuesta 04).
- Reescribir el contenido normativo de governance, TDD o Engram: sólo se reubica.
- Mover la sección `design-mockups`: sigue en `CLAUDE.md` en este change (ver G7).

## Decisions

### D1. Nuevo modo de entrega para Claude: archivo de agente

`ConfigDeliveryPrimaryAgent` no sirve tal cual: escribe `{"agent": {"<id>": {mode, prompt, ...}}}`
en `SettingsPath`, que para Claude es `~/.claude/settings.json` (`adapter.go:46-48`). En Claude Code
`agent` es el setting que elige el agente de la sesión (`claude --help`: "--agent <agent> Agent for
the current session. Overrides the 'agent' setting"), así que reutilizar `InjectPrimaryAgent` sobre
Claude escribiría un objeto donde Claude espera el nombre de un agente y además activaría el
coordinador por defecto. Se agrega un `ConfigDelivery` nuevo (nombre tentativo
`ConfigDeliveryAgentFile`) y un inyector que escribe `<base>/.claude/agents/<id>.md` con
frontmatter y marca de propiedad, más un método del adapter para el directorio de agentes
resuelto por target (regla del proyecto: nunca hardcodear rutas de agente).

- Alternativa rechazada: reutilizar `ConfigDeliveryPrimaryAgent` con un adapter de settings para
  Claude. Mezcla dos formatos y activa el coordinador por defecto.
- Alternativa rechazada como default (queda como opción de G9): `--agents <json>` en un alias de
  shell. No es persistente ni verificable por el instalador. `claude --help` (2.1.318) admite ruta
  de archivo en `--agents` sólo junto con `--print`; en modo interactivo exige JSON en línea.

**Exposición como subagente (hallazgo B1 del plan-reviewer).** Un archivo en un directorio de
agentes descubrible no sirve sólo para `--agent`: también puede quedar disponible como tipo de
subagente para la herramienta Agent de cualquier sesión, y Claude Code puede delegar por
`description` sin pedido explícito (documentación pública, SUPOSICIÓN en 2.1.318). D0 caso C
mostró que la sesión normal lanza agentes definidos por `--agents`. Si eso ocurre, el coordinador
correría como subagente sin poder delegar (los subagentes no lanzarían subagentes, SUPOSICIÓN) y su
regla "NEVER do apply or propose work inline" (`sdd-orchestrator.md:208`) lo dejaría sin salida.
Mitigaciones propuestas, en orden, a validar en la tarea 1.10:
1. Campo de frontmatter que excluya al agente de la delegación o de la lista de subagentes, si
   existe en la versión mínima.
2. `description` que diga que sólo se usa con `claude --agent <id>` y no como subagente.
3. Guarda en el cuerpo del coordinador: si no dispone de la herramienta Agent, no coordina y
   responde con la instrucción de invocación.
Si 1.10 demuestra que no se puede excluir, la vía D1 no se considera cerrada y decide el grupo
(G9): aceptar con 2 y 3, instalar la definición fuera del directorio descubrible e invocarla con un
lanzador que pase `--agents` en línea más `--agent`, o volver a otra vía.

### D2. Contenido de `CLAUDE.md`: sólo bloque universal mínimo

Una sección administrada nueva (ID tentativo `active-stack:universal-rules`) que se registra como
propia (en `ownedSectionIDs` o `sectionFragments`) para que `PurgeStaleSections` no la borre en la
próxima instalación. Contenido candidato: reglas que deben valer igual para sesión, coordinador y
especialista (ej. "nunca crear `openspec/` a mano", protocolo mínimo de memoria, aviso de que los
permisos los impone `settings.json`). Los permisos del harness `permissions` ya son universales por
mecanismo (settings), no por instrucciones.

- Alternativa rechazada: dejar todo en `CLAUDE.md` con secciones por rol. D0 demuestra que todas
  las secciones llegan a todos; los encabezados no aíslan nada.
- Alternativa rechazada: bloque universal vacío. Pierde reglas que hoy recibe la sesión normal y
  rompe el criterio 1 del roadmap.

### D3. Invocación del coordinador: explícita, `claude --agent <id>`

Recomendación: invocación explícita únicamente. El setting `agent` queda documentado como opt-in
manual del usuario; el instalador nunca lo escribe. Fallback: en este change se documenta en README
y no se instala (recomendación dentro de G2); verify sólo nombra vías instaladas (agente nativo o
modo legado por versión). Si el grupo decide entregarlo, se agregan tareas RED de instalación y
detección (tarea 9.3 condicional). Decisión final: G2.

- Alternativa rechazada como default: setting `agent` escrito por el instalador. Equivale a volver
  a inyectar el coordinador en toda sesión, sólo que por otra vía.
- Alternativa rechazada: output style. Cambia el estilo de respuesta, no el rol, y no se probó que
  tenga precedencia controlable frente a `CLAUDE.md` (no verificado).
- Alternativa rechazada: plugin de Claude Code. Agrega un canal de distribución paralelo al
  instalador; útil quizá a futuro, fuera de alcance.
- Alternativa rechazada: hooks (`SessionStart`) que inyecten el coordinador. Activación implícita,
  difícil de auditar y de revertir.

### D4. Módulos por acción

El módulo TDD pasa a la acción apply (constructor), governance a propose, la parte OPSX de Engram
al coordinador. Cómo se cargan es G3. Opciones:
(a) el coordinador conserva los módulos en su prompt y los pega en el brief del especialista según
la acción (patrón actual de "Project Standards", `sdd-orchestrator.md:144-149`);
(b) skills instaladas por Active Stack que el especialista invoca;
(c) un archivo de agente por especialista con su módulo en el prompt.
Recomendación: (a) para este change, porque no depende de los skills `openspec-*` (ver C2) y es
verificable con testigo; (b)/(c) se coordinan con la Propuesta 04.


### D5. Migración y ciclo de vida

Orden por instalación (reflejo de `InjectPrimaryAgent`, con transacción explícita):
1. Chequear versión mínima de Claude (D7). Resultado por caso en la spec
   `claude-coordinator-agent` ("Compatibilidad por versión y entorno sin Claude").
2. Chequear colisión de nombre en el directorio de agentes del target (archivo sin marca propia →
   error).
3. Snapshot de `CLAUDE.md` y del archivo de agente si existe, en un único manifiesto bajo
   `<home>/.active-stack/backups/` también en target Project. No se usa el patrón actual de
   `config.Install`, que arma `snapshotDir` con `homeDir` (`install.go:35-36`) y en Project recibe la
   raíz del proyecto, contra el comentario de `internal/install/plan.go:138-140`. `settings.json` no
   entra al manifiesto: este change sólo lo lee.
4. Escribir el archivo de agente (atómico).
5. Purgar secciones del coordinador en `CLAUDE.md` e inyectar el bloque universal (atómico).
6. Si 4 o 5 fallan, restaurar desde el manifiesto del paso 3.
7. Si fue la primera migración, registrar la ruta del manifiesto como **manifiesto de migración**
   y el hash de `CLAUDE.md` y del archivo de agente resultantes (en `installation.json` o un archivo
   hermano; ubicación a fijar en la tarea 6.8). Reinstalaciones posteriores no lo reemplazan. El
   manifiesto de migración se crea fijado (`Pinned`, `internal/backup/manifest.go:62-64`) para que
   `Prune` (`internal/backup/retention.go:119-131`), que nunca borra respaldos fijados, no lo elimine
   si en el futuro alguien lo llama sobre ese directorio (hoy no encontramos llamadores).

**Update.** El `update` real reemplaza el binario y reaplica la selección registrada con `--sync`
(`cmd/active-stack/update.go:94`, `:159-165`), que vuelve a correr `install --headless --home`
por agente (`update.go:353-360`, sin target Project). `internal/update/state.go:30` registra
`projects`, que hoy se usan para refrescar skills. Por eso: Machine migra por el camino de install;
los proyectos registrados con instalación Project V1 no se reescriben y reciben una advertencia con
el comando de migración explícita (recomendación de G10). Un HOME V1 sin registro de selección
(`installation.json` ausente o sin agentes) no reaplica config: `syncRecordedInstall` sólo refresca
skills e imprime que hay que correr install una vez (`update.go:375-378`). Este change no cambia eso;
agrega a ese mensaje la nota BREAKING de Claude y el comando que migra, para que ese usuario se entere.

**Rollback.** El CLI ya exige `--restore-manifest <path>` con `--strategy restore`
(`cmd/active-stack/uninstall_dispatch.go:30-34`); ese flag manda. El manifiesto de migración
registrado es el que install nombra en su salida y el que documenta README para volver atrás; si el
usuario pasa otro, el restore sigue el comportamiento actual y advierte que no es el de migración.
`internal/backup/restore.go` restaura archivos enteros, así que, con cualquier manifiesto, antes de
escribir `CLAUDE.md` o el archivo de agente se compara el hash actual con el registrado en el paso 7:
si difiere (el usuario editó), no se escribe y se guarda el contenido actual en un snapshot nuevo.
La confirmación para pisar igual es un flag propio (nombre tentativo `--overwrite-edited`), no
`--yes`: `--yes` ya existe para confirmar el uninstall (`uninstall_flags_test.go:156`) y reusarlo
haría que un script con `--yes` pierda ediciones sin aviso. Uninstall sin restore borra sólo el
bloque universal y el archivo con marca propia, con snapshot y restauración si falla a mitad.

**Target Project en verify y uninstall.** Verify ya conoce Project: corre como hook posterior al
install (`cmd/active-stack/main.go:199-207`, `:256-263`) y en Project usa `projectVerifyAdapter`, que
resuelve rutas con `PathsFor(base, Project)` (`main.go:320-345`); no existe un subcomando `verify`
(subcomandos en `main.go:64-112`). Este change sólo agrega al adaptador el método del directorio de
agentes (tarea 3.1) y el check nuevo (tarea 7.4). Uninstall, en cambio, es sólo Machine por decisión
vigente: `cmd/active-stack/headless/uninstall_flags.go:50` ("No --project flag (uninstall is
machine-scope, D7/proposal)"), fijada por `uninstall_flags_test.go:179-182`. Este change no la
revierte: en Project documenta la limpieza manual (README y salida de install) y registra uninstall
Project como DEUDA para el primer change que lo necesite. La opción de revertirla queda como G12,
sin recomendación.

### D6. Permisos

`WriteFileAtomic` fuerza el modo del archivo (`internal/filemerge/writer.go:63`, `tmp.Chmod(perm)`
con `0o644` de los llamadores, `inject.go:125`, `primary_agent.go:83`, `:118`) y hace
`os.Chmod(dir, 0o755)` sobre el directorio padre (`writer.go:40`). Para `~/.claude/` eso puede
ampliar un `0700` del usuario. Ese chmod existe para reparar directorios que una versión vieja del
instalador dejó sin escritura (`writer.go:37-41`, commit 6a40f17 del 2026-05-27); un arreglo que lo
quite reintroduce ese fallo. Regla propuesta: conservar el modo del archivo existente; en el
directorio agregar como máximo el bit de escritura del dueño, nunca grupo u otros. Alcance (sólo
rutas de este change o arreglo global del writer) es G6.

### D7. Versión mínima

Se declara una versión mínima de Claude Code en el adapter. El valor concreto es `(TBD)`: hoy sólo
está verificado 2.1.318. La tarea 1.4 busca la primera versión con `--agent` y agentes de archivo.
Sin binario `claude` (CI, Docker de `e2e/`, instalar antes que Claude) la recomendación es escribir
la entrega por capas y advertir (G11).

**Modo legado después de partir el asset.** El bloque legado es la **composición legado**: la
concatenación, en el orden de hoy, de las partes universal y coordinador del asset partido más los
módulos según toggles, dentro de la sección `active-stack:sdd-orchestrator`. Como la partición sólo
reubica texto (tarea 4.1), la composición legado debe ser byte a byte igual a la de 97d7f6c para los
mismos toggles; lo protege el golden `compose_all_toggles_claude.golden` actual, que se conserva
como golden de la composición legado (tareas 4.2 y 4.3). Un update en modo legado refresca ese
bloque con la composición legado de los assets nuevos, de forma idempotente.

### D8. Oráculo de contexto efectivo

Afirmar ausencia porque el modelo no repite un testigo no prueba ausencia. Oráculo propuesto:
buscar el testigo en el registro de contexto de la transcripción de Claude Code. Evidencia del
plan-reviewer (informe `informes/propose-1-plan-reviewer.md`, sólo conteos): en transcripciones de
2.1.318 el contenido de `~/.claude/CLAUDE.md` aparece en entradas `attachment.type=instructions` y
existen entradas `attachment.type=prompt_snapshot`; `claude --help` documenta
`--system-prompt-snapshot <on|off>`. No está verificado que el cuerpo de un agente o el prompt de un
subagente aparezcan ahí (SUPOSICIÓN): la tarea 1.11 lo valida antes de usarlo. Toda aserción de
ausencia lleva un control positivo en la misma corrida. La respuesta del modelo queda como control
secundario.

### D9. Autenticación en el entorno de prueba

D0 se corrió con el `~/.claude` real del usuario. Las pruebas de contexto usan un directorio de
configuración temporal (mecanismo candidato: variable de entorno de directorio de configuración de
Claude Code, a verificar en la tarea 1.12) y la credencial llega sólo por variable de entorno ya
presente en la sesión del operador. Prohibido copiar archivos de credenciales al directorio
temporal. Los informes registran sólo nombres de variables, nunca valores. El directorio temporal
se borra al terminar y su borrado queda registrado.

## Decisiones pendientes de grupo

No se cierran acá. Las specs están escritas sobre las recomendaciones de esta tabla y cada
requisito o escenario dependiente lo marca con "(depende de Gn, recomendación)". Si el grupo elige
otra opción, el artefacto indicado vuelve al planner y al plan-reviewer.

| ID | Decisión | Opciones | Recomendación | Si se elige otra |
|---|---|---|---|---|
| G1 | Dónde se instala cada pieza | universal en `CLAUDE.md` + agente en `.claude/agents/` (Machine o Project según target) / todo en Project / todo en Machine | según target, como hoy | spec `claude-coordinator-agent` (registro y escenarios de target) |
| G2 | Cómo se invoca el coordinador y fallback | `--agent` explícito / setting `agent` opt-in escrito por el instalador / skill o comando; fallback documentado o instalado | `--agent` explícito; setting sólo manual; fallback sólo documentado | specs `claude-coordinator-agent` (invocación, fallback) y `claude-instruction-layers` (sesión normal) |
| G3 | Cómo llegan los módulos por acción al especialista | (a) brief del coordinador / (b) skills / (c) agente por especialista | (a) ahora, (b)/(c) con 04 | design D4, escenarios de módulos, tareas 4.x |
| G4 | Qué reglas necesitan los especialistas (matriz) | ver borrador abajo | borrador abajo | spec `claude-instruction-layers` (matriz) |
| G5 | ID del agente coordinador | `sdd-orchestrator` / `active-orchestrator` / otro | `sdd-orchestrator` (coincide con `h.ID`, verify y uninstall; `active-orchestrator` ya es una skill de fundación, `harnesses.yaml:92`) | escenario "Update con ID de coordinador nuevo", tareas y docs |
| G6 | Alcance del arreglo de permisos | sólo rutas de Claude de este change / arreglo global de `WriteFileAtomic` / diferirlo como DEUDA | global con tests, respetando la reparación de 6a40f17 | spec `claude-orchestrator-migration` (permisos), tarea 5.4 |
| G7 | Sección `design-mockups` | queda en `CLAUDE.md` / pasa al coordinador o módulo | queda en este change; se mueve en otro | escenario "Secciones administradas hermanas" |
| G8 | Dependencia con 02 | esperar la medición / avanzar spec y dejar la comparación como tarea | avanzar; criterios funcionales no dependen de 02, la comparación antes/después sí | tarea 8.3 |
| G9 | Exposición del coordinador como subagente (B1), sólo si 1.10 muestra que no se puede excluir | aceptar con `description` y guarda / definición fuera del directorio descubrible con lanzador que pase `--agents` en línea y `--agent` / otra vía | decidir con el resultado de 1.10; sin él no se recomienda | specs `claude-coordinator-agent` (registro, sesión normal) y D1 |
| G10 | Migración de proyectos registrados con instalación Project V1 | advertir y migrar a pedido / migrar automáticamente en `update` / ignorar | advertir con el comando de migración | escenario "Proyectos registrados con instalación Project V1", tarea 6.9 |
| G11 | Entrega cuando falta el binario `claude` | entrega por capas con advertencia / legado / no instalar el harness para Claude | entrega por capas con advertencia | escenario "Binario claude ausente" |
| G12 | Uninstall con target Project (revertiría la decisión de `uninstall_flags.go:50`) | mantener Machine-only con limpieza manual documentada y DEUDA / agregar `--project` a uninstall con RED de flag, ayuda, i18n y cambio del test `uninstall_flags_test.go:179-182` | sin recomendación; el plan actual mantiene Machine-only | requisito "Limpieza manual documentada para target Project", tarea 7.5 |

### Borrador de matriz de reglas (G4, a compartir con la Propuesta 04)

Una capa por fila. Cuando un bloque toca dos capas se parte en filas o lleva justificación.

| Origen | Regla o bloque | Capa propuesta | Estado |
|---|---|---|---|
| `sdd-orchestrator.md:3` | "Bind this to the dedicated agent only" | se elimina (lo cumple la entrega) | propuesta |
| `sdd-orchestrator.md:5-9` | Role: coordinador, no ejecutor | coordinador | propuesta |
| `sdd-orchestrator.md:13` | CLI `openspec` dueño del estado | universal | propuesta |
| `sdd-orchestrator.md:14` | Delegar, no inflar | coordinador | propuesta |
| `sdd-orchestrator.md:15` | Engram persiste contexto | universal (mínimo) | (TBD grupo) |
| `sdd-orchestrator.md:17-24` | Starting work, fundación | coordinador | propuesta |
| `sdd-orchestrator.md:26-46` | Delegation rules (toggle) | coordinador | propuesta |
| `sdd-orchestrator.md:48-108` | Workflow, comandos, manejo de pedidos | coordinador | propuesta |
| `sdd-orchestrator.md:109-165` | Patrón de lanzamiento y brief | coordinador | propuesta |
| `sdd-orchestrator.md:167-201` | Ciclo de artefactos y comandos CLI | módulo por acción (propose/apply) | (TBD 04) |
| `sdd-orchestrator.md:205-206` | NEVER guess state; NEVER crear `openspec/` a mano | universal | propuesta |
| `sdd-orchestrator.md:207-210` | NEVER bloquear, NEVER apply/propose inline, ambigüedad, `/sdd-*` | coordinador (contradice al especialista si queda universal) | propuesta |
| `sdd-orchestrator.md:212-226` | Model assignments (toggle) | coordinador | propuesta |
| `governance.md:1-13` | Niveles de autonomía por dominio | módulo propose | (TBD grupo) |
| `governance.md:14-19` | Cómo aplicarlos al decidir delegación y checkpoints | coordinador | (TBD grupo) |
| `engram-protocol.md:1-5` | Encabezado "ALWAYS ACTIVE" | universal | (TBD grupo) |
| `engram-protocol.md:6-12` | Wiring OPSX | coordinador | propuesta |
| `engram-protocol.md:13-59` | Guardado proactivo y búsqueda | universal | (TBD grupo) |
| `engram-protocol.md:60-92` | Cierre de sesión y recuperación tras compactación | coordinador (lo ejecuta el hilo que habla con el usuario) | (TBD grupo) |
| `engram-protocol.md:93-117` | Git sync de memoria | (TBD grupo) | abierto |
| `compose.go:93` | Línea `Strict TDD Mode: enabled` que la composición antepone al módulo | módulo apply (va junto con `strict-tdd.md`) | propuesta |
| `strict-tdd.md` | Módulo TDD (se declara "Apply Phase", `:1`) | módulo apply | propuesta |
| harness `permissions` | Denegaciones en settings | universal por mecanismo | sin cambio |
| sección `design-mockups` | Gates visuales | queda en `CLAUDE.md` (G7) | sin cambio |

Los rangos de `governance.md` y `engram-protocol.md` salen de sus encabezados (`governance.md:1`,
`:5`, `:14`; `engram-protocol.md:1`, `:6`, `:13`, `:48`, `:60`, `:84`, `:93`).

## Contradicciones encontradas en el código

- **C1.** `sdd-orchestrator.md:3` pide atarse a un agente dedicado, pero Claude lo entrega en el
  archivo compartido (`adapter.go:125-127`). Es el problema que motiva el change.
- **C2.** El coordinador delega en skills `openspec-apply-change` y `openspec-archive-change`
  (`sdd-orchestrator.md:100-105`), pero las skills embebidas se llaman `openspec-apply` y
  `openspec-archive` y son marcadores de posición de 4 líneas sin frontmatter
  (`assets/skills/openspec-apply/SKILL.md:3`: "PLACEHOLDER: port from ..."), no listadas en
  `harnesses.yaml`. Además `CHANGES.md:375-390` documenta un borrado de `.claude/skills/openspec-*`
  tras `openspec init`. Por eso "los módulos por acción ya existen como skills embebidas" no se
  sostiene: el origen real de esas skills en Claude queda como tarea 1.6.
- **C3.** Reusar `InjectPrimaryAgent` para Claude escribiría `agent` como objeto en
  `settings.json`, donde Claude usa `agent` para elegir el agente de sesión (D1).
- **C4.** `WriteFileAtomic` ignora el modo existente del archivo y amplía el directorio padre a
  `0o755` (D6).
- **C5.** `PurgeStaleSections` borra toda sección `active-stack:` no registrada: una sección
  universal nueva que no se registre se autodestruye en la próxima instalación (D2).
- **C6.** `CLAUDE.md:9` del repo dice que el orquestador "ya vive en tu `~/.claude/CLAUDE.md`" y
  `AGENTS.md:148-149` dice "El orquestador SDD global (coordinación, delegación, model-routing) vive
  en el `~/.claude/CLAUDE.md` del usuario"; ambas dejarán de ser ciertas (Knowledge Base Impact).
- **C7.** Uninstall no conoce el target Project, por decisión vigente (`uninstall_flags.go:50`);
  verify sí lo conoce, por `projectVerifyAdapter` (`cmd/active-stack/main.go:320-345`). El intento 1
  de este plan afirmó lo contrario sobre verify; corregido en el intento 3 (D5).
- **C8.** `config.Install` guarda el snapshot bajo la raíz del proyecto en target Project
  (`install.go:35-36`), contra `internal/install/plan.go:138-140`. Preexistente; el inyector nuevo
  no lo copia (D5 paso 3).

## Risks / Trade-offs

- [BREAKING: la sesión normal deja de coordinar] → aviso en install/update, nota en README, G2.
- [El coordinador queda expuesto como subagente de la sesión normal] → mitigaciones D1, tarea 1.10,
  G9.
- [El contenido del usuario en `CLAUDE.md` sigue llegando a especialistas] → límite documentado
  (spec), no resuelto.
- [Colisión con agentes propios del usuario en el directorio de agentes] → marca de propiedad y
  error sin sobrescribir.
- [Versión de Claude sin `--agent` o sin agentes de archivo] → versión mínima y modo legado.
- [Fallo a mitad de migración o de uninstall] → transacción con restore (D5).
- [Restore pisa ediciones posteriores del usuario] → hash registrado y confirmación (D5).
- [Permisos ampliados en `~/.claude/`] → D6, G6.
- [Arreglo global de permisos regresiona la reparación de 6a40f17] → test de directorio sin
  escritura del dueño (tarea 5.4).
- [Divergencia con la Propuesta 04 en la matriz] → matriz única con columna de estado y revisión
  conjunta antes de construir.
- [Pruebas con testigo dependen de un modelo real (costo, no determinismo)] → oráculo por registro
  de contexto (D8), testigos aleatorios, control positivo, ejecución manual o en job aparte.
- [Credenciales expuestas en el entorno de prueba] → D9.
- [Engram "ALWAYS ACTIVE" para todos vs. sólo coordinador] → G4.

## Migration Plan

1. Release N: entrega por capas para Claude detrás de la versión mínima; migración automática del
   bloque legado Machine con snapshot y manifiesto de migración (D5). Aviso BREAKING en la salida de
   install/update. Proyectos registrados: advertencia con comando de migración (G10).
2. Rollback del usuario: `active-stack uninstall --strategy restore --restore-manifest <manifiesto
   de migración>` devuelve `CLAUDE.md` byte a byte si no hubo ediciones posteriores; si las hubo,
   pide confirmación (D5). Alternativa: reinstalar la release anterior, que vuelve a inyectar el
   bloque legado; el archivo de agente queda huérfano y uninstall lo elimina por su marca.
3. Rollback del proyecto: revertir el change; OpenCode, Codex y Antigravity no se afectan.

## Open Questions

- Sintaxis exacta de frontmatter admitida por la versión mínima (por ejemplo, si existe un campo
  para precargar skills). No cambia specs salvo por B1, que tiene su propia tarea (1.10) y G9.
