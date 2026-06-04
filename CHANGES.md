# CHANGES.md — Roadmap de construcción de JR Stack

> Backlog técnico canónico (Etapa 3 del MANUAL-METODOLÓGICO). Derivado del
> roadmap de incrementos de `ARCHITECTURE.md` §6 y del bucket mantener/modificar/sacar (§1.1).
> Fuente de verdad del **estado** real: el `openspec` CLI. Este documento es el plan.
>
> Convención de IDs: `C-NN`. Atomicidad: cada change implementable en una sesión, ≤ 12 tareas.

---

## Árbol de dependencias

```
C-01 esqueleto + fundación  [HECHO]
   │
C-02 modelo + catálogo  [HECHO]
   │
   ├──────────────┬───────────────┬───────────────┐
   ▼              ▼               ▼               ▼
C-03 system    C-04 filemerge  C-05 backup     C-06 planner
(detección)    (markers)       (snapshot)      (grafo deps)
   │              │               │               │
   │              └───────┬───────┘               │
   │                      ▼                        │
   │            ┌─────────┴──────────┐             │
   │            ▼                    ▼             │
   │      C-08 harness skill   C-09 harness config │
   │      (clone+copy)         (sdd-orchestrator)  │
   │            │                    │             │
   ▼            │                    │             │
C-07 harness external                │             │
(engram/openspec/ctx7)               │             │
   │            │                    │             │
   └────────────┴────────┬───────────┘             │
                         ▼                          │
                  C-10 agent adapters slim          │
                  (claude + opencode P0)            │
                         │                          │
                         └──────────┬───────────────┘
                                    ▼
                          C-11a install-pipeline
                          (orquestación headless + rollback)
                                    │
                                    ▼
                          C-11b TUI flujo install
                          (lite/full/custom, Bubbletea slim)
                            │            │
                            ▼            ▼
                   C-12 uninstall   C-13 jr-orchestrator
                   harness-aware    (orquestador fundación)
                            │            │
                            └─────┬──────┘
                                  ▼
                          C-14 verify + E2E
```

> Notas de dependencia:
> - C-07/C-08/C-09 son los tres harness installers; comparten `harness/` pero
>   tocan subpaquetes distintos (`external/`, `skill/`, `config/`) → paralelizables.
> - C-08 y C-09 dependen de `filemerge`/`backup` porque escriben config del usuario.
> - C-10 (adapters) necesita que exista al menos un installer que los consuma; arranca P0 con claude+opencode.
> - C-11 (TUI) integra catálogo + planner + installers + adapters → es el gran punto de convergencia.

### Epic: Starter packs (capa proyecto) — rama desde C-02

```
C-02 modelo + catálogo  [HECHO]
   │
   ▼
C-26 modelo de starters  [HECHO]
   │
   ├───────────────────────────────┐
   ▼                               ▼
C-27 target proyecto           C-30 publicar repos
(adapters → .claude/ proj)     por starter + curar
   │                               │
   ▼                               │
C-28 mcp scope proyecto            │
(.mcp.json idempotente)            │
   │                               │
   ▼                               │
C-29 subcomando                    │
jr-stack starter add               │
   │              │                │
   ▼              └────────┬───────┘
C-31 slash               ▼
command            C-32 verify + E2E
(invoca binario)   starters
```

> Notas del epic starters:
> - El epic cuelga del sustrato YA TERMINADO (C-01..C-25): reusa backup/filemerge/pipeline/verify.
> - C-27 es **governance ALTO** (escribe config del usuario en el proyecto) → propone-y-espera-review.
> - C-30 (publicar repos) es paralelizable con C-27/C-28/C-29, pero **debe** estar antes de C-32 (el E2E clona los repos reales).
> - ~~C-28 tiene incógnita abierta: verificar dónde escribe hoy el installer `external`/`method:mcp` antes de tocarlo.~~ → **RESUELTO** (C-28 HECHO): la infra (`installMCP` con backup+merge+atomic) ya existía; el MCP project-scope de Claude va a `<root>/.mcp.json` en la raíz (verificado contra doc oficial), no a `.claude/mcp/`.

---

## Plan óptimo de paralelización

| Ola | Changes en paralelo | Bloquea a | Comentario |
|---|---|---|---|
| 0 | C-01, C-02 | todo | **HECHO** (fundación + dominio) |
| 1 | C-03, C-04, C-05, C-06 | olas 2–4 | **HECHO** — 4 ports independientes del repo viejo |
| 2 | C-07, C-08, C-09, C-09b | C-10 | **HECHO** — 4 installers (external/skill/config/permissions) |
| 3 | C-10 | C-11a | **HECHO** — adapters slim (P0 claude+opencode) |
| 4 | C-11a | C-11b | **HECHO** — install-pipeline: orquestación headless + rollback |
| 5 | C-11b | C-12, C-13 | **HECHO** — TUI Bubbletea slim: punto de convergencia visible |
| 6 | C-12, C-13 | C-14 | **HECHO** — uninstall + orquestador de fundación (C-13 split a/b/c) |
| 7 | C-14 | — | **HECHO** — verify + E2E (cierre del pipeline) |
| 8 | C-26 | C-27, C-30 | **HECHO** — modelo de starters (base del epic starter packs) |
| 9 | C-27, C-30 | C-28/C-29, C-32 | C-27 target proyecto (ALTO) ‖ C-30 publicar repos por starter |
| 10 | C-28 | C-29 | **HECHO** — mcp scope proyecto (.mcp.json raíz para Claude) |
| 11 | C-29 | C-31, C-32 | **HECHO** — subcomando `jr-stack starter add` |
| 12 | C-31, C-32 | — | **HECHO** — C-31 (slash-command) ‖ C-32 verify + E2E starters (**cierre del epic starter packs**) |

**Camino crítico (instalador, HECHO)**: `C-01 → C-02 → C-04/C-05 → C-09 → C-10 → C-11a → C-11b → C-13 → C-14`.
Es la cadena más larga: el merge/backup habilitan el installer de config
(sdd-orchestrator), que alimenta los adapters, que alimentan el install-pipeline
headless, que monta la TUI, que habilita el orquestador de fundación, que se
cierra con verify+E2E.

**Camino crítico (epic starter packs)**: `C-26 → C-27 → C-28 → C-29 → C-32`.
C-30 (publicar repos) corre en paralelo pero bloquea el E2E final.

**Máximo paralelismo útil**: 4 agentes (ola 1). A partir de ahí el grafo se
estrecha hacia la convergencia en C-11b (la TUI).

---

## Fichas de change

### C-01 — Esqueleto + fundación
- **Estado**: HECHO (repo, go.mod, .gitignore, ARCHITECTURE.md, openspec init, AGENTS.md/CLAUDE.md/CHANGES.md).
- **Scope**: Estructura base del repo Go, módulo, gitignore, documentos de fundación de la metodología.
- **Dependencias**: ninguna.
- **Governance**: BAJO.
- **Leer antes**: `ARCHITECTURE.md`, `../MANUAL-METODOLOGICO.md` (Etapas 3 y 4).

### C-02 — Modelo + catálogo
- **Estado**: HECHO (tests en verde).
- **Scope**: Tipos de dominio (`Harness`, `HarnessType`, `InstallMode`, `Agent`, `Source`, `External`) en `internal/model`; parser+validador del `harnesses.yaml` embebido en `internal/catalog`; catálogo inicial de harnesses.
- **Dependencias**: C-01.
- **Governance**: BAJO.
- **Leer antes**: `ARCHITECTURE.md` §2–3, `internal/model/harness.go`, `internal/catalog/`.

### C-03 — Port `system` (detección OS/deps)
- **Estado**: HECHO (archivado `2026-05-26-c03-system-port`).
- **Scope**: Portar `internal/system` del repo viejo: detección OS/arch/WSL/Termux, chequeo de dependencias, guards de plataforma. Limpiar leftovers `gentle-ai`/`Gentleman.Dots`.
- **Dependencias**: C-02.
- **Governance**: BAJO (solo lee el sistema; no escribe config).
- **Leer antes**: repo viejo `internal/system/`, ARCHITECTURE.md §5.

### C-04 — Port `filemerge` (markers)
- **Estado**: HECHO (archivado `2026-05-26-c04-filemerge`).
- **Scope**: Portar `internal/filemerge`: merge por markers idempotente (inyectar bloques sin pisar config del usuario, sin duplicar al reinstalar). Limpiar branding viejo.
- **Dependencias**: C-02.
- **Governance**: **ALTO** (puede destruir/corromper config del usuario).
- **Leer antes**: repo viejo `internal/components/filemerge/`, regla "SIEMPRE markers idempotentes" en AGENTS.md §3.

### C-05 — Port `backup` / rollback
- **Estado**: HECHO (archivado `2026-05-26-c05-backup-port`).
- **Scope**: Portar `internal/backup`: snapshot + restore de configs, manifest, compresión, retención. Es la red de seguridad antes de cualquier escritura.
- **Dependencias**: C-02.
- **Governance**: **ALTO** (snapshot/restore de la config del usuario).
- **Leer antes**: repo viejo `internal/backup/` (manifest, restore, compression, retention).

### C-06 — Port `planner` (grafo de deps)
- **Estado**: HECHO (archivado `2026-05-26-c06-planner`).
- **Scope**: Portar `internal/planner`: resolución del grafo de dependencias entre harnesses (`DependsOn`), orden topológico de instalación, payload de review. Adaptar de "componentes" a "harnesses".
- **Dependencias**: C-02.
- **Governance**: BAJO.
- **Leer antes**: repo viejo `internal/planner/`, modelo `Harness.DependsOn`.

### C-07 — Harness installer `external` (engram/openspec/context7)
- **Estado**: HECHO (archivado `2026-05-26-c07-external`).
- **Scope**: `internal/harness/external`: instalar/configurar tools de terceros según `External.Method` (npm, homebrew, mcp). Cubre OpenSpec CLI, Engram, Context7.
- **Dependencias**: C-03 (detección de deps/OS para elegir método).
- **Governance**: MEDIO.
- **Leer antes**: `internal/catalog/harnesses.yaml` (entradas external), repo viejo `internal/components/engram` y `mcp`, ARCHITECTURE.md §2.

### C-08 — Harness installer `skill` (clone + copy)
- **Estado**: HECHO (archivado `2026-05-26-c08-skill-harness-installer`).
- **Scope**: `internal/harness/skill`: clonar repo de la skill según `Source{Repo,Ref}` y copiar al dir de skills de cada agente (vía adapter). Idempotente, con backup previo.
- **Dependencias**: C-04, C-05.
- **Governance**: MEDIO.
- **Leer antes**: repo viejo `internal/components/skills/` y `internal/assets/skills/`, modelo `Harness.Source`.

### C-09 — Harness installer `config` (sdd-orchestrator componible + permissions)
- **Estado**: HECHO. Se ejecutó en dos changes por vector de escritura distinto:
  `2026-05-26-harness-config-installer` (sdd-orchestrator, inyección markdown por markers)
  y `2026-05-26-c09b-permissions-harness` (permissions, JSON-merge sobre settings.json).
- **Scope**: `internal/harness/config`: componer el bloque `sdd-orchestrator` a partir de toggles (`tdd`, `engram`, `model-routing`, `delegation`, `governance`) e inyectarlo vía markers; instalar `permissions` (security-first, no opcional).
- **Dependencias**: C-04, C-05.
- **Governance**: **ALTO** (escribe config del usuario; el merge mal hecho corrompe CLAUDE.md/AGENTS.md).
- **Leer antes**: repo viejo `internal/assets/*/sdd-orchestrator.md` y `internal/components/permissions/`, ARCHITECTURE.md §2.1.

### C-10 — Agent adapters slim (claude + opencode primero)
- **Estado**: HECHO (archivado `2026-05-27-c10-agent-adapters`). Adapter público = unión de los 4 contratos de installers; ISP preservado vía aserciones compile-time.
- **Scope**: Portar `internal/agents` recortado: resolución de paths de config y dir de skills por agente. **P0: claude + opencode**; resto (gemini/codex/cursor/vscode/windsurf/antigravity) después. Sin lógica de theme/persona-marketing.
- **Dependencias**: C-07, C-08, C-09.
- **Governance**: MEDIO.
- **Leer antes**: repo viejo `internal/agents/{interface.go,factory.go,registry.go,claude,opencode}`, regla "NUNCA hardcodear paths de agente".

> **Nota de split (2026-05-27)**: C-11 excedía la atomicidad (≤12 tareas / una
> sesión): integraba catálogo + planner + 4 installers + backup + pipeline +
> inyección + verify + Bubbletea. Se dividió en **C-11a** (orquestación headless,
> testeable sin TUI) y **C-11b** (la cáscara Bubbletea encima). El repo viejo ya
> separaba `pipeline` de `tui` por la misma razón.

### C-11a — Install-pipeline (orquestación headless + rollback)
- **Estado**: HECHO (archivado `2026-05-27-c11a-install-pipeline`). Camino de rollback testeado explícitamente (harness exitoso + harness que falla → restore disparado). Spec en `openspec/specs/install-pipeline/`.
- **Scope**: Portar `internal/pipeline` (Step/RollbackStep, Runner, Orchestrator, rollback en reversa) + nuevo `internal/install` que cablea el flujo headless: detección → catálogo (`ForMode`/`ForAgent`) → planner (orden topológico) → backup → installers (external/skill/config/permissions) → inyección → verify-hook. Backup-first como Step explícito; cada paso de escritura implementa `Rollback()`. `harnessStep` por tipo envuelve cada installer sin reabrir paquetes de governance ALTO. Verify = hook opcional (impl real en C-14). Progreso vía `pipeline.ProgressFunc` (contrato para C-11b).
- **Dependencias**: C-10 (adapters), C-07/C-08/C-09 (installers), C-05 (backup).
- **Governance**: **ALTO** (orquesta escritura de config del usuario + rollback).
- **Leer antes**: repo viejo `internal/pipeline/`, ARCHITECTURE.md §4.1, firmas reales de catalog/planner/backup/agents/installers.

### C-11b — TUI flujo install (lite/full/custom, Bubbletea slim)
- **Estado**: HECHO (commit `8472cf6` — interactive install flow con Bubbletea).
- **Scope**: Portar `internal/tui` (Bubbletea) **slim**, sin theme cosmético. Flujo: detectar OS/agentes → elegir agente(s) → elegir modo (Lite/Full/Custom) → invocar `internal/install` (C-11a) → render de progreso suscrito al `ProgressFunc`. Selección/agrupación **por harness** (no por componente viejo). Se descarta theme/statusline/keybindings/persona-marketing.
- **Dependencias**: C-11a.
- **Governance**: BAJO (la TUI orquesta; el riesgo vive en el pipeline que invoca).
- **Leer antes**: ARCHITECTURE.md §4.1 (flujo install), repo viejo `internal/tui/`, skill `go-testing` (teatest).

### C-12 — Flujo uninstall harness-aware
- **Estado**: HECHO (archivado `2026-05-27-c12-uninstall`, commit `392984b`). Engine headless `internal/uninstall/` espejo de `internal/install/`, 28 tests en verde. Verificado PASS (governance ALTO).
- **Scope**: Uninstall que entiende **harnesses** (no componentes viejos): revertir inyecciones por markers, restaurar desde backup, remover skills clonadas. Mantener la interfaz de uninstall de la TUI.
- **Dependencias**: C-11.
- **Governance**: **ALTO** (restaura/borra config del usuario).
- **Leer antes**: repo viejo flujo uninstall + `internal/backup/restore.go`, ARCHITECTURE.md §1.1 (nota uninstall).

### C-13 — jr-orchestrator como orquestador de fundación
- **Estado**: HECHO. Se dividió en tres changes por atomicidad (cada skill su contrato):
  - **C-13a** — `jr-orchestrator` thin orchestrator from scratch + congelado del contrato modular (`.jr-starter-state.json` v2 + matriz I/O). Archivado `2026-05-27-c13-jr-starter`.
  - **C-13b** — `kb-creator` integra el contrato (owner del slice `state.kb`, discovery interactiva). Archivado `2026-05-27-c13b-kb-creator`.
  - **C-13c** — `roadmap-generator` integra el contrato (`state.roadmap`, standalone-safe). Archivado `2026-05-27-c13c-roadmap-generator`, verificado PASS.
- **Cierre**: rename `jr-starter` → `jr-orchestrator` (commit `f89c7e8`) + publicación pública de los 3 repos (`JuanCruzRobledo/{jr-orchestrator,kb-creator,roadmap-generator}`). Flujo de fundación instalable end-to-end.
- **Scope**: Crear la skill `jr-orchestrator` (thin orchestrator from scratch) con lazy-loading (ARCHITECTURE.md §4.2): `openspec init` → `kb-creator` → `roadmap-generator` → `find-skill` → `agent-instruction`.
- **Dependencias**: C-11.
- **Governance**: MEDIO.
- **Leer antes**: ARCHITECTURE.md §4.2, MANUAL Etapas 1–4, catálogo (skills propias).

### C-14 — Verify + E2E
- **Estado**: HECHO (split A/B). Cierra el pipeline.
  - **Wave A** — `internal/verify` harness-aware: motor de checks + report + hook wireado al pipeline (el `VerifyHook` quedaba `nil` desde C-11a). Commit `ee26282`.
  - **Wave B** — modo headless del binario (resuelve D4: `jr-stack install --headless --mode --agent --custom --dry-run --yes --home`, sin flags → TUI, cero regresión; commit `d34bb8b`) + suite E2E harness-first con matrix Docker Ubuntu+Arch (Go 1.26). **Tier 1 VERDE 15/15 en ambas distros.** Commit `6e3cd4c`.
- **Hallazgos del E2E** (el install real destapó bugs que los unit tests ocultaban con fakes):
  - **Arreglados** (commit `fddde4d`): panic nil-runner al clonar skills (`skillStep` sin `runner` → inyectado vía `Options.CmdRunner`) + URL de asset con `goos` vacío (`engram_<v>__amd64` → HTTP 404; `Options.Profile` con `runtime.GOOS`).
  - **Follow-ups**: **(C-15)** → HECHO. **(C-16)** → HECHO. **(C-17/C-18/C-19)** → HECHO (cerrados en el Tier 2/3 del E2E real, ver fichas).
- **Dependencias**: C-12, C-13.
- **Governance**: BAJO (verify/E2E); los fixes tocaron `install`/`harness` (MEDIO).
- **Leer antes**: repo viejo `internal/verify/` y `e2e/`, ARCHITECTURE.md §6 punto 8.

### C-16 — Skill clone: SKILL.md en la raíz del repo
- **Estado**: HECHO (commit `ce348b5`, archivado `2026-05-29-c16-skill-clone-layout`).
- **Scope**: `skill/clone.go` esperaba un subdir `tempDir/<skillID>`; la convención real es SKILL.md en la raíz del repo. Fix root-first con fallback a subdir + `copyDir` excluye `.git/` (preserva `.gitignore`). Verificado en E2E: las 5 skills clone instalan.
- **Governance**: MEDIO.

### C-15 — Pre-flight dependency gate + Node en imágenes E2E
- **Estado**: HECHO (commit `feat(install)` C-15, archivar). Cierra el clone+externals del camino `lite`.
- **Scope**: `system.RequiredDependencies` deriva deps según harnesses elegidos; gate en headless+TUI que aborta temprano (sin rollback a mitad) con `InstallHint` si falta una dep required; `npx` agregado a `deps.go`; Node en `Dockerfile.ubuntu`/`arch`. Verificado: `lite+claude` pasa completo.
- **Governance**: MEDIO (gate) + BAJO (Dockerfiles).

### C-17 — Rollback robusto ante directorios no vacíos
- **Estado**: HECHO (commit `53e22e8` — `fix(backup): rollback consciente de directorios`, archivado `c17-rollback-dir-aware`). **Gravedad ALTA** (governance ALTO — propuesto y aprobado antes de tocar). Verificado en E2E real: el rollback no destruye config preexistente.
- **Scope**: el rollback de un step fallaba con `remove "<dir>/skills": directory not empty` cuando otro step llenaba el dir compartido. Fix dir-aware vía nuevo `IsDir bool` en `ManifestEntry`: el rollback distingue dirs preexistentes de dirs creados por el install (el fix obvio `RemoveAll` a secas habría borrado skills del usuario). Un rollback NUNCA debe fallar dejando estado inconsistente.
- **Governance**: **ALTO**.

### C-18 — verify-hook: clave `permissions` ausente en opencode.json
- **Estado**: HECHO (commit `b13d577` — `fix(verify): clave de permisos por agente en el check`). Verificado en E2E real: opencode verde.
- **Scope**: el check `permissions:permissions:opencode` fallaba (`"permissions" key not found in opencode.json`) — el verify hardcodeaba la clave `"permissions"` (plural) para todos los agentes, pero opencode escribe `"permission"` (singular). Fix: nuevo helper `permissionsKeyFor(agent)` en `internal/verify/harness_checks.go` que resuelve la clave JSON por agente.
- **Governance**: MEDIO.

### C-19 — harness best-effort (degradar sin abortar)
- **Estado**: HECHO (commit `355a378` — `feat(install): harness best-effort (degradar sin abortar)`, strict TDD, `internal/pipeline/` intacto). Verificado en E2E real: `full` cierra.
- **Scope**: `find-skill` y `skill-creator` usan `method: npx` con empaquetado upstream TBD — `npx skills add` da `exit status 1` y abortaba el install entero. Fix: estos harness degradan con warning en vez de tumbar el pipeline. El TBD de empaquetado de terceros sigue abierto (no es bug nuestro).
- **Governance**: MEDIO. Relacionado con el TBD "Empaquetado de skills de terceros".

---

## Decisiones tomadas (antes TBD) — 2026-05-30

- ~~**Nombre del repo remoto en GitHub**.~~ → **RESUELTO**: el repo viejo queda **archivado como legacy**; este conserva el nombre **`jr-stack`** (binario y repo). Doc-only.
- ~~**Qué harness entra en cada modo** Lite/Full/Custom.~~ → **RESUELTO** (ver **C-20**, HECHO): **Lite = sustrato** (openspec, engram, context7, sdd-orchestrator, permissions); **Full = sustrato + fundación guiada** (jr-orchestrator + skills que orquesta); **Custom = todos**. `jr-orchestrator` se movió a Full-only (en Lite quedaba huérfano: orquesta skills que no se instalaban).
- ~~**Granularidad de toggles** de `sdd-orchestrator`.~~ → **RESUELTO**: los 5 toggles (`tdd`, `engram`, `model-routing`, `delegation`, `governance`) son la granularidad final (ya implementada en C-09; confirmado en ARCHITECTURE.md §7 "Resueltas").
- ~~**Repos de skills propias** `jr-starter` y `skill-registry`.~~ → **RESUELTO** (`JuanCruzRobledo/jr-orchestrator` y `JuanCruzRobledo/skill-registry`, ambos públicos).
- ~~**Empaquetado de skills de terceros**.~~ → **RESUELTO** (ver **C-22**, HECHO): se **abandonó `npx`** (CLI de terceros interactivo y de contrato inestable) en favor de **`git clone` + copia de subdir** (campo `Source.Path`). `find-skill` y `skill-creator` clonan su repo upstream y copian el `SKILL.md` del subdir correspondiente; siguen `best_effort: true` (C-19) por si el upstream cambia. El comando `npx skills add --skill` quedó **descartado**, no pendiente.
- ~~**Overlay de permisos de Windsurf**.~~ → **RESUELTO**: **no-op explícito** (decisión firme, antes TBD en design.md). Windsurf gestiona permisos vía la UI de la IDE (Cascade), sin schema de `settings.json` inyectable conocido — mismo criterio que Cursor/Codex/Antigravity. Materializado como `case model.AgentWindsurf` explícito en `internal/harness/config/permissions/overlays.go` (antes caía en el `default`) + test de caracterización `TestInstallWindsurfNoOp`. El `default` queda solo para agentes futuros no contemplados.

## Pendientes de implementación (decididos, falta código)

### C-20 — Mapeo de modos: jr-orchestrator a Full-only
- **Estado**: HECHO (catálogo + test `TestForMode_JROrchestratorIsFullOnly`, `go test ./internal/catalog/` verde). Governance BAJO.
- **Scope**: `harnesses.yaml` — `jr-orchestrator` de `[lite, full]` a `[full]`. Lite = sustrato; Full = fundación guiada.

### C-21 — Custom: `permissions` NO desactivable
- **Estado**: HECHO (strict TDD; `go test -count=1 ./internal/...` verde, `go vet` limpio). Governance MEDIO, revisado por el operador en fresco. En modo Custom `permissions` (security-first) queda forzado, no se puede desmarcar (CLAUDE.md §1).
- **Arquitectura (defensa en profundidad)**: la garantía vive en el **chokepoint** `selectHarnesses` (`internal/install/plan.go`), por donde pasan TANTO la TUI como el headless (`--mode custom`): si `permissions` no está en `intent.Custom`, se fuerza al set (salvo que el agente no lo soporte → `filterByAgents` lo descarta, overlay inexistente). Espejo en `selectTUIHarnesses` (`gate.go`) para coherencia del preflight.
- **UX**: el picker (`model.go`) arranca con `permissions` seleccionado, el toggle lo ignora, y el render lo muestra `[x] permissions (requerido — security-first)`. No miente: no hay un toggle muerto sin señal.
- **Scope**: `internal/install/plan.go`, `internal/tui/{gate.go,model.go,screen.go}` + tests en `plan_test.go`, `gate_test.go`, `model_test.go`.

### C-22 — Skill installer terceros: npx → git clone + subdir
- **Estado**: HECHO (strict TDD; `go test ./internal/...` 15/15 verde, `go vet` limpio). Governance MEDIO, revisado por el operador (tests frescos `-count=1`). Cierra el TBD de empaquetado de terceros.
- **Hallazgo**: el plan inicial era arreglar el comando `npx`. La investigación destapó que el CLI `npx skills` tiene contrato **contradictorio entre fuentes** (`skills add` vs `skills i` vs interactivo puro) y es fundamentalmente interactivo. Aunque existe modo headless (`--skill -y -a <agent>`), atar el instalador a un CLI de terceros inestable es frágil, y el destino se controla con `-g`/`-a`, no con un path (rompe el modelo de adapter + tests `--home`). **Decisión: abandonar npx, usar `git clone` + copiar subdir** (mecanismo C-16, probado headless).
- **Scope**: campo `Path string` en `model.Source` (subdir del SKILL.md dentro del repo; vacío = raíz, comportamiento C-16). `clone.go`: si `Path != ""` copia desde `<temp>/<path>` sin fallback. Catálogo: `find-skill`→`{repo: vercel-labs/skills, method: clone, path: skills/find-skills}` (skill upstream = `find-skills`, plural) y `skill-creator`→`{repo: anthropics/skills, method: clone, path: skills/skill-creator}`; ambos siguen `best_effort: true`. ID del harness sin cambios. `npx.go` queda huérfano (sin uso) → limpieza en **C-23**.

### C-24 — Unificar la selección de harnesses (regla security-first en un solo lugar)
- **Estado**: HECHO (strict TDD; `go test -count=1 ./...` verde, `go vet` limpio). Governance MEDIO, revisado por el operador en fresco. La regla "forzar `permissions` en Custom" vivía **duplicada en 4 lugares**; ahora vive en UNO.
- **Solución**: fuente única en `internal/install/plan.go` — `const SecurityFirstHarnessID = "permissions"` + función exportada `SelectHarnesses(cat, intent) ([]Harness, error)` (canónica, estricta: error en id desconocido). `tui.selectTUIHarnesses` (gate.go) y `cmd…collectSelectedHarnesses` (main.go) quedan como wrappers finos que delegan; el picker (`model.go`) referencia `install.SecurityFirstHarnessID`. `filterHarnessesByAgents` huérfano de main.go borrado.
- **Verificado**: grep confirma `ByID(SecurityFirstHarnessID)` sólo en `plan.go:130`. Sin ciclos de import. Tests: `select_harnesses_test.go`, `unified_selection_test.go` (afirma que TUI e install resuelven el mismo set).

### C-23 — Limpiar el método `npx` huérfano
- **Estado**: HECHO (strict TDD; `go test -count=1 ./...` verde, `go vet` limpio). `npx` como **método de instalación** = cero referencias; el contrato del installer queda `clone | embed`. La cola `npx` como **dependencia de sistema** quedó cerrada en **C-25** (HECHO): tampoco era necesaria.
- **Scope**: borrado `internal/harness/skill/npx.go` + casos npx en `skill_test.go`; quitado el `case "npx"` de `installer.go` (doc y error de método pasan a `clone, embed`); en `catalog.go` la inferencia third-party pasa de `npx` a `clone` y `npx` sale de los métodos válidos; `TestMethodInference_ThirdParty` ahora espera `clone`; comentario `clone | npx | embed` → `clone | embed` en `model.Source`. find-skill/skill-creator sin cambios (ya usan clone explícito).

### C-25 — Limpiar `npx` en la capa de pre-flight / dependencias  *(HECHO)*
- **Estado**: HECHO (strict TDD; `go test -count=1 ./...` verde, `go vet` limpio). Removido todo rastro de `npx` como dependencia/método en la capa de pre-flight; el contrato de deps por-harness queda `external npm → node, npm` y `skill clone → git`.
- **Confirmación previa (resuelta)**: ningún `external` con `method: npm` necesita `npx` en runtime. Verificado en `internal/harness/external/npm.go` (install = `npm install -g <pkg>`, sin `npx`) y `verify.go` (`lookPath` del **binario propio** del paquete, ej. `openspec`, sin `npx`). El catálogo además **rechaza** `method: npx` (`catalog_test.go` "skill with removed npx method" → `unknown source.method`), confirmando que el `case "npx"` era inalcanzable.
- **Matiz del hallazgo**: el gate de install (`headless/executor.go`, `tui/gate.go`) consume el **derived set** (`RequiredDependencies` → `DetectDepsFor`), que nunca incluía `npx`. La dep `npx` (`Required:true`) solo vivía en la lista global `defineDependencies`, consumida por el **diagnóstico** (`detect.go`). Por eso quitarla **no cambia el gate de install** — solo elimina un requisito fantasma del diagnóstico.
- **Scope ejecutado**:
  - `internal/system/preflight.go` → borrado el `case "npx"` muerto del switch de skills + actualizado el comentario de mapeo.
  - `internal/system/deps.go` → removida la dep `npx` de `defineDependencies`.
  - `internal/system/preflight_test.go` → eliminados `TestRequiredDependencies_SkillNpx_NodeNpmNpx` y `TestRequiredDependencies_MixedNpxAndConfig_NpxSet` (camino muerto); `TestDefineDependencies_IncludesNpx` invertido a `TestDefineDependencies_ExcludesNpx` (afirma el contrato nuevo).
  - `internal/tui/gate.go` → comentario `node/npm/npx/git` → `node/npm/git`.
  - **No tocado**: `catalog.go` / `catalog_test.go` (comentarios + guard que rechaza `method:npx` son correctos y se conservan como red de seguridad).
- **Governance**: MEDIO. TDD estricto (RED: `ExcludesNpx` falla con `npx` aún presente → GREEN tras la limpieza).

### C-26 — Modelo de starters (base de starter packs)
- **Estado**: HECHO (strict TDD; `go test -count=1 ./...` verde, `go vet` limpio). 12/12 tasks completadas.
- **Scope**: Nueva capa de curación de dominio (estilo Spring Boot starters) sobre los harnesses existentes.
  - Nuevo tipo `model.Starter` (`internal/model/starter.go`) con `ID`, `Name`, `Description`, `Harnesses []string`, `Includes []string`, `MCPs []MCP` (placeholder, shape definitivo en C-28) y método `Validate()`.
  - Sección `starters:` en `internal/catalog/harnesses.yaml` con 3 entradas seed: `active-ia` (compone `ux-ui + backend`), `ux-ui`, `backend`. Entradas placeholder; curación real de contenido en C-30.
  - `catalog.Catalog` extendido con `Starters []model.Starter` + `starterIndex`; `validateStarters()` invocada desde `validate()` con las mismas garantías que harnesses: id no vacío, sin duplicados, harnesses referenciados existentes, includes existentes, **detección de ciclos DFS tri-estado** (autorreferencia + ciclo indirecto ambos cubiertos).
  - Nuevos métodos públicos: `StarterByID(id)` (lookup O(1)) y `ResolveStarter(id)` (expansión recursiva deduplicada, preserva orden de primera aparición, determinista).
- **Decisiones clave**:
  - `Starter` es un tipo NUEVO e independiente, no un `Harness` sobrecargado (D1): contratos y campos son distintos; mezclarlos contaminaría el planner y violaría SRP.
  - DFS de validación (tri-estado, devuelve error) y DFS de resolución (set de visitados, acumula harnesses) son estructuralmente similares pero sirven contratos distintos; no se extrajo helper para evitar sobre-ingeniería (D4).
  - `MCPs []MCP` marcado `(TBD)` — el shape definitivo (`Command`, `Args`, `Transport`, etc.) se decide en C-28 al implementar escritura de `.mcp.json` (D5).
- **Dependencias**: C-02 (modelo + catálogo base). Habilita: C-27 (target proyecto en adapters), C-28 (escritura .mcp.json), C-29 (subcomando `jr-stack starter add`), C-30 (publicar repos por starter), C-31 (slash-command), C-32 (verify + E2E starters).
- **Governance**: BAJO (model/catalog). TDD estricto — 3 archivos nuevos, 12 tests nuevos, 0 regressions.
- **Archivos tocados**: `internal/model/starter.go` (nuevo), `internal/model/starter_test.go` (nuevo), `internal/catalog/catalog.go` (extendido), `internal/catalog/harnesses.yaml` (sección starters:), `internal/catalog/starter_test.go` (nuevo).

### C-27 — Target de proyecto en adapters + pipeline
- **Estado**: HECHO (strict TDD; `go test -count=1 ./...` verde 21/21 paquetes, `go vet` limpio). 12/12 tasks completadas.
- **Scope**: Infraestructura de "target de proyecto" para el epic starter packs. Habilita escribir harnesses bajo `<proyecto>/.claude/` (o `.opencode/`) en lugar del home. NO expone comando nuevo (eso es C-29).
  - Nuevo tipo `model.InstallTarget` (`Machine | Project`, zero-value = `Machine`) en `internal/model/installtarget.go`. La garantía de zero-value = Machine asegura cero regresión en todos los call-sites existentes.
  - Nuevo tipo `model.AgentPaths` en `internal/model/agentpaths.go`: struct con `InstructionsPath`, `SkillsDir`, `SettingsPath` y método `MCPConfigPath(serverName)` via closure. Vive en `model` (no en `agents`) para evitar el ciclo de import `agents/claude → agents`.
  - Nuevo método `PathsFor(base string, t model.InstallTarget) model.AgentPaths` en la interfaz `agents.Adapter` (y en `install.AgentAdapter`). Claude: mismo layout `.claude/` para Machine y Project (solo cambia la base). OpenCode: Machine → `.config/opencode/` (XDG), Project → `.opencode/` (project-local, confirmado contra docs oficiales de OpenCode).
  - `install.Options` gana `Target model.InstallTarget` y `ProjectRoot string`. `BuildPlan` resuelve la base efectiva (`Machine → HomeDir`, `Project → ProjectRoot`) y la propaga a `collectWritePaths`, `buildHarnessStep` y al cálculo del snapshot dir y skill backup dirs.
  - `collectWritePaths` es ahora target-aware: para `Project` usa `PathsFor()` en vez de los métodos individuales. Los 4 helper internos (`resolvedInstructionsPath`, `resolvedSkillsDir`, `resolvedSettingsPath`, `resolvedMCPConfigPath`) encapsulan la bifurcación.
  - Snapshot de proyecto bajo `<ProjectRoot>/.jr-stack/backups/install`. Reusa `internal/backup` con `CreateWithDirHints` exactamente igual que el flujo de máquina. Skills dirs registradas como DirHints para rollback dir-aware (RemoveAll si fue creado por el install, NO-OP si preexistía).
  - Nuevo testseam `SetSnapshotCreateWithHints` en `install/testseams.go` para tests que inspeccionan DirHints.
- **TBD resueltos**:
  - D2 (layout OpenCode proyecto): confirmado `.opencode/` contra docs oficiales de OpenCode (`opencode.ai/docs/config/`). Machine sigue usando `.config/opencode/` (XDG). Diferencia encapsulada en el adapter.
  - Firma del método target-aware: `PathsFor(base string, t model.InstallTarget) model.AgentPaths` — un solo método aditivo (opción elegida de D1), sin multiplicar métodos ni tocar las 4 interfaces de installer (ISP preservado).
- **TBD pendientes**:
  - D1: adapter con estado (`TargetRoot` inyectado, alternativa B) diferido a C-29 si el parámetro resulta incómodo al cablear.
- **Invariante verificado**: cero regresión del flujo de máquina. Los 7 métodos existentes de cada adapter (`Agent`, `InstructionsPath`, `SkillsDir`, `SettingsPath`, `MCPConfigPath`, `MCPStrategy`, `VariantKey`) conservan exactamente su salida anterior. Tests explícitos de regresión en `adapter_regression_test.go` para claude y opencode. Tests `TestBuildPlanMachineTarget_*` verifican que `BuildPlan` sin Target produce paths y snapshot dir idénticos a hoy.
- **Dependencias**: C-26 (modelo starters, base del epic). Habilita: C-28 (mcp scope proyecto), C-29 (subcomando `jr-stack starter add`).
- **Governance**: ALTO. Propuesta aprobada por el operador antes del apply. Cada step de escritura tiene su `Rollback()` vía el pipeline existente. Backup del proyecto antes de escribir usando `internal/backup` (C-17, dir-aware).
- **Archivos nuevos**: `internal/model/installtarget.go`, `internal/model/installtarget_test.go`, `internal/model/agentpaths.go`, `internal/agents/claude/adapter_project_test.go`, `internal/agents/claude/adapter_regression_test.go`, `internal/agents/opencode/adapter_project_test.go`, `internal/agents/opencode/adapter_regression_test.go`, `internal/install/project_target_test.go`, `internal/install/machine_regression_test.go`, `internal/install/collect_write_paths_test.go`, `internal/install/project_rollback_test.go`.
- **Archivos modificados**: `internal/model/agentpaths.go` (nuevo tipo de dominio), `internal/agents/interface.go` (método `PathsFor` + doc), `internal/agents/claude/adapter.go` (implementa `PathsFor`), `internal/agents/opencode/adapter.go` (implementa `PathsFor`), `internal/agents/registry_test.go` (fakeAdapter += PathsFor), `internal/install/types.go` (`AgentAdapter` += `PathsFor`; `Options` += `Target`/`ProjectRoot`), `internal/install/plan.go` (`BuildPlan` + `collectWritePaths` + `buildHarnessStep` target-aware; 4 helpers de resolución), `internal/install/testseams.go` (`SetSnapshotCreateWithHints`), `cmd/jr-stack/headless/executor_test.go` (fakeExecAdapter += PathsFor).

### C-28 — MCP scope proyecto (escritura de `.mcp.json`)
- **Estado**: HECHO (strict TDD; 24/24 tasks, `go test ./...` verde salvo `TestCompose_AllToggles` —falla PRE-EXISTENTE por golden desfasado de esta rama vs main, no causada por C-28, verificada con stash—; `go vet` limpio). Archivado `2026-06-03-c28-mcp-project-scope`.
- **Scope**: Conectar `Starter.MCPs[]` al flujo de install y escribir la config MCP en el path project-scope correcto, target-aware. Foco Claude + OpenCode.
  - **Bug de C-27 corregido**: el MCP project-scope de Claude se escribe en UN solo `<root>/.mcp.json` en la RAÍZ con shape `{"mcpServers":{...}}` (verificado contra doc oficial `code.claude.com/docs/en/mcp`), NO en `.claude/mcp/<server>.json`. `PathsFor(base, Project).MCPConfigPath` → `filepath.Join(base, ".mcp.json")`, ignora serverName.
  - **Estrategia target-aware (D1)**: enum `model.MCPStrategy` (3 valores: `SeparateFile` zero-value/legacy máquina, `MergeIntoSettings` OpenCode, `SingleFileMerge` Claude proyecto) doblada DENTRO de `AgentPaths` (campo poblado por `PathsFor`), no un método aparte. Hace irrepresentable la combinación inconsistente path-archivo-único + estrategia-archivo-separado. Enum en `internal/model` para evitar ciclo `model`↔`external`.
  - **`model.MCP` enriquecido (D3)**: `{Name, Command, Args, Env}` + `Validate()`. `Name` primero (back-compat C-26). Transporte remoto HTTP/SSE marcado TBD/fuera de scope.
  - `buildMCPOverlay()` emite `{"mcpServers":{"<Name>":{command,args,env}}}` reusando backup + `filemerge.MergeJSONObjects` + `WriteFileAtomic` ya existentes en `mcp.go`.
  - `Starter.MCPs[]` cableado en `plan.go` vía `buildMCPSteps()` + `collectStarterMCPPaths()`, espejando `buildHarnessStep`, cada step con `Rollback()`.
  - **Bug cazado**: el early-return `len(selected)==0` no dejaba emitir MCP steps en planes starter-only (sin harnesses) — corregido.
  - **OpenCode confirmado correcto**: `<root>/.opencode/opencode.json` + merge-into-settings, sin cambio de comportamiento (regression test agregado).
- **TBD resuelto post-C-28**: (1) path Machine de Claude — **RESUELTO en PR #4** (commit `d670c2c`, mergeado 2026-06-04): el MCP Machine de Claude se escribía en `~/.claude/mcp/<server>.json`, ruta que Claude Code NO lee (no hay auto-discovery ahí). Ahora se escribe en `~/.claude.json` (`mcpServers`) con discriminador `type:stdio/http`. Corrige también context7 en Claude.
- **TBD pendientes**: (2) shape de transporte MCP remoto HTTP/SSE — fuera de scope.
- **Dependencias**: C-26 (modelo + `model.MCP` placeholder), C-27 (`PathsFor` target-aware). Habilita: C-29 (subcomando `jr-stack starter add`).
- **Governance**: ALTO. Diseño aprobado por el operador antes del apply. Cada step de escritura con `Rollback()`; backup antes de escribir.
- **Archivos nuevos**: `internal/model/agentpaths_test.go`, `internal/install/plan_mcp_test.go`.
- **Archivos modificados**: `internal/model/starter.go` (MCP enriquecido + `Validate`), `internal/model/agentpaths.go` (enum `MCPStrategy` + campo + `WithMCPStrategy`), `internal/agents/claude/adapter.go` (`PathsFor` Project → `.mcp.json` + `SingleFileMerge`), `internal/agents/opencode/adapter.go` (`WithMCPStrategy(MergeIntoSettings)`), `internal/harness/external/mcp.go` (`buildMCPOverlay` + `WriteMCPProjectEntry`), `internal/install/types.go` (`Options` += `Starter`), `internal/install/plan.go` (`buildMCPSteps` + `collectStarterMCPPaths` + fix early-return), `internal/install/steps.go` (`mcpWriteStep` + `writeMCPEntry`), `internal/install/testseams.go` (`SetMCPWriteFn`).

### C-29 — Subcomando `jr-stack starter add`
- **Estado**: HECHO (strict TDD; 33/33 tasks, `go test ./...` verde salvo `TestCompose_AllToggles` PRE-EXISTENTE; `go vet` limpio). Archivado `2026-06-03-c29-starter-add-command`.
- **Scope**: Primera superficie de usuario del epic. Expone `jr-stack starter add <id> [--project <path>] [--dry-run] [--yes] [--agent <list>]`. NO construye maquinaria nueva — CABLEA el CLI al pipeline ya existente (C-26/C-27/C-28). Foco Claude + OpenCode.
  - Nuevo `case "starter"` en el dispatch manual (`switch os.Args[1]`) de `cmd/jr-stack/main.go` — sin cobra, respeta el estilo de `case "install"`.
  - **Default target = PROJECT** (a diferencia de `install` = Machine): un starter scaffoldea un proyecto. `--project` default a cwd, absolutizado, **FALLA si el path no existe** (nunca crea la raíz).
  - **D4 (resuelve D1 diferido de C-27)**: adapter queda STATELESS. Se rechazó la alternativa B (`TargetRoot` inyectado). El handler setea `Options{Target, ProjectRoot}` una vez y `BuildPlan` hace el threading interno → el "threading incómodo" que motivaba alt B nunca se materializa. **No se toca el contrato del adapter.**
  - **D3a (gap encontrado y tapado)**: `ResolveStarter` aplanaba solo harnesses, NO los MCPs → un starter con `includes` perdía MCPs anidados silenciosamente. Nuevo `catalog.ResolveStarterMCPs()` que une + dedup por `MCP.Name` (root gana en colisión).
  - **Reuso (D5/D7)**: rutea por el headless existente extendiendo `ParsedFlags` con `Target`/`ProjectRoot`/`Starter` (aditivo, zero-value = Machine = cero regresión). Reusa snapshot + `Rollback()` por step + dependency gate + dry-run. El `--dry-run` no escribe NADA.
- **TBDs resueltos**: `--agent` default = todos los focales registrados (claude+opencode); sin marcador de proyecto requerido (cualquier dir existente sirve — backup+rollback ya protege); colisión MCP = root-wins silencioso; D5 = extender `ParsedFlags`, no función hermana.
- **Dependencias**: C-26, C-27, C-28. Habilita: C-31 (slash-command), C-32 (verify + E2E starters).
- **Governance**: ALTO. Diseño aprobado por el operador antes del apply. Cada step de escritura con `Rollback()`; backup antes de escribir; `--dry-run` sin efectos.
- **Archivos nuevos**: `internal/catalog/resolve_mcps_test.go`, `cmd/jr-stack/headless/starter_flags.go` (+`_test`), `cmd/jr-stack/headless/project_root_test.go`, `cmd/jr-stack/headless/starter_executor_test.go`, `cmd/jr-stack/starter_add.go` (+`_test`), `cmd/jr-stack/starter_dispatch.go`, `cmd/jr-stack/routing_test.go`.
- **Archivos modificados**: `internal/catalog/catalog.go` (`ResolveStarterMCPs` + `AllStarters`), `cmd/jr-stack/headless/flags.go` (`ParsedFlags` += `Target`/`ProjectRoot`/`Starter`), `cmd/jr-stack/headless/executor.go` (cablea los campos a `install.Options`), `cmd/jr-stack/main.go` (`case "starter"`).

### C-30 — Curación del catálogo de starters (contenido real)
- **Estado**: HECHO (strict TDD; 46/46 tasks completadas, `go test ./internal/catalog/...` 46/46 verde, `go vet ./...` limpio, `openspec validate c30-starter-curation --strict` verde). Archivado pendiente. Se ejecutó en dos pasos: **C-30a** (operativo previo: publicación del monorepo `JuanCruzRobledo/jr-skills` con las 14 skills propias/fork) + **C-30 apply** (curación del catálogo YAML + tests).
- **Scope**: Llenar los 3 starters placeholder (backend, ux-ui, active-ia) con contenido real, introducir el starter `base` como transversal compartido, y modelar las ~30 skills curadas como entradas harness `type: skill`.
  - **Starter `base` (nuevo)**: 6 harnesses transversales compartidos por todos los starters — `test-driven-development`, `systematic-debugging`, `requesting-code-review`, `receiving-code-review`, `code-review-excellence`, `agile-product-owner` (todos terceros).
  - **Starter `backend` recompuesto**: `includes: [base]` + 12 harnesses backend-layer — `clean-architecture`, `fastapi-domain-service`, `fastapi-code-review`, `redis-best-practices`, `websocket-engineer`, `alembic-migrations`, `fastapi-templates`, `python-testing-patterns`, `postgresql-table-design`, `postgresql-optimization`, `api-security-best-practices`, `multi-stage-dockerfile`.
  - **Starter `ux-ui` recompuesto**: `includes: [base]` + 11 harnesses frontend-layer — `vercel-react-best-practices`, `zustand-store-pattern`, `react19-form-pattern`, `dashboard-crud-page`, `ws-frontend-subscription`, `help-system-content`, `interface-design`, `pwa-development`, `typescript-advanced-types`, `tailwind-design-system`, `playwright-best-practices`.
  - **Starter `active-ia` recompuesto**: `includes: [backend, ux-ui]` + `monorepo-scaffold`. Los 6 base transversales aparecen exactamente una vez (dedup por `ResolveStarter`).
  - **14 propias/fork**: `type: skill`, `source.repo: JuanCruzRobledo/jr-skills`, `method: clone`, `path: skills/<nombre>`, sin `third_party`. (8 propias puras + 6 forks expandidos del usuario, publicados en C-30a.)
  - **16 terceros**: `type: skill`, `third_party: true`, `best_effort: true`, `source.repo: <upstream confirmado>`. Incluye los 3 cabos sueltos resueltos en D2: `redis-best-practices` → `Mindrally/skills`; `websocket-engineer` → `Jeffallan/claude-skills`; `pwa-development` → `alinaqi/claude-bootstrap` (WATCH: usa `base.md` del mismo repo, clonar el subdir puede dejar la skill incompleta).
  - **Tests (8 nuevos en `c30_curation_test.go`)**: `TestC30_BaseStarter_ResolvesTo6Transversals`, `TestC30_BackendStarter_ResolvesToBaseUnionBackendLayer`, `TestC30_UxUiStarter_ResolvesToBaseUnionFrontendLayer`, `TestC30_ActiveIa_DeduplicatesBase` (edge case clave: base aparece una sola vez en active-ia), `TestC30_CuratedCatalog_LoadsClean`, `TestC30_CuratedCatalog_NoDuplicateHarnessIDs`, `TestC30_CuratedSkillHarnesses_OwnPointAtMonorepo`, `TestC30_CuratedSkillHarnesses_ThirdPartyAreBestEffort`.
- **TBDs resueltos**:
  - D1: monorepo único `JuanCruzRobledo/jr-skills` (layout `skills/<nombre>/SKILL.md`).
  - D2: `redis-best-practices` → `Mindrally/skills` (100% byte-idéntico); `websocket-engineer` → `Jeffallan/claude-skills` (95%, versión vieja pre-v1.1.0); `pwa-development` → `alinaqi/claude-bootstrap` (99%); los 3 son terceros, NO se publican.
  - D5: `install_modes: [full]` para todas las skills de starter (materializadas vía `jr-stack starter add`, no por modo base).
- **TBDs pendientes**: path de `pwa-development` puede quedar incompleto sin `base.md` del mismo repo (WATCH documentado en el harness). Follow-up en C-32 E2E si se prioriza.
- **Decisiones de clasificación no obvias**: los 3 cabos sueltos se clasificaron como terceros tras investigación por contenido (no por nombre de repo). Los 6 forks expandidos se modelan idéntico a las propias puras en el catálogo (D3 firme: el contenido materializado es siempre la versión del usuario).
- **Dependencias**: C-26 (modelo starters), C-29 (subcomando `starter add` que consume el catálogo). No bloquea C-32 (E2E), que puede correr sobre el contenido curado.
- **Governance**: ALTO (edita `harnesses.yaml`; toda edición validada por `catalog.Load()` — un catálogo malformado rompe el release). Diseño aprobado por el operador (review governance ALTO) antes del apply. TDD estricto activo.
- **Archivos nuevos**: `internal/catalog/c30_curation_test.go`.
- **Archivos modificados**: `internal/catalog/harnesses.yaml` (30 entradas harness nuevas + starters section recompuesta), `openspec/changes/c30-starter-curation/tasks.md` (24/24 tasks marcadas [x]).

### C-31 — Slash-command `starter add` (nuevo `HarnessType "command"`)
- **Estado**: HECHO (strict TDD; 46/46 tasks, `go test ./...` verde salvo `TestCompose_AllToggles` PRE-EXISTENTE; `go vet` limpio; `openspec validate --strict` verde). Archivado `2026-06-03-c31-starter-slash-command`. Implementado en 2 sesiones (la 1ra cortó por límite a 9/46 en estado verde; la 2da reconcilió §3 y completó §4-§7).
- **Scope**: Shipea un slash-command fino que envuelve `jr-stack starter add`. **Capacidad NUEVA**: el installer no shipeaba commands hasta ahora. Foco Claude + OpenCode.
  - **Nuevo `model.HarnessType = "command"`** (4to tipo junto a skill/config/external) + `IsValid()` extendido + entrada `starter-add-command` en `harnesses.yaml` (validada por `catalog.Load()`). Decisión del operador (TBD-2 opción a) sobre el step standalone.
  - **`CommandsDir` aditivo en el adapter** (D1): método `CommandsDir(homeDir)` en la interfaz + campo en `model.AgentPaths` poblado por `PathsFor(base, target)`, target-aware. Cero regresión (regression tests pinean paths byte-idénticos). Espeja C-27 exacto.
  - **Bifurcación per-agente** (D2): el adapter resuelve solo el directorio (Claude `.claude/commands`; OpenCode máquina `.config/opencode/commands`, proyecto `.opencode/commands`); namespace+filename viven en el asset embebido por agente. Claude → `commands/jr/starter-add.md` invocado `/jr:starter-add` (frontmatter completo); OpenCode → `commands/jr-starter-add.md` invocado `/jr-starter-add` (flat, description-only).
  - **`$ARGUMENTS`** (TBD-1, verificado contra doc oficial de AMBOS agentes: code.claude.com/docs/en/agent-sdk/slash-commands + opencode.ai/docs/commands) → una sola implementación, sin branching en los bodies. El body es un wrapper fino que ejecuta `jr-stack starter add $ARGUMENTS` (binario en PATH), sin reimplementar lógica.
  - **Mecanismo de instalación** `internal/harness/command/` (D5): espeja `internal/harness/skill/` — embed read → resuelve dir vía adapter → escritura managed de archivo completo + skip por content-hash + backup antes de pisar. Embebido vía `assets.CommandsFS` (`//go:embed`, TBD-4).
  - **Wiring pipeline** (governance ALTO): `resolvedCommandsDir` en `plan.go`; el dir entra a `collectWritePaths` (snapshot antes de escribir); `commandStep` ruteado en `buildHarnessStep` con `Rollback()` vía manifest. `--dry-run` no escribe.
- **TBDs resueltos**: TBD-1 `$ARGUMENTS` (ambos); TBD-2 `HarnessType command`; TBD-3 siempre-shipeado para agentes focales; TBD-4 `CommandsFS` en `assets/assets.go`.
- **Dependencias**: C-29 (subcomando `starter add`). Habilita: C-32 (verify + E2E starters).
- **Governance**: ALTO (inyección en command-dir del usuario). Diseño aprobado por el operador antes del apply. Snapshot + `Rollback()` por step.
- **Archivos nuevos**: `assets/commands/claude/jr/starter-add.md`, `assets/commands/opencode/jr-starter-add.md`, `assets/commands_test.go`, `internal/harness/command/{types,idempotent,installer,command_test}.go`, `internal/install/command_wiring_test.go`, `internal/install/command_integration_test.go`.
- **Archivos modificados**: `assets/assets.go` (`CommandsFS`), `internal/model/harness.go` (`HarnessCommand` + `IsValid`), `internal/install/steps.go` (`commandStep` + `commandInstallFn` + `toCommandAdapters`), `internal/install/plan.go` (`resolvedCommandsDir` + ruteo `HarnessCommand`), `internal/install/types.go` (`WithEmbeddedCommandsFS`), `internal/install/testseams.go` (`SetCommandInstallFn`), `internal/catalog/harnesses.yaml` (`starter-add-command`), `internal/agents/{interface,claude/adapter,opencode/adapter}.go` (`CommandsDir`).

### harness-scope-model — Scope dimension en el modelo: excluir starter-only del plan global
- **Estado**: HECHO (strict TDD; 22/22 tasks, `go test ./internal/model/... ./internal/catalog/...` 100% verde, `go vet ./...` limpio, `openspec validate harness-scope-model --strict` verde).
- **Bug corregido**: `ForMode(full)` y `ForMode(custom)` devolvían 43 harnesses (13 foundation-global + 30 starter-only C-30) en lugar de 13. Las 30 skills C-30 se filtraban al plan de instalación global cuando su única ruta válida es `jr-stack starter add`. **Antes: lite=6, full=43, custom=43 → Después: lite=6, full=13, custom=13.**
- **Contrato implementado**:
  - **(a)** `model.ScopeKind string` con `ScopeGlobal = "global"` (zero-value, back-compat total) y `ScopeStarterOnly = "starter-only"`. Campo `Scope ScopeKind` (yaml `scope,omitempty`) en `Harness`. Helper `IsStarterOnly() bool` (`return h.Scope == ScopeStarterOnly`).
  - **(b)** `ForMode(m)` filtra `!h.IsStarterOnly() && h.InMode(m)`. El filtro de scope vive en `ForMode` (capa catálogo), NO en `InMode` (que conserva su semántica exacta). `ModeCustom` sigue siendo "todos los foundation-global", dejando de ser "todos".
  - **(c)** `ResolveStarter` INTACTO — ignora `Scope` e `InstallModes`; la ruta `starter add` queda sin cambio.
  - **(d)** `harnesses.yaml`: `scope: starter-only` añadido a las 30 skills C-30 (8 propias puras, 6 forks expandidos, 16 terceros). Los 13 foundation-global omiten el campo (zero-value = global). `install_modes: [full]` intacto en los starter-only.
  - **(e)** `catalog.Load()` valida 2 reglas nuevas: **Regla 1** — harness starter-only no referenciado por ningún starter = error nombrando el id (orphan check, corre DESPUÉS de `validateStarters()` con `validateStarterOnlyReferences()`). **Regla 2** — starter-only + `lite` en install_modes = error nombrando el id (en el loop de harnesses de `validate()`).
- **Verificación de clasificación**: los 30 IDs del listado de starter-only fueron cruzados contra `starters[].harnesses[]` en el YAML — todos están referenciados por al menos un starter (Regla 1 pasa en el catálogo embebido). Ninguno en el listado foundation-global es starter-only.
- **Tests añadidos**: `TestModel_ScopeKind_ZeroValueIsGlobal` (model), `TestForMode_CustomReturnsOnlyGlobal` (reemplaza `TestForMode_CustomReturnsAll`), `TestForMode_ExcludesStarterOnlyHarnesses`, extensión de `TestForMode_LiteIsSubsetOfFull`, `TestCatalogLoad_StarterOnlyMustBeReferenced`, `TestCatalogLoad_StarterOnlyWithLiteModeIsInvalid`, `TestC30_CuratedSkillHarnesses_AllAreStarterOnly` (en c30_curation_test.go).
- **Dependencias**: C-30 (curación del catálogo, define los 30 starter-only). Habilita: C-32 (verify + E2E starters — ahora el plan de instalación global ya no arrastra skills de starter).
- **Governance**: ALTO (edita `harnesses.yaml`; aprobado por el operador antes del apply; TDD estricto; `catalog.Load()` valida en tests antes de cerrar). [[bug/starter-skills-leak-global-install]]
- **Archivos modificados**: `internal/model/harness.go` (ScopeKind + Scope + IsStarterOnly), `internal/catalog/catalog.go` (ForMode filter + Rule 2 + validateStarterOnlyReferences/Rule 1), `internal/catalog/harnesses.yaml` (scope: starter-only en 30 skills C-30).
- **Archivos nuevos**: `internal/model/harness_test.go`, extensiones en `internal/catalog/catalog_test.go`, `internal/catalog/c30_curation_test.go`.

### wire-commands-fs — Cablear `WithEmbeddedCommandsFS` en el entry point del binario
- **Estado**: HECHO (strict TDD; 17/17 tasks, `go test ./internal/install/... ./internal/harness/command/... ./assets/...` verde, `go vet ./...` limpio, `openspec validate wire-commands-fs --strict` verde).
- **Bug corregido**: `"command installer: commandsFS is nil"` en runtime. C-31 shipeó el mecanismo completo (`commandInstallFn` + `embeddedCommandsFS` + `assets.CommandsFS` + `WithEmbeddedCommandsFS`) pero **no hubo ningún call site** de `WithEmbeddedCommandsFS` en `cmd/jr-stack`. El var global `embeddedCommandsFS` quedaba `nil` en todo proceso real; `cmdinstaller.NewInstaller(nil)` → error inmediato (el nil-check es correcto y no se tocó).
- **Causa raíz (asimetría)**: `WithEmbeddedSkillsFS(opts, f) Options` es un functional-option que devuelve `opts` → los 4 sitios de wiring de skills naturalmente "jalan" el call a lo largo del código. `WithEmbeddedCommandsFS(f)` mutaba un global sin devolver nada → no había un `opts` thread que lo arrastrara, así que el call fue omitido silenciosamente.
- **Por qué los tests existentes no lo atraparon**: `command_wiring_test.go` → stub completo vía `SetCommandInstallFn`; `command_integration_test.go` → hand-rolled `fstest.MapFS` inyectado directamente. Ambos bypasean `embeddedCommandsFS` por completo. Verde en CI → rojo en producción.
- **Fix (design D1 — Opción A)**: un call único `install.WithEmbeddedCommandsFS(assets.CommandsFS)` en `wireEmbeddedFS()`, llamado al inicio de `main()`, antes de cualquier branch de dispatch (TUI / headless / starter). El global es leído en tiempo de ejecución del plan (no en BuildPlan), así que un set único cubre los tres caminos.
- **Alternativa descartada (Opción B)**: refactorizar `WithEmbeddedCommandsFS` como functional-option `(opts, f) Options` y cablear en los 4 sitios junto a `WithEmbeddedSkillsFS`. Más uniforme pero fuera de proporción para un fix de una línea; registrado como follow-up.
- **Red test real (D2)**: `TestWireEmbeddedFS_SetsCommandsFS` en `cmd/jr-stack/main_wiring_test.go` — falla en compilación en HEAD actual (`undefined: wireEmbeddedFS`), pasa después del fix. Usa `install.GetEmbeddedCommandsFS()` / `install.ResetEmbeddedCommandsFS()` (seams nuevos en `testseams.go`) y `assets.CommandsFS` real, no un `fstest.MapFS` handrolled.
- **Triangulación**: dos casos más en `internal/install/command_entrypoint_wiring_test.go` — install path (TUI/headless) y starter-add path ambos aseveran que `embeddedCommandsFS` es non-nil cuando corre `commandInstallFn`.
- **Archivos sin tocar** (intactos por diseño): `internal/harness/command/installer.go` (nil-check correcto, se mantiene), `assets/assets.go` (embed correcto, sin cambio).
- **Dependencias**: C-31 (mecanismo + assets). Habilita: C-32 (verify + E2E starters — el command step ya no falla en runtime).
- **Governance**: MEDIO (entry-point + install wiring). TDD estricto. [[opsx/wire-commands-fs/apply]]
- **Archivos nuevos**: `cmd/jr-stack/main_wiring_test.go`, `internal/install/command_entrypoint_wiring_test.go`.
- **Archivos modificados**: `cmd/jr-stack/main.go` (`wireEmbeddedFS()` + call en `main()`), `internal/install/testseams.go` (`GetEmbeddedCommandsFS` + `ResetEmbeddedCommandsFS`).

### planner-resolves-starter-only — Reparar el índice del planner para resolver starter-only harnesses
- **Estado**: HECHO (strict TDD; 11/11 tasks, `go test ./...` 22/22 paquetes verde, `go vet ./...` limpio, `openspec validate planner-resolves-starter-only --strict` verde).
- **Regresión corregida**: `jr-stack starter add <id>` fallaba con `error: build plan: install: resolve dependencies: planner: unknown harness "test-driven-development"`. 3 tests rojos en `cmd/jr-stack`: `TestStarterDispatch_AddSubcommandRoutes`, `TestRunStarterAdd_KnownID_BuildsPlanWithProjectTarget`, `TestRunStarterAdd_DryRun_WritesNothing`.
- **Causa raíz — `internal/install/plan.go:41-44`**: `BuildPlan` construía el índice de resolución de dependencias del planner con `cat.ForMode(model.ModeCustom)`. Antes de `harness-scope-model`, `ForMode(ModeCustom)` devolvía todo el catálogo por accidente (ningún harness excluido). `harness-scope-model` añadió el filtro `!h.IsStarterOnly()` en `ForMode`, que correctamente excluye los 30 harnesses starter-only de la selección global — pero también eliminó esas 30 entradas del índice de resolución del planner. `SelectHarnesses` (vía `ByID`) los encontraba y los metía en el selected-set; el planner no los encontraba en su índice → "unknown harness".
- **Principio del fix**: scope es una propiedad de *qué se instala* (selección), no del *grafo de dependencias* (resolución). El planner necesita el grafo completo; la gate de scope ya la aplicó `SelectHarnesses`.
- **Fix — 3 pasos quirúrgicos**:
  1. `internal/install/types.go`: añadir `AllHarnesses() []model.Harness` a la interfaz `install.Catalog` — sin filtro de scope ni modo. Documentado como accessor de resolución, distinto de `ForMode`.
  2. `internal/catalog/catalog.go`: implementar `func (c *Catalog) AllHarnesses() []model.Harness { return c.Harnesses }` — devuelve el slice completo sin ningún filtro.
  3. `internal/install/plan.go:41-44`: construir `allHarnesses` desde `cat.AllHarnesses()` en lugar de `cat.ForMode(model.ModeCustom)`.
- **Por qué NO reaparece el leak original**: `SelectHarnesses` sigue siendo el único punto de selección. En modo `full`/`lite` usa `ForMode` (scope-filtrado); en modo `Custom` usa `cat.ByID` para las IDs explícitas del intent. Ampliar el índice de resolución solo cambia qué puede *encontrar* el resolver al chasear aristas `DependsOn` — nunca agrega un harness no seleccionado al plan.
- **Verificación — tests de regresión del scope (GUARDAS)**:
  - `TestForMode_ExcludesStarterOnlyHarnesses` y `TestForMode_CustomReturnsOnlyGlobal` en `internal/catalog` — pasan INALTERADOS (el fix no toca `ForMode`).
  - `TestBuildPlan_FullMode_NoStarterOnlyHarness` (nuevo) — asevera que un plan `install --mode full` no contiene ningún harness `Scope == ScopeStarterOnly` ni después del cambio.
- **Test de regresión del install layer**: `TestBuildPlan_StarterOnlyHarnessResolvesViaFullCatalog` (nuevo) — usa el catálogo real, resuelve el starter `base` (incluye `test-driven-development`), construye intent `ModeCustom` espejando `starter_add.go`, llama `BuildPlan`, aserta (a) sin "unknown harness" y (b) `test-driven-development` presente en los Apply steps. FALLÓ en el árbol antiguo con el mensaje exacto de la regresión; PASA tras el fix.
- **Impacto en implementaciones de `install.Catalog`**: 4 fakes en tests necesitaban `AllHarnesses()` — añadido en `install_test.go` (`fakeCatalog`), `internal/tui/gate_test.go` (`fakeTUICatalog`), `internal/tui/review_test.go` (`fakeCatalog`), `cmd/jr-stack/headless/executor_test.go` (`fakeExecCatalog`), `cmd/jr-stack/unify_test.go` (`cmdFakeCatalog`). Todos implementan `return f.harnesses` (sin filtro).
- **Dependencias**: `harness-scope-model` (introduce el filtro que causó la regresión), C-29/C-30/C-31/wire-commands-fs (contexto de la cadena). Habilita: C-32 (verify + E2E starters — ahora `starter add` buildea el plan sin error).
- **Governance**: MEDIO (`internal/install` build-plan path). TDD estricto. [[bug/scope-filter-breaks-starter-planner]] [[opsx/planner-resolves-starter-only/apply]]
- **Archivos nuevos**: `internal/install/plan_starter_only_test.go` (`TestBuildPlan_StarterOnlyHarnessResolvesViaFullCatalog` + `TestBuildPlan_FullMode_NoStarterOnlyHarness`).
- **Archivos modificados**: `internal/install/types.go` (`AllHarnesses()` añadido a interfaz `Catalog`), `internal/catalog/catalog.go` (`AllHarnesses()` implementado), `internal/install/plan.go` (líneas 41-44: `cat.AllHarnesses()` en lugar de `cat.ForMode(model.ModeCustom)`), `internal/install/install_test.go` (`fakeCatalog` += `AllHarnesses`), `internal/tui/gate_test.go` (`fakeTUICatalog` += `AllHarnesses`), `internal/tui/review_test.go` (`fakeCatalog` += `AllHarnesses`), `cmd/jr-stack/headless/executor_test.go` (`fakeExecCatalog` += `AllHarnesses`), `cmd/jr-stack/unify_test.go` (`cmdFakeCatalog` += `AllHarnesses`).

### binary-self-install — Auto-instalar el binario en PATH en `jr-stack install`
- **Estado**: HECHO (strict TDD; 34/34 tasks, `go test ./...` 22/22 paquetes verde, `go vet ./...` limpio, `openspec validate binary-self-install --strict` verde).
- **Problema corregido**: `build.bat`/`build.sh` compilaban el binario a la raíz del repo, pero ningún paso copiaba el resultado al PATH del usuario. `jr-stack` seguía resolviendo a un binario stale en `%LOCALAPPDATA%\jr-stack\bin` (o a nada). El usuario tenía que mover manualmente el `.exe` y re-ejecutar. Ahora `jr-stack install` se auto-instala por defecto.
- **Decisión (default ON + `--no-self-install`)**: la auto-instalación está activa por defecto; se desactiva con `--no-self-install` para CI / builds reproducibles. El operador confirmó los dos TBDs antes del apply: TBD-1 = `system.BinaryInstallDir` compartido; TBD-2 = hint text en `build.*`.
- **Contrato implementado**:
  - **(D1)** `binaryInstallDir` + `isWritableDir` extraídos de `internal/harness/external/homebrew.go` a `internal/system/binary_install_dir.go` como `system.BinaryInstallDir(goos string) string` (+ `binaryInstallDirWritableFn` injectable). `homebrew.go` reapunta su seam al exportado; dead code eliminado.
  - **(D2)** `selfInstallStep` en `internal/install/self_install.go`: `ID()="self-install"`, `Run()` (resolve ejecutable → resolve dir → MkdirAll → `copyBinaryRenameReplace` → `AddToUserPath`), `Rollback()` (via manifest), `setManifest`. Graceful degradation: falla en `os.Executable`, copy o PATH emite WARN y devuelve nil (no aborta el install).
  - **(D3)** `install.Options` += `NoSelfInstall bool` + `SelfInstallBinaryPath string`. `BuildPlan` añade el step DESPUÉS de harness+MCP steps (gated por `!opts.NoSelfInstall`). `collectSelfInstallPath` agrega el target al snapshot (governance ALTO). Manifest wired via `manifestReceiver` loop.
  - **(D4)** Rename-then-replace para Windows exe-in-use: `copyBinaryRenameReplace` escribe `.new`, mueve target a `.old`, renombra `.new` → target, best-effort remove `.old`.
  - **(D5)** Flag `--no-self-install` en `cmd/jr-stack/headless/flags.go` → `ParsedFlags.NoSelfInstall`. `RunHeadless` cablea a `opts.NoSelfInstall`. `main.go` resuelve `os.Executable()` una vez y lo pasa como `BinaryPath` en `ParsedFlags`.
  - **(TBD-2)** `build.bat` y `build.sh`: añadido hint "ejecuta `jr-stack install` para terminar el setup" al final del output.
- **Test seams (tests NUNCA tocan el entorno real)**: `SetExecPathFn`, `SetSelfInstallBinaryInstallDirFn`, `SetAddToUserPathFn` en `testseams.go` + `NewSelfInstallStep` constructor. Todos los tests usan `t.TempDir()`.
- **Tests existentes rotos y reparados**: 3 tests de snapshot (`TestCollectWritePaths_ProjectTarget_ConfigHarness`, `TestBuildPlanMachineTarget_PathsUnderHomeDir`, `TestBuildPlanProjectTarget_PathsUnderProjectRoot`) aseveraban que TODOS los paths capturados estuvieran bajo homeDir/projectRoot. El path de self-install es independiente (bin dir del sistema). Reparados con `NoSelfInstall: true` (estos tests verifican harness paths, no self-install).
- **Governance**: ALTO (escribe al bin dir del usuario + muta PATH). Snapshot antes de escribir + `Rollback()` mandatory. Aprobado por el operador (TBDs confirmados) antes del apply. TDD estricto.
- **Archivos nuevos**: `internal/system/binary_install_dir.go`, `internal/system/binary_install_dir_test.go`, `internal/install/self_install.go`, `internal/install/export_test.go`, `internal/install/self_install_copy_test.go`, `internal/install/self_install_step_test.go`, `internal/install/self_install_seams_test.go`, `internal/install/plan_self_install_test.go`, `cmd/jr-stack/headless/no_self_install_flags_test.go`, `cmd/jr-stack/headless/self_install_executor_test.go`.
- **Archivos modificados**: `internal/harness/external/homebrew.go` (reapunta `binaryInstallDirFn` a `system.BinaryInstallDir`; elimina dead `binaryInstallDir`+`isWritableDir`), `internal/install/types.go` (`Options` += `NoSelfInstall`+`SelfInstallBinaryPath`), `internal/install/plan.go` (step 5c + `collectSelfInstallPath`), `internal/install/testseams.go` (`SetExecPathFn`+`SetSelfInstallBinaryInstallDirFn`+`SetAddToUserPathFn`), `internal/install/collect_write_paths_test.go` (`NoSelfInstall: true`), `internal/install/machine_regression_test.go` (`NoSelfInstall: true`), `internal/install/project_target_test.go` (`NoSelfInstall: true`), `cmd/jr-stack/headless/flags.go` (`ParsedFlags` += `NoSelfInstall`+`BinaryPath`; flag `--no-self-install`), `cmd/jr-stack/headless/executor.go` (cablea `NoSelfInstall`+`SelfInstallBinaryPath`), `cmd/jr-stack/main.go` (resuelve `os.Executable()` una vez en headless path), `build.bat`+`build.sh` (hint text).

### C-32 — Verify + E2E de starters (CIERRE del epic starter packs)
- **Estado**: HECHO (strict TDD; 37/37 tasks, `go test ./...` 23/23 paquetes verde, `go vet ./...` limpio, `openspec validate c32-verify-e2e-starters --strict` verde). Archivado `2026-06-04-c32-verify-e2e-starters`. **Cierra el epic starter packs al 100%.**
- **Scope**: Dos concerns aprobados por el operador. Re-curar los 6 paths upstream podridos quedó FUERA (follow-up separado — drift de terceros).
  - **Concern 1 — E2E híbrido (governance BAJO, test code)**: nuevo paquete Go `e2e/starter/`. Brazo **hermético** (`starter_hermetic_test.go`, SIN build tag → corre siempre en `go test ./...`) usa repos git locales servidos por `file://` en `t.TempDir()`, sin red. Brazo **de red** (`starter_network_test.go`, `//go:build e2e_network` → opt-in, NUNCA en el default) clona el upstream real. Helpers tag-free compartidos en `starter_e2e_helpers.go`. Asevera **post-condiciones en disco** (D4): `SKILL.md` presente y no-vacío en `<adapter.SkillsDir(homeDir)>/<id>/SKILL.md` (path vía adapter real, nunca hardcodeado — CLAUDE.md §3), NO confía en el exit-code/✓.
  - **Concern 2 — Reporte honesto (governance MEDIO, reporting path)**: nuevo `pipeline.StepStatusDegraded`. Un best-effort fallido ahora emite estado `degraded` (glyph `⚠`), distinto de `✗` fatal **y** de `✓` limpio; se enumera en el resumen final ("Degraded (best-effort) harnesses:"); el run sigue saliendo exit 0. La política continue-don't-abort de C-19/C-22 NO se toca.
- **El bug que justificaba C-32 (verificado en fuente)**: un best-effort con `source.path` podrido fallaba en `cloneInstaller` pero `skillStep.Run()` (`internal/install/steps.go:123-145`) emitía `StepStatusFailed` (idéntico a fatal) y devolvía `nil` → el `StepResult` estructurado registraba `Succeeded` (`runner.go:52`, **mentía**) y el resumen daba `Ready==true`/exit 0 sin nombrar nada. Tres mentiras apiladas: glyph que confunde, resultado estructurado falso, resumen que oculta.
- **Decisiones clave**:
  - **D1**: build tag `//go:build e2e_network` para el brazo de red (opt-IN real: el código de red ni compila en el binario default). Rechazado `testing.Short()` (opt-out, se corre por accidente) y env-var gate (compila igual).
  - **D2**: fixtures = repos git locales por `file://` (ejercita el `cloneInstaller` real, mismo code path que `https://`). Rechazado `fstest.MapFS` (saltearía justo el path que se rompe). Requiere `git` en PATH → `t.Skip` limpio si falta (dep dura del propio installer).
  - **D3 (seam resuelto en apply, opción preferida, sin fricción)**: la lista de degradados se **deriva de los progress events** (`Status == StepStatusDegraded`), cero plumbing nuevo en el pipeline core. **Bug de wiring cazado**: `progressFn` se definía DESPUÉS de `opts` en `RunHeadless` → `opts.OnProgress` quedaba `nil` y los eventos degradados iban a stderr sin que nadie los viera. Fix: definir `progressFn` antes y cablear `opts.OnProgress = progressFn`.
  - **D4**: post-condiciones aseveradas en `t.TempDir()` como homeDir (aislamiento); para fixture best-effort roto: run exit 0 + reportado `degraded` + `SKILL.md` correctamente ausente (la degradación no medio-escribe).
- **Specs sincronizados al archivar**: NUEVA capability `starter-e2e-verification`; MERGE en `best-effort-harness` (warning observable + resumen honesto).
- **Dependencias**: C-29 (`starter add`), C-30 (catálogo curado), C-31 (slash-command), harness-scope-model, wire-commands-fs, planner-resolves-starter-only — todas HECHAS. No habilita nada (es el cierre del epic).
- **Governance**: BAJO (E2E) + MEDIO (reporte). Diseño y apply aprobados por el operador. TDD estricto. [[bug/best-effort-reports-ok-on-failed-install]] [[opsx/c32-verify-e2e-starters/apply]]
- **Cambio de producción notable**: `internal/harness/skill/clone.go` ahora acepta URLs ya calificadas (`if strings.Contains(repo, "://")` → usa tal cual; si no, antepone `https://github.com/` como siempre). Comportamiento default IDÉNTICO; solo habilita fixtures `file://`.
- **Archivos nuevos**: `e2e/starter/starter_e2e_helpers.go`, `e2e/starter/starter_hermetic_test.go`, `e2e/starter/starter_network_test.go` (`//go:build e2e_network`), `internal/catalog/parse_external.go` (`ParseForTest()` para catálogos de test sin tocar `harnesses.yaml`), `internal/pipeline/degraded_status_test.go`, `internal/install/degraded_result_test.go`, `cmd/jr-stack/headless/degraded_executor_test.go`, `internal/tui/degraded_tui_test.go`.
- **Archivos modificados**: `internal/pipeline/result.go` (`StepStatusDegraded`), `internal/install/steps.go` (best-effort emite `Degraded`), `cmd/jr-stack/headless/executor.go` (case `Degraded` `⚠` + bloque resumen + wiring `opts.OnProgress`), `internal/tui/model.go` (`statusIcon` + `viewComplete` con degradados), `internal/harness/skill/clone.go` (soporte `file://`), `internal/install/best_effort_test.go` (extendido).
- **Follow-ups abiertos (NO en C-32)**: (1) re-curar los 6 paths upstream third_party podridos (`code-review-excellence`, `agile-product-owner`, `fastapi-templates`, `python-testing-patterns`, `vercel-react-best-practices`, `playwright-best-practices`); (2) contenido de `pwa-development` (`base.md` faltante, WATCH de C-30 — el E2E lo DETECTA, no lo repara).
