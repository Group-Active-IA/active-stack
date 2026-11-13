# propose-2 · planner · codex-native-opsx

- **Autor / rol**: planner, subagente nativo `role-planner` (Claude). Modelo declarado por el entorno: claude-opus-5-5. Es la misma familia que el plan-reviewer.
- **Encargo**: corregir solo los hallazgos de `informes/propose-1-plan-reviewer.md` (B1, B2, M1–M8, m1–m13), sin decidir puntos de grupo.
- **Revisión**: HEAD `1b4639b`. Las citas de código son de `bd63543`, y `internal/` y `cmd/` son idénticos entre ambas. Los artefactos siguen sin trackear.
- **Artefactos modificados** (los cinco reescritos, en `openspec/changes/codex-native-opsx/`):
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `specs/codex-opsx-delivery/spec.md`
  - `specs/codex-sdd-migration/spec.md`
- **Sin cambios**: `propose-1-planner.md`, `LEEME.md`, código, `harnesses.yaml`. No hubo build ni commit. `## Aprobación` sigue vacía.

## EVIDENCIA

| # | Timestamp (UTC) | Comando | Salida relevante |
|---|---|---|---|
| P1 | 2026-11-06 | Lectura de `cmd/active-stack/tui_wiring.go:48-52`, `internal/tui/backups_screen.go:32`, `internal/backup/retention.go:131,175-181` y `cmd/active-stack/headless/windows_backups.go:15-24,84-95`; grep de `Prune(` en código que no es test | La TUI lista todo subdirectorio de `~/.active-stack/backups` que tenga manifiesto (`ListManifests`). La GUI de Windows solo lista los slots `install` y `uninstall`. `Prune` no tiene llamadores en producción, así que no hay retención automática. Esta evidencia sostiene la recomendación G12-a. |
| P2 | 2026-11-06T13:42:02Z | `openspec validate codex-native-opsx --strict` | Exit 1: `[WARNING] codex-opsx-delivery … "La verificación post-instalación…": Requirement text is very long (>500 characters)`. |
| P3 | 2026-11-06T13:42:16Z | `openspec validate codex-native-opsx --strict`, tras acortar la descripción de ese requisito | `Change 'codex-native-opsx' is valid`, exit 0. |
| P4 | 2026-11-06T13:42:16Z | `openspec status --change codex-native-opsx --json` | `isPlanningComplete True`, `isComplete True`; proposal, specs, design y tasks en `done`. |
| P5 | 2026-11-06T13:42:02Z | `openspec list --json`; `git status --short` | `totalTasks: 37`, `completedTasks: 0`, `in-progress`; `?? openspec/changes/codex-native-opsx/`. |

No volví a correr las sondas de OpenSpec. Reutilicé las del intento 1 (E4) y la R3 del revisor, porque los archivos que cubren no cambiaron.

## Hallazgo → cambio aplicado

| Hallazgo | Disposición | Dónde |
|---|---|---|
| **B1** reversión que no sobrevive | Corregido, condicionado a la nueva decisión **G12**, que no tomo. Opciones: (a) snapshot de migración dedicado con sello de tiempo; (b) límite documentado. Recomiendo (a) con la evidencia P1, R6 y R7. Queda como sub-decisión si se cubre la GUI de Windows. Reemplacé el requisito de reversión por dos requisitos excluyentes, con escenarios de update con dos install y de reinstalar tras migrar. Cada rama tiene su tarea con RED y su mutante. | design D12 + tabla G (G12) + Migration Plan §3; spec de migración, requisitos "(G12-a)" y "(G12-b)"; tasks 5.4, 5.5, 8.2, 9.2(k), 10.3, 11.2; proposal "What Changes" |
| **B2** decisiones y aprobación abiertas | Hecho explícito, sin resolver. Puerta: G1–G13 y D6 se cierran en grupo antes de compartir la propuesta como final y antes de `apply`. `## Aprobación` sigue vacía, con un placeholder que ahora nombra G1–G13 y D6. | proposal (estado y Aprobación); design (puerta B2); tasks (encabezado); notas de borrador en ambas specs |
| **M1** `jr-stack:` coexistente | Corregido. Retiro en bucle de todas las ocurrencias del ID en el paquete config, sin tocar `filemerge`. Verify detecta los heredados. Agregué escenario y mutante. | design D3, D10; spec de migración ("Markers actual y heredado coexistentes"); spec de delivery ("Bloque heredado remanente"); tasks 5.2, 7.1, 9.2(h) |
| **M2** alcance del chequeo de firma | Corregido. El alcance es el bloque del ID bajo cualquier prefijo, no el archivo. Agregué el escenario de títulos en contenido del usuario. | spec de delivery R1 + requisito de verify; design D10; task 7.1 |
| **M3** wrapper de uninstall inexistente y `RemoveAll` | Corregido. Uninstall es solo de máquina. El retiro va archivo por archivo y el directorio se borra solo si queda vacío. Agregué escenario con `notas.md`, RED y mutante. | design Context §8, D2, D11; spec de migración (requisito de uninstall); tasks 3.3, 8.1, 9.2(g) |
| **M4** objetivo de proyecto inalcanzable y colisión con Antigravity | Derivado a grupo (**G10**). Retiré la recomendación previa por no fundada, con R9 y R10. El escenario de verify pasa a nivel API, condicionado, y agregué el escenario de coinstalación condicionado. La tarea 5.3 y la 3.3 quedan condicionadas a G10. | design Context §7, G10; spec de delivery (dos escenarios "solo si G10 = sí"); proposal; tasks 3.3, 5.3 |
| **M5** cliente sin mecanismo para deshabilitar la invocación implícita | Corregido. Agregué un requisito propio con una rama de descripción restrictiva medida en 11.1, que pasa a G2 si la medición falla. | spec de delivery (requisito "no se activa por pedidos implícitos"); design D4; tasks 1.2, 4.1, 11.1 |
| **M6** mapeo de toggle a módulo | Corregido. Tabla con columna de toggle (delegation → delegación + subagentes; model-routing → ninguno en Codex). Escenarios "sin delegation" y "model-routing activo". | design D5; spec de delivery (requisito de módulos); tasks 4.2, 7.1 |
| **M7** nombre genérico con despacho solo para Codex | Derivado a grupo como **G13**, una decisión nueva compartida con 03. Opciones: (a) `ConfigDeliveryCodexNative`; (b) un contrato neutral. Mi recomendación queda condicionada y la marco como no fundada, porque 03 no existe. G13 bloquea la tarea 3.1. | design D1, tabla G; proposal; tasks 3.1, 11.3 |
| **M8** spike que solo actualiza el design | Corregido. La tarea 1.3 ahora actualiza specs, design y tasks, con re-revisión obligatoria. Agregué la rama **D6-c** con su comportamiento en install y verify. | design D6 (tabla a/b/c); spec de delivery (escenario D6-c); tasks 1.3, 4.1, 7.1 |
| m1 | `refreshSkills` pasó de [SUP] a [EV] con `update/skills.go:97`. El test de regresión se conserva. | design D9; task 8.3 |
| m2 | Reutilizo los goldens existentes y agrego antigravity. Corregí el filtro `-run` y aclaré que solo hay 4 adapters frente a las variantes de asset. | tasks 2.1, 9.1 |
| m3 | La tarea 3.1 se declara cambio de tipo cubierto por el RED de 3.2 y 5.2. | task 3.1 |
| m4 | El preflight pasa a ser un paso de Prepare antes del snapshot. Que el pipeline aborte Apply es [SUP] y se confirma en 6.4. | design D7, Migration Plan; spec de migración (requisito de propiedad); task 6.4 |
| m5 | Seam de escritura propio del instalador nativo, en lugar de `configInstallFn`. | tasks 5.2, 6.2 |
| m6 | Rutas con `~` o relativas para no versionar el home. Que Codex interprete `~` es [SUP] y se confirma en el spike. | design D5; tasks 1.2, 4.3 |
| m7 | Etiqueté [DOC] el esquema TOML y [SUP] la premisa de que "no se listan como skills". | proposal; spec de delivery; design D5, D6 |
| m8 | G8 referencia la decisión de 01 en lugar de duplicarla. | design G8; proposal Impact; task 1.1 |
| m9 | Corregí "ignorados" a "sin trackear". La desviación de rama queda registrada para que el humano la resuelva; no la decido. | tasks ## Orquestación; nota del proposal |
| m10 | Agregué los mutantes (h) heredado, (i) sin recibo y (j) `design-mockups`. | task 9.2 |
| m11 | Escenario G7 "sin arranque" sin `AGENTS.md`: no se crea el archivo. | spec de delivery R1; design D3; task 5.2 |
| m12 | Derivado a 01/08 como observación, sin corregir acá. | design Non-Goals; tasks "Workflow follow-up" |
| m13 | El disparador de "Cambio de toggles" ahora es un cambio de catálogo entre versiones, vía update. | spec de migración; design D5; task 5.2 |

Aparte de los hallazgos, ajusté la descripción del requisito de verify para el límite de 500 caracteres del modo strict (P2 → P3).

## JUICIO

Autor: planner. Comparé el plan contra el informe del revisor y contra el código, en los puntos nuevos de P1.

- **B1 y M1–M3, M5, M6, M8 (confianza alta).** Corregidos en el plan con escenarios y RED verificables. Las rutas nuevas reutilizan `internal/backup` y `internal/filemerge` sin modificarlos. G12-a usa `backup.NewSnapshotter().Create` desde el paquete config.
- **M4 y M7 (confianza alta en el diagnóstico, baja en la recomendación).** Dependen del grupo y de 03. Dejé las opciones y no decidí.
- **Decisión de grupo nueva (G13).** La agregué porque M7 la requiere y el orquestador citó "M7 con 03". Por eso la puerta B2 pasa a nombrar G1–G13 y no G1–G12.
- **B2 sigue abierto por diseño.** El plan sigue siendo un borrador condicional. El texto condicional de specs y design (por ejemplo "según G…", "solo si G10 = sí") se elimina después de las decisiones; no corresponde quitarlo ahora.
- **Validación (confianza alta).** El modo strict pasa (P3) y el estado de los cuatro artefactos es `done` (P4). Esto valida el formato, no el contenido: el juicio de contenido corresponde a la nueva ronda del plan-reviewer.

## SUPOSICIÓN

- Ningún comportamiento del cliente Codex fue observado. Codex sigue sin estar instalado; no volví a ejecutar `which codex`, porque mi entorno no cambió desde E6.
- No leí `internal/pipeline`. Que un fallo en Prepare impida Apply es [SUP] y queda como chequeo previo en la tarea 6.4.
- No confirmé que la interfaz de restauración de la TUI restaure correctamente un snapshot creado fuera de los slots fijos. Leí que `ListManifests` lo lista, pero no ejecuté la restauración. Queda cubierto por los escenarios de 5.4 y 11.2.
- Tomé de R6–R16 del revisor las citas que no releí: `uninstall/types.go:48-62`, `uninstall/steps.go:199-205`, `headless/flags.go:134-145`, `antigravity/adapter.go:61-62`, `writer.go:16-45`, `update/skills.go:97`, `agents/factory.go:16-46` y `section_test.go:122-131`. En el design están atribuidas a R*n*.
- El tamaño del arranque (273 B) viene de R17. No lo medí.

## Fuentes leídas

- `informes/propose-1-plan-reviewer.md`, completo.
- Mis artefactos del intento 1.
- `cmd/active-stack/tui_wiring.go:45-56`; `cmd/active-stack/headless/windows_backups.go:15-28,84-95`.
- Grep de `ListManifests`, `resolveBackupDir` y `Prune(` en `internal/` y `cmd/`.
- `internal/backup/retention.go:131-185` (grep); `internal/uninstall/plan.go:60-80` (grep de `backups`).

## Contaminación por otros informes

- Leí el informe del plan-reviewer completo. Estructura la corrección y aporta la evidencia R1–R17, que usé atribuida.
- Las citas que verifiqué por mi cuenta son las de P1. Las demás vienen de R*n* y están etiquetadas así.
- Mi informe del intento 1 también es fuente. No consulté Engram ni otros informes.

## Pendientes

- Decisiones de grupo G1–G13 y rama D6.
- Aprobación humana.
- Spike de Codex.
- Changes 02 y 03, que no existen.
- Desviación de rama (m9), a resolver por el humano.
- Nueva ronda obligatoria del plan-reviewer (intento 2).
