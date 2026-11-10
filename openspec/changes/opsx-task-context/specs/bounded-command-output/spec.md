# bounded-command-output

## Purpose

Define el formato de un resumen de salida de comando que conserva los fallos y el código de salida en lugar de truncar por tamaño, y las condiciones bajo las cuales la salida original completa sigue siendo recuperable. En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción más lo declarado en el catálogo; no hay subcomando ni runtime nuevo (`design.md` — D17).

## ADDED Requirements

### Requirement: Una salida verbosa se entrega como resumen

Una salida de comando de más líneas que el umbral declarado en el catálogo (300) SHALL entregarse como resumen más referencia al original. Una salida que no supera el umbral SHALL entregarse completa. El umbral MUST usarse como condición para resumir y MUST NOT usarse como tope de recorte del resumen.

#### Scenario: Salida por encima del umbral

- **WHEN** un comando produce una salida de más de 300 líneas
- **THEN** el agente entrega el resumen y la referencia al original
- **AND** no entrega la salida completa en su contexto de trabajo

#### Scenario: Salida dentro del umbral

- **WHEN** un comando produce una salida de 300 líneas o menos
- **THEN** el agente entrega la salida completa sin resumirla

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

Todo resumen SHALL incluir una referencia con cinco campos: comando, id de corrida, ruta del original, huella sha256 del original ya redactado y fecha de vencimiento. La referencia MUST identificar sin ambigüedad el comando ejecutado y su corrida, y la huella MUST coincidir con el original guardado.

#### Scenario: Recuperación del original a partir del resumen

- **WHEN** se dispone de un resumen y se pide la salida original dentro del plazo de retención
- **THEN** la ruta de la referencia lleva a la salida completa de esa misma corrida
- **AND** su huella sha256 coincide con la declarada en la referencia

#### Scenario: Resumen sin referencia recuperable

- **WHEN** un resumen no incluye la referencia o le falta alguno de sus cinco campos
- **THEN** el resumen queda marcado como insuficiente como evidencia

#### Scenario: Huella que no coincide

- **WHEN** la huella del archivo en la ruta de la referencia no coincide con la declarada
- **THEN** el resumen queda marcado como insuficiente como evidencia

### Requirement: Ubicación, acceso y retención del original declarados

El original SHALL guardarse en `<os.UserCacheDir>/active-stack/runs/<change>/`, fuera del repositorio, y conservarse hasta que el change se archive y siete días más. Solo el usuario local MUST acceder a él; MUST NOT publicarse ni enviarse fuera de la máquina. Vencida la retención, la referencia MUST informar que el original no está disponible y desde qué fecha.

#### Scenario: Original guardado fuera del repositorio

- **WHEN** se guarda el original de una corrida
- **THEN** queda bajo el directorio de caché del usuario
- **AND** no aparece en el árbol de trabajo del repositorio

#### Scenario: Original dentro del plazo de retención

- **WHEN** se pide el original de una corrida de un change activo, o de un change archivado hace siete días o menos
- **THEN** el original está disponible en la ruta de la referencia para el usuario local

#### Scenario: Original fuera del plazo de retención

- **WHEN** se pide el original de una corrida de un change archivado hace más de siete días
- **THEN** el agente informa que el original no está disponible y desde qué fecha

### Requirement: Datos sensibles redactados al capturar

Los valores que coincidan con la lista de patrones sensibles declarada en el catálogo SHALL redactarse antes de escribir el original a disco, no solo al resumirlo. Esa lista MUST ser única, su dueña es esta capacidad, y MUST cubrir todo lo que bloquea el piso de permisos de Claude. Ninguna copia conservada MUST contener el valor sin redactar. La redacción MUST indicar que hubo una omisión en lugar de borrar la línea.

#### Scenario: Salida que contiene un secreto

- **WHEN** la salida original contiene un valor que coincide con un patrón de la lista
- **THEN** el resumen muestra la línea con el valor redactado e indica que el valor fue omitido

#### Scenario: La lista cubre el piso de permisos de Claude

- **WHEN** el piso de permisos de Claude bloquea una ruta que ningún patrón de la lista cubre
- **THEN** la validación del catálogo falla y nombra la ruta no cubierta

#### Scenario: El original conservado tampoco contiene el valor

- **WHEN** se recupera el original de una corrida cuya salida contenía un valor sensible
- **THEN** el original conservado muestra ese valor redactado
- **AND** conserva la línea y su posición para poder diagnosticar el fallo

### Requirement: El resumen declara quién lo produjo

Todo resumen SHALL declarar qué productor lo generó. El contrato MUST evaluarse por lo que el resumen conserva y no por quién lo produjo, de modo que cambiar el productor no invalide un resumen conforme.

#### Scenario: Resumen conforme producido por el agente

- **WHEN** el resumen lo produce el agente bajo instrucción y conserva todo lo que el contrato exige
- **THEN** el resumen se acepta
- **AND** declara al agente como productor

#### Scenario: Resumen no conforme atribuible

- **WHEN** un resumen omite un fallo que la salida original registra
- **THEN** el resumen se rechaza
- **AND** el productor declarado queda registrado junto al rechazo
