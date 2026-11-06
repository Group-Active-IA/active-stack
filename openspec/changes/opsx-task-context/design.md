# Design

## Context

Ver `proposal.md` — Why e Impact para la motivación y la medición de partida.

Tres restricciones del estado actual condicionan todo el diseño:

1. **El mecanismo de composición ya existe y es puro.** `internal/harness/config/compose.go` ensambla el bloque de forma determinista: dos toggles sustractivos (`delegation`, `model-routing`, que quitan secciones marcadas del archivo base) y tres aditivos (`governance`, `engram`, `tdd`, que se concatenan en orden fijo). La función no hace I/O y el mismo input produce siempre el mismo output. No hay que construir un mecanismo de carga modular: hay que darle un selector.
2. **El selector no existe.** `install.go:53` invoca `Compose(h.Toggles, …)` con los toggles literales del catálogo. No hay bandera de CLI, pantalla de TUI ni campo de configuración que los filtre, y `CLAUDE.md` del proyecto lo declara explícitamente: los toggles del orquestador no se exponen en la TUI ni en la GUI. El resultado es que toda instalación produce el bloque máximo.
3. **La decisión ocurre en el momento equivocado.** La composición se congela al instalar y se inyecta con markers en `CLAUDE.md`/`AGENTS.md`. En ese momento no se conoce la acción activa, así que no hay información con la que decidir. Mover la decisión a donde la acción sí se conoce es el núcleo de este change.

## Goals / Non-Goals

**Goals:**

- Especificar la frontera entre lo que es obligatorio en toda sesión y lo que se resuelve por acción, de forma que la frontera sea verificable y no una recomendación.
- Especificar un contrato de resumen de salida que se pueda probar con fixtures antes de que exista implementación.
- Dejar el inventario de partida medido y reproducible, para que 02 pueda comparar contra él y para que la discusión de ahorro deje de ser una estimación.
- Entregar el diseño de los fixtures y el criterio de aceptación; las corridas se planifican como tareas.

**Non-Goals:**

- No se diseña ningún subcomando nuevo del binario. *Alcance del instalador* del roadmap ya dejó los comandos de runtime nuevos sujetos a una decisión de arquitectura separada, marcada `(TBD)` **por el roadmap, no por este change**, y el informe ejecutivo advierte que un envoltorio de estado no debe crear una segunda autoridad frente a OpenSpec. El resumen de salida se especifica como **contrato de formato**: qué debe conservar, qué referencia debe llevar, qué se redacta. Quién lo produce se decide en D8, y el contrato está escrito para no depender de esa elección.
- No se reestructura el contenido de `sdd-orchestrator.md`. Ese artefacto lo tocan seis fichas; su reestructuración en módulos es de 03 y 04. Este change escribe requisitos sobre él.
- No se define ninguna métrica de consumo de modelo. Eso es 02. Los números de este change son de tamaño de texto y están etiquetados como tales.
- No se diseña un almacén de evidencia. La retención del original se decide en D9 con valores concretos, pero sobre una ruta de trabajo local; qué evidencia se considera suficiente para aceptar un cambio sigue siendo de 06.

## Decisions

### D1 — Mover la selección al momento de la acción, conservando la composición determinista

Se conserva el ensamblado determinista actual y se le agrega una entrada: la acción activa. La composición sigue siendo pura y el orden sigue siendo fijo, de modo que la resolución es reproducible y comparable entre corridas.

*Alternativa considerada: exponer los cinco toggles como banderas de instalación.* Es más barata y el plan previo del proyecto la propone. Se descarta como objetivo de este change porque resuelve otro problema: permite a un equipo medir si un componente vale su costo, pero sigue congelando la elección al instalar. Un equipo que active TDD porque lo usa en apply lo carga igual en explore. No obstante, es un buen paso intermedio y queda como tarea de bajo riesgo, porque habilita la comparación con y sin cada módulo que 02 necesita.

*Alternativa considerada: dejar el bloque completo y confiar en la caché del proveedor.* Se descarta porque el informe ejecutivo ya matizó esa afirmación: el proveedor reutiliza bajo condiciones y el ensamblado lo hace el host. Depender de eso convierte una propiedad de diseño en una esperanza sobre un tercero. La estabilidad de prefijos y caché es materia de 09, como variable independiente.

### D2 — El núcleo obligatorio se define por consecuencia, no por tamaño

La frontera entre núcleo y módulo no se fija con un presupuesto de bytes ni de líneas. Una instrucción pertenece al núcleo si su ausencia puede producir daño o invalidar el trabajo con independencia de la acción; pertenece a un módulo si su pertinencia depende de la acción activa.

*Por qué no un presupuesto numérico:* el informe ejecutivo marcó los umbrales de 40 líneas y 1.500 tokens como presupuestos de diseño, no como límites validados. Adoptar un número como criterio reintroduce exactamente la afirmación que la revisión rechazó. Los números siguen siendo útiles como hipótesis a medir con 02, y así quedan registrados en `proposal.md`.

Aplicando el criterio al inventario medido, el reparto candidato es:

| Fragmento | Candidato | Razón |
|---|---|---|
| reglas de cumplimiento obligatorio | núcleo | su ausencia invalida el trabajo en cualquier acción |
| `governance` (1.276 B) | núcleo | define qué autonomía tiene el agente sobre dominios sensibles; no depende de la acción |
| `delegation` (1.005 / 851 B) | módulo | solo aplica cuando la acción coordina subagentes |
| `model-routing` (778 / 450 B) | módulo | solo aplica cuando hay delegación |
| `engram` (6.093 B) | módulo | su protocolo de guardado aplica en puntos de decisión y de cierre |
| `tdd` (3.797 B) | módulo | solo aplica en apply con runner de tests disponible |

**Este reparto queda decidido**, no propuesto. El criterio es el de arriba y la aplicación es la de la tabla: `governance` al núcleo porque define cuánta autonomía tiene el agente sobre dominios sensibles y eso no depende de la acción; los otros cuatro a módulo porque su pertinencia sí depende.

La matriz de módulos de 03 y 04 es un **punto de verificación declarado**, no una condición para decidir: si su inventario contradice este reparto, la tarea 3.1 registra la divergencia y se ajusta. Lo que no se hace es dejar el reparto abierto esperándola.

### D3 — El resumen de salida se especifica por lo que debe conservar, no por cuánto debe recortar

Un límite de tamaño es el modo de fallo que la ficha nombra: una salida extensa puede ocultar un fallo. Por eso el contrato se escribe en términos de retención obligatoria (errores, fallos, código de salida, redacción de datos sensibles) y de recuperabilidad del original, y no en términos de un tope de líneas. Un resumen que cumple la retención y pierde volumen es correcto; uno que cumple un tope y pierde un fallo es incorrecto.

*Alternativa considerada: recortar por cabeza y cola.* Se descarta porque falla justamente en el caso del fallo central, que es uno de los tres que la ficha pide probar.

### D4 — Los fixtures son el entregable verificable de esta ronda

Como el change no implementa, el artefacto que demuestra que la especificación es verificable son los fixtures. Se diseñan con el fallo en tres posiciones (inicio, medio, final), con varios fallos dispersos, y con una salida verbosa sin fallos como control negativo. Cada fixture declara su salida original, su resumen esperado y el fallo que debe sobrevivir.

### D5 — Procedencia obligatoria en lugar de umbral de lectura como regla principal

La regla que más evita trabajo inválido no es "leé menos" sino "toda afirmación sobre el código indica ruta y línea o símbolo". El umbral de lectura completa es un apoyo; la procedencia es lo que hace falsable cada hallazgo. Se especifican las dos, con la procedencia como requisito normativo y el umbral como condición declarada y ajustable.

### D6 — La llamada innecesaria se registra, no se bloquea

Una relectura o una invocación repetida se marca como llamada innecesaria y queda registrada; el sistema no la impide. Bloquearla exigiría decidir con certeza que el estado observado no cambió, y hay casos legítimos donde sí cambió por fuera de la sesión. Registrar convierte el problema en una medida —cuántas llamadas innecesarias por tarea aceptada— que 02 puede comparar entre configuraciones, que es el uso que tiene valor.

*Alternativa considerada: bloquear la segunda llamada.* Se descarta porque un falso positivo deja al agente sin un dato que sí cambió, y el daño de eso es mayor que el de una llamada repetida. La excepción es el sondeo de una integración ya informada como no disponible: ahí el estado sí es conocido dentro de la tarea, y el requisito es no volver a invocarla.

### D7 — El inventario de verbosidad cubre todas las fuentes, no solo el orquestador

La ficha pide localizar dónde aparecen instrucciones largas o resultados verbosos. El orquestador es la fuente más grande y ya está medida, pero no es la única: los cuerpos de skill y los fragmentos de sección propia que el instalador escribe en el proyecto destino también ocupan contexto en toda sesión. El inventario se completa sobre todas las fuentes y se publica como una tabla única ordenada por tamaño, porque el criterio para elegir qué se resuelve por tarea debe apoyarse en el total real y no en el fragmento que resultó más fácil de medir.

Este inventario es una actividad de investigación, no un comportamiento del sistema: por eso vive en `tasks.md` (grupo 1) y en `proposal.md` — Impact, y no como requisito de una spec.

### D8 — El resumen lo produce el agente bajo instrucción, y el contrato no depende del productor

De las tres opciones, el envoltorio ejecutable está vedado por *Alcance del instalador* y la capacidad nativa del cliente no existe por igual en Claude y Codex, así que elegirla rompería la paridad que esta ficha necesita por ser de ambos clientes. Queda el agente bajo instrucción, que está disponible en los dos y no agrega autoridad nueva.

La objeción válida es que una instrucción no es un control: el agente puede no retener el fallo. Se resuelve así: **el contrato y los fixtures verifican el resumen, no al productor.** Un resumen se acepta o se rechaza por lo que conserva, con independencia de quién lo escribió. Y el resumen declara quién lo produjo, de modo que un incumplimiento es atribuible. La consecuencia práctica es que sustituir mañana la instrucción por un envoltorio ejecutable no cambia una sola spec ni un solo fixture: cambia el productor y los mismos fixtures lo evalúan.

*Alternativa descartada: esperar a que exista el envoltorio.* Dejaría la ficha sin entregable verificable en esta ronda, que es justamente lo que D4 evita.

### D9 — El original se redacta al capturarlo, no solo al resumirlo; retención ligada al change

Dos decisiones unidas, porque la segunda depende de la primera.

**Redacción al capturar.** Los valores de las clases de dato sensible declaradas se redactan en el momento de capturar la salida, no solo al resumirla. No existe entonces ninguna copia con secretos esperando a que venza una retención. Nadie necesita el valor de un secreto para diagnosticar un fallo; necesita saber que había uno y dónde.

**Retención.** El original se conserva local, en una ruta de trabajo no versionada, mientras el change esté activo y como mínimo siete días. **Ese número es arbitrario**: lo elegí para que el valor exista y se pueda instrumentar, y no lo respalda ninguna medición —a diferencia del umbral de lectura de D11, que sale de medir el repositorio—. Se calibra con el primer uso real: si nadie pide un original después de tres días, sobra; si alguien lo pide a los diez, falta. Vencido el plazo, la referencia informa que el original no está disponible y desde cuándo, que es lo que la spec ya exige. El acceso es el del sistema de archivos de quien corre el agente: el original no se publica ni se envía a ningún lado.

*Alternativa descartada: conservar el original sin redactar y controlar el acceso.* Un control de acceso sobre un archivo local es una ficción; redactar en origen elimina el problema en lugar de administrarlo.

### D10 — Las clases de tarea son las cinco acciones del ciclo OPSX

El eje de clasificación no se inventa: son `explore`, `propose`, `apply`, `verify` y `archive`. Es el eje que el proyecto ya declara, que OpenSpec ya conoce y que el orquestador ya usa para ubicarse, de modo que declarar un eje propio habría creado un segundo vocabulario para lo mismo.

Las cinco tareas de medición de la ficha 02 (exploración, cambio pequeño, cambio multiarchivo, diagnóstico, revisión) no son clases de tarea: son **cargas de trabajo** que se ejecutan dentro de esas acciones y sirven para comparar. Se mapean sobre las acciones, no las reemplazan.

Esto desacopla a esta ficha de 03 y 04 en lo que respecta al **eje**: el eje ya existe. Lo que sigue necesitando acuerdo con ellas es el **mapeo de módulo a acción**, que es otra cosa y está en la tarea 3.1.

*Alternativa descartada: un eje propio por tipo de trabajo (lectura, escritura, diagnóstico).* Más expresivo, pero obliga a traducir entre dos vocabularios en cada ficha que toque el orquestador.

**La objeción que esta decisión no resuelve bien.** El eje OPSX no distingue trabajos distintos dentro de la misma acción: analizar un log y escribir documentación caen las dos en `apply` y necesitan herramientas distintas. Con este eje, las dos reciben lo mismo. Por eso la decisión va al grupo con la objeción adentro y no solo con su fundamento, y queda una tercera salida sobre la mesa: OPSX como eje principal con un sub-eje de tipo de trabajo dentro de cada acción.

### D11 — El umbral de lectura completa es 300 líneas, medido sobre este repositorio

**En líneas**, porque es la unidad en la que el agente pide un rango y en la que se expresa la procedencia que D5 hace obligatoria. Los bytes no se corresponden con nada que el agente pueda pedir; los símbolos no existen en archivos sin parser disponible.

**300**, porque sale de medir el repositorio y no de estimar. Distribución de los 130 archivos `.go` sin tests:

| | líneas |
|---|---|
| mediana (p50) | 87 |
| p75 | 178 |
| p90 | 344 |
| p95 | 400 |
| máximo | 809 (`internal/tui/model.go`) |

| Umbral | Archivos que se leen enteros |
|---|---|
| 200 | 104 de 130 (80 %) |
| **300** | **116 de 130 (89 %)** |
| 400 | 124 de 130 (95 %) |
| 500 | 126 de 130 (97 %) |

Descarté 400 —que había puesto primero— justamente por medirlo: a 400 la regla toca solo 6 archivos y no cambia el comportamiento en la práctica. Y descarté 200 porque fuerza lectura acotada en archivos donde leer entero sale más barato que dos viajes de ida y vuelta: un archivo de 200 líneas son unos 2.500 tokens, y dos consultas acotadas con riesgo de una tercera no ahorran nada.

300 es la rodilla: deja pasar el 89 % de los archivos, que son chicos y baratos, y reserva la lectura acotada para los 14 que son genuinamente grandes, donde sí paga. La regla tiene que ser **rara y dirigida**, no una fricción constante.

El valor vive en el catálogo y se ajusta sin tocar ninguna spec. Queda etiquetado como calibrable con la medición de 02, pero ya no es un número a dedo.

### D12 — No se adopta un tope de tamaño para los cuerpos de skill en este change

El inventario de las tareas 1.5 a 1.7 publica el tamaño de cada cuerpo de skill, que es el entregable. Fijar un tope es política del catálogo y afecta a toda ficha que instale skills, no solo a esta; adoptarlo acá sería legislar para las otras nueve desde la ficha 05.

La decisión es entonces: **esta ficha entrega la medición y no el tope.** Si el inventario muestra cuerpos desproporcionados, el tope se propone como cambio de catálogo con esos números sobre la mesa. Discutirlo antes de tener el inventario sería discutir un número a ciegas.

### D13 — La señal de pertinencia son el tipo de coincidencia y su cantidad

La orientación a nivel repositorio devuelve rutas candidatas, y cada una viene con dos datos: **qué tipo de coincidencia** tiene y **cuántas**. El tipo es uno de tres, ordenados por peso:

| Tipo | Qué es | Peso |
|---|---|---|
| definición | el símbolo se define en ese archivo | alto |
| referencia | el símbolo se usa en ese archivo | medio |
| texto | coincidencia literal sin estructura | bajo |

Una definición pesa más que cualquier cantidad de referencias, y una referencia más que cualquier cantidad de coincidencias de texto. Dentro del mismo tipo, ordena la cantidad.

*Por qué estos tres y no una puntuación:* una puntuación numérica obliga a justificar los pesos y no es inspeccionable. Tres categorías ordenadas se explican en una línea y el agente puede decidir con ellas sin que nadie calibre nada. Y los tres son obtenibles sin leer el contenido de los archivos descartados, que es el requisito que esta señal tiene que respetar.

### D14 — Una sola lista de datos sensibles: la del piso de permisos

Las clases de dato sensible que la redacción usa **son las mismas que declara el piso de permisos del catálogo**, no una lista propia de esta ficha. Hoy esa lista está en `internal/harness/config/permissions/overlays.go:16` y contiene solo `.env` y `.env.*`.

*Por qué no una lista propia:* dos listas divergentes son peor que una sola flaca. Si la redacción protege algo que el piso no bloquea, o al revés, el usuario no puede saber qué está protegido. Con una sola fuente, cuando la ficha 01 amplíe el piso —el diagnóstico ya señaló que faltan claves SSH, credenciales de nube y `*.pem`— la redacción se amplía con él, sin tocar esta ficha.

*Consecuencia asumida:* mientras el piso sea solo `.env`, la redacción protege solo `.env`. Eso queda declarado, no disimulado, y es un aviso al grupo (sección D del registro), no un pendiente de este change.

### D15 — La lista de condiciones para leer un archivo completo es cerrada

Son cuatro, y lo que no está acá no justifica lectura completa:

1. el archivo está por debajo del umbral de D11
2. la tarea va a reescribir el archivo entero
3. el archivo es el artefacto que la acción activa produce o edita
4. una lectura acotada previa sobre ese archivo devolvió vacío

*Por qué cerrada:* una lista abierta vuelve inaplicable el registro de lectura no justificada — siempre se puede alegar una condición nueva. Cerrada, el registro es decidible: o la lectura entra en una de las cuatro o queda marcada.

La cuarta existe para no dejar al agente trabado: si pidió un símbolo y no estaba, leer entero es la salida razonable y no debería contar como falta.

### D16 — La respuesta del agente es una capacidad aparte, no parte del resumen de comandos

La ficha pide localizar *"instrucciones largas o resultados verbosos"*, y hay dos clases de resultado verboso: la salida de un comando y la respuesta del propio agente. Las dos entierran el hallazgo, pero **no se arreglan con el mismo contrato**, y por eso van en capacidades separadas.

| | Salida de comando | Respuesta del agente |
|---|---|---|
| quién la produce | una herramienta | el modelo |
| hay un original completo | **sí**, y se conserva | **no**, nunca existió |
| cómo se accede al detalle | referencia al original guardado | se produce a pedido |
| qué se puede perder | una línea de fallo | el hallazgo, enterrado en el relato |

La diferencia que decide es la segunda: **no se puede referenciar el original de algo que nunca se produjo.** Para una salida de comando la regla es conservar y referenciar; para una respuesta es poner el hallazgo primero y generar el desarrollo recién cuando se lo piden. Meter las dos en `bounded-command-output` obligaría a que la mitad de sus requisitos no apliquen a la mitad de sus casos.

*Alternativa descartada: ampliar `bounded-command-output` y renombrarla.* Tentador por lo parecido del síntoma, pero el nombre dejaría de describir lo que la capacidad hace y el requisito de original recuperable quedaría sin sentido para la mitad de los casos.

Una de las reglas de esta capacidad sale de una corrección de esta misma sesión: una respuesta no reproduce lo que ya está escrito en un artefacto, lo señala. Vale para lo que el agente le dice a una persona y para lo que un artefacto le copia a otro.

## Risks / Trade-offs

- **El inventario definitivo depende de 03 y 04 →** se entrega el inventario del estado actual, medido y reproducible, y se declara en `tasks.md` el punto exacto donde se revisa contra la matriz de módulos de 03 y 04. El criterio de reparto (D2) no depende de ellas y no necesita revisarse.
- **Seis fichas tocan `sdd-orchestrator.md` y ninguna lo posee →** este change escribe requisitos, no contenido, y lleva al grupo la propuesta de que 03 y 04 sean dueñas del archivo. Si se escribe contenido antes de ese acuerdo, hay reescritura garantizada.
- **Resolver módulos por acción agrega un punto de fallo que hoy no existe →** mitigado por el requisito de que un módulo obligatorio ausente detenga la acción e informe, en lugar de continuar con contexto incompleto. Es el modo de fallo que la aceptación transversal del roadmap ya exige.
- **Mover instrucciones fuera del bloque siempre presente puede degradar el comportamiento del agente en acciones que sí las necesitaban →** por eso ninguna adopción se decide en este change. La comparación con y sin cada módulo, sobre la misma aceptación, es el requisito de `task-tool-selection` y usa el protocolo de 02.
- **El contrato de resumen no sirve de nada si nadie lo produce →** riesgo real y asumido. Esta ronda entrega el contrato y los fixtures; quién lo produce es la pregunta abierta Q1 y la decide el grupo, no este documento.
- **Riesgo de solapamiento con 07 →** la ficha 07 consume el resultado de esta pero no la lista en sus decisiones de grupo. Mitigación: este change define el contrato de carga de módulos de forma que 07 lo consuma sin redefinirlo, y la asimetría queda registrada para resolverse en grupo.

## Migration Plan

No aplica en esta ronda: el change no modifica configuración instalada. Cuando se implemente, la migración del bloque administrado en `CLAUDE.md`/`AGENTS.md` pertenece a 03 y 04, que ya la tienen en su alcance con snapshot, preservación de contenido del usuario y rollback. Este change no debe introducir una segunda ruta de migración sobre los mismos archivos.

## UI Mockups

Decision: no (no UI) — 2026-11-05. El change no agrega ni modifica pantalla, componente, layout ni interacción.

## Decisiones que el grupo tiene que cerrar

Seis puntos. Cada uno con sus opciones, la recomendación, **de dónde sale** esa recomendación y a qué fichas afecta. Se identifican por su nombre, no por un código: así nadie tiene que buscar una leyenda y no se confunden con los puntos abiertos de otra ficha.

Una recomendación puede apoyarse en tres cosas, y cada una lleva su procedencia, igual que D5 exige para cualquier afirmación sobre el código:

| | Qué significa |
|---|---|
| **documentado en** | está escrito en el repositorio o en el roadmap, con archivo y línea |
| **medido acá** | lo midió este change y el número es reproducible |
| **supuesto** | no hay respaldo todavía; se dice por qué y qué lo resolvería |

### Inventario y ubicación de los módulos

| | |
|---|---|
| **Opciones** | archivos dentro de la skill · skills separadas · entradas de catálogo |
| **Recomendación** | archivos dentro de la skill, alineado con lo que propone la 04 |
| **De dónde sale** | **supuesto.** La rama `feat/codex-native-opsx` tiene el inventario abierto en su propia tabla de decisiones de grupo, también sin respaldo. Lo resolvería el inventario de módulos de 03 y 04 |
| **Afecta a** | 03, 04, 05, 07 |
| **Si se rechaza** | cambia la tarea 3.2 y dónde se declaran los módulos; ningún requisito |

**No es una decisión de esta ficha.** El inventario lo produce la 04; acá se consume. Se lista para no apoyar la propuesta en un supuesto que ni su dueño confirmó.

### Eje de clases de tarea

| | |
|---|---|
| **Opciones** | las cinco acciones OPSX · un eje propio por tipo de trabajo (lectura, escritura, diagnóstico) |
| **Recomendación** | las acciones OPSX, ver D10 |
| **De dónde sale** | **documentado en** `CLAUDE.md:26`, `ARCHITECTURE.md:6` y `ARCHITECTURE.md:164`: el ciclo OPSX es `explore → propose → apply → verify → archive`. La rama `feat/codex-native-opsx` propone el mismo eje como alternativa si no hay acuerdo con la 03 |
| **Objeción al eje recomendado** | el eje OPSX **no distingue trabajos distintos dentro de la misma acción**: analizar un log y escribir documentación pueden pasar las dos dentro de `apply` y necesitan herramientas distintas. Un eje por tipo de trabajo (lectura, escritura, diagnóstico, documentación) sí los distingue, al costo de traducir entre dos vocabularios en cada ficha que toque el orquestador. Una tercera salida: OPSX como eje principal más un sub-eje de tipo de trabajo dentro de cada acción — más fiel y más caro |
| **Afecta a** | 03, 04, 05, 07 |
| **Si se rechaza** | el vocabulario del catálogo en las tareas 3.2 y 5.1; los requisitos dicen "la acción activa" y siguen válidos |

### Reparto núcleo / módulo

| | |
|---|---|
| **Opciones** | `governance` al núcleo y los otros cuatro a módulo · `engram` también al núcleo · todo a módulo |
| **Recomendación** | la primera, ver D2 |
| **De dónde sale** | **medido acá.** `proposal.md` — Impact, tabla de atribución por fragmento, obtenida llamando a `Compose` (`internal/harness/config/compose.go`): `engram` son 6.093 B, el 30 % del bloque de Claude. Si entra al núcleo, la ficha pierde la mitad de su efecto. Reproducible por la tarea 1.1 |
| **Afecta a** | 03, 04, 05 |
| **Si se rechaza** | la tabla de D2 y la tarea 3.1; el criterio de reparto no cambia |

### Política de salidas originales

Qué se versiona, cuánto se retiene, quién accede y cómo se redacta.

El commit `1b4639b` no es la fuente de esta recomendación: es lo que volvió urgente la pregunta. Antes `openspec/` estaba en el `.gitignore` y una salida guardada en la carpeta del change no podía llegar al repositorio ni por error; ahora sí. Eso cambia el riesgo, no la decisión.

| | |
|---|---|
| **Opciones** | nada versionado y redacción al capturar · recibo versionado y salida afuera · versionar los dos, redactados |
| **Recomendación** | la primera, ver D9 |
| **De dónde sale** | Tres piezas distintas. **Documentado en** `CHANGES.md:166`: la ficha 05 pone *"dónde queda el original"* entre lo que se acuerda en grupo, así que el punto es nuestro. **Alineado con** la rama `opsx-verifiable-evidence`, cuya spec ya exige que *"full original outputs MUST NOT be committed to the repository by default"*: la recomendación de no versionar no inventa nada, coincide con la 06. **Supuesto**: el mínimo de siete días y el "mientras el change esté activo" los elegí yo, sin nada que los respalde |
| **Afecta a** | 05, 06 |
| **Si se rechaza** | dos valores de la spec de retención y las tareas 2.10 y 2.11 |

### Productor del resumen acotado

| | |
|---|---|
| **Opciones** | el agente bajo instrucción · un envoltorio ejecutable en el binario · una capacidad nativa del cliente |
| **Recomendación** | el agente bajo instrucción, ver D8. La segunda está vedada por *Alcance del instalador*; la tercera no existe por igual en los dos clientes |
| **De dónde sale** | **supuesto.** La objeción —una instrucción no es un control— viene del informe ejecutivo, no de una medición nuestra: nadie probó si el agente retiene el fallo. Por eso el contrato evalúa el resumen y no al productor, y cambiarlo después no toca ninguna spec. Lo resolvería correr los fixtures del grupo 2 contra un agente real |
| **Afecta a** | 05, 06 |
| **Si se rechaza** | el productor y nada más |

### Agentes incluidos

| | |
|---|---|
| **Opciones** | solo Claude y Codex · sumar OpenCode · sumar Antigravity |
| **Recomendación** | solo Claude y Codex, igual que la 06 y la 08 |
| **De dónde sale** | **documentado en** `CHANGES.md:3` —*"plan canónico de mejoras de OPSX para Codex y Claude"*— y `CHANGES.md:158`, que declara esta ficha como *Cliente Ambos* |
| **Afecta a** | 05, 06, 08 |
| **Si se rechaza** | el alcance de la paridad por cliente y la tarea 1.2 |

**No es una decisión de esta ficha**, es del roadmap. Y queda un hecho incómodo sin dueño: el instalador sigue escribiendo para OpenCode y Antigravity, y ninguna ficha dice qué les pasa cuando se retire el bloque global.

### La frontera con la 06, que no es una decisión

La 05 y la 06 son **hermanas**: las dos dependen de 03 y 04, ninguna de la otra. La 06 lo escribió desde su lado, en su *Fuera de alcance*: *"Lectura dirigida y acotado de salidas (propuesta 05). Aquí solo se exige conservar la referencia al original."* Esta ficha lo escribe desde el suyo. Comparten **dos datos**, los dos definidos por la 05: el formato de la referencia al original, y que vence. De eso, lo único que necesita acuerdo es qué se versiona — el punto *política de salidas originales*.
