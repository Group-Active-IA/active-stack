# Design

## Context

El stack instala el orquestador como texto: se compone desde `internal/harness/config/assets/<agente>/sdd-orchestrator.md` con los toggles `tdd`, `engram`, `model-routing`, `delegation` y `governance`, y se inyecta en `CLAUDE.md` o `AGENTS.md` con markers idempotentes. El instalador no corre mientras el agente trabaja. Por eso el contrato de evidencia es, en esta etapa, **instrucciones más comprobaciones que ejecuta el coordinador con comandos de shell**. No hay forma de impedir una violación sin controles nativos, que son trabajo de la propuesta 08.

**Base y fuentes (revisadas el 2026-11-13).** El código, 03 y 04 se citan contra `origin/main` `d5e95bd` (02, 03 y 04 mergeados; PR #8 de 03 y PR #9 de 04). **03 y 04 están en `main`.** 03 (`claude-native-opsx`) está archivado en `openspec/changes/archive/2026-11-13-claude-native-opsx/` y su matriz en `openspec/matriz-reglas-opsx.md`, ambos en `origin/main`. 04 está archivado en `openspec/changes/archive/2026-11-13-codex-native-opsx/` y sus assets en `internal/harness/config/assets/codex/native/` (`main`). Las propuestas todavía no mergeadas (05 y 07) se citan contra la rama donde está su trabajo:

| Change | Rama | Revisión consultada |
|---|---|---|
| 05 `opsx-task-context` | `origin/propose/opsx-task-context` | `4337e36` |
| 07 `opsx-selective-delegation` | `origin/propose/opsx-selective-delegation` | `762e3de` |

Las citas a 05 y 07 son **de sección** (D9, D20, D22, D6, nombre de requisito) y valen contra esa revisión; la tabla es la única que nombra el commit. Las citas a `origin/main` `d5e95bd` (incluidas 03 y 04 archivadas) llevan número de línea, re-verificadas el 2026-11-13.

El roadmap apunta a un orquestador de invocación explícita, con módulos independientes (TDD, delegación y coordinación) que se cargan según el rol y la acción. Esta propuesta escribe el contrato para ese modelo. Los puntos que dependen de 03, 04 y 05 están marcados (TBD).

**Glosario.**

| Término | Dueño | Qué es |
|---|---|---|
| **recibo de ejecución** | 06 (este change) | Registro de **una ejecución de un comando** (D2): evidencia de que algo se corrió, sobre qué estado y con qué resultado. |
| **registro de instalación** | 04 | Registro de qué archivos instaló y es dueño el instalador, usado por `update` y por el retiro (04, `openspec/changes/archive/2026-11-13-codex-native-opsx/design.md`, sección «Evidencia y retorno»). No es evidencia de comandos y no se rige por este contrato. |

Esta propuesta dice siempre "recibo de ejecución". Donde 03, 04 y 07 escriben "recibo" o `recibos` (por ejemplo el campo `recibos` del sobre de 07) se refieren al recibo de ejecución.

## Goals / Non-Goals

**Goals**

- Que un "pasó" de un sub-agente se pueda atar a una ejecución real, sobre el estado exacto del código, con un comando que cubra el requisito.
- Que la duda se resuelva siempre hacia *sin verificar*, y que una ejecución sin recibos nunca quede *verificada*.
- Mismo contrato y misma aceptación en Claude y en Codex, en ejecución directa y delegada.
- No exponer datos sensibles al conservar evidencia.

**Non-Goals**

- Impedir que un worker mienta (eso es 08 y, más adelante, un juez externo).
- Firmar recibos de ejecución, ledger, subcomandos del binario o un runtime nuevo.
- Reutilizar `internal/verify`.
- Acotar o resumir salidas (05), decidir cuándo delegar o definir el sobre de retorno (07).
- Redefinir lo que 03 y 04 ya definieron (capas, módulos, `verify` en Codex): aquí solo se propone y se acuerda.

## Decisions

### D1 — No se reutiliza `internal/verify`

`internal/verify` es el reporte de health checks posteriores a la instalación. En `internal/verify/report.go` (`origin/main` `d5e95bd`), `BuildReport` (`:23`) agrega `CheckResult` con estados `passed`, `failed`, `skipped` y `warning`, y fija `report.Ready = report.Failed == 0` (`:38`); los `skipped`, los `warning` y la ausencia total de checks no bloquean. La semántica que pide esta propuesta es la opuesta (lo incompleto no se da por bueno). Mezclar ambas ocultaría la diferencia, y el caso observado el 02/11 (0 checks, ✓) muestra el riesgo. El contrato se define aparte.

*Alternativa descartada:* extender `Report` con recibos de ejecución. Mezcla salud de instalación con evidencia de tareas y comparte un estado que se comporta al revés.

### D2 — Un recibo de ejecución por ejecución

Cada recibo de ejecución describe **una** ejecución de **un** comando. Si un requisito se cubre con varios comandos, hay varios recibos de ejecución que lo referencian.

| Campo | Contenido | Cómo se obtiene |
|---|---|---|
| `corrida` | Identificador de la corrida | lo asigna la fuente de captura (TBD, 05) |
| `arbol` | Directorio y árbol donde corrió, y **estado exacto probado**: commit base más, si había cambios sin confirmar, una huella del estado completo (preparados, sin preparar y archivos nuevos). `HEAD` más un indicador "dirty" **no alcanza**: dos diffs distintos sobre el mismo `HEAD` deben dar estados distintos | `git rev-parse --show-toplevel`, `git rev-parse HEAD` y la huella del estado (ver **Huella del estado probado**) |
| `comando` | Comando completo, tal como se ejecutó | lo declara el worker; la fuente independiente lo corrobora (D3) |
| `resultado` | Código de salida y marca de tiempo | declarado; corroborado (D3) |
| `requisitos` | IDs de requisitos o escenarios del change que dice cubrir | lo declara el worker |
| `salida` | Referencia a la **copia redactada**: **una** huella de ese archivo (`{algoritmo, valor}`), si fue truncada y el **registro de captura** (qué fuente la capturó y cuándo). Formato de la referencia y del registro de captura: 05 (D9); 06 no define sus campos, solo los estados (D4) | huella sobre los bytes del archivo redactado; registro de captura de 05 D9 |

**Persistencia (lo acordado).** Se persiste solo la copia redactada, con una huella, y el registro de captura. No existe una copia con secretos de la que el recibo de ejecución dependa. El **registro de captura** es el de 05 D9: fuente (`arnés`, `cliente:<fuente>` o `agente`), redactor, cliente, interfaz, sistema, versión, código de salida, cobertura previa al truncamiento y conteo de redacciones por categoría. 06 lo referencia y no agrega campos. El mecanismo de captura independiente por interfaz lo define 05 (matriz D22, ver D3). Ubicación, acceso y vencimiento: **propuestos por 05 (D9), 06 propone aceptar; pendiente de confirmación** (ver "Respuesta de 06 a 05"). Esta propuesta no fija una ruta propia.

**Huella del archivo (definición).** Es el **sha256 de los bytes exactos** del archivo redactado, sin conversión de fines de línea ni filtros, y el campo **nombra el algoritmo**: `{algoritmo: "sha256", valor: "<hex>"}`. Así lo hacen 02 (`eval/baseline/record/record.go:42-45`, `Fingerprint{Algorithm, Value}`; su comentario `:7` remite a 06) y 05, que define el formato de huella de la referencia: `sha256:<hex>`, con el algoritmo nombrado y los bytes sin normalizar fines de línea (05, D9 y requisito «Copia recuperable por referencia estable»). 06 acepta ese formato; `{algoritmo, valor}` (02) y `sha256:<hex>` (05) son equivalentes y 06 no cambia su definición. Bytes iguales dan el mismo valor en Windows y en Linux; bytes distintos, incluidos LF frente a CRLF, dan valores distintos. Eso exige que la captura escriba los bytes sin traducir fines de línea.

*No se usa `git hash-object` como definición.* Hashea un blob de Git (cabecera `blob <tamaño>\0` más contenido, SHA-1 por defecto) y no un digest del archivo: el valor no se puede recomputar con `sha256sum` ni con `Get-FileHash`, y depende de filtros (`--no-filters`) y del formato de objeto del repositorio.

**Huella del estado probado (working tree sucio).** Se alinea con `TreeHash(dir)` de `origin/main` (`eval/baseline/fixture/load.go:224-243`): sha256 sobre las líneas `"<sha256 del archivo>  <ruta>\n"` (dos espacios, ruta relativa con `/`), **ordenadas por ruta**; el mismo formato que `sha256sum`, recomputable con shell. Aplicada a un repositorio, el conjunto de archivos es el de la copia de trabajo (versionados modificados o no, más archivos nuevos no ignorados, sin los borrados y sin `.git`); la lista exacta y su obtención con Git (por ejemplo `git ls-files -z --cached --others --exclude-standard`) se midieron en la tarea 1.3 (`informes/medicion-huella-2026-11-13.md`). Dos advertencias que 02 ya midió: en Git Bash de Windows `sha256sum` imprime ` *archivo` (modo binario) y hay que usar `--text` (`openspec/changes/opsx-baseline/informes/verify-t2-1-breaker.md`, F3, `origin/main`); y la huella depende de los bytes de la copia de trabajo, por lo que `core.autocrlf` cambia el estado probado (eso es correcto: es otro estado). *Decisión con evidencia (propuesta de 06, pendiente de confirmación de Gianfranco con Emiliano):* TreeHash sobre los bytes de la copia de trabajo. Medido en Windows y en Linux (WSL2) con `core.autocrlf` en `true`, `false` e `input` (`informes/medicion-huella-2026-11-13.md`): los valores son idénticos entre configuraciones y plataformas, y LF y CRLF dan huellas distintas. *Alternativa descartada:* el árbol de Git de un índice temporal (`GIT_INDEX_FILE` + `git add -A` + `git write-tree`): el mismo estado da otro árbol según `core.autocrlf` y, con `true` o `input`, un archivo con CRLF y otro con LF dan el mismo árbol. Límites medidos: un archivo ignorado por Git no entra en la huella, y `TreeHash(dir)` recorre `.git`, así que para un repositorio hay que pasarle el conjunto de archivos de Git.

**Plataformas.** Windows y Linux son el mínimo (decidido). La matriz distingue tres columnas por plataforma:

| Columna | Qué significa | Fuente | Estado hoy |
|---|---|---|---|
| *distribuida* | el instalador publica binario | `.github/workflows/release.yml:71-76` (`GOOS=windows` `:71-73`, `GOOS=linux` `:74-76`, ambos `amd64`) | Windows amd64, Linux amd64 |
| *declarada* | la documentación dice que se soporta | `CLAUDE.md:13` ("Windows, macOS, Linux, WSL y Termux") | las cinco |
| *probada* | este contrato se ejecutó en ella | tarea 1.3 y 4.5 | Windows (Git Bash y PowerShell) y Linux **en WSL2** (huella de bytes y del estado probado, `informes/medicion-huella-2026-11-13.md`); Linux nativo o CI: (TBD) |

Para cerrar esta propuesta, Windows y Linux deben figurar como *probadas*; hoy lo están en la medición de la tarea 1.3, con el Linux de WSL2. macOS, WSL y Termux figuran como *declaradas*; no hay binario publicado para macOS en `release.yml` y no se promete aquí que estén probadas (TBD).

### D3 — Tres controles con resultados separados

- **Integridad.** El archivo conservado no cambió: existe y su huella recomputada coincide con la del recibo de ejecución. Es un control mecánico y **no prueba que la ejecución haya ocurrido**: quien fabrica un recibo de ejecución puede fabricar también la huella.
- **Procedencia.** La ejecución es real y corresponde al estado del código. Un resultado declarado por el worker queda *sin verificar* hasta que una **fuente independiente** o una **reejecución** corrobora (a) que el comando se ejecutó y su código de salida, y (b) el estado exacto del árbol en que se ejecutó (D2 `arbol`). La verificación puramente mecánica (huella más relato del worker) no es criterio suficiente. *Mecanismo de captura* (arnés o fuente del cliente, antes de cualquier truncamiento): lo define 05. **Estado actual según 05** (05, D22, citado, no decidido por 06):
  - Solo el **arnés** es referencia por construcción.
  - Claude Code CLI en Windows con comando fallido: salida **truncada**, sin fuente completa.
  - Codex CLI en Windows con `codex exec --json`: **probada como sonda (n=1)**; falta la tarea 8.1 de 05.
  - Codex IDE: **no disponible** como fuente externa.
  - Linux: **no probada**.

  *Consecuencia que escribe 06:* sin una fuente independiente demostrada para la interfaz, un resultado declarado queda *sin verificar*, salvo reejecución por otra parte con captura (05 D20).

  *Regla de reejecución (propuesta de 06, pendiente de confirmación):* una reejecución solo cuenta como fuente independiente si tiene **captura propia de una fuente independiente** (05 D20). La reejecución del coordinador sin esa captura **no corrobora** ni contradice: el recibo queda como si nadie hubiera corroborado. Fixtures: `reejecucion-con-captura-independiente` y `reejecucion-del-coordinador-sin-captura`.
- **Suficiencia.** El comando ejercita lo que el requisito pide y el resultado respalda el cumplimiento. Cada requisito del change debe nombrar qué evidencia lo satisface; lo que no se pueda evaluar queda *sin verificar*. Quién la evalúa: pregunta abierta (Open Questions, 3). 04 (informe `analisis-dependencias-05-06-2026-11-13.md`) y 05 (D20) esperan de 06 un criterio identificable. Propuesta de partida, **de 06 y pendiente de confirmación**: el revisor distinto del constructor que 04 ya especifica (D8) y que cada requisito nombre su evidencia.

Los tres resultados se informan por separado. Integridad aprobada no implica procedencia, y procedencia aprobada no implica suficiencia (tests verdes con un requisito incumplido).

*Coherencia del código de salida.* El código de salida lo escribe **quien captura**, en el registro de captura de 05 D9, no el worker. El código declarado por el worker se compara con ese valor y con la copia redactada. Como la redacción se hace al capturar, 05 exige que conserve la línea y su posición (05, requisito «Datos sensibles redactados antes de persistir»). Que la copia redactada alcance para decidir el código de salida se confirma con un fixture (TBD, tarea 5.1).

*Un detalle redactado no alcanza para verificar.* 06 referencia, sin redefinir, el requisito de 05 "Un detalle redactado no alcanza para verificar" (05, `bounded-command-output`): si verificar un requisito exige un detalle que la redacción descartó, el recibo queda *sin verificar* y no se persiste el original para evitarlo (tarea 4.12).

### D4 — Tres estados, con la duda hacia "sin verificar"

El texto de **rechazado** es el de la ficha 06 del roadmap (`CHANGES.md:194`, "Aceptación futura"): *rechazar recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes. Un recibo incompleto queda sin verificar.*

- **Verificado:** integridad, procedencia (corroborada por fuente independiente o reejecución) y suficiencia aprobadas para cada requisito declarado, y existe al menos un recibo de ejecución para él. **Una ejecución sin recibos nunca queda verificada.**
- **Rechazado**, los cinco casos de la ficha:
  1. recibo **inventado** (la copia no existe o su huella no coincide; o la fuente independiente o la reejecución muestran que la ejecución no ocurrió o dio otro resultado);
  2. recibo **viejo** (el estado del árbol corroborado es distinto del actual);
  3. comando **insuficiente** (no ejercita el requisito declarado);
  4. salida **truncada sin original**: aplica cuando la salida truncada **nunca tuvo** referencia a una copia redactada. Si la referencia existía y la copia venció, es el caso "copia vencida" (más abajo, estado (TBD), propuesta de partida: *sin verificar*);
  5. requisito **incumplido con tests verdes**.
  *Detalle operativo (no cambia la ficha):* un recibo falso solo puede rechazarse como "inventado" cuando algo lo contradice; sin contradicción ni corroboración queda *sin verificar* (acuerdo 05/06). **(TBD)**: si un recibo de **directorio ajeno** (producido fuera del repositorio o worktree del change) es *rechazado* o *sin verificar*, y lo mismo para una **copia vencida** (retención de 05): la ficha no los menciona. Propuesta de partida: ajeno = rechazado, vencida = sin verificar; se acuerda con 05. Los directorios se comparan normalizando separadores (`\` y `/`) y la barra final; límite **(TBD)**: la comparación distingue mayúsculas y minúsculas y no resuelve enlaces simbólicos ni nombres cortos 8.3 de Windows (fixtures `directorio-separadores-windows` y `directorio-ajeno-separadores-windows`). El vencimiento concreto lo propone 05 (D9) y 06 solo exige que la referencia lo informe.
- **Sin verificar:** recibo **incompleto** (ficha), falta la copia redactada por indisponibilidad (ver TBD de vencida), no hay corroboración independiente (acuerdo 05/06), no se pudo evaluar, o no hay ningún recibo de ejecución para el requisito. Es el estado por defecto y, a diferencia de un `skipped` en los health checks, **no cuenta como aprobado**. Un conjunto vacío de recibos de ejecución nunca produce *verificado*.

Un recibo de ejecución falso con la huella correcta pasa integridad; sin corroboración queda *sin verificar*, y si la reejecución o la fuente independiente lo contradicen, *rechazado*.

### D5 — Cuándo se carga el contrato y dónde vive: módulo propio `evidence` (PROPUESTA, pendiente de acuerdo con 03 y 04)

El contrato es texto. **Se carga cuando la acción requiere evidencia**, tanto si el coordinador ejecuta directo como si delega, y hay **el mismo paso de revisión de recibos de ejecución** en ambos casos.

**Estado actual en 03 y 04 (citado, no modificado).**

- La matriz común (`openspec/matriz-reglas-opsx.md:48`, fila 28, `origin/main`) ubica el contrato en la capa "módulo `delegation` y encargo", especialista "sí, si ejecuta comandos", con "pendiente de texto de 06" en Claude y Codex. El inventario de módulos de 03 coloca el extracto `## Evidence Contract` en `delegation.md` (`design.md` archivado, D4, `:181-183`) y D12 (`:342`) lo incluye "en `delegation.md` y en el encargo de acciones que ejecutan comandos, también cuando la acción corre en el hilo".
- El problema: la matriz fija que "un módulo existe sólo con su toggle activo" (`:74`) y `delegation` es un toggle. Con el toggle apagado, la ejecución directa quedaría sin contrato, aunque 03 (D12 `:342`), 04 (`openspec/changes/archive/2026-11-13-codex-native-opsx/design.md`, «Matriz común de reglas», fila 28 `:33`; requisito «Evidencia directa y delegada») y 07 (`design.md` D6, veredicto) piden el mismo paso de revisión en ejecución directa.
- 04 ya marca la fila como "ajuste coordinado pendiente" (`design.md` archivado de 04, `:33`) y carga el contrato en `apply` y `verify` (misma sección, `:42-43`).

**Propuesta.** Un módulo propio `evidence` con el contrato y el paso de revisión de recibos. Se carga en las acciones `apply` y `verify`. Cuando se delega una acción que ejecuta comandos, el coordinador pega **además** un extracto (`## Evidence Contract`) en el encargo, igual que hoy pega `## Project Standards (auto-resolved)`. El worker devuelve los recibos de ejecución en el campo `recibos` del sobre de 07 (D6) y el coordinador los revisa con el mismo paso.

- **Ejecución directa:** el coordinador carga `evidence`, ejecuta con captura (D3) y revisa los recibos que produce.
- **Delegada:** `delegation` sigue siendo el módulo de reglas de delegación y del sobre de 07; el extracto de `evidence` viaja en el encargo.

**Lista exacta de cambios que implicaría** (la propuesta no edita ninguno; los aplican sus dueños si el grupo acuerda):

| # | Dónde (rama) | Cambio |
|---|---|---|
| 1 | `openspec/matriz-reglas-opsx.md:48`, fila 28 (`origin/main`) | Capa: "módulo `delegation` y encargo" → "módulo `evidence`". Especialista: "sí, si ejecuta comandos" → "sí, extracto `## Evidence Contract` pegado en el encargo cuando ejecuta comandos". Claude: → `modules/evidence.md`. Codex: → módulo `evidence`. |
| 2 | Misma matriz, "Tabla acción → módulos" (`:68-70`) | `apply` → `artifacts`, `tdd`, **`evidence`**. Agregar `verify` → **`evidence`** (más `delegation` cuando exige revisor independiente, como ya figura en 04, `design.md` archivado `:43`). `delegar` → `delegation`, `model-assignments` y se aclara que el extracto de `evidence` se pega solo si el encargo ejecuta comandos. Hoy la tabla no tiene fila `verify`. |
| 3 | Misma matriz, regla `:74` ("un módulo existe sólo con su toggle activo") | Excepción para `evidence`: módulo obligatorio sin toggle propio (opción recomendada, porque su ausencia detiene la acción, D11 de 03). Alternativa: toggle `evidence` activo por defecto (TBD). |
| 4 | 03 `design.md` archivado, D4 (`:181-183`) | Quitar "extracto `## Evidence Contract` de 06" de `delegation.md`; agregar `evidence.md` al inventario inicial. |
| 5 | 03 `design.md` archivado, D12 (`:342`) | "Se incluye en `delegation.md` y en el encargo…" → "se incluye en `evidence.md` y en el encargo de acciones que ejecutan comandos, también cuando la acción corre en el hilo". |
| 6 | 03 `tasks.md` archivado, tarea 2.3 (`:59`) | "ubicarlos en `modules/delegation.md`" → ubicar el `## Evidence Contract` en `modules/evidence.md` y el sobre de 07 en `modules/delegation.md`. Nota: 03 está archivado y mergeado en `main`; cambiar la fila 28 requiere un seguimiento de 03 (Juani) sobre `main`, y 03 sigue esperando el texto de 06 (su tarea 2.3). |
| 7 | 04 `openspec/changes/archive/2026-11-13-codex-native-opsx/design.md:33` (`main`) | Cerrar el "ajuste coordinado pendiente" con el texto de la fila 28 nueva. 04 está archivada en `main`: requiere un seguimiento de Matías sobre `main`, como 03. |
| 8 | 04 `openspec/changes/archive/2026-11-13-codex-native-opsx/design.md:43` (`main`) | `verify` → "módulo `evidence`, y `delegation` cuando exige revisor independiente". |

`assets/codex/native/SKILL.md` de 04 (`main`, párrafo de evidencia y sección «Partial evidence requirements») ya exige cargar el contrato en ejecución directa y delegada, y 07 (`entrega-delegation.md`, sección «Evidence») ya contempla el extracto en el encargo: no necesitan cambio de fondo.

**Estado: PROPUESTA.** Pendiente de acuerdo con **03 (Juani)** y **04 (Matías)**; 07 (Neyen) y 05 (Emiliano) se informan porque dependen de dónde vive el texto. Hasta entonces 06 solo redacta el contenido (tarea 2.4) y no edita archivos de 03 ni de 04.

**No puede ser una regla universal.** 03 midió (D0, `design.md` archivado `:59`, Claude Code 2.1.318, 2026-11-06) que todo lo que queda en `CLAUDE.md` llega a la sesión, al coordinador y a los especialistas; un contrato ahí llegaría a quien no lo necesita.

**Dónde vive el archivo.** 03 decidió archivos de apoyo de la skill, `<skills>/sdd-orchestrator/modules/<módulo>.md`, que el coordinador pega en el encargo (D4 `:171-179` y decisión G3 `:365`, de Juani, 2026-11-13). 05 propuso lo mismo (`design.md` de `opsx-task-context`, `modules/<id>.md`, pendiente de revisión del grupo). `evidence.md` sigue ese mecanismo.

Si el módulo requerido no está disponible en una sesión, el coordinador informa la limitación y **detiene** la acción que lo necesita (regla transversal del roadmap; en 03 es D11, `:322-332`), en lugar de seguir sin contrato.

*Alternativa descartada:* un harness `config` propio con sección `active-stack:verifiable-evidence` en `sectionFragments`: viviría en `CLAUDE.md`/`AGENTS.md` y llegaría a todos los especialistas. *Alternativa también descartada:* dejar la fila 28 en `delegation`: con el toggle apagado no hay contrato en ejecución directa.

### D6 — Sobre de retorno y qué se guarda

**Sobre.** 07 es dueña del sobre de retorno (07, D6): `estado`, `partida`, `modificados`, `hallazgos`, `recibos`, `pendiente`. 06 define solo el **contenido** del campo `recibos`: una lista de recibos de ejecución, uno por comando ejecutado. Esta propuesta no define un sobre propio ni atribuye el recibo de ejecución a una sección de retorno de 04 (04 ya referencia el sobre de 07 y no crea otro: `design.md` archivado de 04, «Evidencia y retorno»). Si hacen falta otros campos, se piden a 07. 07 valida los recibos "según 06" como tercer paso (D6, validación) y pide el mismo paso de revisión cuando la acción es directa (D6); coincide con D5.

**Qué se guarda.** Se persiste solo la copia redactada (D2). 05 (D9 de `opsx-task-context`) redacta al capturar y es dueña del formato de la referencia. Ubicación y acceso: **propuestos por 05 (D9); 06 propone aceptar; pendiente de confirmación** ("Respuesta de 06 a 05"). Vencimiento: 05 propone archivo del change más siete días (D9), que 05 declara **supuesto sin medición**; 06 lo acepta como provisorio, solo exige que la referencia informe el vencimiento y no lo adopta como firme. Si la copia vence, la referencia informa que ya no está disponible y el recibo queda *sin verificar* (D4, TBD sobre "vencida").

**Qué se versiona (TBD).** Propuesta de partida: se versiona el texto del recibo de ejecución (campos de D2, sin la salida) y ninguna salida completa. 05 recomienda "nada versionado"; si eso incluye los recibos de ejecución, se acuerda en el grupo. Dónde viven dentro del change (la propuesta previa era `openspec/changes/<change>/evidence/`) queda (TBD) hasta confirmar que OpenSpec tolera la carpeta (tarea 1.4). Lo que no cambia: tras una corrida, ninguna salida completa aparece en `git status`.

### D7 — Paridad entre agentes

Los cambios en los assets de Claude y de Codex se prueban con golden files de composición. Hoy en `internal/harness/config/compose_test.go` (`origin/main` `d5e95bd`) existen `TestCompose_AllToggles` (Claude, `:41`), `TestCompose_AllToggles_opencode` (`:65`), `TestCompose_ClaudeDiagramIncludesVerify` (`:255`), y los golden `TestCompose_AllToggles_codex` (`:274`) y `_antigravity` (`:275`) vía `assertComposeGolden`, además de `TestCompose_VariantCodex` (`:213`), que busca la palabra "Codex". El golden de Codex que cubre el asset legado ya existe; falta el que cubra el contrato de evidencia. Con 03 y 04 la composición cambia (capas y módulos: `ComposeLayers`, `ComposeCodexPackage`), así que el lugar exacto de las pruebas se reconfirma en 3.1 contra la rama integrada. La batería de recibos de ejecución de prueba se aplica a ambos agentes, en ejecución directa y delegada, y los veredictos tienen que coincidir.

### D8 — `verify` en Codex: cubierto por 04 (en `main`)

En `origin/main` `d5e95bd`, el asset legado de Codex no lista `verify` (`internal/harness/config/assets/codex/sdd-orchestrator.md:38-50`, el flujo, y `:85-92`, la tabla de acciones), mientras el de Claude sí lo incluye (`.../claude/sdd-orchestrator.md:71`, `openspec-verify`).

**04 lo cubre**, en `main` (PR #9) y archivada en `openspec/changes/archive/2026-11-13-codex-native-opsx/`:

- **Spec:** `openspec/specs/codex-opsx-delivery/spec.md` (`main`), requisito "Verify con independencia obligatoria": la acción `verify` revisa recibos y cumplimiento y, cuando se requiere independencia, el revisor es distinto del constructor, incluso en una verificación breve (escenario «Verificación breve»); sin revisor habilitado se detiene esa verificación (escenario «Delegación desactivada o revisor ausente»).
- **SKILL:** `internal/harness/config/assets/codex/native/SKILL.md` (`main`): el párrafo de `verify` (carga la skill instalada), el del revisor (debe ser distinto del constructor cuando verify requiere independencia) y el de evidencia (las acciones que requieren evidencia cargan el contrato en ejecución directa y delegada; sin contrato corregido o captura independiente, el resultado queda sin verificar y la acción se detiene).

06 no agrega la acción ni decide quién revisa: consume ese paso para revisar los recibos de ejecución y evaluar la suficiencia (D3). Con 04 en `main`, la paridad de Codex (D7) puede probarse también sobre `verify`; la suficiencia en Codex sigue *sin verificar* hasta que se acuerde el módulo `evidence` y se entregue el contrato completo (TBD, 03 y 04).

**Lo que 04 ya consume de 06.** `assets/codex/native/SKILL.md`, sección «Partial evidence requirements», y `action-table.md` (`apply` y `verify` cargan `bounded-command-output`) consumen D2–D4 de 06 en `f62cf00` como requisitos parciales inline: integridad, procedencia y suficiencia separadas, estado exacto incluidos preparados, sin preparar y nuevos, sha256 de bytes sin normalizar LF/CRLF, detalle descartado → sin verificar, y detener solo la acción afectada. 04 declara que **no** entrega `modules/evidence.md`, un extracto final, un esquema ejecutable ni un validador, y que no inventa el mecanismo de huella del árbol. Esa huella ya no queda pendiente en 06: propuesta de 06 = TreeHash sobre `git ls-files` (`eval/evidence/state.go`, D2; medido en Windows y WSL2). Al entregarse `modules/evidence.md`, esa sección debe alinearse o reemplazarse: seguimiento de Matías (04) sobre `main`. El informe `informes/respuesta-a-04-y-05-2026-11-13.md` contesta punto por punto lo que 04 pedía antes de archivarse.

### D9 — Hooks: complemento, no registro

Los hooks del cliente (08) pueden dejar rastro de ciertas ejecuciones y sirven para corroborar. **Complementan la evidencia pero no son un registro exhaustivo**: no cubren todas las interfaces ni todos los comandos, y su silencio no prueba que algo no se ejecutó. La ausencia de un rastro en un hook no verifica ni rechaza un recibo de ejecución por sí sola. Qué fuente cubre cada interfaz lo demuestra 05 (D3).

### D10 — `eval/evidence/`: código Go nuevo, fuera del binario (PROPUESTA, pendiente de confirmación)

Esta propuesta **ya no es solo texto**. `eval/evidence/` es código Go nuevo que ejecuta el contrato sobre fixtures: el esquema del recibo (`receipt.go`), los veredictos de D3 y D4 (`judge.go`) y la huella del estado desde un worktree de Git (`state.go`), con 22 casos en `testdata/casos/`.

- **Fuera del binario:** no lo importa ningún `cmd/` ni `internal/`. Sigue el patrón de `eval/baseline` de 02 (D6 de 02).
- **Dependencias:** solo la biblioteca estándar, `eval/baseline/record` (`Fingerprint`) y `eval/baseline/fixture` (`TreeHash`, solo en las pruebas) de 02. No toca `go.mod`.
- **No es el verificador del producto.** Es la forma ejecutable del contrato para que los fixtures fijen qué veredicto exige.
- **Gobernanza:** `CLAUDE.md` no clasifica `eval/`. Propuesta de 06, **pendiente de confirmación: MEDIO** (implementar con checkpoints), por analogía con `harness/*`. Si el grupo lo clasifica de otro modo, se ajusta.
- **Aviso pendiente:** a Juan Cruz (01) y a Giuliano (02), dueños de `eval/baseline`, porque `eval/evidence` depende de dos de sus paquetes (tarea 1.9).

### Respuesta de 06 a 05 (propuesta, pendiente de confirmación de Gianfranco)

05 pide a 06 cuatro respuestas (05: D9, D19, D20 y «Puntos de grupo y su estado»). Lo que sigue es **propuesta de 06, pendiente de confirmación de Gianfranco**; no es un acuerdo del grupo.

- **(a) Ubicación y acceso.** Se propone aceptar `<UserCacheDir>/active-stack/runs/<change>/<id de corrida>/` y el acceso local de 05 D9.
- **(b) Vencimiento.** Se propone aceptar archivo del change más siete días como valor provisorio (05 lo declara supuesto sin medición, D9). 06 solo exige que la referencia informe el vencimiento (D6); el estado de la copia vencida sigue en D4.
- **(c) Lista sensible y registro de captura.** Se propone aceptar la lista sensible de 05 D14 y el registro de captura de D9. 06 no agrega categorías ni campos.
- **(d) Contrato en `apply` y `verify`.** 05 propone que se cargue el contrato `bounded-command-output` en `apply` y `verify`, directo o delegado (D19). Coincide con el módulo `evidence` de D5; sigue **pendiente de acuerdo con 03 y 04**.

## Risks / Trade-offs

- **Recibos de ejecución solo formales.** Un worker puede completar los campos y fabricar la huella. Mitigación: sin corroboración independiente o reejecución queda *sin verificar* (D3, D4). No lo elimina; ver 08 y la propuesta de juez externo.
- **Cumplimiento solo por instrucción.** El contrato depende de que el agente lo siga. Mitigación: pruebas de aceptación; controles reales en 08.
- **Fuente independiente no demostrada.** Hasta que 05 demuestre qué fuente cubre cada interfaz, gran parte de lo declarado queda *sin verificar* o exige reejecución.
- **Suficiencia subjetiva.** Mitigación: cada requisito nombra su evidencia; lo no evaluable queda *sin verificar*.
- **Filtración de datos.** Mitigación: solo se persiste la copia redactada y ninguna salida completa entra al repo (D6).
- **Divergencia entre assets.** OpenCode ya pasó por una auditoría de paridad por una divergencia silenciosa. Mitigación: golden también para Codex.
- **Reejecutar** da la garantía más fuerte pero duplica costo y puede fallar en otro entorno. Es una de las dos vías de corroboración; cuándo es obligatoria queda (TBD).
- **Huella del estado con cambios sin confirmar:** medida en Windows y Linux (WSL2) con `core.autocrlf` en tres valores; no se probó en Linux nativo ni en macOS, y un archivo ignorado por Git no entra en la huella.
- **Módulo `evidence` sin acuerdo.** Si 03 o 04 no aceptan la propuesta de D5, el contrato queda en `delegation` y la ejecución directa con el toggle apagado queda sin contrato; en ese caso la acción se detiene (D5) y se reporta como límite.

## Migration Plan

No hay migración de datos: el contrato es texto nuevo en los assets, más `eval/evidence/` (código fuera del binario, D10). Quien reinstala el orquestador recibe la sección reescrita de forma idempotente, y desinstalar sigue siendo limpio. Los changes en curso no tienen recibos de ejecución; se tratan como *sin verificar* hasta que los generen.

## Open Questions

Se eliminaron las resueltas o pasadas a otro change: *agentes incluidos* (roadmap, `CHANGES.md:3` y `:182`), *plataformas mínimas* (decidido, D2), *verify en Codex* (cubierto por 04, D8), *dueño del sobre de retorno* (07, D6), *algoritmo de la huella del archivo* (sha256, D2) y *redacción y retención de las salidas* (05).

1. **Módulo `evidence` y texto del extracto `## Evidence Contract`** (PROPUESTA D5, TBD con 03 y 04): aceptar o ajustar los ocho cambios de la lista de D5. Decide: el grupo, con Juani (03) y Matías (04). 06 redacta el texto.
2. **Mecanismo de captura y fuente independiente por interfaz** (arnés o fuente del cliente, antes del truncamiento). Lo define Emiliano (05); el estado actual es la matriz D22 de 05 (D3). Lo consume Gianfranco (06).
3. **Quién evalúa la suficiencia.** Abierta. 04 y 05 (D20) esperan de 06 un criterio identificable. Propuesta de partida de 06, pendiente de confirmación: el revisor distinto del constructor de 04 (D8) y que cada requisito nombre su evidencia. Decide: el grupo; lo que 07 pueda exigir depende de esto.
4. **Huella del estado probado: conjunto de archivos y mecanismo** (D2). Medido en Windows y Linux (WSL2): propuesta de 06 = TreeHash sobre los bytes de la copia de trabajo, índice temporal descartado. Pendiente de confirmación de Gianfranco con Emiliano (05) y de repetir en Linux nativo o CI.
5. **Ubicación, acceso y vencimiento de la copia redactada, y qué se versiona** (D6). Ubicación, acceso y vencimiento: propuestos por 05 (D9), 06 propone aceptar; pendiente de confirmación. Qué se versiona: (TBD). Decide: el grupo, entre Emiliano (05) y Gianfranco (06).
6. **Que la salida redactada alcance como evidencia** del código de salida (D3). Decide: Gianfranco con Emiliano (05), con un fixture.
7. **Casos fuera de la ficha:** directorio ajeno y copia vencida, ¿rechazado o sin verificar? (D4, TBD con 05).
8. **Por qué el reporte de `verify-hook` quedó con cero checks** en el caso del 02/11 (no investigado). No es una decisión: es la tarea 1.5, de Gianfranco; el resultado informa a 01, 04 y 08.
