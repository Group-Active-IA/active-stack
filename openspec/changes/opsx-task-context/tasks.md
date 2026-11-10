# Tasks

> Todas las tareas quedan **pendientes**: esta ronda es solo propose. El orden refleja
> dependencias reales, no un cronograma.
>
> **Quién ejecuta** (`design.md` — D17): los tests de Go verifican solo lo declarativo —catálogo
> validado por `catalog.Load()`, texto compuesto por `Compose`, fixtures como archivos—. La
> conducta del agente se verifica corriendo los fixtures contra el agente real con el arnés fuera
> del binario (`CHANGES.md:49`). Donde una tarea dice "redactar el texto del módulo", la 05 lo
> entrega a 03 y 04, que lo ubican; la 05 no escribe en `sdd-orchestrator.md`.
>
> **Los grupos 1, 2, 4, 5 y 7 no dependen de ninguna otra ficha** — 55 de las 70 tareas. El grupo 3
> requiere la matriz de módulos de 03 y 04; el grupo 6, el protocolo de medición de la 02.
>
> Las veintiuna decisiones de `design.md` cubren el diseño del change. Aparte quedan **doce
> puntos** en `design.md` — *Decisiones tomadas — pendientes de revisión del grupo*: todos están
> decididos acá y se comparten porque tocan alcance de otras fichas.

## 1. Inventario reproducible del contexto de partida

- [ ] 1.1 Agregar un golden de inventario en `internal/harness/config/testdata/` que registre, para las seis combinaciones acumulativas de toggles en los variantes `claude` y `codex`, los bytes y las líneas (`wc -l`, cantidad de saltos de línea) que produce `Compose`, y verificar que el test compara contra el golden; un cambio de asset de otra ficha se resuelve regenerándolo con `go test -update` (bandera ya existente, `internal/harness/config/compose_test.go:14`), no rompe por diseño. Valores de partida en `1b4639b`: 20.065 B / 448 L en Claude y 18.010 B / 403 L en Codex
- [ ] 1.2 Registrar en el mismo golden la medición de las demás variantes de assets (`antigravity`, `cursor`, `gemini`, `generic`, `opencode`, `windsurf`) como referencia, fuera del alcance de paridad de esta ficha, que es Claude y Codex, y verificar que ninguna variante de `internal/harness/config/assets/` queda sin medición registrada
- [ ] 1.3 Documentar en `docs/` la atribución de bytes por fragmento de la tabla de `proposal.md` — Impact y verificar que el documento reproduce las cifras del golden de 1.1
- [ ] 1.4 Verificar con un test que Claude y Codex componen por la rama `default` de `internal/harness/config/install.go:70` con los toggles literales del catálogo, y que OpenCode lo hace por la rama `ConfigDeliveryPrimaryAgent` (`install.go:53`); el test documenta la ruta de composición actual sin depender de un selector futuro
- [ ] 1.5 Inventariar el tamaño de cada cuerpo de skill del catálogo (propias y de terceros) y verificar que el inventario cubre todas las skills que `harnesses.yaml` declara, sin ninguna sin medir
- [ ] 1.6 Inventariar las demás fuentes de instrucción que el instalador escribe en el proyecto destino, incluidos los fragmentos de sección propia como `design-mockups` (8.865 B), y verificar que la suma por cliente coincide con el total realmente inyectado
- [ ] 1.7 Publicar el inventario completo de 1.1 a 1.6 como una sola tabla ordenada por tamaño descendente y verificar que identifica las cinco fuentes más grandes, que son las candidatas a resolverse por tarea
- [ ] 1.8 Inventariar dónde aparecen las salidas verbosas en el ciclo OPSX —qué comandos de tests, build, logs y OpenSpec CLI corre cada acción y cuántas líneas producen sobre este repositorio— y verificar que cada fuente registrada declara la acción en que aparece, si supera el umbral de salidas (300 líneas o 10.000 caracteres, `design.md` — D3 y D19) y, si lo supera, que se corre con captura a archivo (`design.md` — D8)

## 2. Contrato de resumen de salida y fixtures

- [ ] 2.1 Escribir los fixtures de salida grande con el fallo al inicio, al medio y al final, cada uno con su salida original, su código de salida y el fallo que debe sobrevivir, y verificar que los tres tienen miles de líneas y superan los 10.000 caracteres, de modo que el cliente los recortaría sin captura (`design.md` — D3)
- [ ] 2.2 Agregar el fixture de fallos dispersos en más de una posición y verificar que declara cada fallo relevante por separado
- [ ] 2.3 Agregar el fixture de control negativo: salida verbosa, repetitiva y con código de salida cero, y verificar que no declara ningún fallo esperado
- [ ] 2.4 Agregar el fixture con un valor que coincide con un patrón de la lista sensible de `design.md` — D14 (por ejemplo, una asignación leída de un `.env`) y verificar que su resumen esperado y su original esperado muestran la línea con el valor redactado y la marca de omisión
- [ ] 2.5 Agregar el fixture de ruido que imita el formato de un error con código de salida cero, y verificar que su resumen esperado no declara ningún fallo
- [ ] 2.6 Agregar el fixture con un fallo real rodeado de ruido que lo imita, y verificar que su resumen esperado declara exactamente un fallo y no los de ruido
- [ ] 2.7 Redactar el texto del módulo de resumen acotado conforme a `specs/bounded-command-output` (entrega a 03 y 04) y verificar, corriendo los fixtures 2.1 a 2.6, 2.15 y 2.16 contra el agente real en Claude Code y en Codex —el cliente real, no una simulación— con el arnés fuera del binario, que pasan, incluidos el caso del fallo central, el de stderr y los dos de ruido; esta corrida valida el supuesto de `design.md` — D8
- [ ] 2.8 Declarar en el arnés de evaluación la regla de referencia insuficiente y verificar con un fixture que un resumen sin referencia, con un campo faltante o con una huella que no coincide queda marcado como insuficiente
- [ ] 2.9 Declarar en el catálogo la lista única de patrones de valor sensibles (`design.md` — D14) y verificar con un test que `catalog.Load()` la valida y que todo lo que bloquea el piso de Claude (`internal/harness/config/permissions/overlays.go:16-28`) está cubierto por algún patrón; si el piso agrega una regla no cubierta, el test falla nombrándola
- [ ] 2.10 Redactar en el texto del módulo la regla de redacción antes de conservar —el crudo de la captura se redacta a un segundo archivo y se borra antes de entregar el resumen— y el límite de las copias propias del cliente (`design.md` — D9), y verificar, con el fixture 2.4 corrido en el arnés, que ni el resumen ni el original guardado contienen el valor sin redactar, que el crudo ya no existe al entregar el resumen y que el original mantiene la línea y su posición
- [ ] 2.11 Redactar en el texto del módulo la ubicación (`<os.UserCacheDir>/active-stack/runs/<change>/`) y la retención (hasta archivar el change y siete días más) del original, y verificar en el arnés que el original no queda en el árbol del repositorio y que una petición fuera de plazo informa que no está disponible y desde qué fecha
- [ ] 2.12 Redactar en el texto del módulo la declaración de productor (`design.md` — D8) y verificar en el arnés que un resumen conforme se acepta declarando su productor y que uno que omite un fallo se rechaza quedando el productor registrado
- [ ] 2.13 Documentar el contrato de resumen en `docs/` con un ejemplo de entrada y salida, y verificar que el ejemplo documentado coincide con la salida real del fixture 2.1
- [ ] 2.14 Verificar en el arnés, para cada fixture de 2.1 a 2.6 y para el de 2.15, que a partir de la referencia del resumen se recupera el original completo, que su huella sha256 coincide con la declarada y que el código de salida registrado coincide con el del comando
- [ ] 2.15 Agregar el fixture de un comando que falla escribiendo el fallo **solo por stderr**, con stdout verboso y sin fallos, y verificar que su resumen esperado contiene la línea de stderr y el código de salida distinto de cero; el arnés lo corre en Claude y en Codex, porque en Codex la documentación no dice cómo combina los flujos (`design.md` — D4)
- [ ] 2.16 Agregar el fixture de **original nunca capturado** —comando ya corrido sin redirección, y archivo de captura ausente— y verificar que su resumen esperado está marcado "parcial — sin verificar", dice por qué, no lleva ruta ni huella, y que pedir su original informa "nunca capturado" y no "vencido" (`specs/bounded-command-output` — *Un resumen sin original capturado es parcial*)
- [ ] 2.17 Redactar en el texto del módulo de resumen acotado el paso de captura antes de leer, con la receta para bash y para PowerShell (`design.md` — D8), y verificar en el arnés, en los dos shells, que stdout y stderr quedan en el mismo archivo de la ruta de corrida, que el código de salida registrado es el del comando y no el de una tubería, y que el resumen se produce leyendo ese archivo y no la vista inline del cliente
- [ ] 2.18 Redactar en el texto del módulo la regla del resumen de un worker (`design.md` — D20) y verificar en el arnés que un resumen recibido sin corroborar figura como *sin verificar*, que pasa a *verificado* al corroborarlo contra la referencia y a *rechazado* cuando la re-ejecución con captura muestra un fallo omitido
- [ ] 2.19 Declarar en el catálogo el umbral de salidas de comando en dos unidades (300 líneas y 10.000 caracteres, `design.md` — D3), aparte del umbral de lectura de archivos de la tarea 4.1, y verificar que `catalog.Load()` rechaza un umbral de salidas al que le falte una de las dos unidades

## 3. Resolución de contexto por acción

> Requiere la matriz de módulos de 03 y 04 (ver `design.md` — *Matriz de módulos*). El eje de
> acciones (explore, propose, apply, verify, archive) ya está fijado por D10, y la carga de cada
> módulo por acción, por D19.

- [ ] 3.1 Verificar el reparto núcleo/módulo decidido en `design.md` — D2 contra la matriz única de módulos de 03 y 04, y registrar las divergencias si las hay (hoy, `governance` frente a `claude-native-opsx/design.md:287-288`); el reparto se ajusta, no se vuelve a decidir
- [ ] 3.2 Declarar en el catálogo, para cada módulo, su archivo dentro de la skill (`sdd-orchestrator/modules/<id>.md`, punto *Ubicación de los módulos*) y en qué acciones es obligatorio y en cuáles opcional (tabla de `design.md` — D19), y verificar que ningún archivo de `modules/` lleve metadatos de skill y que `catalog.Load()` rechaza un módulo sin acciones declaradas
- [ ] 3.3 Declarar la acción activa como entrada de la resolución de contexto sobre el catálogo y verificar con tests que para una misma acción, rol e instalación el conjunto de módulos es idéntico entre dos resoluciones
- [ ] 3.4 Verificar con un test que el orden en que se declaran los módulos no altera el resultado de la resolución
- [ ] 3.5 Redactar en el texto del núcleo la regla de módulo ausente —detener si es obligatorio, avisar y resolver directo si es opcional (`design.md` — D19)— y verificar en el arnés que `verify` sin módulo de delegación informa qué módulo falta y no se ejecuta, y que `apply` sin él avisa y resuelve directo; verificar también que las tareas del mismo pedido que no dependen del módulo se completan, que el reporte separa lo completado de lo detenido, y que la acción se retoma cuando el módulo o una superficie compatible queda disponible
- [ ] 3.6 Verificar con tests sobre el catálogo que una acción sin requisito de TDD no resuelve el módulo de TDD y que una acción que delega resuelve el de delegación junto con el de routing
- [ ] 3.7 Verificar que el núcleo obligatorio, con `governance` y la regla corta de memoria, está presente en sesión normal, en sesión con orquestador y en sesión de especialista, y que cuando el especialista no carga `CLAUDE.md` (Explore, Plan u `omitClaudeMd: true` en Claude; en Codex, mientras el spike de la 04 no confirme que el subagente lee `AGENTS.md`) las reglas obligatorias y `governance` viajan en el encargo (`design.md` — D21)
- [ ] 3.8 Declarar la paridad por cliente y verificar que un módulo que un cliente no puede recibir queda registrado como límite en lugar de presentarse como entregado
- [ ] 3.9 Documentar en `docs/` la tabla de acción por módulo y por cliente, y verificar que se genera desde el catálogo en lugar de mantenerse a mano
- [ ] 3.10 Redactar la regla corta de memoria para el núcleo y separar el protocolo detallado de `engram-protocol.md` (formato de guardado, cierre de sesión, compactación, git sync) como módulo, como entrega a 03 y 04, y registrar el tamaño de la regla corta en el golden de 1.1
- [ ] 3.11 Redactar la referencia de invocación del núcleo por cliente con la sintaxis documentada (Claude Code `/nombre`; Codex `$nombre` en CLI e IDE, `@` en la app de ChatGPT) y con los nombres que fijen 03 y 04 (`design.md` — D21), y verificar con un test sobre el texto compuesto que el núcleo de cada cliente no contiene la sintaxis del otro ni `claude --agent` fijo

## 4. Lectura dirigida y procedencia

- [ ] 4.1 Declarar el umbral de lectura completa en el catálogo, en líneas y con 300 como valor (`design.md` — D11), y verificar que `catalog.Load()` rechaza un umbral expresado en otra unidad
- [ ] 4.2 Agregar un test que mida la distribución de líneas (`wc -l`) de los archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en `_test.go` —el mismo conjunto de `design.md` — D11— y falle si el umbral declarado deja de estar entre el p85 y el p92 por rango más cercano, de modo que el número siga justificado cuando el repositorio crezca
- [ ] 4.3 Declarar en el catálogo la lista cerrada de cinco condiciones de lectura completa (`design.md` — D15) y verificar que `catalog.Load()` rechaza una condición fuera de la lista
- [ ] 4.4 Redactar el texto del módulo de lectura dirigida por ruta, rango y símbolo conforme a `specs/directed-code-reading` y verificar en el arnés que una lectura por símbolo devuelve el símbolo completo con su procedencia, la firma de la declaración que lo contiene si está anidado, y nada más
- [ ] 4.5 Redactar en el mismo texto el registro de lectura no justificada y verificar en el arnés las cinco condiciones —no supera el umbral, reescritura completa, artefacto de la acción activa, lectura acotada previa vacía y rangos insuficientes con motivo declarado— y que una lectura fuera de las cinco queda registrada
- [ ] 4.6 Redactar en el mismo texto la exigencia de procedencia y verificar en el arnés que una afirmación sobre el código sin ruta ni línea ni símbolo queda marcada como no verificada
- [ ] 4.7 Redactar en el mismo texto la orientación acotada a nivel repositorio y verificar en el arnés que ubicar una capacidad desconocida devuelve rutas candidatas, cada una con su tipo de coincidencia y su cantidad, sin leer el contenido de los archivos descartados
- [ ] 4.8 Verificar con un fixture el orden entre tipos (`design.md` — D13): una ruta con la definición del símbolo se ordena antes que otra con muchas referencias, y la cantidad solo ordena dentro del mismo tipo
- [ ] 4.9 Verificar en el arnés que recorrer el repositorio archivo por archivo solo para orientarse queda registrado como lectura no justificada
- [ ] 4.10 Redactar en el mismo texto el registro de relectura y verificar en el arnés que una segunda lectura del mismo archivo y rango sin escritura intermedia queda registrada como llamada innecesaria, y que la relectura posterior a una escritura se considera justificada
- [ ] 4.12 Verificar con un fixture la excepción justificada al umbral (`design.md` — D15, quinta condición): una lectura completa de un archivo de más de 300 líneas con `lectura completa: <ruta> — <motivo>` en el reporte se acepta, la misma lectura sin motivo queda registrada como no justificada, y el arnés cuenta cuántas lecturas entraron por la quinta condición como insumo para recalibrar el umbral
- [ ] 4.11 Documentar en `docs/` cuándo corresponde cada modo de lectura, el umbral con su justificación medida y las cinco condiciones, y verificar que los ejemplos documentados corresponden a casos cubiertos por las tareas 4.4 a 4.10 y 4.12

## 5. Herramientas por clase de tarea

> D8, D9 y D18 están decididas. Con 06 solo se acuerda la interfaz de la referencia al original
> (ver `design.md` — *Referencia, ubicación y retención del original*). Este grupo no queda
> bloqueado por eso.

- [ ] 5.1 Declarar en el catálogo la tabla de clase de tarea por herramienta de `design.md` — D18 (herramientas base y opcionales habilitadas por acción) y verificar que `catalog.Load()` rechaza una integración opcional sin clase de tarea declarada
- [ ] 5.2 Declarar en el catálogo el valor por defecto desactivado para integraciones opcionales, redactar la regla en el texto del módulo de selección de herramientas y verificar en el arnés que una integración instalada y no habilitada no se ofrece en la tarea activa
- [ ] 5.3 Redactar en el mismo texto la degradación declarada y verificar en el arnés que al faltar una integración habilitada el agente informa qué capacidad queda sin cubrir y el camino alternativo si existe
- [ ] 5.4 Verificar en el arnés que al faltar una integración sin camino alternativo el resultado parcial no se presenta como completo
- [ ] 5.5 Verificar de punta a punta que la redacción y la retención de las tareas 2.10 y 2.11 se respetan cuando el original se pide desde una clase de tarea distinta de la que lo generó
- [ ] 5.6 Redactar en el mismo texto el registro de invocación repetida y verificar en el arnés que una segunda invocación de la misma integración con la misma entrada, sin cambio en el estado que observa, queda registrada como llamada innecesaria
- [ ] 5.7 Verificar en el arnés que una integración ya informada como no disponible no se vuelve a invocar en la misma tarea y que se usa el camino alternativo declarado si existe
- [ ] 5.8 Documentar en `docs/` la tabla de clase de tarea por herramienta con su estado por defecto, y verificar que se genera desde el catálogo
- [ ] 5.9 Declarar en el catálogo, por integración opcional y por cliente, el nivel de control alcanzable —(a) instrucción, (b) descripción fuera del contexto, (c) el cliente no deja invocarla— con los mecanismos de `design.md` — D18, y verificar que `catalog.Load()` rechaza una integración sin nivel declarado para Claude o para Codex, que `allowed-tools` no se acepta como nivel (c) y que `docs/` (tarea 5.8) muestra el nivel junto al estado

## 6. Comparación de integración y de módulo, con el protocolo de 02

- [ ] 6.1 Preparar la corrida de una misma clase de tarea con y sin cada integración opcional, con la misma definición de aceptación, y verificar que la definición de aceptación es byte a byte idéntica entre ambas corridas
- [ ] 6.2 Registrar el contexto efectivo de cada corrida con el protocolo de 02 y verificar que la diferencia observada queda atribuida a la integración y no a un cambio de criterio
- [ ] 6.3 Verificar la resolución de contexto de punta a punta: comparar una sesión normal con una que invoque al orquestador en cada cliente soportado, en instalación limpia y después de actualizar una instalación anterior, y comprobar qué módulos recibe cada una
- [ ] 6.4 Preparar la corrida de una misma acción con y sin cada módulo (`tdd`, `delegation`, `model-routing`, protocolo detallado de `engram`), con la misma definición de aceptación y el protocolo de 02, y verificar que la diferencia de contexto efectivo y de aceptación queda atribuida al módulo (`specs/task-context-loading` — *Comparabilidad con y sin cada módulo*)

## 7. Respuestas del agente que no entierran el hallazgo

> Lo que se evalúa es conducta del modelo: cada verificación de este grupo aplica una rúbrica
> sobre los fixtures, corridos con el arnés fuera del binario (`design.md` — D16 y D17).

- [ ] 7.1 Escribir los fixtures de respuesta: un hallazgo informado con el recorrido primero y otro con el resultado primero, sobre el mismo caso, y verificar que los dos describen exactamente el mismo hallazgo con la misma procedencia
- [ ] 7.2 Redactar el texto del módulo de reporte conciso conforme a `specs/concise-agent-reporting` y escribir la rúbrica: la primera línea dice el resultado con su ruta y línea, la segunda la consecuencia o el próximo paso, y no se repite el desarrollo de pasos anteriores; verificar que la rúbrica distingue los dos fixtures de 7.1
- [ ] 7.3 Verificar con la rúbrica que varios hallazgos se informan como lista con su procedencia y que la lista precede a cualquier desarrollo
- [ ] 7.4 Verificar con la rúbrica que el desarrollo no viaja en la respuesta inicial y que se produce cuando se lo pide, sobre la misma procedencia
- [ ] 7.5 Verificar con la rúbrica que cuando el contenido ya está en un artefacto del proyecto la respuesta lo señala en lugar de reproducirlo, y que cuando no está escrito en ninguno sí lo contiene
- [ ] 7.6 Verificar con la rúbrica que una revisión sin hallazgos informa en una línea qué se revisó, sin narrar el recorrido
- [ ] 7.7 Documentar en `docs/` el contrato de respuesta con el par de ejemplos del fixture 7.1, y verificar que el ejemplo documentado coincide con la salida real

## Workflow follow-up

> Cada ítem lleva fecha, estado, destinatarios y artefacto. Ningún acuerdo está registrado
> todavía: todo lo de abajo está **pendiente de enviar al grupo**. Cuando se acuerde, se anota
> con quién y en qué artefacto quedó incorporado.

- **2026-11-10 · pendiente de enviar al grupo** · los siete · `design.md` — *Decisiones tomadas — pendientes de revisión del grupo*. Enviar los doce puntos con la decisión, en qué se apoya, las fichas afectadas y qué cambia si se rechaza, antes de compartir la propuesta final (paso 4, `CHANGES.md:16`). Con 06 es una interfaz, no una aprobación: las dos fichas son hermanas. Con 03 y 04 es alineación, porque esta ficha depende de ellas.
- **2026-11-10 · pendiente de enviar al grupo** · Juani (03), Giuliano (04) · `design.md` — *Matriz de módulos*. Proponer la matriz de G4 de la 04 (`codex-native-opsx/design.md:300-313`) como matriz única, con la 03 completando la columna de Claude, para cortar la dependencia circular (`claude-native-opsx/proposal.md:78` ↔ `codex-native-opsx/proposal.md:71`). Pedir las superficies soportadas y qué módulos no puede recibir Codex. Avisar la divergencia en `governance` (D2).
- **2026-11-10 · pendiente de enviar al grupo** · Juani (03), Giuliano (04) · `design.md` — D17. Proponer que 03 y 04 queden como dueñas de `sdd-orchestrator.md`, con las demás fichas escribiendo requisitos y entregando texto de módulos. La 07 ya lo respalda (`opsx-selective-delegation/design.md:173`).
- **2026-11-10 · pendiente de enviar al grupo** · Neyen (07) · `design.md` — D19. Pedir que la señal S1 cite el umbral de 300 líneas en lugar de "el presupuesto de 05" (`opsx-selective-delegation/design.md:40`), y compartir la tabla de carga y que la delegación es obligatoria solo en `verify`, compatible con su D2 (`design.md:58`).
- **2026-11-10 · pendiente de enviar al grupo** · Gianfranco (06) · `design.md` — D9 y D14. Proponer que la 06 adopte el formato de referencia `{comando, id de corrida, ruta, huella sha256, vencimiento}`, la ubicación fuera del repo y la retención hasta archivar más siete días, y que use la lista de patrones sensibles. Señalar las cuatro diferencias de la tabla de D9.
- **2026-11-10 · pendiente de enviar al grupo** · Matías (02), Gianfranco (06), Neyen (10) · `design.md` — D14. Avisar que la lista de patrones sensibles es una sola, declarada en el catálogo, con dueña la 05, para que 02, 06 y 10 la usen en lugar de mantener la propia (la 10, para su derivado redactado, `opsx-shadow-review/design.md:126`).
- **2026-11-10 · pendiente de enviar al grupo** · Juan Cruz (01), Gianfranco (08) · `design.md` — D14. Pedir que una de las dos lleve a Codex un piso equivalente al de Claude: hoy el piso aplica solo a Claude (`overlays.go:73`) y Codex escribe únicamente `approval_policy` y `sandbox_mode` (`permissions/install.go:114-134`).
- **2026-11-10 · pendiente de enviar al grupo** · Matías (02) · `design.md` — D7. Pedir el protocolo de medición que usa el grupo 6 y proponer que su inventario de instrucciones observables (`opsx-baseline/design.md:46`) reutilice el golden del grupo 1 en lugar de duplicarlo. La circularidad entre 02 y 03/04 ya está resuelta en su rama (`opsx-baseline/design.md:47`): no se le pide nada sobre eso. Pedirle además las métricas de lectura dirigida (lecturas, bytes leídos, herramientas cargadas por tarea), registradas por separado para terminal e IDE, según la decisión confirmada para 02 (primera medición con terminal e IDE separados).
- **2026-11-10 · pendiente de enviar al grupo** · los siete · rama `propose/opsx-task-context`. Compartir la rama para la revisión de los siete, con la revisión de partida `1b4639b`. Desde ese commit el repo versiona `openspec/`, así que los artefactos viajan con la rama.
- **2026-11-10 · pendiente de enviar al grupo** · los siete · sin artefacto propio (aviso, no pide decisión). La ficha 08 declara depender de la 06 (`CHANGES.md:229`) cuando el acoplamiento real es con la 01 —ambas tocan `permissions/install.go` y la 01 decide la política que la 08 verifica—; cinco fichas cierran con criterios de ejecución en una ronda de solo propose (02, 05, 07, 09, 10); y el `(TBD)` es el criterio de aceptación en la 01 (`CHANGES.md:69`) y en la 09 (`CHANGES.md:263`), lo que por la regla del paso 4 (`CHANGES.md:16`) les impide ir a revisión así.
- **2026-11-10 · pendiente de enviar al grupo** · Juani (03) · `design.md` — D21 y D18 (aviso de hechos de documentación oficial, no pide decisión). (i) Explore y Plan *"start without your CLAUDE.md files"* y un agente con `omitClaudeMd: true` tampoco los carga (https://code.claude.com/docs/en/sub-agents): lo que la 03 deje en `CLAUDE.md` no llega a esos especialistas, y la 05 repite las reglas obligatorias en el encargo. (ii) La invocación explícita de una skill en Claude es `/nombre` (https://code.claude.com/docs/en/skills); la extensión de VS Code no documenta `--agent` (https://code.claude.com/docs/en/vs-code): pedir el texto de invocación por superficie para el núcleo. (iii) `allowed-tools` de una skill no restringe (*"It does not restrict which tools are available"*); el bloqueo nativo es `permissions.deny` con `Agent(<nombre>)` o `Skill(skill:<nombre>)` (https://code.claude.com/docs/en/permissions).
- **2026-11-10 · pendiente de enviar al grupo** · Giuliano (04) · `design.md` — D21 y D18 (aviso de hechos de documentación oficial, no pide decisión). (i) En la app de ChatGPT la skill se elige con `@`, no con `$` (https://learn.chatgpt.com/docs/build-skills): el núcleo de Codex necesita el texto por superficie. (ii) La invocación con `$` de una skill con `allow_implicit_invocation: false` tiene un reporte abierto de fallo (https://github.com/openai/codex/issues/23454, sin respuesta de mantenedores). (iii) El archivo de un rol no fija `mcp_servers` ni `sandbox_mode` (*"Configure MCP servers in the parent session"*, https://learn.chatgpt.com/docs/agent-configuration/subagents): el control por acción de MCP en Codex es solo por instrucción, y la 05 lo declara como límite. (iv) Pedir que el spike responda si un subagente lee `AGENTS.md` y cómo combina Codex stdout y stderr, que la documentación no dice.
- **2026-11-10 · pendiente de enviar al grupo** · Neyen (07) · `design.md` — D3 y D19. Avisar que la salida verbosa pasa a ser más de 300 líneas **o** más de 10.000 caracteres (límite de Claude para un resultado fallido, https://code.claude.com/docs/en/tools-reference), que el resumen exige capturar antes de leer (D8), que el resumen de un worker queda *sin verificar* hasta que el coordinador lo corrobora (D20), y que `verify` se detiene en una superficie sin subagentes.
- **2026-11-10 · pendiente de enviar al grupo** · Gianfranco (06) · `design.md` — D9 y D20. Avisar que el resumen de un worker sin corroborar usa los tres estados de la 06 (`opsx-verifiable-evidence/design.md:57-61`), que la garantía de redacción se acota a la copia de Active Stack porque el cliente guarda la suya (https://code.claude.com/docs/en/tools-reference), y que un original nunca capturado se distingue de uno vencido.
- **2026-11-10 · pendiente de enviar al grupo** · Juan Cruz (01), Gianfranco (08) · `design.md` — D14 (complementa el pedido anterior). Señalar el mecanismo nativo de Codex para el piso: `permissions.<name>.filesystem` con `"deny"` por patrón (`**/*.env`) y `shell_environment_policy` (https://learn.chatgpt.com/docs/config-file/config-reference), con estado beta no verificable.
- Lo resuelto sin consultar, que no requiere acuerdo: el roadmap canónico se publicó en `CHANGES.md` en `bd63543` y la ficha se revalidó sobre `1b4639b`.
- Archivar el change una vez satisfechos los requisitos de revisión del proyecto.
