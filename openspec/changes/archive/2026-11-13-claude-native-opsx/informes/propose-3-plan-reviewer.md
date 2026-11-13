# Informe original: propose, intento 3, plan-reviewer

Autor: subagente plan-reviewer (Claude Code, despachado por el orquestador el 2026-11-06). Identidad del ejecutor sin registro de despacho propio: SUPOSICIÓN.
Revisión de código: rama `claude-native-opsx`, 97d7f6c, `git status --short` vacío. Artefactos ignorados por git; la revisión se ata a estos hashes sha256 (16 hex), que coinciden con los de `informes/propose-3-planner.md`:
proposal.md 233d08ba07f559c9 · design.md bef205f75a178a52 · tasks.md ee34c9cc9b5f2034 · specs/claude-coordinator-agent 373bc9e54842ac72 · specs/claude-instruction-layers 5bd5245475d3b1a1 · specs/claude-orchestrator-migration c8412e0608c53f64

## Veredicto: LISTO PARA COMPARTIR

N1 a N8 quedaron resueltos en las specs, en design y en tasks. No encontré regresiones que impidan compartir. Esto no autoriza construir. Antes de la sección 3 de tasks siguen pendientes:
- las decisiones de grupo G1 a G12;
- la matriz acordada con la Propuesta 04;
- las verificaciones 1.x;
- la aprobación humana fechada en `## Aprobación`, con la elección ALTO o CRITICAL (m5).

## N → estado

| Hallazgo | Estado | Verificación |
|---|---|---|
| N1 | RESUELTO | Uninstall Project queda fuera del alcance, citando la decisión vigente (`uninstall_flags.go:50`). Requisito nuevo "Limpieza manual documentada para target Project" con escenario "Uninstall sigue rechazando --project" (sin tocar el test `:179-182`). G12 en design:266 sin recomendación, DEUDA-1 en tasks:135, y las tareas 7.5 y 8.2 ajustadas. No queda texto que diga que se agrega target a uninstall (`rg "agrega target"` sin coincidencias). |
| N2 | RESUELTO | El disparador es ahora el "hook de verify posterior a `active-stack install` (no existe un subcomando `verify`)" (spec migration, en el requisito Verify). C7 corregido (design:322). La tarea 7.4 apunta al adaptador recibido, con un caso negativo (un archivo sólo en `<home>` no hace pasar el check Project). `projectVerifyAdapter` (`main.go:333-345`) figura en los consumidores de 3.1. `rg "active-stack verify"` no devuelve usos como comando. |
| N3 | RESUELTO | El texto del requisito separa "versión menor o no parseable → legado" de "binario ausente → G11". Concuerda con los cuatro escenarios. |
| N4 | RESUELTO | La "composición legado" está definida como byte a byte igual a la de 97d7f6c dentro de `active-stack:sdd-orchestrator`. Hay escenario de refresco idempotente en update. Las tareas 4.2 y 4.3 conservan `compose_all_toggles_claude.golden` sin regrabarlo, con mutante, y la 6.4(d) cubre el update. El golden existe y lo afirma `compose_test.go:51`. |
| N5 | RESUELTO | Escenario "HOME V1 sin registro de selección" (cita `update.go:375-378`, verificada en el intento 2) y RED en 6.9. |
| N6 | RESUELTO | Manda `--restore-manifest`. La guarda por hash se aplica también a un manifiesto ajeno, con advertencia. Hay un flag propio de confirmación (`--overwrite-edited`, nombre tentativo, decisión de grupo). `--yes` no saltea la guarda, con escenario y mutante en 7.2. El manifiesto se crea con `pinned: true`; el campo existe (`internal/backup/manifest.go:62-64`) y la tarea 6.8 lo prueba con `Prune` y límite 0 o 1. |
| N7 | RESUELTO | B1 queda en dos ramas. Con exclusión por frontmatter: aserción determinista, una corrida. Sin exclusión: 20 corridas con umbral 0, y con una o más apariciones se aplica "Exclusión no disponible". La 8.1 está alineada. El umbral queda para el grupo, como pidió el orquestador. |
| N8 | RESUELTO | `claude-instruction-layers` tiene "(depende de G2, recomendación)" en el requisito "Sesión normal ...". |

## Hallazgos nuevos

Ninguno que bloquee compartir. Una observación MENOR, que no bloquea y puede ir como nota para el constructor o resolverla el grupo:

- **O1 (MENOR, no bloquea).** El único golden de Claude cubre la combinación con todos los toggles (`compose_all_toggles_claude`). La igualdad byte a byte del modo legado queda protegida sólo para esa combinación; si la partición altera separadores en otras combinaciones (por ejemplo sin `delegation`), no se detecta en RED. Corrección sugerida: en 4.2, una aserción de igualdad del legado contra `Compose` de la base para dos o tres combinaciones más, grabadas desde 97d7f6c igual que la tarea 4.0.

## EVIDENCIA
Todas el 2026-11-06, revisión 97d7f6c, sin ejecutar modelo.
- 09:39:36-03:00 `sha256sum` de los seis artefactos: coinciden con la tabla del planner.
- Lectura completa de los requisitos cambiados:
  - `claude-coordinator-agent`: "La sesión normal no lanza...", "Compatibilidad por versión y entorno sin Claude".
  - `claude-orchestrator-migration`: Rollback, Verify, "Limpieza manual documentada", Update.
  - Tareas 3.1, 4.2, 4.3, 6.4, 6.8, 6.9, 7.2, 7.4, 7.5, 8.1, 8.2 y la sección DEUDA.
  - design.md: G12 (:266), C7 (:322), D5 (:189-197).
- `rg` de afirmaciones viejas ("agrega target", "target a verify y uninstall", "active-stack verify", "G1-G7/G8/G10/G11"): sin restos contradictorios. `design.md:24` cita `harness_checks.go:98-130` sólo como descripción del recorrido actual, que sigue siendo correcto.
- `rg -o "depende de G[0-9]+" specs | sort | uniq -c`: G1, G2 (x3), G3 (x2), G5, G6, G7, G9, G10, G11 y G12.
- Búsqueda de em dash (U+2014) con `rg -c` sobre los artefactos: sin coincidencias (exit 1).
- `internal/backup/manifest.go:62-64`: campo `Pinned`. `internal/harness/config/compose_test.go:13-14, 51`: golden de Claude con flag `-update`.
- `openspec validate claude-native-opsx --strict` exit 0 a las 09:38:26: reusado del planner, no reejecutado (hashes iguales).
- No corrí `go test`: el código no cambió desde la línea base del intento 1 (09:22:25-03:00, todo ok).

## JUICIO
- Autor: plan-reviewer.
- Comparación: cada N del intento 2 contra el texto actual de los artefactos y contra el código citado.
- Confianza alta en que N1 a N8 están resueltos y en que no hay contradicciones nuevas en las secciones que cambiaron. Confianza media en la ausencia total de regresiones fuera de esas secciones: no releí línea por línea las secciones que no cambiaron, y las comparé sólo por búsquedas dirigidas.
- No cuento como defectos, por instrucción y porque son decisiones legítimas de grupo o de la persona: G12 sin recomendación, el nombre `--overwrite-edited`, N = 20 con umbral 0, m5 y la forma de compartir los artefactos (git).

## SUPOSICIÓN (no verificado)
- Que la ayuda de `install` sea localizable (pendiente desde el intento 2; está cubierta por la tarea 6.7, que exige un test de presencia en ambos idiomas).
- Que exista un registro determinista de los tipos de subagente disponibles (depende de 1.10 y 1.11). El plan define qué pasa si no existe.
- No ejecuté sesiones de Claude ni código del change (no existe).

## Fuentes leídas
`informes/propose-3-planner.md`; los seis artefactos (secciones cambiadas completas y búsquedas en el resto); `internal/backup/manifest.go:58-66`; `internal/harness/config/compose_test.go` (búsqueda); mi informe del intento 2.

## Lecciones L-NNN
El repo no tiene base de lecciones (declarado en los intentos anteriores; no reverificado).

## Contaminación
- Leí el informe del planner del intento 3 y los míos de los intentos 1 y 2.
- Usé la tabla del planner como índice y verifiqué cada estado en los artefactos.
- Las citas de código que reusé del intento 2 (`update.go`, `uninstall_flags*.go`, `main.go`, `retention.go`) no cambiaron, porque el código sigue en 97d7f6c.
- Sin Engram.
