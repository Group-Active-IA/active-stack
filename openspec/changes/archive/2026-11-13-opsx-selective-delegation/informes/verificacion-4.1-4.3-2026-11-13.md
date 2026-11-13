# Verificación de la batería de veredictos de 07 (tareas 4.1 a 4.3)

13 de noviembre de 2026. Responsable: Neyen. Pedido del revisor: actualizar la rama desde `main` y usar
el contrato de 06 (un recibo por comando ejecutado), con los casos de `eval/evidence` como referencia.

## Base

| Qué | Valor |
|---|---|
| Rama | `propose/opsx-selective-delegation` (PR #13), `4922498` + merge de `origin/main` `0005a6f` |
| 05 | integrada en `main` `b27b8e0` (PR #10) |
| 06 | integrada en `main` `0005a6f` (PR #12), última corrección `d3ab8b3` |
| Merge | sin conflictos; `go test ./internal/harness/config/... ./eval/...` sin fallos nuevos frente a `main` limpio |

## Qué se entregó

- `eval/delegation/`: paquete externo al binario (no lo importan `cmd/` ni `internal/`), como `eval/evidence`.
  - `envelope.go`: sobre de D6 y `ParseEnvelope` (prosa o campo faltante → error → *sin verificar*).
  - `judge.go`: `JudgeReturn`, que valida en el orden de D6: sobre, partida, propiedad, recibos y criterios.
    **Los recibos no se reimplementan**: cada requisito del encargo se juzga con `evidence.EvaluateRequirement` de 06.
  - `decide.go`: `Decide`, con la regla de `assets/opsx/delegation-rule.md` paso a paso; `HotFiles`.
  - `combine.go`: `JudgeCombination`, para el resultado combinado (D4, paso 4).
  - `testdata/{retornos,decisiones,combinacion}`: 13 + 9 + 4 fixtures. Las copias redactadas son
    bytes de `eval/evidence/testdata/copias` (huellas sha256 iguales).
- Módulo instalado (`assets/opsx/delegation-module.md`) y golden de capas:
  - Archivo caliente corregido a `internal/harness/config/assets/*/sdd-orchestrator.md` (ver hallazgo).
  - Regla de 06 final: la corroboración nombra comando, árbol y resultado; una captura declarada por el
    propio worker no es independiente.
- Tests nuevos en `internal/harness/config`: `TestSelectiveDelegation_HotFilesExist` y dos frases
  más en `TestSelectiveDelegation_ModuleContent`. En `eval/delegation`, `TestHotFiles_MatchTheModule`
  impide que la lista del eval diverja de la del módulo.

## EVIDENCIA

| Hora | Comando | Resultado |
|---|---|---|
| 20:31:15 | `go test ./eval/delegation/` (con stubs) | RED: 25 subtests fallan por aserción (`estado = "", esperado "verificado"`); compila y `go vet` pasa |
| 20:33:05 | ídem, implementado | `ok` |
| — | mutaciones sobre `eval/delegation` | primera tanda: 2 mutantes sobreviven (renombre desde afuera, patrón caliente); se agregan `renombre-desde-fuera-de-alcance` y `archivo-caliente-patron`, y el segundo revela el hallazgo de abajo |
| 20:34:31 | `go test ./internal/harness/config/ -run TestSelectiveDelegation_HotFilesExist` | RED: `hot file "assets/*/sdd-orchestrator.md" matches no file in the repository` |
| 20:34:55 | `go test ./eval/delegation/ -run HotFiles` tras corregir el módulo | RED: `HotFiles` difiere de la lista del módulo; GREEN al alinear |
| 20:36:29 | `go test ./internal/harness/config/ -run TestSelectiveDelegation_ModuleContent` | RED: faltan las dos frases de 06 final |
| 20:38:03 | `go test -count=1 ./internal/harness/config/` | sólo `TestManagedOperationSnapshotSymlinkIsRejected` y `TestManagedInstallRemovePreserveParentPermissions`, que fallan igual en `main` limpio |
| — | `go test ./internal/uninstall/` en `main` limpio y en 07 | `TestManagedSkillRollbackPreservesInstructionBytes` falla en ambos (D07-BASE) |
| — | segunda tanda de mutaciones (renombre desde afuera, patrón caliente, tope de dos workers, `modified` incompleto, fallo combinado) | los 5 mutantes mueren, cada uno en su fixture |

## Hallazgo corregido

El módulo, la spec y D5 declaraban como caliente `assets/*/sdd-orchestrator.md`. Esa ruta no existe en
el repo: los ocho `sdd-orchestrator.md` están en `internal/harness/config/assets/<cliente>/`. La
protección no protegía nada. Decisión de Neyen (13/11): corregir la ruta en el módulo, la spec
(archivo y `openspec/specs`), `design.md` D5 y `entrega-delegation.md`.

## JUICIO (la spec no lo fija; decisión de la implementación del eval, revisable)

- **Acción después del veredicto.** `integrar` si queda verificado; `corroborar` si los recibos
  quedan *sin verificar*; `pedir sobre o reintentar` si falta el sobre; `reintentar` en los demás
  casos. Con el tercer intento agotado, siempre `escalar` con el historial.
- **Tercer worker con escritura** se registra en el paso 2 con la señal «tope de dos workers con escritura».
- **Archivo caliente por patrón.** Dos encargos que tocan el mismo patrón caliente (por ejemplo,
  `claude/` y `codex/sdd-orchestrator.md`) entran en conflicto aunque no compartan ruta: «never split these».
- **Delegation apagado** sin independencia exigida: directo, registrado como paso 5 «delegation desactivado».
- **Criterios de salida.** Sin comprobar → *sin verificar*; incumplidos → *rechazado*. Un `status`
  distinto de `completed` → *rechazado*.
- **Combinación.** Un escenario combinado falla si su recibo queda verificado con código de salida
  distinto de cero. Ahí el fallo se registra contra todos los encargos combinados. Si el recibo se
  rechaza por otro motivo (por ejemplo, inventado), no se imputa a los encargos.
- **`modified`** debe coincidir exactamente con lo atribuible: una omisión o una ruta de más dejan *sin verificar*.

## SUPOSICIÓN y límites

- Los fixtures son sintéticos: SHAs ilustrativos y copias de 06. **No hay copia redactada real de 05**
  ni corridas con clientes reales.
- Los recibos usan el JSON de `evidence.Receipt` (claves en español). El módulo en inglés los llama
  `run`, `tree`, `command`, `result`, `requirements` y `output`; no hay capa de traducción.
- La paridad Claude/Codex sigue siendo por construcción (mismo asset, `TestSelectiveDelegation_ClaudeCodexParity`).
  Los veredictos con cada cliente son la tarea 4.4.

## Lo que queda (seguimiento pedido por el revisor)

- 4.4: conectar la copia redactada real de 05 y probar el flujo directo y el delegado en Claude y en
  Codex, con el mismo veredicto por fixture. Sigue en la deuda D07-04b.
