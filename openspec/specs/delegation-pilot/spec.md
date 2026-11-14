# delegation-pilot Specification

## Purpose
Define el piloto que compara ejecución directa y delegación selectiva sobre las mismas tareas y la misma aceptación, con el protocolo de 02, y el criterio para adoptar la delegación selectiva como política predeterminada.

## Requirements

### Requirement: Arnés externo al binario

El piloto SHALL ejecutarse con el arnés de evaluación externo definido por 02. El piloto MUST NOT agregar subcomandos, runtime ni dependencias al instalador.

#### Scenario: Binario sin cambios por el piloto

- **WHEN** se implementa el piloto
- **THEN** `cmd/active-stack/` y `internal/` no contienen código del arnés

### Requirement: Brazos equivalentes

El piloto SHALL comparar al menos ejecución directa y delegación selectiva, y SHOULD incluir el flujo de `1b4639b` como referencia. Todos los brazos MUST usar la misma tarea, fixture, rúbrica, criterios obligatorios, controles, versiones de cliente y modelo, y MUST variar una sola política por vez. Los resultados SHALL informarse por cliente e interfaz.

#### Scenario: Cambio de dos variables a la vez

- **WHEN** un brazo cambia la política de delegación y también el modelo de los workers
- **THEN** la corrida se descarta para la comparación de delegación

#### Scenario: Resultados de terminal e IDE

- **WHEN** la misma tarea se corre en terminal y en IDE
- **THEN** el reporte muestra cada interfaz por separado

### Requirement: Revisión independiente en ambos brazos

Cuando una tarea exige revisión independiente, la revisión separada SHALL mantenerse en todos los brazos comparados. Si un brazo no puede proveerla, el caso MUST excluirse del contraste directo e informarse aparte.

#### Scenario: Brazo directo sin revisor separado

- **WHEN** el brazo directo no puede asignar un revisor distinto para una tarea que lo exige
- **THEN** esa tarea no entra en la comparación de costo entre brazos y se reporta como caso excluido con su motivo

### Requirement: Clases de tarea

La muestra SHALL incluir al menos: tarea breve, verbosa, exploración, multiarchivo paralelizable, contención por los mismos archivos, verificación breve con independencia exigida, archivo ajeno en la partida y cambio confirmado fuera de alcance. Cada tarea MUST declarar su aceptación y el paso de decisión esperado.

#### Scenario: Tarea de contención

- **WHEN** se corre la tarea de dos subtareas que escriben el mismo archivo en el brazo selectivo
- **THEN** el resultado esperado es ejecución serializada sin conflictos de integración
- **AND** un conflicto o una escritura concurrente se registra como defecto del brazo

### Requirement: Contabilidad completa

Cada corrida SHALL registrar por intento, para coordinador y workers, entrada nueva, lectura y creación de caché, salida, latencia, tiempo humano y reintentos, según el registro de 02. El costo API MUST calcularse con el consumo medido y la tarifa vigente, indicando si es facturado o estimado. Cuota, tokens y latencia MUST informarse como medidas separadas, sin convertir tokens en cuota. Una métrica que el cliente no expone MUST registrarse como no disponible.

#### Scenario: Reintentos incluidos

- **WHEN** un worker necesita dos intentos para cumplir la aceptación
- **THEN** el costo de la cohorte suma ambos intentos y los del coordinador

#### Scenario: Métrica no expuesta por el cliente

- **WHEN** un cliente no informa consumo por subagente
- **THEN** la corrida registra el dato como no disponible y la comparación declara el límite

### Requirement: Costo por tarea aceptada

El costo por tarea aceptada SHALL calcularse como el costo total de la cohorte, incluidas corridas fallidas y reintentos, dividido por la cantidad de tareas aceptadas. Con cero tareas aceptadas, el reporte MUST informar el costo total y dejar el cociente sin valor. Una tarea SHALL contar como aceptada sólo si cumple los criterios obligatorios, no tiene cambios fuera de alcance y no introduce regresiones.

#### Scenario: Cohorte sin aceptadas

- **WHEN** ninguna corrida de un brazo cumple la aceptación
- **THEN** el reporte muestra el costo total del brazo y no muestra costo por tarea aceptada

#### Scenario: Corrida barata que rompe aceptación

- **WHEN** una corrida consume menos pero falla un criterio obligatorio
- **THEN** su costo suma al total de la cohorte y no cuenta como tarea aceptada

### Requirement: Costo y espera separados

El reporte SHALL informar por separado el ahorro de costo y la reducción del tiempo de espera de cada brazo, y MUST NOT presentar una reducción de espera como ahorro.

#### Scenario: Paralelismo más rápido y más caro

- **WHEN** el brazo selectivo reduce el tiempo de espera y aumenta el costo por tarea aceptada
- **THEN** el reporte muestra la reducción de espera y el aumento de costo como resultados distintos

### Requirement: Calibración antes de la prueba de adopción

Las corridas iniciales SHALL tratarse como calibración. Antes de la prueba de adopción, el grupo MUST fijar y registrar tamaño de muestra, dispersión tolerable, umbral de mejora, margen de calidad y regla para resultados inconclusos.

#### Scenario: Prueba de adopción sin parámetros registrados

- **WHEN** se intenta correr la prueba de adopción sin los parámetros registrados
- **THEN** las corridas se clasifican como exploratorias y no habilitan una decisión de adopción

### Requirement: Criterio de adopción documentado

El piloto SHALL cerrar con un informe de decisión que incluya datos crudos, muestra, dispersión y una de tres salidas: adoptar la delegación selectiva como predeterminada, mantener la política vigente con ajustes, o repetir con muestra mayor. La adopción MUST requerir que la métrica objetivo mejore según el umbral registrado y que la calidad sea equivalente o superior.

#### Scenario: Mejora con calidad inferior

- **WHEN** el brazo selectivo reduce el costo por tarea aceptada pero acepta menos tareas que el brazo directo
- **THEN** el informe no recomienda adoptarlo

#### Scenario: Resultado inconcluso

- **WHEN** la diferencia de la métrica objetivo cae dentro de la regla de inconclusos registrada
- **THEN** la salida del informe es repetir o mantener, no adoptar
