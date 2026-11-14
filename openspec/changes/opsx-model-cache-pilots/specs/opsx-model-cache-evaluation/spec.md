# OPSX Model and Cache Evaluation

## Purpose

Determinar de forma reproducible si una política de modelo y esfuerzo por rol o una estrategia de prefijos mejora el consumo por tarea aceptada en OPSX.

## ADDED Requirements

### Requirement: Comparación controlada e independiente

El arnés SHALL identificar el protocolo de 02 y el flujo de 07, y comparar cada intervención con un control de la misma cohorte. SHALL conservar tarea, revisión, rúbrica, cliente, superficie, modalidad, herramientas y política de delegación dentro de cada comparación.

#### Scenario: Comparación válida

- **WHEN** se emparejan corridas de control y candidata que solo difieren en la intervención declarada
- **THEN** el reporte conserva el par, la revisión de cada dependencia y el resultado por criterio

#### Scenario: Diferencia en un control

- **WHEN** una variante cambia otra dimensión experimental o incumple una precondición del protocolo
- **THEN** el resultado indica el desvío y excluye el par de la conclusión de mejora atribuible

### Requirement: Modelo y esfuerzo por rol

El piloto de modelos SHALL registrar configuración solicitada y efectiva del coordinador y los workers, con el alcance de la evidencia. SHALL mantener la estrategia de prefijo fija y distinguir los cambios aislados de modelo o esfuerzo de las variantes que combinen ambos.

#### Scenario: Variante de esfuerzo ejecutada

- **WHEN** el cliente confirma un cambio de esfuerzo en un rol y el modelo permanece fijo
- **THEN** el reporte identifica el rol y atribuye la comparación a la variante de esfuerzo

#### Scenario: Sustitución de modelo

- **WHEN** el cliente aplica un modelo distinto del previsto o no permite confirmar un parámetro requerido
- **THEN** el registro indica el desvío y el par queda fuera de la comparación confirmatoria para ese parámetro

### Requirement: Condiciones de caché verificadas

El piloto de prefijos SHALL fijar modelo, esfuerzo y roles, y registrar el estado previsto y observado de caché. SHALL separar entrada lógica, lecturas y escrituras cuando estén disponibles, y conservar el alcance de cada señal.

#### Scenario: Lectura de un prefijo preparado

- **WHEN** la telemetría confirma la reutilización del prefijo en el alcance y la ventana previstos
- **THEN** el reporte registra la lectura observada y el costo de preparación correspondiente según el protocolo

#### Scenario: Sesión nueva con caché previa

- **WHEN** se prepara una corrida fría pero se observa una lectura atribuible al prefijo bajo prueba
- **THEN** el resultado registra un desvío de la condición fría y conserva la evidencia

#### Scenario: Métrica ausente o solo agregada

- **WHEN** no es posible atribuir la lectura al prefijo que se evalúa
- **THEN** el reporte conserva el uso agregado disponible, marca la condición del prefijo como desconocida y evita atribuirle una mejora confirmada

### Requirement: Resultado por tarea aceptada

El reporte SHALL separar API de suscripción y calcular el indicador acordado por tarea aceptada con cobertura suficiente. SHALL incluir el consumo de coordinador, workers, preparación de caché aplicable, intentos fallidos y reintentos una sola vez. SHALL publicar por separado calidad, tiempo, cuota y cobertura.

#### Scenario: Cohorte completa

- **WHEN** una cohorte tiene tareas aceptadas y se puede medir el consumo de todas las corridas iniciadas
- **THEN** el reporte divide ese consumo y la preparación imputable por la cantidad de tareas aceptadas e informa unidad, fuente y alcance

#### Scenario: Indicador no calculable

- **WHEN** la cohorte no tiene aceptaciones o la métrica principal carece de cobertura suficiente
- **THEN** el reporte indica el motivo y conserva la evidencia y las demás métricas disponibles

### Requirement: Recomendación condicionada a calidad y evidencia

Cada piloto SHALL recomendar adoptar, repetir o descartar por cohorte con base en criterios acordados antes de las corridas. La adopción SHALL requerir una mejora suficiente en la métrica objetivo y calidad comparable o superior.

#### Scenario: Mejora con calidad comparable

- **WHEN** la variante cumple muestra y cobertura, alcanza el umbral de mejora y conserva todos los criterios obligatorios
- **THEN** el piloto puede recomendar su adopción para la cohorte evaluada y cita los resultados que sostienen esa recomendación

#### Scenario: Ahorro con regresión obligatoria

- **WHEN** la variante reduce el consumo pero introduce una regresión en un criterio obligatorio
- **THEN** el piloto descarta la adopción para esa cohorte y registra la regresión

#### Scenario: Resultado inconcluso

- **WHEN** el efecto o la evidencia no bastan para aplicar el criterio acordado
- **THEN** el reporte identifica el resultado como inconcluso y registra qué falta para repetirlo

### Requirement: Consumidor offline de calibración separado de confirmación

El consumidor offline SHALL permitir preparar brazos con decisiones pendientes sin inventar valores y SHALL mantener no disponible la adopción mientras no exista integración independiente de los contratos y resultados externos requeridos. Un Dataset SHALL representar una sola campaña.

#### Scenario: Decisiones pendientes y datos sintéticos
- **WHEN** el operador entrega un manifiesto de calibración con decisiones `null`
- **THEN** el CLI valida y reporta los datos disponibles, marca el contraste inconcluso y mantiene la adopción no disponible

#### Scenario: Identidad de fuentes sin certificación de ejecución
- **WHEN** se congelan archivos designados y se produce un reporte
- **THEN** se conservan bytes SHA256, orden, scope, pins de 02/07, fixtures/rúbrica, IDs de fuentes y huella del Dataset tipado
- **AND** ninguna de esas huellas acredita ejecución, carga del cliente, independencia o suficiencia de un recibo 06

### Requirement: Cobertura íntegra de costos y condiciones

El consumidor SHALL conservar todos los intentos iniciados, fallos, reintentos y preparación imputable una sola vez, SHALL distinguir desconocido de cero y SHALL separar condiciones asignadas, modalidades, basis y unidades antes de concluir sobre una cohorte.

#### Scenario: Agregado con workers incluidos
- **WHEN** el agregado declara incluir workers y también se entregan sus detalles
- **THEN** el total usa el agregado sin volver a sumar los workers y conserva la cobertura de detalle por separado

#### Scenario: Fragmento sin captura o preparación ambigua
- **WHEN** falta un costo o la preparación de un bloque no puede asignarse a una única condición fría/caliente
- **THEN** se conservan intentos y costos disponibles, la preparación se carga una vez y los ratios afectados permanecen incompletos

#### Scenario: Caché no confirmable
- **WHEN** la atribución es global o desconocida, falta aislamiento frío o existe contaminación fría
- **THEN** el cálculo de efecto permanece inconcluso y conserva costos y señales observadas

#### Scenario: Miss genuino y lectura parcial caliente
- **WHEN** dos brazos comparten condición caliente y existe señal atribuida al prefijo, uno con miss y otro con lectura parcial
- **THEN** ambos permanecen en el contraste asignado sin excluir el miss para favorecer la variante

#### Scenario: Control distinto del manifiesto o repetición duplicada
- **WHEN** las celdas observadas difieren del manifiesto congelado o una misma repetición aparece fragmentada por basis o unidad
- **THEN** el contraste se invalida o el Dataset se rechaza respectivamente, sin presentar el subconjunto como comparación completa
