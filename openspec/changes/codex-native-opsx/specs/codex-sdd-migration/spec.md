## Purpose

Define cómo Active Stack migra una configuración de Codex que trae el bloque SDD administrado hacia la entrega nativa, y cómo se revierte sin perder instrucciones ni archivos del usuario. Cubre reinstalación idempotente, rollback ante fallo, propiedad de archivos, punto de reversión, actualización desde instalaciones anteriores y desinstalación.

> Borrador condicionado a G1–G13 y a la rama D6 de `design.md`. Los dos requisitos de reversión (G12-a y G12-b) son excluyentes: cuando el grupo decida G12, se conserva solo el elegido.

## ADDED Requirements

### Requirement: La migración reemplaza solo el bloque administrado

Al instalar `sdd-orchestrator` para Codex, el instalador SHALL reemplazar por el arranque el contenido de todos los bloques del ID `sdd-orchestrator`, tanto bajo el prefijo actual como bajo los heredados (`jr-stack:`), aunque coexistan. El resultado debe ser una sola sección con el prefijo actual. El contenido del usuario y las secciones de otros harnesses MUST quedar intactos.

#### Scenario: Bloque anterior con contenido del usuario alrededor

- **GIVEN** un `~/.codex/AGENTS.md` con:
  - reglas del usuario antes y después del bloque `active-stack:sdd-orchestrator` anterior, que pesa 18.010 bytes con los cinco toggles;
  - una sección `active-stack:design-mockups`
- **WHEN** el usuario ejecuta `active-stack install --agent codex --mode full`
- **THEN** los bytes del usuario antes y después del bloque son idénticos a los originales
- **AND** la sección `design-mockups` es idéntica a la original
- **AND** la sección `sdd-orchestrator` contiene solo el arranque

#### Scenario: Marker heredado aislado

- **GIVEN** un `AGENTS.md` con el bloque bajo `<!-- jr-stack:sdd-orchestrator -->`
- **WHEN** se instala para Codex
- **THEN** el bloque heredado desaparece
- **AND** queda una sola sección con el prefijo `active-stack:`

#### Scenario: Markers actual y heredado coexistentes

- **GIVEN** un `AGENTS.md` con un bloque `active-stack:sdd-orchestrator` y otro `jr-stack:sdd-orchestrator` de 18 KB
- **WHEN** se instala para Codex
- **THEN** no queda ningún marker `jr-stack:sdd-orchestrator`
- **AND** hay exactamente una sección `active-stack:sdd-orchestrator` con el arranque
- **AND** el contenido del usuario queda intacto

#### Scenario: Snapshot antes de escribir

- **WHEN** la migración modifica `AGENTS.md`
- **THEN** existe un snapshot con el contenido previo, creado antes de la escritura

#### Scenario: Falla del snapshot

- **GIVEN** un snapshot que no se puede crear
- **WHEN** se intenta la migración
- **THEN** la instalación devuelve un error
- **AND** `AGENTS.md`, las skills y los roles quedan sin cambios

### Requirement: La reinstalación es idempotente

Reinstalar `sdd-orchestrator` para Codex con la misma selección SHALL dejar exactamente un marker del arranque. Un archivo administrado cuyo contenido no cambió MUST NOT reescribirse.

#### Scenario: Segunda instalación

- **GIVEN** una instalación ya migrada
- **WHEN** se repite la misma instalación
- **THEN** ningún archivo administrado cambia de bytes ni de fecha de modificación
- **AND** el resultado informa que ya estaba instalado

#### Scenario: Cambio de toggles entre versiones del catálogo

- **GIVEN** una instalación migrada hecha con una versión del binario cuyo catálogo incluye el toggle `tdd`
- **WHEN** el usuario ejecuta `active-stack update` y la versión nueva trae un catálogo sin `tdd` (los toggles no se exponen al usuario)
- **THEN** el módulo TDD administrado se retira, con snapshot previo
- **AND** las demás piezas quedan idénticas

### Requirement: Un fallo durante la instalación revierte la configuración

Si falla un paso posterior a la migración, como la escritura de un rol o una verificación dura, el pipeline SHALL restaurar el estado previo:

- `AGENTS.md` vuelve byte a byte;
- los archivos y directorios creados se eliminan;
- los que ya existían se conservan.

#### Scenario: Falla al escribir un rol

- **GIVEN** un `AGENTS.md` con contenido del usuario y el bloque anterior, y un directorio de skills con una skill del usuario
- **WHEN** la escritura de un archivo administrado falla después de haber reescrito `AGENTS.md`
- **THEN** tras el rollback, `AGENTS.md` es idéntico al original
- **AND** no queda ninguna skill, módulo ni rol administrado nuevo
- **AND** la skill del usuario sigue intacta

#### Scenario: Falla de verificación dura

- **WHEN** la verificación post-instalación encuentra que falta un módulo de un toggle activo
- **THEN** el hook de verificación dispara el rollback
- **AND** se restaura el estado previo

### Requirement: No se sobrescriben archivos que Active Stack no administra

Antes de cualquier paso de escritura, el instalador SHALL comprobar que cada destino administrado (skill, módulo o rol) no existe o lleva la marca de propiedad de Active Stack. Si existe sin la marca, la instalación MUST fallar sin ejecutar ningún paso de escritura y MUST nombrar la ruta en conflicto.

#### Scenario: Skill homónima del usuario

- **GIVEN** un `<skills-dir>/sdd-orchestrator/SKILL.md` creado por el usuario, sin marca de propiedad
- **WHEN** se instala `sdd-orchestrator` para Codex
- **THEN** la instalación falla con un error que nombra esa ruta
- **AND** ningún archivo, ni `AGENTS.md`, queda modificado

#### Scenario: Rol homónimo del usuario

- **GIVEN** un archivo de rol del usuario con el mismo nombre que un especialista, sin marca de propiedad
- **WHEN** se instala
- **THEN** la instalación falla y nombra el archivo
- **AND** el archivo del usuario queda idéntico

### Requirement: El punto de reversión previo a la migración sobrevive a reinstalaciones (G12-a)

Solo aplica si G12 = (a). La primera vez que se detecta el bloque anterior, el instalador SHALL crear, antes de migrar, un snapshot de migración dedicado con sello de tiempo en un directorio propio. Ninguna instalación ni actualización posterior MUST sobrescribirlo. El snapshot MUST figurar en la pantalla de backups.

#### Scenario: Update con dos corridas de install conserva el punto previo

- **GIVEN** un home con recibo de Codex en modo Full con harnesses Custom extra, de modo que `update` lanza dos `install` seguidos, y el bloque anterior en `AGENTS.md`
- **WHEN** el usuario ejecuta `active-stack update`
- **THEN** existe exactamente un snapshot de migración
- **AND** al restaurarlo, `AGENTS.md` contiene el bloque anterior byte a byte

#### Scenario: Reinstalar tras migrar no pierde el punto de reversión

- **GIVEN** una instalación migrada con su snapshot de migración
- **WHEN** el usuario reinstala dos veces
- **THEN** el snapshot de migración sigue intacto
- **AND** no se crea otro snapshot de migración

#### Scenario: Restauración del snapshot de migración

- **WHEN** el usuario restaura el snapshot de migración desde la pantalla de backups de la TUI
- **THEN** `AGENTS.md` vuelve a contener el bloque anterior byte a byte
- **AND** las piezas creadas por la migración se eliminan

#### Scenario: Reinstalar tras revertir

- **GIVEN** la restauración anterior
- **WHEN** el usuario vuelve a instalar
- **THEN** el resultado es igual al de la primera migración
- **AND** se crea un nuevo snapshot de migración, sin borrar el anterior

### Requirement: La reversión solo es posible hasta la siguiente instalación (G12-b)

Solo aplica si G12 = (b). El instalador SHALL informar, en la salida de la instalación que migra el bloque, que el snapshot de instalación permite revertir solo hasta la próxima instalación o actualización. La documentación MUST declarar el mismo límite. Ninguna salida ni documento MUST prometer una reversión posterior.

#### Scenario: Aviso al migrar

- **GIVEN** un `AGENTS.md` con el bloque anterior
- **WHEN** se instala para Codex
- **THEN** la salida de la instalación contiene el aviso del límite de reversión

#### Scenario: Reinstalación reemplaza el punto de reversión

- **GIVEN** una instalación migrada
- **WHEN** se reinstala
- **THEN** el snapshot de instalación contiene el estado ya migrado, tal como documenta el límite
- **AND** no se emite otro aviso, porque ya no hay bloque anterior

### Requirement: La actualización migra las instalaciones con recibo

Cuando el estado registrado incluye Codex, `active-stack update` SHALL aplicar la migración, con los mismos requisitos de backup, idempotencia y rollback, porque lo hace reaplicando la instalación. Sin recibo de selección, MUST NOT modificar `AGENTS.md` y MUST informar cómo migrar.

#### Scenario: Actualización con recibo

- **GIVEN** un home con estado registrado para Codex y el bloque anterior en `~/.codex/AGENTS.md`
- **WHEN** el usuario ejecuta `active-stack update`
- **THEN** la sección `sdd-orchestrator` queda migrada al arranque
- **AND** existen la skill, los módulos y los roles que prevé la rama D6
- **AND** la verificación pasa

#### Scenario: Actualización sin recibo

- **GIVEN** una instalación anterior sin estado registrado y el bloque anterior en `AGENTS.md`
- **WHEN** el usuario ejecuta `active-stack update`
- **THEN** `AGENTS.md` queda idéntico
- **AND** la salida indica que la configuración se migra ejecutando `install` con el modo actual

#### Scenario: Actualización repetida

- **WHEN** se ejecuta `active-stack update` dos veces seguidas con recibo
- **THEN** la segunda corrida no reescribe archivos administrados de Codex

### Requirement: La desinstalación retira solo las piezas administradas

Desinstalar `sdd-orchestrator` para Codex en máquina SHALL retirar:

- la sección del arranque;
- uno por uno, los archivos con marca de propiedad (skill, módulos y roles);
- un directorio administrado, solo si queda vacío.

Antes de retirar, MUST tomar un snapshot que incluya cada archivo administrado. MUST preservar el contenido del usuario, las secciones de otros harnesses y todo archivo sin marca, también dentro de un directorio administrado.

#### Scenario: Desinstalación completa

- **GIVEN** una instalación migrada, con reglas del usuario en `AGENTS.md` y una skill propia del usuario
- **WHEN** el usuario desinstala `sdd-orchestrator` para Codex
- **THEN** no quedan la sección `sdd-orchestrator` ni las piezas administradas
- **AND** las reglas del usuario, la sección `design-mockups` y la skill del usuario quedan intactas

#### Scenario: Archivo del usuario dentro de un directorio administrado

- **GIVEN** un archivo `notas.md` del usuario dentro de `<skills-dir>/sdd-orchestrator/`
- **WHEN** se desinstala y, en otra corrida, se fuerza un fallo que dispara el rollback del uninstall
- **THEN** en ambos casos `notas.md` queda intacto
- **AND** tras el uninstall exitoso el directorio sigue existiendo y solo contiene `notas.md`

#### Scenario: Desinstalación sin piezas

- **WHEN** se desinstala en un home donde las piezas no existen
- **THEN** la operación termina sin error y no modifica archivos
