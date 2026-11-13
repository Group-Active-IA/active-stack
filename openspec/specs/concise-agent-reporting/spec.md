# concise-agent-reporting Specification

## Purpose

Define cómo el agente informa lo que encontró: el resultado y su procedencia primero, el relato que llevó hasta él solo si se lo piden. A diferencia de la salida de un comando, una respuesta del agente no tiene un original guardado al que referirse, así que el detalle se produce a pedido. Lo que se evalúa es conducta del modelo: se verifica con una rúbrica aplicada sobre fixtures, corridos con el arnés fuera del binario, y no con tests de Go (`design.md` — D16 y D17).


## Requirements

### Requirement: El resultado va en la primera línea, con su procedencia

Cuando el agente informa un hallazgo, la primera línea de su reporte SHALL decir el resultado con su procedencia, y la segunda SHALL decir la consecuencia o el próximo paso. El recorrido que llevó hasta el resultado MUST NOT precederlo.

#### Scenario: Hallazgo en el código

- **WHEN** el agente encuentra un parámetro inválido en una llamada
- **THEN** la primera línea del reporte dice el hallazgo con su ruta y su línea
- **AND** la segunda dice la consecuencia o el próximo paso
- **AND** ninguna línea anterior explica cómo llegó hasta él

#### Scenario: Varios hallazgos

- **WHEN** el agente encuentra más de un hallazgo en la misma tarea
- **THEN** los informa como una lista, cada uno con su procedencia
- **AND** la lista precede a cualquier desarrollo

### Requirement: El detalle se produce a pedido

El desarrollo de un hallazgo SHALL producirse cuando se lo pide, no por defecto. La respuesta inicial MUST NOT repetir el desarrollo de pasos anteriores de la misma tarea.

#### Scenario: Respuesta inicial

- **WHEN** el agente informa un hallazgo sin que se le pida el desarrollo
- **THEN** la respuesta contiene el resultado, su procedencia y la consecuencia
- **AND** no repite el recorrido de los pasos que llevaron a encontrarlo

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

Cuando el agente no encuentra nada, SHALL decirlo en una línea que nombre qué revisó. El agente MUST NOT narrar la búsqueda que no encontró nada.

#### Scenario: Búsqueda sin resultado

- **WHEN** el agente revisa lo que la tarea pedía y no encuentra ningún hallazgo
- **THEN** informa en una línea que no encontró nada y qué revisó
- **AND** no desarrolla el recorrido de la revisión
