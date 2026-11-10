# shadow-review-pilot

## Purpose

Define el piloto que mide si un juez externo en sombra detecta defectos adicionales frente a la revisión de referencia, con qué tasa de errores y a qué costo, sin influir en la aceptación de ningún change.

## ADDED Requirements

### Requirement: Resultados en sombra

Los hallazgos del juez SHALL guardarse en un almacén sellado con huella y MUST NOT estar disponibles para el coordinador, el auditor ni la persona que acepta hasta que la decisión de aceptación del caso esté registrada.

#### Scenario: Change vivo en revisión

- **WHEN** el juez termina su revisión de un change que todavía no fue aceptado ni rechazado
- **THEN** sus hallazgos quedan sellados y ningún artefacto del change los referencia

#### Scenario: Apertura después de la decisión

- **WHEN** la decisión de aceptación del caso está registrada
- **THEN** los hallazgos pueden abrirse para el emparejamiento y la huella del almacén coincide con la registrada al sellar

### Requirement: Resultados fuera de la aceptación

Durante el piloto, ninguna decisión de aceptación o rechazo de un change SHALL usar los hallazgos del juez como criterio.

#### Scenario: Juez marca un bloqueante en un change aceptado

- **WHEN** al abrir el almacén el juez reporta un bloqueante en un change ya aceptado
- **THEN** el hallazgo se registra para las métricas y, si se confirma, se abre un issue aparte sin revertir la aceptación del piloto

### Requirement: Muestra etiquetada y congelada

La muestra SHALL combinar casos históricos con defecto conocido, casos con defectos sembrados que incluyan los casos de rechazo de 06, y casos de control sin defecto conocido. Cada caso MUST etiquetarse por dos personas de forma independiente, con desacuerdos resueltos por una tercera, y la muestra MUST congelarse en un commit antes de la primera corrida.

#### Scenario: Desacuerdo entre etiquetadores

- **WHEN** dos etiquetadores asignan severidades distintas al mismo defecto
- **THEN** un tercero resuelve y se registra el acuerdo entre etiquetadores de la muestra

#### Scenario: Muestra modificada después de correr

- **WHEN** se agrega o cambia un caso después de la primera corrida
- **THEN** las corridas previas no se comparan con las nuevas sin declarar el cambio de muestra

### Requirement: Referencia y juez medidos contra las etiquetas

El piloto SHALL medir tanto la revisión de referencia como cada juez contra las etiquetas, y MUST reportar por candidato verdaderos positivos, falsos positivos y falsos negativos por severidad, la señal adicional y la tasa de falsos positivos en casos de control.

#### Scenario: Defecto que sólo el juez encuentra

- **WHEN** el juez reporta un defecto etiquetado que la referencia no detectó
- **THEN** cuenta como verdadero positivo del juez y como señal adicional, no como falso positivo

#### Scenario: Hallazgo sin etiqueta

- **WHEN** el juez reporta un hallazgo que no empareja con ninguna etiqueta
- **THEN** el tercer etiquetador lo revisa y, si no es un defecto, cuenta como falso positivo

### Requirement: Costo contabilizado

Cada corrida del juez SHALL registrar tokens de entrada con y sin caché, salida, costo API o consumo de cuota por separado, latencia y tiempo humano de triage, con el protocolo de 02 y 09. Una métrica no disponible MUST registrarse como ausente.

#### Scenario: Triage de falsos positivos

- **WHEN** una persona revisa los hallazgos del juez de un caso
- **THEN** el tiempo de triage se suma al costo del juez para ese caso

### Requirement: Candidatos con criterios de entrada

Cada candidato SHALL cumplir criterios de entrada antes de correr: aceptar el paquete minimizado sin datos adicionales, devolver hallazgos ubicables por ruta y rango o requisito, exponer costo observable y declarar su política de datos. Un candidato opcional que no los cumpla MUST descartarse con el motivo registrado.

#### Scenario: Candidato sin política de datos declarada

- **WHEN** un candidato no documenta retención ni uso de los datos recibidos
- **THEN** no se le envía ningún caso y el informe registra el descarte

### Requirement: Cierre con decisión documentada

El piloto SHALL terminar con un informe que incluya datos crudos, muestra, acuerdo de etiquetado, log de envío de datos y una de tres salidas: seguir explorando, repetir o descartar, según umbrales acordados antes de correr. "Seguir explorando" MUST NOT implicar integración del juez, que requiere una decisión de arquitectura separada.

#### Scenario: Sin señal adicional

- **WHEN** el juez no encuentra defectos etiquetados que la referencia haya omitido
- **THEN** la salida es descartar

#### Scenario: Acuerdo de etiquetado bajo

- **WHEN** el acuerdo entre etiquetadores queda bajo el mínimo acordado
- **THEN** la salida es repetir, sin importar las métricas del juez

### Requirement: Fuera del binario

El arnés del piloto SHALL vivir fuera del binario `active-stack` y MUST NOT modificar `cmd/`, `internal/` ni la configuración que instala.

#### Scenario: Revisión del diff del piloto

- **WHEN** se implementa el arnés
- **THEN** el diff no toca `cmd/` ni `internal/`
