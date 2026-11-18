# Active Stack · Informe unificado de revisión OPSX

Borrador para cerrar decisiones con el equipo. Esta guía reúne los seis informes recibidos y los acuerdos de la auditoría técnica anterior. Cada responsable puede usar su fila para corregir proposal, design, specs y tasks. Los estados que describen los informes corresponden a las revisiones que cada uno inspeccionó, incluidas versiones locales en la auditoría de 06/08.

## Acuerdos que deben reflejar todas las propuestas

- **01.** Estricto en Codex usa solo lectura con aprobación para escalar cuando se puede solicitar. Comprobar la política efectiva y registrar quién aprueba.
- **02.** Medir terminal e IDE desde la primera etapa y publicar resultados separados. Una terminal integrada que ejecuta la CLI pertenece a la cohorte de terminal. Registrar las métricas ausentes y dejar escritorio fuera de la primera comparación.
- **03/04.** Invocación explícita en la sesión actual mediante skill. En Claude, `/sdd-orchestrator`. En Codex, `$sdd-orchestrator`. Validar activación y alternativas por interfaz.
- **05/06.** Un resultado declarado por el agente queda sin verificar hasta corroborar de forma independiente la ejecución y el estado del código. Comprobar además si el resultado alcanza para el requisito.
- **07/08.** Hasta dos agentes con escritura en árboles separados para encargos independientes. Si falta un control obligatorio, frenar solo la acción afectada y permitir retomarla con protección equivalente comprobada. Si el control es opcional, continuar con aviso.
- **09/10.** Separar modelos y caché, y calibrar si 02 todavía no da una referencia estable. El juez empieza con históricos y casos preparados usando material público revisado y datos ficticios.

## Correcciones por propuesta

| Ficha | Corrección que debe quedar escrita | Cómo comprobarla |
|---|---|---|
| 01 · Estricto | Precisar la configuración efectiva y el alcance de la migración. Detectar valores legados en capas ajenas sin sobrescribirlas. Corregir la descripción del tier y definir respaldo y restauración. | Probar lectura, escritura denegada, escalación aprobada y rechazada, valor legado en otra capa y restauración en un perfil aislado. |
| 02 · Medición | Cerrar responsable y ubicación del arnés externo, rúbrica, perfiles y manifiesto. Identificar versión, interfaz, modalidad y permisos. Definir denominador y reconciliación de consumo de padre e hijos. | Comparar tareas idénticas por cohorte. Distinguir celda incompatible antes de iniciar, intento bloqueado y rechazo de calidad. Mostrar un ejemplo numérico por cliente. |
| 03 · Claude | Reescribir el diseño de agente principal dedicado conforme a la invocación en sesión. Revisar las restricciones que dependían de `--agent`, colisiones, ámbito y restauración. | Probar sesión normal, activación manual, especialistas, reanudación y reglas obligatorias. Verificar que el retiro preserve configuración propia. |
| 04 · Codex | Alinear matriz con 03, detallar qué conserva cada fallback, integrar la acción OPSX `verify` y consumir el retorno de 07/06. Distinguir registro de instalación y recibo de ejecución. | Probar invocación y fallback por interfaz, controles ausentes, colisiones y reinstalación. |
| 05 · Contexto | Separar captura de stdout/stderr, resumen y política de datos. Capturar por arnés o fuente del cliente antes del truncamiento. Documentar qué herramientas están solo recomendadas y cuáles restringe el cliente. | Fallo perdido en la salida visible, original recuperable según política, captura ausente y contenido sensible. |
| 06 · Evidencia | Separar integridad, procedencia y suficiencia. Identificar estado exacto del árbol y ejecución. Cargar el contrato al requerir evidencia tanto en modo directo como delegado. Corregir la expectativa LF/CRLF de la huella de bytes. | Recibo inventado con huella coherente, mismo HEAD con dos diffs, original ausente o vencido y tests verdes que no cubren el requisito. |
| 07 · Delegación | Dar prioridad a la independencia obligatoria. Atribuir cambios al árbol y base de cada agente. Integrar el recibo de 06 y remitir el bloqueo preventivo a 08. | Verificación breve que exige independencia, encargos en conflicto, cambios preexistentes y cambios confirmados fuera de alcance. Volver a probar el resultado combinado. |
| 08 · Controles | Matriz por cliente, interfaz, herramienta y fallo. Comprobar qué protege realmente cada control. Definir preservación de configuración, diff sujeto a revisión y comportamiento de Bypass y de actualización desatendida. | Archivo testigo, hook ausente o inválido, control equivalente, JSON malformado intacto, reglas propias preservadas y restauración sin pisar ediciones posteriores. |
| 09 · Pilotos | Fijar el brazo de 07 y el layout modular. Registrar modelo, esfuerzo y caché efectivos. Separar cuotas, costo, calidad y tiempo. | Usar calibración de 02 para acordar muestra y umbral antes de adoptar. Si 07 aún no tiene resultado, calibrar por brazo y diferir la comparación de adopción. |
| 10 · Juez | Elegir casos limpios comprobados y arbitrar hallazgos nuevos por separado. Restringir acceso hasta el cierre del caso. Revisar el paquete y condiciones del proveedor antes del piloto. Jev depende de documentación y criterios de entrada. | Probar etiquetas incompletas, fuga de pistas del fix, acceso prematuro y material sensible. Medir defectos adicionales, falsas alarmas, costo y tiempo humano. |

## Contratos compartidos propuestos

### Guía para corregir los artefactos

**01 · Juan Cruz.** En design, identificar qué capa administra el instalador y qué hace verify si detecta un valor legado en otra capa. En proposal y tareas, incluir la descripción corta y larga del tier. En specs, cubrir lectura, escritura sin escalación, escalación aprobada o rechazada y cliente sin posibilidad de solicitarla. Verificar que un archivo inválido quede intacto y que restaurar no borre ediciones posteriores. [Configuración de Codex](https://learn.chatgpt.com/docs/agent-approvals-security)

**02 · Giuliano.** En design, documentar arnés, rúbrica, perfiles aislados y matriz cliente × versión × sistema × interfaz. En specs, distinguir celda incompatible antes de iniciar, intento bloqueado y rechazo por calidad. Contabilizar intentos fallidos y reintentos en el costo total de la cohorte y dividir por tareas aceptadas. Si hay cero aceptadas, publicar el costo total y dejar el cociente sin valor. Incluir ejemplos de agregación para detectar doble conteo entre coordinador y agentes delegados. Las tareas deben incluir una corrida de calibración por interfaz y una con aprobación de demora conocida. Ver la tabla de telemetría más abajo.

**03 · Juani.** Actualizar proposal, design y specs para que la skill explícita continúe en el hilo actual. Revisar qué restricciones dependían del arranque con `--agent`. Fijar qué reglas quedan disponibles en sesión normal, cuáles se cargan por acción y cuáles recibe cada especialista. Cubrir colisiones con archivos propios, reanudación y retiro de la instalación. [Skills de Claude Code](https://code.claude.com/docs/en/skills) · [Subagentes](https://code.claude.com/docs/en/sub-agents) · [Precedencia de configuración](https://code.claude.com/docs/en/settings)

**04 · Matías.** Precisar invocación explícita, alcance del fallback y reglas por rol en la matriz común con 03. Referenciar el sobre de 07 y los recibos de 06. Agregar escenarios de CLI e IDE para activación manual, acción `verify`, módulo obligatorio ausente y preservación de archivos propios en reinstalación. [Skills de Codex](https://learn.chatgpt.com/docs/build-skills) · [Subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents)

**05 · Emiliano.** Documentar por separado quién captura stdout/stderr, quién redacta y quién genera el resumen. El capturador independiente debe demostrar cobertura antes del truncamiento. Probar una salida con un fallo oculto, otra con datos sensibles ficticios y una sin captura recuperable. Precisar ubicación, acceso y vencimiento de la copia redactada junto con 06. Si hay diferencias de cobertura por interfaz, registrarlas. [Eventos de salida de Codex](https://learn.chatgpt.com/docs/app-server) · [Salida estructurada de Claude Code](https://code.claude.com/docs/en/headless)

**06 · Gianfranco.** Definir en el spec campos de ejecución, identificador de corrida, árbol y estado exacto probado, resultado y referencia a la copia redactada. Separar integridad, procedencia y suficiencia. En tareas, corregir LF/CRLF y probar un recibo falso pero internamente coherente, dos árboles con igual HEAD, evidencia vencida y una prueba que pasa sin cubrir el requisito. Cargar el contrato también cuando la acción con evidencia se ejecuta directamente. [Huella de archivos en Git](https://git-scm.com/docs/git-hash-object)

**07 · Neyen.** Dar prioridad a la independencia de revisión cuando sea obligatoria y definir qué ocurre si falta un revisor habilitado. Documentar base y árbol por agente, máximo de dos agentes con escritura y validación del diff antes de combinar. Probar un cambio confirmado fuera del alcance, un archivo del usuario preexistente y un conflicto entre encargos. Aplicar métricas de 02 por interfaz y verificar el resultado combinado. [Git worktree](https://git-scm.com/docs/git-worktree)

**08 · Gianfranco.** Para cada control, registrar cliente, interfaz, versión, acción cubierta y respuesta ante fallo. Separar cobertura del sandbox, permisos y hooks. Incluir control ausente, equivalente y opcional. Cubrir hook no confiado, error de arranque, timeout y configuración inválida mediante un archivo testigo. Diseñar el diff y la restauración para preservar cambios propios, incluida la actualización desatendida. [Hooks de Codex](https://learn.chatgpt.com/docs/hooks) · [Hooks de Claude Code](https://code.claude.com/docs/en/hooks)

**09 · Matías.** Definir el brazo de referencia de 07, una variante por rol y los factores que quedan fijos. Separar modelo, esfuerzo y caché. Registrar parámetros efectivos y costo de preparar la caché. Usar lecturas observadas para distinguir frío y caliente y fijar tamaño de muestra, umbral y regla de inconclusión antes de decidir adopción. [Caché de OpenAI](https://developers.openai.com/api/docs/guides/prompt-caching) · [Caché de Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)

**10 · Neyen.** Definir controles limpios inspeccionados, etiquetado y adjudicación independiente de hallazgos nuevos. Quitar pistas del fix posterior de los paquetes históricos. Restringir acceso a resultados hasta el cierre del caso y ensayar un intento de acceso prematuro. Documentar proveedor, paquete permitido y costo. Medir defectos adicionales confirmados, falsas alarmas, costo y tiempo humano de revisión. La primera ronda usa material público revisado y datos ficticios. Reutilizar la rúbrica y los criterios de medición de 02.

### Fuentes de telemetría para 02, 07, 09 y 10

| Dato | Fuente y tratamiento |
|---|---|
| Tiempo total, aprobaciones y reintentos | Reloj monotónico del arnés para el total. Para Codex, marcar inicio/fin con eventos de turno y medir esperas observables entre solicitud y respuesta en App Server. Para Claude, contrastar duración del SDK y eventos de permisos que exponga la interfaz. Un evento de solicitud por sí solo no determina el tiempo de espera. [Codex App Server](https://learn.chatgpt.com/docs/app-server) · [Claude Agent SDK](https://code.claude.com/docs/en/agent-sdk/typescript) · [Eventos de permisos de Claude](https://code.claude.com/docs/en/hooks) |
| Consumo en Codex | Capturar uso por turno en `codex exec --json` y actualizaciones por hilo donde se exponga App Server. Registrar entrada, entrada cacheada, salida y razonamiento con la definición de cada campo. Comprobar si el agregado del padre incluye a los hijos antes de sumar. [Codex CLI](https://learn.chatgpt.com/docs/non-interactive-mode) · [Codex App Server](https://learn.chatgpt.com/docs/app-server) |
| Consumo en Claude | Capturar OpenTelemetry o el resultado estructurado del SDK según la interfaz. Registrar entrada nueva, lectura y creación de caché y salida. Correlacionar por sesión y agente solo cuando la versión y la fuente lo permitan. [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage) · [Claude Agent SDK](https://code.claude.com/docs/en/agent-sdk/typescript) |
| IDE, cuota y calidad | Validar la fuente en la extensión real. Si una señal solo está disponible en CLI, marcarla como no disponible en IDE. Registrar cuota por separado cuando sea observable. Para calidad, usar la rúbrica y evidencia de 06. Para 10, agregar etiquetas adjudicadas y minutos de revisión humana. |

| Contrato | Reparto para validar con el equipo |
|---|---|
| Activación, núcleo y módulos | 03/04 mantienen juntos la matriz por cliente, interfaz, rol y acción. 05 define cuándo cargar cada módulo. 04 integra `verify` en Codex. |
| Salidas y recibos | 05 mantiene captura, acceso y política de datos. 06 mantiene procedencia, huellas y estados. 07 mantiene el sobre de retorno que incluye los recibos de 06. |
| Configuración del usuario | 01/03/08 coordinan el ciclo de respaldo y restauración y comparten revisión para cambios en `filemerge`. Separar rollback durante una operación de reversión posterior. |
| Medición | Giuliano lleva 02 y mantiene arnés, manifiesto y rúbrica. Matías lleva 04. 07/09/10 reutilizan el protocolo y declaran las diferencias de sus pilotos. |

## Acuerdos incorporados en esta revisión

- **05/06 · Retención.** Una fuente independiente captura antes del truncamiento. Por defecto se persiste solo la copia redactada, con huella de esa copia y registro de captura. Si verificar un requisito exige un detalle descartado, queda sin verificar hasta conseguir evidencia independiente suficiente. 05 y 06 deben acordar ubicación, acceso y vencimiento de la copia redactada.
- **01/02 · Permisos para medir edición.** Probar Estricto en una serie separada. Para comparar tareas de edición, usar perfiles aislados con permisos equivalentes para la tarea y registrar los permisos efectivos por cliente e interfaz. Descontar la espera de aprobación no elimina los bloqueos, rechazos o reintentos que cambian el curso de una tarea.
- **01/03/04/08 · Plataformas.** Windows y Linux son el mínimo inicial. La matriz distingue plataformas distribuidas, declaradas y probadas, con las combinaciones exigidas de cliente e interfaz identificadas de forma explícita.
- **02/04 · Responsables.** Giuliano queda a cargo de 02 y Matías de 04.

### Cómo medir el tiempo en 02

Claude Code expone métricas de actividad y el SDK informa `duration_ms` y `duration_api_ms`. El tiempo de API solo cubre una parte del trabajo. Codex CLI documenta eventos de inicio y fin de turno y consumo de tokens. Codex App Server también documenta duraciones de algunas herramientas y eventos de solicitud y resolución de aprobaciones. Estas fuentes permiten instrumentar la medición, pero no establecen un reloj común de trabajo sin espera humana para ambas terminales e interfaces IDE. [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage) · [Agent SDK de Claude](https://code.claude.com/docs/en/agent-sdk/typescript) · [Codex CLI](https://learn.chatgpt.com/docs/non-interactive-mode) · [Codex App Server](https://learn.chatgpt.com/docs/app-server)

El arnés mide el tiempo transcurrido con reloj monotónico desde el inicio de la tarea hasta su resultado. Registra aparte las esperas de aprobación observables, la intervención humana y los reintentos. Solo calcula tiempo sin espera de aprobación si puede delimitar los intervalos en que la tarea quedó efectivamente detenida. Une intervalos solapados y conserva el tiempo en que otros agentes siguieron trabajando. Si la interfaz no ofrece suficiente evidencia, el valor ajustado figura como no disponible. Para validar la instrumentación, introducir una pausa conocida en una aprobación controlada y comprobar qué métricas cambian.

## Coordinación pendiente

Las fichas pueden dejar decisiones técnicas pendientes del resultado de un cambio anterior. En cada caso, indicar de qué ficha depende, qué resultado necesita y qué comprobará cuando esté disponible. Registrar el reparto de contratos como referencia para quienes corrijan las propuestas.

- **01 → 03 y 08.** Cuando 01 termine, revisar qué parte del respaldo y la restauración se puede reutilizar. Cada ficha posterior define el ajuste que todavía necesita y lo prueba con sus propios archivos y configuraciones.
- **05 → 06.** 06 mantiene desde ahora el requisito de corroboración independiente. La elección del mecanismo de captura queda pendiente de que 05 demuestre qué fuente cubre cada interfaz antes del truncamiento.
- **06 → 07.** 07 puede definir su retorno y dejar el formato exacto del recibo pendiente de la versión que entregue 06.
- **08 → pruebas de 07 que requieren prevención.** Las pruebas de delegación que exijan impedir escrituras fuera del encargo esperan a que 08 confirme la protección efectiva.
- **07 → 09.** La prueba de adopción de 09 fija su configuración de referencia cuando 07 entregue su resultado. La calibración puede avanzar con las variantes identificadas por separado.

## Fuentes

- Revisión cruzada de propuestas OPSX, 10/11.
- Auditoría Canciani de 06 y 08, 10/11.
- Auditoría del estado de las ramas OPSX, 10/11.
- Informe de revisión de propuestas, 10/11.
- Revisión de propuestas de Juani.
- Auditoría complementaria OPSX, 10/11.
- Auditoría técnica OPSX y decisiones confirmadas en la conversación de revisión.
