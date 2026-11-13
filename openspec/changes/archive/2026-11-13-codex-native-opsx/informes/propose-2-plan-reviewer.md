# propose-2 · plan-reviewer · codex-native-opsx

- **Autor / rol**: plan-reviewer (subagente nativo `role-plan-reviewer`, Claude, claude-opus-5-5). Es la misma familia que el planner: no hay independencia de proveedor.
- **Revisión revisada**: los cinco artefactos reescritos en el intento 2, sin trackear, tal como estaban en disco el 2026-11-06 a las ~13:50Z. HEAD `1b4639b`; `internal/` y `cmd/` sin cambios respecto de `bd63543`.
- **No hice**: ediciones, build, commits ni `go test`.

## EVIDENCIA

- **E1.** Check del planner reutilizado: `openspec validate --strict`, exit 0, a las 13:42:16Z. Lo re-ejecuté a la hora de este informe y sigue válido. Solo valida el formato.
- **E2.** Leí `informes/propose-2-planner.md`, la tabla hallazgo → cambio.
- **E3.** Verifiqué la evidencia nueva de G12:
  - `internal/tui/backups_screen.go:32` llama a `backup.ListManifests(m.deps.BackupDir)`.
  - `internal/backup/retention.go:175-200` lista **un solo nivel**: `backups/<dir>/manifest.json`.
  - `cmd/active-stack/headless/windows_backups.go:24` tiene `windowsBackupSlots = {"install","uninstall"}`.
  - Conclusión: el nombre plano `codex-migration-<ts>` que propone D12 sí aparece en la TUI y no aparece en la GUI de Windows. Coincide con lo que dice el plan.
- **E4.** Leí, sin re-ejecutar, las secciones nuevas o modificadas:
  - design: puerta B2, D1/G13, D5 (tabla de toggles), D6 a/b/c, D12, tabla G y Migration Plan;
  - spec de delivery: los requisitos y escenarios listados por grep, incluido el nuevo de invocación implícita;
  - spec de migración: G12-a/b, escenario de markers coexistentes y escenario de uninstall con `notas.md`;
  - tasks: 1.3, 2.1, 3.1–3.3, 5.4, 5.5, 6.4 y 8.1.

## Estado de los hallazgos del intento 1

| Hallazgo | Estado | Fuente |
|---|---|---|
| B1 | **Cerrado en el plan**, condicionado a G12. Hay requisito, escenarios (dos corridas de install, reinstalación) y RED para cada rama (5.4 y 5.5). | spec de migración :108-155; D12; tasks 5.4 y 5.5 |
| B2 | Puerta humana declarada explícitamente (G1–G13 y D6). No cuenta como defecto. | design :11; proposal ## Aprobación |
| M1 | Cerrado | spec de migración :30; delivery :232; tasks 5.2, 7.1 y 9.2(h) |
| M2 | Cerrado | delivery :35 y requisito de verify |
| M3 | Cerrado | tasks 3.3 y 8.1 (sin `RemoveAll`, escenario `notas.md`); spec de migración :197 |
| M4 | Derivado a G10, con escenarios condicionados | delivery :238 y :244; tasks 3.3 y 5.3 |
| M5 | Cerrado | delivery :77-93; tasks 1.2 y 11.1 |
| M6 | Cerrado | D5 (tabla); delivery :139-177 |
| M7 | Derivado a G13, que bloquea la tarea 3.1 | D1; tasks 3.1 |
| M8 | Cerrado | tasks 1.3 (re-revisión obligatoria); D6-c; delivery :111 |
| m1 a m13 | Cerrados según la tabla del planner. Muestreé m2, m3, m4 (6.4), m5, m6, m11 y m13, y concuerdan. m12 queda derivado. | tasks; D5 |

## JUICIO

Autor: plan-reviewer. Confianza alta en todo lo que es lectura del plan y del código.

- Las correcciones atienden los defectos del intento 1 sin agregar alcance no pedido.
- G12 y G13 están bien planteados como decisiones, con opciones y evidencia.
- La rama D6-c es coherente entre design, spec (:111) y tasks (3.2: `RolesDir` vacío; 7.1: no exige roles).
- No encontré contradicciones nuevas de fondo entre specs, design y tasks.
- El plan está apto para presentarse a la decisión del grupo y del humano.

## SUPOSICIÓN

- No observé nada del cliente Codex, que sigue sin estar instalado.
- No releí cada línea de los 1.207 renglones de los artefactos: revisé por muestreo las zonas tocadas.
- No ejecuté ninguna restauración desde la TUI.
- No leí `internal/pipeline` (queda como tarea 6.4, igual que la declara el plan).

## Fuentes y contaminación

- **Fuentes**: los artefactos citados en E4, `propose-2-planner.md` y el código citado en E3.
- **Contaminación**: conozco mi propio informe del intento 1 y el informe del planner. Las afirmaciones de E3 las verifiqué por mi cuenta. No consulté Engram.

## Hallazgos nuevos

- **n1 · MENOR.** El requisito G12-a dice "MUST figurar en la pantalla de backups", pero la GUI de Windows no lista ese directorio (E3). El escenario solo cubre la TUI. Hay que precisar el texto ("TUI", y la GUI de Windows según la sub-decisión de G12) para que no se pueda leer como incumplido.
- **n2 · MENOR.** En G12-a, si la instalación que migra falla después de crear el snapshot de migración, el pipeline revierte, pero ese snapshot queda fuera del manifiesto: no se borra y el reintento crea otro. Esto contradice "exactamente un snapshot" en el caso de reintento. Falta un escenario: o se elimina en el rollback o se acepta y se documenta que puede haber más de uno.
- **n3 · MENOR.** Restaurar el snapshot de migración borra las rutas de D8 que no existían al momento de tomarlo. Si después de migrar el usuario creó un archivo propio en una de esas rutas, la restauración lo eliminaría. Conviene agregar un escenario, o un aviso en la guía.

## Veredicto

**PLAN APTO.** El plan está listo para ir a las decisiones de grupo (G1–G13 y D6) y a la aprobación humana. Los tres menores se pueden corregir junto con la incorporación de esas decisiones.

`apply` sigue bloqueado por B2 (aprobación humana) y por la re-revisión obligatoria después del spike (tarea 1.3).
