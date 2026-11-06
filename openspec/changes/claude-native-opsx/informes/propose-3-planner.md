# Informe original: propose, intento 3, planner

Autor: subagente planner (Claude Code, despachado por el orquestador el 2026-11-06). Identidad del
ejecutor sin registro de despacho propio: SUPOSICIÓN.
Encargo: corregir sólo N1-N8 de `informes/propose-2-plan-reviewer.md`. Sin código de producción.
Sin cambios en git ni en `.gitignore`. m5 sigue abierto a propósito. Último intento antes de
escalar a la persona (registrado en tasks.md).

Revisión: rama `claude-native-opsx`, código en 97d7f6c sin cambios. Artefactos ignorados por git
(`.gitignore:36`). Hashes sha256 (primeros 16 hex) al cerrar este intento:

| Artefacto | Intento 2 | Intento 3 |
|---|---|---|
| proposal.md | 978f54a75c0d52f7 | 233d08ba07f559c9 |
| design.md | b5fe424064d14bd2 | bef205f75a178a52 |
| tasks.md | c0e7c58adffd8479 | ee34c9cc9b5f2034 |
| specs/claude-coordinator-agent/spec.md | 0bf0ed820ff8ba80 | 373bc9e54842ac72 |
| specs/claude-instruction-layers/spec.md | 5a936c709f378f1d | 5bd5245475d3b1a1 |
| specs/claude-orchestrator-migration/spec.md | 9420051137d9526d | c8412e0608c53f64 |

## N → cambio hecho

| Hallazgo | Cambio | Dónde |
|---|---|---|
| N1 | Uninstall con target Project queda fuera del alcance. Se cita la decisión vigente (`uninstall_flags.go:50`, fijada por `uninstall_flags_test.go:179-182`). Requisito nuevo "Limpieza manual documentada para target Project": la salida de install Project y README nombran qué quitar a mano, y `uninstall --project` sigue rechazado sin cambiar el test. G12 nuevo sin recomendación. DEUDA-1 asignada al primer change que necesite desinstalar en Project. | spec `claude-orchestrator-migration` (requisito nuevo; "Uninstall sin restore..." ahora dice target Machine; escenario Project sólo de verify); design.md D5 "Target Project en verify y uninstall", G12, C7; proposal.md Impact; tasks.md 7.5, 8.2, 9.1, sección "DEUDA registrada por el plan" |
| N2 | Disparador corregido: es el hook de verify posterior a `active-stack install`; no existe subcomando `verify` (`main.go:64-112`). Verify ya conoce Project por `projectVerifyAdapter` (`main.go:320-345`). C7 corregido y atribuido al intento 1. La tarea 7.4 pasa a ser "el check del agente usa el directorio que resuelve el adaptador recibido". `projectVerifyAdapter` (`main.go:333-345`) agregado a los consumidores de 3.1. | spec `claude-orchestrator-migration` (requisito Verify, escenario "Verify en target Project"); design.md D5, C7; proposal.md Impact; tasks.md 3.1, 7.4 |
| N3 | El texto del requisito separa dos casos: "versión menor o no parseable → modo legado" y "binario ausente → según G11 (recomendación: por capas con advertencia)". | spec `claude-coordinator-agent`, requisito "Compatibilidad por versión y entorno sin Claude" |
| N4 | "Composición legado" definida: concatenación de universal, coordinador y módulos en el orden de hoy, byte a byte igual a la composición de 97d7f6c. La protege el golden actual de Claude, que se conserva sin regrabar. Un update en modo legado la refresca de forma idempotente (escenario nuevo). | design.md D7 "Modo legado después de partir el asset"; spec `claude-coordinator-agent` (escenario b reescrito y "Update en modo legado refresca el bloque legado"); tasks.md 4.2, 4.3, 6.4(d) |
| N5 | Escenario "HOME V1 sin registro de selección": update no migra ni escribe `CLAUDE.md` (`update.go:375-378`), y su salida suma la nota BREAKING de Claude y el comando `active-stack install`. RED en 6.9. | spec `claude-orchestrator-migration` (requisito Update); design.md D5 "Update"; tasks.md 6.9 |
| N6 | `--restore-manifest` manda. Con otro manifiesto, la guarda por hash se aplica igual y la salida advierte. Confirmación con un flag propio (nombre tentativo `--overwrite-edited`); `--yes` no saltea la guarda, con escenario y mutante. El manifiesto de migración se crea fijado para que `Prune` (`retention.go:119-131`) no lo borre (`manifest.go:62-64`). | spec `claude-orchestrator-migration` (requisito Rollback y cuatro escenarios); design.md D5 pasos 7 y "Rollback"; tasks.md 6.8, 7.2, 9.1 |
| N7 | B1 en dos escenarios. Si 1.10 encuentra exclusión por frontmatter: aserción determinista de que el agente no figura entre los subagentes disponibles, con una corrida. Si no: N = 20 corridas con umbral 0 de 20; N y el umbral los confirma el grupo con G9. | spec `claude-coordinator-agent` (requisito "La sesión normal no lanza..."); tasks.md 8.1 |
| N8 | Marca "(depende de G2, recomendación)" en el requisito "Sesión normal conserva reglas universales...". | spec `claude-instruction-layers` |

Rango de decisiones actualizado a G1-G12 en proposal, design y tasks.

## EVIDENCIA

Todas el 2026-11-06, código en 97d7f6c.
- 09:38:26-03:00 `openspec validate claude-native-opsx --strict` → "Change 'claude-native-opsx' is valid" (exit 0). Se repitió con el mismo resultado después de agregar la marca G12.
- `openspec status --change claude-native-opsx --json` → `isPlanningComplete: true`, cuatro artefactos `done`.
- Búsqueda de em dash (U+2014) con `rg -c` sobre `*.md` del change → sin coincidencias (exit 1).
- `rg -o "depende de G[0-9]+" specs`: G1, G2 (x3), G3 (x2), G5, G6, G7, G9, G10, G11; G12 se agregó después de ese conteo.
- Código releído para verificar las citas del revisor:
  - `cmd/active-stack/headless/uninstall_flags.go:44-56`: "No --project flag (uninstall is machine-scope, D7/proposal)" en `:50`.
  - `uninstall_flags_test.go:154-158` (caso `--yes`) y `:179-182` (`--project` rechazado).
  - `cmd/active-stack/main.go`: subcomandos en `:64-112`, sin `verify`; hook de verify con `verifyBase = ProjectRoot` en Project en `:199-207`; `projectVerifyAdapter` en `:320-345`.
  - `cmd/active-stack/update.go:370-378`: con `state.Agents` vacío, mensaje y sólo `refreshSkills`.
  - `cmd/active-stack/uninstall_dispatch.go:28-34`: restore exige `--restore-manifest`.
  - `internal/backup/retention.go:13-14` y `:119-131`: los respaldos fijados nunca se podan.
  - `internal/backup/manifest.go:62-64`: campo `Pinned`.

## JUICIO

Autor: planner. Comparé cada N contra el texto de las specs y el código citado.
- Confianza alta en N1, N2, N3, N5, N6 y N8: se apoyan en código leído o son correcciones de texto.
- Confianza media en N4: la igualdad byte a byte del legado supone que la partición sólo reubica texto. Si 4.1 necesita tocar separadores, el golden legado va a fallar y eso se detecta en RED, no en silencio.
- Confianza media en N7: N = 20 con umbral 0 es una elección mía para que el escenario sea comprobable, no un valor estadístico justificado. Lo confirma el grupo con G9.

Decisiones nuevas que no cierro: G12 (sin recomendación) y el nombre `--overwrite-edited` (tentativo).

## SUPOSICIÓN (no verificado)

- Que el contenido de `install --help` sea extensible y localizable (el revisor no encontró `installUsage`; no lo rastreé).
- Que la partición del asset pueda reproducir byte a byte la composición actual (N4).
- Que exista un registro de los tipos de subagente disponibles que sirva de oráculo determinista para N7 (depende de 1.10 y 1.11).
- No ejecuté sesiones de Claude, código del change ni `go test`; no hay cambios de código.

## Fuentes leídas

`informes/propose-2-plan-reviewer.md` (completo); los seis artefactos; el código listado en EVIDENCIA.

## Lecciones L-NNN

No hay base de lecciones ni `kb/` en el repo (declarado en el intento 1; no reverificado).

## Contaminación

Partí del informe del revisor por encargo, incluidas sus citas de `main.go`, `uninstall_flags*.go`, `retention.go` y `update.go:375-378`. Las reverifiqué todas en código antes de usarlas. No consulté Engram.
