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

### Requirement: Recorded Strict installations migrate safely

Active Stack MUST migrar la configuración legada al reinstalar o reaplicar una selección registrada de Codex en Estricto, con respaldo previo y preservación de claves y perfiles ajenos.

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
