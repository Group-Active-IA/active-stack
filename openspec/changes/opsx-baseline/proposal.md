# Proposal · 02 · Medición de referencia de OPSX

**Estado** · Borrador de Matías para acordar con el equipo.

**Revisión de partida de la propuesta** · `1b4639b537a4d0c04dc8c25c0220732095a772d3`.

## Why

Falta una línea base de consumo y calidad por tarea en Claude Code y Codex. El informe ejecutivo midió el tamaño de las plantillas y el reporte del instalador agrega chequeos de salud. Para decidir si la modularización y los pilotos mejoran el flujo, hace falta observar ejecuciones comparables con criterios de aceptación comunes. [Informe ejecutivo](https://active-stack-executive-site.vercel.app/#evidencia) · [Reporte del instalador](../../../internal/verify/report.go)

## What Changes

- Definir un arnés externo al binario, con tareas y fixtures versionados de exploración, cambio pequeño, cambio multiarchivo, diagnóstico y revisión.
- Medir la carga de contexto de una sesión normal y establecer un protocolo común para el OPSX actual y el modular cuando esté disponible. Registrar calidad, entrada, caché, salida, reintentos, intervención humana y consumo atribuible al coordinador y a los workers.
- Entregar un formato de resultados reproducible, con cobertura explícita de las métricas y reportes separados para API y suscripción.
- Usar una configuración compatible con 01 para las corridas de Codex. Los changes 03 y 04 consumirán la referencia para comparar el flujo modular. El 07 y el 09 reutilizarán el protocolo para sus experimentos.

## Capabilities

### New Capabilities

- `opsx-benchmarking` — Ejecutar y evaluar tareas de referencia con criterios verificables, evidencia trazable y resultados comparables.

### Modified Capabilities

Ninguna.

## Impact

El arnés corre sobre perfiles y repositorios de prueba aislados. Quedan fuera del alcance las modificaciones al instalador y a la configuración personal, el enrutamiento de modelos, las optimizaciones de caché y el juez externo. Esta ronda documenta el plan de implementación y de pruebas.

El [registro de decisiones](design.md#decisiones-para-acordar) contiene las opciones y los changes afectados. La propuesta final se comparte una vez que el equipo cierre esas decisiones.
