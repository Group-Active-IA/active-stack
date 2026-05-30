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

**Camino crítico**: `C-01 → C-02 → C-04/C-05 → C-09 → C-10 → C-11a → C-11b → C-13 → C-14`.
Es la cadena más larga: el merge/backup habilitan el installer de config
(sdd-orchestrator), que alimenta los adapters, que alimentan el install-pipeline
headless, que monta la TUI, que habilita el orquestador de fundación, que se
cierra con verify+E2E.

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
- **Empaquetado de skills de terceros** → **CONFIRMADO el comando** (`npx skills add <owner/repo> --skill <name>` — paquete npm `skills` de vercel-labs). Falta **implementar** el `--skill` en el installer (ver **C-22**): hoy `npx skills add <repo>` sin `--skill` entra en selección interactiva y sin TTY (E2E Docker) revienta con `exit 1`. Ya degradado best-effort por C-19, así que no bloquea.

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

### C-23 — Limpiar el método `npx` huérfano  *(PENDIENTE)*
- **Estado**: PENDIENTE. Tras C-22 ningún harness usa `method: npx`. Cleanup.
- **Scope**: borrar `internal/harness/skill/npx.go` + sus tests (`npx_test.go`, casos npx en `skill_test.go`); quitar el `case "npx"` de `installer.go`; en `catalog.go` la inferencia para third-party pasa de `npx` a `clone` y se quita `npx` de los métodos válidos (actualizar `TestMethodInference_ThirdParty` y la validación); ajustar el comentario `clone | npx | embed` en `model.Source`. Governance BAJO/MEDIO; TDD.
