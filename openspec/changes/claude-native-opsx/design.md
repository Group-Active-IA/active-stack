# Design: claude-native-opsx

Motivación en proposal.md (Why). Requisitos en `specs/claude-instruction-layers`,
`specs/claude-coordinator-skill` y `specs/claude-orchestrator-migration`. Cambios de la revisión
del 2026-11-11 en `revision-2026-11-11.md`.

## Context

### Recorrido de InstallSDD para Claude (estado en 35a89f8)

1. `internal/harness/config/install.go:21` `Install` recibe el harness `sdd-orchestrator`
   (catálogo: `internal/catalog/harnesses.yaml:65-72`, toggles `tdd, engram, model-routing,
   delegation, governance`, modos `lite, full`, agentes `claude, codex, opencode, antigravity`).
2. `install.go:43` decide por `adapter.ConfigDelivery()`. Claude devuelve
   `ConfigDeliveryInstructions` (`internal/agents/claude/adapter.go:125-127`), así que va a la rama
   `default` (`install.go:62-76`).
3. `Compose` (`internal/harness/config/compose.go:20-34`, función pura) arma: base
   `assets/claude/sdd-orchestrator.md` (8.899 B); quita `sdd-delegation` y `sdd-model-assignments`
   si el toggle está inactivo; agrega `governance.md` (1.275 B), `engram-protocol.md` (6.091 B) y
   `strict-tdd.md` (3.769 B) si están activos. Con los cinco toggles el bloque mide 20.065 B.
4. `Inject` → `InjectSection(path, "sdd-orchestrator", ...)` (`inject.go:85`, `:101`): snapshot si
   existe, `PurgeStaleSections` (borra toda sección `active-stack:` no registrada como propia,
   `inject.go:66-73`), inyección por markers y `WriteFileAtomic`.
5. Ruta: Machine `~/.claude/CLAUDE.md` (`adapter.go:28-30`); Project `<root>/.claude/CLAUDE.md`
   (`adapter.go:91-98`, resuelto por `internal/install/plan.go:369-372`).
6. Verify: `checkConfig` cuenta el marker una vez (`internal/verify/harness_checks.go:97`);
   la variante primary-agent está en `:275` (líneas corridas por los checks que agregó 01). Uninstall: `internal/uninstall/steps.go:150-170`
   quita la sección y, sólo para `ConfigDeliveryPrimaryAgent`, la entrada JSON.
7. Skills: el adapter ya resuelve `SkillsDir` (`adapter.go:32-35`, `<home>/.claude/skills`; Project
   en `:95`). `refreshSkillDirectory` (`internal/harness/skill/idempotent.go:17-71`) prepara la
   skill en un directorio temporal, la compara, toma snapshot y la activa con `rename`. **No
   comprueba propiedad**: si el usuario tiene una skill propia con el mismo ID, la reemplaza
   (con snapshot).

### Acuerdo de activación (informe unificado, 2026-11-10)

"03/04. Invocación explícita en la sesión actual mediante skill. En Claude, `/sdd-orchestrator`."
La auditoría técnica del mismo día lo detalla: skill con `disable-model-invocation: true`, que
"guía el flujo y carga sus módulos". Esta revisión reescribe el diseño sobre ese acuerdo. La
versión anterior (agente de sesión con `claude --agent`) queda como alternativa descartada.

### Comportamiento documentado de las skills de Claude Code

Fuentes: [Skills](https://code.claude.com/docs/en/skills) y
[Subagentes](https://code.claude.com/docs/en/sub-agents), leídas el 2026-11-11. Son SUPOSICIÓN
hasta que las tareas 1.x las comprueben en la versión fijada.

| Hecho | Consecuencia para el diseño |
|---|---|
| Ubicación: `~/.claude/skills/<id>/SKILL.md` (personal) y `.claude/skills/<id>/SKILL.md` (proyecto). Con el mismo nombre, "personal over project". | En target Project, una skill Machine con el mismo ID tapa a la del proyecto (D5, colisiones). |
| `disable-model-invocation: true`: Claude no la invoca solo, su descripción no entra al contexto, el usuario sí la invoca, y no se puede precargar en subagentes. | La sesión normal no se vuelve coordinador por un pedido OPSX, y el coordinador no puede correr como subagente. Desaparece el problema B1 de la versión anterior (G9). |
| Sin `context: fork`, el contenido entra a la conversación como un mensaje y queda ahí en los turnos siguientes. | El coordinador sigue en el hilo actual. No hay forma documentada de "salir" del rol salvo una sesión nueva o `/clear` (D3). |
| Tras la compactación automática, Claude Code vuelve a adjuntar la invocación más reciente de cada skill "keeping the first 5,000 tokens of each", con 25.000 tokens entre todas. | El `SKILL.md` del coordinador tiene que entrar en ese presupuesto, con lo crítico primero, y los módulos van aparte (D10). |
| Los subagentes reciben su propio prompt más `CLAUDE.md`; no ven la conversación ni las skills invocadas por el padre. Explore y Plan arrancan sin `CLAUDE.md`. | El rol de coordinador no llega a los especialistas. Los módulos les llegan sólo si el coordinador los pega en el encargo. Para Explore y Plan, el encargo repite las reglas obligatorias (igual que propone 05 D21). |
| Un subagente puede lanzar subagentes hasta tres niveles (configurable). | Corrige una suposición de la versión anterior. La regla "no delegar más abajo" es del módulo de delegación (07), no del mecanismo. |
| `allowed-tools` en una skill sólo preaprueba herramientas; no restringe. La allowlist `Agent(tipo)` sólo aplica a un agente que corre como hilo principal con `--agent`. | La skill no puede limitar qué tipos de subagente se lanzan. Esa restricción de la versión anterior se pierde (D3). |
| Archivos de apoyo en el directorio de la skill se cargan a demanda; `${CLAUDE_SKILL_DIR}` resuelve su ruta. | Los módulos por acción viven como archivos de apoyo de la skill (D4). |

### D0. Evidencia de contexto efectivo (EVIDENCIA, 2026-11-06)

Claude Code 2.1.318, modelo sonnet, proyecto temporal con `CLAUDE.md` que contenía el testigo
aleatorio `CANARIO-ceea159f1478`, más el `~/.claude/CLAUDE.md` real del usuario:

| Caso | Invocación | Vio CLAUDE.md de proyecto | Vio CLAUDE.md global |
|---|---|---|---|
| A | `claude -p --tools ""` | sí | sí |
| B | `--agents '{sonda...}' --agent sonda` (hilo principal) | sí | sí |
| C | subagente definido por `--agents` JSON, sin herramientas, lanzado con Agent | sí (transcripción `subagents/agent-*.jsonl`) | sí |

En C el prompt del subagente (355 caracteres) no contenía los testigos. Conclusión: todo lo que
queda en `CLAUDE.md` llega a sesión, coordinador y especialistas. Esto fija el diseño: `CLAUDE.md`
sólo puede contener reglas verdaderamente universales. El caso B ya no corresponde a la entrega
elegida; la tarea 1.1 repite la medición con la skill.

### Precedente en el repo

`InjectPrimaryAgent` (`internal/harness/config/primary_agent.go:25-89`) resolvió lo mismo para
OpenCode: registra el orquestador en `opencode.json` y purga la sección del archivo compartido
(`purgeInstructionsSection`, `:96`) con snapshot previo. Su comentario (`:29-34`) documenta la fuga
("leaks the orchestrator into every agent").

### Lo que dejó 01 en `main`

`codex-strict-compat` agregó `Manifest.GuardRollback(path, hash)` (`internal/backup/restore.go`):
durante una instalación en curso, el rollback no restaura un archivo cuyo contenido cambió desde
que el instalador lo escribió, y avisa con la ruta del respaldo. La guarda vive en memoria ("A
manually loaded snapshot has no live rollback guards"), así que cubre el **rollback durante la
operación** y no la **reversión posterior** con `uninstall --strategy restore`. El informe
unificado pide separar las dos (D5).

## Goals / Non-Goals

**Goals:**
- Tres capas con dueño: universal (en `CLAUDE.md`), coordinador (skill invocable sólo por el
  usuario), módulos por acción (archivos de apoyo de la skill).
- Continuidad del coordinador en la sesión: compactación y reanudación probadas.
- Reutilizar snapshot + purga acotada + escritura atómica del precedente de OpenCode, el
  reemplazo atómico de skills y la guarda de rollback de 01.
- Pruebas de contexto efectivo con testigo reproducibles, además de tests Go del instalador.

**Non-Goals:**
- Filtrar el contenido propio del usuario en `CLAUDE.md`: llega a todos (D0) y se documenta como
  límite, no se resuelve.
- Cambiar la entrega de OpenCode, Codex o Antigravity (Codex es Propuesta 04).
- Reescribir el contenido normativo de governance, TDD, Engram ni los textos que entregan 05, 06 y
  07: sólo se reubican.
- Definir el sobre de retorno del especialista (es de 07) o el formato de recibos (es de 06).
- Mover la sección `design-mockups`: sigue en `CLAUDE.md` en este change (ver G7).

## Decisions

### D1. Nuevo modo de entrega para Claude: skill administrada

Se agrega un `ConfigDelivery` nuevo (nombre tentativo `ConfigDeliveryManagedSkill`) y un inyector
que escribe `<skills>/sdd-orchestrator/` con:

- `SKILL.md`: frontmatter con `name: sdd-orchestrator`, `description`, `disable-model-invocation:
  true` y la marca de propiedad de Active Stack; cuerpo con la definición del rol coordinador.
- `modules/<módulo>.md`: un archivo por módulo por acción (D4), sólo los de toggles activos.

`<skills>` sale de `SkillsDir` del adapter según el target (regla del proyecto: nunca hardcodear
rutas de agente). La escritura reutiliza el patrón de `refreshSkillDirectory` (preparación en
directorio temporal, comparación, snapshot y `rename`) **más** un chequeo de propiedad antes de
reemplazar: un directorio `sdd-orchestrator` sin la marca no se toca (D5).

- Alternativa descartada (acuerdo del grupo): agente de sesión en `.claude/agents/` invocado con
  `claude --agent <id>`. Cambiaba el prompt de sistema de toda la sesión, obligaba a abrir una
  sesión aparte y dejaba el archivo descubrible como subagente (problema B1).
- Alternativa descartada: reutilizar `ConfigDeliveryPrimaryAgent`. Escribe un objeto en la clave
  `agent` de `settings.json`, donde Claude espera el nombre de un agente (C3).
- Alternativa descartada: instalar la skill con el catálogo de skills existente
  (`internal/harness/skill`). Copia archivos embebidos fijos y no compone por toggles; además no
  comprueba propiedad.

### D2. Contenido de `CLAUDE.md`: sólo bloque universal mínimo

Una sección administrada nueva (ID tentativo `active-stack:universal-rules`) que se registra como
propia (en `ownedSectionIDs` o `sectionFragments`) para que `PurgeStaleSections` no la borre en la
próxima instalación (C5). Contenido: reglas que deben valer igual para sesión, coordinador y
especialista (ej. "nunca crear `openspec/` a mano", protocolo mínimo de memoria, aviso de que los
permisos los impone `settings.json`) y la referencia para invocar el coordinador con
`/sdd-orchestrator` (05 D21). Los permisos del harness `permissions` ya son universales por
mecanismo (settings), no por instrucciones.

- Alternativa rechazada: dejar todo en `CLAUDE.md` con secciones por rol. D0 demuestra que todas
  las secciones llegan a todos; los encabezados no aíslan nada.
- Alternativa rechazada: bloque universal vacío. Pierde reglas que hoy recibe la sesión normal y
  rompe el criterio 1 del roadmap.

### D3. Invocación: `/sdd-orchestrator` en la sesión actual

El usuario escribe `/sdd-orchestrator` (con o sin argumentos, que la skill recibe por
`$ARGUMENTS`). Desde ese turno el coordinador guía el flujo en el mismo hilo. Decisión cerrada por
el grupo (ver Context).

Restricciones que dependían de `--agent` y cómo quedan:

| Antes (agente de sesión) | Ahora (skill) | Tratamiento |
|---|---|---|
| El cuerpo del agente reemplazaba el prompt de sistema por defecto. | El cuerpo entra como mensaje de la conversación; el prompt de sistema de Claude Code y `CLAUDE.md` siguen. | Precedencia frente al contenido del usuario se mide (tarea 1.3) y se documenta; el instalador no reescribe texto del usuario. |
| `tools: Agent(tipo, ...)` limitaba los tipos de subagente. | No aplica a skills; `allowed-tools` sólo preaprueba. | Se pierde. Si el grupo necesita limitar tipos, la vía es `permissions.deny` en settings (dominio de 08). Registrado como pendiente con 08. |
| El rol duraba lo que la sesión abierta con `--agent`. | El rol queda en la conversación desde la invocación. | README documenta que para salir del modo coordinador se abre una sesión nueva o se usa `/clear`. |
| El archivo de agente podía aparecer como subagente (B1). | `disable-model-invocation: true` impide que el modelo la invoque o la precargue en subagentes. | G9 y la prueba de 0 en 20 transcripciones se eliminan. Queda un escenario determinista: la skill no figura en las skills disponibles para el modelo (tarea 1.10). |
| `settings.json` `agent` como opt-in documentado. | No corresponde. | El instalador sigue sin escribir `agent`; se conserva sólo como regresión. |

- Alternativa descartada: `context: fork`. Correría el coordinador en un subagente sin la
  conversación, que es lo contrario a "continúa en el hilo actual".
- Alternativas descartadas como en la versión anterior: setting `agent`, output style, plugin y
  hooks `SessionStart` (activación implícita o canal paralelo).

### D4. Módulos por acción: archivos de apoyo de la skill

Los módulos son archivos en `<skills>/sdd-orchestrator/modules/`. El cuerpo de `SKILL.md` tiene
una tabla acción → módulos obligatorios y opcionales, con rutas relativas a `${CLAUDE_SKILL_DIR}`.
Al decidir una acción, el coordinador lee sólo esos archivos y:

- si ejecuta la acción en el hilo, los aplica;
- si delega, pega su contenido en el encargo del especialista (patrón actual de "Project
  Standards", `sdd-orchestrator.md:144-149`), porque el subagente no ve las skills del padre.

Inventario inicial (texto existente, sólo reubicado): `tdd.md` (apply), `governance.md` (según la
matriz, ver G4), `delegation.md` (cuando se delega: reglas de `sdd-delegation`, sobre de retorno
de 07 y extracto `## Evidence Contract` de 06), `model-assignments.md` (toggle), `engram-opsx.md`
(parte OPSX de Engram). Los textos que redacta 05 (lectura dirigida, resumen acotado, selección de
herramientas, reporte conciso, 05 D17) entran como módulos nuevos con el mismo mecanismo. Quién
mantiene cada texto queda en la matriz: el autor lo redacta y lo actualiza; la 03 decide dónde va
y lo instala.

La regla breve de 07 D1 (cuándo delegar) va en el cuerpo de `SKILL.md`, porque tiene que estar
disponible antes de decidir.

- Alternativa: una skill por módulo (`/sdd-tdd`, etc.). Agrega nombres al menú y el modelo podría
  invocarlas solas si no llevan `disable-model-invocation`. Queda para comparar con 04 en la matriz.
- Alternativa: un agente por especialista con su módulo en el prompt. Determinista, pero vuelve a
  poner archivos descubribles en `.claude/agents/`. Descartada por ahora.

### D5. Migración y ciclo de vida

Orden por instalación (reflejo de `InjectPrimaryAgent`, con transacción explícita):
1. Chequear versión mínima de Claude (D7). Resultado por caso en la spec
   `claude-coordinator-skill` ("Compatibilidad por versión y entorno sin Claude").
2. Chequear colisión en `<skills>/sdd-orchestrator` del target: si existe sin marca de propiedad,
   error con la ruta y nada se escribe. En target Project, si existe además una skill Machine con
   el mismo ID, advertir que la personal tapa a la del proyecto ("personal over project") y nombrar
   ambas rutas.
3. Snapshot de `CLAUDE.md` y del directorio de la skill si existe, en un único manifiesto bajo
   `<home>/.active-stack/backups/` también en target Project (no el patrón de `config.Install`,
   ver C8). `settings.json` no entra al manifiesto: este change sólo lo lee.
4. Escribir el directorio de la skill (preparación + `rename`).
5. Purgar secciones del coordinador en `CLAUDE.md` e inyectar el bloque universal (atómico).
6. **Rollback durante la operación.** Si 4 o 5 fallan, restaurar desde el manifiesto del paso 3
   con `GuardRollback` de 01 sobre cada archivo que el instalador escribió, para no pisar un
   cambio hecho por otro proceso mientras tanto.
7. Si fue la primera migración, registrar la ruta del manifiesto como **manifiesto de migración**
   y el hash de `CLAUDE.md` y de cada archivo de la skill resultantes (en `installation.json` o un
   archivo hermano; ubicación a fijar en la tarea 6.8). Reinstalaciones posteriores no lo
   reemplazan. El manifiesto se crea fijado (`Pinned`, `internal/backup/manifest.go`) para que
   `Prune` no lo elimine.

**Update.** El `update` real reemplaza el binario y reaplica la selección registrada con `--sync`
(`cmd/active-stack/update.go:94`, `:159-165`), que vuelve a correr `install --headless --home`
por agente (`update.go:353-360`, sin target Project). Machine migra por el camino de install; los
proyectos registrados con instalación Project V1 no se reescriben y reciben una advertencia con el
comando de migración explícita (G10). Un HOME V1 sin registro de selección no reaplica config
(`update.go:375-378`); este change agrega a ese mensaje la nota BREAKING de Claude y el comando que
migra. Un archivo de la skill con marca editado por el usuario se reemplaza en update, con su
versión editada en un snapshot y la ruta en la salida.

**Reversión posterior.** El CLI ya exige `--restore-manifest <path>` con `--strategy restore`
(`cmd/active-stack/uninstall_dispatch.go:30-34`). Como `GuardRollback` no persiste, para la
reversión se usan los hashes registrados en el paso 7: antes de escribir `CLAUDE.md` o un archivo
de la skill se compara el hash actual con el registrado; si difiere (el usuario editó), no se
escribe y se guarda el contenido actual en un snapshot nuevo. La confirmación para pisar igual es
un flag propio (nombre tentativo `--overwrite-edited`), no `--yes`, que ya confirma el uninstall
(`uninstall_flags_test.go:156`). Si 01 o 08 agregan antes una guarda persistente en
`internal/backup`, esta ficha la reutiliza en lugar de crear otra (coordinación pendiente, G6).
Uninstall sin restore borra sólo el bloque universal y el directorio de la skill con marca propia,
con snapshot y restauración si falla a mitad.

**Target Project en verify y uninstall.** Verify ya conoce Project (`projectVerifyAdapter`,
`cmd/active-stack/main.go:320-345`); este change agrega el check de la skill con el `SkillsDir`
que resuelve ese adaptador. Uninstall es sólo Machine por decisión vigente
(`cmd/active-stack/headless/uninstall_flags.go:50`); en Project se documenta la limpieza manual y
se registra como DEUDA (G12).

### D6. Permisos

`WriteFileAtomic` fuerza el modo del archivo (`internal/filemerge/writer.go:63`, `tmp.Chmod(perm)`
con `0o644` de los llamadores, `inject.go:125`, `primary_agent.go:83`, `:118`) y hace
`os.Chmod(dir, 0o755)` sobre el directorio padre (`writer.go:40`). Para `~/.claude/` eso puede
ampliar un `0700` del usuario. Ese chmod existe para reparar directorios que una versión vieja del
instalador dejó sin escritura (commit 6a40f17 del 2026-05-27). Regla propuesta: conservar el modo
del archivo existente; en el directorio agregar como máximo el bit de escritura del dueño. La
escritura de la skill (`os.MkdirAll(skillsDir, 0o755)`, `idempotent.go:18`) sigue la misma regla.

Coordinación (informe unificado: "01/03/08 coordinan el ciclo de respaldo y restauración y
comparten revisión para cambios en `filemerge`"): este arreglo y el de 08 D9 sobre
`MergeJSONObjects` se integran serializados, con un único revisor ALTO. Alcance (sólo rutas de esta
ficha o arreglo global del writer) es G6.

### D7. Versión mínima y plataformas

Se declara una versión mínima de Claude Code en el adapter: la primera con skills de usuario,
`disable-model-invocation` y el reenganche de skills tras la compactación (tarea 1.4). El valor es
`(TBD · tarea 1.4)`. La versión, la interfaz y la plataforma de cada prueba salen del manifiesto de
02 (D10 de 02: "única fuente de versión, interfaz, tier y permisos"); esta ficha no fija una
versión propia.

Plataformas (acuerdo 01/03/04/08): Windows y Linux como mínimo. La matriz distingue distribuidas
(`.github/workflows/release.yml`), declaradas (`CLAUDE.md:19` del repo) y probadas. Interfaces:
terminal (incluida la terminal integrada del IDE que ejecuta la CLI) y extensión de IDE, con
resultados separados; si una interfaz no expone una señal, queda "no disponible" con motivo.

Sin binario `claude` (CI, Docker de `e2e/`, instalar antes que Claude) la recomendación es escribir
la entrega por capas y advertir (G11).

**Modo legado.** Con versión insuficiente o desconocida, el bloque legado es la **composición
legado**: la concatenación, en el orden de hoy, de las partes universal y coordinador del asset
partido más los módulos según toggles, dentro de la sección `active-stack:sdd-orchestrator`. Debe
ser byte a byte igual a la de 35a89f8 para los mismos toggles (golden
`compose_all_toggles_claude.golden`, tareas 4.2 y 4.3). Un update en modo legado refresca ese bloque
de forma idempotente.

### D8. Oráculo de contexto efectivo

Afirmar ausencia porque el modelo no repite un testigo no prueba ausencia. Oráculo propuesto:
buscar el testigo en el registro de contexto de la transcripción de Claude Code. En 2.1.318 el
contenido de `~/.claude/CLAUDE.md` aparece en entradas `attachment.type=instructions` y existen
entradas `attachment.type=prompt_snapshot` (informe `informes/propose-1-plan-reviewer.md`). Para la
skill, la tarea 1.11 valida en qué entrada queda el cuerpo invocado y el reenganche tras la
compactación. Toda aserción de ausencia lleva un control positivo en la misma corrida. La respuesta
del modelo queda como control secundario.

### D9. Autenticación en el entorno de prueba

Las pruebas de contexto usan un directorio de configuración temporal (mecanismo a verificar en la
tarea 1.12) y la credencial llega sólo por variable de entorno ya presente en la sesión del
operador. Prohibido copiar archivos de credenciales al directorio temporal. Los informes registran
sólo nombres de variables, nunca valores. El directorio temporal se borra al terminar y su borrado
queda registrado. Las corridas medidas siguen además los perfiles aislados y la prohibición de
opt-ins de contenido de 02 (D7 y D10 de 02).

### D10. Continuidad: compactación y reanudación

- **Presupuesto.** El cuerpo de `SKILL.md` tiene que entrar en los primeros 5.000 tokens que Claude
  Code reengancha tras compactar. Con el asset partido, el coordinador sale de los 8.899 B de la
  base menos lo que pasa a universal y módulos; los módulos (hasta ~11 KB) quedan fuera del
  presupuesto porque son archivos aparte. Test sin modelo: el tamaño del `SKILL.md` compuesto con
  todos los toggles queda bajo el umbral (`(TBD · tarea 1.4)` en tokens, medido con el contador que
  acuerde 02). Orden: rol, reglas obligatorias y tabla acción → módulos primero.
- **Compactación.** Después de `/compact` o de la compactación automática, el testigo del
  coordinador sigue en el contexto y el coordinador vuelve a leer el módulo de la acción en curso
  en lugar de suponerlo cargado.
- **Reanudación.** Con `claude --continue` o `--resume` sobre una sesión donde se invocó la skill,
  el rol sigue activo (tarea 1.14 lo verifica). Si no sigue, la salida documentada es volver a
  invocar `/sdd-orchestrator`, y el bloque universal lo recuerda.

### D11. Control o módulo obligatorio ausente

Acuerdo 07/08 del informe unificado: si falta un control obligatorio, se frena sólo la acción
afectada; con protección equivalente comprobada se retoma; si el control es opcional, se continúa
con aviso. En Claude se comunica así:

- **Módulo obligatorio ausente** (archivo de `modules/` faltante o vacío, o toggle apagado para un
  módulo que la acción exige, por ejemplo delegación para un verify con independencia de 07 S4):
  el coordinador detiene esa acción, nombra el módulo y la ruta esperada, indica el comando que lo
  reinstala y sigue con las acciones que no dependen de él. Verify del instalador detecta el
  archivo faltante antes (tarea 7.1).
- **Control de 08 ausente** (hook, permiso o sandbox): el coordinador repite el mensaje que define
  08 para esa superficie. La 03 sólo fija dónde aparece (respuesta del coordinador en el hilo), no
  la detección.

### D12. Contratos que la 03 consume

| Contrato | Dueño | Qué hace la 03 |
|---|---|---|
| Sobre de retorno del especialista | 07 (D6) | El módulo `delegation.md` lo incluye tal cual; la 03 no define otro. |
| Campo `recibos` y `## Evidence Contract` | 06 (D2, D5) | Se incluye en `delegation.md` y en el encargo de acciones que ejecutan comandos, también cuando la acción corre en el hilo. |
| Textos de módulos de contexto | 05 (D17) | Se ubican como módulos nuevos según la matriz. |
| Manifiesto, rúbrica y protocolo de medición | 02 (en `main`) | Las pruebas de contexto registran versión, interfaz y permisos en el manifiesto; la comparación antes y después la corre 02 (tarea 4.1). |
| Guarda de rollback | 01 (en `main`) | Se reutiliza en D5 paso 6. |
| Mensaje de control ausente y protección efectiva | 08 | D11. |

## Decisiones

Cerradas por el grupo o por el acuerdo de 2026-11-10:

| ID | Decisión | Resultado | Fuente |
|---|---|---|---|
| G2 | Cómo se invoca el coordinador | Skill `/sdd-orchestrator` con `disable-model-invocation: true`, en la sesión actual. Sin fallback instalado: con versión insuficiente rige el modo legado (D7). | Informe unificado y auditoría técnica, 2026-11-10 |
| G5 | ID del coordinador | `sdd-orchestrator` (coincide con `h.ID`, verify, uninstall y la sintaxis acordada). | Mismo acuerdo |
| G8 | Dependencia con 02 | 02 integrada en `main`; la comparación antes y después es su tarea 4.1 con la variante de esta ficha. Los criterios funcionales no dependen de esa medición. | `main` 35a89f8 |
| G9 | Exposición del coordinador como subagente | Deja de aplicar: una skill con `disable-model-invocation: true` no se puede precargar en subagentes ni la invoca el modelo. Queda un escenario determinista (tarea 1.10). | Documentación de skills; auditoría complementaria |

Abiertas. Las specs están escritas sobre la recomendación y cada requisito dependiente lo marca
con "(depende de Gn, recomendación)":

| ID | Decisión | Opciones | Recomendación | Con quién |
|---|---|---|---|---|
| G1 | Dónde se instala cada pieza | universal en `CLAUDE.md` + skill en `.claude/skills/` según target / todo en Project / todo en Machine | según target, como hoy, con la advertencia de colisión personal sobre proyecto | grupo |
| G3 | Cómo llegan los módulos al especialista | archivos de apoyo pegados en el encargo / una skill por módulo / agente por especialista | archivos de apoyo (D4) | 04 |
| G4 | Matriz de reglas por cliente, interfaz, rol y acción | ver borrador abajo | borrador abajo; `governance` diverge de 05 (abajo) | 04 (co-dueña), 05 |
| G6 | Alcance del arreglo de permisos y guarda persistente | sólo rutas de esta ficha / arreglo global de `WriteFileAtomic` / diferir como DEUDA | global con tests, serializado con 08 D9 y un único revisor ALTO | 01, 08 |
| G7 | Sección `design-mockups` | queda en `CLAUDE.md` / pasa a módulo | queda en este change | grupo |
| G10 | Proyectos registrados con instalación Project V1 | advertir y migrar a pedido / migrar en `update` / ignorar | advertir con el comando de migración | grupo |
| G11 | Entrega cuando falta el binario `claude` | por capas con advertencia / legado / no instalar | por capas con advertencia | grupo |
| G12 | Uninstall con target Project | mantener Machine-only con limpieza manual y DEUDA / agregar `--project` | sin recomendación; el plan mantiene Machine-only | grupo |
| G13 | Limitar tipos de subagente que lanza el coordinador | no limitar / `permissions.deny` por tipo | no limitar en esta ficha; si se pide, lo implementa 08 | 08 |

### Borrador de matriz de reglas (G4, co-dueña con la Propuesta 04)

Una capa por fila. Cuando un bloque toca dos capas se parte en filas o lleva justificación. La
columna "Especialista" dice si el especialista la recibe y por qué vía.

| Origen | Regla o bloque | Capa propuesta | Especialista | Estado |
|---|---|---|---|---|
| `sdd-orchestrator.md:3` | "Bind this to the dedicated agent only" | se elimina (lo cumple la entrega) | no | propuesta |
| `sdd-orchestrator.md:5-9` | Role: coordinador, no ejecutor | coordinador | no | propuesta |
| `sdd-orchestrator.md:13` | CLI `openspec` dueño del estado | universal | sí (`CLAUDE.md`; repetida en encargo a Explore/Plan) | propuesta |
| `sdd-orchestrator.md:14` | Delegar, no inflar | coordinador | no | propuesta |
| `sdd-orchestrator.md:15` | Engram persiste contexto | universal (mínimo) | sí (`CLAUDE.md`) | (TBD grupo) |
| `sdd-orchestrator.md:17-24` | Starting work, fundación | coordinador | no | propuesta |
| `sdd-orchestrator.md:26-46` | Delegation rules (toggle) | módulo `delegation.md` | sí, cuando se delega (encargo) | propuesta |
| 07 D1 | Cuándo delegar (regla breve) | coordinador (cuerpo de `SKILL.md`) | no | propuesta |
| `sdd-orchestrator.md:48-108` | Workflow, comandos, manejo de pedidos | coordinador | no | propuesta |
| `sdd-orchestrator.md:109-165` | Patrón de lanzamiento y brief | coordinador | no | propuesta |
| `sdd-orchestrator.md:167-201` | Ciclo de artefactos y comandos CLI | módulo por acción (propose/apply) | sí (encargo) | (TBD 04) |
| `sdd-orchestrator.md:205-206` | NEVER guess state; NEVER crear `openspec/` a mano | universal | sí (`CLAUDE.md`) | propuesta |
| `sdd-orchestrator.md:207-210` | NEVER bloquear, NEVER apply/propose inline, ambigüedad, `/sdd-*` | coordinador | no (lo contradiría) | propuesta |
| `sdd-orchestrator.md:212-226` | Model assignments (toggle) | módulo `model-assignments.md` | no | propuesta |
| `governance.md:1-13` | Niveles de autonomía por dominio | módulo propose (03) / núcleo (05) | sí | (TBD grupo: diverge con 05) |
| `governance.md:14-19` | Cómo aplicarlos al decidir delegación y checkpoints | coordinador | no | (TBD grupo) |
| `engram-protocol.md:1-5` | Encabezado "ALWAYS ACTIVE" | universal | sí | (TBD grupo) |
| `engram-protocol.md:6-12` | Wiring OPSX | módulo `engram-opsx.md` | no | propuesta |
| `engram-protocol.md:13-59` | Guardado proactivo y búsqueda | universal | sí | (TBD grupo) |
| `engram-protocol.md:60-92` | Cierre de sesión y recuperación tras compactación | coordinador | no | (TBD grupo) |
| `engram-protocol.md:93-117` | Git sync de memoria | (TBD grupo) | (TBD) | abierto |
| `compose.go:93` | Línea `Strict TDD Mode: enabled` | módulo `tdd.md` | sí, en apply | propuesta |
| `strict-tdd.md` | Módulo TDD | módulo `tdd.md` (apply) | sí, en apply | propuesta |
| 05 D17 | Lectura dirigida, resumen acotado, selección de herramientas, reporte conciso | módulos nuevos | según 05 D2 | pendiente de texto de 05 |
| 06 D5 | Contrato de evidencia y extracto `## Evidence Contract` | módulo `delegation.md` y encargo con comandos | sí, si ejecuta comandos | pendiente de texto de 06 |
| 07 D6 | Sobre de retorno | módulo `delegation.md` | sí, cuando se delega | pendiente de texto de 07 |
| harness `permissions` | Denegaciones en settings | universal por mecanismo | sí | sin cambio |
| sección `design-mockups` | Gates visuales | queda en `CLAUDE.md` (G7) | sí | sin cambio |

Los rangos de `governance.md` y `engram-protocol.md` salen de sus encabezados (`governance.md:1`,
`:5`, `:14`; `engram-protocol.md:1`, `:6`, `:13`, `:48`, `:60`, `:84`, `:93`).

## Contradicciones encontradas en el código

- **C1.** `sdd-orchestrator.md:3` pide atarse a un agente dedicado, pero Claude lo entrega en el
  archivo compartido (`adapter.go:125-127`). Es el problema que motiva el change.
- **C2.** El coordinador delega en skills `openspec-apply-change` y `openspec-archive-change`
  (`sdd-orchestrator.md:100-105`), pero las skills embebidas se llaman `openspec-apply` y
  `openspec-archive` y son marcadores de posición sin frontmatter
  (`assets/skills/openspec-apply/SKILL.md:3`), no listadas en `harnesses.yaml`. El origen real de
  esas skills en Claude queda como tarea 1.6.
- **C3.** Reusar `InjectPrimaryAgent` para Claude escribiría `agent` como objeto en
  `settings.json`, donde Claude usa `agent` para elegir el agente de sesión (D1).
- **C4.** `WriteFileAtomic` ignora el modo existente del archivo y amplía el directorio padre a
  `0o755` (D6).
- **C5.** `PurgeStaleSections` borra toda sección `active-stack:` no registrada: una sección
  universal nueva que no se registre se autodestruye en la próxima instalación (D2).
- **C6.** `CLAUDE.md:9` y `AGENTS.md:148-149` del repo dicen que el orquestador vive en el
  `~/.claude/CLAUDE.md` del usuario; dejarán de ser ciertas (Knowledge Base Impact).
- **C7.** Uninstall no conoce el target Project, por decisión vigente (`uninstall_flags.go:50`);
  verify sí (`projectVerifyAdapter`).
- **C8.** `config.Install` guarda el snapshot bajo la raíz del proyecto en target Project
  (`install.go:35-36`), contra `internal/install/plan.go:138-140`. Preexistente; el inyector nuevo
  no lo copia (D5 paso 3).
- **C9.** `refreshSkillDirectory` reemplaza una skill existente con el mismo ID sin comprobar de
  quién es (`idempotent.go:42-66`). El inyector nuevo agrega ese chequeo (D1, D5 paso 2).

## Risks / Trade-offs

- [BREAKING: la sesión normal deja de coordinar] → aviso en install/update, nota en README.
- [El coordinador se trunca tras compactar] → presupuesto de `SKILL.md`, módulos aparte, relectura
  del módulo en curso (D10).
- [El rol no sobrevive a la reanudación] → tarea 1.14; si falla, volver a invocar y documentarlo.
- [No hay salida del rol dentro de la sesión] → documentado (D3).
- [Se pierde la allowlist de tipos de subagente] → G13 con 08.
- [El contenido del usuario en `CLAUDE.md` sigue llegando a especialistas] → límite documentado.
- [Colisión con una skill propia del usuario, o skill Machine que tapa a la de Project] → marca de
  propiedad, error sin sobrescribir y advertencia de precedencia (D5 paso 2).
- [Versión de Claude sin skills de usuario o sin `disable-model-invocation`] → versión mínima y
  modo legado.
- [Fallo a mitad de migración o de uninstall] → transacción con restore y guarda de 01 (D5).
- [Reversión pisa ediciones posteriores del usuario] → hashes registrados y confirmación (D5).
- [Permisos ampliados en `~/.claude/`] → D6, G6, revisión compartida con 01 y 08.
- [Divergencia con 04 o con 05 en la matriz] → matriz única co-mantenida con 04 y revisión conjunta
  antes de construir; divergencia de `governance` llevada al grupo.
- [Pruebas con testigo dependen de un modelo real] → oráculo por registro de contexto (D8),
  testigos aleatorios, control positivo, corridas con el manifiesto de 02.
- [Credenciales expuestas en el entorno de prueba] → D9.

## Migration Plan

1. Release N: entrega por capas para Claude detrás de la versión mínima; migración automática del
   bloque legado Machine con snapshot y manifiesto de migración (D5). Aviso BREAKING en la salida de
   install/update con `/sdd-orchestrator`. Proyectos registrados: advertencia con comando de
   migración (G10).
2. Reversión del usuario: `active-stack uninstall --strategy restore --restore-manifest <manifiesto
   de migración>` devuelve `CLAUDE.md` byte a byte si no hubo ediciones posteriores; si las hubo,
   pide confirmación (D5). Alternativa: reinstalar la release anterior, que vuelve a inyectar el
   bloque legado; el directorio de la skill queda huérfano y uninstall lo elimina por su marca.
3. Rollback del proyecto: revertir el change; OpenCode, Codex y Antigravity no se afectan.

## Open Questions

- Si la reanudación con `--resume` conserva la skill invocada (tarea 1.14).
- Si el reenganche tras compactar mantiene también los archivos de módulo leídos o sólo el cuerpo
  de `SKILL.md` (tarea 1.11).
- Ubicación de la matriz común con 04 (archivo en `openspec/` o sección de `CHANGES.md`).
