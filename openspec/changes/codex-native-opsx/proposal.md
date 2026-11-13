# Proposal · 04 · Roles y activación explícita de OPSX en Codex

**Responsable** · Matías. **Revisión** · 13 de noviembre de 2026.

Esta revisión incorpora el informe unificado del equipo del 10/11, la matriz compartida publicada por 03 en `ed37113` y la base de implementación `origin/main` en `35a89f8`. Los informes de intentos anteriores se conservan como antecedentes. El estado vigente se describe en estos cuatro artefactos.

## Why

El adapter de Codex sigue entregando la coordinación completa en el `AGENTS.md` compartido. Una sesión común recibe reglas de coordinador aunque el usuario no lo haya invocado. El change entrega la coordinación mediante una skill explícita dentro de la sesión actual, con reglas universales y módulos por acción.

## What Changes

- Instalar `$sdd-orchestrator` con invocación implícita deshabilitada, módulos de apoyo y especialistas, conservando el contexto del hilo actual.
- Separar reglas universales de coordinación conforme a la matriz común de 03/04. Integrar `verify` y la revisión independiente cuando sea obligatoria.
- Usar el sobre de retorno de 07 y los recibos de ejecución de 06, también cuando una acción con evidencia se resuelva directamente. Una declaración del worker conserva estado sin verificar hasta su corroboración independiente.
- Entregar la nueva instalación para Machine. Migrar los bloques administrados, incluidos markers heredados, con revisión, snapshot persistente, propiedad de archivos, reinstalación idempotente y restauración por ruta que preserve ediciones posteriores.
- Definir aceptación separada de CLI y extensión IDE en Windows y Linux. Identificar capacidades no comprobadas y limitar cada fallback a sus garantías verificadas.
- **BREAKING**. La sesión normal conserva las reglas universales y el usuario invoca explícitamente la coordinación.

## Capabilities

### New Capabilities

- `codex-opsx-delivery` — activación, reglas por rol y acción, especialistas, evidencia, verificación y límites por superficie.
- `codex-sdd-migration` — migración, propiedad, reinstalación, actualización, rollback, reversión y retiro de archivos administrados.

### Modified Capabilities

Ninguna. Las capacidades archivadas del instalador conservan su comportamiento.

## Impact

El cambio afecta al adapter de Codex, el despacho de configuración, los assets, el plan de escritura, verify, update y uninstall. Las rutas se resuelven mediante adapters. Los consumidores Claude, OpenCode, Antigravity y `design-mockups` requieren pruebas de regresión. El catálogo mantiene el harness `sdd-orchestrator` y sus toggles.

02 está implementado parcialmente en main y lo lleva Giuliano. Su manifiesto y arnés aportan medición por interfaz. 03 aporta la matriz, 05 los módulos de contexto y captura, 06 los recibos, 07 el retorno y la independencia, y 08 los controles efectivos. El formato pendiente de una dependencia se identifica en design con su resultado requerido y comprobación.

## Revisión humana y preparación de apply

Matías confirmó el 13/11/2026 el alcance Machine inicial, la reversión con snapshot persistente protegido por ruta y el avance de tareas independientes mientras 05/06 completan sus contratos. La aceptación integrada de evidencia espera esas dependencias. El dominio es **ALTO** por migración y retiro de configuración. El alcance concreto de producción está en design. Matías lo aprobó y autorizó iniciar las tareas independientes el 13/11/2026. Las sondas técnicas y las dependencias externas conservan su estado pendiente.
