# task-context-loading

## Purpose

Define qué instrucciones del orquestador recibe un agente en función de la acción activa y del rol que ocupa, de modo que una sesión no cargue módulos que su acción no usa y que la ausencia de un módulo obligatorio sea visible en lugar de silenciosa.

## ADDED Requirements

### Requirement: Núcleo obligatorio en toda sesión

El sistema SHALL entregar en toda sesión un núcleo de instrucciones que no depende de la acción activa, compuesto por las reglas de cumplimiento obligatorio y la referencia para invocar al orquestador. El núcleo MUST estar presente tanto en una sesión normal como en una sesión con el rol de orquestador activo y en la de un especialista.

#### Scenario: Sesión normal recibe el núcleo

- **WHEN** se abre una sesión sin invocar explícitamente al orquestador
- **THEN** el agente recibe el núcleo obligatorio
- **AND** no recibe los módulos asociados a una acción concreta

#### Scenario: El especialista conserva las reglas que le aplican

- **WHEN** el orquestador delega una tarea a un especialista
- **THEN** el especialista recibe el núcleo obligatorio además de su encargo
- **AND** las reglas de cumplimiento obligatorio siguen vigentes en ese rol

### Requirement: Módulos resueltos por acción activa (pendiente de grupo: inventario de módulos y eje de clases de tarea)

El sistema SHALL resolver qué módulos de instrucción se entregan a partir de la acción activa, y no a partir de una selección fijada en el momento de la instalación. Las acciones SHALL ser las del ciclo OPSX: explore, propose, apply, verify y archive. Cada módulo MUST declarar las acciones que lo requieren.

#### Scenario: Una acción sin requisito de TDD no recibe el módulo de TDD

- **WHEN** la acción activa no declara requisito de TDD
- **THEN** el módulo de TDD no se entrega en esa sesión
- **AND** el núcleo obligatorio permanece intacto

#### Scenario: Una acción que requiere delegación recibe su módulo

- **WHEN** la acción activa declara que requiere coordinación de subagentes
- **THEN** el módulo de delegación y coordinación se entrega en esa sesión

### Requirement: Un módulo obligatorio ausente detiene la acción

Cuando una acción declara un módulo como obligatorio y ese módulo no puede entregarse, el sistema SHALL informar la limitación y detener la acción que lo requiere. El sistema MUST NOT continuar la acción con el módulo ausente ni sustituirlo en silencio.

#### Scenario: Módulo obligatorio no disponible

- **WHEN** una acción declara un módulo obligatorio y el módulo no está disponible en la instalación
- **THEN** el sistema informa qué módulo falta y qué acción queda detenida
- **AND** la acción no se ejecuta

### Requirement: Resolución determinista

Para una misma acción, un mismo rol y una misma instalación, el sistema SHALL entregar siempre el mismo conjunto de módulos. El orden de ensamblado MUST ser independiente del orden en que los módulos se declaren.

#### Scenario: Dos resoluciones equivalentes coinciden

- **WHEN** se resuelve el contexto de la misma acción y el mismo rol dos veces sobre la misma instalación
- **THEN** ambas resoluciones producen el mismo conjunto de módulos y el mismo orden

#### Scenario: El orden declarado no altera el resultado

- **WHEN** dos instalaciones declaran los mismos módulos en orden distinto
- **THEN** la resolución para una misma acción produce un resultado idéntico en ambas

### Requirement: Paridad declarada entre clientes (pendiente de grupo: agentes incluidos)

El sistema SHALL declarar, para cada acción, qué conjunto de módulos corresponde en cada cliente soportado. Cuando un cliente no pueda entregar un módulo, esa diferencia MUST quedar declarada en lugar de resolverse como equivalencia.

#### Scenario: Misma acción en los dos clientes

- **WHEN** se resuelve el contexto de la misma acción en Claude y en Codex
- **THEN** el conjunto de módulos declarado coincide, o la diferencia queda registrada como límite del cliente

#### Scenario: Cliente sin soporte para un módulo

- **WHEN** un cliente no puede recibir un módulo que la acción declara
- **THEN** el sistema informa el límite de ese cliente
- **AND** no presenta la entrega como completa
