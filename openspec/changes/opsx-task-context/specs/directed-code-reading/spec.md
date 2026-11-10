# directed-code-reading

## Purpose

Define cómo el agente lee código cuando necesita orientarse, acotando la lectura a la ruta, el rango o el símbolo que la tarea requiere en los archivos que superan el umbral, y exigiendo que toda afirmación sobre el código indique de dónde salió. En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción más lo declarado en el catálogo, y "registrar" es declararlo en el reporte de la acción (`design.md` — D17).

## ADDED Requirements

### Requirement: Lectura acotada en archivos que superan el umbral

Un archivo que no supera el umbral declarado en el catálogo (300 líneas) SHALL leerse completo. Un archivo que lo supera SHALL leerse por ruta, por rango de líneas o por símbolo, salvo que aplique una de las condiciones de lectura completa. Una lectura por símbolo SHALL devolver el símbolo completo —de su primera a su última línea—, la firma o encabezado de la declaración que lo contiene cuando esté anidado, y su procedencia; MUST NOT devolver más que eso.

#### Scenario: Archivo que no supera el umbral

- **WHEN** el agente necesita conocer una parte de un archivo de 300 líneas o menos
- **THEN** lo lee completo

#### Scenario: Orientación dentro de un archivo extenso

- **WHEN** el agente necesita ubicar una función dentro de un archivo que supera el umbral declarado
- **THEN** la lectura devuelve el símbolo completo, de su primera a su última línea, con su procedencia
- **AND** no devuelve el archivo completo ni los símbolos vecinos

#### Scenario: Símbolo anidado

- **WHEN** el símbolo pedido está dentro de otra declaración
- **THEN** la lectura devuelve el símbolo y la firma o encabezado de la declaración que lo contiene
- **AND** no devuelve el resto del cuerpo de esa declaración

#### Scenario: Búsqueda de usos de un símbolo

- **WHEN** el agente necesita saber dónde se usa un símbolo
- **THEN** la búsqueda devuelve las coincidencias con su ruta y su línea
- **AND** no requiere leer cada archivo que las contiene

### Requirement: Orientación en el repositorio sin lectura exhaustiva

Cuando el agente necesita ubicar dónde vive una capacidad y no sabe todavía en qué archivos está, SHALL orientarse con una búsqueda acotada sobre el repositorio. Cada ruta candidata MUST venir con su tipo de coincidencia —definición, referencia o texto— y su cantidad. El agente MUST NOT recorrer el repositorio archivo por archivo para orientarse.

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

### Requirement: Relectura de contenido ya disponible registrada

El agente SHALL registrar como llamada innecesaria toda relectura de un contenido que ya está disponible en la sesión: el mismo archivo y el mismo rango, sin que haya mediado una escritura sobre ese archivo. La relectura MUST NOT bloquearse.

#### Scenario: Misma lectura repetida sin cambios intermedios

- **WHEN** el agente lee el mismo archivo y el mismo rango dos veces y entre ambas lecturas no hubo escritura sobre ese archivo
- **THEN** la segunda lectura queda registrada como llamada innecesaria

#### Scenario: Relectura después de una escritura

- **WHEN** el agente lee un archivo, lo modifica y vuelve a leer el mismo rango
- **THEN** la segunda lectura se considera justificada

### Requirement: Lectura completa bajo condición declarada

Leer un archivo completo SHALL estar justificado solo si: no supera el umbral; la tarea va a reescribirlo entero; es el artefacto de la acción activa; una lectura acotada previa devolvió vacío; o los rangos no alcanzan y el motivo queda declarado en el reporte. La lista MUST ser cerrada. El umbral SHALL expresarse en líneas, con 300 como valor inicial medido. Una lectura completa sin condición aplicable o sin motivo declarado SHALL quedar registrada como no justificada.

#### Scenario: Archivo por debajo del umbral

- **WHEN** el archivo no supera el umbral declarado para lectura completa
- **THEN** la lectura completa se considera justificada

#### Scenario: Archivo que la tarea va a reescribir entero

- **WHEN** la tarea declara que va a reescribir el archivo completo
- **THEN** la lectura completa se considera justificada aunque el archivo supere el umbral

#### Scenario: Artefacto que la acción activa produce o edita

- **WHEN** el archivo supera el umbral y es el artefacto que la acción activa produce o edita
- **THEN** la lectura completa se considera justificada

#### Scenario: Lectura acotada previa sin resultado

- **WHEN** una lectura acotada sobre ese archivo devolvió vacío y el agente lee el archivo completo
- **THEN** la lectura completa se considera justificada

#### Scenario: Excepción justificada porque los rangos no alcanzan

- **WHEN** el agente lee por completo un archivo que supera el umbral porque ningún rango cubre lo que la tarea necesita
- **THEN** la lectura se considera justificada si el reporte declara la ruta y el motivo
- **AND** sin motivo declarado queda registrada como no justificada

#### Scenario: Lectura completa sin condición aplicable

- **WHEN** el agente lee por completo un archivo que supera el umbral y ninguna de las cinco condiciones aplica
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
