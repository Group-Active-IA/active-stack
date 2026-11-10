# review-data-exposure

## Purpose

Define qué datos puede recibir un juez externo durante el piloto, cómo se minimizan y redactan, y qué constancia queda de lo que salió del entorno.

## ADDED Requirements

### Requirement: Paquete por lista de inclusión

El paquete que recibe el juez SHALL armarse con una lista de inclusión de clases de datos permitidas: artefactos y diff del repositorio público y, cuando el caso lo requiera, recibos y salidas acotadas redactadas según 06. Todo dato fuera de la lista MUST quedar excluido.

#### Scenario: Caso con salida original

- **WHEN** un caso tiene una salida original sin redactar
- **THEN** el paquete incluye sólo su versión redactada con ambas huellas, o nada

### Requirement: Datos privados nunca salen

El paquete MUST NOT contener secretos, archivos `.env`, valores de variables de entorno, rutas del home, configuración del usuario (`settings.json`, `config.toml`), memoria de Engram ni transcripciones de sesión.

#### Scenario: Ruta del home en un recibo

- **WHEN** un recibo contiene una ruta bajo el home del usuario
- **THEN** la ruta se redacta antes de incluirlo y se registra la huella redactada

#### Scenario: Secreto detectado al armar el paquete

- **WHEN** el chequeo previo al envío detecta un patrón de secreto en el paquete
- **THEN** el envío se cancela y el caso queda marcado para revisión manual

### Requirement: Registro de envío

Cada envío SHALL registrar caso, huella del paquete, candidato, proveedor y fecha. El registro MUST formar parte del informe de cierre.

#### Scenario: Auditoría del piloto

- **WHEN** el grupo revisa qué datos salieron
- **THEN** el registro permite reconstruir cada paquete enviado por su huella

### Requirement: Condiciones del proveedor acordadas antes de enviar

Ningún caso SHALL enviarse a un proveedor antes de que el grupo acuerde y registre las condiciones aceptables de retención y uso de datos y el proveedor las declare.

#### Scenario: Proveedor sin condiciones verificadas

- **WHEN** las condiciones del proveedor no están registradas como acordadas
- **THEN** el arnés no envía casos a ese proveedor
