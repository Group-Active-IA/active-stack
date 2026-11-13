# Tasks

> Todas las tareas quedan **pendientes**. Revisión del 2026-11-13 contra el informe unificado de
> revisión OPSX (11/11) y la matriz común de reglas publicada por la 03 el 12/11
> (`origin/claude-native-opsx:openspec/matriz-reglas-opsx.md`), más las pruebas reales del 13/11
> y la entrega a 03 y 04 (`entrega-03-04.md`). El orden refleja dependencias reales, no un
> cronograma. Las pruebas reales avanzan varias tareas pero no cumplen la aceptación de ninguna
> (n = 1, generadores propios en lugar de los fixtures del arnés, solo Windows): cada tarea
> afectada cita su prueba y sigue pendiente.
>
> **Estado de cada tarea.** Cada tarea lleva una de estas marcas:
>
> - **[ahora]** — se implementa en este repo sin esperar a que 03 o 04 se integren a `main`:
>   catálogo y su validación (`internal/catalog`, `internal/model`, BAJO), textos de módulo como
>   archivos (`internal/harness/config/assets/modules/`, MEDIO, sin cablearlos), fixtures, el
>   arnés de esta ficha (`eval/taskcontext/`, fuera del binario) y `docs/`.
> - **[bloqueada: …]** — nombra qué la bloquea: la integración de 03 o 04 a `main`, clientes
>   reales con una persona que corra la sesión, una plataforma, o un acuerdo con otra ficha.
>
> **Gobernanza** (`CLAUDE.md` §4). Ninguna tarea toca `internal/backup`, `internal/filemerge` ni
> `internal/pipeline`. Si una implementación descubre que necesita tocarlos, se detiene ahí, lo
> registra en *Workflow follow-up* y espera revisión humana fechada: esos paquetes son ALTO y
> no se tocan en modo autónomo. La migración del bloque instalado es de 03 y 04.
>
> **Quién ejecuta** (`design.md` — D17). Los tests de Go verifican lo declarativo y lo mecánico:
> catálogo validado por `catalog.Load()`, textos de módulo como archivos, el capturador, el
> redactor y el evaluador del arnés. La conducta del agente se verifica en el grupo 8, contra el
> cliente real. El arnés de esta ficha vive en `eval/taskcontext/`, usa solo la biblioteca
> estándar, puede importar `eval/baseline/` y `internal/catalog`, y nada de `cmd/` ni `internal/`
> lo importa (`CHANGES.md:49`; 02 D6 en `openspec/changes/opsx-baseline/design.md`).
>
> **Prioridad.** Primero los grupos 1, 2, 3 y 6 (inventario, salida acotada con captura y carga
> por acción). Los grupos 4, 5 y 7 siguen después.
>
> **Resumen.** 64 tareas: 46 **[ahora]** y 18 **[bloqueadas]** (la 6.8, parcialmente). En los grupos 8 y 9 la marca va
> en el encabezado y vale para cada tarea del grupo. Detalle de bloqueos al final.

## 1. Inventario reproducible del contexto de partida

- [ ] 1.1 **[ahora]** Agregar un golden de inventario en `internal/harness/config/testdata/` que registre, para las seis combinaciones acumulativas de toggles en los variantes `claude` y `codex`, los bytes y las líneas (`wc -l`, cantidad de saltos de línea) que produce `Compose`, y verificar que el test compara contra el golden; un cambio de asset de otra ficha se resuelve regenerándolo con `go test -update` (bandera ya existente, `internal/harness/config/compose_test.go:14`). Valores de partida en `1b4639b`: 20.065 B / 448 L en Claude y 18.010 B / 403 L en Codex. Si para entonces la 03 ya integró sus goldens de `Compose` (`compose_all_toggles_codex.golden`, tarea 4.0 de la 03), el inventario los reutiliza en lugar de duplicarlos
- [ ] 1.2 **[ahora]** Registrar en el mismo golden la medición de las demás variantes de assets (`antigravity`, `cursor`, `gemini`, `generic`, `opencode`, `windsurf`) como referencia, fuera del alcance de paridad de esta ficha, y verificar que ninguna variante de `internal/harness/config/assets/` queda sin medición registrada
- [ ] 1.3 **[ahora]** Documentar en `docs/` la atribución de bytes por fragmento de la tabla de `proposal.md` — Impact y verificar que el documento reproduce las cifras del golden de 1.1
- [ ] 1.4 **[ahora]** Verificar con un test que en `main` Claude y Codex componen por la rama `default` de `internal/harness/config/install.go:70` y OpenCode por la rama `ConfigDeliveryPrimaryAgent` (`install.go:53`). El test documenta la ruta vigente; cuando 03 o 04 integren su modo de entrega (la 03 ya tiene `ConfigDeliveryManagedSkill` en su rama), esa integración lo actualiza en el mismo cambio
- [ ] 1.5 **[ahora]** Inventariar el tamaño de cada cuerpo de skill del catálogo (propias y de terceros) y verificar que el inventario cubre todas las skills que `harnesses.yaml` declara, sin ninguna sin medir
- [ ] 1.6 **[ahora]** Inventariar las demás fuentes de instrucción que el instalador escribe en el proyecto destino, incluidos los fragmentos de sección propia como `design-mockups` (8.865 B), y verificar que la suma por cliente coincide con el total realmente inyectado
- [ ] 1.7 **[ahora]** Publicar el inventario de 1.1 a 1.6 como una sola tabla ordenada por tamaño descendente y verificar que identifica las cinco fuentes más grandes
- [ ] 1.8 **[ahora]** Inventariar las salidas de comando del ciclo OPSX sobre este repositorio —tests, build, vet, logs y OpenSpec CLI por acción— y calibrar con ellas el umbral de salidas de `design.md` — D3: registrar líneas y caracteres de cada salida, en éxito y con un fallo provocado, y verificar que cada fuente declara su acción, si supera 300 líneas o 10.000 caracteres y si el par deja afuera alguna salida que el cliente recortaría. Si la calibración contradice el par, se actualiza D3 con el mismo criterio

## 2. Fixtures de salida y contrato del resumen

> Todo en `eval/taskcontext/`. Los fixtures son comandos deterministas cuya salida el arnés
> conoce de antemano: esa salida es la referencia contra la que se comparan la captura y el
> resumen (`design.md` — D4 y D8).

- [ ] 2.1 **[ahora]** Escribir los fixtures de salida grande con el fallo al inicio, al medio y al final, cada uno con su comando, su salida esperada, su código de salida y el fallo que debe sobrevivir, y verificar con un test que los tres tienen miles de líneas y superan los 10.000 caracteres, de modo que el cliente los recortaría sin captura (`design.md` — D3)
- [ ] 2.2 **[ahora]** Agregar el fixture de fallos dispersos en más de una posición y verificar que declara cada fallo relevante por separado
- [ ] 2.3 **[ahora]** Agregar el fixture de control negativo —salida verbosa, repetitiva y con código cero— y verificar que no declara ningún fallo esperado
- [ ] 2.4 **[ahora]** Agregar el fixture de **datos sensibles ficticios**: un valor inventado por cada patrón de la lista de `design.md` — D14 (asignación de `.env`, token, API key, credencial de nube, clave privada SSH, ruta del home, identificador de cuenta), con un fallo real en la misma salida. Verificar con un test que cada valor coincide con su patrón, que ninguno es un secreto real (todos llevan el prefijo `FICTICIO`), y que el original redactado esperado conserva cada línea en su posición con la marca de omisión
- [ ] 2.5 **[ahora]** Agregar el fixture de ruido que imita el formato de un error con código cero, y verificar que su resumen esperado no declara ningún fallo
- [ ] 2.6 **[ahora]** Agregar el fixture con un fallo real rodeado de ruido que lo imita, y verificar que su resumen esperado declara exactamente un fallo
- [ ] 2.7 **[ahora]** Agregar el fixture de un comando que falla escribiendo el fallo **solo por stderr**, con stdout verboso sin fallos, y verificar que su resumen esperado contiene la línea de stderr y el código distinto de cero
- [ ] 2.8 **[ahora]** Agregar el fixture de **original nunca capturado** —comando ya corrido sin captura y archivo ausente— y verificar que su resumen esperado está marcado "parcial — sin verificar", dice por qué, no lleva ruta ni huella, y que pedir su original informa "nunca capturado" y no "vencido"
- [ ] 2.9 **[ahora]** Escribir el evaluador del contrato del resumen: dado un fixture y un resumen, decide *aceptado*, *insuficiente* o *rechazado* según conserve cada fallo esperado y el código de salida, no declare fallos que no ocurrieron, lleve la referencia de cinco campos con huella coincidente y declare capturador, redactor y productor. Verificar con resúmenes escritos a mano, uno conforme y uno por cada incumplimiento (fallo omitido, fallo inventado, referencia incompleta, huella distinta, capturador sin declarar), que cada uno recibe el veredicto esperado y que el rechazo registra el productor declarado
- [ ] 2.10 **[ahora]** Declarar en el catálogo los patrones de fallo por herramienta (`go test`, `go build`, `go vet` para este repo; `design.md` — D3) y verificar con los fixtures 2.1 a 2.7 que el evaluador identifica exactamente los fallos esperados, y que una salida sin patrón declarado se evalúa solo por código de salida y lo dice
- [ ] 2.11 **[ahora]** Documentar el contrato de resumen en `docs/` con un ejemplo de entrada y salida, y verificar que el ejemplo coincide con el resumen esperado del fixture 2.1

## 3. Captura independiente, redacción y retención

> Tres roles separados (`design.md` — D8): quién captura, quién redacta y quién resume. En el
> arnés captura y redacta código del arnés; el agente solo resume. La matriz de cobertura por
> fuente es D22.

- [ ] 3.1 **[ahora]** Declarar en el catálogo la lista única de patrones de valor sensibles (`design.md` — D14) y verificar con un test que `catalog.Load()` la valida y que todo lo que bloquea el piso de Claude (`internal/harness/config/permissions/overlays.go:16-28`) está cubierto por algún patrón; si el piso agrega una regla no cubierta, el test falla nombrándola
- [ ] 3.2 **[ahora]** Declarar en el catálogo el umbral de salidas de comando en dos unidades (300 líneas y 10.000 caracteres, `design.md` — D3), aparte del umbral de lectura de archivos de la tarea 4.1, y verificar que `catalog.Load()` rechaza un umbral de salidas al que le falte una de las dos unidades
- [ ] 3.3 **[ahora]** Implementar el capturador del arnés (`eval/taskcontext/capture`): ejecuta el comando del fixture, recibe stdout y stderr en un mismo flujo ordenado, registra el código de salida, redacta en memoria con la lista de 3.1 y escribe solo la copia redactada, su huella sha256 y el registro de captura. Verificar con los fixtures 2.1, 2.4 y 2.7 que la copia coincide con la referencia redactada línea por línea, que ningún archivo de la ruta de corrida contiene un valor de 2.4 sin redactar, que la línea de stderr está y que el código registrado es el del comando. Corre en Windows; en Linux, ver 8.1
- [ ] 3.4 **[ahora]** Definir el registro de captura (`design.md` — D9) con fuente de captura (`arnés`, `cliente:<fuente>` o `agente`), redactor, cliente, interfaz, sistema, versión, código de salida, cobertura antes del truncamiento (`demostrada` o `no demostrada`) y conteo de redacciones por categoría, y verificar con un test que un resultado cuya fuente es `agente`, o cuya cobertura es `no demostrada`, nunca supera *sin verificar* (`design.md` — D20)
- [ ] 3.5 **[ahora]** Implementar en el arnés la ubicación y la retención (`design.md` — D9): `<os.UserCacheDir>/active-stack/runs/<change>/<id de corrida>/`, `_sin-change` cuando no hay change activo, y purga de los vencidos al iniciar cada corrida. Verificar con reloj inyectable que un original dentro de plazo se recupera, que uno vencido se informa como no disponible y desde qué fecha, que uno nunca capturado se informa distinto de uno vencido, y que nada queda en el árbol del repositorio
- [ ] 3.6 **[ahora]** Implementar los adaptadores de fuente del cliente sobre muestras derivadas de la documentación (`eval/taskcontext/fixtures/samples/`, con cita por muestra): hook `PostToolUse` y `PostToolUseFailure` de Claude Code; `item/commandExecution/outputDelta` e `item/completed` de Codex App Server; `command_execution` de `codex exec --json`, cuya muestra sale de una corrida real (`aggregated_output` completo, `entrega-03-04.md` §5, X7). Escribir el comparador de cobertura que contrasta lo que entrega la fuente con la referencia del arnés y devuelve `completa`, `truncada` (incluida la marca `... [N characters truncated] ...`), `ausente` o `no comprobada`, y verificar con una muestra de cada caso que el veredicto es el esperado
- [ ] 3.7 **[ahora]** Declarar en el arnés la matriz de cobertura de `design.md` — D22 (cliente × interfaz × sistema × fuente, con estado distribuida, declarada o probada) y verificar con un test que cada celda exigida —Claude y Codex, CLI e IDE, Windows y Linux— tiene una entrada, que una celda solo puede figurar `probada` con el id de una corrida del grupo 8, y que una fuente que solo existe en CLI figura `no disponible` en IDE
- [ ] 3.8 **[ahora]** Fijar y probar las recetas de captura declarada por el agente (`design.md` — D8): bash con redirección conjunta y código del comando; en Windows, `cmd /c` invocado desde PowerShell 5.1 y 7, para no pasar por la codificación ni el envoltorio de errores de PowerShell. Verificar con un test que ejecuta un comando de prueba por cada shell disponible que el archivo queda en bytes del comando, sin BOM ni `NativeCommandError`, con la línea de stderr y el código de salida del comando y no el de una tubería. El caso bash en Linux corre en 8.1

## 4. Lectura dirigida y procedencia

- [ ] 4.1 **[ahora]** Declarar el umbral de lectura completa en el catálogo, en líneas y con 300 como valor (`design.md` — D11), y verificar que `catalog.Load()` rechaza un umbral expresado en otra unidad
- [ ] 4.2 **[ahora]** Agregar un test que mida la distribución de líneas de los archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en `_test.go` —el conjunto de `design.md` — D11— y falle si el umbral declarado deja de estar entre el p85 y el p92 por rango más cercano
- [ ] 4.3 **[ahora]** Declarar en el catálogo la lista cerrada de cinco condiciones de lectura completa (`design.md` — D15) y verificar que `catalog.Load()` rechaza una condición fuera de la lista
- [ ] 4.4 **[ahora]** Pasar a `internal/harness/config/assets/modules/directed-code-reading.md` el texto ya redactado en `entrega-03-04.md` §2, conforme a `specs/directed-code-reading`. Verificar con un test que el archivo existe, no lleva frontmatter de skill, coincide byte a byte con el de la entrega y contiene las formas fijas `full read: <path> — <reason>` y `unnecessary call: <tool> <input>`
- [ ] 4.5 **[ahora]** Verificar con un fixture y el evaluador el orden entre tipos de `design.md` — D13: una ruta con la definición del símbolo se ordena antes que otra con muchas referencias, y la cantidad solo ordena dentro del mismo tipo
- [ ] 4.6 **[ahora]** Escribir los fixtures de lectura que corre la tarea 8.6, con su registro esperado: lectura por símbolo anidado con la firma de la declaración que lo contiene, lectura completa no justificada, quinta condición con motivo y sin motivo, relectura sin escritura intermedia y después de una escritura, y recorrido archivo por archivo para orientarse. Verificar que el evaluador cuenta por separado las lecturas que entran por la quinta condición
- [ ] 4.7 **[ahora]** Documentar en `docs/` cuándo corresponde cada modo de lectura, el umbral con su justificación medida y las cinco condiciones, y verificar que cada ejemplo documentado corresponde a un fixture de 4.5 o 4.6

## 5. Herramientas por clase de tarea

- [ ] 5.1 **[ahora]** Declarar en el catálogo la tabla de herramientas por acción de `design.md` — D18 y el valor por defecto desactivado de las integraciones opcionales, y verificar que `catalog.Load()` rechaza una integración opcional sin acción declarada
- [ ] 5.2 **[ahora]** Declarar en el catálogo, por integración opcional y por cliente, el nivel alcanzable con su nombre —(a) *solo recomendada*, (b) *fuera del contexto*, (c) *restringida por el cliente*— y el mecanismo de `design.md` — D18, y verificar que `catalog.Load()` rechaza una integración sin nivel para Claude o para Codex y que `allowed-tools` no se acepta como nivel (c)
- [ ] 5.3 **[ahora]** Pasar a `internal/harness/config/assets/modules/task-tool-selection.md` el texto de `entrega-03-04.md` §2, conforme a `specs/task-tool-selection`. Verificar con un test que el archivo existe, no lleva frontmatter de skill, no presenta el nivel (a) como bloqueo, contiene `unavailable:`, `unnecessary call:` y `out-of-class use:` en su forma fija, y que su tabla coincide con la del catálogo (5.1)
- [ ] 5.4 **[ahora]** Escribir los fixtures de herramientas que corre la tarea 8.7, con su registro esperado: integración instalada y no habilitada, habilitada y ausente con y sin camino alternativo, invocación repetida sin cambio de estado, sondeo de una integración ya informada como no disponible, y uso fuera de clase con control (a)
- [ ] 5.5 **[ahora]** Generar en `docs/` desde el catálogo la tabla de herramientas por acción, con su estado por defecto y el nivel por cliente, y verificar con un test que el documento coincide con la generación

## 6. Módulos por acción y entrega a 03 y 04

> La 05 define **cuándo** se carga cada módulo (contrato del informe unificado). La matriz común
> fija **en qué capa** vive cada regla; la 05 no la redefine (`design.md` — D2, D19 y D23).

- [ ] 6.1 **[ahora]** Declarar en el catálogo la tabla de carga de `design.md` — D19 —por módulo: acciones en que es obligatorio u opcional y si el especialista lo recibe en el encargo— con los nombres de la matriz común (`artifacts`, `tdd`, `delegation`, `model-assignments`, `engram-opsx`) y los de esta ficha (`directed-code-reading`, `bounded-command-output`, `task-tool-selection`, `concise-agent-reporting`), según la tabla de `entrega-03-04.md` §3. Escribir la función pura que, dados acción, rol y toggles activos, devuelve los módulos, y verificar con tests que dos resoluciones iguales coinciden, que el orden de declaración no altera el resultado, que `catalog.Load()` rechaza un módulo sin acciones, que sin requisito de TDD no aparece `tdd` y que delegar trae `delegation` con `model-assignments` (`specs/task-context-loading`)
- [ ] 6.2 **[ahora]** Generar desde el catálogo la tabla acción → módulos en markdown, con la regla de módulo ausente o vacío y la de compactación debajo, por cliente y con su forma de ruta, y verificar con un golden que la generación es estable y que coincide con el texto de `entrega-03-04.md` §3. Es el texto que 03 y 04 ponen en el `SKILL.md`: así el agente no depende de leer un catálogo embebido
- [ ] 6.3 **[ahora]** Pasar a `internal/harness/config/assets/modules/bounded-command-output.md` el texto de `entrega-03-04.md` §2, conforme a `specs/bounded-command-output`. Verificar con un test que el archivo existe, no lleva frontmatter de skill, contiene las recetas de 3.8 byte a byte y las líneas de forma fija (`exit:`, `failure:`, `ref:`, `capture:`, `status: unverified`, `partial — unverified:`)
- [ ] 6.4 **[ahora]** Generar en `docs/` desde el catálogo la tabla de carga por acción, módulo y cliente, con los límites declarados (por ejemplo, `model-assignments` omitido en Codex hasta la 09), y verificar con un test que coincide con la generación
- [ ] 6.5 **[bloqueada: compositor común 03/04 con registro de módulos (hoy hay dos compositores; ver PR #8/#9 y `design.md` — D23)]** Con 03 y 04, registrar los cuatro módulos de esta ficha y la tabla de 6.2. Verificar con tests sobre lo compuesto que `modules/` contiene los cuatro archivos sin frontmatter en los dos clientes aunque no tengan toggle (hoy la 03 descarta un módulo sin toggle, `entrega-03-04.md` §5, R2), que la tabla del `SKILL.md` es byte a byte la de 6.2 en cada cliente, que un módulo de toggle inactivo no figura y que verify falla con un módulo vacío (R3)
- [ ] 6.6 **[bloqueada: compositor común 03/04 con registro de módulos (hoy hay dos compositores; ver PR #8/#9)]** Verificar sobre lo compuesto que la capa universal contiene los niveles de `governance` y las reglas universales de memoria de la matriz en sesión normal, con coordinador y en especialista, y que para Explore, Plan, un agente con `omitClaudeMd: true` y, mientras la 04 no confirme lo contrario, un subagente de Codex, el encargo repite las reglas obligatorias y `governance` (`design.md` — D21; en Claude ya medido: Explore y Plan no ven `CLAUDE.md`, `entrega-03-04.md` §5, C10)
- [ ] 6.7 **[bloqueada: compositor común 03/04 con registro de módulos (hoy hay dos compositores; ver PR #8/#9)]** Verificar con un test sobre lo compuesto que la referencia de invocación de cada cliente es la acordada (`/sdd-orchestrator` en Claude Code; `$sdd-orchestrator`, con `/skills` como alternativa, en Codex) y no contiene la sintaxis del otro cliente ni `claude --agent`
- [ ] 6.8 **[parcialmente desbloqueada: las superficies ya están declaradas en `codex-native-opsx/design.md:59-63` y la acción `verify` está en el texto de la 04; la paridad depende de su tarea 2.5]** Declarar la paridad por cliente y superficie y verificar que un módulo que un cliente no puede recibir —`delegation` en una superficie sin subagentes, `model-assignments` en Codex, los módulos de esta ficha en el modo legado de Claude (`design.md` — D19)— queda registrado como límite y no como entrega. Ya se puede declarar la tabla de superficies; la verificación espera la 2.5 de la 04

## 7. Respuestas del agente que no entierran el hallazgo

> Prioridad 2. La evaluación de conducta se hace en 8.10 con la rúbrica de 7.2.

- [ ] 7.1 **[ahora]** Escribir los fixtures de respuesta: un hallazgo informado con el recorrido primero y otro con el resultado primero, sobre el mismo caso, más una revisión sin hallazgos, un caso con varios hallazgos y uno cuyo contenido ya está en un artefacto; verificar que el par describe el mismo hallazgo con la misma procedencia
- [ ] 7.2 **[ahora]** Pasar a `internal/harness/config/assets/modules/concise-agent-reporting.md` el texto de `entrega-03-04.md` §2, conforme a `specs/concise-agent-reporting`, y escribir la rúbrica: primera línea con el resultado y su ruta y línea, segunda con la consecuencia o el próximo paso, sin repetir desarrollo, lista con procedencia antes de cualquier desarrollo, señalar en lugar de reproducir lo que ya está en un artefacto, y una línea para una revisión sin hallazgos. Verificar que la rúbrica distingue los dos fixtures del par y que el archivo no lleva frontmatter de skill
- [ ] 7.3 **[ahora]** Documentar en `docs/` el contrato de respuesta con el par de 7.1 y verificar que el ejemplo coincide con el fixture

## 8. Corridas con clientes reales

> **[bloqueada: Claude Code y Codex instalados, en Windows y en Linux (WSL o máquina Linux), con
> una persona que abra cada sesión y, para IDE, la extensión real]**. Hasta que el módulo se
> cargue desde la skill (6.5), el módulo se adjunta al encargo de forma explícita y la corrida lo
> registra. Cada corrida usa el manifiesto y los perfiles de 02 (`eval/baseline/manifest`,
> `eval/baseline/profile`) y registra cliente, versión, interfaz y sistema.

- [ ] 8.1 *(avance del 13/11, sin cumplir la aceptación: en Windows ya se midió que Claude CLI trunca un fallo largo y que `aggregated_output` de `codex exec --json` llega completo, con generadores propios y n = 1; `entrega-03-04.md` §5, C1–C6 y X1–X8)* Demostrar la cobertura de cada fuente candidata de D22 por celda: correr los fixtures 2.1 (fallo al medio) y 2.7 (solo stderr) en Claude Code y Codex, CLI e IDE, Windows y Linux, comparar cada fuente con la referencia del arnés usando 3.6 y registrar la celda como `probada` (con id de corrida), `truncada`, `ausente` o `no disponible`. Correr también 3.3 y 3.8 en Linux. Esta tarea decide qué fuente usa cada interfaz; hasta entonces ninguna captura del cliente cuenta como independiente
- [ ] 8.2 Correr los fixtures 2.1 a 2.8 con el agente real y el módulo `bounded-command-output`, y verificar con el evaluador de 2.9 que cada resumen se acepta, incluidos el fallo al medio, el de stderr y los dos de ruido, y que el de 2.8 se entrega "parcial — sin verificar"
- [ ] 8.3 Verificar con el agente real, en bash y en PowerShell, que la captura declarada sigue las recetas de 3.8 y que el resumen sale de leer el archivo y no de la vista que mostró el cliente
- [ ] 8.4 Verificar en ejecución directa y delegada que un resultado declarado por el agente figura *sin verificar*; que pasa a *verificado* solo con captura independiente de cobertura `probada` (8.1) o re-ejecución por otra parte; que un conjunto archivo-huella-resumen fabricado por el mismo agente y coherente sigue *sin verificar*; y que un fallo omitido lo lleva a *rechazado* (`design.md` — D20)
- [ ] 8.5 Correr el fixture 2.4 de punta a punta: ni el resumen ni ninguna copia de Active Stack contiene un valor ficticio sin redactar, la línea y su posición se conservan, y el reporte declara como límite la copia propia del cliente. Verificar que la retención y la redacción se mantienen cuando el original se pide desde otra acción
- [ ] 8.6 Correr los fixtures de lectura de 4.6 con el módulo `directed-code-reading` y verificar los registros esperados, incluida la cuenta de lecturas por la quinta condición
- [ ] 8.7 Correr los fixtures de herramientas de 5.4 con el módulo `task-tool-selection` y verificar los registros esperados y, por cliente, el nivel efectivo de cada control declarado en 5.2
- [ ] 8.8 **[además bloqueada: 6.5 y 6.8]** *(avance del 13/11 en Claude: módulo ausente detiene e informa; módulo vacío, conducta inconsistente; tras `/compact` el cuerpo de la skill se pierde; `entrega-03-04.md` §5, C11–C13)* Verificar con el módulo cargado desde la skill que `verify` sin `delegation` informa qué falta y no se ejecuta, que un módulo vacío se trata como ausente, que tras compactar el coordinador relee los módulos de la acción, que `apply` sin él avisa y resuelve directo, que las tareas que no dependen del módulo se completan y el reporte las separa, que la acción se retoma al quedar disponible el módulo o una superficie compatible, y que una superficie sin subagentes deja `verify` detenida como límite declarado
- [ ] 8.9 **[además bloqueada: 6.5 y la migración de 03 y 04]** Comparar una sesión normal con una que invoque al orquestador en cada cliente y superficie soportada, en instalación limpia y después de actualizar una instalación anterior, y comprobar qué módulos recibe cada una
- [ ] 8.10 Aplicar la rúbrica de 7.2 a respuestas reales sobre los casos de 7.1 y verificar que distingue resultado primero de recorrido primero

## 9. Comparación con y sin cada integración y cada módulo, con el protocolo de 02

> **[bloqueada: clientes reales, como el grupo 8]**. El protocolo de 02 ya está en `main`
> (`eval/baseline/`); las comparaciones de módulos esperan además a 6.5.

- [ ] 9.1 Preparar la corrida de una misma acción con y sin cada integración opcional, con la misma definición de aceptación, y verificar que es byte a byte idéntica entre ambas corridas
- [ ] 9.2 Registrar el contexto efectivo de cada corrida con el manifiesto y el resumen de 02 y verificar que la diferencia observada queda atribuida a la integración y no a un cambio de criterio
- [ ] 9.3 **[además bloqueada: 6.5]** Preparar la corrida de una misma acción con y sin cada módulo (`tdd`, `delegation`, `model-assignments`, `engram-opsx` y los cuatro de esta ficha: `directed-code-reading`, `bounded-command-output`, `task-tool-selection`, `concise-agent-reporting`), con la misma aceptación, y verificar que la diferencia de contexto efectivo y de aceptación queda atribuida al módulo (`specs/task-context-loading` — *Comparabilidad con y sin cada módulo*)
- [ ] 9.4 **[además bloqueada: 6.5]** Medir con 9.3 el costo de las reglas universales de memoria que fija la matriz (filas 20 y 22) frente a la regla corta que proponía la 05, y llevar el resultado a la matriz común como insumo, sin cambiarla desde esta ficha

## Resumen de bloqueos

| Bloqueo | Tareas |
|---|---|
| Compositor común 03/04 con registro de módulos (hoy hay dos compositores; ver PR #8/#9) | 6.5, 6.6, 6.7 |
| 04: tarea 2.5 (`verify` y evidencia por acción); las superficies ya están declaradas | 6.8 (parcial) |
| Clientes reales en Windows y Linux, con una persona que corra la sesión | 8.1 a 8.10, 9.1 a 9.4 |
| Además, módulos cargados desde la skill (6.5) | 8.8, 8.9, 9.3, 9.4 |

Implementables ahora: 1.1–1.8, 2.1–2.11, 3.1–3.8, 4.1–4.7, 5.1–5.5, 6.1–6.4 y 7.1–7.3.

## Workflow follow-up

> Cada ítem lleva fecha, estado, destinatarios y artefacto. Los acuerdos que el informe unificado
> (11/11) y la matriz común (12/11) ya registran figuran como **acordados**; el resto sigue
> **pendiente de enviar** o **pendiente de respuesta**.

### Acordado

- **2026-11-11 · acordado** · los siete · informe unificado — *Contratos compartidos propuestos*. La 05 mantiene captura, acceso, política de datos y la lista sensible; la 06, procedencia, huellas y estados; la 07, el sobre de retorno. La 05 define cuándo se carga cada módulo. Incorporado en `design.md` — D8, D9, D14, D19 y D23.
- **2026-11-11 · acordado** · 05 y 06 · informe unificado — *05/06*. Un resultado declarado por el agente queda sin verificar hasta corroboración independiente de la ejecución y del estado del código. Incorporado en `design.md` — D20 y `specs/bounded-command-output`.
- **2026-11-11 · acordado** · 01, 03, 04, 08 · informe unificado — *Plataformas*. Windows y Linux como mínimo; matriz distribuida, declarada y probada. Incorporado en `design.md` — D22 y tarea 3.7.
- **2026-11-11 · acordado** · 03 y 04 · informe unificado — *03/04*. Invocación explícita `/sdd-orchestrator` en Claude y `$sdd-orchestrator` en Codex. Incorporado en `design.md` — D21 y tarea 6.7.
- **2026-11-11 · acordado** · los siete · informe unificado — *Responsables*. Giuliano lleva 02 y Matías lleva 04.
- **2026-11-12 · acordado** · Juani (03), Matías (04) · matriz común (`origin/claude-native-opsx:openspec/matriz-reglas-opsx.md`). Matriz única por cliente, interfaz, rol y acción; `governance` en capa universal (fila 18), como proponía la 05; módulos como archivos de la skill. Incorporado en `design.md` — D2, D19 y *Ubicación de los módulos*. La 05 adopta las filas de memoria (8, 20 a 24) y lleva su medición a la tarea 9.4.

### Hecho

- **2026-11-13 · hecho** · Emiliano (05) · `entrega-03-04.md` §5 y §6. Pruebas reales con Claude Code 2.1.319 y Codex 0.167.1 en Windows y sobre las PR #8 y #9. Incorporadas en `design.md` — D3, D8, D18, D19, D21, D22 y D23, y en las specs (formas fijas, módulo vacío, compactación, modo legado, celdas de cobertura). Ninguna tarea se cierra con ellas: avanzan 3.6, 6.6, 6.8, 8.1 y 8.8.

### Pendiente

- **2026-11-13 · listo para enviar** · Juani (03), Matías (04, según el informe unificado) · `entrega-03-04.md` y `design.md` — D23. Texto entregado: los cuatro módulos, la tabla de carga con la columna de especialista, la regla de módulo ausente o vacío, la de compactación y las rutas por cliente. Responde la fila 27 de la matriz, la tarea 2.3 de la 03 y la 1.4 de la 04 (`SKILL.md:20`). Pedir el compositor común con registro de módulos en lugar de los dos actuales, que la tabla del `SKILL.md` sea la de la entrega (con `explore` y `verify`), que verify falle con un módulo vacío en la 03, y que el aviso de modo legado de la 03 nombre que los módulos de contexto no se entregan.
- **2026-11-13 · pendiente de enviar** · Gianfranco (06) · `design.md` — D9, D14 y D20; *Coordinación pendiente* 05 → 06. Proponer ubicación, acceso y vencimiento de la copia redactada, la lista sensible, el registro de captura con su campo de fuente, y que su contrato de evidencia se cargue con la misma condición que `bounded-command-output` (`apply` y `verify`, directa o delegada). La 06 confirma o contrapropone; la ubicación queda propuesta hasta entonces.
- **2026-11-13 · pendiente de enviar** · Giuliano (02) · `design.md` — D9 y D22. Responder el `(TBD · se fija junto con 05)` del plazo de borrado de su D6 con la retención de D9; proponer que `eval/baseline/record` use la lista de D14 en lugar de su lista mínima; avisar que `eval/taskcontext/` reutiliza su manifiesto y sus perfiles. Pedirle además las métricas de lectura dirigida (lecturas, bytes leídos, herramientas por tarea) separadas por terminal e IDE, que su `trace` ya registra.
- **2026-11-13 · pendiente de enviar** · Neyen (07) · `design.md` — D3, D19 y D20. Su S1 ya cita las 300 líneas; avisar que la condición es 300 líneas **o** 10.000 caracteres, que el resultado de un worker queda *sin verificar* salvo captura independiente o re-ejecución, y que `verify` se detiene en una superficie sin subagentes.
- **2026-11-13 · pendiente de enviar** · Juan Cruz (01), Gianfranco (08) · `design.md` — D14 y D22. Pedir que una de las dos lleve a Codex un piso equivalente al de Claude (`permissions.<name>.filesystem` con `"deny"`, `shell_environment_policy`). Avisar a la 08 que un hook `PreToolUse` que envuelva el comando sería un capturador del cliente candidato para Claude, y que su adopción es la decisión de arquitectura de `CHANGES.md:49`, no de esta ficha.
- **2026-11-13 · pendiente de enviar** · Matías (04, según el informe unificado; `CHANGES.md` nombra a Giuliano) · `design.md` — D21 y D22. Avisar lo medido en Codex 0.167.1 sobre Windows (`entrega-03-04.md` §5): `aggregated_output` de `codex exec --json` llega completo (X7), `$sdd-orchestrator` inyecta el cuerpo (X10) y el pedido sin `$` no invoca la skill (X11). Pedir que su sonda P2 cubra lo no probado: si un subagente de Codex lee `AGENTS.md`, la lectura real de módulos y el módulo ausente o vacío, hoy bloqueados por la política de shell de Windows (X12), y Linux.
- **2026-11-13 · pendiente de enviar** · los siete · rama `propose/opsx-task-context`. Compartir esta revisión con el resumen de qué quedó cubierto, qué se corrigió y qué sigue pendiente.
- Archivar el change una vez satisfechos los requisitos de revisión del proyecto.
