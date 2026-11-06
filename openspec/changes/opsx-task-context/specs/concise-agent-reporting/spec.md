# concise-agent-reporting

## Purpose

Define cómo el agente informa lo que encontró: el hallazgo y su procedencia primero, el relato que llevó hasta él solo si se lo piden. A diferencia de la salida de un comando, una respuesta del agente no tiene un original guardado al que referirse, así que el detalle no se conserva: se produce a pedido.

## ADDED Requirements

### Requirement: El hallazgo va primero, con su procedencia

Cuando el agente informa un hallazgo, lo primero que entrega SHALL ser el hallazgo y su procedencia. El recorrido que llevó hasta él MUST NOT precederlo.

#### Scenario: Hallazgo en el código

- **WHEN** el agente encuentra un parámetro inválido en una llamada
- **THEN** informa primero el hallazgo con su ruta y su línea
- **AND** no precede el hallazgo con la explicación de cómo llegó hasta él

#### Scenario: Varios hallazgos

- **WHEN** el agente encuentra más de un hallazgo en la misma tarea
- **THEN** los informa como una lista, cada uno con su procedencia
- **AND** la lista precede a cualquier desarrollo

### Requirement: El detalle se produce a pedido

El desarrollo de un hallazgo SHALL producirse cuando se lo pide, no por defecto. El agente MUST NOT conservar un desarrollo previo para entregarlo después: a diferencia de la salida de un comando, una respuesta no tiene original guardado.

#### Scenario: Respuesta inicial

- **WHEN** el agente informa un hallazgo sin que se le pida el desarrollo
- **THEN** la respuesta contiene el hallazgo, su procedencia y la consecuencia
- **AND** no contiene el recorrido completo que llevó a encontrarlo

#### Scenario: Desarrollo pedido después

- **WHEN** se pide el desarrollo de un hallazgo ya informado
- **THEN** el agente lo produce en ese momento sobre la misma procedencia

### Requirement: Una respuesta no repite lo que ya está escrito

Cuando el contenido ya está en un artefacto del proyecto, la respuesta SHALL señalarlo en lugar de reproducirlo. El agente MUST NOT copiar en su respuesta un bloque que el artefacto ya contiene.

#### Scenario: Contenido que vive en un artefacto

- **WHEN** lo que se informa ya está escrito en un artefacto del proyecto
- **THEN** la respuesta nombra el artefacto y la parte que corresponde
- **AND** no reproduce su contenido

#### Scenario: Contenido que todavía no está escrito

- **WHEN** lo que se informa no está en ningún artefacto
- **THEN** la respuesta lo contiene

### Requirement: Sin hallazgo, no hay informe

Cuando el agente no encuentra nada, SHALL decirlo en una línea. El agente MUST NOT narrar la búsqueda que no encontró nada.

#### Scenario: Búsqueda sin resultado

- **WHEN** el agente revisa lo que la tarea pedía y no encuentra ningún hallazgo
- **THEN** informa en una línea que no encontró nada y qué revisó
- **AND** no desarrolla el recorrido de la revisión
