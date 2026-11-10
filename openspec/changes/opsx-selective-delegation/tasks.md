# Tasks: opsx-selective-delegation

> Ronda `propose`: todas las tareas quedan pendientes. Gobernanza MEDIO: implementar con checkpoints.
> Prerrequisitos de implementación: 05 y 06 aprobadas e integradas; skill de invocación explícita de 03/04 acordada; protocolo y arnés de 02 (`240e1b2`) disponibles para el grupo 5.

## 0. Decisiones previas (grupo)

- [ ] 0.1 Cerrar las decisiones pendientes de `design.md` (señales, métrica objetivo, toggle en `lite`, dueño de `sdd-orchestrator.md`) y volcarlas a los artefactos.
- [ ] 0.2 Confirmar con 05 el nombre del disparador "la acción requiere coordinación de subagentes"; con 06 el texto del extracto `## Evidence Contract` y la huella del estado de trabajo que reutiliza la partida; con 08 qué escrituras fuera de alcance tienen control preventivo obligatorio.

## 1. Inventario reproducible

- [ ] 1.1 Listar con `Compose` cada obligación de delegar dentro y fuera de los markers `sdd-delegation`, para Claude y Codex, con y sin el toggle. Aceptación: tabla con línea, texto y si sobrevive al toggle.
- [ ] 1.2 Documentar por cliente, versión e interfaz cómo se lanza, espera y cierra un worker y qué consumo expone por worker. Aceptación: matriz con fuente y versión probada; lo no observado se marca SUPOSICIÓN.

## 2. Semántica del toggle (TDD)

- [ ] 2.1 RED: golden en `compose_test.go` que falla si el bloque sin `delegation` contiene `ALWAYS delegate`, `Delegate, don't inflate` o la tabla, para Claude y Codex.
- [ ] 2.2 RED: test que falla si el bloque con `delegation` contiene el umbral "1-3 / 4+" o no contiene la regla con independencia obligatoria primero.
- [ ] 2.3 GREEN: mover las obligaciones al módulo y dejar en la skill la regla de decisión y la declaración sin subagentes. Depende de 1.1 y de la skill de 03/04.
- [ ] 2.4 Regresión: `go test ./internal/harness/config/... ./internal/filemerge/... ./internal/uninstall/...` sin fallos nuevos; el purge de markers sigue reconociendo `sdd-delegation`.

## 3. Contenido del módulo (TDD sobre texto compuesto)

- [ ] 3.1 RED: test de presencia de los nueve campos del encargo, el contrato de retorno, la regla de partida y propiedad, y el tope de dos workers con escritura, en el módulo de cada cliente.
- [ ] 3.2 GREEN: redactar regla de decisión (skill), encargo, partida, propiedad, archivos calientes, paralelismo, retorno, veredicto y reintentos (módulo).
- [ ] 3.3 RED/GREEN: golden de paridad que compara la sección de delegación de Claude y Codex salvo la sintaxis de lanzamiento declarada.

## 4. Batería de veredictos

- [ ] 4.1 Fixtures de retorno: completo corroborado, completo sin corroborar, prosa libre, partida distinta, ajeno en la partida, cambio confirmado fuera de alcance, renombre fuera de alcance, archivo nuevo fuera de alcance, `modificados` incompleto, tope agotado.
- [ ] 4.2 Fixtures de decisión: verificación breve con independencia exigida, independencia exigida con `delegation` apagado, módulo ausente, contención por archivo caliente, tercer worker con escritura.
- [ ] 4.3 Fixture de combinación: dos retornos válidos por separado que rompen un escenario al combinarse.
- [ ] 4.4 Correr 4.1–4.3 con el flujo de Claude y el de Codex. Aceptación: mismo veredicto por fixture; diferencias de cliente declaradas como límite.

## 5. Piloto (depende de 02, 05 y 06)

- [ ] 5.1 Usar el arnés de 02; verificar que `cmd/` e `internal/` no cambian.
- [ ] 5.2 Fixtures T1–T8 sobre un commit fijo, cada uno con aceptación y paso de decisión esperado; T2 reutiliza los fixtures de salida verbosa de 05.
- [ ] 5.3 Calibración: tres corridas por celda (brazo × tarea × cliente × interfaz), orden alternado y caché registrada según 02.
- [ ] 5.4 Con la calibración, registrar en grupo tamaño de muestra, dispersión tolerable, umbral de mejora, margen de calidad y regla de inconclusos. Sin este registro, 5.5 no habilita adopción.
- [ ] 5.5 Prueba de adopción con los parámetros de 5.4. Excluir del contraste directo los casos con independencia exigida que un brazo no pueda proveer.
- [ ] 5.6 Reportar costo por tarea aceptada (total de cohorte ÷ aceptadas; sin cociente con cero aceptadas), cuota, tokens, latencia, ahorro de costo y reducción de espera por separado, y tasa de acierto de D1.

## 6. Decisión

- [ ] 6.1 Informe de decisión con datos crudos, muestra, dispersión y una de tres salidas (adoptar, mantener con ajustes, repetir).
- [ ] 6.2 Si se adopta: cambiar la política predeterminada en los assets y en la descripción del catálogo, validando con `catalog.Load()`; si no, registrar la decisión y dejar el módulo opt-in.
- [ ] 6.3 Entregar a 09 la revisión de 07 usada como control A0 y los datos del piloto.
