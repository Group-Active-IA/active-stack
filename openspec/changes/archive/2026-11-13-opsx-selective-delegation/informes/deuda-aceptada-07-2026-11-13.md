# Deuda aceptada de 07 al archivar `opsx-selective-delegation`

13 de noviembre de 2026. Responsable de 07: Neyen. Sigue la convención de 03, 04 y 06: lo que
depende de otros changes, de acuerdos del grupo o de mediciones que 07 no puede hacer sola sale del
checklist activo. Queda registrado con su criterio de cierre y **no se marca como hecho**.

Verificado sobre `91e3fa5` (`informes/verificacion-07-2026-11-13.md`). Base: `origin/main` `d5e95bd`.
06 se consulta en `origin/propose/opsx-verifiable-evidence` `3c7d908` (PR #12); 05 en
`origin/propose/opsx-task-context` `4337e36` (PR #10).

## Qué significa archivar con decisiones pendientes

Al archivar, la spec `selective-delegation` se sincroniza a `openspec/specs/`. Las decisiones que
`design.md` lista como **pendientes de grupo** (señales, métrica de adopción, toggle en `lite`, dueño
de `sdd-orchestrator.md`, profundidad de D10) siguen pendientes: la spec describe la política
propuesta, no un acuerdo. **La política predeterminada no cambia por archivar**: «directo por
defecto» se adopta sólo con el piloto (sección 5) y la decisión de 6.1.

## Lo entregado y verificado (no es deuda)

Tareas 1.1, 2.1 a 2.5, 3.1 a 3.3 y, tras integrarse 05 y 06 en `main` (`0005a6f`), 4.1 a 4.3 (`eval/delegation/`, `informes/verificacion-4.1-4.3-2026-11-13.md`). Están en `assets/opsx/delegation-rule.md`,
`delegation-rule-disabled.md` y `delegation-module.md`, `compose_layers.go`, `codex_package.go`,
`selective_delegation_test.go` y los goldens de capas. Inventario en
`informes/inventario-1.1-1.2-2026-11-13.md`. El acuse a 06 está en `informes/acuse-a-06-2026-11-13.md`.

## Registro de deuda

**Estado de todas las filas:** deuda aceptada por Neyen el 2026-11-13 para cerrar el circuito y que
el grupo siga con 08. Ninguna fila se cierra por existir una propuesta o un informe.

| ID | Tarea | Proveedor / dueño | Artefacto o decisión faltante | Criterio de cierre | Comportamiento mientras tanto |
|---|---|---|---|---|---|
| D07-0001 | 0.1 | Grupo | Decisiones de `design.md`: señales S1–S4 y umbral de 300 líneas, métrica objetivo de adopción, toggle `delegation` en `lite`, dueño de `sdd-orchestrator.md` | Decisión registrada por tema y artefactos ajustados | Rige la propuesta de 07: S1–S4 con orden, 300 líneas como supuesto, toggle en `lite` sin cambios |
| D07-0002a | 0.2 | Juani (03) y Matías (04) | Acuerdo sobre el módulo propio `evidence` (06 D5, D06-0101) | Respuesta registrada y matriz fila 28 actualizada | El módulo de 07 dice que, sin contrato instalado, todo recibo queda *sin verificar*. Si se rechaza, el extracto vuelve a `delegation-module.md` con el límite de D8 |
| D07-0002b | 0.2 | Gianfranco (06), grupo | Quién evalúa la suficiencia (06 Open Questions 3, D06-0102 y D06-0502) | Decisión registrada | 07 no fija qué evidencia esperada puede exigir el encargo más allá de los criterios de salida |
| D07-0002c | 0.2 | Gianfranco (08) | Sumar «escritura fuera del conjunto de escritura del encargo» al inventario de acciones a bloquear de 08 | Fila incorporada en 08 | Sólo hay comprobación posterior (D4); declarada no controlada preventivamente |
| D07-0002d | 0.2, D10 | Grupo; 03, 04 y 09 afectados | Profundidad de delegación (ninguna / un nivel de sólo lectura / tope de 03) | Decisión registrada | El módulo dice «Only you launch workers» y la plantilla de Claude dice «Do not launch sub-agents» (propuesta de 07) |
| D07-0002e | 0.2 | Emiliano (05) | Nombre del disparador «la acción requiere coordinación de subagentes» | Confirmación de 05 | Se usa la fila `delegate` de la tabla acción → módulos de 03 y 04 |
| D07-0102 | 1.2 | Neyen; Matías (04) para Codex | Matriz por cliente, versión e interfaz: lanzar, esperar, cerrar y consumo por worker. Claude: falta una corrida observada. Codex: no instalado; depende de P2 y de C05/C06 de 04 | Corrida observada de Claude terminal registrada, y celdas de Codex de 04 o corrida propia en Windows/Linux | Matriz parcial con SUPOSICIÓN marcada en `inventario-1.1-1.2-2026-11-13.md` §1.2. El piloto no puede contabilizar consumo por worker |
| D07-04b | 4.4 | Neyen | Correr la batería de `eval/delegation` con el flujo de Claude y el de Codex, directo y delegado, sobre la copia redactada real de 05. Las tareas 4.1 a 4.3 se cerraron el 13/11 tras integrarse 06 (`informes/verificacion-4.1-4.3-2026-11-13.md`) | Mismo veredicto por fixture en ambos clientes, con la copia de 05 conectada; las diferencias de cliente se declaran como límite | Los veredictos se prueban sobre fixtures sintéticos (`go test ./eval/delegation`); no hay corrida con clientes reales |
| D07-05 | 5.1 a 5.6 | Neyen, tras 02, 05 (PR #10) y 06 (PR #12) en `main` | Piloto: arnés de 02, fixtures T1–T8, calibración, parámetros de adopción, prueba de adopción y reporte de costo | Las seis tareas con sus datos crudos | Sin datos de adopción. La política nueva está instalada sólo en las rutas de 07 y no se declara mejor que la anterior |
| D07-06 | 6.1 a 6.3 | Neyen, tras D07-05 | Informe de decisión, cambio o no de la política predeterminada (validado con `catalog.Load()`), entrega del control A0 a 09 | Informe con una de tres salidas y A0 entregado a 09 | 09 no tiene A0 fijado: según `design.md` D9, «repetir» implica que 09 sólo puede calibrar con brazo declarado |
| D07-LEG | D8 (límite) | Neyen con Juani (03) | Regla de 07 en el bloque legacy (`Compose`) | Seguimiento acordado con 03 que cambie su golden | OpenCode, Antigravity, Codex de proyecto y Claude < 2.1.110 o sin versión detectable conservan «ALWAYS delegate» y la tabla 1-3 / 4+ |
| D07-BASE | 2.4 (suite completa) | 04 (D04-BASE01) | `go test ./...` en verde: 6 pruebas de `backup`, `harness/config`, `install` y `uninstall` fallan, igual que en `main` limpio | Disposición de D04-BASE01 y suite en verde en Windows y Linux | 07 no agrega fallos (comparado con `main`). La suite completa no se declara verde |
| D07-PR | cierre | Neyen; revisión de Juan Cruz | PR del change archivado y revisión «contrato y evidencia sobre el último commit» | PR abierto, revisado y mergeado | El change queda archivado en la rama |
