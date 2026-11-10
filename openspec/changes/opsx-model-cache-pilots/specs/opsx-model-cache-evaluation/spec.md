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
