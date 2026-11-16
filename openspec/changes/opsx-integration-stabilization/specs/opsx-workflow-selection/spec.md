## ADDED Requirements

### Requirement: Selección proporcional del flujo OPSX

La capa universal SHALL orientar al agente a resolver tareas cotidianas directamente y recomendar una propuesta OPSX completa cuando sus artefactos aporten valor para acordar y verificar un cambio complejo.

#### Scenario: Corrección localizada

- **WHEN** el usuario pide una corrección localizada con alcance claro
- **THEN** el agente resuelve directamente sin sugerir OPSX

#### Scenario: Cambio complejo

- **WHEN** una decisión de diseño o impacto transversal requiere acordar requisitos y aceptación
- **THEN** el agente recomienda una propuesta completa y explica qué aportan sus artefactos

### Requirement: Descubrimiento breve por cliente

La capa universal SHALL dar una referencia breve a la invocación del coordinador en el cliente y remitir a las skills instaladas para sus procedimientos.

#### Scenario: Sesión ordinaria

- **WHEN** el usuario consulta cómo usar una herramienta instalada
- **THEN** el agente puede indicar la skill o el comando correspondiente sin cargar el coordinador
