# bounded-command-output

## Purpose

Define cómo se captura una salida de comando antes de que el cliente la recorte, el formato de un resumen que conserva los fallos y el código de salida en lugar de truncar por tamaño, y las condiciones bajo las cuales la salida original completa sigue siendo recuperable. Los clientes no le entregan al modelo la salida completa de un comando largo: Claude Code entrega un resultado fallido de más de unos 10.000 caracteres como cabeza y cola, sin ruta al resto (https://code.claude.com/docs/en/tools-reference), y Codex recorta el medio dentro de `tool_output_token_limit` (https://learn.chatgpt.com/docs/config-file/config-reference); por eso el resumen se produce desde el original capturado (`design.md` — D8). En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción más lo declarado en el catálogo; no hay subcomando ni runtime nuevo (`design.md` — D17).

## ADDED Requirements

### Requirement: Una salida verbosa se entrega como resumen

Una salida de comando que supere el umbral declarado en el catálogo —300 líneas o 10.000 caracteres, lo que ocurra primero (`design.md` — D3)— SHALL entregarse como resumen más referencia al original. Una salida que no supera el umbral SHALL entregarse completa. El umbral MUST usarse como condición para capturar y resumir y MUST NOT usarse como tope de recorte del resumen. Este umbral MUST declararse aparte del umbral de lectura de archivos, aunque compartan las 300 líneas.

#### Scenario: Salida por encima del umbral en líneas

- **WHEN** un comando produce una salida de más de 300 líneas
- **THEN** el agente entrega el resumen y la referencia al original
- **AND** no entrega la salida completa en su contexto de trabajo

#### Scenario: Salida corta en líneas y larga en caracteres

- **WHEN** un comando produce menos de 300 líneas que suman más de 10.000 caracteres
- **THEN** el agente la trata como salida verbosa
- **AND** entrega el resumen y la referencia al original

#### Scenario: Salida dentro del umbral

- **WHEN** un comando produce una salida de 300 líneas o menos y de 10.000 caracteres o menos
- **THEN** el agente entrega la salida completa sin resumirla

### Requirement: La salida se captura antes de que el cliente la recorte

Un comando largo o verboso SHALL ejecutarse redirigiendo stdout y stderr al mismo archivo de la ruta de corrida y capturando su código de salida en la misma invocación, con la receta declarada para bash y PowerShell y sin subcomando nuevo del binario. El agente SHALL resumir desde ese archivo con lectura dirigida, y MUST NOT resumir desde lo que el cliente mostró inline ni citar como referencia la copia propia del cliente.

#### Scenario: Comando verboso capturado a archivo

- **WHEN** el agente corre una suite de tests que el inventario declara verbosa
- **THEN** la salida combinada y el código de salida quedan en un archivo de la ruta de corrida
- **AND** el resumen se produce leyendo ese archivo, no la vista que el cliente mostró

#### Scenario: Fallo al medio que el cliente habría recortado

- **WHEN** un comando que falla produce miles de líneas con el error en el tramo central
- **THEN** el error está en el archivo capturado aunque el cliente haya mostrado solo cabeza y cola
- **AND** el resumen lo contiene

#### Scenario: Fallo que aparece solo por stderr

- **WHEN** un comando escribe su fallo únicamente por stderr y por stdout solo líneas sin fallos
- **THEN** el archivo capturado contiene la línea de stderr
- **AND** el resumen conserva ese fallo y el código de salida distinto de cero, en Claude y en Codex

#### Scenario: El código de salida sobrevive a la receta del shell

- **WHEN** el agente captura la salida de un comando que termina con código distinto de cero, en bash o en PowerShell
- **THEN** el código registrado es el del comando y no el del último paso de una tubería

### Requirement: Un resumen sin original capturado es parcial

Cuando el original no se pudo capturar, el resumen SHALL marcarse "parcial — sin verificar" y SHALL decir por qué; MUST NOT llevar ruta ni huella ni presentarse como completo. Si el comando no tiene efectos laterales, el agente MAY volver a correrlo con captura; si los tiene, MUST NOT repetirlo. Un original nunca capturado MUST distinguirse de uno vencido.

#### Scenario: Comando ya corrido sin redirección

- **WHEN** el agente ya corrió un comando sin redirigir su salida y el cliente la mostró recortada
- **THEN** el resumen queda marcado "parcial — sin verificar"
- **AND** dice que el original no se capturó y que lo visible es la vista recortada del cliente

#### Scenario: Archivo de captura ausente

- **WHEN** la ruta de corrida no contiene el archivo de la captura que el agente declaró
- **THEN** el resumen queda marcado "parcial — sin verificar"
- **AND** dice que el archivo falta

#### Scenario: Original nunca capturado frente a original vencido

- **WHEN** se pide el original de un resumen parcial cuyo comando nunca se capturó
- **THEN** el agente informa que el original nunca existió en la ruta de corrida
- **AND** no lo informa como vencido ni da una fecha de vencimiento

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

El resumen SHALL conservar un fallo relevante con independencia de su posición en la salida original, y SHALL producirse desde el original capturado. Un fallo al inicio, al medio o al final MUST aparecer en el resumen. Las tres posiciones se prueban por separado porque lo que cae en el tramo central de una salida larga es lo que más se pierde.

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

Los valores que coincidan con la lista única de patrones sensibles del catálogo, cuya dueña es esta capacidad y que MUST cubrir el piso de permisos de Claude, SHALL redactarse antes de conservar el original. Ninguna copia que conserve Active Stack MUST contener el valor sin redactar, y el crudo de la captura MUST borrarse antes de entregar el resumen. La redacción MUST marcar la omisión sin borrar la línea. Las copias propias del cliente MUST declararse como límite.

#### Scenario: Salida que contiene un secreto

- **WHEN** la salida original contiene un valor que coincide con un patrón de la lista
- **THEN** el resumen muestra la línea con el valor redactado e indica que el valor fue omitido

#### Scenario: La lista cubre el piso de permisos de Claude

- **WHEN** el piso de permisos de Claude bloquea una ruta que ningún patrón de la lista cubre
- **THEN** la validación del catálogo falla y nombra la ruta no cubierta

#### Scenario: El crudo de la captura no se conserva

- **WHEN** termina la redacción de una salida capturada
- **THEN** en la ruta de corrida queda solo el original redactado
- **AND** el archivo crudo ya no existe cuando se entrega el resumen

#### Scenario: Copia propia del cliente

- **WHEN** el cliente guardó por su cuenta una copia de la salida con el valor sensible
- **THEN** el agente no la presenta como redactada ni la cita como referencia
- **AND** el límite queda declarado en el reporte de la acción

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

### Requirement: Un resumen de un worker sin corroborar queda sin verificar

El resumen que un worker entrega al coordinador SHALL quedar *sin verificar* hasta que el coordinador lo corrobore contra el original o su referencia, o con una re-ejecución con captura si el comando no tiene efectos laterales. Corroborado, SHALL pasar a *verificado*; con un fallo omitido o una huella distinta, a *rechazado*. Los estados MUST ser los de la ficha 06, y el coordinador MUST NOT presentar como verificado un resumen que no corroboró.

#### Scenario: Resumen de un worker recibido sin corroborar

- **WHEN** un worker entrega al coordinador un resumen de salida y el coordinador todavía no lo comparó con nada
- **THEN** el resumen figura como *sin verificar* en el reporte del coordinador

#### Scenario: Corroboración contra la referencia

- **WHEN** el coordinador comprueba que la ruta de la referencia existe, que la huella coincide y que los fallos declarados están en el original
- **THEN** el resumen pasa a *verificado*

#### Scenario: Corroboración que encuentra un fallo omitido

- **WHEN** el coordinador vuelve a correr el comando con captura y aparece un fallo que el resumen del worker no declaraba
- **THEN** el resumen pasa a *rechazado*
- **AND** el productor declarado queda registrado junto al rechazo
