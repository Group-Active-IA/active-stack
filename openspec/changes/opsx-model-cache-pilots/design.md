# Design · Pilotos de modelos y caché

## Context

Las [plantillas de Claude](../../../internal/harness/config/assets/claude/sdd-orchestrator.md) y [Codex](../../../internal/harness/config/assets/codex/sdd-orchestrator.md) contienen asignaciones por fase. Claude además describe el paso de un alias al delegar y una sustitución si falta acceso. La tabla expresa intención de routing. La selección efectiva de cada cliente y superficie se verifica en el arnés.

El protocolo de 02 aporta fixtures, rúbricas, estructura de resultados y reglas de atribución. El flujo de 07 aporta la política de delegación y el contrato de retorno. Fijar las revisiones aprobadas de ambos antes de correr el 09. El control se vuelve a medir sobre 07 para evitar atribuir al piloto cambios de flujo ocurridos desde la referencia inicial de 02.

## Goals / Non-Goals

El objetivo es poder recomendar por separado una política de modelo o esfuerzo y una estrategia de prefijo, con alcance por cliente, superficie y modalidad. El 09 entrega evidencia y una decisión de piloto. Un cambio predeterminado del instalador o un despliegue de routing se evalúa aparte.

## Decisions

### Controles comunes

- Fijar cliente, versión, superficie, modalidad, modelo y cuenta cuando corresponda, revisión del fixture, rúbrica, herramientas, permisos, límites y política de delegación. Seleccionar tareas directas y delegadas de 02 que representen el flujo de 07. Mantener el encargo y contrato del worker equivalentes. Registrar roles, cantidad de workers y reintentos observados para explicar diferencias de ejecución.
- Calibrar con tareas fuera de la muestra medida. Emparejar por tarea, fixture y bloque temporal, y alternar con una semilla registrada el orden de los bloques de control y variante. Mantener cada secuencia de preparación y lectura de caché dentro de su bloque. Reiniciar sesión y checkout por repetición.
- Registrar intención y ejecución efectiva, cobertura, motivo de desvío, resultado por criterio y vínculo a evidencia depurada. Conservar los pares excluidos y reportar cuántos quedan comparables por brazo. Aplicar la misma rúbrica y ocultar la variante al revisor cuando sea viable.

### Piloto A · Modelo y esfuerzo por rol

| Brazo | Cambio evaluado |
|---|---|
| A0 · Control | Configuración efectiva de 07 por cliente y modalidad. |
| A1 · Modelo | Cambiar el modelo de un rol elegido, manteniendo el esfuerzo donde tenga un significado comparable y pueda fijarse. |
| A2 · Esfuerzo | Cambiar el esfuerzo de un rol elegido con el mismo modelo. |

Mantener fija la estrategia de prefijo y la preparación de caché. Usar como comparación inicial una condición de caché controlada y comprobable donde sea posible. Registrar lecturas y escrituras por brazo, ya que cambiar de modelo puede cambiar la elegibilidad o el costo de caché. Separar desvíos de caché del contraste confirmatorio si impiden atribuir el efecto. Si una superficie obliga a variar modelo y esfuerzo juntos, etiquetar el brazo como combinado y reportarlo por separado.

En Claude Code, las definiciones de subagente documentan `model` y `effort`, con precedencia de overrides según la versión. En Codex, los archivos de agente personalizado documentan `model` y `model_reasoning_effort`. Son candidatos para perfiles aislados. Verificar versiones, permisos, valores efectivos, herencia y fallback en el coordinador y en los workers. Si una superficie carece del control o de evidencia suficiente, marcar la variante como no evaluable para esa superficie. [Subagentes de Claude Code](https://code.claude.com/docs/en/sub-agents) · [Subagentes de Codex](https://learn.chatgpt.com/docs/agent-configuration/subagents)

Si el equipo decide incluir Jev, plantear un brazo opcional frente al routing nativo con el mismo contrato y rúbrica, y contabilizar su tiempo y costo. Confirmar primero compatibilidad, datos que recibiría y evidencia efectiva de routing. Su evaluación de routing queda separada del juez en sombra de 10.

### Piloto B · Estabilidad de prefijos y caché

| Brazo | Condición prevista |
|---|---|
| B0 · Prefijo actual | Fría y caliente bajo el procedimiento acordado. |
| B1 · Prefijo estable | Fría y caliente con las mismas reglas, herramientas y entradas de tarea, colocando el contenido común estable antes de los datos variables donde el cliente lo permita. |

Fijar modelo, esfuerzo, roles, versión, política de delegación y parámetros de caché disponibles entre B0 y B1. Mantener equivalencia semántica y documentar cualquier diferencia de longitud. Si el cliente administra contenido que el arnés no puede inspeccionar, declarar el alcance controlado y evitar conclusiones sobre ese contenido.

- **Fría**. Partir de un bloque aislado por el mecanismo disponible o esperar la expiración verificable. Confirmar el estado en la primera solicitud elegible del prefijo. El resto de la tarea puede reutilizar caché y se registra por separado.
- **Caliente**. Preparar el prefijo del brazo con una solicitud separada y medir en una sesión nueva dentro de la ventana elegible. Registrar tiempo, parámetros y costo de preparación. Confirmar la lectura del prefijo con la señal disponible y mantener bloques independientes para evitar contaminación entre brazos. Una lectura parcial o un miss se conserva como resultado observado del brazo asignado. Una sesión nueva por sí sola no prueba aislamiento.
- **Análisis**. Comparar B0 con B1 dentro de cada condición asignada y usar las lecturas observadas para explicar el efecto. Mantener los misses genuinos en el rendimiento de la estrategia. Identificar desvíos de protocolo y no inferir un hit del prefijo a partir de latencia o de uso agregado sin atribución. Informar entrada lógica, lectura, escritura, salida y costo de preparación por separado. Incluir la preparación imputable al horizonte de reutilización acordado y mostrar el costo completo del bloque. La reutilización de cómputo y el tamaño de contexto se miden por separado.

Las guías de API describen las categorías de uso de caché y sus condiciones por modelo. La calibración deberá comprobar qué controles y señales se exponen en Claude Code y Codex para cada superficie y modalidad. [Caché de OpenAI API](https://developers.openai.com/api/docs/guides/prompt-caching) · [Caché de Claude API](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) · [Uso de Codex CLI](https://learn.chatgpt.com/docs/non-interactive-mode) · [Telemetría de Claude Code](https://code.claude.com/docs/en/monitoring-usage)

### Evaluación y decisión

Usar el formato de 02 con campos adicionales para piloto, brazo, par, parámetros solicitados y efectivos, condición de caché prevista y observada, preparación y motivos de exclusión. El total de una cohorte incluye todas las corridas iniciadas según el protocolo, incluso fallidas y reintentos. Los agregados de sesión y detalle por worker se reconcilian sin duplicar uso. Informar cualquier desvío o exclusión del contraste y su consumo aparte para transparentar el gasto del piloto.

En API, comparar costo total por tarea aceptada con fuente y alcance, distinguiendo importes observados de estimaciones. En suscripción, usar la cuota atribuible si la superficie la ofrece o una métrica técnica acordada que se reporte por separado como tal. Publicar tiempo transcurrido, tiempo humano, cuota, calidad, dispersión y cobertura por cohorte. Con cero aceptaciones, cobertura insuficiente o muestra menor a la acordada, indicar que falta evidencia para adoptar.

La recomendación es independiente para A y B. Adoptar exige alcanzar la mejora acordada y conservar calidad comparable o superior, sin regresiones en criterios obligatorios. Una regresión obligatoria descarta la adopción. Si faltan señales o la dispersión impide aplicar el criterio, recomendar repetir con el faltante explicitado. Un resultado con evidencia suficiente que quede por debajo del umbral puede descartarse para esa cohorte.

### Decisiones para acordar

Todas las filas están pendientes. Los valores y variantes de este borrador son candidatos para discusión.

| ID | Qué acordar | Punto de partida | Changes afectados |
|---|---|---|---|
| D1 | Tareas, revisiones y alcance por cliente | Reusar fixtures y rúbricas de 02 y el contrato de 07. Empezar con los CLI y cubrir tareas directas y delegadas. Fijar versiones, modelos, perfiles y modalidades probados. | 02, 03, 04, 07, 09 |
| D2 | Variantes de modelo y esfuerzo | Empezar por un rol por contraste, con routing nativo. Elegir modelos, niveles, evidencia mínima y tratamiento de fallback. Decidir por separado si vale incluir Jev como brazo opcional. | 03, 04, 07, 09, 10 |
| D3 | Variantes y preparación de caché | Comparar prefijo actual y estable en frío y caliente. Acordar controles disponibles, ventana y tamaño del prefijo, señal exigida, aislamiento y horizonte de imputación de la preparación. | 02, 09 |
| D4 | Métrica objetivo y cobertura | Acordar por modalidad costo o consumo por tarea aceptada, cobertura de coordinador y workers, unidad de cuota y criterios de comparabilidad. Fijar qué se puede concluir con señales parciales. | 02, 07, 09 |
| D5 | Muestra, presupuesto y decisión | Acordar repeticiones, topes de dinero, cuota y tiempo, regla de parada, margen de calidad y mejora mínima por cohorte. Fijar el criterio antes de medir y registrar quién revisa la recomendación. | 02, 07, 09 |

Matías incorpora los acuerdos y los responsables antes de compartir la propuesta final para revisión de los siete.

## Risks / Trade-offs

- **Contaminación entre brazos**. Aislar la preparación y registrar la evidencia de lectura. Si no se puede verificar el aislamiento, limitar la conclusión al alcance observable.
- **Ahorro aparente por cobertura desigual**. Publicar fallos, desvíos, exclusiones y preparación. Usar el mismo criterio de comparabilidad para los dos brazos.
- **Cambio de versión o fallback durante la muestra**. Segregar el bloque afectado y repetirlo si el protocolo lo permite.
- **Orden del contexto y calidad**. Mantener contenido y herramientas equivalentes y evaluar con la misma rúbrica. Registrar si la variante de prefijo modifica el tamaño de contexto.
- **Exposición de trazas o costos de prueba**. Heredar la política de evidencia de 02 y fijar perfiles aislados, cuentas y límites antes de ejecutar.

## Migration Plan

Versionar variantes y resultados en el arnés externo, junto con el manifiesto de 02 y la revisión de 07. La limpieza de perfiles y la desactivación de las variantes dejan el entorno listo para volver al control. Una recomendación aceptada se lleva a una decisión de adopción posterior.

## UI Mockups

Decision: no (no UI)
