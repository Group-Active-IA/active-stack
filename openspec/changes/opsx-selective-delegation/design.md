# Design: opsx-selective-delegation

## Context

El orquestador se compone en la instalación (`internal/harness/config/compose.go`) desde `assets/<agente>/sdd-orchestrator.md` más toggles, y se inyecta con markers en `CLAUDE.md`/`AGENTS.md`. La delegación aparece en dos lugares: la sección marcada `sdd-delegation` (tabla de umbrales, removible) y el texto base del rol (obligaciones incondicionales, no removibles). La proposal cita las líneas.

Las fichas vecinas ya fijaron sus piezas, y esta las consume sin redefinirlas:

| Pieza | Dueña | Qué toma 07 |
|---|---|---|
| Activación explícita del coordinador | 03 / 04 (decisión confirmada: skill invocable `/sdd-orchestrator` en Claude con `disable-model-invocation: true`, `$sdd-orchestrator` en Codex con `allow_implicit_invocation: false`) | dónde vive la regla de decisión |
| Carga de módulos por acción, módulo obligatorio ausente detiene la acción | 05 `task-context-loading` | el disparador: "la acción declara que requiere coordinación de subagentes" |
| Salida acotada con referencia al original | 05 `bounded-command-output` | el formato de los hallazgos verbosos del worker |
| Recibo, integridad, procedencia, suficiencia | 06 `execution-receipts` (decisión confirmada: un reporte del worker por sí solo queda sin verificar) | el campo `recibos` del retorno y el veredicto *sin verificar* |
| Controles preventivos por cliente | 08 `native-controls-pilot` | la prevención de escrituras fuera de alcance donde sea obligatoria |
| Protocolo de medición, registro por intento, costo por tarea aceptada | 02 `opsx-benchmarking` (`240e1b2`) | la contabilidad del piloto |
| Experimentos de modelo y caché | 09 (`89f1fa5`) | usa la configuración de esta ficha como control A0 |

## Goals / Non-Goals

**Goals**

- Que delegar sea una decisión con señal registrada, no un reflejo.
- Que la independencia de revisión, cuando es obligatoria, no pueda perderse por ahorro.
- Que un worker no pueda pisar lo de otro, ni cargar con cambios ajenos, sin que el coordinador lo detecte.
- Que el retorno se compruebe contra un punto de partida y los recibos, no contra la prosa.
- Que la política predeterminada se adopte con datos.

**Non-Goals**

- Bloquear escrituras fuera de alcance antes de que ocurran (08).
- Elegir modelo por rol (09).
- Construir un runtime de coordinación o un ledger.

## Decisions

### D1 — Directo por defecto, delegar por señal, con independencia obligatoria primero

La decisión se evalúa en este orden. El primer paso que aplica la cierra.

1. **Independencia obligatoria (S4).** Si la acción exige que quien revisa no sea quien construyó (por ejemplo verify sobre un apply), la revisión se asigna a un agente distinto. Ninguna contraseñal la anula: una tarea breve o un conflicto de archivos no habilitan que el constructor se revise a sí mismo. Si no hay módulo de delegación o revisor habilitado (toggle apagado, cliente sin subagentes, módulo ausente), esa verificación se detiene y el coordinador explica qué falta; las demás tareas siguen (decisión confirmada sobre controles obligatorios).
2. **Contención (C3).** Si el conjunto de escritura se solapa con el de un encargo abierto, se serializa o se redefine el alcance hasta que sea disjunto. Nunca se resuelve corriendo en paralelo.
3. **Contraseñales C1, C2, C4.** Si alguna aplica, se resuelve directo.
4. **Señales S1–S3.** Si alguna aplica, se delega.
5. **Por defecto**, directo.

| Señal | Qué la dispara | Por qué justifica delegar |
|---|---|---|
| S1 salida verbosa | la tarea va a producir más de 300 líneas de salida (tests, build, logs), umbral de 05 D19; supuesto a calibrar con los fixtures de salida de 02/05 | el contexto principal recibe sólo el resumen acotado y el recibo |
| S2 exploración con conclusión | la respuesta es una conclusión (dónde está X, qué rompe Y) y no el contenido leído | el principal no necesita los archivos, sólo el hallazgo |
| S3 trabajo paralelizable | dos o más subtareas cumplen D4 | gana tiempo de espera; el consumo se mide aparte |
| S4 separación de rol | la acción exige que quien revisa no sea quien construyó | independencia, no ahorro; tiene prioridad |

| Contraseñal | Por qué obliga a resolver directo |
|---|---|
| C1 tarea breve | el costo de armar el encargo y releer en frío supera el trabajo |
| C2 depende del contexto vivo | la tarea necesita decisiones de la conversación que habría que transcribir enteras |
| C3 conflicto de archivos | se solapa con un encargo abierto; serializar o redefinir alcance |
| C4 el principal va a releer todo | si el resultado se va a leer completo igual, delegar duplica |

**Alternativas.** (a) Mantener el umbral de archivos: simple, pero mide lo que se lee y no lo que se ahorra. (b) Delegar siempre (estado actual): sin datos que lo respalden. (c) Umbral en tokens: más fiel, pero el agente no conoce el tamaño antes de leer. (d) Señales sin orden (versión `0597ebc`): permitía que C1 o C3 anularan S4; corregido por la auditoría del 10/11.

**Registro.** Cada decisión escribe `{acción, paso que la cerró, señal|contraseñal, directo|delegado|detenido}` en el informe de la ronda. Es el insumo del piloto y de la auditoría.

### D2 — La regla en la skill del coordinador, el módulo bajo demanda

La regla de D1 tiene que estar disponible **antes** de decidir, así que va en la skill de invocación explícita del coordinador acordada para 03/04 (texto breve, presupuesto `(TBD con 03/04)`). Todo lo demás —encargo, propiedad, paralelismo, retorno, reintentos— va en el módulo de delegación y coordinación, que la skill carga con el mecanismo de 05 sólo si la decisión fue delegar.

Si el módulo no está disponible: con S4 la verificación se detiene y se explica (D1 paso 1); sin S4, el coordinador resuelve directo y lo registra.

**Alternativa.** Todo en la skill: más simple, pero carga el módulo en cada invocación, que es justo lo que 05 quiere evitar.

### D3 — Encargo mínimo

| Campo | Contenido |
|---|---|
| objetivo | una frase verificable |
| acción | explore / propose / apply / verify / archive |
| partida | commit base y huella del estado de trabajo entregado al worker (D4) |
| escritura | lista exclusiva de rutas o globs que el worker puede crear, modificar, renombrar o borrar (vacía = sólo lectura) |
| lectura | punteros por ruta o símbolo (05), no contenido pegado |
| límites | lo que no se toca y lo que no se decide |
| salida | criterios de "listo" observables |
| retorno | contrato de D6, incluido el extracto `## Evidence Contract` de 06 cuando ejecuta comandos |
| intentos | tope (D7) |

**Archivos calientes.** `go.mod`, `go.sum`, `internal/catalog/harnesses.yaml`, `assets/*/sdd-orchestrator.md`, `CLAUDE.md`/`AGENTS.md` nunca se reparten entre workers concurrentes: los toca uno solo o el coordinador.

### D4 — Propiedad de archivos contra un punto de partida

El diff global de un árbol compartido puede atribuirle al worker cambios previos del usuario o de otro encargo. Por eso la propiedad se comprueba contra la partida de cada encargo y no contra el árbol actual:

1. **Partida identificada.** Antes de lanzar, el coordinador registra el commit base y una huella del estado de trabajo (cambios preparados, sin preparar y archivos nuevos) que recibe el worker. Un worker con escritura trabaja en su propio árbol de trabajo creado desde esa partida.
2. **Cambios atribuibles.** Al retorno, se calcula lo que cambió entre la partida y el estado final de ese árbol, incluyendo: commits nuevos del worker, cambios preparados, cambios sin preparar, archivos nuevos sin trackear, borrados y renombres (origen y destino cuentan como rutas tocadas).
3. **Comparación.** Toda ruta atribuible fuera de `escritura` → retorno `out_of_scope`, aunque esté confirmada en un commit y los tests pasen. Lo que ya existía en la partida no se le atribuye.
4. **Integración.** Antes de integrar se vuelve a comprobar el alcance sobre el resultado combinado y se re-ejecutan los escenarios que dependen de combinar cambios de más de un encargo.

La comprobación es posterior a la escritura. Donde 08 declare un control preventivo obligatorio para escrituras fuera de alcance, se usa además de esta; si no está disponible, aplica la regla de controles obligatorios (detener sólo esa acción y explicar).

**Prevención sin dueño todavía.** El inventario de 08 (`opsx-native-controls/tasks.md`, tarea 2.3) clasifica hoy sólo `.env` y git destructivo. Se le pide a 08 que sume "escritura fuera del conjunto de escritura del encargo" como candidata a "debe bloquearse". Hasta que 08 la incorpore, esta ficha la declara **no controlada preventivamente**: la única defensa es la comprobación posterior de este punto.

### D5 — Paralelismo (decisión confirmada)

Hasta **dos** agentes con escritura en paralelo durante el piloto, cada uno en su propio árbol de trabajo. Se paralelizan sólo encargos que cumplen las tres condiciones:

1. conjuntos de escritura disjuntos (incluidos archivos calientes);
2. ninguno depende de lo que el otro escribe en ese encargo;
3. cada uno tiene criterio de salida y pruebas propias.

Los que dependen entre sí o comparten archivos se serializan. El resultado combinado se verifica antes de integrar (D4 paso 4). Workers de sólo lectura pueden compartir árbol.

**Alternativa descartada.** Paralelizar en el mismo árbol con conjuntos disjuntos: un `go test ./...` de un worker ve a medio escribir lo del otro y produce recibos inválidos para 06.

### D6 — Retorno estructurado y veredicto

```
estado:        completed | blocked | out_of_scope | failed
partida:       commit base + huella declarada
modificados:   [rutas, con tipo: añadido | modificado | borrado | renombrado(origen→destino)]
hallazgos:     [{severidad, ruta:línea, descripción}]   # acotados según 05
recibos:       [recibo 06]                              # uno por comando ejecutado
pendiente:     texto breve | —
```

**Dueño del sobre.** Esta ficha es dueña del sobre de retorno; 06 es dueña del campo `recibos`. Las fichas de roles nativos (03, 04) lo referencian en lugar de definir otro: `codex-native-opsx/design.md:216` hoy propone un retorno propio (estado, hallazgos, evidencia, módulos cargados), que se le pide alinear con este. Si el grupo necesita `módulos cargados`, se agrega acá como campo opcional.

El coordinador valida en orden: (1) la partida coincide con la registrada; (2) `modificados` coincide con los cambios atribuibles de D4 y está dentro de `escritura`; (3) recibos según 06; (4) criterios de salida.

**Veredicto.** Lo que el worker declara se conserva, pero queda *sin verificar* hasta que una fuente independiente o una re-ejecución corrobora la ejecución y el estado del código, y la evidencia alcanza para el requisito (decisión confirmada para 05/06). Un retorno con "tests verdes" sin esa corroboración sigue sin verificar. El mismo contrato y el mismo paso de revisión de recibos aplican cuando la acción se resuelve directo y requiere evidencia.

### D7 — Reintentos

Hasta 3 intentos en total por encargo (alineado con "tres intentos fallidos requieren dirección humana"). Cada reintento recibe el motivo del rechazo y parte de una nueva partida registrada. Agotado el tope, el coordinador escala con el historial. Todos los intentos cuentan en el costo de la cohorte.

### D8 — Toggle `delegation` sin residuo

Hoy el toggle sólo quita la tabla. Se mueve al módulo toda obligación de delegar del texto base y se deja en la skill sólo la regla de D1. Sin el toggle: el bloque compuesto no contiene `ALWAYS delegate`, `Delegate, don't inflate` ni la tabla; el coordinador declara que trabaja sin subagentes y, cuando una acción exige S4, detiene esa verificación y lo explica. Se prueba con golden en `compose_test.go`.

### D9 — Piloto

**Arnés.** El de 02, externo al binario. Sin subcomando nuevo.

**Brazos.**

| Brazo | Configuración |
|---|---|
| A directo | coordinador sin delegar salvo S4 |
| B selectivo | regla D1 + módulo bajo demanda |
| R referencia | flujo de `1b4639b` (delegar siempre), tomado de la referencia de 02 cuando exista |

**Equivalencia entre brazos.** Misma tarea, fixture, rúbrica, criterios obligatorios y controles en todos los brazos, con resultados por cliente e interfaz (terminal e IDE por separado, según la decisión confirmada para 02). Cuando una tarea exige revisión independiente, la revisión separada se mantiene en ambos brazos; si un brazo no puede proveerla, el caso se excluye del contraste directo y se informa aparte.

**Clases de tarea** (fixtures de 02 cuando existan):

| Clase | Ejemplo | Paso de D1 esperado |
|---|---|---|
| T1 breve | corregir un mensaje y su test en un paquete | C1 → directo |
| T2 verbosa | diagnosticar un fallo con salida de miles de líneas (fixtures de 05) | S1 |
| T3 exploración | "¿qué rutas escriben en `settings.json`?" | S2 |
| T4 paralelizable | dos cambios en paquetes disjuntos | S3, dos worktrees |
| T5 contención | dos subtareas que tocan `harnesses.yaml` | C3 → serializar |
| T6 verificación breve con independencia | verificar un apply de una línea | S4 prevalece sobre C1 |
| T7 ajeno en la partida | árbol con un archivo modificado por el usuario fuera del encargo | no atribuido al worker |
| T8 commit fuera de alcance | el worker confirma un cambio en una ruta ajena | `out_of_scope` |

**Medidas** (registro por intento y corrida de 02, contando principal, workers, fallos y reintentos):

- costo API = consumo medido por categoría × tarifa vigente, informado como facturado o estimado;
- costo por tarea aceptada = costo total de la cohorte ÷ tareas aceptadas; con cero aceptadas se informa el total y el cociente queda sin valor;
- cuota, tokens (entrada nueva, lectura y creación de caché, salida) y latencia como medidas separadas; nunca se convierten tokens en cuota;
- **ahorro de costo** y **reducción del tiempo de espera** se reportan por separado: el paralelismo puede bajar la espera y subir el costo;
- puerta de calidad: criterios obligatorios cumplidos, cero cambios fuera de alcance, cero regresiones;
- secundarias: pico de contexto del principal, tiempo humano, reintentos, conflictos detectados, tasa de acierto de D1 contra el paso esperado.

**Calibración y adopción.** Las tres corridas por celda y el 10 % de mejora son una **propuesta de calibración**, no el criterio. Tras la calibración, el grupo fija tamaño de muestra, dispersión tolerable, umbral de mejora, margen de calidad y regla para resultados inconclusos **antes** de la prueba de adopción.

**Resultado.** Informe de decisión con datos crudos y una de tres salidas: adoptar B como predeterminado, mantener con ajustes, o repetir con muestra mayor.

**Qué recibe 09 como control A0** (`opsx-model-cache-pilots/design.md:25`), según la salida:

| Salida de 07 | Configuración que recibe 09 |
|---|---|
| adoptar | brazo B (selectivo) en la revisión de 07 que se adoptó |
| mantener con ajustes | la política vigente con los ajustes registrados en el informe, como revisión identificada |
| repetir | 09 no corre su prueba de adopción; puede calibrar sobre A y B por separado, con tareas directas y delegadas estratificadas y el brazo declarado en cada corrida |

## Risks / Trade-offs

- **El agente ignora la regla de D1.** Una instrucción no es un control. Mitigación: el registro permite medir la tasa de acierto contra el paso esperado de cada fixture.
- **La comprobación de propiedad es posterior.** Mitigación: árbol propio por worker, partida registrada, comprobación antes de integrar y controles preventivos de 08 donde sean obligatorios.
- **Huella del estado de trabajo.** Si el worker recibe cambios sin commit, la partida depende de una huella bien definida; se reutiliza la de 06 (huella del estado probado) para no tener dos definiciones.
- **Seis fichas tocan `sdd-orchestrator.md`.** Mitigación: esta ficha escribe contenido de módulo y una regla breve de skill, y se integra después de 03/04.
- **Directo por defecto puede inflar el contexto en tareas largas.** Por eso S1 y S2 existen y el piloto mide el pico de contexto del principal.

## Migration Plan

No aplica en esta ronda. En la implementación, el texto llega por el mismo camino de inyección con markers y backup que ya usan 03/04; esta ficha no agrega una ruta de migración propia.

## UI Mockups

Decision: no (no UI) — 2026-11-10. El change no agrega ni modifica pantalla.

## Decisiones

### Confirmadas por el grupo (10/11)

| Tema | Decisión |
|---|---|
| Paralelismo | Hasta dos agentes con escritura en paralelo en el piloto, cada uno en su árbol; serializar dependientes o con archivos compartidos; verificar el resultado combinado antes de integrar; la revisión con independencia la hace un agente distinto del que implementó. |
| Activación del coordinador (03/04) | Skill invocable explícitamente; la regla de D1 vive ahí. |
| Reporte del worker (05/06) | Queda sin verificar hasta corroboración independiente o re-ejecución. |
| Controles obligatorios | Si falta, se detiene sólo la acción que depende de él, se explica y se sigue con lo demás. |

### Pendientes

| Tema | Opciones | Recomendación | Afecta a |
|---|---|---|---|
| Señales que justifican delegar | S1–S4 con orden de D1 / sumar umbral numérico | S1–S4 con orden; el piloto las valida | 05, 09 |
| Métrica objetivo de adopción | costo por tarea aceptada / tiempo de espera / pico de contexto | costo por tarea aceptada (API y cuota aparte), con puerta de calidad; espera como secundaria | 02, 09 |
| Muestra, umbral y regla inconclusa | — | fijar tras la calibración, antes de la prueba de adopción | 02, 09 |
| Toggle `delegation` en `lite` | mantener / sacar de `lite` | mantener hasta tener datos | catálogo |
| Dueño de `sdd-orchestrator.md` | 03+04 / sin dueño | 03+04, como propone 05 | 02–07, 09 |

## Cambios por la auditoría del 10/11

| Observación | Dónde se incorporó |
|---|---|
| S4 podía ser anulada por C1/C3 | D1 con orden de evaluación; requisito "Independencia obligatoria primero"; T6 |
| Delegación desactivada con independencia exigida | D1 paso 1, D8; escenario en spec |
| Diff global atribuye cambios ajenos | D4 con partida identificada y árbol propio; T7 |
| Preparados, sin preparar, confirmados, renombrados, nuevos | D4 paso 2; formato de `modificados` en D6; T8 |
| Re-ejecutar escenarios del resultado combinado | D4 paso 4; D5 |
| Prevención de 08 donde sea obligatoria | D4 último párrafo |
| Alinear con 02, 06 y 09 | D9 (equivalencia de brazos, medidas, costo de cohorte, calibración) |
| Pruebas sin corroboración siguen sin verificar | D6 veredicto |
| Separar ahorro de costo y tiempo de espera | D9 medidas |

## Cambios por la auditoría complementaria del 10/11

| Observación | Dónde se incorporó |
|---|---|
| S1 apuntaba a un "presupuesto de 05" que no existe | D1 tabla de señales y escenario de salida verbosa: más de 300 líneas (05 D19), como supuesto a calibrar |
| Prevención de escrituras fuera de alcance sin dueño | D4 "Prevención sin dueño todavía"; tarea 0.2 |
| Tres formatos de retorno (04, 06, 07) | D6 "Dueño del sobre" |
| Qué configuración recibe 09 como A0 | D9 tabla "Qué recibe 09"; tarea 6.3 |
