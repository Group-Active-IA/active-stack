# Design · Medición de referencia de OPSX

## Context

La propuesta parte de `1b4639b` y se actualizó sobre `main` con el 01 ya integrado (PR #5, `6e8d3b8`). En `main`, Estricto en Codex escribe `sandbox_mode = "read-only"` y `approval_policy = "on-request"`; las acciones que requieren permisos adicionales necesitan aprobación explícita cuando la interfaz admite solicitarla y quedan bloqueadas en las demás. Por eso las tareas con edición no pueden correr bajo Estricto sin agregar aprobaciones humanas que sesgan tiempo y calidad, y en ejecución no interactiva no hay revisor. En Claude, el modo `default` también pide confirmar ediciones sin ser equivalente a solo lectura. El manifiesto de corridas cita el contrato de 01 vigente al correr y registra los permisos efectivos de cada celda.

Las plantillas actuales de [Claude](../../../internal/harness/config/assets/claude/sdd-orchestrator.md) y [Codex](../../../internal/harness/config/assets/codex/sdd-orchestrator.md) incluyen asignaciones por fase y reglas de delegación. Ambas traen la misma tabla de alias de modelos de Claude. El manifiesto registra el modelo efectivo por rol y cualquier sustitución. Las comparaciones se hacen dentro de cada cliente, así que no hace falta declarar equivalencia de modelos entre clientes. La revisión de partida de los fixtures es un dato independiente del commit de esta propuesta.

## Goals / Non-Goals

El protocolo tiene que permitir que otra persona reconstruya una corrida y audite su aceptación y sus agregados. Los adaptadores se limitan a capturar señales de los clientes. Los cambios de comportamiento, modelos y caché se evaluarán como variantes de changes posteriores. La captura independiente de stdout/stderr, la redacción y la retención son de 05; las huellas y los estados de procedencia, de 06. Este change consume esos contratos y no los redefine.

## Decisions

### Recorrido propuesto

1. Cargar un manifiesto versionado con la matriz, las tareas y sus pesos, la composición de cada cohorte, la semilla de orden, los límites y la revisión de los fixtures. Hacer una comprobación previa por celda de versión, interfaz, permisos efectivos, señales disponibles y opt-ins de contenido desactivados. Una celda que no pasa queda incompatible antes de iniciar.
2. Preparar un checkout descartable y un perfil de prueba, iniciar una sesión nueva y capturar los eventos, el código de salida, el diff y los tiempos. Usar un workspace sintético y credenciales de prueba según la modalidad autorizada.
3. Evaluar con chequeos y la rúbrica común. Para criterios que requieren juicio humano, ocultar la variante cuando sea posible y registrar veredicto, criterio y evidencia. El tiempo humano suma intervención activa, como autorizaciones y correcciones, y revisión activa.
4. Normalizar las unidades y generar registros por intento, por corrida y un resumen por cohorte. Se propone JSONL para los registros y Markdown para el resumen. El arnés vive fuera del binario del instalador; la ubicación concreta y el lenguaje quedan en D6.

### Matriz de cohortes

- **Clientes e interfaces.** Claude Code y Codex, en terminal y en IDE desde la primera medición, con resultados separados por interfaz. Una terminal integrada del IDE que ejecuta la CLI pertenece a la cohorte de terminal. La interfaz registrada es la real, no la declarada. Las aplicaciones de escritorio quedan fuera de la primera comparación.
- **Plataformas.** Windows y Linux son el mínimo. La matriz distingue plataformas distribuidas, declaradas y probadas, e identifica de forma explícita las combinaciones de cliente e interfaz exigidas en cada una.
- **Métricas ausentes.** Si una señal existe en la CLI pero no en el IDE, la métrica figura como no disponible en la cohorte IDE con su motivo. La corrida se conserva para las métricas válidas.

### Permisos de medición

- El perfil Estricto de 01 se mide en una serie separada, con aprobaciones registradas como intervención humana. Donde la interfaz no puede solicitar aprobación, la acción bloqueada se registra como corrida bloqueada por permisos.
- Las tareas con edición corren en perfiles aislados con permisos equivalentes para la tarea. Equivalente significa que la tarea puede leer el fixture y escribir solo dentro del checkout descartable sin pedir aprobación por cada edición.
- El manifiesto registra por cliente e interfaz el modo de permisos y los flags efectivos, comprobados con el cliente real en la comprobación previa.
- Descontar la espera de aprobación no elimina los bloqueos, rechazos o reintentos que cambian el curso de una tarea. Una corrida bloqueada por permisos después de iniciar cuenta como corrida no aceptada.

### Tareas y calidad

| Fixture candidato | Evidencia mínima |
|---|---|
| Exploración de un flujo existente | Mapa de entrada, archivos y efectos corroborados por una clave de referencia. El checkout queda intacto. |
| Cambio pequeño | Corrección de un comportamiento acotado con una prueba de regresión y control de archivos tocados. |
| Cambio multiarchivo | Cambio de contrato pequeño entre dos componentes, con compatibilidad verificada en llamadas existentes y un caso de borde. |
| Diagnóstico de una falla sembrada | Reproducción, causa raíz y descarte de un síntoma alternativo con evidencia verificable. |
| Revisión de un diff sembrado | Hallazgos contra defectos conocidos con la regla de emparejamiento de abajo y evidencia por hallazgo. |

Congelar prompt, estado inicial, herramientas permitidas, solución o clave de referencia, condiciones de tiempo y rúbrica antes de las corridas. En los cambios de código se evalúa el comportamiento observable y el alcance. Para exploración, diagnóstico y revisión se evalúan hechos y evidencia. Un criterio obligatorio incumplido da un veredicto de rechazo por calidad. Reportar calidad por criterio y tasa de aceptación junto al consumo. Una afirmación del agente sin corroboración independiente queda sin verificar según el contrato de 06.

### Rúbrica de revisión de un diff sembrado

02 mantiene esta regla; 10 la reutiliza.

- La clave de referencia lista cada defecto sembrado con identificador, archivo, rango de líneas y mecanismo de falla.
- Un hallazgo empareja con un defecto cuando señala el mismo archivo, un rango que se solapa con el del defecto y el mismo mecanismo de falla. El revisor decide el mecanismo y registra la evidencia.
- Cada defecto empareja con un solo hallazgo y cada hallazgo con un solo defecto. Los hallazgos repetidos sobre el mismo defecto no suman.
- **Verdadero positivo**: hallazgo emparejado con un defecto sembrado.
- **Omisión**: defecto sembrado sin hallazgo emparejado.
- **Falso positivo**: hallazgo sin emparejar que el revisor rechaza.
- Un hallazgo sin emparejar que el revisor considera un defecto real no sembrado se registra como hallazgo nuevo. Un árbitro lo adjudica por separado y no cambia los conteos del caso hasta su adjudicación.

### Señales y fuentes

| Cliente e interfaz | Fuente candidata | Comprobación pendiente en la versión fijada |
|---|---|---|
| Codex terminal | `codex exec --json` documenta uso por turno con entrada, entrada cacheada, salida y razonamiento, y eventos de inicio y fin de turno. App Server documenta uso por hilo, duraciones de algunas herramientas y solicitud y resolución de aprobaciones. La sección `[otel]` de la configuración es una fuente adicional. | Si el agregado del padre incluye a los hilos hijos, con una tarea directa y otra delegada. El código de salida se guarda como dato bruto junto a los eventos `turn.failed` y `error`; no determina el veredicto. |
| Claude Code terminal | `claude -p` y el SDK entregan resultado estructurado con `duration_ms` y `duration_api_ms`. OpenTelemetry documenta contadores de entrada, salida, lectura y creación de caché, con categoría de origen para principal, subagente o auxiliar. | Atribución por worker solo desde la versión 2.1.310, con `agent_id`. Los agentes propios se reportan con `agent.name = "custom"`, así que el nombre no sirve para atribuir. El `total_tokens` del cierre de un subagente cubre solo su última solicitud. La categoría subagente puede incluir hooks basados en agentes. |
| IDE de ambos clientes | Las fuentes anteriores cuando la extensión real las expone, y eventos de permisos o exportaciones de la interfaz. | Validar cada fuente en la extensión real. Lo que solo exista en la CLI figura como no disponible en IDE. |

Fuentes oficiales para diseñar la captura. [Codex no interactivo](https://learn.chatgpt.com/docs/non-interactive-mode) · [Codex App Server](https://learn.chatgpt.com/docs/app-server) · [Claude Code programático](https://code.claude.com/docs/en/headless) · [Agent SDK de Claude](https://code.claude.com/docs/en/agent-sdk/typescript) · [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage) · [Eventos de permisos de Claude](https://code.claude.com/docs/en/hooks)

### Matriz y medidas

- **Sesión normal.** Usar una sonda breve y fija en una sesión nueva sin invocar OPSX ni solicitar herramientas. Preferir la entrada de la primera solicitud o un desglose nativo. Si solo existe un total de turno, usarlo como aproximación identificada cuando se pueda comprobar que representa una única solicitud. Registrar el alcance de la señal y el inventario de instrucciones observables. Los bytes del inventario se reportan en su propia unidad. Repetir la sonda en las instalaciones actual y modular cuando ambas estén disponibles.
- **Tareas OPSX.** Capturar primero el flujo actual con los permisos de medición definidos arriba. Al llegar 03 y 04, comparar con el flujo modular dentro de cada cliente, interfaz y modalidad, con la misma revisión del fixture, rúbrica y política de modelos, herramientas y permisos. Alternar el orden de las variantes y registrar la semilla. Si una diferencia de capacidad obliga a cambiar un control, identificarla en los resultados y separar la comparación.
- **Caché y contexto.** Entrada lógica por solicitud incluye lecturas y escrituras de caché. Separar el contexto observable en una solicitud del total de tokens de entrada procesados durante la tarea. En Claude Code la entrada documentada excluye lecturas y escrituras de caché y se suman al total lógico. En Codex la entrada cacheada forma parte de la entrada, previa confirmación con la versión fijada; la entrada sin caché se deriva por diferencia. Registrar el estado de caché previsto y el observado. Una sesión nueva puede reutilizar prefijos. Clasificar por evidencia y separar o estratificar las corridas cuyo estado no pueda confirmarse. El 09 diseñará el experimento específico de caché.
- **Lecturas y herramientas.** Registrar por intento la cantidad de lecturas de archivos y de llamadas a herramientas por tipo y rol, con la fuente y su alcance. 05 usa estas métricas para evaluar la lectura dirigida.
- **Consumo y reintentos.** Cada repetición parte de cero y puede contener varios intentos bajo un límite preacordado. Registrar eventos por coordinador o principal, worker y consumo auxiliar o sin atribuir, con IDs de sesión, intento, rol y agente cuando existan. Conservar snapshots acumulados y derivar deltas solo si el alcance es conocido. Registrar reintentos automáticos observables y reintentos explícitos por separado.

### Tiempo

El arnés mide el tiempo transcurrido con reloj monotónico desde el inicio de la tarea hasta su resultado. Registra aparte las esperas de aprobación observables, la intervención humana y los reintentos. En Codex, el inicio y el fin se marcan con los eventos de turno y las esperas con la solicitud y la resolución de aprobaciones de App Server. En Claude se contrasta la duración del SDK con los eventos de permisos que exponga la interfaz. Un evento de solicitud por sí solo no determina el tiempo de espera, y el tiempo de API cubre solo una parte del trabajo.

Solo se calcula el tiempo sin espera de aprobación si se pueden delimitar los intervalos en que la tarea quedó efectivamente detenida. Los intervalos solapados se unen y se conserva el tiempo en que otros agentes siguieron trabajando. Si la interfaz no ofrece suficiente evidencia, el valor ajustado figura como no disponible. La instrumentación se valida introduciendo una pausa conocida en una aprobación controlada y comprobando qué métricas cambian.

### Denominadores

| Caso | Momento | Tratamiento |
|---|---|---|
| Celda incompatible | Antes de iniciar, en la comprobación previa | Fuera del denominador y sin consumo. El resumen la lista aparte con su motivo. |
| Corrida bloqueada por permisos | Después de iniciar | Iniciada y no aceptada. Su consumo entra en el costo total de la cohorte. |
| Rechazo por calidad | Al evaluar | Iniciada y no aceptada. Su consumo entra en el costo total de la cohorte. |
| Timeout o fallo del cliente | Después de iniciar | Iniciada y no aceptada, con la causa registrada. |

El costo por tarea aceptada es el costo total de todas las corridas iniciadas de la cohorte, incluidos fallos, bloqueos y reintentos, dividido por las tareas aceptadas. Con cero aceptadas se publica el costo total y el cociente queda sin valor. Con cobertura incompleta de consumo, el cociente queda sin valor y se publica la cobertura. Las tareas y pesos de cada cohorte vienen del manifiesto y son los mismos para ambos brazos.

### Reconciliación de padre e hijos

Regla de precedencia ante discrepancias entre fuentes del mismo cliente:

1. El total de la corrida sale de la fuente de mayor alcance comprobado para la versión fijada, por ejemplo un agregado que se verificó que incluye a los hijos.
2. La atribución por rol sale de la fuente más granular con IDs correlacionables: eventos por solicitud o por turno.
3. Si el detalle suma menos que el total, la diferencia va a consumo sin atribuir y la atribución se marca parcial.
4. Si el detalle suma más que el total, el alcance del agregado no está comprobado. La métrica se marca parcial, se registran ambas cifras y la corrida queda fuera de los indicadores que exigen cobertura completa.

Los ejemplos usan cifras ilustrativas en tokens lógicos de una corrida delegada con un worker.

**Claude.** Eventos de OpenTelemetry por solicitud: principal 40.000 tokens; worker con `agent_id` propio, dos solicitudes de 12.000 y 8.000. El resultado del SDK informa 60.000. El cierre del worker informa `total_tokens` 8.000, que solo cubre su última solicitud.

| Cálculo | Total | Error |
|---|---|---|
| Resultado del SDK + suma del worker | 80.000 | Doble conteo de 20.000 si el resultado ya incluye al worker |
| Principal + `total_tokens` del cierre | 48.000 | Subconteo de 12.000 |
| Regla de precedencia | 60.000 | Principal 40.000, worker 20.000, sin atribuir 0 |

**Codex.** Uso por turno del hilo padre 30.000 tokens; hilo hijo 18.000. El agregado del hilo padre informa 48.000 y la comprobación confirmó que incluye al hijo.

| Cálculo | Total | Error |
|---|---|---|
| Agregado del padre + hilo hijo | 66.000 | Doble conteo de 18.000 |
| Regla de precedencia | 48.000 | Padre 30.000, hijo 18.000, sin atribuir 0 |

Si la comprobación muestra que el agregado del padre no incluye al hijo, el total es la suma de ambos hilos. Si un hilo hijo no se puede correlacionar, su consumo queda sin atribuir y la atribución se marca parcial.

### API y cuota

Agrupar por cliente, versión, interfaz, modalidad y variante, con las mismas tareas y pesos para ambos brazos. Para API, usar costo atribuible observado o una estimación identificada con modelo, tarifas fechadas, categorías disponibles y alcance. Para suscripción, registrar unidad y ventana de cuota cuando la fuente permite atribuirla a la corrida. Los tokens se reportan también como consumo técnico. Publicar cobertura y cantidades válidas, parciales y excluidas.

### Datos sensibles

La lista de datos sensibles, la captura, la redacción, el acceso y la retención son de 05. Las huellas de la evidencia y sus estados de procedencia son de 06. Como mínimo, el resultado compartido excluye correo, UUID de cuenta, identificador de usuario, rutas del host, detalle de herramientas y cuerpos de solicitudes. En corridas medidas no se activan opt-ins de contenido, como las variables `OTEL_LOG_*` de Claude Code o las opciones equivalentes de `[otel]` en Codex. La comprobación previa marca como incompatible una celda con alguno activo.

### Estructura de resultados propuesta

| Registro | Campos mínimos |
|---|---|
| Manifiesto | Versión de protocolo y rúbrica, revisión de fixture, ID y huella de prompt, revisión de Active Stack, perfil y huella de configuración sin secretos, cliente, versión, sistema y plataforma, interfaz real, modalidad, modo de permisos y flags efectivos, tier, modelo por rol y esfuerzo observable, herramientas, tareas y pesos por cohorte, semilla y límites. Es la única fuente de estos datos para 07, 09 y 10. |
| Corrida | ID de tarea, repetición, variante, IDs de intento, tiempos, estado de ejecución, aceptación por criterio, motivo de rechazo, bloqueo o incompatibilidad, intervención y evidencia revisable. |
| Intento | IDs de corrida, sesión y agente, número y causa del intento, roles y modelo observado, consumo por categoría, lecturas y llamadas a herramientas, cuota o costo si corresponde, estado de caché observado, código de salida y eventos de error, tiempos y alcance del agregado. |
| Métrica y evidencia | Valor o `null`, unidad, fuente y alcance, estado observado, estimado, parcial o no disponible, y motivo cuando corresponda. Referencia a evidencia autorizada con la huella definida por 06. |
| Resumen | Cohorte y versión de protocolo, iniciadas, aceptadas, rechazadas y bloqueadas, celdas incompatibles aparte, cobertura de señales, dispersión por tarea, agregados separados por interfaz y modalidad, y comparación emparejada cuando corresponda. |

### Registro de decisiones

| ID | Decisión | Estado | Acuerdo o propuesta | Fuente y fecha | Changes afectados |
|---|---|---|---|---|---|
| D1 | Tareas, fixtures y calidad | Parcial | Acordado: un fixture por familia sobre una revisión fija del repositorio de Active Stack, rúbrica con criterios obligatorios, y revisión de los resultados a cargo de Giuliano como responsable de 02. El árbitro de desacuerdos es un integrante distinto del revisor: (TBD · nombre a confirmar con el grupo). Los fixtures concretos se nombran al implementar la tarea 1.3, congelados antes de las corridas. Alternativa si la variabilidad lo exige: un segundo fixture en las familias con mayor variabilidad. | Responsable de 02, 2026-11-11 | 02, 03, 04, 07, 09, 10 |
| D2 | Clientes, interfaces, plataformas y versiones | Acordada | Terminal e IDE desde la primera medición, con resultados separados. Terminal integrada que ejecuta la CLI = cohorte terminal. Métricas ausentes en IDE = no disponibles. Escritorio fuera de la primera comparación. Windows y Linux como mínimo; la matriz distingue distribuidas, declaradas y probadas. El manifiesto registra cliente, versión, sistema, interfaz real, modalidad y permisos efectivos. | Informe unificado, 2026-11-10 | 01, 02, 03, 04, 07, 09 |
| D3 | Repeticiones, orden y presupuesto | Pendiente | Propuesta: una corrida de calibración fuera de la muestra y tres repeticiones medidas por tarea y celda, como referencia descriptiva. La muestra, la dispersión tolerable y el margen de calidad se fijan antes de usar la referencia para decidir adopción en 07 o 09. Montos de presupuesto, tope de cuota, tiempo humano, timeout, reintentos, regla de parada y tamaño de una ampliación: (TBD · pendiente de acuerdo). | — | 02, 07, 09 |
| D4 | Métricas y cobertura mínima | Acordada | Veredicto, revisión y tiempo en toda corrida iniciada. Una señal ausente se publica como no disponible con su motivo y la corrida se conserva para las métricas válidas. Un total incompleto queda fuera de los indicadores que exigen cobertura completa. | Informe unificado, 2026-11-10 | 02, 03, 04, 05, 07, 09 |
| D5 | API, suscripción y cuentas de prueba | Pendiente | Recomendación: cuenta aislada para medir cuota; si no hay, medir solo tokens y declarar la cuota como no disponible. Falta acordar qué cuentas se usan y quién las aporta. Cohortes de API y suscripción separadas. | — | 02, 07, 09 |
| D6 | Ubicación y lenguaje del arnés | Acordada | El arnés es externo al binario y vive en un directorio de evaluación separado dentro del repositorio de Active Stack, fuera de `cmd/` e `internal/`, escrito en Go (mismo toolchain, sin dependencias nuevas) y sin incluirse en el binario del instalador. Fixtures y resultados depurados versionados; las trazas originales quedan locales, no versionadas, con acceso y retención según 05. Plazo de borrado de originales: (TBD · se fija junto con 05). Alternativa evaluada: repositorio auxiliar, si el acceso o la retención lo exigen. | Arnés externo: CHANGES.md e informe unificado, 2026-11-10. Ubicación y lenguaje: responsable de 02, 2026-11-11 | 02, 05, 06, 07, 09 |
| D7 | Permisos para medir edición | Acordada | Estricto en una serie separada. Las tareas con edición usan perfiles aislados con permisos equivalentes y registran los permisos efectivos (modo y flags) por cliente e interfaz. | Informe unificado, 2026-11-10 | 01, 02, 07, 09 |
| D8 | Denominadores y reconciliación | Acordada | Celda incompatible fuera del denominador; bloqueada por permisos y rechazo por calidad dentro del costo de la cohorte. Costo total con fallos y reintentos / tareas aceptadas; con cero aceptadas, costo total publicado y cociente sin valor. Reconciliación padre/hijos con ejemplo por cliente y regla de precedencia. | Informe unificado, 2026-11-10 | 02, 07, 09, 10 |
| D9 | Medición del tiempo | Acordada | Reloj monotónico, esperas observables aparte, unión de intervalos solapados, ajustado no disponible si no se puede delimitar, validación con una pausa conocida. | Informe unificado, 2026-11-10 | 02, 07, 09, 10 |
| D10 | Contratos con otras fichas | Acordada | 02 mantiene arnés, manifiesto y rúbrica, incluida la de revisión de un diff sembrado. El manifiesto es la única fuente de versión, interfaz, tier y permisos y modalidad para 07, 09 y 10. 05 mantiene la lista sensible, la captura y la política de datos; 06, huellas y estados de procedencia. Opt-ins de contenido prohibidos en corridas medidas. | Informe unificado, 2026-11-10 | 02, 05, 06, 07, 09, 10 |

Las decisiones pendientes se cierran con el grupo antes de compartir la propuesta final. Al cerrarlas, registrar acuerdo, responsables, fuente y fecha en esta tabla.

## Risks / Trade-offs

- **Variabilidad de modelos y cuota** → fijar versiones cuando sea posible, registrar sustituciones y separar cambios de perfil, modelo o ventana de cuota. Informar dispersión y tamaño de muestra.
- **Métricas incompletas o mezcladas** → validar con muestras reales de eventos de cada versión fijada. Si todavía no hay muestras reales, validar con eventos sintéticos y declarar el riesgo en los resultados. Un cambio en instrumentación exige identificar la versión del protocolo.
- **Fuentes distintas por interfaz** → resultados separados por interfaz y métricas no disponibles marcadas, en vez de completar el IDE con datos de la CLI.
- **Sesgo de evaluación** → congelar la rúbrica, preservar fallos y alternar el orden. Las corridas interrumpidas después de iniciar mantienen su estado, evidencia y consumo disponible.
- **Permisos que cambian la tarea** → perfiles equivalentes para edición y Estricto en serie separada. Los bloqueos se cuentan como corridas no aceptadas.
- **Exposición de datos** → fixtures sintéticos, lista sensible de 05, opt-ins de contenido desactivados y comprobados antes de iniciar.

## Migration Plan

El arnés se incorpora por separado del instalador y se ejecuta en entornos descartables. Una vez aceptada la referencia, se conserva su versión junto a sus resultados. Si cambia la semántica de un adaptador, versionar el protocolo y repetir las celdas afectadas o etiquetarlas como incompatibles.

## UI Mockups

Decision: no (no UI)
