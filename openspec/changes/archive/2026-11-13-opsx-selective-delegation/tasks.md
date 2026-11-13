# Tasks: opsx-selective-delegation

> Estado al archivar: ver «Alcance de 07 al archivar y deuda aceptada» (9 entregadas, 16 en deuda). Gobernanza MEDIO: implementar con checkpoints.
> Prerrequisitos de implementación: 05 (PR #10) y 06 (PR #12) integradas en `main`; skill de invocación explícita de 03/04 ya integrada (`d5e95bd`); protocolo y arnés de 02 (`240e1b2`) disponibles para el grupo 5.


## Alcance de 07 al archivar y deuda aceptada

07 se archiva con **alcance parcial** (13/11), igual que 03, 04 y 06. Lo que depende de otros
changes, de acuerdos del grupo o de mediciones pasa a deuda aceptada y no se marca como hecho.

**25 tareas: 9 entregadas y 16 pendientes (deuda aceptada).** Entregadas: 1.1, 2.1 a 2.5 y 3.1 a
3.3. Pendientes: 0.1, 0.2, 1.2, 4.1 a 4.4, 5.1 a 5.6 y 6.1 a 6.3. El detalle por tarea, con
proveedor, artefacto faltante, criterio de cierre y comportamiento mientras tanto, está en
[informes/deuda-aceptada-07-2026-11-13.md](informes/deuda-aceptada-07-2026-11-13.md). La
verificación con comandos y salidas reales está en
[informes/verificacion-07-2026-11-13.md](informes/verificacion-07-2026-11-13.md).

- **Suite Go completa: no está en verde.** Las 6 pruebas que fallan, fallan igual en `main` limpio (D07-BASE).
- **Política predeterminada sin adoptar.** Falta el piloto (D07-05); la spec sincronizada no es un acuerdo del grupo.
- **Límite legacy** (D8, D07-LEG).

## 0. Decisiones previas (grupo)

- [ ] 0.1 Cerrar las decisiones pendientes de `design.md` (señales, métrica objetivo, toggle en `lite`, dueño de `sdd-orchestrator.md`) y volcarlas a los artefactos.
- [ ] 0.2 Confirmar con 05 el nombre del disparador "la acción requiere coordinación de subagentes"; con 06 el módulo `evidence` (06 D5, adoptado por 07 si 03/04 lo aceptan; sigue sin acuerdo) y quién evalúa la suficiencia (06, Open Questions 3). Ya resuelto con 06 (`3c7d908`): texto del `## Evidence Contract`, huella de la partida, directorio ajeno y regla de reejecución, con acuse en `informes/acuse-a-06-2026-11-13.md`; con 08 que sume "escritura fuera del conjunto de escritura del encargo" a su inventario de acciones a bloquear (hasta entonces queda declarada como no controlada preventivamente); 03 y 04 ya referencian el sobre de D6 (03 D12; 04 `assets/codex/native/delegation.md` en `main`); con el grupo, la profundidad de D10.

## 1. Inventario reproducible

- [x] 1.1 Listar con `Compose` (Codex y legacy) y `ComposeLayers` (Claude, capas de 03) cada obligación de delegar dentro y fuera de los markers `sdd-delegation`, con y sin el toggle. Aceptación: tabla con línea, texto, capa y si sobrevive al toggle. Borrador sobre goldens en `entrega-delegation.md` §1; falta confirmarlo componiendo sin el toggle. *Hecho 13/11: `informes/inventario-1.1-1.2-2026-11-13.md` §1.1, compuesto con y sin el toggle. Hallazgo: el límite legacy de D8 también alcanza a Claude < 2.1.110 o sin versión detectable y a Codex en instalación de proyecto.*
- [ ] 1.2 Documentar por cliente, versión e interfaz cómo se lanza, espera y cierra un worker y qué consumo expone por worker. Aceptación: matriz con fuente y versión probada; lo no observado se marca SUPOSICIÓN. *Parcial 13/11: matriz en `informes/inventario-1.1-1.2-2026-11-13.md` §1.2. Falta una corrida observada de Claude (consumo por worker) y Codex completo (no instalado; depende de P2/C05/C06 de 04).*

## 2. Semántica del toggle (TDD)

- [x] 2.1 RED: golden en `compose_test.go` y `compose_layers_test.go` que falla si el bloque (Codex) o la skill y los módulos (Claude) sin `delegation` contienen `ALWAYS delegate`, `Delegate, don't inflate` o la tabla. *Hecho: `TestSelectiveDelegation_NoObligationToDelegate` (Claude por capas y paquete nativo de Codex, con y sin toggle). El bloque legacy de `Compose` queda idéntico byte a byte por el invariante 4.2 de 03; ver nota de alcance abajo.*
- [x] 2.2 RED: test que falla si el bloque con `delegation` contiene el umbral "1-3 / 4+" o no contiene la regla con independencia obligatoria primero. *Hecho: `TestSelectiveDelegation_RuleOrderWithToggle` y `_RuleWithoutToggle`.*
- [x] 2.3 GREEN: reemplazar `modules/delegation.md` (hoy la tabla de umbrales), condicionar `coordinator/principle-delegate` al toggle y dejar en `SKILL.md` la regla de decisión y la declaración sin subagentes. Partir del texto de `entrega-delegation.md` §2–§3. Depende de 1.1 y, para Codex, de 04. *Hecho: `assets/opsx/delegation-rule.md`, `delegation-rule-disabled.md` y `delegation-module.md`, compartidos por Claude (`compose_layers.go`) y Codex (`codex_package.go`); se elimina `assets/codex/native/delegation.md`. Goldens de capas regenerados y revisados.*
- [x] 2.5 RED/GREEN: la regla de profundidad de D10 está en el módulo y un encargo de worker no incluye instrucciones de lanzar subagentes. *Hecho: sección Depth del módulo y plantilla de lanzamiento con "Do not launch sub-agents." (`TestSelectiveDelegation_LaunchTemplate*`).*
- [x] 2.4 Regresión: `go test ./internal/harness/config/... ./internal/filemerge/... ./internal/uninstall/...` sin fallos nuevos; el purge de markers sigue reconociendo `sdd-delegation`. *Hecho 13/11: `internal/harness/config/selective_delegation_test.go`; RED observado entre 13:20 y 13:24 (fallos de aserción, no de compilación), GREEN y regresión sin fallos nuevos frente a la línea base 13:20:49 (3 fallos previos de Windows: symlink, permisos, migración).*

## 3. Contenido del módulo (TDD sobre texto compuesto)

- [x] 3.1 RED: test de presencia de los nueve campos del encargo, el contrato de retorno, la regla de partida y propiedad, y el tope de dos workers con escritura, en el módulo de cada cliente. *Hecho: `TestSelectiveDelegation_ModuleContent`.*
- [x] 3.2 GREEN: redactar regla de decisión (skill), encargo, partida, propiedad, archivos calientes, paralelismo, retorno, veredicto y reintentos (módulo). *Hecho: texto de `entrega-delegation.md` §2–§3 en los assets compartidos, con la regla de reejecución de 06.*
- [x] 3.3 RED/GREEN: golden de paridad que compara la sección de delegación de Claude y Codex salvo la sintaxis de lanzamiento declarada. *Hecho por construcción (mismo asset) y probado: `TestSelectiveDelegation_ClaudeCodexParity`.*

> **Nota de alcance de la sección 2 (13/11).** El bloque único legacy (`Compose`, usado por agentes sin capas y por `primary-agent`) no cambia: 03 fija que sea idéntico byte a byte a su golden (`TestComposeLegacy_ByteIdenticalToGolden`). Por eso Claude legacy y `assets/codex/sdd-orchestrator.md` conservan las obligaciones de delegar. **Decisión de Neyen (13/11): límite declarado.** No se toca el golden de 03; ver `design.md`, D8.

## 4. Batería de veredictos

- [ ] 4.1 Fixtures de retorno: completo corroborado (reejecución con captura independiente), completo sin corroborar, reejecutado por el coordinador sin captura, prosa libre, partida distinta, ajeno en la partida, cambio confirmado fuera de alcance, renombre fuera de alcance, archivo nuevo fuera de alcance, `modificados` incompleto, tope agotado.
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
- [ ] 6.3 Entregar a 09 la configuración de control A0 que corresponde a la salida de 6.1 según la tabla "Qué recibe 09" de `design.md` (D9), con su revisión identificada y los datos del piloto.
