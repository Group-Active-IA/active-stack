# Proposal · 02 · Medición de referencia de OPSX

**Estado** · Propuesta de Giuliano, revisada el 2026-11-11 con los acuerdos del grupo. Quedan pendientes de acuerdo D1, los valores de D3, D5 y la ubicación del arnés en D6 ([registro de decisiones](design.md#registro-de-decisiones)).

**Revisión de partida de la propuesta** · `1b4639b537a4d0c04dc8c25c0220732095a772d3`.

## Why

Falta una línea base de consumo y calidad por tarea en Claude Code y Codex. El informe ejecutivo midió el tamaño de las plantillas y el reporte del instalador agrega chequeos de salud. Para decidir si la modularización y los pilotos mejoran el flujo, hace falta observar ejecuciones comparables con criterios de aceptación comunes. [Informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia) · [Reporte del instalador](../../../internal/verify/report.go)

## What Changes

- Definir un arnés externo al binario, con tareas y fixtures versionados de exploración, cambio pequeño, cambio multiarchivo, diagnóstico y revisión.
- Medir terminal e IDE desde la primera etapa, con resultados separados por interfaz. Una terminal integrada que ejecuta la CLI pertenece a la cohorte de terminal. Las aplicaciones de escritorio quedan fuera de la primera comparación.
- Medir la carga de contexto de una sesión normal y establecer un protocolo común para el OPSX actual y el modular cuando esté disponible. Registrar calidad, entrada, caché, salida, lecturas y llamadas a herramientas, reintentos, tiempo, intervención humana y consumo atribuible al coordinador y a los workers.
- Mantener el manifiesto de corridas como única fuente de versión, interfaz, permisos efectivos y modalidad para 07, 09 y 10. Mantener la rúbrica común, incluida la regla de emparejamiento de la revisión de un diff sembrado que reutiliza 10.
- Medir el perfil Estricto de 01 en una serie separada. Las tareas con edición corren en perfiles aislados con permisos equivalentes y registran los permisos efectivos por cliente e interfaz.
- Entregar un formato de resultados reproducible, con cobertura explícita de las métricas, denominadores definidos y reportes separados para API y suscripción. Los changes 03 y 04 consumirán la referencia para comparar el flujo modular. El 07, el 09 y el 10 reutilizarán el protocolo y declararán sus diferencias.

## Capabilities

### New Capabilities

- `opsx-benchmarking` — Ejecutar y evaluar tareas de referencia con criterios verificables, evidencia trazable y resultados comparables.

### Modified Capabilities

Ninguna.

## Impact

El arnés corre sobre perfiles y repositorios de prueba aislados, fuera del binario. Quedan fuera del alcance las modificaciones al instalador y a la configuración personal, el enrutamiento de modelos, las optimizaciones de caché y el juez externo. La captura, el acceso y la política de datos sensibles los define 05; las huellas y los estados de procedencia, 06. Esta ronda documenta el plan de implementación y de pruebas.

El [registro de decisiones](design.md#registro-de-decisiones) contiene el estado de cada decisión, su fuente y los changes afectados.
