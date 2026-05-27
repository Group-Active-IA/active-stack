# CLAUDE.md — JR Stack (instalador del harness metodológico)

> Constitución operativa del proyecto. Versión canónica (Claude).
> Espejo de `AGENTS.md`: **si modificás uno, actualizás el otro.**
> Leé este archivo antes de cualquier acción no trivial.

> **Nota sobre el orquestador SDD.** El orquestador SDD global (coordinación,
> delegación a sub-agentes, model-routing por fase, protocolo Engram, governance)
> ya vive en tu `~/.claude/CLAUDE.md`. **Este archivo NO lo duplica.** Si necesitás
> el detalle del orquestador, referenciá ese archivo global. Acá viven solo las
> reglas **específicas de este proyecto**.

---

## 1. Stack y topología

- **Lenguaje**: Go 1.26.
- **TUI**: Bubbletea + Lipgloss (sin el theme cosmético del repo viejo).
- **Distribución**: binario único, cross-platform — Windows, macOS, Linux, WSL y Termux.
- **Entrypoint**: `cmd/jr-stack/`.
- **Catálogo**: embebido en el binario vía `//go:embed` (`internal/catalog/harnesses.yaml`).

Qué es esto: un **instalador methodology-first**. Materializa el `MANUAL-METODOLOGICO.md`.
Un comando (`jr-stack install`) instala/configura el sustrato (harnesses); un único
orquestador de fundación deja el proyecto listo para el ciclo OPSX
(`explore → propose → apply → verify → archive`).

**NO es** (fuera de scope, decisión firme — ver ARCHITECTURE.md §1):
GGA (code review en commit), themes/statusline/keybindings, persona como producto
de marketing, framing "supercharge any agent" / relación con Gentleman.Dots.

### Estructura de paquetes (ARCHITECTURE.md §5)

```
cmd/jr-stack/            entrypoint CLI
internal/
  system/                detección OS/arch/WSL/Termux, deps, guards   [PORT]
  catalog/               parseo del harnesses.yaml embebido           [NEW]  ← existe
  model/                 tipos de dominio (harness, agente, modo)     [NEW]  ← existe
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

`[PORT]` = traer del repo viejo (`E:\ESCRITORIO\programar\2026\jr-stack`) y limpiar.
`[NEW]` = construir. Se descarta: `components/gga`, `components/theme`,
`components/persona` (marketing). `components/permissions` **se mantiene** (la
seguridad no es opcional).

---

## 2. Modelo de dominio (resumen)

Detalle real en `internal/model/harness.go`. Tipos clave:

- **`Harness`** — módulo instalable/configurable. Campos: `ID`, `Name`, `Type`,
  `Source` (skill), `External` (tool), `Toggles` (config), `InstallModes`,
  `DependsOn`, `Agents`. Métodos: `InMode(mode)`, `SupportsAgent(agent)`.
- **`HarnessType`** — cómo se materializa un harness:
  - `skill` → `SKILL.md` + assets, clonado de un repo y copiado al dir de skills del agente.
  - `config` → texto/archivos bundleados que configuran el agente (ej. `sdd-orchestrator`, `permissions`).
  - `external` → binario/servicio de terceros que instalamos/configuramos pero no son nuestros (Engram, OpenSpec, Context7).
- **`InstallMode`** — `lite` | `full` | `custom`. Convención: un harness Lite lista
  `[lite, full]` (Full incluye a Lite); uno exclusivo de Full lista `[full]`; Custom matchea todos.
- **`Agent`** — `claude`, `opencode`, `gemini`, `codex`, `cursor`, `vscode`, `windsurf`, `antigravity`.
- **Catálogo** (`internal/catalog`) — `Load()` parsea y valida el YAML embebido;
  un catálogo malformado es error de build/release (falla ruidoso). Métodos:
  `ByID`, `ForMode`, `ForAgent`.

**El harness `sdd-orchestrator` es clave**: es de tipo `config` y se compone a
partir de **toggles modulares** (`tdd`, `engram`, `model-routing`, `delegation`,
`governance`). El resultado es el bloque de instrucciones del orquestador que se
inyecta en `CLAUDE.md`/`AGENTS.md` del proyecto destino.

---

## 3. Reglas críticas NO negociables

Formuladas como "NUNCA X → hacer Y". Estas reglas son duras: violarlas invalida el trabajo.

- **NUNCA pisar config del usuario sin backup** → SIEMPRE snapshot vía `internal/backup` antes de escribir.
- **SIEMPRE inyectar con markers idempotentes** → usar `internal/filemerge` (merge por markers); reinstalar no debe duplicar bloques.
- **NUNCA hardcodear paths de agente** → resolver SIEMPRE vía el adapter del agente (`internal/agents`).
- **NUNCA commitear sin pedido explícito** del responsable; *conventional commits* exclusivamente; **NUNCA** atribución de coautoría a la IA.
- **NUNCA meter en el repo lo que se saca** → GGA, theme, statusline, keybindings, persona-marketing y el framing "supercharge any agent" NO van. La persona puede sobrevivir solo como harness `config` opcional, sin envoltorio de marketing.
- **SIEMPRE limpiar leftovers `gentle-ai` / `gentle-stack` / `Gentleman.Dots`** al portar código del repo viejo (paths, strings, nombres, branding).
- **NUNCA instalar "repos"** → el instalador instala **harnesses**, y cada harness sabe cómo se materializa (skill/config/external).
- **NUNCA editar `harnesses.yaml` sin pasar por `catalog.Load()` validando** → un catálogo inválido rompe el release.
- **NUNCA build después de cambios** salvo pedido explícito (regla del operador).
- **SIEMPRE marcar `(TBD)`** cuando una decisión no está tomada; nunca inventar.

---

## 4. Governance por dominio

El nivel de autonomía es proporcional a la criticidad del paquete tocado.

| Nivel | Dominios (paquetes) | Comportamiento del agente |
|---|---|---|
| **ALTO** | `backup` (snapshot/restore), `filemerge` (merge por markers), `pipeline` (rollback) | Propone y espera review. Pueden **destruir config del usuario**. |
| **MEDIO** | `agents` (adapters por agente), `harness/*` (installers) | Implementa con checkpoints. |
| **BAJO** | `catalog`, `model`, `tui` | Autonomía completa si los tests pasan. |

> No hay dominios CRÍTICO en este proyecto (no manejamos auth/billing/secrets de
> producción). El máximo es ALTO porque destruir la config del usuario es el peor
> daño posible del instalador.

---

## 5. Mapa de navegación ("Necesito X → Leer Y")

| Necesito… | Leer / mirar |
|---|---|
| El blueprint del diseño (fuente de verdad) | `ARCHITECTURE.md` |
| La metodología (etapas, governance, flujo) | `../MANUAL-METODOLOGICO.md` |
| Roadmap de changes, deps, camino crítico | `CHANGES.md` |
| Qué harnesses existen y de qué tipo | `internal/catalog/harnesses.yaml` |
| Los tipos de dominio | `internal/model/harness.go` |
| Cargar/validar el catálogo | `internal/catalog/catalog.go` |
| Infra a portar (backup, filemerge, planner, agents, pipeline, verify, tui) | repo viejo: `E:\ESCRITORIO\programar\2026\jr-stack\internal\` |
| Configs bundleadas (sdd-orchestrator por agente) | repo viejo: `internal/assets/` |
| Estado real de los changes | `openspec` CLI (`openspec list`, `openspec status`) — fuente de verdad |
| El orquestador SDD global | `~/.claude/CLAUDE.md` (NO duplicar acá) |

---

## 6. Notas

- `openspec/` está versionado-ignorado por git (dogfooding interno).
- Cada incremento del roadmap es un change OPSX (ver `CHANGES.md`).
- Skills de dominio relevantes para este repo: `go-testing` (tests Go + Bubbletea TUI).
