# propose-1 · plan-reviewer · codex-native-opsx

- **Autor / rol**: plan-reviewer (subagente nativo `role-plan-reviewer`, Claude, modelo declarado por el entorno: claude-opus-5-5). Es la misma familia que el planner, así que **no hay independencia de proveedor**.
- **Revisión revisada**: artefactos sin trackear de `openspec/changes/codex-native-opsx/` (proposal, design, tasks y las dos specs), tal como estaban en disco el 2026-11-06 entre las 13:19 y las 13:30Z. HEAD `1b4639b537a4d0c04dc8c25c0220732095a772d3`. Código citado: `bd63543`. `git diff --quiet bd63543 HEAD -- internal cmd` → IDENTICAL (re-ejecutado).
- **No hice**: ediciones a los artefactos ni al código, build, commit ni `go test`. No compilé nada: lo prohíben la regla del operador y el brief.
- **KB / L-NNN / skill registry**: no existen en el repo (según el brief y el planner; no lo re-verifiqué).

## EVIDENCIA

| # | Timestamp (UTC) | Comando / observación | Salida relevante |
|---|---|---|---|
| R1 | 2026-11-06T13:20:43Z | `git rev-parse HEAD; git diff --quiet bd63543 HEAD -- internal cmd && echo IDENTICAL` | `1b4639b…`; `IDENTICAL` |
| R2 | 13:19:41Z (orquestador, reutilizado); 13:29:09Z (re-ejecución propia, artefactos sin cambios) | `openspec validate codex-native-opsx --strict` | `Change 'codex-native-opsx' is valid`, exit 0. Valida el formato, no el contenido. |
| R3 | 2026-11-06T13:25:23Z | Sonda propia e independiente: `openspec init --tools codex --no-animation .` en `scratchpad/review-probe/proj`, con HOME, USERPROFILE, CODEX_HOME, APPDATA, LOCALAPPDATA y XDG_CONFIG_HOME aislados | Reproduce la E4 del planner: "6 skills in .agents/"; "Commands skipped for: codex (uses skills)"; "Start your first change: $openspec-propose … (Codex CLI or IDE); in the Codex desktop app, select openspec-propose from Skills in the sidebar". Versión 1.16.0. La sonda del planner (`scratchpad/osprobe`) coincide con lo que afirma: 6 `SKILL.md`, `.openspec-target`=`codex`, `home/.config/openspec/config.json` y `metadata.generatedBy: "1.16.0"`. El exit 1 del comando vino de `which codex`. OpenSpec avisa que envía telemetría anónima. |
| R4 | 2026-11-06 | `which codex` | No encontrado. Ningún comportamiento del cliente Codex fue ejecutado. |
| R5 | lectura | `backup/restore.go:35-40`, `:83-84`; `snapshot.go:117-160` | **Confirma D8.** `Existed && IsDir` → NO-OP en las ramas comprimida y plana. El snapshot no recorre directorios: un archivo dentro de un directorio preexistente solo se restaura si tiene su propia entrada. |
| R6 | lectura | `install/plan.go:141` (`…/.active-stack/backups/install`), `install/steps.go:35-36`, `backup/snapshot.go:43-82`, `harness/config/install.go:35-36` (`backups/sdd-orchestrator-<agent>`), `uninstall/plan.go:74` | Los directorios de snapshot son **fijos**: cada corrida reescribe `manifest.json` y `snapshot.tar.gz` en el mismo directorio. |
| R7 | lectura | `cmd/active-stack/update.go:362-368` (`installRuns`) | Un `update` con recibo puede lanzar **dos** `install` seguidos (modo más Custom extra), y cada uno reescribe el snapshot de R6. |
| R8 | lectura | `filemerge/section.go:127-139` (`locateSection`, gana el prefijo actual), `:62-79`; `harness/config/inject.go:34-72` (la purga salta IDs propios); test existente `filemerge/section_test.go:122-131` (`TestInjectMarkdownSection_PrefersCurrentPrefixWhenBothExist`); `verify/harness_checks.go:106,120` | Si coexisten `active-stack:sdd-orchestrator` y `jr-stack:sdd-orchestrator`, se reemplaza solo el actual y **el `jr-stack:` queda**. El test lo codifica como esperado, y verify cuenta solo el prefijo actual. |
| R9 | lectura | `agents/antigravity/adapter.go:61-62` vs `agents/codex/adapter.go:47-48` | En Project, Codex y Antigravity resuelven el **mismo** `<root>/AGENTS.md` y el **mismo** `<root>/.agents/skills`. |
| R10 | lectura | `headless/flags.go:134-145` (install no tiene `--project`); `headless/starter_params.go:85-104` (`Target: model.Project` solo vía `starter add`); `grep sdd-orchestrator harnesses.yaml` → :65, :89, :112 (ningún starter lo incluye) | `active-stack install --project` **no existe**. El objetivo de proyecto del `sdd-orchestrator` no se alcanza hoy desde la CLI. |
| R11 | lectura | `uninstall/types.go:48-62`, `uninstall/plan.go:150-241`; grep de `PathsFor|Project|Target` en código no test → nada; `uninstall/steps.go:185-187, 199-205` | Uninstall **no tiene** wrapper de proyecto: es solo máquina. `skillRemovalStep` hace `os.RemoveAll(SkillsDir/<id>)`. |
| R12 | lectura | `internal/update/skills.go:97` | `scanSkills` filtra `h.Type != model.HarnessSkill` → `refreshSkills` nunca toma `sdd-orchestrator` (tipo config). Resuelve la [SUP] de D9. |
| R13 | lectura | `harness/config/testdata/compose_all_toggles_{claude,opencode}.golden`; `compose_test.go:13-54` (flag `-update`) | Ya existe un golden de Compose para claude y opencode. |
| R14 | lectura | `agents/factory.go:16-46` | Solo hay 4 adapters (claude, codex, opencode, antigravity). gemini, cursor, windsurf y generic son variantes de asset. |
| R15 | lectura | `filemerge/writer.go:16-45` | `WriteFileAtomic` crea el archivo aunque el contenido sea vacío y hace `chmod 0755` del directorio padre. Con G7 = "sin arranque" sobre un home sin `AGENTS.md`, se crearía un `AGENTS.md` vacío (razonado, no ejecutado). |
| R16 | `git check-ignore -v …/proposal.md` → exit 1; `git status --short` → `?? openspec/changes/codex-native-opsx/` | Los artefactos están sin trackear, **no ignorados**. |
| R17 | conteo | `wc -c` de los assets (6844/1275/6091/3769), `compose.go:96-102`, cola `-->\n` de la plantilla, `.gitattributes` `*.md eol=lf`, 0 CR | La aritmética de 18.010 B es consistente (Compose no se ejecutó). El borrador del arranque mide 273 B. |

### Muestreo de citas archivo:línea (todas las que sostienen decisiones, más una muestra)

Correctas:

- Adapter Codex: `adapter.go:18-20, :22-24, :26-28, :44-53, :64-66`; `codex/adapter_test.go:56-67`.
- Plantilla Codex: `:3`, `:17-36`, `:24-25`, `:55-67`, `:63-67`, `:80`, `:87-92`, `:94`, `:100`, `:129-159`, `:169-181`, `:174-179`.
- Backup: `restore.go:37-40`; `snapshot.go:117-160`.
- Config: `install.go:43-61, :63-78`; `inject.go:85-87, :107-111, :119, :122, :125`; `types.go:8-28`; `assets.go:29-31`.
- Filemerge: `section.go:28-30, :127-139, :155-175`.
- Modelo y catálogo: `model/harness.go:68-89`; `harnesses.yaml:16-22, :68, :70`.
- Install y verify: `plan.go:141-160, :307-310, :556-575`; `main.go:333-346`; `harness_checks.go:13, :97-131, :100-102, :240-279`.
- Uninstall y update: `uninstall/plan.go:162-176, :229-241`; `update.go:353-368, :375-377, :399-405, :416-418`.
- Fragmentos: `governance.md:3`; `engram-protocol.md:4`; `claude/sdd-orchestrator.md:109`.

Incorrectas:

- El "wrapper equivalente de `internal/uninstall`" (D2, tarea 3.3) no existe (R11).
- "Ignorados por git" en tasks.md, sección Orquestación, es falso (R16).

## JUICIO

Autor: plan-reviewer. Comparé el plan contra el código en `bd63543`, contra CHANGES.md (ficha 04, "Cómo trabajar", "Comportamiento objetivo", "Aceptación transversal", "Límites compartidos" y las fichas 01, 03, 05 y 06) y contra CLAUDE.md §2 a §4.

- **Calidad general (confianza alta).** El recorrido está bien trazado y las citas se sostienen. La contradicción central está demostrada (plantilla :3 frente a adapter :64-66). La sonda es reproducible (R3).
  - El riesgo D8 es real (R5), y el plan lo cubre: D8, la tarea 6.1 (cada archivo como entrada propia y dir hints para directorios nuevos) y la tarea 6.2 (rollback end-to-end con una skill del usuario en un directorio preexistente).
  - El rigor HIGH está presente: regresiones de consumidores en 2.x, 7.3, 8.x y 9.1, y mutantes en 9.2.
- **Defectos de fondo (confianza alta).**
  - El plan promete una reversión explícita que el mecanismo reutilizado no retiene (R6, R7).
  - Deja sin cubrir caminos que conservan el bloque de 18 KB o pisan a otro agente (R8, R9).
  - Tiene dos errores de factibilidad sobre consumidores (R10, R11).
- **¿Es construible tal como está?** No.
  - G1–G11, D6 a/b y el spike condicionan D4/D5/D6, los specs y la aceptación. El planner hizo bien en no decidirlas y en bloquear `apply`.
  - Pero CHANGES.md "Cómo trabajar" §4 exige resolver **antes de compartir la propuesta final** las decisiones que cambian alcance, diseño o aceptación, y reserva `(TBD)` para excepciones no bloqueantes *acordadas por el grupo*.
  - Hoy la propuesta es un borrador previo a las decisiones.

### Juicio por decisión de grupo

| ID | ¿Cambia alcance/diseño/aceptación? | ¿Admite `(TBD)` no bloqueante? | ¿Recomendación fundada? |
|---|---|---|---|
| G1 superficies | Sí (aceptación 11.x y docs) | No | Parcial. CLI es [DOC]. IDE y app tienen [EV indirecta] por el texto de OpenSpec (R3). Cloud es [SUP]. |
| G2 arranque del coordinador | Sí (D4) | No; depende del spike | [SUP]. Solo hay evidencia indirecta de la sintaxis `$skill`. |
| G3 fallback | Sí | Puede acordarse ya | **EV** (R3). Es la de menor riesgo. |
| G4 universal vs módulo | Sí (toda sesión; CHANGES.md: "las reglas generales obligatorias siguen disponibles") | No | La tensión es EV (`governance.md:3`, `engram-protocol.md:4`). La recomendación depende de 02, que no existe: [SUP]. |
| G5 módulos y ubicación | Sí (D5, 03/05/07) | No | [SUP] sin etiquetar ("no se listan como skills"). |
| G6 especialistas | Sí (D6, 03/06/07) | No | Depende de 03, que no existe. El fallback se funda en EV (:87-92). |
| G7 presupuesto | Sí (verify y spec) | No, pero es trivial | El borrador mide 273 B (R17). Falta el caso de archivo inexistente (R15). |
| G8 versión de Codex | Sí | No | Se superpone con la decisión de 01 sobre versiones y entornos: debería referenciarla. |
| G9 model routing | Sí (contenido de assets) | Puede acordarse ya | **EV** (:174-179). |
| G10 objetivo de proyecto | Sí | No | **No fundada.** Es inalcanzable desde la CLI (R10) y colisiona con Antigravity (R9). |
| G11 sin recibo | Sí (update) | El aviso blando sí | **EV** (`update.go:375-377`; R12). |
| D6 a/b | Sí (b toca `config.toml` compartido con 01/08) | No | [SUP], depende del spike. |
| G12 (nuevo) retención del punto de reversión | Sí (B1) | No | EV (R6, R7). |

## SUPOSICIÓN

- **No observé ningún comportamiento del cliente Codex** (R4). Todo lo del cliente es SUP o [DOC] heredado del brief del planner, sin acceso propio: carga de `AGENTS.md` en sesión y subagentes, descubrimiento de skills y roles, sintaxis `$sdd-orchestrator`, clave de invocación implícita, esquema TOML de roles, spawn, feature flags, superficies y telemetría. No consulté Context7 ni la web.
- **No ejecuté `go test` ni ningún código.**
  - R8 (bloque dual), R9 (colisión) y R15 (archivo vacío) son razonados desde el código, no reproducidos en ejecución.
  - R8 tiene además un test existente que codifica el comportamiento, pero no lo corrí.
- No sé cuántos usuarios reales tienen el bloque dual. `section.go:20-23` indica que hubo upgrades sin migración.
- No leí `internal/pipeline` para confirmar el orden de `Rollback()`. Asumo, por `install/steps.go:165-167, :199-203`, que cada paso restaura el manifiesto completo, lo que sería idempotente.
- No validé el informe ejecutivo externo ni 01, 02 y 03, que no existen como changes.

## Fuentes leídas

- **Artefactos** (completos): proposal, design, tasks, las dos specs, `informes/LEEME.md`, `informes/propose-1-planner.md`.
- **Proyecto**: CHANGES.md (completo), CLAUDE.md (inyectado), grep de `ARCHITECTURE.md:74-82` y de README.
- **Código: agents.** `agents/codex/adapter.go` (completo), `adapter_test.go:54-68`, `antigravity/adapter.go:15-80`, `opencode/adapter.go:85-115`, `factory.go`.
- **Código: harness/config.** `install.go:1-100`, `inject.go:25-140`, `compose.go:60-104`, `types.go`, `assets.go:26-32`, `primary_agent.go:40-46`, plantilla Codex completa, cabeceras de governance, engram y claude (:100-170), `compose_test.go:13-54, :205-245`, `purge_test.go:76-129`.
- **Código: catálogo.** `harnesses.yaml:1-120, :538-560`.
- **Código: backup y filemerge.** `backup/restore.go` (completo), `snapshot.go:1-160`; `filemerge/section.go` (completo), `writer.go:16-45`, funciones de `toml.go`, `section_test.go:118-131`.
- **Código: install, verify, uninstall.**
  - `install/plan.go:130-170, :260-340, :447-455, :545-585`, `install/steps.go:25-44, :165-215, :329-353`, `codex_project_internal_test.go`.
  - `verify/harness_checks.go:85-131, :238-242`, `adapter.go`, `codex_test.go:1-48`.
  - `uninstall/plan.go:150-260`, `steps.go:149-210`, `types.go:48-62`.
- **Código: CLI, update, TUI y modelo.**
  - `cmd/active-stack/update.go:345-425, :495-505`, `main.go:320-350`, `headless/flags.go:60-80, :134-145`, `headless/starter_params.go:85-104`.
  - `update/skills.go:1-140`; `tui/backups_screen.go:150-165`; `model/agentpaths.go:34-42`, `harness.go:68-88`.
- **Sondas**: `scratchpad/osprobe` (del planner) y `scratchpad/review-probe` (propia).

## Contaminación por otros informes

- Leí `propose-1-planner.md` completo antes de revisar el código y lo usé como mapa de citas. Verifiqué contra el código cada cita que sostiene una decisión y re-ejecuté la sonda (R3) de forma independiente.
- Las etiquetas [DOC] sobre Codex vienen del brief del orquestador al planner. No tuve acceso a esa fuente: no las doy por verificadas.
- No consulté Engram ni a otros revisores.

## Hallazgos

### B1 · BLOQUEANTE — La reversión explícita no sobrevive a la siguiente instalación o actualización

- **Fuente**: spec de migración, requisito "La migración se puede revertir explícitamente" y escenario "Restauración del snapshot"; design, Migration Plan §3 y "Reutilizar internal/backup … sin modificarlos"; proposal, "Una restauración explícita del snapshot vuelve al bloque anterior". Contrastado con R6 y R7.
- **Observación**:
  - El snapshot de install y el backup por archivo de InjectSection viven en directorios fijos, que se sobrescriben en cada corrida.
  - Tras la migración, cualquier reinstalación o update reemplaza el punto de reversión por el estado ya migrado. El propio escenario "Actualización repetida" lo destruye.
  - Un update con Custom extra corre dos install (update.go:362-368), así que el punto previo a la migración puede perderse dentro de UN solo update.
  - El escenario de spec ("GIVEN una instalación migrada y su snapshot previo") está construido para pasar y oculta este caso.
- **Criterio incumplido**: CHANGES.md, "la migración … admite rollback" y "migración reversible" (ficha 04); dominio ALTO de CLAUDE.md §4.
- **Impacto**: el usuario pierde en silencio la posibilidad de volver a su configuración anterior.
- **Corrección verificable**: registrar G12 como decisión humana o de grupo, con dos opciones:
  - (a) Snapshot de migración dedicado, con sello de tiempo y no sobrescribible, creado desde el paquete config solo cuando se detecta el bloque anterior, sin tocar internal/backup. Hay que comprobar que lo listan la pantalla de backups (tui_wiring.go:51) y la acción de Windows (windows_backups.go:90).
  - (b) Reformular el requisito como "revertible hasta la próxima instalación o actualización" y documentar ese límite.

  En ambos casos, agregar los escenarios "update con dos corridas conserva el punto previo a la migración" y "reinstalar tras migrar no pierde el punto de reversión", cada uno con su tarea RED.

### B2 · BLOQUEANTE (puerta de proceso) — Decisiones y aprobación abiertas

- **Fuente**: CHANGES.md "Cómo trabajar" §4 y §5; proposal.md ## Aprobación vacía; G1–G11 y D6 a/b.
- **Observación**: las decisiones abiertas cambian alcance, diseño y aceptación (ver la tabla del JUICIO). Mientras sigan abiertas, la propuesta no puede compartirse como final ni construirse (HIGH). Mi visto bueno no reemplaza la aprobación humana.
- **Corrección verificable**:
  - El humano o el grupo registra las decisiones en ## Aprobación.
  - El planner las incorpora y elimina la redacción condicional de specs y design ("o ninguna, si G7…", "según G5", etc.).
  - Después corre una nueva ronda de plan-reviewer.

### M1 · MAYOR — El bloque heredado jr-stack:sdd-orchestrator sobrevive cuando coexiste con el actual

- **Fuente**: spec de migración, escenario "Marker heredado" (solo cubre el heredado aislado); design D3. Contrastado con R8.
- **Observación**:
  - Con ambos bloques presentes (caso que describe section.go:20-23), InjectSection reemplaza el actual y deja el jr-stack:, que puede ser el de 18 KB.
  - La purga no lo toca porque el ID es propio, y verify no lo detecta.
  - La sesión normal seguiría cargando la coordinación: incumple el requisito 1 de delivery.
- **Corrección verificable**:
  - Escenario nuevo: "GIVEN active-stack: y jr-stack: sdd-orchestrator presentes WHEN install THEN no queda ningún jr-stack: y hay una sola sección actual".
  - Implementarlo en el paquete config: retirar en bucle todas las ocurrencias del ID propio bajo cualquier prefijo antes de inyectar. Así no se toca filemerge y se respeta section_test.go:122.
  - Verify detecta cualquier ocurrencia heredada.
  - Agregar un mutante en 9.2.

### M2 · MAYOR — El alcance del chequeo de títulos de firma es contradictorio

- **Fuente**: spec delivery R1 "Instalación limpia" ("el ARCHIVO no contiene…"), frente a "Bloque anterior remanente" y D10, que hablan de la sección o del arranque.
- **Observación**:
  - A nivel archivo, un "## Delegation Rules" o "## OPSX Workflow" del propio usuario hace fallar el chequeo duro y dispara rollback. Es probable en un AGENTS.md que documente OPSX.
  - A nivel sección, el chequeo no detecta M1.
  - No hay escenario con esos títulos en contenido del usuario.
- **Corrección verificable**:
  - Fijar el alcance: la sección propia más toda sección del ID propio bajo cualquier prefijo.
  - Escenario: "títulos de firma en contenido del usuario fuera de markers → verify pasa y los bytes quedan intactos".
  - Agregar el caso correspondiente en 7.1.

### M3 · MAYOR — Uninstall: el wrapper citado no existe y el patrón de borrado puede perder archivos sin posibilidad de restaurarlos

- **Fuente**: D2 ("el adaptador de uninstall implementan RolesDir"); tareas 3.3 y 8.1; spec "La desinstalación retira solo las piezas administradas". Contrastado con R11 y R5.
- **Observación**:
  - (i) internal/uninstall no tiene objetivo de proyecto ni wrapper: la tarea 3.3 modifica algo que no existe.
  - (ii) El patrón actual es RemoveAll(SkillsDir/<id>). Reutilizado para sdd-orchestrator/, borraría archivos sin marca que el usuario dejó dentro.
  - El rollback no los recupera: el directorio preexistente es NO-OP y esos archivos no tienen entrada en el manifiesto.
  - Ni install (D7) ni uninstall tienen escenario para un archivo no administrado dentro de un directorio administrado.
- **Corrección verificable**:
  - Corregir D2 y 3.3: uninstall es solo máquina; un objetivo de proyecto sería alcance nuevo.
  - Especificar el borrado archivo por archivo (solo los marcados) y el del directorio solo si queda vacío.
  - Escenario y RED en 8.1: "archivo del usuario dentro de sdd-orchestrator/ → intacto tras uninstall y tras rollback".
  - Agregar un mutante.

### M4 · MAYOR — El objetivo de proyecto (G10) es inalcanzable desde la CLI y colisiona con Antigravity

- **Fuente**: escenario de spec "Objetivo de proyecto" ("WHEN se instala con --project <root>"); tareas 5.3 y 3.3; recomendación de G10. Contrastado con R10 y R9.
- **Observación**:
  - install no tiene --project y ningún starter incluye sdd-orchestrator: el escenario no se puede ejecutar tal como está escrito.
  - Si se habilita, Codex y Antigravity comparten <root>/AGENTS.md (con el mismo marker) y <root>/.agents/skills. Uno pisa al otro:
    - si gana el arranque de Codex, Antigravity pierde su bloque y se incumple "La entrega de los demás agentes no cambia";
    - si gana el bloque de Antigravity, Codex vuelve a leer 18 KB.
  - Además, Antigravity vería la skill y los módulos de Codex.
- **Corrección verificable**: llevar R9 y R10 al grupo (G10).
  - Si la decisión es "sí": escenario a nivel API (install.Options{Target: Project}) más un escenario de coinstalación Codex+Antigravity con la resolución acordada.
  - Si es "no": declarar el límite y quitar 5.3 y el escenario.

### M5 · MAYOR — Falta el camino "el cliente no permite deshabilitar la invocación implícita"

- **Fuente**: spec delivery, requisito de invocación explícita ("Cuando el cliente admite…") y escenario "Pedido implícito…", cuyo GIVEN dice "deshabilitada y confirmada por el spike"; CHANGES.md "Comportamiento objetivo" ("permanece inactivo").
- **Observación**: el objetivo central depende de un supuesto [SUP] del cliente. Si el spike lo refuta, no hay escenario, alternativa ni criterio. La description de la skill podría activarla en sesiones normales.
- **Corrección verificable**:
  - Definir una rama explícita: una description que no matchee pedidos comunes, verificada manualmente con pedidos implícitos, o pasar a G2 = perfil.
  - Escenario de fallo con aceptación medible en 11.1.

### M6 · MAYOR — El mapeo toggle → módulo no está definido

- **Fuente**: requisito de módulos de la spec y escenario "Toggles completos"; D5; tareas 4.2 y 7.1 ("cada módulo de toggle activo").
- **Observación**: el catálogo tiene cinco toggles (harnesses.yaml:70). "Coordinación de subagentes" no tiene toggle, y model-routing activo no tiene módulo (G9). Verify y los tests no pueden saber qué esperar con delegation apagado o model-routing encendido.
- **Corrección verificable**:
  - Agregar una columna "toggle" a D5 y su equivalente en la spec (por ejemplo, delegation → delegación + subagentes; model-routing → ignorado en Codex hasta 09).
  - Escenarios "sin delegation" y "model-routing activo".

### M7 · MAYOR — ConfigDeliveryNativeRoles tiene nombre genérico pero despacho exclusivo de Codex, en paralelo con 03

- **Fuente**: D1, D8 y D10 (checkCodexNative, CodexManagedPaths, "instalador Codex nativo"); CHANGES.md (03 y 04 en paralelo sobre install y verify); install.go:43, harness_checks.go:100.
- **Observación**:
  - Si 03 reutiliza ese valor para Claude, el switch lo rutearía al instalador de Codex.
  - D1 rechaza "if agent == codex" por acoplamiento, pero el valor elegido acopla de forma implícita.
  - La coordinación con 03 recién está prevista en 11.3, después de implementar.
- **Corrección verificable**:
  - Nombre específico (ConfigDeliveryCodexNative) o contrato neutral en el que el adapter provee el paquete y las rutas.
  - Registrar el acuerdo de código compartido con 03 antes de apply, como bloqueo de 3.1.

### M8 · MAYOR — El spike alimenta los specs, pero solo se prevé actualizar el design y la re-revisión es condicional

- **Fuente**: tarea 1.3 ("Actualizar design.md … si cambia el enfoque, devolver…"); escenarios de spec con "lo que fije el spike"; D6 solo tiene ramas a y b.
- **Observación**: cada resultado del spike convierte supuestos en requisitos. Además falta la rama (c), "ninguna superficie admite roles nativos": con ella, "SHALL registrar cada especialista como rol nativo" sería incumplible.
- **Corrección verificable**:
  - 1.3 actualiza specs, design y tareas, y la re-revisión del plan pasa a ser OBLIGATORIA.
  - Agregar D6-c con el comportamiento de install y verify.

### Menores

- **m1 · MENOR** — D9 y la tarea 8.3 dejan como [SUP] que refreshSkills no toca sdd-orchestrator. R12 lo prueba con update/skills.go:97. Convertirlo a [EV] y conservar el test de regresión.
- **m2 · MENOR** — La tarea 2.1 crea un testdata/compose-golden/ nuevo e ignora los goldens existentes (R13). Además, "-run Golden" no matchea TestCompose_AllToggles*. Solo hay 4 adapters (R14): gemini, cursor, windsurf y generic son variantes de asset. Reutilizar y extender los goldens (agregar antigravity) y aclarar lo de las variantes.
- **m3 · MENOR** — El RED de la tarea 3.1 ("exige el nuevo valor") es necesariamente una falla de compilación, y el propio plan dice que eso no cuenta como RED. Reformularlo como RED de comportamiento (despacho, tarea 3.2) o declarar 3.1 como cambio de tipo cubierto por 3.2.
- **m4 · MENOR** — El orden del Migration Plan ("preflight → snapshot") contradice el pipeline: el snapshot corre en Prepare (plan.go:149-160) y el preflight D7 en config.Install (Apply). Mover el preflight a Prepare, o corregir el texto y aclarar que "ningún archivo modificado" se garantiza por rollback.
- **m5 · MENOR** — La tarea 6.2 sugiere el seam configInstallFn, pero reemplazarlo saltea las escrituras reales y no simula "falla al escribir un rol después de reescribir AGENTS.md". Especificar un seam de escritura dentro del instalador nativo.
- **m6 · MENOR** — D5 pone rutas absolutas en developer_instructions. En objetivo de proyecto, eso versiona en el repo la ruta del home y el nombre de usuario. Decidir entre rutas relativas o con ~.
- **m7 · MENOR** — Hay afirmaciones de Codex presentadas como hecho sin etiqueta: el esquema TOML name/description/developer_instructions (proposal "What Changes" y requisito de especialistas) y "no se listan como skills en la sesión normal" (D5). Etiquetarlas [DOC] o [SUP] y condicionarlas al spike.
- **m8 · MENOR** — G8 se superpone con la decisión de grupo de 01 sobre versiones y entornos. Referenciarla en vez de duplicarla.
- **m9 · MENOR** — tasks.md (Orquestación) dice "ignorados por git", pero los artefactos están sin trackear (R16). Además, la rama es feat/codex-native-opsx y CHANGES.md "Cómo trabajar" §2 pide propose/<nombre>. Corregir el texto y registrar o justificar la desviación.
- **m10 · MENOR** — Faltan mutantes en 9.2 para tres aserciones: la migración del marker heredado (5.2, y M1), "update sin recibo no modifica AGENTS.md" (8.3) y "design-mockups intacto" (5.2/8.1).
- **m11 · MENOR** — Con G7 "sin arranque" sobre un home sin AGENTS.md, el camino reutilizado crearía un archivo vacío (R15, razonado). Agregar el escenario y su resultado esperado.
- **m12 · MENOR (observación preexistente, fuera de alcance)** — WriteFileAtomic hace chmod 0755 del directorio padre (writer.go:39-42), lo que afecta a ~/.codex/ y a los directorios de roles nuevos. Derivarlo a 01/08 sin corregirlo acá.
- **m13 · MENOR** — El escenario "Cambio de toggles" no nombra su disparador. Los toggles no se exponen al usuario (CLAUDE.md §2): el disparador real es un cambio de catálogo entre versiones, vía update.

### Cobertura de los escenarios de la ficha (foco 4)

| Escenario | Estado |
|---|---|
| Sesión normal sin bloque SDD | Presente (spec R1, manual 11.1). Huecos M1 y M2. |
| Invocación explícita | Presente (manual). Falta el camino de fallo (M5). |
| Especialista con sus reglas | Presente. Falta D6-c (M8). |
| Acción con TDD | Presente. |
| Módulo faltante → informa y detiene | Presente (manual más verify 7.1). |
| Migración idempotente | Presente. Hueco M1. |
| Rollback con contenido del usuario | Presente (6.2, D8 cubierto). La reversión explícita falla (B1). |
| Update desde instalación anterior, con y sin recibo | Presente. B1 en la variante de dos corridas. |
| Uninstall | Presente. Ver M3. |

### Rigor HIGH en las tareas

- **Regresión de consumidores**: presente (2.1, 2.2, 7.3, 8.2, 8.3, 9.1).
- **Mutantes**: presentes en 9.2, con las faltas que señala m10.
- **RED observable**: en general sí, salvo m3 y m5.
- **Dependencias infactibles**: M3 (i) y M4.
- **Dependencias abiertas**:
  - 02, que no existe, bloquea 11.1.
  - 03, que tampoco existe, bloquea la revisión conjunta 11.3. `openspec list` solo muestra este change.

## Disposición propuesta

| Hallazgo | Disposición |
|---|---|
| B1 | Corregir en plan + decisión humana o de grupo (G12) |
| B2 | Decisión de grupo + aprobación humana |
| M1, M2, M3, M5, M6, M8 | Corregir en plan |
| M4, M7 | Decisión de grupo, después corregir en plan |
| m1–m11, m13 | Corregir en plan |
| m12 | Aceptar como observación y derivar a 01/08 |

## Veredicto

**PLAN APTO CON CORRECCIONES.** No es apto para apply ni para compartirse como propuesta final.

- El diseño base es sólido, el riesgo D8 está cubierto y las citas se sostienen.
- B1 y M1–M8 vuelven al planner.
- Las decisiones G1–G11, más G12 y D6 a/b, y la aprobación humana fechada siguen cerradas (B2).
- Corregido el plan, hace falta una nueva ronda de plan-reviewer.
- Sin esa ronda y sin ## Aprobación completada por el humano, la construcción no empieza.
