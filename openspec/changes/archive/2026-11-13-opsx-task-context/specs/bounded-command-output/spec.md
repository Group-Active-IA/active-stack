# bounded-command-output

## Purpose

Define quién captura una salida de comando antes de que el cliente la recorte, quién la redacta y quién la resume; el formato de un resumen que conserva los fallos y el código de salida en lugar de truncar por tamaño; y las condiciones bajo las cuales la copia redactada sigue siendo recuperable. Los clientes no le entregan al modelo la salida completa de un comando largo. Medido el 2026-11-13 (`entrega-03-04.md` §5): Claude Code 2.1.319 entrega un resultado fallido como unos 10.040 caracteres de cabeza y cola, sin ruta al resto, y esa cola no es el final real; Codex 0.167.1 entrega cabeza y cola sin avisarle al modelo que falta el medio. En `codex exec --json`, `aggregated_output` trae la salida completa (`design.md` — D3 y D22). En esta capacidad, "el sistema" es el agente siguiendo los módulos de instrucción, más lo declarado en el catálogo, más el arnés de evaluación fuera del binario; no hay subcomando ni runtime nuevo (`design.md` — D8 y D17).

## ADDED Requirements

### Requirement: Una salida verbosa se entrega como resumen

Una salida de comando que supere el umbral declarado en el catálogo —300 líneas o 10.000 caracteres, lo que ocurra primero (`design.md` — D3)— SHALL entregarse como resumen más referencia a la copia redactada. Una salida que no supera el umbral SHALL entregarse completa. El umbral MUST usarse como condición para resumir y MUST NOT usarse como tope de recorte del resumen. Este umbral MUST declararse aparte del umbral de lectura de archivos, aunque compartan las 300 líneas.

#### Scenario: Salida por encima del umbral en líneas

- **WHEN** un comando produce una salida de más de 300 líneas
- **THEN** el agente entrega el resumen y la referencia
- **AND** no entrega la salida completa en su contexto de trabajo

#### Scenario: Salida corta en líneas y larga en caracteres

- **WHEN** un comando produce menos de 300 líneas que suman más de 10.000 caracteres
- **THEN** el agente la trata como salida verbosa
- **AND** entrega el resumen y la referencia

#### Scenario: Salida dentro del umbral

- **WHEN** un comando produce una salida de 300 líneas o menos y de 10.000 caracteres o menos
- **THEN** el agente entrega la salida completa sin resumirla

### Requirement: Captura, redacción y resumen son roles declarados por separado

El registro de captura SHALL declarar quién capturó la salida, quién la redactó y quién produjo el resumen. La captura SHALL incluir stdout, stderr y el código de salida, y SHALL hacerla una fuente independiente del agente: el arnés o una fuente del cliente con cobertura probada para la celda. Una captura hecha por el agente MUST registrarse con fuente `agente`. El resumen SHALL salir de la copia capturada y MUST NOT citar la vista inline ni la copia del cliente.

#### Scenario: Captura del arnés en evaluación

- **WHEN** el arnés corre un fixture
- **THEN** el capturador del arnés registra stdout, stderr y el código de salida del comando
- **AND** el registro de captura declara `arnés` como fuente y como redactor, y al agente como productor del resumen
- **AND** el resumen lo repite en la línea de forma fija `capture: <source> · redactor: <source> · producer: <who>`

#### Scenario: Fallo al medio que el cliente habría recortado

- **WHEN** un comando que falla produce miles de líneas con el error en el tramo central
- **THEN** el error está en la copia capturada aunque el cliente haya mostrado solo cabeza y cola
- **AND** el resumen lo contiene

#### Scenario: Fallo que aparece solo por stderr

- **WHEN** un comando escribe su fallo únicamente por stderr y por stdout solo líneas sin fallos
- **THEN** la copia capturada contiene la línea de stderr
- **AND** el resumen conserva ese fallo y el código de salida distinto de cero, en Claude y en Codex

#### Scenario: Captura declarada por el agente

- **WHEN** no hay fuente independiente con cobertura probada para la celda y el agente corre un comando de verificación
- **THEN** el agente lo corre con la receta de su shell y el registro declara `agente` como fuente
- **AND** el resultado no supera *sin verificar*

#### Scenario: El código de salida sobrevive a la receta del shell

- **WHEN** el agente captura la salida de un comando que termina con código distinto de cero, en bash o en PowerShell
- **THEN** el código registrado es el del comando y no el del último paso de una tubería
- **AND** el archivo capturado conserva los bytes del comando, sin cambio de codificación ni envoltorio de errores del shell

### Requirement: La cobertura de cada fuente se demuestra por celda

Una fuente del cliente SHALL contar como independiente solo en las celdas (cliente, interfaz, sistema y versión) donde se demostró, frente a la referencia del arnés, que entrega la salida completa antes del truncamiento. La matriz SHALL cubrir como mínimo Claude y Codex, CLI e IDE, Windows y Linux, y distinguir celdas distribuidas, declaradas, probadas, truncadas, no disponibles y no probadas con su motivo. Una fuente que solo existe en CLI MUST figurar como no disponible en IDE.

#### Scenario: Fuente que coincide con la referencia

- **WHEN** la fuente de una celda entrega, para el fixture del fallo central y el de stderr, el mismo contenido que la referencia del arnés
- **THEN** la celda queda `probada` con el id de la corrida

#### Scenario: Fuente que trunca

- **WHEN** la fuente entrega la salida con una marca de truncamiento o sin el tramo central
- **THEN** la celda queda registrada como `truncada`
- **AND** esa fuente no cuenta como independiente en esa celda

#### Scenario: Celda sin corrida

- **WHEN** una celda tiene la fuente documentada pero ninguna corrida que la compare con la referencia
- **THEN** la celda figura `declarada` y no `probada`

#### Scenario: Sonda que coincide sin fixtures del arnés

- **WHEN** una fuente entregó la salida completa en una sonda con generadores propios, sin los fixtures del arnés
- **THEN** la celda figura `probada (sonda)`
- **AND** esa fuente todavía no cuenta como independiente

#### Scenario: Fuente que solo existe en CLI

- **WHEN** una fuente de captura está documentada para el CLI y no puede observarse desde la extensión de IDE
- **THEN** la celda de IDE figura `no disponible`
- **AND** la diferencia queda registrada en la matriz

### Requirement: Un resumen sin original capturado es parcial

Cuando el original no se pudo capturar, el resumen SHALL marcarse "parcial — sin verificar" con la línea de forma fija `partial — unverified: <reason>` y SHALL decir por qué; MUST NOT llevar ruta ni huella ni presentarse como completo. Si el comando no tiene efectos laterales, el agente MAY volver a correrlo con captura; si los tiene, MUST NOT repetirlo. Un original nunca capturado MUST distinguirse de uno vencido.

#### Scenario: Comando ya corrido sin captura

- **WHEN** el agente ya corrió un comando sin capturarlo y el cliente lo mostró recortado
- **THEN** el resumen queda marcado "parcial — sin verificar"
- **AND** dice que el original no se capturó y que lo visible es la vista recortada del cliente

#### Scenario: Archivo de captura ausente

- **WHEN** la ruta de corrida no contiene la copia que el registro declara
- **THEN** el resumen queda marcado "parcial — sin verificar"
- **AND** dice que la copia falta

#### Scenario: Original nunca capturado frente a original vencido

- **WHEN** se pide el original de un resumen parcial cuyo comando nunca se capturó
- **THEN** el agente informa que el original nunca existió en la ruta de corrida
- **AND** no lo informa como vencido ni da una fecha de vencimiento

#### Scenario: Superficie sin disco persistente

- **WHEN** la acción corre en una superficie sin disco persistente del usuario
- **THEN** el resumen queda marcado "parcial — sin verificar"
- **AND** la celda figura `no disponible` en la matriz de cobertura

### Requirement: El resumen retiene fallos y código de salida

El resumen de una salida de comando SHALL conservar los errores, los fallos de prueba y el código de salida. El resumen MUST NOT descartar una línea de fallo por razones de tamaño. Un fallo SHALL identificarse por el código de salida distinto de cero y por los patrones de fallo que el catálogo declara para la herramienta; si la herramienta no tiene patrones declarados, el resumen MUST decirlo.

#### Scenario: Suite de tests con un fallo

- **WHEN** se resume la salida de una suite de pruebas que terminó con un fallo
- **THEN** el resumen contiene ese fallo y el código de salida distinto de cero

#### Scenario: Comando exitoso y verboso

- **WHEN** se resume la salida de un comando que terminó con código cero y produjo muchas líneas repetitivas
- **THEN** el resumen informa el código de salida
- **AND** descarta la repetición sin fallos

#### Scenario: Herramienta sin patrones declarados

- **WHEN** se resume la salida de una herramienta para la que el catálogo no declara patrones de fallo
- **THEN** el resumen informa el código de salida
- **AND** dice que los fallos no se identificaron por patrón

### Requirement: El resumen no declara un fallo que no ocurrió

El resumen SHALL declarar como fallo únicamente lo que la salida original registra como fallo. Una línea que se parece a un error sin serlo, o un texto de ruido que imita el formato de un fallo, MUST NOT aparecer en el resumen como fallo.

#### Scenario: Ruido que imita el formato de un fallo

- **WHEN** la salida original contiene líneas que imitan el formato de un error pero el comando terminó con código cero y sin fallos
- **THEN** el resumen no declara ningún fallo: su única línea de fallo es la forma fija `failure: none`, sin agregados
- **AND** informa el código de salida cero

#### Scenario: Fallo real entre ruido que lo imita

- **WHEN** la salida original contiene un fallo real rodeado de líneas que imitan el formato de un error
- **THEN** el resumen declara el fallo real
- **AND** no declara las líneas de ruido como fallos adicionales

### Requirement: Un fallo sobrevive al resumen en cualquier posición

El resumen SHALL conservar un fallo relevante con independencia de su posición en la salida original. Un fallo al inicio, al medio o al final MUST aparecer en el resumen. Las tres posiciones se prueban por separado porque lo que cae en el tramo central de una salida larga es lo que más se pierde.

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

### Requirement: Copia recuperable por referencia estable

Todo resumen SHALL incluir una referencia con cinco campos: comando, id de corrida, ruta de la copia redactada, huella sha256 de esa copia con el algoritmo nombrado, y fecha de vencimiento, en la línea de forma fija `ref: cmd=<command> run=<run-id> path=<redacted copy> sha256:<hex> expires=<YYYY-MM-DD>`. Junto a la copia SHALL guardarse el registro de captura. La referencia MUST identificar sin ambigüedad el comando ejecutado y su corrida, y la huella MUST coincidir con la copia guardada.

#### Scenario: Recuperación de la copia a partir del resumen

- **WHEN** se dispone de un resumen y se pide la salida dentro del plazo de retención
- **THEN** la ruta de la referencia lleva a la copia redactada completa de esa misma corrida
- **AND** su huella sha256 coincide con la declarada en la referencia

#### Scenario: Resumen sin referencia recuperable

- **WHEN** un resumen no incluye la referencia o le falta alguno de sus cinco campos
- **THEN** el resumen queda marcado como insuficiente como evidencia

#### Scenario: Huella que no coincide

- **WHEN** la huella del archivo en la ruta de la referencia no coincide con la declarada
- **THEN** el resumen queda marcado como insuficiente como evidencia

### Requirement: Ubicación, acceso y retención de la copia declarados

La copia redactada SHALL guardarse en `<os.UserCacheDir>/active-stack/runs/<change>/<id de corrida>/`, con `_sin-change` cuando no hay change activo, fuera del repositorio. SHALL conservarse hasta que el change se archive y siete días más, o siete días desde la captura en `_sin-change`. Solo el usuario local MUST acceder a la copia; MUST NOT publicarse ni enviarse fuera de la máquina. Vencida la retención, la referencia MUST informar que la copia no está disponible y desde qué fecha.

#### Scenario: Copia guardada fuera del repositorio

- **WHEN** se guarda la copia de una corrida
- **THEN** queda bajo el directorio de caché del usuario
- **AND** no aparece en el árbol de trabajo del repositorio

#### Scenario: Acción sin change activo

- **WHEN** se captura una salida durante una acción sin change activo
- **THEN** la copia queda bajo `_sin-change` y vence siete días después de la captura

#### Scenario: Copia dentro del plazo de retención

- **WHEN** se pide la copia de una corrida de un change activo, o de un change archivado hace siete días o menos
- **THEN** la copia está disponible en la ruta de la referencia para el usuario local

#### Scenario: Copia fuera del plazo de retención

- **WHEN** se pide la copia de una corrida de un change archivado hace más de siete días
- **THEN** se informa que la copia no está disponible y desde qué fecha

### Requirement: Quien escribe la copia la borra al vencer

El capturador que escribió una copia SHALL borrarla cuando vence. Cuando no hay un capturador que pueda borrarla, la retención MUST declararse como recomendación y la referencia SHALL seguir informando el vencimiento.

#### Scenario: El arnés purga lo vencido

- **WHEN** el arnés inicia una corrida y hay copias vencidas en su directorio
- **THEN** las borra antes de capturar
- **AND** sus referencias siguen informando la fecha de vencimiento

### Requirement: Datos sensibles redactados antes de persistir

Los valores que coincidan con la lista única de patrones sensibles del catálogo, cuya dueña es esta capacidad y que MUST cubrir el piso de permisos de Claude, SHALL redactarse antes de persistir la copia. Por defecto SHALL persistirse solo la copia redactada, y ninguna copia que administre Active Stack MUST contener el valor sin redactar. La redacción MUST marcar la omisión sin borrar la línea. Las copias propias del cliente MUST declararse como límite.

#### Scenario: Fixture con datos sensibles ficticios

- **WHEN** el arnés captura el fixture que contiene un valor ficticio por cada patrón de la lista y un fallo real
- **THEN** ningún archivo de la ruta de corrida contiene un valor ficticio sin redactar
- **AND** cada línea afectada conserva su posición con la marca de omisión, y el fallo real sigue en la copia

#### Scenario: Salida que contiene un secreto

- **WHEN** la salida original contiene un valor que coincide con un patrón de la lista
- **THEN** el resumen muestra la línea con el valor redactado e indica que el valor fue omitido

#### Scenario: La lista cubre el piso de permisos de Claude

- **WHEN** el piso de permisos de Claude bloquea una ruta que ningún patrón de la lista cubre
- **THEN** la validación del catálogo falla y nombra la ruta no cubierta

#### Scenario: Captura declarada con crudo transitorio

- **WHEN** el agente captura con la receta de su shell y redacta el archivo
- **THEN** en la ruta de corrida queda solo la copia redactada cuando entrega el resumen
- **AND** el registro declara al agente como redactor

#### Scenario: Copia propia del cliente

- **WHEN** el cliente guardó por su cuenta una copia de la salida con el valor sensible
- **THEN** el agente no la presenta como redactada ni la cita como referencia
- **AND** el límite queda declarado en el reporte de la acción

### Requirement: Un detalle redactado no alcanza para verificar

Si verificar un requisito exige un detalle que la redacción descartó, el resultado SHALL quedar sin verificar hasta conseguir evidencia independiente suficiente.

#### Scenario: Detalle descartado necesario para verificar

- **WHEN** verificar un requisito exige un valor que la redacción reemplazó
- **THEN** el resultado queda *sin verificar*
- **AND** el reporte dice qué detalle falta y que hace falta evidencia independiente

### Requirement: Un resultado declarado por el agente queda sin verificar

Un resultado sobre una salida de comando que declara el agente —el coordinador sobre su propio comando o un worker al coordinador— SHALL quedar *sin verificar* hasta que se corrobore de forma independiente la ejecución y el estado del código. Una huella coincidente prueba la integridad de la copia y MUST NOT tomarse por sí sola como prueba de procedencia. Los estados MUST ser los de la ficha 06, y qué evidencia alcanza para el requisito MUST quedar a cargo de esa ficha.

#### Scenario: Resultado recibido sin corroborar

- **WHEN** un worker entrega al coordinador un resumen de salida y el coordinador todavía no lo comparó con nada
- **THEN** el resumen figura como *sin verificar* en el reporte del coordinador

#### Scenario: Ejecución directa con captura declarada

- **WHEN** el coordinador corre un comando sin fuente independiente y resume su propia captura
- **THEN** el resultado figura como *sin verificar*

#### Scenario: Conjunto coherente fabricado por la misma parte

- **WHEN** la copia, la huella y el resumen los produjo el mismo agente y son coherentes entre sí
- **THEN** el resultado sigue *sin verificar*
- **AND** el reporte dice que falta procedencia independiente

### Requirement: La corroboración exige procedencia independiente

Un resultado declarado por el agente SHALL pasar a *verificado* solo si la copia la capturó una fuente independiente con cobertura probada para esa celda, o si otra parte volvió a correr el comando con captura independiente. Con un fallo omitido o una huella distinta, SHALL pasar a *rechazado*.

#### Scenario: Corroboración con captura independiente

- **WHEN** la copia la capturó el arnés o una fuente del cliente con cobertura probada, la huella coincide y los fallos del resumen son los de la copia
- **THEN** el resultado pasa a *verificado* en cuanto a ejecución y estado del código

#### Scenario: Corroboración que encuentra un fallo omitido

- **WHEN** otra parte vuelve a correr el comando con captura independiente y aparece un fallo que el resumen no declaraba
- **THEN** el resultado pasa a *rechazado*
- **AND** el productor declarado queda registrado junto al rechazo

#### Scenario: Resumen no conforme atribuible

- **WHEN** un resumen omite un fallo que la copia registra
- **THEN** el resumen se rechaza
- **AND** el productor declarado queda registrado junto al rechazo
