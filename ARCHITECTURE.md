# Active Stack — Arquitectura (blueprint v0)

> Instalador **methodology-first** del harness de desarrollo asistido por IA.
> Materializa el `MANUAL-METODOLOGICO.md`: un comando instala y configura el
> sustrato (harnesses), y un único orquestador de fundación deja el proyecto
> listo para el ciclo OPSX (`explore → propose → apply → verify → archive`).

Estado: **planificación**. Este documento es la fuente de verdad del diseño
mientras arrancamos. Se construye desde cero, portando la infraestructura
probada del active-stack actual (no reescribiéndola a ciegas).

---

## 1. Qué es (y qué NO es)

**Es:** un configurador del harness metodológico. Toma el/los agente(s) de IA
del usuario y les inyecta, de forma modular y actualizable, los harnesses que
exige la metodología.

**NO es** (decisión: sacar del scope, no van en la metodología):

- ❌ GGA (code review en commit) — harness aparte, no parte del ciclo SDD.
- ❌ Themes / statusline / keybindings — cosmético.
- ❌ Persona como producto de marketing ("Your own Gentleman!").
- ❌ Framing "supercharge any agent" / relación con Gentleman.Dots.

La persona puede sobrevivir como harness **opcional** (configuración), sin el
envoltorio de marketing.

### 1.1 Mantener / Modificar / Sacar

El repo es nuevo por el volumen de cambios, pero NO se tira todo. Buckets:

| Bucket | Qué | Detalle |
|--------|-----|---------|
| **MANTENER** | Núcleo del instalador e interfaz | Detección OS/deps, adapters por agente, backup/rollback, merge por markers, pipeline, verify, y la **interfaz de install/uninstall** de la TUI. |
| **MODIFICAR** | Flujo de instalación + branding | El **flujo de install se rediseña**: antes era selección de componentes por agente; ahora hay **muchos más módulos (harnesses)** → selección/agrupación por harness, árbol de deps entre harnesses, modos. Limpiar todo leftover `gentle-ai` / `gentle-stack` / `Gentleman.Dots`. |
| **SACAR** | Lo no-metodológico | GGA, theme, persona-marketing, statusline, keybindings, framing "supercharge any agent". |

> Nota: el flujo de install **no es un port directo** — es un rediseño. La
> mecánica de bajo nivel (merge, backup, adapters) se reusa; la **orquestación
> de qué se instala y cómo se presenta** se rehace alrededor del catálogo de
> harnesses (§3). El uninstall se mantiene pero debe entender harnesses, no
> componentes viejos.

---

## 2. Definición de Harness

Un **harness** es cualquier módulo que prepara o guía el entorno de la IA.
**No es sinónimo de "repo".** Se materializa de tres formas:

| Forma | Qué es | Origen | Ejemplos |
|-------|--------|--------|----------|
| **Skill** | `SKILL.md` + assets, cargada bajo demanda | repo propio o de terceros, se baja al instalar | kb-creator, roadmap-generator, agent-instruction, active-orchestrator, skill-registry (propias); find-skill, skill-creator (terceros) |
| **Configuración** | Texto/archivos que configuran el agente | bundleado en el instalador | `sdd-orchestrator` (orquestador con toggles: TDD, Engram, …), config de `AGENTS.md`/`CLAUDE.md`, permisos, MCP |
| **Herramienta externa** | Binario/servicio de terceros | instalado/configurado, no es nuestro repo | Engram, OpenSpec CLI, Context7 (MCP) |

Principio: el instalador no instala "repos", instala **harnesses**, y cada
harness sabe cómo se materializa.

### 2.1 El harness `sdd-orchestrator` (clave)

Es un harness de **configuración** que arma el orquestador SDD con piezas
**toggleables modularmente**. El usuario elige qué capacidades activar:

- Estricto TDD (sí/no)
- Memoria Engram (sí/no)
- Routing de modelos por fase (sí/no)
- Delegación a sub-agentes (sí/no)
- Governance por dominio (sí/no)
- … (extensible)

El resultado es el bloque de instrucciones del orquestador que se inyecta en
`CLAUDE.md` / `AGENTS.md`, compuesto a partir de los toggles elegidos.

---

## 3. Catálogo de harnesses (embebido)

El instalador lleva embebido un **catálogo maestro** (p. ej. `harnesses.yaml`)
que es la fuente de verdad de qué harnesses existen. Cada entrada declara:

```yaml
- id: kb-creator
  type: skill            # skill | config | external
  source:
    repo: Group-Active-IA/kb-creator
    ref: latest          # tag/branch/commit
  install_modes: [full]  # lite | full | ...
  depends_on: [openspec] # otros harness id
  agents: [claude, opencode, ...]   # dónde aplica
- id: sdd-orchestrator
  type: config
  toggles: [tdd, engram, model-routing, delegation, governance]
  install_modes: [lite, full]
  ...
- id: engram
  type: external
  install: { method: homebrew | download, ... }
```

La TUI lee este catálogo para construir la lista de selección. Para
agregar/actualizar un harness se toca el catálogo y se saca release del
instalador (tradeoff aceptado: simple y versionado con el instalador).

---

## 4. Flujo de uso

### 4.1 Instalación (una vez por máquina)

```
active-stack install  →  TUI (Bubbletea)
  1. Detectar OS/arch/agentes/deps
  2. Elegir agente(s)
  3. Elegir modo (Lite / Full / Custom)
  4. Resolver catálogo → árbol de dependencias de harnesses
  5. Backup de configs existentes
  6. Instalar deps externas (Engram, OpenSpec) + harnesses
  7. Inyectar configs por agente (merge por markers)
  8. Verificar (health checks)
```

### 4.2 Fundación del proyecto (una vez por proyecto)

Un **único comando** que orquesta la fase de fundación con lazy-loading de
skills, en este orden (evolución de la skill `active-orchestrator`):

```
1. openspec init        (solo la carpeta openspec/)
2. kb-creator           (armar knowledge-base/)
3. roadmap-generator    (armar CHANGES.md)
4. find-skill           (instalar skills faltantes según stack)
5. agent-instruction    (armar AGENTS.md/CLAUDE.md con todas las referencias)
```

### 4.3 Ciclo iterativo (por change)

`explore → propose → apply → verify → archive`, vía OPSX. El orquestador
delega a sub-agentes; OpenSpec CLI es la fuente de verdad del estado.

---

## 5. Arquitectura del instalador (Go + Bubbletea)

Se **porta** la infraestructura probada del active-stack actual (260+ tests):

```
cmd/active-stack/            entrypoint CLI
internal/
  system/                detección OS/arch/WSL/Termux, deps, guards   [PORT]
  catalog/               parseo del harnesses.yaml embebido           [NEW]
  model/                 tipos de dominio (harness, agente, modo)     [PORT/adapt]
  planner/               grafo de dependencias, orden, review payload [PORT]
  agents/                adapters por agente (claude/opencode/...)    [PORT, slim]
  harness/               install/inject por tipo de harness           [NEW]
    skill/  config/  external/
  filemerge/             merge por markers (inyectar sin pisar)       [PORT]
  backup/                snapshot + restore de configs                [PORT]
  pipeline/              ejecución por etapas + rollback              [PORT]
  verify/                health checks post-install                   [PORT]
  tui/                   Bubbletea (sin theme cosmético del viejo)    [PORT, slim]
assets/                  catálogo + configs bundleadas (sdd-orchestrator, etc.)
```

`[PORT]` = traer del repo actual y limpiar. `[NEW]` = construir.
Se descarta: `components/gga`, `components/theme`, `components/persona`
(marketing), `components/permissions` se mantiene (seguridad no es opcional).

---

## 6. Roadmap de construcción (incrementos)

1. **Esqueleto + fundación** ← (en curso) repo, go.mod, .gitignore, este doc, openspec init.
2. **Modelo + catálogo**: tipos `Harness`, parser de `harnesses.yaml`, catálogo inicial.
3. **Port de infra**: system, filemerge, backup, planner desde el active-stack actual (limpios).
4. **Adapters slim**: claude + opencode primero (P0), resto después.
5. **Harness installers**: `skill` (clone+copy), `external` (engram/openspec), `config` (sdd-orchestrator componible).
6. **TUI**: flujo Lite/Full/Custom leyendo catálogo.
7. **active-orchestrator como orquestador de fundación**: lazy-load del flujo §4.2.
8. **Verify + E2E**: health checks + Docker E2E portados.

Cada incremento será un **change OPSX** (dogfooding) en `openspec/` (ignorado por git).

---

## 7. Decisiones tomadas y abiertas

### Resueltas
- **Fetch de skills (C-08) — mixto por tipo**: skills propias (active-orchestrator, etc.)
  → `git clone` del repo; terceros (find-skill, skill-creator) → `npx skills add`;
  core openspec (atadas a la versión del orquestador) → embebidas. El modelo de
  skill necesita un campo de método de fetch (clone/npx/embed), análogo a
  `External.Method`.
- **active-orchestrator — activador modular**: orquesta la fase de fundación y activa/llama
  los módulos (kb-creator, roadmap-generator, etc.) según cuáles estén instalados/
  activos (lazy-loading, §4.2). No es una skill monolítica.
- **Toggles del `sdd-orchestrator`**: el texto sale de los assets del stack viejo.
  base (siempre) + `delegation` (en base) + `model-routing` (bloque ya marcado) +
  `engram` (engram-protocol) + `tdd` (strict-tdd) + `governance` (niveles, Etapa 4).

### Follow-ups / abiertas
- ~~**Fix engram download**~~ → **RESUELTO**: `model.External` ya separa `Repo`
  (`Gentleman-Programming/engram`) del `Pkg` (ver `internal/model/harness.go` y el catálogo).
- ~~**Nombre del repo remoto en GitHub**~~ → **RESUELTO** (2026-05-30): el viejo
  `JuanCruzRobledo/jr-stack` (repo fuente original) queda **archivado como legacy**; este conserva `active-stack`.
- ~~**Mapeo harness↔modo**~~ → **RESUELTO** (CHANGES.md C-20): Lite = sustrato,
  Full = sustrato + fundación guiada, Custom = todos. `active-orchestrator` movido a Full-only.
- **Empaquetado de skills de terceros** → comando confirmado (`npx skills add <owner/repo> --skill <name>`);
  falta corregir el installer (CHANGES.md **C-22**, 3 bugs en `npx.go`). Skill de Vercel = `find-skills` (plural).
- **Custom + `permissions`** → decisión: NO desactivable (security-first); falta implementar (CHANGES.md **C-21**).
