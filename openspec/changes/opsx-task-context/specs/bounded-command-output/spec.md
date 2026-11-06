# bounded-command-output

## Purpose

Define el formato de un resumen de salida de comando que conserva los fallos y el código de salida en lugar de truncar por tamaño, y las condiciones bajo las cuales la salida original completa sigue siendo recuperable.

## ADDED Requirements

### Requirement: El resumen retiene fallos y código de salida

El resumen de una salida de comando SHALL conservar los errores, los fallos de prueba y el código de salida. El resumen MUST NOT descartar una línea de fallo por razones de tamaño.

#### Scenario: Suite de tests con un fallo

- **WHEN** se resume la salida de una suite de pruebas que terminó con un fallo
- **THEN** el resumen contiene ese fallo y el código de salida distinto de cero

#### Scenario: Comando exitoso y verboso

- **WHEN** se resume la salida de un comando que terminó con código cero y produjo muchas líneas repetitivas
- **THEN** el resumen informa el código de salida
- **AND** descarta la repetición sin fallos

### Requirement: El resumen no declara un fallo que no ocurrió

El resumen SHALL declarar como fallo únicamente lo que la salida original registra como fallo. Una línea que se parece a un error sin serlo, o un texto de ruido que imita el formato de un fallo, MUST NOT aparecer en el resumen como fallo.

#### Scenario: Ruido que imita el formato de un fallo

- **WHEN** la salida original contiene líneas que imitan el formato de un error pero el comando terminó con código cero y sin fallos
- **THEN** el resumen no declara ningún fallo
- **AND** informa el código de salida cero

#### Scenario: Fallo real entre ruido que lo imita

- **WHEN** la salida original contiene un fallo real rodeado de líneas que imitan el formato de un error
- **THEN** el resumen declara el fallo real
- **AND** no declara las líneas de ruido como fallos adicionales

### Requirement: Un fallo sobrevive al resumen en cualquier posición

El resumen SHALL conservar un fallo relevante con independencia de su posición dentro de la salida original. Un fallo al inicio, al medio o al final MUST aparecer en el resumen. La regla existe porque la pérdida depende de la posición: lo que cae en el tramo central de una salida larga es lo que más se pierde, y por eso las tres posiciones se prueban por separado en lugar de dar por buena una sola.

#### Scenario: Fallo al inicio de una salida grande

- **WHEN** la salida original tiene miles de líneas y el fallo aparece en las primeras
- **THEN** el resumen contiene ese fallo

#### Scenario: Fallo al medio de una salida grande

- **WHEN** la salida original tiene miles de líneas y el fallo aparece en el tramo central
- **THEN** el resumen contiene ese fallo

#### Scenario: Fallo al final de una salida grande

- **WHEN** la salida original tiene miles de líneas y el fallo aparece en las últimas
- **THEN** el resumen contiene ese fallo

#### Scenario: Varios fallos dispersos

- **WHEN** la salida original contiene fallos en más de una posición
- **THEN** el resumen contiene cada uno de los fallos relevantes

### Requirement: Original recuperable por referencia estable

Todo resumen SHALL incluir una referencia que permita recuperar la salida original completa. La referencia MUST identificar sin ambigüedad el comando ejecutado y su corrida.

#### Scenario: Recuperación del original a partir del resumen

- **WHEN** se dispone de un resumen y se pide la salida original
- **THEN** la referencia del resumen permite obtener la salida completa de esa misma corrida

#### Scenario: Resumen sin referencia recuperable

- **WHEN** un resumen no incluye una referencia que permita recuperar el original
- **THEN** el resumen queda marcado como insuficiente como evidencia

### Requirement: Acceso y retención del original declarados (pendiente de grupo: política de salidas originales)

El original SHALL conservarse en una ruta de trabajo local no versionada, mientras el change esté activo y por un mínimo de siete días. El original MUST NOT publicarse ni enviarse fuera de la máquina. Vencida la retención, la referencia MUST informar que el original ya no está disponible en lugar de fallar en silencio.

#### Scenario: Original dentro del plazo de retención

- **WHEN** se pide el original de una corrida dentro del plazo declarado
- **THEN** el original está disponible para quien tiene acceso declarado

#### Scenario: Original fuera del plazo de retención

- **WHEN** se pide el original de una corrida cuyo plazo de retención ya venció
- **THEN** el sistema informa que el original no está disponible y desde cuándo

### Requirement: Datos sensibles redactados al capturar

Los valores que coincidan con las clases de dato sensible SHALL redactarse en el momento de capturar la salida, no solo al resumirla. Esas clases MUST ser las mismas que declara el piso de permisos del catálogo: el sistema MUST NOT mantener una lista propia para la redacción. Ninguna copia conservada MUST contener el valor sin redactar. La redacción MUST indicar que hubo una omisión en lugar de borrar la línea.

#### Scenario: Salida que contiene un secreto

- **WHEN** la salida original contiene un valor de una clase de dato sensible declarada
- **THEN** el resumen muestra la línea con el valor redactado e indica que el valor fue omitido

#### Scenario: El piso de permisos se amplía

- **WHEN** el piso de permisos agrega una clase de dato sensible
- **THEN** la redacción al capturar cubre esa clase nueva sin cambiar nada de esta capacidad

#### Scenario: El original conservado tampoco contiene el valor

- **WHEN** se recupera el original de una corrida cuya salida contenía un valor de clase sensible
- **THEN** el original conservado muestra ese valor redactado
- **AND** conserva la línea y su posición para poder diagnosticar el fallo

### Requirement: El resumen declara quién lo produjo (pendiente de grupo: productor del resumen)

Todo resumen SHALL declarar qué productor lo generó. El contrato MUST evaluarse por lo que el resumen conserva y no por quién lo produjo, de modo que cambiar el productor no invalide un resumen conforme.

#### Scenario: Resumen conforme producido por el agente

- **WHEN** el resumen lo produce el agente bajo instrucción y conserva todo lo que el contrato exige
- **THEN** el resumen se acepta
- **AND** declara al agente como productor

#### Scenario: Resumen no conforme atribuible

- **WHEN** un resumen omite un fallo que la salida original registra
- **THEN** el resumen se rechaza
- **AND** el productor declarado queda registrado junto al rechazo
