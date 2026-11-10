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
> **Los grupos 1, 2, 4, 5 y 7 no dependen de ninguna otra ficha** — 48 de las 62 tareas. El grupo 3
> requiere la matriz de módulos de 03 y 04; el grupo 6, el protocolo de medición de la 02.
>
> Las diecinueve decisiones de `design.md` cubren el diseño del change. Aparte quedan **diez
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
- [ ] 1.8 Inventariar dónde aparecen las salidas verbosas en el ciclo OPSX —qué comandos de tests, build, logs y OpenSpec CLI corre cada acción y cuántas líneas producen sobre este repositorio— y verificar que cada fuente registrada declara la acción en que aparece y si supera el umbral de 300 líneas (`design.md` — D19)

## 2. Contrato de resumen de salida y fixtures

- [ ] 2.1 Escribir los fixtures de salida grande con el fallo al inicio, al medio y al final, cada uno con su salida original y el fallo que debe sobrevivir, y verificar que los tres tienen miles de líneas
- [ ] 2.2 Agregar el fixture de fallos dispersos en más de una posición y verificar que declara cada fallo relevante por separado
- [ ] 2.3 Agregar el fixture de control negativo: salida verbosa, repetitiva y con código de salida cero, y verificar que no declara ningún fallo esperado
- [ ] 2.4 Agregar el fixture con un valor que coincide con un patrón de la lista sensible de `design.md` — D14 (por ejemplo, una asignación leída de un `.env`) y verificar que su resumen esperado y su original esperado muestran la línea con el valor redactado y la marca de omisión
- [ ] 2.5 Agregar el fixture de ruido que imita el formato de un error con código de salida cero, y verificar que su resumen esperado no declara ningún fallo
- [ ] 2.6 Agregar el fixture con un fallo real rodeado de ruido que lo imita, y verificar que su resumen esperado declara exactamente un fallo y no los de ruido
- [ ] 2.7 Redactar el texto del módulo de resumen acotado conforme a `specs/bounded-command-output` (entrega a 03 y 04) y verificar, corriendo los fixtures 2.1 a 2.6 contra un agente real con el arnés fuera del binario, que pasan, incluidos el caso del fallo central y los dos de ruido; esta corrida valida el supuesto de `design.md` — D8
- [ ] 2.8 Declarar en el arnés de evaluación la regla de referencia insuficiente y verificar con un fixture que un resumen sin referencia, con un campo faltante o con una huella que no coincide queda marcado como insuficiente
- [ ] 2.9 Declarar en el catálogo la lista única de patrones de valor sensibles (`design.md` — D14) y verificar con un test que `catalog.Load()` la valida y que todo lo que bloquea el piso de Claude (`internal/harness/config/permissions/overlays.go:16-28`) está cubierto por algún patrón; si el piso agrega una regla no cubierta, el test falla nombrándola
- [ ] 2.10 Redactar en el texto del módulo la regla de redacción antes de escribir a disco (`design.md` — D9) y verificar, con el fixture 2.4 corrido en el arnés, que ni el resumen ni el original guardado contienen el valor sin redactar y que el original mantiene la línea y su posición
- [ ] 2.11 Redactar en el texto del módulo la ubicación (`<os.UserCacheDir>/active-stack/runs/<change>/`) y la retención (hasta archivar el change y siete días más) del original, y verificar en el arnés que el original no queda en el árbol del repositorio y que una petición fuera de plazo informa que no está disponible y desde qué fecha
- [ ] 2.12 Redactar en el texto del módulo la declaración de productor (`design.md` — D8) y verificar en el arnés que un resumen conforme se acepta declarando su productor y que uno que omite un fallo se rechaza quedando el productor registrado
- [ ] 2.13 Documentar el contrato de resumen en `docs/` con un ejemplo de entrada y salida, y verificar que el ejemplo documentado coincide con la salida real del fixture 2.1
- [ ] 2.14 Verificar en el arnés, para cada fixture de 2.1 a 2.6, que a partir de la referencia del resumen se recupera el original completo y que su huella sha256 coincide con la declarada

## 3. Resolución de contexto por acción

> Requiere la matriz de módulos de 03 y 04 (ver `design.md` — *Matriz de módulos*). El eje de
> acciones (explore, propose, apply, verify, archive) ya está fijado por D10, y la carga de cada
> módulo por acción, por D19.

- [ ] 3.1 Verificar el reparto núcleo/módulo decidido en `design.md` — D2 contra la matriz única de módulos de 03 y 04, y registrar las divergencias si las hay (hoy, `governance` frente a `claude-native-opsx/design.md:287-288`); el reparto se ajusta, no se vuelve a decidir
- [ ] 3.2 Declarar en el catálogo, para cada módulo, en qué acciones es obligatorio y en cuáles opcional (tabla de `design.md` — D19), y verificar que `catalog.Load()` rechaza un módulo sin acciones declaradas
- [ ] 3.3 Declarar la acción activa como entrada de la resolución de contexto sobre el catálogo y verificar con tests que para una misma acción, rol e instalación el conjunto de módulos es idéntico entre dos resoluciones
- [ ] 3.4 Verificar con un test que el orden en que se declaran los módulos no altera el resultado de la resolución
- [ ] 3.5 Redactar en el texto del núcleo la regla de módulo ausente —detener si es obligatorio, avisar y resolver directo si es opcional (`design.md` — D19)— y verificar en el arnés que `verify` sin módulo de delegación informa qué módulo falta y no se ejecuta, y que `apply` sin él avisa y resuelve directo
- [ ] 3.6 Verificar con tests sobre el catálogo que una acción sin requisito de TDD no resuelve el módulo de TDD y que una acción que delega resuelve el de delegación junto con el de routing
- [ ] 3.7 Verificar que el núcleo obligatorio, con `governance` y la regla corta de memoria, está presente en sesión normal, en sesión con orquestador y en sesión de especialista
- [ ] 3.8 Declarar la paridad por cliente y verificar que un módulo que un cliente no puede recibir queda registrado como límite en lugar de presentarse como entregado
- [ ] 3.9 Documentar en `docs/` la tabla de acción por módulo y por cliente, y verificar que se genera desde el catálogo en lugar de mantenerse a mano
- [ ] 3.10 Redactar la regla corta de memoria para el núcleo y separar el protocolo detallado de `engram-protocol.md` (formato de guardado, cierre de sesión, compactación, git sync) como módulo, como entrega a 03 y 04, y registrar el tamaño de la regla corta en el golden de 1.1

## 4. Lectura dirigida y procedencia

- [ ] 4.1 Declarar el umbral de lectura completa en el catálogo, en líneas y con 300 como valor (`design.md` — D11), y verificar que `catalog.Load()` rechaza un umbral expresado en otra unidad
- [ ] 4.2 Agregar un test que mida la distribución de líneas (`wc -l`) de los archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en `_test.go` —el mismo conjunto de `design.md` — D11— y falle si el umbral declarado deja de estar entre el p85 y el p92 por rango más cercano, de modo que el número siga justificado cuando el repositorio crezca
- [ ] 4.3 Declarar en el catálogo la lista cerrada de cuatro condiciones de lectura completa (`design.md` — D15) y verificar que `catalog.Load()` rechaza una condición fuera de la lista
- [ ] 4.4 Redactar el texto del módulo de lectura dirigida por ruta, rango y símbolo conforme a `specs/directed-code-reading` y verificar en el arnés que una lectura por símbolo devuelve el símbolo completo con su procedencia, la firma de la declaración que lo contiene si está anidado, y nada más
- [ ] 4.5 Redactar en el mismo texto el registro de lectura no justificada y verificar en el arnés las cuatro condiciones —no supera el umbral, reescritura completa, artefacto de la acción activa y lectura acotada previa vacía— y que una lectura fuera de las cuatro queda registrada
- [ ] 4.6 Redactar en el mismo texto la exigencia de procedencia y verificar en el arnés que una afirmación sobre el código sin ruta ni línea ni símbolo queda marcada como no verificada
- [ ] 4.7 Redactar en el mismo texto la orientación acotada a nivel repositorio y verificar en el arnés que ubicar una capacidad desconocida devuelve rutas candidatas, cada una con su tipo de coincidencia y su cantidad, sin leer el contenido de los archivos descartados
- [ ] 4.8 Verificar con un fixture el orden entre tipos (`design.md` — D13): una ruta con la definición del símbolo se ordena antes que otra con muchas referencias, y la cantidad solo ordena dentro del mismo tipo
- [ ] 4.9 Verificar en el arnés que recorrer el repositorio archivo por archivo solo para orientarse queda registrado como lectura no justificada
- [ ] 4.10 Redactar en el mismo texto el registro de relectura y verificar en el arnés que una segunda lectura del mismo archivo y rango sin escritura intermedia queda registrada como llamada innecesaria, y que la relectura posterior a una escritura se considera justificada
- [ ] 4.11 Documentar en `docs/` cuándo corresponde cada modo de lectura, el umbral con su justificación medida y las cuatro condiciones, y verificar que los ejemplos documentados corresponden a casos cubiertos por las tareas 4.4 a 4.10

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

- **2026-11-10 · pendiente de enviar al grupo** · los siete · `design.md` — *Decisiones tomadas — pendientes de revisión del grupo*. Enviar los diez puntos con la decisión, en qué se apoya, las fichas afectadas y qué cambia si se rechaza, antes de compartir la propuesta final (paso 4, `CHANGES.md:16`). Con 06 es una interfaz, no una aprobación: las dos fichas son hermanas. Con 03 y 04 es alineación, porque esta ficha depende de ellas.
- **2026-11-10 · pendiente de enviar al grupo** · Juani (03), Giuliano (04) · `design.md` — *Matriz de módulos*. Proponer la matriz de G4 de la 04 (`codex-native-opsx/design.md:300-313`) como matriz única, con la 03 completando la columna de Claude, para cortar la dependencia circular (`claude-native-opsx/proposal.md:78` ↔ `codex-native-opsx/proposal.md:71`). Pedir las superficies soportadas y qué módulos no puede recibir Codex. Avisar la divergencia en `governance` (D2).
- **2026-11-10 · pendiente de enviar al grupo** · Juani (03), Giuliano (04) · `design.md` — D17. Proponer que 03 y 04 queden como dueñas de `sdd-orchestrator.md`, con las demás fichas escribiendo requisitos y entregando texto de módulos. La 07 ya lo respalda (`opsx-selective-delegation/design.md:173`).
- **2026-11-10 · pendiente de enviar al grupo** · Neyen (07) · `design.md` — D19. Pedir que la señal S1 cite el umbral de 300 líneas en lugar de "el presupuesto de 05" (`opsx-selective-delegation/design.md:40`), y compartir la tabla de carga y que la delegación es obligatoria solo en `verify`, compatible con su D2 (`design.md:58`).
- **2026-11-10 · pendiente de enviar al grupo** · Gianfranco (06) · `design.md` — D9 y D14. Proponer que la 06 adopte el formato de referencia `{comando, id de corrida, ruta, huella sha256, vencimiento}`, la ubicación fuera del repo y la retención hasta archivar más siete días, y que use la lista de patrones sensibles. Señalar las cuatro diferencias de la tabla de D9.
- **2026-11-10 · pendiente de enviar al grupo** · Matías (02), Gianfranco (06), Neyen (10) · `design.md` — D14. Avisar que la lista de patrones sensibles es una sola, declarada en el catálogo, con dueña la 05, para que 02, 06 y 10 la usen en lugar de mantener la propia (la 10, para su derivado redactado, `opsx-shadow-review/design.md:126`).
- **2026-11-10 · pendiente de enviar al grupo** · Juan Cruz (01), Gianfranco (08) · `design.md` — D14. Pedir que una de las dos lleve a Codex un piso equivalente al de Claude: hoy el piso aplica solo a Claude (`overlays.go:73`) y Codex escribe únicamente `approval_policy` y `sandbox_mode` (`permissions/install.go:114-134`).
- **2026-11-10 · pendiente de enviar al grupo** · Matías (02) · `design.md` — D7. Pedir el protocolo de medición que usa el grupo 6 y proponer que su inventario de instrucciones observables (`opsx-baseline/design.md:46`) reutilice el golden del grupo 1 en lugar de duplicarlo. La circularidad entre 02 y 03/04 ya está resuelta en su rama (`opsx-baseline/design.md:47`): no se le pide nada sobre eso.
- **2026-11-10 · pendiente de enviar al grupo** · los siete · rama `propose/opsx-task-context`. Compartir la rama para la revisión de los siete, con la revisión de partida `1b4639b`. Desde ese commit el repo versiona `openspec/`, así que los artefactos viajan con la rama.
- **2026-11-10 · pendiente de enviar al grupo** · los siete · sin artefacto propio (aviso, no pide decisión). La ficha 08 declara depender de la 06 (`CHANGES.md:229`) cuando el acoplamiento real es con la 01 —ambas tocan `permissions/install.go` y la 01 decide la política que la 08 verifica—; cinco fichas cierran con criterios de ejecución en una ronda de solo propose (02, 05, 07, 09, 10); y el `(TBD)` es el criterio de aceptación en la 01 (`CHANGES.md:69`) y en la 09 (`CHANGES.md:263`), lo que por la regla del paso 4 (`CHANGES.md:16`) les impide ir a revisión así.
- Lo resuelto sin consultar, que no requiere acuerdo: el roadmap canónico se publicó en `CHANGES.md` en `bd63543` y la ficha se revalidó sobre `1b4639b`.
- Archivar el change una vez satisfechos los requisitos de revisión del proyecto.
