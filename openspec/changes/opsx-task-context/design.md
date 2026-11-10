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

El único número del contrato es **cuándo** se resume, no **cuánto**: una salida de más de 300 líneas **o** de más de 10.000 caracteres, lo que ocurra primero, se entrega como resumen más referencia (D19). Es una condición de entrada, no un tope de recorte.

**Por qué dos unidades** (**decidido por la 05 — pendiente de revisión del grupo**). Los clientes no miden en líneas. Claude Code entrega inline un resultado válido hasta unos 30.000 caracteres y uno fallido hasta unos 10.000; pasado eso, el válido llega como ruta de archivo más una vista previa de los primeros 2.000 caracteres y el fallido como un extracto de cabeza y cola, *"with no file path"* (https://code.claude.com/docs/en/tools-reference, *Output limits*). Codex recorta el medio de cada salida dentro de un presupuesto por salida, `tool_output_token_limit` (https://learn.chatgpt.com/docs/config-file/config-reference), cuyo valor por defecto la documentación no da. Con 300 líneas de unos 100 caracteres, un comando fallido ya se recorta en Claude alrededor de la línea 100: el umbral en líneas solo, que es el de la lectura de archivos (D11), llegaba tarde justo en el caso que importa. Los 10.000 caracteres son el límite documentado más bajo de los dos clientes, y el umbral en líneas queda para que la regla siga siendo legible en la unidad del rango. El umbral de lectura de archivos (D11) y el de salidas de comando son **dos reglas distintas**, aunque compartan las 300 líneas.

*Alternativa descartada: recortar por cabeza y cola.* Falla justamente en el caso del fallo central, que es uno de los tres que la ficha pide probar. Es además lo que el cliente ya hace por su cuenta con una salida fallida: por eso el resumen se arma desde el original capturado (D8) y no desde lo que el cliente mostró.

### D4 — Los fixtures son el entregable verificable de esta ronda

Como el change no implementa, el artefacto que demuestra que la especificación es verificable son los fixtures. Se diseñan con el fallo en tres posiciones (inicio, medio, final), con varios fallos dispersos, con una salida verbosa sin fallos como control negativo, con un fallo que aparece **solo por stderr**, y con un comando cuyo original **nunca se capturó**. Cada fixture declara su salida original, su resumen esperado y el fallo que debe sobrevivir.

El fixture de stderr existe porque los clientes no tratan igual ese flujo: en Claude la salida llega *"with stderr interleaved into stdout"* (https://code.claude.com/docs/en/monitoring-usage); en Codex la documentación no lo dice. La captura de D8 redirige los dos flujos al mismo archivo, así que el fallo de stderr tiene que sobrevivir en ambos clientes; el fixture se corre en los dos. Los fixtures se corren **contra el cliente real** (D17): el recorte propio del cliente forma parte de lo que miden.

### D5 — Procedencia obligatoria en lugar de umbral de lectura como regla principal

La regla que más evita trabajo inválido no es "leé menos" sino "toda afirmación sobre el código indica ruta y línea o símbolo". El umbral de lectura completa es un apoyo; la procedencia es lo que hace falsable cada hallazgo. Se especifican las dos, con la procedencia como requisito normativo y el umbral como condición declarada en el catálogo.

### D6 — La llamada innecesaria se registra, no se bloquea

Una relectura o una invocación repetida se marca como llamada innecesaria y queda registrada; nada la impide. Bloquearla exigiría decidir con certeza que el estado observado no cambió, y hay casos legítimos donde sí cambió por fuera de la sesión. Registrar convierte el problema en una medida —cuántas llamadas innecesarias por tarea aceptada— que 02 puede comparar entre configuraciones, que es el uso que tiene valor. Qué significa "registrar" lo fija D17.

*Alternativa descartada: bloquear la segunda llamada.* Un falso positivo deja al agente sin un dato que sí cambió, y el daño de eso es mayor que el de una llamada repetida. La excepción es el sondeo de una integración ya informada como no disponible: ahí el estado sí es conocido dentro de la tarea, y el requisito es no volver a invocarla.

### D7 — El inventario de verbosidad cubre todas las fuentes, no solo el orquestador

La ficha pide localizar dónde aparecen instrucciones largas o resultados verbosos (`CHANGES.md:162`). El orquestador es la fuente más grande y ya está medida, pero no es la única: los cuerpos de skill y los fragmentos de sección propia que el instalador escribe en el proyecto destino también ocupan contexto en toda sesión, y las salidas verbosas de comandos (tests, build, logs) aparecen en puntos concretos del ciclo. El inventario se completa sobre todas las fuentes —instrucciones (tareas 1.1 a 1.7) y salidas (tarea 1.8)— y se publica como una tabla única ordenada por tamaño, porque el criterio para elegir qué se resuelve por tarea debe apoyarse en el total real y no en el fragmento que resultó más fácil de medir.

Este inventario es una actividad de investigación, no un comportamiento del sistema: por eso vive en `tasks.md` (grupo 1) y en `proposal.md` — Impact, y no como requisito de una spec. Se propone a la 02 reutilizarlo para su inventario de instrucciones (`opsx-baseline/design.md:46`) en lugar de duplicarlo.

### D8 — El resumen lo produce el agente bajo instrucción, y el contrato no depende del productor

*Límite declarado:* el manejo de salidas grandes de un hook de Codex (https://learn.chatgpt.com/docs/hooks) no prueba que toda salida de shell se conserve. Por eso la captura de esta decisión no depende de hooks: la hace el agente redirigiendo el comando a un archivo.

De las tres opciones, el envoltorio ejecutable sería un runtime nuevo, que `CHANGES.md:49` deja sujeto a una decisión de arquitectura separada. La capacidad nativa del cliente existe en los dos, pero con otra forma y otro fin: Claude acota con `BASH_MAX_OUTPUT_LENGTH` (https://code.claude.com/docs/en/env-vars) y guarda el resto en archivo o entrega cabeza y cola; Codex acota con `tool_output_token_limit` y recorta el medio (https://learn.chatgpt.com/docs/config-file/config-reference). Ninguna de las dos resume conservando fallos, que es lo que pide el contrato, así que elegirla tampoco resuelve el problema. Queda el agente bajo instrucción, que está disponible en los dos y no agrega autoridad nueva.

**El agente captura antes de leer** (**decidido por la 05 — pendiente de revisión del grupo**). La versión anterior de esta decisión suponía que el agente ve la salida completa y la resume. No la ve: el cliente la recorta antes (D3), y el caso que más recorta —un comando que falla, con el error al medio— es el que el contrato quiere proteger. Por eso el módulo de resumen acotado prescribe un paso previo, sin subcomando nuevo (`CHANGES.md:49`):

1. Un comando largo o verboso —los que el inventario de la tarea 1.8 declara así para cada acción, y cualquier otro cuya salida se espere por encima del umbral de D3— se ejecuta **redirigiendo stdout y stderr al mismo archivo** dentro de la ruta de corrida de D9, y **capturando el código de salida** en la misma invocación.
2. El agente redacta, calcula la huella y resume **desde ese archivo**, leyéndolo con lectura dirigida (rango, búsqueda de patrones de fallo, cola), y no desde lo que el cliente mostró inline.
3. Si el original no se pudo capturar —el comando ya corrió sin redirección, el archivo falta o quedó vacío sin explicación—, el resumen se marca **"parcial — sin verificar"** y dice por qué. Si el comando no tiene efectos laterales (una suite de tests, un `go vet`), el agente puede volver a correrlo con captura en lugar de entregar el parcial; si los tiene, no lo repite.

La receta es por shell, porque el código de salida se pierde de forma distinta en cada una cuando se encadena con una tubería:

| Shell | Receta | Cuidado |
|---|---|---|
| bash / sh | `cmd > "$f" 2>&1; echo "exit=$?" >> "$f.exit"` | no usar `cmd 2>&1 \| tee "$f"`: sin `set -o pipefail`, `$?` es el de `tee` |
| PowerShell | `& cmd *> $f; "exit=$LASTEXITCODE" \| Out-File "$f.exit"` | `$LASTEXITCODE` vale para comandos nativos; `$?` solo dice si falló el último paso |

La copia que el cliente guarda por su cuenta (en Claude, el archivo de trabajo y el del directorio de sesión, https://code.claude.com/docs/en/tools-reference) **no** es el original de Active Stack: no está redactada ni lleva la huella de D9. El resumen nunca la cita como referencia.

La objeción válida es que una instrucción no es un control: el agente puede no retener el fallo. Se resuelve así: **el contrato y los fixtures verifican el resumen, no al productor.** Un resumen se acepta o se rechaza por lo que conserva, con independencia de quién lo escribió. Y el resumen declara quién lo produjo, de modo que un incumplimiento es atribuible. Sustituir mañana la instrucción por un envoltorio ejecutable no cambia una sola spec ni un solo fixture: cambia el productor y los mismos fixtures lo evalúan.

**Supuesto declarado:** que el agente bajo instrucción captura antes de leer y retiene los fallos de los fixtures. Nadie lo midió todavía. Se valida corriendo los fixtures del grupo 2 contra un agente real, en el cliente real, con el arnés fuera del binario (tareas 2.7 y 2.17); si los fixtures del fallo central, de stderr o de ruido fallan, el resultado se lleva al grupo como insumo para la decisión de arquitectura que `CHANGES.md:49` reserva.

*Alternativa descartada: esperar a que exista el envoltorio.* Dejaría la ficha sin entregable verificable en esta ronda, que es justamente lo que D4 evita.

### D9 — El original se redacta antes de guardarse, vive fuera del repo y vence después de archivar

Tres decisiones unidas (**decidido por la 05 — pendiente de revisión del grupo**).

**Redacción antes de conservar.** Los valores que coincidan con la lista de patrones sensibles (D14) se redactan antes de conservar el original, no solo al resumirlo. Como la captura de D8 es una redirección del shell, el primer archivo que se escribe es **crudo**: el agente lo redacta a un segundo archivo, que es el original que se conserva, y borra el crudo antes de entregar el resumen. Ninguna copia que guarde Active Stack queda con secretos esperando a que venza una retención. Nadie necesita el valor de un secreto para diagnosticar un fallo; necesita saber que había uno y dónde.

**Límite declarado: el cliente guarda sus propias copias.** Claude Code *"streams a command's output to a working file as the command runs"* y deja las salidas grandes en *"a file saved to the session directory"* (https://code.claude.com/docs/en/tools-reference); Codex guarda la salida, ya recortada, *"in history"* (`tool_output_token_limit`, https://learn.chatgpt.com/docs/config-file/config-reference), y dónde persiste no está documentado. Esas copias no pasan por nuestra redacción. La garantía se acota entonces a **la copia que guarda Active Stack**, y la protección real se apoya en no leer el secreto en origen: las reglas de denegación de lectura del piso de Claude (`Read(.env)`, `internal/harness/config/permissions/overlays.go:16-28`) y su equivalente nativo de Codex, que hoy el instalador no escribe (D14). Cuando hay dos copias, **gobierna la de Active Stack**: es la única con huella y vencimiento.

**Referencia y ubicación.** Todo resumen lleva una referencia con cinco campos: `{comando, id de corrida, ruta, huella sha256 del original ya redactado, fecha de vencimiento}`. Una sola huella alcanza porque el crudo no se conserva: no hay un original sin redactar que haya que distinguir del redactado. El original vive en la caché del usuario, fuera del repositorio: `<os.UserCacheDir>/active-stack/runs/<change>/`. Desde `1b4639b` `openspec/` se versiona, así que una salida guardada dentro del change podría commitearse por error; fuera del repo no puede.

**Retención y acceso.** El original se conserva hasta que el change se archive **y siete días más**; no hay otro plazo, así que no hay prelación que resolver. Los siete días son un **supuesto**: no los respalda ninguna medición —a diferencia del umbral de D11—. Se recalibran con el primer uso real: si nadie pide un original más de tres días después de archivar, se baja; si alguien lo pide a los diez, se sube. Vencido el plazo, la referencia informa que el original no está disponible y desde qué fecha. Un original **vencido** y uno **nunca capturado** son casos distintos: el primero existió y tiene referencia completa; el segundo no tiene ruta ni huella, y su resumen es "parcial — sin verificar" (D8). El acceso es el del usuario local con los permisos de sistema de archivos de su directorio de caché: el original no se publica ni se envía a ningún lado.

**Diferencias con la 06**, que es hermana y define la evidencia. Coincidimos en no versionar las salidas completas por defecto (`opsx-verifiable-evidence/design.md:78`). Diferimos en cuatro puntos, y se le propone a la 06 adoptar los de acá:

| | 06 (`opsx-verifiable-evidence/design.md`) | 05 |
|---|---|---|
| cuándo se redacta | antes de compartir la salida (`:79`) | antes de conservarla; el crudo de la captura se borra |
| huellas | dos: la original y la de la versión redactada (`:79`) | una: la del original ya redactado |
| dónde vive | subdirectorio local de la evidencia, bajo `openspec/changes/<change>/evidence/` (`:73`, `:78`) | `<os.UserCacheDir>/active-stack/runs/<change>/`, fuera del repo |
| retención | `(TBD)`, *"por ejemplo, hasta archivar el change"* (`:80`) | hasta archivar y siete días más, compatible con su ejemplo |
| formato de referencia | ruta, huella y si fue truncada (`:46`) | comando, id de corrida, ruta, huella sha256, vencimiento |

*Alternativa descartada: conservar el original sin redactar y controlar el acceso.* Un control de acceso sobre un archivo local es una ficción; redactar antes de conservar elimina el problema en la copia propia en lugar de administrarlo.

*Alternativa descartada: redactar en la tubería, antes de que toque el disco.* Exigiría un filtro ejecutable entre el comando y el archivo, que es el envoltorio que `CHANGES.md:49` reserva. El crudo transitorio es el costo de no tenerlo, y queda declarado.

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

**Cabe en lo que los clientes leen.** En Claude Code una lectura devuelve como máximo 25.000 tokens y rechaza un archivo entero de más de 256 KB; Read acepta `offset`/`limit` y Grep `head_limit`/`offset` (https://code.claude.com/docs/en/tools-reference). Un archivo de 300 líneas queda muy por debajo, y los rangos que la regla pide existen como parámetros de la herramienta. En Codex no hay una herramienta de lectura documentada con límite propio: es **no verificable** con la documentación, y la tarea 4.4 lo comprueba en el arnés.

Este umbral es de **lectura de archivos**. El de salidas de comando es otra regla (D3), porque los clientes recortan salidas en caracteres y no en líneas.

El valor vive en el catálogo y se ajusta sin tocar ninguna spec; 300 es el **valor inicial medido**, no un límite fijo. La regla de recalibración es mecánica y no depende de otra ficha: la tarea 4.2 mide el mismo conjunto en cada corrida y falla si 300 sale del intervalo p85–p92; en ese caso se elige el nuevo valor dentro del intervalo con el mismo criterio y se actualiza este cuadro.

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

La redacción no usa las clases del piso de permisos: el piso bloquea **rutas** (`Read(.env)`, `Edit(.env.*)`, `internal/harness/config/permissions/overlays.go:16-28`) y la redacción tiene que reconocer **valores** dentro de una salida. Además el piso aplica solo a Claude (`composeClaudeOverlay`, `overlays.go:73`); en Codex el instalador escribe únicamente `approval_policy` y `sandbox_mode` (`permissions/install.go:114-134`), así que no hay lista que heredar en ese cliente. Codex **sí tiene** un mecanismo nativo equivalente que el instalador no usa: `permissions.<name>.filesystem` acepta `"deny"` por ruta o patrón (*"glob subpaths such as `"**/*.env"` can deny reads"*) y `shell_environment_policy` filtra variables de entorno (https://learn.chatgpt.com/docs/config-file/config-reference). La documentación marca los perfiles de permisos como beta en algunos apartados; su estado actual es **no verificable**.

**Decisión** (**decidido por la 05 — pendiente de revisión del grupo**): una sola lista de **patrones de valor**, declarada en el catálogo, cuya dueña es esta ficha y que usan también 02, 06 y 10. Contenido inicial:

- `.env` y `.env.*`: el archivo y los valores de sus asignaciones
- tokens y API keys
- credenciales de nube
- claves SSH y `*.pem`
- valores de variables de entorno
- rutas del home del usuario
- identificadores de cuenta

La lista coincide con lo que la 06 ya redacta antes de compartir (variables de entorno, tokens, credenciales, rutas del home, `opsx-verifiable-evidence/design.md:79`) y con lo que la 10 necesita para su *"derivado redactado con lista de inclusión"* (`opsx-shadow-review/design.md:126`). Un test verifica que todo lo que bloquea el piso de Claude está cubierto por la lista (tarea 2.9): si el piso agrega una regla que la lista no cubre, la validación falla en lugar de divergir en silencio. Llevar un piso equivalente a Codex no es de esta ficha: se le pide a la 01 o a la 08, señalando las claves de arriba como el mecanismo concreto. La 01 (`propose/codex-strict-compat`) hoy no toca la lista sensible.

*Alternativa descartada: reusar el piso de permisos como única lista.* Era la decisión anterior. Se descarta porque mezcla dos objetos —rutas que se bloquean y valores que se redactan—, porque el piso es de Claude solo, y porque con el piso actual la redacción protegería únicamente `.env`.

### D15 — La lista de condiciones para leer un archivo completo es cerrada

Son cinco, y lo que no está acá no justifica lectura completa:

1. el archivo no supera el umbral de D11
2. la tarea va a reescribir el archivo entero
3. el archivo es el artefacto que la acción activa produce o edita
4. una lectura acotada previa sobre ese archivo devolvió vacío
5. los rangos no alcanzan, y el motivo queda declarado en el reporte con forma fija: `lectura completa: <ruta> — <motivo>`

La primera es la regla general: un archivo de 300 líneas o menos se lee completo. Las otras cuatro son las únicas excepciones para un archivo que supera el umbral, que en cualquier otro caso se lee por ruta, rango o símbolo.

*Por qué cerrada:* una lista abierta vuelve inaplicable el registro de lectura no justificada — siempre se puede alegar una condición nueva. Cerrada, el registro es decidible: o la lectura entra en una de las cinco o queda marcada.

La cuarta existe para no dejar al agente trabado: si pidió un símbolo y no estaba, leer entero es la salida razonable y no debería contar como falta.

La quinta (**decidido por la 05 — pendiente de revisión del grupo**) es la excepción justificada: hay archivos donde ningún rango sirve —una tabla de dispatch que cruza todo el archivo, un archivo generado cuyo orden no sigue al símbolo— y forzar rangos sale más caro que leer entero. No abre la lista, porque solo vale con el motivo escrito en el reporte: el arnés cuenta cuántas lecturas entran por la quinta y con qué motivos, y eso es el insumo para recalibrar el umbral de D11 sin tocar ninguna spec. Una lectura completa sin motivo declarado sigue quedando registrada como no justificada. 300 sigue siendo el valor inicial medido.

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
- **El comportamiento del agente** se verifica corriendo los fixtures contra el agente real con un arnés fuera del binario, como manda `CHANGES.md:49` y como lo planifica la 02. Es etapa posterior a esta ronda. Las corridas son **contra el cliente real** (Claude Code y Codex), no contra una simulación: el recorte de salidas que hace cada cliente (D3) forma parte del resultado que se mide.
- **"Resumir una salida"** incluye capturarla antes de leerla (D8): el módulo de resumen acotado prescribe la redirección a archivo, la captura del código de salida y la lectura dirigida del archivo. Un resumen hecho sobre lo que el cliente mostró inline, sin captura, es "parcial — sin verificar".
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

**A qué nivel queda "no habilitada" una herramienta** (**decidido por la 05 — pendiente de revisión del grupo**). "No se ofrece" puede querer decir tres cosas distintas, y la tabla no vale lo mismo en cada una:

| Nivel | Qué es | Claude Code | Codex |
|---|---|---|---|
| **(a) instrucción** | el módulo le pide al agente que no la use | `CLAUDE.md` o cuerpo del agente; la documentación dice que esas instrucciones se tratan *"as context, not enforced configuration"* (https://code.claude.com/docs/en/memory) | `AGENTS.md`, `developer_instructions` o cuerpo de la skill |
| **(b) descripción fuera del contexto** | el agente no ve que existe | skill con `disable-model-invocation: true`: *"Description not in context"* (https://code.claude.com/docs/en/skills); MCP con búsqueda de herramientas, activa por defecto, difiere las definiciones (https://code.claude.com/docs/en/mcp) | skill con `policy.allow_implicit_invocation: false`: la documentación no dice si la descripción sale del contexto (https://learn.chatgpt.com/docs/build-skills); un reporte de usuarios indica que sí sale (https://github.com/openai/codex/issues/23454) — **no verificable** en la documentación |
| **(c) el cliente no deja invocarla** | aunque el agente lo intente, la llamada no ocurre | `permissions.deny` con `Skill(skill:<nombre>)`, `Agent(<nombre>)` o `mcp__<servidor>__*`; un nombre pelado *"removes the tool from Claude's context entirely"* (https://code.claude.com/docs/en/permissions); `tools`, `disallowedTools` y `mcpServers` en el archivo de un subagente (https://code.claude.com/docs/en/sub-agents) | `mcp_servers.<id>.enabled_tools` / `disabled_tools` y `enabled` en `config.toml` (https://learn.chatgpt.com/docs/config-file/config-reference); `skills.config` y `features.*=false` en el archivo de un rol (https://learn.chatgpt.com/docs/agent-configuration/subagents) |

Lo que decide la tabla de arriba es **por acción**, dentro de una misma sesión, y eso cambia el nivel alcanzable:

- **Claude:** por acción, (c) solo si la acción corre en un subagente con su propia lista (`tools`, `mcpServers`, `skills`). En el hilo principal la lista queda fija para toda la sesión, así que ahí el control por acción es (a).
- **Codex:** el archivo de un rol no puede fijar `mcp_servers` —*"Configure MCP servers in the parent session"*, https://learn.chatgpt.com/docs/agent-configuration/subagents—, así que el control por acción de una integración MCP es **solo (a)**, y se declara como límite del cliente. Para skills, `skills.config` en el rol da (c) por rol.
- Una skill de Claude con `allowed-tools` **no** restringe: *"It does not restrict which tools are available"* (https://code.claude.com/docs/en/skills). Solo preaprueba, y no se cuenta como (c).

El catálogo declara, por integración opcional, por cliente y por superficie (CLI, IDE, app), qué nivel alcanza la tabla (tarea 5.9). Que una skill o un MCP esté instalado, o no, no prueba aislamiento por tarea: el nivel se declara por lo que el cliente documenta que impide, no por lo que está instalado. El nivel (a) nunca se presenta como aislamiento: una integración "no habilitada" al nivel (a) sigue invocable, y si el agente la invoca, queda registrado como uso fuera de clase.

### D19 — Interfaz con la 07: umbral de salida verbosa y obligatoriedad de módulos

**Decidido por la 05 — pendiente de revisión del grupo.** La 07 consume esta ficha (`opsx-selective-delegation/proposal.md:4`) y cita dos cosas que hay que fijar:

- **"Salida verbosa"** es una salida de más de 300 líneas o de más de 10.000 caracteres, lo que ocurra primero (D3). Las 300 líneas coinciden con el umbral medido de lectura de archivos (D11); los 10.000 caracteres son el límite documentado más bajo con que un cliente recorta una salida fallida. Es la condición para capturar y resumir, no un tope de recorte. Reemplaza "el presupuesto de 05" de la señal S1 de la 07 (`opsx-selective-delegation/design.md:40`), que no existe: D3 rechaza presupuestos de tamaño.
- **El módulo de delegación es obligatorio solo en `verify`**, por la separación de rol de la señal S4 de la 07 (`design.md:43`). En el resto de las acciones es opcional: si falta, el agente avisa y resuelve directo. Así la regla de detener de `task-context-loading` —respaldada por la *Aceptación transversal* (`CHANGES.md:47`: *"Si falta un módulo obligatorio, informar la limitación y detener la acción que lo requiere"*)— aplica a los módulos obligatorios y es compatible con la D2 de la 07 (`design.md:58`: *"se informa y se resuelve directo o se detiene la acción si la acción exige delegar (S4)"*).

Tabla de carga, que es el mapeo que pide la ficha (`CHANGES.md:164`: *"Mapeá con 03, 04 y 07 cuándo se cargan TDD, delegación y coordinación de subagentes"*):

| Módulo | `explore` | `propose` | `apply` | `verify` | `archive` |
|---|---|---|---|---|---|
| `tdd` | — | — | obligatorio si el proyecto tiene Strict TDD activo y runner | — | — |
| `delegation` (con coordinación de subagentes) | opcional, si la regla de la 07 decide delegar | opcional, ídem | opcional, ídem | **obligatorio** (S4) | opcional, ídem |
| `model-routing` | con `delegation` | con `delegation` | con `delegation` | con `delegation` | con `delegation` |
| `engram` — protocolo detallado | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre | al inicio y al cierre |
| `governance`, `engram` — regla corta | núcleo | núcleo | núcleo | núcleo | núcleo |

**Superficie sin subagentes.** Codex documenta la actividad de subagentes en la app de escritorio, el CLI y la extensión del IDE, y no en la nube (https://learn.chatgpt.com/docs/agent-configuration/subagents). En una superficie sin subagentes, `verify` queda **detenida por diseño**: el módulo de delegación es obligatorio ahí y no puede entregarse. Se declara como límite de esa superficie, no como fallo; qué superficies soporta cada cliente lo fijan 03 y 04.

### D20 — Un resumen de un worker sin corroborar queda "sin verificar"

**Decidido por la 05 — pendiente de revisión del grupo.** Cuando un especialista (worker) corre un comando y le devuelve al coordinador un resumen, el coordinador recibe una afirmación, no evidencia: no vio el original y no sabe si el worker capturó antes de leer (D8). Ese resumen queda **sin verificar** hasta que el coordinador lo corrobora de una de dos formas:

1. contra el original o su referencia: la ruta existe, la huella sha256 coincide y los fallos declarados están en el original; o
2. volviendo a correr el comando con captura, si no tiene efectos laterales.

Corroborado, el resumen pasa a **verificado**; si la corroboración encuentra un fallo omitido o una huella que no coincide, a **rechazado**. Los tres estados son los de la 06 (`opsx-verifiable-evidence/design.md:57-61`), con la duda hacia *sin verificar*, y no se inventa un vocabulario propio. La 05 no define qué evidencia alcanza para aceptar un cambio —eso es de la 06—: define solo que un resumen sin corroborar no cuenta como corroborado.

*Alternativa descartada: confiar en el resumen del worker si declara productor.* La declaración de productor (D8) hace atribuible un incumplimiento; no lo detecta. Un worker que no capturó produce un resumen bien formado y falso.

### D21 — El núcleo nombra la invocación por cliente, no una sola sintaxis

**Decidido por la 05 — pendiente de revisión del grupo.** El núcleo incluye *"la referencia para invocar al orquestador"* (`specs/task-context-loading`). Esa referencia no puede ser un texto único, porque la invocación explícita difiere por cliente y por superficie:

| Cliente | Invocación explícita documentada | Fuente |
|---|---|---|
| Claude Code | skill: `/nombre` al inicio del mensaje | https://code.claude.com/docs/en/skills |
| Codex CLI e IDE | `$nombre` o `/skills` | https://learn.chatgpt.com/docs/build-skills |
| Codex en la app de ChatGPT | `@` para elegir la skill | https://learn.chatgpt.com/docs/build-skills |

El núcleo lleva **un texto por cliente**, y la referencia a las skills `openspec-*` con la sintaxis de cada uno (`/openspec-propose` en Claude, `$openspec-propose` en Codex). **Qué es el orquestador** en cada cliente es de 03 y 04. La 03 lo proponía como agente elegido con `claude --agent <id>` (`claude-native-opsx/design.md:116`) y la 04 como skill `$sdd-orchestrator` (`codex-native-opsx/design.md:153-160`). La decisión confirmada del equipo para 03/04 (auditoría técnica del 10/11) es activarlo **explícitamente dentro de la sesión actual con una skill**:

- En Claude Code, `/sdd-orchestrator` con `disable-model-invocation: true` (https://code.claude.com/docs/en/skills).
- En Codex, `$sdd-orchestrator` con `policy.allow_implicit_invocation: false` en `agents/openai.yaml` (https://learn.chatgpt.com/docs/build-skills).

El núcleo nombra esas dos invocaciones. Hay dos salvedades:

- En Codex, la invocación con `$` de una skill con `allow_implicit_invocation: false` tiene un bug abierto (https://github.com/openai/codex/issues/23454). El texto del núcleo indica el camino alternativo `/skills`, y el spike de la 04 lo verifica por versión.
- Si 03 y 04 cambian el nombre, el núcleo toma el que fijen por superficie.

**Qué especialistas reciben el núcleo.** En Claude el núcleo vive en `CLAUDE.md`, y los subagentes integrados Explore y Plan *"start without your CLAUDE.md files"*, igual que un agente con `omitClaudeMd: true` (https://code.claude.com/docs/en/sub-agents). Para esos especialistas, el encargo que arma el coordinador **repite** las reglas de cumplimiento obligatorio y `governance`. En Codex no está documentado si un subagente lee `AGENTS.md` (https://learn.chatgpt.com/docs/agent-configuration/agents-md): hasta que el spike de la 04 lo responda, el encargo también las repite.

## Risks / Trade-offs

- **La matriz de módulos de 03 y 04 todavía no está cerrada →** la 05 usa su propio reparto (D2) hasta que exista la matriz única propuesta (punto *Matriz de módulos*), y la tarea 3.1 registra las divergencias contra ella. El criterio de reparto no depende de esa matriz.
- **Seis fichas tocan `sdd-orchestrator.md` y ninguna lo posee →** este change escribe requisitos y redacta texto de módulos como entrega (D17), no contenido en el archivo, y lleva al grupo la propuesta de que 03 y 04 sean dueñas. La 07 ya la respalda (`opsx-selective-delegation/design.md:173`: *"03+04, como propone 05"*).
- **Resolver módulos por acción agrega un punto de fallo que hoy no existe →** mitigado por el requisito de que un módulo obligatorio ausente detenga la acción e informe, que es el modo de fallo que la aceptación transversal ya exige (`CHANGES.md:47`). D19 fija cuáles son obligatorios para que la regla no frene acciones que pueden resolverse directo.
- **Mover instrucciones fuera del bloque siempre presente puede degradar el comportamiento del agente en acciones que sí las necesitaban →** por eso ninguna adopción se decide en este change. La comparación con y sin cada módulo, sobre la misma aceptación, es requisito de `task-context-loading` y usa el protocolo de 02 (tarea 6.4).
- **El contrato de resumen no sirve de nada si el productor no retiene el fallo →** el productor está decidido (D8) y su supuesto se valida corriendo los fixtures del grupo 2 contra un agente real, en el cliente real (tareas 2.7 y 2.17). El contrato evalúa el resumen, así que cambiar el productor no invalida nada.
- **El cliente recorta la salida antes de que el agente la vea →** el agente captura a archivo antes de leer (D8), con el umbral en la unidad del cliente (D3). Lo que no se capturó se entrega como "parcial — sin verificar", nunca como resumen completo.
- **El agente puede no capturar, o un worker puede resumir sin capturar →** un resumen sin original capturado es parcial (D8) y el de un worker queda sin verificar hasta que el coordinador lo corrobora (D20).
- **Una tabla de herramientas que en un cliente es solo instrucción puede leerse como aislamiento →** D18 declara el nivel alcanzable por cliente y prohíbe presentar el nivel (a) como control.
- **Solapamiento con 07 →** la 07 consume esta ficha y ya la declara dependencia. D19 fija el umbral de salida verbosa y qué módulo es obligatorio en qué acción, y la tabla de carga queda acá para que la 07 la consuma sin redefinirla.
- **La redacción por instrucción puede fallar →** un agente puede conservar un valor sensible, y la captura de D8 escribe un crudo antes de redactar. Mitigación: la lista de D14 es declarativa y testeable, el fixture 2.4 lo evalúa, el crudo se borra antes de entregar el resumen, y el original vive fuera del repo (D9), de modo que un fallo de redacción no termina en un commit. Las copias propias del cliente quedan fuera de esta garantía y se declaran como límite (D9); la protección en origen es no leer el secreto.

## Migration Plan

No aplica en esta ronda: el change no modifica configuración instalada. Cuando se implemente, la migración del bloque administrado en `CLAUDE.md`/`AGENTS.md` pertenece a 03 y 04, que ya la tienen en su alcance con snapshot, preservación de contenido del usuario y rollback. Este change no debe introducir una segunda ruta de migración sobre los mismos archivos.

**OpenCode y Antigravity** mantienen el bloque global actual. Es un límite declarado de la 05: el roadmap es de Claude y Codex (`CHANGES.md:3`) y ninguna ficha tiene asignado qué les pasa al retirar el bloque global; queda sin dueño hasta que el roadmap lo asigne.

## UI Mockups

Decision: no (no UI) — 2026-11-05. El change no agrega ni modifica pantalla, componente, layout ni interacción.

## Decisiones tomadas — pendientes de revisión del grupo

Doce puntos. Todos están decididos acá; se comparten porque tocan alcance de otras fichas. Cada uno con la decisión, en qué se apoya, a qué fichas afecta y qué cambia si el grupo la rechaza. Se identifican por su nombre, no por un código: así nadie tiene que buscar una leyenda y no se confunden con los puntos de otra ficha.

Una decisión puede apoyarse en tres cosas, y cada una lleva su procedencia, igual que D5 exige para cualquier afirmación sobre el código:

| | Qué significa |
|---|---|
| **documentado** | está escrito en el repositorio, en el roadmap o en la rama de otra ficha, con archivo y línea, o en la documentación oficial del cliente, con su URL |
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
| **Decisión** | la tabla de D18: todo opcional apagado por defecto y habilitado por acción, con el nivel de control —(a) instrucción, (b) descripción fuera del contexto, (c) el cliente no deja invocarla— declarado por cliente; (a) nunca se presenta como aislamiento |
| **En qué se apoya** | **documentado**: `CHANGES.md:166` pone *"qué capacidad es necesaria para cada tarea"* entre lo que se acuerda en grupo; el catálogo declara Context7 y Playwright (`harnesses.yaml:36`, `:50`). Los niveles por cliente: **documentado** en https://code.claude.com/docs/en/permissions, https://code.claude.com/docs/en/skills, https://code.claude.com/docs/en/sub-agents y https://learn.chatgpt.com/docs/agent-configuration/subagents (el archivo de un rol de Codex no fija `mcp_servers`). La asignación por acción es **supuesto** razonado; se recalibra con la comparación con y sin cada integración del grupo 6 |
| **Afecta a** | 03, 04, 05, 07, 08 |
| **Si se rechaza** | la tabla de D18 y las tareas 5.1, 5.8 y 5.9 |

### Referencia, ubicación y retención del original

| | |
|---|---|
| **Decisión** | referencia `{comando, id de corrida, ruta, huella sha256, vencimiento}`; original en `<os.UserCacheDir>/active-stack/runs/<change>/`; retención hasta archivar el change y siete días más; acceso del usuario local; el crudo de la captura se borra y la garantía de redacción cubre solo la copia de Active Stack, no las del cliente (D9). Se propone a la 06 adoptar el formato |
| **En qué se apoya** | **documentado**: `CHANGES.md:166` pone *"dónde queda el original"* entre lo que se acuerda en grupo y pide coordinar la referencia con 06; la 06 tampoco versiona originales (`opsx-verifiable-evidence/design.md:78`); el cliente guarda su propia copia (https://code.claude.com/docs/en/tools-reference). **Supuesto**: los siete días; se recalibran con el primer uso real (D9) |
| **Afecta a** | 05, 06 |
| **Si se rechaza** | los valores del requisito de ubicación y retención y las tareas 2.10, 2.11 y 2.14 |

El commit `1b4639b` no es la fuente de esta decisión: es lo que volvió urgente la pregunta. Antes `openspec/` estaba en el `.gitignore` y una salida guardada en la carpeta del change no podía llegar al repositorio ni por error; ahora sí. Por eso el original vive fuera del repo.

### Datos sensibles

| | |
|---|---|
| **Decisión** | una sola lista de patrones de valor en el catálogo, dueña la 05, usada por 02, 06 y 10; un test garantiza que cubre el piso de Claude (D14) |
| **En qué se apoya** | **documentado**: el piso bloquea rutas y solo en Claude (`overlays.go:16-28`, `:73`); el instalador no escribe piso en Codex (`permissions/install.go:114-134`), aunque Codex tiene `permissions.<name>.filesystem` con `"deny"` (https://learn.chatgpt.com/docs/config-file/config-reference); la 06 redacta valores (`opsx-verifiable-evidence/design.md:79`) y la 10 consume una redacción (`opsx-shadow-review/design.md:126`) |
| **Afecta a** | 01, 02, 05, 06, 08, 10 |
| **Si se rechaza** | el requisito de datos sensibles y las tareas 2.4 y 2.9 |

### Productor del resumen acotado

| | |
|---|---|
| **Decisión** | el agente bajo instrucción, que **captura antes de leer**: redirige stdout y stderr a un archivo de la ruta de corrida con el código de salida, y resume desde ese archivo; sin captura, el resumen es "parcial — sin verificar" (D8) |
| **En qué se apoya** | **documentado**: un envoltorio ejecutable sería un runtime nuevo (`CHANGES.md:49`); la capacidad nativa existe en los dos pero recorta sin conservar fallos (https://code.claude.com/docs/en/tools-reference, https://learn.chatgpt.com/docs/config-file/config-reference). **Supuesto**: que el agente captura y retiene el fallo; se valida con las tareas 2.7 y 2.17 |
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
| **Decisión** | salida verbosa = más de 300 líneas o más de 10.000 caracteres, lo que ocurra primero; delegación obligatoria solo en `verify`, que queda detenida en una superficie sin subagentes; tabla de carga de D19 |
| **En qué se apoya** | **documentado**: `CHANGES.md:47` y `:164`; la 07 en `opsx-selective-delegation/design.md:40`, `:43` y `:58`. **Medido**: las 300 líneas salen de D11. **Documentado**: los 10.000 caracteres son el límite de Claude para un resultado fallido (https://code.claude.com/docs/en/tools-reference) |
| **Afecta a** | 05, 07 |
| **Si se rechaza** | el requisito de umbral de salida verbosa, la tabla de D19 y la tarea 3.5 |

### Agentes incluidos

| | |
|---|---|
| **Decisión** | solo Claude y Codex; OpenCode y Antigravity mantienen el bloque global actual como límite declarado (Migration Plan) |
| **En qué se apoya** | **documentado** en `CHANGES.md:3` —*"plan canónico de mejoras de OPSX para Codex y Claude"*— y `CHANGES.md:158` (*Cliente Ambos*). La 07 declara lo mismo (`opsx-selective-delegation/proposal.md:61`); la 06 lo tiene como pregunta abierta nº 5 (`opsx-verifiable-evidence/design.md:114`) |
| **Afecta a** | 05, 06, 07, 08 |
| **Si se rechaza** | el alcance de la paridad por cliente y la tarea 1.2 |

### Estado de un resumen sin corroborar

| | |
|---|---|
| **Decisión** | el resumen que un worker le entrega al coordinador queda *sin verificar* hasta que el coordinador lo corrobora contra el original o su referencia, o con una re-ejecución; corroborado pasa a *verificado*, y con un fallo omitido o una huella distinta, a *rechazado* (D20) |
| **En qué se apoya** | **documentado**: los tres estados y la duda hacia *sin verificar* son de la 06 (`opsx-verifiable-evidence/design.md:57-61`); el cliente recorta salidas antes de que el worker las vea (https://code.claude.com/docs/en/tools-reference) |
| **Afecta a** | 05, 06, 07 |
| **Si se rechaza** | el requisito de resumen de un worker y la tarea 2.18 |

### Referencia de invocación por cliente

| | |
|---|---|
| **Decisión** | el núcleo lleva un texto por cliente con la invocación explícita documentada —Claude Code `/nombre`; Codex `$nombre` en CLI e IDE y `@` en la app de ChatGPT—; el nombre final del orquestador y si es agente o skill son de 03 y 04; no se fija `claude --agent` en el núcleo; los especialistas que no cargan `CLAUDE.md` reciben las reglas obligatorias en el encargo (D21) |
| **En qué se apoya** | **documentado** en https://code.claude.com/docs/en/skills, https://learn.chatgpt.com/docs/build-skills y https://code.claude.com/docs/en/sub-agents (Explore y Plan sin `CLAUDE.md`; `omitClaudeMd`); `claude-native-opsx/design.md:116` y `codex-native-opsx/design.md:153-160` |
| **Afecta a** | 03, 04, 05 |
| **Si se rechaza** | los requisitos de núcleo y de referencia de invocación de `task-context-loading` y las tareas 3.7 y 3.11 |

### La frontera con la 06, que no es una decisión

La 05 y la 06 son **hermanas**: las dos dependen de 03 y 04, ninguna de la otra. La 06 lo escribió desde su lado, en su *Fuera de alcance*: *"Lectura dirigida y acotado de salidas (propuesta 05). Aquí solo se exige conservar la referencia al original."* Esta ficha lo escribe desde el suyo. Comparten el formato de la referencia al original y su vencimiento, los dos definidos por la 05 (D9); lo que se le propone a la 06 es adoptarlos, junto con la lista de patrones de D14.
