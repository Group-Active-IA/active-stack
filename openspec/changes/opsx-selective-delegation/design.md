# Design: opsx-selective-delegation

## Context

El orquestador se compone en la instalación (`internal/harness/config/compose.go`) desde `assets/<agente>/sdd-orchestrator.md` más toggles, y se inyecta con markers en `CLAUDE.md`/`AGENTS.md`. La delegación aparece en dos lugares: la sección marcada `sdd-delegation` (tabla de umbrales, removible) y el texto base del rol (obligaciones incondicionales, no removibles). La proposal cita las líneas.

Las fichas vecinas ya fijaron sus piezas, y esta las consume sin redefinirlas:

| Pieza | Dueña | Qué toma 07 |
|---|---|---|
| Carga de módulos por acción, módulo obligatorio ausente detiene la acción | 05 `task-context-loading` | el disparador: "la acción declara que requiere coordinación de subagentes" |
| Salida acotada con referencia al original | 05 `bounded-command-output` | el formato de los hallazgos verbosos del worker |
| Recibo de ejecución, procedencia, suficiencia, estados | 06 `execution-receipts` | el campo `recibos` del retorno y el veredicto *sin verificar* |
| Layout de módulos (`modules/{tdd,delegation,subagents,...}`), roles nativos | 03 / 04 | dónde se escribe el contenido |
| Protocolo de medición, métricas por rol, API vs cuota | 02 `opsx-baseline` (sin rama al 10/11) | la contabilidad del piloto |

## Goals / Non-Goals

**Goals**

- Que delegar sea una decisión con señal registrada, no un reflejo.
- Que un worker no pueda pisar lo de otro sin que el coordinador lo detecte.
- Que el retorno se compruebe con el diff y los recibos, no con la prosa.
- Que la política predeterminada se adopte con datos.

**Non-Goals**

- Bloquear escrituras fuera de alcance antes de que ocurran (08).
- Elegir modelo por rol (09).
- Construir un runtime de coordinación o un ledger.

## Decisions

### D1 — Directo por defecto, delegar por señal

**Decisión.** El coordinador resuelve directo salvo que se cumpla al menos una señal y ninguna contraseñal.

| Señal | Qué la dispara | Por qué justifica delegar |
|---|---|---|
| S1 salida verbosa | la tarea va a producir salida que supera el presupuesto de 05 (tests, build, logs) | el contexto principal recibe sólo el resumen acotado + recibo |
| S2 exploración con conclusión | la respuesta es una conclusión (dónde está X, qué rompe Y) y no el contenido leído | el principal no necesita los archivos, sólo el hallazgo |
| S3 trabajo paralelizable | dos o más subtareas cumplen D4 | gana tiempo de pared; el consumo se mide aparte |
| S4 separación de rol | la acción exige que quien revisa no sea quien construyó (verify) | independencia, no ahorro |

| Contraseñal | Por qué obliga a resolver directo |
|---|---|
| C1 tarea breve | el costo de armar el encargo y releer en frío supera el trabajo |
| C2 depende del contexto vivo | la tarea necesita decisiones de la conversación que habría que transcribir enteras |
| C3 conflicto de archivos | el conjunto de escritura se solapa con el de un worker en curso (D3) |
| C4 el principal va a releer todo | si el resultado se va a leer completo igual, delegar duplica |

**Alternativas.** (a) Mantener el umbral de archivos: simple, pero mide lo que se lee y no lo que se ahorra; un archivo de 5.000 líneas pasa "inline". (b) Delegar siempre (estado actual): sin datos que lo respalden. (c) Umbral en tokens: más fiel, pero el agente no conoce el tamaño antes de leer. Se elige señales cualitativas + registro, y el piloto valida si funcionan.

**Registro.** Cada decisión escribe `{acción, señal|contraseñal, directo|delegado}` en el informe de la ronda. Es el insumo del piloto y de la auditoría; no requiere código nuevo.

### D2 — La regla en el rol, el módulo bajo demanda

La regla de D1 tiene que estar disponible **antes** de decidir, así que va en la definición del coordinador (texto breve, presupuesto `(TBD con 03/04)`). Todo lo demás —encargo, propiedad, paralelismo, retorno, reintentos— va en el módulo de delegación y coordinación, que se carga con el mecanismo de 05 sólo si la decisión fue delegar. Si el módulo no está disponible, aplica el requisito de 05: se informa y se resuelve directo o se detiene la acción si la acción exige delegar (S4).

**Alternativa.** Todo en el rol: más simple, pero carga el módulo en cada turno, que es justo lo que 05 quiere evitar.

### D3 — Encargo mínimo y propiedad de archivos

Campos obligatorios del encargo:

| Campo | Contenido |
|---|---|
| objetivo | una frase verificable |
| acción | explore / propose / apply / verify / archive |
| escritura | lista exclusiva de rutas o globs que el worker puede modificar (vacía = sólo lectura) |
| lectura | punteros por ruta o símbolo (05), no contenido pegado |
| límites | lo que no se toca y lo que no se decide |
| salida | criterios de "listo" observables |
| retorno | contrato de D5, incluido el extracto `## Evidence Contract` de 06 cuando ejecuta comandos |
| intentos | tope (D6) |

**Propiedad.** El coordinador mantiene un mapa `ruta → worker` mientras el encargo está abierto. Al recibir el retorno compara `git diff --name-only` (y archivos nuevos sin trackear) contra `escritura`. Cualquier ruta afuera → retorno rechazado con estado `out_of_scope`, aunque el resto esté bien. La comprobación es posterior a la escritura: prevenirla es tarea de 08.

**Archivos calientes.** `go.mod`, `go.sum`, `internal/catalog/harnesses.yaml`, `assets/*/sdd-orchestrator.md`, `CLAUDE.md`/`AGENTS.md` nunca se reparten entre workers concurrentes: los toca uno solo o el coordinador.

### D4 — Paralelismo

Dos workers corren en paralelo sólo si se cumplen las tres condiciones:

1. conjuntos de escritura disjuntos (incluidos archivos calientes);
2. ninguno lee lo que el otro escribe en ese encargo;
3. cada uno tiene criterio de salida y verificación propios.

Workers con escritura en paralelo trabajan en worktrees separados para que tests y builds no se pisen; workers de sólo lectura pueden compartir árbol. Concurrencia máxima: decisión de grupo (recomendación: 2 en el piloto). Si una condición falla, se serializa y se registra `C3`.

**Alternativa.** Paralelizar en el mismo árbol con conjuntos disjuntos: más barato, pero un `go test ./...` de un worker ve a medio escribir lo del otro y produce recibos inválidos para 06.

### D5 — Retorno estructurado

```
estado:        completed | blocked | out_of_scope | failed
modificados:   [rutas]                  # se contrasta con el diff
hallazgos:     [{severidad, ruta:línea, descripción}]   # acotados según 05
recibos:       [recibo 06]              # uno por comando ejecutado
pendiente:     texto breve | —
```

El coordinador valida en orden: (1) `modificados` ⊆ `escritura` y coincide con el diff; (2) recibos según 06; (3) criterios de salida. Campo faltante o no comprobable → *sin verificar*, nunca aceptado por inferencia. Es el mismo default que fija 06.

### D6 — Reintentos

Un encargo admite hasta 3 intentos en total (alineado con la regla vigente de "tres intentos fallidos requieren dirección humana"). Cada reintento recibe el motivo del rechazo y nada más. Agotado el tope, el coordinador escala con el historial. Los tokens de todos los intentos se suman al costo de la tarea.

### D7 — Toggle `delegation` sin residuo

Hoy el toggle sólo quita la tabla. Se propone mover al módulo (o a la sección marcada) toda obligación de delegar del texto base y dejar en el rol sólo la regla de D1. Sin el toggle: el bloque compuesto no contiene `ALWAYS delegate`, `Delegate, don't inflate` ni la tabla, y el coordinador declara que trabaja sin subagentes (coincide con `codex-native-opsx` design:197). Se prueba con golden en `compose_test.go`.

### D8 — Piloto

**Arnés.** Externo al binario, en un directorio de evaluación del repo o fuera de él `(TBD con 02)`. Sin subcomando nuevo.

**Brazos.**

| Brazo | Configuración |
|---|---|
| A directo | coordinador sin módulo de delegación; todo inline |
| B selectivo | regla D1 + módulo bajo demanda |
| R referencia | flujo actual de `1b4639b` (delegar siempre), medido con 02 |

A vs B es la comparación de la ficha. R ancla contra lo que hoy se instala.

**Clases de tarea** (fixtures sobre una copia fija del repo en un commit dado):

| Clase | Ejemplo | Señal esperada |
|---|---|---|
| T1 breve | corregir un mensaje y su test en un paquete | C1 → directo |
| T2 verbosa | diagnosticar un fallo con salida de miles de líneas (fixtures de 05) | S1 |
| T3 exploración | "¿qué rutas escriben en `settings.json`?" | S2 |
| T4 paralelizable | dos cambios en paquetes disjuntos | S3 |
| T5 contención | dos subtareas que tocan `harnesses.yaml` | C3 → serializar |

**Métricas** (por tarea aceptada, contando principal + workers + reintentos):

- primaria: costo por tarea aceptada = entrada sin caché + entrada cacheada + salida, en costo API y en cuota por separado (02);
- puerta de calidad: aceptación igual o mejor, cero cambios fuera de alcance, cero regresiones;
- secundarias: pico de contexto del principal, tiempo de pared, tiempo humano, reintentos, conflictos detectados, decisiones D1 que contradicen la señal esperada.

**Diseño.** Orden alternado entre brazos, caché fría y caliente separadas, mismas versiones de cliente y modelo, un cambio de política por vez. Repeticiones y umbral de adopción: decisión de grupo (recomendación: ≥3 corridas por tarea y brazo; adoptar B si baja ≥10 % la métrica primaria sin perder la puerta de calidad).

**Resultado.** Un informe de decisión con los datos crudos y una de tres salidas: adoptar B como predeterminado, mantener R/A con ajustes, o repetir con muestra mayor si la incertidumbre cruza el umbral.

## Risks / Trade-offs

- **El agente ignora la regla de D1.** Una instrucción no es un control. Mitigación: el registro de decisiones permite medir la tasa de acierto contra la señal esperada de cada fixture; si es baja, el piloto lo muestra.
- **La comprobación de propiedad es posterior.** Un worker puede romper algo antes del rechazo. Mitigación: worktrees separados en paralelo y diff antes de integrar; la prevención es de 08.
- **02 no tiene rama todavía.** La contabilidad del piloto depende de un protocolo no escrito. Mitigación: las tareas del piloto quedan bloqueadas por 02 y esta ficha sólo declara qué métricas necesita.
- **Seis fichas tocan `sdd-orchestrator.md`.** Riesgo de reescritura. Mitigación: esta ficha escribe contenido de módulo y una regla breve de rol, y se integra después de 03/04.
- **Directo por defecto puede inflar el contexto en tareas largas.** Por eso S1 y S2 existen y el piloto mide el pico de contexto del principal.

## Migration Plan

No aplica en esta ronda. En la implementación, el texto llega por el mismo camino de inyección con markers y backup que ya usan 03/04; esta ficha no agrega una ruta de migración propia.

## UI Mockups

Decision: no (no UI) — 2026-11-10. El change no agrega ni modifica pantalla.

## Decisiones que el grupo tiene que cerrar

| Tema | Opciones | Recomendación | Afecta a |
|---|---|---|---|
| Señales que justifican delegar | S1–S4 / sumar umbral numérico / mantener conteo de archivos | S1–S4 + registro; el piloto las valida | 05, 09 |
| Qué se paraleliza | nada / sólo lectura / escritura disjunta en worktrees | escritura disjunta en worktrees, máx. 2 | 03, 04, 08 |
| Métrica de adopción | costo por tarea aceptada / tiempo de pared / pico de contexto | costo por tarea aceptada (API y cuota aparte), con puerta de calidad | 02, 09 |
| Umbral y repeticiones | — | ≥3 corridas, ≥10 % de mejora | 02 |
| Dónde vive la regla de D1 | rol / módulo | rol (breve) + módulo bajo demanda | 03, 04, 05 |
| Toggle `delegation` en `lite` | mantener / sacar de `lite` | mantener hasta tener datos del piloto | catálogo |
| Dueño de `sdd-orchestrator.md` | 03+04 / sin dueño | 03+04, como propone 05 | 02, 03, 04, 05, 07, 09 |

## Open Questions

- ¿El brazo R se mide en este piloto o se reutiliza la corrida de referencia de 02? Depende de cuándo exista 02.
- ¿Codex expone en todas las superficies el consumo por subagente? Si no, el límite se declara y la comparación en Codex se hace a nivel sesión.
