# task-tool-selection

## Purpose

Define qué skills e integraciones opcionales queda habilitado usar en cada clase de tarea, de modo que una integración que la tarea no necesita no cargue contexto ni amplíe la superficie disponible sin motivo. La tabla de clase por herramienta está en `design.md` — D18 y se declara en el catálogo. En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción más lo declarado en el catálogo, y "registrar" es declararlo en el reporte de la acción (`design.md` — D17).

## ADDED Requirements

### Requirement: Herramientas habilitadas por clase de tarea

El catálogo SHALL declarar, para cada clase de tarea, qué herramientas base corresponden y qué skills e integraciones opcionales quedan habilitadas, conforme a la tabla de `design.md` — D18. Las clases de tarea SHALL ser las acciones del ciclo OPSX: explore, propose, apply, verify y archive. Una integración que no figure como habilitada para la clase de tarea activa MUST NOT presentarse como disponible en esa tarea, al nivel de control que el catálogo declare para el cliente activo.

#### Scenario: Acción explore

- **WHEN** la acción activa es `explore`
- **THEN** quedan disponibles la lectura dirigida, la búsqueda y Context7
- **AND** Playwright y las skills de dominio no se ofrecen

#### Scenario: Acción que no toca UI

- **WHEN** la acción activa es `apply` o `verify` y no toca UI
- **THEN** Playwright no se ofrece en esa tarea

#### Scenario: Integraciones de diseño sin el harness de mockups

- **WHEN** el harness `design-mockups` no está activo en el proyecto
- **THEN** las integraciones de diseño no se ofrecen en ninguna acción

### Requirement: Nivel de control declarado por cliente

El catálogo SHALL declarar, por integración opcional y por cliente, en qué nivel queda no habilitada: (a) *solo recomendada*; (b) *fuera del contexto*; (c) *restringida por el cliente*, con los mecanismos que documenta cada cliente (`design.md` — D18). El nivel (a) MUST NOT presentarse como aislamiento, y un mecanismo que solo preaprueba, como `allowed-tools`, MUST NOT contarse como (c). Lo que un cliente no alcance por acción MUST declararse como límite.

#### Scenario: MCP por acción en Codex

- **WHEN** la tabla deshabilita una integración MCP para una acción y el cliente activo es Codex
- **THEN** el catálogo declara el nivel (a) para ese control por acción
- **AND** el límite queda registrado: el archivo de un rol de Codex no puede fijar `mcp_servers`

#### Scenario: Acción que corre en un subagente de Claude con lista propia

- **WHEN** la acción corre en un subagente de Claude cuyo archivo declara `tools` o `mcpServers` sin la integración
- **THEN** el catálogo declara el nivel (c) para esa acción en Claude

#### Scenario: Preaprobación que no restringe

- **WHEN** una skill de Claude declara `allowed-tools`
- **THEN** esa declaración no se cuenta como nivel (c) para ninguna integración

#### Scenario: Uso de una integración no habilitada al nivel (a)

- **WHEN** el agente invoca una integración que la acción activa no habilita y el control declarado es solo (a)
- **THEN** la invocación queda registrada como uso fuera de clase con `out-of-class use: <integration> in <action>`
- **AND** el reporte no presenta la integración como bloqueada

### Requirement: Recomendadas y restringidas documentadas

La documentación generada desde el catálogo SHALL mostrar, por integración opcional y por cliente, si queda solo recomendada, fuera del contexto o restringida por el cliente.

#### Scenario: Documentación de recomendadas y restringidas

- **WHEN** se genera la documentación de herramientas desde el catálogo
- **THEN** cada integración opcional figura, por cliente, como *solo recomendada*, *fuera del contexto* o *restringida por el cliente*
- **AND** ninguna figura como restringida sin un mecanismo del cliente que lo impida

### Requirement: Integración opcional desactivada por defecto

Toda integración declarada como opcional SHALL estar desactivada mientras no se la habilite para la clase de tarea activa. El sistema MUST NOT activar una integración opcional por el solo hecho de estar instalada.

#### Scenario: Integración opcional instalada y no habilitada

- **WHEN** una integración opcional está instalada pero no habilitada para la tarea activa
- **THEN** la tarea se ejecuta sin esa integración

### Requirement: Degradación declarada cuando falta una integración

Cuando una integración habilitada para la clase de tarea no está disponible, el sistema SHALL informar qué capacidad queda sin cubrir y cuál es el camino alternativo si existe, con la línea de forma fija `unavailable: <integration> — uncovered: <capability>; alternative: <path | none>`. El sistema MUST NOT presentar el resultado como equivalente al obtenido con la integración.

#### Scenario: Integración habilitada pero ausente

- **WHEN** la clase de tarea declara una integración y esa integración no está disponible
- **THEN** el sistema informa qué capacidad queda sin cubrir
- **AND** ofrece el camino alternativo declarado si existe

#### Scenario: Sin camino alternativo declarado

- **WHEN** falta una integración habilitada y no hay camino alternativo declarado
- **THEN** el sistema informa el límite
- **AND** no presenta el resultado parcial como completo

### Requirement: Invocación repetida registrada

El agente SHALL registrar como llamada innecesaria, con la línea de forma fija `unnecessary call: <tool> <input>`, toda invocación de una herramienta o integración con la misma entrada cuyo resultado ya está disponible en la sesión, sin que haya cambiado el estado que observa. La invocación repetida MUST NOT bloquearse, salvo el sondeo de una integración ya informada como no disponible, que MUST NOT repetirse en la misma tarea.

#### Scenario: Integración consultada dos veces por lo mismo

- **WHEN** el agente invoca una integración con la misma entrada dos veces y el estado que esa integración observa no cambió entre ambas
- **THEN** la segunda invocación queda registrada como llamada innecesaria

#### Scenario: Sondeo de una integración declarada no disponible

- **WHEN** una integración ya quedó informada como no disponible para la tarea activa
- **THEN** el agente no vuelve a invocarla en esa tarea
- **AND** usa el camino alternativo declarado si existe

### Requirement: Comparabilidad con y sin cada integración opcional

Una misma clase de tarea SHALL poder ejecutarse con y sin cada integración opcional, conservando la misma definición de aceptación. La diferencia observada MUST poder atribuirse a esa integración y no a un cambio de criterio.

#### Scenario: Misma tarea con y sin una integración opcional

- **WHEN** se ejecuta la misma clase de tarea con una integración opcional habilitada y luego deshabilitada
- **THEN** la definición de aceptación es idéntica en ambas corridas
- **AND** la diferencia de contexto efectivo queda atribuida a esa integración
