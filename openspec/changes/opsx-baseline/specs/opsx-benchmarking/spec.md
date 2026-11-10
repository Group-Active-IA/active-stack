# OPSX Benchmarking

## Purpose

Establecer una referencia reproducible de calidad, contexto y consumo por tarea de OPSX en Claude Code y Codex para evaluar cambios posteriores.

## ADDED Requirements

### Requirement: Corridas reproducibles y aisladas

El arnés SHALL ejecutar cada repetición desde un fixture inmutable y un perfil de prueba aislado. El registro MUST identificar la revisión, la rúbrica, la configuración efectiva, la variante, el cliente, la superficie, la versión, el modelo y la modalidad de acceso.

#### Scenario: Repeticiones del mismo caso

- **WHEN** se repite una tarea con el mismo manifiesto
- **THEN** ambas corridas parten de la misma revisión, prompt y rúbrica
- **AND** cada una recibe un checkout de trabajo y una sesión nuevos

#### Scenario: Configuración incompatible

- **WHEN** la versión o el perfil elegido no cumple las precondiciones verificadas para la celda
- **THEN** el arnés registra la celda como no ejecutable con el motivo y no la contabiliza como una tarea intentada

### Requirement: Evidencia de aceptación por tarea

Cada tarea SHALL tener una rúbrica versionada con criterios obligatorios observables. Cada corrida iniciada SHALL terminar con un resultado de aceptación y referencias a la evidencia que lo justifica.

#### Scenario: Tarea completada

- **WHEN** los chequeos y la revisión prevista confirman todos los criterios obligatorios
- **THEN** el resultado registra la tarea como aceptada y conserva el resultado por criterio

#### Scenario: Verificación insuficiente

- **WHEN** una prueba pasa pero un criterio obligatorio queda sin evidencia, o la corrida termina por timeout
- **THEN** la tarea se registra como no aceptada y se conserva la causa

### Requirement: Contexto de sesión normal y variantes de OPSX

El arnés SHALL medir la carga de contexto observable de una sesión normal con una sonda fija sin invocar OPSX. Para la comparación de tareas, SHALL emparejar las variantes actual y modular dentro de una misma superficie y modalidad y con los mismos criterios de aceptación.

#### Scenario: Primera interacción de una sesión normal

- **WHEN** se ejecuta la sonda en una sesión nueva sin invocar el orquestador
- **THEN** se informa la entrada observable de la primera interacción y se identifica el alcance de la medición o de su aproximación

#### Scenario: Flujo modular todavía ausente

- **WHEN** el flujo modular no está disponible en la superficie probada
- **THEN** la referencia actual se conserva y la comparación se marca pendiente

### Requirement: Registro de consumo y cobertura

Cada resultado SHALL registrar por intento la entrada, lecturas y escrituras de caché cuando se distingan, salida, duración, intervención humana y atribución por rol según lo observable. Cada métrica MUST declarar su fuente, unidad, alcance y disponibilidad.

#### Scenario: Uso atribuible a coordinador y workers

- **WHEN** el cliente entrega eventos correlacionables de coordinador y workers, incluidas las llamadas de un reintento
- **THEN** el arnés conserva su atribución y contabiliza una sola vez cada unidad de uso en el total de la corrida

#### Scenario: Atribución parcial o ausente

- **WHEN** solo se dispone de un total de sesión o se pierde parte de la captura
- **THEN** el resultado conserva el agregado con su alcance y declara el detalle faltante o parcial con su motivo
- **AND** un total incompleto queda excluido del indicador que exige cobertura completa

### Requirement: Comparación por tarea aceptada

El reporte SHALL separar API y suscripción, mostrar la cobertura de cada indicador y contabilizar los intentos fallidos y reintentos dentro del consumo de la cohorte evaluada. SHALL vincular cada agregado con las corridas y el protocolo que lo produjeron.

#### Scenario: API con consumo completo y tareas aceptadas

- **WHEN** una cohorte de API tiene un costo medible para todas sus corridas y al menos una tarea aceptada
- **THEN** el reporte divide el costo total de todas las corridas por la cantidad de tareas aceptadas e indica si el costo es facturado o estimado

#### Scenario: Cuota no observable o ninguna aceptación

- **WHEN** la cuota atribuible no está disponible o una cohorte no tiene tareas aceptadas
- **THEN** el indicador correspondiente se informa como no calculable con su motivo
- **AND** el reporte mantiene visibles los resultados de calidad, tiempo y consumo disponibles

### Requirement: Evidencia apta para compartir

El resultado compartido SHALL contener solo los datos necesarios para reproducir y revisar el protocolo. El arnés MUST aplicar la política de acceso y retención acordada a cualquier evidencia original.

#### Scenario: Evidencia con datos sensibles

- **WHEN** un evento incluye credenciales, identificadores de cuenta o rutas personales
- **THEN** los datos sensibles se excluyen del resultado compartido y se conserva una referencia revisable a evidencia autorizada o se registra su ausencia
