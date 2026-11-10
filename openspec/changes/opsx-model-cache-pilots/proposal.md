# Proposal · 09 · Pilotos de modelos y caché

**Estado** · Borrador de Matías para acordar con el equipo.

**Revisión de partida de la propuesta** · `1b4639b537a4d0c04dc8c25c0220732095a772d3`.

## Why

Las plantillas actuales asignan modelos por fase, pero falta medir si una política por rol mejora el costo o consumo por tarea aceptada. También falta medir cuánto aporta estabilizar los prefijos y en qué condiciones hay lecturas de caché. Los efectos se estimarán por separado para distinguir sus causas. [Roadmap, ficha 09](../../../CHANGES.md#change-09) · [Informe ejecutivo](https://active-stack-executive-site.vercel.app/#plan)

## What Changes

- Diseñar dos pilotos independientes para Claude Code y Codex. El primero compara configuraciones de modelo y esfuerzo por rol. El segundo compara estrategias de prefijo con condiciones frías y calientes, manteniendo fijos modelos, esfuerzo y roles.
- Reutilizar el protocolo y la referencia de 02, así como el flujo y el contrato de delegación de 07. Comparar variantes con la misma tarea y rúbrica, incluyendo coordinador, workers y reintentos.
- Entregar una recomendación por piloto y cohorte, con evidencia, calidad, cobertura, consumo por tarea aceptada, tiempo y cuota por separado. La adopción exige una mejora de la métrica acordada y calidad comparable o superior.

## Capabilities

### New Capabilities

- `opsx-model-cache-evaluation` — Evaluación controlada y reproducible de políticas de modelo y esfuerzo por rol y de estabilidad de prefijos en OPSX.

### Modified Capabilities

Ninguna.

## Impact

El arnés externo de 02 será la base de ejecución y reporte. El diseño del 09 puede redactarse en paralelo. Sus corridas requieren el protocolo de 02 y la disponibilidad del flujo de 07. Los pilotos se preparan en perfiles aislados. La configuración predeterminada del instalador, la incorporación permanente de un router y el juez en sombra del 10 quedan fuera de alcance. El [registro de decisiones](design.md#decisiones-para-acordar) reúne los acuerdos previos a la revisión formal.
