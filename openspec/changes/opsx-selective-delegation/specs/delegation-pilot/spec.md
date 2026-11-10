# delegation-pilot

## Purpose

Define el piloto que compara ejecución directa y delegación selectiva sobre las mismas tareas y la misma aceptación, y el criterio para adoptar la delegación selectiva como política predeterminada.

## ADDED Requirements

### Requirement: Arnés externo al binario

El piloto SHALL ejecutarse con un arnés de evaluación externo al binario `active-stack`. El piloto MUST NOT agregar subcomandos, runtime ni dependencias al instalador.

#### Scenario: Binario sin cambios por el piloto

- **WHEN** se implementa el arnés del piloto
- **THEN** `cmd/active-stack/` y `internal/` no contienen código del arnés

### Requirement: Brazos comparables

El piloto SHALL comparar al menos ejecución directa y delegación selectiva, y SHOULD incluir el flujo instalado en `1b4639b` como referencia. Todos los brazos MUST usar las mismas tareas, el mismo commit de partida, los mismos criterios de aceptación y las mismas versiones de cliente y modelo, y MUST variar una sola política por vez.

#### Scenario: Cambio de dos variables a la vez

- **WHEN** un brazo cambia la política de delegación y también el modelo de los workers
- **THEN** la corrida se descarta para la comparación de delegación

### Requirement: Clases de tarea

La muestra SHALL incluir al menos una tarea de cada clase: breve, verbosa, exploración, multiarchivo paralelizable y contención por los mismos archivos. Cada tarea MUST declarar su aceptación y la señal o contraseñal esperada.

#### Scenario: Tarea de contención

- **WHEN** se corre la tarea de dos subtareas que escriben el mismo archivo en el brazo selectivo
- **THEN** el resultado esperado es ejecución serializada sin conflictos de integración
- **AND** un conflicto o una escritura concurrente se registra como defecto del brazo

### Requirement: Contabilidad completa

Cada corrida SHALL registrar, por coordinador y por cada worker, entrada sin caché, entrada cacheada, salida, reintentos, tiempo de pared y tiempo humano, con costo API y consumo de cuota por separado según el protocolo de 02. Una métrica que el cliente no expone MUST registrarse como ausente con su motivo y MUST NOT estimarse en silencio.

#### Scenario: Reintentos incluidos

- **WHEN** un worker necesita dos intentos para cumplir la aceptación
- **THEN** el costo de la tarea suma ambos intentos y los del coordinador

#### Scenario: Métrica no expuesta por el cliente

- **WHEN** un cliente no informa consumo por subagente
- **THEN** la corrida registra el dato como ausente y la comparación declara el límite

### Requirement: Puerta de calidad antes del costo

Una corrida SHALL contar como tarea aceptada sólo si cumple su aceptación, no tiene cambios fuera de alcance y no introduce regresiones. El costo MUST compararse únicamente entre tareas aceptadas.

#### Scenario: Corrida barata que rompe aceptación

- **WHEN** una corrida consume menos pero falla un criterio de aceptación
- **THEN** no entra en el costo por tarea aceptada y cuenta como fallo de calidad del brazo

### Requirement: Criterio de adopción documentado

El piloto SHALL cerrar con un informe de decisión que incluya los datos por corrida, la muestra, la incertidumbre y una de tres salidas: adoptar la delegación selectiva como predeterminada, mantener la política vigente con ajustes, o repetir con muestra mayor. La adopción MUST requerir mejora de la métrica primaria acordada por el grupo con calidad equivalente o superior.

#### Scenario: Mejora con calidad inferior

- **WHEN** el brazo selectivo reduce el costo por tarea aceptada pero acepta menos tareas que el brazo directo
- **THEN** el informe no recomienda adoptarlo

#### Scenario: Resultado dentro de la incertidumbre

- **WHEN** la diferencia de la métrica primaria cae dentro del margen de incertidumbre declarado
- **THEN** la salida del informe es repetir o mantener, no adoptar
