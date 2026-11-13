# Design · Codex Native OPSX

## Context

Base de código `35a89f8`. `internal/agents/codex/adapter.go` devuelve `ConfigDeliveryInstructions`, resuelve instrucciones de máquina y proyecto y usa `.agents/skills`. `internal/harness/config/install.go` despacha el bloque completo mediante `Compose` e `Inject`.

Main ya integra 01 y parte de 02. `internal/backup/restore.go` incorpora guardas de rollback en memoria, pero un manifiesto cargado posteriormente carece de esas guardas. La reversión posterior requiere identidad persistente del estado esperado. No alcanza con asumir que la corrección de 01 cubre los archivos de 04.

## Goals / Non-Goals

La entrega debe preservar reglas obligatorias, iniciar coordinación sólo por pedido explícito y mantener el contrato en ejecución directa y delegada. La matriz de soporte distingue distribución, declaración y prueba efectiva. Escritorio y cloud quedan fuera de la primera aceptación acordada; se documenta el límite.

## Decisions

### Activación y rutas

La skill conserva el hilo actual y se instala en la ruta que resuelve el adapter. Su `agents/openai.yaml` declara `policy.allow_implicit_invocation: false`. La política está documentada y debe verificarse en cada versión y superficie aceptada. Una descripción restrictiva por sí sola no satisface el requisito de activación explícita. [Metadata de skills](https://learn.chatgpt.com/docs/build-skills#optional-metadata)

El paquete contiene `SKILL.md`, `agents/openai.yaml`, módulos de apoyo y, donde sea viable, roles nativos. Se propone una estrategia de entrega propia para Codex mediante interfaz opcional de rutas. La estructura de skill administrada está confirmada. El contrato técnico compartido con 03 se alinea antes de modificar tipos o despacho. Los roles y sus permisos efectivos deben comprobarse con el cliente. Escribir “solo lectura” en el prompt no configura un sandbox. [Subagentes de Codex](https://learn.chatgpt.com/docs/agent-configuration/subagents)

### Matriz común de reglas

La fuente es `openspec/matriz-reglas-opsx.md` publicada por 03 y consultada en `ed37113`. El 04 completa únicamente su columna Codex y documenta la salvedad de evidencia directa. La columna Claude conserva el contenido de referencia. La compatibilidad de código de ambas entregas requiere seleccionar el recorrido por cliente al combinar los PR.

| Filas de la matriz común | Entrega Codex |
|---|---|
| 1, 2 | Reglas del usuario y permisos efectivos heredados de la superficie. El encargo incluye las reglas aplicables cuyo acceso del worker no esté demostrado. |
| 3, 6, 8, 15, 18, 20, 22 | Bloque universal administrado en `AGENTS.md`. Estado OpenSpec, reglas de autonomía por dominio y protocolo universal de memoria según toggles. El contenido propio queda intacto. |
| 5, 7, 9, 11–13, 16, 19, 23 | Cuerpo de la skill coordinadora, incluida la regla breve de selección de delegación de 07. |
| 10, 14, 21, 24–26 | Archivos de apoyo `delegation`, `artifacts`, `engram-opsx` y `tdd`, con el corte por toggle de la matriz. |
| 17 | Asignaciones de modelo de Claude omitidas en Codex hasta 09. |
| 27 | Módulos de contexto y captura definidos por 05. Su ubicación se acuerda con 03. |
| 28 | Contrato de evidencia de 06 requerido por la acción, incluso sin delegación. Ajuste coordinado pendiente de la fila común, que hoy lo ubica en `delegation`. |
| 29, 30 | Sobre de retorno de 07 en los encargos delegados. Gates visuales en su sección propia, conservada. |

El bloque universal puede incluir governance y memoria, conforme a la matriz común. Se retiran sólo las obligaciones de coordinación globales. Los antiguos tests que prohibían todo título de governance o Engram en la sesión normal se sustituyen por aserciones sobre reglas y capas concretas.

| Acción | Contenido requerido |
|---|---|
| explore | Reglas universales y módulos de contexto de 05 que correspondan. |
| propose | `artifacts`, reglas aplicables y módulos de contexto. |
| apply | `artifacts`, `tdd` si está habilitado y contrato de evidencia por la acción. |
| verify | Contrato de evidencia y delegación cuando exige revisor independiente. El constructor no verifica su propio trabajo en ese caso. |
| archive | `artifacts` y comprobación del estado de aceptación registrado. |
| delegar | `delegation`, encargo acotado, reglas aplicables y sobre de 07 con campo `recibos` de 06. |

Un toggle inactivo retira módulos y referencias de su alcance, pero no permite saltar un requisito obligatorio. Si verify exige independencia y no hay revisor habilitado, se detiene esa verificación con el motivo. Se conservan las acciones que sí pueden completarse.

### Evidencia y retorno

07 define el sobre `estado`, `partida`, `modificados`, `hallazgos`, `recibos`, `pendiente`. 06 define los recibos. 04 referencia ambos contratos y no crea un tercer formato. Se distingue **registro de instalación** (update y ownership) de **recibo de ejecución** (evidencia de comandos).

Aplicar directamente una tarea no elimina la obligación de evidencia ni su revisión. La integridad de un archivo no acredita por sí sola la ejecución. La procedencia y el estado exacto del código deben corroborarse mediante fuente independiente o reejecución, y la evidencia debe cubrir el requisito. La nueva revisión de 06, `f62cf00c09dc8a3aa96ddb4648801250e1a67a8c`, incorpora esas correcciones normativas en D2–D4. Se consumen como requisitos parciales inline en la skill, identificados como adaptación de 04. No se entrega `modules/evidence.md`, extracto final, esquema ejecutable o algoritmo propio para el árbol: D5 y sus acuerdos operativos siguen pendientes.

05 define captura independiente antes del truncamiento y política de datos. Su entrega `e4ec0f54536760f7fca6692e27bafcb8cd380df1`, `entrega-03-04.md` §2–3, aporta cuatro módulos literales y la tabla de acciones. Se instalan únicamente en el paquete Codex y sin toggle propio; se conservan los toggles previos para TDD, delegación y memoria. Por defecto sólo se persiste la copia redactada y su huella con registro de captura. Si falta un detalle necesario descartado, el requisito queda sin verificar. Incorporar instrucciones de captura no instala ni demuestra un capturador independiente. Ubicación, acceso y vencimiento siguen sujetos a los acuerdos de 05/06, sin fijar plazos locales.

### Consumo disponible de contexto y evidencia

`ComposeCodexPackage` entrega los módulos `directed-code-reading`, `bounded-command-output`, `task-tool-selection` y `concise-agent-reporting` con los bytes de la fuente 05 fijada. La tabla Codex usa rutas relativas `modules/`, sin `model-assignments` ni variables de ruta Claude. `apply` y `verify` requieren `bounded-command-output`; la tabla conserva requisitos condicionales de TDD, delegación y memoria, describe módulos ausentes o vacíos y exige reinvocar la skill y releer módulos después de compactación o reanudación. Una ausencia opcional permite continuar sólo si los requisitos obligatorios de esa acción siguen satisfechos.

Las reglas parciales de 06 permanecen disponibles en la skill explícita con delegación desactivada, sin volver universales las instrucciones de evidencia. Separan integridad, procedencia y suficiencia, exigen corroborar ejecución y estado exacto —incluidos preparados, sin preparar y nuevos— y distinguen SHA-256 de bytes LF/CRLF sin normalización. No se implementa un validador de recibos. Sin contrato completo requerido, captura suficiente o revisor obligatorio, se detiene sólo la acción afectada y el resultado no se verifica.

Matías amplió expresamente la autorización a registrar como deuda de entrega todo insumo o disposición que no dependa de él, incluidos 02, revisiones de otros responsables y fallos ajenos, además de 05/06/07. El registro y las condiciones de reanudación están en `informes/deudas-aceptadas-05-06-07-2026-11-13.md`; las pruebas implementadas están en `informes/apply-modulos-05-06-2026-11-13.md`. Esa aceptación no ratifica contratos propuestos ni marca mediciones o revisiones humanas como realizadas. Esas partes externas salen del checklist activo a deuda con disparador; sondas propias P2, especialistas y aceptación propia CLI/IDE siguen pendientes de Matías. No se archiva el change. No se modifica el compositor Claude ni se implementa el registro común propuesto en D23.

### Superficies y fallback

| Superficie | Objetivo de aceptación | Alternativa y límite |
|---|---|---|
| CLI Windows / Linux | Skill explícita y módulos, especialistas según versión, verify y reinstalación comprobados por separado. | Si no hay roles personalizados pero sí subagentes, usar especialista genérico con encargo completo, conservando independencia y permisos efectivos. |
| Extensión IDE Windows / Linux | Mismos escenarios en la extensión real. Una terminal integrada ejecutando CLI pertenece a CLI. | Usar el mecanismo explícito confirmado en esa extensión. Si falta una capacidad, informar exactamente la acción afectada. |
| Escritorio / cloud | Sin promesa inicial de aceptación de 04. | Informar límite y ofrecer ejecutar la acción en una superficie compatible con sus precondiciones. |

Las skills `$openspec-*` sólo ofrecen las acciones que realmente estén instaladas. No garantizan coordinación, evidencia o revisor independiente por su nombre. La alternativa verifica reglas, módulos, evidencia y controles de la acción antes de ejecutarla. Si falta un obligatorio, detiene sólo esa acción y explica la ruta o protección faltante. Una protección equivalente debe estar comprobada. La ausencia de un control opcional permite continuar con aviso. No se restaura silenciosamente el bloque coordinador global.

La versión local `0.167.0-alpha.3.2` en macOS demuestra disponibilidad del binario en este entorno. No acredita aceptación de Windows, Linux o IDE. La versión mínima, edición de extensión y flags por celda se fijan mediante P2 y la matriz del manifiesto de 02.

### Instalación y ciclo de vida

Alcance confirmado por Matías el 13/11/2026: Machine para la entrega nueva, con rutas administradas enumeradas por el adapter y colisiones detectadas antes de toda escritura. La migración de Project se difiere a una decisión posterior por la coinstalación con Antigravity y su ciclo de retiro. Las instalaciones de proyecto preexistentes requieren diagnóstico visible si siguen cargando coordinación legada; la aceptación de sesión normal limpia se realiza en perfiles sin ese legado o con su tratamiento explícitamente registrado.

Preflight valida configuración, propiedad de cada archivo y módulos disponibles. Snapshot antes de escribir. Retirar todas las ocurrencias administradas del ID `sdd-orchestrator`, incluidas heredadas; conservar contenido propio y `design-mockups`. Escribir bloque universal, skill y módulos con operaciones atómicas y propiedad reconocible. Reinstalar no reescribe bytes idénticos. Un archivo homónimo ajeno bloquea el paquete antes de migrar.

Se incluye snapshot de migración dedicado y persistente, conservado al reinstalar y actualizar, más huellas esperadas persistentes por ruta. Durante rollback se reutiliza la guarda vigente de 01 donde corresponda. Para restauración posterior, si una ruta difiere del estado esperado, preservarla y comunicar conflicto con el respaldo disponible. No retirar árboles completos que puedan contener archivos nuevos del usuario. La primera copia previa a la migración se conserva; el estado esperado por ruta se actualiza sólo tras una escritura administrada exitosa. El historial identifica cada operación y las rutas creadas. Los manifiestos anteriores sin estado esperado requieren un conflicto explícito y recuperación revisada. La visibilidad del respaldo se integra en los mecanismos existentes.

Update con registro de instalación reaplica mediante el mismo plan. Sin registro, informa adopción explícita y no migra silenciosamente. Verify de instalación comprueba estructura y propiedad; la aceptación con cliente confirma activación y carga efectiva. Uninstall retira sólo entradas y archivos propios y elimina directorios sólo si quedan vacíos.

## Decisiones confirmadas y condiciones pendientes

Matías confirmó las tres recomendaciones en esta conversación el 13 de noviembre de 2026.

| Tema | Decisión confirmada | Condición técnica que conserva |
|---|---|---|
| P1 · Alcance | Machine inicial y módulos dentro de la skill. Project se difiere. | Alinear con 03 los IDs de módulos, el contrato del adapter y la fila de evidencia directa en la matriz. |
| P2 · Compatibilidad | CLI y extensión IDE en Windows/Linux como matriz objetivo. | Sondas aisladas antes de afirmar versión mínima y capacidades por celda. macOS local sirve para preparación. |
| P3 · Reversión | Snapshot persistente con estado esperado por ruta, preservación de ediciones posteriores y conservación de la primera copia previa. | Revisar plan de escritura y probar restauración de un manifiesto recargado, reinicios y actualizaciones sucesivas. |
| P4 · Secuencia | Avanzar composición y migración independientes. La evidencia operativa conserva dependencias de 05/06/07. | Consumo disponible de 05 `e4ec0f5` y normas corregidas de 06 `f62cf00`; referencia de 07 `3fb24ee`. Contrato final, captura y acuerdos de retorno pendientes registrados como deuda autorizada. La base compartida de 03 ya se incorporó desde main. |

La aceptación completa exige las pruebas funcionales y las dependencias aplicables. Una tarea independiente puede avanzar con su contrato definido; una acción que necesita evidencia o protección ausente queda detenida con el límite informado.

## Alcance de producción aprobado para apply

### Rutas y componentes

| Área | Cambio previsto | Aceptación necesaria |
|---|---|---|
| `internal/model/harness.go`, `internal/agents/codex/adapter.go`, contratos de `internal/harness/config/` | Declarar entrega de skill administrada y resolver instrucciones, skill, módulos y roles mediante adapter. Acordar interfaz compartida con 03. | Valor cero compatible, rutas Machine, paths con espacios y Windows y regresiones de los demás adapters. |
| `internal/harness/config/assets/codex/` y composición | Skill con `agents/openai.yaml`, política explícita, bloque universal y archivos de apoyo por acción/toggle. | Sesión normal, activación manual, módulos requeridos, toggles, continuidad y mapa de reglas. |
| `internal/harness/config/`, `internal/install/` | Preflight, ownership, snapshot previo y migración de markers actuales y heredados. | Colisión o configuración inválida sin escritura; contenido propio y sección visual conservados; reinstalación idéntica; fallo parcial recuperable. |
| `internal/backup/` y registro de migración | Persistir identidad, primera copia previa, estado esperado por ruta y estado de cada operación. Reutilizar backup existente y añadir el mínimo soporte necesario. | Snapshot original conservado tras actualización doble; manifiesto recargado; archivo nuevo, editado o eliminado por usuario preservado como conflicto; fallo de escritura consistente y reintento seguro. |
| `internal/verify/harness_checks.go`, `internal/uninstall/`, integración de update | Chequeo estructural, retiro por archivo propio y replay según registro de instalación. Revisar callbacks de `cmd/active-stack/` sólo donde el recorrido lo requiera. | Componentes ausentes detectados; notas propias conservadas; update sin registro exige adopción explícita; directorios sólo se eliminan vacíos. |
| Interfaces de respaldo existentes y documentación | Exponer respaldo y conflictos mediante mecanismos existentes. Actualizar README, arquitectura y AGENTS/CLAUDE en espejo. | Recuperación identificable y límites por interfaz/target. Si una modificación visual fuera necesaria, revisar su alcance y política de diseño antes de esa tarea. |

### Escrituras sobre el perfil destino

Las rutas exactas se obtienen del adapter y del sistema de backups, y se muestran en el plan previo.

- Archivo de instrucciones de máquina de Codex, normalmente `~/.codex/AGENTS.md`: modificar sólo secciones administradas y conservar el contenido del usuario.
- Directorio de skill resuelto por el adapter, normalmente `~/.agents/skills/sdd-orchestrator/`: escribir sólo archivos propios inventariados, incluida política y módulos. Una colisión se detecta antes de migrar.
- Snapshot y registro de migración en el almacén de respaldo del instalador: conservar primera copia y persistir estado esperado, ownership y operación.
- Roles de Codex y su registro sólo si la sonda determina el mecanismo compatible. Cualquier cambio de `config.toml` necesita enumeración exacta de claves, respaldo y revisión coordinada con 01/08 antes de esa tarea.

### Orden de ejecución y límites de aceptación

1. Alinear contrato y matriz con 03 y realizar sondas por celda. Registrar qué capacidades están demostradas.
2. Implementar composición, rutas y preflight con safety net y TDD. Implementar snapshot persistente y migración después de revisar este plan de escritura.
3. Integrar update, verify de instalación y retiro. Probar fallos, reintentos, reinstalación y ediciones posteriores.
4. Integrar evidencia y revisión independiente cuando 05/06 entreguen contratos y captura suficientes. Hasta entonces, preservar explícitamente el estado sin verificar y detener la acción afectada.
5. Completar aceptación CLI/IDE Windows/Linux y comparación modular con 02. Revisar en conjunto con 03 antes de promover la migración.

Matías aprobó este alcance y autorizó iniciar las tareas independientes el 13/11/2026. Las comprobaciones pendientes de producción y las dependencias externas permanecen identificadas.

## Risks / Trade-offs

- Cambios de comportamiento entre interfaces requieren evidencia por celda y un límite visible.
- Reglas universales demasiado reducidas pueden perder governance o memoria. Comparar contra filas concretas de la matriz compartida.
- Una skill descubierta por otro agente puede colisionar con archivos propios. Preflight y ownership se prueban antes de migrar.
- Restaurar un snapshot antiguo puede borrar ediciones nuevas. Persistir estado esperado y preservar conflictos.

## Migration Plan

Resolver pendientes aplicables, aprobar el diff de escritura, caracterizar consumidores y seguir TDD. Validar instalación limpia, legado, update, reinstalación, rollback y retiro. Comparar la sesión normal y coordinada con el protocolo de 02 por interfaz. Promover la migración después de revisión conjunta 03/04.

## UI Mockups

Decision: no (no UI)

## Integración de la base tras PR #8

El 13/11/2026 se incorporó main `7a600895e8f1fe1f3728cffa1fcb57dca5663a59`, que ya contiene la entrega Claude `01f944c`, mediante merge sobre la rama publicada de 04. No se implementaron funciones nuevas de 03. El enum `ConfigDeliveryManagedSkill` permanece único y conserva su valor 2; los valores anteriores siguen siendo 0 y 1.

La interfaz común de configuración conserva la compatibilidad con adapters sin `SkillsDir`: cada recorrido nativo comprueba localmente esa capacidad. `Install` conserva el wrapper a `InstallWith`; sus opciones de versión, composición, home y shadowing siguen controlando solo la entrega Claude. Codex Machine usa su compositor, inventario y snapshot persistente, mientras Project conserva el recorrido legado y su aviso. El resultado combina `Modes`, `Warnings` y `ManagedWrites`, incluidas las salidas parciales ante error de un adapter posterior.

El despacho de verificación y retiro y la enumeración de rutas de snapshot seleccionan el cliente; los markers siguen siendo propios de cada uno. Los avisos de entrega Claude no se emiten para Codex por compartir el modo `skill`. La conciliación técnica y sus pruebas están en `informes/integracion-main-pr8-2026-11-13.md`; no cierran evidencia directa de 06 ni aceptación humana 03/04 o CLI/IDE real.

Los cambios propuestos de visibilidad de respaldo reutilizan interfaces existentes. Si P3 incorpora una modificación visual, aplicar la política de diseño del proyecto antes de esa tarea.

## Seguimiento vigente por responsable

El alcance de entrega parcial y la aceptación funcional completa son distintos. El [checklist activo](tasks.md) sigue 20 unidades propias, 17 entregadas y 3 pendientes. Las unidades `a` describen solo partes probadas o paquetes de revisión preparados; no reemplazan el comportamiento original. La [trazabilidad original](informes/trazabilidad-alcance-y-deuda-04-2026-11-13.md) conserva las 21 tareas y su estado previo 11/21.

La deuda externa queda fuera de los checks activos: D04-0201 para medición de 02; D04-0301 para disposición de matriz/03; D04-010308 para revisión compartida de snapshot; D04-BASE01 para fallos preexistentes de main; D04-REV01 para disposición humana del paquete/PR; y los IDs previos de 05/06/07. Todos tienen proveedor, disparador, criterio y riesgo en el [registro vigente](informes/deudas-aceptadas-05-06-07-2026-11-13.md). Matías retoma la integración consumidora cuando llega el insumo. Si aparecen defectos propios, vuelven al trabajo activo, no se atribuyen al proveedor por conveniencia.

Las tres obligaciones propias restantes son P2 (1.2), especialistas y rutas/encargo (2.4) y ejecución de aceptación real de Codex (4.2), separando los casos bloqueados por una deuda externa concreta. No se promete soporte funcional probado en celdas no ejecutadas. Los requisitos completos de las specs conservan su criterio de aceptación; el informe de entrega indica qué parte se comprobó y qué depende de otra disposición. Publicación, aprobación humana, merge y archivo son acciones diferentes; este ajuste no ejecuta ninguna de ellas.
