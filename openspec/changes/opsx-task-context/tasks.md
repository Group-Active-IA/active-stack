# Tasks

> Todas las tareas quedan **pendientes**: esta ronda es solo propose. El orden refleja
> dependencias reales, no un cronograma.
>
> **Los grupos 1, 2, 4 y 7 no dependen de ninguna otra ficha** — 38 de las 58 tareas. El grupo 3
> requiere el mapeo de módulo a acción de 03 y 04; el grupo 6, el protocolo de medición de la 02.
>
> Las dieciséis decisiones de `design.md` cubren el diseño del change. Aparte quedan **seis
> puntos que el grupo tiene que cerrar**, listados al final de `design.md`: cuatro están
> decididos acá y se comparten porque tocan alcance ajeno, y dos no son de esta ficha — el
> inventario de módulos lo produce la 04 y los agentes incluidos los define el roadmap.

## 1. Inventario reproducible del contexto de partida

- [ ] 1.1 Agregar un test de inventario en `internal/harness/config` que llame a `Compose` para las seis combinaciones acumulativas de toggles en los variantes `claude` y `codex` y verifique que el bloque con los cinco toggles mide 20.065 B / 449 L y 18.010 B / 404 L respectivamente; el test falla si una edición de los assets cambia el tamaño sin actualizar la cifra registrada
- [ ] 1.2 Extender ese test a los ocho variantes del catálogo (`antigravity`, `claude`, `codex`, `cursor`, `gemini`, `generic`, `opencode`, `windsurf`) y verificar que ninguno queda sin medición registrada
- [ ] 1.3 Documentar en `docs/` la atribución de bytes por fragmento de la tabla de `proposal.md` — Impact y verificar que el documento reproduce las cifras que emite el test de 1.1
- [ ] 1.4 Verificar con un test que `install.go` compone con los toggles literales del catálogo y que hoy no existe ninguna ruta que produzca un bloque menor; el test documenta el estado de partida y debe fallar cuando se agregue el selector del grupo 3
- [ ] 1.5 Inventariar el tamaño de cada cuerpo de skill del catálogo (propias y de terceros) y verificar que el inventario cubre todas las skills que `harnesses.yaml` declara, sin ninguna sin medir
- [ ] 1.6 Inventariar las demás fuentes de instrucción verbosa que el instalador escribe en el proyecto destino, incluidos los fragmentos de sección propia como `design-mockups`, y verificar que la suma por cliente coincide con el total realmente inyectado
- [ ] 1.7 Publicar el inventario completo de 1.1 a 1.6 como una sola tabla ordenada por tamaño descendente y verificar que identifica las cinco fuentes más grandes, que son las candidatas a resolverse por tarea

## 2. Contrato de resumen de salida y fixtures

- [ ] 2.1 Escribir los fixtures de salida grande con el fallo al inicio, al medio y al final, cada uno con su salida original y el fallo que debe sobrevivir, y verificar que los tres superan las mil líneas
- [ ] 2.2 Agregar el fixture de fallos dispersos en más de una posición y verificar que declara cada fallo relevante por separado
- [ ] 2.3 Agregar el fixture de control negativo: salida verbosa, repetitiva y con código de salida cero, y verificar que no declara ningún fallo esperado
- [ ] 2.4 Agregar el fixture con un valor de clase de dato sensible y verificar que su resumen esperado muestra la línea con el valor redactado y la marca de omisión
- [ ] 2.5 Agregar el fixture de ruido que imita el formato de un error con código de salida cero, y verificar que su resumen esperado no declara ningún fallo
- [ ] 2.6 Agregar el fixture con un fallo real rodeado de ruido que lo imita, y verificar que su resumen esperado declara exactamente un fallo y no los de ruido
- [ ] 2.7 Implementar el resumen conforme a `specs/bounded-command-output` y verificar que los fixtures 2.1 a 2.6 pasan, incluidos el caso del fallo central y los dos de ruido
- [ ] 2.8 Verificar con un test que un resumen sin referencia recuperable al original queda marcado como insuficiente
- [ ] 2.9 Conectar la redacción a las clases que declara el piso de permisos (`design.md` — D14) y verificar con un test que al agregar una clase al piso la redacción la cubre sin tocar esta capacidad, y que no existe una segunda lista propia
- [ ] 2.10 Implementar la redacción al capturar (`design.md` — D9) y verificar con un test que ni el resumen ni el original conservado contienen el valor sin redactar, y que el original mantiene la línea y su posición
- [ ] 2.11 Implementar la retención del original en ruta de trabajo local no versionada, mientras el change esté activo y con mínimo de siete días, y verificar que una petición fuera de plazo informa que no está disponible y desde cuándo
- [ ] 2.12 Implementar la declaración de productor en el resumen (`design.md` — D8) y verificar con tests que un resumen conforme se acepta declarando su productor y que uno que omite un fallo se rechaza quedando el productor registrado
- [ ] 2.13 Documentar el contrato de resumen en `docs/` con un ejemplo de entrada y salida, y verificar que el ejemplo documentado coincide con la salida real del fixture 2.1

## 3. Resolución de contexto por acción

> Requiere el mapeo de módulo a acción acordado con 03 y 04 (ver `design.md` — Risks). El eje
> de acciones (explore, propose, apply, verify, archive) ya está fijado por D10.

- [ ] 3.1 Verificar el reparto núcleo/módulo decidido en `design.md` — D2 contra el inventario de módulos de 03 y 04, y registrar las divergencias si las hay; el reparto se ajusta, no se vuelve a decidir
- [ ] 3.2 Declarar en el catálogo qué acciones requieren cada módulo y verificar que `catalog.Load()` rechaza un módulo sin acciones declaradas
- [ ] 3.3 Agregar la acción activa como entrada de la resolución de contexto y verificar con tests que para una misma acción, rol e instalación el conjunto de módulos es idéntico entre dos resoluciones
- [ ] 3.4 Verificar con un test que el orden en que se declaran los módulos no altera el resultado de la resolución
- [ ] 3.5 Implementar la detención por módulo obligatorio ausente y verificar que el sistema informa qué módulo falta y qué acción queda detenida, sin ejecutar la acción
- [ ] 3.6 Verificar con tests que una acción sin requisito de TDD no recibe el módulo de TDD y que una acción que coordina subagentes sí recibe el de delegación
- [ ] 3.7 Verificar que el núcleo obligatorio está presente en sesión normal, en sesión con orquestador y en sesión de especialista
- [ ] 3.8 Declarar la paridad por cliente y verificar que un módulo que un cliente no puede recibir queda registrado como límite en lugar de presentarse como entregado
- [ ] 3.9 Documentar en `docs/` la tabla de acción por módulo y por cliente, y verificar que se genera desde el catálogo en lugar de mantenerse a mano

## 4. Lectura dirigida y procedencia

- [ ] 4.1 Declarar el umbral de lectura completa en el catálogo, en líneas y con 300 como valor (`design.md` — D11), y verificar que `catalog.Load()` rechaza un umbral expresado en otra unidad
- [ ] 4.2 Agregar un test que mida la distribución de líneas de los archivos del repositorio y falle si el umbral declarado deja de estar entre el p85 y el p92, de modo que el número siga justificado cuando el repositorio crezca
- [ ] 4.3 Declarar en el catálogo la lista cerrada de cuatro condiciones de lectura completa (`design.md` — D15) y verificar que `catalog.Load()` rechaza una condición fuera de la lista
- [ ] 4.4 Declarar las reglas de lectura por ruta, rango y símbolo conforme a `specs/directed-code-reading` y verificar con un test que una lectura por símbolo devuelve el símbolo completo con su procedencia, la declaración que lo contiene si está anidado, y nada más
- [ ] 4.5 Implementar el registro de lectura no justificada y verificar con tests las cuatro condiciones: por debajo del umbral, reescritura completa, artefacto de la acción activa y lectura acotada previa vacía; y que una lectura fuera de las cuatro queda registrada
- [ ] 4.6 Implementar la exigencia de procedencia y verificar con un test que una afirmación sobre el código sin ruta ni línea ni símbolo queda marcada como no verificada
- [ ] 4.7 Implementar la orientación acotada a nivel repositorio y verificar con un test que ubicar una capacidad desconocida devuelve rutas candidatas, cada una con su tipo de coincidencia y su cantidad, sin leer el contenido de los archivos descartados
- [ ] 4.8 Verificar con un test el orden entre tipos (`design.md` — D13): una ruta con la definición del símbolo se ordena antes que otra con muchas referencias, y la cantidad solo ordena dentro del mismo tipo
- [ ] 4.9 Verificar con un test que recorrer el repositorio archivo por archivo solo para orientarse queda registrado como lectura no justificada
- [ ] 4.10 Implementar el registro de relectura y verificar con tests que una segunda lectura del mismo archivo y rango sin escritura intermedia queda registrada como llamada innecesaria, y que la relectura posterior a una escritura se considera justificada
- [ ] 4.11 Documentar en `docs/` cuándo corresponde cada modo de lectura, el umbral con su justificación medida y las cuatro condiciones, y verificar que los ejemplos documentados corresponden a casos cubiertos por los tests de 4.4 a 4.10

## 5. Herramientas por clase de tarea

> D8 y D9 están decididas y son internas a esta ficha. Con 06 solo se acuerda la interfaz de
> la referencia al original (ver `design.md` — *Decisiones que el grupo tiene que cerrar*). Este grupo no
> queda bloqueado por eso.

- [ ] 5.1 Declarar las clases de tarea acordadas en 3.1 y qué skills e integraciones habilita cada una, y verificar que `catalog.Load()` rechaza una integración sin clase de tarea declarada
- [ ] 5.2 Implementar el valor por defecto desactivado para integraciones opcionales y verificar con un test que una integración instalada y no habilitada no se ofrece en la tarea activa
- [ ] 5.3 Implementar la degradación declarada y verificar que al faltar una integración habilitada el sistema informa qué capacidad queda sin cubrir y el camino alternativo si existe
- [ ] 5.4 Verificar con un test que al faltar una integración sin camino alternativo el resultado parcial no se presenta como completo
- [ ] 5.5 Verificar de punta a punta que el acceso y la retención de la tarea 2.10 se respetan cuando el original se pide desde una clase de tarea distinta de la que lo generó
- [ ] 5.6 Implementar el registro de invocación redundante y verificar con un test que una segunda invocación de la misma integración con la misma entrada, sin cambio en el estado que observa, queda registrada como llamada innecesaria
- [ ] 5.7 Verificar con un test que una integración ya informada como no disponible no se vuelve a invocar en la misma tarea y que se usa el camino alternativo declarado si existe
- [ ] 5.8 Documentar en `docs/` la tabla de clase de tarea por herramienta con su estado por defecto, y verificar que se genera desde el catálogo

## 6. Comparación de integración, con el protocolo de 02

- [ ] 6.1 Preparar la corrida de una misma clase de tarea con y sin cada integración opcional, con la misma definición de aceptación, y verificar que la definición de aceptación es byte a byte idéntica entre ambas corridas
- [ ] 6.2 Registrar el contexto efectivo de cada corrida con el protocolo de 02 y verificar que la diferencia observada queda atribuida a la integración y no a un cambio de criterio
- [ ] 6.3 Verificar la resolución de contexto de punta a punta: comparar una sesión normal con una que invoque al orquestador en cada cliente soportado, en instalación limpia y después de actualizar una instalación anterior, y comprobar qué módulos recibe cada una

## 7. Respuestas del agente que no entierran el hallazgo

- [ ] 7.1 Escribir los fixtures de respuesta: un hallazgo informado con el recorrido primero y otro con el hallazgo primero, sobre el mismo caso, y verificar que los dos describen exactamente el mismo hallazgo con la misma procedencia
- [ ] 7.2 Implementar el orden hallazgo-primero conforme a `specs/concise-agent-reporting` y verificar con un test que la respuesta abre con el hallazgo y su ruta y línea, y que el recorrido no lo precede
- [ ] 7.3 Verificar con un test que varios hallazgos se informan como lista con su procedencia y que la lista precede a cualquier desarrollo
- [ ] 7.4 Verificar con un test que el desarrollo no viaja en la respuesta inicial y que se produce cuando se lo pide, sobre la misma procedencia
- [ ] 7.5 Verificar con un test que cuando el contenido ya está en un artefacto del proyecto la respuesta lo señala en lugar de reproducirlo, y que cuando no está escrito en ninguno sí lo contiene
- [ ] 7.6 Verificar con un test que una revisión sin hallazgos informa en una línea qué se revisó, sin narrar el recorrido
- [ ] 7.7 Documentar en `docs/` el contrato de respuesta con el par de ejemplos del fixture 7.1, y verificar que el ejemplo documentado coincide con la salida real

## Workflow follow-up

- Resolver en el grupo los seis puntos de `design.md` — *Decisiones que el grupo tiene que cerrar* antes de compartir la propuesta final, según el paso 4 de la guía. El inventario de módulos y los agentes incluidos no son de esta ficha: se alinean con la 04 y con el roadmap. Los otros cuatro están decididos acá y lo que se confirma es que nadie va en otra dirección. Compartir el ID del change, la decisión tomada, la alternativa descartada y el impacto si se rechaza. Con 06 es una interfaz —el formato de la referencia al original y su vencimiento—, no una aprobación: las dos fichas son hermanas y ninguna depende de la otra. Con 03 y 04 es alineación, porque esta ficha sí depende de ellas.
- Proponer en grupo que 03 y 04 queden como dueñas de `sdd-orchestrator.md`, con las demás fichas escribiendo requisitos sobre el archivo en lugar de contenido.
- Registrar en grupo la asimetría de coordinación con 07, que consume el resultado de esta ficha sin listarla en sus decisiones.
- Compartir la rama `propose/opsx-task-context` para la revisión de los siete. Desde `1b4639b` el repo versiona `openspec/`, así que los artefactos viajan con la rama y ya no hace falta adjuntar la carpeta a mano; indicar la revisión de partida `1b4639b`.
- Pedir lo que esta ficha necesita de las fichas de las que depende: a Matías (02) el protocolo
  de medición, y plantearle que el criterio de cierre de su ficha exige comparar contra el flujo
  modular, que no existe hasta que 03 y 04 estén implementadas, mientras ellas declaran depender
  de la 02; a Juani (03) y Giuliano (04) el inventario de módulos, las superficies soportadas y
  qué módulos no puede recibir Codex, con la matriz compartida saliendo una sola y no dos.
- Avisar, sin pedir decisión: la ficha 08 declara depender de la 06 cuando el acoplamiento real
  es con la 01 —ambas tocan `permissions/install.go` y la 01 decide la política que la 08
  verifica—; cinco fichas cierran con criterios de ejecución en una ronda de solo propose (02,
  05, 07, 09, 10) y solo la 02 lo resuelve internamente; el `(TBD)` es el criterio de aceptación
  en la 01 y en la 09, lo que por la regla del paso 5 les impide ir a revisión; y el piso de
  permisos declara como dato sensible solo `.env` y `.env.*` (`overlays.go:16`), así que la
  redacción de esta ficha protege solo eso hasta que la 01 lo amplíe.
- Registrar lo acordado en el grupo —fecha, con quién y en qué artefacto quedó incorporado—
  antes de cerrar cada punto. Lo ya resuelto sin consultar: la lista de datos sensibles se tomó
  del código (`overlays.go:16`) en lugar de pedirla, y el roadmap canónico pasó a `CHANGES.md`
  en `1b4639b`.
- Archivar el change una vez satisfechos los requisitos de revisión del proyecto.
