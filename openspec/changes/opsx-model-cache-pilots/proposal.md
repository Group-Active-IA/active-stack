# Proposal · 09 · Pilotos de modelos y caché

**Estado** · Apply incremental offline; acuerdos D1–D5 y experimentos reales pendientes.

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

El arnés externo de 02 será la base de ejecución y reporte. El diseño del 09 puede redactarse en paralelo. Sus corridas confirmatorias requieren el protocolo aprobado de 02 y la disponibilidad del flujo y resultados de 07; la preparación/calibración offline por brazo puede avanzar sin autorizar adopción. Los pilotos se preparan en perfiles aislados. La configuración predeterminada del instalador, la incorporación permanente de un router y el juez en sombra del 10 quedan fuera de alcance. El [registro de decisiones](design.md#decisiones-para-acordar) reúne los acuerdos previos a la revisión formal.


## Entrega incremental offline · 2026-11-13

Se incorpora main `d5e95bd81d3bbab2d54a39d1c07d1256954b5384` y se implementa `eval/modelcache`, consumidor de APIs públicas del 02, sin reescribir otros cambios. El resultado permite preparar/calibrar por brazo aunque 02 no tenga referencia estable y 07 no entregue A0/resultados. No permite experimentos confirmatorios ni adopción: esas acciones requieren acuerdos D1–D5 y evaluación independiente todavía no integrada.

El CLI lee JSON y archivos designados localmente; no llama proveedores, instala perfiles ni configura routing. Todos los ejemplos son sintéticos y no fijan modelos, esfuerzos, tamaño muestral, presupuesto ni umbrales reales. `null` mantiene un parámetro pendiente; cero observado tiene valor propio.

El layout conserva rutas, hashes de bytes y orden explícito. `repo-source` no prueba el payload de un cliente ni que una skill se haya leído. `materialized-prompt` es una declaración del dataset, no una certificación independiente. El piloto B implementado sólo reordena el mismo conjunto de archivos/bytes y exige una referencia declarada de revisión semántica; no valida reescrituras generales.

El reporte separa piloto, brazo, cohorte, condición asignada de caché, modalidad, basis y unidades. Conserva fallos/reintentos y preparación identificada una sola vez. Los agregados incluyen workers sólo según alcance explícito; los fragmentos incompletos no producen costos normalizados completos. Una preparación ambigua entre frío/caliente queda visible una vez y no habilita ratios condicionales. Una campaña ocupa un Dataset; repeticiones duplicadas dentro de brazo/cohorte/tarea/condición se rechazan, pero reintentos son eventos del intento.

La señal de caché debe atribuirse al prefijo designado; un contador global, una sesión nueva o latencia no lo demuestran. Se conservan misses/lecturas parciales genuinos y contaminación. Los contrastes de caché comparan la misma condición asignada; contaminación fría, aislamiento faltante o atribución desconocida vuelven la conclusión inconclusa sin eliminar su costo. La ventana caliente y la carga real del cliente requieren calibración externa.

El cálculo condicional exige controles/pares íntegros, objetivo explícito, cobertura, criterios de calidad y umbrales declarados. Una regresión obligatoria lo rechaza. `candidate_for_review` no es inferencia verificada ni autorización de adopción; `adoption.status` permanece `unavailable`, sin confiar en booleanos de procedencia introducidos por el propio dataset.

El reporte conserva pins02/07, fixture/rúbrica, layouts y fuentes declaradas. Su SHA256 identifica la serialización JSON tipada modelcache/v1; no es un TreeHash06, recibo, comprobación de ejecución ni prueba de suficiencia. Los contratos/implementaciones de 05/06/07 fuera de main no se incorporan automáticamente.

Design mockups: Decision: no (no UI).
