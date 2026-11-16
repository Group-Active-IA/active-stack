# Proposal

## Why

La activación explícita deja abierta la recomendación de OPSX para tareas simples. La verificación puede bloquear changes de bajo riesgo y la evidencia integrada sigue parcial.

## What Changes

- Resolver tareas cotidianas directamente. Recomendar una propuesta completa cuando sus artefactos permitan acordar un cambio complejo.
- Publicar referencias breves a comandos y skills instalados por cliente.
- Ajustar verify al riesgo e integrar la evidencia de ejecución.
- Corregir la validación estricta, reproducir fallas Windows y probar migración y reanudación.

## Capabilities

### New Capabilities

- `opsx-workflow-selection` — recomendación y descubrimiento de OPSX.

### Modified Capabilities

- `selective-delegation` — revisión según el riesgo.
- `task-context-loading` — módulos requeridos según el tipo de revisión.

## Impact

Composición de instrucciones Claude y Codex, captura y recibos, validación de specs y pruebas del instalador. Base inicial 8bfca74. El PR 16 sigue abierto. Los pilotos 02, 08, 09 y 10 mantienen sus criterios de cierre.
