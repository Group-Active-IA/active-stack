# OPSX Benchmarking

## Purpose

Establecer una referencia reproducible de calidad, contexto y consumo por tarea de OPSX en Claude Code y Codex, medidos por separado en terminal e IDE, para evaluar cambios posteriores.

## ADDED Requirements

### Requirement: Corridas reproducibles y aisladas

El arnés SHALL ejecutar cada repetición desde un fixture inmutable y un perfil de prueba aislado. El manifiesto MUST identificar la revisión, la rúbrica, la configuración efectiva, la variante, el cliente, la versión, el sistema, la interfaz real, el modo de permisos y los flags efectivos, el modelo y la modalidad de acceso. El manifiesto SHALL ser la única fuente de esos datos para los changes que reutilizan el protocolo.

#### Scenario: Repeticiones del mismo caso

- **WHEN** se repite una tarea con el mismo manifiesto
- **THEN** ambas corridas parten de la misma revisión, prompt y rúbrica
- **AND** cada una recibe un checkout de trabajo y una sesión nuevos

#### Scenario: Celda incompatible antes de iniciar

- **WHEN** la comprobación previa detecta que la versión, la interfaz, los permisos efectivos o la configuración no cumplen las precondiciones de la celda
- **THEN** el arnés no inicia corridas en esa celda y la registra como incompatible con su motivo
- **AND** la celda queda fuera del denominador y el resumen la lista aparte

#### Scenario: Opt-in de contenido activo

- **WHEN** la comprobación previa detecta un opt-in de registro de contenido activo en el cliente
- **THEN** la celda se registra como incompatible con ese motivo

### Requirement: Cohortes por interfaz

El arnés SHALL medir terminal e IDE desde la primera medición y publicar resultados separados por interfaz. Una terminal integrada del IDE que ejecuta la CLI SHALL pertenecer a la cohorte de terminal. Las aplicaciones de escritorio quedan fuera de la primera comparación.

#### Scenario: CLI en la terminal integrada del IDE

- **WHEN** una corrida ejecuta la CLI desde la terminal integrada de un IDE
- **THEN** el manifiesto registra la interfaz real como terminal y la corrida se agrega a la cohorte de terminal

#### Scenario: Métrica disponible solo en la CLI

- **WHEN** una señal existe en la cohorte de terminal pero la extensión del IDE no la expone
- **THEN** la métrica figura como no disponible en la cohorte IDE con su motivo
- **AND** la corrida se conserva para las métricas válidas

### Requirement: Permisos de medición

Las tareas con edición SHALL correr en perfiles aislados con permisos equivalentes para la tarea en cada cliente e interfaz. El perfil Estricto SHALL medirse en una serie separada. Cada corrida MUST registrar los permisos efectivos comprobados con el cliente real.

#### Scenario: Tarea con edición en una serie comparable

- **WHEN** se compara una tarea de cambio pequeño o multiarchivo entre variantes
- **THEN** ambas variantes corren con los mismos permisos efectivos registrados en el manifiesto

#### Scenario: Corrida bloqueada por permisos

- **WHEN** una corrida iniciada queda bloqueada por una denegación de permisos
- **THEN** se registra como bloqueada y no aceptada, con la causa
- **AND** su consumo entra en el costo total de la cohorte

### Requirement: Evidencia de aceptación por tarea

Cada tarea SHALL tener una rúbrica versionada con criterios obligatorios observables. Cada corrida iniciada SHALL terminar con un resultado de aceptación y referencias a la evidencia que lo justifica. El código de salida del cliente SHALL guardarse como dato bruto y no determina el veredicto.

#### Scenario: Tarea completada

- **WHEN** los chequeos y la revisión prevista confirman todos los criterios obligatorios
- **THEN** el resultado registra la tarea como aceptada y conserva el resultado por criterio

#### Scenario: Rechazo por calidad

- **WHEN** una prueba pasa pero un criterio obligatorio queda sin evidencia
- **THEN** la tarea se registra como rechazada por calidad y se conserva el criterio incumplido
- **AND** el estado es distinto del de una corrida bloqueada por permisos

#### Scenario: Timeout o fallo del cliente

- **WHEN** la corrida termina por timeout o el cliente emite un evento de fallo de turno o de error
- **THEN** la tarea se registra como no aceptada con la causa, el código de salida y los eventos de error

### Requirement: Rúbrica de revisión de un diff sembrado

La familia de revisión SHALL evaluarse contra una clave de defectos sembrados con archivo, rango de líneas y mecanismo de falla. Un hallazgo SHALL emparejar con un defecto cuando coinciden archivo, rango solapado y mecanismo, con emparejamiento uno a uno.

#### Scenario: Hallazgo emparejado

- **WHEN** un hallazgo señala el archivo, un rango solapado y el mecanismo de un defecto sembrado
- **THEN** se cuenta como verdadero positivo con su evidencia
- **AND** otro hallazgo sobre el mismo defecto no suma

#### Scenario: Defecto sin hallazgo

- **WHEN** un defecto sembrado no tiene hallazgo emparejado
- **THEN** se cuenta como omisión

#### Scenario: Hallazgo sin emparejar

- **WHEN** un hallazgo no empareja con ningún defecto sembrado
- **THEN** se cuenta como falso positivo si el revisor lo rechaza
- **AND** se registra como hallazgo nuevo para adjudicación separada si el revisor lo considera un defecto real

### Requirement: Contexto de sesión normal y variantes de OPSX

El arnés SHALL medir la carga de contexto observable de una sesión normal con una sonda fija sin invocar OPSX. Para la comparación de tareas, SHALL emparejar las variantes actual y modular dentro de un mismo cliente, interfaz y modalidad y con los mismos criterios de aceptación.

#### Scenario: Primera interacción de una sesión normal

- **WHEN** se ejecuta la sonda en una sesión nueva sin invocar el orquestador
- **THEN** se informa la entrada observable de la primera interacción y se identifica el alcance de la medición o de su aproximación

#### Scenario: Flujo modular todavía ausente

- **WHEN** el flujo modular no está disponible en la interfaz probada
- **THEN** la referencia actual se conserva y la comparación se marca pendiente

### Requirement: Registro de consumo y cobertura

Cada resultado SHALL registrar por intento la entrada, lecturas y escrituras de caché cuando se distingan, salida, lecturas de archivos, llamadas a herramientas por tipo, duración, intervención humana y atribución por rol según lo observable. Cada métrica MUST declarar su fuente, unidad, alcance y disponibilidad.

#### Scenario: Uso atribuible a coordinador y workers

- **WHEN** el cliente entrega eventos correlacionables de coordinador y workers, incluidas las llamadas de un reintento
- **THEN** el arnés conserva su atribución y contabiliza una sola vez cada unidad de uso en el total de la corrida

#### Scenario: Agregado que ya incluye a los hijos

- **WHEN** el agregado del coordinador incluye el consumo de los workers según la comprobación de la versión fijada
- **THEN** el total de la corrida es ese agregado y el detalle de los workers solo se usa para atribuir

#### Scenario: Cierre de worker con total parcial

- **WHEN** el evento de cierre de un worker informa un total que cubre solo su última solicitud
- **THEN** ese valor no se usa como consumo total del worker

#### Scenario: Discrepancia entre fuentes

- **WHEN** la suma del detalle por rol difiere del total de mayor alcance comprobado
- **THEN** si el detalle suma menos, la diferencia se registra como consumo sin atribuir y la atribución se marca parcial
- **AND** si el detalle suma más, la métrica se marca parcial con ambas cifras y queda fuera de los indicadores que exigen cobertura completa

#### Scenario: Atribución parcial o ausente

- **WHEN** solo se dispone de un total de sesión o se pierde parte de la captura
- **THEN** el resultado conserva el agregado con su alcance y declara el detalle faltante o parcial con su motivo
- **AND** un total incompleto queda excluido del indicador que exige cobertura completa

### Requirement: Medición del tiempo

El arnés SHALL medir el tiempo transcurrido de cada corrida con reloj monotónico desde el inicio de la tarea hasta su resultado y registrar aparte las esperas de aprobación observables, la intervención humana y los reintentos.

#### Scenario: Pausa conocida en una aprobación

- **WHEN** una corrida de calibración introduce una demora conocida en una aprobación controlada
- **THEN** el tiempo transcurrido incluye la demora y la espera registrada la refleja
- **AND** el tiempo ajustado la descuenta solo si los intervalos detenidos quedaron delimitados

#### Scenario: Esperas solapadas

- **WHEN** dos esperas se solapan o otro agente sigue trabajando durante una espera
- **THEN** los intervalos solapados se unen y el tiempo en que otro agente trabajó no se descuenta

#### Scenario: Espera no delimitable

- **WHEN** la interfaz no ofrece evidencia suficiente para delimitar los intervalos detenidos
- **THEN** el tiempo ajustado figura como no disponible y el transcurrido se conserva

### Requirement: Comparación por tarea aceptada

El reporte SHALL separar API y suscripción, mostrar la cobertura de cada indicador y contabilizar los intentos fallidos, bloqueados y reintentos dentro del consumo de la cohorte evaluada. SHALL vincular cada agregado con las corridas y el protocolo que lo produjeron.

#### Scenario: API con consumo completo y tareas aceptadas

- **WHEN** una cohorte de API tiene un costo medible para todas sus corridas iniciadas y al menos una tarea aceptada
- **THEN** el reporte divide el costo total de todas las corridas iniciadas por la cantidad de tareas aceptadas e indica si el costo es facturado o estimado

#### Scenario: Ninguna tarea aceptada

- **WHEN** una cohorte no tiene tareas aceptadas
- **THEN** el reporte publica el costo total de la cohorte y deja el cociente sin valor con su motivo

#### Scenario: Cuota no observable

- **WHEN** la cuota atribuible no está disponible
- **THEN** el indicador de cuota se informa como no disponible con su motivo
- **AND** el reporte mantiene visibles los resultados de calidad, tiempo y consumo disponibles

### Requirement: Evidencia apta para compartir

El resultado compartido SHALL contener solo los datos necesarios para reproducir y revisar el protocolo y SHALL aplicar la lista de datos sensibles y la política de captura, acceso y retención de 05. Las referencias a evidencia MUST usar las huellas y estados de procedencia de 06.

#### Scenario: Evidencia con datos sensibles

- **WHEN** un evento incluye correo, UUID de cuenta, identificador de usuario, rutas del host, detalle de herramientas o cuerpos de solicitudes
- **THEN** esos datos se excluyen del resultado compartido y se conserva una referencia revisable a evidencia autorizada o se registra su ausencia
