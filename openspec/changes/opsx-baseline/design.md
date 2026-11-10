# Design · Medición de referencia de OPSX

## Context

La revisión de partida del borrador es `1b4639b`. La rama del 01, `propose/codex-strict-compat` en `9778802`, propone un perfil Estricto de solo lectura con escalaciones bajo solicitud y verificación con el cliente real. El manifiesto de corridas deberá fijar la revisión y el perfil finalmente aceptados. Para tareas con edición, la matriz debe registrar una ruta autorizada y comprobada dentro del fixture aislado. La equivalencia de permisos se mantiene al comparar variantes dentro de cada tarea.

Las plantillas actuales de [Claude](../../../internal/harness/config/assets/claude/sdd-orchestrator.md) y [Codex](../../../internal/harness/config/assets/codex/sdd-orchestrator.md) incluyen asignaciones por fase y reglas de delegación. El manifiesto registra su configuración efectiva y cualquier sustitución de modelo. La revisión de partida de los fixtures será un dato independiente del commit de esta propuesta.

## Goals / Non-Goals

El protocolo tiene que permitir que otra persona reconstruya una corrida y audite su aceptación y sus agregados. Los adaptadores se limitan a capturar señales de los clientes. Los cambios de comportamiento, modelos y caché se evaluarán como variantes de changes posteriores.

## Decisions

### Recorrido propuesto

1. Cargar un manifiesto versionado con la matriz, la semilla de orden, los límites y la revisión de los fixtures. Validar versiones, permisos efectivos y señales disponibles con una comprobación previa por celda.
2. Preparar un checkout descartable y un perfil de prueba, iniciar una sesión nueva y capturar los eventos, el código de salida, el diff y los tiempos. Usar un workspace sintético y credenciales de prueba según la modalidad autorizada.
3. Evaluar con chequeos y una rúbrica común. Para criterios que requieren juicio humano, ocultar la variante cuando sea posible y registrar veredicto, criterio y evidencia. El tiempo humano suma intervención activa, como autorizaciones y correcciones, y revisión activa. El tiempo de espera y el transcurrido se reportan aparte.
4. Normalizar las unidades y generar registros por intento, por corrida y un resumen por cohorte. Se propone JSONL para los registros y Markdown para el resumen. El arnés podría vivir en un directorio de evaluación separado del código del instalador o en un repositorio auxiliar si los permisos y la retención lo requieren. El equipo debe cerrar la ubicación y dependencias.

### Tareas y calidad

| Fixture candidato | Evidencia mínima |
|---|---|
| Exploración de un flujo existente | Mapa de entrada, archivos y efectos corroborados por una clave de referencia. El checkout queda intacto. |
| Cambio pequeño | Corrección de un comportamiento acotado con una prueba de regresión y control de archivos tocados. |
| Cambio multiarchivo | Cambio de contrato pequeño entre dos componentes, con compatibilidad verificada en llamadas existentes y un caso de borde. |
| Diagnóstico de una falla sembrada | Reproducción, causa raíz y descarte de un síntoma alternativo con evidencia verificable. |
| Revisión de un diff sembrado | Hallazgos contra defectos conocidos, registro de omisiones y falsos positivos y evidencia por hallazgo. |

Congelar prompt, estado inicial, herramientas permitidas, solución o clave de referencia, condiciones de tiempo y rúbrica antes de las corridas. En los cambios de código se evalúa el comportamiento observable y el alcance. Para exploración, diagnóstico y revisión se evalúan hechos y evidencia. Un criterio obligatorio incumplido da un veredicto de rechazo. Reportar calidad por criterio y tasa de aceptación junto al consumo.

### Señales candidatas

| Cliente y superficie | Fuente candidata | Comprobación pendiente |
|---|---|---|
| Codex CLI | `codex exec --json` documenta uso del turno, incluida entrada cacheada y salida. App Server documenta actualizaciones de uso por hilo y queda como fuente alternativa. | Verificar el alcance del agregado con una tarea directa y otra delegada, la correlación con hilos hijos, la entrada de la primera solicitud y la disponibilidad de cuota atribuible en cada perfil. |
| Claude Code CLI | `claude -p` entrega salida estructurada. OpenTelemetry documenta contadores de entrada, salida, lectura y escritura de caché, con una categoría de origen para principal, subagente o auxiliar. | Verificar correlación por sesión e intento en la versión fijada. La categoría subagente puede incluir hooks basados en agentes. Confirmar si se puede atribuir a un worker individual. |
| Superficies interactivas | Herramientas de inspección de contexto, exportaciones o eventos de la superficie, si están disponibles. | Acordar con 03 y 04 cuáles requieren validación. Un resultado del CLI solo representa el CLI probado. |

Fuentes oficiales para diseñar la captura. [Codex no interactivo](https://learn.chatgpt.com/docs/non-interactive-mode) · [Codex App Server](https://learn.chatgpt.com/docs/app-server) · [Claude Code programático](https://code.claude.com/docs/en/headless) · [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage)

### Matriz y medidas

- **Sesión normal.** Usar una sonda breve y fija en una sesión nueva sin invocar OPSX ni solicitar herramientas. Preferir la entrada de la primera solicitud o un desglose nativo. Si solo existe un total de turno, usarlo como aproximación identificada cuando se pueda comprobar que representa una única solicitud. Registrar el alcance de la señal y el inventario de instrucciones observables. Los bytes del inventario se reportan en su propia unidad. Repetir la sonda en las instalaciones actual y modular cuando ambas estén disponibles.
- **Tareas OPSX.** Capturar primero el flujo actual compatible con 01. Al llegar 03 y 04, comparar con el flujo modular dentro de cada cliente, superficie y modalidad, con la misma revisión del fixture, rúbrica y política de modelos, herramientas y permisos cuando sean equivalentes. Alternar el orden de las variantes en la comparación y registrar la semilla. Si una diferencia de capacidad obliga a cambiar un control, identificarla en los resultados y separar la comparación.
- **Caché y contexto.** Entrada lógica por solicitud incluye lecturas y escrituras de caché. Separar el contexto observable en una solicitud del total de tokens de entrada procesados durante la tarea. Normalizar por fuente. En Claude Code la entrada documentada excluye lecturas y escrituras y se suman al total lógico. En el agregado JSON de Codex, las lecturas cacheadas se registran como parte de la entrada, previa confirmación con la versión fijada. La entrada sin lectura cacheada se puede derivar por diferencia donde la semántica esté comprobada. Registrar el estado de caché previsto y el observado. Una sesión nueva puede reutilizar prefijos. Clasificar por evidencia y separar o estratificar las corridas cuyo estado no pueda confirmarse. El 09 diseñará el experimento específico de caché.
- **Consumo y reintentos.** La unidad de aceptación es la corrida de una tarea. Cada repetición parte de cero y puede contener varios intentos bajo un límite preacordado. Registrar eventos por coordinador o principal, worker y consumo auxiliar o sin atribuir, con IDs de sesión, intento y rol cuando existan. Conservar snapshots acumulados y derivar deltas solo si el alcance es conocido. Un agregado que ya contiene consumo hijo se usa para reconciliar el detalle, sin sumarlo de nuevo. Registrar reintentos automáticos observables y reintentos explícitos por separado.
- **API y cuota.** Agrupar por cliente, versión, superficie, modalidad y variante, manteniendo las mismas tareas y pesos para ambos brazos. Para API, usar costo atribuible observado o una estimación identificada con modelo, tarifas fechadas, categorías disponibles y alcance. Para suscripción, registrar unidad y ventana de cuota cuando la fuente permite atribuirla a la corrida. Los tokens se reportan también como consumo técnico. Para una cohorte completa, dividir el total de todas las corridas iniciadas, incluidas fallidas, por las aceptadas. Con cero aceptadas o cobertura incompleta, indicar que el cociente no se puede calcular. Publicar cobertura y cantidades válidas, parciales y excluidas.

### Estructura de resultados propuesta

| Registro | Campos mínimos |
|---|---|
| Manifiesto | Versión de protocolo y rúbrica, revisión de fixture, ID y huella de prompt, revisión de Active Stack, perfil y huella de configuración sin secretos, cliente, versión, superficie, modalidad, modelo por rol y esfuerzo observable, herramientas, semilla y límites. |
| Corrida | ID de tarea, repetición, variante, IDs de intento, tiempos, estado de ejecución, aceptación por criterio, motivo de rechazo o bloqueo, intervención y evidencia revisable. |
| Intento | IDs de corrida y sesión, número y causa del intento, roles y modelo observado, consumo por categoría, cuota o costo si corresponde, estado de caché observado, tiempos y alcance del agregado. |
| Métrica y evidencia | Valor o `null`, unidad, fuente y alcance, estado observado, estimado, parcial o no disponible, y motivo cuando corresponda. Referencia a evidencia autorizada y huella para verificar integridad. |
| Resumen | Cohorte y versión de protocolo, iniciadas, aceptadas, rechazadas y bloqueadas, cobertura de señales, dispersión por tarea, agregados separados por modalidad y comparación emparejada cuando corresponda. |

### Decisiones para acordar

Todas las filas están **pendientes de acuerdo**. La columna de propuesta sirve como punto de partida para la discusión.

| ID | Decisión | Propuesta inicial y alternativa | Changes afectados |
|---|---|---|---|
| D1 | Tareas, fixtures y calidad | Usar un fixture por familia y una rúbrica con criterios obligatorios. Acordar los fixtures concretos, quién revisa y cómo arbitrar desacuerdos. Alternativa para ampliar cobertura, agregar un segundo fixture a las familias con mayor variabilidad. | 02, 03, 04, 07, 09 |
| D2 | Clientes, superficies y versiones | Empezar con ambos CLI y fijar versiones, modelos y OS por celda. Usar la revisión aprobada de 01 para Codex. Acordar con 03 y 04 si se incorporan superficies interactivas y qué sonda es equivalente. | 01, 02, 03, 04 |
| D3 | Repeticiones, orden y presupuesto | Proponer una corrida de calibración fuera de la muestra y tres repeticiones medidas por tarea y celda para una referencia descriptiva. Acordar montos, límite de cuota, tiempo humano, timeout, reintentos, regla de parada y tamaño de una ampliación antes de correr. Alternativa, aumentar la muestra de antemano si se usará para decisiones cuantitativas. | 02, 07, 09 |
| D4 | Métricas, caché y cobertura mínima | Exigir veredicto, revisión y tiempo en toda corrida iniciada. Acordar qué cobertura de tokens, caché y atribución se exige por indicador, qué aproximación de contexto es aceptable y qué celdas se consideran comparables. Si falta una señal, publicar el motivo y conservar la corrida para las métricas válidas. | 02, 03, 04, 07, 09 |
| D5 | API, suscripción y cuentas de prueba | Acordar qué modalidades se pueden medir, quién aporta las cuentas, el método de atribución y los topes por modalidad. Mantener cohortes separadas. Si la cuota solo existe en ventanas compartidas, registrar la limitación o acordar una cuenta aislada para medirla. | 02, 07, 09 |
| D6 | Ubicación del arnés y tratamiento de evidencia | Proponer fixtures y resultados depurados versionados, con trazas originales locales de acceso restringido. Acordar ubicación, lenguaje, dependencias, responsables, redacción y plazo de borrado. Considerar repositorio auxiliar si facilita el acceso y la retención. | 02, 06, 07, 09 |

Registrar el acuerdo, responsables, revisión y motivo en esta tabla. Cerrar los puntos que cambien alcance, diseño o aceptación antes de compartir la propuesta final para revisión de los siete.

## Risks / Trade-offs

- **Variabilidad de modelos y cuota** → fijar versiones cuando sea posible, registrar sustituciones y separar cambios de perfil, modelo o ventana de cuota. Informar dispersión y tamaño de muestra.
- **Métricas incompletas o mezcladas** → validar con fixtures de eventos, conservar el alcance de la fuente y publicar cobertura por variante. Un cambio en instrumentación exige identificar la versión del protocolo.
- **Sesgo de evaluación** → congelar la rúbrica, preservar fallos y alternar el orden. Las corridas interrumpidas después de iniciar mantienen su estado, evidencia y consumo disponible.
- **Exposición de datos** → limitar la captura a fixtures sintéticos, redactar campos sensibles y restringir acceso y retención de originales según D6.

## Migration Plan

El arnés se incorpora por separado del instalador y se ejecuta en entornos descartables. Una vez aceptada la referencia, se conserva su versión junto a sus resultados. Si cambia la semántica de un adaptador, versionar el protocolo y repetir las celdas afectadas o etiquetarlas como incompatibles.

## UI Mockups

Decision: no (no UI)
