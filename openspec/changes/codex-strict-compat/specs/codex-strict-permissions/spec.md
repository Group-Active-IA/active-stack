# Spec Delta

## Purpose

Define la configuración efectiva del tier Estricto de Codex y cómo migrarla preservando las preferencias del usuario.

## ADDED Requirements

### Requirement: Codex Strict loads with a read-only sandbox

Active Stack MUST configurar Codex en solo lectura con solicitudes de aprobación para las escalaciones admitidas cuando se elige Estricto. La descripción para Codex MUST reflejar ese comportamiento.

#### Scenario: Clean install with Strict

- **WHEN** se instala Codex con Estricto en un entorno soportado
- **THEN** el cliente carga la configuración, puede leer un archivo permitido sin aprobación y muestra una descripción de Estricto que explica la restricción de escritura en el sandbox

#### Scenario: Escalation is denied

- **WHEN** una escritura de prueba requiere escalación y el revisor la rechaza
- **THEN** el archivo de prueba queda intacto

#### Scenario: Write without escalation

- **WHEN** se intenta una escritura dentro del sandbox sin escalación
- **THEN** se bloquea la escritura y el archivo testigo queda intacto

#### Scenario: Escalation is approved

- **WHEN** el revisor habilitado aprueba la escalación de la escritura de prueba
- **THEN** se permite esa operación y se verifica el resultado sobre el archivo testigo

#### Scenario: Client cannot request escalation

- **WHEN** la superficie carece de un mecanismo para solicitar escalación
- **THEN** se informa la limitación y el archivo testigo queda intacto

### Requirement: Recorded Strict installations migrate safely

Active Stack MUST migrar la configuración legada al reinstalar o reaplicar una selección registrada de Codex en Estricto, con respaldo previo y preservación de claves y perfiles ajenos.

#### Scenario: Legacy policy in an unmanaged layer

- **WHEN** verify detecta la política legada en una capa ajena al instalador
- **THEN** informa la capa y el impacto observado, preserva esa capa e indica si la política efectiva sigue sin verificar

#### Scenario: Invalid TOML

- **WHEN** el TOML que se va a migrar es inválido
- **THEN** informa el diagnóstico y conserva el archivo intacto

#### Scenario: Rollback would overwrite later edits

- **WHEN** el archivo difiere del estado que dejó la operación que se intenta revertir
- **THEN** conserva el archivo actual e identifica el respaldo para recuperación manual

#### Scenario: Recorded update and repeat

- **WHEN** una actualización reaplica Estricto sobre una configuración legada con preferencias adicionales y se repite la operación
- **THEN** la primera ejecución deja una configuración compatible con las preferencias intactas y la segunda conserva el mismo contenido

#### Scenario: Update without an installation record

- **WHEN** se ejecuta una actualización completa sin selección registrada
- **THEN** los permisos quedan intactos y se indica cómo reaplicar el tier mediante instalación explícita

#### Scenario: Backup fails

- **WHEN** falla el respaldo de una configuración existente
- **THEN** la operación informa el fallo y deja el archivo intacto

#### Scenario: Migration fails after writing

- **WHEN** la migración falla tras modificar la configuración y corresponde revertirla
- **THEN** el rollback restaura el contenido anterior o informa el fallo de restauración e identifica el respaldo para recuperación
