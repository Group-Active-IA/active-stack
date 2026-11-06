# directed-code-reading

## Purpose

Define cómo el agente lee código cuando necesita orientarse, acotando la lectura a la ruta, el rango o el símbolo que la tarea requiere, y exigiendo que toda afirmación sobre el código indique de dónde salió.

## ADDED Requirements

### Requirement: Lectura acotada por defecto

Cuando el agente necesita conocer una parte del código, el sistema SHALL ofrecer lectura acotada por ruta, por rango de líneas o por símbolo. La lectura acotada MUST ser el comportamiento esperado para orientarse dentro de un archivo cuyo contenido no se conoce todavía. Una lectura por símbolo SHALL devolver el símbolo completo —de su primera a su última línea—, la declaración que lo contiene cuando esté anidado, y su procedencia; MUST NOT devolver más que eso.

#### Scenario: Orientación dentro de un archivo extenso

- **WHEN** el agente necesita ubicar una función dentro de un archivo que excede el umbral declarado
- **THEN** la lectura devuelve el símbolo completo, de su primera a su última línea, con su procedencia
- **AND** no devuelve el archivo completo ni los símbolos vecinos

#### Scenario: Símbolo anidado

- **WHEN** el símbolo pedido está dentro de otra declaración
- **THEN** la lectura devuelve el símbolo y la declaración que lo contiene
- **AND** no devuelve el resto del cuerpo de esa declaración

#### Scenario: Búsqueda de usos de un símbolo

- **WHEN** el agente necesita saber dónde se usa un símbolo
- **THEN** la lectura devuelve las coincidencias con su ruta y su línea
- **AND** no requiere leer cada archivo que las contiene

### Requirement: Orientación en el repositorio sin lectura exhaustiva

Cuando el agente necesita ubicar dónde vive una capacidad y no sabe todavía en qué archivos está, el sistema SHALL ofrecer una orientación acotada sobre el repositorio. Cada ruta candidata MUST venir con su tipo de coincidencia —definición, referencia o texto— y su cantidad. El agente MUST NOT recorrer el repositorio archivo por archivo para orientarse.

#### Scenario: Ubicar dónde vive una capacidad

- **WHEN** el agente necesita saber en qué archivos se implementa una capacidad que no conoce
- **THEN** la orientación devuelve las rutas candidatas, cada una con su tipo de coincidencia y su cantidad
- **AND** el agente no lee el contenido de los archivos que la orientación descarta

#### Scenario: Orden entre tipos de coincidencia

- **WHEN** una ruta contiene la definición del símbolo y otra contiene muchas referencias a él
- **THEN** la ruta con la definición se ordena antes que la de las referencias
- **AND** la cantidad solo ordena entre rutas del mismo tipo

#### Scenario: Orientación por recorrido exhaustivo

- **WHEN** el agente lee el contenido de archivos del repositorio uno por uno solo para orientarse
- **THEN** ese recorrido queda registrado como lectura no justificada

### Requirement: Ninguna relectura de contenido ya disponible

El sistema SHALL evitar volver a leer un contenido que ya está disponible en la sesión. Una relectura del mismo archivo y el mismo rango, sin que haya mediado una escritura sobre ese archivo, MUST quedar registrada como llamada innecesaria.

#### Scenario: Misma lectura repetida sin cambios intermedios

- **WHEN** el agente lee el mismo archivo y el mismo rango dos veces y entre ambas lecturas no hubo escritura sobre ese archivo
- **THEN** la segunda lectura queda registrada como llamada innecesaria

#### Scenario: Relectura después de una escritura

- **WHEN** el agente lee un archivo, lo modifica y vuelve a leer el mismo rango
- **THEN** la segunda lectura se considera justificada

### Requirement: Lectura completa bajo condición declarada

El sistema SHALL permitir la lectura de un archivo completo solo bajo una de estas cuatro condiciones: el archivo está por debajo del umbral declarado; la tarea va a reescribirlo entero; es el artefacto que la acción activa produce o edita; o una lectura acotada previa sobre ese archivo devolvió vacío. La lista MUST ser cerrada. El umbral SHALL expresarse en líneas. Una lectura completa sin condición aplicable SHALL quedar registrada como lectura no justificada.

#### Scenario: Archivo por debajo del umbral

- **WHEN** el archivo está por debajo del umbral declarado para lectura completa
- **THEN** la lectura completa se considera justificada

#### Scenario: Archivo que la tarea va a reescribir entero

- **WHEN** la tarea declara que va a reescribir el archivo completo
- **THEN** la lectura completa se considera justificada aunque el archivo exceda el umbral

#### Scenario: Lectura acotada previa sin resultado

- **WHEN** una lectura acotada sobre ese archivo devolvió vacío y el agente lee el archivo completo
- **THEN** la lectura completa se considera justificada

#### Scenario: Lectura completa sin condición aplicable

- **WHEN** el agente lee por completo un archivo que excede el umbral y ninguna de las cuatro condiciones aplica
- **THEN** la lectura queda registrada como no justificada

### Requirement: Procedencia obligatoria en afirmaciones sobre el código

Toda afirmación del agente sobre el contenido del código SHALL indicar su procedencia como ruta más línea o ruta más símbolo. Una afirmación sin procedencia MUST quedar marcada como no verificada.

#### Scenario: Hallazgo con procedencia

- **WHEN** el agente afirma que una función tiene un comportamiento determinado
- **THEN** la afirmación incluye la ruta y la línea o el símbolo donde se observa

#### Scenario: Hallazgo sin procedencia

- **WHEN** el agente afirma algo sobre el código sin indicar ruta ni línea ni símbolo
- **THEN** la afirmación queda marcada como no verificada
- **AND** no se cuenta como evidencia para aceptar un cambio
