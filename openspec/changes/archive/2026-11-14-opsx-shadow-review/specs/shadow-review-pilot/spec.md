# shadow-review-pilot

## Purpose

Define el piloto que mide si un juez externo en sombra detecta defectos adicionales frente a la revisión de referencia, con qué tasa de errores y a qué costo, sin influir en la aceptación de ningún change.

## ADDED Requirements

### Requirement: Primera ronda sin changes vivos

La primera ronda SHALL usar sólo casos históricos y casos preparados, con material público revisado y datos ficticios. La primera ronda MUST NOT incluir changes vivos ni evidencia real del usuario.

#### Scenario: Change vivo propuesto para la primera ronda

- **WHEN** se intenta agregar a la muestra un change en curso
- **THEN** el caso se rechaza y queda anotado como candidato para una ronda posterior

### Requirement: Resultados en sombra con acceso restringido

Los hallazgos del juez SHALL guardarse fuera del árbol del change, con huella, en un almacén de acceso restringido a un custodio designado que no participa de la referencia ni del etiquetado. El custodio MUST abrir los resultados de un caso sólo cuando el caso esté cerrado definitivamente, con decisión de referencia y etiquetas congeladas registradas.

#### Scenario: Acceso anticipado

- **WHEN** alguien del flujo de referencia o del etiquetado intenta leer los hallazgos de un caso todavía abierto
- **THEN** el acceso se deniega y el intento queda registrado

#### Scenario: Apertura al cierre del caso

- **WHEN** el caso tiene decisión de referencia y etiquetas congeladas registradas
- **THEN** el custodio abre los hallazgos y la huella coincide con la registrada al guardarlos

### Requirement: Resultados fuera de la aceptación

Durante el piloto, ninguna decisión de aceptación o rechazo de un change SHALL usar los hallazgos del juez como criterio. La aceptación MUST seguir su flujo habitual.

#### Scenario: Juez marca un bloqueante en un caso histórico

- **WHEN** el juez reporta un bloqueante confirmado en código ya integrado
- **THEN** el hallazgo se registra para las métricas y, si corresponde, se abre un issue aparte sin modificar decisiones del flujo

### Requirement: Muestra etiquetada y congelada

La muestra SHALL combinar casos históricos con defecto conocido, casos preparados que incluyan los casos de rechazo de 06, y controles inspeccionados o fixtures limpios. Un change sin fixes posteriores conocidos MUST NOT usarse como control sin inspección previa. Cada caso MUST etiquetarse por dos personas de forma independiente, con desacuerdos resueltos por una tercera, y la muestra MUST congelarse en un commit antes de la primera corrida.

#### Scenario: Control limpio

- **WHEN** un control inspeccionado se envía al juez
- **THEN** cualquier hallazgo del juez sobre ese control pasa al arbitraje aparte y, si se descarta como defecto, cuenta como falsa alarma

#### Scenario: Control sin inspección

- **WHEN** se propone como control un change archivado sólo porque no tiene fixes posteriores
- **THEN** no entra a la muestra hasta ser inspeccionado

### Requirement: Paquetes históricos sin pistas

El paquete de un caso histórico SHALL armarse desde el estado previo al fix y MUST excluir el fix posterior, sus mensajes de commit, referencias a issues o PRs del fix, texto escrito después, nombres de rama y etiquetas.

#### Scenario: Paquete con mensaje del fix

- **WHEN** el paquete armado contiene el mensaje de commit del fix que corrige el defecto
- **THEN** el paquete se rechaza y se rearma

### Requirement: Medición contra etiquetas congeladas

El análisis principal SHALL medir la revisión de referencia y cada juez contra las etiquetas congeladas, y MUST reportar por candidato verdaderos positivos, falsos positivos y falsos negativos por severidad, la señal adicional y la proporción de controles con al menos una falsa alarma.

#### Scenario: Defecto que sólo el juez encuentra

- **WHEN** el juez reporta un defecto etiquetado que la referencia no detectó
- **THEN** cuenta como verdadero positivo del juez y suma a la señal adicional

### Requirement: Arbitraje aparte

Los hallazgos que no emparejan con etiquetas congeladas SHALL resolverse en un arbitraje separado, ciego respecto del origen del hallazgo, que asigna defecto nuevo confirmado, falsa alarma o indeterminado. El arbitraje MUST reportarse aparte y MUST NOT modificar las métricas del análisis principal.

#### Scenario: Defecto válido ausente de las etiquetas

- **WHEN** el juez reporta un defecto real que no estaba en las etiquetas
- **THEN** el árbitro, sin saber de qué candidato vino, lo clasifica como defecto nuevo confirmado
- **AND** se reporta aparte sin cambiar la exhaustividad ni la señal adicional principales

### Requirement: Unidades y costo definidos

La señal adicional SHALL medirse como cantidad de defectos distintos de severidad bloqueante o mayor detectados por el juez y no por la referencia; un porcentaje MAY informarse como contexto sin decidir. Cada caso SHALL registrar tokens, costo API o cuota por separado, latencia y minutos de triage, según el protocolo de 02.

#### Scenario: Triage de falsas alarmas

- **WHEN** una persona revisa los hallazgos del juez de un caso
- **THEN** el tiempo de triage se suma al costo de ese caso

### Requirement: Candidatos con criterios de entrada y calibración

Cada candidato SHALL cumplir criterios de entrada antes de correr: aceptar el paquete minimizado sin datos adicionales, devolver hallazgos ubicables, exponer costo observable y declarar retención y uso para entrenamiento. Cada candidato SHALL calibrarse con un conjunto propio fuera de la muestra. Un candidato opcional que no cumpla MUST descartarse con el motivo registrado.

#### Scenario: Candidato sin política de datos declarada

- **WHEN** un candidato no documenta retención o uso para entrenamiento
- **THEN** no se le envía ningún caso y el informe registra el descarte

### Requirement: Cierre con decisión documentada

El piloto SHALL terminar con un informe que incluya datos crudos, muestra, acuerdo de etiquetado, arbitraje aparte, log de envío y una de tres salidas: seguir explorando, repetir o descartar, según umbrales registrados antes de correr. "Seguir explorando" MUST NOT implicar integración del juez.

#### Scenario: Sin señal adicional

- **WHEN** el juez no alcanza la señal adicional mínima registrada
- **THEN** la salida es descartar

#### Scenario: Acuerdo de etiquetado bajo

- **WHEN** el acuerdo entre etiquetadores queda bajo el mínimo registrado
- **THEN** la salida es repetir, sin importar las métricas del juez

#### Scenario: Señal suficiente pero costo excedido

- **WHEN** el juez alcanza la señal mínima y supera el presupuesto por caso
- **THEN** la salida es repetir con otro candidato o configuración, no seguir explorando

### Requirement: Fuera del binario

El arnés del piloto SHALL vivir fuera del binario `active-stack` y MUST NOT modificar `cmd/`, `internal/` ni la configuración que instala.

#### Scenario: Revisión del diff del piloto

- **WHEN** se implementa el arnés
- **THEN** el diff no toca `cmd/` ni `internal/`
