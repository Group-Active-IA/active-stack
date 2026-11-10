# Tasks: opsx-selective-delegation

> Ronda `propose`: todas las tareas quedan pendientes. Gobernanza MEDIO: implementar con checkpoints.
> Prerrequisitos de implementación: 05 y 06 aprobadas e integradas; layout de módulos de 03/04 acordado; protocolo de 02 disponible para el grupo 4.

## 0. Decisiones previas (grupo)

- [ ] 0.1 Cerrar en grupo las decisiones de la tabla de `design.md` (señales, paralelismo, métrica, umbral, ubicación de la regla, toggle en `lite`, dueño de `sdd-orchestrator.md`) y volcarlas a los artefactos.
- [ ] 0.2 Confirmar con 05 el nombre del disparador "la acción requiere coordinación de subagentes" y con 06 el texto del extracto `## Evidence Contract`.

## 1. Inventario reproducible

- [ ] 1.1 Listar con `Compose` cada obligación de delegar dentro y fuera de los markers `sdd-delegation`, para Claude y Codex, con y sin el toggle. Aceptación: tabla con línea, texto y si sobrevive al toggle.
- [ ] 1.2 Documentar por cliente cómo se lanza, espera y cierra un worker y qué consumo expone por worker (Claude Agent tool; Codex spawn/`wait_agent`/`close_agent`). Aceptación: matriz con fuente y versión probada; lo no observado se marca SUPOSICIÓN.

## 2. Semántica del toggle (TDD)

- [ ] 2.1 RED: golden en `compose_test.go` que falla si el bloque sin `delegation` contiene `ALWAYS delegate`, `Delegate, don't inflate` o la tabla, para Claude y Codex.
- [ ] 2.2 RED: test que falla si el bloque con `delegation` contiene el umbral "1-3 / 4+".
- [ ] 2.3 GREEN: mover las obligaciones al módulo y dejar en el rol la regla de decisión y la declaración sin subagentes. Depende de 1.1 y del layout de 03/04.
- [ ] 2.4 Regresión: `go test ./internal/harness/config/... ./internal/filemerge/... ./internal/uninstall/...` sin fallos nuevos; purge de markers sigue reconociendo `sdd-delegation`.

## 3. Contenido del módulo (TDD sobre texto compuesto)

- [ ] 3.1 RED: test de presencia de los ocho campos del encargo, el contrato de retorno y la regla de propiedad en el módulo de cada cliente.
- [ ] 3.2 GREEN: redactar regla de decisión (rol), encargo, propiedad, archivos calientes, paralelismo, retorno y reintentos (módulo).
- [ ] 3.3 RED/GREEN: golden de paridad que compara la sección de delegación de Claude y Codex salvo la sintaxis de lanzamiento declarada.

## 4. Batería de veredictos

- [ ] 4.1 Armar fixtures de retorno: completo, prosa libre, fuera de alcance, sin recibos, `modificados` incompleto, tope de intentos agotado.
- [ ] 4.2 Correr la batería con el flujo de Claude y el de Codex. Aceptación: mismo veredicto por fixture; diferencias de cliente declaradas como límite.

## 5. Piloto (depende de 02, 05 y 06)

- [ ] 5.1 Arnés externo al binario; verificar que `cmd/` e `internal/` no cambian.
- [ ] 5.2 Fixtures T1–T5 sobre un commit fijo, cada uno con aceptación y señal esperada; T2 reutiliza los fixtures de salida verbosa de 05.
- [ ] 5.3 Corridas por brazo (A directo, B selectivo, R referencia) con orden alternado y caché fría/caliente separadas; repeticiones según 0.1.
- [ ] 5.4 Registrar por coordinador y worker: entrada sin caché, cacheada, salida, reintentos, tiempo de pared y humano; API y cuota por separado; ausencias declaradas.
- [ ] 5.5 Calcular tasa de acierto de la regla D1 contra la señal esperada de cada fixture.

## 6. Decisión

- [ ] 6.1 Informe de decisión con datos crudos, muestra, incertidumbre y una de tres salidas (adoptar, mantener con ajustes, repetir).
- [ ] 6.2 Si se adopta: cambiar la política predeterminada en los assets y en la descripción del catálogo, validando con `catalog.Load()`; si no, registrar la decisión y dejar el módulo opt-in.
- [ ] 6.3 Entregar a 09 el flujo y los datos del piloto.
