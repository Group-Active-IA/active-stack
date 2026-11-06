# task-tool-selection

## Purpose

Define qué skills e integraciones opcionales queda habilitado usar en cada clase de tarea, de modo que una integración que la tarea no necesita no cargue contexto ni amplíe la superficie disponible sin motivo.

## ADDED Requirements

### Requirement: Herramientas habilitadas por clase de tarea (pendiente de grupo: eje de clases de tarea)

El sistema SHALL declarar, para cada clase de tarea, qué skills e integraciones quedan habilitadas. Las clases de tarea SHALL ser las acciones del ciclo OPSX: explore, propose, apply, verify y archive. Una integración que no figure como habilitada para la clase de tarea activa MUST NOT presentarse como disponible en esa tarea.

#### Scenario: Tarea de exploración

- **WHEN** la clase de tarea activa es de exploración de código
- **THEN** las herramientas de lectura declaradas para esa clase quedan disponibles
- **AND** las integraciones no declaradas para esa clase no se ofrecen

#### Scenario: Tarea sin componente visual

- **WHEN** la clase de tarea activa no declara trabajo visual
- **THEN** las integraciones de navegador y de captura no se ofrecen en esa tarea

### Requirement: Integración opcional desactivada por defecto

Toda integración declarada como opcional SHALL estar desactivada mientras no se la habilite explícitamente. El sistema MUST NOT activar una integración opcional por el solo hecho de estar instalada.

#### Scenario: Integración opcional instalada y no habilitada

- **WHEN** una integración opcional está instalada pero no habilitada para la tarea activa
- **THEN** la tarea se ejecuta sin esa integración

### Requirement: Degradación declarada cuando falta una integración

Cuando una integración habilitada para la clase de tarea no está disponible, el sistema SHALL informar qué capacidad queda sin cubrir y cuál es el camino alternativo si existe. El sistema MUST NOT presentar el resultado como equivalente al obtenido con la integración.

#### Scenario: Integración habilitada pero ausente

- **WHEN** la clase de tarea declara una integración y esa integración no está disponible
- **THEN** el sistema informa qué capacidad queda sin cubrir
- **AND** ofrece el camino alternativo declarado si existe

#### Scenario: Sin camino alternativo declarado

- **WHEN** falta una integración habilitada y no hay camino alternativo declarado
- **THEN** el sistema informa el límite
- **AND** no presenta el resultado parcial como completo

### Requirement: Ninguna invocación cuyo resultado ya está disponible

El sistema SHALL evitar invocar una herramienta o integración cuando su resultado para esa misma entrada ya está disponible en la sesión. Una invocación repetida sin que haya cambiado el estado que observa MUST quedar registrada como llamada innecesaria.

#### Scenario: Integración consultada dos veces por lo mismo

- **WHEN** el agente invoca una integración con la misma entrada dos veces y el estado que esa integración observa no cambió entre ambas
- **THEN** la segunda invocación queda registrada como llamada innecesaria

#### Scenario: Sondeo de una integración declarada no disponible

- **WHEN** una integración ya quedó informada como no disponible para la tarea activa
- **THEN** el agente no vuelve a invocarla en esa tarea
- **AND** usa el camino alternativo declarado si existe

### Requirement: Comparabilidad con y sin cada integración opcional

El sistema SHALL permitir ejecutar una misma clase de tarea con y sin cada integración opcional, conservando la misma definición de aceptación. La diferencia observada MUST poder atribuirse a esa integración y no a un cambio de criterio.

#### Scenario: Misma tarea con y sin una integración opcional

- **WHEN** se ejecuta la misma clase de tarea con una integración opcional habilitada y luego deshabilitada
- **THEN** la definición de aceptación es idéntica en ambas corridas
- **AND** la diferencia de contexto efectivo queda atribuida a esa integración
