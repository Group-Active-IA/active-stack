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
