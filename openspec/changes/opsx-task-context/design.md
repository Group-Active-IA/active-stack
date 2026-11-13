# Design

## Context

Ver `proposal.md` — Why e Impact para la motivación y la medición de partida.

Tres restricciones del estado actual condicionan todo el diseño:

1. **El mecanismo de composición ya existe y es puro.** `internal/harness/config/compose.go` ensambla el bloque de forma determinista: dos toggles sustractivos (`delegation`, `model-routing`, que quitan secciones marcadas del archivo base) y tres aditivos (`governance`, `engram`, `tdd`, que se concatenan en orden fijo). La función no hace I/O y el mismo input produce siempre el mismo output.
2. **El selector no existe en `main`.** Claude y Codex componen en la rama `default` de `internal/harness/config/install.go:70`, que invoca `Compose(h.Toggles, …)` con los toggles literales del catálogo; OpenCode pasa por la rama `ConfigDeliveryPrimaryAgent` (`install.go:53`). No hay bandera de CLI, pantalla de TUI ni campo de configuración que los filtre (`CLAUDE.md:82`). Toda instalación produce el bloque máximo.
3. **La decisión ocurre en el momento equivocado.** La composición se congela al instalar y se inyecta con markers en `CLAUDE.md`/`AGENTS.md`. En ese momento no se conoce la acción activa. La entrega por skill que parte el bloque en capa universal, cuerpo de la skill y módulos existe hoy **dos veces y fuera de `main`**: la 03 (PR #8, change archivado en su rama) con `ComposeLayers`, que parte el asset con anotaciones `<!-- as:module/<id> -->`, y la 04 (PR #9, borrador) con `ComposeCodexPackage`, que lo parte con `splitCodexAsset` y escribe la tabla de acciones a mano. Las dos agregan la constante `ConfigDeliveryManagedSkill` con semánticas distintas y no se integran sin conflicto (documentado y medido: `entrega-03-04.md` §5, R5 y R6). Lo que esta ficha agrega es **cuándo** se lee cada módulo, y propone el compositor común que lo resuelve una sola vez (D23).

Fuentes de esta revisión (2026-11-13): el informe unificado de revisión OPSX del 11/11 (`informe-unificado-propuestas-opsx.md`), la matriz común de reglas que publicó la 03 el 12/11 (`origin/claude-native-opsx:openspec/matriz-reglas-opsx.md`, en adelante *la matriz común*) y las pruebas reales del 13/11 con Claude Code 2.1.319 y Codex 0.167.1 en Windows y sobre las PR #8 y #9 (`entrega-03-04.md` §5 y §6, citadas por id: C1…, X1…, R1…).

## Goals / Non-Goals

**Goals:**

- Especificar la frontera entre lo que es obligatorio en toda sesión y lo que se resuelve por acción, de forma que la frontera sea verificable y no una recomendación.
- Especificar un contrato de resumen de salida que se pueda probar con fixtures antes de que exista implementación, con roles separados para capturar, redactar y resumir.
- Demostrar, por cliente, interfaz y sistema, qué fuente independiente captura la salida antes del truncamiento.
- Dejar el inventario de partida medido y reproducible, para que 02 pueda comparar contra él.

**Non-Goals:**

- No se diseña ningún subcomando nuevo del binario. Los *Límites compartidos* del roadmap (`CHANGES.md:49`) dicen que *"un ledger firmado, un runtime nuevo o una integración permanente del juez requieren una decisión de arquitectura separada"* y que *"el arnés de evaluación queda fuera del binario"*. La captura independiente de esta ficha vive en el arnés (`eval/taskcontext/`) o es una fuente que el cliente ya ofrece (D8, D22); un capturador permanente instalado en la sesión del usuario es esa decisión separada.
- No se reestructura `sdd-orchestrator.md`. La partición en capas es de 03 y 04 (matriz común). Esta ficha **redacta el texto de sus propios módulos** y la tabla de carga, y los entrega (D23).
- No se define ninguna métrica de consumo de modelo. Eso es 02. Los números de este change son de tamaño de texto.
- No se define qué evidencia alcanza para aceptar un cambio: eso es de 06. Esta ficha define captura, acceso y política de datos.

## Estado de las decisiones

Cada decisión lleva su estado:

| Estado | Qué significa |
|---|---|
| **acordado** | lo registra el informe unificado (11/11), la matriz común (12/11) o el roadmap |
| **propuesto a <ficha>** | toca alcance de otra ficha y espera su confirmación; queda en *Coordinación pendiente* |
| **de la 05** | está dentro del alcance de esta ficha y no cambia el de otra |

Cada afirmación sobre un cliente o una rama lleva además su respaldo: **medido** (prueba real, con su id en `entrega-03-04.md` §5; n = 1), **documentado** (fuente citada) o **no probado** (motivo en `entrega-03-04.md` §6).

## Decisions

### D1 — La selección ocurre al leer la skill, no al componer

**Acordado** (matriz común, *Capas*). La composición sigue siendo pura y determinista al instalar: produce la capa universal, el cuerpo de la skill y los módulos como archivos (`ComposeLayers` de la 03). La acción activa **no** es una entrada de `Compose`: el coordinador la conoce en la sesión y lee los módulos que la tabla acción → módulos del `SKILL.md` indica para ella. Esa tabla la genera el catálogo (D19), así que la resolución sigue siendo reproducible y comparable entre corridas.

*Alternativa descartada: exponer los cinco toggles como banderas de instalación.* Resuelve otro problema —permite medir si un componente vale su costo, pero sigue congelando la elección al instalar— y el proyecto decidió no exponer los toggles en la TUI ni en la GUI (`CLAUDE.md:82`). La comparación con y sin cada módulo se arma con el arnés fuera del binario (tarea 9.3).

*Alternativa descartada: dejar el bloque completo y confiar en la caché del proveedor.* El proveedor reutiliza bajo condiciones y el ensamblado lo hace el host. La estabilidad de prefijos y caché es materia de 09.

### D2 — El reparto por capa lo fija la matriz común; la 05 aporta el criterio

**Acordado** (matriz común, *Decisiones tomadas al publicar*). El criterio de la 05 es por consecuencia, no por tamaño: una instrucción va a la capa universal si su ausencia puede producir daño o invalidar el trabajo con independencia de la acción; va a módulo si su pertinencia depende de la acción. La matriz común aplica ese criterio fila por fila y es la única fuente del reparto: la 05 no redefine filas.

| Fragmento | Capa en la matriz común | Relación con la propuesta de la 05 |
|---|---|---|
| `governance` — niveles (`governance.md:1-13`) | universal (fila 18) | converge: la 05 lo proponía en el núcleo |
| `governance` — aplicación a delegación (`:14-fin`) | coordinador (fila 19) | compatible: la 05 no lo separaba |
| `engram` — encabezado, guardado proactivo y búsqueda | universal (filas 8, 20, 22) | la 05 proponía una regla corta; la matriz deja universal más texto |
| `engram` — wiring OPSX y git sync | módulo `engram-opsx` (filas 21, 24) | converge |
| `engram` — cierre de sesión y compactación | coordinador (fila 23) | compatible |
| delegación | módulo `delegation` (fila 10) | converge |
| model routing | módulo `model-assignments` (fila 17); en Codex se omite hasta la 09 | converge; cambia el nombre |
| `tdd` | módulo `tdd` (filas 25, 26) | converge |
| ciclo de artefactos | módulo `artifacts` (fila 14) | nuevo de la matriz; la 05 fija cuándo se carga (D19) |
| módulos de esta ficha | módulos nuevos (fila 27) | los entrega la 05 (D23) |
| `design-mockups` | sección propia (fila 30) | fuera de alcance |

**Por qué se adopta la fila de memoria aunque difiera.** `engram` son 6.093 B, el 30,4 % del bloque de Claude (`proposal.md` — Impact). La matriz deja en la capa universal el guardado y la búsqueda, que son más que la regla corta que proponía la 05. La 05 no reabre la fila: lleva la medición con y sin ese texto a la tarea 9.4 y entrega el resultado a la matriz como insumo.

*Por qué no un presupuesto numérico:* el informe ejecutivo marcó los umbrales de 40 líneas y 1.500 tokens como presupuestos de diseño, no como límites validados. Siguen siendo hipótesis a medir con 02.

### D3 — El resumen de salida se especifica por lo que debe conservar, no por cuánto debe recortar

**De la 05.** Un límite de tamaño es el modo de fallo que la ficha nombra: una salida extensa puede ocultar un fallo. Por eso el contrato se escribe en términos de retención obligatoria (errores, fallos, código de salida, redacción de datos sensibles) y de recuperabilidad del original, y no en términos de un tope de líneas.

El único número del contrato es **cuándo** se resume, no **cuánto**: una salida de más de 300 líneas **o** de más de 10.000 caracteres, lo que ocurra primero, se entrega como resumen más referencia (D19). Es una condición de entrada, no un tope de recorte.

**Por qué dos unidades.** Los clientes no miden en líneas. Lo que recibe el modelo, por cliente:

| | Claude Code 2.1.319 | Codex 0.167.1 |
|---|---|---|
| Comando fallido con salida larga | **medido** (C1–C5): unos 10.040 caracteres de cabeza y cola con la marca `[N characters truncated]`, sin ruta al resto. Bash conserva primero unos 30.000 caracteres, así que la "cola" **no es el final real**: de 3.000 líneas, lo visible termina en la 912, un fallo en la última línea no aparece y la marca dice 20.012 cuando se omitieron ~90.000. Con 750 líneas (~25.000 caracteres) se pierden las líneas ~155 a 599. Solo stderr: igual | **medido** (X1–X6): cabeza y cola, mitad y mitad aproximada, con `Warning: truncated output (original token count: N)` y `…N tokens truncated…`; el final real se conserva. Igual con exit 0 o 1 y con solo stderr. El tamaño lo elige el modelo por llamada (`max_output_tokens`) con un tope externo de ~10.000 tokens (~40 KB); con 750 líneas (32.297 caracteres) ya se pierde el medio. El modelo no avisa que falta el medio. `-c tool_output_token_limit` en 2.000 y 50.000 no tuvo efecto observable (no concluyente) |
| Comando exitoso con salida larga | **medido** (C6): vista previa de ~2 KB más la ruta a un archivo del cliente con la salida completa (3.000 líneas, 100.897 B). **Documentado**: *"with no file path"* para el fallido (https://code.claude.com/docs/en/tools-reference) | igual que el fallido |
| Original completo disponible | solo en éxito, y es copia del cliente sin redactar (D9) | **medido** (X7): `aggregated_output` de `codex exec --json` trae la salida completa (D22); el historial de la sesión guarda la recortada (X8) |

Consecuencias. Los 10.000 caracteres se confirman: es el tamaño que Claude deja de un comando fallido y queda debajo del recorte de Codex. Para el caso que importa —un comando que falla con salida larga— **Claude no ofrece en la sesión ninguna fuente completa**, así que la captura la hace el arnés o el agente con la receta de D8. Y como ni Claude (marca que subestima) ni Codex (sin aviso) le dicen al modelo cuánto perdió, el módulo prohíbe resumir desde la vista del cliente. El par se sigue calibrando con las salidas reales del ciclo (tarea 1.8). El umbral de lectura de archivos (D11) y el de salidas son **dos reglas distintas**, aunque compartan las 300 líneas.

**Qué es un fallo.** El código de salida distinto de cero es siempre un fallo. Además, el catálogo declara patrones de fallo por herramienta, empezando por las de este repositorio (`go test`: `--- FAIL`, `FAIL\t`, `panic:`; `go build` y `go vet`: `<archivo>:<línea>:<columna>:`), y cada fixture declara su conjunto de fallos esperado (tarea 2.10). Una salida de una herramienta sin patrones declarados se evalúa solo por código de salida, y el resumen lo dice. Así la comparación de fallos no depende del juicio del modelo.

*Alternativa descartada: recortar por cabeza y cola.* Falla justamente en el caso del fallo central, y es lo que el cliente ya hace con una salida fallida.

### D4 — Los fixtures son el entregable verificable

**De la 05.** Los fixtures son comandos deterministas cuya salida conoce el arnés. Se diseñan con el fallo en tres posiciones (inicio, medio, final), con fallos dispersos, con una salida verbosa sin fallos como control negativo, con ruido que imita un error, con un fallo real entre ruido, con un fallo **solo por stderr**, con **datos sensibles ficticios** —un valor inventado por cada patrón de D14, marcado como ficticio— y con un comando cuyo original **nunca se capturó**. Cada fixture declara su salida, su código de salida, su resumen esperado y el fallo que debe sobrevivir.

Los fixtures sirven para dos cosas: medir **la fuente de captura** (si lo que entrega coincide con la salida conocida, D22) y medir **el resumen** (si conserva lo que el contrato exige, tarea 2.9). Los dos se corren contra el cliente real en el grupo 8.

### D5 — Procedencia obligatoria en lugar de umbral de lectura como regla principal

**De la 05.** La regla que más evita trabajo inválido no es "leé menos" sino "toda afirmación sobre el código indica ruta y línea o símbolo". El umbral de lectura completa es un apoyo; la procedencia es lo que hace falsable cada hallazgo.

### D6 — La llamada innecesaria se registra, no se bloquea

**De la 05.** Una relectura o una invocación repetida se marca como llamada innecesaria y queda registrada; nada la impide. Bloquearla exigiría decidir con certeza que el estado observado no cambió. Registrar convierte el problema en una medida que 02 puede comparar. La excepción es el sondeo de una integración ya informada como no disponible: ahí el requisito es no volver a invocarla.

### D7 — El inventario de verbosidad cubre todas las fuentes, no solo el orquestador

**De la 05.** El inventario cubre instrucciones (tareas 1.1 a 1.7) y salidas de comandos (tarea 1.8), y se publica como una tabla única ordenada por tamaño. Es investigación, no comportamiento: vive en `tasks.md` y en `proposal.md` — Impact. Se propone a la 02 reutilizarlo en lugar de duplicarlo.

### D8 — Tres roles separados: captura independiente, redacción determinista y resumen del agente

**Acordado** (informe unificado: *05 · Contexto*, *05/06 · Retención*, contrato *Salidas y recibos*). El informe pide documentar por separado quién captura stdout y stderr, quién redacta y quién genera el resumen, y que la captura la haga una fuente independiente antes del truncamiento. La versión anterior de esta decisión ponía los tres roles en el agente; eso hace que la misma parte que reporta produzca el original, y por eso no alcanza para verificar nada.

| Rol | En el arnés de evaluación | En una sesión real, hoy |
|---|---|---|
| **Captura** | el capturador del arnés ejecuta el comando del fixture y recibe stdout y stderr en un mismo flujo, con el código de salida (tarea 3.3). Es independiente por construcción y su salida es la referencia | una fuente del cliente solo cuando D22 la declara `probada` para esa celda; si no, la **captura declarada por el agente**, que no es independiente |
| **Redacción** | código del arnés, en memoria, con la lista de D14, antes de escribir a disco; no queda crudo | quien captura: el código de la fuente, o el agente si la captura es declarada, y entonces la redacción también es declarada |
| **Resumen** | el agente, desde la copia capturada y con lectura dirigida | el agente |

El registro de captura (D9) dice quién cumplió cada rol, y eso fija el estado máximo que puede alcanzar el resultado (D20).

**Fuentes candidatas del cliente.** Documentadas, con su cobertura por demostrar en D22: el hook `PostToolUse` de Claude Code, cuyo `tool_response` de Bash trae `stdout`, `stderr`, `interrupted` e `isImage`, y `PostToolUseFailure`, cuyo `error` empieza con `Exit code N` y se *"middle-truncates"* con una marca `... [N characters truncated] ...` (https://code.claude.com/docs/en/hooks); en Codex App Server, `item/commandExecution/outputDelta`, que *"streams stdout/stderr for a command"*, y el item final con `aggregatedOutput?` y `exitCode?` (https://learn.chatgpt.com/docs/app-server); y en `codex exec --json`, el item `command_execution` con `aggregated_output` y `exit_code`. La documentación no dice en ningún caso si la salida llega completa. **Medido** (X7): en Codex CLI 0.167.1 sobre Windows, `aggregated_output` trae la salida completa —129.047 caracteres de 3.000 líneas, con el fallo central y el final— en stdout, en solo stderr y con exit 0 o 1. Es la primera fuente del cliente con cobertura medida; vale para lo que el arnés conduce con `codex exec --json`, no para una sesión interactiva. En Claude, los hooks siguen **no probados**; lo que recibe el modelo (y por eso `stream-json`) está recortado (C1). Ninguna fuente cuenta como independiente en una celda hasta que la tarea 8.1 la corra con los fixtures del arnés (D22).

**Captura declarada por el agente.** Sigue siendo útil: recupera el fallo central que el cliente recorta aunque el resultado quede *sin verificar*. El módulo `bounded-command-output` la prescribe **siempre** para los comandos de verificación (tests, build, vet, lint) y para los que el inventario de la tarea 1.8 declara verbosos, sin depender de adivinar el tamaño antes de correr. Si un comando no capturado resultó verboso y no tiene efectos laterales, se vuelve a correr con captura; si los tiene, el resumen es "parcial — sin verificar". Recetas:

| Shell | Receta | Por qué |
|---|---|---|
| bash / sh | `cmd > "$f" 2>&1; echo "exit=$?" > "$f.exit"` | sin tubería, `$?` es el del comando |
| PowerShell 5.1 y 7 | `cmd /c "<comando> > ""$f"" 2>&1"; "exit=$LASTEXITCODE" \| Set-Content -Encoding ascii "$f.exit"` | `*>` en PowerShell 5.1 escribe UTF-16 LE con BOM y envuelve stderr como `NativeCommandError` (reproducido por la auditoría del 10/11); `cmd /c` deja los bytes del comando |

La tarea 3.8 prueba las dos recetas en cada shell disponible.

**La copia propia del cliente no es el original de Active Stack**: no está redactada ni tiene huella ni vencimiento. El resumen nunca la cita.

*Alternativa descartada: un envoltorio ejecutable instalado en la sesión.* Es el runtime nuevo que `CHANGES.md:49` reserva. Un hook `PreToolUse` que reescriba el comando para capturarlo sería un capturador del cliente bajo control del instalador; es candidato para esa decisión separada, con la 08 como dueña de los hooks, y no se adopta acá.

*Alternativa descartada: confiar en la captura del agente si declara productor.* La declaración hace atribuible un incumplimiento; no lo detecta.

### D9 — Copia redactada fuera del repo, registro de captura y vencimiento

**Propuesto a 06** (informe unificado, *05/06 · Retención*: *"05 y 06 deben acordar ubicación, acceso y vencimiento de la copia redactada"*). Lo que ya está acordado: una fuente independiente captura antes del truncamiento; por defecto se persiste solo la copia redactada, con la huella de esa copia y el registro de captura; si verificar un requisito exige un detalle descartado, el resultado queda sin verificar hasta conseguir evidencia independiente suficiente.

**Qué se persiste.** Solo la copia redactada. En el arnés no existe crudo en disco: el capturador redacta en memoria (D8). En la captura declarada por el agente el primer archivo es crudo; el módulo manda redactarlo a un segundo archivo y borrar el crudo antes de entregar el resumen, y eso queda declarado, no garantizado.

**Referencia.** Todo resumen lleva `{comando, id de corrida, ruta, huella sha256 de la copia redactada, vencimiento}`. La huella nombra el algoritmo (`sha256:<hex>`) y es de bytes, sin normalizar finales de línea.

**Registro de captura**, junto a la copia: fuente de captura (`arnés`, `cliente:<fuente>` o `agente`), redactor, cliente, interfaz, sistema, versión, código de salida, cobertura antes del truncamiento (`demostrada` según D22, o `no demostrada`) y conteo de redacciones por categoría. El código de salida vive acá, escrito por quien captura, y no en un archivo aparte del agente.

**Ubicación.** `<os.UserCacheDir>/active-stack/runs/<change>/<id de corrida>/`, fuera del repositorio, y `_sin-change` en lugar del nombre cuando la acción no tiene change activo (por ejemplo, `explore`). Desde `1b4639b` `openspec/` se versiona y `openspec archive` mueve la carpeta del change, así que una salida guardada dentro del change podría commitearse; fuera del repo no puede. En una superficie sin disco persistente del usuario (Codex en la nube) no hay ubicación: la celda figura `no disponible` en D22 y el resumen es "parcial — sin verificar".

**Retención y quién borra.** Hasta que el change se archive y siete días más; para `_sin-change`, siete días desde la captura. Borra quien escribió: el arnés purga los vencidos al iniciar cada corrida (tarea 3.5). En una sesión real sin capturador independiente nadie borra, porque una instrucción no vence archivos: ahí la retención es una **recomendación declarada** y la referencia igual informa el vencimiento. Los siete días son un **supuesto**; se recalibran con el primer uso real. Un original **vencido** y uno **nunca capturado** son distintos: el primero tiene referencia completa; el segundo no tiene ruta ni huella.

**Acceso.** El del usuario local con los permisos de su directorio de caché. La copia no se publica ni se envía.

**Límite declarado: el cliente guarda sus propias copias** sin redactar —la transcripción de Claude Code en `~/.claude/projects/` y el archivo de trabajo (https://code.claude.com/docs/en/tools-reference); el historial de Codex—. La garantía se acota a **las copias que administra Active Stack**. La protección en origen es no leer el secreto: el piso de Claude (`Read(.env)`, `internal/harness/config/permissions/overlays.go:16-28`) y su equivalente nativo de Codex, que hoy el instalador no escribe (D14).

**Diferencias con la 06** (su rama `opsx-verifiable-evidence` no cambió desde `f3f005e`). Coincidimos en no versionar las salidas completas (`opsx-verifiable-evidence/design.md:78`).

| | 06 (`opsx-verifiable-evidence/design.md`) | 05 |
|---|---|---|
| cuándo se redacta | antes de compartir (`:79`) | antes de persistir |
| huellas | dos (`:79`), con `git hash-object` | una, sha256 de la copia redactada, con el algoritmo nombrado |
| dónde vive | `openspec/changes/<change>/evidence/` (`:73`, `:78`) | `<os.UserCacheDir>/active-stack/runs/…`, fuera del repo |
| retención | `(TBD)` (`:80`) | hasta archivar y siete días más |
| referencia | ruta, huella y si fue truncada (`:46`) | cinco campos más el registro de captura |

La 02 dejó el plazo de borrado de sus trazas originales como `(TBD · se fija junto con 05)` en su D6; esta decisión le da el valor.

*Alternativa descartada: conservar el crudo con dos huellas para poder comparar la redacción.* La redacción del arnés es código determinista y se prueba con el fixture de datos ficticios (tarea 3.3), así que no hace falta conservar secretos para auditarla.

### D10 — Las clases de tarea son las cinco acciones del ciclo OPSX

**Acordado** (la matriz común usa las acciones OPSX en su tabla acción → módulos). El eje es `explore`, `propose`, `apply`, `verify` y `archive` (`CLAUDE.md:26`, `ARCHITECTURE.md:6`, `ARCHITECTURE.md:164`). Las cinco tareas de medición de la ficha 02 son **cargas de trabajo** que se ejecutan dentro de esas acciones, no clases.

*Alternativa descartada: un eje propio por tipo de trabajo, o un sub-eje dentro de cada acción.* Agregaría otro vocabulario sin cambiar la tabla de herramientas (D18) ni las señales de delegación de la 07.

### D11 — El umbral de lectura completa es 300 líneas, medido sobre este repositorio

**De la 05.** **En líneas**, porque es la unidad en la que el agente pide un rango y en la que se expresa la procedencia. **300**, porque sale de medir: los 130 archivos `.go` versionados bajo `internal/` y `cmd/` que no terminan en `_test.go`, en `1b4639b`, con `wc -l`. Percentiles por rango más cercano:

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

A 400 la regla toca solo 6 archivos; a 200 fuerza lectura acotada donde leer entero sale más barato. 300 queda entre el p85 y el p92 y reserva la lectura acotada para los 14 archivos grandes.

En Claude Code una lectura devuelve como máximo 25.000 tokens y Read acepta `offset`/`limit` (https://code.claude.com/docs/en/tools-reference). En Codex no hay herramienta de lectura documentada con límite propio: es **no verificable** con la documentación y lo comprueba la tarea 8.6. El valor vive en el catálogo; la tarea 4.2 falla si 300 sale del intervalo p85–p92 y entonces se elige el nuevo valor con el mismo criterio.

### D12 — Esta ficha no fija ni propone un tope para los cuerpos de skill

**De la 05.** El inventario de las tareas 1.5 a 1.7 publica el tamaño de cada cuerpo de skill. Fijar un tope es política del catálogo para todas las fichas; la 05 entrega la medición y lo declara fuera de su alcance.

### D13 — La señal de pertinencia son el tipo de coincidencia y su cantidad

**De la 05.** La orientación a nivel repositorio devuelve rutas candidatas con **qué tipo de coincidencia** tienen y **cuántas**:

| Tipo | Qué es | Peso |
|---|---|---|
| definición | el símbolo se define en ese archivo | alto |
| referencia | el símbolo se usa en ese archivo | medio |
| texto | coincidencia literal sin estructura | bajo |

Una definición pesa más que cualquier cantidad de referencias, y una referencia más que cualquier cantidad de coincidencias de texto. Dentro del mismo tipo, ordena la cantidad. Tres categorías se explican en una línea y se obtienen sin leer el contenido de los archivos descartados.

### D14 — Una sola lista de patrones de valor sensibles, declarada en el catálogo y con dueña la 05

**Acordado** (informe unificado, contrato *Salidas y recibos*: *"05 mantiene captura, acceso y política de datos"*; 02 D10: *"05 mantiene la lista sensible"*). La redacción no usa las clases del piso de permisos: el piso bloquea **rutas** (`overlays.go:16-28`) y solo en Claude (`composeClaudeOverlay`, `overlays.go:73`); en Codex el instalador escribe únicamente `approval_policy` y `sandbox_mode` (`permissions/install.go:114-134`). Codex sí tiene un mecanismo nativo que el instalador no usa: `permissions.<name>.filesystem` con `"deny"` por patrón y `shell_environment_policy` (https://learn.chatgpt.com/docs/config-file/config-reference), con estado beta **no verificable**.

Contenido inicial de la lista de **patrones de valor**:

- `.env` y `.env.*`: el archivo y los valores de sus asignaciones
- tokens y API keys
- credenciales de nube
- claves SSH y `*.pem`
- valores de variables de entorno
- rutas del home del usuario
- identificadores de cuenta, correo y UUID de cuenta (los que ya redacta `eval/baseline/record`)

La usan 02, 06 y 10. Un test verifica que todo lo que bloquea el piso de Claude está cubierto (tarea 3.1). Llevar un piso equivalente a Codex se le pide a la 01 o a la 08.

### D15 — La lista de condiciones para leer un archivo completo es cerrada

**De la 05.** Son cinco:

1. el archivo no supera el umbral de D11
2. la tarea va a reescribir el archivo entero
3. el archivo es el artefacto que la acción activa produce o edita
4. una lectura acotada previa sobre ese archivo devolvió vacío
5. los rangos no alcanzan, y el motivo queda declarado en el reporte con forma fija: `full read: <path> — <reason>` (las formas fijas van en inglés porque viven en el texto del módulo; están en `entrega-03-04.md` §2)

Cerrada, el registro de lectura no justificada es decidible. La quinta es la excepción justificada para archivos donde ningún rango sirve; el arnés cuenta cuántas lecturas entran por ella y con qué motivos, como insumo para recalibrar D11. Una lectura por símbolo devuelve el símbolo completo y, si está anidado, la **firma** de la declaración que lo contiene.

### D16 — La respuesta del agente es una capacidad aparte, no parte del resumen de comandos

**De la 05.** La ficha pide localizar *"instrucciones largas o resultados verbosos"* (`CHANGES.md:162`), y hay dos clases de resultado verboso:

| | Salida de comando | Respuesta del agente |
|---|---|---|
| quién la produce | una herramienta | el modelo |
| hay un original completo | **sí**, y se conserva | **no**, nunca existió |
| cómo se accede al detalle | referencia a la copia guardada | se produce a pedido |
| qué se puede perder | una línea de fallo | el hallazgo, enterrado en el relato |

La regla observable: **la primera línea del reporte dice el resultado; la segunda, la consecuencia o el próximo paso; no se repite el desarrollo de pasos anteriores.** Se evalúa con una rúbrica sobre los fixtures del grupo 7. Tiene prioridad 2 dentro de esta ficha (tasks.md).

### D17 — "El sistema" es el agente siguiendo los módulos, más el catálogo, más el arnés

**Acordado** (`CHANGES.md:49`; la 03 espera los textos de esta ficha en su tarea 2.3 y en la fila 27 de la matriz común).

- **"El sistema"** es el agente siguiendo los módulos de instrucción, más lo declarado en el catálogo (`internal/catalog/harnesses.yaml`). No hay subcomando, hook ni proceso nuevo en el binario.
- **"Registrar"** es que el agente lo declare en su reporte en una línea con forma fija (por ejemplo, `unnecessary call: <tool> <input>`), para que el arnés lo cuente. Las formas fijas son las de los módulos (`entrega-03-04.md` §2) y las especificaciones las citan literales.
- **El arnés** (`eval/taskcontext/`, fuera del binario, reutiliza `eval/baseline/`) captura, redacta, evalúa y purga en evaluación (D8, D9).
- **Tests de Go** verifican lo declarativo y lo mecánico: el catálogo validado por `catalog.Load()`, los textos de módulo como archivos, el capturador, el redactor, el evaluador y la generación de tablas.
- **La conducta del agente** se verifica contra el cliente real (Claude Code y Codex), no contra una simulación (grupo 8).
- **La 05 redacta** los módulos `directed-code-reading`, `bounded-command-output`, `task-tool-selection` y `concise-agent-reporting` (antes `lectura-dirigida`, `resumen-acotado`, `herramientas` y `reporte`; el id coincide con la capacidad) y la tabla de carga, y los entrega a 03 y 04 (D23). No escribe en `sdd-orchestrator.md`.

*Alternativa descartada: implementar los registros y la orientación como código del instalador.* Es el runtime nuevo que el roadmap reserva.

### D18 — Tabla de herramientas por clase de tarea

**Propuesto a 03, 04, 07 y 08** (`CHANGES.md:166` pone *"qué capacidad es necesaria para cada tarea"* entre lo que se acuerda en grupo). Toda integración opcional está apagada por defecto y se habilita por acción. Las herramientas base del cliente no son opcionales: la tabla dice cuáles corresponden a cada acción, pero no las instala ni las apaga.

| Acción | Base | Opcionales habilitadas |
|---|---|---|
| `explore` | lectura dirigida, búsqueda | Context7 |
| `propose` | lectura dirigida, OpenSpec CLI | Context7 |
| `apply` | edición, tests | Context7; skills de dominio del registro (`go-testing` en este repo); Playwright solo si la acción toca UI |
| `verify` | tests, lectura dirigida | Playwright solo si hay UI |
| `archive` | OpenSpec CLI | ninguna |

- **Engram** no es integración opcional: llega por la capa universal y el módulo `engram-opsx` (D2).
- **Integraciones de diseño** (Stitch, `emil-design-eng`, `design-taste-frontend-v1`) se habilitan solo con el harness `design-mockups` activo, según la política del proyecto.

La tabla se declara en el catálogo (tarea 5.1), `docs/` la genera desde ahí (tarea 5.5) y el módulo `task-tool-selection` la lleva en inglés (`entrega-03-04.md` §2); hasta que exista el catálogo, la de este D18 manda.

**Solo recomendada o restringida por el cliente.** El informe pide documentar qué herramientas están solo recomendadas y cuáles restringe el cliente. "No habilitada" tiene tres niveles, con nombre:

| Nivel | Nombre | Qué es | Claude Code | Codex |
|---|---|---|---|---|
| **(a)** | *solo recomendada* | el módulo le pide al agente que no la use | `CLAUDE.md`, cuerpo del agente o de la skill: *"as context, not enforced configuration"* (https://code.claude.com/docs/en/memory) | `AGENTS.md`, `developer_instructions` o cuerpo de la skill |
| **(b)** | *fuera del contexto* | el agente no ve que existe | skill con `disable-model-invocation: true` (https://code.claude.com/docs/en/skills); MCP con búsqueda de herramientas (https://code.claude.com/docs/en/mcp) | skill con `policy.allow_implicit_invocation: false` (https://learn.chatgpt.com/docs/build-skills). **Medido** (X9, X11): la invocación implícita queda bloqueada y el cuerpo no se carga, pero nombre y descripción siguen listados; no es (b) completo |
| **(c)** | *restringida por el cliente* | aunque el agente lo intente, la llamada no ocurre | `permissions.deny` con `Skill(skill:<nombre>)`, `Agent(<nombre>)` o `mcp__<servidor>__*` (https://code.claude.com/docs/en/permissions); `tools`, `disallowedTools` y `mcpServers` de un subagente (https://code.claude.com/docs/en/sub-agents) | `mcp_servers.<id>.enabled_tools` / `disabled_tools` y `enabled` (https://learn.chatgpt.com/docs/config-file/config-reference); `skills.config` y `features.*=false` en un rol (https://learn.chatgpt.com/docs/agent-configuration/subagents) |

Por acción, dentro de una misma sesión:

- **Claude:** (c) solo si la acción corre en un subagente con su propia lista. En el hilo principal la lista queda fija para la sesión, así que ahí el control por acción es (a).
- **Codex:** el archivo de un rol no puede fijar `mcp_servers` (*"Configure MCP servers in the parent session"*), así que el control por acción de una integración MCP es **solo (a)**, declarado como límite. Para skills, `skills.config` en el rol da (c).
- `allowed-tools` de una skill de Claude **no** restringe (*"It does not restrict which tools are available"*): no cuenta como (c).

El catálogo declara el nivel por integración, cliente y superficie (tarea 5.2). Una integración *solo recomendada* sigue invocable: si el agente la invoca, queda registrado como uso fuera de clase, y el reporte nunca la presenta como bloqueada.

### D19 — Tabla de carga de módulos por acción

**Acordado** que la 05 define cuándo se carga cada módulo (informe unificado, contrato *Activación, núcleo y módulos*). La tabla —nueve módulos por cinco acciones, con la columna *Especialista* y sus condiciones— vive en `entrega-03-04.md` §3, junto con el texto en inglés que va al `SKILL.md`; es la única fuente de esa regla hasta que el catálogo la declare (tarea 6.1), una función pura la resuelva y el texto del `SKILL.md` se **genere** desde ella (tarea 6.2), porque el agente no lee el catálogo embebido en el binario. Reemplaza la tabla que hoy escriben a mano la 03 (`actionTable`, sin `explore` ni `verify`, medido R1) y la 04 (string `actions`). La matriz común dice en qué capa vive cada regla; esta tabla dice cuándo se lee cada módulo. Un módulo de toggle inactivo no figura (matriz común); si el inactivo es obligatorio para una acción, la fila dice que esa acción se detiene. `model-assignments` se omite en Codex hasta la 09: límite declarado.

**Por qué `bounded-command-output` es obligatorio en `apply` y `verify`.** Son las acciones que corren tests y producen evidencia; sin el módulo, la captura y la referencia no existen y el resultado no puede verificarse. Se propone a la 06 que su contrato de evidencia se cargue con la misma condición, en ejecución directa y delegada (*Coordinación pendiente*).

**Regla de módulo ausente** (la generación de la tabla la escribe debajo; texto en `entrega-03-04.md` §3): un módulo obligatorio ausente detiene solo la acción que lo requiere, informa qué falta y deja seguir las tareas que no dependen de él; un módulo opcional ausente se avisa y la acción se resuelve directo. **Un módulo vacío cuenta como ausente**: con un módulo de 0 B Claude tuvo tres conductas distintas en tres corridas (medido, C13), mientras que con el módulo borrado se detuvo e informó (medido, C12); y el verify de la 03 deja pasar un módulo vacío (medido, R3). Es compatible con la *Aceptación transversal* (`CHANGES.md:47`) y con la D2 de la 07 (`opsx-selective-delegation/design.md:58`).

**Compactación y reanudación.** Después de `/compact` el cuerpo de la skill y los módulos leídos salen del contexto y solo sobrevive el aviso universal de reinvocar que escribe la 03 (medido, C11). Regla: tras compactar o reanudar, el coordinador vuelve a invocar la skill y **relee los módulos de la acción en curso** antes de seguirla. El texto va debajo de la tabla del `SKILL.md`.

**Modo legado** (Claude Code anterior a 2.1.110; la 03 deja el coordinador como bloque único en `CLAUDE.md`, medido con versión simulada, C14). No hay skill ni tabla, así que los módulos de la 05 **no se entregan** y ninguna acción puede detenerse por su ausencia. Comportamiento declarado: la paridad (tarea 6.8) registra el modo legado como límite; todo resultado de `apply` o `verify` en esa instalación queda *sin verificar* (D20), porque falta el `bounded-command-output` obligatorio; y se le pide a la 03 que el aviso de instalación degradada nombre que los módulos de contexto no se entregan. La 05 no agrega texto al bloque legado, que la 03 conserva byte a byte.

**"Salida verbosa"**, para la señal S1 de la 07, es la condición de D3: más de 300 líneas o más de 10.000 caracteres. La 07 ya cita las 300 líneas (`opsx-selective-delegation/design.md:49`); falta avisarle los caracteres.

**Superficie sin subagentes.** Codex documenta subagentes en la app de escritorio, el CLI y la extensión del IDE, y no en la nube (https://learn.chatgpt.com/docs/agent-configuration/subagents). Ahí `verify` queda **detenida por diseño**, como límite de esa superficie. Qué superficies soporta cada cliente lo fijan 03 y 04.

### D20 — Un resultado declarado por el agente queda sin verificar hasta corroboración independiente

**Acordado** (informe unificado, *05/06*: *"Un resultado declarado por el agente queda sin verificar hasta corroborar de forma independiente la ejecución y el estado del código"*). Vale igual en ejecución directa y delegada: el resumen del coordinador sobre su propio comando y el que un worker le devuelve.

Se separan tres preguntas:

| Pregunta | Qué la responde | De quién es |
|---|---|---|
| **integridad**: ¿la copia es la que dice la referencia? | la huella sha256 coincide | 05 (D9) |
| **procedencia**: ¿la ejecutó y capturó alguien que no es quien reporta? | el registro de captura: fuente `arnés`, o `cliente:<fuente>` con cobertura `demostrada` (D22), o re-ejecución por otra parte con captura | 05 aporta el registro; los estados y la regla son de 06 |
| **suficiencia**: ¿alcanza para el requisito? | la revisión de 06 | 06 |

Un resultado pasa a **verificado** solo con integridad y procedencia independiente, y la suficiencia la decide 06. Con fuente `agente`, o con cobertura `no demostrada`, queda **sin verificar** aunque la huella coincida: un archivo, una huella y un resumen fabricados por la misma parte son coherentes entre sí y no prueban nada (auditoría de ramas del 10/11, N5.1). Si la corroboración encuentra un fallo omitido o una huella distinta, pasa a **rechazado**. Si verificar exige un detalle que la redacción descartó, queda sin verificar hasta conseguir evidencia independiente. Los tres estados son los de la 06 (`opsx-verifiable-evidence/design.md:57-61`).

### D21 — Referencia de invocación por cliente

**Acordado** (informe unificado, *03/04*). La invocación explícita es una skill en la sesión actual: `/sdd-orchestrator` en Claude Code, con `disable-model-invocation: true` (https://code.claude.com/docs/en/skills), y `$sdd-orchestrator` en Codex, con `policy.allow_implicit_invocation: false` (https://learn.chatgpt.com/docs/build-skills); en la app de ChatGPT la skill se elige con `@`. El texto que la nombra es la fila 3 de la matriz común y lo escriben 03 y 04; esta ficha solo exige que cada cliente nombre la suya y no la del otro (tarea 6.7). En Codex, la invocación con `$` de una skill con `allow_implicit_invocation: false` tiene un reporte abierto (https://github.com/openai/codex/issues/23454): la alternativa es `/skills` y la 04 la verifica por versión. **Medido** en Codex 0.167.1 sobre Windows: `$sdd-orchestrator` inyecta el cuerpo (X10) y el pedido sin `$` no la invoca (X11); `/skills` sin interfaz quedó **no probado**. En Claude Code 2.1.319, **medido**: la skill no figura para el modelo hasta invocarla (C7) y `/sdd-orchestrator` la carga (C8).

**Qué especialistas reciben las reglas universales.** Explore y Plan *"start without your CLAUDE.md files"*, igual que un agente con `omitClaudeMd: true` (documentado, https://code.claude.com/docs/en/sub-agents; **medido**, C10: Explore y Plan no ven `CLAUDE.md`, y un general-purpose ve `CLAUDE.md` pero no `SKILL.md`): para ellos el encargo repite las reglas obligatorias y `governance` (matriz común, fila 1), y a todo especialista el coordinador le pega los módulos. En Codex, si un subagente lee `AGENTS.md` sigue **no probado** (la política de shell de Windows impide lanzarlo, X12); hasta que la 04 lo responda, el encargo también las repite.

### D22 — Matriz de cobertura de captura por cliente, interfaz y sistema

**De la 05**, con el formato acordado para plataformas (informe unificado, *Plataformas*: Windows y Linux como mínimo; distribuida, declarada y probada). Es lo que la *Coordinación pendiente* 05 → 06 espera: *"La elección del mecanismo de captura queda pendiente de que 05 demuestre qué fuente cubre cada interfaz antes del truncamiento."*

Estados de una celda: **distribuida** (el cliente se instala ahí), **declarada** (la documentación dice que la fuente existe), **probada** (la fuente coincidió con la salida conocida en una corrida real; *sonda* si fue con generadores propios, *8.1* cuando la tarea 8.1 la corre con los fixtures 2.1 y 2.7 y deja id de corrida), **truncada** (la corrida mostró recorte), **no disponible** (la fuente no existe en esa interfaz) y **no probada** (con su motivo). Solo una celda *probada 8.1* cuenta como independiente (D20).

Estado al 2026-11-13 (`entrega-03-04.md` §5 y §6):

| Cliente · interfaz · sistema | Fuente | Estado | Respaldo |
|---|---|---|---|
| cualquiera · arnés | capturador del arnés sobre el comando del fixture | referencia por construcción | — |
| Claude Code · CLI · Windows | lo que recibe el modelo (`stream-json`, `tool_result`) en un comando **fallido** | **truncada**: ~10.040 caracteres de cabeza y cola, la cola no es el final, sin ruta | medido (C1–C5) |
| Claude Code · CLI · Windows | archivo del cliente en un comando **exitoso** (`tool-results/<id>.txt`) | **probada (sonda)** para exit 0; copia del cliente sin redactar (D9), y no existe en un fallo | medido (C6) |
| Claude Code · CLI · Windows | hooks `PostToolUse` (`tool_response.stdout`, `stderr`) y `PostToolUseFailure` (`error`, documentado como *"middle-truncates"*, https://code.claude.com/docs/en/hooks) | declarada; no probada (no se instalaron hooks) | documentado |
| Claude Code · IDE · Windows | los mismos hooks (*"fires the same hook events wherever it runs"*) | declarada; no probada (sin IDE operable) | documentado |
| Codex · CLI · Windows | `codex exec --json`, `item.completed` de `command_execution`: `aggregated_output`, `exit_code` | **probada (sonda)**: salida completa (129.047 caracteres) en stdout, solo stderr, exit 0 y 1; falta 8.1 con los fixtures del arnés | medido (X7) |
| Codex · CLI · Windows | lo que recibe el modelo; historial de la sesión | **truncada**: cabeza y cola sin aviso al modelo; el historial guarda la recortada | medido (X1–X6, X8) |
| Codex · CLI · Windows | App Server conducido por el arnés: `item/commandExecution/outputDelta`, `aggregatedOutput?`, `exitCode?` | declarada; no probada | documentado (https://learn.chatgpt.com/docs/app-server) |
| Codex · IDE · Windows | App Server interno de la extensión | **no disponible** como fuente externa: el arnés no puede suscribirse a la instancia de la extensión | documentado |
| Codex · escritorio | — | no probada; la 04 no promete aceptación en escritorio (`codex-native-opsx/design.md:59-63`) | documentado |
| Codex · nube | sin disco persistente del usuario | **no disponible** (D9); la 04 tampoco la promete | documentado |
| Claude Code y Codex · CLI e IDE · Linux | las mismas fuentes | **no probada** (sin cliente nativo en WSL) | — |

Conclusiones. En Claude, el caso que más importa —un comando que falla con salida larga— **no tiene fuente completa en la sesión**: la captura independiente es la del arnés, y en una sesión real la del agente (D8), que no supera *sin verificar*. En Codex CLI el arnés tiene una fuente del cliente que, en la sonda, entregó todo: la tarea 8.1 la convierte en celda *probada 8.1* y la tarea 3.6 la usa con muestras reales en lugar de derivadas de la documentación. Donde no haya una celda *probada 8.1*, el resultado de una sesión real queda *sin verificar* (D20).

### D23 — Qué entrega la 05 a 03 y 04

**Acordado** que la 05 entrega texto y la tabla de carga (D17); el contenido de la entrega es **de la 05**. Responde la fila 27 de la matriz común (*"Lectura dirigida, resumen acotado, herramientas, reporte conciso · módulos nuevos · según 05 D2 · pendiente de texto de 05"*) y la tarea 2.3 de la 03.

**Qué se entrega hoy.** El texto, no el código: los cuatro módulos en inglés, la tabla acción → módulos con su regla de módulo ausente o vacío y la de compactación, y las rutas por cliente, todo en `entrega-03-04.md` (§2 y §3). Estado: *texto listo para integrar; integración en código pendiente de la ronda de apply*.

| Entrega | Texto hoy | En código (apply) | Tarea | Qué hacen 03 y 04 |
|---|---|---|---|---|
| módulo `directed-code-reading` | `entrega-03-04.md` §2 | `internal/harness/config/assets/modules/directed-code-reading.md` | 4.4 | lo registran en el compositor común |
| módulo `bounded-command-output` | ídem | `…/modules/bounded-command-output.md` | 6.3 | ídem |
| módulo `task-tool-selection` | ídem | `…/modules/task-tool-selection.md` | 5.3 | ídem |
| módulo `concise-agent-reporting` | ídem | `…/modules/concise-agent-reporting.md` | 7.2 | ídem |
| tabla acción → módulos, regla de módulo ausente o vacío, compactación | `entrega-03-04.md` §3 | generada desde el catálogo | 6.1, 6.2 | la incluyen en el `SKILL.md` sin editarla |

**Compositor común con registro de módulos (propuesto a 03 y 04).** Hoy hay dos compositores que se superponen (medido y documentado, R5 y R6) y los dos fijan la tabla en código. Un sexto módulo anotado sin toggle se descarta en silencio en la 03 (medido, R2), así que copiar los archivos a `modules/` no alcanza. La 05 propone que 03 y 04 compartan un registro con esta forma, sin fijar la implementación:

| Campo | Qué es |
|---|---|
| id | nombre estable del módulo (`artifacts`, `bounded-command-output`, …), igual al del archivo `modules/<id>.md` |
| texto | contenido del archivo, desde un asset o un fragmento partido de un asset |
| carga por acción | para cada acción: obligatorio, opcional o no se carga, con su condición (por ejemplo, `tdd` solo con Strict TDD y runner) |
| especialista | si el especialista lo recibe en el encargo y cuándo |
| toggle | opcional: si existe y está inactivo, el módulo no se entrega y una acción que lo tenga como obligatorio se detiene. Un módulo sin toggle (los cuatro de la 05) se entrega siempre |
| clientes | en qué clientes se entrega (`model-assignments` no en Codex hasta la 09) |

Con ese registro, la tabla de `entrega-03-04.md` §3 es la **única fuente** y reemplaza `actionTable`, `exists` y `ExpectedModules` de la 03 y el string `actions` de la 04; el verify de la instalación compara lo instalado con el registro y falla con un módulo ausente o vacío (el de la 04 ya compara byte a byte, R4). El texto del `SKILL.md` se genera desde el registro por cliente, con la ruta de cada uno (`${CLAUDE_SKILL_DIR}/modules/<id>.md` en Claude, medido C8; `modules/<id>.md` relativa en Codex, no probado X12).

Los cuatro archivos no llevan frontmatter de skill, para que el cliente no los liste como skills propias, y no usan los encabezados `## Core Principles` ni `## Rules` (matriz común, *Reglas de la tabla*). La 05 **no** entrega texto de la capa universal ni de la invocación: son filas de la matriz común que escriben 03 y 04.

## Ubicación de los módulos

**Acordado** (matriz común, *Capas*: *"Módulo `<x>`: archivo de apoyo de la skill, se lee sólo para la acción que lo pide"*). Los módulos son archivos dentro de la skill del orquestador, `sdd-orchestrator/modules/<id>.md`, en los dos clientes. Una skill puede traer archivos de apoyo que se leen bajo demanda (https://code.claude.com/docs/en/skills, https://learn.chatgpt.com/docs/build-skills). La materialización —que el instalador entregue una skill con `modules/`— es de 03 y 04, hoy en dos ramas con dos compositores (D23). Si se delega, el coordinador pega el módulo en el encargo: el especialista no ve la skill (medido, C9 y C10).

## Risks / Trade-offs

- **Ninguna fuente del cliente queda probada →** el arnés sigue siendo la fuente independiente de la evaluación, y en sesiones reales todo resultado queda *sin verificar* (D20). Es una conclusión válida de la ficha, no un fallo: lleva al grupo el insumo para la decisión de arquitectura de `CHANGES.md:49`.
- **La captura declarada por el agente puede faltar o estar mal →** queda *sin verificar* por construcción; sin captura, el resumen es "parcial — sin verificar".
- **Seis fichas tocan `sdd-orchestrator.md` →** la matriz común es la única fuente del reparto; esta ficha entrega archivos nuevos y una tabla generada (D23), no ediciones al asset.
- **03 y 04 integran con compositores distintos →** cada módulo de la 05 tendría que cablearse dos veces y un módulo sin toggle se perdería en la 03 (R2). D23 propone el registro común; mientras no exista, `entrega-03-04.md` §1 dice qué cambiar en cada rama.
- **El cliente le oculta al modelo cuánto perdió →** Claude subestima la marca y Codex no avisa (D3). El módulo prohíbe resumir desde la vista del cliente y prescribe la captura antes de leer.
- **Resolver módulos por acción agrega un punto de fallo →** un módulo obligatorio ausente detiene la acción e informa (`CHANGES.md:47`); D19 fija cuáles son obligatorios.
- **Mover instrucciones fuera del bloque siempre presente puede degradar el comportamiento →** ninguna adopción se decide acá; la comparación con y sin cada módulo usa el protocolo de 02 (tareas 9.3 y 9.4).
- **Una tabla de herramientas que en un cliente es solo recomendada puede leerse como aislamiento →** D18 nombra cada nivel y prohíbe presentar (a) como control.
- **La redacción puede fallar →** en el arnés es código determinista probado con datos ficticios (tarea 3.3); en la captura declarada es instrucción y lo dice. Las copias del cliente quedan fuera y se declaran.
- **La retención en sesiones reales no tiene quien borre →** se declara como recomendación (D9); en el arnés purga el arnés.
- **Las recetas de captura dependen del shell →** se fijan por shell y se prueban en Windows y Linux (tareas 3.8 y 8.1).

## Migration Plan

No aplica a esta ficha: no modifica configuración instalada. La migración del bloque administrado en `CLAUDE.md`/`AGENTS.md` pertenece a 03 y 04, con snapshot, preservación de contenido del usuario y rollback; esta ficha no introduce una segunda ruta de migración. Ninguna tarea toca `internal/backup`, `internal/filemerge` ni `internal/pipeline`.

**OpenCode y Antigravity** mantienen el bloque global actual. El roadmap es de Claude y Codex (`CHANGES.md:3`).

## UI Mockups

Decision: no (no UI) — 2026-11-05. El change no agrega ni modifica pantalla, componente, layout ni interacción.

## Coordinación pendiente

Lo que depende de otra ficha, qué resultado necesita y qué comprobará esta ficha cuando llegue. El envío de cada pedido está en `tasks.md` — *Workflow follow-up*; lo que cada ficha tiene que hacer en código con la entrega, en `entrega-03-04.md` §7.

**Estado real al 2026-11-13.** La 03 está archivada en su rama (PR #8) con 15 tareas abiertas, entre ellas la 2.3, que espera los textos de esta ficha. La 04 es un borrador (PR #9) con 10 de 21 tareas cerradas; su tarea 1.4 pide fijar las revisiones de 05 y 06 y su `SKILL.md:20` dice *"Consume context and capture modules from change 05 when available"*. Ninguna está en `main` y las dos ramas chocan en 9 archivos (R5). La entrega de esta ficha (`entrega-03-04.md`) deja satisfacibles la 2.3 de la 03 y la 1.4 de la 04 en lo que toca a la 05; falta integrarla en código.

| De → a | Qué falta | Qué se comprueba cuando llegue |
|---|---|---|
| 05 → 06 | La 06 confirma o contrapropone ubicación, acceso y vencimiento de la copia redactada (D9), la lista sensible (D14), el registro de captura y que su contrato de evidencia se cargue con `bounded-command-output` en `apply` y `verify`, directa o delegada (D19). La elección del mecanismo de captura ya tiene la matriz medida de D22 | que la referencia de cinco campos y el campo de fuente de captura entren en su recibo sin redefinirse, y que su regla de *verificado* use la procedencia de D20 |
| 05 → 03 y 04 | Compositor común con registro de módulos (D23) en lugar de dos compositores; registrar los cuatro módulos y la tabla de `entrega-03-04.md` | tareas 6.5 a 6.7 |
| 03 → 05 | Que el aviso de instalación degradada (modo legado) nombre que los módulos de contexto no se entregan (D19) | tarea 6.8 |
| 04 (Matías según el informe unificado) → 05 | Tarea 2.5 de la 04 (`verify` y contrato de evidencia por acción); si un subagente de Codex lee `AGENTS.md`; lectura real de módulos y módulo ausente o vacío en Codex, hoy no probados por la política de shell de Windows (X12). La lista de superficies ya está declarada en su design (`codex-native-opsx/design.md:59-63`) y `aggregated_output` ya está medido (X7) | tareas 6.8, 8.1 y la regla de D21 sobre el encargo |
| 02 → 05 | Uso de la lista de D14 en `eval/baseline/record` y del plazo de D9 en su D6 | que las trazas de 02 y las salidas de 05 compartan política |
| 05 → 07 | La 07 actualiza S1 con los 10.000 caracteres y adopta D20 para el retorno del worker | que su escenario de salida verbosa use el par de D3 |
| 05 → 01 u 08 | Piso equivalente al de Claude en Codex (D14) | que el test de cobertura del piso pueda extenderse a Codex |
| 05 → matriz común | Resultado de la tarea 9.4 sobre las filas de memoria | que la matriz lo use como insumo, sin cambiarla desde esta ficha |

## Puntos de grupo y su estado

| Punto | Decisión | Estado | Si cambia |
|---|---|---|---|
| Matriz de módulos | la matriz común de la 03, con la columna de Codex de la 04 | acordado (12/11) | — |
| Ubicación de los módulos | archivos `modules/<id>.md` dentro de la skill | acordado (12/11) | la ruta de las tareas 4.4, 5.3, 6.3 y 7.2 |
| Compositor común | un registro de módulos compartido por 03 y 04, con la tabla de la 05 como única fuente (D23) | propuesto a 03 y 04 (13/11) | las tareas 6.5 a 6.7 y lo que 03 y 04 cambian en código (`entrega-03-04.md` §7) |
| Ids de los módulos de la 05 | el nombre de la capacidad, en inglés (D17) | de la 05 (13/11) | los nombres en la tabla de `entrega-03-04.md` §3 |
| Eje de clases de tarea | las cinco acciones OPSX (D10) | acordado (matriz común) | — |
| Reparto por capa | el de la matriz común; la 05 mide memoria (D2) | acordado (12/11) | la tarea 9.4 |
| Herramientas por clase | tabla y niveles con nombre (D18) | propuesto a 03, 04, 07 y 08 | la tabla de D18 y las tareas 5.1 a 5.5 |
| Copia redactada | ubicación, acceso, vencimiento y registro de captura (D9) | propuesto a 06 | los requisitos de ubicación y retención y las tareas 3.4, 3.5 y 8.5 |
| Lista sensible | una sola, en el catálogo, dueña la 05 (D14) | acordado (11/11) | — |
| Roles de captura | tres roles separados y captura independiente (D8) | acordado (11/11) | — |
| Cobertura por interfaz | matriz de D22 | de la 05; la espera la 06 | las celdas, según la tarea 8.1 |
| Quién ejecuta | agente con módulos, catálogo y arnés (D17) | acordado (`CHANGES.md:49`) | — |
| Interfaz con la 07 | salida verbosa de D3; tabla de carga de D19 | propuesto a 07 (300 líneas ya adoptadas) | el escenario de S1 de la 07 |
| Resultado sin verificar | D20 en ejecución directa y delegada | acordado (11/11) | — |
| Invocación por cliente | D21 | acordado (11/11) | — |
| Agentes incluidos | Claude y Codex; OpenCode y Antigravity mantienen el bloque global | acordado (`CHANGES.md:3`) | — |

## Decisiones para el grupo

Mejoras que surgieron de la revisión del 13/11 y que **no** se aplican en esta ronda porque
cambian convenciones de más de una ficha. Quedan para decidir en grupo:

| Decisión | Qué cambiaría | Por qué no se aplica ahora |
|---|---|---|
| Ids en inglés para todos los artefactos | nombres de requisitos, estados y formas en specs y tasks de todas las fichas | es convención común; esta ficha solo pasó a inglés los ids de sus módulos y las líneas de forma fija, porque se insertan en assets en inglés (D17, D23) |
| Método testigo en todos los escenarios de las specs | cada escenario de conducta nombraría su testigo (texto único, oráculo de transcripción) como hacen las pruebas de la 03 | cambia el formato de los escenarios de todas las fichas |
| `governance` solo a los especialistas que escriben | hoy el encargo de Explore y Plan repite `governance` (D21); se podría limitar a los que editan | toca la fila 1 y la 18 de la matriz común |
| Guarda de tamaño del `SKILL.md` y de los módulos | un test que falle si el texto compuesto supera un tope | D12 deja los topes fuera del alcance de la 05; la 03 ya fijó 4.000 tokens para su `SKILL.md` |
