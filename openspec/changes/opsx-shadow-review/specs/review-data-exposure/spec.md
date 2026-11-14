# review-data-exposure

## Purpose

Define qué datos puede recibir un juez externo durante el piloto, cómo se examinan y minimizan, y qué constancia queda de lo que salió del entorno.

## ADDED Requirements

### Requirement: Paquete por lista de inclusión

El paquete que recibe el juez SHALL armarse con una lista de inclusión. En la primera ronda MUST contener sólo material público revisado y datos ficticios. Todo dato fuera de la lista MUST quedar excluido.

#### Scenario: Recibo real en la primera ronda

- **WHEN** un caso de la primera ronda incluiría un recibo o salida real, aunque esté redactado
- **THEN** se reemplaza por un equivalente ficticio o el caso se excluye

### Requirement: Examen previo de todo el paquete

Antes de cada envío, el paquete completo SHALL examinarse con la lista única de patrones de valor sensibles definida por 05 y en busca de archivos de configuración de usuario, incluidos los diffs del repositorio público. El piloto MUST NOT usar una lista propia de patrones. Un positivo MUST cancelar el envío del caso.

#### Scenario: Secreto sembrado en un diff público

- **WHEN** un diff público del paquete contiene un secreto sembrado
- **THEN** el examen lo detecta, el envío se cancela y el caso queda marcado para revisión manual

#### Scenario: Patrón agregado a la lista de 05

- **WHEN** 05 agrega un patrón a su lista de valores sensibles
- **THEN** el examen previo del piloto lo aplica sin cambios en esta ficha

### Requirement: Datos privados nunca salen

El paquete MUST NOT contener salidas originales, archivos `.env`, configuración del usuario (`settings.json`, `config.toml`), memoria de Engram ni transcripciones de sesión.

#### Scenario: Configuración de usuario en un fixture

- **WHEN** un fixture incluye un `settings.json` real de un integrante
- **THEN** el paquete se rechaza y el fixture se rehace con datos ficticios

### Requirement: Registro de envío

Cada envío SHALL registrar caso, huella del paquete, candidato, proveedor y fecha. El registro MUST formar parte del informe de cierre.

#### Scenario: Auditoría del piloto

- **WHEN** el grupo revisa qué datos salieron
- **THEN** el registro permite reconstruir cada paquete enviado por su huella

### Requirement: Condiciones del proveedor verificadas por separado

Ningún caso SHALL enviarse a un proveedor antes de que el grupo registre como acordadas, por separado, la condición de retención y la condición de uso para entrenamiento, y el proveedor declare ambas.

#### Scenario: Proveedor que declara retención pero no entrenamiento

- **WHEN** el proveedor documenta su retención y no documenta si usa los datos para entrenamiento
- **THEN** el arnés no envía casos a ese proveedor
