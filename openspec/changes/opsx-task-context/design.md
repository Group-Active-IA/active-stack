# Design

## Context

Ver `proposal.md` — Why e Impact para la motivación y la medición de partida.

Tres restricciones del estado actual condicionan todo el diseño:

1. **El mecanismo de composición ya existe y es puro.** `internal/harness/config/compose.go` ensambla el bloque de forma determinista: dos toggles sustractivos (`delegation`, `model-routing`, que quitan secciones marcadas del archivo base) y tres aditivos (`governance`, `engram`, `tdd`, que se concatenan en orden fijo). La función no hace I/O y el mismo input produce siempre el mismo output. No hay que construir un mecanismo de carga modular: hay que darle un selector.
2. **El selector no existe.** Claude y Codex componen en la rama `default` de `internal/harness/config/install.go:70`, que invoca `Compose(h.Toggles, …)` con los toggles literales del catálogo; OpenCode pasa por la rama `ConfigDeliveryPrimaryAgent` (`install.go:53`, que elige `internal/agents/opencode/adapter.go:113`) con los mismos toggles. No hay bandera de CLI, pantalla de TUI ni campo de configuración que los filtre, y el `CLAUDE.md` del proyecto lo declara explícitamente (`CLAUDE.md:82`): los toggles del orquestador no se exponen en la TUI ni en la GUI. El resultado es que toda instalación produce el bloque máximo.
3. **La decisión ocurre en el momento equivocado.** La composición se congela al instalar y se inyecta con markers en `CLAUDE.md`/`AGENTS.md`. En ese momento no se conoce la acción activa, así que no hay información con la que decidir. Mover la decisión a donde la acción sí se conoce es el núcleo de este change.

## Goals / Non-Goals

**Goals:**

- Especificar la frontera entre lo que es obligatorio en toda sesión y lo que se resuelve por acción, de forma que la frontera sea verificable y no una recomendación.
- Especificar un contrato de resumen de salida que se pueda probar con fixtures antes de que exista implementación.
- Dejar el inventario de partida medido y reproducible, para que 02 pueda comparar contra él y para que la discusión de ahorro deje de ser una estimación.
- Entregar el diseño de los fixtures y el criterio de aceptación; las corridas se planifican como tareas.

**Non-Goals:**

- No se diseña ningún subcomando nuevo del binario. Los *Límites compartidos* del roadmap (`CHANGES.md:49`) dicen que *"un ledger firmado, un runtime nuevo o una integración permanente del juez requieren una decisión de arquitectura separada"* y que *"el arnés de evaluación queda fuera del binario"*. Un subcomando que capture y resuma salidas sería ese runtime nuevo. Además, el informe ejecutivo advierte que un envoltorio de estado no debe crear una segunda autoridad frente a OpenSpec, que `ARCHITECTURE.md:165` declara fuente de verdad del estado. El resumen de salida se especifica como **contrato de formato**: qué debe conservar, qué referencia debe llevar, qué se redacta. Quién lo produce se decide en D8 y quién ejecuta cada regla en D17.
- No se reestructura `sdd-orchestrator.md`. Ese artefacto lo tocan seis fichas; su reestructuración en módulos es de 03 y 04. Este change escribe requisitos sobre él y **redacta el texto de sus propios módulos** (lectura dirigida, resumen acotado, selección de herramientas, reporte conciso) como entrega a 03 y 04, que lo ubican (D17).
- No se define ninguna métrica de consumo de modelo. Eso es 02. Los números de este change son de tamaño de texto y están etiquetados como tales.
- No se diseña un almacén de evidencia. La ubicación y la retención del original se deciden en D9 con valores concretos; qué evidencia se considera suficiente para aceptar un cambio sigue siendo de 06.

## Decisions

### D1 — Mover la selección al momento de la acción, conservando la composición determinista

Se conserva el ensamblado determinista actual y se le agrega una entrada: la acción activa. La composición sigue siendo pura y el orden sigue siendo fijo, de modo que la resolución es reproducible y comparable entre corridas.

*Alternativa descartada: exponer los cinco toggles como banderas de instalación.* Es más barata y el plan previo del proyecto la propone. Se descarta porque resuelve otro problema —permite medir si un componente vale su costo, pero sigue congelando la elección al instalar: un equipo que active TDD porque lo usa en apply lo carga igual en explore— y porque el proyecto decidió que los toggles del orquestador no se exponen en la TUI ni en la GUI (`CLAUDE.md:82`). Este change no la propone ni como paso intermedio. La comparación con y sin cada módulo que 02 necesita no depende de banderas: se arma con el arnés de evaluación fuera del binario (tarea 6.4).

*Alternativa descartada: dejar el bloque completo y confiar en la caché del proveedor.* El informe ejecutivo ya matizó esa afirmación: el proveedor reutiliza bajo condiciones y el ensamblado lo hace el host. Depender de eso convierte una propiedad de diseño en una esperanza sobre un tercero. La estabilidad de prefijos y caché es materia de 09, como variable independiente.

### D2 — El núcleo obligatorio se define por consecuencia, no por tamaño

La frontera entre núcleo y módulo no se fija con un presupuesto de bytes ni de líneas. Una instrucción pertenece al núcleo si su ausencia puede producir daño o invalidar el trabajo con independencia de la acción; pertenece a un módulo si su pertinencia depende de la acción activa.

*Por qué no un presupuesto numérico:* el informe ejecutivo marcó los umbrales de 40 líneas y 1.500 tokens como presupuestos de diseño, no como límites validados. Adoptar un número como criterio reintroduce exactamente la afirmación que la revisión rechazó. Los números siguen siendo útiles como hipótesis a medir con 02, y así quedan registrados en `proposal.md`.

Aplicando el criterio al inventario medido, el reparto es el siguiente (**decidido por la 05 — pendiente de revisión del grupo**):

| Fragmento | Reparto | Razón |
|---|---|---|
| reglas de cumplimiento obligatorio | núcleo | su ausencia invalida el trabajo en cualquier acción |
| `governance` (1.276 B) | núcleo | `governance.md:3`: *"It applies to **all agents**, regardless of provider"*; define cuánta autonomía tiene el agente sobre dominios sensibles y eso no depende de la acción |
| `engram` — regla corta | núcleo | `engram-protocol.md:4`: *"This protocol is MANDATORY and ALWAYS ACTIVE"*. Al núcleo va solo lo que lo hace siempre activo: la memoria existe, se consulta al empezar y se guarda ante decisiones |
| `engram` — protocolo detallado | módulo | formato de guardado, cierre de sesión, recuperación tras compactación y git sync; se carga al inicio y al cierre de cada acción |
| `delegation` (1.005 / 851 B) | módulo | solo aplica cuando la acción coordina subagentes; incluye la coordinación de subagentes, como lo trata la 07 (D19) |
| `model-routing` (778 / 450 B) | módulo | solo aplica cuando hay delegación; se carga junto con `delegation` |
| `tdd` (3.797 B) | módulo | solo aplica en `apply` con Strict TDD activo y runner de tests disponible |
| `design-mockups` (8.865 B, `assets/design-mockups.md`) | fuera de alcance | es un harness propio con su sección `active-stack:design-mockups` (registrada en `sectionFragments`, `internal/harness/config/assets.go:29`) y política por proyecto; no es un toggle del orquestador |

**Por qué `engram` se parte en lugar de ir entero a un lado.** El dato medido es que `engram` son 6.093 B, el 30,4 % del bloque de Claude y el 33,8 % del de Codex (`proposal.md` — Impact). Entero al núcleo, esa fracción vuelve a estar en toda sesión; entero a módulo, contradice `engram-protocol.md:4`. Partido, al núcleo va solo la regla corta; su tamaño queda registrado en el golden de la tarea 1.1 cuando se redacte (tarea 3.10).

**Divergencias registradas con 03 y 04.** La 03 propone `governance` dividido —`governance.md:1-13` como módulo de propose y `:14-19` como coordinador— y `engram` partido entre universal (encabezado y guardado proactivo) y coordinador (wiring OPSX, cierre de sesión), todo marcado `(TBD grupo)` en su tabla (`claude-native-opsx/design.md:287-292`). La 04 deja G4 abierto entre *universal · módulo · híbrido*, con recomendación híbrido sujeta a 02 (`codex-native-opsx/design.md:331`). El reparto de la 05 **converge** con las dos en `engram` (partido, coincide con el corte universal/coordinador de la 03 y con el híbrido de la 04) y **diverge** de la 03 en `governance`: acá va entero al núcleo porque su propio texto dice que aplica a todo agente. La tarea 3.1 registra la divergencia contra la matriz única (punto *Matriz de módulos* abajo); el reparto se ajusta si el grupo lo rechaza, no queda abierto esperándola.

### D3 — El resumen de salida se especifica por lo que debe conservar, no por cuánto debe recortar

Un límite de tamaño es el modo de fallo que la ficha nombra: una salida extensa puede ocultar un fallo. Por eso el contrato se escribe en términos de retención obligatoria (errores, fallos, código de salida, redacción de datos sensibles) y de recuperabilidad del original, y no en términos de un tope de líneas. Un resumen que cumple la retención y pierde volumen es correcto; uno que cumple un tope y pierde un fallo es incorrecto.

El único número del contrato es **cuándo** se resume, no **cuánto**: una salida de más de 300 líneas se entrega como resumen más referencia (D19). Es una condición de entrada, no un tope de recorte.

*Alternativa descartada: recortar por cabeza y cola.* Falla justamente en el caso del fallo central, que es uno de los tres que la ficha pide probar.

### D4 — Los fixtures son el entregable verificable de esta ronda

Como el change no implementa, el artefacto que demuestra que la especificación es verificable son los fixtures. Se diseñan con el fallo en tres posiciones (inicio, medio, final), con varios fallos dispersos, y con una salida verbosa sin fallos como control negativo. Cada fixture declara su salida original, su resumen esperado y el fallo que debe sobrevivir.

### D5 — Procedencia obligatoria en lugar de umbral de lectura como regla principal

La regla que más evita trabajo inválido no es "leé menos" sino "toda afirmación sobre el código indica ruta y línea o símbolo". El umbral de lectura completa es un apoyo; la procedencia es lo que hace falsable cada hallazgo. Se especifican las dos, con la procedencia como requisito normativo y el umbral como condición declarada en el catálogo.

### D6 — La llamada innecesaria se registra, no se bloquea

Una relectura o una invocación repetida se marca como llamada innecesaria y queda registrada; nada la impide. Bloquearla exigiría decidir con certeza que el estado observado no cambió, y hay casos legítimos donde sí cambió por fuera de la sesión. Registrar convierte el problema en una medida —cuántas llamadas innecesarias por tarea aceptada— que 02 puede comparar entre configuraciones, que es el uso que tiene valor. Qué significa "registrar" lo fija D17.

*Alternativa descartada: bloquear la segunda llamada.* Un falso positivo deja al agente sin un dato que sí cambió, y el daño de eso es mayor que el de una llamada repetida. La excepción es el sondeo de una integración ya informada como no disponible: ahí el estado sí es conocido dentro de la tarea, y el requisito es no volver a invocarla.

### D7 — El inventario de verbosidad cubre todas las fuentes, no solo el orquestador

La ficha pide localizar dónde aparecen instrucciones largas o resultados verbosos (`CHANGES.md:162`). El orquestador es la fuente más grande y ya está medida, pero no es la única: los cuerpos de skill y los fragmentos de sección propia que el instalador escribe en el proyecto destino también ocupan contexto en toda sesión, y las salidas verbosas de comandos (tests, build, logs) aparecen en puntos concretos del ciclo. El inventario se completa sobre todas las fuentes —instrucciones (tareas 1.1 a 1.7) y salidas (tarea 1.8)— y se publica como una tabla única ordenada por tamaño, porque el criterio para elegir qué se resuelve por tarea debe apoyarse en el total real y no en el fragmento que resultó más fácil de medir.

Este inventario es una actividad de investigación, no un comportamiento del sistema: por eso vive en `tasks.md` (grupo 1) y en `proposal.md` — Impact, y no como requisito de una spec. Se propone a la 02 reutilizarlo para su inventario de instrucciones (`opsx-baseline/design.md:46`) en lugar de duplicarlo.

### D8 — El resumen lo produce el agente bajo instrucción, y el contrato no depende del productor

De las tres opciones, el envoltorio ejecutable sería un runtime nuevo, que `CHANGES.md:49` deja sujeto a una decisión de arquitectura separada, y la capacidad nativa del cliente no existe por igual en Claude y Codex, así que elegirla rompería la paridad que esta ficha necesita por ser de ambos clientes. Queda el agente bajo instrucción, que está disponible en los dos y no agrega autoridad nueva.

La objeción válida es que una instrucción no es un control: el agente puede no retener el fallo. Se resuelve así: **el contrato y los fixtures verifican el resumen, no al productor.** Un resumen se acepta o se rechaza por lo que conserva, con independencia de quién lo escribió. Y el resumen declara quién lo produjo, de modo que un incumplimiento es atribuible. Sustituir mañana la instrucción por un envoltorio ejecutable no cambia una sola spec ni un solo fixture: cambia el productor y los mismos fixtures lo evalúan.

**Supuesto declarado:** que el agente bajo instrucción retiene los fallos de los fixtures. Nadie lo midió todavía. Se valida corriendo los fixtures del grupo 2 contra un agente real con el arnés fuera del binario (tarea 2.7); si los fixtures del fallo central o de ruido fallan, el resultado se lleva al grupo como insumo para la decisión de arquitectura que `CHANGES.md:49` reserva.

*Alternativa descartada: esperar a que exista el envoltorio.* Dejaría la ficha sin entregable verificable en esta ronda, que es justamente lo que D4 evita.

### D9 — El original se redacta antes de guardarse, vive fuera del repo y vence después de archivar

Tres decisiones unidas (**decidido por la 05 — pendiente de revisión del grupo**).

**Redacción al capturar.** Los valores que coincidan con la lista de patrones sensibles (D14) se redactan antes de escribir el original a disco, no solo al resumirlo. No existe entonces ninguna copia en disco con secretos esperando a que venza una retención. Nadie necesita el valor de un secreto para diagnosticar un fallo; necesita saber que había uno y dónde.

**Referencia y ubicación.** Todo resumen lleva una referencia con cinco campos: `{comando, id de corrida, ruta, huella sha256 del original ya redactado, fecha de vencimiento}`. Una sola huella alcanza porque nunca existe en disco un original sin redactar que haya que distinguir del redactado. El original vive en la caché del usuario, fuera del repositorio: `<os.UserCacheDir>/active-stack/runs/<change>/`. Desde `1b4639b` `openspec/` se versiona, así que una salida guardada dentro del change podría commitearse por error; fuera del repo no puede.

**Retención y acceso.** El original se conserva hasta que el change se archive **y siete días más**; no hay otro plazo, así que no hay prelación que resolver. Los siete días son un **supuesto**: no los respalda ninguna medición —a diferencia del umbral de D11—. Se recalibran con el primer uso real: si nadie pide un original más de tres días después de archivar, se baja; si alguien lo pide a los diez, se sube. Vencido el plazo, la referencia informa que el original no está disponible y desde qué fecha. El acceso es el del usuario local con los permisos de sistema de archivos de su directorio de caché: el original no se publica ni se envía a ningún lado.

**Diferencias con la 06**, que es hermana y define la evidencia. Coincidimos en no versionar las salidas completas por defecto (`opsx-verifiable-evidence/design.md:78`). Diferimos en cuatro puntos, y se le propone a la 06 adoptar los de acá:

| | 06 (`opsx-verifiable-evidence/design.md`) | 05 |
|---|---|---|
| cuándo se redacta | antes de compartir la salida (`:79`) | antes de escribirla a disco |
| huellas | dos: la original y la de la versión redactada (`:79`) | una: la del original ya redactado |
| dónde vive | subdirectorio local de la evidencia, bajo `openspec/changes/<change>/evidence/` (`:73`, `:78`) | `<os.UserCacheDir>/active-stack/runs/<change>/`, fuera del repo |
| retención | `(TBD)`, *"por ejemplo, hasta archivar el change"* (`:80`) | hasta archivar y siete días más, compatible con su ejemplo |
| formato de referencia | ruta, huella y si fue truncada (`:46`) | comando, id de corrida, ruta, huella sha256, vencimiento |

*Alternativa descartada: conservar el original sin redactar y controlar el acceso.* Un control de acceso sobre un archivo local es una ficción; redactar en origen elimina el problema en lugar de administrarlo.

### D10 — Las clases de tarea son las cinco acciones del ciclo OPSX

El eje de clasificación no se inventa: son `explore`, `propose`, `apply`, `verify` y `archive` (`CLAUDE.md:26`, `ARCHITECTURE.md:6`, `ARCHITECTURE.md:164`). Es el eje que el proyecto ya declara, que OpenSpec ya conoce y que el orquestador ya usa para ubicarse, de modo que declarar un eje propio habría creado un segundo vocabulario para lo mismo (**decidido por la 05 — pendiente de revisión del grupo**).

Las cinco tareas de medición de la ficha 02 (exploración, cambio pequeño, cambio multiarchivo, diagnóstico, revisión) no son clases de tarea: son **cargas de trabajo** que se ejecutan dentro de esas acciones y sirven para comparar. Se mapean sobre las acciones, no las reemplazan.

*Alternativa descartada: un eje propio por tipo de trabajo (lectura, escritura, diagnóstico).* Más expresivo, pero obliga a traducir entre dos vocabularios en cada ficha que toque el orquestador.

*Alternativa descartada: OPSX más un sub-eje de tipo de trabajo dentro de cada acción.* La objeción que la motivaba —analizar un log y escribir documentación caen las dos en `apply` y necesitan herramientas distintas— queda cubierta sin sub-eje por dos piezas que ya existen: la tabla de herramientas por clase (D18), que habilita las condicionales solo cuando la acción las necesita (Playwright si toca UI), y las señales de delegación de la 07 (`opsx-selective-delegation/design.md:40-43`), que distinguen una tarea verbosa de una que no lo es. Un sub-eje agregaría un tercer vocabulario sin cambiar ninguna de las dos.

### D11 — El umbral de lectura completa es 300 líneas, medido sobre este repositorio

**En líneas**, porque es la unidad en la que el agente pide un rango y en la que se expresa la procedencia que D5 hace obligatoria. Los bytes no se corresponden con nada que el agente pueda pedir; los símbolos no existen en archivos sin parser disponible.

**300**, porque sale de medir el repositorio y no de estimar. Conjunto medido: los 130 archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en `_test.go`, en `1b4639b` (código idéntico en `b74fd76`), contando líneas con `wc -l`. Percentiles por rango más cercano; la mediana, como promedio de los dos valores centrales:

| | líneas |
|---|---|
| mediana (p50) | 85,5 |
| p75 | 178 |
| p85 | 244 |
| p90 | 311 |
| p92 | 375 |
| p95 | 400 |
| máximo | 809 (`internal/tui/model.go`) |

| Umbral | Archivos que se leen enteros (≤ umbral) |
|---|---|
| 200 | 104 de 130 (80 %) |
| **300** | **116 de 130 (89 %)** |
| 400 | 124 de 130 (95 %) |
| 500 | 126 de 130 (97 %) |

Descarté 400 —que había puesto primero— justamente por medirlo: a 400 la regla toca solo 6 archivos y no cambia el comportamiento en la práctica. Y descarté 200 porque fuerza lectura acotada en archivos donde leer entero sale más barato que dos viajes de ida y vuelta: un archivo de 200 líneas son unos 2.500 tokens, y dos consultas acotadas con riesgo de una tercera no ahorran nada.

300 queda entre el p85 (244) y el p92 (375), justo debajo del p90: deja pasar el 89 % de los archivos, que son chicos y baratos, y reserva la lectura acotada para los 14 que son genuinamente grandes, donde sí paga. La regla tiene que ser **rara y dirigida**, no una fricción constante.

El valor vive en el catálogo y se ajusta sin tocar ninguna spec. La regla de recalibración es mecánica y no depende de otra ficha: la tarea 4.2 mide el mismo conjunto en cada corrida y falla si 300 sale del intervalo p85–p92; en ese caso se elige el nuevo valor dentro del intervalo con el mismo criterio y se actualiza este cuadro.

### D12 — Esta ficha no fija ni propone un tope para los cuerpos de skill

El inventario de las tareas 1.5 a 1.7 publica el tamaño de cada cuerpo de skill, que es el entregable. Fijar un tope es política del catálogo y afecta a toda ficha que instale skills; adoptarlo acá sería legislar para las otras nueve desde la ficha 05.

**Decisión:** la 05 entrega la medición y declara el tope fuera de su alcance. No lo propone como cambio de catálogo. El inventario queda publicado para que cualquier ficha o change posterior que quiera proponer un tope parta de esos números y no de una estimación.

### D13 — La señal de pertinencia son el tipo de coincidencia y su cantidad

La orientación a nivel repositorio devuelve rutas candidatas, y cada una viene con dos datos: **qué tipo de coincidencia** tiene y **cuántas**. El tipo es uno de tres, ordenados por peso:

| Tipo | Qué es | Peso |
|---|---|---|
| definición | el símbolo se define en ese archivo | alto |
| referencia | el símbolo se usa en ese archivo | medio |
| texto | coincidencia literal sin estructura | bajo |

Una definición pesa más que cualquier cantidad de referencias, y una referencia más que cualquier cantidad de coincidencias de texto. Dentro del mismo tipo, ordena la cantidad.

*Por qué estos tres y no una puntuación:* una puntuación numérica obliga a justificar los pesos y no es inspeccionable. Tres categorías ordenadas se explican en una línea y el agente puede decidir con ellas sin que nadie calibre nada. Y los tres son obtenibles sin leer el contenido de los archivos descartados, que es el requisito que esta señal tiene que respetar.

### D14 — Una sola lista de patrones de valor sensibles, declarada en el catálogo y con dueña la 05

La redacción no usa las clases del piso de permisos: el piso bloquea **rutas** (`Read(.env)`, `Edit(.env.*)`, `internal/harness/config/permissions/overlays.go:16-28`) y la redacción tiene que reconocer **valores** dentro de una salida. Además el piso aplica solo a Claude (`composeClaudeOverlay`, `overlays.go:73`); en Codex el instalador escribe únicamente `approval_policy` y `sandbox_mode` (`permissions/install.go:114-134`), así que no hay lista que heredar en ese cliente.

**Decisión** (**decidido por la 05 — pendiente de revisión del grupo**): una sola lista de **patrones de valor**, declarada en el catálogo, cuya dueña es esta ficha y que usan también 02, 06 y 10. Contenido inicial:

- `.env` y `.env.*`: el archivo y los valores de sus asignaciones
- tokens y API keys
- credenciales de nube
- claves SSH y `*.pem`
- valores de variables de entorno
- rutas del home del usuario
- identificadores de cuenta

La lista coincide con lo que la 06 ya redacta antes de compartir (variables de entorno, tokens, credenciales, rutas del home, `opsx-verifiable-evidence/design.md:79`) y con lo que la 10 necesita para su *"derivado redactado con lista de inclusión"* (`opsx-shadow-review/design.md:126`). Un test verifica que todo lo que bloquea el piso de Claude está cubierto por la lista (tarea 2.9): si el piso agrega una regla que la lista no cubre, la validación falla en lugar de divergir en silencio. Llevar un piso equivalente a Codex no es de esta ficha: se le pide a la 01 o a la 08. La 01 (`propose/codex-strict-compat`) hoy no toca la lista sensible.

*Alternativa descartada: reusar el piso de permisos como única lista.* Era la decisión anterior. Se descarta porque mezcla dos objetos —rutas que se bloquean y valores que se redactan—, porque el piso es de Claude solo, y porque con el piso actual la redacción protegería únicamente `.env`.

### D15 — La lista de condiciones para leer un archivo completo es cerrada

Son cuatro, y lo que no está acá no justifica lectura completa:

1. el archivo no supera el umbral de D11
2. la tarea va a reescribir el archivo entero
3. el archivo es el artefacto que la acción activa produce o edita
4. una lectura acotada previa sobre ese archivo devolvió vacío

La primera es la regla general: un archivo de 300 líneas o menos se lee completo. Las otras tres son las únicas excepciones para un archivo que supera el umbral, que en cualquier otro caso se lee por ruta, rango o símbolo.

*Por qué cerrada:* una lista abierta vuelve inaplicable el registro de lectura no justificada — siempre se puede alegar una condición nueva. Cerrada, el registro es decidible: o la lectura entra en una de las cuatro o queda marcada.

La cuarta existe para no dejar al agente trabado: si pidió un símbolo y no estaba, leer entero es la salida razonable y no debería contar como falta.

Una lectura por símbolo devuelve el símbolo completo y, si está anidado, la **firma o encabezado** de la declaración que lo contiene —no su cuerpo—, para que la procedencia sea interpretable sin leer los vecinos.

### D16 — La respuesta del agente es una capacidad aparte, no parte del resumen de comandos

La ficha pide localizar *"instrucciones largas o resultados verbosos"* (`CHANGES.md:162`), y hay dos clases de resultado verboso: la salida de un comando y la respuesta del propio agente. Las dos entierran el hallazgo, pero **no se arreglan con el mismo contrato**, y por eso van en capacidades separadas.

| | Salida de comando | Respuesta del agente |
|---|---|---|
| quién la produce | una herramienta | el modelo |
| hay un original completo | **sí**, y se conserva | **no**, nunca existió |
| cómo se accede al detalle | referencia al original guardado | se produce a pedido |
| qué se puede perder | una línea de fallo | el hallazgo, enterrado en el relato |

La diferencia que decide es la segunda: **no se puede referenciar el original de algo que nunca se produjo.** Para una salida de comando la regla es conservar y referenciar; para una respuesta es poner el hallazgo primero y generar el desarrollo recién cuando se lo piden.

La regla se vuelve observable así: **la primera línea del reporte dice el resultado; la segunda, la consecuencia o el próximo paso; no se repite el desarrollo de pasos anteriores.** Se evalúa con una rúbrica aplicada sobre los fixtures del grupo 7, corridos con el arnés fuera del binario; no con un test de Go, porque lo que se evalúa es conducta del modelo.

*Alternativa descartada: ampliar `bounded-command-output` y renombrarla.* Tentador por lo parecido del síntoma, pero el nombre dejaría de describir lo que la capacidad hace y el requisito de original recuperable quedaría sin sentido para la mitad de los casos.

Una de las reglas de esta capacidad sale de una corrección de esta misma sesión: una respuesta no reproduce lo que ya está escrito en un artefacto, lo señala. Vale para lo que el agente le dice a una persona y para lo que un artefacto le copia a otro.

### D17 — "El sistema" es el agente siguiendo los módulos más lo declarado en el catálogo

Las specs dicen "el sistema registra", "el sistema informa", "el sistema rechaza". Sin runtime nuevo (`CHANGES.md:49`) hay que decir quién es (**decidido por la 05 — pendiente de revisión del grupo**):

- **"El sistema"** es el agente siguiendo los módulos de instrucción, más lo declarado en el catálogo (`internal/catalog/harnesses.yaml`). No hay subcomando, hook ni proceso nuevo.
- **"Registrar"** es que el agente lo declare en su reporte de la acción, en una línea con forma fija (por ejemplo, `llamada innecesaria: <herramienta> <entrada>`), para que el arnés de evaluación lo cuente.
- **Tests de Go** verifican solo lo declarativo: el catálogo validado por `catalog.Load()`, el texto compuesto por `Compose`, y los fixtures como archivos (que existen, que declaran lo que dicen declarar).
- **El comportamiento del agente** se verifica corriendo los fixtures contra el agente real con un arnés fuera del binario, como manda `CHANGES.md:49` y como lo planifica la 02. Es etapa posterior a esta ronda.
- **La 05 redacta el texto** de los módulos de lectura dirigida, resumen acotado, selección de herramientas y reporte conciso, y lo entrega a 03 y 04, que lo ubican en la estructura de módulos. La 05 no escribe en `sdd-orchestrator.md`.

Cada tarea que antes decía "Implementar…" pasa a ser tres piezas: redactar el texto del módulo, declarar en el catálogo lo que corresponda y escribir el fixture que lo evalúa.

*Alternativa descartada: implementar los registros y la orientación como código del instalador.* Es el runtime nuevo que el roadmap reserva para una decisión de arquitectura separada.

### D18 — Tabla de herramientas por clase de tarea

**Decidido por la 05 — pendiente de revisión del grupo.** Toda integración opcional está apagada por defecto y se habilita por clase. Las herramientas base del cliente (lectura, búsqueda, edición, ejecución de tests, OpenSpec CLI) no son integraciones opcionales: la tabla dice cuáles corresponden a cada acción para que el agente no las use fuera de lugar, pero no las instala ni las apaga.

| Acción | Base | Opcionales habilitadas |
|---|---|---|
| `explore` | lectura dirigida, búsqueda | Context7 (docs de librerías) |
| `propose` | lectura dirigida, OpenSpec CLI | Context7 |
| `apply` | edición, tests | Context7; skills de dominio del registro (`go-testing` en este repo, `CLAUDE.md:160`); Playwright solo si la acción toca UI |
| `verify` | tests, lectura dirigida | Playwright solo si hay UI |
| `archive` | OpenSpec CLI | ninguna |

- **Engram** no figura como integración opcional: llega por el núcleo y el módulo de D2.
- **Integraciones de diseño** (Stitch, `emil-design-eng`, `design-taste-frontend-v1`) se habilitan solo con el harness `design-mockups` activo, según la política del proyecto; fuera de eso no se ofrecen en ninguna acción.

La tabla se declara en el catálogo (tarea 5.1) y `docs/` la genera desde ahí (tarea 5.8).

### D19 — Interfaz con la 07: umbral de salida verbosa y obligatoriedad de módulos

**Decidido por la 05 — pendiente de revisión del grupo.** La 07 consume esta ficha (`opsx-selective-delegation/proposal.md:4`) y cita dos cosas que hay que fijar:

- **"Salida verbosa"** es una salida de más de 300 líneas, el mismo umbral medido de D11. Es la condición para resumir, no un tope de recorte (D3). Reemplaza "el presupuesto de 05" de la señal S1 de la 07 (`opsx-selective-delegation/design.md:40`), que no existe: D3 rechaza presupuestos de tamaño.
- **El módulo de delegación es obligatorio solo en `verify`**, por la separación de rol de la señal S4 de la 07 (`design.md:43`). En el resto de las acciones es opcional: si falta, el agente avisa y resuelve directo. Así la regla de detener de `task-context-loading` —respaldada por la *Aceptación transversal* (`CHANGES.md:47`: *"Si falta un módulo obligatorio, informar la limitación y detener la acción que lo requiere"*)— aplica a los módulos obligatorios y es compatible con la D2 de la 07 (`design.md:58`: *"se informa y se resuelve directo o se detiene la acción si la acción exige delegar (S4)"*).

Tabla de carga, que es el mapeo que pide la ficha (`CHANGES.md:164`: *"Mapeá con 03, 04 y 07 cuándo se cargan TDD, delegación y coordinación de subagentes"*):

| Módulo | `explore` | `propose` | `apply` | `verify` | `archive` |
|---|---|---|---|---|---|
| `tdd` | — | — | obligatorio si el proyecto tiene Strict TDD activo y runner | — | — |
| `delegation` (con coordinación de subagentes) | opcional, si la regla de la 07 decide delegar | opcional, ídem | opcional, ídem | **obligatorio** (S4) | opcional, ídem |
| `model-routing` | con `delegation` | con `delegation` | con `delegation` | con `delegation` | con `delegation` |
| `engram` — protocolo detallado | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre |
| `governance`, `engram` — regla corta | núcleo | núcleo | núcleo | núcleo | núcleo |

## Risks / Trade-offs

- **La matriz de módulos de 03 y 04 todavía no está cerrada →** la 05 usa su propio reparto (D2) hasta que exista la matriz única propuesta (punto *Matriz de módulos*), y la tarea 3.1 registra las divergencias contra ella. El criterio de reparto no depende de esa matriz.
- **Seis fichas tocan `sdd-orchestrator.md` y ninguna lo posee →** este change escribe requisitos y redacta texto de módulos como entrega (D17), no contenido en el archivo, y lleva al grupo la propuesta de que 03 y 04 sean dueñas. La 07 ya la respalda (`opsx-selective-delegation/design.md:173`: *"03+04, como propone 05"*).
- **Resolver módulos por acción agrega un punto de fallo que hoy no existe →** mitigado por el requisito de que un módulo obligatorio ausente detenga la acción e informe, que es el modo de fallo que la aceptación transversal ya exige (`CHANGES.md:47`). D19 fija cuáles son obligatorios para que la regla no frene acciones que pueden resolverse directo.
- **Mover instrucciones fuera del bloque siempre presente puede degradar el comportamiento del agente en acciones que sí las necesitaban →** por eso ninguna adopción se decide en este change. La comparación con y sin cada módulo, sobre la misma aceptación, es requisito de `task-context-loading` y usa el protocolo de 02 (tarea 6.4).
- **El contrato de resumen no sirve de nada si el productor no retiene el fallo →** el productor está decidido (D8) y su supuesto se valida corriendo los fixtures del grupo 2 contra un agente real (tarea 2.7). El contrato evalúa el resumen, así que cambiar el productor no invalida nada.
- **Solapamiento con 07 →** la 07 consume esta ficha y ya la declara dependencia. D19 fija el umbral de salida verbosa y qué módulo es obligatorio en qué acción, y la tabla de carga queda acá para que la 07 la consuma sin redefinirla.
- **La redacción por instrucción puede fallar →** un agente puede escribir un valor sensible al disco. Mitigación: la lista de D14 es declarativa y testeable, el fixture 2.4 lo evalúa, y el original vive fuera del repo (D9), de modo que un fallo de redacción no termina en un commit.

## Migration Plan

No aplica en esta ronda: el change no modifica configuración instalada. Cuando se implemente, la migración del bloque administrado en `CLAUDE.md`/`AGENTS.md` pertenece a 03 y 04, que ya la tienen en su alcance con snapshot, preservación de contenido del usuario y rollback. Este change no debe introducir una segunda ruta de migración sobre los mismos archivos.

**OpenCode y Antigravity** mantienen el bloque global actual. Es un límite declarado de la 05: el roadmap es de Claude y Codex (`CHANGES.md:3`) y ninguna ficha tiene asignado qué les pasa al retirar el bloque global; queda sin dueño hasta que el roadmap lo asigne.

## UI Mockups

Decision: no (no UI) — 2026-11-05. El change no agrega ni modifica pantalla, componente, layout ni interacción.

## Decisiones tomadas — pendientes de revisión del grupo

Diez puntos. Todos están decididos acá; se comparten porque tocan alcance de otras fichas. Cada uno con la decisión, en qué se apoya, a qué fichas afecta y qué cambia si el grupo la rechaza. Se identifican por su nombre, no por un código: así nadie tiene que buscar una leyenda y no se confunden con los puntos de otra ficha.

Una decisión puede apoyarse en tres cosas, y cada una lleva su procedencia, igual que D5 exige para cualquier afirmación sobre el código:

| | Qué significa |
|---|---|
| **documentado** | está escrito en el repositorio, en el roadmap o en la rama de otra ficha, con archivo y línea |
| **medido** | lo midió este change y el número es reproducible |
| **supuesto** | no hay respaldo todavía; se declara junto con la decisión y con cuándo se recalibra |

### Matriz de módulos

| | |
|---|---|
| **Decisión** | proponer la matriz de G4 de la 04 (`codex-native-opsx/design.md:300-313`, columnas sesión normal / orquestador / especialista / Codex / Claude) como **matriz única** de 03 y 04; la 03 completa la columna de Claude. Hasta que esté, la 05 usa su reparto (D2) y la tarea 3.1 registra divergencias |
| **En qué se apoya** | **documentado.** La dependencia es circular: la 03 dice que el inventario de módulos lo da la 04 (`claude-native-opsx/proposal.md:78`) y la 04 que lo da la 03 (`codex-native-opsx/proposal.md:71`). La matriz de la 04 es la única publicada con columnas para los dos clientes |
| **Afecta a** | 03, 04, 05, 07 |
| **Si se rechaza** | la tarea 3.1 compara contra la matriz que el grupo elija; ningún requisito cambia |

### Eje de clases de tarea

| | |
|---|---|
| **Decisión** | las cinco acciones OPSX, sin sub-eje (D10) |
| **En qué se apoya** | **documentado** en `CLAUDE.md:26`, `ARCHITECTURE.md:6` y `ARCHITECTURE.md:164`. La 04 usa las acciones OPSX como opción para su inventario de especialistas (G6, `codex-native-opsx/design.md:333`), que es otro eje —quién ejecuta, no qué se carga— y no lo contradice |
| **Afecta a** | 03, 04, 05, 07 |
| **Si se rechaza** | el vocabulario del catálogo en las tareas 3.2 y 5.1 y la tabla de D18; los requisitos dicen "la acción activa" |

### Reparto núcleo / módulo

| | |
|---|---|
| **Decisión** | `governance` y la regla corta de `engram` al núcleo; protocolo detallado de `engram`, `tdd`, `delegation` y `model-routing` a módulo; `design-mockups` fuera de alcance (D2) |
| **En qué se apoya** | **documentado** en `governance.md:3` y `engram-protocol.md:4`; **medido** en `proposal.md` — Impact: `engram` es el 30,4 % del bloque de Claude, reproducible por la tarea 1.1 |
| **Afecta a** | 03, 04, 05 |
| **Si se rechaza** | la tabla de D2, la de D19 y la tarea 3.1; el criterio de reparto no cambia |

### Herramientas por clase de tarea

| | |
|---|---|
| **Decisión** | la tabla de D18: todo opcional apagado por defecto y habilitado por acción |
| **En qué se apoya** | **documentado**: `CHANGES.md:166` pone *"qué capacidad es necesaria para cada tarea"* entre lo que se acuerda en grupo; el catálogo declara Context7 y Playwright (`harnesses.yaml:36`, `:50`). La asignación por acción es **supuesto** razonado; se recalibra con la comparación con y sin cada integración del grupo 6 |
| **Afecta a** | 05, 07 |
| **Si se rechaza** | la tabla de D18 y las tareas 5.1 y 5.8 |

### Referencia, ubicación y retención del original

| | |
|---|---|
| **Decisión** | referencia `{comando, id de corrida, ruta, huella sha256, vencimiento}`; original en `<os.UserCacheDir>/active-stack/runs/<change>/`; retención hasta archivar el change y siete días más; acceso del usuario local (D9). Se propone a la 06 adoptar el formato |
| **En qué se apoya** | **documentado**: `CHANGES.md:166` pone *"dónde queda el original"* entre lo que se acuerda en grupo y pide coordinar la referencia con 06; la 06 tampoco versiona originales (`opsx-verifiable-evidence/design.md:78`). **Supuesto**: los siete días; se recalibran con el primer uso real (D9) |
| **Afecta a** | 05, 06 |
| **Si se rechaza** | los valores del requisito de ubicación y retención y las tareas 2.10, 2.11 y 2.14 |

El commit `1b4639b` no es la fuente de esta decisión: es lo que volvió urgente la pregunta. Antes `openspec/` estaba en el `.gitignore` y una salida guardada en la carpeta del change no podía llegar al repositorio ni por error; ahora sí. Por eso el original vive fuera del repo.

### Datos sensibles

| | |
|---|---|
| **Decisión** | una sola lista de patrones de valor en el catálogo, dueña la 05, usada por 02, 06 y 10; un test garantiza que cubre el piso de Claude (D14) |
| **En qué se apoya** | **documentado**: el piso bloquea rutas y solo en Claude (`overlays.go:16-28`, `:73`); Codex no tiene piso (`permissions/install.go:114-134`); la 06 redacta valores (`opsx-verifiable-evidence/design.md:79`) y la 10 consume una redacción (`opsx-shadow-review/design.md:126`) |
| **Afecta a** | 01, 02, 05, 06, 08, 10 |
| **Si se rechaza** | el requisito de datos sensibles y las tareas 2.4 y 2.9 |

### Productor del resumen acotado

| | |
|---|---|
| **Decisión** | el agente bajo instrucción (D8) |
| **En qué se apoya** | **documentado**: un envoltorio ejecutable sería un runtime nuevo (`CHANGES.md:49`); la capacidad nativa no existe por igual en los dos clientes. **Supuesto**: que el agente retiene el fallo; se valida con la tarea 2.7 |
| **Afecta a** | 05, 06, 07 |
| **Si se rechaza** | el productor y nada más: los fixtures evalúan el resumen |

### Quién ejecuta

| | |
|---|---|
| **Decisión** | "el sistema" es el agente con los módulos más el catálogo; tests de Go solo sobre lo declarativo; conducta con arnés fuera del binario; la 05 redacta el texto de sus módulos y lo entrega a 03 y 04 (D17) |
| **En qué se apoya** | **documentado** en `CHANGES.md:49` (*"el arnés de evaluación queda fuera del binario"*; runtime nuevo sujeto a decisión separada) |
| **Afecta a** | 02, 03, 04, 05 |
| **Si se rechaza** | las tareas que dicen "redactar el texto del módulo" pasan a implementación en el binario, previa decisión de arquitectura |

### Interfaz con la 07

| | |
|---|---|
| **Decisión** | salida verbosa = más de 300 líneas; delegación obligatoria solo en `verify`; tabla de carga de D19 |
| **En qué se apoya** | **documentado**: `CHANGES.md:47` y `:164`; la 07 en `opsx-selective-delegation/design.md:40`, `:43` y `:58`. **Medido**: el umbral de 300 sale de D11 |
| **Afecta a** | 05, 07 |
| **Si se rechaza** | el requisito de umbral de salida verbosa, la tabla de D19 y la tarea 3.5 |

### Agentes incluidos

| | |
|---|---|
| **Decisión** | solo Claude y Codex; OpenCode y Antigravity mantienen el bloque global actual como límite declarado (Migration Plan) |
| **En qué se apoya** | **documentado** en `CHANGES.md:3` —*"plan canónico de mejoras de OPSX para Codex y Claude"*— y `CHANGES.md:158` (*Cliente Ambos*). La 07 declara lo mismo (`opsx-selective-delegation/proposal.md:61`); la 06 lo tiene como pregunta abierta nº 5 (`opsx-verifiable-evidence/design.md:114`) |
| **Afecta a** | 05, 06, 07, 08 |
| **Si se rechaza** | el alcance de la paridad por cliente y la tarea 1.2 |

### La frontera con la 06, que no es una decisión

La 05 y la 06 son **hermanas**: las dos dependen de 03 y 04, ninguna de la otra. La 06 lo escribió desde su lado, en su *Fuera de alcance*: *"Lectura dirigida y acotado de salidas (propuesta 05). Aquí solo se exige conservar la referencia al original."* Esta ficha lo escribe desde el suyo. Comparten el formato de la referencia al original y su vencimiento, los dos definidos por la 05 (D9); lo que se le propone a la 06 es adoptarlos, junto con la lista de patrones de D14.
