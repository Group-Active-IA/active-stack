# task-context-loading

## Purpose

Define qué instrucciones del orquestador recibe un agente en función de la acción activa y del rol que ocupa, de modo que una sesión no cargue módulos que su acción no usa y que la ausencia de un módulo obligatorio sea visible en lugar de silenciosa. En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción más lo declarado en el catálogo (`design.md` — D17).

## ADDED Requirements

### Requirement: Núcleo obligatorio en toda sesión

El sistema SHALL entregar en toda sesión un núcleo independiente de la acción: las filas de capa universal de la matriz común de 03 y 04 (`design.md` — D2), entre ellas las reglas obligatorias, los niveles de `governance` y la memoria universal. El núcleo MUST estar en la sesión normal, en la del orquestador y en la de un especialista; si el especialista no carga `CLAUDE.md` ni `AGENTS.md`, las reglas obligatorias y `governance` MUST viajar en su encargo.

#### Scenario: Sesión normal recibe el núcleo

- **WHEN** se abre una sesión sin invocar explícitamente al orquestador
- **THEN** el agente recibe el núcleo obligatorio
- **AND** no recibe los módulos asociados a una acción concreta

#### Scenario: El núcleo incluye governance y la memoria universal

- **WHEN** se compone el núcleo para cualquier acción y cualquier rol
- **THEN** contiene los niveles de `governance` y las reglas universales de memoria de la matriz común
- **AND** no contiene ningún módulo, incluido `engram-opsx`

#### Scenario: El especialista conserva las reglas que le aplican

- **WHEN** el orquestador delega una tarea a un especialista
- **THEN** el especialista recibe el núcleo obligatorio además de su encargo
- **AND** las reglas de cumplimiento obligatorio siguen vigentes en ese rol

#### Scenario: Especialista que no carga CLAUDE.md

- **WHEN** el orquestador de Claude delega en Explore, en Plan o en un agente con `omitClaudeMd: true`
- **THEN** el encargo repite las reglas de cumplimiento obligatorio y `governance`
- **AND** el especialista no depende de `CLAUDE.md` para recibirlas

### Requirement: Referencia de invocación por cliente

La referencia del núcleo para invocar al orquestador SHALL redactarse por cliente, con la invocación explícita acordada para 03 y 04: `/sdd-orchestrator` en Claude Code y `$sdd-orchestrator` en Codex CLI e IDE, con `/skills` como alternativa, y `@` en la app de ChatGPT. La referencia MUST NOT fijar una sola sintaxis para los dos clientes ni fijar `claude --agent`; el texto lo escriben 03 y 04 (`design.md` — D21).

#### Scenario: Referencia de invocación por cliente

- **WHEN** se compone el núcleo para Claude y para Codex
- **THEN** cada núcleo nombra la invocación explícita documentada para su cliente
- **AND** ninguno manda a escribir la sintaxis del otro cliente

### Requirement: Módulos resueltos por acción activa

El sistema SHALL resolver qué módulos de instrucción se entregan a partir de la acción activa, y no a partir de una selección fijada en el momento de la instalación. Las acciones SHALL ser las del ciclo OPSX: explore, propose, apply, verify y archive. Cada módulo MUST declarar en el catálogo las acciones en que es obligatorio, aquellas en que es opcional y si el especialista lo recibe en el encargo (`design.md` — D19).

#### Scenario: Una acción sin requisito de TDD no recibe el módulo de TDD

- **WHEN** la acción activa no declara requisito de TDD
- **THEN** el módulo de TDD no se entrega en esa sesión
- **AND** el núcleo obligatorio permanece intacto

#### Scenario: Una acción que delega recibe su módulo

- **WHEN** la acción activa coordina subagentes
- **THEN** el módulo `delegation` se entrega en esa sesión junto con `model-assignments`
- **AND** en Codex, mientras `model-assignments` se omita, la ausencia queda declarada como límite del cliente

#### Scenario: El módulo de memoria OPSX al inicio y al cierre

- **WHEN** empieza o cierra cualquier acción con el toggle de memoria activo
- **THEN** el agente recibe el módulo `engram-opsx`

#### Scenario: Resumen acotado en las acciones que producen evidencia

- **WHEN** la acción activa es `apply` o `verify`
- **THEN** el módulo `bounded-command-output` se entrega como obligatorio

### Requirement: La tabla de carga se genera desde el catálogo

La tabla acción → módulos que lee la skill del orquestador SHALL generarse desde el catálogo y MUST NOT mantenerse a mano, porque el agente no lee el catálogo embebido en el binario.

#### Scenario: Tabla generada desde el catálogo

- **WHEN** se compone la skill del orquestador para Claude o para Codex
- **THEN** la tabla acción → módulos del `SKILL.md` coincide byte a byte con la generada desde el catálogo

### Requirement: Un módulo obligatorio ausente detiene la acción

Cuando una acción declara un módulo como obligatorio y ese módulo no puede entregarse, el sistema SHALL detener solo la acción que lo requiere, explicar qué módulo o protección falta y seguir con las tareas que no dependen de él. El sistema MUST NOT continuar esa acción con el módulo ausente ni sustituirlo en silencio. Si el módulo es opcional, SHALL avisar y resolver la acción directo.

#### Scenario: Módulo obligatorio no disponible

- **WHEN** la acción `verify` requiere el módulo de delegación y el módulo no está disponible en la instalación
- **THEN** el sistema informa qué módulo falta y qué acción queda detenida, con la línea de forma fija `missing module: <id> — <action> stopped`
- **AND** la acción no se ejecuta

#### Scenario: Las tareas que no dependen del módulo siguen

- **WHEN** la acción `verify` queda detenida por falta del módulo de delegación y el pedido incluye además tareas que no lo requieren
- **THEN** el sistema completa esas tareas
- **AND** el reporte separa lo completado de la acción detenida y su motivo

#### Scenario: La acción detenida se retoma con un equivalente comprobado

- **WHEN** después de detenerse una acción, el módulo queda disponible o se usa una superficie compatible que lo entrega
- **THEN** la acción se puede retomar
- **AND** el reporte indica con qué módulo o superficie se retomó

#### Scenario: Módulo opcional no disponible

- **WHEN** una acción distinta de `verify` decide delegar y el módulo de delegación no está disponible
- **THEN** el sistema avisa que el módulo falta, con la línea de forma fija `missing optional module: <id> — continuing without it`
- **AND** resuelve la acción directo, sin delegar

#### Scenario: Superficie sin subagentes

- **WHEN** la acción `verify` se pide en una superficie que no documenta subagentes
- **THEN** el sistema informa que la superficie no puede entregar el módulo de delegación obligatorio
- **AND** la acción queda detenida como límite declarado de esa superficie

### Requirement: Un módulo vacío cuenta como ausente

Un módulo cuyo archivo existe sin texto (cero bytes o solo espacios) SHALL tratarse como ausente, en la sesión y en la verificación de la instalación (medido: `entrega-03-04.md` §5, C13 y R3).

#### Scenario: Módulo obligatorio vacío

- **WHEN** el archivo de un módulo obligatorio existe pero no tiene texto
- **THEN** el sistema lo trata como ausente y detiene la acción que lo requiere
- **AND** la verificación de la instalación falla nombrando el módulo

### Requirement: Módulos releídos después de compactar

Cuando la sesión se compacta o se reanuda durante una acción, el coordinador SHALL volver a invocar la skill y releer los módulos de la acción en curso antes de continuarla, porque la compactación saca del contexto el cuerpo de la skill y los módulos leídos (medido: `entrega-03-04.md` §5, C11).

#### Scenario: Acción que sigue después de compactar

- **WHEN** la sesión se compacta en medio de `apply`
- **THEN** antes de seguir, el coordinador invoca la skill otra vez y relee los módulos de `apply`
- **AND** no continúa la acción con reglas que ya no están en su contexto

### Requirement: Modo legado declarado

En una instalación en modo legado, sin skill ni módulos (Claude Code anterior a la versión mínima de la 03), el sistema SHALL declarar como límite que los módulos de esta ficha no se entregan, y todo resultado de `apply` o `verify` en esa instalación SHALL quedar *sin verificar*.

#### Scenario: Instalación en modo legado

- **WHEN** el instalador deja el coordinador como bloque único por versión insuficiente del cliente
- **THEN** la paridad por cliente registra que los módulos de contexto no se entregan
- **AND** un resultado de `apply` o `verify` en esa instalación figura *sin verificar*

### Requirement: Resolución determinista

Para una misma acción, un mismo rol y una misma instalación, el sistema SHALL entregar siempre el mismo conjunto de módulos. El orden de ensamblado MUST ser independiente del orden en que los módulos se declaren.

#### Scenario: Dos resoluciones equivalentes coinciden

- **WHEN** se resuelve el contexto de la misma acción y el mismo rol dos veces sobre la misma instalación
- **THEN** ambas resoluciones producen el mismo conjunto de módulos y el mismo orden

#### Scenario: El orden declarado no altera el resultado

- **WHEN** dos instalaciones declaran los mismos módulos en orden distinto
- **THEN** la resolución para una misma acción produce un resultado idéntico en ambas

### Requirement: Paridad declarada entre clientes

El sistema SHALL declarar, para cada acción, qué conjunto de módulos corresponde en cada cliente soportado (Claude y Codex). Cuando un cliente no pueda entregar un módulo, esa diferencia MUST quedar declarada en lugar de resolverse como equivalencia.

#### Scenario: Misma acción en los dos clientes

- **WHEN** se resuelve el contexto de la misma acción en Claude y en Codex
- **THEN** el conjunto de módulos declarado coincide, o la diferencia queda registrada como límite del cliente

#### Scenario: Cliente sin soporte para un módulo

- **WHEN** un cliente no puede recibir un módulo que la acción declara
- **THEN** el sistema informa el límite de ese cliente
- **AND** no presenta la entrega como completa

### Requirement: Comparabilidad con y sin cada módulo

Una misma acción SHALL poder ejecutarse con y sin cada módulo, conservando la misma definición de aceptación, para que la diferencia observada se atribuya al módulo y no a un cambio de criterio. La comparación MUST correr con el arnés fuera del binario y el protocolo de medición de 02.

#### Scenario: Misma acción con y sin un módulo

- **WHEN** se ejecuta la misma acción con un módulo cargado y luego sin él
- **THEN** la definición de aceptación es idéntica en ambas corridas
- **AND** la diferencia de contexto efectivo y de aceptación queda atribuida a ese módulo
