# Verificación de 07 antes de archivar

13 de noviembre de 2026. Commit verificado: `91e3fa5dd31859ad9d698647d30c5767bdc2c11d` (rama
`propose/opsx-selective-delegation`, igual a su remoto), sobre `origin/main`
`d5e95bd81d3bbab2d54a39d1c07d1256954b5384` (0 commits de `main` sin integrar tras `git fetch`).
Entorno: Windows 11, Git Bash, Claude Code 2.1.323. Los commits posteriores del cierre sólo tocan
documentos de este change. Por la regla del operador **no se ejecutó build**.

## 1. Comandos y resultados (13:39–13:42, -03:00)

| Comando | Resultado real |
|---|---|
| `go vet ./internal/harness/config/...` | exit 0 |
| `go test -count=1 -v ./internal/harness/config/ -run SelectiveDelegation` | exit 0; 7 pruebas en PASS, 0 FAIL |
| `go test -count=1 ./internal/harness/config/... ./internal/filemerge/... ./internal/uninstall/...` | exit 1; `filemerge` y `config/permissions` ok; fallan `TestManagedInstallRemovePreserveParentPermissions`, `TestManagedOperationSnapshotSymlinkIsRejected` y `TestManagedSkillRollbackPreservesInstructionBytes`, los mismos 3 de la línea base tomada antes de tocar código (13:20:49) |
| `go test -count=1 ./...` | **exit 1**: 34 paquetes `ok` y 4 `FAIL` (`internal/backup`, `internal/harness/config`, `internal/install`, `internal/uninstall`), con 6 pruebas fallidas |
| Mismos 4 paquetes sobre `main` limpio (`C:\Users\LENOVO\active-stack`, `d5e95bd`) | exit 1 con **las mismas 6 pruebas fallidas** (comparación con `diff` de los nombres: idénticas) |
| `openspec validate opsx-selective-delegation --strict` | «Change 'opsx-selective-delegation' is valid» |
| `git diff origin/main --stat -- internal/backup internal/filemerge internal/pipeline` | vacío: no se tocaron los paquetes ALTO |

Las 6 pruebas fallidas son permisos de symlink en Windows («A required privilege is not held by the
client»), directorios padre que se ensanchan a 777 y rollbacks con edición preservada. Fallan igual
en `main` y son de 04 (D04-BASE01). **La suite Go completa no está en verde**, y 07 no la declara verde.

## 2. Criterios de las tareas entregadas contra su prueba

| Tarea | Criterio | Prueba | Resultado |
|---|---|---|---|
| 1.1 | Tabla de obligaciones por ruta, con y sin el toggle | test temporal de composición, 13:33:56 (`inventario-1.1-1.2-2026-11-13.md`) | tabla registrada |
| 2.1 | Sin obligación de delegar en Claude por capas ni en Codex nativo, con y sin el toggle | `TestSelectiveDelegation_NoObligationToDelegate` | PASS |
| 2.2 | Regla en orden con independencia primero; sin el toggle, trabajo directo y verify detenido | `TestSelectiveDelegation_RuleOrderWithToggle`, `_RuleWithoutToggle` | PASS |
| 2.3 | Módulo reemplazado; `principle-delegate` y «ALWAYS delegate» fuera de la skill | goldens `compose_layers_skill` y `compose_layers_module_delegation` (diff revisado) | PASS |
| 2.4 | Regresión sin fallos nuevos; purge de markers intacto | comparación con la línea base y con `main` | sin fallos nuevos |
| 2.5 | Profundidad en el módulo; el encargo no pide lanzar subagentes | `TestSelectiveDelegation_LaunchTemplate`, `_LaunchTemplateWithoutToggle` | PASS |
| 3.1 | Nueve campos, sobre, partida, propiedad, tope de dos workers | `TestSelectiveDelegation_ModuleContent` | PASS |
| 3.2 | Texto del módulo y de la regla | `assets/opsx/delegation-*.md` | entregado |
| 3.3 | Paridad entre Claude y Codex | `TestSelectiveDelegation_ClaudeCodexParity` (mismo asset) | PASS |

RED observado entre 13:20 y 13:24, antes del GREEN, con fallos de aserción y no de compilación.

## 3. Límite declarado

El bloque legacy (`Compose`) conserva las obligaciones de delegar, por decisión de Neyen (D8). Lo
reciben OpenCode, Antigravity, Codex en instalación de proyecto y Claude Code anterior a 2.1.110 o
sin versión detectable (`inventario-1.1-1.2-2026-11-13.md`).
