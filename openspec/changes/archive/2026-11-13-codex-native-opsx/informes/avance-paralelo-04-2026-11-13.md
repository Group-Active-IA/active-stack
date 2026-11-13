# Avance independiente del cambio 04 — 13/11/2026

## Alcance y resultado

Se trabajó exclusivamente sobre 04 en `propose/codex-native-opsx`, partiendo del HEAD `9a6eb2a`. No se incorporó ni implementó código del cambio 03. No se ejecutaron build, commit ni push.

La tarea 2.2 está implementada y probada en perfiles temporales: Machine instala la skill nativa; un adapter de entrega nativa sin la capacidad opcional de rutas de skills devuelve error; Codex Project conserva el recorrido legacy e informa su limitación en stderr con la ruta resuelta por el adapter. El target se obtiene mediante una capacidad opcional explícita, sin inferirlo del nombre del archivo. El aviso también aparece en la reinstalación idempotente. Claude Project y Codex Machine no reciben este diagnóstico.

`Result.Warnings` transporta el aviso hasta el paso de instalación. No se emite `StepStatusDegraded`: ese estado representa fallos best-effort y Project legacy sigue siendo comportamiento soportado. No se amplió el contrato de pipeline ni la UI.

## Evidencia TDD

| Comportamiento | Prueba | Safety net | RED | GREEN / triangulación |
|---|---|---|---|---|
| Diagnóstico y límites de target | `internal/harness/config/project_diagnostic_test.go` | 139 casos/subcasos de config | Falló: diagnóstico Codex Project ausente | Project, Machine, Claude, Machine nativo, capacidad ausente y reinstalación con un único bloque legacy |
| Aviso sin degradar instalación | `internal/install/codex_project_internal_test.go` | 113 casos/subcasos de install | Falló: el aviso generaba un evento degraded | stderr contiene el aviso, la instalación termina sin error y no emite eventos de degradación |

Validación posterior: 146 casos/subcasos de config y 114 de install aprobados en `go test ./internal/harness/config ./internal/install`. `go vet ./internal/harness/config ./internal/install` y `git diff --check` aprobaron. Los tests existentes de entrega Machine y las pruebas nuevas de capacidad ausente y Project cubren los tres criterios técnicos de 2.2. El enum existente conserva su valor cero; esta iteración no lo modificó.

## Coordinación que sigue pendiente

La compatibilidad conjunta 03/04 no se probó. Antes de integrar ambas entregas, revisar con su responsable:

- Mantener un único enum de entrega y acordar identificadores y valores, preservando el valor cero compatible.
- Despachar cada entrega por agente y target sin sustituir la ruta de otro agente.
- Conservar las opciones aportadas por 03 y la información `ManagedWrites` de 04 al combinar contratos de instalación y resultados parciales; las guardas de rollback deben seguir recibiendo esos estados.

Estos puntos son una guía de revisión de los archivos compartidos, no una implementación del cambio 03 ni evidencia de su compatibilidad conjunta.

## Límites de aceptación

Este informe acredita el comportamiento del instalador en perfiles aislados. No acredita aceptación de Codex CLI o extensión IDE, Windows/Linux, roles especializados ni sesión normal limpia en un cliente real. Esas celdas continúan pendientes en 1.2 y 4.2; el acuerdo compartido con 03 sigue en 1.1 y 4.6. Los contratos de captura y recibos corroborados de 05/06, la comparación con 02 y el cierre global de regresiones siguen pendientes en sus tareas específicas.
