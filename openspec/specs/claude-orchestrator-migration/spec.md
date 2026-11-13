# claude-orchestrator-migration Specification

## Purpose
Garantiza que pasar del bloque legado en `CLAUDE.md` a la entrega por capas, y repetir, actualizar
o deshacer esa instalación, nunca pierda ni exponga contenido del usuario.

## Requirements

### Requirement: Migración desde el bloque legado
Al instalar sobre un `CLAUDE.md` con el bloque legado en una versión compatible, el instalador MUST
tomar un snapshot antes de escribir, registrar ese manifiesto como manifiesto de migración, purgar
sólo las secciones del coordinador y dejar el bloque universal y la skill del coordinador. Actor:
instalador. Disparador: install o update. Resultado: manifiesto registrado y archivos nuevos.

#### Scenario: Migración con snapshot previo
- **WHEN** `CLAUDE.md` contiene `active-stack:sdd-orchestrator` y texto del usuario, y se ejecuta install
- **THEN** existe un snapshot con el contenido previo exacto bajo `<home>/.active-stack/backups/` (también en target Project), su ruta queda registrada como manifiesto de migración, el bloque legado ya no está, el bloque universal aparece una vez, existe `<skills>/sdd-orchestrator/SKILL.md` con marca de propiedad y el texto del usuario es idéntico

#### Scenario: Secciones administradas hermanas se conservan
- **WHEN** `CLAUDE.md` contiene también `active-stack:design-mockups`
- **THEN** esa sección queda intacta después de la migración

### Requirement: Reinstalación idempotente
Repetir install con la misma versión y opciones MUST NOT cambiar ningún archivo ni el manifiesto de
migración registrado. Actor: instalador. Disparador: segunda ejecución de install. Resultado:
"sin cambios" y bytes idénticos.

#### Scenario: Segunda instalación no escribe
- **WHEN** se ejecuta install dos veces seguidas con las mismas opciones
- **THEN** la segunda reporta que no hubo cambios, el marker universal aparece una sola vez, cada archivo de la skill es idéntico byte a byte y el manifiesto de migración registrado no cambia

### Requirement: Update real migra y reemplaza sólo contenido administrado
`active-stack update` (camino `--sync` que reaplica la selección registrada) MUST migrar la entrega
Machine de Claude y reemplazar bloque universal y skill propia sin tocar contenido del usuario ni
skills ajenas. Actor: usuario. Disparador: `active-stack update`. Resultado: testigo nuevo
presente y viejo ausente.

#### Scenario: update --sync sobre un HOME con bloque legado
- **WHEN** el HOME tiene `installation.json` con selección para Claude y `CLAUDE.md` con el bloque legado, y corre el camino `update --sync`
- **THEN** el resultado es el mismo que el escenario "Migración con snapshot previo"

#### Scenario: Update cambia el testigo
- **WHEN** la instalación tiene el testigo T1 en coordinador, módulos y universal, y el update aplica assets con T2
- **THEN** `CLAUDE.md` y todos los archivos de la skill contienen T2, ninguno contiene T1 y el texto del usuario es idéntico

#### Scenario: Update quita un módulo de un toggle desactivado
- **WHEN** la instalación tiene `modules/tdd.md` y el update reaplica la selección con `tdd` inactivo
- **THEN** `modules/tdd.md` ya no existe, los demás módulos quedan y la versión anterior del directorio queda en un snapshot

#### Scenario: Usuario editó un archivo de la skill con marca
- **WHEN** el usuario modificó un archivo dentro del directorio de la skill que tiene marca de propiedad y corre un update con contenido distinto
- **THEN** el update lo reemplaza, la versión editada queda en un snapshot y la salida nombra la ruta del snapshot (comportamiento declarado: el directorio con marca pertenece a Active Stack)

#### Scenario: Proyectos registrados con instalación Project V1
- **WHEN** `installation.json` lista proyectos y alguno tiene el bloque legado en `<root>/.claude/CLAUDE.md`
- **THEN** el update no escribe en esos proyectos y emite una advertencia por proyecto con la ruta y el comando de migración explícita (`active-stack install` con target Project)

#### Scenario: HOME V1 sin registro de selección
- **WHEN** el HOME no tiene `installation.json` (o lo tiene sin agentes registrados), `CLAUDE.md` tiene el bloque legado y corre el camino `update --sync`
- **THEN** el update no migra ni escribe `CLAUDE.md` (hoy sólo refresca skills, `cmd/active-stack/update.go:375-378`) y su salida, además del mensaje actual, nombra el cambio de entrega del coordinador de Claude (BREAKING), `/sdd-orchestrator` y el comando `active-stack install` que lo aplica

### Requirement: Rollback durante la operación
Si una escritura de la entrega para Claude falla a mitad de install, update o uninstall, el
instalador MUST restaurar el estado previo desde el snapshot de esa operación y MUST NOT pisar un
archivo que cambió después de que el instalador lo escribió (guarda `GuardRollback` de 01). Actor:
instalador. Disparador: error de E/S. Resultado: archivos iguales al estado previo, o preservados
con aviso, y código de salida no cero.

#### Scenario: Falla el snapshot
- **WHEN** la creación del snapshot falla antes de escribir
- **THEN** ningún archivo cambia y el error nombra la ruta afectada

#### Scenario: Falla la escritura de la skill después de purgar CLAUDE.md
- **WHEN** la purga de `CLAUDE.md` tuvo éxito y la activación del directorio de la skill falla
- **THEN** `CLAUDE.md` se restaura al contenido previo y el usuario no queda sin coordinador ni con el bloque a medias

#### Scenario: Otro proceso cambió CLAUDE.md durante la operación
- **WHEN** el instalador escribió `CLAUDE.md`, otro proceso lo modificó, y después falla la activación de la skill
- **THEN** el rollback no escribe `CLAUDE.md`, el error nombra la ruta y la del respaldo, y los demás archivos sí vuelven al estado previo

#### Scenario: Uninstall parcial
- **WHEN** uninstall quitó el bloque universal y falla al borrar el directorio de la skill
- **THEN** el bloque universal se restaura desde el snapshot de esa desinstalación y el error nombra la ruta de la skill

### Requirement: Reversión posterior que no pisa ediciones del usuario
Uninstall con estrategia restore MUST restaurar `CLAUDE.md` y el directorio de la skill
administrada desde el manifiesto que recibe en `--restore-manifest`, MUST NOT restaurar
`settings.json` (este change sólo lo lee) y MUST NOT sobrescribir un archivo administrado editado
después de la migración sin la confirmación específica de esta guarda. Actor: usuario. Disparador:
`uninstall --strategy restore --restore-manifest <m>`. Resultado: estado previo o aviso.

#### Scenario: Restore sin ediciones posteriores
- **WHEN** se migró una instalación legada, el usuario no editó nada y se ejecuta uninstall con restore
- **THEN** `CLAUDE.md` es idéntico byte a byte al previo a la migración, el directorio de la skill creado por Active Stack ya no existe y `settings.json` no cambia

#### Scenario: Usuario editó CLAUDE.md después de migrar
- **WHEN** el hash actual de `CLAUDE.md` difiere del registrado al terminar la migración y se ejecuta restore sin el flag de confirmación de la guarda (nombre tentativo `--overwrite-edited`)
- **THEN** el restore no escribe `CLAUDE.md`, guarda el contenido actual en un snapshot nuevo, muestra la ruta del diff contra el manifiesto y termina con código no cero; con `--overwrite-edited` restaura después de guardar ese snapshot

#### Scenario: --yes no saltea la guarda por hash
- **WHEN** se da el caso anterior y el usuario pasa `--yes` sin `--overwrite-edited`
- **THEN** el resultado es el mismo que sin confirmación: `--yes` mantiene su significado actual y no autoriza pisar ediciones posteriores

#### Scenario: Manifiesto distinto del registrado
- **WHEN** `--restore-manifest` apunta a un manifiesto distinto del manifiesto de migración registrado
- **THEN** el restore de ese manifiesto sigue el comportamiento actual para los demás archivos, la guarda por hash se aplica igual a `CLAUDE.md` y a los archivos de la skill, y la salida advierte que no es el manifiesto de migración y nombra su ruta

#### Scenario: El manifiesto de migración no se poda
- **WHEN** se crea el manifiesto de migración y luego corre la poda de respaldos (`Prune`) con cualquier límite
- **THEN** el manifiesto de migración sigue existiendo porque se creó fijado (`pinned: true`)

#### Scenario: Varias reinstalaciones después de migrar
- **WHEN** hubo reinstalaciones o updates que generaron snapshots nuevos después de la migración
- **THEN** el manifiesto de migración registrado sigue siendo el que nombra la salida de install y la documentación de reversión, no el snapshot más reciente

#### Scenario: Uninstall sin restore limpia sólo lo propio (target Machine)
- **WHEN** se ejecuta uninstall sin estrategia restore sobre una instalación Machine
- **THEN** se eliminan el bloque universal y el directorio de la skill con marca de propiedad, y el resto de `CLAUDE.md` y del directorio de skills queda intacto

### Requirement: Permisos del usuario preservados
Al reescribir un archivo existente del usuario, el instalador MUST conservar su modo de permisos y,
sobre el directorio que lo contiene, MUST agregar como máximo el bit de escritura del dueño, nunca
permisos de grupo u otros. Actor: instalador. Disparador: escritura sobre archivo existente.
Resultado: modos comparados antes y después.

#### Scenario: CLAUDE.md con 0600
- **WHEN** `CLAUDE.md` tiene modo 0600 y su directorio 0700, y se ejecuta install
- **THEN** después del install el archivo sigue en 0600 y el directorio en 0700

#### Scenario: Directorio de skills con 0700
- **WHEN** `<home>/.claude/skills` tiene modo 0700 y se instala la skill del coordinador
- **THEN** el directorio sigue en 0700

#### Scenario: Directorio sin escritura del dueño sigue siendo reparable
- **WHEN** el directorio padre tiene modo 0500 (dejado por una versión vieja del instalador)
- **THEN** el install agrega sólo el bit de escritura del dueño (0700) y completa la escritura

### Requirement: Verify refleja la entrega por capas en ambos targets
El verify posterior a install MUST fallar, en Machine y Project y en versión compatible, si la
skill del coordinador está incompleta (`SKILL.md`, módulos activos o `disable-model-invocation`),
si el bloque universal no aparece una vez o si queda una sección legada; en modo legado MUST
advertir. Actor: instalador. Disparador: hook de verify posterior a install. Resultado: chequeo
con ID y motivo.

#### Scenario: Bloque legado residual
- **WHEN** en versión compatible `CLAUDE.md` todavía contiene `active-stack:sdd-orchestrator` después del install
- **THEN** verify falla con un mensaje que nombra el marker residual y la ruta

#### Scenario: Módulo faltante
- **WHEN** `tdd` está activo y falta `<skills>/sdd-orchestrator/modules/tdd.md`
- **THEN** verify falla nombrando el módulo y la ruta

#### Scenario: Instalación correcta
- **WHEN** la skill existe con marca de propiedad, `disable-model-invocation: true` y sus módulos, y el bloque universal aparece una vez
- **THEN** los chequeos de Claude para `sdd-orchestrator` pasan

#### Scenario: Verify en target Project
- **WHEN** la instalación se hizo con target Project en `<root>`
- **THEN** el verify posterior al install revisa `<root>/.claude/CLAUDE.md` y `<root>/.claude/skills/sdd-orchestrator/`, resueltos por el adaptador de verify de Project, y una skill presente sólo en `<home>` no hace pasar el check Project

### Requirement: Limpieza manual documentada para target Project
Uninstall sigue siendo sólo Machine en este change; el sistema MUST documentar en README los pasos
manuales para quitar en target Project el directorio de la skill con marca y la sección universal,
y la salida de install Project MUST nombrar esas rutas. Actor: usuario. Disparador: install
Project o lectura de README. Resultado: rutas y pasos visibles.

#### Scenario: Install Project informa la limpieza manual
- **WHEN** termina un install para Claude con target Project en `<root>`
- **THEN** la salida nombra `<root>/.claude/skills/sdd-orchestrator/` y la sección universal de `<root>/.claude/CLAUDE.md` como lo que hay que quitar a mano para desinstalar, y README contiene los mismos pasos

#### Scenario: Uninstall sigue rechazando --project
- **WHEN** se ejecuta `active-stack uninstall --project <root>`
- **THEN** el comportamiento es el actual (flag no reconocido) y el test existente que lo fija no cambia su aserción

### Requirement: OpenCode y otros agentes sin cambios
La entrega para OpenCode, Codex y Antigravity MUST producir los mismos archivos que antes del
change. Actor: instalador. Disparador: install para esos agentes. Resultado: igualdad con la salida
registrada antes de tocar `compose.go`.

#### Scenario: Composición de otros agentes intacta
- **WHEN** se ejecutan los tests de composición de OpenCode, Codex y Antigravity contra goldens grabados antes del change (los de Codex y Antigravity se agregan en este change, porque hoy no existen)
- **THEN** no hay diff y OpenCode sigue registrando `agent.sdd-orchestrator` con `mode: primary`
