# Proposal

> Ficha 05 del roadmap · `opsx-task-context` · Emiliano · Ambos clientes · P1
> Base: `1b4639b`. **Depende de 03 (`claude-native-opsx`) y 04 (`codex-native-opsx`). Habilita 07.**
> Hermana de 06 (`opsx-verifiable-evidence`): ninguna depende de la otra.
> Revisado el 2026-11-13 contra el informe unificado de revisión OPSX (11/11) y la matriz común de
> reglas de 03 y 04 (12/11). Próxima ronda: **apply**; `tasks.md` marca qué se implementa ya y qué
> queda bloqueado y por qué.

## Clasificación

**MEDIO.** El roadmap marca riesgo medio y la gobernanza del repo (`CLAUDE.md` §4) lo confirma: las tareas tocan `internal/harness/config` y `internal/catalog`, que son MEDIO y BAJO, y el arnés de evaluación en `eval/taskcontext/`, fuera del binario. Esta ficha **no** migra configuración del usuario —eso es de 03 y 04, que por eso son ALTO— ni toca `backup`, `filemerge` ni `pipeline`; si una tarea descubriera que los necesita, se detiene y espera revisión humana. Por ser MEDIO: se implementa con checkpoints y las decisiones no obvias se suben a revisión, sin aprobación humana fechada previa a `apply`.

Lo único que roza datos del usuario es la retención de la copia redactada de las salidas, y por eso queda fuera del repositorio, en la caché del usuario (D9), propuesta a la 06.

## Why

El bloque del orquestador se compone una vez en la instalación con los cinco toggles activos y se inyecta completo en `CLAUDE.md`/`AGENTS.md`, así que toda sesión arranca cargando instrucciones de TDD, memoria, delegación, routing y gobernanza aunque la acción activa no use ninguna. A eso se suma que una lectura amplia de archivos o una salida de comando extensa pueden desplazar del contexto el fallo que importaba, sin que quede registro de que se perdió. Y en el caso de las salidas el problema empieza antes del agente: los dos clientes recortan la salida de un comando largo antes de entregarla al modelo —medido el 13/11: Claude Code deja unos 10.040 caracteres de cabeza y cola de un resultado fallido, sin ruta al resto y con una cola que no es el final real; Codex deja cabeza y cola sin avisar que falta el medio (`entrega-03-04.md` §5, C1–C5 y X1–X6)—, así que el fallo central ya no existe cuando el agente resume.

## What Changes

El cambio especifica ocho comportamientos. La propuesta entrega el diseño, los escenarios y los fixtures; la implementación se ordena en `tasks.md`, que separa lo que se hace ya en este repositorio de lo que espera a 03 y 04 o a corridas con clientes reales.

- **Carga por acción, no por instalación.** La matriz común de 03 y 04 parte el bloque en capa universal, cuerpo de la skill y módulos. Esta ficha define **cuándo** se lee cada módulo: una tabla por acción, declarada en el catálogo y generada en el `SKILL.md`, con qué módulo es obligatorio, cuál opcional y qué recibe el especialista (`design.md` — D1 y D19).
- **Reglas de lectura por ruta o símbolo.** Se define cuándo el agente puede leer un archivo completo y cuándo debe pedir un rango, un símbolo o una coincidencia estructural, y qué procedencia (archivo, línea, símbolo) acompaña a cada lectura. Incluye la orientación a nivel repositorio: ubicar dónde vive una capacidad sin recorrer el repositorio archivo por archivo.
- **Ninguna llamada innecesaria.** Se define que una relectura del mismo archivo y rango sin escritura intermedia, y una invocación repetida de una integración sobre un estado que no cambió, queden registradas como llamadas innecesarias. El sondeo de una integración ya informada como no disponible queda prohibido dentro de la misma tarea.
- **Respuestas del agente que no entierran el hallazgo.** La ficha pide localizar *"instrucciones largas o resultados verbosos"*, y los resultados verbosos son dos cosas: lo que devuelve un comando y lo que devuelve el agente. Se define que el hallazgo y su procedencia van primero, que el desarrollo se produce a pedido, y que una respuesta no reproduce lo que ya está escrito en un artefacto.
- **Formato de salida acotado con copia recuperable.** Se separan tres roles: quién captura stdout y stderr, quién redacta y quién resume. Captura una fuente independiente antes del truncamiento —el arnés, o una fuente del cliente cuya cobertura se demostró por cliente, interfaz y sistema—; si no la hay, el agente captura con una receta por shell y el resultado queda *sin verificar*. Una salida de más de 300 líneas o de más de 10.000 caracteres se entrega como resumen que retiene errores, fallos y códigos de salida, más una referencia a la copia redactada (`{comando, id de corrida, ruta, huella sha256, vencimiento}`), el registro de captura y las condiciones de ubicación, acceso y retención. Un resultado declarado por el agente, directo o delegado, queda *sin verificar* hasta corroboración independiente.
- **Herramientas por clase de tarea.** Se define, con una tabla por acción OPSX (`design.md` — D18), qué skills e integraciones opcionales habilita cada clase de tarea y cuáles quedan fuera por defecto, de modo que una integración innecesaria no cargue contexto ni superficie. Para cada cliente se declara a qué nivel queda "no habilitada" —instrucción, descripción fuera del contexto o invocación impedida por el cliente—, y el nivel de instrucción nunca se presenta como aislamiento.
- **Fixtures de pérdida de señal y de ruido.** Se especifican fixtures con fallos ubicados al inicio, al medio y al final de una salida grande, con fallos dispersos, con ruido que imita el formato de un error, con un fallo real rodeado de ese ruido, con un fallo que aparece solo por stderr, con datos sensibles ficticios y con un original nunca capturado. El criterio es doble: cada fallo relevante y su salida completa son recuperables, y el resumen no declara un fallo que no ocurrió.
- **Inventario de verbosidad sobre todas las fuentes.** Se mide el tamaño de cada cuerpo de skill y de cada fragmento que el instalador escribe en el proyecto destino, no solo el bloque del orquestador, y se ubica dónde aparecen las salidas verbosas de comandos en el ciclo. Se publica como una tabla única ordenada por tamaño.

Fuera de alcance: cualquier subcomando nuevo de runtime en el binario. Los *Límites compartidos* del roadmap (`CHANGES.md:49`) dicen que *"un ledger firmado, un runtime nuevo o una integración permanente del juez requieren una decisión de arquitectura separada"* y que *"el arnés de evaluación queda fuera del binario"*. El resumen de salida se especifica como **contrato de formato**, no como comando. Los roles se deciden en `design.md` — D8, y quién ejecuta cada regla en D17: el agente siguiendo los módulos que esta ficha redacta, lo declarado en el catálogo y, en evaluación, el arnés fuera del binario. Si más adelante se quiere un capturador permanente en la sesión del usuario, es esa decisión de arquitectura separada, y la matriz de cobertura de D22 es su insumo.

## Capabilities

### New Capabilities

- `task-context-loading`: qué instrucciones del orquestador recibe el agente según la acción activa y el rol, qué parte es obligatoria en toda sesión y cómo se informa un módulo obligatorio ausente.
- `directed-code-reading`: reglas de lectura de código por ruta, rango o símbolo en lugar de archivo completo, con la procedencia que acompaña a cada lectura.
- `bounded-command-output`: formato de resumen de salida de comando que retiene fallos y códigos de salida, con referencia al original completo, su acceso y su retención.
- `task-tool-selection`: qué skills e integraciones opcionales habilita cada clase de tarea y cuál es el comportamiento por defecto cuando una integración no está disponible.
- `concise-agent-reporting`: cómo el agente informa lo que encontró — el hallazgo y su procedencia primero, el desarrollo a pedido — para que una respuesta larga no entierre lo que importa.

### Modified Capabilities

Ninguna. Desde `1b4639b` el repositorio versiona catorce especificaciones, pero ninguna cubre el comportamiento de esta ficha: todas describen el instalador y la GUI de Windows (`catalog-*`, `design-*`, `gui-installer-*`, `installer-*`, `playwright-mcp-harness`, `skill-harness-install`, `tui-environment-detection`, `windows-*`).

La más cercana es `skill-harness-install`, y la frontera es nítida: esa especifica **cómo se clona e instala** una skill (ref por SHA, agentes soportados); esta ficha especifica **cuáles quedan habilitadas** según la tarea activa, una vez instaladas. No hay requisito existente que cambie, así que las cinco capacidades son nuevas y no hay deltas.

## Impact

**Evidencia medida sobre `97d7f6c` y revalidada en `1b4639b`** (los commits intermedios solo tocaron documentación y artefactos de OpenSpec: `CHANGES.md`, `CLAUDE.md`, `AGENTS.md`, `ARCHITECTURE.md`, `README.md`, `.gitignore`, `docs/archive/CHANGES-2026-11-05-post-starters.md` y `openspec/`; ni el código ni los assets cambiaron, y las cifras salen idénticas). El bloque instalado es el máximo posible en los dos clientes, porque Claude y Codex componen en la rama `default` de `internal/harness/config/install.go:70`, que invoca `Compose(h.Toggles, …)`, y `h.Toggles` sale literal del catálogo (`internal/catalog/harnesses.yaml:70`) con los cinco toggles activos. OpenCode compone con los mismos toggles por la rama `ConfigDeliveryPrimaryAgent` (`install.go:53`). No existe hoy ninguna ruta de código, bandera de CLI ni pantalla de la TUI que produzca un bloque menor.

Tamaño del bloque compuesto por combinación de toggles, obtenido llamando a la función real `Compose`. Las líneas se cuentan como `wc -l` (cantidad de saltos de línea), la misma unidad que el informe ejecutivo:

| Toggles activos | Claude | Codex |
|---|---|---|
| ninguno | 7.116 B · 188 L | 5.543 B · 147 L |
| +delegation | 8.121 B · 210 L | 6.394 B · 168 L |
| +model-routing | 8.899 B · 226 L | 6.844 B · 181 L |
| +governance | 10.175 B · 246 L | 8.120 B · 201 L |
| +engram | 16.268 B · 365 L | 14.213 B · 320 L |
| +tdd (**lo que se instala hoy**) | **20.065 B · 448 L** | **18.010 B · 403 L** |

Las dos cifras finales coinciden con las del informe ejecutivo (448 y 403 líneas), lo que confirma que ese informe midió la composición completa y permite ahora atribuir cada byte a su fragmento. Con los dos toggles sustractivos activos, el bloque base mide lo mismo que el archivo base del variante (226 y 181 líneas):

| Fragmento | Claude | % | Codex | % |
|---|---|---|---|---|
| base del variante | 7.116 B | 35,5 % | 5.543 B | 30,8 % |
| `delegation` | 1.005 B | 5,0 % | 851 B | 4,7 % |
| `model-routing` | 778 B | 3,9 % | 450 B | 2,5 % |
| `governance` | 1.276 B | 6,4 % | 1.276 B | 7,1 % |
| `engram` | 6.093 B | 30,4 % | 6.093 B | 33,8 % |
| `tdd` | 3.797 B | 18,9 % | 3.797 B | 21,1 % |

`engram` y `tdd` juntos son el 49,3 % del bloque de Claude y el 54,9 % del de Codex. `tdd` es el fragmento cuya pertinencia depende más claramente de la acción activa; `engram` se parte (D2): una regla corta al núcleo, porque `engram-protocol.md:4` lo declara siempre activo, y el protocolo detallado a módulo. Son por eso los primeros candidatos a resolverse por tarea, y la razón por la que este change puede justificarse con una medición y no con una estimación.

**Aclaración sobre una discrepancia de cifras.** La afirmación de "~200 líneas por variante" y la medición de 448/403 líneas del informe ejecutivo no se contradicen: la primera describe el archivo base del variante (226 L en Claude, 181 L en Codex, `wc -l`) y la segunda el bloque compuesto con los cinco toggles. Son dos objetos distintos. Cualquier comparación posterior debe decir cuál de los dos mide.

**Código y artefactos afectados.**

- `internal/harness/config/compose.go` — la función de ensamblado y su orden determinista; es el mecanismo que esta propuesta quiere seleccionar por acción en lugar de por instalación.
- `internal/harness/config/assets/{claude,codex}/sdd-orchestrator.md` — artefacto compartido con las fichas 02, 03, 04, 07 y 09. Esta propuesta escribe **requisitos** sobre él y redacta el texto de sus propios módulos como entrega a 03 y 04, que son quienes lo reestructuran (D17); no escribe contenido en el archivo.
- `internal/harness/config/assets/{strict-tdd,engram-protocol,governance}.md` — los tres fragmentos aditivos.
- `internal/catalog/harnesses.yaml` — declaración de toggles y, a futuro, de módulos por acción, herramientas por clase de tarea, umbral de lectura y lista de patrones sensibles.
- `internal/harness/skill/` — instalación y verificación de skills; frontera donde se decide qué skill está disponible por tarea.
- `internal/harness/config/assets/modules/` — destino en código de los cuatro módulos que esta ficha redacta; el texto ya está en `entrega-03-04.md` §2 y se entrega a 03 y 04 (D23).
- `eval/taskcontext/` — arnés de esta ficha, fuera del binario, que reutiliza `eval/baseline/` de la 02: fixtures, capturador, redactor, evaluador y matriz de cobertura.

**Qué intercambia esta ficha con las demás.** El tipo de relación está en el encabezado; acá va el contenido concreto.

- **03 y 04** → mantienen la matriz común por cliente, interfaz, rol y acción (publicada por la 03 el 12/11) y la entrega por skill. Esta ficha les da el criterio de reparto (D2), la tabla de carga por acción (D19), la atribución de bytes por fragmento y el texto de sus cuatro módulos (D23), entregados en `entrega-03-04.md`, y les propone un compositor común con registro de módulos en lugar de los dos que hoy tienen sus ramas.
- **06** → esta ficha mantiene captura, acceso y política de datos; la 06, procedencia, huellas y estados, y qué evidencia alcanza (informe unificado). Se le propone la ubicación, el acceso y el vencimiento de la copia redactada, el registro de captura y la lista sensible (D9, D14); la elección del mecanismo de captura espera la matriz de cobertura (D22).
- **07** → consume el contrato de carga de módulos y el formato de salida acotada de esta ficha y la declara dependencia (`opsx-selective-delegation/proposal.md:4`). Su señal S1 ya cita las 300 líneas; esta ficha le agrega los 10.000 caracteres, que el módulo de delegación es obligatorio solo en `verify` (D19) y que el resultado de un worker queda *sin verificar* sin corroboración independiente (D20). Comparar delegación contra ejecución directa es de la 07, no de acá.
- **02** (Giuliano) → su arnés `eval/baseline/` ya está en `main`: manifiesto, perfiles y registros que usan los grupos 8 y 9. Esta ficha le da la lista sensible, el plazo de retención que su D6 dejó pendiente y el inventario medido de partida como línea base de tamaño de texto, y le propone reutilizarlo para su inventario de instrucciones (`opsx-baseline/design.md:46`). **Los números de este documento son tamaño de texto, no consumo de modelo**, y no deben leerse como ahorro.
