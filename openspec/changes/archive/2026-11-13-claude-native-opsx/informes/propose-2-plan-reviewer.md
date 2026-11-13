# Informe original: propose, intento 2, plan-reviewer

Autor: subagente plan-reviewer (Claude Code, despachado por el orquestador el 2026-11-06). Identidad del ejecutor sin registro de despacho propio: SUPOSICIÓN.
Revisión de código: rama `claude-native-opsx`, 97d7f6c, `git status --short` vacío. Artefactos ignorados por git; la revisión se ata a estos hashes sha256 (16 hex), que coinciden con los declarados en `informes/propose-2-planner.md`:
proposal.md 978f54a75c0d52f7 · design.md b5fe424064d14bd2 · tasks.md c0e7c58adffd8479 · specs/claude-coordinator-agent 0bf0ed820ff8ba80 · specs/claude-instruction-layers 5a936c709f378f1d · specs/claude-orchestrator-migration 9420051137d9526d

## Veredicto: VUELVE AL PLANNER (ronda corta)

De los 15 hallazgos que le tocaban al planner, 11 quedaron RESUELTOS y 4 PARCIALES; m5 sigue abierto a propósito. B1, M4, M5 y M6 quedan bien como tareas de verificación previas a construir (1.10, 1.11, 1.12, 1.3, 1.5), y el plan dice qué pasa si fallan (escenario "Exclusión no disponible", G9, tarea 2.4). No las cuento como defectos.

Lo que impide compartir son tres defectos MAYOR, todos sobre el código existente y acotados:
- N1: el plan revierte sin decirlo una decisión vigente del repo (uninstall sólo Machine).
- N2: la spec usa un comando `active-stack verify` que no existe y "corrige" un verify Project que ya existe. Este error nace de mi hallazgo M3 del intento 1: lo reconozco.
- N3: el requisito de compatibilidad se contradice con su propio escenario de binario ausente.

Arreglarlos no exige rediseño.

## Hallazgos del intento 1 → estado

| Hallazgo | Estado | Observación |
|---|---|---|
| B1 | RESUELTO (como plan) | Requisito nuevo, tarea 1.10 (a-d), mitigaciones en orden, G9 y salida explícita si falla. Queda el detalle de N7. |
| M1 | RESUELTO | Marcas "(depende de Gn, recomendación)" para G1, G2, G3, G5, G6, G7, G9, G10 y G11 (`rg -o "depende de G[0-9]+" specs`), más la línea explícita en proposal y design. Detalle en N8. |
| M2 | PARCIAL | Los escenarios (a), (b) y (c) y verify como advertencia están bien, y verify ya admite checks `Soft` (`internal/verify/checks.go:27-28`, estado `warning` en `:17`). Pero el texto normativo contradice el escenario (c) (N3) y la composición del bloque legado después de partir el asset no está definida (N4). |
| M3 | PARCIAL | El manifiesto en `<home>` y `projectTargetAdapter` quedaron bien. Uninstall Project choca con una decisión vigente (N1). En verify mi hallazgo original estaba equivocado (N2). |
| M4 | RESUELTO (como plan) | Requisito de oráculo con control positivo, tarea 1.11 y regreso al planner por 2.4 si el oráculo no cubre alguna capa. |
| M5 | RESUELTO | D9 y tarea 1.12: sin copiar credenciales, sólo nombres de variables, borrado registrado. |
| M6 | RESUELTO | Requisito "Precedencia de capas" con tres escenarios alimentados por 1.3 y 1.5. |
| M7 | PARCIAL | El camino `--sync` y G10 están bien. Falta el HOME V1 sin recibo de selección (N5). |
| M8 | RESUELTO | Fallback sólo documentado, verify nombra sólo vías instaladas, tarea 9.3 condicional. |
| M9 | PARCIAL | Manifiesto de migración, hash y `settings.json` fuera del restore están bien. Falta la relación con `--restore-manifest` y el mecanismo de confirmación (N6). |
| m1 | RESUELTO | Matriz partida; los rangos coinciden con los encabezados (`governance.md:1,5,14`; `engram-protocol.md:1,6,13,48,60,84,93`) y con `compose.go:93`. |
| m2 | RESUELTO | Tarea 4.0 con goldens de Codex y Antigravity grabados desde la base. |
| m3 | RESUELTO | Escenarios y RED para colisión Project, cambio de ID, archivo con marca editado, uninstall parcial, marker mal formado e `install --help`. |
| m4 | RESUELTO | Regla "sólo bit de escritura del dueño" y escenario 0500 → 0700. |
| m5 | ABIERTO a propósito | Decisión humana; la tensión ALTO vs CRITICAL quedó explícita y la tarea 1.13 aporta el hecho. No cuenta como defecto. |
| m6 | RESUELTO | G1-G11; citas de `CLAUDE.md:9` y `AGENTS.md:148-149` correctas. Compartir por git queda como pendiente humano y no se cuenta. |

## Hallazgos nuevos

### N1 (MAYOR) Uninstall con target Project revierte sin decirlo una decisión vigente
- Fuente: `cmd/active-stack/headless/uninstall_flags.go:50` ("No --project flag (uninstall is machine-scope, D7/proposal)") y un test que la fija: `uninstall_flags_test.go:179-182` ("--project is not a recognized flag").
- Frente a: tarea 7.5, escenario "Verify y uninstall en target Project" y design D5 "Target Project en verify y uninstall".
- Impacto: el equipo leería como algo nuevo una reversión de una decisión de diseño anterior. Tampoco hay tarea de cableado del CLI (flag, ayuda, i18n) ni se dice que el test existente cambia su aserción, lo que choca con la regla de regresión de consumidores.
- Corrección verificable, una de dos:
  - (a) Nueva decisión de grupo, G12, "uninstall Project", con tareas RED de flag y del test que cambia.
  - (b) Sacar del alcance la desinstalación Project y documentar la limpieza manual de `<root>/.claude/agents/<id>.md` y del bloque universal.
- En ambos casos, citar `uninstall_flags.go:50` en design.

### N2 (MAYOR) No existe `active-stack verify`, y verify ya conoce el target Project
- Fuente: los subcomandos de `cmd/active-stack/main.go` son `version`, `update`, `install`, `starter`, `uninstall` y `windows` (`:64-112`). Verify corre sólo como hook posterior al install (`:199-207`, `:256-263`), y en Project usa `projectVerifyAdapter`, que resuelve con `PathsFor(base, Project)` (`main.go:320-345`).
- Frente a: el disparador de la spec `claude-orchestrator-migration` (`:117`, "`active-stack verify`"), C7 de design ("Verify y uninstall no conocen el target Project") y la tarea 7.4 ("hoy `harness_checks.go:98-130` usa sólo `homeDir`").
- Origen: mi M3 del intento 1 afirmó que verify no conocía el target porque miré `harness_checks.go` sin seguir el adaptador que recibe. El planner lo copió. Es mi error.
- Impacto: un disparador que no existe es un criterio no verificable, y la RED de 7.4 apunta a un problema que no existe (riesgo de test que nace en verde). Además falta un consumidor: `projectVerifyAdapter` tiene que delegar el método nuevo del directorio de agentes, y la tarea 3.1 no lo lista.
- Corrección verificable:
  - Disparador "verify posterior a install (hook)", o una tarea explícita si el grupo quiere un subcomando nuevo.
  - C7 corregido a "uninstall no conoce Project; verify sí, por `projectVerifyAdapter`".
  - 7.4 reescrita como "el check del agente usa el directorio de agentes que resuelve `projectVerifyAdapter`".
  - `cmd/active-stack/main.go:333-345` agregado a los consumidores de 3.1.

### N3 (MAYOR) El requisito de compatibilidad contradice su escenario (c)
- Fuente: spec `claude-coordinator-agent`, requisito "Compatibilidad por versión y entorno sin Claude": "si la versión es menor, desconocida o el binario falta, MUST aplicar el modo legado". Su escenario "Binario claude ausente" (G11) pide en cambio la entrega por capas.
- Impacto: el texto normativo y el escenario piden lo contrario, y el constructor no sabe a cuál escribirle la RED de 6.4(c).
- Corrección verificable: el texto del requisito separa "versión menor o no parseable → modo legado" de "binario ausente → según G11 (recomendación: por capas)".

### N4 (MENOR) Modo legado después de partir el asset
- Después de la tarea 4.1 no existe un asset único del orquestador para Claude. Los escenarios "Versión insuficiente ..." piden "el bloque legado completo como hoy", y no queda dicho:
  - si ese bloque es la concatenación de universal y coordinador o una copia del asset viejo;
  - si un update en modo legado lo refresca;
  - qué golden lo protege.
- Corrección: una línea en D5/D7, más un golden o una aserción de la composición legado en 4.2/4.3.

### N5 (MENOR) Update sobre un HOME V1 sin recibo de selección
- `syncRecordedInstall` no reaplica config si `state.Agents` está vacío: imprime "ejecutá install una vez" y sólo refresca skills (`cmd/active-stack/update.go:375-378`).
- Un usuario de V1 sin `installation.json` no migra, no ve el aviso BREAKING y sigue con el coordinador en `CLAUDE.md`.
- Corrección: escenario con ese caso (resultado esperado: sin migración y con mensaje que nombre el cambio de Claude) y su RED en 6.9.

### N6 (MENOR) Restore: manifiesto registrado frente a `--restore-manifest`, y confirmación
- Hoy el CLI exige `--restore-manifest <path>` (`uninstall_dispatch.go:30-34`), y la spec dice que restore "usa el manifiesto de migración registrado". No está definido qué pasa si el usuario pasa otro manifiesto.
- La "confirmación explícita" no tiene flag nombrado. Ya existe `--yes` (`uninstall_flags_test.go:156`), y usarlo para saltear la protección por hash sería riesgoso.
- Además, `internal/backup/retention.go:119-131` (`Prune`) borra respaldos viejos no fijados (pinned). Hoy no encontré quien llame a `Prune`, así que es riesgo futuro.
- Corrección:
  - definir la precedencia entre el flag y el registro;
  - nombrar el mecanismo de confirmación, distinto de `--yes` o justificado;
  - marcar el manifiesto de migración como fijado (pinned).

### N7 (MENOR) B1, escenario 1: negativo de una sola corrida sobre conducta del modelo
- "La transcripción no registra ningún subagente del tipo `<id>`" depende de que el modelo no decida delegar en esa corrida. Si la mitigación 1 (frontmatter) existe, hay una aserción determinista: el agente no figura entre los tipos disponibles.
- Corrección: el THEN prioriza esa aserción determinista. Si sólo quedan las mitigaciones 2 y 3, declarar N corridas y el umbral.

### N8 (MENOR) Falta la marca G2 en "Sesión normal ..."
- La tabla de design dice que G2 afecta a la spec `claude-instruction-layers` (sesión normal), pero ese requisito no tiene marca "(depende de G2 ...)".

## EVIDENCIA
Todas el 2026-11-06, revisión 97d7f6c, sin ejecutar modelo.
- 09:32:59-03:00 `sha256sum` de los seis artefactos: coinciden con la tabla de `propose-2-planner.md`.
- `rg -o "depende de G[0-9]+" specs | sort | uniq -c`: marcas para G1, G2 (x2), G3 (x2), G5, G6, G7, G9, G10 y G11.
- búsqueda de em dash (U+2014) con `rg -c` sobre los artefactos y el informe del planner: sin coincidencias (exit 1).
- `cmd/active-stack/update.go:155-168` (`--sync` → `syncRecordedInstall`), `:353-360` (`installArgs` sin target Project) y `:370-414` (con `state.Agents` vacío no reaplica config). `internal/update/state.go:26-34` (`Projects`, `installation.json`). Las citas M7 del planner son correctas.
- `rg -n "^#"` en `governance.md` y `engram-protocol.md`: los rangos de la matriz coinciden. `compose.go:93`: `"Strict TDD Mode: enabled\n\n"`.
- `cmd/active-stack/headless/uninstall_flags.go:44-56` (machine-scope declarado); `uninstall_flags_test.go:179-182` (`--project` rechazado).
- `rg -n 'case "[a-z-]+"' cmd/active-stack/main.go`: no hay subcomando `verify`. `main.go:199-207`, `:256-263` y `:320-345`: `projectVerifyAdapter`.
- `internal/verify/checks.go:16-29`: estados `skipped`/`warning` y campo `Soft`.
- `internal/backup/retention.go:13-14`, `:119-131` (`Prune`, pinned nunca se borra). `rg "Prune\("` fuera de `retention.go` y tests: sin coincidencias.
- `openspec validate claude-native-opsx --strict` exit 0: reusado del planner, no reejecutado (hashes iguales a los suyos).

## JUICIO
- Autor: plan-reviewer.
- Comparación: cada hallazgo del intento 1 contra los artefactos con los hashes de arriba, y las afirmaciones nuevas (update real, Project, verify, matriz) contra el código en 97d7f6c.
- Confianza:
  - Alta en N1, N2, N3, N5 y N6: observados en código o en el texto de la spec.
  - Media en N4: depende de cómo se implemente la partición.
  - Media en N7.
- Las tareas 1.x con resultado incierto no son defectos: tienen consecuencia escrita. N1 a N3 sí lo son, porque describen mal el sistema actual o se contradicen dentro de la spec, y eso confundiría al equipo.

## SUPOSICIÓN (no verificado)
- Que `install --help` esté localizado en ambos idiomas: no encontré `installUsage` en `cmd/active-stack` y no rastreé de dónde sale la ayuda.
- Que algún camino futuro llame a `Prune` sobre `backups/install`.
- No corrí sesiones de Claude con modelo, ni código del change (no existe), ni `go test` en este intento (sin cambios de código desde la línea base del intento 1, 09:22:25-03:00, todo ok).

## Fuentes leídas
`informes/propose-2-planner.md`, los seis artefactos completos (proposal en las secciones What Changes, Clasificación, Aprobación y KB Impact), y el código listado en EVIDENCIA.

## Lecciones L-NNN
El repo no tiene base de lecciones (declarado por el planner; no reverificado con listado propio).

## Contaminación
- Leí `propose-2-planner.md` por encargo y mi propio informe del intento 1.
- La tabla de cambios del planner fue la guía de qué mirar, pero verifiqué cada estado contra los artefactos y el código, no contra su tabla.
- N2 corrige un error mío que el planner heredó.
- Sin Engram.
