# Auditoría técnica de las propuestas OPSX

10 de noviembre de 2026 · Revisión documental para el equipo

Se revisaron las diez propuestas del [roadmap](https://active-stack-roadmap.vercel.app/) contra el [informe ejecutivo](https://active-stack-executive-site.vercel.app/). Las diez tienen rama en `origin` y pasan `openspec validate --strict` en los commits enlazados. La auditoría es documental. Las pruebas de aceptación por versión, sistema operativo e interfaz quedan para la implementación.

**Decisión transversal.** El ejemplo del informe ejecutivo deja SDD activo por defecto y muestra `agent` en settings de Claude. El roadmap pide que una sesión normal conserve reglas generales y que OPSX se inicie explícitamente. Esta revisión aplica el roadmap; corregir o etiquetar ese ejemplo como histórico. Usar la misma aceptación para Claude y Codex, con mecanismos nativos distintos.

## 01 · Compatibilidad del tier estricto — Juan Cruz

[Artefactos: 9778802](https://github.com/Group-Active-IA/active-stack/tree/9778802/openspec/changes/codex-strict-compat)

- **Decisión de producto.** `read-only + on-request` es la migración documentada para uso interactivo. Los comandos permitidos por el sandbox pueden ejecutarse sin pedir aprobación; una escalación puede habilitar acciones fuera de esa restricción. Si se quiere aprobación para casi todo comando, OpenAI documenta la alternativa de confianza por proyecto sin política explícita; también desactiva su configuración local. Si se quiere impedir escalaciones, evaluar `read-only + never`. No son garantías equivalentes. Distinguir además aprobación humana y auto-review; el plan debe comprobar cuál queda efectiva o documentar que respeta la elección preexistente. [Aprobaciones y seguridad de Codex](https://learn.chatgpt.com/docs/agent-approvals-security)
- **Cambiar `design.md`, specs y tareas.** Definir cómo se trata `untrusted` fuera de la entrada administrada: perfil ajeno, override de arranque o configuración gestionada. Detectar e informar el origen cuando siga bloqueando el arranque; la migración no debería reescribir a ciegas todas esas capas. Declarar qué significa el tier en CLI interactiva, en ejecución sin interacción y frente a restricciones administradas.
- **Criterio de cierre.** Con perfil aislado: lectura y comando permitido; escritura denegada; si se admite escalación, rechazo y aprobación explícitos sobre archivo temporal; red según la política acordada. Verificar arranque con configuración efectiva y distinguir actualización con registro, sin registro y con valor legado en otra capa. Separar rollback ante fallo de la reversión posterior.

## 02 · Medición de referencia — Matías

[Artefactos: 240e1b2](https://github.com/Group-Active-IA/active-stack/tree/240e1b2/openspec/changes/opsx-baseline)

- **Cerrar el esquema antes de correr.** Codex CLI documenta uso por turno; App Server emite uso por hilo. Eso no acredita que la cifra del padre incluya o excluya todos los hijos. Calibrar ambos caminos con una ejecución directa y una delegada, correlacionar identificadores y evitar sumar padre e hijos si ya se solapan. En Claude Code, la entrada total se reconstruye sumando entrada nueva, lectura y creación de caché; mantener las tres categorías. Distinguir lo observado de lo estimado y de lo no disponible. [Codex CLI](https://learn.chatgpt.com/docs/non-interactive-mode) · [Codex App Server](https://learn.chatgpt.com/docs/app-server) · [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage)
- **Cambiar `design.md` y el spec.** Publicar un ejemplo numérico por cliente con entradas totales, caché, intentos fallidos y workers; fijar qué fuente prevalece ante discrepancias. Medir sesión normal aparte, con una sonda idéntica de arranque y señal de primera solicitud cuando exista. Los bytes de instrucciones instaladas son un inventario; no imputarlos como tokens de esa solicitud.
- **Criterio de cierre.** Versionar fixtures, rúbrica, modalidad, configuración efectiva y denominadores. La muestra inicial de tres repeticiones por celda es descriptiva. Para decidir adopción en 07/09, predefinir muestra, dispersión tolerable y margen de calidad. Usar todo el costo de la cohorte, incluidos fallos y reintentos, dividido por tareas aceptadas; con cero aceptadas, informar el total y dejar el cociente sin valor. Mantener API y suscripción separadas.

## 03 · Roles nativos en Claude — Juani

[Artefactos: 9809b25](https://github.com/Group-Active-IA/active-stack/tree/9809b25/openspec/changes/claude-native-opsx)

- **Corregir la hipótesis de activación y contexto.** `claude --agent` con un cuerpo no vacío reemplaza el prompt de sistema predeterminado; `CLAUDE.md` sigue cargándose por su vía normal. El mismo archivo de agente puede usarse como subagente. Un nombre o una descripción que diga “solo invocación explícita” no prueba exclusión. Comparar el agente de sesión con una entrada manual de skill (`disable-model-invocation: true`) si el equipo acepta ese diseño; si se necesita mantener un coordinador principal dedicado, no dar por equivalente una skill en el hilo normal o con `context: fork`. [Subagentes de Claude](https://code.claude.com/docs/en/sub-agents) · [Skills y comandos de Claude](https://code.claude.com/docs/en/skills)
- **Cambiar `design.md` D1 y los specs.** Usar la entrada explícita en la sesión actual acordada para 03/04 y revisar las restricciones que el diseño anterior asignaba al agente principal dedicado. La allowlist `Agent(tipo)` documentada para el agente que corre como hilo principal con `--agent` no se puede dar por aplicada al invocar una skill. Cerrar el inventario de reglas universales, el ámbito Machine/Project y la restauración. Comprobar que cada especialista recibe las reglas obligatorias que le correspondan. [Configuración y precedencia](https://code.claude.com/docs/en/settings) · [Subagentes de Claude](https://code.claude.com/docs/en/sub-agents)
- **Criterio de cierre.** En versión y superficie fijadas, probar una sesión normal con un pedido relacionado con OPSX, invocación explícita, intento de delegar al coordinador, especialista, módulo faltante y reanudación. Verificar reglas mediante testigos de contexto y comportamiento observable. Si se ofrece Project, incluir ciclo de retiro/restauración o declarar qué operación queda fuera.

## 04 · Roles nativos en Codex — Giuliano

[Artefactos: aa784dd](https://github.com/Group-Active-IA/active-stack/tree/aa784dd/openspec/changes/codex-native-opsx)

- **Precisar D4 y la tarea que entrega la skill.** La rama ya menciona `agents/openai.yaml` como hipótesis. La documentación vigente confirma la ruta relativa a la skill y la clave `policy.allow_implicit_invocation: false`. Esto conserva `$sdd-orchestrator` y deshabilita selección implícita de esa skill. Pasar el mecanismo de supuesto a documentado y conservar la prueba por versión y superficie. El TOML del especialista no selecciona al agente principal ni crea su propio sandbox de solo lectura: los subagentes heredan los controles de la sesión padre. No llamar “read-only” a una restricción que solo está escrita en el prompt. [Skills de Codex](https://learn.chatgpt.com/docs/build-skills) · [Subagentes de Codex](https://learn.chatgpt.com/docs/agent-configuration/subagents)
- **Cerrar G1/G3/G8/G12/G13 y D6.** Fijar soporte, versión y mecanismo efectivo por superficie. Si solo funcionan `$openspec-*`, informar qué módulos, restricciones y evidencias se pierden y qué acciones no se pueden completar. Alinear inventario de módulos con 03/05/06; actualizar la referencia de que 03 no existe. Proteger la reversión de la migración frente a reinstalaciones si se promete rollback después de actualizar otra vez.
- **Criterio de cierre.** Sesión normal no activa el coordinador; invocación explícita carga los módulos correctos; versión sin roles o módulo obligatorio ausente informa el límite; colisión con skill o rol del usuario no borra nada. Probar instalación, actualización, reinstalación y reversión sobre contenido propio, incluidos markers heredados y sección de mockups.

## 05 · Lectura dirigida y herramientas por tarea — Emiliano

[Artefactos: b74fd76](https://github.com/Group-Active-IA/active-stack/tree/b74fd76/openspec/changes/opsx-task-context)

- **Corregir D8/D9.** Un agente no recupera texto que el host ya descartó. Separar captura antes del truncamiento, persistencia autorizada y resumen. Para el piloto, usar un mecanismo existente o un capturador externo en fixtures, sin añadir un subcomando al instalador. Si el original no quedó disponible, mostrar estado parcial o sin verificar. El manejo de salidas grandes de un hook de Codex no acredita que toda salida de shell se conserve. [Hooks de Codex](https://learn.chatgpt.com/docs/hooks)
- **Corregir D14 y el alcance de herramientas.** Un piso de permisos para rutas no detecta secretos impresos por un comando permitido. Definir con 06 qué se captura, quién accede, qué se redacta y cuándo vence. Diferenciar “la guía pide no usar esta herramienta”, “no se carga su descripción” y “el cliente no permite invocarla”. Ajustar los requisitos por superficie; la instalación de una skill o MCP no prueba aislamiento por tarea. [Skills de Codex](https://learn.chatgpt.com/docs/build-skills) · [Herramientas y MCP en subagentes de Claude](https://code.claude.com/docs/en/sub-agents)
- **Criterio de cierre.** Fixtures con fallo al inicio, medio y final, salida por stderr, falso positivo y original perdido; comprobar recuperación por identificador, código de salida y referencias. Tratar 300 líneas como valor inicial medible, con una excepción justificada cuando los rangos no bastan.

## 06 · Evidencia ligada a la ejecución — Gianfranco

[Artefactos: f3f005e](https://github.com/Group-Active-IA/active-stack/tree/f3f005e/openspec/changes/opsx-verifiable-evidence)

- **Corregir D2–D4 y el spec.** Huella más relato del worker permite comprobar integridad del archivo, no autenticidad de la ejecución. Un recibo autoconsistente inventado pasa el chequeo actual. Separar integridad, procedencia y suficiencia. Para aceptar procedencia, exigir un capturador bajo control del coordinador/revisor, una traza confiable con cobertura demostrada o una reejecución independiente. Registrar revisión, huella del estado probado si hay cambios sin commit, directorio, ID de ejecución, comando, resultado y artefacto. HEAD más un booleano “dirty” no distingue dos estados de trabajo diferentes.
- **Corregir la tarea 4.3.** Con `git hash-object --no-filters` LF y CRLF deben dar huellas distintas si cambian los bytes. Si se quiere comparar texto normalizado, definir una segunda huella con algoritmo explícito y conservar la de los bytes originales. [Git: hash-object](https://git-scm.com/docs/git-hash-object)
- **Cerrar D5/D6/D8 con 04, 05 y 07.** Definir el mismo contrato para ejecución directa y delegada cuando una acción requiere evidencia; que exista un paso de revisión de recibos en ambas variantes. Acordar qué campos pueden viajar a Git, incluidas rutas y argumentos sensibles. Alinear cuándo se redacta, quién conserva el original y cómo se verifica después de vencer. Los hooks pueden complementar la evidencia; no son un registro exhaustivo. [Hooks de Codex](https://learn.chatgpt.com/docs/hooks) · [Hooks de Claude](https://code.claude.com/docs/en/hooks)
- **Criterio de cierre.** Recibo falso con hash correcto, mismo HEAD con diff distinto, salida redactada con referencia válida, original vencido, cero recibos y tests verdes sin cubrir el requisito. Ninguno de los casos insuficientes puede terminar en verificado.

## 07 · Delegación selectiva — Neyen

[Artefactos: 0597ebc](https://github.com/Group-Active-IA/active-stack/tree/0597ebc/openspec/changes/opsx-selective-delegation)

- **Corregir D1, D2 y el requisito de decisión.** D1 permite que una tarea breve o un conflicto de archivos anule S4, que exige separar a quien construye de quien verifica. Dar prioridad a la independencia obligatoria. Si falta el módulo o un revisor habilitado, detener esa verificación y explicar el motivo. Si hay contención, serializar los encargos o redefinir su alcance. Añadir el caso de una verificación breve con separación obligatoria y el caso con delegación desactivada.
- **Completar D3–D5 y los escenarios de propiedad.** El diff global de un árbol compartido puede atribuirle al worker cambios preexistentes del usuario o de otro encargo. Comparar cada retorno con un punto de partida identificado y su propio árbol de trabajo. Incluir cambios preparados, sin preparar, confirmados, renombrados y archivos nuevos. Antes de integrar, comprobar alcance y volver a ejecutar los escenarios que dependan de combinar cambios. La detección posterior de una escritura fuera de alcance se complementa con los controles preventivos de 08 donde sean obligatorios.
- **Alinear D8 con 02, 06 y 09.** 02 ya está publicado. Usar la misma tarea, criterios y controles en ambos brazos, con resultados por cliente e interfaz. Si se exige revisión independiente, conservarla en ambos brazos o excluir ese caso del contraste directo. Sumar costos de API a partir de tarifas y consumo medido. Informar cuota, tokens y latencia como medidas separadas. Tratar las tres corridas y el 10 % sugerido como propuesta de calibración, con umbral y tamaño de muestra acordados antes de la prueba de adopción.
- **Criterio de cierre.** Probar verificación breve que exige independencia, conflicto entre encargos, archivo preexistente ajeno al worker y cambio confirmado fuera de alcance. Un retorno con pruebas sin corroboración sigue sin verificar. El piloto debe distinguir ahorro de costo y reducción del tiempo de espera.

## 08 · Controles efectivos por cliente — Gianfranco

[Artefactos: f3ba5d3](https://github.com/Group-Active-IA/active-stack/tree/f3ba5d3/openspec/changes/opsx-native-controls)

- **Corregir D6 y la matriz.** Codex documenta que ciertos errores, timeouts o respuestas inválidas de `PreToolUse` pueden dejar correr la herramienta. Un hook nuevo o modificado no gestionado se omite hasta que se confía en su definición; la capa de proyecto también depende de la confianza del repositorio. En Claude Code, `onFailure: "block"` está documentado para hooks `command` o `http` desde v2.1.295; no se extiende automáticamente a todos los handlers, ni bloquea una ruta que no dispara el hook. Usar permisos o sandbox para el piso obligatorio cuando puedan aplicarlo; marcar sin cobertura los caminos que no se puedan proteger. [Hooks de Codex](https://learn.chatgpt.com/docs/hooks) · [Hooks de Claude Code](https://code.claude.com/docs/en/hooks#block-the-action-when-a-hook-fails)
- **Corregir D8/D9.** Distinguir el piloto aislado de la decisión de integrar configuración permanente. Diseñar propiedad y reversión por entrada, precedencia frente a reglas del usuario y controles gestionados, y comportamiento de Bypass. Mostrar el diff real antes de escribir. Probar tanto la preservación de preferencias como el comportamiento efectivo de las reglas. [Permisos de Claude Code](https://code.claude.com/docs/en/permissions)
- **Criterio de cierre.** Cubrir herramienta, evento, versión, shell y superficie; incluir no confiado, redefinido, deshabilitado, no coincidente, timeout, error de arranque y JSON inválido. Confirmar el efecto sobre un archivo testigo y distinguir bloqueo previo de aviso posterior.

## 09 · Pilotos de modelos y caché — Matías

[Artefactos: 89f1fa5](https://github.com/Group-Active-IA/active-stack/tree/89f1fa5/openspec/changes/opsx-model-cache-pilots)

- **Mantener los dos experimentos separados; cerrar D2/D3.** Modelo por rol y esfuerzo son tratamientos distintos. Registrar modelo y esfuerzo efectivos, rerouting y fallback. Si un brazo no se puede fijar u observar, informar el límite o evaluar el paquete entero sin atribuirle el efecto a una sola variable. [Subagentes de Codex](https://learn.chatgpt.com/docs/agent-configuration/subagents) · [Subagentes de Claude](https://code.claude.com/docs/en/sub-agents)
- **Endurecer el piloto de caché.** Reiniciar sesión y checkout no garantiza caché fría: clasificar por lectura observada, conservar el estado previsto y contabilizar la preparación. Usar señales nativas por cliente y modalidad. En Claude API contabilizar también la creación de caché y el TTL elegido; en OpenAI medir entrada cacheada con su señal de uso. El orden de archivos puede cambiar lo que el host recibe, pero no prueba qué prefijo efectivo reutilizó. Separar contexto lógico, latencia, costo de API y cuota. [Caché de OpenAI API](https://developers.openai.com/api/docs/guides/prompt-caching) · [Caché de Claude API](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
- **Criterio de cierre.** Fijar antes de correr cobertura mínima, umbral de mejora, margen de calidad, presupuesto y regla para resultados inconclusos. Verificar un bloque frío/caliente y un caso de preparación sin lectura efectiva. No recomendar adopción en una cohorte sin señal ni convertir tokens en cuota de suscripción. El piloto usa el contrato de delegación de 07 una vez revisado.

## 10 · Juez externo en sombra — Neyen

[Artefactos: 27bec41](https://github.com/Group-Active-IA/active-stack/tree/27bec41/openspec/changes/opsx-shadow-review)

- **Corregir D2 y D4 y la tarea 1.3.** Un cambio sin fixes posteriores conocidos no acredita que esté limpio. Usar controles inspeccionados y fixtures limpios. Mantener las etiquetas congeladas para el análisis principal y arbitrar aparte hallazgos nuevos con revisión ciega respecto del origen. Separar defectos nuevos confirmados, falsas alarmas y casos indeterminados. Excluir de los paquetes históricos el fix posterior y cualquier etiqueta o pista que revele la respuesta.
- **Precisar D3, el requisito de sombra y la tarea 4.2.** Guardar fuera del árbol y calcular una huella permite comprobar integridad. Para mantener la sombra hace falta restringir acceso hasta el cierre definitivo del caso y definir quién puede abrir los resultados. Unificar el momento de apertura entre diseño y spec. El primer piloto se limita a históricos y casos preparados, según la decisión confirmada. Si se agregan cambios vivos, probar que el equipo que decide la aceptación no pueda ver hallazgos prematuros.
- **Precisar D5–D7.** Examinar todo el paquete autorizado, incluso diffs públicos, antes de enviarlo. Verificar retención y uso para entrenamiento como condiciones separadas. 02 y 09 ya tienen rama. Aprovechar su protocolo de medición y calibrar los candidatos de revisión con tareas propias. Jev queda condicionado a documentación y criterios de entrada. Fijar una unidad para la señal adicional, una para falsas alarmas y un presupuesto por caso. El texto alterna cantidad de casos y porcentaje de defectos sin definir cómo resolver diferencias.
- **Criterio de cierre.** Probar controles limpios, un defecto válido ausente de las etiquetas, un diff público con un secreto sembrado y acceso anticipado al resultado. Informar precisión, exhaustividad, costo y tiempo humano de triage. Un resultado favorable abre una decisión de arquitectura posterior. La aceptación de los cambios sigue su flujo habitual durante el piloto.

## Orden recomendado para la revisión del equipo

1. Trasladar el contrato de Estricto a 01 y concretar el esquema de medición de terminal e IDE en 02.
2. Revisar 03/04 en conjunto: activación explícita, reglas por rol, matriz de soporte y reversión.
3. Resolver 05/06 con un caso común: comando con un fallo fuera de la vista, original recuperable y procedencia independiente. Acordar qué pasa cuando falta cualquiera de los tres.
4. Incorporar en 08 la decisión sobre controles obligatorios y definir cómo se detecta su disponibilidad efectiva por cliente, versión y superficie. Aplicar el protocolo de 02 a 07 y 09 antes de sacar conclusiones. Usar 10 solo en sombra.

**Decisión de producto confirmada para 01.** Estricto permite ejecutar comandos dentro del sandbox de solo lectura sin pedir aprobación por cada comando, y pide aprobación para escalar cuando esa escalación está disponible. Quedan por comprobar el comportamiento efectivo y la matriz de soporte. Las demás opciones de 01 se conservan como alternativas evaluadas.

**Decisión de producto confirmada sobre controles obligatorios.** Si OPSX no puede aplicar un control obligatorio, debe detener solo la acción que depende de él, explicar qué protección falta y continuar con las tareas que sí puede realizar. La acción se puede retomar con una protección equivalente comprobada o en una superficie compatible. Si falta un control opcional, el flujo puede seguir con un aviso. La ficha 08 debe cubrir tres casos de aceptación: control obligatorio ausente, protección equivalente disponible y control opcional ausente. Las fichas 03 y 04 deben indicar cómo se comunica ese límite en cada superficie.

## Decisiones confirmadas para 02–06

| Fichas | Decisión | Cómo incorporarla |
|---|---|---|
| 02 | Incluir terminal e IDE en la primera medición, con métricas separadas | Registrar cliente, interfaz real, versión y modalidad por ejecución. Publicar resultados y cobertura por separado. Una terminal integrada sigue siendo CLI. Si una métrica del IDE falta, marcarla como no disponible. Escritorio queda fuera de la primera medición. |
| 03 y 04 | Activar OPSX explícitamente dentro de la sesión actual | Entregar una skill invocable como comando. En Claude Code, proponer `/sdd-orchestrator` con `disable-model-invocation: true`. En Codex, proponer `$sdd-orchestrator` con `allow_implicit_invocation: false`. Después de invocarlo, la skill guía el flujo y carga sus módulos. Comprobar invocación y continuidad de contexto por interfaz. [Claude Code](https://code.claude.com/docs/en/skills) · [Codex](https://learn.chatgpt.com/docs/build-skills) |
| 05 y 06 | Un reporte del worker por sí solo queda sin verificar | Conservar el resultado declarado. Marcarlo como verificado solo cuando se corrobore la ejecución y el estado del código por una fuente independiente o una reejecución, y la evidencia alcance para el requisito. |

Las decisiones y correcciones se consolidan en este informe antes de incorporarlas en los artefactos de cada propuesta.

## Decisiones confirmadas para los pilotos 07, 09 y 10

| Ficha | Decisión | Cómo incorporarla |
|---|---|---|
| 07 · Delegación selectiva | Permitir hasta dos agentes con escritura en paralelo durante el piloto, cada uno en su propio árbol de trabajo. | Paralelizar solo encargos independientes con alcances de escritura disjuntos y pruebas propias. Serializar los encargos que dependen entre sí o comparten archivos. Verificar el resultado combinado antes de integrarlo. Cuando la revisión requiera independencia, asignarla a un agente distinto del que implementó. |
| 09 · Modelos y caché | Usar la referencia de 02. Si todavía no permite estimar la variación normal, hacer primero una calibración exploratoria. | Fijar la mejora mínima, el tamaño de muestra y la regla para resultados inconclusos antes de la prueba que decida la adopción. Evaluar modelos y caché por separado y exigir que cada ajuste cumpla todos los criterios obligatorios de calidad. |
| 10 · Juez en sombra | Empezar con casos históricos y casos de prueba preparados, usando material público revisado y datos ficticios. | Excluir cambios vivos y evidencia real del usuario de la primera ronda. Acordar proveedor, paquete de datos, acceso y métricas antes de ejecutar el piloto. Evaluar en una etapa posterior si conviene incluir evidencia real redactada o cambios vivos. |

Para preparar los pilotos, quedan por fijar los valores numéricos de muestra, mejora mínima y presupuesto, y las condiciones concretas del proveedor del juez. Estos acuerdos deben quedar registrados antes de las corridas que se usarán para decidir adopción o, para el juez, antes de enviar el primer paquete.
