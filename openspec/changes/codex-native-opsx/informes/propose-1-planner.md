# propose-1 · planner · codex-native-opsx

- **Autor / rol**: planner (subagente nativo `role-planner`, Claude, modelo declarado por el entorno: claude-opus-5-5). El orquestador eligió el ejecutor; no hay tabla Role executors en el proyecto.
- **Revisión revisada**: `bd6354349ca6fce29d4e4786d75ca9a4aed7ea60`. HEAD del checkout: `1b4639b537a4d0c04dc8c25c0220732095a772d3`. `git diff --quiet bd63543 HEAD -- internal cmd` → idénticos. Las citas `archivo:línea` valen para ambos.
- **Artefactos entregados**:
  - `openspec/changes/codex-native-opsx/proposal.md`
  - `openspec/changes/codex-native-opsx/specs/codex-opsx-delivery/spec.md`
  - `openspec/changes/codex-native-opsx/specs/codex-sdd-migration/spec.md`
  - `openspec/changes/codex-native-opsx/design.md`
  - `openspec/changes/codex-native-opsx/tasks.md`
- **No hice**: código de producción, edición de `harnesses.yaml`, build, commit, ni edición de `informes/LEEME.md` (es del orquestador).

## EVIDENCIA

| # | Timestamp (UTC) | Comando | Salida relevante |
|---|---|---|---|
| E1 | 2026-11-06T13:06:53Z | `git rev-parse HEAD; openspec list --json; openspec status --change codex-native-opsx --json` | HEAD `1b4639b…`. Un único change `codex-native-opsx` con `no-tasks`. Artefactos: proposal `ready`, el resto `blocked`. |
| E2 | 2026-11-06 (≈13:07) | `git diff --quiet bd63543 HEAD -- internal cmd && echo identical` | `internal+cmd identical` |
| E3 | 2026-11-06 | `wc -lc internal/harness/config/assets/*/sdd-orchestrator.md internal/harness/config/assets/*.md` | codex 181 líneas / 6.844 B; governance 1.275 B; engram-protocol 6.091 B; strict-tdd 3.769 B. Con `compose.go:97-103` la suma da 6.844 + 1 + (1.275 + 2 + 6.091 + 2 + 26 + 3.769) = **18.010 B**, igual al informe ejecutivo. Es aritmética sobre tamaños medidos: no ejecuté `Compose`. |
| E4 | 2026-11-06T13:08:47Z | `openspec init --tools codex --no-animation .` en `scratchpad/osprobe/proj`, con `HOME`, `USERPROFILE`, `CODEX_HOME`, `APPDATA`, `LOCALAPPDATA` y `XDG_CONFIG_HOME` apuntando a un home temporal | "6 skills in .agents/"; "Commands skipped for: codex (uses skills)"; "Start your first change: $openspec-propose … (Codex CLI or IDE); in the Codex desktop app, select openspec-propose from Skills in the sidebar". Archivos creados: `.agents/skills/{openspec-apply-change,openspec-archive-change,openspec-explore,openspec-propose,openspec-sync-specs,openspec-update-change}/SKILL.md` y `.agents/skills/.openspec-target` (contenido `codex`). En el home temporal solo apareció `.config/openspec/config.json`. Frontmatter: `name`, `description`, `allowed-tools`, `metadata.generatedBy: "1.16.0"`. No hay `agents/openai.yaml`. |
| E5 | 2026-11-06 | `openspec --version` | `1.16.0` |
| E6 | 2026-11-06 | `which codex` | no encontrado (Codex CLI no instalado) |
| E7 | 2026-11-06 | `openspec list --specs` | 14 specs; ninguna del `sdd-orchestrator`. `design-mockups-harness` revisada en vista filtrada (`openspec show … --no-scenarios`). |
| E8 | 2026-11-06T13:18:13Z | `openspec status --change codex-native-opsx --json` | `isPlanningComplete: true`, `isComplete: true`; proposal, specs, design y tasks en `done`. |
| E9 | 2026-11-06T13:18:13Z | `openspec validate codex-native-opsx` y `openspec validate codex-native-opsx --strict` | `Change 'codex-native-opsx' is valid` en ambos, exit 0. |
| E10 | 2026-11-06T13:18:13Z | `openspec list --json` | `totalTasks: 34`, `completedTasks: 0`, `status: in-progress` |

Observaciones de código citadas en `design.md` (lectura directa):

- `adapter.go:18-20, :22-24, :26-28, :44-53, :64-66`
- `install.go:43-79`; `inject.go:34-72, :85-134`; `compose.go:34-106`; `primary_agent.go:43-122`
- `model/harness.go:68-89`
- `install/plan.go:141-160, :307-310, :556-575`
- `verify/harness_checks.go:97-131, :240-279`; `verify/adapter.go:13-37`
- `uninstall/plan.go:162-176, :229-241`
- `update.go:353-368, :370-413, :416-418`
- `backup/snapshot.go:117-160`; `backup/restore.go:35-104`
- `filemerge/section.go:28-30, :127-216`
- `assets/assets.go:12-21`; `assets/skills/openspec-propose/SKILL.md:3` (placeholder)
- `main.go:333-346`
- plantilla Codex `:3, :17-36, :55-67, :71-83, :87-92, :94, :100, :129-159, :169-181`
- `governance.md:3`; `engram-protocol.md:4`

## JUICIO

Autor: planner. Confianza indicada en cada punto.

1. **Contradicción central (alta).** La plantilla presupone un agente dedicado (`:3`), pero la entrega es `ConfigDeliveryInstructions` hacia `~/.codex/AGENTS.md`. Es la misma fuga que el código ya documenta y resolvió para OpenCode (`model/harness.go:70-76`). Propongo una `ConfigDelivery` propia para Codex (D1), siguiendo ese precedente.
2. **Instalado no es lo mismo que activado (alta).** Los `/opsx:*` que anuncia la plantilla (`:63-67`) no existen en Codex: el adapter no tiene directorio de comandos y OpenSpec 1.16.0 los omite para Codex (E4). Las skills `openspec-*` que la plantilla usa solo aparecen si el proyecto corrió `openspec init --tools codex`, y van a `<root>/.agents/skills`, no a `$HOME` (`:94`). Active Stack no las instala.
3. **Hallazgo lateral (alta).** La variante Codex asigna modelos de Claude (`opus`/`sonnet`/`haiku`, `:174-179`). Lo dejé como G9 sin decidirlo.
4. **Riesgo de rollback (alta).** `restore.go:37-40` hace NO-OP sobre directorios preexistentes. Escribir dentro de `~/.agents/skills` exige registrar cada archivo en el manifiesto (D8). Si no, un fallo deja piezas nuevas sin limpiar.
5. **Tensión de diseño con Engram y governance (media).** Sus fragmentos se declaran "ALWAYS ACTIVE" y "applies to all agents". Sacarlos de la sesión normal cambia comportamiento. Lo dejé como G4, no como hecho.
6. **Alcance de infraestructura (media-alta).** El diseño no requiere modificar `backup`, `filemerge` ni `pipeline`. Si la rama D6-b (registro en `config.toml`) resulta necesaria, el riesgo sube y el plan vuelve a revisión.
7. **Comparación con el brief.** Cubrí todos los escenarios pedidos:
   - sesión normal;
   - invocación explícita;
   - especialista;
   - TDD;
   - módulo faltante;
   - migración idempotente;
   - rollback con contenido del usuario;
   - update con y sin recibo.
   
   Además agregué propiedad de archivos, desinstalación y no regresión de otros agentes.
8. **Tensión con la instrucción del schema (media).** El schema pide resolver antes de las tareas las preguntas que cambian el plan. El brief prohíbe decidir los puntos de grupo. Los dejé como G1–G11 bloqueantes de `apply` (no como Open Questions), y las tareas que dependen de ellos están condicionadas. Lo tiene que juzgar el plan-reviewer.

## SUPOSICIÓN

No verificadas, con su motivo:

- Comportamiento de Codex: carga de `~/.codex/AGENTS.md` en sesión principal y en subagentes, rutas de descubrimiento de skills y roles, clave de `agents/openai.yaml` para deshabilitar la invocación implícita, parámetro de selección de rol al hacer spawn, feature flags, comportamiento por superficie (IDE, app, cloud). **Motivo**: Codex CLI no está instalado (E6) y el subagente no tiene web. Todo pasa al spike (tasks 1.x) con versión `(TBD, G8)`.
- Convención `name = "sdd_explorer"` y la rama D6-a o D6-b. **Motivo**: tomadas del informe del orquestador, no confirmadas en la documentación.
- Que `internal/update/skills.go` (`refreshSkills`) no trate `sdd-orchestrator/` como skill de catálogo. **Motivo**: no lo leí en detalle; queda como tarea 8.3.
- Que la extensión IDE comparta el home de Codex. **Motivo**: inferencia indirecta del mensaje de OpenSpec (E4).
- El total de 18.010 B no salió de ejecutar `Compose`: es suma de tamaños (E3). No corrí `go test` para no compilar, por la regla de no hacer build.
- No hay medición de tokens ni de consumo: depende de 02, que no existe.

## Fuentes leídas

- Instrucciones: `C:\tools\active-stack\CLAUDE.md` (inyectado); `CHANGES.md:1-204` (cómo trabajar, comportamiento objetivo, aceptación transversal, límites, fichas 01–06); `openspec/changes/codex-native-opsx/informes/LEEME.md`; `.openspec.yaml`; `openspec/config.yaml`.
- Código: `internal/agents/codex/adapter.go`, `adapter_test.go`, `internal/agents/interface.go`, `internal/agents/claude/adapter.go`; `internal/harness/config/{compose,install,inject,assets,primary_agent,types}.go`; assets `codex/sdd-orchestrator.md` (completo), `claude/sdd-orchestrator.md` (encabezados y grep), cabeceras de `strict-tdd.md`, `governance.md` y `engram-protocol.md`; `internal/catalog/harnesses.yaml:1-205`; `internal/install/steps.go` (completo), `internal/install/plan.go:130-340, :548-580`; `internal/verify/harness_checks.go` (completo), `codex_test.go`, `adapter.go`; `internal/uninstall/plan.go:150-254` y grep de `steps.go`; `cmd/active-stack/update.go:350-424` y grep; `cmd/active-stack/main.go:325-346`; `internal/filemerge/section.go` (completo) y lista de funciones; `internal/backup/snapshot.go:100-160`, `restore.go:20-107`; `internal/model/harness.go:60-95` y grep de `agentpaths.go`; `internal/harness/skill/embed.go`; `assets/assets.go`; `assets/skills/openspec-propose/SKILL.md` (cabecera).
- Specs y archivo: `openspec/specs/design-mockups-harness/spec.md:1-30`; `openspec/changes/archive/2026-11-06-codex-design-mockups/proposal.md:1-40`.
- Documentación: `ARCHITECTURE.md:60-82` y grep de Codex en `ARCHITECTURE.md`, `README.md` y `AGENTS.md`.
- Externas: la evidencia del brief del orquestador (informe ejecutivo y doc Codex rust-v0.165.0 vía Context7), usada como [DOC], sin acceso propio.

## Contaminación por otros informes

- No hay informes previos de roles en este change.
- Usé el **resumen del informe ejecutivo** y los **extractos de documentación de Codex** que trajo el brief del orquestador. Cada afirmación que proviene de ahí está marcada [DOC] o [SUP] en `design.md`, y las que pude volver a comprobar sobre `bd63543` (adapter.go:65 y :28; plantilla :3, :66, :80, :25; 18.010 B) quedaron como [EV] con mi propia cita.
- No consulté Engram ni conclusiones de otros revisores.

## Pendientes explícitos

- Aprobación humana en `proposal.md ## Aprobación` (vacía).
- Decisiones de grupo G1–G11 (`design.md`).
- Spike de Codex (tasks 1.x).
- Change 02 `opsx-baseline` inexistente, que bloquea 11.1.
- Revisión conjunta con 03.
