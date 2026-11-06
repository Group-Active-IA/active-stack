# Spec Delta

## Purpose

Define qué reglas administradas por Active Stack recibe cada consumidor de Claude Code (sesión
normal, coordinador y especialista), en qué orden prevalecen, y garantiza que el contenido propio
del usuario en `CLAUDE.md` se conserve intacto.

## ADDED Requirements

### Requirement: Oráculo de contexto efectivo independiente del modelo
Toda prueba de contexto efectivo MUST decidir presencia o ausencia de un testigo por el registro de contexto de la transcripción de Claude Code (entradas `attachment.type=instructions`, `attachment.type=prompt_snapshot` o el registro que valide la tarea 1.11), no por la respuesta del modelo, y toda aserción de ausencia MUST tener un control positivo en la misma corrida. Actor: quien prueba. Disparador: corrida con testigos. Resultado: veredicto reproducible.

#### Scenario: Ausencia con control positivo
- **WHEN** una corrida afirma que el testigo X no está en el contexto de un consumidor
- **THEN** el registro de contexto de esa misma corrida contiene otro testigo Y que debía estar, y no contiene X

#### Scenario: La respuesta del modelo no decide
- **WHEN** el modelo repite u omite un testigo en su respuesta
- **THEN** el veredicto sale sólo del registro de contexto y la respuesta queda como control secundario

### Requirement: Sesión normal conserva reglas universales sin rol de coordinador (depende de G2, recomendación)
Cuando el usuario abre Claude Code sin invocar el coordinador, la sesión MUST recibir el bloque de
reglas universales administrado y MUST NOT recibir la definición del rol coordinador ni los
módulos por acción. Actor: usuario. Disparador: `claude` sin `--agent`. Resultado: registro de
contexto de la transcripción.

#### Scenario: Sesión normal registra el testigo universal y no el del coordinador
- **WHEN** el instalador escribió un testigo único en el bloque universal y otro en la definición del coordinador, y el usuario ejecuta `claude -p` sin `--agent`
- **THEN** el registro de contexto de la transcripción contiene el testigo universal (control positivo) y no contiene el testigo del coordinador

#### Scenario: El archivo de instrucciones no contiene el rol de coordinador
- **WHEN** termina una instalación para Claude con el harness `sdd-orchestrator` en una versión compatible
- **THEN** el `CLAUDE.md` del target contiene exactamente una sección administrada de reglas universales y ninguna sección `active-stack:sdd-orchestrator`, `active-stack:sdd-delegation` ni `active-stack:sdd-model-assignments`

### Requirement: Especialista conserva las reglas que le aplican
Un subagente lanzado por el coordinador MUST recibir las reglas universales y el módulo de su
acción, y MUST NOT recibir la definición del rol coordinador. Actor: coordinador. Disparador:
lanzamiento con la herramienta Agent. Resultado: registro de contexto de la transcripción propia
del subagente.

#### Scenario: Subagente registra el testigo universal y no el del coordinador
- **WHEN** el coordinador lanza un subagente sin herramientas cuyo prompt no contiene ningún testigo
- **THEN** el registro de contexto de la transcripción propia del subagente contiene el testigo universal (control positivo) y no contiene el testigo del coordinador

#### Scenario: Especialista de apply recibe el módulo TDD (depende de G3, recomendación)
- **WHEN** el coordinador delega una acción apply con el toggle `tdd` activo
- **THEN** el contexto del especialista contiene el testigo del módulo TDD y no contiene el testigo del módulo de governance
- **AND** con G3 opción (a) este escenario prueba que el coordinador pegó el módulo en el brief (comportamiento del modelo, no del instalador) y se verifica sobre el prompt registrado del subagente; con G3 opción (c) es determinista por archivo y se reescribe

### Requirement: Contenido del usuario en CLAUDE.md llega a todos los consumidores
El sistema MUST documentar, en la ayuda de `install` y en README, que todo texto del usuario fuera
de markers en `CLAUDE.md` llega a sesión, coordinador y especialistas; el instalador MUST NOT
moverlo ni filtrarlo. Actor: usuario. Disparador: lectura de la ayuda o del README. Resultado:
texto de la advertencia presente y testigo del usuario en el contexto del especialista.

#### Scenario: Preferencia del usuario visible en el especialista
- **WHEN** el usuario tiene un testigo propio fuera de markers en `CLAUDE.md` y el coordinador lanza un subagente
- **THEN** el registro de contexto del subagente contiene el testigo del usuario

#### Scenario: Advertencia en la ayuda de install
- **WHEN** el usuario ejecuta `active-stack install --help`
- **THEN** la salida contiene la advertencia de que el contenido propio de `CLAUDE.md` llega a todos los agentes, en ambos idiomas soportados

### Requirement: Contenido del usuario fuera de markers se preserva byte a byte
Toda escritura del instalador sobre `CLAUDE.md` MUST dejar idéntico el contenido fuera de las
secciones administradas por Active Stack, incluidas secciones con markers ajenos o mal formados.
Actor: instalador. Disparador: install, update o uninstall. Resultado: diff fuera de markers vacío.

#### Scenario: Install preserva texto previo y posterior
- **WHEN** `CLAUDE.md` tiene texto del usuario antes y después del bloque legado y se ejecuta install
- **THEN** el texto fuera de secciones administradas es idéntico byte a byte al previo

#### Scenario: Marker mal formado no se toca y se advierte
- **WHEN** `CLAUDE.md` contiene un marker de apertura `active-stack:` sin cierre
- **THEN** el instalador no elimina ese fragmento y emite una advertencia que nombra la ruta y el ID del marker (hoy `PurgeStaleSections` lo ignora en silencio; la advertencia es comportamiento nuevo)

### Requirement: Precedencia de capas declarada y probada
El sistema MUST documentar el orden efectivo entre contenido del usuario en `CLAUDE.md`, bloque
universal, prompt del coordinador y `CLAUDE.md` global frente a proyecto, medido en las tareas 1.3
y 1.5, y el instalador MUST NOT reescribir texto del usuario para resolver un choque. Actor:
equipo y usuario. Disparador: sesión con capas en conflicto. Resultado: registro de contexto y
documentación coherentes.

#### Scenario: Choque entre regla del usuario y regla del coordinador
- **WHEN** el `CLAUDE.md` del usuario contiene una regla con testigo U que contradice una regla del coordinador con testigo C, y se ejecuta `claude -p --agent <id-coordinador>`
- **THEN** el registro de contexto contiene U y C, el texto del usuario queda sin cambios, y README documenta qué capa prevalece según el resultado de 1.3; la respuesta del modelo se registra sólo como control

#### Scenario: Agente Project y Machine con el mismo ID (depende de 1.5)
- **WHEN** existen `<home>/.claude/agents/<id>.md` con testigo M y `<root>/.claude/agents/<id>.md` con testigo P, y se ejecuta `claude -p --agent <id>` en `<root>`
- **THEN** el registro de contexto contiene sólo el testigo de la capa que 1.5 determine como prevalente, y esa regla queda documentada; si 1.5 contradice el supuesto de design, el planner ajusta este escenario

#### Scenario: CLAUDE.md global y de proyecto
- **WHEN** hay testigos distintos en `<home>/.claude/CLAUDE.md` y en el `CLAUDE.md` del proyecto
- **THEN** el registro de contexto contiene ambos y la documentación declara su orden de carga

### Requirement: Matriz de reglas publicada y trazable
Cada regla del bloque compuesto actual MUST quedar en una matriz con capa asignada (universal,
coordinador, módulo por acción o especialista), origen archivo:línea y estado de acuerdo con la
Propuesta 04. Una regla asignada a más de una capa MUST partirse o llevar justificación escrita.
Actor: equipo. Disparador: revisión del plan. Resultado: matriz sin reglas huérfanas.

#### Scenario: Ninguna regla sin capa
- **WHEN** se compara la lista de encabezados y reglas de `assets/claude/sdd-orchestrator.md`, `governance.md`, `engram-protocol.md`, `strict-tdd.md` y las líneas agregadas por la composición (por ejemplo `Strict TDD Mode: enabled`) con la matriz
- **THEN** cada regla aparece con una capa asignada o con estado `(TBD grupo)` explícito

#### Scenario: Regla en dos capas
- **WHEN** una fila de la matriz asigna una regla a dos capas
- **THEN** la fila tiene justificación escrita o la regla está partida en filas de una sola capa
