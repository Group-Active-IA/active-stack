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

### Requirement: Select one saved documentation diagram
When an ambiguous request matches multiple saved Mermaid blocks, the agent SHALL offer one labeled option per matching diagram and wait for a selection. The agent SHALL use the native single-choice interaction when exposed by the current client. In a client without that interaction, it SHALL use a numbered list. After selection it SHALL return only the exact selected block.

#### Scenario: Ambiguous request with three diagrams
- **WHEN** the user asks to see the diagram in a source containing three Mermaid blocks
- **THEN** the agent presents the three diagram options and waits before returning a block

#### Scenario: Named diagram in a client without rendering
- **WHEN** the user identifies one saved diagram and the client does not render Mermaid
- **THEN** the agent returns the exact editable fenced block for that diagram

### Requirement: Preserve the diagram across discovery and delivery
The agent SHALL propose a diagram during discovery when it clarifies relationships supported by the source. When a diagram contributes to the delivery, the agent SHALL identify its saved source in the handoff.

#### Scenario: Discovery source describes a workflow
- **WHEN** a discovery document describes a workflow that would benefit from a diagram
- **THEN** the agent proposes an editable Mermaid block expressing the documented relationships

#### Scenario: Recover a delivery diagram
- **WHEN** a later session opens the handoff and selects one of its saved diagrams
- **THEN** the agent retrieves the selected block from the source identified in the handoff
