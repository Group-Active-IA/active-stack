# Verificación de 06 (Fase A) antes de archivar

13 de noviembre de 2026. Commit verificado: `ac87dd74201294a0bd31edc3f05d56ea172fe17c` (rama `opsx-verifiable-evidence`, igual a `origin/propose/opsx-verifiable-evidence`), sobre `origin/main` `d5e95bd81d3bbab2d54a39d1c07d1256954b5384` (sin cambios al hacer `git fetch origin '+refs/heads/*:refs/remotes/origin/*'`, por lo que no hubo que integrar nada). Entorno: Windows 11, Git Bash. Los commits posteriores de este cierre solo tocan documentos de este change; el código de `eval/evidence/` es el del commit verificado.

Salidas completas en `C:\Proyectos\informes-opsx\06-opsx-verifiable-evidence\` (archivos `v10-*.txt`).

## 1. Comandos y resultados

| Comando | Resultado real | Salida |
|---|---|---|
| `go build ./...` | exit 0 | `v10-go-build.txt` |
| `go vet ./...` | exit 0 | `v10-go-vet.txt` |
| `go test -count=1 -v ./eval/evidence` | exit 0; `ok ... eval/evidence 3.586s`; 11 pruebas de primer nivel y 25 subpruebas en PASS, 0 FAIL | `v10-go-test-evidence.txt` |
| `go test -count=1 ./...` | **exit 1**: 35 paquetes `ok` y 4 `FAIL` (`internal/backup`, `internal/harness/config`, `internal/install`, `internal/uninstall`) | `v10-go-test-all.txt` |
| `openspec validate opsx-verifiable-evidence` | «Change 'opsx-verifiable-evidence' is valid», con 6 advertencias de requisitos largos (>500 caracteres) en `execution-receipts/spec.md` | `v10-validate-pre.txt` |
| `git diff origin/main --stat -- internal/backup internal/filemerge internal/pipeline` | vacío: no se tocaron los paquetes ALTO | (en este informe) |

Los 4 paquetes que fallan son de 04 y **fallan igual en `origin/main` limpio** (`go-test-main-limpio.txt`): permisos de symlink en Windows («A required privilege is not held by the client»), permisos de directorio que se ensanchan a 777 y rollback con edición preservada. No tienen relación con 06 (06 solo agrega archivos en `eval/evidence/` y en este change). **La suite Go completa no está en verde**; por eso la tarea 4.11 queda sin cumplir y pasa a deuda.

## 2. Criterios de aceptación de la ficha (`CHANGES.md:194`) contra su fixture

Criterio de la ficha: «Rechazar recibos inventados o viejos, comandos insuficientes, salidas truncadas sin original y requisitos incumplidos con tests verdes. Un recibo incompleto queda sin verificar.» Cada fixture está en `eval/evidence/testdata/casos/<nombre>.json`; la prueba `TestCases_VerdictMatchesContract` compara el veredicto calculado con el campo `esperado` y pasó en los 22 casos.

| Criterio | Fixture | Estado obtenido (integridad / procedencia / suficiencia) | Resultado de la prueba |
|---|---|---|---|
| Recibo inventado | `inventado-huella-no-coincide` | rechazado (fallida / corroborada / suficiente) | PASS |
| Recibo inventado | `inventado-copia-ausente` | rechazado (fallida / corroborada / suficiente) | PASS |
| Recibo inventado | `inventado-codigo-de-salida` | rechazado (aprobada / contradicha / suficiente) | PASS |
| Recibo inventado | `inventado-fuente-contradice` | rechazado (aprobada / contradicha / suficiente) | PASS |
| Recibo viejo | `viejo` | rechazado (aprobada / corroborada / suficiente) | PASS |
| Recibo viejo (mismo HEAD, otro diff) | `mismo-head-otro-diff` | rechazado | PASS |
| Comando insuficiente | `comando-insuficiente` | rechazado (aprobada / corroborada / insuficiente) | PASS |
| Salida truncada sin original | `truncado-sin-original` | rechazado (fallida / no evaluable / no evaluable) | PASS |
| Requisito incumplido con tests verdes | `tests-verdes-requisito-incumplido` | rechazado (aprobada / corroborada / insuficiente) | PASS |
| Recibo incompleto = sin verificar | `incompleto` | sin verificar (no evaluable en los tres controles) | PASS |
| Cero recibos nunca verificado | `cero-recibos` | sin verificar | PASS |
| Recibo falso con huella correcta, sin corroboración | `falso-huella-correcta-sin-corroboracion` | sin verificar (aprobada / sin corroborar / suficiente) | PASS |
| Recibo falso con huella correcta, contradicho | `falso-huella-correcta-reejecucion-contradice` | rechazado (aprobada / contradicha / suficiente) | PASS |
| Reejecución del coordinador sin captura propia | `reejecucion-del-coordinador-sin-captura` | sin verificar (no corrobora) | PASS |
| Reejecución con captura independiente | `reejecucion-con-captura-independiente` | verificado | PASS |
| Detalle descartado por la redacción | `detalle-redactado` | sin verificar (aprobada / corroborada / no evaluable) | PASS |
| Copia vencida | `copia-vencida` | sin verificar (propuesta de partida, (TBD) con 05) | PASS |
| Caso válido | `valido` | verificado | PASS |
| Huella LF frente a CRLF | `copia-crlf-huella-lf` | rechazado (integridad fallida) | PASS |
| Directorio ajeno (propuesta de 06) | `directorio-ajeno`, `directorio-ajeno-separadores-windows` | rechazado | PASS |
| Directorio con separadores de Windows | `directorio-separadores-windows` | verificado | PASS |

Huella del estado probado: `TestStateFingerprint_RealWorktree` (limpio, modificado sin preparar, mismo HEAD con otro diff preparado, archivo nuevo, borrado, CRLF frente a LF, ignorado; `core.autocrlf` en `true`, `false` e `input`), `TestStateFingerprint_SameBaseDifferentDiffs`, `TestStateFingerprint_CleanMatchesMeasuredAndTreeHash` y `TestStateFingerprint_NotARepository`: PASS.

## 3. Qué se verificó y qué no

**Verificado en este commit (Windows):** build, vet y las pruebas de `eval/evidence`; los cinco criterios de la ficha y el default «sin verificar» contra sus fixtures; la huella del estado probado sobre un worktree de Git real; los paquetes ALTO sin cambios; el change válido para `openspec validate`.

**Medido antes y registrado en `informes/medicion-huella-2026-11-13.md`:** la huella en Linux sobre **WSL2** (Ubuntu) con `core.autocrlf` en tres valores, y el recomputo con `sha256sum --text` y `Get-FileHash`. No se repitió en este paso.

**No verificado:**
- **Linux nativo y CI** (no hay ejecución en runner de Linux), **macOS y Termux**.
- **Captura real de 05:** los fixtures son datos; no existe todavía el redactor ni el capturador de 05, así que `4.6` y `4.9` no se pudieron correr con salida real.
- **Assets de Claude y de Codex** (sección 3): no existen; el contrato no está cargado en ningún orquestador y por eso no hay prueba de `apply` directo ni delegado (4.8, 4.10).
- **Suite Go completa:** exit 1 por los 4 paquetes de 04 descritos arriba; no está verificada en verde ni en Windows ni en Linux.
- **Acuerdos del grupo:** nada de lo marcado «propuesta de 06, pendiente de confirmación» fue confirmado (módulo `evidence`, quién evalúa la suficiencia, directorio ajeno, reejecución con captura independiente, TreeHash, D9 y D14 de 05).

## 4. Correcciones de la revisión del PR #12

**Vínculo entre los juicios y la ejecución concreta.** `eval/evidence/judge.go` consultaba la corroboración y la suficiencia solo por id de corrida: con `valido.json`, cambiar `Command` a `echo no-test` conservando la corrida seguía dando «verificado». Ahora:

- La corroboración (`CorroborationRecord`) identifica comando, directorio, commit, huella del estado y código de salida, y una fuente. Si confirma pero describe otra ejecución que la del recibo, el recibo queda **rechazado** (la fuente independiente vio otra cosa). Si no hay registro independiente para esa ejecución, queda **sin verificar**. Un registro sin fuente, o cuya fuente es la de la captura del recibo (el agente), no es independiente y no cuenta.
- El juicio de suficiencia (`SufficiencyRecord`) corresponde a corrida, requisito y comando; si alguno difiere, no aplica y la suficiencia queda «no evaluable» (sin verificar).
- Los 22 fixtures migraron a ese formato (listas en lugar de mapas por id) y siguen con el mismo veredicto.
- Regresiones nuevas en `eval/evidence/evidence_test.go`: recibo original verde como control; cambio de comando con la misma corrida (rechazado, y sin verificar si falta la corroboración); cambio de directorio, commit, huella y resultado; suficiencia por requisito; captura declarada por el agente sin corroboración independiente (sin registro, registro de la misma fuente y registro de otra corrida).

**Validación estricta de OpenSpec.** Los seis requisitos de más de 500 caracteres de `execution-receipts` se acortaron (el detalle pasó a escenarios nuevos; los escenarios existentes se mantienen) en el delta archivado y en `openspec/specs/execution-receipts/spec.md`, que quedan sincronizados salvo los encabezados propios de cada formato. El change ya está archivado: para validar el delta se copió temporalmente a `openspec/changes/opsx-verifiable-evidence/` y se borró la copia.

Resultados de `v11-*.txt` (en `C:\Proyectos\informes-opsx-opsx-verifiable-evidence\`), corridos sobre el commit `ddbe71371a5d54481bde63393901828e4b6cf04a` (árbol limpio salvo este informe y `tasks.md`), en Windows:

| Comando | Resultado real | Salida |
|---|---|---|
| `go build ./...` | exit 0 | `v11-go-build.txt` |
| `go vet ./...` | exit 0 | `v11-go-vet.txt` |
| `go test -count=1 -v ./eval/evidence` | `ok`, PASS, exit 0 | `v11-go-test-evidence.txt` |
| `openspec validate opsx-verifiable-evidence --type change --strict` (delta, OpenSpec 2.0.0) | «Change 'opsx-verifiable-evidence' is valid», exit 0 | `v11-validate-change.txt` |
| `openspec validate execution-receipts --type spec --strict` (spec final, OpenSpec 2.0.0) | «Specification 'execution-receipts' is valid», exit 0 | `v11-validate-spec.txt` |

El delta se validó copiando la carpeta archivada a `openspec/changes/opsx-verifiable-evidence/` y borrando la copia después. Se comprobó que los nueve requisitos de la spec final y del delta archivado son idénticos (solo difieren los encabezados).

Antes del ajuste, la spec final devolvía exit 1 en modo estricto con 6 advertencias de requisitos largos. La CLI instalada aquí es 2.0.0 y la revisión usó 1.16.0; el resultado en 1.16.0 no se repitió.

## 5. Lo que sigue pendiente

El fixture `detalle-redactado` prueba el veredicto sobre una copia **ya redactada**. La recuperación y el vencimiento de la copia real del 05, y la carga efectiva del contrato en Claude y Codex para ejecución directa y delegada, siguen **pendientes**.

**Límite:** un recibo capturado solo por el arnés necesita un registro de otra fuente independiente para quedar verificado; el evaluador no cuenta como independiente a la fuente de la propia captura. A confirmar con 05 D22.
