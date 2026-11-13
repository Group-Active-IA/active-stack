## Purpose

Define migración, actualización y retiro de las piezas administradas de Codex con preservación de contenido propio, respaldo y restauración protegida frente a ediciones posteriores.

**Alcance de entrega y revisión.**

Las pruebas técnicas de migración/rollback y su paquete de revisión están entregados. La disposición compartida de 01/03/08 sigue como deuda externa D04-010308 y D04-REV01, no como tarea abierta de implementación de Matías ni como aprobación recibida. Esto no cambia los requisitos de preservación ni acredita aceptación funcional de clientes. Ver `informes/deudas-aceptadas-05-06-07-2026-11-13.md`.

## ADDED Requirements

### Requirement: Preflight y propiedad

Antes de escribir, el instalador SHALL resolver rutas mediante adapters y comprobar propiedad de cada pieza existente. Una colisión con archivo ajeno o una configuración inválida SHALL detener el paquete sin modificar archivos. Cada nueva pieza SHALL tener propiedad reconocible.

#### Scenario: Skill homónima del usuario
- **WHEN** existe una skill o rol homónimo sin marca de Active Stack
- **THEN** se informa la ruta y no se modifica ni esa pieza ni las instrucciones de usuario

#### Scenario: Configuración inválida
- **WHEN** una configuración que debe actualizarse no se puede interpretar de forma segura
- **THEN** se conserva intacta y se informa el error

### Requirement: Migración acotada e idempotente

La migración SHALL respaldar antes de escribir y reemplazar únicamente las secciones administradas del orquestador por las capas nuevas. SHALL retirar copias heredadas reconocidas, preservar bytes propios y secciones hermanas y evitar reescrituras o duplicados al reinstalar.

#### Scenario: Markers actuales y heredados coexistentes
- **WHEN** coexisten bloques administrados actuales y heredados
- **THEN** queda una entrega vigente coherente, sin coordinación global legada y con contenido propio y design-mockups conservados

#### Scenario: Reinstalación con notas propias
- **WHEN** se reinstala el mismo paquete y hay notas del usuario dentro del directorio de skill
- **THEN** los archivos administrados idénticos no se reescriben y las notas quedan intactas

#### Scenario: Fallo de respaldo
- **WHEN** falla el snapshot previo
- **THEN** se informa el fallo y no se inicia la escritura

### Requirement: Rollback de una operación

Un fallo parcial de instalación SHALL restaurar las rutas de la operación cuando su estado coincide con lo escrito por ella. Si una ruta cambió después, SHALL preservarse y reportar el conflicto junto al respaldo recuperable.

#### Scenario: Fallo después de migrar instrucciones
- **WHEN** falla la escritura de un módulo después de migrar AGENTS.md
- **THEN** se recupera el estado previo en rutas sin cambios posteriores y se retiran sólo las piezas creadas por la operación

#### Scenario: Edición concurrente durante rollback
- **WHEN** una ruta fue editada después de escribirla el instalador
- **THEN** se preserva la edición y se informa la restauración incompleta y el respaldo disponible

### Requirement: Reversión posterior protegida

El punto de migración para reversión posterior SHALL conservar la primera copia previa y sobrevivir reinstalaciones y actualizaciones. La restauración SHALL comprobar estado esperado persistente y preservar cambios posteriores y archivos ajenos. La interfaz SHALL explicar el alcance y los conflictos antes de reemplazar contenido.

#### Scenario: Actualización repetida
- **WHEN** se migra y luego se reinstala o actualiza dos veces
- **THEN** se conserva el punto original de reversión y se puede identificar qué estado esperaba restaurar

#### Scenario: Edición posterior del usuario
- **WHEN** se solicita revertir un manifiesto cargado desde disco y un archivo difiere del estado esperado
- **THEN** ese archivo se conserva y el usuario recibe el conflicto y la ruta del respaldo

### Requirement: Update según registro de instalación

La actualización SHALL reaplicar una selección registrada mediante el mismo plan de propiedad y migración. Si falta el registro de instalación, SHALL informar adopción explícita y conservar la configuración existente. El registro de instalación SHALL distinguirse de los recibos de ejecución.

#### Scenario: Update con registro
- **WHEN** se reaplica Codex registrado
- **THEN** se usa el plan de migración con los mismos respaldos, controles y garantías de reinstalación

#### Scenario: Update sin registro
- **WHEN** no hay selección registrada para migrar
- **THEN** se informa la acción explícita necesaria y no se adopta ni sobrescribe la configuración silenciosamente

### Requirement: Retiro de archivos propios

Uninstall SHALL retirar sólo secciones y archivos administrados y eliminar directorios únicamente si quedan vacíos. La entrega nueva SHALL limitarse a Machine, con preservación de instrucciones y archivos de otros agentes. El diagnóstico SHALL identificar una instalación Project legada que siga aportando coordinación.

#### Scenario: Archivo propio dentro de la skill
- **WHEN** se desinstala y existe un archivo del usuario junto a módulos administrados
- **THEN** el archivo y su directorio quedan conservados y sólo se retiran piezas propias

#### Scenario: Target no disponible
- **WHEN** se solicita retiro de un target fuera del alcance implementado
- **THEN** se informa el límite y las instrucciones de limpieza segura sin afirmar que fue retirado

### Requirement: Alcance Machine de la migración

La nueva entrega SHALL migrar únicamente la instalación Machine y MUST informar si una configuración Project legada impide demostrar una sesión normal libre de coordinación.

#### Scenario: Proyecto con instrucciones legadas
- **WHEN** se comprueba la entrega Machine desde un proyecto con coordinación legada propia
- **THEN** se identifica la fuente Project y la limitación de la aceptación sin modificarla automáticamente

#### Scenario: Perfil Machine aislado
- **WHEN** se instala la entrega nueva en un perfil sin coordinación Project legada
- **THEN** se verifica la activación explícita y el retiro Machine con sus rutas administradas

#### Scenario: Instalación Codex Project con entrega legada
- **WHEN** se instala o reaplica el harness de Codex con target Project
- **THEN** el resultado y la salida del instalador informan que conserva la entrega legada y que la nueva migración corresponde a Machine
- **AND** los archivos Project conservan su estrategia de instalación anterior
