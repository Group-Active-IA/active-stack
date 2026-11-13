# Informe original: propose, intento 2, planner

Autor: subagente planner (Claude Code, despachado por el orquestador el 2026-11-06). Identidad del
ejecutor sin registro de despacho propio: SUPOSICIÓN.
Encargo: corregir sólo los hallazgos de `informes/propose-1-plan-reviewer.md` (1 BLOQUEANTE,
9 MAYOR, 6 MENOR). Sin código de producción. Sin cambios en git ni en `.gitignore`.

Revisión: rama `claude-native-opsx`, código en 97d7f6c sin cambios. Artefactos ignorados por git
(`.gitignore:36`); hashes sha256 (primeros 16 hex) al cerrar este intento:

| Artefacto | sha256 |
|---|---|
| proposal.md | 978f54a75c0d52f7 |
| design.md | b5fe424064d14bd2 |
| tasks.md | c0e7c58adffd8479 |
| specs/claude-coordinator-agent/spec.md | 0bf0ed820ff8ba80 |
| specs/claude-instruction-layers/spec.md | 5a936c709f378f1d |
| specs/claude-orchestrator-migration/spec.md | 9420051137d9526d |

## Hallazgo → cambio hecho

| Hallazgo | Cambio | Dónde |
|---|---|---|
| B1 | Requisito nuevo "La sesión normal no lanza el coordinador como subagente" con tres escenarios (pedido OPSX sin nombrarlo, coordinador lanzado como subagente con guarda, exclusión no disponible → no se construye la vía hasta G9). Mitigaciones en orden (frontmatter, `description`, guarda en el cuerpo). Tarea de verificación 1.10 (a-d). G9 nuevo. Caso en 8.1. | spec `claude-coordinator-agent`; design.md D1 "Exposición como subagente", tabla G9, Risks; tasks.md 1.10, 8.1, Estado (bloqueante 5) |
| M1 | Marca "(depende de Gn, recomendación)" en cada requisito o escenario dependiente (G1, G2, G3, G5, G6, G7, G9, G10, G11). Línea explícita "las specs están escritas sobre las recomendaciones" en proposal y design. | las tres specs; proposal.md What Changes; design.md "Decisiones pendientes de grupo" |
| M2 | Requisito "Compatibilidad por versión y entorno sin Claude" con escenarios (a) insuficiente con legado, (b) insuficiente o desconocida sin entrega previa, (c) binario ausente, y resultado de verify en cada uno (advertencia, no fallo). G11 para el caso sin binario. RED por caso. | spec `claude-coordinator-agent`; spec `claude-orchestrator-migration` (Verify); design.md D7, G11; tasks.md 3.3, 6.4, 7.1 |
| M3 | Verify y uninstall con target Project declarados en alcance, con RED propias; `projectTargetAdapter` (`internal/install/plan.go:556-580`) en consumidores de 3.1; manifiesto bajo `<home>` también en Project (contradicción C8, `install.go:35-36` vs `plan.go:138-140`). 8.2 condicionada a 7.4 y 7.5. | spec `claude-orchestrator-migration` (Verify en ambos targets, Migración); design.md D5, C7, C8; tasks.md 3.1, 5.5, 7.4, 7.5, 8.2 |
| M4 | Requisito "Oráculo de contexto efectivo independiente del modelo" (registro de contexto de la transcripción, control positivo en la misma corrida). THEN reescritos a "registro de contexto". Tarea 1.11 que valida el oráculo antes de usarlo. | spec `claude-instruction-layers` (primer requisito y todos los THEN con testigo); spec `claude-coordinator-agent`; design.md D8; tasks.md 1.11, 8.1 |
| M5 | Mecanismo de autenticación: directorio de configuración temporal por variable de entorno y credencial por variable ya presente; prohibido copiar archivos de credenciales; sólo nombres de variables; borrar el directorio. Tarea 1.12. | design.md D9, Risks; tasks.md 1.12, preámbulos de secciones 1 y 8 |
| M6 | Requisito "Precedencia de capas declarada y probada" con escenarios usuario vs coordinador, agente Project vs Machine con el mismo ID (condicionado a 1.5) y CLAUDE.md global vs proyecto. 1.3 y 1.5 alimentan la spec. | spec `claude-instruction-layers`; tasks.md 1.3, 1.5, 8.2, 9.1 |
| M7 | Requisito "Update real migra..." sobre el camino `update --sync` (`cmd/active-stack/update.go:94`, `:159-165`, `:353-360`) y proyectos registrados (`internal/update/state.go:30`); proyectos Project V1 reciben advertencia y comando de migración (G10). | spec `claude-orchestrator-migration`; design.md D5 "Update", G10; tasks.md 6.9; proposal.md Impact |
| M8 | Fallback: en este change se documenta, no se instala (recomendación dentro de G2); verify nombra sólo vías instaladas; escenario "El fallback documentado funciona"; tarea 9.3, con RED de instalación y detección sólo si el grupo decide instalarlo. | spec `claude-coordinator-agent` (Fallback); design.md D3; tasks.md 9.3 |
| M9 | Rollback con manifiesto de migración registrado (no el último snapshot), hash post-migración, sin escribir si el usuario editó salvo confirmación, `settings.json` fuera del restore. Escenarios nuevos. | spec `claude-orchestrator-migration` (Rollback que no pisa ediciones posteriores, Migración, Reinstalación); design.md D5 pasos 3 y 7, "Rollback"; tasks.md 6.8, 7.2 |
| m1 | Matriz de una capa por fila: `governance.md` partida (`:1-13` módulo propose, `:14-19` coordinador), `engram-protocol.md` partida por encabezados; línea `Strict TDD Mode: enabled` (`compose.go:93`) agregada. Spec exige partir o justificar. | design.md matriz; spec `claude-instruction-layers` (Matriz, escenario "Regla en dos capas") |
| m2 | Tarea 4.0 que agrega goldens de Codex y Antigravity desde 97d7f6c antes de tocar `compose.go`; escenario reescrito. | tasks.md 4.0, 4.3; spec `claude-orchestrator-migration` (OpenCode y otros agentes) |
| m3 | Escenarios: colisión en `<root>/.claude/agents`, cambio de ID (G5), archivo de agente con marca editado por el usuario (declarado), uninstall parcial, advertencia por marker mal formado, advertencia en `install --help`. RED correspondientes. | specs `claude-coordinator-agent` y `claude-orchestrator-migration` y `claude-instruction-layers`; tasks.md 4.5, 5.2, 6.7, 6.10, 6.11, 7.6 |
| m4 | Riesgo del commit 6a40f17 (`writer.go:37-41`): regla "sólo bit de escritura del dueño" y escenario de directorio 0500 reparable. | design.md D6, Risks; spec `claude-orchestrator-migration` (Permisos); tasks.md 5.4 |
| m5 | No resuelto, como pidió el orquestador. Clasificación reescrita con la tensión ALTO vs CRITICAL explícita para la persona en `## Aprobación`; "si settings.json puede guardar credenciales" pasó a tarea de verificación 1.13. | proposal.md Clasificación y comentario de `## Aprobación`; tasks.md 1.13, 2.5, Estado |
| m6 | `G1-G7` corregido a `G1-G11`; cita literal de `AGENTS.md:148-149` y `CLAUDE.md:9`. Compartir con el equipo: sin tocar git, registrado como pendiente humano (`AGENTS.md:150` declara el ignore intencional). | proposal.md (Aprobación, Knowledge Base Impact); design.md C6; tasks.md Estado |

Cambio adicional, por la validación estricta: las notas antes del primer requisito generaban
advertencias de `openspec validate --strict`. El oráculo pasó a ser requisito propio y la nota
"specs sobre recomendaciones" quedó sólo en proposal.md y design.md.

## EVIDENCIA

Todas el 2026-11-06, código en 97d7f6c.
- 09:31:05-03:00 `openspec validate claude-native-opsx --strict`: tres WARNING "Text before the
  first requirement" (exit 1). Corregido.
- Después de la corrección: `openspec validate claude-native-opsx --strict` → "Change
  'claude-native-opsx' is valid" (exit 0).
- `openspec status --change claude-native-opsx --json` → `isPlanningComplete: true`, cuatro
  artefactos `done`.
- 09:31:20-03:00 búsqueda de em dash (U+2014) con `rg -c` sobre `*.md` en el change → sin coincidencias (exit 1).
- `rg "depende de[^)]*\bGn\b" specs/` → al menos una coincidencia para G1, G2, G3, G5, G6, G7, G9,
  G10 y G11 (corrida antes de mover las notas; las marcas por requisito no cambiaron).
- Código releído para las correcciones: `cmd/active-stack/update.go:80-110,155-175,353-370,510-520`
  (`--sync` reaplica selección con `install --headless --home` sin target Project);
  `internal/update/state.go:25-40` (`Projects []string`); `internal/install/plan.go:136-142`
  (snapshot contra el home real) y `:552-582` (`projectTargetAdapter`);
  `internal/harness/config/install.go:35` (`snapshotDir` con `homeDir`); `AGENTS.md:147-151`;
  `CLAUDE.md:9`; `internal/harness/config/compose.go:93` (`"Strict TDD Mode: enabled\n\n"`).
- `claude --help` (2.1.318): `--agents <json-or-file>` admite ruta de archivo sólo "with --print".

## JUICIO

Autor: planner. Comparación: cada hallazgo del informe del plan-reviewer contra los artefactos
corregidos y el código citado. Confianza: alta en M2, M3, M7, M9, m1, m2, m3, m6 (anclados en
código y escenarios concretos); media en B1, M4, M5 y M6, porque su cierre depende de tareas de
verificación (1.10, 1.11, 1.12, 1.3, 1.5) que todavía no se ejecutaron. Las specs dejan explícito
que, si esas tareas contradicen el supuesto, el plan vuelve al planner (tasks 2.4, escenario
"Exclusión no disponible").

Decisiones de recomendación que agregué y que el grupo debe confirmar: G9 (sin recomendación hasta
1.10), G10 (advertir a proyectos registrados, no reescribirlos) y G11 (entrega por capas sin binario
`claude`). Elegí G11 así para que CI y el Docker de `e2e/` ejerciten el camino nuevo; la alternativa
legado es igual de defendible.

## SUPOSICIÓN (no verificado)

- Que exista una variable de entorno de directorio de configuración y una de credencial que
  permitan correr `claude -p` sin copiar archivos (D9, tarea 1.12).
- Que el registro de contexto de la transcripción incluya el cuerpo del agente de sesión y el
  prompt del subagente (D8, tarea 1.11). La evidencia de partida es del plan-reviewer (conteos de
  tipos de entrada), no la reproduje.
- Delegación automática por `description` y subagentes sin herramienta Agent en 2.1.318 (B1).
- Que `update --sync` no tenga otro camino que reinstale config en proyectos registrados: leí
  `installArgs` y el uso de `state.Projects` en `syncRecordedInstall`, no el flujo completo de
  refresco de skills.
- No ejecuté ninguna sesión de Claude con modelo ni código del change (no existe).

## Fuentes leídas

`informes/propose-1-plan-reviewer.md` (completo); los seis artefactos del change; código listado en
EVIDENCIA; `claude --help`.

## Lecciones L-NNN

El repo no tiene base de lecciones ni `kb/` (declarado en el intento 1; no reverificado con listado
nuevo en este intento).

## Contaminación

Este intento parte del informe del plan-reviewer por encargo: tomé de ahí los hallazgos, la
evidencia de transcripciones (`attachment.type=instructions` y `prompt_snapshot`) y el commit
6a40f17 con su motivo. Reverifiqué en código M3 (`projectTargetAdapter`, `plan.go:138-140`,
`install.go:35`), M7 (`update.go`, `state.go`), m1 (`compose.go:93`) y m6 (`AGENTS.md:148-150`). No
reverifiqué 6a40f17 en `git log` ni las transcripciones. No consulté Engram.
