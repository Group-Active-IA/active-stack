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
| 4 | C-11a | C-11b | install-pipeline: orquestación headless + rollback |
| 5 | C-11b | C-12, C-13 | TUI Bubbletea slim: punto de convergencia visible |
| 6 | C-12, C-13 | C-14 | uninstall + orquestador de fundación, en paralelo |
| 7 | C-14 | — | verify + E2E (cierre) |

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
- **Estado**: PENDIENTE.
- **Scope**: Portar `internal/tui` (Bubbletea) **slim**, sin theme cosmético. Flujo: detectar OS/agentes → elegir agente(s) → elegir modo (Lite/Full/Custom) → invocar `internal/install` (C-11a) → render de progreso suscrito al `ProgressFunc`. Selección/agrupación **por harness** (no por componente viejo). Se descarta theme/statusline/keybindings/persona-marketing.
- **Dependencias**: C-11a.
- **Governance**: BAJO (la TUI orquesta; el riesgo vive en el pipeline que invoca).
- **Leer antes**: ARCHITECTURE.md §4.1 (flujo install), repo viejo `internal/tui/`, skill `go-testing` (teatest).

### C-12 — Flujo uninstall harness-aware
- **Estado**: PENDIENTE.
- **Scope**: Uninstall que entiende **harnesses** (no componentes viejos): revertir inyecciones por markers, restaurar desde backup, remover skills clonadas. Mantener la interfaz de uninstall de la TUI.
- **Dependencias**: C-11.
- **Governance**: **ALTO** (restaura/borra config del usuario).
- **Leer antes**: repo viejo flujo uninstall + `internal/backup/restore.go`, ARCHITECTURE.md §1.1 (nota uninstall).

### C-13 — jr-orchestrator como orquestador de fundación
- **Estado**: PENDIENTE.
- **Scope**: Crear la skill `jr-orchestrator` (thin orchestrator from scratch) con lazy-loading (ARCHITECTURE.md §4.2): `openspec init` → `kb-creator` → `roadmap-generator` → `find-skill` → `agent-instruction`.
- **Dependencias**: C-11.
- **Governance**: MEDIO.
- **Leer antes**: ARCHITECTURE.md §4.2, MANUAL Etapas 1–4, catálogo (skills propias).

### C-14 — Verify + E2E
- **Estado**: PENDIENTE.
- **Scope**: Portar `internal/verify` (health checks post-install) + suite E2E Docker (Ubuntu + Arch) adaptada a harnesses. Cierre del pipeline.
- **Dependencias**: C-12, C-13.
- **Governance**: BAJO.
- **Leer antes**: repo viejo `internal/verify/` y `e2e/`, ARCHITECTURE.md §6 punto 8.

---

## Decisiones abiertas (TBD)

- **Nombre del repo remoto en GitHub** (colisión con el `jr-stack` actual): renombrar / archivar / nombre nuevo. (ARCHITECTURE.md §7).
- **Granularidad final de los toggles** de `sdd-orchestrator` (afecta C-09).
- **Qué harness entra exactamente en cada modo** Lite/Full/Custom (afecta C-11; el catálogo tiene un mapeo provisional).
- **Empaquetado de skills de terceros** (`find-skill`, `skill-creator`): repos `vercel-labs/skills` y `anthropics/skills` marcados TBD en el catálogo (afecta C-08).
- ~~**Repos de skills propias** `jr-starter` y `skill-registry`: marcados TBD en el catálogo.~~ → **RESUELTO** (`JuanCruzRobledo/jr-orchestrator` y `JuanCruzRobledo/skill-registry`, ambos públicos).
