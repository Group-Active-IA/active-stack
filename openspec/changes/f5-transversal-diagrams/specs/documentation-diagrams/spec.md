# Diagramas de documentación

## Purpose

Conservar visualizaciones editables que expliquen documentación de proyecto y puedan recuperarse desde Codex.

## ADDED Requirements

### Requirement: Incorporación y persistencia de diagramas

Codex SHALL incorporar un bloque Mermaid editable cuando el diagrama aclare relaciones relevantes de la documentación y SHALL guardarlo en el origen autorizado.

#### Scenario: Roadmap con dependencias

- **WHEN** se documenta un roadmap cuyas dependencias se entienden mejor visualmente
- **THEN** el Markdown conserva el bloque Mermaid con las dependencias respaldadas por el roadmap

#### Scenario: Arquitectura en ClickUp

- **WHEN** un flujo de arquitectura se documenta en una tarea y el usuario confirma la escritura remota
- **THEN** la descripción Markdown conserva la fuente Mermaid editable

#### Scenario: Documento breve

- **WHEN** el gráfico no aclara ninguna relación relevante
- **THEN** la documentación puede entregarse solo como texto

### Requirement: Recuperación visual desde la fuente guardada

Cuando el usuario solicite mostrar un diagrama, Codex SHALL recuperar el Mermaid guardado y presentarlo como diagrama.

#### Scenario: Sesión nueva y archivo del repo

- **WHEN** el usuario pide el diagrama de un roadmap guardado en el repositorio
- **THEN** Codex lee el Markdown y muestra la misma fuente Mermaid, conservando nodos y flechas

#### Scenario: Sesión nueva y tarea de ClickUp

- **WHEN** el usuario pide el diagrama de una tarea accesible
- **THEN** Codex lee markdown_description, presenta el Mermaid guardado y referencia la tarea con el contexto disponible

#### Scenario: Origen no accesible

- **WHEN** el origen o el bloque solicitado no está disponible
- **THEN** Codex indica qué falta para recuperar el diagrama
